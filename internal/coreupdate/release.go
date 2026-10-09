package coreupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const maxArchiveBytes int64 = 512 << 20

type releaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
	Size int64  `json:"size"`
}
type release struct {
	ID         int64          `json:"id"`
	Tag        string         `json:"tag_name"`
	URL        string         `json:"html_url"`
	Draft      bool           `json:"draft"`
	Prerelease bool           `json:"prerelease"`
	Assets     []releaseAsset `json:"assets"`
}
type selection struct {
	preview  Preview
	asset    releaseAsset
	checksum string
}

func (m *Manager) get(ctx context.Context, address string) (*http.Response, error) {
	u, err := url.Parse(address)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
		return nil, errors.New("update download URL must use HTTPS without credentials")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, errors.New("invalid update URL")
	}
	req.Header.Set("User-Agent", "BloraPanel-Updater/1")
	req.Header.Set("Accept", "application/vnd.github+json")
	response, err := m.client.Do(req)
	if err != nil {
		return nil, errors.New("could not contact release source; check the network or mirror")
	}
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		return nil, fmt.Errorf("release source returned HTTP %d", response.StatusCode)
	}
	return response, nil
}

func (m *Manager) readURL(ctx context.Context, address string, limit int64) ([]byte, error) {
	response, err := m.get(ctx, address)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, errors.New("release download was interrupted")
	}
	if int64(len(data)) > limit {
		return nil, errors.New("release metadata exceeds its size limit")
	}
	return data, nil
}

func (m *Manager) discover(ctx context.Context, source Config) (*selection, error) {
	var best *release
	var bestVersion semanticVersion
	for page := 1; page <= 10; page++ {
		data, err := m.readURL(ctx, source.APIURL+"/releases?per_page=100&page="+strconv.Itoa(page), 4<<20)
		if err != nil {
			return nil, err
		}
		var releases []release
		if json.Unmarshal(data, &releases) != nil {
			return nil, errors.New("release source returned an invalid release list")
		}
		for _, candidate := range releases {
			version, err := parseVersion(candidate.Tag)
			if candidate.Draft || err != nil || (source.Channel == "stable" && (candidate.Prerelease || len(version.pre) > 0)) {
				continue
			}
			if best == nil || compareVersions(version, bestVersion) > 0 {
				value := candidate
				best = &value
				bestVersion = version
			}
		}
		if len(releases) < 100 {
			break
		}
	}
	if best == nil {
		return nil, errors.New("no published release is available for this channel")
	}
	commit, err := m.resolveTag(ctx, source.APIURL, best.Tag)
	if err != nil {
		return nil, err
	}
	preview := Preview{Revision: commit, CurrentRevision: m.options.CurrentRevision, Version: strings.TrimPrefix(best.Tag, "v"), Tag: best.Tag, ReleaseURL: best.URL, CheckedAt: time.Now().UTC()}
	if _, err := secureURL(best.URL); err != nil {
		preview.ReleaseURL = ""
	}
	current, versionErr := parseVersion(m.options.CurrentVersion)
	if versionErr != nil && m.options.CurrentVersion != "development" {
		preview.Reason = "Installed version cannot be identified; use a maintenance installation."
	}
	if versionErr == nil {
		comparison := compareVersions(bestVersion, current)
		preview.UpToDate = comparison <= 0
		if comparison < 0 {
			preview.Reason = "The selected channel only contains older releases; automatic downgrades are not allowed."
		}
		if comparison == 0 && commit != m.options.CurrentRevision && m.options.CurrentRevision != "" {
			preview.Reason = "This release version has a different source revision; reissued versions require maintenance."
		}
	}
	if commit == m.options.CurrentRevision {
		preview.UpToDate = true
	}
	selected := &selection{preview: preview}
	platform := runtime.GOOS + "-" + runtime.GOARCH
	suffix := ".tar.gz"
	if runtime.GOOS == "windows" {
		suffix = ".zip"
	}
	var archive, checksums, metadata releaseAsset
	seen := map[string]bool{}
	for _, asset := range best.Assets {
		if seen[asset.Name] {
			return nil, errors.New("release contains duplicate asset names")
		}
		seen[asset.Name] = true
		if asset.Name == "SHA256SUMS" {
			checksums = asset
		}
		if asset.Name == "CORE-UPDATE.json" {
			metadata = asset
		}
		if asset.Name == "blora-"+m.options.Component+"-"+best.Tag+"-"+platform+suffix || asset.Name == "blora-"+m.options.Component+"-"+strings.TrimPrefix(best.Tag, "v")+"-"+platform+suffix {
			if archive.Name != "" {
				return nil, errors.New("release contains ambiguous installation archives")
			}
			archive = asset
		}
	}
	if archive.URL == "" || checksums.URL == "" || metadata.URL == "" {
		selected.preview.Reason = "Release is missing a platform archive, checksums or managed-update metadata; use a maintenance installation."
		return selected, nil
	}
	if archive.Size <= 0 || archive.Size > maxArchiveBytes {
		selected.preview.Reason = "The release installation exceeds the supported download size."
		return selected, nil
	}
	rawHashes, err := m.readURL(ctx, checksums.URL, 1<<20)
	if err != nil {
		return nil, err
	}
	hashes, err := parseChecksums(string(rawHashes))
	if err != nil {
		return nil, err
	}
	if hashes[archive.Name] == "" || hashes[metadata.Name] == "" {
		return nil, errors.New("release checksums do not cover the installation and compatibility metadata")
	}
	rawMetadata, err := m.readURL(ctx, metadata.URL, 1<<20)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(rawMetadata)
	if hex.EncodeToString(sum[:]) != hashes[metadata.Name] {
		return nil, errors.New("release compatibility metadata checksum mismatch")
	}
	var manifest Manifest
	if err := decodeStrict(rawMetadata, &manifest); err != nil {
		return nil, errors.New("release compatibility metadata is invalid")
	}
	selected.asset, selected.checksum = archive, hashes[archive.Name]
	selected.preview.Manifest, selected.preview.DownloadBytes = manifest, archive.Size
	if strings.TrimPrefix(manifest.Version, "v") != strings.TrimPrefix(best.Tag, "v") || manifest.Revision != commit {
		selected.preview.Reason = "Release tag and compatibility metadata disagree; do not apply this release."
		return selected, nil
	}
	selected.preview.Version = manifest.Version
	if err := compatibleManifest(manifest); err != nil {
		selected.preview.Reason = "Release changes protocol, database or instance ownership; a maintenance update is required."
		return selected, nil
	}
	if selected.preview.Reason == "" {
		selected.preview.Compatible = true
	}
	return selected, nil
}

func (m *Manager) resolveTag(ctx context.Context, api, tag string) (string, error) {
	if tag == "" || len(tag) > 128 || strings.ContainsAny(tag, "/\\\x00\r\n") {
		return "", errors.New("release has an invalid tag")
	}
	address := api + "/git/ref/tags/" + url.PathEscape(tag)
	for depth := 0; depth < 6; depth++ {
		data, err := m.readURL(ctx, address, 1<<20)
		if err != nil {
			return "", err
		}
		var object struct {
			Object struct {
				Type string `json:"type"`
				SHA  string `json:"sha"`
			} `json:"object"`
		}
		if json.Unmarshal(data, &object) != nil || !revisionPattern.MatchString(object.Object.SHA) {
			return "", errors.New("release tag did not resolve to a source commit")
		}
		if object.Object.Type == "commit" {
			return object.Object.SHA, nil
		}
		if object.Object.Type != "tag" {
			return "", errors.New("release tag points to an unsupported object")
		}
		address = api + "/git/tags/" + object.Object.SHA
	}
	return "", errors.New("release tag nesting exceeds its limit")
}

func parseChecksums(raw string) (map[string]string, error) {
	result := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 || !validArchiveName(parts[1]) || strings.Contains(parts[1], "/") || len(parts[0]) != 64 {
			return nil, errors.New("release checksum file is invalid")
		}
		if _, err := hex.DecodeString(parts[0]); err != nil {
			return nil, errors.New("release checksum file is invalid")
		}
		if _, exists := result[parts[1]]; exists {
			return nil, errors.New("release checksum file contains duplicate entries")
		}
		result[parts[1]] = strings.ToLower(parts[0])
	}
	return result, nil
}

func (m *Manager) downloadArchive(ctx context.Context, asset releaseAsset, checksum, destination string) error {
	response, err := m.get(ctx, asset.URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	f, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	h := sha256.New()
	progress := &progressWriter{manager: m, last: time.Now()}
	n, copyErr := io.Copy(io.MultiWriter(f, h, progress), io.LimitReader(response.Body, maxArchiveBytes+1))
	syncErr := f.Sync()
	closeErr := f.Close()
	if copyErr != nil {
		return errors.New("installation download was interrupted; the current installation was not changed")
	}
	if syncErr != nil || closeErr != nil {
		return errors.New("could not persist installation download")
	}
	if n > maxArchiveBytes || n != asset.Size {
		return errors.New("release installation download size mismatch")
	}
	if hex.EncodeToString(h.Sum(nil)) != checksum {
		return errors.New("release installation checksum mismatch")
	}
	if err := m.downloadProgress(n); err != nil {
		return errors.New("could not persist installation progress")
	}
	return nil
}

type progressWriter struct {
	manager *Manager
	n       int64
	last    time.Time
}

func (w *progressWriter) Write(p []byte) (int, error) {
	w.n += int64(len(p))
	if time.Since(w.last) >= time.Second {
		if err := w.manager.downloadProgress(w.n); err != nil {
			return 0, err
		}
		w.last = time.Now()
	}
	return len(p), nil
}

var versionExpression = regexp.MustCompile(`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z.-]+))?(?:\+([0-9A-Za-z.-]+))?$`)

type semanticVersion struct {
	major, minor, patch uint64
	pre                 []string
}

func parseVersion(raw string) (semanticVersion, error) {
	var v semanticVersion
	parts := versionExpression.FindStringSubmatch(raw)
	if parts == nil {
		return v, errors.New("invalid semantic release version")
	}
	var err error
	v.major, err = strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return v, err
	}
	v.minor, err = strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return v, err
	}
	v.patch, err = strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		return v, err
	}
	if parts[4] != "" {
		v.pre = strings.Split(parts[4], ".")
		for _, part := range v.pre {
			if part == "" {
				return v, errors.New("invalid prerelease identifier")
			}
			if digits(part) && len(part) > 1 && part[0] == '0' {
				return v, errors.New("invalid numeric prerelease identifier")
			}
		}
	}
	if parts[5] != "" {
		for _, part := range strings.Split(parts[5], ".") {
			if part == "" {
				return v, errors.New("invalid build metadata identifier")
			}
		}
	}
	return v, nil
}
func digits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
func compareVersions(a, b semanticVersion) int {
	for _, pair := range [][2]uint64{{a.major, b.major}, {a.minor, b.minor}, {a.patch, b.patch}} {
		if pair[0] < pair[1] {
			return -1
		}
		if pair[0] > pair[1] {
			return 1
		}
	}
	if len(a.pre) == 0 && len(b.pre) > 0 {
		return 1
	}
	if len(b.pre) == 0 && len(a.pre) > 0 {
		return -1
	}
	for i := 0; i < len(a.pre) && i < len(b.pre); i++ {
		x, y := a.pre[i], b.pre[i]
		if x == y {
			continue
		}
		nx, ny := digits(x), digits(y)
		if nx && ny {
			if len(x) < len(y) {
				return -1
			}
			if len(x) > len(y) {
				return 1
			}
			if x < y {
				return -1
			}
			return 1
		}
		if nx {
			return -1
		}
		if ny {
			return 1
		}
		if x < y {
			return -1
		}
		return 1
	}
	if len(a.pre) < len(b.pre) {
		return -1
	}
	if len(a.pre) > len(b.pre) {
		return 1
	}
	return 0
}

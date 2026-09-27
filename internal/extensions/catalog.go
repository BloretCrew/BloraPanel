package extensions

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const catalogMaxPackageBytes = 16 << 20

// CatalogSource is the narrow source contract used by the Master. A source
// may be a local release directory or a remote HTTPS registry; package
// authenticity is still checked by extensions.Manager at installation time.
type CatalogSource interface {
	List() ([]CatalogEntry, error)
	Load(appID, version string) (Package, error)
}

// Catalog is a local, read-only package source. Release tooling can populate
// it after its own signing step, while the Master still verifies every
// package at install.
type Catalog struct {
	root string
}

type CatalogEntry struct {
	Manifest         Manifest `json:"manifest"`
	SHA256           string   `json:"sha256"`
	SignaturePresent bool     `json:"signaturePresent"`
	Size             int64    `json:"size"`
}

type catalogEnvelope struct {
	Manifest  Manifest `json:"manifest"`
	SHA256    string   `json:"sha256"`
	Payload   []byte   `json:"payload"`
	Signature []byte   `json:"signature,omitempty"`
}

func NewCatalog(root string) (*Catalog, error) {
	if root == "" {
		return nil, errors.New("extension catalog root required")
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	return &Catalog{root: filepath.Clean(root)}, nil
}

func (c *Catalog) packagePath(appID, version string) (string, error) {
	if !appIDPattern.MatchString(appID) || !versionPattern.MatchString(version) {
		return "", errors.New("invalid catalog package identity")
	}
	return filepath.Join(c.root, appID, version+".json"), nil
}

func (c *Catalog) Load(appID, version string) (Package, error) {
	path, err := c.packagePath(appID, version)
	if err != nil {
		return Package{}, err
	}
	stat, err := os.Lstat(path)
	if err != nil {
		return Package{}, err
	}
	if !stat.Mode().IsRegular() || stat.Size() > catalogMaxPackageBytes {
		return Package{}, errors.New("catalog package is not a regular bounded file")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return Package{}, err
	}
	var in catalogEnvelope
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		return Package{}, fmt.Errorf("catalog package JSON: %w", err)
	}
	var trailing any
	if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) {
		return Package{}, errors.New("catalog package has trailing JSON")
	}
	if in.Manifest.AppID != appID || in.Manifest.PackageVersion != version {
		return Package{}, errors.New("catalog package identity mismatch")
	}
	digest := sha256.Sum256(in.Payload)
	if !strings.EqualFold(hex.EncodeToString(digest[:]), in.SHA256) {
		return Package{}, errors.New("catalog package integrity check failed")
	}
	return Package{Manifest: in.Manifest, SHA256: in.SHA256, Payload: in.Payload, Signature: in.Signature}, nil
}

// RemoteCatalog reads the same envelope as Catalog over an administrator
// configured HTTPS base URL. The registry is intentionally pull-only: it
// cannot install anything by itself, and every acquired package is passed to
// Manager.Validate before it reaches the extension directory.
type RemoteCatalog struct {
	base   *url.URL
	client *http.Client
}

// NewRemoteCatalog creates a bounded HTTPS catalog client. The base URL is a
// directory containing index.json and <appId>/<version>.json. Redirects are
// rejected so a valid-looking registry cannot silently move acquisition to a
// different host. A custom client is useful for a deployment's pinned CA or
// for an isolated TLS test; its redirect policy is replaced with the safe one.
func NewRemoteCatalog(rawURL string, custom *http.Client) (*RemoteCatalog, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("extension catalog URL must be an HTTPS directory without credentials or query")
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/"
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}, Proxy: nil}, Timeout: 20 * time.Second}
	if custom != nil {
		copy := *custom
		client = &copy
		if client.Timeout <= 0 {
			client.Timeout = 20 * time.Second
		}
	}
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return errors.New("extension catalog redirects are not allowed")
	}
	return &RemoteCatalog{base: u, client: client}, nil
}

func (c *RemoteCatalog) endpoint(parts ...string) string {
	u := *c.base
	path := strings.TrimSuffix(u.Path, "/")
	for _, part := range parts {
		path += "/" + part
	}
	u.Path = path
	u.RawPath = ""
	return u.String()
}

func (c *RemoteCatalog) get(path string, limit int64) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("extension catalog returned HTTP %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, errors.New("extension catalog response is too large")
	}
	return b, nil
}

func (c *RemoteCatalog) List() ([]CatalogEntry, error) {
	b, err := c.get(c.endpoint("index.json"), 1<<20)
	if err != nil {
		return nil, err
	}
	var in struct {
		Items []CatalogEntry `json:"items"`
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		return nil, fmt.Errorf("remote catalog index JSON: %w", err)
	}
	var trailing any
	if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, errors.New("remote catalog index has trailing JSON")
	}
	if len(in.Items) > 4096 {
		return nil, errors.New("remote catalog index has too many entries")
	}
	seen := make(map[string]struct{}, len(in.Items))
	for i := range in.Items {
		x := &in.Items[i]
		if !appIDPattern.MatchString(x.Manifest.AppID) || !versionPattern.MatchString(x.Manifest.PackageVersion) || len(x.Manifest.AppID) > 128 || len(x.Manifest.PackageVersion) > 64 {
			return nil, errors.New("remote catalog index contains an invalid package identity")
		}
		if len(x.SHA256) != sha256.Size*2 {
			return nil, errors.New("remote catalog index contains an invalid package digest")
		}
		if _, err := hex.DecodeString(x.SHA256); err != nil {
			return nil, errors.New("remote catalog index contains an invalid package digest")
		}
		if x.Size < 0 || x.Size > catalogMaxPackageBytes {
			return nil, errors.New("remote catalog index contains an invalid package size")
		}
		key := x.Manifest.AppID + "@" + x.Manifest.PackageVersion
		if _, ok := seen[key]; ok {
			return nil, errors.New("remote catalog index contains a duplicate package")
		}
		seen[key] = struct{}{}
	}
	sort.Slice(in.Items, func(i, j int) bool {
		if in.Items[i].Manifest.AppID != in.Items[j].Manifest.AppID {
			return in.Items[i].Manifest.AppID < in.Items[j].Manifest.AppID
		}
		return compareVersion(in.Items[i].Manifest.PackageVersion, in.Items[j].Manifest.PackageVersion) > 0
	})
	return in.Items, nil
}

func (c *RemoteCatalog) Load(appID, version string) (Package, error) {
	if !appIDPattern.MatchString(appID) || !versionPattern.MatchString(version) {
		return Package{}, errors.New("invalid catalog package identity")
	}
	b, err := c.get(c.endpoint(url.PathEscape(appID), url.PathEscape(version)+".json"), catalogMaxPackageBytes)
	if err != nil {
		return Package{}, err
	}
	var in catalogEnvelope
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		return Package{}, fmt.Errorf("remote catalog package JSON: %w", err)
	}
	var trailing any
	if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) {
		return Package{}, errors.New("remote catalog package has trailing JSON")
	}
	if in.Manifest.AppID != appID || in.Manifest.PackageVersion != version {
		return Package{}, errors.New("catalog package identity mismatch")
	}
	digest := sha256.Sum256(in.Payload)
	if !strings.EqualFold(hex.EncodeToString(digest[:]), in.SHA256) {
		return Package{}, errors.New("catalog package integrity check failed")
	}
	return Package{Manifest: in.Manifest, SHA256: in.SHA256, Payload: in.Payload, Signature: in.Signature}, nil
}

// Publish adds a package to the local source without replacing an existing
// version. Release tooling may call this after its own signing step; install
// still performs the authoritative capability, compatibility and trust checks.
func (c *Catalog) Publish(p Package) error {
	path, err := c.packagePath(p.Manifest.AppID, p.Manifest.PackageVersion)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(p.Payload)
	if !strings.EqualFold(hex.EncodeToString(digest[:]), p.SHA256) {
		return errors.New("catalog package integrity check failed")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	b, err := json.Marshal(catalogEnvelope{Manifest: p.Manifest, SHA256: p.SHA256, Payload: p.Payload, Signature: p.Signature})
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".catalog-")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err = tmp.Chmod(0600); err == nil {
		_, err = tmp.Write(b)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err := os.Link(tmpName, path); err != nil {
		return fmt.Errorf("publish catalog package: %w", err)
	}
	return syncDir(filepath.Dir(path))
}

func (c *Catalog) List() ([]CatalogEntry, error) {
	apps, err := os.ReadDir(c.root)
	if err != nil {
		return nil, err
	}
	entries := make([]CatalogEntry, 0)
	for _, app := range apps {
		if !app.IsDir() || !appIDPattern.MatchString(app.Name()) {
			continue
		}
		versions, err := os.ReadDir(filepath.Join(c.root, app.Name()))
		if err != nil {
			return nil, err
		}
		for _, version := range versions {
			if version.IsDir() || filepath.Ext(version.Name()) != ".json" {
				continue
			}
			name := version.Name()[:len(version.Name())-len(filepath.Ext(version.Name()))]
			if !versionPattern.MatchString(name) || !version.Type().IsRegular() {
				continue
			}
			pkg, err := c.Load(app.Name(), name)
			if err != nil {
				// A malformed catalog entry is not silently presented as
				// installable; leave it out of the browse response and let a
				// direct acquisition report the precise error.
				continue
			}
			stat, _ := version.Info()
			var size int64
			if stat != nil {
				size = stat.Size()
			}
			entries = append(entries, CatalogEntry{Manifest: pkg.Manifest, SHA256: pkg.SHA256, SignaturePresent: len(pkg.Signature) != 0, Size: size})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Manifest.AppID != entries[j].Manifest.AppID {
			return entries[i].Manifest.AppID < entries[j].Manifest.AppID
		}
		return compareVersion(entries[i].Manifest.PackageVersion, entries[j].Manifest.PackageVersion) > 0
	})
	return entries, nil
}

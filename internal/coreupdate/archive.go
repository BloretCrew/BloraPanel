package coreupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const maxExpandedBytes int64 = 2 << 30
const maxFileBytes int64 = 512 << 20
const maxArchiveEntries = 10000

type fileRecord struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	Mode   string `json:"mode"`
	SHA256 string `json:"sha256"`
}
type installationManifest struct {
	FormatVersion int          `json:"formatVersion"`
	Version       string       `json:"version"`
	Component     string       `json:"component"`
	Platform      string       `json:"platform"`
	Files         []fileRecord `json:"files"`
}
type extractedFile struct {
	size int64
	hash string
	mode int64
}

func validArchiveName(name string) bool {
	if name == "" || len(name) > 4096 || strings.HasPrefix(name, "/") || strings.ContainsAny(name, "\\:\x00\r\n\t") || path.Clean(name) != name {
		return false
	}
	for _, part := range strings.Split(name, "/") {
		if part == "" || part == "." || part == ".." || len(part) > 255 || strings.TrimRight(part, ". ") != part {
			return false
		}
		for _, c := range part {
			if c < 32 || c == 127 {
				return false
			}
		}
		stem := strings.ToUpper(strings.SplitN(part, ".", 2)[0])
		if stem == "CON" || stem == "PRN" || stem == "AUX" || stem == "NUL" || (len(stem) == 4 && (strings.HasPrefix(stem, "COM") || strings.HasPrefix(stem, "LPT")) && stem[3] >= '1' && stem[3] <= '9') {
			return false
		}
	}
	return true
}

func decodeStrict(data []byte, value any) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(value); err != nil {
		return err
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return errors.New("expected one JSON object")
	}
	return nil
}

func extractVerified(archive, name, destination, component, version, platform string) (installationManifest, error) {
	var manifest installationManifest
	if err := os.Mkdir(destination, 0700); err != nil {
		return manifest, err
	}
	files := map[string]extractedFile{}
	seen := map[string]bool{}
	var total int64
	entries := 0
	write := func(name string, size int64, mode int64, directory bool, reader io.Reader) error {
		entries++
		if entries > maxArchiveEntries {
			return errors.New("release archive has too many entries")
		}
		name = strings.TrimSuffix(name, "/")
		if !validArchiveName(name) {
			return errors.New("release archive contains an unsafe path")
		}
		lower := strings.ToLower(name)
		if seen[lower] {
			return errors.New("release archive contains colliding paths")
		}
		seen[lower] = true
		rootName := strings.ToLower(strings.SplitN(name, "/", 2)[0])
		for _, reserved := range []string{".git", ".local", ".codex", ".agents", "state", "updates", "master.json", "daemon.json", "initial-password", "node.enrollment"} {
			if rootName == reserved {
				return errors.New("release archive attempts to replace private configuration or state")
			}
		}
		target := filepath.Join(destination, filepath.FromSlash(name))
		if directory {
			return os.MkdirAll(target, 0700)
		}
		if size < 0 || size > maxFileBytes || total > maxExpandedBytes-size {
			return errors.New("release archive exceeds its expanded size limit")
		}
		total += size
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		h := sha256.New()
		n, copyErr := io.Copy(io.MultiWriter(out, h), io.LimitReader(reader, size+1))
		syncErr := out.Sync()
		closeErr := out.Close()
		if copyErr != nil || syncErr != nil || closeErr != nil || n != size {
			return errors.New("release archive contains a truncated or invalid file")
		}
		files[name] = extractedFile{size: n, hash: hex.EncodeToString(h.Sum(nil)), mode: mode & 0777}
		return nil
	}
	if strings.HasSuffix(name, ".zip") {
		reader, err := zip.OpenReader(archive)
		if err != nil {
			return manifest, errors.New("release ZIP is invalid")
		}
		defer reader.Close()
		for _, file := range reader.File {
			if file.Mode()&os.ModeSymlink != 0 || (!file.Mode().IsRegular() && !file.Mode().IsDir()) {
				return manifest, errors.New("release archive contains a link or special file")
			}
			if file.UncompressedSize64 > uint64(maxFileBytes) {
				return manifest, errors.New("release ZIP file exceeds its size limit")
			}
			input, err := file.Open()
			if err != nil {
				return manifest, errors.New("release ZIP entry is invalid")
			}
			err = write(file.Name, int64(file.UncompressedSize64), int64(file.Mode().Perm()), file.Mode().IsDir(), input)
			input.Close()
			if err != nil {
				return manifest, err
			}
		}
	} else if strings.HasSuffix(name, ".tar.gz") {
		file, err := os.Open(archive)
		if err != nil {
			return manifest, err
		}
		defer file.Close()
		compressed, err := gzip.NewReader(file)
		if err != nil {
			return manifest, errors.New("release gzip archive is invalid")
		}
		defer compressed.Close()
		reader := tar.NewReader(compressed)
		for {
			header, err := reader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return manifest, errors.New("release tar archive is invalid")
			}
			if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA && header.Typeflag != tar.TypeDir {
				return manifest, errors.New("release archive contains a link or special file")
			}
			if err := write(header.Name, header.Size, header.Mode, header.Typeflag == tar.TypeDir, reader); err != nil {
				return manifest, err
			}
		}
		// Python's release packager pads tar records with zeros. Read through
		// that bounded padding to the gzip trailer, rejecting hidden trailing
		// streams as well as corruption after tar's end marker.
		tail, err := io.ReadAll(io.LimitReader(compressed, (64<<10)+1))
		if err != nil {
			return manifest, errors.New("release gzip checksum mismatch")
		}
		if len(tail) > 64<<10 {
			return manifest, errors.New("release archive has excessive trailing data")
		}
		for _, value := range tail {
			if value != 0 {
				return manifest, errors.New("release archive has nonzero trailing data")
			}
		}
	} else {
		return manifest, errors.New("unsupported release archive format")
	}
	data, err := os.ReadFile(filepath.Join(destination, "MANIFEST.json"))
	if err != nil || len(data) > 1<<20 {
		return manifest, errors.New("release installation manifest is missing or too large")
	}
	if err := decodeStrict(data, &manifest); err != nil {
		return manifest, errors.New("release installation manifest is invalid")
	}
	if manifest.FormatVersion != 1 || manifest.Component != component || manifest.Version != version || manifest.Platform != platform {
		return manifest, errors.New("release installation has the wrong component, version or platform")
	}
	if len(manifest.Files) > maxArchiveEntries || len(files) != len(manifest.Files)+1 {
		return manifest, errors.New("release installation manifest does not cover every file")
	}
	listed := map[string]bool{}
	for _, expected := range manifest.Files {
		if expected.Path == "MANIFEST.json" || !validArchiveName(expected.Path) || listed[expected.Path] {
			return manifest, errors.New("release manifest contains invalid or duplicate paths")
		}
		listed[expected.Path] = true
		actual, exists := files[expected.Path]
		mode, err := strconv.ParseInt(expected.Mode, 8, 32)
		if !exists || err != nil || (mode != 0644 && mode != 0755) || actual.mode != mode || actual.size != expected.Size || actual.hash != strings.ToLower(expected.SHA256) {
			return manifest, errors.New("release installation file does not match its manifest")
		}
		if err := os.Chmod(filepath.Join(destination, filepath.FromSlash(expected.Path)), os.FileMode(mode)); err != nil {
			return manifest, err
		}
	}
	return manifest, nil
}

type runtimeInfo struct {
	FormatVersion     int    `json:"formatVersion"`
	Component         string `json:"component"`
	Version           string `json:"version"`
	Revision          string `json:"revision"`
	ProtocolVersion   uint32 `json:"protocolVersion"`
	SchemaVersion     int    `json:"schemaVersion"`
	SchemaFingerprint string `json:"schemaFingerprint"`
	PreserveInstances bool   `json:"preserveInstances"`
	ManagedRestart    bool   `json:"managedRestart"`
}

type cappedWriter struct{ buffer bytes.Buffer }

func (w *cappedWriter) Bytes() []byte { return w.buffer.Bytes() }

func (w *cappedWriter) Write(p []byte) (int, error) {
	if w.buffer.Len()+len(p) > 64<<10 {
		return 0, errors.New("runtime information exceeds size limit")
	}
	return w.buffer.Write(p)
}

func probeExecutable(ctx context.Context, binary string) (runtimeInfo, error) {
	var info runtimeInfo
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "--core-update-info")
	cmd.Dir = filepath.Dir(binary)
	cmd.Stderr = io.Discard
	var output cappedWriter
	cmd.Stdout = &output
	if cmd.Run() != nil {
		return info, errors.New("replacement runtime compatibility probe failed")
	}
	if decodeStrict(output.Bytes(), &info) != nil {
		return info, errors.New("replacement runtime compatibility information is invalid")
	}
	return info, nil
}

func validateRuntime(info runtimeInfo, component string, manifest Manifest) error {
	if info.FormatVersion != 1 || info.Component != component || info.Version != manifest.Version || info.Revision != manifest.Revision || info.ProtocolVersion != manifest.ProtocolVersion || info.SchemaVersion != manifest.SchemaVersion || info.SchemaFingerprint != manifest.SchemaFingerprint || !info.PreserveInstances || !info.ManagedRestart {
		return errors.New("replacement binary does not match its managed-update compatibility metadata")
	}
	return nil
}

func validateSourceNotice(path, revision string) error {
	data, err := os.ReadFile(path)
	if err != nil || len(data) > 64<<10 {
		return errors.New("release source provenance is missing")
	}
	if !strings.Contains(string(data), "Source revision: "+revision+"\n") || !strings.Contains(string(data), "Modified working tree: false\n") {
		return errors.New("release source provenance does not match its tag or contains uncommitted code")
	}
	return nil
}

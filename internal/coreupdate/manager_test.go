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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"blora.dev/panel/internal/storage"
)

const testVersion = "0.2.0-beta.2"

var testRevision = strings.Repeat("b", 40)

// This subprocess behaves like the real no-side-effect compatibility endpoint.
// Its executable is downloaded, hash-verified, extracted and launched by the
// updater, rather than mocking the executable-launch path.
func TestMain(m *testing.M) {
	if len(os.Args) == 2 && os.Args[1] == "--core-update-info" {
		_ = json.NewEncoder(os.Stdout).Encode(runtimeInfo{FormatVersion: 1, Component: "daemon", Version: testVersion, Revision: testRevision, ProtocolVersion: 1, SchemaVersion: storage.SchemaVersion(), SchemaFingerprint: storage.SchemaFingerprint(), PreserveInstances: true, ManagedRestart: true})
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func testManifest() Manifest {
	return Manifest{FormatVersion: 1, Version: testVersion, Revision: testRevision, ProtocolVersion: 1, SchemaVersion: storage.SchemaVersion(), SchemaFingerprint: storage.SchemaFingerprint(), PeerProtocolMin: 1, PeerProtocolMax: 1, PreserveInstances: true}
}

type fixtureFile struct {
	name     string
	data     []byte
	mode     int64
	typeflag byte
}

func archiveBytes(t *testing.T, files []fixtureFile, component, version, platform string, zipFormat bool, editManifest func(*installationManifest)) []byte {
	t.Helper()
	manifest := installationManifest{FormatVersion: 1, Version: version, Component: component, Platform: platform}
	for _, file := range files {
		sum := sha256.Sum256(file.data)
		manifest.Files = append(manifest.Files, fileRecord{Path: file.name, Size: int64(len(file.data)), Mode: fmt.Sprintf("%04o", file.mode), SHA256: hex.EncodeToString(sum[:])})
	}
	if editManifest != nil {
		editManifest(&manifest)
	}
	raw, _ := json.Marshal(manifest)
	files = append(files, fixtureFile{name: "MANIFEST.json", data: raw, mode: 0644})
	var output bytes.Buffer
	if zipFormat {
		writer := zip.NewWriter(&output)
		for _, file := range files {
			header := &zip.FileHeader{Name: file.name, Method: zip.Deflate}
			mode := os.FileMode(file.mode)
			if file.typeflag == tar.TypeSymlink {
				mode |= os.ModeSymlink
			}
			header.SetMode(mode)
			entry, err := writer.CreateHeader(header)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = entry.Write(file.data); err != nil {
				t.Fatal(err)
			}
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
	} else {
		compressed, err := gzip.NewWriterLevel(&output, gzip.BestSpeed)
		if err != nil {
			t.Fatal(err)
		}
		writer := tar.NewWriter(compressed)
		for _, file := range files {
			typeflag := file.typeflag
			if typeflag == 0 {
				typeflag = tar.TypeReg
			}
			size := int64(len(file.data))
			if typeflag == tar.TypeSymlink {
				size = 0
			}
			if err := writer.WriteHeader(&tar.Header{Name: file.name, Mode: file.mode, Size: size, Typeflag: typeflag, Linkname: "../../outside"}); err != nil {
				t.Fatal(err)
			}
			if size > 0 {
				if _, err := writer.Write(file.data); err != nil {
					t.Fatal(err)
				}
			}
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		if err := compressed.Close(); err != nil {
			t.Fatal(err)
		}
	}
	return output.Bytes()
}

type releaseFixture struct {
	server         *httptest.Server
	manager        *Manager
	archive        []byte
	manifest       Manifest
	corruptArchive bool
	slowCheck      chan struct{}
}

func newReleaseFixture(t *testing.T, ready func(context.Context, Prepared) error) *releaseFixture {
	t.Helper()
	manifest := testManifest()
	suffix := ""
	zipFormat := runtime.GOOS == "windows"
	if zipFormat {
		suffix = ".exe"
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	notice := []byte("Source revision: " + testRevision + "\nModified working tree: false\n")
	metadata, _ := json.Marshal(manifest)
	extension := ".tar.gz"
	if zipFormat {
		extension = ".zip"
	}
	assetName := "blora-daemon-" + testVersion + "-" + runtime.GOOS + "-" + runtime.GOARCH + extension
	archive := archiveBytes(t, []fixtureFile{{name: "blora-daemon" + suffix, data: binary, mode: 0755}, {name: "SOURCE-REVISION.txt", data: notice, mode: 0644}, {name: "CORE-UPDATE.json", data: metadata, mode: 0644}}, "daemon", testVersion, runtime.GOOS+"-"+runtime.GOARCH, zipFormat, nil)
	f := &releaseFixture{archive: archive, manifest: manifest}
	f.server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases":
			if f.slowCheck != nil {
				select {
				case <-f.slowCheck:
				case <-r.Context().Done():
					return
				}
			}
			_ = json.NewEncoder(w).Encode([]release{{ID: 1, Tag: "v" + testVersion, URL: "https://github.com/BloretCrew/BloraPanel/releases/tag/v" + testVersion, Prerelease: true, Assets: []releaseAsset{{Name: assetName, URL: f.server.URL + "/archive", Size: int64(len(archive))}, {Name: "CORE-UPDATE.json", URL: f.server.URL + "/metadata"}, {Name: "SHA256SUMS", URL: f.server.URL + "/checksums"}}}})
		case "/git/ref/tags/v" + testVersion:
			fmt.Fprintf(w, `{"object":{"type":"tag","sha":"%s"}}`, strings.Repeat("a", 40))
		case "/git/tags/" + strings.Repeat("a", 40):
			fmt.Fprintf(w, `{"object":{"type":"commit","sha":"%s"}}`, testRevision)
		case "/metadata":
			raw, _ := json.Marshal(f.manifest)
			_, _ = w.Write(raw)
		case "/checksums":
			raw, _ := json.Marshal(f.manifest)
			metadataSum := sha256.Sum256(raw)
			archiveSum := sha256.Sum256(archive)
			fmt.Fprintf(w, "%x  CORE-UPDATE.json\n%x  %s\n", metadataSum, archiveSum, assetName)
		case "/archive":
			if f.corruptArchive {
				_, _ = w.Write(append([]byte{0}, archive[1:]...))
			} else {
				_, _ = w.Write(archive)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(f.server.Close)
	manager, err := New(Options{Root: filepath.Join(t.TempDir(), "updates"), Component: "daemon", CurrentRevision: strings.Repeat("c", 40), CurrentVersion: "0.1.0-beta.1", Config: Config{APIURL: f.server.URL}, Ready: ready})
	if err != nil {
		t.Fatal(err)
	}
	client := f.server.Client()
	client.CheckRedirect = manager.client.CheckRedirect
	manager.client = client
	f.manager = manager
	t.Cleanup(manager.Close)
	return f
}

func waitState(t *testing.T, m *Manager, state string) Job {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		status := m.Status()
		if status.Job != nil && status.Job.State == state {
			return *status.Job
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("job did not reach %s: %+v", state, m.Status())
	return Job{}
}

func TestPrebuiltReleaseWithoutDeveloperToolsAndActivationConfirmation(t *testing.T) {
	ready := make(chan Prepared, 1)
	f := newReleaseFixture(t, func(ctx context.Context, p Prepared) error { ready <- p; return nil })
	preview, err := f.manager.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !preview.Compatible || preview.Revision != testRevision || preview.UpToDate {
		t.Fatalf("bad preview: %+v", preview)
	}
	// Production updates must not require Git, Go, Node, npm or a shell.
	t.Setenv("PATH", t.TempDir())
	job, err := f.manager.Start("actor:request-1", preview.Revision)
	if err != nil {
		t.Fatal(err)
	}
	var prepared Prepared
	select {
	case prepared = <-ready:
	case <-time.After(20 * time.Second):
		t.Fatalf("no prepared installation: %+v", f.manager.Status())
	}
	if prepared.Revision != testRevision || !strings.HasPrefix(prepared.Directory, filepath.Join(f.manager.options.Root, "revisions")+string(os.PathSeparator)) {
		t.Fatalf("bad preparation: %+v", prepared)
	}
	if got := f.manager.Status().Job; got.State != "running" || got.Phase != "activating" {
		t.Fatalf("preparation incorrectly reported success: %+v", got)
	}
	if got := f.manager.Status().Job; got.DownloadedBytes != int64(len(f.archive)) || got.DownloadedBytes != got.TotalBytes {
		t.Fatalf("download progress was not persisted: %+v", got)
	}
	if _, err := os.Stat(prepared.Binary); err != nil {
		t.Fatal(err)
	}
	if replay, err := f.manager.Start(job.ID, job.Revision); err != nil || replay.ID != job.ID {
		t.Fatalf("request replay: %+v %v", replay, err)
	}
	if _, err := f.manager.Start(job.ID, strings.Repeat("d", 40)); err == nil {
		t.Fatal("request ID reused for a different update")
	}
	options := f.manager.options
	options.CurrentRevision = testRevision
	options.CurrentVersion = testVersion
	replacement, err := New(options)
	if err != nil {
		t.Fatal(err)
	}
	defer replacement.Close()
	if err := replacement.ConfirmActivation(); err != nil {
		t.Fatal(err)
	}
	completed := replacement.Status().Job
	if completed.State != "completed" || completed.Phase != "complete" {
		t.Fatalf("activation confirmation missing: %+v", completed)
	}
	if historic, found, err := replacement.Lookup(job.ID, testRevision); err != nil || !found || historic.State != "completed" {
		t.Fatalf("durable replay: %+v %v %v", historic, found, err)
	}
}

func TestDownloadChecksumFailureNeverActivates(t *testing.T) {
	called := false
	f := newReleaseFixture(t, func(context.Context, Prepared) error { called = true; return nil })
	f.corruptArchive = true
	preview, err := f.manager.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.manager.Start("corruption", preview.Revision); err != nil {
		t.Fatal(err)
	}
	job := waitState(t, f.manager, "failed")
	if called || !strings.Contains(job.Detail, "checksum mismatch") {
		t.Fatalf("corrupt release activated: %+v called=%v", job, called)
	}
}

func TestChangedSchemaBlockedBeforeDownload(t *testing.T) {
	f := newReleaseFixture(t, func(context.Context, Prepared) error { return errors.New("unexpected activation") })
	f.manifest.SchemaFingerprint = strings.Repeat("0", 64)
	preview, err := f.manager.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if preview.Compatible || !strings.Contains(preview.Reason, "maintenance") {
		t.Fatalf("schema change accepted: %+v", preview)
	}
	if _, err := f.manager.Start("schema", preview.Revision); !errors.Is(err, ErrIncompatible) {
		t.Fatalf("incompatible Start: %v", err)
	}
}

func TestDowngradesReissuedVersionsAndDevelopmentAdoption(t *testing.T) {
	f := newReleaseFixture(t, nil)
	for _, scenario := range []struct {
		version    string
		compatible bool
		upToDate   bool
	}{
		{version: "development", compatible: true},
		{version: "0.3.0", upToDate: true},
		{version: testVersion, upToDate: true},
		{version: "unidentified build"},
	} {
		f.manager.options.CurrentVersion = scenario.version
		preview, err := f.manager.Check(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if preview.Compatible != scenario.compatible || preview.UpToDate != scenario.upToDate {
			t.Fatalf("version %q: %+v", scenario.version, preview)
		}
	}
	f.manager.options.CurrentVersion = "0.1.0-beta.1"
	f.manifest.Revision = strings.Repeat("d", 40)
	preview, err := f.manager.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if preview.Compatible || !strings.Contains(preview.Reason, "disagree") {
		t.Fatalf("retagged release accepted: %+v", preview)
	}
}

func TestHTTPSRedirectPolicyRejectsCleartext(t *testing.T) {
	m, err := New(Options{Root: t.TempDir(), Component: "daemon"})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://127.0.0.1:1/private", http.StatusFound)
	}))
	defer server.Close()
	client := server.Client()
	client.CheckRedirect = m.client.CheckRedirect
	m.client = client
	if _, err := m.readURL(context.Background(), server.URL, 1<<10); err == nil {
		t.Fatal("HTTPS release redirected to cleartext")
	}
}

func TestBackgroundCheckSerializationAndCancellation(t *testing.T) {
	f := newReleaseFixture(t, nil)
	f.slowCheck = make(chan struct{})
	status, err := f.manager.BeginCheck()
	if err != nil || !status.Checking {
		t.Fatalf("async check: %+v %v", status, err)
	}
	if _, err := f.manager.SetSource(context.Background(), Config{}); !errors.Is(err, ErrBusy) {
		t.Fatalf("source changed while checking: %v", err)
	}
	if _, err := f.manager.BeginCheck(); !errors.Is(err, ErrBusy) {
		t.Fatalf("parallel check: %v", err)
	}
	f.manager.Close()
	deadline := time.Now().Add(5 * time.Second)
	for f.manager.Status().Checking && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if status = f.manager.Status(); status.Checking || status.CheckError == "" {
		t.Fatalf("cancelled check not settled: %+v", status)
	}
}

func TestActivationFailureRecoveryAllowsRetry(t *testing.T) {
	root := t.TempDir()
	job := Job{ID: "old-attempt", Revision: testRevision, State: "running", Phase: "activating", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := writeJSON(filepath.Join(root, "job.json"), job); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(root, "activation-error.json"), map[string]string{"revision": testRevision, "error": "failed readiness"}); err != nil {
		t.Fatal(err)
	}
	m, err := New(Options{Root: root, Component: "daemon", CurrentRevision: strings.Repeat("c", 40)})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if m.Status().Job.State != "failed" {
		t.Fatalf("activation left stuck: %+v", m.Status().Job)
	}
	if historical, found, err := m.Lookup(job.ID, testRevision); err != nil || !found || historical.State != "failed" {
		t.Fatalf("recovery not durable: %+v %v %v", historical, found, err)
	}
}

func TestSourceValidationAndPersistence(t *testing.T) {
	for _, source := range []Config{{Repository: "http://github.com/BloretCrew/BloraPanel"}, {Repository: "https://user:secret@github.com/BloretCrew/BloraPanel"}, {APIURL: "https://api.example.invalid/repos/panel?token=secret"}, {Channel: "nightly"}, {Repository: "https://mirror.invalid/blora/panel"}, {Repository: "https://github.com/BloretCrew/BloraPanel/extra"}} {
		if _, err := normalizeConfig(source); err == nil {
			t.Fatalf("unsafe source accepted: %+v", source)
		}
	}
	root := t.TempDir()
	m, err := New(Options{Root: root, Component: "master"})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	configured := Config{Repository: "https://mirror.invalid/blora/panel", Channel: "stable", APIURL: "https://mirror.invalid/api/repos/blora/panel"}
	if _, err := m.SetSource(context.Background(), configured); err != nil {
		t.Fatal(err)
	}
	reopened, err := New(Options{Root: root, Component: "master"})
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if reopened.Status().Source != configured {
		t.Fatalf("source was not persisted: %+v", reopened.Status().Source)
	}
}

func TestSemanticVersionOrdering(t *testing.T) {
	ordered := []string{"0.1.0-alpha", "0.1.0-alpha.1", "0.1.0-alpha.beta", "0.1.0-beta.2", "0.1.0-beta.11", "0.1.0-rc.1", "v0.1.0", "0.1.1", "0.2.0-beta.1", "1.0.0"}
	for i := 1; i < len(ordered); i++ {
		a, err := parseVersion(ordered[i-1])
		if err != nil {
			t.Fatal(err)
		}
		b, err := parseVersion(ordered[i])
		if err != nil {
			t.Fatal(err)
		}
		if compareVersions(a, b) >= 0 {
			t.Fatalf("bad version ordering %s -> %s", ordered[i-1], ordered[i])
		}
	}
	for _, invalid := range []string{"0.01.0", "1.2", "1.2.3-beta.01", "1.2.3-beta..1", "1.2.3+build..1", "1.2.3\n"} {
		if _, err := parseVersion(invalid); err == nil {
			t.Fatalf("invalid version accepted: %s", invalid)
		}
	}
}

func TestArchiveVerificationAndUnsafePaths(t *testing.T) {
	for _, zipFormat := range []bool{false, true} {
		t.Run(fmt.Sprintf("zip=%v", zipFormat), func(t *testing.T) {
			suffix := ".tar.gz"
			if zipFormat {
				suffix = ".zip"
			}
			for _, scenario := range []struct {
				name   string
				files  []fixtureFile
				mutate func(*installationManifest)
				valid  bool
			}{
				{name: "normal", files: []fixtureFile{{name: "blora-daemon", data: []byte("binary"), mode: 0755}}, valid: true},
				{name: "traversal", files: []fixtureFile{{name: "../outside", data: []byte("bad"), mode: 0644}}},
				{name: "backslash", files: []fixtureFile{{name: "docs\\escape", data: []byte("bad"), mode: 0644}}},
				{name: "private state", files: []fixtureFile{{name: "state/database", data: []byte("bad"), mode: 0644}}},
				{name: "case collision", files: []fixtureFile{{name: "README", data: []byte("a"), mode: 0644}, {name: "readme", data: []byte("b"), mode: 0644}}},
				{name: "symlink", files: []fixtureFile{{name: "link", mode: 0755, typeflag: tar.TypeSymlink}}},
				{name: "digest mismatch", files: []fixtureFile{{name: "normal", data: []byte("a"), mode: 0644}}, mutate: func(m *installationManifest) { m.Files[0].SHA256 = strings.Repeat("0", 64) }},
				{name: "extra file", files: []fixtureFile{{name: "normal", data: []byte("a"), mode: 0644}}, mutate: func(m *installationManifest) { m.Files = nil }},
				{name: "wrong platform", files: []fixtureFile{{name: "normal", data: []byte("a"), mode: 0644}}, mutate: func(m *installationManifest) { m.Platform = "wrong" }},
			} {
				t.Run(scenario.name, func(t *testing.T) {
					directory := t.TempDir()
					archive := filepath.Join(directory, "archive"+suffix)
					data := archiveBytes(t, scenario.files, "daemon", testVersion, "linux-amd64", zipFormat, scenario.mutate)
					if err := os.WriteFile(archive, data, 0600); err != nil {
						t.Fatal(err)
					}
					_, err := extractVerified(archive, "archive"+suffix, filepath.Join(directory, "extracted"), "daemon", testVersion, "linux-amd64")
					if (err == nil) != scenario.valid {
						t.Fatalf("valid=%v error=%v", scenario.valid, err)
					}
				})
			}
		})
	}
}

func TestChecksumAndRuntimeMetadataBoundaries(t *testing.T) {
	for _, raw := range []string{strings.Repeat("0", 64) + "  ../outside\n", strings.Repeat("0", 64) + "  a\n" + strings.Repeat("1", 64) + "  a\n", "invalid  file\n"} {
		if _, err := parseChecksums(raw); err == nil {
			t.Fatalf("invalid checksums accepted: %q", raw)
		}
	}
	manifest := testManifest()
	info := runtimeInfo{FormatVersion: 1, Component: "daemon", Version: manifest.Version, Revision: manifest.Revision, ProtocolVersion: 1, SchemaVersion: manifest.SchemaVersion, SchemaFingerprint: manifest.SchemaFingerprint, PreserveInstances: true, ManagedRestart: true}
	if err := validateRuntime(info, "daemon", manifest); err != nil {
		t.Fatal(err)
	}
	info.ManagedRestart = false
	if err := validateRuntime(info, "daemon", manifest); err == nil {
		t.Fatal("unmanaged binary accepted")
	}
	var writer cappedWriter
	if _, err := io.Copy(&writer, strings.NewReader(strings.Repeat("x", (64<<10)+1))); err == nil {
		t.Fatal("runtime output limit not enforced")
	}
}

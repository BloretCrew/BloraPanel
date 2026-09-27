package extensions

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestUserDataIsolationCASMigrationAndRecovery(t *testing.T) {
	root := t.TempDir()
	m, err := New(root, []string{"window.open", "data.read", "data.write"})
	if err != nil {
		t.Fatal(err)
	}
	modulePath := filepath.Join(t.TempDir(), "migration.wasm")
	build := exec.Command("go", "build", "-buildvcs=false", "-o", modulePath, "./testdata/migration")
	build.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm", "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build migration: %v %s", err, out)
	}
	module, err := os.ReadFile(modulePath)
	if err != nil {
		t.Fatal(err)
	}
	bundle, _ := json.Marshal(Bundle{Frontend: "export function start(){}", Backend: module})
	p := pkg("data.example", "1.0.0", append([]byte(BundlePrefix), bundle...))
	p.Manifest.Capabilities = []string{"data.read", "data.write"}
	if _, err := m.Install(p); err != nil {
		t.Fatal(err)
	}
	first, err := m.WriteUserData(p.Manifest.AppID, "alice", "request-1", 0, 1, json.RawMessage(`{"note":"alice"}`))
	if err != nil {
		t.Fatal(err)
	}
	if first.Revision != 1 {
		t.Fatal("bad revision")
	}
	bob, err := m.ReadUserData(p.Manifest.AppID, "bob")
	if err != nil || string(bob.Data) != "{}" {
		t.Fatalf("cross-user leak: %v", err)
	}
	if _, err := m.WriteUserData(p.Manifest.AppID, "bob", "bob-request", 0, 1, json.RawMessage(`{"note":"bob private"}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := m.WriteUserData(p.Manifest.AppID, "alice", "stale", 0, 1, json.RawMessage(`{}`)); !errors.Is(err, ErrDataConflict) {
		t.Fatal("stale write accepted")
	}
	if _, err := m.WriteUserData(p.Manifest.AppID, "alice", "request-2", 1, 1, json.RawMessage(`{"note":"latest"}`)); err != nil {
		t.Fatal(err)
	}
	replay, err := m.WriteUserData(p.Manifest.AppID, "alice", "request-1", 0, 1, json.RawMessage(`{"note":"alice"}`))
	if err != nil || !strings.Contains(string(replay.Data), "latest") {
		t.Fatal("replay overwrote newer data")
	}
	p.Manifest.PackageVersion = "2.0.0"
	p.Manifest.DataSchemaVersion = 2
	if _, err := m.Upgrade(p); err != nil {
		t.Fatal(err)
	}
	got, err := m.ReadUserData(p.Manifest.AppID, "alice")
	if err != nil || got.SchemaVersion != 2 || got.Revision != 3 || !strings.Contains(string(got.Data), "latest") {
		t.Fatalf("migration: %+v %v", got, err)
	}
	bob, err = m.ReadUserData(p.Manifest.AppID, "bob")
	if err != nil || bob.SchemaVersion != 2 || !strings.Contains(string(bob.Data), "bob private") || strings.Contains(string(bob.Data), "latest") {
		t.Fatal("batch migration crossed users")
	}
	p.Manifest.PackageVersion = "3.0.0"
	p.Manifest.DataSchemaVersion = 3
	if _, err := m.Upgrade(p); err == nil {
		t.Fatal("failed guest migration accepted")
	}
	installed, err := m.Get(p.Manifest.AppID)
	if err != nil || installed.Manifest.PackageVersion != "2.0.0" {
		t.Fatal("failed migration replaced package")
	}
	if _, err := m.Rollback(p.Manifest.AppID); err != nil {
		t.Fatal(err)
	}
	got, err = m.ReadUserData(p.Manifest.AppID, "alice")
	if err != nil || got.SchemaVersion != 1 || !strings.Contains(string(got.Data), "latest") {
		t.Fatal("rollback lost data")
	}
	// Simulate interruption after data changed and before package commit.
	tx, err := m.beginTransaction(p.Manifest.AppID)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(m.dataFile(p.Manifest.AppID), []byte("broken incomplete data"), 0600); err != nil {
		t.Fatal(err)
	}
	_ = tx
	reopened, err := New(root, []string{"data.read", "data.write"})
	if err != nil {
		t.Fatal(err)
	}
	got, err = reopened.ReadUserData(p.Manifest.AppID, "alice")
	if err != nil || !strings.Contains(string(got.Data), "latest") {
		t.Fatal("data recovery failed")
	}
	if err := reopened.Uninstall(p.Manifest.AppID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(reopened.dataFile(p.Manifest.AppID)); err != nil {
		t.Fatal("uninstall removed retained data")
	}
}

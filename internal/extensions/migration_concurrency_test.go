package extensions

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMigrationDoesNotBlockRegistryAndRejectsConcurrentWrite(t *testing.T) {
	modulePath := filepath.Join(t.TempDir(), "migration.wasm")
	build := exec.Command("go", "build", "-buildvcs=false", "-o", modulePath, "./testdata/migration")
	build.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm", "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, out)
	}
	module, err := os.ReadFile(modulePath)
	if err != nil {
		t.Fatal(err)
	}
	bundle, _ := json.Marshal(Bundle{Frontend: "export function start(){}", Backend: module})
	m, err := New(t.TempDir(), []string{"data.read", "data.write", "window.open"})
	if err != nil {
		t.Fatal(err)
	}
	p := pkg("migration.example", "1.0.0", append([]byte(BundlePrefix), bundle...))
	p.Manifest.Capabilities = []string{"data.read", "data.write"}
	if _, err := m.Install(p); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Install(pkg("other.example", "1.0.0", []byte("other"))); err != nil {
		t.Fatal(err)
	}
	if _, err := m.WriteUserData(p.Manifest.AppID, "alice", "first", 0, 1, json.RawMessage(`{"note":"old","slow":true}`)); err != nil {
		t.Fatal(err)
	}
	p.Manifest.PackageVersion = "2.0.0"
	p.Manifest.DataSchemaVersion = 2
	finished := make(chan error, 1)
	go func() { _, err := m.Upgrade(p); finished <- err }()
	deadline := time.Now().Add(10 * time.Second)
	for {
		active := false
		if m.mu.TryRLock() {
			active = m.migrationActive
			m.mu.RUnlock()
		}
		if active {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("migration never released registry lock")
		}
		time.Sleep(5 * time.Millisecond)
	}
	access := make(chan error, 1)
	if _, err := m.Upgrade(p); err == nil || !strings.Contains(err.Error(), "another extension data migration") {
		t.Fatalf("concurrent migration not bounded: %v", err)
	}
	go func() {
		if err := m.AuthorizeCapability("other.example", "window.open"); err != nil {
			access <- err
			return
		}
		_, err := m.WriteUserData(p.Manifest.AppID, "alice", "concurrent", 1, 1, json.RawMessage(`{"note":"new"}`))
		access <- err
	}()
	select {
	case err := <-access:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("migration blocked another extension or user write")
	}
	select {
	case err := <-finished:
		if !errors.Is(err, ErrDataConflict) {
			t.Fatalf("expected conflict, got %v", err)
		}
	case <-time.After(65 * time.Second):
		t.Fatal("migration did not terminate")
	}
	current, err := m.Get(p.Manifest.AppID)
	if err != nil || current.Manifest.PackageVersion != "1.0.0" {
		t.Fatal("conflicting migration replaced package")
	}
	data, err := m.ReadUserData(p.Manifest.AppID, "alice")
	if err != nil || data.Revision != 2 || !strings.Contains(string(data.Data), "new") {
		t.Fatal("conflicting migration overwrote new data")
	}
	if _, err := m.Upgrade(p); err != nil {
		t.Fatalf("retry failed: %v", err)
	}
}

//go:build linux

package daemon

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"blora.dev/panel/internal/containers"
)

func TestDaemonRefusesStartupWithUnconfirmedComposeOwnership(t *testing.T) {
	root := t.TempDir()
	m, err := containers.New(containers.Options{StateRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	dir := filepath.Join(root, "cli-runs")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, strings.Repeat("a", 32))
	if err := os.WriteFile(path, []byte(`{"marker":"wrong"}`), 0600); err != nil {
		t.Fatal(err)
	}
	// No runtime or connection exists: Run must stop before it can accept work.
	d := &Daemon{containers: m}
	if err := d.Run(context.Background()); err == nil || !strings.Contains(err.Error(), "recover Compose CLI ownership") {
		t.Fatalf("startup bypassed ownership recovery: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("startup removed unconfirmed evidence", err)
	}
}

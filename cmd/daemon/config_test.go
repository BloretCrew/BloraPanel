package main

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestDaemonSiblingConfigAndPaths(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "daemon.json")
	if err := os.WriteFile(path, []byte(`{"enrollmentFile":"node.enrollment","rotationFile":"rotation","backupRoot":"backup"}`), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := parseConfigAt(nil, io.Discard, base)
	if err != nil || c.StateDir != filepath.Join(base, "state/daemon") || c.MasterURL != "http://127.0.0.1:37861" ||
		c.EnrollmentFile != filepath.Join(base, "node.enrollment") || c.RotationFile != filepath.Join(base, "rotation") ||
		c.BackupRoot != filepath.Join(base, "backup") || c.CAFile != "" {
		t.Fatalf("sibling configuration defaults: %+v, %v", c, err)
	}
	other := t.TempDir()
	c, err = parseConfigAt([]string{"--config", path}, io.Discard, other)
	if err != nil || c.StateDir != filepath.Join(other, "state/daemon") || c.EnrollmentFile != filepath.Join(other, "node.enrollment") {
		t.Fatalf("external configuration paths must still follow executable: %+v, %v", c, err)
	}
}

func TestDaemonConfigErrorsAndHelp(t *testing.T) {
	base := t.TempDir()
	if _, err := parseConfigAt(nil, io.Discard, base); err == nil {
		t.Fatal("missing sibling configuration accepted")
	}
	for _, body := range []string{"null", "[]", "{} {}", `{"unknown":true}`, `{"masterUrl":`} {
		if err := os.WriteFile(filepath.Join(base, "daemon.json"), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := parseConfigAt(nil, io.Discard, base); err == nil {
			t.Fatal("invalid configuration accepted")
		}
	}
	if _, err := parseConfigAt([]string{"--help"}, io.Discard, base); err != flag.ErrHelp {
		t.Fatalf("help: %v", err)
	}
}

package bootstrap

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	child, _, err := ParseServiceChild(os.Args[1:])
	if err != nil {
		os.Exit(82)
	}
	if child != nil {
		if strings.Contains(filepath.Base(os.Args[0]), "fail") {
			os.Exit(83)
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cwd, err := os.Getwd()
		if err != nil || cwd != child.InstallRoot {
			os.Exit(85)
		}
		if child.ReadyAndWait(ctx) != nil {
			os.Exit(84)
		}
		child.WatchStop(ctx, cancel)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func copyServiceTestBinary(t *testing.T, path string) {
	t.Helper()
	b, err := os.ReadFile(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, b, 0700); err != nil {
		t.Fatal(err)
	}
}

func waitServiceFile(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("service did not publish %s", filepath.Base(path))
}

func stopCurrentTestWorker(t *testing.T, root string) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(root, "ready-*.json"))
	if err != nil || len(files) != 1 {
		t.Fatalf("worker readiness: %v %v", files, err)
	}
	token := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(files[0]), "ready-"), ".json")
	if err := serviceWrite(root, "stop-"+token+".json", true); err != nil {
		t.Fatal(err)
	}
}

func TestSupervisorActivatesAndRetainsInstallationRoot(t *testing.T) {
	root := t.TempDir()
	original := filepath.Join(root, "initial")
	copyServiceTestBinary(t, original)
	updates := filepath.Join(root, "updates")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- Supervise(ctx, ServiceOptions{Executable: original, InstallRoot: root, UpdateRoot: updates, ReadyTimeout: 2 * time.Second})
	}()
	deadline := time.Now().Add(8 * time.Second)
	for {
		files, _ := filepath.Glob(filepath.Join(updates, "ready-*.json"))
		if len(files) == 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("initial worker missing")
		}
		time.Sleep(10 * time.Millisecond)
	}
	revision := strings.Repeat("a", 40)
	directory := filepath.Join(updates, "revisions", revision)
	binary := filepath.Join(directory, "replacement")
	copyServiceTestBinary(t, binary)
	if err := RequestActivation(updates, Activation{revision, binary, directory}); err != nil {
		t.Fatal(err)
	}
	stopCurrentTestWorker(t, updates)
	waitServiceFile(t, filepath.Join(updates, "active.json"))
	active, err := readActivation(updates, "active.json")
	if err != nil || active.Binary != binary {
		t.Fatalf("activation: %+v %v", active, err)
	}
	if err := Supervise(ctx, ServiceOptions{Executable: original, InstallRoot: root, UpdateRoot: updates}); err == nil {
		t.Fatal("duplicate state writer accepted")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("graceful launcher stop timed out")
	}
}

func TestSupervisorCandidateFailureRetainsOldVersion(t *testing.T) {
	root := t.TempDir()
	original := filepath.Join(root, "initial")
	copyServiceTestBinary(t, original)
	updates := filepath.Join(root, "updates")
	if err := os.MkdirAll(updates, 0700); err != nil {
		t.Fatal(err)
	}
	revision := strings.Repeat("b", 40)
	directory := filepath.Join(updates, "revisions", revision)
	binary := filepath.Join(directory, "fail-candidate")
	copyServiceTestBinary(t, binary)
	if err := RequestActivation(updates, Activation{revision, binary, directory}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- Supervise(ctx, ServiceOptions{Executable: original, InstallRoot: root, UpdateRoot: updates, ReadyTimeout: time.Second})
	}()
	waitServiceFile(t, filepath.Join(updates, "activation-error.json"))
	if _, err := os.Stat(filepath.Join(updates, "active.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("failed candidate became active: %v", err)
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("launcher did not stop")
	}
}

func TestActivationRefusesUnownedExecutable(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "binary")
	copyServiceTestBinary(t, outside)
	if err := RequestActivation(root, Activation{strings.Repeat("a", 40), outside, filepath.Dir(outside)}); err == nil {
		t.Fatal("accepted executable outside update revisions")
	}
}

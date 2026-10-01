//go:build linux

package runlog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func duplicateFile(f *os.File) (*os.File, error) {
	fd, err := unix.Dup(int(f.Fd()))
	if err != nil {
		return nil, err
	}
	unix.CloseOnExec(fd)
	return os.NewFile(uintptr(fd), "duplicate-writer"), nil
}

func waitTestProcessExit(pid int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, state, err := procBirth(pid)
		if os.IsNotExist(err) || errors.Is(err, unix.ESRCH) || state == 'Z' || state == 'X' {
			return nil
		}
		if err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("own test process %d did not exit", pid)
}

// The launcher is a short-lived substitute for the daemon process; the parent
// test never inherits its business pipe, and reconnects only through disk+IPC.
func TestLauncherProcess(t *testing.T) {
	root := os.Getenv("BLORA_TEST_LOG_LAUNCHER")
	if root == "" {
		return
	}
	c, err := Start(context.Background(), Options{Root: root, RunID: "orphan-run", Command: helperCommand()})
	if err != nil {
		os.Exit(41)
	}
	cmd := exec.Command("/bin/sh", "-c", "printf 'before-daemon-exit\\n'; sleep 0.5; printf '\\033[32m守护进程已退出，业务仍写入\\033[0m\\n'; sleep 0.1; printf 'after-daemon-exit\\n'")
	cmd.Stdout = c.Stdout()
	cmd.Stderr = c.Stdout()
	if err = cmd.Start(); err != nil {
		os.Exit(42)
	}
	if err = c.ReleaseWriter(); err != nil {
		os.Exit(43)
	}
	_ = json.NewEncoder(os.Stdout).Encode(map[string]int{"businessPID": cmd.Process.Pid})
	// Abrupt process exit: no daemon goroutines can continue capturing output.
	os.Exit(0)
}

func TestDaemonExitDoesNotCloseBusinessPipe(t *testing.T) {
	root := t.TempDir()
	launcher := exec.Command(os.Args[0], "-test.run=^TestLauncherProcess$")
	launcher.Env = append(os.Environ(), "BLORA_TEST_LOG_LAUNCHER="+root, "GORACE=atexit_sleep_ms=0")
	out, err := launcher.Output()
	if err != nil {
		t.Fatalf("launcher: %v", err)
	}
	var result struct {
		BusinessPID int `json:"businessPID"`
	}
	if err = json.Unmarshal(out, &result); err != nil || result.BusinessPID <= 1 {
		t.Fatalf("launcher result: %q %v", out, err)
	}
	c, err := Open(root, "orphan-run")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Finish(context.Background()) })
	deadline := time.Now().Add(5 * time.Second)
	for {
		s, err := c.Status(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if s.Phase == "complete" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("orphan capture did not finish")
		}
		time.Sleep(10 * time.Millisecond)
	}
	b, err := c.Read(context.Background(), 0, 64<<10)
	if err != nil {
		t.Fatal(err)
	}
	var raw []byte
	for _, event := range b.Events {
		raw = append(raw, event.Data...)
	}
	if !strings.Contains(string(raw), "before-daemon-exit") || !strings.Contains(string(raw), "\x1b[32m守护进程已退出，业务仍写入\x1b[0m") || !strings.Contains(string(raw), "after-daemon-exit") {
		t.Fatalf("output lost after launcher exit: %q", raw)
	}
	if err = c.Finish(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Logf("launcher exited before later output; business PID %d completed and replay has %d raw bytes", result.BusinessPID, len(raw))
}

func TestStorageFailureKeepsDraining(t *testing.T) {
	c := startCapture(t, 64<<10)
	archive := filepath.Join(c.root, "archive")
	if err := os.Rename(archive, archive+"-before-fault"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(archive, []byte("controlled test storage fault"), 0600); err != nil {
		t.Fatal(err)
	}
	finished := make(chan error, 1)
	go func() { _, err := fmt.Fprint(c.Stdout(), strings.Repeat("still-running\n", 100000)); finished <- err }()
	select {
	case err := <-finished:
		if err != nil {
			t.Fatalf("business pipe lost after disk error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("storage failure blocked business")
	}
	s, err := c.Status(context.Background())
	if err != nil || s.Phase != "degraded" || s.Diagnostic == "" {
		t.Fatalf("missing degraded state: %+v %v", s, err)
	}
	_ = c.ReleaseWriter()
	if err = c.Finish(context.Background()); err == nil {
		t.Fatal("storage failure reported success")
	}
}

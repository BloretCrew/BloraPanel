//go:build windows

package runlog

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func duplicateFile(f *os.File) (*os.File, error) {
	var handle windows.Handle
	err := windows.DuplicateHandle(windows.CurrentProcess(), windows.Handle(f.Fd()), windows.CurrentProcess(), &handle, 0, false, windows.DUPLICATE_SAME_ACCESS)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(handle), "duplicate-writer"), nil
}

func waitTestProcessExit(pid int, timeout time.Duration) error {
	h, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
		return nil
	}
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	state, err := windows.WaitForSingleObject(h, uint32(timeout.Milliseconds()))
	if err != nil {
		return err
	}
	if state != windows.WAIT_OBJECT_0 {
		return fmt.Errorf("own test process %d did not exit", pid)
	}
	return nil
}

func TestWindowsLogBusinessProcess(t *testing.T) {
	if os.Getenv("BLORA_TEST_LOG_BUSINESS") != "1" {
		return
	}
	fmt.Fprint(os.Stdout, "before-daemon-exit\n")
	time.Sleep(500 * time.Millisecond)
	fmt.Fprint(os.Stdout, "\x1b[32mDaemon退出后中文输出\x1b[0m\n")
	os.Exit(0)
}

func TestWindowsLogLauncherProcess(t *testing.T) {
	root := os.Getenv("BLORA_TEST_LOG_LAUNCHER")
	if root == "" {
		return
	}
	c, err := Start(context.Background(), Options{Root: root, RunID: "orphan-run", Command: helperCommand()})
	if err != nil {
		os.Exit(41)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestWindowsLogBusinessProcess$")
	cmd.Env = append(os.Environ(), "BLORA_TEST_LOG_BUSINESS=1")
	cmd.Stdout = c.Stdout()
	cmd.Stderr = c.Stdout()
	if err = cmd.Start(); err != nil {
		os.Exit(42)
	}
	_ = c.ReleaseWriter()
	os.Exit(0)
}

// Execute this on a Windows machine; cross-compilation is not runtime evidence.
func TestWindowsDaemonExitDoesNotCloseBusinessPipe(t *testing.T) {
	root := t.TempDir()
	launcher := exec.Command(os.Args[0], "-test.run=^TestWindowsLogLauncherProcess$")
	launcher.Env = append(os.Environ(), "BLORA_TEST_LOG_LAUNCHER="+root, "GORACE=atexit_sleep_ms=0")
	if err := launcher.Run(); err != nil {
		t.Fatal(err)
	}
	c, err := Open(root, "orphan-run")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Finish(context.Background()) })
	deadline := time.Now().Add(5 * time.Second)
	for {
		s, e := c.Status(context.Background())
		if e != nil {
			t.Fatal(e)
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
	if !strings.Contains(string(raw), "before-daemon-exit") || !strings.Contains(string(raw), "\x1b[32mDaemon退出后中文输出\x1b[0m") {
		t.Fatalf("output lost after daemon exit: %q", raw)
	}
}

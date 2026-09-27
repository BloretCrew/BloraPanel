//go:build windows

package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
	"golang.org/x/sys/windows"
)

func windowsTestOptions(root string) Options {
	return Options{StateRoot: root, JobKeeperCommand: []string{os.Args[0], "-test.run=^TestWindowsJobKeeperProcess$", "--"}}
}

func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == "--job-keeper" {
		if err := RunJobKeeper(os.Args[2:]); err != nil {
			os.Exit(71)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestWindowsJobKeeperProcess(t *testing.T) {
	for i, arg := range os.Args {
		if arg == "--state-root" {
			if err := RunJobKeeper(os.Args[i:]); err != nil {
				os.Exit(71)
			}
			os.Exit(0)
		}
	}
}

func TestWindowsRuntimeHelper(t *testing.T) {
	mode := os.Getenv("BLORA_WINDOWS_TEST_HELPER")
	if mode == "" {
		return
	}
	if mode == "parent" {
		child := exec.Command(os.Args[0], "-test.run=^TestWindowsRuntimeHelper$")
		child.Env = append(os.Environ(), "BLORA_WINDOWS_TEST_HELPER=child")
		child.Stdout = os.Stdout
		child.Stderr = os.Stderr
		if err := child.Start(); err != nil {
			os.Exit(2)
		}
		os.Exit(0)
	}
	time.Sleep(30 * time.Second)
	os.Exit(0)
}

func TestWindowsJobKeepsDescendantsAfterParentExits(t *testing.T) {
	m, err := New(windowsTestOptions(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	r, err := m.Start(context.Background(), model.ID(), model.ID(), model.InstanceConfig{Mode: "native", Command: []string{os.Args[0], "-test.run=^TestWindowsRuntimeHelper$"}, Environment: map[string]string{"BLORA_WINDOWS_TEST_HELPER": "parent"}}, IO{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = m.Stop(context.Background(), r, StopPolicy{Force: true, KillWait: time.Second}, nil)
		_ = m.Cleanup(context.Background(), r)
	})
	deadline := time.Now().Add(5 * time.Second)
	found := false
	for time.Now().Before(deadline) {
		o, e := m.Observe(context.Background(), r)
		if e != nil {
			t.Fatal(e)
		}
		parent := false
		for _, p := range o.Processes {
			if p.PID == r.PID {
				parent = true
			}
		}
		if !parent && !o.Exited && len(o.Processes) > 0 {
			found = true
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !found {
		t.Fatal("could not confirm surviving descendant in original Job")
	}
	o, err := m.Stop(context.Background(), r, StopPolicy{Force: true, KillWait: 2 * time.Second}, nil)
	if err != nil || !o.Exited {
		t.Fatalf("Job termination unconfirmed: %+v %v", o, err)
	}
}

func TestWindowsJobCrashRecoveryPreservesRun(t *testing.T) {
	m, err := New(windowsTestOptions(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	r, err := m.Start(context.Background(), model.ID(), model.ID(), model.InstanceConfig{Command: []string{os.Args[0], "-test.run=^TestWindowsRuntimeHelper$"}, Environment: map[string]string{"BLORA_WINDOWS_TEST_HELPER": "child"}}, IO{})
	if err != nil {
		t.Fatal(err)
	}
	m.mu.Lock()
	live := m.live[r.RunID]
	delete(m.live, r.RunID)
	m.mu.Unlock()
	if err = windows.CloseHandle(live.platform.(*windowsRun).job); err != nil {
		t.Fatal(err)
	}
	other, err := New(m.options)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = other.Stop(context.Background(), r, StopPolicy{Force: true, KillWait: time.Second}, nil)
		_ = other.Cleanup(context.Background(), r)
	})
	_, o, err := other.Recover(context.Background(), r.RunID)
	if err != nil || o.Exited || o.State != "RUNNING" {
		t.Fatalf("keeper did not preserve run after management handle closed: %+v %v", o, err)
	}
	o, err = other.Stop(context.Background(), r, StopPolicy{Force: true, KillWait: 2 * time.Second}, nil)
	if err != nil || !o.Exited {
		t.Fatalf("reopened Job stop unconfirmed: %+v %v", o, err)
	}
}

func TestWindowsRuntimeDaemonProcess(t *testing.T) {
	root := os.Getenv("BLORA_WINDOWS_RUNTIME_DAEMON")
	if root == "" {
		return
	}
	m, err := New(windowsTestOptions(root))
	if err != nil {
		os.Exit(72)
	}
	r, err := m.Start(context.Background(), "survive-daemon", "own-test-instance", model.InstanceConfig{Mode: "native", Command: []string{os.Args[0], "-test.run=^TestWindowsRuntimeHelper$"}, Environment: map[string]string{"BLORA_WINDOWS_TEST_HELPER": "parent", "GORACE": "atexit_sleep_ms=0"}}, IO{})
	if err != nil {
		os.Exit(73)
	}
	if err = json.NewEncoder(os.Stdout).Encode(map[string]string{"runId": r.RunID}); err != nil {
		os.Exit(74)
	}
	// Real process exit closes every daemon Job handle, not just one test handle.
	os.Exit(0)
}

// Windows machine entry point: this launches a real daemon subprocess, which
// exits while a child/grandchild remain. The new Manager must reopen and verify
// the original Job before it can stop those exact processes.
func TestWindowsDaemonExitReopenAndStop(t *testing.T) {
	root := t.TempDir()
	daemon := exec.Command(os.Args[0], "-test.run=^TestWindowsRuntimeDaemonProcess$")
	daemon.Env = append(os.Environ(), "BLORA_WINDOWS_RUNTIME_DAEMON="+root, "GORACE=atexit_sleep_ms=0")
	out, err := daemon.Output()
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		RunID string `json:"runId"`
	}
	if err = json.Unmarshal(out, &result); err != nil || result.RunID != "survive-daemon" {
		t.Fatalf("daemon launch result: %v", err)
	}
	m, err := New(windowsTestOptions(root))
	if err != nil {
		t.Fatal(err)
	}
	r, err := m.Record(result.RunID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = m.Stop(context.Background(), r, StopPolicy{Force: true, KillWait: time.Second}, nil)
		_ = m.Cleanup(context.Background(), r)
	})
	deadline := time.Now().Add(5 * time.Second)
	for {
		_, o, e := m.Recover(context.Background(), r.RunID)
		if e != nil || o.Exited || o.State != "RUNNING" {
			t.Fatalf("run did not survive daemon exit: %+v %v", o, e)
		}
		parent := false
		for _, p := range o.Processes {
			if p.PID == r.PID {
				parent = true
			}
		}
		if !parent && len(o.Processes) > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("could not verify surviving descendant after both parents exited")
		}
		time.Sleep(20 * time.Millisecond)
	}
	k, err := loadKeeper(root, r)
	if err != nil {
		t.Fatal(err)
	}
	keeperHandle, err := keeperProcess(k, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(keeperHandle)
	control := exec.Command(os.Args[0], "-test.run=^TestWindowsRuntimeHelper$")
	control.Env = append(os.Environ(), "BLORA_WINDOWS_TEST_HELPER=child")
	if err = control.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = control.Process.Kill(); _ = control.Wait() }()
	controlHandle, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(control.Process.Pid))
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(controlHandle)
	o, err := m.Stop(context.Background(), r, StopPolicy{Force: true, KillWait: 2 * time.Second}, nil)
	if err != nil || !o.Exited {
		t.Fatalf("confirmed Job stop failed: %+v %v", o, err)
	}
	if err = m.Cleanup(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	state, err := windows.WaitForSingleObject(keeperHandle, 3000)
	if err != nil || state != windows.WAIT_OBJECT_0 {
		t.Fatalf("keeper failed to retire after confirmed empty Job: %d %v", state, err)
	}
	state, err = windows.WaitForSingleObject(controlHandle, 0)
	if err != nil || state != uint32(windows.WAIT_TIMEOUT) {
		t.Fatalf("unrelated test control was stopped: %d %v", state, err)
	}
}

func TestWindowsKeeperBirthMismatchFailsClosed(t *testing.T) {
	m, err := New(windowsTestOptions(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	r, err := m.Start(context.Background(), model.ID(), model.ID(), model.InstanceConfig{Mode: "native", Command: []string{os.Args[0], "-test.run=^TestWindowsRuntimeHelper$"}, Environment: map[string]string{"BLORA_WINDOWS_TEST_HELPER": "child"}}, IO{})
	if err != nil {
		t.Fatal(err)
	}
	k, err := loadKeeper(m.options.StateRoot, r)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = saveKeeper(m.options.StateRoot, k)
		_, _ = m.Stop(context.Background(), r, StopPolicy{Force: true, KillWait: time.Second}, nil)
		_ = m.Cleanup(context.Background(), r)
	})
	wrong := k
	wrong.Birth++
	if err = saveKeeper(m.options.StateRoot, wrong); err != nil {
		t.Fatal(err)
	}
	if _, err = m.Observe(context.Background(), r); !errors.Is(err, ErrUnknown) {
		t.Fatalf("mismatched keeper birth accepted: %v", err)
	}
	if _, err = m.Stop(context.Background(), r, StopPolicy{Force: true, KillWait: time.Second}, nil); !errors.Is(err, ErrUnknown) {
		t.Fatalf("mismatched keeper was used to signal: %v", err)
	}
}

func TestWindowsKeeperStartupFailureCleansEmptyJob(t *testing.T) {
	m, err := New(windowsTestOptions(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	r, err := m.Start(context.Background(), model.ID(), model.ID(), model.InstanceConfig{Mode: "native", Command: []string{os.Args[0]}, Directory: m.options.StateRoot + `\missing-directory`}, IO{})
	if err == nil {
		t.Fatal("missing working directory accepted")
	}
	k, e := loadKeeper(m.options.StateRoot, r)
	if e != nil {
		t.Fatalf("test failed before creating the intended keeper: %v (start: %v)", e, err)
	}
	if handle, e := keeperProcess(k, 0); e == nil {
		windows.CloseHandle(handle)
		t.Fatal("startup failure left keeper alive")
	}
	if handle, e := openJob(r.Unit); !errors.Is(e, windows.ERROR_FILE_NOT_FOUND) {
		if handle != 0 {
			windows.CloseHandle(handle)
		}
		t.Fatalf("startup failure left empty Job: %v", e)
	}
}

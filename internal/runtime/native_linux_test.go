//go:build linux

package runtime

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
	"golang.org/x/sys/unix"
)

func testManager(t *testing.T) *Manager {
	t.Helper()
	m, err := New(Options{StateRoot: t.TempDir(), AllowPGIDFallback: true})
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func testStart(t *testing.T, m *Manager, script string, streams IO) Record {
	t.Helper()
	r, err := m.Start(context.Background(), model.ID(), model.ID(), model.InstanceConfig{Mode: "native", Command: []string{"/bin/sh", "-c", script}}, streams)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, _ = m.Stop(ctx, r, StopPolicy{Force: true, KillWait: time.Second}, nil)
		_ = m.Cleanup(ctx, r)
	})
	return r
}
func waitCondition(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("condition was not reached")
}

func TestParentExitAndInheritedOutputDoesNotMeanRunExit(t *testing.T) {
	m := testManager(t)
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer read.Close()
	defer write.Close()
	r := testStart(t, m, "sleep 30 & exit 0", IO{Stdout: write, Stderr: write})
	waitCondition(t, func() bool { _, e := os.Stat(filepath.Join("/proc", itoa(r.PID))); return errors.Is(e, os.ErrNotExist) })
	o, err := m.Observe(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	if o.Exited || len(o.Processes) == 0 {
		t.Fatalf("parent exit incorrectly reported whole run exited: %+v", o)
	}
	start := time.Now()
	o, err = m.Stop(context.Background(), r, StopPolicy{Force: true, KillWait: time.Second}, nil)
	if err != nil || !o.Exited {
		t.Fatalf("stop failed: %+v %v", o, err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("stop blocked on inherited output")
	}
}

func TestStopEscalatesAndConfirmsAllDescendants(t *testing.T) {
	m := testManager(t)
	ready := filepath.Join(t.TempDir(), "ready")
	r := testStart(t, m, "trap '' TERM; sleep 30 & touch "+ready+"; wait", IO{})
	waitCondition(t, func() bool { _, e := os.Stat(ready); return e == nil })
	var phases []string
	o, err := m.Stop(context.Background(), r, StopPolicy{Grace: 100 * time.Millisecond, KillWait: time.Second, Escalate: true}, func(p string) error { phases = append(phases, p); return nil })
	if err != nil || !o.Exited {
		t.Fatalf("%+v %v", o, err)
	}
	if strings.Join(phases, ",") != "STOPPING,KILLING,VERIFYING_EXIT" {
		t.Fatalf("unexpected phases: %v", phases)
	}
}

func TestStopDeadlineWithoutEscalation(t *testing.T) {
	m := testManager(t)
	ready := filepath.Join(t.TempDir(), "ready")
	r := testStart(t, m, "trap '' TERM; touch "+ready+"; while :; do sleep 30; done", IO{})
	waitCondition(t, func() bool { _, e := os.Stat(ready); return e == nil })
	start := time.Now()
	o, err := m.Stop(context.Background(), r, StopPolicy{Grace: 100 * time.Millisecond}, nil)
	if !errors.Is(err, ErrStopTimeout) || o.Exited || o.State != "STOP_FAILED" {
		t.Fatalf("timeout falsely succeeded: %+v %v", o, err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("unbounded stop")
	}
}

func TestRecoverUsesPersistentIdentityAfterManagerReplacement(t *testing.T) {
	m := testManager(t)
	r := testStart(t, m, "sleep 30", IO{})
	other, err := New(m.options)
	if err != nil {
		t.Fatal(err)
	}
	record, o, err := other.Recover(context.Background(), r.RunID)
	if err != nil || o.Exited || record.StartTicks == 0 {
		t.Fatalf("recovery failed: %+v %+v %v", record, o, err)
	}
	_, err = other.Start(context.Background(), r.RunID, r.InstanceID, model.InstanceConfig{Command: []string{"/bin/true"}}, IO{})
	if err == nil {
		t.Fatal("duplicate run ID accepted")
	}
	o, err = other.Stop(context.Background(), record, StopPolicy{Force: true, KillWait: time.Second}, nil)
	if err != nil || !o.Exited {
		t.Fatalf("recovered stop failed %+v %v", o, err)
	}
}

func TestUnknownOwnershipNeverSignalsProcess(t *testing.T) {
	m := testManager(t)
	r := testStart(t, m, "sleep 30", IO{})
	forged := r
	forged.Token = model.ID()
	forged.PID = 0
	forged.StartTicks = 0
	if _, err := m.Stop(context.Background(), forged, StopPolicy{Force: true}, nil); !errors.Is(err, ErrUnknown) {
		t.Fatalf("unconfirmed group accepted: %v", err)
	}
	o, err := m.Observe(context.Background(), r)
	if err != nil || o.Exited {
		t.Fatalf("real process affected: %+v %v", o, err)
	}
}

func TestExistingPIDWithDifferentBirthIdentityIsNeverAdoptedOrSignalled(t *testing.T) {
	m := testManager(t)
	r := testStart(t, m, "sleep 30", IO{})
	stale := r
	stale.StartTicks++
	if _, err := m.Observe(context.Background(), stale); !errors.Is(err, ErrUnknown) {
		t.Fatalf("existing PID with wrong birth identity was adopted: %v", err)
	}
	if _, err := m.Stop(context.Background(), stale, StopPolicy{Force: true, KillWait: time.Second}, nil); !errors.Is(err, ErrUnknown) {
		t.Fatalf("stale identity was allowed to signal: %v", err)
	}
	actual, err := m.Observe(context.Background(), r)
	if err != nil || actual.Exited || len(actual.Processes) == 0 {
		t.Fatalf("live original run affected by stale record: %+v %v", actual, err)
	}
}

func TestRecoveredIntentFindsRunMarker(t *testing.T) {
	m := testManager(t)
	r := testStart(t, m, "sleep 30", IO{})
	r.PID = 0
	r.PGID = 0
	r.StartTicks = 0
	r.Phase = "intent"
	o, err := m.Observe(context.Background(), r)
	if err != nil || o.Exited || len(o.Processes) == 0 {
		t.Fatalf("intent recovery failed: %+v %v", o, err)
	}
}

func TestPhasePersistenceFailurePreventsSignal(t *testing.T) {
	m := testManager(t)
	r := testStart(t, m, "sleep 30", IO{})
	persistErr := errors.New("disk full")
	_, err := m.Stop(context.Background(), r, StopPolicy{Force: true}, func(string) error { return persistErr })
	if !errors.Is(err, persistErr) {
		t.Fatal(err)
	}
	o, err := m.Observe(context.Background(), r)
	if err != nil || o.Exited {
		t.Fatalf("signal sent after persistence failure: %+v %v", o, err)
	}
}

func TestApplicationInputUsesParentPipeAndDoesNotRepeat(t *testing.T) {
	m := testManager(t)
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer read.Close()
	defer write.Close()
	r := testStart(t, m, "read line; test \"$line\" = stop", IO{Stdin: read, Input: write})
	o, err := m.Stop(context.Background(), r, StopPolicy{Grace: time.Second, Input: "stop\n"}, nil)
	if err != nil || !o.Exited {
		t.Fatalf("application stop input failed: %+v %v", o, err)
	}
}

func TestNativeFallbackRefusesUnenforcedLimits(t *testing.T) {
	m := testManager(t)
	r, err := m.Start(context.Background(), model.ID(), model.ID(), model.InstanceConfig{Command: []string{"/bin/true"}, MemoryBytes: 1 << 20}, IO{})
	if !errors.Is(err, ErrCapability) {
		t.Fatalf("unenforced resource limit accepted: %v", err)
	}
	_, o, err := m.Recover(context.Background(), r.RunID)
	if err != nil || !o.Exited {
		t.Fatalf("rejected launch cannot be reconciled: %+v %v", o, err)
	}
}

func TestCleanupPreservesExitEvidenceAndIsIdempotent(t *testing.T) {
	m := testManager(t)
	r := testStart(t, m, "sleep 30", IO{})
	if _, err := m.Stop(context.Background(), r, StopPolicy{Force: true, KillWait: time.Second}, nil); err != nil {
		t.Fatal(err)
	}
	if err := m.Cleanup(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if err := m.Cleanup(context.Background(), r); err != nil {
		t.Fatalf("cleanup not idempotent: %v", err)
	}
	record, o, err := m.Recover(context.Background(), r.RunID)
	if err != nil || !o.Exited || record.Phase != "cleaned" {
		t.Fatalf("exit evidence lost: %+v %+v %v", record, o, err)
	}
}

func TestPIDBirthMismatchFailsClosed(t *testing.T) {
	m := testManager(t)
	r := testStart(t, m, "sleep 30", IO{})
	forged := r
	forged.StartTicks++
	if _, err := m.Stop(context.Background(), forged, StopPolicy{Force: true}, nil); !errors.Is(err, ErrUnknown) {
		t.Fatalf("mismatched PID birth accepted: %v", err)
	}
}

func TestBlockedApplicationInputHasDeadline(t *testing.T) {
	m := testManager(t)
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer read.Close()
	defer write.Close()
	connection, err := write.SyscallConn()
	if err != nil {
		t.Fatal(err)
	}
	var sizeErr error
	if err = connection.Control(func(fd uintptr) { _, sizeErr = unix.FcntlInt(fd, unix.F_SETPIPE_SZ, 4096) }); err != nil {
		t.Fatal(err)
	}
	if sizeErr != nil {
		t.Fatal(sizeErr)
	}
	r := testStart(t, m, "sleep 30", IO{Stdin: read, Input: write})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err = m.WriteInput(ctx, r, make([]byte, 64<<10))
	if err == nil {
		t.Fatal("expected bounded pipe write failure")
	}
	if time.Since(start) > time.Second {
		t.Fatal("stdin write exceeded deadline")
	}
}

func TestDelegatedCgroupRealLaunch(t *testing.T) {
	root := os.Getenv("BLORA_TEST_CGROUP_ROOT")
	if root == "" {
		t.Skip("environment missing: set BLORA_TEST_CGROUP_ROOT to an explicitly delegated cgroup v2 subtree")
	}
	m, err := New(Options{StateRoot: t.TempDir(), CgroupRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	r := testStart(t, m, "trap '' TERM; sleep 30 & wait", IO{})
	if r.WeakContainment || r.UnitIdentity == 0 {
		t.Fatal("did not start inside real cgroup")
	}
	b, err := os.ReadFile(filepath.Join("/proc", itoa(r.PID), "cgroup"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "blora-"+r.RunID) {
		t.Fatalf("process not assigned before execution: %s", b)
	}
	o, err := m.Stop(context.Background(), r, StopPolicy{Force: true, KillWait: time.Second}, nil)
	if err != nil || !o.Exited {
		t.Fatalf("cgroup kill not confirmed: %+v %v", o, err)
	}
}

func TestDelegatedCgroupResourceLimits(t *testing.T) {
	root := os.Getenv("BLORA_TEST_CGROUP_ROOT")
	if root == "" {
		t.Skip("environment missing: explicitly delegated cgroup v2 subtree required")
	}
	m, err := New(Options{StateRoot: t.TempDir(), CgroupRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	r, err := m.Start(context.Background(), model.ID(), model.ID(), model.InstanceConfig{
		Mode: "native", Command: []string{"/bin/sh", "-c", "while :; do :; done"},
		MemoryBytes: 32 << 20, PidsLimit: 8, CPUQuota: 50000,
	}, IO{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if outcome, err := m.Stop(ctx, r, StopPolicy{Force: true, KillWait: time.Second}, nil); err != nil || !outcome.Exited {
			t.Errorf("cleanup limited run: %+v %v", outcome, err)
		}
		if err := m.Cleanup(ctx, r); err != nil {
			t.Errorf("cleanup limited cgroup: %v", err)
		}
	})
	for file, want := range map[string]string{"memory.max": "33554432", "pids.max": "8", "cpu.max": "50000 100000"} {
		value, err := os.ReadFile(filepath.Join(r.Unit, file))
		if err != nil || strings.TrimSpace(string(value)) != want {
			t.Fatalf("kernel limit %s=%q, want %q: %v", file, value, want, err)
		}
	}
	// A bounded CPU loop must actually be throttled by this run's own cgroup.
	// Memory/PID values above verify configuration, not an OOM/exhaustion test.
	waitCondition(t, func() bool {
		data, err := os.ReadFile(filepath.Join(r.Unit, "cpu.stat"))
		if err != nil {
			return false
		}
		for _, line := range strings.Split(string(data), "\n") {
			fields := strings.Fields(line)
			if len(fields) == 2 && fields[0] == "nr_throttled" && fields[1] != "0" {
				return true
			}
		}
		return false
	})
}

func itoa(n int) string {
	const digits = "0123456789"
	if n == 0 {
		return "0"
	}
	var b [24]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = digits[n%10]
		n /= 10
	}
	return string(b[i:])
}

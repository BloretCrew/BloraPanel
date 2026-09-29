//go:build linux

package runtime

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
)

func limitedKernelProbe(t *testing.T, program string) (*Manager, Record) {
	t.Helper()
	root := os.Getenv("BLORA_TEST_CGROUP_ROOT")
	if root == "" || os.Getenv("BLORA_SYSTEMD_E2E") != "1" {
		t.Skip("requires the isolated systemd/cgroup runner")
	}
	if _, err := os.Stat("/run/blora-systemd-e2e"); err != nil {
		t.Fatal("isolated runner marker missing")
	}
	if _, err := os.Stat("/.dockerenv"); err != nil {
		t.Fatal("exhaustion probes must run in the disposable container")
	}
	m, err := New(Options{StateRoot: t.TempDir(), CgroupRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	ready := filepath.Join(t.TempDir(), "ready")
	// Both probes are finite even if isolation unexpectedly fails: at most a
	// 128 MiB allocation or twelve child processes, each with a short lifetime.
	guard := `import os,pathlib,time
p=pathlib.Path('/sys/fs/cgroup'+pathlib.Path('/proc/self/cgroup').read_text().strip().split('::',1)[1])
assert (p/'memory.max').read_text().strip()=='33554432'
assert (p/'pids.max').read_text().strip()=='8'
deadline=time.monotonic()+5
while not pathlib.Path(` + "'" + ready + "'" + `).exists():
 if time.monotonic()>deadline: raise RuntimeError('probe release timed out')
 time.sleep(.01)
`
	r, err := m.Start(context.Background(), model.ID(), model.ID(), model.InstanceConfig{
		Mode: "native", Command: []string{"/usr/bin/python3", "-c", guard + program},
		MemoryBytes: 32 << 20, PidsLimit: 8,
	}, IO{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if outcome, err := m.Stop(ctx, r, StopPolicy{Force: true, KillWait: time.Second}, nil); err != nil || !outcome.Exited {
			t.Errorf("stop probe: %+v %v", outcome, err)
		}
		if err := m.Cleanup(ctx, r); err != nil {
			t.Errorf("remove probe cgroup: %v", err)
		}
	})
	// Exclude swap only inside this owned test group so OOM is deterministic;
	// this is a test condition, not a claim about the product's swap policy.
	if err := os.WriteFile(filepath.Join(r.Unit, "memory.swap.max"), []byte("0"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ready, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	return m, r
}

func kernelEvent(t *testing.T, unit, file, key string) {
	t.Helper()
	waitCondition(t, func() bool {
		data, err := os.ReadFile(filepath.Join(unit, file))
		if err != nil {
			return false
		}
		for _, line := range strings.Split(string(data), "\n") {
			fields := strings.Fields(line)
			if len(fields) == 2 && fields[0] == key && fields[1] != "0" {
				return true
			}
		}
		return false
	})
}

func TestDelegatedCgroupMemoryExhaustion(t *testing.T) {
	m, r := limitedKernelProbe(t, "block=bytearray(128*1024*1024)\ntime.sleep(2)\n")
	kernelEvent(t, r.Unit, "memory.events", "oom_kill")
	waitCondition(t, func() bool {
		observed, err := m.Observe(context.Background(), r)
		return err == nil && observed.Exited
	})
}

func TestDelegatedCgroupPidsExhaustion(t *testing.T) {
	_, r := limitedKernelProbe(t, `for _ in range(12):
 try: pid=os.fork()
 except BlockingIOError: break
 if pid==0:
  time.sleep(3)
  os._exit(0)
time.sleep(2)
`)
	kernelEvent(t, r.Unit, "pids.events", "max")
	value, err := os.ReadFile(filepath.Join(r.Unit, "pids.current"))
	if err != nil || strings.TrimSpace(string(value)) != "8" {
		t.Fatalf("expected exactly eight admitted processes, got %q: %v", value, err)
	}
}

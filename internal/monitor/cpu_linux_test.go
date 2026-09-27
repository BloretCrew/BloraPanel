package monitor

import (
	"os"
	"slices"
	"testing"
	"time"
)

func TestOwnedProcessCPUSamplingAndIdentity(t *testing.T) {
	pid := os.Getpid()
	identity, err := readProcessCPU(pid)
	if err != nil {
		t.Fatal(err)
	}
	c := New()
	first, err := c.Process(pid, "run-a", identity.birth)
	if err != nil || !slices.Contains(first.Unavailable, "cpu") {
		t.Fatal("first CPU sample must be unavailable", err)
	}
	// Burn CPU in this owned test process so a zero reading cannot pass.
	deadline := time.Now().Add(250 * time.Millisecond)
	for time.Now().Before(deadline) {
	}
	second, err := c.Process(pid, "run-a", identity.birth)
	if err != nil || slices.Contains(second.Unavailable, "cpu") || second.CPUPercent <= 0 || second.CPUPercent > 100 {
		t.Fatalf("real CPU delta unavailable: %+v %v", second, err)
	}
	if _, err := c.Process(pid, "run-a", identity.birth+1); err == nil {
		t.Fatal("foreign birth accepted")
	}
	changed, err := c.Process(pid, "run-b", identity.birth)
	if err != nil || !slices.Contains(changed.Unavailable, "cpu") {
		t.Fatal("new run reused old CPU baseline", err)
	}
}

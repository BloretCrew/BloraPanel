//go:build linux

package monitor

import (
	"os/exec"
	"testing"
	"time"
)

func TestProcessIdentityAndTermination(t *testing.T) {
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()
	p, err := readProcessDetails(cmd.Process.Pid)
	if err != nil {
		t.Fatal(err)
	}
	if p.StartTicks == 0 || p.Command == "" {
		t.Fatalf("incomplete process %+v", p)
	}
	if err := Terminate(p.PID, p.StartTicks); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("process did not terminate")
	}
}

//go:build windows

package monitor

import (
	"os"
	"os/exec"
	"testing"
	"time"
	"unsafe"
)

func TestWindowsMonitorChild(t *testing.T) {
	if os.Getenv("BLORA_MONITOR_CHILD") != "1" {
		return
	}
	time.Sleep(time.Minute)
}

func TestWindowsNativeMetricsAndProcessIdentity(t *testing.T) {
	if unsafe.Sizeof(memoryStatusEx{}) != 64 {
		t.Fatal("MEMORYSTATUSEX layout mismatch")
	}
	if unsafe.Sizeof(processMemoryCounters{}) != 8+8*unsafe.Sizeof(uintptr(0)) {
		t.Fatal("PROCESS_MEMORY_COUNTERS layout mismatch")
	}
	c := New()
	point, err := c.Node(t.TempDir())
	if err != nil || point.MemoryTotal <= 0 || point.DiskTotal <= 0 {
		t.Fatalf("native node metrics: %+v %v", point, err)
	}
	current, err := c.Process(os.Getpid(), "test-run")
	if err != nil || current.RSS <= 0 {
		t.Fatalf("native process metrics: %+v %v", current, err)
	}
	items, err := ListProcesses(100)
	if err != nil || len(items) == 0 {
		t.Fatalf("process snapshot: %v", err)
	}
	child := exec.Command(os.Args[0], "-test.run=^TestWindowsMonitorChild$")
	child.Env = append(os.Environ(), "BLORA_MONITOR_CHILD=1")
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer child.Process.Kill()
	details, err := readProcessDetails(child.Process.Pid)
	if err != nil {
		t.Fatal(err)
	}
	if details.StartTicks == 0 || details.StartTicks > 1<<53 {
		t.Fatal("birth token cannot round-trip through browser JSON")
	}
	if err := Terminate(child.Process.Pid, details.StartTicks+1); err == nil {
		t.Fatal("wrong process identity accepted")
	}
	if err := Terminate(child.Process.Pid, details.StartTicks); err != nil {
		t.Fatal(err)
	}
	_ = child.Wait()
}

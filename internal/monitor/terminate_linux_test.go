//go:build linux

package monitor

import (
	"os"
	"os/exec"
	"testing"

	"golang.org/x/sys/unix"
)

func TestTerminateStableIdentityAndConfirmedExit(t *testing.T) {
	child := exec.Command("sleep", "60")
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer child.Process.Kill()
	defer child.Wait()
	p, err := readProcessDetails(child.Process.Pid)
	if err != nil {
		t.Fatal(err)
	}
	if err := Terminate(p.PID, p.StartTicks+1); err == nil {
		t.Fatal("accepted wrong birth identity")
	}
	if err := child.Process.Signal(unix.Signal(0)); err != nil {
		t.Fatalf("wrong identity killed child: %v", err)
	}
	if err := Terminate(os.Getpid(), p.StartTicks); err == nil {
		t.Fatal("accepted own daemon PID")
	}
	fd, err := unix.PidfdOpen(p.PID, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer unix.Close(fd)
	if err := Terminate(p.PID, p.StartTicks); err != nil {
		t.Fatal(err)
	}
	fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	n, err := unix.Poll(fds, 0)
	if err != nil || n != 1 || fds[0].Revents&unix.POLLIN == 0 {
		t.Fatalf("returned before child exit: %v, %v", fds, err)
	}
}

func TestProcessPagesSearchAllProcesses(t *testing.T) {
	children := []*exec.Cmd{exec.Command("sleep", "59.876543"), exec.Command("sleep", "59.876543")}
	for _, child := range children {
		if err := child.Start(); err != nil {
			t.Fatal(err)
		}
		defer child.Wait()
		defer child.Process.Kill()
	}
	first, err := ListProcessPage(ProcessQuery{Limit: 1, Search: "59.876543"})
	if err != nil || len(first.Items) != 1 || first.NextAfterPID == 0 {
		t.Fatalf("first: %+v %v", first, err)
	}
	second, err := ListProcessPage(ProcessQuery{Limit: 1, AfterPID: first.NextAfterPID, Search: "59.876543"})
	if err != nil || len(second.Items) != 1 || second.NextAfterPID != 0 {
		t.Fatalf("second: %+v %v", second, err)
	}
	if first.Items[0].PID != children[0].Process.Pid || second.Items[0].PID != children[1].Process.Pid {
		t.Fatalf("unexpected process identities: %+v %+v", first, second)
	}
	for _, q := range []ProcessQuery{{Limit: 101}, {AfterPID: -1}, {Limit: -1}} {
		if _, err := ListProcessPage(q); err == nil {
			t.Fatalf("accepted invalid query %+v", q)
		}
	}
}

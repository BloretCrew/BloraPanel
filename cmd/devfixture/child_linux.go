package main

import (
	"os"
	"os/exec"
	"syscall"
)

// Keep terminal Ctrl+C on the fixture supervisor. It must stop owned instances
// while Master and Daemons still serve requests, before interrupting children.
func isolateFixtureChild(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true} }

// masterRestartSignal requests a test-only restart while the fixture supervisor
// stays alive and retains ownership of every child process.
func masterRestartSignal() os.Signal { return syscall.SIGUSR1 }

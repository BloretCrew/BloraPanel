//go:build linux

package helper

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/creack/pty"
	"golang.org/x/sys/unix"
)

func TestHelperChildProcess(t *testing.T) {
	if os.Getenv("BLORA_HELPER_TEST_CHILD") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Exit(Run(os.Args[i+1:], os.Stdin, os.Stdout, os.Stderr))
		}
	}
	os.Exit(125)
}

func TestHelperRealPTYBirthGuardAndSessionCleanup(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	const token = "0123456789abcdef0123456789abcdef"
	cmd := exec.Command(executable, "-test.run=^TestHelperChildProcess$", "--", "serve", "--", "/bin/sh", "-c", `sleep 120 & child=$!; printf 'CHILD=%s\n' "$child"; wait`)
	cmd.Env = append(os.Environ(), "BLORA_HELPER_TEST_CHILD=1", TokenEnvironment+"="+token)
	f, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()
	reader := bufio.NewReader(f)
	line, err := reader.ReadString('\n')
	if err != nil || !strings.HasPrefix(line, Prefix) {
		t.Fatal("helper handshake missing")
	}
	var identity Identity
	if err = json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(line, Prefix))), &identity); err != nil {
		t.Fatal(err)
	}
	if identity.PID != cmd.Process.Pid || identity.SessionID != identity.PID || identity.Version != Version || identity.Token != token {
		t.Fatal("helper identity not bound to the spawned process")
	}
	if _, err = f.Write([]byte("BLORA-EXEC-GO " + token + "\n")); err != nil {
		t.Fatal(err)
	}
	line, err = reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	child, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "CHILD=")))
	if err != nil {
		t.Fatal("child PID missing")
	}
	childFD, err := unix.PidfdOpen(child, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer unix.Close(childFD)
	wrong := identity
	wrong.StartTicks++
	if err = closeIdentity(wrong); err == nil {
		t.Fatal("reused birth was accepted")
	}
	wrong = identity
	wrong.Token = "fedcba9876543210fedcba9876543210"
	if err = closeIdentity(wrong); err == nil {
		t.Fatal("foreign ownership marker accepted")
	}
	control := exec.Command("/bin/sleep", "120")
	if err = control.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = control.Process.Kill(); _ = control.Wait() }()
	controlFD, err := unix.PidfdOpen(control.Process.Pid, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer unix.Close(controlFD)
	if err = closeIdentity(identity); err != nil {
		t.Fatal(err)
	}
	poll := []unix.PollFd{{Fd: int32(childFD), Events: unix.POLLIN}}
	if n, err := unix.Poll(poll, 1000); err != nil || n != 1 {
		t.Fatalf("owned child exit was not confirmed: %v", err)
	}
	poll = []unix.PollFd{{Fd: int32(controlFD), Events: unix.POLLIN}}
	if n, err := unix.Poll(poll, 0); err != nil || n != 0 {
		t.Fatal("unrelated test process was affected")
	}
	if err = cmd.Wait(); err != nil {
		t.Fatal(err)
	}
}

func TestHelperRequiresPTYAndChecksPIDFDCapability(t *testing.T) {
	if err := probe(); err != nil {
		t.Fatal(err)
	}
	if err := closeIdentity(Identity{PID: 1, StartTicks: 1, Token: strings.Repeat("a", 32)}); err == nil {
		t.Fatal("PID 1 accepted")
	}
	file, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	_, err = serve([]string{"/bin/true"}, file, file, file)
	if err == nil {
		t.Fatal("non-PTY helper launch accepted")
	}
	if errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
}

func TestHelperCommandFailureIsVisibleAndCleanupConfirmed(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	const token = "0123456789abcdef0123456789abcdef"
	cmd := exec.CommandContext(ctx, executable, "-test.run=^TestHelperChildProcess$", "--", "serve", "--", "/bin/sh", "-c", "exit 7")
	cmd.Env = append(os.Environ(), "BLORA_HELPER_TEST_CHILD=1", TokenEnvironment+"="+token)
	f, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	defer func() { _ = cmd.Process.Kill() }()
	reader := bufio.NewReader(f)
	line, err := reader.ReadString('\n')
	if err != nil || !strings.HasPrefix(line, Prefix) {
		t.Fatal("helper handshake missing")
	}
	if _, err = f.Write([]byte("BLORA-EXEC-GO " + token + "\n")); err != nil {
		t.Fatal(err)
	}
	out, err := io.ReadAll(reader)
	if err != nil && !errors.Is(err, unix.EIO) {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "command exited with status 7") {
		t.Fatal("terminal command failure was hidden")
	}
	if err = cmd.Wait(); err != nil {
		t.Fatalf("verified cleanup reported as an unknown supervisor failure: %v", err)
	}
}

//go:build linux

package helper

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/creack/pty"
	"golang.org/x/sys/unix"
)

func TestHelperClosesSetsidDescendantsWithoutTouchingAnotherProcess(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	const token = "fedcba9876543210fedcba9876543210"
	cmd := exec.Command(executable, "-test.run=^TestHelperChildProcess$", "--", "serve", "--", "/bin/sh", "-c", `setsid /bin/sh -c 'sleep 120 & child=$!; printf "ESCAPED=%s\n" "$child"; wait' & wait`)
	cmd.Env = append(os.Environ(), "BLORA_HELPER_TEST_CHILD=1", TokenEnvironment+"="+token)
	f, err := pty.Start(cmd)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	defer func() { cmd.Process.Kill(); cmd.Wait() }()
	reader := bufio.NewReader(f)
	line, err := reader.ReadString('\n')
	if err != nil || !strings.HasPrefix(line, Prefix) {
		t.Fatal("missing helper handshake", err)
	}
	var identity Identity
	if err = json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(line, Prefix))), &identity); err != nil {
		t.Fatal(err)
	}
	if _, err = f.Write([]byte("BLORA-EXEC-GO " + token + "\n")); err != nil {
		t.Fatal(err)
	}
	line, err = reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	child, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "ESCAPED=")))
	if err != nil {
		t.Fatal("missing escaped child PID", line, err)
	}
	actual, err := processIdentity(child)
	if err != nil {
		t.Fatal(err)
	}
	if actual.Session == identity.SessionID {
		t.Fatal("test child did not escape the original session")
	}
	fd, err := unix.PidfdOpen(child, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer unix.Close(fd)
	// The descriptor pins only this controlled child for failure-path cleanup.
	defer unix.PidfdSendSignal(fd, unix.SIGKILL, nil, 0)
	other := exec.Command("/bin/sleep", "120")
	if err = other.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { other.Process.Kill(); other.Wait() }()
	otherFD, err := unix.PidfdOpen(other.Process.Pid, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer unix.Close(otherFD)
	if err = closeIdentity(identity); err != nil {
		t.Fatal(err)
	}
	if n, err := unix.Poll([]unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}, 1000); err != nil || n != 1 {
		t.Fatal("setsid descendant survived helper success", err)
	}
	if n, err := unix.Poll([]unix.PollFd{{Fd: int32(otherFD), Events: unix.POLLIN}}, 0); err != nil || n != 0 {
		t.Fatal("unrelated process affected", err)
	}
	if err = cmd.Wait(); err != nil {
		t.Fatal(err)
	}
}

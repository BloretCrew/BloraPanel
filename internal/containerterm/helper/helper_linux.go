//go:build linux

package helper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

type procIdentity struct {
	PID, Parent, Session int
	Start                uint64
	State                string
}

func processIdentity(pid int) (procIdentity, error) {
	b, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return procIdentity{}, err
	}
	end := bytes.LastIndexByte(b, ')')
	if end < 0 {
		return procIdentity{}, errors.New("invalid process stat")
	}
	f := strings.Fields(string(b[end+1:]))
	if len(f) < 20 {
		return procIdentity{}, errors.New("incomplete process stat")
	}
	sid, err := strconv.Atoi(f[3])
	if err != nil {
		return procIdentity{}, err
	}
	start, err := strconv.ParseUint(f[19], 10, 64)
	if err != nil {
		return procIdentity{}, err
	}
	parent, err := strconv.Atoi(f[1])
	if err != nil {
		return procIdentity{}, err
	}
	return procIdentity{PID: pid, Parent: parent, Session: sid, Start: start, State: f[0]}, nil
}
func probe() error {
	fd, err := unix.PidfdOpen(os.Getpid(), 0)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	return unix.PidfdSendSignal(fd, 0, nil, 0)
}
func tokenValid(token string) bool {
	if len(token) != 32 {
		return false
	}
	for _, c := range token {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

// Run only understands fixed probe/serve/close operations. Serve requires its
// own PTY session and a launch acknowledgement before starting user code.
func Run(args []string, in *os.File, out, diagnostic io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(diagnostic, "exec-helper operation required")
		return 125
	}
	var err error
	switch args[0] {
	case "probe":
		if len(args) != 1 {
			err = errors.New("unexpected probe arguments")
			break
		}
		if err = probe(); err == nil {
			err = unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0)
		}
		if err == nil {
			err = json.NewEncoder(out).Encode(Probe{Version: Version, PIDFD: true, ProcessTree: true})
		}
	case "serve":
		if len(args) < 3 || args[1] != "--" {
			err = errors.New("serve requires -- command")
			break
		}
		var code int
		code, err = serve(args[2:], in, out, diagnostic)
		if err == nil {
			return code
		}
	case "close":
		if len(args) != 4 {
			err = errors.New("close requires PID, birth, and token")
			break
		}
		pid, e := strconv.Atoi(args[1])
		if e != nil {
			err = e
			break
		}
		birth, e := strconv.ParseUint(args[2], 10, 64)
		if e != nil {
			err = e
			break
		}
		err = closeIdentity(Identity{PID: pid, StartTicks: birth, SessionID: pid, Token: args[3]})
		if err == nil {
			err = json.NewEncoder(out).Encode(map[string]bool{"closed": true})
		}
	default:
		err = errors.New("unknown exec-helper operation")
	}
	if err != nil {
		fmt.Fprintln(diagnostic, "exec-helper:", err)
		return 125
	}
	return 0
}

func serve(command []string, in *os.File, out, diagnostic io.Writer) (int, error) {
	if err := probe(); err != nil {
		return 125, fmt.Errorf("pidfd capability unavailable: %w", err)
	}
	if err := unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0); err != nil {
		return 125, fmt.Errorf("child subreaper capability unavailable: %w", err)
	}
	token := os.Getenv(TokenEnvironment)
	if !tokenValid(token) {
		return 125, errors.New("missing exec ownership token")
	}
	identity, err := processIdentity(os.Getpid())
	if err != nil {
		return 125, err
	}
	if identity.Session != identity.PID {
		return 125, errors.New("exec helper must own a distinct PTY session")
	}
	termios, err := unix.IoctlGetTermios(int(in.Fd()), unix.TCGETS)
	if err != nil {
		return 125, fmt.Errorf("PTY required: %w", err)
	}
	handshakeMode := *termios
	handshakeMode.Lflag &^= unix.ECHO | unix.ICANON
	handshakeMode.Cc[unix.VMIN] = 1
	handshakeMode.Cc[unix.VTIME] = 0
	if err = unix.IoctlSetTermios(int(in.Fd()), unix.TCSETS, &handshakeMode); err != nil {
		return 125, err
	}
	b, _ := json.Marshal(Identity{Version: Version, PID: identity.PID, SessionID: identity.Session, StartTicks: identity.Start, Token: token})
	if _, err = fmt.Fprintf(out, "%s%s\n", Prefix, b); err != nil {
		return 125, err
	}
	ack := make(chan string, 1)
	go func() {
		var line strings.Builder
		one := make([]byte, 1)
		for line.Len() < 128 {
			n, err := in.Read(one)
			if n > 0 {
				line.WriteByte(one[0])
				if one[0] == '\n' {
					ack <- line.String()
					return
				}
			}
			if err != nil {
				break
			}
		}
		ack <- ""
	}()
	select {
	case line := <-ack:
		if line != "BLORA-EXEC-GO "+token+"\n" {
			return 125, errors.New("exec launch acknowledgement invalid")
		}
	case <-time.After(8 * time.Second):
		return 125, errors.New("exec launch acknowledgement expired")
	}
	if err = unix.IoctlSetTermios(int(in.Fd()), unix.TCSETS, termios); err != nil {
		return 125, err
	}
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = diagnostic
	cmd.Env = os.Environ()
	signals := make(chan os.Signal, 8)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT)
	defer signal.Stop(signals)
	if err = cmd.Start(); err != nil {
		return 125, err
	}
	wait := make(chan error, 1)
	go func() { wait <- cmd.Wait() }()
	for {
		select {
		case err = <-wait:
			if cleanupErr := cleanupSession(identity.Session, time.Now().Add(3*time.Second)); cleanupErr != nil {
				return 125, cleanupErr
			}
			if err == nil {
				return 0, nil
			}
			var exit *exec.ExitError
			if errors.As(err, &exit) {
				code := exit.ExitCode()
				if code < 0 {
					code = 128
					if status, ok := exit.Sys().(syscall.WaitStatus); ok && status.Signaled() {
						code += int(status.Signal())
					}
				}
				// The helper's status describes verified session cleanup, not the
				// success of arbitrary terminal commands. Preserve the command's
				// actual failure in terminal output while returning cleanup success.
				if _, err = fmt.Fprintf(diagnostic, "\r\n[Blora terminal command exited with status %d]\r\n", code); err != nil {
					return 125, err
				}
				return 0, nil
			}
			return 125, err
		case sig := <-signals:
			// Terminal interrupts already reach the foreground process group.
			// The supervisor stays alive to observe and clean that process.
			if sig == syscall.SIGINT || sig == syscall.SIGQUIT {
				continue
			}
			if err = cleanupSession(identity.Session, time.Now().Add(3*time.Second)); err != nil {
				return 125, err
			}
			select {
			case <-wait:
				return 0, nil
			case <-time.After(time.Second):
				return 125, errors.New("exec child exit not confirmed")
			}
		}
	}
}

func closeIdentity(expected Identity) error {
	if expected.PID <= 1 || expected.StartTicks == 0 || !tokenValid(expected.Token) {
		return errors.New("invalid exec identity")
	}
	fd, err := unix.PidfdOpen(expected.PID, 0)
	if err != nil {
		return fmt.Errorf("open exec identity: %w", err)
	}
	defer unix.Close(fd)
	actual, err := processIdentity(expected.PID)
	if err != nil {
		return err
	}
	if actual.Start != expected.StartTicks || actual.Session != expected.PID {
		return errors.New("exec birth or session identity changed")
	}
	env, err := os.ReadFile(fmt.Sprintf("/proc/%d/environ", expected.PID))
	if err != nil {
		return err
	}
	found := false
	for _, value := range bytes.Split(env, []byte{0}) {
		if string(value) == TokenEnvironment+"="+expected.Token {
			found = true
			break
		}
	}
	if !found {
		return errors.New("exec ownership marker differs")
	}
	// Re-check birth after opening the pidfd. Signals use that pinned kernel
	// object, so a PID reused after this point cannot redirect the operation.
	if err = unix.PidfdSendSignal(fd, unix.SIGTERM, nil, 0); err != nil {
		return err
	}
	poll := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		n, err := unix.Poll(poll, 100)
		if errors.Is(err, unix.EINTR) {
			continue
		}
		if err != nil {
			return err
		}
		if n > 0 && poll[0].Revents&unix.POLLIN != 0 {
			return nil
		}
	}
	return errors.New("exec helper did not confirm cleanup before deadline")
}

func cleanupSession(session int, deadline time.Time) error {
	if session != os.Getpid() {
		return errors.New("refuse to clean another process session")
	}
	quiet := 0
	for time.Now().Before(deadline) {
		entries, err := os.ReadDir("/proc")
		if err != nil {
			return err
		}
		if len(entries) > 100000 {
			return errors.New("exec process inspection budget reached")
		}
		processes := map[int]procIdentity{}
		for _, entry := range entries {
			if !time.Now().Before(deadline) {
				return errors.New("exec process inspection deadline elapsed")
			}
			pid, err := strconv.Atoi(entry.Name())
			if err != nil || pid <= 1 {
				continue
			}
			p, err := processIdentity(pid)
			if err == nil {
				processes[pid] = p
			}
		}
		active := 0
		ownership := map[int]bool{session: true, 1: false, 0: false}
		for pid, p := range processes {
			if !time.Now().Before(deadline) {
				return errors.New("exec process cleanup deadline elapsed")
			}
			if pid == os.Getpid() || p.State == "Z" || p.State == "X" {
				continue
			}
			// Subreaper ancestry survives setsid and double-fork. A distinct
			// terminal/exec is never our child and cannot enter this set.
			owned := p.Session == session
			ancestor := p.Parent
			chain := []int{pid}
			for depth := 0; !owned && ancestor > 1; depth++ {
				if depth >= 1024 {
					return errors.New("exec process ancestry exceeds inspection budget")
				}
				if known, ok := ownership[ancestor]; ok {
					owned = known
					break
				}
				parent, ok := processes[ancestor]
				if !ok || parent.Parent == ancestor {
					break
				}
				chain = append(chain, ancestor)
				ancestor = parent.Parent
			}
			for _, member := range chain {
				ownership[member] = owned
			}
			if !owned {
				continue
			}
			fd, err := unix.PidfdOpen(pid, 0)
			if errors.Is(err, unix.ESRCH) {
				continue
			}
			if err != nil {
				return err
			}
			current, err := processIdentity(pid)
			if err != nil || current.Start != p.Start {
				unix.Close(fd)
				continue
			}
			active++
			// Stop before killing to restrict further forks; repeated scans also
			// cover children created immediately before that stop was delivered.
			err = unix.PidfdSendSignal(fd, unix.SIGSTOP, nil, 0)
			if err == nil {
				err = unix.PidfdSendSignal(fd, unix.SIGKILL, nil, 0)
			}
			unix.Close(fd)
			if err != nil && !errors.Is(err, unix.ESRCH) {
				return err
			}
		}
		if active == 0 {
			quiet++
			if quiet >= 2 {
				return nil
			}
		} else {
			quiet = 0
		}
		time.Sleep(10 * time.Millisecond)
	}
	return errors.New("processes remain in exec session after cleanup deadline")
}

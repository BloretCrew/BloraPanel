//go:build linux

package terminal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
	"golang.org/x/sys/unix"
)

type nativeBackend struct{}
type linuxProcess struct {
	file      *os.File
	cmd       *exec.Cmd
	done      chan struct{}
	mu        sync.Mutex
	waitErr   error
	closed    bool
	closeOnce sync.Once
	closeErr  error
}

// Native terminals run as the daemon's identity and are host.manage only.
// A session/process group is lifecycle containment, not a tenant sandbox.
func (nativeBackend) Spawn(ctx context.Context, s SpawnSpec) (Process, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cmd := exec.Command(s.Command[0], s.Command[1:]...)
	cmd.Dir = s.Directory
	env := map[string]string{"PATH": "/usr/local/bin:/usr/bin:/bin", "TERM": "xterm-256color", "LANG": "C.UTF-8", "BLORA_TERMINAL_ID": s.SessionID}
	for k, v := range s.Environment {
		env[k] = v
	}
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		cmd.Env = append(cmd.Env, k+"="+env[k])
	}
	f, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: s.Cols, Rows: s.Rows})
	if err != nil {
		return nil, fmt.Errorf("Linux PTY: %w", err)
	}
	// The nonblocking duplicate is recognized by os.NewFile's poller; deadlines
	// bound input even when a foreground program stops reading its terminal.
	fd, err := unix.FcntlInt(f.Fd(), unix.F_DUPFD_CLOEXEC, 0)
	if err == nil {
		err = unix.SetNonblock(fd, true)
	}
	if err != nil {
		if fd >= 0 {
			_ = unix.Close(fd)
		}
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		_ = f.Close()
		_ = cmd.Wait()
		return nil, err
	}
	pollable := os.NewFile(uintptr(fd), "blora-pty")
	_ = f.Close()
	p := &linuxProcess{file: pollable, cmd: cmd, done: make(chan struct{})}
	go func() { err := cmd.Wait(); p.mu.Lock(); p.waitErr = err; p.mu.Unlock(); close(p.done) }()
	return p, nil
}
func (p *linuxProcess) Read(b []byte) (int, error) {
	n, err := p.file.Read(b)
	if errors.Is(err, syscall.EIO) {
		err = io.EOF
	}
	return n, err
}
func (p *linuxProcess) Write(b []byte) (int, error) {
	if err := p.file.SetWriteDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return 0, err
	}
	return p.file.Write(b)
}
func (p *linuxProcess) Resize(cols, rows uint16) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return os.ErrClosed
	}
	return pty.Setsize(p.file, &pty.Winsize{Cols: cols, Rows: rows})
}
func (p *linuxProcess) Wait() error { <-p.done; p.mu.Lock(); defer p.mu.Unlock(); return p.waitErr }
func (p *linuxProcess) Close() error {
	p.closeOnce.Do(func() {
		p.mu.Lock()
		p.closed = true
		p.mu.Unlock()
		// An interactive shell gives a foreground job its own process group.
		// TIOCGPGRP identifies that job through this PTY; only signal it after
		// confirming that it belongs to the session created for this shell.
		foreground, fgErr := unix.IoctlGetInt(int(p.file.Fd()), unix.TIOCGPGRP)
		if fgErr == nil && foreground > 0 && foreground != p.cmd.Process.Pid {
			if sid, err := unix.Getsid(foreground); err == nil && sid == p.cmd.Process.Pid {
				if err = unix.Kill(-foreground, unix.SIGKILL); err != nil && !errors.Is(err, unix.ESRCH) {
					p.closeErr = err
				}
			}
		}
		// Only the group created by StartWithSize is addressed. Holding the PTY
		// until signalling also retains its controlling session association.
		err := syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL)
		if err != nil && !errors.Is(err, syscall.ESRCH) {
			p.closeErr = err
		}
		select {
		case <-p.done:
		case <-time.After(3 * time.Second):
			p.closeErr = errors.Join(p.closeErr, errors.New("PTY shell did not exit before deadline"))
		}
		if err = p.file.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			p.closeErr = errors.Join(p.closeErr, err)
		}
	})
	return p.closeErr
}

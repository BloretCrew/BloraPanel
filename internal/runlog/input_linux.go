//go:build linux

package runlog

import (
	"context"
	"errors"
	"golang.org/x/sys/unix"
	"os"
	"strconv"
	"time"
)

type heldInput struct {
	read, writer *os.File
	writerFD     int
}

func newHeldInput() (*heldInput, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	fd := int(w.Fd()) // Capture once: later File.Fd calls can restore blocking mode.
	if err = unix.SetNonblock(fd, true); err != nil {
		r.Close()
		w.Close()
		return nil, err
	}
	return &heldInput{read: r, writer: w, writerFD: fd}, nil
}
func pipeDescriptor(f *os.File) (inputDescriptor, error) {
	var stat unix.Stat_t
	err := unix.Fstat(int(f.Fd()), &stat)
	if err == nil && stat.Mode&unix.S_IFMT != unix.S_IFIFO {
		err = ErrUnknown
	}
	return inputDescriptor{Handle: uint64(f.Fd()), Device: uint64(stat.Dev), Inode: stat.Ino}, err
}
func (h *heldInput) descriptor() (inputDescriptor, error) { return pipeDescriptor(h.read) }
func (h *heldInput) close()                               { h.writer.Close(); h.read.Close() }
func (h *heldInput) write(ctx context.Context, data []byte) (int, error) {
	// Exactly one successful write syscall. EAGAIN means zero bytes were
	// accepted and may be waited on; a partial write is returned, never retried.
	for {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		n, err := unix.Write(h.writerFD, data)
		if !errors.Is(err, unix.EAGAIN) && !errors.Is(err, unix.EINTR) {
			return max(n, 0), err
		}
		delay := 50
		if deadline, ok := ctx.Deadline(); ok {
			delay = max(1, min(delay, int(time.Until(deadline).Milliseconds())))
		}
		fds := []unix.PollFd{{Fd: int32(h.writerFD), Events: unix.POLLOUT}}
		if _, err = unix.Poll(fds, delay); err != nil && !errors.Is(err, unix.EINTR) {
			return 0, err
		}
	}
}
func duplicateInput(r record) (*os.File, error) {
	alive, err := identityAlive(r.Identity)
	if err != nil || !alive {
		return nil, ErrUnknown
	}
	f, err := os.Open("/proc/" + strconv.Itoa(r.Identity.PID) + "/fd/" + strconv.FormatUint(r.Input.Handle, 10))
	if err != nil {
		return nil, err
	}
	d, err := pipeDescriptor(f)
	if err == nil && (d.Device != r.Input.Device || d.Inode != r.Input.Inode || d.Inode == 0) {
		err = ErrUnknown
	}
	if err == nil {
		alive, err = identityAlive(r.Identity)
		if !alive {
			err = ErrUnknown
		}
	}
	if err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

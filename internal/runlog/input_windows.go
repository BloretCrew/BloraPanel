//go:build windows

package runlog

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"golang.org/x/sys/windows"
	"os"
	"runtime"
	"sync"
	"time"
)

type heldInput struct {
	read     *os.File
	writer   windows.Handle
	once     sync.Once
	poisoned bool
}

func newHeldInput() (*heldInput, error) {
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return nil, err
	}
	name, err := windows.UTF16PtrFromString(`\\.\pipe\Blora-Input-` + hex.EncodeToString(random[:]))
	if err != nil {
		return nil, err
	}
	writer, err := windows.CreateNamedPipe(name, windows.PIPE_ACCESS_OUTBOUND|windows.FILE_FLAG_OVERLAPPED|windows.FILE_FLAG_FIRST_PIPE_INSTANCE, windows.PIPE_TYPE_BYTE|windows.PIPE_READMODE_BYTE|windows.PIPE_WAIT|windows.PIPE_REJECT_REMOTE_CLIENTS, 1, 64<<10, 0, 2000, nil)
	if err != nil {
		return nil, err
	}
	reader, err := windows.CreateFile(name, windows.GENERIC_READ, 0, nil, windows.OPEN_EXISTING, 0, 0)
	if err != nil {
		windows.CloseHandle(writer)
		return nil, err
	}
	connected, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		windows.CloseHandle(reader)
		windows.CloseHandle(writer)
		return nil, err
	}
	ov := &windows.Overlapped{HEvent: connected}
	err = windows.ConnectNamedPipe(writer, ov)
	if errors.Is(err, windows.ERROR_IO_PENDING) {
		// CreateFile above already connected our only client. Retain any
		// unexpected pending operation until cancellation actually completes.
		_ = windows.CancelIoEx(writer, ov)
		go func() {
			defer windows.CloseHandle(connected)
			windows.WaitForSingleObject(connected, windows.INFINITE)
			runtime.KeepAlive(ov)
		}()
	} else {
		windows.CloseHandle(connected)
	}
	if err != nil && !errors.Is(err, windows.ERROR_PIPE_CONNECTED) {
		windows.CloseHandle(reader)
		windows.CloseHandle(writer)
		return nil, err
	}
	return &heldInput{read: os.NewFile(uintptr(reader), "business-stdin"), writer: writer}, nil
}
func (h *heldInput) descriptor() (inputDescriptor, error) {
	return inputDescriptor{Handle: uint64(h.read.Fd())}, nil
}
func (h *heldInput) close() {
	h.once.Do(func() { windows.CancelIoEx(h.writer, nil); windows.CloseHandle(h.writer); h.read.Close() })
}
func (h *heldInput) write(ctx context.Context, data []byte) (int, error) {
	if h.poisoned {
		return 0, ErrIncomplete
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	event, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return 0, err
	}
	ov := &windows.Overlapped{HEvent: event}
	var written uint32
	err = windows.WriteFile(h.writer, data, &written, ov)
	if !errors.Is(err, windows.ERROR_IO_PENDING) {
		windows.CloseHandle(event)
		return int(written), err
	}
	for {
		if err = ctx.Err(); err != nil {
			cancelErr := windows.CancelIoEx(h.writer, ov)
			state, waitErr := windows.WaitForSingleObject(event, 100)
			if waitErr != nil || state != windows.WAIT_OBJECT_0 {
				// Retain the exact overlapped storage until the kernel retires it.
				// No further write may reuse this pipe while completion is unknown.
				h.poisoned = true
				go func() {
					defer windows.CloseHandle(event)
					windows.WaitForSingleObject(event, windows.INFINITE)
					runtime.KeepAlive(ov)
					runtime.KeepAlive(data)
				}()
				return 0, ErrIncomplete
			}
			resultErr := windows.GetOverlappedResult(h.writer, ov, &written, false)
			windows.CloseHandle(event)
			if resultErr == nil {
				return int(written), nil
			}
			if cancelErr != nil && !errors.Is(cancelErr, windows.ERROR_NOT_FOUND) {
				return int(written), cancelErr
			}
			return int(written), err
		}
		delay := uint32(50)
		if deadline, ok := ctx.Deadline(); ok {
			delay = uint32(max(1, min(50, time.Until(deadline).Milliseconds())))
		}
		state, e := windows.WaitForSingleObject(event, delay)
		if e != nil {
			windows.CancelIoEx(h.writer, ov)
			h.poisoned = true
			go func() {
				defer windows.CloseHandle(event)
				windows.WaitForSingleObject(event, windows.INFINITE)
				runtime.KeepAlive(ov)
				runtime.KeepAlive(data)
			}()
			return 0, e
		}
		if state == windows.WAIT_OBJECT_0 {
			err = windows.GetOverlappedResult(h.writer, ov, &written, false)
			windows.CloseHandle(event)
			return int(written), err
		}
	}
}
func duplicateInput(r record) (*os.File, error) {
	h, alive, err := openIdentity(r.Identity, windows.PROCESS_DUP_HANDLE)
	if err != nil || !alive {
		return nil, ErrUnknown
	}
	defer windows.CloseHandle(h)
	var copy windows.Handle
	// Request only read access, even though the source is already read-only.
	if err = windows.DuplicateHandle(h, windows.Handle(r.Input.Handle), windows.CurrentProcess(), &copy, windows.GENERIC_READ, false, 0); err != nil {
		return nil, err
	}
	typeID, err := windows.GetFileType(copy)
	if err != nil || typeID != windows.FILE_TYPE_PIPE {
		windows.CloseHandle(copy)
		return nil, ErrUnknown
	}
	return os.NewFile(uintptr(copy), "business-stdin"), nil
}

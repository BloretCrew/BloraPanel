//go:build windows

package runlog

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var isInJob = windows.NewLazySystemDLL("kernel32.dll").NewProc("IsProcessInJob")

func processInJob(handle windows.Handle) (bool, error) {
	var result int32
	ok, _, err := isInJob.Call(uintptr(handle), 0, uintptr(unsafe.Pointer(&result)))
	if ok == 0 {
		return false, err
	}
	return result != 0, nil
}
func independentProcess(cmd *exec.Cmd) error {
	flags := uint32(windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP)
	inside, err := processInJob(windows.CurrentProcess())
	if err != nil {
		return err
	}
	if inside {
		flags |= windows.CREATE_BREAKAWAY_FROM_JOB
	}
	// An enclosing Job that forbids breakaway causes CreateProcess to fail;
	// inheriting daemon kill-on-close lifetime is never silently accepted.
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: flags}
	return nil
}
func validateIndependent() error {
	inside, err := processInJob(windows.CurrentProcess())
	if err != nil {
		return err
	}
	if inside {
		return errors.New("log helper is still in an enclosing Windows Job")
	}
	return nil
}
func processBirth(handle windows.Handle) (uint64, error) {
	var created, exited, kernel, user windows.Filetime
	err := windows.GetProcessTimes(handle, &created, &exited, &kernel, &user)
	return uint64(created.HighDateTime)<<32 | uint64(created.LowDateTime), err
}
func selfIdentity() (Identity, error) {
	birth, err := processBirth(windows.CurrentProcess())
	return Identity{PID: os.Getpid(), Birth: birth}, err
}
func openIdentity(i Identity, access uint32) (windows.Handle, bool, error) {
	if i.PID <= 0 || i.Birth == 0 {
		return 0, false, ErrUnknown
	}
	h, err := windows.OpenProcess(access|windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.SYNCHRONIZE, false, uint32(i.PID))
	if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	birth, err := processBirth(h)
	if err != nil {
		windows.CloseHandle(h)
		return 0, false, err
	}
	if birth != i.Birth {
		windows.CloseHandle(h)
		return 0, false, nil
	}
	state, err := windows.WaitForSingleObject(h, 0)
	if err != nil {
		windows.CloseHandle(h)
		return 0, false, err
	}
	if state == windows.WAIT_OBJECT_0 {
		windows.CloseHandle(h)
		return 0, false, nil
	}
	return h, true, nil
}
func identityAlive(i Identity) (bool, error) {
	h, alive, err := openIdentity(i, 0)
	if h != 0 {
		windows.CloseHandle(h)
	}
	return alive, err
}
func terminateIdentity(i Identity) error {
	h, alive, err := openIdentity(i, windows.PROCESS_TERMINATE)
	if err != nil || !alive {
		return err
	}
	defer windows.CloseHandle(h)
	if err = windows.TerminateProcess(h, 1); err != nil {
		return err
	}
	state, err := windows.WaitForSingleObject(h, 2000)
	if err != nil {
		return err
	}
	if state != windows.WAIT_OBJECT_0 {
		return fmt.Errorf("log helper termination unconfirmed: %w", ErrUnknown)
	}
	return nil
}
func syncDirectory(string) error { return nil } // File.Sync uses FlushFileBuffers.

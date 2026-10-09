//go:build windows

package bootstrap

import (
	"fmt"
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
)

func lockService(root string) (func(), error) {
	f, err := os.OpenFile(filepath.Join(root, "launcher.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	o := new(windows.Overlapped)
	if err = windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, o); err != nil {
		f.Close()
		return nil, fmt.Errorf("another service launcher owns this state: %w", err)
	}
	return func() { _ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, o); _ = f.Close() }, nil
}
func replaceServiceFile(from, to string) error {
	f, e := windows.UTF16PtrFromString(from)
	if e != nil {
		return e
	}
	t, e := windows.UTF16PtrFromString(to)
	if e != nil {
		return e
	}
	return windows.MoveFileEx(f, t, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH)
}

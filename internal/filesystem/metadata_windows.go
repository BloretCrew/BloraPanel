package filesystem

import (
	"golang.org/x/sys/windows"
	"os"
	"time"
)

func setFileTimes(f *os.File, modified time.Time) error {
	v := windows.NsecToFiletime(modified.UnixNano())
	h, err := reopenMetadata(f)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	return windows.SetFileTime(h, nil, &v, &v)
}

func reopenMetadata(f *os.File) (windows.Handle, error) {
	// Reopen the exact object handle with attribute access. Resolving its path
	// again could follow a replacement directory or link during a rename race.
	reopen := windows.NewLazySystemDLL("kernel32.dll").NewProc("ReOpenFile")
	h, _, err := reopen.Call(f.Fd(), windows.FILE_WRITE_ATTRIBUTES, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, windows.FILE_FLAG_BACKUP_SEMANTICS)
	if windows.Handle(h) == windows.InvalidHandle {
		return windows.InvalidHandle, err
	}
	return windows.Handle(h), nil
}

func applyHandleMetadata(f *os.File, mode uint32, modified time.Time) error {
	h, err := reopenMetadata(f)
	if err != nil {
		return err
	}
	attributes := os.NewFile(uintptr(h), f.Name())
	defer attributes.Close()
	if err = attributes.Chmod(os.FileMode(mode)); err != nil {
		return err
	}
	v := windows.NsecToFiletime(modified.UnixNano())
	// Windows directory metadata has the same process-crash guarantee as
	// syncDir; FlushFileBuffers cannot flush a directory attribute handle.
	return windows.SetFileTime(h, nil, &v, &v)
}

package filesystem

import (
	"fmt"
	"os"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
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
	info, err := f.Stat()
	if err != nil {
		return windows.InvalidHandle, err
	}
	if info.IsDir() {
		// Reopen the retained directory object with explicit
		// directory options; the previous file reopen path failed for directories.
		// Never resolve f.Name(): a renamed/replaced ancestor must not redirect
		// an already validated metadata operation to a different object.
		// NT's empty relative name means the RootDirectory object itself. Go's
		// Windows os.Root adapter uses the same translation for the "." path.
		name, err := windows.NewNTUnicodeString("")
		if err != nil {
			return windows.InvalidHandle, err
		}
		attrs := windows.OBJECT_ATTRIBUTES{RootDirectory: windows.Handle(f.Fd()), ObjectName: name}
		attrs.Length = uint32(unsafe.Sizeof(attrs))
		var h windows.Handle
		err = windows.NtCreateFile(&h, windows.FILE_READ_ATTRIBUTES|windows.FILE_WRITE_ATTRIBUTES|windows.SYNCHRONIZE, &attrs, &windows.IO_STATUS_BLOCK{}, nil, 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, windows.FILE_OPEN, windows.FILE_DIRECTORY_FILE|windows.FILE_OPEN_REPARSE_POINT|windows.FILE_SYNCHRONOUS_IO_NONALERT|windows.FILE_OPEN_FOR_BACKUP_INTENT, 0, 0)
		if err != nil {
			return windows.InvalidHandle, fmt.Errorf("open directory attribute handle: %w", err)
		}
		return h, nil
	}
	reopen := windows.NewLazySystemDLL("kernel32.dll").NewProc("ReOpenFile")
	// os.File.Chmod first reads existing attributes before changing READONLY.
	h, _, err := reopen.Call(f.Fd(), windows.FILE_READ_ATTRIBUTES|windows.FILE_WRITE_ATTRIBUTES, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, 0)
	if windows.Handle(h) == windows.InvalidHandle {
		return windows.InvalidHandle, fmt.Errorf("reopen file attribute handle: %w", err)
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
		return fmt.Errorf("set file attributes: %w", err)
	}
	v := windows.NsecToFiletime(modified.UnixNano())
	// Windows directory metadata has the same process-crash guarantee as
	// syncDir; FlushFileBuffers cannot flush a directory attribute handle.
	if err = windows.SetFileTime(h, nil, &v, &v); err != nil {
		return fmt.Errorf("set file times: %w", err)
	}
	return nil
}

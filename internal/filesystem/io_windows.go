package filesystem

import (
	"errors"
	"os"
	"syscall"
	"time"
)

const readNonblock = 0

// Windows FlushFileBuffers requires a writable file handle, and directory
// handles cannot be flushed with the access rights exposed by os.Root.Open.
// File Sync still precedes the rename. Power-loss directory durability is not
// claimed on Windows; process-crash checkpoints remain recoverable.
func syncDir(r *os.Root, p string) error { return nil }
func crossDevice(e error) bool           { return errors.Is(e, syscall.Errno(17)) }
func setModified(f *os.File, t time.Time) error {
	ft := syscall.NsecToFiletime(t.UnixNano())
	return syscall.SetFileTime(syscall.Handle(f.Fd()), nil, nil, &ft)
}

// Root's Windows metadata methods acquire a rooted FILE_WRITE_ATTRIBUTES
// handle. A generic-read handle cannot call SetFileTime. The documented Unix
// chmod/chtimes path race does not apply to this Windows implementation.
func applyPathMetadata(r *os.Root, p string, item Entry) error {
	if e := r.Chmod(p, os.FileMode(item.Mode).Perm()); e != nil {
		return e
	}
	return r.Chtimes(p, item.Modified, item.Modified)
}

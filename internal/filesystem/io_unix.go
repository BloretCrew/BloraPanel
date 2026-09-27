//go:build !windows

package filesystem

import (
	"errors"
	"os"
	"syscall"
	"time"
)

const readNonblock = syscall.O_NONBLOCK

func syncDir(r *os.Root, p string) error {
	f, e := r.Open(p)
	if e != nil {
		return e
	}
	defer f.Close()
	return f.Sync()
}
func crossDevice(e error) bool { return errors.Is(e, syscall.EXDEV) }
func setModified(f *os.File, t time.Time) error {
	tv := syscall.NsecToTimeval(t.UnixNano())
	return syscall.Futimes(int(f.Fd()), []syscall.Timeval{tv, tv})
}
func applyPathMetadata(r *os.Root, p string, item Entry) error {
	f, e := r.OpenFile(p, os.O_RDONLY|readNonblock, 0)
	if e != nil {
		return e
	}
	defer f.Close()
	fi, e := f.Stat()
	if e != nil {
		return e
	}
	if !fi.IsDir() && !fi.Mode().IsRegular() {
		return ErrUnsupported
	}
	if e = f.Chmod(os.FileMode(item.Mode).Perm()); e != nil {
		return e
	}
	if e = setModified(f, item.Modified); e != nil {
		return e
	}
	if fi.IsDir() {
		return syncDir(r, p)
	}
	return f.Sync()
}

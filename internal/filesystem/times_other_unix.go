//go:build !windows && !linux

package filesystem

import (
	"os"
	"time"

	"golang.org/x/sys/unix"
)

func setFileTimes(f *os.File, modified time.Time) error {
	v := unix.NsecToTimeval(modified.UnixNano())
	return unix.Futimes(int(f.Fd()), []unix.Timeval{v, v})
}

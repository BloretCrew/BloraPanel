package filesystem

import (
	"errors"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

func setFileTimes(f *os.File, modified time.Time) error {
	v := unix.NsecToTimespec(modified.UnixNano())
	err := unix.UtimesNanoAt(int(f.Fd()), "", []unix.Timespec{v, v}, unix.AT_EMPTY_PATH)
	if !errors.Is(err, unix.EINVAL) && !errors.Is(err, unix.ENOSYS) && !errors.Is(err, unix.ENOENT) {
		return err
	}
	// Older kernels lack AT_EMPTY_PATH. The fd-based fallback retains the
	// safety boundary with the microsecond precision of the older syscall.
	tv := unix.NsecToTimeval(modified.UnixNano())
	return unix.Futimes(int(f.Fd()), []unix.Timeval{tv, tv})
}

//go:build !windows

package filesystem

import (
	"os"
	"time"
)

func applyHandleMetadata(f *os.File, mode uint32, modified time.Time) error {
	if err := f.Chmod(os.FileMode(mode)); err != nil {
		return err
	}
	if err := setFileTimes(f, modified); err != nil {
		return err
	}
	return f.Sync()
}

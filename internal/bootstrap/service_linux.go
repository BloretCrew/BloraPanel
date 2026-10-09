//go:build !windows

package bootstrap

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
)

func lockService(root string) (func(), error) {
	f, err := os.OpenFile(filepath.Join(root, "launcher.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		f.Close()
		return nil, fmt.Errorf("another service launcher owns this state: %w", err)
	}
	return func() { _ = unix.Flock(int(f.Fd()), unix.LOCK_UN); _ = f.Close() }, nil
}
func replaceServiceFile(from, to string) error {
	if err := os.Rename(from, to); err != nil {
		return err
	}
	directory, err := os.Open(filepath.Dir(to))
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

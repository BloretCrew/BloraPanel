//go:build !windows

package daemon

import "os"

func syncContainerLogDirectory(root string) error {
	dir, err := os.Open(root)
	if err != nil {
		return err
	}
	err = dir.Sync()
	closeErr := dir.Close()
	if err != nil {
		return err
	}
	return closeErr
}

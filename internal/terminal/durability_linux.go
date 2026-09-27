//go:build linux

package terminal

import "os"

func syncDirectory(root string) error {
	f, err := os.Open(root)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}

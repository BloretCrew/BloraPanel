package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func seedPerformanceResource(root string, source bool, transferBytes int64) error {
	if err := os.Mkdir(root, 0700); err != nil {
		return err
	}
	if !source {
		return nil
	}
	directory := filepath.Join(root, "ten-thousand")
	if err := os.Mkdir(directory, 0700); err != nil {
		return err
	}
	for index := 0; index < 10000; index++ {
		path := filepath.Join(directory, fmt.Sprintf("entry-%05d.txt", index))
		if err := os.WriteFile(path, []byte(fmt.Sprintf("Blora performance fixture entry %d\n", index)), 0600); err != nil {
			return err
		}
	}
	f, err := os.OpenFile(filepath.Join(root, "transfer-load.bin"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	// Sparse source avoids unnecessary fixture disk allocation. Transfer still
	// reads, hashes, transports, writes and verifies every source byte.
	if err = f.Truncate(transferBytes); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

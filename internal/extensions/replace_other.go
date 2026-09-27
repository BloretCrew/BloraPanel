//go:build !windows

package extensions

import "os"

func replaceRegistryFile(from, to string) error { return os.Rename(from, to) }

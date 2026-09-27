//go:build !linux

package terminal

// Files are flushed before replacement. Windows has no directory fsync API.
func syncDirectory(string) error { return nil }

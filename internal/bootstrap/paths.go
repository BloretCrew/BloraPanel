package bootstrap

import (
	"os"
	"path/filepath"
)

// ExecutableDir is independent of the caller's working directory.
func ExecutableDir() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	return filepath.Dir(path), nil
}

func ResolvePath(base, path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(base, path)
}

//go:build !linux && !windows

package containers

import (
	"context"
	"io"
)

func syncDir(string) error                     { return nil }
func recoverCLI(context.Context, string) error { return nil }
func runCLI(context.Context, []string, string, string, []string, io.Writer, io.Writer) error {
	return ErrCapability
}

//go:build !linux

package helper

import (
	"fmt"
	"io"
	"os"
)

func Run(_ []string, _ *os.File, _ io.Writer, diagnostic io.Writer) int {
	fmt.Fprintln(diagnostic, "exec-helper requires Linux and pidfd support")
	return 125
}

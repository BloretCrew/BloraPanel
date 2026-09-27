package master

import (
	"fmt"
	"os"
	"testing"

	"blora.dev/panel/internal/runlog"
	run "blora.dev/panel/internal/runtime"
)

func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == "--job-keeper" {
		if err := run.RunJobKeeper(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if len(os.Args) > 1 && os.Args[1] == "--log-helper" {
		if err := runlog.RunHelper(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

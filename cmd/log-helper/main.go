package main

import (
	"blora.dev/panel/internal/runlog"
	"fmt"
	"os"
)

func main() {
	if err := runlog.RunHelper(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Blora log helper:", err)
		os.Exit(1)
	}
}

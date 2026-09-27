package main

import (
	"blora.dev/panel/internal/containerterm/helper"
	"os"
)

func main() { os.Exit(helper.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }

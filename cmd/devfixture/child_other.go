//go:build !linux

package main

import (
	"os"
	"os/exec"
)

func isolateFixtureChild(*exec.Cmd)  {}
func masterRestartSignal() os.Signal { return nil }

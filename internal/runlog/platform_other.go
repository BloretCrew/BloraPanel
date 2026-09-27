//go:build !linux && !windows

package runlog

import "os/exec"

func independentProcess(*exec.Cmd) error   { return ErrUnknown }
func validateIndependent() error           { return ErrUnknown }
func selfIdentity() (Identity, error)      { return Identity{}, ErrUnknown }
func identityAlive(Identity) (bool, error) { return false, ErrUnknown }
func terminateIdentity(Identity) error     { return ErrUnknown }
func syncDirectory(string) error           { return nil }

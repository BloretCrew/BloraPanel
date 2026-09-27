//go:build !linux && !windows

package monitor

import "errors"

func visitProcesses(visit func(SystemProcess)) error {
	return errors.New("system process listing unavailable on this platform")
}
func Terminate(pid int, start uint64) error {
	return errors.New("system process termination unavailable on this platform")
}

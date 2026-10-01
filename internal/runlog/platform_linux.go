//go:build linux

package runlog

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

func independentProcess(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return nil
}
func validateIndependent() error {
	sid, err := unix.Getsid(0)
	if err != nil {
		return err
	}
	if sid != os.Getpid() {
		return errors.New("log helper must own an independent OS session")
	}
	// Require a real exact-process signal handle before accepting any output.
	fd, err := unix.PidfdOpen(os.Getpid(), 0)
	if err != nil {
		return fmt.Errorf("log helper requires pidfd: %w", err)
	}
	defer unix.Close(fd)
	return unix.PidfdSendSignal(fd, 0, nil, 0)
}
func procBirth(pid int) (uint64, byte, error) {
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return 0, 0, err
	}
	end := strings.LastIndexByte(string(b), ')')
	if end < 0 {
		return 0, 0, ErrUnknown
	}
	fields := strings.Fields(string(b[end+1:]))
	if len(fields) < 20 || len(fields[0]) != 1 {
		return 0, 0, ErrUnknown
	}
	birth, err := strconv.ParseUint(fields[19], 10, 64)
	return birth, fields[0][0], err
}
func currentBoot() (string, error) {
	b, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	return strings.TrimSpace(string(b)), err
}
func selfIdentity() (Identity, error) {
	i := Identity{PID: os.Getpid()}
	var err error
	i.Birth, _, err = procBirth(i.PID)
	if err != nil {
		return i, err
	}
	i.BootID, err = currentBoot()
	return i, err
}
func identityAlive(i Identity) (bool, error) {
	if i.PID <= 1 || i.Birth == 0 || i.BootID == "" {
		return false, ErrUnknown
	}
	boot, err := currentBoot()
	if err != nil {
		return false, err
	}
	if boot != i.BootID {
		return false, nil
	}
	birth, state, err := procBirth(i.PID)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, unix.ESRCH) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if birth != i.Birth {
		return false, nil
	}
	return state != 'Z' && state != 'X', nil
}
func terminateIdentity(i Identity) error {
	if i.PID <= 1 {
		return ErrUnknown
	}
	fd, err := unix.PidfdOpen(i.PID, 0)
	if errors.Is(err, unix.ESRCH) {
		return nil
	}
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	alive, err := identityAlive(i)
	if err != nil {
		return err
	}
	if !alive {
		return nil
	}
	// pidfd was acquired BEFORE rereading the birth identity, closing PID reuse.
	if err = unix.PidfdSendSignal(fd, unix.SIGKILL, nil, 0); err != nil && !errors.Is(err, unix.ESRCH) {
		return err
	}
	poll := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	_, err = unix.Poll(poll, 2000)
	if err != nil {
		return err
	}
	if poll[0].Revents&unix.POLLIN == 0 {
		return ErrUnknown
	}
	return nil
}
func syncDirectory(root string) error {
	f, err := os.Open(root)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}

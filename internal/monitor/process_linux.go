//go:build linux

package monitor

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

func visitProcesses(visit func(SystemProcess)) error {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return err
	}
	for _, e := range entries {
		pid, er := strconv.Atoi(e.Name())
		if er != nil || pid <= 0 {
			continue
		}
		p, er := readProcessDetails(pid)
		if er != nil {
			p = SystemProcess{PID: pid, Command: "[unavailable]", Unavailable: []string{"identity", "memory", "user"}}
		}
		visit(p)
	}
	return nil
}
func readProcessDetails(pid int) (SystemProcess, error) {
	p := SystemProcess{PID: pid}
	b, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
	if err != nil {
		return p, err
	}
	s := string(b)
	close := strings.LastIndex(s, ")")
	if close < 0 {
		return p, errors.New("invalid process stat")
	}
	f := strings.Fields(s[close+2:])
	if len(f) < 20 {
		return p, errors.New("short process stat")
	}
	p.StartTicks, _ = strconv.ParseUint(f[19], 10, 64)
	if cmd, er := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "cmdline")); er == nil {
		p.Command = strings.TrimSpace(strings.ReplaceAll(string(cmd), "\x00", " "))
	}
	if p.Command == "" {
		p.Command = strings.TrimSpace(s[strings.Index(s, "(")+1 : close])
	}
	file, er := os.Open(filepath.Join("/proc", strconv.Itoa(pid), "status"))
	if er == nil {
		defer file.Close()
		sc := bufio.NewScanner(file)
		for sc.Scan() {
			v := strings.Fields(sc.Text())
			if len(v) >= 2 && v[0] == "Uid:" {
				p.UID = v[1]
			}
			if len(v) >= 2 && v[0] == "VmRSS:" {
				k, _ := strconv.ParseInt(v[1], 10, 64)
				p.RSS = k << 10
			}
		}
	}
	return p, nil
}
func Terminate(pid int, start uint64) error {
	if pid <= 1 || pid == os.Getpid() || start == 0 {
		return errors.New("protected process or missing identity")
	}
	// Acquire the stable kernel identity before checking /proc. Never fall
	// back to kill(pid), which could signal a reused PID after validation.
	fd, err := unix.PidfdOpen(pid, 0)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	p, err := readProcessDetails(pid)
	if err != nil {
		return err
	}
	if start == 0 || p.StartTicks != start {
		return errors.New("process identity changed")
	}
	if err := unix.PidfdSendSignal(fd, unix.SIGKILL, nil, 0); err != nil {
		return err
	}
	deadline := time.Now().Add(10 * time.Second)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return errors.New("process exit confirmation timed out")
		}
		fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
		_, err := unix.Poll(fds, int((remaining+time.Millisecond-1)/time.Millisecond))
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			return err
		}
		if fds[0].Revents&(unix.POLLIN|unix.POLLHUP) != 0 {
			return nil
		}
		if fds[0].Revents&(unix.POLLERR|unix.POLLNVAL) != 0 {
			return errors.New("process exit observation failed")
		}
	}
}

//go:build linux

package containers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type cliRunRecord struct {
	Marker  string `json:"marker"`
	Session int    `json:"session"`
}

func syncDir(root string) error {
	f, err := os.Open(root)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
func runCLI(ctx context.Context, command []string, dir, stateRoot string, env []string, stdout, stderr io.Writer) (returnErr error) {
	if len(command) == 0 {
		return ErrCapability
	}
	fd, err := unix.PidfdOpen(os.Getpid(), 0)
	if err != nil {
		return fmt.Errorf("CLI cancellation requires pidfd: %w", ErrCapability)
	}
	unix.Close(fd)
	marker := randomID()
	records := filepath.Join(stateRoot, "cli-runs")
	if err := os.MkdirAll(records, 0700); err != nil {
		return err
	}
	if err := syncDir(stateRoot); err != nil {
		return err
	}
	record := filepath.Join(records, marker)
	// The token is durable before fork, covering death before a PID can be saved.
	if err := atomicJSON(record, cliRunRecord{Marker: marker}); err != nil {
		return err
	}
	session := 0
	defer func() {
		if session > 1 {
			if err := terminateCLISession(session, marker); err != nil {
				returnErr = errors.Join(returnErr, ErrUnknown, err)
				return
			}
		}
		if err := terminateCLIToken(context.Background(), marker); err != nil {
			returnErr = errors.Join(returnErr, ErrUnknown, err)
			return
		}
		if err := os.Remove(record); err != nil {
			returnErr = errors.Join(returnErr, err)
			return
		}
		returnErr = errors.Join(returnErr, syncDir(records))
	}()
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = dir
	cmd.Env = append(env, "BLORA_CONTAINER_CLI_TOKEN="+marker)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.WaitDelay = time.Second
	cmd.Cancel = func() error { return terminateCLISession(cmd.Process.Pid, marker) }
	if err = cmd.Start(); err != nil {
		return err
	}
	pid := cmd.Process.Pid
	session = pid
	if saveErr := atomicJSON(record, cliRunRecord{Marker: marker, Session: pid}); saveErr != nil {
		cleanupErr := terminateCLISession(pid, marker)
		if cleanupErr != nil {
			cleanupErr = errors.Join(cleanupErr, cmd.Process.Kill())
		}
		waitErr := cmd.Wait()
		return errors.Join(saveErr, cleanupErr, waitErr)
	}
	err = cmd.Wait()
	if ctx.Err() != nil {
		cleanupErr := terminateCLISession(pid, marker)
		return errors.Join(ctx.Err(), err, cleanupErr)
	}
	if errors.Is(err, exec.ErrWaitDelay) {
		return errors.Join(err, terminateCLISession(pid, marker))
	}
	return err
}

func recoverCLI(ctx context.Context, root string) error {
	dir := filepath.Join(root, "cli-runs")
	f, err := os.Open(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	entries, err := f.ReadDir(1025)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	if len(entries) > 1024 {
		return errors.New("CLI recovery record budget reached")
	}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		marker := entry.Name()
		// atomicFile can leave a temporary file if interrupted before rename.
		// No process is started until the final record has been committed.
		if strings.HasPrefix(marker, ".pending-") {
			continue
		}
		if len(marker) != 32 || strings.Trim(marker, "0123456789abcdef") != "" || !entry.Type().IsRegular() {
			return ErrIdentity
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Size() > 1024 {
			return ErrIdentity
		}
		b, err := os.ReadFile(filepath.Join(dir, marker))
		if err != nil {
			return err
		}
		var record cliRunRecord
		if json.Unmarshal(b, &record) != nil || record.Marker != marker || record.Session < 0 || record.Session == 1 {
			return ErrIdentity
		}
		if record.Session > 1 {
			if err := terminateCLISession(record.Session, marker); err != nil {
				return err
			}
		}
		if err := terminateCLIToken(ctx, marker); err != nil {
			return err
		}
		if err := os.Remove(filepath.Join(dir, marker)); err != nil {
			return err
		}
		if err := syncDir(dir); err != nil {
			return err
		}
	}
	return nil
}

func terminateCLIToken(ctx context.Context, marker string) error {
	deadline := time.Now().Add(2 * time.Second)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		entries, err := os.ReadDir("/proc")
		if err != nil {
			return err
		}
		if len(entries) > 100000 {
			return errors.New("CLI process inspection budget reached")
		}
		live := 0
		for _, entry := range entries {
			pid, err := strconv.Atoi(entry.Name())
			if err != nil || pid <= 1 {
				continue
			}
			_, birth, state, err := cliBirth(pid)
			if err != nil || state == 'Z' || state == 'X' {
				continue
			}
			path := "/proc/" + entry.Name()
			b, err := os.ReadFile(path + "/environ")
			if errors.Is(err, os.ErrNotExist) || errors.Is(err, unix.ESRCH) {
				continue
			}
			if err != nil {
				info, statErr := os.Stat(path)
				if errors.Is(statErr, os.ErrNotExist) {
					continue
				}
				if statErr == nil && info.Sys().(*syscall.Stat_t).Uid != uint32(os.Geteuid()) {
					continue
				}
				return fmt.Errorf("CLI ownership inspection failed: %w", err)
			}
			if !bytes.Contains(append([]byte{0}, b...), []byte("\x00BLORA_CONTAINER_CLI_TOKEN="+marker+"\x00")) {
				continue
			}
			fd, err := unix.PidfdOpen(pid, 0)
			if errors.Is(err, unix.ESRCH) {
				continue
			}
			if err != nil {
				return err
			}
			_, again, _, err := cliBirth(pid)
			if err != nil || again != birth {
				unix.Close(fd)
				continue
			}
			live++
			err = unix.PidfdSendSignal(fd, unix.SIGKILL, nil, 0)
			unix.Close(fd)
			if err != nil && !errors.Is(err, unix.ESRCH) {
				return err
			}
		}
		if live == 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return ErrUnknown
		}
		time.Sleep(20 * time.Millisecond)
	}
}
func cliBirth(pid int) (sid int, birth uint64, state byte, err error) {
	b, e := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if e != nil {
		err = e
		return
	}
	end := strings.LastIndexByte(string(b), ')')
	if end < 0 {
		err = ErrIdentity
		return
	}
	fields := strings.Fields(string(b[end+1:]))
	if len(fields) < 20 || len(fields[0]) != 1 {
		err = ErrIdentity
		return
	}
	sid, err = strconv.Atoi(fields[3])
	if err != nil {
		return
	}
	birth, err = strconv.ParseUint(fields[19], 10, 64)
	state = fields[0][0]
	return
}
func terminateCLISession(sid int, marker string) error {
	deadline := time.Now().Add(2 * time.Second)
	for {
		entries, err := os.ReadDir("/proc")
		if err != nil {
			return err
		}
		if len(entries) > 100000 {
			return errors.New("CLI process inspection budget reached")
		}
		live := 0
		for _, entry := range entries {
			pid, e := strconv.Atoi(entry.Name())
			if e != nil || pid <= 1 {
				continue
			}
			session, birth, state, e := cliBirth(pid)
			if e != nil || session != sid || state == 'Z' || state == 'X' {
				continue
			}
			live++
			fd, e := unix.PidfdOpen(pid, 0)
			if errors.Is(e, unix.ESRCH) {
				continue
			}
			if e != nil {
				return e
			}
			again, second, _, e := cliBirth(pid)
			if e != nil || again != sid || second != birth {
				unix.Close(fd)
				continue
			}
			environ, e := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/environ")
			owned := false
			for _, value := range bytes.Split(environ, []byte{0}) {
				if string(value) == "BLORA_CONTAINER_CLI_TOKEN="+marker {
					owned = true
					break
				}
			}
			if e != nil || !owned {
				unix.Close(fd)
				return fmt.Errorf("CLI descendant ownership unconfirmed: %w", ErrUnknown)
			}
			e = unix.PidfdSendSignal(fd, unix.SIGKILL, nil, 0)
			unix.Close(fd)
			if e != nil && !errors.Is(e, unix.ESRCH) {
				return e
			}
		}
		if live == 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return ErrUnknown
		}
		time.Sleep(20 * time.Millisecond)
	}
}

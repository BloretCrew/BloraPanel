//go:build linux

package runtime

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"blora.dev/panel/internal/model"
	"golang.org/x/sys/unix"
)

func nativeBackend() string { return "linux" }
func prepareContainerDirectory(path string, uid, gid uint32) error {
	created, err := newContainerDirectory(path)
	if err != nil {
		return err
	}
	if created {
		return os.Chown(path, int(uid), int(gid))
	}
	return nil
}
func syncDirectory(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
func bootID() (string, error) {
	b, e := os.ReadFile("/proc/sys/kernel/random/boot_id")
	return strings.TrimSpace(string(b)), e
}

func (m *Manager) startNative(ctx context.Context, r Record, c model.InstanceConfig, streams IO) (Record, error) {
	var err error
	r.BootID, err = bootID()
	if err != nil {
		return r, err
	}
	attr := &syscall.SysProcAttr{Setpgid: true}
	if (c.UID == nil) != (c.GID == nil) {
		return r, errors.New("native UID and GID must be set together")
	}
	if c.UID != nil {
		attr.Credential = &syscall.Credential{Uid: *c.UID, Gid: *c.GID, Groups: []uint32{}}
	}
	var group *os.File
	if m.options.CgroupRoot != "" {
		root, e := filepath.EvalSymlinks(m.options.CgroupRoot)
		if e != nil {
			return r, e
		}
		var stat unix.Statfs_t
		if e = unix.Statfs(root, &stat); e != nil {
			return r, e
		}
		if stat.Type != unix.CGROUP2_SUPER_MAGIC {
			return r, fmt.Errorf("delegated path is not cgroup v2: %w", ErrCapability)
		}
		r.Unit = filepath.Join(root, "blora-"+r.RunID)
		if e = os.Mkdir(r.Unit, 0700); e != nil {
			return r, fmt.Errorf("create delegated cgroup: %w", e)
		}
		group, e = os.Open(r.Unit)
		if e != nil {
			return r, e
		}
		defer group.Close()
		info, e := group.Stat()
		if e != nil {
			return r, e
		}
		r.UnitIdentity = info.Sys().(*syscall.Stat_t).Ino
		limits := map[string]string{}
		if c.MemoryBytes > 0 {
			limits["memory.max"] = strconv.FormatInt(c.MemoryBytes, 10)
		}
		if c.PidsLimit > 0 {
			limits["pids.max"] = strconv.FormatInt(c.PidsLimit, 10)
		}
		if c.CPUQuota > 0 {
			limits["cpu.max"] = strconv.FormatInt(c.CPUQuota, 10) + " 100000"
		}
		for name, value := range limits {
			if e = os.WriteFile(filepath.Join(r.Unit, name), []byte(value), 0600); e != nil {
				return r, fmt.Errorf("set cgroup %s: %w", name, e)
			}
		}
		attr.UseCgroupFD = true
		attr.CgroupFD = int(group.Fd())
	} else {
		if !m.options.AllowPGIDFallback {
			return r, fmt.Errorf("no delegated cgroup and trusted PGID fallback is disabled: %w", ErrCapability)
		}
		if c.MemoryBytes > 0 || c.CPUQuota > 0 || c.PidsLimit > 0 {
			return r, fmt.Errorf("PGID fallback cannot enforce resource limits: %w", ErrCapability)
		}
		r.WeakContainment = true
	}
	if err = m.save(r); err != nil {
		return r, err
	}
	if err = ctx.Err(); err != nil {
		return r, err
	}
	cmd := exec.Command(c.Command[0], c.Command[1:]...)
	cmd.Dir = c.Directory
	cmd.Env = environment(c, r.Token)
	cmd.SysProcAttr = attr
	// Files are passed straight through. cmd.Wait is never coupled to io.Copy or
	// EOF on a pipe inherited by an unrelated descendant.
	if streams.Stdin != nil {
		cmd.Stdin = streams.Stdin
	}
	if streams.Stdout != nil {
		cmd.Stdout = streams.Stdout
	}
	if streams.Stderr != nil {
		cmd.Stderr = streams.Stderr
	}
	if err = cmd.Start(); err != nil {
		return r, fmt.Errorf("launch in running unit: %w", err)
	}
	r.PID = cmd.Process.Pid
	r.PGID = r.PID
	if p, e := readProcess(r.PID); e == nil {
		r.StartTicks = p.StartTicks
	} else {
		// A very short command may already have exited. Keep the PID unreaped
		// until the first identity read; /proc usually retains zombie identity.
		err = e
	}
	m.mu.Lock()
	m.live[r.RunID] = &liveRun{stdin: streams.Input, platform: cmd}
	m.mu.Unlock()
	go func() { _ = cmd.Wait() }()
	if err != nil {
		return r, fmt.Errorf("launch identity read failed; reconcile ownership marker: %w", err)
	}
	return r, nil
}

type procInfo struct {
	Process
	PGID int
}

func processGone(err error) bool {
	if err == nil {
		return false
	}
	return os.IsNotExist(err) || errors.Is(err, syscall.ESRCH) || strings.Contains(strings.ToLower(err.Error()), "no such process")
}

func readProcess(pid int) (procInfo, error) {
	b, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return procInfo{}, err
	}
	end := bytes.LastIndexByte(b, ')')
	if end < 0 {
		return procInfo{}, ErrUnknown
	}
	f := strings.Fields(string(b[end+1:]))
	if len(f) < 20 {
		return procInfo{}, ErrUnknown
	}
	pgid, e := strconv.Atoi(f[2])
	if e != nil {
		return procInfo{}, e
	}
	ticks, e := strconv.ParseUint(f[19], 10, 64)
	if e != nil {
		return procInfo{}, e
	}
	return procInfo{Process: Process{PID: pid, StartTicks: ticks, State: f[0]}, PGID: pgid}, nil
}
func hasMarker(pid int, token string) (bool, error) {
	b, e := os.ReadFile(fmt.Sprintf("/proc/%d/environ", pid))
	if e != nil {
		return false, e
	}
	for _, v := range bytes.Split(b, []byte{0}) {
		if string(v) == "BLORA_RUN_TOKEN="+token {
			return true, nil
		}
	}
	return false, nil
}

func (m *Manager) verifyCgroup(r Record) (bool, error) {
	if m.options.CgroupRoot == "" {
		return false, ErrUnknown
	}
	root, err := filepath.EvalSymlinks(m.options.CgroupRoot)
	if err != nil {
		return false, err
	}
	if r.Unit != filepath.Join(root, "blora-"+r.RunID) {
		return false, ErrUnknown
	}
	info, err := os.Lstat(r.Unit)
	if os.IsNotExist(err) {
		return false, ErrUnknown
	}
	if err != nil {
		return false, err
	}
	if !info.IsDir() || info.Sys().(*syscall.Stat_t).Ino != r.UnitIdentity {
		return false, ErrUnknown
	}
	return true, nil
}
func (m *Manager) observeNative(ctx context.Context, r Record) (Observation, error) {
	o := Observation{State: "UNKNOWN", ObservedAt: time.Now().UTC()}
	if r.Backend != "linux" {
		return o, ErrCapability
	}
	boot, err := bootID()
	if err != nil {
		return o, err
	}
	if r.BootID == "" {
		return o, ErrUnknown
	}
	if r.BootID != boot {
		o.State = "STOPPED"
		o.Exited = true
		o.Diagnostic = "operating system boot identity changed"
		return o, nil
	}
	if r.Unit != "" {
		if _, err = m.verifyCgroup(r); err != nil {
			return o, err
		}
		b, e := os.ReadFile(filepath.Join(r.Unit, "cgroup.events"))
		if e != nil {
			return o, e
		}
		populated := ""
		for _, line := range strings.Split(string(b), "\n") {
			v := strings.Fields(line)
			if len(v) == 2 && v[0] == "populated" {
				populated = v[1]
			}
		}
		if populated == "0" {
			o.State = "STOPPED"
			o.Exited = true
			return o, nil
		}
		if populated != "1" {
			return o, ErrUnknown
		}
		pids, e := cgroupPIDs(r.Unit)
		if e != nil {
			return o, e
		}
		for _, pid := range pids {
			p, e := readProcess(pid)
			if e == nil {
				o.Processes = append(o.Processes, p.Process)
			} else if !processGone(e) {
				return o, e
			}
		}
		o.State = "RUNNING"
		return o, nil
	}
	if !r.WeakContainment {
		return o, ErrUnknown
	}
	entries, e := os.ReadDir("/proc")
	if e != nil {
		return o, e
	}
	for _, entry := range entries {
		if err = ctx.Err(); err != nil {
			return o, err
		}
		pid, e := strconv.Atoi(entry.Name())
		if e != nil {
			continue
		}
		p, e := readProcess(pid)
		if e != nil {
			if processGone(e) {
				continue
			}
			return o, e
		}
		// Zombies no longer execute, fork or hold file descriptors. Their reaping
		// belongs to the parent; they do not block a new resource run.
		if p.State == "Z" || p.State == "X" {
			continue
		}
		if r.PGID != 0 && p.PGID != r.PGID {
			continue
		}
		if pid == r.PID && r.StartTicks != 0 && p.StartTicks != r.StartTicks {
			return o, fmt.Errorf("original PID birth identity changed: %w", ErrUnknown)
		}
		owned := pid == r.PID && r.StartTicks != 0 && p.StartTicks == r.StartTicks
		if !owned {
			owned, e = hasMarker(pid, r.Token)
			if e != nil {
				if processGone(e) {
					continue
				}
				if r.PGID != 0 {
					return o, fmt.Errorf("cannot inspect group member %d: %w", pid, e)
				}
				continue
			}
		}
		if !owned {
			if r.PGID != 0 {
				return o, fmt.Errorf("process group contains unconfirmed process %d: %w", pid, ErrUnknown)
			}
			continue
		}
		o.Processes = append(o.Processes, p.Process)
	}
	if len(o.Processes) == 0 {
		o.State = "STOPPED"
		o.Exited = true
	} else {
		o.State = "RUNNING"
	}
	o.Diagnostic = "trusted PGID fallback; descendants that deliberately leave the group are outside containment"
	return o, nil
}
func cgroupPIDs(root string) ([]int, error) {
	var pids []int
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		b, e := os.ReadFile(filepath.Join(path, "cgroup.procs"))
		if e != nil {
			return e
		}
		for _, v := range strings.Fields(string(b)) {
			n, e := strconv.Atoi(v)
			if e != nil {
				return e
			}
			pids = append(pids, n)
		}
		return nil
	})
	return pids, err
}
func signalIdentity(p Process, sig unix.Signal) error {
	// Hold a pidfd before checking birth identity to avoid signaling a recycled PID.
	fd, err := unix.PidfdOpen(p.PID, 0)
	if errors.Is(err, unix.ESRCH) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("pidfd is required for safe signaling: %w", err)
	}
	defer unix.Close(fd)
	now, err := readProcess(p.PID)
	if processGone(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if now.StartTicks != p.StartTicks {
		return ErrUnknown
	}
	err = unix.PidfdSendSignal(fd, sig, nil, 0)
	if errors.Is(err, unix.ESRCH) {
		return nil
	}
	return err
}
func (m *Manager) signalNative(ctx context.Context, r Record, force bool) error {
	o, err := m.observeNative(ctx, r)
	if err != nil || o.Exited {
		return err
	}
	if force && r.Unit != "" {
		if _, err = m.verifyCgroup(r); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(r.Unit, "cgroup.kill"), []byte("1"), 0600)
	}
	sig := unix.SIGTERM
	if force {
		sig = unix.SIGKILL
	}
	for _, p := range o.Processes {
		if err = signalIdentity(p, sig); err != nil {
			return err
		}
	}
	return nil
}
func (m *Manager) cleanupNative(r Record) error {
	if r.Unit != "" {
		if _, err := os.Lstat(r.Unit); errors.Is(err, os.ErrNotExist) && r.Phase == "exited" {
			return nil
		}
		if _, err := m.verifyCgroup(r); err != nil {
			return err
		}
		if err := os.Remove(r.Unit); err != nil {
			return err
		}
	}
	m.mu.Lock()
	delete(m.live, r.RunID)
	m.mu.Unlock()
	return nil
}

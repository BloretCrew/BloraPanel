//go:build linux

package terminal

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func awaitOutput(t *testing.T, m *Manager, id, view, needle string, after uint64) (Batch, []byte) {
	t.Helper()
	deadline := time.Now().Add(8 * time.Second)
	var out []byte
	var last Batch
	for time.Now().Before(deadline) {
		b, err := m.Read(context.Background(), id, "owner", view, after, 0)
		if err != nil {
			t.Fatal(err)
		}
		last = b
		after = b.Next
		for _, e := range b.Events {
			if e.Kind == "output" {
				out = append(out, e.Data...)
			}
		}
		if bytes.Contains(out, []byte(needle)) {
			return last, out
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("PTY output did not contain %q: %q", needle, out)
	return last, out
}

func TestLinuxPTYChineseANSIResizeAndReattach(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m, err := New(Options{Root: t.TempDir(), AllowNative: true})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	r := testRequest()
	r.Backend = "native"
	r.Command = []string{"/bin/sh", "-c", `stty -echo; printf '\033[?1049h\033[32m中文-ready\033[0m\n'; while IFS= read -r line; do if [ "$line" = size ]; then stty size; else printf 'received:%s\n' "$line"; fi; done`}
	s, err := m.Create(ctx, r)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	if _, err = m.Attach(context.Background(), s.ID, "owner", "view", 0, 0); err != nil {
		t.Fatal(err)
	}
	b, out := awaitOutput(t, m, s.ID, "view", "中文-ready", 0)
	if !bytes.Contains(out, []byte("\x1b[?1049h\x1b[32m")) {
		t.Fatalf("ANSI missing: %q", out)
	}
	if _, err = m.AcquireLease(s.ID, "owner", "view", time.Minute, false); err != nil {
		t.Fatal(err)
	}
	if err = m.Resize(context.Background(), s.ID, "owner", "view", 132, 42); err != nil {
		t.Fatal(err)
	}
	if err = m.WriteInput(context.Background(), s.ID, "owner", "view", []byte("size\n")); err != nil {
		t.Fatal(err)
	}
	b, _ = awaitOutput(t, m, s.ID, "view", "42 132", b.Next)
	if err = m.Detach(s.ID, "owner", "view"); err != nil {
		t.Fatal(err)
	}
	if _, err = m.Attach(context.Background(), s.ID, "owner", "moved-view", b.Next, 0); err != nil {
		t.Fatal(err)
	}
	if _, err = m.AcquireLease(s.ID, "owner", "moved-view", time.Minute, false); err != nil {
		t.Fatal(err)
	}
	if err = m.WriteInput(context.Background(), s.ID, "owner", "moved-view", []byte("continuing-中文\n")); err != nil {
		t.Fatal(err)
	}
	_, out = awaitOutput(t, m, s.ID, "moved-view", "received:continuing-中文", b.Next)
	if strings.Count(string(out), "received:continuing-中文") != 1 {
		t.Fatal("command duplicated")
	}
	closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer closeCancel()
	if err = m.CloseSession(closeCtx, s.ID, "owner"); err != nil {
		t.Fatal(err)
	}
	got, err := m.Get(context.Background(), s.ID, "owner")
	if err != nil || got.State != "exited" {
		t.Fatalf("close not confirmed: %+v %v", got, err)
	}
}

func TestLinuxPTYNoSubscriberOutputRemainsBounded(t *testing.T) {
	m, err := New(Options{Root: t.TempDir(), AllowNative: true, Archive: ArchiveOptions{MaxBytes: 64 << 10, SegmentBytes: 16 << 10}})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	r := testRequest()
	r.Backend = "native"
	r.Command = []string{"/bin/sh", "-c", `head -c 2097152 /dev/zero; printf 'output-complete'`}
	s, err := m.Create(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	live, _ := m.find(s.ID)
	select {
	case <-live.done:
	case <-time.After(10 * time.Second):
		t.Fatal("unobserved PTY stdout blocked")
	}
	b, err := m.Attach(context.Background(), s.ID, "owner", "slow-browser", 0, MaxBatchBytes)
	if err != nil {
		t.Fatal(err)
	}
	var out []byte
	for _, e := range b.Events {
		out = append(out, e.Data...)
	}
	if !b.Gap || !bytes.Contains(out, []byte("output-complete")) {
		t.Fatalf("late subscriber gap/final output absent: gap=%v bytes=%d", b.Gap, len(out))
	}
	if live.archive.Bytes() > 64<<10 {
		t.Fatal("PTY archive exceeded disk budget")
	}
	t.Logf("2 MiB raw output without browser; retained=%d bytes earliest=%d latest=%d gap=%v", live.archive.Bytes(), b.Earliest, b.Latest, b.Gap)
}

func TestLinuxPTYRepeatedArchiveRotationWithoutConsumer(t *testing.T) {
	m, err := New(Options{Root: t.TempDir(), AllowNative: true, Archive: ArchiveOptions{MaxBytes: 64 << 10, SegmentBytes: 16 << 10}})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	r := testRequest()
	r.Backend = "native"
	r.Command = []string{"/bin/sh", "-c", `i=0; while [ "$i" -lt 32 ]; do head -c 2097152 /dev/zero; i=$((i+1)); sleep .5; done; printf 'rotation-complete'`}
	s, err := m.Create(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	live, err := m.find(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	timeout := time.NewTimer(60 * time.Second)
	defer timeout.Stop()
	var peak int64
	process := live.process.(*linuxProcess)
	baseline, err := linuxProcessRSS(process.cmd.Process.Pid)
	if err != nil {
		t.Fatal(err)
	}
	var processPeak int64 = baseline
	samples := 0
loop:
	for {
		select {
		case <-live.done:
			break loop
		case <-timeout.C:
			t.Fatal("64 MiB unobserved PTY did not finish")
		case <-ticker.C:
			size := live.archive.Bytes()
			if size > peak {
				peak = size
			}
			if size > 64<<10 {
				t.Fatal("archive budget exceeded", size)
			}
			if rss, err := linuxProcessRSS(process.cmd.Process.Pid); err == nil {
				if rss > processPeak {
					processPeak = rss
				}
				if rss > baseline+32<<20 {
					t.Fatalf("PTY process RSS grew beyond bound: baseline=%d peak=%d", baseline, rss)
				}
			}
			samples++
		}
	}
	batch, err := m.Attach(context.Background(), s.ID, "owner", "late-reader", 0, MaxBatchBytes)
	if err != nil {
		t.Fatal(err)
	}
	var output []byte
	for _, event := range batch.Events {
		output = append(output, event.Data...)
	}
	if !batch.Gap || batch.Earliest <= 1 || !bytes.Contains(output, []byte("rotation-complete")) {
		t.Fatal("rolling archive lost gap or final marker")
	}
	entries, err := os.ReadDir(live.archive.root)
	if err != nil {
		t.Fatal(err)
	}
	var disk int64
	for _, entry := range entries {
		info, err := os.Stat(filepath.Join(live.archive.root, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		disk += info.Size()
	}
	if disk > 64<<10 || disk != live.archive.Bytes() || samples == 0 {
		t.Fatalf("disk=%d tracked=%d samples=%d", disk, live.archive.Bytes(), samples)
	}
	t.Logf("64 MiB real PTY, elapsed=%s samples=%d archivePeak=%d disk=%d processRSS=%d->%d earliest=%d latest=%d gap=%v", time.Since(started), samples, peak, disk, baseline, processPeak, batch.Earliest, batch.Latest, batch.Gap)
}

func linuxProcessRSS(pid int) (int64, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "VmRSS:") {
			continue
		}
		var value int64
		if _, err := fmt.Sscanf(strings.TrimSpace(strings.TrimPrefix(line, "VmRSS:")), "%d kB", &value); err != nil {
			return 0, err
		}
		return value * 1024, nil
	}
	return 0, fmt.Errorf("VmRSS missing for pid %d", pid)
}

func TestLinuxPTYCloseTerminatesSeparateForegroundJob(t *testing.T) {
	ctx := context.Background()
	m, err := New(Options{Root: t.TempDir(), AllowNative: true})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	r := testRequest()
	r.Backend = "native"
	r.Command = []string{"/bin/sh", "-i"}
	s, err := m.Create(ctx, r)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.Attach(ctx, s.ID, "owner", "view", 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err = m.AcquireLease(s.ID, "owner", "view", time.Minute, false); err != nil {
		t.Fatal(err)
	}
	if err = m.WriteInput(ctx, s.ID, "owner", "view", []byte("sleep 120\n")); err != nil {
		t.Fatal(err)
	}
	live, _ := m.find(s.ID)
	p := live.process.(*linuxProcess)
	deadline := time.Now().Add(3 * time.Second)
	foreground := 0
	for time.Now().Before(deadline) {
		fg, err := unix.IoctlGetInt(int(p.file.Fd()), unix.TIOCGPGRP)
		if err == nil && fg > 0 && fg != p.cmd.Process.Pid {
			foreground = fg
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if foreground == 0 {
		t.Fatal("interactive shell did not assign a foreground job group")
	}
	pidfd, err := unix.PidfdOpen(foreground, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer unix.Close(pidfd)
	closeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err = m.CloseSession(closeCtx, s.ID, "owner"); err != nil {
		t.Fatal(err)
	}
	poll := []unix.PollFd{{Fd: int32(pidfd), Events: unix.POLLIN}}
	if n, err := unix.Poll(poll, 1000); err != nil || n != 1 || poll[0].Revents&unix.POLLIN == 0 {
		t.Fatalf("foreground process did not exit: poll=%+v error=%v", poll, err)
	}
}

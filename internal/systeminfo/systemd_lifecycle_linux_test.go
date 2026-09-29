//go:build linux

package systeminfo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// This opt-in test requires a dedicated disposable container, never the host
// manager. The runner provisions the marker before booting its own PID 1.
func TestRealSystemdServiceAndTimerLifecycle(t *testing.T) {
	if os.Getenv("BLORA_SYSTEMD_E2E") != "1" {
		t.Skip("requires the dedicated systemd E2E container")
	}
	marker, err := os.ReadFile("/run/blora-systemd-e2e")
	if err != nil || string(marker) != "disposable-blora-systemd-e2e\n" {
		t.Fatal("refusing to use a manager without the disposable-container marker")
	}
	if _, err := os.Stat("/.dockerenv"); err != nil {
		t.Fatal("requires a disposable Docker container")
	}
	pid1, err := os.ReadFile("/proc/1/comm")
	if err != nil || strings.TrimSpace(string(pid1)) != "systemd" {
		t.Fatal("the isolated container must actually boot systemd as PID 1")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	ctl := func(args ...string) string {
		t.Helper()
		output, err := exec.CommandContext(ctx, "systemctl", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("systemctl %v: %v: %s", args, err, output)
		}
		return strings.TrimSpace(string(output))
	}
	t.Setenv("LC_ALL", "C")
	t.Setenv("SYSTEMD_COLORS", "0")
	var nonce [8]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		t.Fatal(err)
	}
	base := "blora-e2e-" + hex.EncodeToString(nonce[:])
	service, firedService, timer := base+".service", base+"-fired.service", base+".timer"
	fired := filepath.Join("/run", base+"-fired")
	units := map[string]string{
		service:      "[Unit]\nDescription=Blora isolated lifecycle probe\n[Service]\nType=simple\nExecStart=/usr/bin/sleep 600\n",
		firedService: "[Unit]\nDescription=Blora isolated timer probe\n[Service]\nType=oneshot\nExecStart=/usr/bin/touch " + fired + "\n",
		timer:        "[Unit]\nDescription=Blora isolated timer\n[Timer]\nOnActiveSec=1s\nAccuracySec=100ms\nUnit=" + firedService + "\n[Install]\nWantedBy=timers.target\n",
	}
	created := []string{}
	t.Cleanup(func() {
		cleanup, done := context.WithTimeout(context.Background(), 15*time.Second)
		defer done()
		for _, args := range [][]string{{"stop", "--", timer, service, firedService}, {"disable", "--", timer}} {
			if output, err := exec.CommandContext(cleanup, "systemctl", args...).CombinedOutput(); err != nil {
				t.Errorf("cleanup systemctl %v: %v: %s", args, err, output)
			}
		}
		for _, path := range append(created, fired) {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				t.Errorf("cleanup %s: %v", path, err)
			}
		}
		if output, err := exec.CommandContext(cleanup, "systemctl", "daemon-reload").CombinedOutput(); err != nil {
			t.Errorf("cleanup reload: %v: %s", err, output)
		}
	})
	for name, text := range units {
		path := filepath.Join("/run/systemd/system", name)
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		created = append(created, path)
		_, writeErr := file.WriteString(text)
		closeErr := file.Close()
		if writeErr != nil || closeErr != nil {
			t.Fatalf("write unit: %v / %v", writeErr, closeErr)
		}
	}
	alias := base + "-alias.service"
	aliasPath := filepath.Join("/run/systemd/system", alias)
	if err := os.Symlink(service, aliasPath); err != nil {
		t.Fatal(err)
	}
	created = append(created, aliasPath)
	ctl("daemon-reload")
	for _, action := range []string{"start", "restart", "stop"} {
		priorPID := ctl("show", service, "--property=MainPID", "--value")
		if err := ServiceAction(ctx, service, action); err != nil {
			t.Fatalf("production ServiceAction %s: %v", action, err)
		}
		pid := ctl("show", service, "--property=MainPID", "--value")
		state := ctl("show", service, "--property=ActiveState", "--value")
		want := "active"
		if action == "stop" {
			want = "inactive"
			if pid != "0" {
				t.Fatalf("stop retained process %s", pid)
			}
		} else if pid == "0" || pid == priorPID {
			t.Fatalf("%s did not create a fresh running process: %s -> %s", action, priorPID, pid)
		}
		if state != want {
			t.Fatalf("%s state %s, want %s", action, state, want)
		}
		page, err := ListServicePage(ctx, 10, base)
		if err != nil {
			t.Fatal(err)
		}
		found, aliasFound := false, false
		for _, item := range page.Items {
			if item.Name == service {
				found = item.State == want && item.Description == "Blora isolated lifecycle probe"
			}
			if item.Name == alias {
				aliasFound = item.State == want && item.Description == "Blora isolated lifecycle probe"
			}
		}
		if !found || !aliasFound {
			t.Fatalf("service enumeration did not reflect %s: %+v", action, page)
		}
	}
	for _, action := range []string{"enable", "disable", "enable"} {
		before, err := ListScheduledTaskPage(ctx, 10, base)
		if err != nil {
			t.Fatal(err)
		}
		visible := false
		for _, item := range before.Items {
			visible = visible || item.Name == timer
		}
		if !visible {
			t.Fatalf("installed timer disappeared before %s: %+v", action, before)
		}
		if err := TaskAction(ctx, timer, action); err != nil {
			t.Fatalf("production TaskAction %s: %v", action, err)
		}
		want := "enabled"
		if action == "disable" {
			want = "disabled"
		}
		if got := ctl("show", timer, "--property=UnitFileState", "--value"); got != want {
			t.Fatalf("timer %s state %q, want %q", action, got, want)
		}
		after, err := ListScheduledTaskPage(ctx, 10, base)
		if err != nil {
			t.Fatal(err)
		}
		observed := ""
		for _, item := range after.Items {
			if item.Name == timer {
				observed = item.State
			}
		}
		if observed != want {
			t.Fatalf("production timer list after %s reports %q, want %q", action, observed, want)
		}
	}
	// Enablement alone does not start a timer. Start only our fixture's unit to
	// prove actual scheduling and parsing, without changing TaskAction semantics.
	ctl("start", "--", timer)
	page, err := ListScheduledTaskPage(ctx, 10, base)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range page.Items {
		if item.Name == timer && item.Schedule != "" {
			found = true
		}
	}
	if !found {
		t.Fatalf("actual timer missing from production enumeration: %+v", page)
	}
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(fired); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timer did not execute its real oneshot service")
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Log(fmt.Sprintf("actual systemd start/restart/stop, enable/disable and timer execution passed (%s)", base))
}

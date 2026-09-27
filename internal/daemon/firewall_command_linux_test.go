//go:build linux

package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/systeminfo"
)

// Runs in an actual child process behind the test-only firewall-cmd wrapper.
// No host firewall command is invoked.
func TestFirewallCommandHelper(t *testing.T) {
	pos := -1
	for i, a := range os.Args {
		if a == "--" {
			pos = i
			break
		}
	}
	if pos < 0 || os.Getenv("BLORA_FW_TEST_STATE") == "" {
		return
	}
	args := os.Args[pos+1:]
	path := os.Getenv("BLORA_FW_TEST_STATE")
	var state systeminfo.FirewallSnapshot
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &state); err != nil {
		t.Fatal(err)
	}
	if len(args) == 1 && args[0] == "--get-default-zone" {
		fmt.Print(state.Zone)
		os.Exit(0)
	}
	// The legacy initial runtime observation is read-only.
	if len(args) == 1 && args[0] == "--list-ports" {
		fmt.Print(strings.Join(state.Runtime, " "))
		os.Exit(0)
	}
	if len(args) < 2 || args[0] != "--zone="+state.Zone {
		t.Fatalf("unbound/global firewall command: %v", args)
	}
	ports := &state.Runtime
	if len(args) == 3 && args[2] == "--permanent" {
		ports = &state.Permanent
	}
	if args[1] == "--list-ports" {
		fmt.Print(strings.Join(*ports, " "))
		os.Exit(0)
	}
	set := map[string]bool{}
	for _, p := range *ports {
		set[p] = true
	}
	switch {
	case strings.HasPrefix(args[1], "--add-port="):
		set[strings.TrimPrefix(args[1], "--add-port=")] = true
	case strings.HasPrefix(args[1], "--remove-port="):
		delete(set, strings.TrimPrefix(args[1], "--remove-port="))
	default:
		t.Fatalf("unexpected firewall command %v", args)
	}
	*ports = []string{}
	for p := range set {
		*ports = append(*ports, p)
	}
	sort.Strings(*ports)
	b, err = json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	os.Exit(0)
}

func TestFirewallDaemonCommandSnapshotAndStalePreview(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	before := systeminfo.FirewallSnapshot{Zone: "public", Runtime: []string{"22/tcp"}, Permanent: []string{"80/tcp"}}
	b, _ := json.Marshal(before)
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("BLORA_FW_TEST_BINARY", binary)
	t.Setenv("BLORA_FW_TEST_STATE", path)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	wrapper := "#!/bin/sh\nexec \"$BLORA_FW_TEST_BINARY\" -test.run=^TestFirewallCommandHelper$ -- \"$@\"\n"
	if err := os.WriteFile(filepath.Join(dir, "firewall-cmd"), []byte(wrapper), 0700); err != nil {
		t.Fatal(err)
	}
	d, cleanup := newFirewallLeaseTestDaemon(t)
	defer cleanup()
	newTask := func(id, hash string) model.Task {
		task := firewallTask(id, "actor", time.Now().Add(50*time.Second))
		task.Payload, _ = json.Marshal(map[string]any{"desired": []string{"443/tcp"}, "confirmUntil": time.Now().Add(50 * time.Second), "planHash": hash})
		persistFirewallTask(t, d, task)
		return task
	}
	stale := newTask("stale-command", strings.Repeat("0", 64))
	if _, err := d.runFirewallApply(context.Background(), stale); err == nil || !strings.Contains(err.Error(), "preview changed") {
		t.Fatalf("accepted stale preview: %v", err)
	}
	if _, err := d.store.Record(context.Background(), firewallLeaseNamespace, stale.ID, &firewallLeaseRecord{}); err == nil {
		t.Fatal("stale preview persisted a mutation lease")
	}
	task := newTask("command-rollback", systeminfo.FirewallPlanHash(before, []string{"443/tcp"}))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := d.runFirewallApply(ctx, task); done <- err }()
	deadline := time.Now().Add(30 * time.Second)
	for {
		d.firewallMu.Lock()
		lease := d.firewallLeases[task.ID]
		applied := lease != nil && lease.record.State == "applied"
		d.firewallMu.Unlock()
		if applied {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("command apply timed out")
		}
		time.Sleep(10 * time.Millisecond)
	}
	var record firewallLeaseRecord
	if _, err := d.store.Record(context.Background(), firewallLeaseNamespace, task.ID, &record); err != nil || record.Snapshot == nil || !sameFirewallLeaseSet(record.Snapshot.Permanent, before.Permanent) {
		t.Fatalf("missing durable dual snapshot %+v %v", record, err)
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("cancelled task succeeded")
		}
	case <-time.After(30 * time.Second):
		t.Fatal("rollback timeout")
	}
	after, err := systeminfo.CaptureFirewall(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !sameFirewallLeaseSet(after.Runtime, before.Runtime) || !sameFirewallLeaseSet(after.Permanent, before.Permanent) {
		t.Fatalf("incorrect restored configs %+v", after)
	}
	stored, err := d.store.Task(context.Background(), task.ID)
	if err != nil || stored.State != model.Cancelled {
		t.Fatalf("incorrect task result %+v %v", stored, err)
	}
}

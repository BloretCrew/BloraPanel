package systeminfo

import (
	"context"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestPreviewFirewallUsesManagedPortsAndRejectsInvalidRules(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux firewalld adapter")
	}
	old := firewallCommandContext
	defer func() { firewallCommandContext = old }()
	calls := 0
	firewallCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		calls++
		if reflect.DeepEqual(args, []string{"--get-default-zone"}) {
			return exec.CommandContext(ctx, "printf", "public")
		}
		if name != "firewall-cmd" || (len(args) != 2 && len(args) != 3) || args[0] != "--zone=public" || args[1] != "--list-ports" {
			t.Fatalf("unexpected preview command %s %v", name, args)
		}
		return exec.CommandContext(ctx, "sh", "-c", "printf '80/tcp 22/tcp 80/tcp 1000-2000/tcp'")
	}
	d, err := PreviewFirewall(context.Background(), []string{"443/tcp", "22/tcp"})
	if err != nil || !d.Apply || !reflect.DeepEqual(d.Add, []string{"443/tcp"}) || !reflect.DeepEqual(d.Remove, []string{"80/tcp"}) {
		t.Fatalf("wrong preview %+v %v", d, err)
	}
	if _, err := PreviewFirewall(context.Background(), []string{"70000/tcp"}); err == nil {
		t.Fatal("accepted rule apply would reject")
	}
	if d.Zone != "public" || len(d.PlanHash) != 64 || !reflect.DeepEqual(d.PermanentAdd, d.Add) {
		t.Fatalf("missing permanent preview %+v", d)
	}
	if calls != 3 {
		t.Fatalf("invalid rules reached backend: %d", calls)
	}
}

func snapshotFailureFixture(t *testing.T, fail func(bool, string) bool) (FirewallSnapshot, func() FirewallSnapshot) {
	t.Helper()
	old := firewallCommandContext
	t.Cleanup(func() { firewallCommandContext = old })
	state := FirewallSnapshot{Zone: "public", Runtime: []string{"80/tcp"}, Permanent: []string{"22/tcp"}}
	before := state
	firewallCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		if name != "firewall-cmd" || len(args) < 2 || args[0] != "--zone=public" {
			t.Fatalf("unbound command %s %v", name, args)
		}
		permanent := len(args) == 3 && args[2] == "--permanent"
		ports := &state.Runtime
		if permanent {
			ports = &state.Permanent
		}
		if args[1] == "--list-ports" {
			return exec.CommandContext(ctx, "printf", "%s", strings.Join(*ports, " "))
		}
		if fail != nil && fail(permanent, args[1]) {
			return exec.CommandContext(ctx, "false")
		}
		set := portSet(*ports)
		if strings.HasPrefix(args[1], "--add-port=") {
			set[strings.TrimPrefix(args[1], "--add-port=")] = true
		} else if strings.HasPrefix(args[1], "--remove-port=") {
			delete(set, strings.TrimPrefix(args[1], "--remove-port="))
		} else {
			t.Fatal(args)
		}
		*ports = []string{}
		for p := range set {
			*ports = append(*ports, p)
		}
		return exec.CommandContext(ctx, "true")
	}
	return before, func() FirewallSnapshot { return state }
}

func TestApplyFirewallReportsRollbackFailure(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux firewalld adapter")
	}
	before, _ := snapshotFailureFixture(t, func(permanent bool, action string) bool {
		return permanent && action == "--add-port=443/tcp" || !permanent && action == "--remove-port=443/tcp"
	})
	if err := ApplyFirewallSnapshot(context.Background(), before, []string{"443/tcp"}); err == nil {
		t.Fatal("apply failure hidden")
	}
	if err := RestoreFirewallSnapshot(context.Background(), before, []string{"443/tcp"}); err == nil || !strings.Contains(err.Error(), "restore firewall") {
		t.Fatalf("rollback failure hidden: %v", err)
	}
}

func TestApplyFirewallRollsBackAfterCommandFailure(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux firewalld adapter")
	}
	before, state := snapshotFailureFixture(t, func(permanent bool, action string) bool { return permanent && action == "--add-port=443/tcp" })
	if err := ApplyFirewallSnapshot(context.Background(), before, []string{"443/tcp"}); err == nil {
		t.Fatal("apply failure hidden")
	}
	if err := RestoreFirewallSnapshot(context.Background(), before, []string{"443/tcp"}); err != nil {
		t.Fatal(err)
	}
	after := state()
	if !samePortSet(after.Runtime, before.Runtime) || !samePortSet(after.Permanent, before.Permanent) {
		t.Fatalf("wrong rollback %+v", after)
	}
}

func TestApplyFirewallRollbackSurvivesCancelledParent(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux firewalld adapter")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	before, state := snapshotFailureFixture(t, func(permanent bool, action string) bool {
		if permanent && action == "--add-port=443/tcp" {
			cancel()
			return true
		}
		return false
	})
	if err := ApplyFirewallSnapshot(ctx, before, []string{"443/tcp"}); err == nil {
		t.Fatal("cancelled apply succeeded")
	}
	// The Daemon owns the independent bounded recovery context; exercise the
	// snapshot interface with that same separation after parent cancellation.
	rollbackCtx, stop := context.WithTimeout(context.Background(), 10*time.Second)
	defer stop()
	if err := RestoreFirewallSnapshot(rollbackCtx, before, []string{"443/tcp"}); err != nil {
		t.Fatal(err)
	}
	after := state()
	if !samePortSet(after.Runtime, before.Runtime) || !samePortSet(after.Permanent, before.Permanent) {
		t.Fatal("cancelled parent prevented restoration")
	}
}

func TestRollbackFirewallRefusesUnexpectedCurrentState(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux firewalld adapter")
	}
	before, _ := snapshotFailureFixture(t, nil)
	before.Runtime = []string{"53/udp"}
	if err := RestoreFirewallSnapshot(context.Background(), before, []string{"443/tcp"}); err == nil || !strings.Contains(err.Error(), "changed outside lease") {
		t.Fatalf("conflict ignored: %v", err)
	}
}

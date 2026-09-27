//go:build linux

package systeminfo

import (
	"context"
	"os/exec"
	"sort"
	"strings"
	"testing"
)

func TestFirewallPlanBindsBothConfigurationsZoneAndDesired(t *testing.T) {
	s := FirewallSnapshot{Zone: "public", Runtime: []string{"80/tcp", "22/tcp"}, Permanent: []string{"53/udp"}}
	h := FirewallPlanHash(s, []string{"443/tcp"})
	if h != FirewallPlanHash(FirewallSnapshot{Zone: "public", Runtime: []string{"22/tcp", "80/tcp", "22/tcp"}, Permanent: s.Permanent}, []string{"443/tcp", "443/tcp"}) {
		t.Fatal("canonical equivalent changed plan")
	}
	for _, changed := range []FirewallSnapshot{{Zone: "trusted", Runtime: s.Runtime, Permanent: s.Permanent}, {Zone: s.Zone, Runtime: []string{"22/tcp"}, Permanent: s.Permanent}, {Zone: s.Zone, Runtime: s.Runtime, Permanent: []string{"25/tcp"}}} {
		if h == FirewallPlanHash(changed, []string{"443/tcp"}) {
			t.Fatal("unbound snapshot field")
		}
	}
	if h == FirewallPlanHash(s, []string{"444/tcp"}) {
		t.Fatal("unbound desired rules")
	}
}

func TestFirewallSnapshotSeparatesConfigurationsAndRecoversPartialApply(t *testing.T) {
	for _, failPermanent := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "partial-failure"}[failPermanent], func(t *testing.T) {
			old := firewallCommandContext
			defer func() { firewallCommandContext = old }()
			states := map[bool]map[string]bool{false: portSet([]string{"22/tcp", "1000-2000/tcp"}), true: portSet([]string{"80/tcp", "3000-4000/tcp"})}
			fail := failPermanent
			firewallCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
				if name != "firewall-cmd" {
					t.Fatal(name)
				}
				if len(args) == 1 && args[0] == "--get-default-zone" {
					return exec.CommandContext(ctx, "printf", "public")
				}
				if len(args) < 2 || args[0] != "--zone=public" {
					t.Fatalf("unbound or global command: %v", args)
				}
				permanent := len(args) == 3 && args[2] == "--permanent"
				set := states[permanent]
				if args[1] == "--list-ports" {
					ports := []string{}
					for p := range set {
						ports = append(ports, p)
					}
					sort.Strings(ports)
					return exec.CommandContext(ctx, "printf", "%s", strings.Join(ports, " "))
				}
				if fail && permanent {
					fail = false
					return exec.CommandContext(ctx, "false")
				}
				if strings.HasPrefix(args[1], "--add-port=") {
					set[strings.TrimPrefix(args[1], "--add-port=")] = true
				} else if strings.HasPrefix(args[1], "--remove-port=") {
					delete(set, strings.TrimPrefix(args[1], "--remove-port="))
				} else {
					t.Fatalf("unexpected command %v", args)
				}
				return exec.CommandContext(ctx, "true")
			}
			ctx := context.Background()
			before, err := CaptureFirewall(ctx)
			if err != nil {
				t.Fatal(err)
			}
			err = ApplyFirewallSnapshot(ctx, before, []string{"443/tcp"})
			if (err != nil) != failPermanent {
				t.Fatalf("apply result %v", err)
			}
			if !failPermanent {
				if !states[false]["443/tcp"] || !states[true]["443/tcp"] || !states[false]["1000-2000/tcp"] || !states[true]["3000-4000/tcp"] {
					t.Fatal("lost separate configurations")
				}
			}
			if err := RestoreFirewallSnapshot(ctx, before, []string{"443/tcp"}); err != nil {
				t.Fatal(err)
			}
			after, err := readFirewallSnapshot(ctx, before.Zone)
			if err != nil || !samePortSet(after.Runtime, before.Runtime) || !samePortSet(after.Permanent, before.Permanent) {
				t.Fatalf("rollback mismatch %+v %v", after, err)
			}
			states[false]["53/udp"] = true
			if err := RestoreFirewallSnapshot(ctx, before, []string{"443/tcp"}); err == nil {
				t.Fatal("overwrote independent modification")
			}
		})
	}
}

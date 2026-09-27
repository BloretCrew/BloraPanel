package systeminfo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// FirewallSnapshot binds recovery to one zone and preserves both independent
// configurations. No reload or runtime-to-permanent operation is needed.
type FirewallSnapshot struct {
	Zone      string   `json:"zone"`
	Runtime   []string `json:"runtime"`
	Permanent []string `json:"permanent"`
}

func FirewallPlanHash(snapshot FirewallSnapshot, desired []string) string {
	canonical := func(xs []string) []string {
		set := portSet(xs)
		out := make([]string, 0, len(set))
		for x := range set {
			out = append(out, x)
		}
		sort.Strings(out)
		return out
	}
	snapshot.Runtime = canonical(snapshot.Runtime)
	snapshot.Permanent = canonical(snapshot.Permanent)
	b, _ := json.Marshal(struct {
		Snapshot FirewallSnapshot
		Desired  []string
	}{snapshot, canonical(desired)})
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

var firewallZone = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

func CaptureFirewall(ctx context.Context) (FirewallSnapshot, error) {
	b, err := firewallCommandContext(ctx, "firewall-cmd", "--get-default-zone").Output()
	if err != nil {
		return FirewallSnapshot{}, err
	}
	return readFirewallSnapshot(ctx, strings.TrimSpace(string(b)))
}
func readFirewallSnapshot(ctx context.Context, zone string) (FirewallSnapshot, error) {
	s := FirewallSnapshot{Zone: zone}
	if !firewallZone.MatchString(zone) {
		return s, errors.New("invalid firewall zone")
	}
	for _, permanent := range []bool{false, true} {
		args := []string{"--zone=" + zone, "--list-ports"}
		if permanent {
			args = append(args, "--permanent")
		}
		b, err := firewallCommandContext(ctx, "firewall-cmd", args...).Output()
		if err != nil {
			return s, err
		}
		if len(b) > 65536 {
			return s, errors.New("firewall port snapshot too large")
		}
		ports := strings.Fields(string(b))
		sort.Strings(ports)
		if permanent {
			s.Permanent = ports
		} else {
			s.Runtime = ports
		}
	}
	return s, nil
}
func portSet(xs []string) map[string]bool {
	out := map[string]bool{}
	for _, x := range xs {
		out[x] = true
	}
	return out
}
func samePortSet(a, b []string) bool {
	x, y := portSet(a), portSet(b)
	if len(x) != len(y) {
		return false
	}
	for v := range x {
		if !y[v] {
			return false
		}
	}
	return true
}
func targetPorts(previous, desired []string) []string {
	out := append([]string(nil), desired...)
	for _, p := range previous {
		if !validFirewalldPort(p) {
			out = append(out, p)
		}
	}
	return out
}
func validateDesiredPorts(desired []string) error {
	if len(desired) > 256 {
		return errors.New("too many firewall rules")
	}
	for _, p := range desired {
		if !validFirewalldPort(p) {
			return errors.New("only PORT/(tcp|udp) rules are supported")
		}
	}
	return nil
}
func updateZonePorts(ctx context.Context, zone string, permanent bool, current, target []string) error {
	a, b := portSet(current), portSet(target)
	for _, add := range []bool{true, false} {
		changes := []string{}
		if add {
			for p := range b {
				if !a[p] {
					changes = append(changes, p)
				}
			}
		} else {
			for p := range a {
				if !b[p] {
					changes = append(changes, p)
				}
			}
		}
		sort.Strings(changes)
		for _, p := range changes {
			if !validFirewalldPort(p) {
				return errors.New("refusing to mutate unmanaged firewall rule")
			}
			action := "--remove-port="
			if add {
				action = "--add-port="
			}
			args := []string{"--zone=" + zone, action + p}
			if permanent {
				args = append(args, "--permanent")
			}
			if err := firewallCommandContext(ctx, "firewall-cmd", args...).Run(); err != nil {
				return err
			}
		}
	}
	return nil
}

func ApplyFirewallSnapshot(ctx context.Context, before FirewallSnapshot, desired []string) error {
	if err := validateDesiredPorts(desired); err != nil {
		return err
	}
	current, err := readFirewallSnapshot(ctx, before.Zone)
	if err != nil {
		return err
	}
	if !samePortSet(current.Runtime, before.Runtime) || !samePortSet(current.Permanent, before.Permanent) {
		return errors.New("firewall changed since snapshot")
	}
	if err := updateZonePorts(ctx, before.Zone, false, current.Runtime, targetPorts(before.Runtime, desired)); err != nil {
		return err
	}
	if err := updateZonePorts(ctx, before.Zone, true, current.Permanent, targetPorts(before.Permanent, desired)); err != nil {
		return err
	}
	after, err := readFirewallSnapshot(ctx, before.Zone)
	if err != nil {
		return err
	}
	if !samePortSet(after.Runtime, targetPorts(before.Runtime, desired)) || !samePortSet(after.Permanent, targetPorts(before.Permanent, desired)) {
		return errors.New("firewall applied result not confirmed")
	}
	return nil
}

// Partial progress is recoverable only when every differing membership belongs
// to this change. An unrelated rule change is a conflict, never overwritten.
func allowedPartial(current, previous, target []string) bool {
	c, p, t := portSet(current), portSet(previous), portSet(target)
	for x := range c {
		if !p[x] && !t[x] {
			return false
		}
	}
	for x := range p {
		if t[x] && !c[x] {
			return false
		}
	}
	return true
}
func RestoreFirewallSnapshot(ctx context.Context, before FirewallSnapshot, desired []string) error {
	if err := validateDesiredPorts(desired); err != nil {
		return err
	}
	current, err := readFirewallSnapshot(ctx, before.Zone)
	if err != nil {
		return err
	}
	if !allowedPartial(current.Runtime, before.Runtime, targetPorts(before.Runtime, desired)) || !allowedPartial(current.Permanent, before.Permanent, targetPorts(before.Permanent, desired)) {
		return errors.New("firewall changed outside lease")
	}
	for _, permanent := range []bool{false, true} {
		from, to := current.Runtime, before.Runtime
		if permanent {
			from, to = current.Permanent, before.Permanent
		}
		if err := updateZonePorts(ctx, before.Zone, permanent, from, to); err != nil {
			return fmt.Errorf("restore firewall: %w", err)
		}
	}
	after, err := readFirewallSnapshot(ctx, before.Zone)
	if err != nil {
		return err
	}
	if !samePortSet(after.Runtime, before.Runtime) || !samePortSet(after.Permanent, before.Permanent) {
		return errors.New("firewall rollback result not confirmed")
	}
	return nil
}

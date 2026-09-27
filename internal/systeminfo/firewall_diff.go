package systeminfo

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

var firewalldPort = regexp.MustCompile(`^[0-9]{1,5}/(tcp|udp)$`)

var firewallCommandContext = exec.CommandContext

func validFirewalldPort(rule string) bool {
	if !firewalldPort.MatchString(rule) {
		return false
	}
	parts := strings.Split(rule, "/")
	var n int
	_, _ = fmt.Sscanf(parts[0], "%d", &n)
	return n > 0 && n <= 65535
}

// CurrentFirewallPorts returns the managed firewalld port subset in a stable
// order. Rich rules and other forms are deliberately left to the host and are
// not silently interpreted by the reversible port adapter.
func CurrentFirewallPorts(ctx context.Context) ([]string, error) {
	if runtime.GOOS != "linux" {
		return nil, errors.New("firewall port backend unavailable")
	}
	b, err := firewallCommandContext(ctx, "firewall-cmd", "--list-ports").Output()
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	ports := make([]string, 0)
	for _, rule := range strings.Fields(string(b)) {
		if validFirewalldPort(rule) && !seen[rule] {
			seen[rule] = true
			ports = append(ports, rule)
		}
	}
	sort.Strings(ports)
	return ports, nil
}

type FirewallDiff struct {
	Zone             string   `json:"zone"`
	PlanHash         string   `json:"planHash"`
	PermanentCurrent []string `json:"permanentCurrent"`
	PermanentAdd     []string `json:"permanentAdd"`
	PermanentRemove  []string `json:"permanentRemove"`
	Backend          string   `json:"backend"`
	Current          []string `json:"current"`
	Add              []string `json:"add"`
	Remove           []string `json:"remove"`
	Apply            bool     `json:"apply"`
}

func PreviewFirewall(ctx context.Context, desired []string) (FirewallDiff, error) {
	if len(desired) > 256 {
		return FirewallDiff{}, errors.New("too many firewall rules")
	}
	for _, r := range desired {
		if len(r) == 0 || len(r) > 512 || strings.ContainsAny(r, "\x00\r\n") {
			return FirewallDiff{}, errors.New("invalid firewall rule")
		}
	}
	if runtime.GOOS != "linux" {
		return FirewallDiff{Backend: "unavailable"}, errors.New("firewall port preview backend unavailable")
	}
	for _, rule := range desired {
		if !validFirewalldPort(rule) {
			return FirewallDiff{}, errors.New("only PORT/(tcp|udp) rules are supported")
		}
	}
	snapshot, e := CaptureFirewall(ctx)
	if e != nil {
		return FirewallDiff{Backend: "firewalld"}, e
	}
	managed := func(xs []string) []string {
		out := []string{}
		for _, p := range xs {
			if validFirewalldPort(p) {
				out = append(out, p)
			}
		}
		return out
	}
	current := managed(snapshot.Runtime)
	set := func(xs []string) map[string]bool {
		m := map[string]bool{}
		for _, x := range xs {
			if strings.TrimSpace(x) != "" {
				m[x] = true
			}
		}
		return m
	}
	a, b := set(current), set(desired)
	d := FirewallDiff{Backend: "firewalld", Current: current, Apply: true, Add: []string{}, Remove: []string{}}
	d.Zone = snapshot.Zone
	d.PlanHash = FirewallPlanHash(snapshot, desired)
	d.PermanentCurrent = managed(snapshot.Permanent)
	d.PermanentAdd = []string{}
	d.PermanentRemove = []string{}
	permanent := set(d.PermanentCurrent)
	for x := range b {
		if !permanent[x] {
			d.PermanentAdd = append(d.PermanentAdd, x)
		}
	}
	for x := range permanent {
		if !b[x] {
			d.PermanentRemove = append(d.PermanentRemove, x)
		}
	}
	sort.Strings(d.PermanentAdd)
	sort.Strings(d.PermanentRemove)
	for x := range b {
		if !a[x] {
			d.Add = append(d.Add, x)
		}
	}
	for x := range a {
		if !b[x] {
			d.Remove = append(d.Remove, x)
		}
	}
	sort.Strings(d.Add)
	sort.Strings(d.Remove)
	return d, nil
}

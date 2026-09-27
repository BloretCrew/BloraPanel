package systeminfo

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"strings"
)

type FirewallStatus struct {
	Backend string `json:"backend"`
	State   string `json:"state"`
	Detail  string `json:"detail,omitempty"`
}

func Firewall(ctx context.Context) (FirewallStatus, error) {
	x := FirewallStatus{Backend: "unavailable", State: "unknown"}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.CommandContext(ctx, "firewall-cmd", "--state")
		x.Backend = "firewalld"
	case "windows":
		cmd = exec.CommandContext(ctx, "netsh.exe", "advfirewall", "show", "allprofiles", "state")
		x.Backend = "netsh"
	case "darwin":
		cmd = exec.CommandContext(ctx, "pfctl", "-s", "info")
		x.Backend = "pf"
	default:
		return x, errors.New("firewall backend unavailable")
	}
	b, e := cmd.Output()
	if e != nil {
		return x, e
	}
	x.Detail = strings.TrimSpace(string(b))
	if strings.Contains(strings.ToLower(x.Detail), "running") || strings.Contains(strings.ToLower(x.Detail), "on") || strings.Contains(strings.ToLower(x.Detail), "enabled") {
		x.State = "active"
	} else {
		x.State = "reported"
	}
	return x, nil
}

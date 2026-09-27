//go:build linux

package systeminfo

import (
	"context"
	"strings"
	"testing"
)

func TestLinuxUnitActionsRejectPathsAndOtherUnitTypesBeforeExecution(t *testing.T) {
	// No systemctl exists here. A missing-executable error would mean a bad
	// target reached execution instead of being rejected by the adapter.
	t.Setenv("PATH", t.TempDir())
	for _, name := range []string{"network.target", "data.mount", "demo.socket", "demo.timer", "sshd"} {
		err := ServiceAction(context.Background(), name, "start")
		if err == nil || !strings.Contains(err.Error(), "service action must identify") {
			t.Errorf("service target %q reached execution: %v", name, err)
		}
	}
	for _, name := range []string{"/tmp/untrusted.timer", `folder\untrusted.timer`, "folder/untrusted.timer", "demo.service"} {
		err := TaskAction(context.Background(), name, "enable")
		if err == nil || !strings.Contains(err.Error(), "scheduled task must identify") {
			t.Errorf("timer target %q reached execution: %v", name, err)
		}
	}
	for _, name := range []string{"sshd.service", "worker@one.service"} {
		err := ServiceAction(context.Background(), name, "start")
		if err == nil || !strings.Contains(err.Error(), "executable file not found") {
			t.Errorf("valid service %q rejected before command lookup: %v", name, err)
		}
	}
	for _, name := range []string{"backup.timer", "backup@one.timer"} {
		err := TaskAction(context.Background(), name, "enable")
		if err == nil || !strings.Contains(err.Error(), "executable file not found") {
			t.Errorf("valid timer %q rejected before command lookup: %v", name, err)
		}
	}
}

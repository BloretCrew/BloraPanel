//go:build linux

package master

import (
	"strings"
	"testing"

	"blora.dev/panel/internal/model"
)

func TestSystemUnitTasksRejectPathsAndWrongTypesAtDaemon(t *testing.T) {
	// An empty executable search path also ensures a regression cannot invoke
	// the host's service manager during this authorization/dispatch test.
	t.Setenv("PATH", t.TempDir())
	f := newFileFixture(t)
	base := "/nodes/" + f.instances[0].NodeID + "/system/"
	for _, sample := range []struct{ endpoint, name, action, diagnostic string }{
		{"services/actions", "blora-test.mount", "start", "service action must identify a service unit"},
		{"services/actions", "blora-test.target", "stop", "service action must identify a service unit"},
		{"tasks/actions", "/tmp/blora-test.timer", "enable", "scheduled task must identify a timer unit"},
		{"tasks/actions", `folder\blora-test.timer`, "disable", "scheduled task must identify a timer unit"},
	} {
		task := parseFileTask(t, f.admin.request("POST", base+sample.endpoint, map[string]string{"name": sample.name, "action": sample.action}, model.ID(), 202))
		task = f.awaitFile(t, task, model.Failed)
		if !strings.Contains(task.Error, sample.diagnostic) {
			t.Fatalf("target %q reached command execution: %s", sample.name, task.Error)
		}
	}
}

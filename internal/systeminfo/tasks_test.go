package systeminfo

import (
	"context"
	"runtime"
	"strings"
	"testing"
)

func TestTaskActionRejectsUnsafeInputBeforeBackend(t *testing.T) {
	if err := TaskAction(context.Background(), "../../etc", "enable"); err == nil {
		t.Fatal("path traversal task accepted")
	}
	if err := TaskAction(context.Background(), "safe.timer", "delete"); err == nil {
		t.Fatal("unsupported task action accepted")
	}
	if runtime.GOOS == "linux" {
		if err := TaskAction(context.Background(), "sshd.service", "disable"); err == nil {
			t.Fatal("service accepted as timer")
		}
	}
}

func TestWindowsTaskPathsSupportUnicodeAndRejectTraversal(t *testing.T) {
	for _, name := range []string{`\我的任务\每日 备份`, `\Vendor\Nightly task`, `\User's task`} {
		if !ValidTaskName(name) {
			t.Fatalf("rejected valid path %q", name)
		}
	}
	for _, name := range []string{`\`, `\folder\\task`, `\folder\..\task`, `\folder\?task`, "\\folder\\bad\nname", `\folder/other\task`} {
		if ValidTaskName(name) {
			t.Fatalf("accepted unsafe path %q", name)
		}
	}
}

func TestTaskCOMOutputIsBounded(t *testing.T) {
	var output taskOutput
	if _, err := output.Write([]byte(strings.Repeat("x", 1<<20))); err != nil {
		t.Fatal(err)
	}
	if _, err := output.Write([]byte("x")); err == nil {
		t.Fatal("output limit not enforced")
	}
	if output.Len() != 1<<20 {
		t.Fatal("overflow changed retained output")
	}
}

func TestLinuxTimerListUsesUnitRatherThanActivatedService(t *testing.T) {
	items := parseLinuxTimers("Sat 2026-09-12 12:00:00 UTC 3h left Fri 2026-09-11 12:00:00 UTC 21h ago backup.timer backup.service\nn/a n/a n/a n/a inactive.timer inactive.service\n2 timers listed.\n", 100)
	if len(items) != 2 || items[0].Name != "backup.timer" || items[1].Name != "inactive.timer" {
		t.Fatalf("incorrect timer identities %+v", items)
	}
	if items[0].Schedule == "" {
		t.Fatal("lost schedule observation")
	}
}

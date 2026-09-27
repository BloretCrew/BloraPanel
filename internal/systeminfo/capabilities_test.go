package systeminfo

import "testing"

func TestDetectReportsPlatformAndBoundedBackends(t *testing.T) {
	c := Detect()
	if c.Platform == "" {
		t.Fatal("platform missing")
	}
	for _, v := range []string{c.Services, c.Firewall, c.Tasks} {
		if v == "" {
			t.Fatal("empty capability")
		}
	}
}

func TestTaskNameBoundaries(t *testing.T) {
	for _, name := range []string{"systemd-timer.service", `\Microsoft\Windows\Update`} {
		if !ValidTaskName(name) {
			t.Fatalf("valid task rejected: %q", name)
		}
	}
	for _, name := range []string{"", "../../etc", "task name", "x;reboot"} {
		if ValidTaskName(name) {
			t.Fatalf("unsafe task accepted: %q", name)
		}
	}
}

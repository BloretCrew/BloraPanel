package systeminfo

import (
	"context"
	"strings"
	"testing"
)

func TestServiceActionRejectsUnsafeInputs(t *testing.T) {
	if err := ServiceAction(context.Background(), "../../etc/passwd", "start"); err == nil {
		t.Fatal("path traversal service name accepted")
	}
	if err := ServiceAction(context.Background(), "demo.service", "exec /bin/sh"); err == nil {
		t.Fatal("arbitrary action accepted")
	}
}

func TestServiceIdentitySupportsWindowsNames(t *testing.T) {
	for _, name := range []string{"Windows Update", "中文服务", "Vendor's service"} {
		if !ValidServiceName(name) {
			t.Fatalf("valid service rejected %q", name)
		}
	}
	for _, name := range []string{"", `folder\service`, "service/name", "bad\nname", strings.Repeat("x", 257)} {
		if ValidServiceName(name) {
			t.Fatalf("invalid service accepted %q", name)
		}
	}
}

func TestFirewallPreviewRejectsUnboundedRules(t *testing.T) {
	rules := make([]string, 257)
	if _, err := PreviewFirewall(context.Background(), rules); err == nil {
		t.Fatal("unbounded rules accepted")
	}
	if _, err := PreviewFirewall(context.Background(), []string{"bad\nrule"}); err == nil {
		t.Fatal("newline rule accepted")
	}
}

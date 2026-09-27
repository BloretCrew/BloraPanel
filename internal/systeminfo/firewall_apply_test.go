package systeminfo

import (
	"context"
	"testing"
)

func TestApplyFirewallRejectsUnsupportedRulesBeforeBackend(t *testing.T) {
	for _, rules := range [][]string{{"80/tcp; reboot"}, {"70000/tcp"}, {"80/tcp\n22/tcp"}} {
		if err := ApplyFirewallSnapshot(context.Background(), FirewallSnapshot{}, rules); err == nil {
			t.Fatalf("unsafe firewall rule accepted: %q", rules)
		}
	}
}

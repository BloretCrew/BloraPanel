package runtime

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestNativeCapabilitiesCreatesAndCleansOnlyItsProbe(t *testing.T) {
	if nativeBackend() == "unsupported" {
		t.Skip("unsupported native platform")
	}
	root := t.TempDir()
	m, err := New(Options{StateRoot: root, AllowPGIDFallback: true})
	if err != nil {
		t.Fatal(err)
	}
	capabilities := m.Capabilities(context.Background())
	if capabilities["native"] != "available" {
		t.Fatalf("actual native process probe failed: %s", capabilities["native.reason"])
	}
	if nativeBackend() == "linux" && capabilities["native.mode"] != "pgid_trusted" {
		t.Fatal("PGID probe claimed cgroup isolation")
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "cap-") {
			t.Fatal("completed capability probe left runtime records")
		}
	}
}

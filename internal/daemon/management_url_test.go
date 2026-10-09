package daemon

import (
	"path/filepath"
	"testing"
)

func TestManagementWebSocketSchemes(t *testing.T) {
	for input, want := range map[string]string{"https://panel.example/": "wss://panel.example", "http://127.0.0.1:37861/": "ws://127.0.0.1:37861", "http://[::1]:37861": "ws://[::1]:37861"} {
		if got := managementWebSocketURL(input); got != want {
			t.Fatalf("%s: %s", input, got)
		}
	}
}

func TestManagementHTTPRejectsExternalBeforeState(t *testing.T) {
	for _, endpoint := range []string{"http://192.0.2.1:37861", "http://panel.example", "ftp://localhost:37861"} {
		state := filepath.Join(t.TempDir(), "state")
		if d, err := New(Config{StateDir: state, MasterURL: endpoint}); err == nil {
			d.Close()
			t.Fatalf("accepted non-local insecure management endpoint %s", endpoint)
		}
	}
}

package master

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"blora.dev/panel/internal/daemon"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/storage"
)

func TestLoopbackHTTPSourceAndProxyModes(t *testing.T) {
	proxies, _ := ParseTrustedProxies("127.0.0.1")
	for _, tc := range []struct {
		name, peer, host, forwarded, proto, want string
		all, strict, reject                      bool
	}{
		{name: "direct", peer: "127.0.0.1:9", host: "localhost:37861", want: "http://localhost:37861"},
		{name: "HTTP default port", peer: "[::1]:9", host: "[::1]:80", want: "http://[::1]"},
		{name: "default all proxies", peer: "127.0.0.1:9", host: "localhost:37861", forwarded: "panel.example:443", proto: "https", all: true, want: "https://panel.example"},
		{name: "strict trusted", peer: "127.0.0.1:9", host: "localhost:37861", forwarded: "panel.example", proto: "https", strict: true, want: "https://panel.example"},
		{name: "strict untrusted", peer: "127.0.0.2:9", host: "localhost:37861", forwarded: "panel.example", proto: "https", strict: true, want: "http://localhost:37861"},
		{name: "non-loopback cannot forge TLS", peer: "192.0.2.9:7", host: "localhost", forwarded: "panel.example", proto: "https", all: true, reject: true},
		{name: "external HTTP forbidden", peer: "127.0.0.1:9", host: "localhost", forwarded: "panel.example", proto: "http", all: true, reject: true},
		{name: "missing proto", peer: "127.0.0.1:9", host: "localhost", forwarded: "panel.example", all: true, reject: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &Server{allowLoopbackHTTP: true, trustAllProxies: tc.all}
			if tc.strict {
				s.trustedProxies = proxies
			}
			r := httptest.NewRequest("GET", "http://localhost/", nil)
			r.RemoteAddr, r.Host = tc.peer, tc.host
			if tc.forwarded != "" {
				r.Header.Set("X-Forwarded-Host", tc.forwarded)
			}
			if tc.proto != "" {
				r.Header.Set("X-Forwarded-Proto", tc.proto)
			}
			got, err := s.resolveSource(r)
			if tc.reject {
				if err == nil {
					t.Fatal("unsafe transport/source accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if s.requestOrigin(got) != tc.want || !managementTransport(got) || got.TLS != nil {
				t.Fatalf("incorrect source/transport: %s, %v", s.requestOrigin(got), got.TLS)
			}
		})
	}
}

func TestLoopbackHTTPDaemonControlAndData(t *testing.T) {
	base := t.TempDir()
	store, err := storage.Open(filepath.Join(base, "master.db"))
	if err != nil {
		t.Fatal(err)
	}
	app := New(Options{Store: store, AllowLoopbackHTTP: true, TrustAllProxies: true})
	server := httptest.NewServer(app)
	ctx, cancel := context.WithCancel(context.Background())
	token, err := store.Enrollment(ctx, "loopback-node")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(base, "node.enrollment")
	if err := os.WriteFile(path, []byte(token), 0600); err != nil {
		t.Fatal(err)
	}
	d, err := daemon.New(daemon.Config{StateDir: filepath.Join(base, "node"), MasterURL: server.URL, EnrollmentFile: path, AllowPGIDFallback: true})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- d.Run(ctx) }()
	t.Cleanup(func() {
		cancel()
		_ = d.Close()
		_ = app.Close()
		server.Close()
		_ = store.Close()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Error("daemon shutdown timed out")
		}
	})
	eventually(t, 10*time.Second, func() bool {
		node, _, err := store.Node(ctx, d.NodeID())
		app.mu.Lock()
		defer app.mu.Unlock()
		return err == nil && node.State == "ONLINE" && app.peers[d.NodeID()] != nil && app.links[dataKey(d.NodeID(), protocol.ChannelBulk)] != nil
	})
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(base, "node", "tls.crt")); !os.IsNotExist(err) {
		t.Fatal("loopback Daemon unexpectedly needs a local TLS certificate")
	}
}

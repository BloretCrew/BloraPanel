package master

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
	"github.com/coder/websocket"
	"golang.org/x/crypto/bcrypt"
)

func TestRequestSourceTrustBoundary(t *testing.T) {
	proxies, err := ParseTrustedProxies("127.0.0.1,10.0.0.0/8,::1")
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"*", "localhost", "127.0.0.1,", "bad/8"} {
		if _, err := ParseTrustedProxies(value); err == nil {
			t.Fatalf("accepted proxy %q", value)
		}
	}
	for _, tc := range []struct {
		name, peer, host, xfhost, proto, xff, real, origin, wantIP, wantOrigin string
		reject                                                                 bool
	}{
		{name: "local IP", peer: "127.0.0.1:4", host: "127.0.0.1:8443", wantIP: "127.0.0.1", wantOrigin: "https://127.0.0.1:8443"},
		{name: "local IPv6", peer: "[::1]:4", host: "[::1]:8443", wantIP: "::1", wantOrigin: "https://[::1]:8443"},
		{name: "untrusted spoof", peer: "192.0.2.8:4", host: "localhost:8443", xfhost: "evil.example", proto: "http", xff: "fake", wantIP: "192.0.2.8", wantOrigin: "https://localhost:8443"},
		{name: "trusted external host", peer: "127.0.0.1:4", host: "localhost:8443", xfhost: "panel.example:444", proto: "https", xff: "198.51.100.7", wantIP: "198.51.100.7", wantOrigin: "https://panel.example:444"},
		{name: "default HTTPS port", peer: "127.0.0.1:4", host: "localhost", xfhost: "panel.example:443", proto: "https", wantIP: "127.0.0.1", wantOrigin: "https://panel.example"},
		{name: "default IPv6 HTTPS port", peer: "127.0.0.1:4", host: "localhost", xfhost: "[2001:db8::1]:443", proto: "https", wantIP: "127.0.0.1", wantOrigin: "https://[2001:db8::1]"},
		{name: "forged leftmost chain", peer: "127.0.0.1:4", host: "localhost", xff: "203.0.113.9,198.51.100.7,10.1.2.3", wantIP: "198.51.100.7", wantOrigin: "https://localhost"},
		{name: "real IPv6", peer: "[::1]:4", host: "localhost", real: "2001:db8::8", wantIP: "2001:db8::8", wantOrigin: "https://localhost"},
		{name: "pinned origin", peer: "127.0.0.1:4", host: "localhost", xfhost: "other.example", proto: "https", origin: "https://fixed.example", wantIP: "127.0.0.1", wantOrigin: "https://fixed.example"},
		{name: "direct external", peer: "192.0.2.8:4", host: "evil.example", reject: true},
		{name: "insecure external", peer: "127.0.0.1:4", host: "localhost", xfhost: "panel.example", proto: "http", reject: true},
		{name: "host list", peer: "127.0.0.1:4", host: "localhost", xfhost: "one.example,two.example", proto: "https", reject: true},
		{name: "host credentials", peer: "127.0.0.1:4", host: "localhost", xfhost: "user@panel.example", proto: "https", reject: true},
		{name: "bad port", peer: "127.0.0.1:4", host: "localhost", xfhost: "panel.example:99999", proto: "https", reject: true},
		{name: "bad chain", peer: "127.0.0.1:4", host: "localhost", xff: "bad,198.51.100.7", reject: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &Server{trustedProxies: proxies, origin: tc.origin}
			r := httptest.NewRequest("GET", "https://localhost/", nil)
			r.Host, r.RemoteAddr = tc.host, tc.peer
			for k, v := range map[string]string{"X-Forwarded-Host": tc.xfhost, "X-Forwarded-Proto": tc.proto, "X-Forwarded-For": tc.xff, "X-Real-IP": tc.real} {
				if v != "" {
					r.Header.Set(k, v)
				}
			}
			got, err := s.resolveSource(r)
			if tc.reject {
				if err == nil {
					t.Fatal("accepted invalid source")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.RemoteAddr != "["+tc.wantIP+"]:0" && got.RemoteAddr != tc.wantIP+":0" {
				t.Fatalf("IP: %s", got.RemoteAddr)
			}
			if s.requestOrigin(got) != tc.wantOrigin {
				t.Fatalf("origin: %s", s.requestOrigin(got))
			}
			if r.RemoteAddr != tc.peer || r.Host != tc.host {
				t.Fatal("mutated input request")
			}
		})
	}
	s := &Server{trustedProxies: proxies}
	// Proxy trust is opt-in even for local peers.
	localRequest := httptest.NewRequest("GET", "https://localhost:8443/", nil)
	localRequest.RemoteAddr = "127.0.0.1:4"
	localRequest.Header.Set("X-Forwarded-Host", "attacker.example")
	localRequest.Header.Set("X-Forwarded-Proto", "https")
	localRequest.Header.Set("X-Forwarded-For", "203.0.113.9")
	untrusted := &Server{}
	resolved, err := untrusted.resolveSource(localRequest)
	if err != nil || untrusted.requestOrigin(resolved) != "https://localhost:8443" || resolved.RemoteAddr != "127.0.0.1:0" {
		t.Fatalf("default trusted a local peer's headers: %v", err)
	}
	r := httptest.NewRequest("GET", "https://localhost/", nil)
	r.RemoteAddr = "127.0.0.1:4"
	r.Header.Add("X-Forwarded-Host", "one.example")
	r.Header.Add("X-Forwarded-Host", "two.example")
	if _, err := s.resolveSource(r); err == nil {
		t.Fatal("accepted duplicate host headers")
	}
	r.Header = make(http.Header)
	r.TLS = nil
	if _, err := s.resolveSource(r); err == nil {
		t.Fatal("manufactured TLS")
	}
}

func TestProxySourceRealTLSLoginAndWebSocket(t *testing.T) {
	db, err := storage.Open(filepath.Join(t.TempDir(), "master.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	hash, err := bcrypt.GenerateFromPassword([]byte("private-test-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.CreateUser(context.Background(), model.User{Name: "admin", Admin: true}, hash); err != nil {
		t.Fatal(err)
	}
	proxies, _ := ParseTrustedProxies("127.0.0.1")
	app := New(Options{Store: db, TrustedProxies: proxies})
	defer app.Close()
	backend := httptest.NewTLSServer(app)
	defer backend.Close()
	u, _ := url.Parse(backend.URL)
	p := httputil.NewSingleHostReverseProxy(u)
	p.Transport = backend.Client().Transport
	director := p.Director
	p.Director = func(r *http.Request) {
		externalHost := r.Host
		director(r)
		r.Host = u.Host
		r.Header.Set("X-Forwarded-Host", externalHost)
		r.Header.Set("X-Forwarded-Proto", "https")
		// nil prevents ReverseProxy from appending: this simulates Nginx overwriting XFF.
		r.Header["X-Forwarded-For"] = nil
		r.Header.Set("X-Real-IP", "198.51.100.44")
	}
	proxy := httptest.NewTLSServer(p)
	defer proxy.Close()
	c := proxy.Client()
	c.Jar, _ = cookiejar.New(nil)
	request := func(origin string, want int) {
		r, _ := http.NewRequest("POST", proxy.URL+"/api/v1/login", bytes.NewBufferString(`{"name":"admin","password":"private-test-password"}`))
		r.Header.Set("Origin", origin)
		r.Header.Set("Content-Type", "application/json")
		resp, e := c.Do(r)
		if e != nil {
			t.Fatal(e)
		}
		defer resp.Body.Close()
		if resp.StatusCode != want {
			b, _ := io.ReadAll(resp.Body)
			t.Fatalf("login %d: %s", resp.StatusCode, b)
		}
	}
	request("https://attacker.example", 403)
	request(proxy.URL, 200)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ws, _, err := websocket.Dial(ctx, "wss"+proxy.URL[5:]+"/api/v1/events", &websocket.DialOptions{HTTPClient: c, HTTPHeader: http.Header{"Origin": []string{proxy.URL}}})
	if err != nil {
		t.Fatal(err)
	}
	ws.CloseNow()
	if _, _, err = websocket.Dial(ctx, "wss"+proxy.URL[5:]+"/api/v1/events", &websocket.DialOptions{HTTPClient: c, HTTPHeader: http.Header{"Origin": []string{"https://attacker.example"}}}); err == nil {
		t.Fatal("cross-origin WSS accepted")
	}
	app.mu.Lock()
	attempts := app.attempts["198.51.100.44"].Count
	app.mu.Unlock()
	if attempts != 1 {
		t.Fatalf("rate limiter client attribution: %d", attempts)
	}
	// No fixed-origin setting: direct loopback access also works with its own port.
	r, _ := http.NewRequest("GET", backend.URL+"/api/v1/session", nil)
	r.Header.Set("Origin", backend.URL)
	resp, err := backend.Client().Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var payload map[string]any
	json.NewDecoder(resp.Body).Decode(&payload)
	if resp.StatusCode != 401 {
		t.Fatalf("local origin rejected: %d", resp.StatusCode)
	}
	if r.TLS != nil {
		t.Fatal("unexpected client-side request mutation")
	}
}

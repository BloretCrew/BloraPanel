package containers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestOperationLostEngineAcknowledgementNeverReplays(t *testing.T) {
	var creates atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/version" {
			json.NewEncoder(w).Encode(map[string]string{"ApiVersion": "1.54", "MinAPIVersion": "1.40", "Os": "linux"})
			return
		}
		if r.Method == "POST" && r.URL.Path == "/v1.45/containers/create" {
			creates.Add(1)
			conn, _, err := w.(http.Hijacker).Hijack()
			if err == nil {
				conn.Close()
			}
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	root := t.TempDir()
	m := saveTestManager(t, root)
	m.engine = &engine{client: server.Client(), base: server.URL}
	op := Operation{TaskID: "lost_ack", Action: "container.create", Name: "test-only", Config: json.RawMessage(`{"Image":"test-only"}`)}
	r, err := m.Execute(context.Background(), "admin", op, nil)
	if err == nil || !r.Unknown || r.State != "INTERRUPTED" || creates.Load() != 1 {
		t.Fatal(r, err, creates.Load())
	}
	m = saveTestManager(t, root)
	m.engine = &engine{client: server.Client(), base: server.URL}
	r, err = m.Execute(context.Background(), "admin", op, nil)
	if !errors.Is(err, ErrUnknown) || creates.Load() != 1 {
		t.Fatal("side effect replayed", r, err, creates.Load())
	}
	op.Name = "different"
	if _, err = m.Execute(context.Background(), "admin", op, nil); !errors.Is(err, ErrConflict) {
		t.Fatal("task identity reused", err)
	}
}

func TestVolumeIdentityRecheckedAfterProgress(t *testing.T) {
	var generation atomic.Int32
	var deletes atomic.Int32
	generation.Store(1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/version" {
			json.NewEncoder(w).Encode(map[string]string{"ApiVersion": "1.54", "MinAPIVersion": "1.40", "Os": "linux"})
			return
		}
		if r.Method == "DELETE" {
			deletes.Add(1)
			w.WriteHeader(204)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/v1.45/volumes/") {
			v := volumeInspect{Name: "test-volume", CreatedAt: "2026-09-09T00:00:00Z", Labels: map[string]string{"generation": "one"}}
			if generation.Load() == 2 {
				v.Labels["generation"] = "two"
			}
			json.NewEncoder(w).Encode(v)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	m := saveTestManager(t, t.TempDir())
	m.engine = &engine{client: server.Client(), base: server.URL}
	o, err := m.engine.inspect(context.Background(), Target{Kind: "volume", Name: "test-volume"}, false)
	if err != nil {
		t.Fatal(err)
	}
	r, err := m.Execute(context.Background(), "admin", Operation{TaskID: "delete_replaced_volume", Action: "volume.delete", Target: o.Target}, func(p Progress) error {
		if p.Phase == "delete-volume" {
			generation.Store(2)
		}
		return nil
	})
	if !errors.Is(err, ErrIdentity) || deletes.Load() != 0 || r.Unknown {
		t.Fatal(r, err, deletes.Load())
	}
}

func TestIdleDockerLogStreamClosesOnPermissionRevocation(t *testing.T) {
	var permitted atomic.Bool
	permitted.Store(true)
	id := strings.Repeat("a", 64)
	attached := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/version":
			json.NewEncoder(w).Encode(map[string]string{"ApiVersion": "1.54", "MinAPIVersion": "1.40", "Os": "linux"})
		case strings.HasSuffix(r.URL.Path, "/json"):
			json.NewEncoder(w).Encode(map[string]any{"Id": id, "Config": map[string]any{"Tty": false}})
		case strings.HasSuffix(r.URL.Path, "/logs"):
			w.WriteHeader(200)
			w.(http.Flusher).Flush()
			close(attached)
			<-r.Context().Done()
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	m, err := New(Options{StateRoot: t.TempDir(), Authorize: func(context.Context, string) error {
		if permitted.Load() {
			return nil
		}
		return ErrForbidden
	}})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	m.engine = &engine{client: server.Client(), base: server.URL}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- m.Logs(ctx, "admin", Target{Kind: "container", ID: id}, LogOptions{Follow: true}, func(LogFrame) error { return nil })
	}()
	select {
	case <-attached:
	case <-ctx.Done():
		t.Fatal("logs did not attach")
	}
	permitted.Store(false)
	select {
	case err := <-done:
		if !errors.Is(err, ErrForbidden) {
			t.Fatal("revoked stream returned wrong result", err)
		}
	case <-ctx.Done():
		t.Fatal("idle stream survived revocation")
	}
}

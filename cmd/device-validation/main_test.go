package main

import (
	"blora.dev/panel/internal/model"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Failed cleanup must preserve private recovery state and propagate a failure,
// even when the main workflow itself had already succeeded.
func TestCleanupFailureKeepsPrivateStateAndFailsReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(202)
			_ = json.NewEncoder(w).Encode(map[string]any{"task": model.Task{ID: "owned-task", State: model.Failed}})
			return
		}
		if r.URL.Path == "/api/v1/tasks/owned-task" {
			_ = json.NewEncoder(w).Encode(map[string]any{"task": model.Task{ID: "owned-task", State: model.Failed}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"instance": model.Instance{ID: "owned-instance", State: "RUNNING"}})
	}))
	defer server.Close()
	root := t.TempDir()
	private := filepath.Join(root, "identity")
	if e := os.WriteFile(private, []byte("private fixture data"), 0600); e != nil {
		t.Fatal(e)
	}
	// The original workflow context is already cancelled. Cleanup must use its
	// own bounded context to inspect the real task outcome and resource state.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	f := &fixture{ctx: ctx, root: root, origin: server.URL, http: server.Client(), csrf: "owned-csrf", instance: "owned-instance"}
	if e := f.cleanup(); e == nil {
		t.Fatal("unconfirmed live resource cleanup accepted")
	}
	if f.r.Outcome != "FAILED" {
		t.Fatal("cleanup failure not reflected in report")
	}
	if _, e := os.Stat(private); e != nil {
		t.Fatal("unconfirmed recovery state removed")
	}
}

func TestFailedBackendTaskIsNeverAcceptedAsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(202)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"task": model.Task{ID: "owned-task", State: model.Interrupted}})
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	f := &fixture{ctx: ctx, origin: server.URL, http: server.Client()}
	if _, e := f.task("POST", "/api/v1/owned", map[string]any{}); e == nil {
		t.Fatal("interrupted task counted as success")
	}
}

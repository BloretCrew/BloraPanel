package containers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestPullProgressPreservesLayerCountersAndJournal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/version":
			json.NewEncoder(w).Encode(map[string]string{"ApiVersion": "1.54", "MinAPIVersion": "1.40", "Os": "linux"})
		case "/v1.45/images/create":
			json.NewEncoder(w).Encode(map[string]any{"id": "layer-a", "status": "Downloading", "progressDetail": map[string]int64{"current": 128, "total": 1024}})
			json.NewEncoder(w).Encode(map[string]any{"id": "layer-b", "status": "Extracting", "progressDetail": map[string]int64{"current": 1024, "total": 1024}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	m := saveTestManager(t, t.TempDir())
	m.engine = &engine{client: server.Client(), base: server.URL}
	var notified []Progress
	x := execution{manager: m, actor: "admin", ctx: context.Background(), entry: journal{Result: Result{TaskID: "pull_progress", State: "RUNNING"}}, notify: func(p Progress) error { notified = append(notified, p); return nil }}
	// Deliberately fail final inspection: progress is not a success receipt.
	if err := x.pullImage(context.Background(), "example.invalid/image:one"); err == nil {
		t.Fatal("missing final image must not succeed")
	}
	record, err := m.loadOperation("pull_progress")
	if err != nil {
		t.Fatal(err)
	}
	var observed bool
	for _, p := range record.Result.Progress {
		if p.Layer == "layer-a" {
			observed = true
			if p.Current != 128 || p.Total != 1024 || p.Message != "Downloading" {
				t.Fatal("lost layer measurement", p)
			}
		}
	}
	if !observed || len(notified) < 2 {
		t.Fatal("progress not persisted and notified")
	}
	final := record.Result.Progress[len(record.Result.Progress)-1]
	if final.Layer != "layer-b" || final.Current != 1024 || final.Total != 1024 {
		t.Fatal("last sampled progress lost", final)
	}
	if err := x.emitProgress(Progress{Phase: "pull", Current: -1, Total: 1024}); err != nil {
		t.Fatal(err)
	}
	last := notified[len(notified)-1]
	if last.Current != 0 || last.Total != 0 {
		t.Fatal("invalid measurement was exposed")
	}
}

func TestPullProgressBudgetIsNotSuccessfulEOF(t *testing.T) {
	var inspections atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/version":
			json.NewEncoder(w).Encode(map[string]string{"ApiVersion": "1.54", "MinAPIVersion": "1.40", "Os": "linux"})
		case "/v1.45/images/create":
			// A valid record ends before the exact old limit; a later remote error
			// must not be hidden by a synthetic EOF from a bounded reader.
			record := `{"status":"Downloading","id":"layer-a"}`
			w.Write([]byte(record))
			remaining := (32 << 20) - len(record)
			block := strings.Repeat(" ", 8192)
			for remaining > 0 {
				n := min(remaining, len(block))
				if _, err := w.Write([]byte(block[:n])); err != nil {
					return
				}
				remaining -= n
			}
			w.Write([]byte(`{"error":"late failure"}`))
		default:
			inspections.Add(1)
			json.NewEncoder(w).Encode(map[string]string{"Id": "old-image"})
		}
	}))
	defer server.Close()
	m := saveTestManager(t, t.TempDir())
	m.engine = &engine{client: server.Client(), base: server.URL}
	x := execution{manager: m, actor: "admin", ctx: context.Background(), entry: journal{Result: Result{TaskID: "pull_budget", State: "RUNNING"}}}
	err := x.pullImage(context.Background(), "example.invalid/image:one")
	if err == nil || !strings.Contains(err.Error(), "exceeds 32 MiB") || !x.entry.Result.Unknown || inspections.Load() != 0 {
		t.Fatal("truncated pull was reconciled as success", err, inspections.Load())
	}
}

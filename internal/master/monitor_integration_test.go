package master

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/monitor"
)

func TestMetricsUseLiveNodeAndOwnedProcessScope(t *testing.T) {
	f := newFileFixture(t)
	i := f.instances[0]
	node := i.NodeID
	f.reader.request("GET", "/nodes/"+node+"/metrics", nil, "", 403)
	pointBody := f.admin.request("GET", "/nodes/"+node+"/metrics", nil, "", 200)
	var point monitor.Point
	if err := json.Unmarshal(mustJSON(t, pointBody), &point); err != nil {
		t.Fatal(err)
	}
	if point.ObservedAt.IsZero() || point.MemoryTotal <= 0 || point.DiskTotal <= 0 {
		t.Fatalf("not a real node point: %+v", point)
	}
	// A live native process is the sole process whose RSS may be reported by
	// the instance route; stopped instances cannot be mapped to host PID data.
	i.Config.Command = []string{"/bin/sh", "-c", "while :; do sleep .1; done"}
	i.Config.Escalate = true
	i = settingsResult(t, f.admin.request("PATCH", "/instances/"+i.ID, map[string]any{"revision": i.ConfigRevision, "config": i.Config}, model.ID(), 200))
	task := parseFileTask(t, f.admin.request("POST", "/instances/"+i.ID+"/actions", map[string]string{"action": "start"}, model.ID(), 202))
	f.awaitFile(t, task, model.Succeeded)
	t.Cleanup(func() {
		task := parseFileTask(t, f.admin.request("POST", "/instances/"+i.ID+"/actions", map[string]string{"action": "stop"}, model.ID(), 202))
		f.awaitFile(t, task, model.Succeeded)
	})
	eventually(t, 5*time.Second, func() bool {
		response := f.admin.request("GET", "/instances/"+i.ID+"/metrics", nil, "", 200)
		var p monitor.ProcessPoint
		err := json.Unmarshal(mustJSON(t, response), &p)
		t.Logf("metric point=%+v err=%v", p, err)
		return err == nil && p.RunID != "" && p.PID > 0
	})
	member := f.reader.request("GET", "/instances/"+i.ID+"/metrics", nil, "", 200)
	var visible monitor.ProcessPoint
	if err := json.Unmarshal(mustJSON(t, member), &visible); err != nil || visible.RunID == "" {
		t.Fatal("instance.read metric scope lost", err)
	}
}

func TestInstanceMetricDisconnectAndRunReplacement(t *testing.T) {
	f := newFileFixture(t)
	i := f.instances[0]
	i.Config.Command = []string{"/bin/sh", "-c", "while :; do sleep .1; done"}
	i.Config.Escalate = true
	i = settingsResult(t, f.admin.request("PATCH", "/instances/"+i.ID, map[string]any{"revision": i.ConfigRevision, "config": i.Config}, model.ID(), 200))
	action := func(action string) {
		task := parseFileTask(t, f.admin.request("POST", "/instances/"+i.ID+"/actions", map[string]string{"action": action}, model.ID(), 202))
		f.awaitFile(t, task, model.Succeeded)
	}
	offline := false
	t.Cleanup(func() {
		if offline {
			f.restartNode(0)
		}
		action("kill")
	})
	action("start")
	path := "/instances/" + i.ID + "/metrics"
	var first monitor.ProcessPoint
	if err := json.Unmarshal(mustJSON(t, f.reader.request("GET", path, nil, "", 200)), &first); err != nil || first.RunID == "" || first.Stale {
		t.Fatal("missing live metric", err)
	}
	f.stopNode(0)
	offline = true
	var cached monitor.ProcessPoint
	if err := json.Unmarshal(mustJSON(t, f.reader.request("GET", path, nil, "", 200)), &cached); err != nil || !cached.Stale || cached.RunID != first.RunID || !cached.ObservedAt.Equal(first.ObservedAt) {
		t.Fatal("disconnect changed sample identity", err)
	}
	grant := model.Grant{UserID: f.readerUser.ID, Resource: instanceRef(i), Action: "instance.read"}
	f.admin.request("DELETE", "/grants", grant, model.ID(), 200)
	f.reader.request("GET", path, nil, "", 403)
	f.admin.request("POST", "/grants", grant, model.ID(), 200)
	f.restartNode(0)
	offline = false
	var restored monitor.ProcessPoint
	if err := json.Unmarshal(mustJSON(t, f.reader.request("GET", path, nil, "", 200)), &restored); err != nil || restored.Stale || restored.RunID != first.RunID {
		t.Fatal("reconnect failed to resume live sampling", err)
	}
	action("restart")
	current, err := f.store.Instance(context.Background(), i.ID)
	if err != nil || current.RunID == first.RunID {
		t.Fatal("restart did not replace run", err)
	}
	// No sample exists for the new run. Disconnect must not return the old run.
	f.stopNode(0)
	offline = true
	f.reader.request("GET", path, nil, "", 503)
}

func TestMetricsAggregateOwnedDescendants(t *testing.T) {
	f := newFileFixture(t)
	i := f.instances[0]
	i.Config.Command = []string{"/bin/sh", "-c", "(while :; do :; done) & sleep 30 & wait"}
	i.Config.Escalate = true
	i = settingsResult(t, f.admin.request("PATCH", "/instances/"+i.ID, map[string]any{"revision": i.ConfigRevision, "config": i.Config}, model.ID(), 200))
	action := func(action string) {
		task := parseFileTask(t, f.admin.request("POST", "/instances/"+i.ID+"/actions", map[string]string{"action": action}, model.ID(), 202))
		f.awaitFile(t, task, model.Succeeded)
	}
	t.Cleanup(func() { action("kill") })
	action("start")
	eventually(t, 5*time.Second, func() bool {
		var point monitor.ProcessPoint
		if err := json.Unmarshal(mustJSON(t, f.reader.request("GET", "/instances/"+i.ID+"/metrics", nil, "", 200)), &point); err != nil {
			t.Fatal(err)
		}
		return point.Scope == "process-tree" && point.ProcessCount >= 3 && point.RSS > 0 && point.CPUPercent > 0 && !point.Stale
	})
}

func TestMetricHistoryPersistsAndPrunes(t *testing.T) {
	f := newFileFixture(t)
	node := f.instances[0].NodeID
	base := time.Date(2026, 9, 20, 23, 59, 56, 0, time.UTC)
	for i := 0; i < 125; i++ {
		f.app.cacheMetric(context.Background(), node, monitor.Point{ObservedAt: base.Add(time.Duration(i) * time.Second), MemoryTotal: int64(i + 1)})
	}
	items, err := f.app.cachedMetrics(context.Background(), node)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 120 || items[0].MemoryTotal != 6 || !items[0].Stale {
		t.Fatalf("unexpected persisted history: len=%d first=%+v", len(items), items[0])
	}
	if !items[0].ObservedAt.Equal(time.Date(2026, 9, 21, 0, 0, 1, 0, time.UTC)) || !items[len(items)-1].ObservedAt.Equal(time.Date(2026, 9, 21, 0, 2, 0, 0, time.UTC)) {
		t.Fatalf("cross-day UTC history order: first=%s last=%s", items[0].ObservedAt, items[len(items)-1].ObservedAt)
	}
}

func TestInstanceMetricHistoryIsBoundedAndMarkedStale(t *testing.T) {
	f := newFileFixture(t)
	base := time.Date(2026, 9, 20, 23, 59, 56, 0, time.UTC)
	for n := 0; n < 125; n++ {
		f.app.cacheInstanceMetric(context.Background(), f.instances[0].ID, monitor.ProcessPoint{Point: monitor.Point{ObservedAt: base.Add(time.Duration(n) * time.Second)}, PID: n + 1, RunID: "run"})
	}
	got, err := f.app.cachedInstanceMetric(context.Background(), f.instances[0].ID, "run")
	if err != nil {
		t.Fatal(err)
	}
	if got.PID != 125 || got.RunID != "run" || !got.Stale || !got.ObservedAt.Equal(time.Date(2026, 9, 21, 0, 2, 0, 0, time.UTC)) {
		t.Fatalf("unexpected stale instance point: %+v", got)
	}
	ids, err := f.store.RecordIDs(context.Background(), instanceMetricHistoryNS)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, id := range ids {
		if strings.HasPrefix(id, f.instances[0].ID+"-") {
			count++
		}
	}
	if count != 120 {
		t.Fatalf("instance metric retention=%d", count)
	}
	f.app.cacheInstanceMetric(context.Background(), f.instances[0].ID, monitor.ProcessPoint{Point: monitor.Point{ObservedAt: base.Add(time.Hour)}, RunID: "other-run"})
	got, err = f.app.cachedInstanceMetric(context.Background(), f.instances[0].ID, "run")
	if err != nil || got.RunID != "run" || got.PID != 125 {
		t.Fatal("another run's newer timestamp displaced matching history", err)
	}
	if _, err := f.app.cachedInstanceMetric(context.Background(), f.instances[0].ID, "unsampled-run"); err == nil {
		t.Fatal("unknown run received old cache")
	}
}
func mustJSON(t *testing.T, value map[string]json.RawMessage) []byte {
	t.Helper()
	b, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

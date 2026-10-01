package master

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
)

func TestNodeMaintenancePreservesRuntimeAndSettingsAcrossHeartbeat(t *testing.T) {
	f := newFileFixture(t)
	ctx := context.Background()
	nodeID := f.instances[0].NodeID
	node, _, err := f.store.Node(ctx, nodeID)
	if err != nil {
		t.Fatal(err)
	}
	stale := node
	f.reader.request("PATCH", "/nodes/"+nodeID, map[string]any{"quota": 3, "revision": node.ConfigRevision}, model.ID(), 403)
	result := f.admin.request("POST", "/instances", map[string]any{"nodeId": nodeID, "name": "maintenance-process", "config": model.InstanceConfig{Mode: "native", Command: nativeTestCommand(t, "idle"), Environment: nativeTestEnvironment(), StopSeconds: 1, KillSeconds: 1, Escalate: true}}, model.ID(), 201)
	var instance model.Instance
	if err := json.Unmarshal(result["instance"], &instance); err != nil {
		t.Fatal(err)
	}
	action := func(action string, status int) map[string]json.RawMessage {
		return f.admin.request("POST", "/instances/"+instance.ID+"/actions", map[string]string{"action": action}, model.ID(), status)
	}
	f.awaitFile(t, parseFileTask(t, action("start", 202)), model.Succeeded)
	before, err := f.store.Instance(ctx, instance.ID)
	if err != nil {
		t.Fatal(err)
	}
	key := model.ID()
	patch := map[string]any{"name": "renamed-maintenance-node", "group": "production", "tags": []string{"edge", "中文"}, "maintenance": true, "quota": 2, "revision": node.ConfigRevision}
	f.admin.request("PATCH", "/nodes/"+nodeID, patch, key, 200)
	f.admin.request("PATCH", "/nodes/"+nodeID, patch, key, 200)
	// Replay of a previous connection snapshot must not restore old settings.
	stale.LastSeen = time.Now().UTC()
	if err := f.store.PutNode(ctx, stale); err != nil {
		t.Fatal(err)
	}
	node, _, err = f.store.Node(ctx, nodeID)
	if err != nil {
		t.Fatal(err)
	}
	if !node.Maintenance || node.Name != "renamed-maintenance-node" || node.Group != "production" || len(node.Tags) != 2 || node.Tags[1] != "中文" || node.Quota != 2 || node.ConfigRevision != 2 {
		t.Fatalf("heartbeat overwrote configuration: %+v", node)
	}
	action("restart", 409)
	after, err := f.store.Instance(ctx, instance.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.RunID != before.RunID || after.State != "RUNNING" {
		t.Fatalf("maintenance changed actual running unit: %+v", after)
	}
	f.awaitFile(t, parseFileTask(t, action("stop", 202)), model.Succeeded)
	f.admin.request("PATCH", "/nodes/"+nodeID, map[string]any{"maintenance": false, "revision": 2}, model.ID(), 200)
	f.admin.request("POST", "/instances", map[string]any{"nodeId": nodeID, "name": "quota-excess", "config": instance.Config}, model.ID(), 409)
}

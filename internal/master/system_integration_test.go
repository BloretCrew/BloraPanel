package master

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
)

func TestSystemServiceActionValidationAndPermission(t *testing.T) {
	for _, action := range []string{"system.service.start", "system.service.stop", "system.service.restart"} {
		if !actions[action] {
			t.Fatalf("action %s missing authority mapping", action)
		}
	}
	for _, action := range []string{"system.task.enable", "system.task.disable"} {
		if !actions[action] {
			t.Fatalf("action %s missing authority mapping", action)
		}
	}
	if !actions["system.firewall.apply"] {
		t.Fatal("firewall apply action missing authority mapping")
	}
	f := newFileFixture(t)
	node := f.instances[0].NodeID
	f.admin.request("GET", "/nodes/"+node+"/system/tasks?limit=201", nil, "", 400)
	f.admin.request("GET", "/nodes/"+node+"/system/tasks?after=..%2Fbad", nil, "", 400)
	f.reader.request("POST", "/nodes/"+node+"/system/services/actions", map[string]string{"name": "demo.service", "action": "start"}, "", 403)
	f.admin.request("POST", "/nodes/"+node+"/system/services/actions", map[string]string{"name": "demo.service", "action": "exec"}, "", 400)
	f.admin.request("POST", "/nodes/"+node+"/system/services/actions", map[string]string{"name": "../../passwd", "action": "start"}, "", 400)
	f.reader.request("POST", "/nodes/"+node+"/system/tasks/actions", map[string]string{"name": "demo.timer", "action": "enable"}, "", 403)
	f.admin.request("POST", "/nodes/"+node+"/system/tasks/actions", map[string]string{"name": "demo.timer", "action": "delete"}, "", 400)
	f.admin.request("POST", "/nodes/"+node+"/system/tasks/actions", map[string]string{"name": "../../timer", "action": "enable"}, "", 400)
	f.admin.request("POST", "/nodes/"+node+"/system/firewall/apply", map[string]any{"desired": []string{"80/tcp"}, "planHash": strings.Repeat("a", 64), "confirmUntil": time.Now().UTC().Add(-time.Minute).Format(time.RFC3339)}, "", 400)
	f.admin.request("POST", "/nodes/"+node+"/system/firewall/apply", map[string]any{"desired": []string{"80/tcp"}, "confirmUntil": time.Now().UTC().Add(30 * time.Second).Format(time.RFC3339)}, "", 400)
	f.admin.request("POST", "/nodes/"+node+"/system/firewall/confirm", map[string]string{"taskId": "missing"}, "", 400)
}

func TestFirewallConfirmationReplayAfterTaskTerminal(t *testing.T) {
	f := newFileFixture(t)
	node := f.instances[0].NodeID
	key := model.ID()
	payload, err := json.Marshal(map[string]any{
		"desired":      []string{"443/tcp"},
		"confirmUntil": time.Now().UTC().Add(time.Minute),
		"planHash":     strings.Repeat("a", 64),
	})
	if err != nil {
		t.Fatal(err)
	}
	task, _, err := f.store.Accept(context.Background(), model.Task{
		ActorID:   f.adminUser.ID,
		RequestID: model.ID(),
		Resource:  nodeRef(node),
		Action:    "system.firewall.apply",
		Payload:   payload,
	})
	if err != nil {
		t.Fatal(err)
	}
	task, err = f.store.UpdateTask(context.Background(), task.ID, task.Revision, model.Running, "awaiting_confirmation", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	result, err := json.Marshal(map[string]any{"confirmed": true, "confirmRequestId": key})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.store.UpdateTask(context.Background(), task.ID, task.Revision, model.Succeeded, "confirmed", result, ""); err != nil {
		t.Fatal(err)
	}
	body := map[string]string{"taskId": task.ID}
	replayed := f.admin.request("POST", "/nodes/"+node+"/system/firewall/confirm", body, key, 200)
	var confirmed bool
	if err := json.Unmarshal(replayed["confirmed"], &confirmed); err != nil || !confirmed {
		t.Fatalf("same confirmation key did not replay success: %s %v", replayed["confirmed"], err)
	}
	f.admin.request("POST", "/nodes/"+node+"/system/firewall/confirm", body, model.ID(), 409)
}

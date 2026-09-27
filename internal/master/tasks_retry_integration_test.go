package master

import (
	"encoding/json"
	"strings"
	"testing"

	"blora.dev/panel/internal/model"
)

func TestTaskRetryCreatesRelatedInstanceAttemptAndIsIdempotent(t *testing.T) {
	f := newFileFixture(t)
	instance := f.instances[0]

	// The fixture's command is intentionally unavailable, so the real daemon
	// accepts the lifecycle task and records a durable failure.
	failed := parseFileTask(t, f.admin.request("POST", "/instances/"+instance.ID+"/actions", map[string]string{"action": "start"}, model.ID(), 202))
	failed = f.awaitFile(t, failed, model.Failed)
	f.admin.request("POST", "/tasks/"+failed.ID+"/retry", map[string]any{}, " ", 400)
	f.admin.request("POST", "/tasks/"+failed.ID+"/retry", map[string]any{}, strings.Repeat("x", 129), 400)

	requestID := model.ID()
	retried := parseFileTask(t, f.admin.request("POST", "/tasks/"+failed.ID+"/retry", map[string]any{}, requestID, 202))
	if retried.ID == failed.ID || retried.RetryOf != failed.ID || retried.Action != failed.Action || retried.Resource != failed.Resource {
		t.Fatalf("retry lost relationship: original=%+v retry=%+v", failed, retried)
	}
	repeated := parseFileTask(t, f.admin.request("POST", "/tasks/"+failed.ID+"/retry", map[string]any{}, requestID, 200))
	if repeated.ID != retried.ID || repeated.RetryOf != failed.ID {
		t.Fatalf("retry idempotency returned a different task: first=%+v repeated=%+v", retried, repeated)
	}
	page := f.admin.request("GET", "/tasks?offset=0&limit=1", nil, "", 200)
	var pageItems []model.Task
	if err := json.Unmarshal(page["items"], &pageItems); err != nil || len(pageItems) != 1 || string(page["nextOffset"]) != "1" {
		t.Fatalf("task pagination is not bounded: items=%s next=%s err=%v", page["items"], page["nextOffset"], err)
	}

	// A reader cannot see an administrator-owned task, so the retry endpoint
	// preserves the same not-found authorization boundary as GET /tasks/{id}.
	f.reader.request("POST", "/tasks/"+failed.ID+"/retry", map[string]any{}, model.ID(), 404)
	f.awaitFile(t, retried, model.Failed)
}

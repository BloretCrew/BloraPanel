package master

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"blora.dev/panel/internal/model"
)

func TestCloudWorkspaceSnapshotsOwnerRevisionAndContentPolicy(t *testing.T) {
	f := newFileFixture(t)
	state := map[string]any{
		"userId":      f.adminUser.ID,
		"workspaceId": "cloud-1",
		"windows":     []any{map[string]any{"appId": "overview", "windowId": "w1"}},
		"views":       []any{map[string]any{"viewTabId": "v1", "resourceRef": map[string]any{"kind": "node", "id": "n1"}}},
		"drafts":      map[string]any{"file": "secret draft"},
		"terminals":   map[string]any{"v1": "screen"},
		"uploads":     map[string]any{"u1": "handle"},
		"preferences": map[string]any{"notifications": []any{map[string]string{"message": "private notification"}}, "reduceMotion": true},
	}
	firstKey := model.ID()
	put := f.admin.request("PUT", "/workspaces/cloud-1", map[string]any{
		"title": "Layout copy", "deviceId": "device-a", "baseRevision": 0,
		"schemaVersion": 1, "state": state,
	}, firstKey, 200)
	var meta struct {
		Revision       int64 `json:"revision"`
		IncludeContent bool  `json:"includeContent"`
	}
	if err := json.Unmarshal(put["metadata"], &meta); err != nil {
		t.Fatal(err)
	}
	if meta.Revision != 1 || meta.IncludeContent {
		t.Fatalf("unexpected first metadata: %+v", meta)
	}
	replayed := f.admin.request("PUT", "/workspaces/cloud-1", map[string]any{
		"title": "Layout copy", "deviceId": "device-a", "baseRevision": 0,
		"schemaVersion": 1, "state": state,
	}, firstKey, 200)
	var replayMeta struct {
		Revision int64 `json:"revision"`
	}
	if err := json.Unmarshal(replayed["metadata"], &replayMeta); err != nil || replayMeta.Revision != 1 {
		t.Fatalf("workspace request replay changed result: %+v (%v)", replayMeta, err)
	}
	got := f.admin.request("GET", "/workspaces/cloud-1", nil, "", 200)
	var restored map[string]json.RawMessage
	if err := json.Unmarshal(got["state"], &restored); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"drafts", "terminals", "uploads"} {
		if string(restored[key]) != "{}" {
			t.Fatalf("default sync retained %s: %s", key, restored[key])
		}
	}
	if len(restored["windows"]) == 0 || len(restored["views"]) == 0 {
		t.Fatal("layout and resource references were not preserved")
	}
	var preferences map[string]json.RawMessage
	if err := json.Unmarshal(restored["preferences"], &preferences); err != nil || preferences["notifications"] != nil || string(preferences["reduceMotion"]) != "true" {
		t.Fatal("layout sync must remove notification content and retain preferences")
	}
	list := f.admin.request("GET", "/workspaces", nil, "", 200)
	var items []workspaceMetadata
	if err := json.Unmarshal(list["items"], &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].WorkspaceID != "cloud-1" {
		t.Fatalf("workspace list=%+v", items)
	}
	readerList := f.reader.request("GET", "/workspaces", nil, "", 200)
	var readerItems []workspaceMetadata
	if err := json.Unmarshal(readerList["items"], &readerItems); err != nil {
		t.Fatal(err)
	}
	if len(readerItems) != 0 {
		t.Fatalf("reader saw another user's workspace: %+v", readerItems)
	}
	f.reader.request("GET", "/workspaces/cloud-1", nil, "", 404)
	f.admin.request("PUT", "/workspaces/cloud-1", map[string]any{
		"title": "stale", "deviceId": "device-b", "baseRevision": 0,
		"schemaVersion": 1, "state": state,
	}, model.ID(), 409)
	state["drafts"] = map[string]any{"file": "kept by explicit content sync"}
	second := f.admin.request("PUT", "/workspaces/cloud-1", map[string]any{
		"title": "Content copy", "deviceId": "device-a", "baseRevision": 1,
		"schemaVersion": 1, "includeContent": true, "state": state,
	}, model.ID(), 200)
	if err := json.Unmarshal(second["metadata"], &meta); err != nil {
		t.Fatal(err)
	}
	if meta.Revision != 2 || !meta.IncludeContent {
		t.Fatalf("unexpected content metadata: %+v", meta)
	}
	got = f.admin.request("GET", "/workspaces/cloud-1", nil, "", 200)
	if err := json.Unmarshal(got["state"], &restored); err != nil {
		t.Fatal(err)
	}
	var drafts map[string]string
	if err := json.Unmarshal(restored["preferences"], &preferences); err != nil || !bytes.Contains(preferences["notifications"], []byte("private notification")) {
		t.Fatal("explicit content sync lost notifications")
	}
	if err := json.Unmarshal(restored["drafts"], &drafts); err != nil || drafts["file"] != "kept by explicit content sync" {
		t.Fatalf("explicit content was not retained: %s (%v)", restored["drafts"], err)
	}
	deleteWorkspace := func(c *testClient, id string, status int) {
		t.Helper()
		req, err := http.NewRequest(http.MethodDelete, c.base+"/api/v1/workspaces/"+id, bytes.NewReader(nil))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("X-CSRF-Token", c.csrf)
		req.Header.Set("Idempotency-Key", model.ID())
		resp, err := c.client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != status {
			t.Fatalf("DELETE %s status=%d want=%d body=%s", id, resp.StatusCode, status, body)
		}
	}
	deleteWorkspace(f.admin, "cloud-1", http.StatusNoContent)
	f.admin.request("GET", "/workspaces/cloud-1", nil, "", 404)
	deleteWorkspace(f.admin, "cloud-1", http.StatusNoContent)
}

func TestCloudWorkspaceRejectsInvalidOwnerAndID(t *testing.T) {
	f := newFileFixture(t)
	f.admin.request("PUT", "/workspaces/missing-key", map[string]any{}, "", 400)
	f.admin.request("PUT", "/workspaces/missing-owner", map[string]any{
		"title": "bad", "deviceId": "device", "baseRevision": 0, "schemaVersion": 1,
		"state": map[string]any{"workspaceId": "missing-owner"},
	}, model.ID(), 400)
	f.admin.request("PUT", "/workspaces/bad%2Fid", map[string]any{
		"title": "bad", "deviceId": "device", "baseRevision": 0, "schemaVersion": 1,
		"state": map[string]any{"userId": f.readerUser.ID},
	}, model.ID(), 400)
	f.admin.request("PUT", "/workspaces/owner-mismatch", map[string]any{
		"title": "bad", "deviceId": "device", "baseRevision": 0, "schemaVersion": 1,
		"state": map[string]any{"userId": f.readerUser.ID},
	}, model.ID(), 403)
	f.admin.request("PUT", "/workspaces/path-mismatch", map[string]any{
		"title": "bad", "deviceId": "device", "baseRevision": 0, "schemaVersion": 1,
		"state": map[string]any{"userId": f.adminUser.ID, "workspaceId": "other"},
	}, model.ID(), 400)
}

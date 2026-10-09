package master

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/coreupdate"
	"blora.dev/panel/internal/model"
)

func TestCoreUpdatesRequireAdministratorAndConfiguredService(t *testing.T) {
	f := newFileFixture(t)
	for _, route := range []struct{ method, path string }{
		{"GET", "/system/updates"},
		{"POST", "/system/updates/check"},
		{"POST", "/system/updates/apply"},
		{"PUT", "/system/updates/source"},
		{"GET", "/nodes/unknown/updates/status"},
		{"POST", "/nodes/unknown/updates/check"},
		{"POST", "/nodes/unknown/updates/apply"},
		{"PUT", "/nodes/unknown/updates/source"},
	} {
		denied := f.reader.request(route.method, route.path, map[string]string{}, model.ID(), http.StatusForbidden)
		if !strings.Contains(string(denied["error"]), "FORBIDDEN") {
			t.Fatalf("permission result: %s", denied["error"])
		}
	}
	for _, path := range []string{"/system/updates", "/system/updates/check", "/system/updates/apply", "/system/updates/source"} {
		method := "POST"
		if path == "/system/updates" {
			method = "GET"
		}
		if strings.HasSuffix(path, "/source") {
			method = "PUT"
		}
		result := f.admin.request(method, path, map[string]string{}, model.ID(), http.StatusServiceUnavailable)
		if !strings.Contains(string(result["error"]), "UPDATES_UNAVAILABLE") {
			t.Fatalf("disabled updater result: %s", result["error"])
		}
	}
	f.admin.request("POST", "/nodes/unknown/updates/apply", map[string]string{"revision": "anything"}, model.ID(), http.StatusNotFound)
}

func TestCoreUpdateActivationRechecksAdministrator(t *testing.T) {
	f := newFileFixture(t)
	ctx := context.Background()
	requestID := f.adminUser.ID + ":accepted-update"
	if err := f.app.ValidateCoreUpdateActor(ctx, requestID); err != nil {
		t.Fatal(err)
	}
	nodeID := f.instances[0].NodeID
	ref := nodeRef(nodeID)
	f.admin.request("POST", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: ref, Action: "host.manage"}, model.ID(), http.StatusOK)
	check := func(actor string) bool {
		t.Helper()
		result, err := f.app.nodeAuthority(ctx, nodeID, bridge.Request{Method: "authority.check", ActorID: actor, Resource: ref, Args: []byte(`{"action":"core.update"}`)})
		if err != nil {
			t.Fatal(err)
		}
		return result.(map[string]bool)["allowed"]
	}
	if check(f.readerUser.ID) {
		t.Fatal("host.manage grant must not grant core update authority")
	}
	if !check(f.adminUser.ID) {
		t.Fatal("administrator update authority denied")
	}
	if _, err := f.store.DB.ExecContext(ctx, "UPDATE users SET admin=0 WHERE id=?", f.adminUser.ID); err != nil {
		t.Fatal(err)
	}
	if err := f.app.ValidateCoreUpdateActor(ctx, requestID); err == nil {
		t.Fatal("revoked administrator can still activate accepted update")
	}
	if check(f.adminUser.ID) {
		t.Fatal("daemon activation authority retained revoked administrator role")
	}
	if err := f.app.ValidateCoreUpdateActor(ctx, "unscoped-request"); err == nil {
		t.Fatal("unscoped update identity accepted")
	}
}

func TestCoreUpdateCompatibilityIncludesOfflineAndLegacyNodes(t *testing.T) {
	f := newFileFixture(t)
	ctx := context.Background()
	nodes, err := f.store.Nodes(ctx)
	if err != nil || len(nodes) == 0 {
		t.Fatal("no real enrolled nodes", err)
	}
	request := httptest.NewRequest("POST", "/", nil)
	compatible := coreupdate.Preview{Manifest: coreupdate.Manifest{PeerProtocolMin: 1, PeerProtocolMax: 1}}
	if reason := f.app.coreUpdateFleetReason(request, compatible); reason != "" {
		t.Fatal(reason)
	}
	f.stopNode(0)
	node, _, err := f.store.Node(ctx, f.instances[0].NodeID)
	if err != nil {
		t.Fatal(err)
	}
	node.Capabilities["core.protocol"] = "2"
	data, _ := json.Marshal(node)
	if _, err := f.store.DB.ExecContext(ctx, "UPDATE nodes SET document=? WHERE id=?", data, node.ID); err != nil {
		t.Fatal(err)
	}
	if reason := f.app.coreUpdateFleetReason(request, compatible); !strings.Contains(reason, "不兼容") {
		t.Fatalf("offline incompatible node not blocked: %q", reason)
	}
	node.Capabilities["core.protocol"] = "invalid"
	data, _ = json.Marshal(node)
	if _, err := f.store.DB.ExecContext(ctx, "UPDATE nodes SET document=? WHERE id=?", data, node.ID); err != nil {
		t.Fatal(err)
	}
	if reason := f.app.coreUpdateFleetReason(request, compatible); !strings.Contains(reason, "无效") {
		t.Fatalf("invalid protocol not blocked: %q", reason)
	}
}

func TestCoreUpdateReplaySurvivesMissingPreviewAndSourceChange(t *testing.T) {
	f := newFileFixture(t)
	root := t.TempDir()
	revision := strings.Repeat("a", 40)
	key := "same-accepted-update"
	old := coreupdate.Job{ID: f.adminUser.ID + ":" + key, Revision: revision, State: "failed", Phase: "failed", Detail: "Recorded attempt"}
	data, err := json.Marshal(old)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, "job.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	manager, err := coreupdate.New(coreupdate.Options{Root: root, Component: "master", CurrentRevision: strings.Repeat("b", 40)})
	if err != nil {
		t.Fatal(err)
	}
	f.app.coreUpdater = manager
	if manager.Status().Preview != nil {
		t.Fatal("fixture must have no preview")
	}
	replayed := f.admin.request("POST", "/system/updates/apply", map[string]string{"revision": revision}, key, http.StatusAccepted)
	var result coreupdate.Job
	data, _ = json.Marshal(replayed)
	if err := json.Unmarshal(data, &result); err != nil || result.ID != old.ID || result.State != old.State {
		t.Fatalf("replayed result changed: %+v (%v)", result, err)
	}
	f.admin.request("POST", "/system/updates/apply", map[string]string{"revision": strings.Repeat("c", 40)}, key, http.StatusConflict)
	f.admin.request("POST", "/system/updates/apply", map[string]string{"revision": revision}, "", http.StatusBadRequest)
	f.admin.request("PUT", "/system/updates/source", map[string]string{"repository": "http://insecure.example/repo", "channel": "beta"}, model.ID(), http.StatusConflict)
	f.admin.request("PUT", "/system/updates/source", map[string]string{"repository": coreupdate.DefaultRepository, "channel": "stable", "apiUrl": "https://example.com/repos/BloretCrew/BloraPanel"}, model.ID(), http.StatusOK)
	f.admin.request("POST", "/system/updates/apply", map[string]string{"revision": revision}, key, http.StatusAccepted)
	if got := manager.Status().Source.APIURL; got != "https://example.com/repos/BloretCrew/BloraPanel" {
		t.Fatalf("source mirror not persisted: %s", got)
	}
}

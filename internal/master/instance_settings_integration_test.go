package master

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"blora.dev/panel/internal/filesystem"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
)

func settingsResult(t *testing.T, result map[string]json.RawMessage) model.Instance {
	t.Helper()
	var i model.Instance
	if err := json.Unmarshal(result["instance"], &i); err != nil {
		t.Fatal(err)
	}
	return i
}

func TestInstanceSettingsAreVersionedAndRootsRemainSeparate(t *testing.T) {
	f := newFileFixture(t)
	ctx := context.Background()
	i := f.instances[0]
	f.awaitFile(t, f.save(t, 0, "marker.txt", "original-root", filesystem.MissingVersion, model.ID()), model.Succeeded)
	url := "/instances/" + i.ID
	f.reader.request("PATCH", url, map[string]any{"revision": i.ConfigRevision, "name": "denied"}, model.ID(), 403)
	key := model.ID()
	body := map[string]any{"revision": i.ConfigRevision, "name": "中文-renamed", "group": "games", "tags": []string{"one", "two"}}
	i = settingsResult(t, f.admin.request("PATCH", url, body, key, 200))
	if again := settingsResult(t, f.admin.request("PATCH", url, body, key, 200)); again.ConfigRevision != i.ConfigRevision {
		t.Fatal("retry applied settings twice")
	}
	f.admin.request("PATCH", url, map[string]any{"revision": i.ConfigRevision - 1, "name": "stale"}, model.ID(), 409)
	filtered := f.admin.request("GET", "/instances?group=games&tag=two", nil, "", 200)
	var found []model.Instance
	if err := json.Unmarshal(filtered["items"], &found); err != nil || len(found) != 1 || found[0].ID != i.ID {
		t.Fatal("group/tag filter mismatch")
	}
	oldConfig := i.Config
	newRoot := filepath.Join(t.TempDir(), "new-root")
	if err := os.Mkdir(newRoot, 0750); err != nil {
		t.Fatal(err)
	}
	i.Config.Directory = newRoot
	i = settingsResult(t, f.admin.request("PATCH", url, map[string]any{"revision": i.ConfigRevision, "config": i.Config}, model.ID(), 200))
	f.admin.request("GET", fileURL(i)+"/stat?path=marker.txt", nil, "", 404)
	f.awaitFile(t, f.save(t, 0, "marker.txt", "new-root", filesystem.MissingVersion, model.ID()), model.Succeeded)
	assertDisk(t, f.roots[0], "marker.txt", "original-root")
	assertDisk(t, newRoot, "marker.txt", "new-root")
	// A delayed fresh task cannot use the configuration it read before PATCH.
	stale, _ := json.Marshal(oldConfig)
	if _, _, err := f.store.Accept(ctx, model.Task{ActorID: f.adminUser.ID, RequestID: model.ID(), Resource: instanceRef(i), Action: "instance.start", Payload: stale}); !errors.Is(err, storage.ErrConflict) {
		t.Fatalf("stale launch admitted: %v", err)
	}
	i = settingsResult(t, f.admin.request("PATCH", url, map[string]any{"revision": i.ConfigRevision, "config": oldConfig}, model.ID(), 200))
	if e := f.stat(t, 0, "marker.txt"); e.Size != int64(len("original-root")) {
		t.Fatal("returning to original root reused another root")
	}
	// A real running process keeps its launch configuration; changing the
	// graceful-stop command is effective for the next explicit stop request.
	i.Config.Command = []string{"/bin/sh", "-c", "while IFS= read -r line; do if [ \"$line\" = quit ]; then printf stopped > stop-policy; exit 0; fi; done"}
	i = settingsResult(t, f.admin.request("PATCH", url, map[string]any{"revision": i.ConfigRevision, "config": i.Config}, model.ID(), 200))
	control := func(action string) model.Task {
		return parseFileTask(t, f.admin.request("POST", url+"/actions", map[string]string{"action": action}, model.ID(), 202))
	}
	f.awaitFile(t, control("start"), model.Succeeded)
	t.Cleanup(func() { f.awaitFile(t, control("stop"), model.Succeeded) })
	bad := i.Config
	bad.Command = []string{"/bin/false"}
	f.admin.request("PATCH", url, map[string]any{"revision": i.ConfigRevision, "config": bad}, model.ID(), 409)
	i.Config.StopInput = "quit\n"
	i = settingsResult(t, f.admin.request("PATCH", url, map[string]any{"revision": i.ConfigRevision, "config": i.Config}, model.ID(), 200))
	f.awaitFile(t, control("stop"), model.Succeeded)
	assertDisk(t, f.roots[0], "stop-policy", "stopped")
	current, err := f.store.Instance(ctx, i.ID)
	if err != nil || current.ConfigRevision != i.ConfigRevision || current.Name != "中文-renamed" {
		t.Fatal("node observations overwrote Master settings")
	}
}

func TestAutostartRunsOncePerDaemonStartupAndRechecksOwner(t *testing.T) {
	f := newFileFixture(t)
	ctx := context.Background()
	i := f.instances[0]
	for _, action := range []string{"instance.configure", "instance.start"} {
		f.admin.request("POST", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: instanceRef(i), Action: action}, model.ID(), 200)
	}
	f.admin.request("POST", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: model.ResourceRef{Kind: "node", ID: i.NodeID}, Action: "host.manage"}, model.ID(), 200)
	i.Config.Command = []string{"/bin/sh", "-c", "trap 'exit 0' TERM; while :; do sleep .1; done"}
	i.Config.Escalate = true
	i.Config.Autostart = true
	i = settingsResult(t, f.reader.request("PATCH", "/instances/"+i.ID, map[string]any{"revision": i.ConfigRevision, "config": i.Config}, model.ID(), 200))
	n, _, err := f.store.Node(ctx, i.NodeID)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.app.enqueueAutostart(ctx, n); err != nil {
		t.Fatal(err)
	}
	pending, err := f.store.Pending(ctx)
	if err != nil || len(pending) != 0 {
		t.Fatal("saving autostart started a process immediately")
	}
	control := func(action string) model.Task {
		return parseFileTask(t, f.admin.request("POST", "/instances/"+i.ID+"/actions", map[string]string{"action": action}, model.ID(), 202))
	}
	t.Cleanup(func() { f.awaitFile(t, control("stop"), model.Succeeded) })
	f.restartNode(0)
	eventually(t, 8*time.Second, func() bool {
		current, err := f.store.Instance(ctx, i.ID)
		return err == nil && current.State == "RUNNING"
	})
	n, _, err = f.store.Node(ctx, i.NodeID)
	if err != nil {
		t.Fatal(err)
	}
	key := "autostart-" + storage.Hash([]byte(i.ID+":"+n.StartupID))
	task, err := f.store.TaskByRequest(ctx, f.readerUser.ID, key)
	if err != nil {
		t.Fatal(err)
	}
	f.awaitFile(t, task, model.Succeeded)
	f.awaitFile(t, control("stop"), model.Succeeded)
	f.app.disconnectNode(i.NodeID)
	eventually(t, 8*time.Second, func() bool {
		now, _, err := f.store.Node(ctx, i.NodeID)
		return err == nil && now.Generation > n.Generation
	})
	if err := f.app.enqueueAutostart(ctx, n); err != nil {
		t.Fatal(err)
	}
	current, err := f.store.Instance(ctx, i.ID)
	if err != nil || current.State != "STOPPED" {
		t.Fatal("management reconnect restarted manually stopped instance")
	}
	f.admin.request("DELETE", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: instanceRef(i), Action: "instance.start"}, model.ID(), 200)
	f.restartNode(0)
	eventually(t, 8*time.Second, func() bool {
		now, _, err := f.store.Node(ctx, i.NodeID)
		if err != nil {
			return false
		}
		var decision struct {
			Reason string `json:"reason"`
		}
		_, err = f.store.Record(ctx, "autostart", i.ID+":"+now.StartupID, &decision)
		return err == nil && decision.Reason == "authorization_revoked"
	})
	current, err = f.store.Instance(ctx, i.ID)
	if err != nil || current.State != "STOPPED" {
		t.Fatal("revoked owner kept automatic startup authority")
	}
}

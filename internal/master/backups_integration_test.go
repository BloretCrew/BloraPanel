package master

import (
	"encoding/json"
	"testing"

	"blora.dev/panel/internal/backup"
	"blora.dev/panel/internal/filesystem"
	"blora.dev/panel/internal/model"
)

func TestBackupAPIRealArchiveConditionalRestoreAndRevocation(t *testing.T) {
	f := newFileFixture(t)
	i := f.instances[0]
	base := "/instances/" + i.ID
	f.awaitFile(t, f.save(t, 0, "中文.txt", "备份原文\n", filesystem.MissingVersion, model.ID()), model.Succeeded)
	version := f.stat(t, 0, "中文.txt").Version
	body := map[string]any{"path": "中文.txt", "version": version, "compression": "deflate", "consistency": map[string]string{"mode": "files"}}
	f.reader.request("POST", base+"/backups", body, model.ID(), 403)
	f.admin.request("POST", base+"/backups", body, "", 400)
	key := model.ID()
	task := parseFileTask(t, f.admin.request("POST", base+"/backups", body, key, 202))
	task = f.awaitFile(t, task, model.Succeeded)
	var created backup.Result
	if err := json.Unmarshal(task.Result, &created); err != nil || created.Snapshot == nil {
		t.Fatal("backup receipt missing", err)
	}
	repeated := parseFileTask(t, f.admin.request("POST", base+"/backups", body, key, 202))
	if repeated.ID != task.ID {
		t.Fatal("repeated backup created another task")
	}
	f.restartNode(0)
	listed := f.admin.request("GET", base+"/backups", nil, "", 200)
	var items []backup.SnapshotInfo
	if err := json.Unmarshal(listed["items"], &items); err != nil || len(items) != 1 || items[0].ID != created.ID {
		t.Fatal("archive lost after daemon restart", err)
	}
	f.awaitFile(t, f.save(t, 0, "中文.txt", "后来修改", version, model.ID()), model.Succeeded)
	changed := f.stat(t, 0, "中文.txt").Version
	// A different authorized actor may restore an accessible archive; the
	// operation must retain that actor rather than impersonate its creator.
	for _, action := range []string{"backup.restore", "file.write"} {
		f.admin.request("POST", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: instanceRef(i), Action: action}, model.ID(), 200)
	}
	result := f.reader.request("POST", base+"/restore-plans", map[string]any{"backupId": created.ID, "path": "中文.txt", "version": changed}, model.ID(), 201)
	var plan backup.RestorePlan
	if err := json.Unmarshal(result["plan"], &plan); err != nil || plan.Request.OwnerID != f.readerUser.ID {
		t.Fatal("restore plan actor mismatch", err)
	}
	denied := parseFileTask(t, f.reader.request("POST", base+"/restores", map[string]any{"planId": plan.Request.ID, "planHash": plan.Hash}, model.ID(), 202))
	f.awaitFile(t, denied, model.Failed)
	assertDisk(t, f.roots[0], "中文.txt", "后来修改")
	// Permission disappears after the plan is prepared. A queued restore
	// cannot turn the durable plan into perpetual authority over its source.
	f.admin.request("DELETE", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: instanceRef(i), Action: "file.read"}, model.ID(), 200)
	revoked := parseFileTask(t, f.reader.request("POST", base+"/restores", map[string]any{"planId": plan.Request.ID, "planHash": plan.Hash, "overwritePlanHash": plan.Hash}, model.ID(), 202))
	f.awaitFile(t, revoked, model.Failed)
	assertDisk(t, f.roots[0], "中文.txt", "后来修改")
	f.admin.request("POST", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: instanceRef(i), Action: "file.read"}, model.ID(), 200)
	restored := parseFileTask(t, f.reader.request("POST", base+"/restores", map[string]any{"planId": plan.Request.ID, "planHash": plan.Hash, "overwritePlanHash": plan.Hash}, model.ID(), 202))
	f.awaitFile(t, restored, model.Succeeded)
	assertDisk(t, f.roots[0], "中文.txt", "备份原文\n")
	// The old target version no longer authorizes a second overwrite.
	stale := parseFileTask(t, f.reader.request("POST", base+"/restores", map[string]any{"planId": plan.Request.ID, "planHash": plan.Hash, "overwritePlanHash": plan.Hash}, model.ID(), 202))
	f.awaitFile(t, stale, model.Failed)
}

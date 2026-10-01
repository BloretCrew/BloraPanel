package master

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"blora.dev/panel/internal/backup"
	"blora.dev/panel/internal/model"
)

func scheduleFromResponse(t *testing.T, result map[string]json.RawMessage) backup.Schedule {
	t.Helper()
	var item backup.Schedule
	if err := json.Unmarshal(result["schedule"], &item); err != nil {
		t.Fatal(err)
	}
	return item
}
func apiSchedule(i model.Instance, action string, args any) scheduleInput {
	b, _ := json.Marshal(args)
	return scheduleInput{Resource: instanceRef(i), Action: action, Args: b, Cron: "* * * * *", Timezone: "Asia/Shanghai", Misfire: "run_once", Overlap: "wait_one", Enabled: true}
}
func TestSchedulesAPIControlsOwnershipRevocationAndCAS(t *testing.T) {
	f := newFileFixture(t)
	f.app.closeSchedules()
	ctx := context.Background()
	in := apiSchedule(f.instances[0], "instance.stop", struct{}{})
	f.reader.request("POST", "/schedules", in, model.ID(), 403)
	f.admin.request("POST", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: in.Resource, Action: "schedule.create"}, model.ID(), 200)
	f.reader.request("POST", "/schedules", in, model.ID(), 403)
	f.admin.request("POST", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: in.Resource, Action: "instance.stop"}, model.ID(), 200)
	f.reader.request("POST", "/schedules", in, "   ", 400)
	key := model.ID()
	item := scheduleFromResponse(t, f.reader.request("POST", "/schedules", in, key, 201))
	repeat := scheduleFromResponse(t, f.reader.request("POST", "/schedules", in, key, 200))
	if repeat.Spec.ID != item.Spec.ID || !repeat.NextRun.Equal(item.NextRun) || repeat.ConfigRevision != 1 {
		t.Fatal("creation response loss resets schedule")
	}
	adminOnly := scheduleFromResponse(t, f.admin.request("POST", "/schedules", apiSchedule(f.instances[1], "instance.stop", struct{}{}), model.ID(), 201))
	listed := f.reader.request("GET", "/schedules", nil, "", 200)
	var items []backup.Schedule
	if err := json.Unmarshal(listed["items"], &items); err != nil || len(items) != 1 || items[0].Spec.ID != item.Spec.ID {
		t.Fatalf("owner list leaked %v %v", items, err)
	}
	f.reader.request("DELETE", "/schedules/"+adminOnly.Spec.ID, map[string]int{"revision": 1}, model.ID(), 404)
	f.admin.request("DELETE", "/schedules/"+adminOnly.Spec.ID, map[string]int{"revision": 1}, "", 400)
	f.admin.request("DELETE", "/schedules/"+adminOnly.Spec.ID, map[string]int{"revision": 1}, model.ID(), 200)
	in.Revision = 0
	f.reader.request("PUT", "/schedules/"+item.Spec.ID, in, model.ID(), 409)
	f.admin.request("DELETE", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: in.Resource, Action: "instance.stop"}, model.ID(), 200)
	if fires, err := f.app.schedules.engine.Tick(ctx, item.NextRun); err != nil || len(fires) != 0 {
		t.Fatalf("revoked schedule fired %v %v", fires, err)
	}
	updated, err := f.app.schedules.engine.Get(ctx, item.Spec.ID)
	if err != nil || updated.LastError != "authorization_denied" || updated.Skipped != 1 {
		t.Fatalf("revocation fact %+v %v", updated, err)
	}
	f.admin.request("DELETE", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: in.Resource, Action: "schedule.create"}, model.ID(), 200)
	in.Revision = 1
	in.Enabled = false
	disableKey := model.ID()
	disabled := scheduleFromResponse(t, f.reader.request("PUT", "/schedules/"+item.Spec.ID, in, disableKey, 200))
	replayedDisable := scheduleFromResponse(t, f.reader.request("PUT", "/schedules/"+item.Spec.ID, in, disableKey, 200))
	if replayedDisable.ConfigRevision != disabled.ConfigRevision || !reflect.DeepEqual(replayedDisable.Spec, disabled.Spec) {
		t.Fatalf("schedule update replay changed result: first=%+v replay=%+v", disabled, replayedDisable)
	}
	changed := in
	changed.Cron = "5 * * * *"
	f.reader.request("PUT", "/schedules/"+item.Spec.ID, changed, disableKey, 409)
	if disabled.Spec.Enabled {
		t.Fatal("owner cannot disable after revocation")
	}
	deleteKey := model.ID()
	f.reader.request("DELETE", "/schedules/"+item.Spec.ID, map[string]int64{"revision": disabled.ConfigRevision}, deleteKey, 200)
	f.reader.request("DELETE", "/schedules/"+item.Spec.ID, map[string]int64{"revision": disabled.ConfigRevision}, deleteKey, 200)
	f.reader.request("DELETE", "/schedules/"+item.Spec.ID, map[string]int64{"revision": disabled.ConfigRevision + 1}, deleteKey, 409)
}

func TestSchedulesRealNodeOfflineCoalescesAndRechecksPermission(t *testing.T) {
	for _, revoke := range []bool{false, true} {
		t.Run(map[bool]string{false: "recover_once", true: "revoked_while_offline"}[revoke], func(t *testing.T) {
			f := newFileFixture(t)
			f.app.closeSchedules()
			ctx := context.Background()
			i := f.instances[0]
			transferWrite(t, f.roots[0], "scheduled-source", []byte("before disconnect"))
			for _, action := range []string{"schedule.create", "backup.create"} {
				f.admin.request("POST", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: instanceRef(i), Action: action}, model.ID(), 200)
			}
			item := scheduleFromResponse(t, f.reader.request("POST", "/schedules", apiSchedule(i, "backup.create", scheduledBackup{Path: "scheduled-source", Compression: "store", Consistency: backup.Consistency{Mode: "files"}}), model.ID(), 201))
			f.stopNode(0)
			for minute := 0; minute < 5; minute++ {
				fires, err := f.app.schedules.engine.Tick(ctx, item.NextRun.Add(time.Duration(minute)*time.Minute))
				if err != nil || len(fires) != 0 {
					t.Fatalf("offline schedule accepted work: %+v %v", fires, err)
				}
			}
			pending, err := f.app.schedules.engine.Get(ctx, item.Spec.ID)
			if err != nil || pending.Pending.IsZero() {
				t.Fatalf("coalesced pending occurrence missing: %+v %v", pending, err)
			}
			if tasks, err := f.store.Tasks(ctx); err != nil || len(tasks) != 0 {
				t.Fatalf("offline node accumulated tasks: %+v %v", tasks, err)
			}
			transferWrite(t, f.roots[0], "scheduled-source", []byte("fresh after offline period"))
			if revoke {
				f.admin.request("DELETE", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: instanceRef(i), Action: "backup.create"}, model.ID(), 200)
			}
			f.restartNode(0)
			at := item.NextRun.Add(4*time.Minute + time.Second)
			fires, err := f.app.schedules.engine.Tick(ctx, at)
			if err != nil {
				t.Fatal(err)
			}
			if revoke {
				if len(fires) != 0 {
					t.Fatalf("revoked offline schedule fired: %+v", fires)
				}
				current, err := f.app.schedules.engine.Get(ctx, item.Spec.ID)
				if err != nil || current.LastError != "authorization_denied" {
					t.Fatalf("missing revocation diagnostic: %+v %v", current, err)
				}
				if tasks, err := f.store.Tasks(ctx); err != nil || len(tasks) != 0 {
					t.Fatalf("revoked occurrence left an accepted task: %+v %v", tasks, err)
				}
				return
			}
			if len(fires) != 1 {
				t.Fatalf("expected one coalesced occurrence, got %+v", fires)
			}
			task := f.awaitFile(t, fires[0].Task, model.Succeeded)
			var payload backup.TaskPayload
			if err := json.Unmarshal(task.Payload, &payload); err != nil || payload.Create == nil || payload.Create.Version != f.stat(t, 0, "scheduled-source").Version || task.ActorID != f.readerUser.ID {
				t.Fatalf("recovered schedule did not bind fresh source and actor: %+v %v", payload, err)
			}
			if more, err := f.app.schedules.engine.Tick(ctx, at.Add(time.Second)); err != nil || len(more) != 0 {
				t.Fatalf("offline occurrences replayed more than once: %+v %v", more, err)
			}
		})
	}
}

func TestScheduledBackupMasterReopenKeepsAcceptedOccurrence(t *testing.T) {
	f, restart := newRestartTransferFixture(t)
	f.app.closeSchedules()
	ctx := context.Background()
	transferWrite(t, f.roots[0], "scheduled-restart-source", []byte("durable scheduled backup"))
	item := scheduleFromResponse(t, f.admin.request("POST", "/schedules", apiSchedule(f.instances[0], "backup.create", scheduledBackup{Path: "scheduled-restart-source", Compression: "store", Consistency: backup.Consistency{Mode: "files"}}), model.ID(), 201))
	fires, err := f.app.schedules.engine.Tick(ctx, item.NextRun)
	if err != nil || len(fires) != 1 {
		t.Fatalf("initial scheduled occurrence: %+v %v", fires, err)
	}
	original := fires[0].Task
	restart() // Close the actual Master and SQLite, reopen, reconnect both nodes.
	f.app.closeSchedules()
	if more, err := f.app.schedules.engine.Tick(ctx, item.NextRun.Add(time.Second)); err != nil || len(more) != 0 {
		t.Fatalf("Master restart replayed accepted occurrence: %+v %v", more, err)
	}
	task := f.awaitFile(t, original, model.Succeeded)
	if task.ID != original.ID || task.RequestID != original.RequestID {
		t.Fatal("Master restart replaced scheduled task identity")
	}
	tasks, err := f.store.Tasks(ctx)
	if err != nil || len(tasks) != 1 {
		t.Fatalf("duplicate scheduled tasks after reopen: %+v %v", tasks, err)
	}
	var result backup.Result
	if err := json.Unmarshal(task.Result, &result); err != nil || result.Snapshot == nil {
		t.Fatalf("scheduled archive did not complete after reopen: %+v %v", result, err)
	}
	listed := f.admin.request("GET", "/instances/"+f.instances[0].ID+"/backups", nil, "", 200)
	var items []backup.SnapshotInfo
	if err := json.Unmarshal(listed["items"], &items); err != nil || len(items) != 1 || items[0].ID != result.ID {
		t.Fatalf("restarted schedule produced extra or missing archive: %+v %v", items, err)
	}
}

func TestSchedulesAcceptRealLifecycleConsoleAndFreshBackup(t *testing.T) {
	f := newFileFixture(t)
	f.app.closeSchedules()
	ctx := context.Background()
	created := f.admin.request("POST", "/instances", map[string]any{"nodeId": f.instances[0].NodeID, "name": "scheduled-runtime", "config": model.InstanceConfig{Mode: "native", Directory: f.roots[0], Command: nativeTestCommand(t, "scheduled-input"), Environment: nativeTestEnvironment(), StopSeconds: 1, KillSeconds: 1, Escalate: true}}, model.ID(), 201)
	var i model.Instance
	if err := json.Unmarshal(created["instance"], &i); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		f.awaitFile(t, parseFileTask(t, f.admin.request("POST", "/instances/"+i.ID+"/actions", map[string]string{"action": "stop"}, model.ID(), 202)), model.Succeeded)
	})
	run := func(action string, args any) model.Task {
		t.Helper()
		item := scheduleFromResponse(t, f.admin.request("POST", "/schedules", apiSchedule(i, action, args), model.ID(), 201))
		fires, err := f.app.schedules.engine.Tick(ctx, item.NextRun)
		if err != nil || len(fires) != 1 {
			t.Fatalf("trigger %s %v %v", action, fires, err)
		}
		task := f.awaitFile(t, fires[0].Task, model.Succeeded)
		if more, err := f.app.schedules.engine.Tick(ctx, item.NextRun.Add(time.Second)); err != nil || len(more) != 0 {
			t.Fatalf("occurrence repeated %v %v", more, err)
		}
		f.admin.request("DELETE", "/schedules/"+item.Spec.ID, map[string]int64{"revision": item.ConfigRevision}, model.ID(), 200)
		return task
	}
	run("instance.start", struct{}{})
	current, err := f.store.Instance(ctx, i.ID)
	if err != nil || current.RunID == "" {
		t.Fatal("real scheduled start missing run")
	}
	input := run("console.input", map[string]string{"data": "中文-scheduled\n"})
	var payload model.ConsoleInput
	if err = json.Unmarshal(input.Payload, &payload); err != nil || payload.RunID != current.RunID {
		t.Fatalf("unbound console run %+v %v", payload, err)
	}
	eventually(t, 3*time.Second, func() bool {
		b, e := os.ReadFile(filepath.Join(f.roots[0], "scheduled-input"))
		return e == nil && string(b) == "中文-scheduled\n"
	})
	run("instance.restart", struct{}{})
	next, err := f.store.Instance(ctx, i.ID)
	if err != nil || next.RunID == current.RunID {
		t.Fatal("scheduled restart did not change actual run")
	}
	run("instance.stop", struct{}{})
	if err = os.WriteFile(filepath.Join(f.roots[0], "backup-source"), []byte("fresh-at-fire"), 0600); err != nil {
		t.Fatal(err)
	}
	task := run("backup.create", scheduledBackup{Path: "backup-source", Compression: "deflate", Consistency: backup.Consistency{Mode: "files"}})
	var result backup.Result
	if err = json.Unmarshal(task.Result, &result); err != nil || result.Snapshot == nil || result.Stage != "succeeded" {
		t.Fatalf("actual backup failed %+v %v", result, err)
	}
	var bp backup.TaskPayload
	if err = json.Unmarshal(task.Payload, &bp); err != nil || bp.Create == nil || bp.Create.Version == "" || bp.Create.OwnerID != f.adminUser.ID {
		t.Fatal("backup payload not concretely bound")
	}
}

package master

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"blora.dev/panel/internal/backup"
	"blora.dev/panel/internal/daemon"
	"blora.dev/panel/internal/filesystem"
	"blora.dev/panel/internal/model"
)

// A real fixed-argv save executable used only by the private test Daemon.
// It publishes application state atomically and records its bound resource.
func TestBackupApplicationSaveExecutable(t *testing.T) {
	phase := os.Getenv("BLORA_BACKUP_PHASE")
	if phase == "" {
		return
	}
	if len(os.Args) < 5 || os.Getenv("BLORA_BACKUP_RESOURCE_ID") == "" {
		t.Fatal("missing fixed target or resource identity")
	}
	target, outcome := os.Args[len(os.Args)-2], os.Args[len(os.Args)-1]
	log, err := os.OpenFile(target+".hooks", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = log.WriteString(phase + "|" + os.Getenv("BLORA_BACKUP_RESOURCE_ID") + "\n")
	if err == nil {
		err = log.Sync()
	}
	closeErr := log.Close()
	if err != nil || closeErr != nil {
		t.Fatalf("hook receipt output: %v %v", err, closeErr)
	}
	if phase != "before" {
		return
	}
	if outcome == "fail" {
		t.Fatal("application refused save")
	}
	f, err := os.OpenFile(target+".saving", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString("application committed fresh state\n")
	if err == nil {
		err = f.Sync()
	}
	closeErr = f.Close()
	if err != nil || closeErr != nil {
		t.Fatalf("application save: %v %v", err, closeErr)
	}
	if err = os.Rename(target+".saving", target); err != nil {
		t.Fatal(err)
	}
}

func TestBackupFixedSaveHooksCapturePublishedStateAndRejectFailure(t *testing.T) {
	for _, outcome := range []string{"success", "fail"} {
		t.Run(outcome, func(t *testing.T) {
			index := 0
			f := newFileFixtureWithConfig(t, func(config *daemon.Config) {
				root := []string{"resource-a", "resource-b"}[index]
				index++
				target := filepath.Join(filepath.Dir(config.StateDir), root, "save.state")
				command := []string{os.Args[0], "-test.run=^TestBackupApplicationSaveExecutable$", "--", target, outcome}
				config.BackupHookCommands = map[string][]string{"backup.save.before": command, "backup.save.after": command}
			})
			i := f.instances[0]
			transferWrite(t, f.roots[0], "save.state", []byte("stale on-disk state"))
			version := f.stat(t, 0, "save.state").Version
			base := "/instances/" + i.ID
			body := map[string]any{"path": "save.state", "version": version, "compression": "store", "consistency": map[string]string{"mode": "save"}}
			key := model.ID()
			task := parseFileTask(t, f.admin.request("POST", base+"/backups", body, key, 202))
			want := model.Succeeded
			if outcome == "fail" {
				want = model.Failed
			}
			task = f.awaitFile(t, task, want)
			if replay := parseFileTask(t, f.admin.request("POST", base+"/backups", body, key, 202)); replay.ID != task.ID {
				t.Fatal("same backup request created another task")
			}
			calls, err := os.ReadFile(filepath.Join(f.roots[0], "save.state.hooks"))
			if err != nil || string(calls) != "before|"+i.ID+"\nafter|"+i.ID+"\n" {
				t.Fatalf("save/compensation repeated or bound incorrectly: %q %v", calls, err)
			}
			var result backup.Result
			if err := json.Unmarshal(task.Result, &result); err != nil || result.CleanupPending || result.Unknown {
				t.Fatalf("hook cleanup result: %+v %v", result, err)
			}
			if outcome == "fail" {
				if result.Snapshot != nil || !strings.Contains(task.Error, "exit status") {
					t.Fatalf("save failure reported successful backup: %+v", task)
				}
				assertDisk(t, f.roots[0], "save.state", "stale on-disk state")
				return
			}
			if result.Snapshot == nil {
				t.Fatal("saved backup missing snapshot")
			}
			prepared := f.admin.request("POST", base+"/restore-plans", map[string]any{"backupId": result.ID, "path": "restored.state", "version": filesystem.MissingVersion}, model.ID(), 201)
			var plan backup.RestorePlan
			if err := json.Unmarshal(prepared["plan"], &plan); err != nil {
				t.Fatal(err)
			}
			restore := parseFileTask(t, f.admin.request("POST", base+"/restores", map[string]any{"planId": plan.Request.ID, "planHash": plan.Hash}, model.ID(), 202))
			f.awaitFile(t, restore, model.Succeeded)
			assertDisk(t, f.roots[0], "restored.state", "application committed fresh state\n")
		})
	}
}

func TestScheduledBackupSaveHookCapturesPublishedState(t *testing.T) {
	index := 0
	f := newFileFixtureWithConfig(t, func(config *daemon.Config) {
		root := []string{"resource-a", "resource-b"}[index]
		index++
		target := filepath.Join(filepath.Dir(config.StateDir), root, "scheduled.state")
		command := []string{os.Args[0], "-test.run=^TestBackupApplicationSaveExecutable$", "--", target, "success"}
		config.BackupHookCommands = map[string][]string{"backup.save.before": command, "backup.save.after": command}
	})
	f.app.closeSchedules()
	transferWrite(t, f.roots[0], "scheduled.state", []byte("stale scheduled state"))
	i := f.instances[0]
	item := scheduleFromResponse(t, f.admin.request("POST", "/schedules", apiSchedule(i, "backup.create", scheduledBackup{
		Path: "scheduled.state", Compression: "store", Consistency: backup.Consistency{Mode: "save"},
	}), model.ID(), 201))
	fires, err := f.app.schedules.engine.Tick(context.Background(), item.NextRun)
	if err != nil || len(fires) != 1 {
		t.Fatalf("scheduled save occurrence: %+v %v", fires, err)
	}
	task := f.awaitFile(t, fires[0].Task, model.Succeeded)
	var result backup.Result
	if err := json.Unmarshal(task.Result, &result); err != nil || result.Snapshot == nil || result.Snapshot.Consistency.Mode != "save" {
		t.Fatalf("scheduled save hook result: %+v %v", result, err)
	}
	assertDisk(t, f.roots[0], "scheduled.state", "application committed fresh state\n")
	calls, err := os.ReadFile(filepath.Join(f.roots[0], "scheduled.state.hooks"))
	if err != nil || string(calls) != "before|"+i.ID+"\nafter|"+i.ID+"\n" {
		t.Fatalf("scheduled hook calls repeated or missing: %q %v", calls, err)
	}
}

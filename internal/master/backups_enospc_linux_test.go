//go:build linux

package master

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"blora.dev/panel/internal/backup"
	"blora.dev/panel/internal/daemon"
	"blora.dev/panel/internal/model"
	"golang.org/x/sys/unix"
)

func TestBackupRepositoryENOSPCDoesNotPublishArchive(t *testing.T) {
	if !privateENOSPCNamespace(t) {
		return
	}
	repository := t.TempDir()
	if err := unix.Mount("blora-backup-enospc", repository, "tmpfs", unix.MS_NODEV|unix.MS_NOSUID|unix.MS_NOEXEC, "size=1048576,mode=0700"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := unix.Unmount(repository, unix.MNT_DETACH); err != nil {
			t.Errorf("unmount private backup filesystem: %v", err)
		}
	})
	configured := false
	f := newFileFixtureWithConfig(t, func(config *daemon.Config) {
		if !configured {
			config.BackupRoot = repository
			configured = true
		}
	})
	body := bytes.Repeat([]byte("source data retained when backup repository fills\n"), 65536)
	transferWrite(t, f.roots[0], "source.bin", body)
	version := f.stat(t, 0, "source.bin").Version
	base := "/instances/" + f.instances[0].ID
	request := map[string]any{"path": "source.bin", "version": version, "compression": "store", "consistency": map[string]string{"mode": "files"}}
	task := parseFileTask(t, f.admin.request("POST", base+"/backups", request, model.ID(), 202))
	task = f.awaitFile(t, task, model.Failed)
	if !strings.Contains(strings.ToLower(task.Error), "no space left") {
		t.Fatalf("expected kernel ENOSPC: %s", task.Error)
	}
	var result backup.Result
	if err := json.Unmarshal(task.Result, &result); err != nil || result.Snapshot != nil || result.CleanupPending || result.Unknown {
		t.Fatalf("unsafe backup receipt: %+v %v", result, err)
	}
	var items []backup.SnapshotInfo
	listed := f.admin.request("GET", base+"/backups", nil, "", 200)
	if err := json.Unmarshal(listed["items"], &items); err != nil || len(items) != 0 {
		t.Fatalf("failed archive published: %+v %v", items, err)
	}
	got, err := os.ReadFile(filepath.Join(f.roots[0], "source.bin"))
	if err != nil || !bytes.Equal(got, body) {
		t.Fatalf("backup failure changed source: %v", err)
	}
	for _, name := range []string{"pending", "objects"} {
		entries, err := os.ReadDir(filepath.Join(repository, name))
		if err != nil || len(entries) != 0 {
			t.Fatalf("backup %s not cleaned: %v %v", name, entries, err)
		}
	}
}

func TestBackupRestoreENOSPCKeepsExistingTarget(t *testing.T) {
	if !privateENOSPCNamespace(t) {
		return
	}
	f := newFileFixture(t)
	if err := unix.Mount("blora-restore-enospc", f.roots[0], "tmpfs", unix.MS_NODEV|unix.MS_NOSUID|unix.MS_NOEXEC, "size=1048576,mode=0750"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := unix.Unmount(f.roots[0], unix.MNT_DETACH); err != nil {
			t.Errorf("unmount private restore filesystem: %v", err)
		}
	})
	transferWrite(t, f.roots[0], "target.bin", bytes.Repeat([]byte{0x37}, 768<<10))
	version := f.stat(t, 0, "target.bin").Version
	base := "/instances/" + f.instances[0].ID
	request := map[string]any{"path": "target.bin", "version": version, "compression": "store", "consistency": map[string]string{"mode": "files"}}
	task := f.awaitFile(t, parseFileTask(t, f.admin.request("POST", base+"/backups", request, model.ID(), 202)), model.Succeeded)
	var created backup.Result
	if err := json.Unmarshal(task.Result, &created); err != nil || created.Snapshot == nil {
		t.Fatalf("backup prerequisite: %+v %v", created, err)
	}
	transferWrite(t, f.roots[0], "target.bin", []byte("current target must survive"))
	transferWrite(t, f.roots[0], "other.bin", bytes.Repeat([]byte{0x42}, 512<<10))
	version = f.stat(t, 0, "target.bin").Version
	prepared := f.admin.request("POST", base+"/restore-plans", map[string]any{"backupId": created.ID, "path": "target.bin", "version": version}, model.ID(), 201)
	var plan backup.RestorePlan
	if err := json.Unmarshal(prepared["plan"], &plan); err != nil {
		t.Fatal(err)
	}
	task = parseFileTask(t, f.admin.request("POST", base+"/restores", map[string]any{"planId": plan.Request.ID, "planHash": plan.Hash, "overwritePlanHash": plan.Hash}, model.ID(), 202))
	task = f.awaitFile(t, task, model.Failed)
	if !strings.Contains(strings.ToLower(task.Error), "no space left") {
		t.Fatalf("expected kernel ENOSPC during restore: %s", task.Error)
	}
	assertDisk(t, f.roots[0], "target.bin", "current target must survive")
	other, err := os.ReadFile(filepath.Join(f.roots[0], "other.bin"))
	if err != nil || !bytes.Equal(other, bytes.Repeat([]byte{0x42}, 512<<10)) {
		t.Fatalf("unrelated target file changed: %v", err)
	}
	var items []backup.SnapshotInfo
	listed := f.admin.request("GET", base+"/backups", nil, "", 200)
	if err := json.Unmarshal(listed["items"], &items); err != nil || len(items) != 1 || items[0].ID != created.ID {
		t.Fatalf("restore failure lost source archive: %+v %v", items, err)
	}
	for _, name := range []string{"work", "uploads", "trash"} {
		entries, err := os.ReadDir(filepath.Join(f.roots[0], ".blora-files", name))
		if err != nil || len(entries) != 0 {
			t.Fatalf("restore %s not cleaned: %v %v", name, entries, err)
		}
	}
}

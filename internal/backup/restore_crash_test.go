package backup

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"blora.dev/panel/internal/filesystem"
	"blora.dev/panel/internal/model"
)

const backupRestoreCrashFixtureEnv = "BLORA_BACKUP_RESTORE_CRASH_BASE"

// TestBackupRestoreProcessKillDuringUpload kills a real restore subprocess
// after it has written target staging bytes but before the restored file is
// committed to its destination.
func TestBackupRestoreProcessKillDuringUpload(t *testing.T) {
	if base := os.Getenv(backupRestoreCrashFixtureEnv); base != "" {
		runBackupRestoreCrashHelper(t, base)
		return
	}

	f := newBackupFixture(t, Options{})
	ctx := context.Background()
	const payloadBytes int64 = 8 << 20
	if err := os.Mkdir(filepath.Join(f.sourceRoot, "world"), 0750); err != nil {
		t.Fatal(err)
	}
	largePath := filepath.Join(f.sourceRoot, "world", "large.bin")
	large, err := os.OpenFile(largePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0640)
	if err != nil {
		t.Fatal(err)
	}
	if err = large.Truncate(payloadBytes); err == nil {
		err = large.Chmod(0640)
	}
	if closeErr := large.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		t.Fatal(err)
	}
	entry, err := f.source.Stat(ctx, "world")
	if err != nil {
		t.Fatal(err)
	}
	create := CreateSpec{ID: model.ID(), OwnerID: "owner", Source: backupRef("source"), Path: "world", Version: entry.Version, Compression: "store", Consistency: Consistency{Mode: "files"}}
	created, err := f.manager.Create(ctx, f.source, create)
	if err != nil || created.Stage != "succeeded" || created.Snapshot == nil {
		t.Fatalf("create restore crash fixture snapshot: result=%+v err=%v", created, err)
	}
	plan := f.plan(t, create.ID, "restore")
	restore := restoreSpec(plan)
	restoreBody, err := json.Marshal(restore)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(f.base, "restore-spec.json"), restoreBody, 0600); err != nil {
		t.Fatal(err)
	}
	if err = f.manager.Close(); err != nil {
		t.Fatal(err)
	}
	if err = f.target.Close(); err != nil {
		t.Fatal(err)
	}

	uploadID := hashBytes([]byte(restore.ID + ":1"))[7:39]
	partPath := filepath.Join(f.targetRoot, "restore", ".blora-upload-"+uploadID)
	cmd := exec.Command(os.Args[0], "-test.run=^TestBackupRestoreProcessKillDuringUpload$", "-test.count=1")
	cmd.Env = append(os.Environ(), backupRestoreCrashFixtureEnv+"="+f.base)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	wait := make(chan error, 1)
	go func() { wait <- cmd.Wait() }()
	t.Cleanup(func() {
		if cmd.ProcessState == nil {
			_ = cmd.Process.Kill()
			<-wait
		}
	})

	deadline := time.NewTimer(30 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	var partialBytes int64
	for {
		if info, statErr := os.Stat(partPath); statErr == nil && info.Size() >= 1<<20 && info.Size() < payloadBytes {
			partialBytes = info.Size()
			if err = cmd.Process.Kill(); err != nil {
				t.Fatalf("kill restore writer at %d bytes: %v", partialBytes, err)
			}
			break
		}
		select {
		case waitErr := <-wait:
			t.Fatalf("restore helper exited before partial upload was observed: wait=%v stdout=%s stderr=%s", waitErr, stdout.String(), stderr.String())
		case <-deadline.C:
			t.Fatalf("restore helper did not produce a partial target upload; stdout=%s stderr=%s", stdout.String(), stderr.String())
		case <-ticker.C:
		}
	}
	if waitErr := <-wait; waitErr == nil {
		t.Fatal("restore helper exited normally instead of being killed")
	} else {
		var exitErr *exec.ExitError
		if !errors.As(waitErr, &exitErr) {
			t.Fatalf("wait for killed restore helper: %v", waitErr)
		}
	}

	f.manager, err = New(f.repo, f.options)
	if err != nil {
		t.Fatalf("reopen backup state after restore process kill: %v", err)
	}
	f.target, err = filesystem.New(f.targetRoot, filesystem.Options{StateDir: filepath.Join(f.base, "target-state")})
	if err != nil {
		t.Fatalf("reopen restore target after process kill: %v", err)
	}
	partial, err := f.manager.Restore(ctx, f.target, restore)
	if !errors.Is(err, ErrInterrupted) || partial.Stage != "interrupted" || partial.Unknown || partial.CleanupPending || !partial.Partial || partial.Completed != 1 {
		t.Fatalf("reconcile interrupted restore: result=%+v err=%v", partial, err)
	}
	upload, err := f.target.UploadStatus(ctx, uploadID)
	if err != nil || upload.Stage != "cancelled" || upload.Offset == 0 {
		t.Fatalf("partial staging upload was not cancelled: upload=%+v err=%v", upload, err)
	}
	if _, err = os.Stat(partPath); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("partial restore bytes remain after reconciliation: %v", err)
	}
	if _, err = f.target.Stat(ctx, "restore/large.bin"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("interrupted file was published to the target: %v", err)
	}
	again, err := f.manager.Restore(ctx, f.target, restore)
	if !errors.Is(err, ErrInterrupted) || again.Stage != "interrupted" || again.Completed != partial.Completed || !again.Partial || again.Unknown || again.CleanupPending {
		t.Fatalf("replaying interrupted restore changed outcome: result=%+v err=%v", again, err)
	}

	newPlan := f.plan(t, create.ID, "restore")
	recovered, err := f.manager.Restore(ctx, f.target, restoreSpec(newPlan))
	if err != nil || recovered.Stage != "succeeded" || recovered.Completed != 2 || recovered.Bytes != payloadBytes {
		t.Fatalf("new restore after interrupted cleanup: result=%+v err=%v", recovered, err)
	}
	wantHash := ""
	for _, planned := range newPlan.Entries {
		if planned.TargetPath == "restore/large.bin" {
			wantHash = planned.Hash
			break
		}
	}
	final, err := f.target.Stat(ctx, "restore/large.bin")
	if err != nil || wantHash == "" || final.Size != payloadBytes || final.Version != wantHash {
		t.Fatalf("restored target did not match source archive: entry=%+v err=%v", final, err)
	}
	t.Logf("SIGKILL observed with %d bytes staged of %d; partial upload cancelled, no destination published, fresh restore verified", partialBytes, payloadBytes)
}

func runBackupRestoreCrashHelper(t *testing.T, base string) {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(base, "restore-spec.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec RestoreSpec
	if err = json.Unmarshal(body, &spec); err != nil {
		t.Fatal(err)
	}
	manager, err := New(filepath.Join(base, "repository"), Options{StateDir: filepath.Join(base, "backup-state")})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()
	target, err := filesystem.New(filepath.Join(base, "target"), filesystem.Options{StateDir: filepath.Join(base, "target-state")})
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	result, err := manager.Restore(context.Background(), target, spec)
	if err != nil || result.Stage != "succeeded" {
		t.Fatalf("restore child outcome: result=%+v err=%v", result, err)
	}
}

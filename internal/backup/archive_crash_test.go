package backup

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"blora.dev/panel/internal/filesystem"
	"blora.dev/panel/internal/model"
)

const backupArchiveCrashFixtureEnv = "BLORA_BACKUP_PACK_CRASH_BASE"

// TestBackupCreateProcessKillDuringPacking kills a real backup subprocess
// after its temporary ZIP contains payload bytes but before the archive is
// finalized or published as a snapshot.
func TestBackupCreateProcessKillDuringPacking(t *testing.T) {
	if base := os.Getenv(backupArchiveCrashFixtureEnv); base != "" {
		runBackupArchiveCrashHelper(t, base)
		return
	}

	base := t.TempDir()
	sourceRoot := filepath.Join(base, "source")
	sourceState := filepath.Join(base, "source-state")
	repository := filepath.Join(base, "repository")
	backupState := filepath.Join(base, "backup-state")
	if err := os.MkdirAll(sourceRoot, 0700); err != nil {
		t.Fatal(err)
	}
	const payloadBytes int64 = 64 << 20
	largePath := filepath.Join(sourceRoot, "large.bin")
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
	smallPath := filepath.Join(sourceRoot, "small.bin")
	if err = os.WriteFile(smallPath, bytes.Repeat([]byte("retry-after-crash\n"), 64), 0640); err != nil {
		t.Fatal(err)
	}

	source, err := filesystem.New(sourceRoot, filesystem.Options{StateDir: sourceState})
	if err != nil {
		t.Fatal(err)
	}
	largeEntry, err := source.Stat(context.Background(), "large.bin")
	if err != nil {
		_ = source.Close()
		t.Fatal(err)
	}
	spec := CreateSpec{ID: model.ID(), OwnerID: "owner", Source: backupRef("source"), Path: "large.bin", Version: largeEntry.Version, Compression: "store", Consistency: Consistency{Mode: "files"}}
	if err = source.Close(); err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(base, "spec.json"), body, 0600); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestBackupCreateProcessKillDuringPacking$", "-test.count=1")
	cmd.Env = append(os.Environ(), backupArchiveCrashFixtureEnv+"="+base)
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

	pendingArchive := filepath.Join(repository, filepath.FromSlash(pendingPath(spec.ID)))
	deadline := time.NewTimer(30 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	var partialBytes int64
	for {
		if info, statErr := os.Stat(pendingArchive); statErr == nil && info.Size() >= 1<<20 && info.Size() < payloadBytes {
			partialBytes = info.Size()
			if err = cmd.Process.Kill(); err != nil {
				t.Fatalf("kill backup packer at %d bytes: %v", partialBytes, err)
			}
			break
		}
		select {
		case waitErr := <-wait:
			t.Fatalf("backup helper exited before partial archive was observed: wait=%v stdout=%s stderr=%s", waitErr, stdout.String(), stderr.String())
		case <-deadline.C:
			t.Fatalf("backup helper did not produce a partial archive; stdout=%s stderr=%s", stdout.String(), stderr.String())
		case <-ticker.C:
		}
	}
	if waitErr := <-wait; waitErr == nil {
		t.Fatal("backup helper exited normally instead of being killed")
	} else {
		var exitErr *exec.ExitError
		if !errors.As(waitErr, &exitErr) {
			t.Fatalf("wait for killed backup helper: %v", waitErr)
		}
	}

	manager, err := New(repository, Options{StateDir: backupState})
	if err != nil {
		t.Fatalf("reopen repository after process kill: %v", err)
	}
	defer manager.Close()
	source, err = filesystem.New(sourceRoot, filesystem.Options{StateDir: sourceState})
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	ctx := context.Background()
	result, err := manager.Create(ctx, source, spec)
	if !errors.Is(err, ErrInterrupted) || result.Stage != "interrupted" || !result.Unknown || result.CleanupPending {
		t.Fatalf("interrupted archive outcome: result=%+v err=%v", result, err)
	}
	if _, err = os.Stat(pendingArchive); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("partial archive was not removed: err=%v", err)
	}
	if _, err = manager.Inspect(ctx, spec.ID); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("interrupted archive was published as a snapshot: err=%v", err)
	}

	smallEntry, err := source.Stat(ctx, "small.bin")
	if err != nil {
		t.Fatal(err)
	}
	retry := CreateSpec{ID: model.ID(), OwnerID: "owner", Source: backupRef("source"), Path: "small.bin", Version: smallEntry.Version, Compression: "store", Consistency: Consistency{Mode: "files"}}
	recovered, err := manager.Create(ctx, source, retry)
	if err != nil || recovered.Stage != "succeeded" || recovered.Snapshot == nil {
		t.Fatalf("fresh backup after interrupted cleanup: result=%+v err=%v", recovered, err)
	}
	if _, err = manager.Inspect(ctx, retry.ID); err != nil {
		t.Fatalf("verify backup after interrupted cleanup: %v", err)
	}
	t.Logf("SIGKILL observed while pending ZIP was %d bytes of a %d-byte source", partialBytes, payloadBytes)
}

func runBackupArchiveCrashHelper(t *testing.T, base string) {
	body, err := os.ReadFile(filepath.Join(base, "spec.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec CreateSpec
	if err = json.Unmarshal(body, &spec); err != nil {
		t.Fatal(err)
	}
	source, err := filesystem.New(filepath.Join(base, "source"), filesystem.Options{StateDir: filepath.Join(base, "source-state")})
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	manager, err := New(filepath.Join(base, "repository"), Options{StateDir: filepath.Join(base, "backup-state")})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()
	if _, err = manager.Create(context.Background(), source, spec); err != nil {
		t.Fatal(err)
	}
	t.Fatal("backup completed before parent terminated packer")
}

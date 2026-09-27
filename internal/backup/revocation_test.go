package backup

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"blora.dev/panel/internal/filesystem"
	"blora.dev/panel/internal/model"
)

func TestBackupMidStreamRevocationAndSourceChange(t *testing.T) {
	for _, change := range []bool{false, true} {
		name := "revocation"
		if change {
			name = "source-change"
		}
		t.Run(name, func(t *testing.T) {
			f := newBackupFixture(t, Options{})
			diskWrite(t, filepath.Join(f.sourceRoot, "large"), bytes.Repeat([]byte("data"), 100000))
			spec := f.spec(t, "large")
			spec.Compression = "store"
			triggered := false
			f.manager.opts.AuthorizeSnapshotRead = func(context.Context, string, model.ResourceRef) error {
				info, err := os.Stat(filepath.Join(f.repo, filepath.FromSlash(pendingPath(spec.ID))))
				if err == nil && info.Size() > 0 && !triggered {
					triggered = true
					if change {
						diskWrite(t, filepath.Join(f.sourceRoot, "large"), []byte("source replaced while reading"))
						return nil
					}
					return errors.New("file.read revoked")
				}
				return nil
			}
			result, err := f.manager.Create(context.Background(), f.source, spec)
			if err == nil || !triggered || result.Stage == "succeeded" {
				t.Fatalf("stream change not detected: %+v %v trigger=%v", result, err, triggered)
			}
			if _, err = os.Stat(filepath.Join(f.repo, filepath.FromSlash(archivePath(spec.ID)))); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("failed snapshot published %v", err)
			}
			if _, err = os.Stat(filepath.Join(f.repo, filepath.FromSlash(pendingPath(spec.ID)))); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("owned partial not cleaned %v", err)
			}
		})
	}
}

func TestRestoreReadWriteRevocationCancelsOnlyOwnStaging(t *testing.T) {
	for _, targetDenied := range []bool{false, true} {
		name := "source-read"
		if targetDenied {
			name = "target-write"
		}
		t.Run(name, func(t *testing.T) {
			f := newBackupFixture(t, Options{})
			ctx := context.Background()
			diskWrite(t, filepath.Join(f.sourceRoot, "tree", "a"), []byte("done"))
			diskWrite(t, filepath.Join(f.sourceRoot, "tree", "b"), bytes.Repeat([]byte("binary"), 50000))
			spec := f.spec(t, "tree")
			if _, err := f.manager.Create(ctx, f.source, spec); err != nil {
				t.Fatal(err)
			}
			plan := f.plan(t, spec.ID, "destination")
			restore := restoreSpec(plan)
			other := filesystem.UploadSpec{ID: model.ID(), OwnerID: "unrelated", Path: "other", Total: 5, Hash: hashBytes([]byte("other")), ExpectedVersion: filesystem.MissingVersion, SourceName: "other", SourceFingerprint: "other"}
			if _, err := f.target.BeginUpload(ctx, other); err != nil {
				t.Fatal(err)
			}
			denied := false
			authorize := func(context.Context, string, model.ResourceRef) error {
				entries, _ := os.ReadDir(filepath.Join(f.targetRoot, "destination"))
				for _, entry := range entries {
					if strings.HasPrefix(entry.Name(), ".blora-upload-") {
						info, err := entry.Info()
						if err != nil {
							return err
						}
						if info.Size() >= chunkBytes {
							denied = true
							return errors.New("permission revoked after synced chunk")
						}
					}
				}
				return nil
			}
			if targetDenied {
				f.manager.opts.AuthorizeRestoreWrite = authorize
			} else {
				f.manager.opts.AuthorizeSnapshotRead = authorize
			}
			result, err := f.manager.Restore(ctx, f.target, restore)
			if err == nil || !denied || result.Stage != "interrupted" || !result.Partial || result.Unknown || result.Completed != 2 {
				t.Fatalf("revocation receipt %+v %v denied=%v", result, err, denied)
			}
			if b, err := os.ReadFile(filepath.Join(f.targetRoot, "destination", "a")); err != nil || string(b) != "done" {
				t.Fatalf("completed file lost %q %v", b, err)
			}
			if _, err = os.Stat(filepath.Join(f.targetRoot, "destination", "b")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("unconfirmed file committed %v", err)
			}
			entries, err := os.ReadDir(filepath.Join(f.targetRoot, "destination"))
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".blora-upload-") {
					t.Fatal("own staging retained despite acknowledged cancel")
				}
			}
			if status, err := f.target.UploadStatus(ctx, other.ID); err != nil || status.Stage != "receiving" {
				t.Fatalf("other upload affected %+v %v", status, err)
			}
		})
	}
}

func TestRestoreCrossOwnerAuthorizationAndRepeatedReceiptRevocation(t *testing.T) {
	f := newBackupFixture(t, Options{})
	ctx := context.Background()
	diskWrite(t, filepath.Join(f.sourceRoot, "file"), []byte("data"))
	spec := f.spec(t, "file")
	if _, err := f.manager.Create(ctx, f.source, spec); err != nil {
		t.Fatal(err)
	}
	request := RestoreRequest{ID: model.ID(), OwnerID: "reader", BackupID: spec.ID, Target: backupRef("target"), Path: "dest", Version: filesystem.MissingVersion}
	if _, err := f.manager.PlanRestore(ctx, f.target, request); !errors.Is(err, ErrConflict) {
		t.Fatalf("default cross-owner access %v", err)
	}
	allowed := true
	f.manager.opts.AuthorizeSnapshotRead = func(_ context.Context, actor string, source model.ResourceRef) error {
		if actor != "reader" || source != spec.Source || !allowed {
			return errors.New("read denied")
		}
		return nil
	}
	plan, err := f.manager.PlanRestore(ctx, f.target, request)
	if err != nil {
		t.Fatal(err)
	}
	restore := restoreSpec(plan)
	restore.OwnerID = "reader"
	result, err := f.manager.Restore(ctx, f.target, restore)
	if err != nil || result.Stage != "succeeded" {
		t.Fatalf("authorized cross owner %+v %v", result, err)
	}
	allowed = false
	if _, err = f.manager.PlanRestore(ctx, f.target, request); err == nil {
		t.Fatal("old plan bypassed revoked read")
	}
	if _, err = f.manager.Restore(ctx, f.target, restore); err == nil {
		t.Fatal("receipt bypassed revoked read")
	}
	var checkpoint restoreCheckpoint
	if err = f.manager.readRecord("restore/"+restore.ID+".json", &checkpoint); err != nil {
		t.Fatal(err)
	}
	if checkpoint.Spec.OwnerID != "reader" {
		t.Fatal("audit actor impersonated snapshot creator")
	}
}

func TestRestorePlanningRechecksTargetAuthorizationBeforePublishing(t *testing.T) {
	f := newBackupFixture(t, Options{})
	ctx := context.Background()
	diskWrite(t, filepath.Join(f.sourceRoot, "tree", "a"), []byte("data"))
	spec := f.spec(t, "tree")
	if _, err := f.manager.Create(ctx, f.source, spec); err != nil {
		t.Fatal(err)
	}
	request := RestoreRequest{ID: model.ID(), OwnerID: spec.OwnerID, BackupID: spec.ID, Target: backupRef("target"), Path: "dest", Version: filesystem.MissingVersion}
	calls := 0
	denied := errors.New("target write revoked during restore preview")
	f.manager.opts.AuthorizeRestoreWrite = func(_ context.Context, actor string, ref model.ResourceRef) error {
		if actor != request.OwnerID || ref != request.Target {
			t.Fatal("wrong restore preview authority binding")
		}
		calls++
		if calls > 1 {
			return denied
		}
		return nil
	}
	if _, err := f.manager.PlanRestore(ctx, f.target, request); !errors.Is(err, denied) {
		t.Fatalf("revoked target preview published: %v", err)
	}
	if _, err := f.manager.GetRestorePlan(ctx, request.ID); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("revoked preview left plan: %v", err)
	}
}

package backup

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

type atomicSaveHooks struct {
	durableTestHooks
	source string
}

func (h *atomicSaveHooks) Run(ctx context.Context, call HookCall) (HookReceipt, error) {
	if call.Phase == "before" {
		file, err := os.OpenFile(h.source+".new", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0640)
		if err != nil {
			return HookReceipt{}, err
		}
		_, err = file.Write([]byte("saved by application"))
		if err == nil {
			err = file.Sync()
		}
		if e := file.Close(); err == nil {
			err = e
		}
		if err != nil {
			return HookReceipt{}, err
		}
		if err = os.Rename(h.source+".new", h.source); err != nil {
			return HookReceipt{}, err
		}
	}
	return h.durableTestHooks.Run(ctx, call)
}

func TestBackupSaveHookBindsPostSaveVersionAndObject(t *testing.T) {
	f := newBackupFixture(t, Options{})
	ctx := context.Background()
	p := filepath.Join(f.sourceRoot, "savefile")
	diskWrite(t, p, []byte("before save"))
	spec := f.spec(t, "savefile")
	spec.Consistency.Mode = "save"
	original, err := f.source.RelationFacts(ctx, "savefile")
	if err != nil {
		t.Fatal(err)
	}
	f.manager.opts.Hooks = &atomicSaveHooks{durableTestHooks: durableTestHooks{root: t.TempDir()}, source: p}
	result, err := f.manager.Create(ctx, f.source, spec)
	if err != nil {
		t.Fatal(err)
	}
	if result.Snapshot.CapturedVersion != hashBytes([]byte("saved by application")) || result.Snapshot.CapturedVersion == spec.Version {
		t.Fatalf("pre-save content certified %+v", result)
	}
	var snap Snapshot
	if err = f.manager.readRecord("snapshots/"+spec.ID+".json", &snap); err != nil {
		t.Fatal(err)
	}
	if snap.Manifest.SourceObjectID == original.ObjectID {
		t.Fatal("post-save object not rebound")
	}
	plan := f.plan(t, spec.ID, "restored")
	if _, err = f.manager.Restore(ctx, f.target, restoreSpec(plan)); err != nil {
		t.Fatal(err)
	}
	text, err := f.target.ReadText(ctx, "restored")
	if err != nil || text.Content != "saved by application" {
		t.Fatalf("saved content %v %v", text, err)
	}
}

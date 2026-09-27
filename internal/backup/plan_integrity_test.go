package backup

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRestoreRejectsMutatedPlanAndUnboundOverwriteConfirmation(t *testing.T) {
	f := newBackupFixture(t, Options{})
	ctx := context.Background()
	diskWrite(t, filepath.Join(f.sourceRoot, "file"), []byte("backup"))
	spec := f.spec(t, "file")
	if _, err := f.manager.Create(ctx, f.source, spec); err != nil {
		t.Fatal(err)
	}
	diskWrite(t, filepath.Join(f.targetRoot, "target"), []byte("baseline"))
	plan := f.plan(t, spec.ID, "target")
	restore := restoreSpec(plan)
	restore.Overwrite = nil
	restore.OverwritePlanHash = hashBytes([]byte("different plan"))
	if _, err := f.manager.Restore(ctx, f.target, restore); !errors.Is(err, ErrConflict) {
		t.Fatalf("unbound overwrite confirmation %v", err)
	}
	restore.OverwritePlanHash = plan.Hash
	plan.Entries[0].TargetPath = "unconfirmed-path"
	if err := f.manager.writeRecord("plans/"+plan.Request.ID+".json", plan); err != nil {
		t.Fatal(err)
	}
	if _, err := f.manager.Restore(ctx, f.target, restore); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("mutated manifest retained confirmation %v", err)
	}
	if body, err := os.ReadFile(filepath.Join(f.targetRoot, "target")); err != nil || string(body) != "baseline" {
		t.Fatalf("unconfirmed overwrite %q %v", body, err)
	}
}

//go:build linux

package containers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCLIRecoveryRetainsUnconfirmedRecords(t *testing.T) {
	m := saveTestManager(t, t.TempDir())
	dir := filepath.Join(m.options.StateRoot, "cli-runs")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	marker := randomID()
	path := filepath.Join(dir, marker)
	if err := atomicJSON(path, cliRunRecord{Marker: randomID()}); err != nil {
		t.Fatal(err)
	}
	if err := m.RecoverCLI(context.Background()); !errors.Is(err, ErrIdentity) {
		t.Fatalf("corrupt identity accepted: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("unconfirmed evidence removed", err)
	}
	_, err := m.Execute(context.Background(), "admin", Operation{TaskID: "blocked_mutation", Action: "volume.create", Name: "must-not-create"}, nil)
	if !errors.Is(err, ErrUnknown) {
		t.Fatalf("mutation accepted with unconfirmed CLI: %v", err)
	}
	if _, err := os.Stat(m.operationPath("blocked_mutation")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("blocked mutation entered execution")
	}
	if err := atomicJSON(path, cliRunRecord{Marker: marker}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := m.RecoverCLI(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("recovery ignored cancellation", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("cancelled recovery removed evidence", err)
	}
	if err := m.RecoverCLI(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("confirmed empty launch not cleared", err)
	}
}

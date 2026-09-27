package daemon

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"blora.dev/panel/internal/backup"
	"blora.dev/panel/internal/model"
)

// TestBackupHookHelper is launched as a fixed argv by the runner test. It is
// intentionally not a shell: the test proves the adapter forwards structured
// environment values while keeping command arguments immutable.
func TestBackupHookHelper(t *testing.T) {
	if os.Getenv("BLORA_BACKUP_HOOK_HELPER") != "1" {
		return
	}
	path := os.Getenv("BLORA_HOOK_OUTPUT")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		os.Exit(2)
	}
	_, _ = f.WriteString(os.Getenv("BLORA_BACKUP_PHASE") + "|" + os.Getenv("BLORA_BACKUP_RESOURCE_ID") + "\n")
	_ = f.Close()
}

func TestBackupHookRunnerUsesFixedCommandsAndNeverRepeatsUnknownSideEffects(t *testing.T) {
	root := t.TempDir()
	output := filepath.Join(root, "effects.log")
	oldHelper, oldOutput := os.Getenv("BLORA_BACKUP_HOOK_HELPER"), os.Getenv("BLORA_HOOK_OUTPUT")
	if err := os.Setenv("BLORA_BACKUP_HOOK_HELPER", "1"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("BLORA_HOOK_OUTPUT", output); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Setenv("BLORA_BACKUP_HOOK_HELPER", oldHelper)
		_ = os.Setenv("BLORA_HOOK_OUTPUT", oldOutput)
	})

	runner, err := newBackupHookRunner(filepath.Join(root, "state"), map[string][]string{
		"backup.save.before": {os.Args[0], "-test.run=^TestBackupHookHelper$", "--"},
	})
	if err != nil {
		t.Fatal(err)
	}
	call := backup.HookCall{ID: "backup-1:before", BackupID: "backup-1", OwnerID: "owner-1", Resource: model.ResourceRef{Kind: "instance", ID: "instance-1", NodeID: "node-1"}, Strategy: backup.Consistency{Mode: "save"}, Phase: "before"}
	receipt, err := runner.Run(context.Background(), call)
	if err != nil || receipt.State != "succeeded" {
		t.Fatalf("hook did not succeed: %+v %v", receipt, err)
	}
	// Repeating the same durable hook ID returns the stored result and does not
	// invoke the executable a second time.
	if again, err := runner.Run(context.Background(), call); err != nil || again.State != "succeeded" {
		t.Fatalf("repeated hook was not an idempotent receipt: %+v %v", again, err)
	}
	b, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.Split(strings.TrimSpace(string(b)), "\n"); len(lines) != 1 || lines[0] != "before|instance-1" {
		t.Fatalf("fixed command or environment binding failed: %q", b)
	}
	if observed, err := runner.Reconcile(context.Background(), call); err != nil || observed.State != "succeeded" {
		t.Fatalf("successful hook was not reconciled: %+v %v", observed, err)
	}

	unknown := call
	unknown.ID = "backup-2:before"
	unknown.BackupID = "backup-2"
	unknown.Strategy = backup.Consistency{Mode: "hooks"}
	if _, err := runner.Run(context.Background(), unknown); !errors.Is(err, backup.ErrUnsupported) {
		t.Fatalf("unconfigured hook was executed or reported the wrong error: %v", err)
	}
	// A persisted running receipt is always converted to unknown and never
	// launched again, even if the original process may have completed.
	crash := call
	crash.ID = "backup-3:before"
	crash.BackupID = "backup-3"
	if err := runner.write(crash.ID, backupHookRecord{Call: crash, Receipt: backup.HookReceipt{ID: crash.ID, State: "running"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Reconcile(context.Background(), crash); !errors.Is(err, backup.ErrInterrupted) {
		t.Fatalf("running receipt was not made unknown: %v", err)
	}
	if _, err := runner.Run(context.Background(), crash); !errors.Is(err, backup.ErrInterrupted) {
		t.Fatalf("unknown side effect was repeated: %v", err)
	}
}

func TestBackupHookRunnerRejectsMalformedConfiguration(t *testing.T) {
	if _, err := newBackupHookRunner(t.TempDir(), map[string][]string{"bad ref": {"/bin/true"}}); err == nil {
		t.Fatal("malformed hook reference accepted")
	}
	if _, err := newBackupHookRunner(t.TempDir(), map[string][]string{"ok": {"/bin/true", "bad\x00arg"}}); err == nil {
		t.Fatal("NUL command argument accepted")
	}
}

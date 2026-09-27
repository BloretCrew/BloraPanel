package backup

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"blora.dev/panel/internal/filesystem"
)

type exitingHooks struct{ durableTestHooks }

func (h *exitingHooks) Run(ctx context.Context, call HookCall) (HookReceipt, error) {
	receipt, err := h.durableTestHooks.Run(ctx, call)
	if err == nil && call.Phase == "before" {
		os.Exit(73)
	}
	return receipt, err
}

func TestBackupActualProcessExitBeforeHookReceipt(t *testing.T) {
	if base := os.Getenv("BLORA_BACKUP_CRASH_FIXTURE"); base != "" {
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
		manager, err := New(filepath.Join(base, "repository"), Options{StateDir: filepath.Join(base, "backup-state"), Hooks: &exitingHooks{durableTestHooks{root: filepath.Join(base, "hook-effects"), unknown: true}}})
		if err != nil {
			t.Fatal(err)
		}
		defer manager.Close()
		if _, err = manager.Create(context.Background(), source, spec); err != nil {
			t.Fatal(err)
		}
		t.Fatal("crash point did not execute")
	}
	f := newBackupFixture(t, Options{})
	diskWrite(t, filepath.Join(f.sourceRoot, "file"), []byte("data"))
	spec := f.spec(t, "file")
	spec.Consistency.Mode = "hooks"
	body, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	diskWrite(t, filepath.Join(f.base, "spec.json"), body)
	hookDir := filepath.Join(f.base, "hook-effects")
	if err = os.Mkdir(hookDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err = f.manager.Close(); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestBackupActualProcessExitBeforeHookReceipt$", "-test.count=1")
	cmd.Env = append(os.Environ(), "BLORA_BACKUP_CRASH_FIXTURE="+f.base)
	output, err := cmd.CombinedOutput()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 73 {
		t.Fatalf("helper exit %v output=%s", err, output)
	}
	f.options.Hooks = &durableTestHooks{root: hookDir, unknown: true}
	f.manager, err = New(f.repo, f.options)
	if err != nil {
		t.Fatal(err)
	}
	result, err := f.manager.Create(context.Background(), f.source, spec)
	if !errors.Is(err, ErrInterrupted) || !result.Unknown || result.Stage == "succeeded" {
		t.Fatalf("unknown effect replayed %+v %v", result, err)
	}
	effects, err := os.ReadDir(hookDir)
	if err != nil || len(effects) != 2 {
		t.Fatalf("original effect/compensation %v %v", effects, err)
	}
	if _, err = f.manager.Inspect(context.Background(), spec.ID); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("crashed capture certified snapshot %v", err)
	}
}

package daemon

// This file contains the deliberately small, opt-in bridge from the backup
// consistency contract to an administrator supplied executable.  Commands
// are configured by the Daemon operator, never assembled from request text or
// passed through a shell.  The persisted receipt is written before Start so a
// crash can only leave an operation unknown; Reconcile never runs it again.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"blora.dev/panel/internal/backup"
)

var backupHookRef = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,128}$`)

type backupHookRecord struct {
	Call    backup.HookCall    `json:"call"`
	Receipt backup.HookReceipt `json:"receipt"`
	PID     int                `json:"pid,omitempty"`
}

type backupHookRunner struct {
	root     string
	commands map[string][]string
	mu       sync.Mutex
}

func newBackupHookRunner(root string, configured map[string][]string) (*backupHookRunner, error) {
	if root == "" {
		return nil, errors.New("backup hook state directory is required")
	}
	commands := make(map[string][]string, len(configured))
	for ref, command := range configured {
		if !backupHookRef.MatchString(ref) || len(command) == 0 || len(command) > 64 {
			return nil, fmt.Errorf("invalid backup hook command reference %q", ref)
		}
		fixed := append([]string(nil), command...)
		for _, arg := range fixed {
			if arg == "" || len(arg) > 4096 || strings.IndexByte(arg, 0) >= 0 {
				return nil, fmt.Errorf("invalid backup hook command argument for %q", ref)
			}
		}
		commands[ref] = fixed
	}
	if len(commands) == 0 {
		return nil, errors.New("backup hook command map is empty")
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	return &backupHookRunner{root: root, commands: commands}, nil
}

func (h *backupHookRunner) command(call backup.HookCall) ([]string, error) {
	ref := call.Strategy.Before
	if call.Phase == "after" {
		ref = call.Strategy.After
	}
	if ref == "" {
		if call.Strategy.Mode == "hooks" {
			return nil, backup.ErrUnsupported
		}
		ref = "backup." + call.Strategy.Mode + "." + call.Phase
	}
	if !backupHookRef.MatchString(ref) {
		return nil, backup.ErrUnsupported
	}
	command := h.commands[ref]
	if len(command) == 0 {
		return nil, backup.ErrUnsupported
	}
	return append([]string(nil), command...), nil
}

func (h *backupHookRunner) recordPath(id string) string {
	sum := sha256.Sum256([]byte(id))
	return filepath.Join(h.root, hex.EncodeToString(sum[:])+".json")
}

func sameHookCall(a, b backup.HookCall) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return string(x) == string(y)
}

func (h *backupHookRunner) read(id string) (backupHookRecord, error) {
	var record backupHookRecord
	b, err := os.ReadFile(h.recordPath(id))
	if err != nil {
		return record, err
	}
	err = json.Unmarshal(b, &record)
	if err == nil && record.Call.ID != id {
		err = backup.ErrInterrupted
	}
	return record, err
}

func (h *backupHookRunner) write(id string, record backupHookRecord) error {
	b, err := json.Marshal(record)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(h.root, ".hook-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err = tmp.Chmod(0600); err == nil {
		_, err = tmp.Write(b)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	// A missing record is always treated as unknown during Reconcile, so the
	// remove-before-rename fallback remains safe on Windows where replacing an
	// open destination is not consistently supported.
	if err = os.Remove(h.recordPath(id)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(tmpName, h.recordPath(id))
}

func (h *backupHookRunner) Run(ctx context.Context, call backup.HookCall) (backup.HookReceipt, error) {
	command, err := h.command(call)
	if err != nil {
		return backup.HookReceipt{ID: call.ID, State: "failed", At: time.Now().UTC()}, err
	}
	if call.ID == "" || call.BackupID == "" || call.OwnerID == "" || call.Phase != "before" && call.Phase != "after" {
		return backup.HookReceipt{ID: call.ID, State: "failed", At: time.Now().UTC()}, backup.ErrInvalid
	}
	h.mu.Lock()
	previous, readErr := h.read(call.ID)
	if readErr == nil {
		if !sameHookCall(previous.Call, call) {
			h.mu.Unlock()
			return backup.HookReceipt{ID: call.ID, State: "unknown", At: time.Now().UTC()}, backup.ErrConflict
		}
		if previous.Receipt.State == "succeeded" || previous.Receipt.State == "failed" || previous.Receipt.State == "unknown" {
			h.mu.Unlock()
			if previous.Receipt.State == "succeeded" {
				return previous.Receipt, nil
			}
			return previous.Receipt, backup.ErrInterrupted
		}
		// A prior running invocation may have performed the side effect before a
		// process or daemon crash. Never start a second invocation.
		h.mu.Unlock()
		previous.Receipt.State = "unknown"
		previous.Receipt.At = time.Now().UTC()
		_ = h.write(call.ID, previous)
		return previous.Receipt, backup.ErrInterrupted
	}
	if !errors.Is(readErr, os.ErrNotExist) {
		h.mu.Unlock()
		return backup.HookReceipt{ID: call.ID, State: "unknown", At: time.Now().UTC()}, readErr
	}
	record := backupHookRecord{Call: call, Receipt: backup.HookReceipt{ID: call.ID, State: "running", At: time.Now().UTC()}}
	if err = h.write(call.ID, record); err != nil {
		h.mu.Unlock()
		return backup.HookReceipt{ID: call.ID, State: "unknown", At: time.Now().UTC()}, err
	}
	h.mu.Unlock()

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Env = append(os.Environ(),
		"BLORA_BACKUP_HOOK_ID="+call.ID,
		"BLORA_BACKUP_ID="+call.BackupID,
		"BLORA_BACKUP_PHASE="+call.Phase,
		"BLORA_BACKUP_MODE="+call.Strategy.Mode,
		"BLORA_BACKUP_RESOURCE_KIND="+call.Resource.Kind,
		"BLORA_BACKUP_RESOURCE_ID="+call.Resource.ID,
		"BLORA_BACKUP_NODE_ID="+call.Resource.NodeID,
	)
	if err = cmd.Start(); err != nil {
		record.Receipt = backup.HookReceipt{ID: call.ID, State: "failed", Detail: "hook process could not start", At: time.Now().UTC()}
		_ = h.write(call.ID, record)
		return record.Receipt, err
	}
	record.PID = cmd.Process.Pid
	_ = h.write(call.ID, record)
	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		record.Receipt = backup.HookReceipt{ID: call.ID, State: "unknown", Detail: "hook context ended before outcome was recorded", At: time.Now().UTC()}
		_ = h.write(call.ID, record)
		return record.Receipt, backup.ErrInterrupted
	}
	if waitErr != nil {
		record.Receipt = backup.HookReceipt{ID: call.ID, State: "failed", Detail: "hook command returned a failure", At: time.Now().UTC()}
		_ = h.write(call.ID, record)
		return record.Receipt, waitErr
	}
	record.Receipt = backup.HookReceipt{ID: call.ID, State: "succeeded", At: time.Now().UTC()}
	if err = h.write(call.ID, record); err != nil {
		return backup.HookReceipt{ID: call.ID, State: "unknown", At: time.Now().UTC()}, backup.ErrInterrupted
	}
	return record.Receipt, nil
}

func (h *backupHookRunner) Reconcile(_ context.Context, call backup.HookCall) (backup.HookReceipt, error) {
	h.mu.Lock()
	record, err := h.read(call.ID)
	h.mu.Unlock()
	if err != nil {
		return backup.HookReceipt{ID: call.ID, State: "unknown", At: time.Now().UTC()}, backup.ErrInterrupted
	}
	if !sameHookCall(record.Call, call) {
		return backup.HookReceipt{ID: call.ID, State: "unknown", At: time.Now().UTC()}, backup.ErrConflict
	}
	if record.Receipt.State == "succeeded" {
		return record.Receipt, nil
	}
	if record.Receipt.State == "failed" {
		return record.Receipt, errors.New(record.Receipt.Detail)
	}
	// running/unknown is intentionally not re-executed. The backup manager will
	// expose an interrupted/unknown result and an operator can compensate it.
	if record.Receipt.State != "unknown" {
		record.Receipt.State = "unknown"
		record.Receipt.At = time.Now().UTC()
		_ = h.write(call.ID, record)
	}
	return record.Receipt, backup.ErrInterrupted
}

package daemon

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/runlog"
)

// PrepareCoreRestart freezes new task execution and drains accepted operations
// before a replacement daemon starts. It never stops a business instance,
// cancels a task, or sends application input. On success the caller MUST retain
// the gate through Close; on failure the gate is released automatically.
// Native/container terminal sessions are maintenance sessions and may close;
// they are not the managed instance's independent runtime or stdin/log helper.
func (d *Daemon) PrepareCoreRestart(ctx context.Context) (result error) {
	d.mu.Lock()
	if d.closing || d.coreUpdating {
		d.mu.Unlock()
		return errors.New("daemon is closing or another core update is in progress")
	}
	d.coreUpdating = true
	d.mu.Unlock()
	defer func() {
		if result != nil {
			d.AbortCoreRestart()
		}
	}()
	// A forgotten caller deadline must not freeze task acceptance indefinitely.
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("core update waiting for active operations; none were cancelled: %w", err)
		}
		d.mu.Lock()
		active, closing := len(d.active), d.closing
		d.mu.Unlock()
		if closing {
			return errors.New("daemon closed while preparing core update")
		}
		if active == 0 {
			break
		}
		select {
		case <-ctx.Done():
		case <-ticker.C:
		}
	}
	// Queued tasks contain no executed side effects and are safe to reconcile
	// after reconnecting. A nonterminal executed task absent from active could
	// have lost its durable outcome; never call that a safe restart boundary.
	pending, err := d.store.Pending(ctx)
	if err != nil {
		return fmt.Errorf("inspect accepted operations: %w", err)
	}
	for _, task := range pending {
		if task.State != model.Queued {
			return fmt.Errorf("core update blocked by unreconciled operation %s (%s)", task.ID, task.State)
		}
	}
	instanceIDs, err := d.store.RecordIDs(ctx, "instance")
	if err != nil {
		return fmt.Errorf("inspect managed instances: %w", err)
	}
	for _, instanceID := range instanceIDs {
		var instance model.Instance
		if _, err := d.store.Record(ctx, "instance", instanceID, &instance); err != nil {
			return fmt.Errorf("read instance %s: %w", instanceID, err)
		}
		if instance.RunID == "" {
			if instance.State == "RUNNING" || instance.State == "STARTING" || instance.State == "STOPPING" {
				return fmt.Errorf("instance %s has no recoverable run identity", instance.ID)
			}
			continue
		}
		record, observation, err := d.runtime.Recover(ctx, instance.RunID)
		if err != nil {
			return fmt.Errorf("cannot verify instance %s survives update: %w", instance.ID, err)
		}
		if record.InstanceID != instance.ID {
			return fmt.Errorf("instance %s run ownership does not match", instance.ID)
		}
		if observation.Exited {
			continue
		}
		if observation.State != "RUNNING" || record.Phase != "running" {
			return fmt.Errorf("instance %s run is not confirmed stable (%s/%s)", instance.ID, observation.State, record.Phase)
		}
		// Recover validates cgroup/PGID process birth or the Windows Job keeper,
		// and verifies Docker identity with the Engine. Native output AND stdin
		// must additionally belong to a live independent authenticated helper.
		if record.Backend != "docker" {
			capture, err := runlog.Open(filepath.Join(d.config.StateDir, "logs"), instance.RunID)
			if err != nil {
				return fmt.Errorf("instance %s independent log helper unavailable: %w", instance.ID, err)
			}
			status, err := capture.Status(ctx)
			if err != nil {
				return fmt.Errorf("instance %s independent log helper cannot be verified: %w", instance.ID, err)
			}
			if status.RunID != instance.RunID || !status.InputAvailable || (status.Phase != "running" && status.Phase != "output_complete") {
				return fmt.Errorf("instance %s independent stdin/output helper is not ready (%s)", instance.ID, status.Phase)
			}
		}
	}
	return nil
}

// AbortCoreRestart resumes task acceptance after a failed/cancelled update.
// The update coordinator must call it only for the preparation it owns.
func (d *Daemon) AbortCoreRestart() {
	d.mu.Lock()
	d.coreUpdating = false
	d.mu.Unlock()
}

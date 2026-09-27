package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"time"

	"blora.dev/panel/internal/backup"
	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/terminal"
)

func (d *Daemon) initBackups() error {
	var err error
	repository := d.config.BackupRoot
	if repository == "" {
		repository = filepath.Join(d.config.StateDir, "backups")
	}
	var hooks backup.HookRunner
	if len(d.config.BackupHookCommands) > 0 {
		hooks, err = newBackupHookRunner(filepath.Join(d.config.StateDir, "backup-hooks"), d.config.BackupHookCommands)
		if err != nil {
			return err
		}
	}
	d.backups, err = backup.New(repository, backup.Options{
		StateDir: filepath.Join(d.config.StateDir, "backup-state"),
		Hooks:    hooks,
		AuthorizeSnapshotRead: func(ctx context.Context, actor string, ref model.ResourceRef) error {
			return d.authorizeTerminal(ctx, actor, ref, "file.read")
		},
		AuthorizeRestoreWrite: func(ctx context.Context, actor string, ref model.ResourceRef) error {
			if err := d.authorizeTerminal(ctx, actor, ref, "backup.restore"); err != nil {
				return err
			}
			return d.authorizeTerminal(ctx, actor, ref, "file.write")
		},
	})
	return err
}

type backupRPCArgs struct {
	ID      string                `json:"id"`
	Offset  int                   `json:"offset"`
	Limit   int                   `json:"limit"`
	Request backup.RestoreRequest `json:"request"`
}

func (d *Daemon) backupRPC(ctx context.Context, r bridge.Request) (any, error) {
	if r.Resource.Kind != "instance" {
		return nil, terminal.ErrForbidden
	}
	var a backupRPCArgs
	if err := json.Unmarshal(r.Args, &a); err != nil {
		return nil, err
	}
	permission := "file.read"
	if r.Method == "backup.plan" || r.Method == "backup.plan.get" {
		permission = "backup.restore"
	}
	if err := d.authorizeTerminal(ctx, r.ActorID, r.Resource, permission); err != nil {
		return nil, err
	}
	switch r.Method {
	case "backup.list":
		items, err := d.backups.List(ctx)
		if err != nil {
			return nil, err
		}
		out := []backup.SnapshotInfo{}
		for _, item := range items {
			if item.Source == r.Resource {
				out = append(out, item)
			}
		}
		if a.Offset < 0 || a.Offset > len(out) {
			return nil, backup.ErrInvalid
		}
		if a.Limit < 1 || a.Limit > 100 {
			a.Limit = 100
		}
		end := a.Offset + a.Limit
		if end > len(out) {
			end = len(out)
		}
		next := -1
		if end < len(out) {
			next = end
		}
		return map[string]any{"items": out[a.Offset:end], "nextOffset": next, "repository": "node-local"}, nil
	case "backup.plan", "backup.plan.get":
		if err := d.authorizeTerminal(ctx, r.ActorID, r.Resource, "file.write"); err != nil {
			return nil, err
		}
		var plan backup.RestorePlan
		var err error
		if r.Method == "backup.plan" {
			if a.Request.OwnerID != r.ActorID || a.Request.Target != r.Resource {
				return nil, terminal.ErrForbidden
			}
			service, release, e := d.filesFor(r)
			if e != nil {
				return nil, e
			}
			defer release()
			plan, err = d.backups.PlanRestore(ctx, service, a.Request)
		} else {
			plan, err = d.backups.GetRestorePlan(ctx, a.ID)
		}
		if err != nil {
			return nil, err
		}
		if plan.Request.OwnerID != r.ActorID || plan.Request.Target != r.Resource {
			return nil, terminal.ErrForbidden
		}
		snapshot, err := d.backups.Inspect(ctx, plan.Request.BackupID)
		if err != nil {
			return nil, err
		}
		if err = d.authorizeTerminal(ctx, r.ActorID, snapshot.Source, "file.read"); err != nil {
			return nil, err
		}
		if a.Offset < 0 || a.Offset > len(plan.Entries) {
			return nil, backup.ErrInvalid
		}
		if a.Limit < 1 || a.Limit > 100 {
			a.Limit = 100
		}
		total := len(plan.Entries)
		end := a.Offset + a.Limit
		if end > total {
			end = total
		}
		plan.Entries = plan.Entries[a.Offset:end]
		next := -1
		if end < total {
			next = end
		}
		return map[string]any{"plan": plan, "entryCount": total, "nextOffset": next}, nil
	default:
		return nil, backup.ErrUnsupported
	}
}

func (d *Daemon) executeBackupTask(ctx context.Context, t model.Task) {
	var p backup.TaskPayload
	var result any
	err := json.Unmarshal(t.Payload, &p)
	if err == nil {
		service, release, e := d.filesFor(bridge.Request{Resource: t.Resource, Config: &p.Config})
		err = e
		if err == nil {
			defer release()
			switch t.Action {
			case "backup.create":
				if p.Create == nil || p.Create.OwnerID != t.ActorID || p.Create.Source != t.Resource {
					err = backup.ErrInvalid
					break
				}
				result, err = d.backups.Create(ctx, service, *p.Create)
			case "backup.restore":
				if p.Restore == nil || p.Restore.OwnerID != t.ActorID {
					err = backup.ErrInvalid
					break
				}
				plan, e := d.backups.GetRestorePlan(ctx, p.Restore.PlanID)
				if e != nil {
					err = e
					break
				}
				if plan.Request.Target != t.Resource || plan.Request.OwnerID != t.ActorID {
					err = terminal.ErrForbidden
					break
				}
				result, err = d.backups.Restore(ctx, service, *p.Restore)
			case "backup.prune":
				if p.Retention == nil || p.Retention.OwnerID != t.ActorID || p.Retention.Source != t.Resource {
					err = backup.ErrInvalid
					break
				}
				result, err = d.backups.Prune(ctx, *p.Retention, time.Now().UTC())
			default:
				err = backup.ErrUnsupported
			}
		}
	}
	state, phase, failure := model.Succeeded, "completed", ""
	if err != nil {
		state, phase, failure = model.Failed, "backup_failed", err.Error()
	}
	if errors.Is(err, context.Canceled) {
		state, phase = model.Cancelled, "cancelled_remaining_work"
	}
	if r, ok := result.(backup.Result); ok {
		if r.Stage != "" {
			phase = r.Stage
		}
		if r.Unknown || r.CleanupPending || errors.Is(err, backup.ErrInterrupted) {
			state = model.Interrupted
		}
	}
	b, e := json.Marshal(result)
	if e != nil {
		state, failure = model.Failed, e.Error()
	}
	_, _ = d.transition(context.Background(), t.ID, state, phase, b, failure)
}

package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"runtime"
	"strings"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/filesystem"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/monitor"
	"blora.dev/panel/internal/systeminfo"
	"blora.dev/panel/internal/terminal"
)

func (d *Daemon) executeAdapter(ctx context.Context, t model.Task) bool {
	if strings.HasPrefix(t.Action, "backup.") {
		d.executeBackupTask(ctx, t)
		return true
	}
	if model.IsContainerAction(t.Action) {
		d.executeContainerTask(ctx, t)
		return true
	}
	if !strings.HasPrefix(t.Action, "file.") && !strings.HasPrefix(t.Action, "terminal.") && !strings.HasPrefix(t.Action, "process.") && !strings.HasPrefix(t.Action, "system.service.") && !strings.HasPrefix(t.Action, "system.task.") && t.Action != "system.firewall.apply" && t.Action != "console.input" && t.Action != "extension.task" {
		return false
	}
	value, err := d.runAdapter(ctx, t)
	state := model.Succeeded
	phase := "completed"
	failure := ""
	if err != nil {
		state = model.Failed
		phase = "failed"
		failure = err.Error()
		var leaseErr *firewallLeaseError
		if errors.As(err, &leaseErr) && leaseErr.phase != "" {
			phase = leaseErr.phase
			if phase == "confirmation_commit_pending" {
				// The host confirmation is already durable. Leave the task
				// recoverable instead of replacing it with a false failure.
				_, _ = d.transition(context.Background(), t.ID, model.Running, phase, nil, failure)
				return true
			}
		}
		if errors.Is(err, context.Canceled) {
			state = model.Cancelled
			if leaseErr == nil || leaseErr.phase == "" {
				phase = "cancelled_remaining_work"
			}
		}
	}
	b, encodeErr := json.Marshal(value)
	if encodeErr != nil {
		state = model.Failed
		failure = encodeErr.Error()
	}
	_, _ = d.transition(context.Background(), t.ID, state, phase, b, failure)
	return true
}
func (d *Daemon) runAdapter(ctx context.Context, t model.Task) (any, error) {
	if t.Action == "extension.task" {
		return d.runExtension(ctx, t)
	}
	if t.Action == "console.input" {
		return d.consoleInput(ctx, t)
	}
	if t.Action == "process.terminate" {
		var p struct {
			PID        int    `json:"pid"`
			StartTicks uint64 `json:"startTicks"`
		}
		if err := json.Unmarshal(t.Payload, &p); err != nil {
			return nil, err
		}
		if err := d.authorizeTerminal(ctx, t.ActorID, model.ResourceRef{Kind: "node", ID: d.identity.NodeID, NodeID: d.identity.NodeID}, "host.manage"); err != nil {
			return nil, err
		}
		if err := monitor.Terminate(p.PID, p.StartTicks); err != nil {
			return nil, err
		}
		return map[string]any{"pid": p.PID, "startTicks": p.StartTicks}, nil
	}
	if strings.HasPrefix(t.Action, "system.service.") {
		name := strings.TrimPrefix(t.Action, "system.service.")
		if name == "" {
			return nil, errors.New("missing service action")
		}
		var p struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(t.Payload, &p); err != nil {
			return nil, err
		}
		if err := d.authorizeTerminal(ctx, t.ActorID, model.ResourceRef{Kind: "node", ID: d.identity.NodeID, NodeID: d.identity.NodeID}, "host.manage"); err != nil {
			return nil, err
		}
		if err := systeminfo.ServiceAction(ctx, p.Name, name); err != nil {
			return nil, err
		}
		return map[string]any{"name": p.Name, "action": name}, nil
	}
	if strings.HasPrefix(t.Action, "system.task.") {
		action := strings.TrimPrefix(t.Action, "system.task.")
		var p struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(t.Payload, &p); err != nil {
			return nil, err
		}
		if err := d.authorizeTerminal(ctx, t.ActorID, model.ResourceRef{Kind: "node", ID: d.identity.NodeID, NodeID: d.identity.NodeID}, "host.manage"); err != nil {
			return nil, err
		}
		if err := systeminfo.TaskAction(ctx, p.Name, action); err != nil {
			return nil, err
		}
		return map[string]any{"name": p.Name, "action": action}, nil
	}
	if t.Action == "system.firewall.apply" {
		return d.runFirewallApply(ctx, t)
	}
	if strings.HasPrefix(t.Action, "file.") {
		var p model.FileTaskPayload
		if err := json.Unmarshal(t.Payload, &p); err != nil {
			return nil, err
		}
		service, release, err := d.filesFor(bridge.Request{Resource: t.Resource, Config: &p.Config})
		if err != nil {
			return nil, err
		}
		defer release()
		switch t.Action {
		case "file.metadata":
			return service.Metadata(ctx, p.Path, p.Version, p.Mode, p.Modified)
		case "file.save", "file.upload":
			status, err := service.UploadStatus(ctx, p.UploadID)
			if err != nil {
				return nil, err
			}
			if status.Spec.OwnerID != t.ActorID || status.Spec.Path != p.Path || status.Spec.Hash != p.Hash || status.Spec.Total != p.Total {
				return nil, filesystem.ErrTransfer
			}
			return service.CommitUpload(ctx, p.UploadID)
		case "file.mkdir":
			return service.Mkdir(ctx, p.Path)
		case "file.copy":
			return service.Copy(ctx, p.Path, p.Target, p.Version, p.TargetVersion)
		case "file.move":
			return service.Move(ctx, p.Path, p.Target, p.Version, p.TargetVersion)
		case "file.delete":
			return service.DeleteChecked(ctx, p.Path, p.Version, p.ExpectedObjectID)
		case "file.restore":
			return service.Restore(ctx, p.TrashID, p.Target, p.TargetVersion)
		case "file.compress":
			return service.Compress(ctx, p.Path, p.Target, p.Version, p.TargetVersion)
		case "file.extract":
			return service.Extract(ctx, p.Path, p.Target, p.Version, filesystem.ExtractOptions{TargetVersion: p.TargetVersion, Overwrite: p.Overwrite})
		default:
			return nil, errors.New("unsupported file task")
		}
	}
	var p model.TerminalTaskPayload
	if err := json.Unmarshal(t.Payload, &p); err != nil {
		return nil, err
	}
	switch t.Action {
	case "terminal.create":
		if t.Resource.Kind == "node" {
			return d.createHostTerminal(ctx, t, p)
		}
		backend, runID := "native", ""
		dir := p.Config.Directory
		if dir == "" {
			dir = filepath.Join(d.config.StateDir, "instances", t.Resource.ID)
		}
		_, release, err := d.filesFor(bridge.Request{Resource: t.Resource, Config: &p.Config})
		if err != nil {
			return nil, err
		}
		release()
		command := []string{"/bin/sh"}
		if runtime.GOOS == "windows" {
			command = []string{"cmd.exe"}
		}
		if p.Config.Mode == "container" {
			var i model.Instance
			if _, err := d.store.Record(ctx, "instance", t.Resource.ID, &i); err != nil {
				return nil, err
			}
			backend, runID, dir = "docker", i.RunID, "/workspace"
			command = []string{"/bin/sh"}
			if _, err := d.resolveTerminalRun(ctx, t.Resource, runID); err != nil {
				return nil, err
			}
		}
		session, err := d.terminals.Create(ctx, terminal.CreateRequest{OwnerID: t.ActorID, Resource: t.Resource, RunID: runID, Backend: backend, Command: command, Directory: dir, Cols: p.Cols, Rows: p.Rows})
		return map[string]any{"session": session}, err
	case "terminal.close":
		session, err := d.terminals.Get(ctx, p.SessionID, t.ActorID)
		if err != nil {
			return nil, err
		}
		if session.Resource.Key() != t.Resource.Key() || (session.Backend == "docker-host" && session.OwnerID != t.ActorID) {
			return nil, terminal.ErrForbidden
		}
		return nil, d.terminals.CloseSession(ctx, p.SessionID, t.ActorID)
	default:
		return nil, errors.New("unsupported terminal task")
	}
}

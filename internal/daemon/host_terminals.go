package daemon

import (
	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/containers"
	"blora.dev/panel/internal/containerterm"
	"blora.dev/panel/internal/model"
	run "blora.dev/panel/internal/runtime"
	"blora.dev/panel/internal/terminal"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

type hostTerminalBinding struct {
	Actor  string               `json:"actor"`
	Target run.HostDockerTarget `json:"target"`
}

func (d *Daemon) hostTerminalTarget(ctx context.Context, r bridge.Request) (any, error) {
	if r.Resource.Kind != "node" || r.Resource.ID != d.identity.NodeID {
		return nil, terminal.ErrForbidden
	}
	var target containers.Target
	if err := json.Unmarshal(r.Args, &target); err != nil {
		return nil, err
	}
	checked, err := d.containers.InspectExecTarget(ctx, r.ActorID, target)
	if err != nil {
		return nil, err
	}
	return model.HostTerminalTarget{ContainerID: checked.Target.ID, CreatedAt: checked.Target.CreatedAt, StartedAt: checked.StartedAt}, nil
}

func (d *Daemon) createHostTerminal(ctx context.Context, t model.Task, p model.TerminalTaskPayload) (any, error) {
	if p.Host == nil || t.Resource.ID != d.identity.NodeID || t.Resource.NodeID != d.identity.NodeID {
		return nil, terminal.ErrForbidden
	}
	if d.hostExecutor == nil {
		return nil, terminal.ErrCapability
	}
	fixed := run.HostDockerTarget{ContainerID: p.Host.ContainerID, CreatedAt: p.Host.CreatedAt, StartedAt: p.Host.StartedAt}
	checked, err := d.containers.InspectExecTarget(ctx, t.ActorID, containers.Target{Kind: "container", ID: fixed.ContainerID, CreatedAt: fixed.CreatedAt})
	if err != nil {
		return nil, err
	}
	if checked.StartedAt != fixed.StartedAt {
		return nil, run.ErrUnknown
	}
	key := t.ActorID + ":" + fixed.Reference()
	var old hostTerminalBinding
	revision, err := d.store.Record(ctx, "host_terminal_binding", key, &old)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	binding := hostTerminalBinding{Actor: t.ActorID, Target: fixed}
	if err == nil && old != binding {
		return nil, terminal.ErrForbidden
	}
	if errors.Is(err, sql.ErrNoRows) {
		if _, err = d.store.PutRecord(ctx, "host_terminal_binding", key, revision, binding); err != nil {
			return nil, err
		}
	}
	session, err := d.terminals.Create(ctx, terminal.CreateRequest{ContainerID: fixed.ContainerID, OwnerID: t.ActorID, Resource: t.Resource, RunID: fixed.Reference(), Backend: "docker-host", Command: []string{"/bin/sh"}, Directory: "/", Cols: p.Cols, Rows: p.Rows})
	return map[string]any{"session": session}, err
}

func (d *Daemon) resolveHostTerminal(ctx context.Context, actor string, ref model.ResourceRef, runRef string) (containerterm.HostBinding, error) {
	if actor == "" || ref.Kind != "node" || ref.ID != d.identity.NodeID || ref.NodeID != ref.ID {
		return containerterm.HostBinding{}, terminal.ErrForbidden
	}
	var binding hostTerminalBinding
	if _, err := d.store.Record(ctx, "host_terminal_binding", actor+":"+runRef, &binding); err != nil {
		return containerterm.HostBinding{}, err
	}
	if binding.Actor != actor || binding.Target.Reference() != runRef {
		return containerterm.HostBinding{}, terminal.ErrForbidden
	}
	return containerterm.HostBinding{Target: binding.Target, Authorize: func(ctx context.Context) error { return d.authorizeTerminal(ctx, actor, ref, "host.manage") }}, nil
}

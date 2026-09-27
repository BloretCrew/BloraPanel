package containerterm

import (
	"context"

	"blora.dev/panel/internal/model"
	run "blora.dev/panel/internal/runtime"
	"blora.dev/panel/internal/terminal"
)

// HostBinding is loaded by the trusted Daemon from its actor-bound immutable
// terminal intent. No browser-provided reference is decoded into permissions.
type HostBinding struct {
	Target    run.HostDockerTarget
	Authorize func(context.Context) error
}
type HostResolver func(context.Context, string, model.ResourceRef, string) (HostBinding, error)
type HostAdapter struct {
	Executor *run.HostDockerExecutor
	Resolve  HostResolver
}

func (a HostAdapter) Spawn(ctx context.Context, s terminal.SpawnSpec) (terminal.Process, error) {
	if a.Executor == nil || a.Resolve == nil {
		return nil, terminal.ErrCapability
	}
	if s.OwnerID == "" || s.Resource.Kind != "node" || s.Resource.ID == "" || s.Resource.NodeID != s.Resource.ID || s.RunID == "" {
		return nil, terminal.ErrForbidden
	}
	binding, err := a.Resolve(ctx, s.OwnerID, s.Resource, s.RunID)
	if err != nil {
		return nil, err
	}
	if binding.Authorize == nil || binding.Target.Reference() != s.RunID {
		return nil, terminal.ErrForbidden
	}
	if err = binding.Authorize(ctx); err != nil {
		return nil, err
	}
	process, err := a.Executor.Start(ctx, binding.Target, run.DockerExecOptions{SessionID: s.SessionID, Command: s.Command, Directory: s.Directory, Environment: s.Environment, Cols: s.Cols, Rows: s.Rows}, binding.Authorize)
	if err != nil {
		return nil, err
	}
	return process, nil
}

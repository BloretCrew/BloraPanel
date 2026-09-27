// Package containerterm binds authorized terminal resources to runtime-owned
// container runs. Browser callers never choose a Docker container/exec ID.
package containerterm

import (
	"context"
	"errors"

	"blora.dev/panel/internal/model"
	run "blora.dev/panel/internal/runtime"
	"blora.dev/panel/internal/terminal"
)

type Resolver func(context.Context, model.ResourceRef, string) (run.Record, error)
type Adapter struct {
	Runtime *run.Manager
	Resolve Resolver
}

func (a Adapter) Spawn(ctx context.Context, s terminal.SpawnSpec) (terminal.Process, error) {
	if a.Runtime == nil || a.Resolve == nil {
		return nil, terminal.ErrCapability
	}
	if s.Resource.Kind != "instance" || s.Resource.ID == "" || s.Resource.NodeID == "" || s.RunID == "" {
		return nil, terminal.ErrForbidden
	}
	r, err := a.Resolve(ctx, s.Resource, s.RunID)
	if err != nil {
		return nil, err
	}
	if r.InstanceID != s.Resource.ID || r.RunID != s.RunID || r.Backend != "docker" || r.ContainerID == "" {
		return nil, errors.Join(terminal.ErrForbidden, run.ErrUnknown)
	}
	process, err := a.Runtime.StartDockerExec(ctx, r, run.DockerExecOptions{SessionID: s.SessionID, Command: s.Command, Directory: s.Directory, Environment: s.Environment, Cols: s.Cols, Rows: s.Rows})
	if err != nil {
		return nil, err
	}
	return process, nil
}

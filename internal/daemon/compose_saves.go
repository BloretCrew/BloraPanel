package daemon

import (
	"context"
	"encoding/json"
	"errors"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/containers"
	"blora.dev/panel/internal/model"
)

func (d *Daemon) composeRPC(ctx context.Context, r bridge.Request) (any, error) {
	if r.Resource.Kind != "node" || r.Resource.ID != d.identity.NodeID {
		return nil, containers.ErrForbidden
	}
	var in model.ComposeSaveRequest
	if err := json.Unmarshal(r.Args, &in); err != nil {
		return nil, err
	}
	switch r.Method {
	case "container.project.prepare":
		return d.containers.PrepareProjectSave(ctx, r.ActorID, in.SaveID, in.ProjectID, in.ExpectedRevision, in.Total, in.SHA256)
	case "container.project.status":
		return d.containers.ProjectSaveStatus(ctx, r.ActorID, in.SaveID)
	case "container.project.chunk":
		return d.containers.WriteProjectChunk(ctx, r.ActorID, in.SaveID, in.Offset, in.Data, in.SHA256)
	case "container.project.read":
		return d.containers.ReadProjectChunk(ctx, r.ActorID, in.ProjectID, in.Revision, in.Offset, in.Length)
	default:
		return nil, errors.New("unsupported project save operation")
	}
}

func (d *Daemon) executeComposeSave(ctx context.Context, t model.Task) {
	var in model.ComposeSaveRequest
	if err := json.Unmarshal(t.Payload, &in); err != nil {
		_, _ = d.transition(context.Background(), t.ID, model.Failed, "invalid_project_save", nil, err.Error())
		return
	}
	result, err := d.containers.CommitProjectSave(ctx, t.ActorID, in.SaveID)
	state, phase, diagnostic := model.Succeeded, "compose_source_saved", ""
	if err != nil {
		state, phase, diagnostic = model.Failed, "compose_save_failed", err.Error()
		if result.State == "COMMITTING" || errors.Is(err, containers.ErrUnknown) {
			state, phase = model.Interrupted, "compose_save_requires_reconciliation"
		}
	}
	b, _ := json.Marshal(result)
	_, _ = d.transition(context.Background(), t.ID, state, phase, b, diagnostic)
}

package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/containers"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/terminal"
)

func (d *Daemon) containerQuery(ctx context.Context, r bridge.Request) (any, error) {
	if r.Resource.Kind != "node" || r.Resource.ID != d.identity.NodeID {
		return nil, containers.ErrForbidden
	}
	var query containers.Query
	if err := json.Unmarshal(r.Args, &query); err != nil {
		return nil, err
	}
	if query.Limit < 1 {
		query.Limit = 50
	}
	if query.Limit > 100 {
		query.Limit = 100
	}
	snapshot, err := d.containers.Query(ctx, r.ActorID, query)
	if err != nil {
		return nil, err
	}
	if snapshot.Operation != nil {
		// The complete on-node journal is retained. This query is a bounded
		// recent-progress view; callers can inspect target resources separately.
		op := snapshot.Operation
		if len(op.Facts) > 0 {
			snapshot.Truncated = true
		}
		op.Facts = nil
		if len(op.Progress) > 24 {
			op.Progress = op.Progress[len(op.Progress)-24:]
			snapshot.Truncated = true
		}
	}
	b, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}
	if len(b) > 96<<10 {
		return nil, terminal.ErrLimit
	}
	return snapshot, nil
}

func (d *Daemon) executeContainerTask(ctx context.Context, t model.Task) {
	if t.Action == "compose.save" {
		d.executeComposeSave(ctx, t)
		return
	}
	var op containers.Operation
	if err := json.Unmarshal(t.Payload, &op); err != nil || op.Action != t.Action {
		_, _ = d.transition(context.Background(), t.ID, model.Failed, "invalid_container_operation", nil, "操作合同不匹配")
		return
	}
	op.TaskID = t.ID
	if op.Target.Kind == "container" && op.Target.ID != "" {
		// A platform instance must retain its run-generation and stop policy.
		// Its Docker Center entry routes lifecycle control to the instance app.
		ids, err := d.store.RecordIDs(ctx, "instance")
		if err == nil {
			for _, id := range ids {
				var i model.Instance
				if _, e := d.store.Record(ctx, "instance", id, &i); e != nil || i.RunID == "" {
					continue
				}
				r, e := d.runtime.Record(i.RunID)
				if e == nil && r.ContainerID == op.Target.ID {
					b, _ := json.Marshal(map[string]any{"instanceId": id, "runId": i.RunID, "managed": true})
					_, _ = d.transition(context.Background(), t.ID, model.Failed, "managed_instance_control_required", b, "此容器属于托管实例，请通过实例控制保留运行代次和退出确认")
					return
				}
			}
		}
		if err != nil {
			_, _ = d.transition(context.Background(), t.ID, model.Failed, "runtime_index_unavailable", nil, err.Error())
			return
		}
	}
	result, err := d.containers.Execute(ctx, t.ActorID, op, func(p containers.Progress) error {
		b, _ := json.Marshal(map[string]any{"phase": p.Phase, "message": p.Message, "sequence": p.Sequence, "current": p.Current, "total": p.Total, "layer": p.Layer})
		_, err := d.transition(ctx, t.ID, model.Running, p.Phase, b, "")
		return err
	})
	state := model.TaskState(result.State)
	if !state.Terminal() {
		state = model.Failed
	}
	if result.Unknown || errors.Is(err, containers.ErrUnknown) {
		state = model.Interrupted
	}
	if err != nil && state == model.Succeeded {
		state = model.Failed
	}
	diagnostic := result.Diagnostic
	if err != nil && diagnostic == "" {
		diagnostic = err.Error()
	}
	if len(diagnostic) > 4096 {
		diagnostic = diagnostic[:4096]
	}
	targets := result.Targets
	if len(targets) > 16 {
		targets = targets[:16]
	}
	b, _ := json.Marshal(map[string]any{"taskId": t.ID, "action": t.Action, "state": state, "phase": result.Phase, "changed": result.Changed, "unknown": result.Unknown, "targets": targets, "targetCount": len(result.Targets), "diagnostic": diagnostic})
	phase := result.Phase
	if phase == "" {
		phase = "container_operation_failed"
	}
	if _, e := d.transition(context.Background(), t.ID, state, phase, b, diagnostic); e != nil {
		slog.Error("container task result could not be persisted", "taskId", t.ID, "error", e)
	}
	if state == model.Succeeded && op.Action == "container.delete" && op.Target.Kind == "container" {
		if err := d.purgeContainerLogArchive(op.Target.ID); err != nil {
			slog.Warn("container log archive cleanup failed", "containerId", op.Target.ID, "error", err)
		}
	}
}

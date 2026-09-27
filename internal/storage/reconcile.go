package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"blora.dev/panel/internal/model"
)

// ReconcileTask records an observed daemon fact. Source revisions are separate
// from the Master index revision; duplicates and stale reports cannot roll back it.
func (s *Store) ReconcileTask(ctx context.Context, nodeID string, remote model.Task) (model.Task, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return model.Task{}, err
	}
	defer tx.Rollback()
	var local model.Task
	var b []byte
	if err := tx.QueryRowContext(ctx, "SELECT document FROM tasks WHERE id=?", remote.ID).Scan(&b); err != nil {
		return local, err
	}
	if err := json.Unmarshal(b, &local); err != nil {
		return local, err
	}
	if local.Resource.NodeID != nodeID || local.ActorID != remote.ActorID || local.RequestID != remote.RequestID || local.Digest != remote.Digest {
		return local, errors.New("remote task identity mismatch")
	}
	var old int64
	err = tx.QueryRowContext(ctx, "SELECT revision FROM records WHERE namespace='remote_task' AND id=?", remote.ID).Scan(&old)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return local, err
	}
	if (remote.Revision <= old && local.State != model.WaitingNode) || local.State.Terminal() {
		return local, nil
	}
	switch remote.State {
	case model.Queued, model.Running, model.WaitingNode, model.WaitingClient, model.CancelRequested, model.Succeeded, model.Failed, model.Cancelled, model.Interrupted:
	default:
		return local, errors.New("invalid remote task state")
	}
	if local.CancellationRequested && !remote.State.Terminal() {
		return local, nil
	}
	local.State = remote.State
	local.Phase = remote.Phase
	local.Result = remote.Result
	local.Error = remote.Error
	local.Revision++
	local.UpdatedAt = time.Now().UTC()
	b, err = json.Marshal(local)
	if err != nil {
		return local, err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE tasks SET state=?,revision=?,document=? WHERE id=?", local.State, local.Revision, b, local.ID); err != nil {
		return local, err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO task_events(task_id,revision,recorded_at,document) VALUES(?,?,?,?)", local.ID, local.Revision, local.UpdatedAt.UnixMilli(), b); err != nil {
		return local, err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO records(namespace,id,revision,document) VALUES('remote_task',?,?,?) ON CONFLICT(namespace,id) DO UPDATE SET revision=excluded.revision,document=excluded.document", local.ID, remote.Revision, []byte(`{}`)); err != nil {
		return local, err
	}
	if local.Action == "terminal.create" && local.State == model.Succeeded {
		var result struct {
			Session json.RawMessage `json:"session"`
		}
		if err := json.Unmarshal(local.Result, &result); err != nil {
			return local, err
		}
		var session struct {
			ID       string            `json:"sessionId"`
			Resource model.ResourceRef `json:"resource"`
		}
		if err := json.Unmarshal(result.Session, &session); err != nil {
			return local, err
		}
		if session.ID == "" || session.Resource != local.Resource {
			return local, errors.New("terminal result scope mismatch")
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO records(namespace,id,revision,document) VALUES('terminal',?,1,?) ON CONFLICT(namespace,id) DO NOTHING", session.ID, []byte(result.Session)); err != nil {
			return local, err
		}
	}
	if local.State.Terminal() {
		if _, err := tx.ExecContext(ctx, "INSERT INTO audit(actor_id,node_id,resource_key,action,request_id,result,recorded_at) VALUES(?,?,?,?,?,?,?)", local.ActorID, local.Resource.NodeID, local.Resource.Key(), local.Action, local.RequestID, string(local.State), local.UpdatedAt.UnixMilli()); err != nil {
			return local, err
		}
	}
	return local, tx.Commit()
}

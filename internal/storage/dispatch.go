package storage

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"blora.dev/panel/internal/model"
)

// mutateTask serializes an acceptance decision with the phase journal. Callbacks
// cannot perform external effects or access this store (the transaction owns it).
func (s *Store) mutateTask(ctx context.Context, id string, expected int64, change func(*model.Task) error) (model.Task, error) {
	var t model.Task
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return t, err
	}
	defer tx.Rollback()
	var b []byte
	if err := tx.QueryRowContext(ctx, "SELECT document FROM tasks WHERE id=?", id).Scan(&b); err != nil {
		return t, err
	}
	if err := json.Unmarshal(b, &t); err != nil {
		return t, err
	}
	if expected != 0 && expected != t.Revision {
		return t, ErrConflict
	}
	wasTerminal := t.State.Terminal()
	if err := change(&t); err != nil {
		return t, err
	}
	t.Revision++
	t.UpdatedAt = time.Now().UTC()
	b, err = json.Marshal(t)
	if err != nil {
		return t, err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE tasks SET state=?,revision=?,document=? WHERE id=?", t.State, t.Revision, b, t.ID); err != nil {
		return t, err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO task_events(task_id,revision,recorded_at,document) VALUES(?,?,?,?)", t.ID, t.Revision, t.UpdatedAt.UnixMilli(), b); err != nil {
		return t, err
	}
	if !wasTerminal && t.State.Terminal() {
		if _, err := tx.ExecContext(ctx, "INSERT INTO audit(actor_id,node_id,resource_key,action,request_id,result,recorded_at) VALUES(?,?,?,?,?,?,?)", t.ActorID, t.Resource.NodeID, t.Resource.Key(), t.Action, t.RequestID, string(t.State), t.UpdatedAt.UnixMilli()); err != nil {
			return t, err
		}
	}
	return t, tx.Commit()
}

// MarkDispatch is committed before a socket write. Even a failed write may have
// an unknown remote outcome; cancellation must then be reconciled by the daemon.
func (s *Store) MarkDispatch(ctx context.Context, id string, expected int64) (model.Task, error) {
	return s.mutateTask(ctx, id, expected, func(t *model.Task) error {
		if t.State.Terminal() {
			return ErrTransition
		}
		if t.DispatchedAt.IsZero() {
			t.DispatchedAt = time.Now().UTC()
		}
		if !t.CancellationRequested {
			t.Phase = "awaiting_node_acceptance"
		}
		return nil
	})
}

var errCancelAlreadyRequested = errors.New("task cancellation already requested")

// RequestCancel records the first cancellation request on the durable task
// receipt. A lost response can therefore be retried and resolved to the same
// state, including after the task has reached CANCELLED.
func (s *Store) RequestCancel(ctx context.Context, id string, requestIDs ...string) (model.Task, error) {
	requestID := ""
	if len(requestIDs) > 0 {
		requestID = requestIDs[0]
	}
	t, err := s.mutateTask(ctx, id, 0, func(t *model.Task) error {
		if t.CancellationRequested {
			return errCancelAlreadyRequested
		}
		if t.State.Terminal() {
			return ErrTransition
		}
		t.CancellationRequested = true
		t.CancellationRequestID = requestID
		if t.DispatchedAt.IsZero() && (t.State == model.Queued || t.State == model.WaitingNode) {
			t.State = model.Cancelled
			t.Phase = "cancelled_before_dispatch"
		} else {
			t.State = model.CancelRequested
			t.Phase = "awaiting_node_cancellation"
		}
		return nil
	})
	if errors.Is(err, errCancelAlreadyRequested) {
		return t, nil
	}
	if errors.Is(err, ErrTransition) && t.CancellationRequested {
		return t, nil
	}
	return t, err
}

// WaitingNode changes the management observation, never the runtime fact.
func (s *Store) WaitingNode(ctx context.Context, id string, expected int64) (model.Task, error) {
	return s.mutateTask(ctx, id, expected, func(t *model.Task) error {
		if t.State.Terminal() {
			return ErrTransition
		}
		if t.State == model.WaitingNode {
			return errors.New("already waiting for node")
		}
		t.State = model.WaitingNode
		if t.DispatchedAt.IsZero() {
			t.Phase = "waiting_node_before_dispatch"
		} else {
			t.Phase = "node_disconnected_outcome_unknown"
		}
		return nil
	})
}

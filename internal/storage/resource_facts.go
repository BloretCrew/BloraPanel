package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"blora.dev/panel/internal/model"
)

func (s *Store) ReconcileInstance(ctx context.Context, nodeID string, remote model.Instance) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var b []byte
	var local model.Instance
	if err := tx.QueryRowContext(ctx, "SELECT document FROM instances WHERE id=? AND node_id=?", remote.ID, nodeID).Scan(&b); err != nil {
		return err
	}
	if err := json.Unmarshal(b, &local); err != nil {
		return err
	}
	if local.Deleted {
		// A deleted resource is a tombstone.  Ignore stale daemon snapshots so
		// reconnect cannot resurrect it or make the management link fail.
		return nil
	}
	if remote.NodeID != nodeID {
		return errors.New("instance fact node mismatch")
	}
	var old int64
	err = tx.QueryRowContext(ctx, "SELECT revision FROM records WHERE namespace='instance_fact' AND id=?", remote.ID).Scan(&old)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if remote.Revision <= old {
		return nil
	}
	switch remote.State {
	case "STOPPED", "STARTING", "RUNNING", "STOPPING", "KILLING", "STOP_FAILED", "START_FAILED", "UNKNOWN", "RECOVERING":
	default:
		return errors.New("invalid observed instance state")
	}
	local.State = remote.State
	local.RunID = remote.RunID
	local.Revision++
	b, err = json.Marshal(local)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE instances SET document=? WHERE id=? AND node_id=?", b, remote.ID, nodeID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO records(namespace,id,revision,document) VALUES('instance_fact',?,?,?) ON CONFLICT(namespace,id) DO UPDATE SET revision=excluded.revision,document=excluded.document", remote.ID, remote.Revision, []byte(`{}`)); err != nil {
		return err
	}
	return tx.Commit()
}

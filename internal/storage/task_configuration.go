package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"blora.dev/panel/internal/model"
)

// A configuration edit and a task admission use the same SQLite boundary.
// Existing receipts are resolved first; fresh stale requests cannot cross it.
func checkTaskConfiguration(ctx context.Context, tx *sql.Tx, t model.Task) error {
	check := func(ref model.ResourceRef, config model.InstanceConfig) error {
		var b []byte
		err := tx.QueryRowContext(ctx, "SELECT document FROM instances WHERE id=?", ref.ID).Scan(&b)
		// A Daemon has no Master instance index; its immutable accepted task
		// is checked against current authority again before execution.
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		var i model.Instance
		if err := json.Unmarshal(b, &i); err != nil {
			return err
		}
		if i.Deleted {
			// A tombstoned identity cannot accept a task that raced with its
			// deletion.  Keep the durable request from reaching a daemon that
			// still has an old local snapshot.
			return ErrConflict
		}
		if i.NodeID != ref.NodeID || model.LaunchConfiguration(i.Config) != model.LaunchConfiguration(config) {
			return ErrConflict
		}
		return nil
	}
	switch {
	case t.Action == "instance.start" || t.Action == "instance.restart":
		var config model.InstanceConfig
		if err := json.Unmarshal(t.Payload, &config); err != nil {
			return err
		}
		return check(t.Resource, config)
	case strings.HasPrefix(t.Action, "file.") || strings.HasPrefix(t.Action, "backup.") || t.Action == "terminal.create":
		var p struct {
			Config model.InstanceConfig `json:"config"`
		}
		if err := json.Unmarshal(t.Payload, &p); err != nil {
			return err
		}
		return check(t.Resource, p.Config)
	case t.Action == "transfer.copy" || t.Action == "transfer.move":
		var p model.TransferTaskPayload
		if err := json.Unmarshal(t.Payload, &p); err != nil {
			return err
		}
		if err := check(p.SourceResource, p.SourceConfig); err != nil {
			return err
		}
		return check(p.TargetResource, p.TargetConfig)
	}
	return nil
}

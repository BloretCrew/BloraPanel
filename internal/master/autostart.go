package master

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
)

// One decision per daemon process startup, after its recovered inventory. A
// management reconnect or Master restart cannot restart a manually stopped run.
func (s *Server) enqueueAutostart(ctx context.Context, node model.Node) error {
	if node.StartupID == "" {
		return nil
	}
	items, err := s.store.Instances(ctx)
	if err != nil {
		return err
	}
	for _, i := range items {
		if i.NodeID != node.ID || !i.Config.Autostart || i.AutostartActor == "" || i.AutostartAfter == node.StartupID {
			continue
		}
		key := i.ID + ":" + node.StartupID
		var prior map[string]any
		if _, err := s.store.Record(ctx, "autostart", key, &prior); err == nil {
			continue
		} else if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		reason := "accepted"
		current, _, err := s.store.Node(ctx, node.ID)
		if err != nil || current.Maintenance {
			reason = "node_unavailable"
		}
		u, err := s.store.User(ctx, i.AutostartActor)
		if err != nil || !s.store.Allowed(ctx, u, instanceRef(i), "instance.start") || !s.store.Allowed(ctx, u, instanceRef(i), "instance.configure") {
			reason = "authorization_revoked"
		}
		if i.State == "RUNNING" {
			reason = "already_running"
		} else if i.State != "STOPPED" && i.State != "START_FAILED" {
			reason = "state_unconfirmed"
		}
		var pending int
		if err := s.store.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE resource_key=? AND state NOT IN ('SUCCEEDED','FAILED','CANCELLED','INTERRUPTED')", instanceRef(i).Key()).Scan(&pending); err != nil {
			return err
		}
		if pending > 0 {
			reason = "existing_operation"
		}
		taskID := ""
		if reason == "accepted" {
			payload, _ := json.Marshal(i.Config)
			task, _, err := s.store.Accept(ctx, model.Task{ActorID: i.AutostartActor, RequestID: "autostart-" + storage.Hash([]byte(key)), Resource: instanceRef(i), Action: "instance.start", Payload: payload})
			if err != nil {
				return err
			}
			taskID = task.ID
		}
		if _, err := s.store.PutRecord(ctx, "autostart", key, 0, map[string]any{"instanceId": i.ID, "startupId": node.StartupID, "reason": reason, "taskId": taskID, "at": time.Now().UTC()}); err != nil && !errors.Is(err, storage.ErrConflict) {
			return err
		}
	}
	return nil
}

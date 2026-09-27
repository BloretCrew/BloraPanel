package storage

import (
	"context"
	"encoding/json"
	"testing"

	"blora.dev/panel/internal/model"
)

func TestRemoteAndUndispatchedTerminalAuditIsAtomicAndNotReplayed(t *testing.T) {
	for _, mode := range []string{"remote", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			s, err := Open(":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			ctx := context.Background()
			task, _, err := s.Accept(ctx, model.Task{ActorID: "actor", RequestID: model.ID(), Resource: model.ResourceRef{Kind: "node", ID: "node", NodeID: "node"}, Action: "extension.task", Payload: json.RawMessage(`{"private":"not for audit"}`)})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.DB.ExecContext(ctx, `CREATE TRIGGER audit_failure BEFORE INSERT ON audit WHEN NEW.result <> 'accepted' BEGIN SELECT RAISE(ABORT,'injected audit failure'); END`); err != nil {
				t.Fatal(err)
			}
			complete := func() (model.Task, error) {
				if mode == "cancel" {
					return s.RequestCancel(ctx, task.ID)
				}
				remote := task
				remote.State = model.Succeeded
				remote.Revision = 5
				remote.Result = json.RawMessage(`{"private":"not for audit"}`)
				return s.ReconcileTask(ctx, "node", remote)
			}
			if _, err := complete(); err == nil {
				t.Fatal("audit failure did not abort terminal update")
			}
			stored, err := s.Task(ctx, task.ID)
			if err != nil {
				t.Fatal(err)
			}
			if stored.State != model.Queued || stored.Revision != task.Revision {
				t.Fatal("failed audit left a committed terminal task")
			}
			if _, err := s.DB.ExecContext(ctx, "DROP TRIGGER audit_failure"); err != nil {
				t.Fatal(err)
			}
			terminal, err := complete()
			if err != nil {
				t.Fatal(err)
			}
			_, _ = complete() // Duplicate remote report or already terminal cancellation.
			items, err := s.Audits(ctx, AuditFilter{ActorID: "actor", NodeID: "node", Limit: 10})
			if err != nil {
				t.Fatal(err)
			}
			if len(items) != 2 {
				t.Fatalf("expected acceptance and one terminal audit, got %d", len(items))
			}
			found := 0
			for _, item := range items {
				if item.Result == string(terminal.State) {
					found++
				}
				if item.RequestID != task.RequestID || item.Action != task.Action {
					t.Fatal("audit identity mismatch")
				}
			}
			if found != 1 {
				t.Fatal("terminal state missing or duplicated")
			}
		})
	}
}

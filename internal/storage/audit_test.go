package storage

import (
	"context"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
)

func TestAuditsFilterByActorNodeResourceActionAndTime(t *testing.T) {
	store, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	base := time.Now().UTC().Add(-time.Second)
	if err := store.Audit(ctx, "alice", model.ResourceRef{Kind: "node", ID: "node-a", NodeID: "node-a"}, "node.settings", "request-a", "applied"); err != nil {
		t.Fatal(err)
	}
	if err := store.Audit(ctx, "alice", model.ResourceRef{Kind: "instance", ID: "instance-a", NodeID: "node-a"}, "instance.start", "request-b", "accepted"); err != nil {
		t.Fatal(err)
	}
	if err := store.Audit(ctx, "bob", model.ResourceRef{Kind: "node", ID: "node-b", NodeID: "node-b"}, "node.settings", "request-c", "applied"); err != nil {
		t.Fatal(err)
	}
	items, err := store.Audits(ctx, AuditFilter{ActorID: "alice", NodeID: "node-a", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].ActorID != "alice" || items[0].NodeID != "node-a" {
		t.Fatalf("filtered records = %+v", items)
	}
	if items[0].RecordedAt.Before(base) {
		t.Fatalf("recorded timestamp did not use milliseconds: %v", items[0].RecordedAt)
	}
	items, err = store.Audits(ctx, AuditFilter{ResourceKey: "instance:instance-a", Action: "instance.start", From: base, To: time.Now().UTC().Add(time.Second), Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].RequestID != "request-b" {
		t.Fatalf("resource/action/time filter = %+v", items)
	}
	items, err = store.Audits(ctx, AuditFilter{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].RequestID != "request-b" {
		t.Fatalf("pagination = %+v", items)
	}
}

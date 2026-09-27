package storage

import (
	"blora.dev/panel/internal/model"
	"context"
	"path/filepath"
	"testing"
)

func TestTaskHistoryCrossesThousandAndDoesNotShiftOnInsertion(t *testing.T) {
	ctx := context.Background()
	store, err := Open(filepath.Join(t.TempDir(), "history.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	expected := map[string]bool{}
	for range 1005 {
		task, _, err := store.Accept(ctx, model.Task{ActorID: "alice", RequestID: model.ID(), Resource: model.ResourceRef{Kind: "instance", ID: "one"}, Action: "test", Payload: []byte(`{}`)})
		if err != nil {
			t.Fatal(err)
		}
		expected[task.ID] = true
	}
	before := int64(0)
	seen := map[string]bool{}
	for page := 0; ; page++ {
		rows, more, err := store.RootTaskPage(ctx, before, 100)
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) > 100 {
			t.Fatal("unbounded history page")
		}
		for _, row := range rows {
			if seen[row.Task.ID] || !expected[row.Task.ID] {
				t.Fatal("shifted/duplicated history")
			}
			seen[row.Task.ID] = true
			before = row.Cursor
		}
		if page == 0 {
			_, _, err = store.Accept(ctx, model.Task{ActorID: "alice", RequestID: model.ID(), Resource: model.ResourceRef{Kind: "instance", ID: "one"}, Action: "new-at-head", Payload: []byte(`{}`)})
			if err != nil {
				t.Fatal(err)
			}
		}
		if !more {
			break
		}
	}
	if len(seen) != len(expected) {
		t.Fatalf("missing history: %d of %d", len(seen), len(expected))
	}
}

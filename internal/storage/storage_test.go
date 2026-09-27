package storage

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"blora.dev/panel/internal/model"
)

func TestTaskCentreKeepsParentAfterLargeTransfer(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "tasks.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	parent, _, err := s.Accept(ctx, model.Task{ActorID: "alice", RequestID: "copy-tree", Resource: model.ResourceRef{Kind: "instance", ID: "target"}, Action: "transfer.copy", Payload: []byte(`{}`)})
	if err != nil {
		t.Fatal(err)
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	var last model.Task
	for range 1001 {
		payload, _ := json.Marshal(map[string]string{"transferParentId": parent.ID})
		last = model.Task{ID: model.ID(), ActorID: "alice", RequestID: model.ID(), Resource: parent.Resource, Action: "file.metadata", Payload: payload, State: model.Succeeded, Revision: 3}
		b, _ := json.Marshal(last)
		if _, err := tx.ExecContext(ctx, "INSERT INTO tasks(id,actor_id,request_id,resource_key,digest,state,revision,document) VALUES(?,?,?,?,?,?,?,?)", last.ID, last.ActorID, last.RequestID, last.Resource.Key(), "fixture", last.State, last.Revision, b); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	items, err := s.RootTasks(ctx)
	if err != nil || len(items) != 1 || items[0].ID != parent.ID {
		t.Fatalf("parent displaced: count=%d error=%v", len(items), err)
	}
	child, err := s.Task(ctx, last.ID)
	if err != nil || child.ID != last.ID {
		t.Fatal("internal step is no longer addressable")
	}
}

func TestDurableAcceptanceAndConcurrentRetry(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	in := model.Task{ActorID: "alice", RequestID: "same-request", Resource: model.ResourceRef{Kind: "instance", ID: "a"}, Action: "restart", Payload: []byte(`{"b":2,"a":1}`)}
	var wg sync.WaitGroup
	ids := make(chan string, 20)
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			task, _, err := s.Accept(ctx, in)
			if err != nil {
				t.Error(err)
				return
			}
			ids <- task.ID
		}()
	}
	wg.Wait()
	close(ids)
	var first string
	for id := range ids {
		if first == "" {
			first = id
		}
		if first != id {
			t.Fatal("duplicate task")
		}
	}
	in.Payload = []byte(`{"a":1,"b":2}`)
	if _, fresh, err := s.Accept(ctx, in); err != nil || fresh {
		t.Fatalf("canonical retry: %v %v", fresh, err)
	}
	in.Payload = []byte(`{"a":3}`)
	if _, _, err := s.Accept(ctx, in); !errors.Is(err, ErrRequestMismatch) {
		t.Fatalf("mutated request accepted: %v", err)
	}
	task, err := s.Task(ctx, first)
	if err != nil {
		t.Fatal(err)
	}
	task, err = s.UpdateTask(ctx, first, task.Revision, model.Running, "stopping", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	restored, err := s.Task(ctx, first)
	if err != nil {
		t.Fatal(err)
	}
	if restored.State != model.Running || restored.Revision != task.Revision {
		t.Fatal("acceptance lost on reopen")
	}
	if _, err := s.UpdateTask(ctx, first, 1, model.Succeeded, "done", nil, ""); !errors.Is(err, ErrConflict) {
		t.Fatal("stale writer accepted")
	}
	task, err = s.UpdateTask(ctx, first, restored.Revision, model.Interrupted, "reconcile_required", nil, "outcome requires observation")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateTask(ctx, first, task.Revision, model.Running, "replay", nil, ""); !errors.Is(err, ErrTransition) {
		t.Fatal("interrupted side effects replayed")
	}
	events, err := s.Events(ctx, 0, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 {
		t.Fatalf("expected durable phase log, got %d", len(events))
	}
}

func TestVersionedRecordAndCancellation(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.PutRecord(ctx, "draft", "x", 0, map[string]string{"body": "未保存"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.PutRecord(ctx, "draft", "x", 0, map[string]string{}); !errors.Is(err, ErrConflict) {
		t.Fatal("duplicate creation overwrites")
	}
	if _, err := s.PutRecord(ctx, "draft", "x", 2, map[string]string{}); !errors.Is(err, ErrConflict) {
		t.Fatal("stale record overwrites")
	}
	if CanTransition(model.CancelRequested, model.Running) || CanTransition(model.Succeeded, model.Running) || CanTransition(model.Queued, model.Succeeded) {
		t.Fatal("invalid transition permitted")
	}
}

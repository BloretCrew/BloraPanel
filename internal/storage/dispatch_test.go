package storage

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"blora.dev/panel/internal/model"
)

func TestCancellationFencesDispatchAndPreservesUnknownOutcome(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	newTask := func() model.Task {
		t.Helper()
		v, _, err := s.Accept(ctx, model.Task{ActorID: "alice", RequestID: model.ID(), Resource: model.ResourceRef{Kind: "instance", ID: "i", NodeID: "node"}, Action: "instance.restart", Payload: []byte(`{}`)})
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	first := newTask()
	cancelled, err := s.RequestCancel(ctx, first.ID, "cancel-first")
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.State != model.Cancelled {
		t.Fatal("unsent task must be cancelled locally")
	}
	if cancelled.CancellationRequestID != "cancel-first" {
		t.Fatal("cancellation request key was not persisted")
	}
	replayed, err := s.RequestCancel(ctx, first.ID, "cancel-first-retry")
	if err != nil || replayed.Revision != cancelled.Revision || replayed.State != model.Cancelled {
		t.Fatalf("cancellation retry did not return original receipt: %+v, %v", replayed, err)
	}
	if _, err := s.MarkDispatch(ctx, first.ID, first.Revision); !errors.Is(err, ErrConflict) {
		t.Fatal("cancelled task was dispatched from stale read")
	}
	second := newTask()
	dispatched, err := s.MarkDispatch(ctx, second.ID, second.Revision)
	if err != nil {
		t.Fatal(err)
	}
	if dispatched.DispatchedAt.IsZero() {
		t.Fatal("dispatch intent missing")
	}
	cancelling, err := s.RequestCancel(ctx, second.ID, "cancel-second")
	if err != nil {
		t.Fatal(err)
	}
	if cancelling.State != model.CancelRequested {
		t.Fatal("unknown node acceptance was falsely cancelled")
	}
	waiting, err := s.WaitingNode(ctx, second.ID, cancelling.Revision)
	if err != nil {
		t.Fatal(err)
	}
	if !waiting.CancellationRequested || waiting.State != model.WaitingNode {
		t.Fatal("disconnect erased cancellation intent")
	}
	remote := second
	remote.State = model.Running
	remote.Revision = 2
	merged, err := s.ReconcileTask(ctx, "node", remote)
	if err != nil {
		t.Fatal(err)
	}
	if merged.State != model.WaitingNode || !merged.CancellationRequested {
		t.Fatal("remote progress erased cancellation")
	}
	remote.State = model.Succeeded
	remote.Revision = 3
	merged, err = s.ReconcileTask(ctx, "node", remote)
	if err != nil {
		t.Fatal(err)
	}
	if merged.State != model.Succeeded {
		t.Fatal("completed side effect must be reported even when cancellation arrives late")
	}
}

func TestConcurrentCancellationReturnsOneDurableReceipt(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	task, _, err := s.Accept(ctx, model.Task{ActorID: "alice", RequestID: model.ID(), Resource: model.ResourceRef{Kind: "instance", ID: "i", NodeID: "node"}, Action: "instance.start", Payload: []byte(`{}`)})
	if err != nil {
		t.Fatal(err)
	}
	type result struct {
		task model.Task
		err  error
	}
	results := make(chan result, 2)
	for _, key := range []string{"cancel-a", "cancel-b"} {
		go func(key string) {
			v, e := s.RequestCancel(ctx, task.ID, key)
			results <- result{task: v, err: e}
		}(key)
	}
	var first model.Task
	for range 2 {
		got := <-results
		if got.err != nil {
			t.Fatal(got.err)
		}
		if first.ID == "" {
			first = got.task
		} else if got.task.Revision != first.Revision || got.task.CancellationRequestID != first.CancellationRequestID || got.task.State != first.State {
			t.Fatalf("concurrent cancellation receipts differ: first=%+v got=%+v", first, got.task)
		}
	}
	if first.State != model.Cancelled || first.CancellationRequestID == "" {
		t.Fatalf("invalid durable cancellation receipt: %+v", first)
	}
}

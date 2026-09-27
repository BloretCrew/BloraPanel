package storage

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"blora.dev/panel/internal/model"
)

func TestScheduledAdmissionAtomicallyRejectsOverlappingTriggers(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "tasks.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	ref := model.ResourceRef{Kind: "instance", ID: "shared", NodeID: "node"}
	var wg sync.WaitGroup
	accepted := make(chan model.Task, 2)
	failures := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			task, _, err := s.Accept(ctx, model.Task{ActorID: "owner", RequestID: model.ID(), Resource: ref, Action: "instance.restart", Payload: []byte(`{"scheduleId":"scheduled"}`)})
			if err == nil {
				accepted <- task
			} else {
				failures <- err
			}
		}()
	}
	wg.Wait()
	close(accepted)
	close(failures)
	if len(accepted) != 1 || len(failures) != 1 {
		t.Fatalf("accepted=%d rejected=%d", len(accepted), len(failures))
	}
	if err := <-failures; !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	task := <-accepted
	if same, fresh, err := s.Accept(ctx, task); err != nil || fresh || same.ID != task.ID {
		t.Fatalf("existing receipt rejected: %v", err)
	}
	// Unrelated resources remain admissible; the guard is not a global lock.
	ref.ID = "independent"
	if _, _, err := s.Accept(ctx, model.Task{ActorID: "owner", RequestID: model.ID(), Resource: ref, Action: task.Action, Payload: task.Payload}); err != nil {
		t.Fatal(err)
	}
}

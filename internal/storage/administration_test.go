package storage

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"blora.dev/panel/internal/model"
)

func TestUserChangesAtomicReceiptsAndLastAdministrator(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	a, err := s.CreateUser(ctx, model.User{Name: "admin-a", Admin: true}, []byte("unused"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.CreateUser(ctx, model.User{Name: "admin-b", Admin: true}, []byte("unused"))
	if err != nil {
		t.Fatal(err)
	}
	member, err := s.CreateUser(ctx, model.User{Name: "member"}, []byte("unused"))
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := s.NewSession(ctx, member.ID)
	if err != nil {
		t.Fatal(err)
	}
	disabled := true
	key := model.ID()
	patch := model.UserPatch{Disabled: &disabled, Revision: 1}
	first, err := s.UpdateUser(ctx, a.ID, member.ID, key, patch)
	if err != nil {
		t.Fatal(err)
	}
	again, err := s.UpdateUser(ctx, a.ID, member.ID, key, patch)
	if err != nil || again.Revision != first.Revision {
		t.Fatalf("metadata replay: %+v %v", again, err)
	}
	if _, _, err := s.Session(ctx, token); err == nil {
		t.Fatal("disabled account retained an authenticated session")
	}
	newName := "other"
	patch.Name = &newName
	if _, err := s.UpdateUser(ctx, a.ID, member.ID, key, patch); !errors.Is(err, ErrRequestMismatch) {
		t.Fatalf("different request accepted: %v", err)
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, u := range []model.User{a, b} {
		wg.Add(1)
		go func(u model.User) {
			defer wg.Done()
			_, err := s.UpdateUser(ctx, u.ID, u.ID, model.ID(), model.UserPatch{Disabled: &disabled, Revision: 1})
			results <- err
		}(u)
	}
	wg.Wait()
	close(results)
	passed, rejected := 0, 0
	for err := range results {
		if err == nil {
			passed++
		} else if errors.Is(err, ErrLastAdmin) {
			rejected++
		} else {
			t.Fatal(err)
		}
	}
	if passed != 1 || rejected != 1 {
		t.Fatalf("concurrent admin removal: successful=%d rejected=%d", passed, rejected)
	}
}

func TestEnrollmentMutationReplaysTicketByRequest(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "enrollment.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	admin, err := s.CreateUser(ctx, model.User{Name: "enrollment-admin", Admin: true}, []byte("unused"))
	if err != nil {
		t.Fatal(err)
	}
	first, err := s.EnrollmentMutation(ctx, admin.ID, "enrollment-key", "node-a")
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := s.EnrollmentMutation(ctx, admin.ID, "enrollment-key", "node-a")
	if err != nil || replayed != first {
		t.Fatalf("enrollment replay changed ticket: first=%q replay=%q err=%v", first, replayed, err)
	}
	if _, err := s.EnrollmentMutation(ctx, admin.ID, "enrollment-key", "node-b"); !errors.Is(err, ErrRequestMismatch) {
		t.Fatalf("different enrollment payload accepted: %v", err)
	}
	var count int
	if err := s.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM enrollments").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("request replay inserted %d enrollment tickets", count)
	}
}

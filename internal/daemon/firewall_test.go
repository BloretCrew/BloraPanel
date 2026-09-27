package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"blora.dev/panel/internal/containers"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
)

func newFirewallLeaseTestDaemon(t *testing.T) (*Daemon, func()) {
	t.Helper()
	store, err := storage.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	d := &Daemon{store: store, identity: identity{NodeID: "node-1"}, firewallLeases: map[string]*firewallLease{}}
	d.containers, err = containers.New(containers.Options{StateRoot: t.TempDir()})
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { d.containers.Close() })
	d.firewallAuthorize = func(context.Context, string) error { return nil }
	return d, func() { _ = store.Close() }
}

func firewallTask(id, actor string, until time.Time) model.Task {
	payload, _ := json.Marshal(map[string]any{"desired": []string{"443/tcp"}, "confirmUntil": until})
	return model.Task{ID: id, ActorID: actor, RequestID: "request-" + id, Resource: model.ResourceRef{Kind: "node", ID: "node-1", NodeID: "node-1"}, Action: "system.firewall.apply", Payload: payload}
}

func persistFirewallTask(t *testing.T, d *Daemon, task model.Task) {
	t.Helper()
	if _, _, err := d.store.Accept(context.Background(), task); err != nil {
		t.Fatal(err)
	}
}

func waitFirewallApplied(t *testing.T, d *Daemon, id string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		d.firewallMu.Lock()
		lease := d.firewallLeases[id]
		applied := lease != nil && lease.record.State == "applied"
		d.firewallMu.Unlock()
		if applied {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("firewall lease did not reach applied state")
}

func TestFirewallLeaseRequiresExplicitConfirmation(t *testing.T) {
	d, cleanup := newFirewallLeaseTestDaemon(t)
	defer cleanup()
	var mu sync.Mutex
	current := []string{"22/tcp"}
	d.firewallCurrent = func(context.Context) ([]string, error) {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), current...), nil
	}
	d.firewallApply = func(context.Context, []string) error {
		mu.Lock()
		current = []string{"443/tcp"}
		mu.Unlock()
		return nil
	}
	d.firewallRollback = func(_ context.Context, expected, previous []string) error {
		mu.Lock()
		defer mu.Unlock()
		if len(current) != 1 || current[0] != expected[0] {
			return errors.New("unexpected current firewall set")
		}
		current = append([]string(nil), previous...)
		return nil
	}
	task := firewallTask("fw-confirm", "actor", time.Now().Add(time.Second))
	persistFirewallTask(t, d, task)
	resultCh := make(chan error, 1)
	go func() { _, err := d.runFirewallApply(context.Background(), task); resultCh <- err }()
	waitFirewallApplied(t, d, task.ID)
	if _, err := d.confirmFirewallLease(context.Background(), task.ID, task.ActorID, "confirm-key"); err != nil {
		t.Fatal(err)
	}
	if value, err := d.confirmFirewallLease(context.Background(), task.ID, task.ActorID, "confirm-key"); err != nil || value == nil {
		t.Fatalf("same confirmation request was not replayable: value=%v err=%v", value, err)
	}
	if _, err := d.confirmFirewallLease(context.Background(), task.ID, task.ActorID, "different-key"); !errors.Is(err, errFirewallLeaseConflict) {
		t.Fatalf("different confirmation request was not rejected: %v", err)
	}
	if err := <-resultCh; err != nil {
		t.Fatalf("confirmed lease failed: %v", err)
	}
	stored, err := d.store.Task(context.Background(), task.ID)
	if err != nil || stored.Phase != "confirmed" || stored.State != model.Succeeded {
		t.Fatalf("confirmation phase was not persisted: %+v %v", stored, err)
	}
	var confirmationResult struct {
		ConfirmRequestID string `json:"confirmRequestId"`
	}
	if err := json.Unmarshal(stored.Result, &confirmationResult); err != nil || confirmationResult.ConfirmRequestID != "confirm-key" {
		t.Fatalf("confirmation request identity was not persisted in result: %+v %v", confirmationResult, err)
	}
	mu.Lock()
	if len(current) != 1 || current[0] != "443/tcp" {
		t.Fatalf("confirmed rules were rolled back: %v", current)
	}
	mu.Unlock()
	if _, err := d.store.Record(context.Background(), firewallLeaseNamespace, task.ID, &firewallLeaseRecord{}); err == nil {
		t.Fatal("confirmed lease record was not removed")
	}
}

func TestConfirmedFirewallRetainsRecoveryUntilTaskCommit(t *testing.T) {
	d, cleanup := newFirewallLeaseTestDaemon(t)
	defer cleanup()
	ctx := context.Background()
	task := firewallTask("fw-confirm-recovery", "actor", time.Now().Add(time.Minute))
	persistFirewallTask(t, d, task)
	if _, err := d.transition(ctx, task.ID, model.Running, "awaiting_confirmation", nil, ""); err != nil {
		t.Fatal(err)
	}
	record := firewallLeaseRecord{TaskID: task.ID, ActorID: task.ActorID, State: "confirmed", Desired: []string{"443/tcp"}}
	if _, err := d.store.PutRecord(ctx, firewallLeaseNamespace, task.ID, 0, record); err != nil {
		t.Fatal(err)
	}
	if _, err := d.store.DB.ExecContext(ctx, `CREATE TRIGGER fail_firewall_terminal BEFORE UPDATE ON tasks BEGIN SELECT RAISE(ABORT,'injected task write failure'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err := d.finishConfirmedFirewall(record); err == nil {
		t.Fatal("reported success after failed durable task write")
	}
	if _, err := d.store.Record(ctx, firewallLeaseNamespace, task.ID, &firewallLeaseRecord{}); err != nil {
		t.Fatal("lost recovery evidence", err)
	}
	if err := d.recover(ctx); err == nil {
		t.Fatal("startup discarded failed confirmation commit")
	}
	if task, err := d.store.Task(ctx, record.TaskID); err != nil || task.State != model.Running {
		t.Fatalf("lost recoverable task: %+v %v", task, err)
	}
	if _, err := d.store.DB.ExecContext(ctx, `DROP TRIGGER fail_firewall_terminal`); err != nil {
		t.Fatal(err)
	}
	d.firewallApply = func(context.Context, []string) error { t.Fatal("replayed host mutation"); return nil }
	if err := d.recoverFirewallLeases(ctx); err != nil {
		t.Fatal(err)
	}
	stored, err := d.store.Task(ctx, task.ID)
	if err != nil || stored.State != model.Succeeded {
		t.Fatalf("not recovered: %+v %v", stored, err)
	}
	if _, err := d.store.Record(ctx, firewallLeaseNamespace, task.ID, &firewallLeaseRecord{}); err == nil {
		t.Fatal("terminal record not pruned")
	}
}

func TestFirewallLeaseRollsBackOnCancellationAndExpiry(t *testing.T) {
	for _, tc := range []struct {
		name   string
		cancel bool
	}{
		{name: "cancel", cancel: true},
		{name: "expiry"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, cleanup := newFirewallLeaseTestDaemon(t)
			defer cleanup()
			var mu sync.Mutex
			current := []string{"22/tcp"}
			d.firewallCurrent = func(context.Context) ([]string, error) {
				mu.Lock()
				defer mu.Unlock()
				return append([]string(nil), current...), nil
			}
			d.firewallApply = func(context.Context, []string) error {
				mu.Lock()
				current = []string{"443/tcp"}
				mu.Unlock()
				return nil
			}
			d.firewallRollback = func(_ context.Context, expected, previous []string) error {
				mu.Lock()
				defer mu.Unlock()
				if len(current) != 1 || current[0] != expected[0] {
					return errors.New("unexpected current firewall set")
				}
				current = append([]string(nil), previous...)
				return nil
			}
			until := time.Now().Add(80 * time.Millisecond)
			ctx, cancel := context.WithCancel(context.Background())
			if tc.cancel {
				until = time.Now().Add(time.Second)
			}
			task := firewallTask("fw-"+tc.name, "actor", until)
			persistFirewallTask(t, d, task)
			result := make(chan error, 1)
			go func() { _, err := d.runFirewallApply(ctx, task); result <- err }()
			waitFirewallApplied(t, d, task.ID)
			if tc.cancel {
				cancel()
			}
			if err := <-result; err == nil {
				t.Fatal("unconfirmed lease unexpectedly succeeded")
			}
			stored, storeErr := d.store.Task(context.Background(), task.ID)
			if storeErr != nil || (tc.cancel && stored.Phase != "cancelled_rolled_back") || (!tc.cancel && stored.Phase != "confirmation_expired_rolled_back") {
				t.Fatalf("unexpected rollback phase: %+v %v", stored, storeErr)
			}
			mu.Lock()
			if len(current) != 1 || current[0] != "22/tcp" {
				t.Fatalf("rollback did not restore previous rules: %v", current)
			}
			mu.Unlock()
			cancel()
		})
	}
}

func TestDurableFirewallConfirmationCASIsIdempotent(t *testing.T) {
	d, cleanup := newFirewallLeaseTestDaemon(t)
	defer cleanup()
	ctx := context.Background()
	task := firewallTask("fw-confirm-cas", "actor", time.Now().Add(time.Minute))
	persistFirewallTask(t, d, task)
	if _, err := d.transition(ctx, task.ID, model.Running, "awaiting_confirmation", nil, ""); err != nil {
		t.Fatal(err)
	}
	record := firewallLeaseRecord{TaskID: task.ID, ActorID: task.ActorID, Previous: []string{"22/tcp"}, Desired: []string{"443/tcp"}, ConfirmUntil: time.Now().Add(time.Minute), State: "applied"}
	if _, err := d.store.PutRecord(ctx, firewallLeaseNamespace, task.ID, 0, record); err != nil {
		t.Fatal(err)
	}
	type outcome struct {
		key string
		err error
	}
	outcomes := make(chan outcome, 2)
	for _, key := range []string{"confirm-cas-a", "confirm-cas-b"} {
		go func(key string) {
			_, err := d.confirmFirewallLease(ctx, task.ID, task.ActorID, key)
			outcomes <- outcome{key: key, err: err}
		}(key)
	}
	var winner string
	for range 2 {
		got := <-outcomes
		if got.err == nil {
			if winner != "" {
				t.Fatal("two different confirmation keys both succeeded")
			}
			winner = got.key
		} else if !errors.Is(got.err, errFirewallLeaseConflict) {
			t.Fatalf("unexpected concurrent confirmation error: %v", got.err)
		}
	}
	if winner == "" {
		t.Fatal("no confirmation key succeeded")
	}
	if _, err := d.confirmFirewallLease(ctx, task.ID, task.ActorID, winner); err != nil {
		t.Fatalf("winning confirmation was not replayable: %v", err)
	}
	var saved firewallLeaseRecord
	if _, err := d.store.Record(ctx, firewallLeaseNamespace, task.ID, &saved); err != nil || saved.State != "confirmed" || saved.ConfirmRequestID != winner {
		t.Fatalf("unexpected durable confirmation record: %+v %v", saved, err)
	}
}

func TestRecoverFirewallLeaseRollsBackDurableMutation(t *testing.T) {
	d, cleanup := newFirewallLeaseTestDaemon(t)
	defer cleanup()
	current := []string{"443/tcp"}
	d.firewallCurrent = func(context.Context) ([]string, error) { return append([]string(nil), current...), nil }
	d.firewallRollback = func(_ context.Context, expected, previous []string) error {
		if len(current) != 1 || current[0] != expected[0] {
			return errors.New("unexpected current firewall set")
		}
		current = append([]string(nil), previous...)
		return nil
	}
	record := firewallLeaseRecord{TaskID: "fw-recover", ActorID: "actor", Previous: []string{"22/tcp"}, Desired: []string{"443/tcp"}, ConfirmUntil: time.Now().Add(time.Minute), State: "applied"}
	persistFirewallTask(t, d, firewallTask(record.TaskID, record.ActorID, record.ConfirmUntil))
	if _, err := d.store.PutRecord(context.Background(), firewallLeaseNamespace, record.TaskID, 0, record); err != nil {
		t.Fatal(err)
	}
	if err := d.recoverFirewallLeases(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(current) != 1 || current[0] != "22/tcp" {
		t.Fatalf("recovery did not rollback: %v", current)
	}
	if _, err := d.store.Record(context.Background(), firewallLeaseNamespace, record.TaskID, &firewallLeaseRecord{}); err == nil {
		t.Fatal("recovered lease record was not removed")
	}
}

func TestRollbackCompletionSurvivesTaskWriteFailure(t *testing.T) {
	d, cleanup := newFirewallLeaseTestDaemon(t)
	defer cleanup()
	ctx := context.Background()
	task := firewallTask("rollback-commit", "actor", time.Now().Add(time.Minute))
	persistFirewallTask(t, d, task)
	if _, err := d.transition(ctx, task.ID, model.Running, "awaiting_confirmation", nil, ""); err != nil {
		t.Fatal(err)
	}
	record := firewallLeaseRecord{TaskID: task.ID, ActorID: task.ActorID, State: "applied", Previous: []string{"22/tcp"}, Desired: []string{"443/tcp"}}
	revision, err := d.store.PutRecord(ctx, firewallLeaseNamespace, task.ID, 0, record)
	if err != nil {
		t.Fatal(err)
	}
	current := append([]string(nil), record.Desired...)
	d.firewallCurrent = func(context.Context) ([]string, error) { return current, nil }
	calls := 0
	d.firewallRollback = func(context.Context, []string, []string) error { calls++; current = record.Previous; return nil }
	if err := d.rollbackFirewallLease(&firewallLease{record: record, revision: revision}); err != nil {
		t.Fatal(err)
	}
	if _, err := d.store.DB.ExecContext(ctx, `CREATE TRIGGER fail_rollback_terminal BEFORE UPDATE ON tasks BEGIN SELECT RAISE(ABORT,'injected failure'); END`); err != nil {
		t.Fatal(err)
	}
	d.finishFirewallLeaseTask(task, &firewallLeaseError{phase: "confirmation_expired_rolled_back", err: errors.New("expired")})
	var saved firewallLeaseRecord
	if _, err := d.store.Record(ctx, firewallLeaseNamespace, task.ID, &saved); err != nil || saved.State != "rolled_back" {
		t.Fatalf("lost rollback receipt %+v %v", saved, err)
	}
	if err := d.recover(ctx); err == nil {
		t.Fatal("ignored terminal commit failure")
	}
	if _, err := d.store.DB.ExecContext(ctx, `DROP TRIGGER fail_rollback_terminal`); err != nil {
		t.Fatal(err)
	}
	if err := d.recover(ctx); err != nil {
		t.Fatal(err)
	}
	stored, err := d.store.Task(ctx, task.ID)
	if err != nil || stored.State != model.Failed || calls != 1 {
		t.Fatalf("unexpected recovery task=%+v calls=%d err=%v", stored, calls, err)
	}
}

func TestRecoverPreparedFirewallLeaseChecksHostBeforeDiscarding(t *testing.T) {
	for _, tc := range []struct {
		name         string
		state        string
		current      []string
		wantRollback bool
	}{
		{name: "not-applied", current: []string{"22/tcp"}},
		{name: "applied-before-crash", current: []string{"443/tcp"}, wantRollback: true},
		{name: "restored-before-receipt", state: "rollback_in_progress", current: []string{"22/tcp"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, cleanup := newFirewallLeaseTestDaemon(t)
			defer cleanup()
			current := append([]string(nil), tc.current...)
			rolledBack := false
			d.firewallCurrent = func(context.Context) ([]string, error) { return append([]string(nil), current...), nil }
			d.firewallRollback = func(_ context.Context, expected, previous []string) error {
				if !sameFirewallLeaseSet(current, expected) {
					return errors.New("unexpected current firewall set")
				}
				rolledBack = true
				current = append([]string(nil), previous...)
				return nil
			}
			record := firewallLeaseRecord{TaskID: "prepared-" + tc.name, ActorID: "actor", Previous: []string{"22/tcp"}, Desired: []string{"443/tcp"}, ConfirmUntil: time.Now().Add(time.Minute), State: "prepared"}
			if tc.state != "" {
				record.State = tc.state
			}
			persistFirewallTask(t, d, firewallTask(record.TaskID, record.ActorID, record.ConfirmUntil))
			if _, err := d.store.PutRecord(context.Background(), firewallLeaseNamespace, record.TaskID, 0, record); err != nil {
				t.Fatal(err)
			}
			if err := d.recoverFirewallLeases(context.Background()); err != nil {
				t.Fatal(err)
			}
			if !sameFirewallLeaseSet(current, []string{"22/tcp"}) || rolledBack != tc.wantRollback {
				t.Fatalf("rollback state mismatch current=%v rolledBack=%v", current, rolledBack)
			}
			_, err := d.store.Record(context.Background(), firewallLeaseNamespace, record.TaskID, &firewallLeaseRecord{})
			if err == nil {
				t.Fatal("prepared lease record was not removed")
			}
		})
	}
}

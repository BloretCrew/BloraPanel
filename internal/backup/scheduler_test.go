package backup

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
)

func scheduleTime(t *testing.T, value string) time.Time {
	t.Helper()
	v, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func scheduleFixture(t *testing.T) (*storage.Store, SchedulerOptions, *bool, *bool) {
	t.Helper()
	db, err := storage.Open(filepath.Join(t.TempDir(), "schedule.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	authorized, online := true, true
	options := SchedulerOptions{Authorize: func(context.Context, ScheduleSpec) error {
		if !authorized {
			return errors.New("permission revoked")
		}
		return nil
	}, NodeOnline: func(context.Context, string) bool { return online }, BuildTask: func(_ context.Context, s ScheduleSpec, slot time.Time) (model.Task, error) {
		payload, _ := json.Marshal(map[string]any{"source": "real-intent", "slot": slot})
		return model.Task{Resource: s.Resource, Action: s.Action, Payload: payload}, nil
	}}
	return db, options, &authorized, &online
}
func scheduleSpec() ScheduleSpec {
	return ScheduleSpec{ID: model.ID(), OwnerID: "owner", Resource: model.ResourceRef{Kind: "instance", ID: "instance", NodeID: "node"}, Action: "backup.create", Args: json.RawMessage(`{}`), Cron: "* * * * *", Timezone: "UTC", Misfire: "run_once", Overlap: "wait_one", Enabled: true}
}
func completeScheduledTask(t *testing.T, db *storage.Store, task model.Task) {
	t.Helper()
	var err error
	task, err = db.UpdateTask(context.Background(), task.ID, task.Revision, model.Running, "running", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.UpdateTask(context.Background(), task.ID, task.Revision, model.Succeeded, "completed", nil, ""); err != nil {
		t.Fatal(err)
	}
}

func TestSchedulerTimezoneDSTAndFiveFieldContract(t *testing.T) {
	spec := scheduleSpec()
	spec.Cron = "30 2 * * *"
	spec.Timezone = "America/New_York"
	calendar, err := parseSchedule(spec)
	if err != nil {
		t.Fatal(err)
	}
	if got := calendar.Next(scheduleTime(t, "2026-03-08T06:00:00Z")); !got.Equal(scheduleTime(t, "2026-03-09T06:30:00Z")) {
		t.Fatalf("nonexistent spring wall time was not skipped: %v", got)
	}
	spec.Cron = "30 1 * * *"
	calendar, err = parseSchedule(spec)
	if err != nil {
		t.Fatal(err)
	}
	first := calendar.Next(scheduleTime(t, "2026-11-01T04:00:00Z"))
	second := calendar.Next(first)
	if !first.Equal(scheduleTime(t, "2026-11-01T05:30:00Z")) || !second.Equal(scheduleTime(t, "2026-11-01T06:30:00Z")) {
		t.Fatalf("fall-back UTC slots %v %v", first, second)
	}
	for _, expression := range []string{"@every 1s", "* * * * * *", "CRON_TZ=UTC * * * * *"} {
		spec.Cron = expression
		if _, err = parseSchedule(spec); err == nil {
			t.Fatalf("non-five-field cron accepted: %s", expression)
		}
	}
}

func TestSchedulerOfflineCoalescingOverlapAndRevocation(t *testing.T) {
	ctx := context.Background()
	db, options, allowed, online := scheduleFixture(t)
	engine, err := NewScheduler(db, options)
	if err != nil {
		t.Fatal(err)
	}
	now := scheduleTime(t, "2026-09-09T09:00:00Z")
	plan, err := engine.Save(ctx, scheduleSpec(), 0, now)
	if err != nil {
		t.Fatal(err)
	}
	*online = false
	if fires, err := engine.Tick(ctx, now.Add(5*time.Minute+30*time.Second)); err != nil || len(fires) != 0 {
		t.Fatalf("offline tasks: %v %v", fires, err)
	}
	if _, err = engine.Tick(ctx, now.Add(8*time.Minute)); err != nil {
		t.Fatal(err)
	}
	*online = true
	fires, err := engine.Tick(ctx, now.Add(8*time.Minute+30*time.Second))
	if err != nil || len(fires) != 1 {
		t.Fatalf("single catch-up: %v %v", fires, err)
	}
	if !fires[0].Slot.Equal(now.Add(5 * time.Minute)) {
		t.Fatalf("pending durable occurrence changed: %v", fires[0].Slot)
	}
	if again, err := engine.Tick(ctx, now.Add(8*time.Minute+40*time.Second)); err != nil || len(again) != 0 {
		t.Fatalf("catch-up duplicated: %v %v", again, err)
	}
	if waiting, err := engine.Tick(ctx, now.Add(9*time.Minute)); err != nil || len(waiting) != 0 {
		t.Fatalf("resource overlapped: %v %v", waiting, err)
	}
	current, err := engine.Get(ctx, plan.Spec.ID)
	if err != nil || current.Pending.IsZero() || current.LastError != "resource_overlap" {
		t.Fatalf("%+v %v", current, err)
	}
	completeScheduledTask(t, db, fires[0].Task)
	*allowed = false
	if revoked, err := engine.Tick(ctx, now.Add(9*time.Minute+30*time.Second)); err != nil || len(revoked) != 0 {
		t.Fatalf("revoked plan executed: %v %v", revoked, err)
	}
	current, err = engine.Get(ctx, plan.Spec.ID)
	if err != nil || current.LastError != "authorization_denied" || !current.Pending.IsZero() {
		t.Fatalf("%+v %v", current, err)
	}
	tasks, err := db.Tasks(ctx)
	if err != nil || len(tasks) != 1 {
		t.Fatalf("unexpected task count %d %v", len(tasks), err)
	}
}

func TestSchedulerRestartReconcilesAcceptedTaskBeforeNewAuthorization(t *testing.T) {
	ctx := context.Background()
	db, options, allowed, online := scheduleFixture(t)
	engine, err := NewScheduler(db, options)
	if err != nil {
		t.Fatal(err)
	}
	now := scheduleTime(t, "2026-09-09T09:00:00Z")
	plan, err := engine.Save(ctx, scheduleSpec(), 0, now)
	if err != nil {
		t.Fatal(err)
	}
	fires, err := engine.Tick(ctx, now.Add(time.Minute))
	if err != nil || len(fires) != 1 {
		t.Fatalf("%v %v", fires, err)
	}
	// Restore only the checkpoints that can lag a successfully committed Accept.
	// The durable task remains real in SQLite, and must not be accepted twice.
	var lagging Schedule
	revision, err := db.Record(ctx, schedulesNS, plan.Spec.ID, &lagging)
	if err != nil {
		t.Fatal(err)
	}
	lagging.Pending = fires[0].Slot
	lagging.LastTaskID = ""
	lagging.LastSlot = time.Time{}
	if _, err = db.PutRecord(ctx, schedulesNS, plan.Spec.ID, revision, lagging); err != nil {
		t.Fatal(err)
	}
	var fire Fire
	fireRevision, err := db.Record(ctx, firesNS, fires[0].ID, &fire)
	if err != nil {
		t.Fatal(err)
	}
	fire.State = "prepared"
	if _, err = db.PutRecord(ctx, firesNS, fire.ID, fireRevision, fire); err != nil {
		t.Fatal(err)
	}
	*allowed = false
	*online = false
	var seq int
	var label, dbPath string
	if err = db.DB.QueryRowContext(ctx, "PRAGMA database_list").Scan(&seq, &label, &dbPath); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = storage.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	restarted, err := NewScheduler(db, options)
	if err != nil {
		t.Fatal(err)
	}
	recovered, err := restarted.Tick(ctx, now.Add(2*time.Minute))
	if err != nil || len(recovered) != 1 || recovered[0].Task.ID != fires[0].Task.ID {
		t.Fatalf("reconcile %v %v", recovered, err)
	}
	tasks, err := db.Tasks(ctx)
	if err != nil || len(tasks) != 1 {
		t.Fatalf("duplicate persisted task %d %v", len(tasks), err)
	}
}

func TestSchedulerSaveConflictCancellationAndScopedRetention(t *testing.T) {
	ctx := context.Background()
	db, options, _, _ := scheduleFixture(t)
	engine, err := NewScheduler(db, options)
	if err != nil {
		t.Fatal(err)
	}
	now := scheduleTime(t, "2026-09-09T09:00:00Z")
	plan, err := engine.Save(ctx, scheduleSpec(), 0, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = engine.Save(ctx, plan.Spec, 0, now); !errors.Is(err, ErrConflict) {
		t.Fatalf("lost update accepted: %v", err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err = engine.Tick(cancelled, now.Add(time.Minute)); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled tick: %v", err)
	}
	fires, err := engine.Tick(ctx, now.Add(time.Minute))
	if err != nil || len(fires) != 1 {
		t.Fatalf("%v %v", fires, err)
	}
	if n, err := engine.PruneFires(ctx, now.Add(time.Hour)); err != nil || n != 0 {
		t.Fatalf("active fire pruned %d %v", n, err)
	}
	completeScheduledTask(t, db, fires[0].Task)
	if _, err = db.PutRecord(ctx, "unrelated_records", model.ID(), 0, map[string]string{"keep": "yes"}); err != nil {
		t.Fatal(err)
	}
	if n, err := engine.PruneFires(ctx, now.Add(time.Hour)); err != nil || n != 1 {
		t.Fatalf("terminal fire retention %d %v", n, err)
	}
	ids, err := db.RecordIDs(ctx, "unrelated_records")
	if err != nil || len(ids) != 1 {
		t.Fatal("retention crossed namespace")
	}
}

func TestSchedulerMutationReceiptsSurviveStoreRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "schedule-mutation.db")
	db, err := storage.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	admin, err := db.CreateUser(ctx, model.User{Name: "schedule-admin", Admin: true}, []byte("unused"))
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	options := SchedulerOptions{Authorize: func(context.Context, ScheduleSpec) error { return nil }, NodeOnline: func(context.Context, string) bool { return true }, BuildTask: func(_ context.Context, s ScheduleSpec, slot time.Time) (model.Task, error) {
		return model.Task{Resource: s.Resource, Action: s.Action, Payload: []byte(`{"slot":"stable"}`)}, nil
	}}
	engine, err := NewScheduler(db, options)
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	spec := scheduleSpec()
	spec.OwnerID = admin.ID
	key := "schedule-save-restart"
	first, err := engine.SaveMutation(ctx, admin.ID, key, spec, 0, scheduleTime(t, "2026-09-09T09:00:00Z"))
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = storage.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	engine, err = NewScheduler(db, options)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := engine.SaveMutation(ctx, admin.ID, key, spec, 0, time.Now().UTC())
	if err != nil || replayed.ConfigRevision != first.ConfigRevision || !replayed.NextRun.Equal(first.NextRun) {
		t.Fatalf("save receipt lost after restart: first=%+v replay=%+v err=%v", first, replayed, err)
	}
	deleteKey := "schedule-delete-restart"
	if deleted, err := engine.DeleteMutation(ctx, admin.ID, deleteKey, spec.ID, first.ConfigRevision); err != nil || !deleted {
		t.Fatalf("delete mutation failed: %v", err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = storage.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	engine, err = NewScheduler(db, options)
	if err != nil {
		t.Fatal(err)
	}
	deleted, err := engine.DeleteMutation(ctx, admin.ID, deleteKey, spec.ID, first.ConfigRevision)
	if err != nil || !deleted {
		t.Fatalf("delete receipt lost after restart: deleted=%v err=%v", deleted, err)
	}
}

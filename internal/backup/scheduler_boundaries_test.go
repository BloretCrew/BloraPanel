package backup

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
)

func TestSchedulerSkipPoliciesAndBoundedLongDowntime(t *testing.T) {
	ctx := context.Background()
	db, options, _, online := scheduleFixture(t)
	engine, err := NewScheduler(db, options)
	if err != nil {
		t.Fatal(err)
	}
	now := scheduleTime(t, "2026-09-09T09:00:00Z")
	spec := scheduleSpec()
	spec.Misfire = "skip"
	spec.Overlap = "skip"
	plan, err := engine.Save(ctx, spec, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	*online = false
	if fires, err := engine.Tick(ctx, now.Add(time.Minute)); err != nil || len(fires) != 0 {
		t.Fatalf("offline skip %v %v", fires, err)
	}
	current, err := engine.Get(ctx, plan.Spec.ID)
	if err != nil || !current.Pending.IsZero() || current.Skipped != 1 {
		t.Fatalf("offline catchup retained %+v %v", current, err)
	}
	*online = true
	manual, _, err := db.Accept(ctx, model.Task{ActorID: "other-owner", RequestID: "manual", Action: spec.Action, Resource: spec.Resource, Payload: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatal(err)
	}
	if fires, err := engine.Tick(ctx, now.Add(2*time.Minute)); err != nil || len(fires) != 0 {
		t.Fatalf("overlap skip %v %v", fires, err)
	}
	current, err = engine.Get(ctx, plan.Spec.ID)
	if err != nil || !current.Pending.IsZero() || current.Skipped != 2 || current.LastError != "resource_overlap" {
		t.Fatalf("same resource cross-user overlap %+v %v", current, err)
	}
	completeScheduledTask(t, db, manual)
	if fires, err := engine.Tick(ctx, now.Add(40*24*time.Hour)); err != nil || len(fires) != 0 {
		t.Fatalf("unbounded catchup %v %v", fires, err)
	}
	current, err = engine.Get(ctx, plan.Spec.ID)
	if err != nil || !current.NextRun.After(now.Add(40*24*time.Hour)) || !current.MissedCountTruncated {
		t.Fatalf("long downtime cap %+v %v", current, err)
	}
}

func TestSchedulerConfigurationChangedDuringPrepareNeverAcceptsAndPrunesIntent(t *testing.T) {
	ctx := context.Background()
	db, options, _, _ := scheduleFixture(t)
	now := scheduleTime(t, "2026-09-09T09:00:00Z")
	var engine *Scheduler
	original := options.BuildTask
	options.BuildTask = func(ctx context.Context, spec ScheduleSpec, slot time.Time) (model.Task, error) {
		current, err := engine.Get(ctx, spec.ID)
		if err != nil {
			return model.Task{}, err
		}
		spec.Enabled = false
		if _, err = engine.Save(ctx, spec, current.ConfigRevision, slot); err != nil {
			return model.Task{}, err
		}
		return original(ctx, spec, slot)
	}
	var err error
	engine, err = NewScheduler(db, options)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = engine.Save(ctx, scheduleSpec(), 0, now); err != nil {
		t.Fatal(err)
	}
	if fires, err := engine.Tick(ctx, now.Add(time.Minute)); err != nil || len(fires) != 0 {
		t.Fatalf("disabled prepared task accepted %+v %v", fires, err)
	}
	if tasks, err := db.Tasks(ctx); err != nil || len(tasks) != 0 {
		t.Fatalf("disabled task persisted %+v %v", tasks, err)
	}
	if count, err := engine.PruneFires(ctx, now.Add(time.Hour)); err != nil || count != 1 {
		t.Fatalf("unaccepted obsolete intent quota leak %d %v", count, err)
	}
}

func TestSchedulerManualTaskDuringPreparationDefersSameStableFire(t *testing.T) {
	ctx := context.Background()
	db, options, _, _ := scheduleFixture(t)
	now := scheduleTime(t, "2026-09-09T09:00:00Z")
	original := options.BuildTask
	var manual model.Task
	builds := 0
	options.BuildTask = func(ctx context.Context, spec ScheduleSpec, slot time.Time) (model.Task, error) {
		builds++
		var err error
		manual, _, err = db.Accept(ctx, model.Task{ActorID: "another", RequestID: "racing-manual", Action: spec.Action, Resource: spec.Resource, Payload: json.RawMessage(`{}`)})
		if err != nil {
			return model.Task{}, err
		}
		return original(ctx, spec, slot)
	}
	engine, err := NewScheduler(db, options)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := engine.Save(ctx, scheduleSpec(), 0, now)
	if err != nil {
		t.Fatal(err)
	}
	if fires, err := engine.Tick(ctx, now.Add(time.Minute)); err != nil || len(fires) != 0 {
		t.Fatalf("manual preparation race %+v %v", fires, err)
	}
	current, err := engine.Get(ctx, plan.Spec.ID)
	if err != nil || current.Pending.IsZero() {
		t.Fatalf("deferred slot lost %+v %v", current, err)
	}
	completeScheduledTask(t, db, manual)
	fires, err := engine.Tick(ctx, now.Add(time.Minute+30*time.Second))
	if err != nil || len(fires) != 1 || builds != 1 || !fires[0].Slot.Equal(now.Add(time.Minute)) {
		t.Fatalf("prepared intent rebuilt/repeated %+v %v builds=%d", fires, err, builds)
	}
}

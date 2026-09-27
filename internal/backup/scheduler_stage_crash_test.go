package backup

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
)

const (
	schedulerStageCrashBaseEnv  = "BLORA_SCHEDULE_STAGE_CRASH_BASE"
	schedulerStageCrashPhaseEnv = "BLORA_SCHEDULE_STAGE_CRASH_PHASE"
	schedulerStageReadyPrefix   = "SCHEDULER_STAGE_CRASH_READY "
)

type schedulerStageCrashReady struct {
	Phase      string    `json:"phase"`
	ScheduleID string    `json:"scheduleId"`
	FireID     string    `json:"fireId,omitempty"`
	TaskID     string    `json:"taskId,omitempty"`
	RequestID  string    `json:"requestId,omitempty"`
	State      string    `json:"state"`
	Slot       time.Time `json:"slot"`
}

// TestSchedulerProcessCrashReconcilesDurableOccurrenceStages exercises the
// actual Tick path before task acceptance: once with only the pending slot
// persisted, and once after a prepared fire is persisted but before Accept.
func TestSchedulerProcessCrashReconcilesDurableOccurrenceStages(t *testing.T) {
	if base := os.Getenv(schedulerStageCrashBaseEnv); base != "" {
		runSchedulerStageCrashHelper(t, base, os.Getenv(schedulerStageCrashPhaseEnv))
		return
	}
	for _, phase := range []string{"pending", "prepared"} {
		t.Run(phase, func(t *testing.T) { verifySchedulerStageCrashRecovery(t, phase) })
	}
}

func verifySchedulerStageCrashRecovery(t *testing.T, phase string) {
	base := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestSchedulerProcessCrashReconcilesDurableOccurrenceStages$", "-test.count=1")
	cmd.Env = append(os.Environ(), schedulerStageCrashBaseEnv+"="+base, schedulerStageCrashPhaseEnv+"="+phase)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if cmd.ProcessState == nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	})

	readyCh := make(chan schedulerStageCrashReady, 1)
	scanDone := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, schedulerStageReadyPrefix) {
				continue
			}
			var ready schedulerStageCrashReady
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, schedulerStageReadyPrefix)), &ready); err != nil {
				scanDone <- err
				return
			}
			readyCh <- ready
		}
		scanDone <- scanner.Err()
	}()

	var ready schedulerStageCrashReady
	select {
	case ready = <-readyCh:
	case scanErr := <-scanDone:
		waitErr := cmd.Wait()
		t.Fatalf("scheduler helper exited before crash stage %q: scan=%v wait=%v stderr=%s", phase, scanErr, waitErr, stderr.String())
	case <-time.After(15 * time.Second):
		t.Fatalf("scheduler helper did not reach crash stage %q; stderr=%s", phase, stderr.String())
	}
	if ready.Phase != phase || ready.ScheduleID == "" || ready.Slot.IsZero() {
		t.Fatalf("incomplete durable stage receipt: %+v", ready)
	}
	if phase == "pending" && ready.State != "pending" {
		t.Fatalf("wrong pre-build crash state: %+v", ready)
	}
	if phase == "prepared" && (ready.State != "prepared" || ready.FireID == "" || ready.TaskID == "" || ready.RequestID == "") {
		t.Fatalf("incomplete prepared-fire receipt: %+v", ready)
	}
	if err = cmd.Process.Kill(); err != nil {
		t.Fatalf("kill scheduler at %s stage: %v", phase, err)
	}
	if err = cmd.Wait(); err == nil {
		t.Fatal("scheduler helper exited normally instead of being killed")
	} else {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("wait for killed scheduler helper: %v", err)
		}
	}
	if scanErr := <-scanDone; scanErr != nil {
		t.Fatalf("read scheduler helper output after process kill: %v", scanErr)
	}

	db, err := storage.Open(filepath.Join(base, "scheduler.db"))
	if err != nil {
		t.Fatalf("reopen scheduler database after %s crash: %v", phase, err)
	}
	defer db.Close()
	ctx := context.Background()
	var schedule Schedule
	if revision, recordErr := db.Record(ctx, schedulesNS, ready.ScheduleID, &schedule); recordErr != nil || revision == 0 || !schedule.Pending.Equal(ready.Slot) || !schedule.LastSlot.IsZero() || schedule.LastTaskID != "" {
		t.Fatalf("pending checkpoint did not survive %s crash: revision=%d schedule=%+v err=%v", phase, revision, schedule, recordErr)
	}
	fireIDs, err := db.RecordIDs(ctx, firesNS)
	if err != nil {
		t.Fatal(err)
	}
	var prepared Fire
	switch phase {
	case "pending":
		if len(fireIDs) != 0 {
			t.Fatalf("fire unexpectedly exists before task construction: %v", fireIDs)
		}
	case "prepared":
		if len(fireIDs) != 1 || fireIDs[0] != ready.FireID {
			t.Fatalf("prepared fire not durably bound: %v want=%s", fireIDs, ready.FireID)
		}
		if revision, recordErr := db.Record(ctx, firesNS, ready.FireID, &prepared); recordErr != nil || revision == 0 || prepared.State != "prepared" || prepared.Task.ID != ready.TaskID || prepared.Task.RequestID != ready.RequestID {
			t.Fatalf("prepared fire changed after kill: revision=%d fire=%+v err=%v", revision, prepared, recordErr)
		}
		if _, recordErr := db.TaskByRequest(ctx, prepared.Task.ActorID, prepared.Task.RequestID); !errors.Is(recordErr, sql.ErrNoRows) {
			t.Fatalf("task was accepted before crash point: %v", recordErr)
		}
	}
	if tasks, taskErr := db.Tasks(ctx); taskErr != nil || len(tasks) != 0 {
		t.Fatalf("unaccepted task survived %s stage: tasks=%+v err=%v", phase, tasks, taskErr)
	}

	buildCalls := 0
	engine, err := NewScheduler(db, SchedulerOptions{
		Authorize:  func(context.Context, ScheduleSpec) error { return nil },
		NodeOnline: func(context.Context, string) bool { return true },
		BuildTask: func(_ context.Context, spec ScheduleSpec, _ time.Time) (model.Task, error) {
			buildCalls++
			return model.Task{Resource: spec.Resource, Action: spec.Action, Payload: json.RawMessage(`{"recovered":true}`)}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	fires, err := engine.Tick(ctx, ready.Slot.Add(30*time.Second))
	if err != nil || len(fires) != 1 || fires[0].State != "accepted" || !fires[0].Slot.Equal(ready.Slot) {
		t.Fatalf("reconcile %s occurrence: fires=%+v err=%v", phase, fires, err)
	}
	if phase == "pending" && buildCalls != 1 {
		t.Fatalf("pending occurrence build calls=%d, want exactly one", buildCalls)
	}
	if phase == "prepared" && (buildCalls != 0 || fires[0].ID != ready.FireID || fires[0].Task.ID != ready.TaskID || fires[0].Task.RequestID != ready.RequestID) {
		t.Fatalf("prepared task identity was rebuilt: calls=%d ready=%+v fire=%+v", buildCalls, ready, fires[0])
	}
	if replay, tickErr := engine.Tick(ctx, ready.Slot.Add(45*time.Second)); tickErr != nil || len(replay) != 0 {
		t.Fatalf("same occurrence replayed after reconciliation: fires=%+v err=%v", replay, tickErr)
	}

	var accepted Fire
	if _, recordErr := db.Record(ctx, firesNS, fires[0].ID, &accepted); recordErr != nil || accepted.State != "accepted" || accepted.Task.ID != fires[0].Task.ID {
		t.Fatalf("accepted fire receipt missing: %+v %v", accepted, recordErr)
	}
	var current Schedule
	if _, recordErr := db.Record(ctx, schedulesNS, ready.ScheduleID, &current); recordErr != nil || !current.Pending.IsZero() || current.LastTaskID != fires[0].Task.ID || !current.LastSlot.Equal(ready.Slot) {
		t.Fatalf("schedule checkpoint not finalized: %+v %v", current, recordErr)
	}
	if tasks, taskErr := db.Tasks(ctx); taskErr != nil || len(tasks) != 1 || tasks[0].ID != fires[0].Task.ID || tasks[0].RequestID != fires[0].Task.RequestID {
		t.Fatalf("occurrence produced duplicate/missing tasks: %+v %v", tasks, taskErr)
	}
}

func runSchedulerStageCrashHelper(t *testing.T, base, phase string) {
	db, err := storage.Open(filepath.Join(base, "scheduler.db"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	authorizeCalls := 0
	options := SchedulerOptions{
		Authorize: func(ctx context.Context, spec ScheduleSpec) error {
			authorizeCalls++
			if phase == "prepared" && authorizeCalls == 3 {
				ids, listErr := db.RecordIDs(ctx, firesNS)
				if listErr != nil || len(ids) != 1 {
					t.Fatalf("prepared fire list at final authorization: ids=%v err=%v", ids, listErr)
				}
				var fire Fire
				if _, recordErr := db.Record(ctx, firesNS, ids[0], &fire); recordErr != nil || fire.State != "prepared" {
					t.Fatalf("expected durable prepared fire before accept: %+v err=%v", fire, recordErr)
				}
				notifySchedulerStageAndBlock(t, schedulerStageCrashReady{Phase: phase, ScheduleID: fire.ScheduleID, FireID: fire.ID, TaskID: fire.Task.ID, RequestID: fire.Task.RequestID, State: fire.State, Slot: fire.Slot})
			}
			return nil
		},
		NodeOnline: func(context.Context, string) bool { return true },
		BuildTask: func(_ context.Context, spec ScheduleSpec, slot time.Time) (model.Task, error) {
			if phase == "pending" {
				var current Schedule
				if _, recordErr := db.Record(ctx, schedulesNS, spec.ID, &current); recordErr != nil || !current.Pending.Equal(slot) {
					return model.Task{}, fmt.Errorf("pending slot not durable before task build: schedule=%+v err=%v", current, recordErr)
				}
				notifySchedulerStageAndBlock(t, schedulerStageCrashReady{Phase: phase, ScheduleID: spec.ID, State: "pending", Slot: slot})
			}
			return model.Task{Resource: spec.Resource, Action: spec.Action, Payload: json.RawMessage(`{"stage":"prepared-crash"}`)}, nil
		},
	}
	engine, err := NewScheduler(db, options)
	if err != nil {
		t.Fatal(err)
	}
	now := scheduleTime(t, "2026-09-21T09:00:01Z")
	plan, err := engine.Save(ctx, scheduleSpec(), 0, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = engine.Tick(ctx, plan.NextRun); err == nil {
		t.Fatal("scheduler returned instead of blocking at the requested crash stage")
	}
}

func notifySchedulerStageAndBlock(t *testing.T, ready schedulerStageCrashReady) {
	t.Helper()
	body, err := json.Marshal(ready)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = fmt.Fprintf(os.Stdout, "%s%s\n", schedulerStageReadyPrefix, body); err != nil {
		t.Fatal(err)
	}
	select {}
}

package backup

import (
	"bufio"
	"context"
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

const schedulerCrashFixtureEnv = "BLORA_SCHEDULE_CRASH_FIXTURE"

type schedulerCrashReady struct {
	FireID     string    `json:"fireId"`
	ScheduleID string    `json:"scheduleId"`
	TaskID     string    `json:"taskId"`
	RequestID  string    `json:"requestId"`
	Slot       time.Time `json:"slot"`
}

// TestSchedulerProcessCrashKeepsAcceptedOccurrence kills a separate test
// process after it durably accepts a scheduled task. Reopening the database
// must retain the exact task receipt, avoid replaying that occurrence, and
// still re-authorize a later occurrence.
func TestSchedulerProcessCrashKeepsAcceptedOccurrence(t *testing.T) {
	if base := os.Getenv(schedulerCrashFixtureEnv); base != "" {
		runSchedulerCrashHelper(t, base)
		return
	}

	base := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestSchedulerProcessCrashKeepsAcceptedOccurrence$", "-test.count=1")
	cmd.Env = append(os.Environ(), schedulerCrashFixtureEnv+"="+base)
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

	readyCh := make(chan schedulerCrashReady, 1)
	scanDone := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			const prefix = "SCHEDULER_CRASH_READY "
			if !strings.HasPrefix(line, prefix) {
				continue
			}
			var ready schedulerCrashReady
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, prefix)), &ready); err != nil {
				scanDone <- err
				return
			}
			readyCh <- ready
		}
		scanDone <- scanner.Err()
	}()

	var ready schedulerCrashReady
	select {
	case ready = <-readyCh:
	case scanErr := <-scanDone:
		waitErr := cmd.Wait()
		t.Fatalf("scheduler helper exited before accepting an occurrence: scan=%v wait=%v stderr=%s", scanErr, waitErr, stderr.String())
	case <-time.After(15 * time.Second):
		t.Fatalf("scheduler helper did not accept an occurrence; stderr=%s", stderr.String())
	}
	if ready.FireID == "" || ready.ScheduleID == "" || ready.TaskID == "" || ready.RequestID == "" || ready.Slot.IsZero() {
		t.Fatalf("incomplete helper receipt: %+v", ready)
	}
	if err = cmd.Process.Kill(); err != nil {
		t.Fatalf("kill scheduler process: %v", err)
	}
	if err = cmd.Wait(); err == nil {
		t.Fatal("scheduler helper exited normally instead of being killed")
	} else {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("scheduler helper wait: %v", err)
		}
	}
	if scanErr := <-scanDone; scanErr != nil {
		t.Fatalf("read scheduler helper output after process kill: %v", scanErr)
	}

	db, err := storage.Open(filepath.Join(base, "scheduler.db"))
	if err != nil {
		t.Fatalf("reopen database after process kill: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	built := false
	engine, err := NewScheduler(db, SchedulerOptions{
		Authorize:  func(context.Context, ScheduleSpec) error { return errors.New("permission revoked") },
		NodeOnline: func(context.Context, string) bool { return false },
		BuildTask: func(context.Context, ScheduleSpec, time.Time) (model.Task, error) {
			built = true
			return model.Task{}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var fire Fire
	if revision, recordErr := db.Record(ctx, firesNS, ready.FireID, &fire); recordErr != nil || revision == 0 || fire.State != "accepted" || fire.Task.ID != ready.TaskID || fire.Task.RequestID != ready.RequestID {
		t.Fatalf("accepted fire did not survive process kill: revision=%d fire=%+v err=%v", revision, fire, recordErr)
	}
	schedule, err := engine.Get(ctx, ready.ScheduleID)
	if err != nil || schedule.LastTaskID != ready.TaskID || !schedule.LastSlot.Equal(ready.Slot) || !schedule.Pending.IsZero() {
		t.Fatalf("schedule checkpoint did not survive process kill: schedule=%+v err=%v", schedule, err)
	}

	if fires, tickErr := engine.Tick(ctx, ready.Slot.Add(30*time.Second)); tickErr != nil || len(fires) != 0 {
		t.Fatalf("restart replayed the accepted slot: fires=%+v err=%v", fires, tickErr)
	}
	if fires, tickErr := engine.Tick(ctx, ready.Slot.Add(time.Minute)); tickErr != nil || len(fires) != 0 {
		t.Fatalf("later occurrence bypassed re-authorization: fires=%+v err=%v", fires, tickErr)
	}
	if built {
		t.Fatal("scheduler built a new task after authorization was revoked")
	}
	tasks, err := db.Tasks(ctx)
	if err != nil || len(tasks) != 1 || tasks[0].ID != ready.TaskID || tasks[0].RequestID != ready.RequestID {
		t.Fatalf("accepted task was duplicated or lost: tasks=%+v err=%v", tasks, err)
	}
}

func runSchedulerCrashHelper(t *testing.T, base string) {
	db, err := storage.Open(filepath.Join(base, "scheduler.db"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	engine, err := NewScheduler(db, SchedulerOptions{
		Authorize:  func(context.Context, ScheduleSpec) error { return nil },
		NodeOnline: func(context.Context, string) bool { return true },
		BuildTask: func(_ context.Context, spec ScheduleSpec, slot time.Time) (model.Task, error) {
			payload, marshalErr := json.Marshal(map[string]any{"slot": slot.UTC(), "proof": "accepted-before-process-kill"})
			if marshalErr != nil {
				return model.Task{}, marshalErr
			}
			return model.Task{Resource: spec.Resource, Action: spec.Action, Payload: payload}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	now := scheduleTime(t, "2026-09-21T09:00:01Z")
	plan, err := engine.Save(ctx, scheduleSpec(), 0, now)
	if err != nil {
		t.Fatal(err)
	}
	fires, err := engine.Tick(ctx, plan.NextRun)
	if err != nil || len(fires) != 1 || fires[0].State != "accepted" {
		t.Fatalf("prepare accepted occurrence: fires=%+v err=%v", fires, err)
	}
	ready := schedulerCrashReady{FireID: fires[0].ID, ScheduleID: fires[0].ScheduleID, TaskID: fires[0].Task.ID, RequestID: fires[0].Task.RequestID, Slot: fires[0].Slot}
	body, err := json.Marshal(ready)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = fmt.Fprintf(os.Stdout, "SCHEDULER_CRASH_READY %s\n", body); err != nil {
		t.Fatal(err)
	}
	select {}
}

package backup

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	_ "time/tzdata"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
	"github.com/robfig/cron/v3"
)

const schedulesNS = "platform_schedules_v1"
const firesNS = "platform_schedule_fires_v1"

type ScheduleSpec struct {
	ID       string            `json:"id"`
	OwnerID  string            `json:"ownerId"`
	Resource model.ResourceRef `json:"resource"`
	Action   string            `json:"action"`
	Args     json.RawMessage   `json:"args"`
	Cron     string            `json:"cron"`
	Timezone string            `json:"timezone"`
	Misfire  string            `json:"misfire"` // skip or run_once
	Overlap  string            `json:"overlap"` // skip or wait_one
	Enabled  bool              `json:"enabled"`
}
type Schedule struct {
	Spec                 ScheduleSpec `json:"spec"`
	ConfigRevision       int64        `json:"configRevision"`
	NextRun              time.Time    `json:"nextRun"`
	Pending              time.Time    `json:"pending,omitempty"`
	LastSlot             time.Time    `json:"lastSlot,omitempty"`
	LastTaskID           string       `json:"lastTaskId,omitempty"`
	Skipped              uint64       `json:"skipped"`
	MissedCountTruncated bool         `json:"missedCountTruncated"`
	LastError            string       `json:"lastError,omitempty"`
}
type Fire struct {
	ID             string     `json:"id"`
	ScheduleID     string     `json:"scheduleId"`
	ConfigRevision int64      `json:"configRevision"`
	Slot           time.Time  `json:"slot"`
	Created        time.Time  `json:"created"`
	Task           model.Task `json:"task"`
	State          string     `json:"state"`
	Error          string     `json:"error,omitempty"`
}
type SchedulerOptions struct {
	MaxSchedules   int
	MaxFireRecords int
	// Authorize checks current user/resource/root configuration and capabilities
	// at every trigger. BuildTask performs reads only and binds a concrete task.
	Authorize  func(context.Context, ScheduleSpec) error
	BuildTask  func(context.Context, ScheduleSpec, time.Time) (model.Task, error)
	NodeOnline func(context.Context, string) bool
}
type Scheduler struct {
	store *storage.Store
	opts  SchedulerOptions
	mu    sync.Mutex
	tick  sync.Mutex
}

type scheduleSaveMutationInput struct {
	ID                string       `json:"id"`
	Spec              ScheduleSpec `json:"spec"`
	ExpectedConfigRev int64        `json:"expectedConfigRevision"`
}

type scheduleDeleteMutationInput struct {
	ID                string `json:"id"`
	ExpectedConfigRev int64  `json:"expectedConfigRevision"`
}

func NewScheduler(store *storage.Store, opts SchedulerOptions) (*Scheduler, error) {
	if store == nil || opts.Authorize == nil || opts.BuildTask == nil || opts.NodeOnline == nil {
		return nil, ErrInvalid
	}
	if opts.MaxSchedules <= 0 {
		opts.MaxSchedules = 256
	}
	if opts.MaxFireRecords <= 0 {
		opts.MaxFireRecords = 4096
	}
	return &Scheduler{store: store, opts: opts}, nil
}
func parseSchedule(spec ScheduleSpec) (cron.Schedule, error) {
	if len(spec.Cron) > 256 || len(spec.Timezone) > 128 || len(strings.Fields(spec.Cron)) != 5 || strings.ContainsAny(spec.Timezone, " \t\r\n") || spec.Timezone == "" || spec.Timezone == "Local" {
		return nil, ErrInvalid
	}
	if _, err := time.LoadLocation(spec.Timezone); err != nil {
		return nil, ErrInvalid
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	return parser.Parse("CRON_TZ=" + spec.Timezone + " " + spec.Cron)
}
func (s *Scheduler) Save(ctx context.Context, spec ScheduleSpec, expectedConfigRevision int64, now time.Time) (Schedule, error) {
	if spec.ID == "" {
		spec.ID = model.ID()
	}
	if spec.Misfire == "" {
		spec.Misfire = "skip"
	}
	if spec.Overlap == "" {
		spec.Overlap = "skip"
	}
	if !validID(spec.ID) || spec.OwnerID == "" || !validResource(spec.Resource) || spec.Action == "" || len(spec.Args) > 16<<10 || !json.Valid(spec.Args) || (spec.Misfire != "skip" && spec.Misfire != "run_once") || (spec.Overlap != "skip" && spec.Overlap != "wait_one") {
		return Schedule{}, ErrInvalid
	}
	calendar, err := parseSchedule(spec)
	if err != nil {
		return Schedule{}, err
	}
	next := calendar.Next(now).UTC()
	if next.IsZero() {
		return Schedule{}, ErrInvalid
	}
	if spec.Enabled {
		if err = s.opts.Authorize(ctx, spec); err != nil {
			return Schedule{}, err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var old Schedule
	revision, err := s.store.Record(ctx, schedulesNS, spec.ID, &old)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return old, err
	}
	if old.ConfigRevision != expectedConfigRevision {
		return old, ErrConflict
	}
	if revision == 0 {
		ids, err := s.store.RecordIDs(ctx, schedulesNS)
		if err != nil {
			return old, err
		}
		if len(ids) >= s.opts.MaxSchedules {
			return old, ErrLimit
		}
	} else if old.Spec.OwnerID != spec.OwnerID {
		return old, ErrConflict
	}
	updated := Schedule{Spec: spec, ConfigRevision: old.ConfigRevision + 1, NextRun: next, LastTaskID: old.LastTaskID, LastSlot: old.LastSlot, Skipped: old.Skipped}
	_, err = s.store.PutRecord(ctx, schedulesNS, spec.ID, revision, updated)
	return updated, err
}

// SaveMutation atomically applies a schedule update and stores its response
// under the caller's request key. Repeating the same key returns the original
// schedule even when the configuration revision has since advanced.
func (s *Scheduler) SaveMutation(ctx context.Context, actor, requestID string, spec ScheduleSpec, expectedConfigRevision int64, now time.Time) (Schedule, error) {
	if spec.ID == "" {
		spec.ID = model.ID()
	}
	if spec.Misfire == "" {
		spec.Misfire = "skip"
	}
	if spec.Overlap == "" {
		spec.Overlap = "skip"
	}
	if !validID(spec.ID) || spec.OwnerID == "" || !validResource(spec.Resource) || spec.Action == "" || len(spec.Args) > 16<<10 || !json.Valid(spec.Args) || (spec.Misfire != "skip" && spec.Misfire != "run_once") || (spec.Overlap != "skip" && spec.Overlap != "wait_one") {
		return Schedule{}, ErrInvalid
	}
	calendar, err := parseSchedule(spec)
	if err != nil {
		return Schedule{}, err
	}
	next := calendar.Next(now)
	if next.IsZero() {
		return Schedule{}, ErrInvalid
	}
	if spec.Enabled {
		if err = s.opts.Authorize(ctx, spec); err != nil {
			return Schedule{}, err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := s.store.UserMetadataMutation(ctx, actor, requestID, "schedule.save", scheduleSaveMutationInput{ID: spec.ID, Spec: spec, ExpectedConfigRev: expectedConfigRevision}, func(tx *sql.Tx) (any, error) {
		var old Schedule
		var revision int64
		var document []byte
		err := tx.QueryRowContext(ctx, "SELECT revision,document FROM records WHERE namespace=? AND id=?", schedulesNS, spec.ID).Scan(&revision, &document)
		if err == nil {
			if err = json.Unmarshal(document, &old); err != nil {
				return nil, err
			}
		} else if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if old.ConfigRevision != expectedConfigRevision {
			return old, ErrConflict
		}
		if revision == 0 {
			var count int
			if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM records WHERE namespace=?", schedulesNS).Scan(&count); err != nil {
				return nil, err
			}
			if count >= s.opts.MaxSchedules {
				return old, ErrLimit
			}
		} else if old.Spec.OwnerID != spec.OwnerID {
			return old, ErrConflict
		}
		updated := Schedule{Spec: spec, ConfigRevision: old.ConfigRevision + 1, NextRun: next.UTC(), LastTaskID: old.LastTaskID, LastSlot: old.LastSlot, Skipped: old.Skipped}
		b, err := json.Marshal(updated)
		if err != nil {
			return nil, err
		}
		if revision == 0 {
			if _, err = tx.ExecContext(ctx, "INSERT INTO records(namespace,id,revision,document) VALUES(?,?,1,?)", schedulesNS, spec.ID, b); err != nil {
				return nil, err
			}
		} else {
			result, updateErr := tx.ExecContext(ctx, "UPDATE records SET revision=revision+1,document=? WHERE namespace=? AND id=? AND revision=?", b, schedulesNS, spec.ID, revision)
			if updateErr != nil {
				return nil, updateErr
			}
			n, rowsErr := result.RowsAffected()
			if rowsErr != nil {
				return nil, rowsErr
			}
			if n != 1 {
				return nil, ErrConflict
			}
		}
		return updated, nil
	})
	if err != nil {
		return Schedule{}, err
	}
	var result Schedule
	if err = json.Unmarshal(data, &result); err != nil {
		return Schedule{}, err
	}
	return result, nil
}
func (s *Scheduler) Get(ctx context.Context, id string) (Schedule, error) {
	var record Schedule
	_, err := s.store.Record(ctx, schedulesNS, id, &record)
	return record, err
}
func (s *Scheduler) List(ctx context.Context) ([]Schedule, error) {
	ids, err := s.store.RecordIDs(ctx, schedulesNS)
	if err != nil {
		return nil, err
	}
	if len(ids) > s.opts.MaxSchedules {
		return nil, ErrLimit
	}
	items := make([]Schedule, 0, len(ids))
	for _, id := range ids {
		item, err := s.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Spec.ID < items[j].Spec.ID })
	return items, nil
}

// Delete only removes this module's schedule configuration. Already accepted
// tasks retain their durable identities and require their normal cancellation.
// The HTTP caller authenticates ownership/admin access before invoking it.
func (s *Scheduler) Delete(ctx context.Context, id string, expectedConfigRevision int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var item Schedule
	revision, err := s.store.Record(ctx, schedulesNS, id, &item)
	if err != nil {
		return err
	}
	if item.ConfigRevision != expectedConfigRevision {
		return ErrConflict
	}
	result, err := s.store.DB.ExecContext(ctx, "DELETE FROM records WHERE namespace=? AND id=? AND revision=?", schedulesNS, id, revision)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err == nil && n != 1 {
		return ErrConflict
	}
	return err
}

// DeleteMutation atomically removes a schedule and stores a replayable result
// under the caller's request key. The callback checks ownership again inside
// the same transaction so an admin or owner cannot delete a different record.
func (s *Scheduler) DeleteMutation(ctx context.Context, actor, requestID, id string, expectedConfigRevision int64) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := s.store.UserMetadataMutation(ctx, actor, requestID, "schedule.delete", scheduleDeleteMutationInput{ID: id, ExpectedConfigRev: expectedConfigRevision}, func(tx *sql.Tx) (any, error) {
		var revision int64
		var document []byte
		if err := tx.QueryRowContext(ctx, "SELECT revision,document FROM records WHERE namespace=? AND id=?", schedulesNS, id).Scan(&revision, &document); err != nil {
			return nil, err
		}
		var item Schedule
		if err := json.Unmarshal(document, &item); err != nil {
			return nil, err
		}
		var admin bool
		if err := tx.QueryRowContext(ctx, "SELECT admin FROM users WHERE id=?", actor).Scan(&admin); err != nil {
			return nil, err
		}
		if item.Spec.OwnerID != actor && !admin {
			return nil, sql.ErrNoRows
		}
		if item.ConfigRevision != expectedConfigRevision {
			return nil, ErrConflict
		}
		result, err := tx.ExecContext(ctx, "DELETE FROM records WHERE namespace=? AND id=? AND revision=?", schedulesNS, id, revision)
		if err != nil {
			return nil, err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return nil, err
		}
		if n != 1 {
			return nil, ErrConflict
		}
		return true, nil
	})
	if err != nil {
		return false, err
	}
	var deleted bool
	if err = json.Unmarshal(data, &deleted); err != nil {
		return false, err
	}
	return deleted, nil
}

// Tick creates at most one concrete task per schedule. It is called by Master;
// Daemons never carry schedules or produce new work from an offline plan.
func (s *Scheduler) Tick(ctx context.Context, now time.Time) ([]Fire, error) {
	s.tick.Lock()
	defer s.tick.Unlock()
	items, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	var results []Fire
	var failures []error
	for _, item := range items {
		if err = ctx.Err(); err != nil {
			return results, err
		}
		if !item.Spec.Enabled {
			continue
		}
		fire, err := s.tickOne(ctx, item.Spec.ID, now.UTC())
		if fire.ID != "" && fire.State == "accepted" {
			results = append(results, fire)
		}
		if err != nil {
			failures = append(failures, err)
		}
	}
	return results, errors.Join(failures...)
}
func (s *Scheduler) tickOne(ctx context.Context, id string, now time.Time) (Fire, error) {
	var item Schedule
	revision, err := s.store.Record(ctx, schedulesNS, id, &item)
	if err != nil {
		return Fire{}, err
	}
	if !item.Spec.Enabled {
		return Fire{}, nil
	}
	calendar, err := parseSchedule(item.Spec)
	if err != nil {
		return Fire{}, err
	}
	if item.NextRun.IsZero() {
		return Fire{}, ErrInvalid
	}
	due := !now.Before(item.NextRun)
	if !due && item.Pending.IsZero() {
		return Fire{}, nil
	}
	if due {
		last := item.NextRun
		count := uint64(0)
		for !now.Before(item.NextRun) && count < 4096 {
			last = item.NextRun
			item.NextRun = calendar.Next(item.NextRun).UTC()
			count++
			if item.NextRun.IsZero() {
				return Fire{}, ErrInvalid
			}
		}
		if count == 4096 && !now.Before(item.NextRun) {
			item.MissedCountTruncated = true
			item.NextRun = calendar.Next(now).UTC()
		}
		// A slot is current for its minute. Older slots are explicit misfires;
		// run_once remembers one latest observed slot, never an unbounded queue.
		if !item.Pending.IsZero() {
			item.Skipped += count // keep the one durable pending intent
		} else if now.Sub(last) < time.Minute {
			item.Pending = last
			if count > 1 {
				item.Skipped += count - 1
			}
		} else if item.Spec.Misfire == "run_once" {
			item.Pending = last
			if count > 1 {
				item.Skipped += count - 1
			}
		} else {
			item.Skipped += count
			item.LastError = "missed_skipped"
		}
	}
	// Save the chosen slot and next time before preparing or accepting a task.
	// Recovery therefore reconciles the same occurrence even after hours offline.
	revision, err = s.store.PutRecord(ctx, schedulesNS, id, revision, item)
	if err != nil {
		return Fire{}, err
	}
	if item.Pending.IsZero() {
		return Fire{}, nil
	}
	slot := item.Pending
	fireID := hashBytes([]byte(fmt.Sprintf("%s:%d:%d", id, item.ConfigRevision, slot.Unix())))[7:39]
	var fire Fire
	fireRevision, err := s.store.Record(ctx, firesNS, fireID, &fire)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fire, err
	}
	if fireRevision > 0 {
		if fire.ID != fireID || fire.ScheduleID != id || fire.ConfigRevision != item.ConfigRevision || !fire.Slot.Equal(slot) || fire.Task.ActorID != item.Spec.OwnerID || fire.Task.Resource != item.Spec.Resource || fire.Task.Action != item.Spec.Action || fire.Task.RequestID != "schedule:"+fireID {
			return fire, ErrConflict
		}
		if old, lookupErr := s.store.TaskByRequest(ctx, fire.Task.ActorID, fire.Task.RequestID); lookupErr == nil {
			if old.Action != fire.Task.Action || old.Resource != fire.Task.Resource || !bytes.Equal(old.Payload, fire.Task.Payload) {
				return fire, ErrConflict
			}
			// This is a receipt for work already accepted, including when the
			// actor is now revoked or the node is offline. No new task is created.
			fire.Task, fire.State = old, "accepted"
			if _, err = s.store.PutRecord(ctx, firesNS, fire.ID, fireRevision, fire); err != nil {
				return fire, err
			}
			item.LastTaskID, item.LastSlot, item.Pending, item.LastError = old.ID, slot, time.Time{}, ""
			_, err = s.store.PutRecord(ctx, schedulesNS, id, revision, item)
			return fire, err
		} else if !errors.Is(lookupErr, sql.ErrNoRows) {
			return fire, lookupErr
		} else if fire.State == "accepted" {
			return fire, ErrInterrupted
		}
	}
	if err = s.opts.Authorize(ctx, item.Spec); err != nil {
		item.Skipped++
		item.Pending = time.Time{}
		item.LastError = "authorization_denied"
		_, saveErr := s.store.PutRecord(ctx, schedulesNS, id, revision, item)
		return Fire{}, saveErr
	}
	if !s.opts.NodeOnline(ctx, item.Spec.Resource.NodeID) {
		item.LastError = "node_offline"
		if item.Spec.Misfire == "skip" {
			item.Skipped++
			item.Pending = time.Time{}
		}
		_, err = s.store.PutRecord(ctx, schedulesNS, id, revision, item)
		return Fire{}, err
	}
	if busy, err := s.busy(ctx, item.Spec); err != nil {
		return Fire{}, err
	} else if busy {
		item.LastError = "resource_overlap"
		if item.Spec.Overlap == "skip" {
			item.Skipped++
			item.Pending = time.Time{}
		}
		_, err = s.store.PutRecord(ctx, schedulesNS, id, revision, item)
		return Fire{}, err
	}
	if fireRevision == 0 {
		ids, err := s.store.RecordIDs(ctx, firesNS)
		if err != nil {
			return fire, err
		}
		if len(ids) >= s.opts.MaxFireRecords {
			return fire, ErrLimit
		}
		task, err := s.opts.BuildTask(ctx, item.Spec, slot)
		if err != nil {
			item.LastError = "task_preparation_failed"
			_, saveErr := s.store.PutRecord(ctx, schedulesNS, id, revision, item)
			return fire, saveErr
		}
		if task.Action != item.Spec.Action || task.Resource != item.Spec.Resource || len(task.Payload) > 20<<10 || !json.Valid(task.Payload) {
			return fire, ErrInvalid
		}
		var payload map[string]json.RawMessage
		if json.Unmarshal(task.Payload, &payload) != nil || payload == nil {
			return fire, ErrInvalid
		}
		payload["scheduleId"], _ = json.Marshal(id)
		payload["scheduleSlot"], _ = json.Marshal(slot)
		payload["scheduleConfigRevision"], _ = json.Marshal(item.ConfigRevision)
		task.Payload, _ = json.Marshal(payload)
		task.ActorID = item.Spec.OwnerID
		if task.ID == "" {
			task.ID = model.ID()
		}
		if !validID(task.ID) {
			return fire, ErrInvalid
		}
		task.State = model.Queued
		task.Result = nil
		task.RequestID = "schedule:" + fireID
		fire = Fire{ID: fireID, ScheduleID: id, ConfigRevision: item.ConfigRevision, Slot: slot, Created: now, Task: task, State: "prepared"}
		fireRevision, err = s.store.PutRecord(ctx, firesNS, fireID, 0, fire)
		if err != nil {
			return fire, err
		}
	}
	// Configuration changes/disable and the final accept are serialized. Reads
	// or node queries above do not hold this short administrative mutex.
	s.mu.Lock()
	defer s.mu.Unlock()
	var current Schedule
	currentRevision, err := s.store.Record(ctx, schedulesNS, id, &current)
	if errors.Is(err, sql.ErrNoRows) {
		fire.State = "discarded"
		fire.Error = "configuration_removed"
		_, err = s.store.PutRecord(ctx, firesNS, fire.ID, fireRevision, fire)
		return fire, err
	}
	if err != nil {
		return fire, err
	}
	if !current.Spec.Enabled || current.ConfigRevision != item.ConfigRevision {
		fire.State = "discarded"
		fire.Error = "configuration_changed"
		_, err = s.store.PutRecord(ctx, firesNS, fire.ID, fireRevision, fire)
		return fire, err
	}
	if err = s.opts.Authorize(ctx, current.Spec); err != nil {
		item.LastError = "authorization_denied"
		item.Pending = time.Time{}
		item.Skipped++
		_, err = s.store.PutRecord(ctx, schedulesNS, id, currentRevision, item)
		return fire, err
	}
	if fire.ScheduleID != id || fire.ConfigRevision != item.ConfigRevision || fire.Slot != slot || fire.Task.ActorID != item.Spec.OwnerID || fire.Task.Resource != item.Spec.Resource || fire.Task.Action != item.Spec.Action {
		return fire, ErrConflict
	}
	if old, lookupErr := s.store.TaskByRequest(ctx, fire.Task.ActorID, fire.Task.RequestID); lookupErr == nil {
		if old.Action != fire.Task.Action || old.Resource != fire.Task.Resource || !bytes.Equal(old.Payload, fire.Task.Payload) {
			return fire, ErrConflict
		}
		fire.Task = old
	} else if !errors.Is(lookupErr, sql.ErrNoRows) {
		return fire, lookupErr
	} else {
		if busy, e := s.busy(ctx, current.Spec); e != nil {
			return fire, e
		} else if busy {
			item.LastError = "resource_overlap"
			if item.Spec.Overlap == "skip" {
				item.Pending = time.Time{}
				item.Skipped++
			}
			_, err = s.store.PutRecord(ctx, schedulesNS, id, currentRevision, item)
			return Fire{}, err
		}
		fire.Task, _, err = s.store.Accept(ctx, fire.Task)
		if err != nil {
			return fire, err
		}
	}
	fire.State = "accepted"
	if _, err = s.store.PutRecord(ctx, firesNS, fire.ID, fireRevision, fire); err != nil {
		return fire, err
	}
	item.LastTaskID, item.LastSlot, item.Pending, item.LastError = fire.Task.ID, slot, time.Time{}, ""
	_, err = s.store.PutRecord(ctx, schedulesNS, id, currentRevision, item)
	return fire, err
}
func (s *Scheduler) busy(ctx context.Context, spec ScheduleSpec) (bool, error) {
	tasks, err := s.store.Pending(ctx)
	if err != nil {
		return false, err
	}
	for _, task := range tasks {
		if task.Resource == spec.Resource && task.Action == spec.Action {
			return true, nil
		}
	}
	return false, nil
}

// PruneFires only removes this scheduler's completed fire metadata. The task
// journal remains owned by storage. Unaccepted records are only removable once
// their occurrence is no longer pending in the current configuration.
func (s *Scheduler) PruneFires(ctx context.Context, before time.Time) (int, error) {
	s.tick.Lock()
	defer s.tick.Unlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	ids, err := s.store.RecordIDs(ctx, firesNS)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, id := range ids {
		var fire Fire
		revision, err := s.store.Record(ctx, firesNS, id, &fire)
		if err != nil {
			return count, err
		}
		if !fire.Created.Before(before) {
			continue
		}
		task, err := s.store.TaskByRequest(ctx, fire.Task.ActorID, fire.Task.RequestID)
		if err == nil {
			if task.Action != fire.Task.Action || task.Resource != fire.Task.Resource || !bytes.Equal(task.Payload, fire.Task.Payload) {
				return count, ErrConflict
			}
			if !task.State.Terminal() {
				continue
			}
		} else if !errors.Is(err, sql.ErrNoRows) {
			return count, err
		} else {
			if fire.State == "accepted" {
				return count, ErrInterrupted
			}
			current, e := s.Get(ctx, fire.ScheduleID)
			if e != nil && !errors.Is(e, sql.ErrNoRows) {
				return count, e
			}
			if e == nil && current.Spec.Enabled && current.ConfigRevision == fire.ConfigRevision && current.Pending.Equal(fire.Slot) {
				continue
			}
		}
		result, err := s.store.DB.ExecContext(ctx, "DELETE FROM records WHERE namespace=? AND id=? AND revision=?", firesNS, id, revision)
		if err != nil {
			return count, err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return count, err
		}
		count += int(n)
	}
	return count, nil
}

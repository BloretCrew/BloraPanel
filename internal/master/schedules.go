package master

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"sync"
	"time"
	"unicode/utf8"

	"blora.dev/panel/internal/backup"
	"blora.dev/panel/internal/filesystem"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
)

type scheduleService struct {
	engine *backup.Scheduler
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}
type scheduleInput struct {
	Revision int64             `json:"revision"`
	Resource model.ResourceRef `json:"resource"`
	Action   string            `json:"action"`
	Args     json.RawMessage   `json:"args"`
	Cron     string            `json:"cron"`
	Timezone string            `json:"timezone"`
	Misfire  string            `json:"misfire"`
	Overlap  string            `json:"overlap"`
	Enabled  bool              `json:"enabled"`
}
type scheduledBackup struct {
	Path        string             `json:"path"`
	Compression string             `json:"compression"`
	Consistency backup.Consistency `json:"consistency"`
}

var scheduledHookRef = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,128}$`)

func validateScheduledConsistency(c backup.Consistency) error {
	// An omitted mode has the same durable meaning as the regular backup API:
	// capture the files as observed and do not invoke application hooks.
	if c.Mode == "" {
		c.Mode = "files"
	}
	switch c.Mode {
	case "files":
		if c.Before != "" || c.After != "" {
			return backup.ErrInvalid
		}
	case "save", "pause", "stop", "hooks":
		for _, ref := range []string{c.Before, c.After} {
			if ref != "" && !scheduledHookRef.MatchString(ref) {
				return backup.ErrInvalid
			}
		}
	default:
		return backup.ErrInvalid
	}
	return nil
}

func strictScheduleArgs(raw json.RawMessage, out any) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		return backup.ErrInvalid
	}
	var tail any
	if !errors.Is(d.Decode(&tail), io.EOF) {
		return backup.ErrInvalid
	}
	return nil
}
func validateScheduleAction(spec backup.ScheduleSpec) error {
	if spec.Resource.Kind != "instance" {
		return backup.ErrUnsupported
	}
	switch spec.Action {
	case "instance.start", "instance.stop", "instance.restart":
		return strictScheduleArgs(spec.Args, &struct{}{})
	case "console.input":
		var in struct {
			Data string `json:"data"`
		}
		if err := strictScheduleArgs(spec.Args, &in); err != nil {
			return err
		}
		if len(in.Data) == 0 || len(in.Data) > 8192 || !utf8.ValidString(in.Data) {
			return backup.ErrInvalid
		}
	case "backup.create":
		var in scheduledBackup
		if err := strictScheduleArgs(spec.Args, &in); err != nil {
			return err
		}
		if filesystem.ValidatePath(in.Path) != nil || (in.Compression != "store" && in.Compression != "deflate") {
			return backup.ErrInvalid
		}
		if err := validateScheduledConsistency(in.Consistency); err != nil {
			return err
		}
	default:
		return backup.ErrUnsupported
	}
	return nil
}
func (s *Server) scheduleAuthorize(ctx context.Context, spec backup.ScheduleSpec) error {
	if err := validateScheduleAction(spec); err != nil {
		return err
	}
	u, err := s.store.User(ctx, spec.OwnerID)
	if err != nil || u.Disabled {
		return storage.ErrForbidden
	}
	i, err := s.store.Instance(ctx, spec.Resource.ID)
	if err != nil {
		return err
	}
	if instanceRef(i) != spec.Resource {
		return backup.ErrConflict
	}
	for _, action := range []string{"schedule.create", taskPermission(spec.Action)} {
		if !s.store.Allowed(ctx, u, spec.Resource, action) {
			return storage.ErrForbidden
		}
	}
	if spec.Action == "backup.create" && !s.store.Allowed(ctx, u, spec.Resource, "file.read") {
		return storage.ErrForbidden
	}
	n, _, err := s.store.Node(ctx, i.NodeID)
	if err != nil {
		return err
	}
	if n.Maintenance {
		return storage.ErrNodeUnavailable
	}
	return nil
}
func (s *Server) buildScheduledTask(ctx context.Context, spec backup.ScheduleSpec, slot time.Time) (model.Task, error) {
	if err := s.scheduleAuthorize(ctx, spec); err != nil {
		return model.Task{}, err
	}
	i, err := s.store.Instance(ctx, spec.Resource.ID)
	if err != nil {
		return model.Task{}, err
	}
	var payload any = i.Config
	switch spec.Action {
	case "console.input":
		var in model.ConsoleInput
		if err = json.Unmarshal(spec.Args, &in); err != nil {
			return model.Task{}, err
		}
		if i.RunID == "" || i.State != "RUNNING" {
			return model.Task{}, backup.ErrConflict
		}
		in.RunID = i.RunID
		payload = in
	case "backup.create":
		var in scheduledBackup
		if err = json.Unmarshal(spec.Args, &in); err != nil {
			return model.Task{}, err
		}
		u, err := s.store.User(ctx, spec.OwnerID)
		if err != nil {
			return model.Task{}, err
		}
		var entry filesystem.Entry
		if err = s.callFile(ctx, u, i, "file.stat", map[string]string{"path": in.Path}, &entry); err != nil {
			return model.Task{}, err
		}
		payload = backup.TaskPayload{Config: i.Config, Create: &backup.CreateSpec{ID: model.ID(), OwnerID: spec.OwnerID, Source: instanceRef(i), Path: in.Path, Version: entry.Version, Compression: in.Compression, Consistency: in.Consistency}}
	}
	b, err := json.Marshal(payload)
	return model.Task{ActorID: spec.OwnerID, Resource: spec.Resource, Action: spec.Action, Payload: b}, err
}
func (s *Server) registerSchedules() {
	engine, err := backup.NewScheduler(s.store, backup.SchedulerOptions{Authorize: s.scheduleAuthorize, BuildTask: s.buildScheduledTask, NodeOnline: func(ctx context.Context, id string) bool {
		n, _, err := s.store.Node(ctx, id)
		if err != nil || n.Maintenance {
			return false
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		return s.peers[id] != nil
	}})
	if err != nil {
		panic(err)
	} // All required callbacks are fixed above.
	ctx, cancel := context.WithCancel(context.Background())
	service := &scheduleService{engine: engine, cancel: cancel, done: make(chan struct{})}
	s.schedules = service
	s.mux.HandleFunc("GET /api/v1/schedules", s.auth(s.listSchedules))
	s.mux.HandleFunc("POST /api/v1/schedules", s.auth(s.saveSchedule))
	s.mux.HandleFunc("PUT /api/v1/schedules/{id}", s.auth(s.saveSchedule))
	s.mux.HandleFunc("DELETE /api/v1/schedules/{id}", s.auth(s.deleteSchedule))
	go func() {
		defer close(service.done)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		prune := time.NewTicker(time.Hour)
		defer prune.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				call, done := context.WithTimeout(ctx, 30*time.Second)
				_, err := engine.Tick(call, now)
				done()
				if err != nil && ctx.Err() == nil {
					slog.Warn("schedule tick incomplete", "error", err)
				}
			case now := <-prune.C:
				call, done := context.WithTimeout(ctx, 30*time.Second)
				_, err := engine.PruneFires(call, now.Add(-30*24*time.Hour))
				done()
				if err != nil && ctx.Err() == nil {
					slog.Warn("schedule receipt retention incomplete", "error", err)
				}
			}
		}
	}()
}
func (s *Server) closeSchedules() {
	if s.schedules != nil {
		s.schedules.once.Do(func() { s.schedules.cancel(); <-s.schedules.done })
	}
}
func (s *Server) listSchedules(w http.ResponseWriter, r *http.Request, u model.User) {
	items, err := s.schedules.engine.List(r.Context())
	if err != nil {
		scheduleError(w, err)
		return
	}
	out := []backup.Schedule{}
	for _, item := range items {
		if item.Spec.OwnerID == u.ID || u.Admin {
			out = append(out, item)
		}
	}
	reply(w, 200, map[string]any{"items": out})
}
func scheduleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrForbidden):
		fail(w, 403, "FORBIDDEN", "调度或实际操作权限不足")
	case errors.Is(err, sql.ErrNoRows):
		fail(w, 404, "NOT_FOUND", "计划或资源不存在")
	case errors.Is(err, backup.ErrConflict), errors.Is(err, storage.ErrConflict):
		fail(w, 409, "SCHEDULE_CONFLICT", "计划修订或资源已变更")
	case errors.Is(err, storage.ErrRequestMismatch):
		fail(w, 409, "IDEMPOTENCY_CONFLICT", "同一请求键不能复用不同内容")
	case errors.Is(err, backup.ErrLimit):
		fail(w, 429, "SCHEDULE_LIMIT", "调度记录达到上限")
	case errors.Is(err, backup.ErrUnsupported):
		fail(w, 400, "CAPABILITY_UNAVAILABLE", "此调度动作或一致性策略尚未接入")
	default:
		fail(w, 400, "INVALID_SCHEDULE", err.Error())
	}
}
func (s *Server) saveSchedule(w http.ResponseWriter, r *http.Request, u model.User) {
	if !requireRequestID(w, r) {
		return
	}
	var in scheduleInput
	if !decode(w, r, &in) {
		return
	}
	id := r.PathValue("id")
	owner := u.ID
	if id == "" {
		if in.Revision != 0 {
			scheduleError(w, backup.ErrConflict)
			return
		}
		id = storage.Hash([]byte(u.ID + ":schedule:" + r.Header.Get("Idempotency-Key")))[:32]
	} else {
		old, err := s.schedules.engine.Get(r.Context(), id)
		if err != nil {
			scheduleError(w, err)
			return
		}
		if old.Spec.OwnerID != u.ID && !u.Admin {
			scheduleError(w, sql.ErrNoRows)
			return
		}
		owner = old.Spec.OwnerID
	}
	if in.Misfire == "" {
		in.Misfire = "skip"
	}
	if in.Overlap == "" {
		in.Overlap = "skip"
	}
	if len(in.Args) == 0 {
		in.Args = json.RawMessage(`{}`)
	}
	spec := backup.ScheduleSpec{ID: id, OwnerID: owner, Resource: in.Resource, Action: in.Action, Args: in.Args, Cron: in.Cron, Timezone: in.Timezone, Misfire: in.Misfire, Overlap: in.Overlap, Enabled: in.Enabled}
	if err := validateScheduleAction(spec); err != nil {
		scheduleError(w, err)
		return
	}
	// Owners can still disable an accepted schedule after action permissions
	// are revoked; a changed or newly enabled configuration requires scope.
	old, oldErr := s.schedules.engine.Get(r.Context(), id)
	disabling := oldErr == nil && !spec.Enabled && old.Spec.Resource == spec.Resource && (owner == u.ID || u.Admin)
	if !disabling && !s.allowed(w, r, u, spec.Resource, "schedule.create") {
		return
	}
	i, err := s.store.Instance(r.Context(), spec.Resource.ID)
	if err != nil {
		scheduleError(w, err)
		return
	}
	if instanceRef(i) != spec.Resource {
		scheduleError(w, backup.ErrConflict)
		return
	}
	if spec.Enabled {
		if err = s.scheduleAuthorize(r.Context(), spec); err != nil {
			scheduleError(w, err)
			return
		}
	}
	// Stable POST identity makes a lost creation response inspectable without
	// advancing the configuration revision or resetting the next occurrence.
	if r.Method == "POST" && oldErr == nil {
		a, _ := json.Marshal(old.Spec)
		b, _ := json.Marshal(spec)
		if bytes.Equal(a, b) {
			reply(w, 200, map[string]any{"schedule": old})
			return
		}
		scheduleError(w, backup.ErrConflict)
		return
	}
	saved, err := s.schedules.engine.SaveMutation(r.Context(), u.ID, r.Header.Get("Idempotency-Key"), spec, in.Revision, time.Now().UTC())
	if err != nil {
		scheduleError(w, err)
		return
	}
	status := 200
	if r.Method == "POST" {
		status = 201
	}
	reply(w, status, map[string]any{"schedule": saved})
}
func (s *Server) deleteSchedule(w http.ResponseWriter, r *http.Request, u model.User) {
	if !requireRequestID(w, r) {
		return
	}
	var in struct {
		Revision int64 `json:"revision"`
	}
	if !decode(w, r, &in) {
		return
	}
	deleted, err := s.schedules.engine.DeleteMutation(r.Context(), u.ID, r.Header.Get("Idempotency-Key"), r.PathValue("id"), in.Revision)
	if err != nil {
		scheduleError(w, err)
		return
	}
	reply(w, 200, map[string]bool{"deleted": deleted})
}

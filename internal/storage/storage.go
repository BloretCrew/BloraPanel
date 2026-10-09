// Package storage owns explicit SQLite migrations and durable acceptance boundaries.
package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"blora.dev/panel/internal/model"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

// SchemaFingerprint identifies the precise migration code compiled into this
// executable. Online code rollback is allowed only when that code is unchanged.
func SchemaFingerprint() string {
	h := sha256.New()
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		b, err := migrations.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return ""
		}
		h.Write([]byte("internal/storage/migrations/" + entry.Name()))
		h.Write([]byte{0})
		h.Write(b)
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func SchemaVersion() int {
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return 0
	}
	return len(entries)
}

var (
	ErrConflict        = errors.New("revision conflict")
	ErrRequestMismatch = errors.New("request ID already accepted with different content")
	ErrTransition      = errors.New("invalid task transition")
	ErrQueueLimit      = errors.New("pending task budget exhausted")
	ErrSchemaVersion   = errors.New("database migration history is not supported by this binary")
)

type Store struct{ DB *sql.DB }

func Open(path string) (*Store, error) {
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return nil, err
		}
		f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			return nil, err
		}
		f.Close()
		if err := os.Chmod(path, 0600); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // PRAGMAs apply to the sole connection; short transactions only.
	db.SetMaxIdleConns(1)
	for _, statement := range []string{"PRAGMA foreign_keys=ON", "PRAGMA busy_timeout=5000", "PRAGMA journal_mode=WAL", "PRAGMA synchronous=FULL", "CREATE TABLE IF NOT EXISTS schema_migrations(version TEXT PRIMARY KEY)"} {
		if _, err := db.Exec(statement); err != nil {
			db.Close()
			return nil, err
		}
	}
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		db.Close()
		return nil, err
	}
	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version LIMIT ?", len(entries)+1)
	if err != nil {
		db.Close()
		return nil, err
	}
	index := 0
	for rows.Next() {
		var version string
		if err = rows.Scan(&version); err != nil {
			err = fmt.Errorf("%w: unreadable version record", ErrSchemaVersion)
			break
		}
		if index >= len(entries) || version != entries[index].Name() {
			err = fmt.Errorf("%w: unexpected version %q; use a compatible binary or restore its matching backup", ErrSchemaVersion, version)
			break
		}
		index++
	}
	if err == nil {
		err = rows.Err()
	}
	rows.Close()
	if err != nil {
		db.Close()
		return nil, err
	}
	for _, entry := range entries {
		var count int
		if err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version=?", entry.Name()).Scan(&count); err != nil {
			break
		}
		if count != 0 {
			continue
		}
		var data []byte
		data, err = migrations.ReadFile("migrations/" + entry.Name())
		if err != nil {
			break
		}
		var tx *sql.Tx
		tx, err = db.Begin()
		if err != nil {
			break
		}
		if _, err = tx.Exec(string(data)); err == nil {
			_, err = tx.Exec("INSERT INTO schema_migrations(version) VALUES(?)", entry.Name())
		}
		if err != nil {
			tx.Rollback()
			break
		}
		err = tx.Commit()
		if err != nil {
			break
		}
	}
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("migration: %w", err)
	}
	return &Store{DB: db}, nil
}

func (s *Store) Close() error { return s.DB.Close() }
func Hash(b []byte) string    { sum := sha256.Sum256(b); return hex.EncodeToString(sum[:]) }

func taskDigest(t model.Task) (string, error) {
	var payload any
	decoder := json.NewDecoder(strings.NewReader(string(t.Payload)))
	decoder.UseNumber()
	if len(t.Payload) != 0 {
		if err := decoder.Decode(&payload); err != nil {
			return "", err
		}
	}
	b, err := json.Marshal(struct {
		Resource model.ResourceRef
		Action   string
		Payload  any
	}{t.Resource, t.Action, payload})
	if err != nil {
		return "", err
	}
	return Hash(b), nil
}

// Accept commits both the task and first event before any executor may run.
// Deduplication is scoped to actor, not a browser window or connection.
func (s *Store) Accept(ctx context.Context, task model.Task) (model.Task, bool, error) {
	if task.ActorID == "" || task.RequestID == "" || len(task.RequestID) > 128 || task.Resource.ID == "" || task.Action == "" {
		return task, false, errors.New("actor, request, resource and action required")
	}
	if len(task.Payload) > 24*1024 {
		return task, false, errors.New("task payload exceeds limit")
	}
	digest, err := taskDigest(task)
	if err != nil {
		return task, false, err
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return task, false, err
	}
	defer tx.Rollback()
	var old []byte
	err = tx.QueryRowContext(ctx, "SELECT document FROM tasks WHERE actor_id=? AND request_id=?", task.ActorID, task.RequestID).Scan(&old)
	if err == nil {
		var existing model.Task
		if err := json.Unmarshal(old, &existing); err != nil {
			return task, false, err
		}
		if existing.Digest != digest {
			return existing, false, ErrRequestMismatch
		}
		return existing, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return task, false, err
	}
	if err := checkTaskConfiguration(ctx, tx, task); err != nil {
		return task, false, err
	}
	// Scheduler admission and the pending-work check share the transaction.
	// Manual work accepted immediately before a trigger cannot slip between
	// the scheduler's read-only overlap check and this durable admission.
	var scheduled struct {
		ScheduleID string `json:"scheduleId"`
	}
	_ = json.Unmarshal(task.Payload, &scheduled)
	if scheduled.ScheduleID != "" {
		var busy int
		if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE resource_key=? AND json_extract(document,'$.action')=? AND state NOT IN ('SUCCEEDED','FAILED','CANCELLED','INTERRUPTED')", task.Resource.Key(), task.Action).Scan(&busy); err != nil {
			return task, false, err
		}
		if busy != 0 {
			return task, false, ErrConflict
		}
	}
	var pending int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE state NOT IN ('SUCCEEDED','FAILED','CANCELLED','INTERRUPTED')").Scan(&pending); err != nil {
		return task, false, err
	}
	if pending >= 1024 {
		return task, false, ErrQueueLimit
	}
	if task.ID == "" {
		task.ID = model.ID()
	}
	task.Digest = digest
	if task.State == model.WaitingClient {
		task.Phase = "waiting_client_data"
	} else {
		task.State = model.Queued
		task.Phase = "accepted"
	}
	task.Revision = 1
	task.CreatedAt = time.Now().UTC()
	task.UpdatedAt = task.CreatedAt
	data, err := json.Marshal(task)
	if err != nil {
		return task, false, err
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO tasks(id,actor_id,request_id,resource_key,digest,state,revision,document) VALUES(?,?,?,?,?,?,?,?)", task.ID, task.ActorID, task.RequestID, task.Resource.Key(), task.Digest, task.State, task.Revision, data)
	if err != nil {
		return task, false, err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO task_events(task_id,revision,recorded_at,document) VALUES(?,?,?,?)", task.ID, task.Revision, task.UpdatedAt.UnixMilli(), data); err != nil {
		return task, false, err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO audit(actor_id,node_id,resource_key,action,request_id,result,recorded_at) VALUES(?,?,?,?,?,?,?)", task.ActorID, task.Resource.NodeID, task.Resource.Key(), task.Action, task.RequestID, "accepted", task.UpdatedAt.UnixMilli()); err != nil {
		return task, false, err
	}
	return task, true, tx.Commit()
}

func (s *Store) Task(ctx context.Context, id string) (model.Task, error) {
	var t model.Task
	var b []byte
	if err := s.DB.QueryRowContext(ctx, "SELECT document FROM tasks WHERE id=?", id).Scan(&b); err != nil {
		return t, err
	}
	return t, json.Unmarshal(b, &t)
}
func (s *Store) TaskByRequest(ctx context.Context, actor, request string) (model.Task, error) {
	var t model.Task
	var b []byte
	if err := s.DB.QueryRowContext(ctx, "SELECT document FROM tasks WHERE actor_id=? AND request_id=?", actor, request).Scan(&b); err != nil {
		return t, err
	}
	return t, json.Unmarshal(b, &t)
}
func (s *Store) Tasks(ctx context.Context) ([]model.Task, error) {
	return s.listTasks(ctx, false)
}

// RootTasks keeps internal transfer steps from displacing their durable parent
// in the task centre. Children remain directly addressable for diagnostics.
func (s *Store) RootTasks(ctx context.Context) ([]model.Task, error) {
	return s.listTasks(ctx, true)
}

type TaskPageEntry struct {
	Cursor int64
	Task   model.Task
}

// RootTaskPage scans a bounded, stable rowid page. Authorization is applied by
// the caller; insertions at the head cannot shift subsequent pages.
func (s *Store) RootTaskPage(ctx context.Context, before int64, limit int) ([]TaskPageEntry, bool, error) {
	return s.RootTaskPageFiltered(ctx, before, limit, "", false)
}

func (s *Store) RootTaskPageFiltered(ctx context.Context, before int64, limit int, state model.TaskState, activeOnly bool) ([]TaskPageEntry, bool, error) {
	if before < 0 || limit < 1 || limit > 1000 {
		return nil, false, errors.New("invalid task page")
	}
	query := "SELECT rowid,document FROM tasks WHERE COALESCE(json_extract(document, '$.payload.transferParentId'), '')=''"
	args := []any{}
	if state != "" {
		query += " AND state=?"
		args = append(args, state)
	}
	if activeOnly {
		query += " AND state NOT IN ('SUCCEEDED','FAILED','CANCELLED','INTERRUPTED')"
	}
	if before > 0 {
		query += " AND rowid<?"
		args = append(args, before)
	}
	args = append(args, limit+1)
	rows, err := s.DB.QueryContext(ctx, query+" ORDER BY rowid DESC LIMIT ?", args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	items := []TaskPageEntry{}
	for rows.Next() {
		var item TaskPageEntry
		var data []byte
		if err := rows.Scan(&item.Cursor, &data); err != nil {
			return nil, false, err
		}
		if err := json.Unmarshal(data, &item.Task); err != nil {
			return nil, false, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}
	more := len(items) > limit
	if more {
		items = items[:limit]
	}
	return items, more, nil
}

func (s *Store) listTasks(ctx context.Context, rootsOnly bool) ([]model.Task, error) {
	query := "SELECT document FROM tasks"
	if rootsOnly {
		query += " WHERE COALESCE(json_extract(document, '$.payload.transferParentId'), '')=''"
	}
	rows, err := s.DB.QueryContext(ctx, query+" ORDER BY rowid DESC LIMIT 1000")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.Task{}
	for rows.Next() {
		var t model.Task
		var b []byte
		if err := rows.Scan(&b); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, rows.Err()
}
func (s *Store) Pending(ctx context.Context) ([]model.Task, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT document FROM tasks WHERE state NOT IN ('SUCCEEDED','FAILED','CANCELLED','INTERRUPTED') ORDER BY rowid LIMIT 10000")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []model.Task{}
	for rows.Next() {
		var t model.Task
		var b []byte
		if err := rows.Scan(&b); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

func CanTransition(from, to model.TaskState) bool {
	if from.Terminal() {
		return false
	}
	if from == to {
		return true
	}
	if from == model.CancelRequested {
		return to == model.Cancelled || to == model.Succeeded || to == model.Failed || to == model.Interrupted || to == model.WaitingNode
	}
	switch to {
	case model.Failed, model.Interrupted, model.CancelRequested:
		return true
	case model.Cancelled:
		return from == model.Queued || from == model.WaitingClient || from == model.WaitingNode
	case model.Running:
		return from == model.Queued || from == model.WaitingNode || from == model.WaitingClient
	case model.Queued:
		return from == model.WaitingClient
	case model.WaitingNode, model.WaitingClient:
		return true
	case model.Succeeded:
		return from == model.Running
	default:
		return false
	}
}

// UpdateTask is an optimistic transition plus event in a single durable transaction.
func (s *Store) UpdateTask(ctx context.Context, id string, expected int64, state model.TaskState, phase string, result json.RawMessage, failure string) (model.Task, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return model.Task{}, err
	}
	defer tx.Rollback()
	var t model.Task
	var b []byte
	if err = tx.QueryRowContext(ctx, "SELECT document FROM tasks WHERE id=?", id).Scan(&b); err != nil {
		return t, err
	}
	if err = json.Unmarshal(b, &t); err != nil {
		return t, err
	}
	if t.Revision != expected {
		return t, ErrConflict
	}
	if !CanTransition(t.State, state) {
		return t, ErrTransition
	}
	t.State = state
	t.Phase = phase
	t.Result = result
	t.Error = failure
	t.Revision++
	t.UpdatedAt = time.Now().UTC()
	b, err = json.Marshal(t)
	if err != nil {
		return t, err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE tasks SET state=?,revision=?,document=? WHERE id=?", state, t.Revision, b, id); err != nil {
		return t, err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO task_events(task_id,revision,recorded_at,document) VALUES(?,?,?,?)", t.ID, t.Revision, t.UpdatedAt.UnixMilli(), b); err != nil {
		return t, err
	}
	if state.Terminal() {
		if _, err = tx.ExecContext(ctx, "INSERT INTO audit(actor_id,node_id,resource_key,action,request_id,result,recorded_at) VALUES(?,?,?,?,?,?,?)", t.ActorID, t.Resource.NodeID, t.Resource.Key(), t.Action, t.RequestID, string(state), t.UpdatedAt.UnixMilli()); err != nil {
			return t, err
		}
	}
	return t, tx.Commit()
}

type Event struct {
	Sequence int64      `json:"sequence"`
	Task     model.Task `json:"task"`
}

func (s *Store) LatestEventSequence(ctx context.Context) (int64, error) {
	var sequence int64
	err := s.DB.QueryRowContext(ctx, "SELECT COALESCE(MAX(sequence),0) FROM task_events").Scan(&sequence)
	return sequence, err
}

func (s *Store) Events(ctx context.Context, after int64, limit int) ([]Event, error) {
	if limit < 1 || limit > 256 {
		limit = 256
	}
	rows, err := s.DB.QueryContext(ctx, "SELECT sequence,document FROM task_events WHERE sequence>? ORDER BY sequence LIMIT ?", after, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Event{}
	for rows.Next() {
		var e Event
		var b []byte
		if err := rows.Scan(&e.Sequence, &b); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(b, &e.Task); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

// PutRecord implements conditional versioned records, including daemon run facts.
func (s *Store) PutRecord(ctx context.Context, namespace, id string, expected int64, value any) (int64, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return 0, err
	}
	if len(b) > 8*1024*1024 {
		return 0, errors.New("record exceeds metadata budget")
	}
	if expected == 0 {
		r, err := s.DB.ExecContext(ctx, "INSERT INTO records(namespace,id,revision,document) VALUES(?,?,1,?) ON CONFLICT DO NOTHING", namespace, id, b)
		if err != nil {
			return 0, err
		}
		n, _ := r.RowsAffected()
		if n != 1 {
			return 0, ErrConflict
		}
		return 1, nil
	}
	r, err := s.DB.ExecContext(ctx, "UPDATE records SET revision=revision+1,document=? WHERE namespace=? AND id=? AND revision=?", b, namespace, id, expected)
	if err != nil {
		return 0, err
	}
	n, _ := r.RowsAffected()
	if n != 1 {
		return 0, ErrConflict
	}
	return expected + 1, nil
}
func (s *Store) Record(ctx context.Context, namespace, id string, value any) (int64, error) {
	var rev int64
	var b []byte
	err := s.DB.QueryRowContext(ctx, "SELECT revision,document FROM records WHERE namespace=? AND id=?", namespace, id).Scan(&rev, &b)
	if err != nil {
		return 0, err
	}
	return rev, json.Unmarshal(b, value)
}
func (s *Store) RecordIDs(ctx context.Context, namespace string) ([]string, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT id FROM records WHERE namespace=?", namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	sort.Strings(items)
	return items, rows.Err()
}

// DeleteRecord removes a private durable record after its owner has completed
// the associated lifecycle. Callers that need compare-and-swap semantics keep
// the revision in the record itself; this helper is intentionally limited to
// internal cleanup where the namespace/id are already authenticated.
func (s *Store) DeleteRecord(ctx context.Context, namespace, id string) error {
	_, err := s.DB.ExecContext(ctx, "DELETE FROM records WHERE namespace=? AND id=?", namespace, id)
	return err
}
func (s *Store) Audit(ctx context.Context, actor string, r model.ResourceRef, action, request, result string) error {
	_, err := s.DB.ExecContext(ctx, "INSERT INTO audit(actor_id,node_id,resource_key,action,request_id,result,recorded_at) VALUES(?,?,?,?,?,?,?)", actor, r.NodeID, r.Key(), action, request, result, time.Now().UnixMilli())
	return err
}

// AuditFilter bounds an audit query. Empty string fields are ignored. The
// timestamps are inclusive and are compared against UTC millisecond storage.
type AuditFilter struct {
	ActorID     string
	NodeID      string
	ResourceKey string
	Action      string
	From        time.Time
	To          time.Time
	Limit       int
	Offset      int
}

// Audits returns newest audit records first. The query is deliberately
// read-only and uses parameterized predicates for every user supplied value.
func (s *Store) Audits(ctx context.Context, filter AuditFilter) ([]model.Audit, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 200 {
		limit = 200
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	query := "SELECT sequence,actor_id,node_id,resource_key,action,request_id,result,recorded_at FROM audit WHERE 1=1"
	args := make([]any, 0, 8)
	if filter.ActorID != "" {
		query += " AND actor_id=?"
		args = append(args, filter.ActorID)
	}
	if filter.NodeID != "" {
		query += " AND node_id=?"
		args = append(args, filter.NodeID)
	}
	if filter.ResourceKey != "" {
		query += " AND resource_key=?"
		args = append(args, filter.ResourceKey)
	}
	if filter.Action != "" {
		query += " AND action=?"
		args = append(args, filter.Action)
	}
	if !filter.From.IsZero() {
		query += " AND recorded_at>=?"
		args = append(args, filter.From.UnixMilli())
	}
	if !filter.To.IsZero() {
		query += " AND recorded_at<=?"
		args = append(args, filter.To.UnixMilli())
	}
	query += " ORDER BY recorded_at DESC, sequence DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.Audit, 0, limit)
	for rows.Next() {
		var item model.Audit
		var recorded int64
		if err := rows.Scan(&item.Sequence, &item.ActorID, &item.NodeID, &item.ResourceKey, &item.Action, &item.RequestID, &item.Result, &recorded); err != nil {
			return nil, err
		}
		item.RecordedAt = time.UnixMilli(recorded).UTC()
		items = append(items, item)
	}
	return items, rows.Err()
}

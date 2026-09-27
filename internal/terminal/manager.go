package terminal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"blora.dev/panel/internal/model"
)

var (
	ErrForbidden  = errors.New("terminal access denied")
	ErrNotFound   = errors.New("terminal session not found")
	ErrLease      = errors.New("terminal input and size lease belongs to another view or has expired")
	ErrInactive   = errors.New("terminal session is not running")
	ErrLimit      = errors.New("terminal session or attachment budget exhausted")
	ErrCapability = errors.New("terminal backend capability unavailable")
)

type Options struct {
	Root           string
	MaxSessions    int
	MaxAttachments int
	Archive        ArchiveOptions
	AllowNative    bool
	Backends       map[string]Backend
	// Authorize must check current resource grants. A nil callback only permits
	// the owner. Enabling native mode is an explicit administrator deployment
	// decision; API gateways must independently require host.manage on creation.
	Authorize func(context.Context, string, model.ResourceRef, string) error
}
type CreateRequest struct {
	ContainerID string            `json:"containerId,omitempty"`
	OwnerID     string            `json:"ownerId"`
	Resource    model.ResourceRef `json:"resource"`
	RunID       string            `json:"runId,omitempty"`
	Backend     string            `json:"backend"`
	Command     []string          `json:"command"`
	Directory   string            `json:"directory,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Cols        uint16            `json:"cols"`
	Rows        uint16            `json:"rows"`
}
type SpawnSpec struct {
	SessionID   string
	OwnerID     string // Trusted CreateRequest owner, never inferred from RunID.
	Resource    model.ResourceRef
	RunID       string
	Command     []string
	Directory   string
	Environment map[string]string
	Cols, Rows  uint16
}

// Backend implementations must honor ctx before launching. Once Spawn returns,
// the process outlives that request. Write and Close must be bounded; Close
// explicitly terminates only the spawned shell/exec, never its host instance.
type Backend interface {
	Spawn(context.Context, SpawnSpec) (Process, error)
}
type Process interface {
	io.Reader
	io.Writer
	Resize(cols, rows uint16) error
	Wait() error
	Close() error
}
type Lease struct {
	UserID    string    `json:"userId"`
	ViewID    string    `json:"viewId"`
	ExpiresAt time.Time `json:"expiresAt"`
}
type Session struct {
	ContainerID    string            `json:"containerId,omitempty"`
	ID             string            `json:"sessionId"`
	OwnerID        string            `json:"ownerId"`
	Resource       model.ResourceRef `json:"resource"`
	RunID          string            `json:"runId,omitempty"`
	Backend        string            `json:"backend"`
	State          string            `json:"state"`
	Diagnostic     string            `json:"diagnostic,omitempty"`
	Cols           uint16            `json:"cols"`
	Rows           uint16            `json:"rows"`
	CreatedAt      time.Time         `json:"createdAt"`
	EndedAt        *time.Time        `json:"endedAt,omitempty"`
	Lease          *Lease            `json:"lease,omitempty"`
	Archive        ArchiveOptions    `json:"archive"`
	MaxSessions    int               `json:"maxSessions"`
	MaxAttachments int               `json:"maxAttachments"`
}
type sessionRecord struct {
	Session Session         `json:"session"`
	Revoked map[string]bool `json:"revoked,omitempty"`
}
type liveSession struct {
	mu       sync.Mutex
	events   sync.Mutex
	session  Session
	revoked  map[string]bool
	views    map[string]string
	process  Process
	archive  *Archive
	done     chan struct{}
	stopOnce sync.Once
	stopDone chan struct{}
	stopErr  error
}
type Manager struct {
	mu       sync.Mutex
	options  Options
	sessions map[string]*liveSession
	closed   bool
}

func New(o Options) (*Manager, error) {
	if o.Root == "" {
		return nil, errors.New("terminal Root is required")
	}
	if o.MaxSessions == 0 {
		o.MaxSessions = 64
	}
	if o.MaxAttachments == 0 {
		o.MaxAttachments = 16
	}
	if o.MaxSessions < 1 || o.MaxSessions > 4096 || o.MaxAttachments < 1 || o.MaxAttachments > 256 {
		return nil, ErrLimit
	}
	var err error
	if o.Archive, err = archiveOptions(o.Archive); err != nil {
		return nil, err
	}
	if o.Root, err = filepath.Abs(o.Root); err != nil {
		return nil, err
	}
	if err = os.MkdirAll(o.Root, 0700); err != nil {
		return nil, err
	}
	backends := make(map[string]Backend, len(o.Backends))
	for k, v := range o.Backends {
		backends[k] = v
	}
	o.Backends = backends
	m := &Manager{options: o, sessions: map[string]*liveSession{}}
	entries, err := os.ReadDir(o.Root)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".json")
		if !safeID(id) || entry.Type()&os.ModeSymlink != 0 {
			return nil, errors.New("invalid terminal state filename")
		}
		b, err := os.ReadFile(filepath.Join(o.Root, entry.Name()))
		if err != nil {
			return nil, err
		}
		var record sessionRecord
		if err = json.Unmarshal(b, &record); err != nil {
			return nil, err
		}
		if record.Session.ID != id {
			return nil, errors.New("terminal state identity mismatch")
		}
		a, err := OpenArchive(filepath.Join(o.Root, id), o.Archive)
		if err != nil {
			return nil, err
		}
		s := &liveSession{session: record.Session, revoked: record.Revoked, views: map[string]string{}, archive: a, done: make(chan struct{}), stopDone: make(chan struct{})}
		if s.revoked == nil {
			s.revoked = map[string]bool{}
		}
		s.session.Lease = nil
		s.session.Archive = o.Archive
		s.session.MaxSessions = o.MaxSessions
		s.session.MaxAttachments = o.MaxAttachments
		if s.session.State == "running" || s.session.State == "starting" || s.session.State == "closing" {
			s.session.State = "invalid"
			s.session.Diagnostic = "Daemon restarted; original PTY cannot be reattached. No replacement shell was started."
			now := time.Now().UTC()
			s.session.EndedAt = &now
			if err = m.save(s); err != nil {
				return nil, err
			}
		}
		close(s.done)
		m.sessions[id] = s
	}
	if len(m.sessions) > o.MaxSessions {
		return nil, fmt.Errorf("retained terminal sessions exceed configured budget: %w", ErrLimit)
	}
	return m, nil
}

func safeID(id string) bool {
	if len(id) == 0 || len(id) > 96 {
		return false
	}
	for _, c := range id {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}
func (m *Manager) save(s *liveSession) error {
	b, err := json.Marshal(sessionRecord{Session: s.session, Revoked: s.revoked})
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(m.options.Root, ".terminal-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(b); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err = os.Rename(name, filepath.Join(m.options.Root, s.session.ID+".json")); err != nil {
		return err
	}
	return syncDirectory(m.options.Root)
}
func (m *Manager) find(id string) (*liveSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.sessions[id]
	if s == nil {
		return nil, ErrNotFound
	}
	return s, nil
}
func (m *Manager) authorize(ctx context.Context, s *liveSession, user, action string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if user == "" || s.revoked[user] {
		return ErrForbidden
	}
	if m.options.Authorize != nil {
		if err := m.options.Authorize(ctx, user, s.session.Resource, action); err != nil {
			return errors.Join(ErrForbidden, err)
		}
		return nil
	}
	if user != s.session.OwnerID {
		return ErrForbidden
	}
	return nil
}
func (m *Manager) Create(ctx context.Context, r CreateRequest) (Session, error) {
	if !safeID(r.OwnerID) || r.Resource.Kind == "" || r.Resource.ID == "" || len(r.Command) == 0 || r.Command[0] == "" {
		return Session{}, errors.New("terminal owner, resource, and command are required")
	}
	if r.Cols == 0 {
		r.Cols = 100
	}
	if r.Rows == 0 {
		r.Rows = 28
	}
	if !validSize(r.Cols, r.Rows) {
		return Session{}, errors.New("invalid terminal dimensions")
	}
	for _, v := range r.Command {
		if strings.ContainsRune(v, 0) {
			return Session{}, errors.New("invalid command")
		}
	}
	for k, v := range r.Environment {
		if k == "" || strings.ContainsAny(k, "=\x00") || strings.ContainsRune(v, 0) {
			return Session{}, errors.New("invalid environment variable")
		}
	}
	if r.Backend == "" {
		r.Backend = "native"
	}
	var backend Backend
	if r.Backend == "native" {
		if !m.options.AllowNative {
			return Session{}, ErrForbidden
		}
		backend = nativeBackend{}
	} else {
		backend = m.options.Backends[r.Backend]
	}
	if backend == nil {
		return Session{}, ErrCapability
	}
	s := &liveSession{session: Session{ContainerID: r.ContainerID, ID: model.ID(), OwnerID: r.OwnerID, Resource: r.Resource, RunID: r.RunID, Backend: r.Backend, State: "starting", Cols: r.Cols, Rows: r.Rows, CreatedAt: time.Now().UTC(), Archive: m.options.Archive, MaxSessions: m.options.MaxSessions, MaxAttachments: m.options.MaxAttachments}, revoked: map[string]bool{}, views: map[string]string{}, done: make(chan struct{}), stopDone: make(chan struct{})}
	action := "terminal.input"
	if r.Backend == "native" {
		action = "host.manage"
	}
	if err := m.authorize(ctx, s, r.OwnerID, action); err != nil {
		return Session{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return Session{}, os.ErrClosed
	}
	if len(m.sessions) >= m.options.MaxSessions {
		m.mu.Unlock()
		return Session{}, ErrLimit
	}
	m.sessions[s.session.ID] = s
	m.mu.Unlock()
	a, err := OpenArchive(filepath.Join(m.options.Root, s.session.ID), m.options.Archive)
	if err == nil {
		s.archive = a
		err = m.save(s)
	}
	if err == nil {
		err = a.AppendResize(r.Cols, r.Rows)
	}
	if err == nil {
		s.process, err = backend.Spawn(ctx, SpawnSpec{SessionID: s.session.ID, OwnerID: r.OwnerID, Resource: r.Resource, RunID: r.RunID, Command: append([]string(nil), r.Command...), Directory: r.Directory, Environment: r.Environment, Cols: r.Cols, Rows: r.Rows})
	}
	if err != nil {
		s.session.State = "failed"
		s.session.Diagnostic = err.Error()
		now := time.Now().UTC()
		s.session.EndedAt = &now
		_ = m.save(s)
		close(s.done)
		return s.session, err
	}
	s.session.State = "running"
	if err = m.save(s); err != nil {
		s.session.State = "invalid"
		s.session.Diagnostic = "launch recording failed"
		// Drain ConPTY output while closing even when recording failed.
		go m.run(s)
		m.stop(s)
		return s.session, err
	}
	go m.run(s)
	return snapshot(s), nil
}

func snapshot(s *liveSession) Session {
	v := s.session
	if v.Lease != nil {
		copy := *v.Lease
		v.Lease = &copy
	}
	if v.EndedAt != nil {
		copy := *v.EndedAt
		v.EndedAt = &copy
	}
	return v
}
func (m *Manager) run(s *liveSession) {
	readDone := make(chan error, 1)
	go func() {
		buf := make([]byte, MaxEventBytes)
		for {
			n, err := s.process.Read(buf)
			if n > 0 {
				s.events.Lock()
				archiveErr := s.archive.AppendOutput(buf[:n])
				s.events.Unlock()
				if archiveErr != nil {
					readDone <- archiveErr
					// ConPTY can emit while closing. Keep draining with constant
					// memory after reporting the archival failure to the owner.
					_, _ = io.Copy(io.Discard, s.process)
					return
				}
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					err = nil
				}
				readDone <- err
				return
			}
		}
	}()
	waitDone := make(chan error, 1)
	go func() { waitDone <- s.process.Wait() }()
	var waitErr, readErr error
	waitSeen, readSeen := false, false
	select {
	case waitErr = <-waitDone:
		waitSeen = true
		// Pipe ownership and process Wait are independent. A child retaining
		// output cannot keep session completion stuck without a deadline.
		select {
		case readErr = <-readDone:
			readSeen = true
		case <-time.After(2 * time.Second):
		}
	case readErr = <-readDone:
		readSeen = true
		// Read failure must not become silent output loss. End this PTY only.
	}
	m.stop(s)
	var cleanupErr error
	select {
	case <-s.stopDone:
	case <-time.After(5 * time.Second):
		cleanupErr = errors.New("PTY cleanup deadline exceeded")
	}
	if !waitSeen {
		select {
		case waitErr = <-waitDone:
		case <-time.After(2 * time.Second):
			cleanupErr = errors.Join(cleanupErr, errors.New("PTY process exit was not confirmed"))
		}
	}
	if !readSeen {
		select {
		case readErr = <-readDone:
		case <-time.After(2 * time.Second):
			cleanupErr = errors.Join(cleanupErr, errors.New("PTY output drain was not confirmed"))
		}
	}
	s.mu.Lock()
	if s.session.State != "invalid" {
		s.session.State = "exited"
		if cleanupErr != nil || s.stopErr != nil {
			s.session.State = "invalid"
		}
		if err := errors.Join(readErr, waitErr, cleanupErr, s.stopErr); err != nil {
			s.session.Diagnostic = err.Error()
		}
	}
	s.session.Lease = nil
	now := time.Now().UTC()
	s.session.EndedAt = &now
	if err := m.save(s); err != nil {
		s.session.Diagnostic = "terminal exit recording failed: " + err.Error()
	}
	s.mu.Unlock()
	close(s.done)
}
func (m *Manager) stop(s *liveSession) {
	s.stopOnce.Do(func() {
		go func() { err := s.process.Close(); s.mu.Lock(); s.stopErr = err; s.mu.Unlock(); close(s.stopDone) }()
	})
}
func (m *Manager) Get(ctx context.Context, id, user string) (Session, error) {
	s, err := m.find(id)
	if err != nil {
		return Session{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err = m.authorize(ctx, s, user, "terminal.read"); err != nil {
		return Session{}, err
	}
	return snapshot(s), nil
}
func (m *Manager) List(ctx context.Context, user string) ([]Session, error) {
	m.mu.Lock()
	ids := make([]string, 0, len(m.sessions))
	for id := range m.sessions {
		ids = append(ids, id)
	}
	m.mu.Unlock()
	result := []Session{}
	for _, id := range ids {
		s, err := m.Get(ctx, id, user)
		if err == nil {
			result = append(result, s)
		} else if !errors.Is(err, ErrForbidden) && !errors.Is(err, ErrNotFound) {
			return nil, err
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.Before(result[j].CreatedAt) })
	return result, nil
}
func (m *Manager) Attach(ctx context.Context, id, user, view string, after uint64, maxBytes int) (Batch, error) {
	if !safeID(view) {
		return Batch{}, errors.New("invalid terminal view identity")
	}
	s, err := m.find(id)
	if err != nil {
		return Batch{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err = m.authorize(ctx, s, user, "terminal.read"); err != nil {
		return Batch{}, err
	}
	if owner, ok := s.views[view]; ok && owner != user {
		return Batch{}, ErrForbidden
	}
	if _, ok := s.views[view]; !ok && len(s.views) >= m.options.MaxAttachments {
		return Batch{}, ErrLimit
	}
	if s.archive == nil {
		return Batch{}, ErrInactive
	}
	b, err := s.archive.Read(after, maxBytes)
	if err != nil {
		return b, err
	}
	s.views[view] = user
	return b, nil
}
func (m *Manager) Read(ctx context.Context, id, user, view string, after uint64, maxBytes int) (Batch, error) {
	s, err := m.find(id)
	if err != nil {
		return Batch{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err = m.authorize(ctx, s, user, "terminal.read"); err != nil {
		return Batch{}, err
	}
	if s.views[view] != user {
		return Batch{}, ErrForbidden
	}
	if s.archive == nil {
		return Batch{}, ErrInactive
	}
	return s.archive.Read(after, maxBytes)
}
func (m *Manager) Detach(id, user, view string) error {
	s, err := m.find(id)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.views[view] != user {
		return ErrForbidden
	}
	delete(s.views, view)
	if l := s.session.Lease; l != nil && l.UserID == user && l.ViewID == view {
		s.session.Lease = nil
	}
	return nil
}
func (m *Manager) AcquireLease(id, user, view string, ttl time.Duration, takeover bool) (Lease, error) {
	s, err := m.find(id)
	if err != nil {
		return Lease{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err = m.authorize(context.Background(), s, user, "terminal.input"); err != nil {
		return Lease{}, err
	}
	if s.views[view] != user {
		return Lease{}, ErrForbidden
	}
	if s.session.State != "running" {
		return Lease{}, ErrInactive
	}
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	if ttl > 2*time.Minute {
		ttl = 2 * time.Minute
	}
	now := time.Now().UTC()
	old := s.session.Lease
	if old != nil && old.ExpiresAt.After(now) && (old.UserID != user || old.ViewID != view) && !takeover {
		return *old, ErrLease
	}
	l := Lease{UserID: user, ViewID: view, ExpiresAt: now.Add(ttl)}
	s.session.Lease = &l
	return l, nil
}
func (m *Manager) checkInput(ctx context.Context, s *liveSession, user, view string) error {
	if err := m.authorize(ctx, s, user, "terminal.input"); err != nil {
		return err
	}
	if s.views[view] != user {
		return ErrForbidden
	}
	if s.session.State != "running" {
		return ErrInactive
	}
	l := s.session.Lease
	if l == nil || l.UserID != user || l.ViewID != view || !l.ExpiresAt.After(time.Now()) {
		return ErrLease
	}
	return nil
}
func (m *Manager) WriteInput(ctx context.Context, id, user, view string, data []byte) error {
	if len(data) > MaxEventBytes {
		return errors.New("terminal input frame exceeds 32 KiB")
	}
	s, err := m.find(id)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err = m.checkInput(ctx, s, user, view); err != nil {
		return err
	}
	n, err := s.process.Write(data)
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	return err
}
func (m *Manager) Resize(ctx context.Context, id, user, view string, cols, rows uint16) error {
	if !validSize(cols, rows) {
		return errors.New("invalid terminal dimensions")
	}
	s, err := m.find(id)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err = m.checkInput(ctx, s, user, view); err != nil {
		return err
	}
	s.events.Lock()
	defer s.events.Unlock()
	if err = s.process.Resize(cols, rows); err != nil {
		return err
	}
	if err = s.archive.AppendResize(cols, rows); err != nil {
		s.session.State = "invalid"
		s.session.Diagnostic = "resize succeeded but recovery archive failed"
		_ = m.save(s)
		m.stop(s)
		return err
	}
	s.session.Cols = cols
	s.session.Rows = rows
	return m.save(s)
}

// Revoke is called by the trusted authorization gateway. It synchronously
// removes all user views and the writer lease, and prevents stale reattaches.
func (m *Manager) Revoke(id, user string) error {
	s, err := m.find(id)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revoked[user] = true
	for view, owner := range s.views {
		if owner == user {
			delete(s.views, view)
		}
	}
	if s.session.Lease != nil && s.session.Lease.UserID == user {
		s.session.Lease = nil
	}
	return m.save(s)
}
func (m *Manager) RestoreAccess(id, user string) error {
	s, err := m.find(id)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.revoked, user)
	return m.save(s)
}
func (m *Manager) CloseSession(ctx context.Context, id, user string) error {
	s, err := m.find(id)
	if err != nil {
		return err
	}
	s.mu.Lock()
	if err = m.authorize(ctx, s, user, "terminal.input"); err != nil {
		s.mu.Unlock()
		return err
	}
	if s.process == nil && s.session.State == "invalid" {
		s.mu.Unlock()
		return ErrInactive
	}
	if s.process == nil || s.session.State == "exited" || s.session.State == "failed" {
		s.mu.Unlock()
		return nil
	}
	s.session.State = "closing"
	s.session.Lease = nil
	err = m.save(s)
	s.mu.Unlock()
	if err != nil {
		return err
	}
	m.stop(s)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.done:
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.session.State == "invalid" {
			return fmt.Errorf("terminal cleanup could not be confirmed: %s: %w", s.session.Diagnostic, ErrInactive)
		}
		return s.stopErr
	}
}

// DeleteSession explicitly forgets a finished session and its retained output,
// freeing the published session/disk budget. Running PTYs cannot be deleted.
func (m *Manager) DeleteSession(ctx context.Context, id, user string) error {
	s, err := m.find(id)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err = m.authorize(ctx, s, user, "terminal.input"); err != nil {
		return err
	}
	select {
	case <-s.done:
	default:
		return ErrInactive
	}
	if err = os.Remove(filepath.Join(m.options.Root, id+".json")); err != nil {
		return err
	}
	if s.archive != nil {
		_ = s.archive.Close()
	}
	if err = os.RemoveAll(filepath.Join(m.options.Root, id)); err != nil {
		return err
	}
	m.mu.Lock()
	delete(m.sessions, id)
	m.mu.Unlock()
	return nil
}

// Close is explicit daemon maintenance, with a common deadline. View Detach
// never calls it. Retained metadata and raw archives remain available.
func (m *Manager) Close() error {
	m.mu.Lock()
	m.closed = true
	all := make([]*liveSession, 0, len(m.sessions))
	for _, s := range m.sessions {
		all = append(all, s)
	}
	m.mu.Unlock()
	for _, s := range all {
		s.mu.Lock()
		if s.process != nil {
			m.stop(s)
		}
		s.mu.Unlock()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var result error
	for _, s := range all {
		select {
		case <-s.done:
		case <-ctx.Done():
			result = errors.Join(result, ctx.Err())
			continue
		}
		s.mu.Lock()
		result = errors.Join(result, s.stopErr)
		if s.archive != nil {
			_ = s.archive.Close()
		}
		s.mu.Unlock()
	}
	return result
}

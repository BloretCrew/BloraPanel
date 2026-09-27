package terminal

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
)

type controlledBackend struct{ p *controlledProcess }

func (b controlledBackend) Spawn(context.Context, SpawnSpec) (Process, error) { return b.p, nil }

type controlledProcess struct {
	r          *io.PipeReader
	w          *io.PipeWriter
	mu         sync.Mutex
	input      bytes.Buffer
	done       chan struct{}
	once       sync.Once
	closeError error
}

func newControlledProcess() *controlledProcess {
	r, w := io.Pipe()
	return &controlledProcess{r: r, w: w, done: make(chan struct{})}
}
func (p *controlledProcess) Read(b []byte) (int, error) { return p.r.Read(b) }
func (p *controlledProcess) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.input.Write(b)
}
func (p *controlledProcess) Resize(uint16, uint16) error { return nil }
func (p *controlledProcess) Wait() error                 { <-p.done; return nil }
func (p *controlledProcess) Close() error {
	p.once.Do(func() { p.w.Close(); close(p.done) })
	return p.closeError
}
func (p *controlledProcess) writes() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.input.String()
}
func testRequest() CreateRequest {
	return CreateRequest{OwnerID: "owner", Resource: model.ResourceRef{Kind: "instance", ID: "instance-one", NodeID: "node-one"}, Backend: "test", Command: []string{"test-shell"}, Cols: 80, Rows: 24}
}

func TestManagerLeaseObserversDetachRevoke(t *testing.T) {
	ctx := context.Background()
	p := newControlledProcess()
	m, err := New(Options{Root: t.TempDir(), Backends: map[string]Backend{"test": controlledBackend{p}}, Authorize: func(_ context.Context, user string, _ model.ResourceRef, action string) error {
		if user == "reader" && action != "terminal.read" {
			return ErrForbidden
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	s, err := m.Create(ctx, testRequest())
	if err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]string{{"owner", "a"}, {"owner", "b"}, {"reader", "watch"}, {"second", "c"}} {
		if _, err = m.Attach(ctx, s.ID, pair[0], pair[1], 0, 0); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = m.AcquireLease(s.ID, "reader", "watch", time.Minute, true); !errors.Is(err, ErrForbidden) {
		t.Fatal("reader obtained write")
	}
	if _, err = m.AcquireLease(s.ID, "owner", "a", time.Minute, false); err != nil {
		t.Fatal(err)
	}
	if _, err = m.AcquireLease(s.ID, "second", "c", time.Minute, false); !errors.Is(err, ErrLease) {
		t.Fatal("concurrent lease accepted")
	}
	if err = m.WriteInput(ctx, s.ID, "owner", "a", []byte("secret-input")); err != nil {
		t.Fatal(err)
	}
	if _, err = m.AcquireLease(s.ID, "second", "c", time.Minute, true); err != nil {
		t.Fatal(err)
	}
	if err = m.WriteInput(ctx, s.ID, "owner", "a", []byte("stale")); !errors.Is(err, ErrLease) {
		t.Fatal("old lease retained")
	}
	if err = m.Resize(ctx, s.ID, "second", "c", 132, 42); err != nil {
		t.Fatal(err)
	}
	if err = m.Detach(s.ID, "second", "c"); err != nil {
		t.Fatal(err)
	}
	select {
	case <-p.done:
		t.Fatal("view close killed PTY")
	default:
	}
	if _, err = m.Attach(ctx, s.ID, "second", "c", 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err = m.AcquireLease(s.ID, "second", "c", time.Millisecond, false); err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Millisecond)
	if err = m.WriteInput(ctx, s.ID, "second", "c", []byte("expired")); !errors.Is(err, ErrLease) {
		t.Fatal("expired lease accepted")
	}
	if err = m.Revoke(s.ID, "owner"); err != nil {
		t.Fatal(err)
	}
	if _, err = m.Read(ctx, s.ID, "owner", "a", 0, 0); !errors.Is(err, ErrForbidden) {
		t.Fatal("revoked stream readable")
	}
	if _, err = m.Attach(ctx, s.ID, "owner", "new", 0, 0); !errors.Is(err, ErrForbidden) {
		t.Fatal("revoked attach accepted")
	}
	b, err := m.Read(ctx, s.ID, "reader", "watch", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Events) != 2 || b.Events[1].Kind != "resize" {
		t.Fatalf("input was archived or resize missing: %+v", b)
	}
	if p.writes() != "secret-input" {
		t.Fatalf("unexpected write/replay: %q", p.writes())
	}
}

func TestManagerBudgetsIdentityAndNativeGate(t *testing.T) {
	ctx := context.Background()
	p := newControlledProcess()
	m, err := New(Options{Root: t.TempDir(), MaxSessions: 1, MaxAttachments: 1, Backends: map[string]Backend{"test": controlledBackend{p}}})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	r := testRequest()
	r.Backend = "native"
	if _, err = m.Create(ctx, r); !errors.Is(err, ErrForbidden) {
		t.Fatal("default allowed host shell")
	}
	s, err := m.Create(ctx, testRequest())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.Create(ctx, testRequest()); !errors.Is(err, ErrLimit) {
		t.Fatal("session limit ignored")
	}
	if _, err = m.Get(ctx, s.ID, "other"); !errors.Is(err, ErrForbidden) {
		t.Fatal("session identity leaked")
	}
	if _, err = m.Attach(ctx, s.ID, "owner", "view", 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err = m.Attach(ctx, s.ID, "owner", "view-two", 0, 0); !errors.Is(err, ErrLimit) {
		t.Fatal("attachment limit ignored")
	}
	if err = m.DeleteSession(ctx, s.ID, "owner"); !errors.Is(err, ErrInactive) {
		t.Fatal("running PTY deleted")
	}
	if err = m.CloseSession(ctx, s.ID, "owner"); err != nil {
		t.Fatal(err)
	}
	if err = m.DeleteSession(ctx, s.ID, "owner"); err != nil {
		t.Fatal(err)
	}
	if _, err = m.Get(ctx, s.ID, "owner"); !errors.Is(err, ErrNotFound) {
		t.Fatal("deleted session retained")
	}
}

func TestManagerDaemonRestartInvalidatesWithoutSpawn(t *testing.T) {
	root := t.TempDir()
	p := newControlledProcess()
	m, err := New(Options{Root: root, Backends: map[string]Backend{"test": controlledBackend{p}}})
	if err != nil {
		t.Fatal(err)
	}
	s, err := m.Create(context.Background(), testRequest())
	if err != nil {
		t.Fatal(err)
	}
	// Capture persisted running state to simulate the daemon crashing. Keep the
	// fake process's resources separate; no OS process is orphaned by this test.
	live, _ := m.find(s.ID)
	live.mu.Lock()
	record := snapshot(live)
	live.mu.Unlock()
	if err = m.Close(); err != nil {
		t.Fatal(err)
	}
	live.mu.Lock()
	live.session = record
	if err = m.save(live); err != nil {
		t.Fatal(err)
	}
	live.mu.Unlock()
	recovered, err := New(Options{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	defer recovered.Close()
	got, err := recovered.Get(context.Background(), s.ID, "owner")
	if err != nil {
		t.Fatal(err)
	}
	if got.State != "invalid" || got.ID != s.ID || got.Lease != nil {
		t.Fatalf("not invalidated: %+v", got)
	}
	if _, err = recovered.Attach(context.Background(), s.ID, "owner", "restored", 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err = recovered.AcquireLease(s.ID, "owner", "restored", time.Minute, false); !errors.Is(err, ErrInactive) {
		t.Fatal("invalid session writable")
	}
	if err = recovered.CloseSession(context.Background(), s.ID, "owner"); !errors.Is(err, ErrInactive) {
		t.Fatal("lost PTY pretended to close successfully")
	}
}

func TestManagerRevocationPersistsAndArchiveFailureIsVisible(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	p := newControlledProcess()
	o := Options{Root: root, Backends: map[string]Backend{"test": controlledBackend{p}}, Authorize: func(context.Context, string, model.ResourceRef, string) error { return nil }}
	m, err := New(o)
	if err != nil {
		t.Fatal(err)
	}
	s, err := m.Create(ctx, testRequest())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.Attach(ctx, s.ID, "observer", "view", 0, 0); err != nil {
		t.Fatal(err)
	}
	if err = m.Revoke(s.ID, "observer"); err != nil {
		t.Fatal(err)
	}
	live, _ := m.find(s.ID)
	live.archive.Close()
	if _, err = p.w.Write([]byte("cannot-be-archived")); err != nil {
		t.Fatal(err)
	}
	select {
	case <-live.done:
	case <-time.After(3 * time.Second):
		t.Fatal("archive failure left PTY running")
	}
	got, err := m.Get(ctx, s.ID, "owner")
	if err != nil || got.Diagnostic == "" || got.State == "running" {
		t.Fatalf("archive failure hidden: %+v %v", got, err)
	}
	if err = m.Close(); err != nil {
		t.Fatal(err)
	}
	o.Backends = nil
	recovered, err := New(o)
	if err != nil {
		t.Fatal(err)
	}
	defer recovered.Close()
	if _, err = recovered.Attach(ctx, s.ID, "observer", "stale", 0, 0); !errors.Is(err, ErrForbidden) {
		t.Fatal("revocation forgotten after restart")
	}
	if err = recovered.RestoreAccess(s.ID, "observer"); err != nil {
		t.Fatal(err)
	}
	if _, err = recovered.Attach(ctx, s.ID, "observer", "new", 0, 0); err != nil {
		t.Fatal(err)
	}
}

func TestManagerCloseFailureDoesNotReportSuccess(t *testing.T) {
	p := newControlledProcess()
	p.closeError = errors.New("backend could not confirm exec descendants exited")
	m, err := New(Options{Root: t.TempDir(), Backends: map[string]Backend{"test": controlledBackend{p}}})
	if err != nil {
		t.Fatal(err)
	}
	s, err := m.Create(context.Background(), testRequest())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err = m.CloseSession(ctx, s.ID, "owner"); !errors.Is(err, ErrInactive) {
		t.Fatalf("unconfirmed close returned success: %v", err)
	}
	got, err := m.Get(ctx, s.ID, "owner")
	if err != nil || got.State != "invalid" || got.Diagnostic == "" {
		t.Fatalf("unconfirmed close hidden: %+v %v", got, err)
	}
	if err = m.Close(); err == nil {
		t.Fatal("maintenance hid cleanup failure")
	}
}

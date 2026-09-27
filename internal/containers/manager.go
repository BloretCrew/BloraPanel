package containers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Manager struct {
	options             Options
	engine              *engine
	mu                  sync.Mutex
	locks               map[string]chan struct{}
	active              map[string]bool
	cliRecoveryRequired bool
	token               string
}
type journal struct {
	ActorID   string          `json:"actorId"`
	Digest    string          `json:"digest"`
	Operation Operation       `json:"operation"`
	Result    Result          `json:"result"`
	Outputs   []CommandOutput `json:"outputs,omitempty"`
}

// RecoverCLI must run before accepting work after an executor restart.
// Failure blocks startup: a prior CLI may still mutate the remote Engine.
func (m *Manager) RecoverCLI(ctx context.Context) error {
	err := recoverCLI(ctx, m.options.StateRoot)
	m.mu.Lock()
	m.cliRecoveryRequired = err != nil
	m.mu.Unlock()
	return err
}

func (m *Manager) runCLI(ctx context.Context, command []string, dir string, env []string, stdout, stderr io.Writer) error {
	err := runCLI(ctx, command, dir, m.options.StateRoot, env, stdout, stderr)
	if errors.Is(err, ErrUnknown) {
		m.mu.Lock()
		m.cliRecoveryRequired = true
		m.mu.Unlock()
	}
	return err
}

type execution struct {
	manager *Manager
	actor   string
	ctx     context.Context
	entry   journal
	notify  func(Progress) error
}

func New(options Options) (*Manager, error) {
	if options.StateRoot == "" {
		return nil, errors.New("container management StateRoot is required")
	}
	root, err := filepath.Abs(options.StateRoot)
	if err != nil {
		return nil, err
	}
	options.StateRoot = root
	if options.OperationTimeout <= 0 || options.OperationTimeout > time.Hour {
		options.OperationTimeout = 10 * time.Minute
	}
	if options.MaxProjects <= 0 {
		options.MaxProjects = 128
	}
	if options.MaxOperations <= 0 {
		options.MaxOperations = 10000
	}
	if options.MaxProjectSaves <= 0 {
		options.MaxProjectSaves = 1024
	}
	for _, dir := range []string{root, filepath.Join(root, "operations"), filepath.Join(root, "projects"), filepath.Join(root, "project-saves"), filepath.Join(root, "docker-client")} {
		if err = os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
	}
	tokenPath := filepath.Join(root, "owner")
	b, err := os.ReadFile(tokenPath)
	if errors.Is(err, os.ErrNotExist) {
		token := randomID()
		f, e := os.OpenFile(tokenPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			return nil, e
		}
		_, e = f.WriteString(token)
		if e == nil {
			e = f.Sync()
		}
		ce := f.Close()
		if e != nil {
			return nil, e
		}
		if ce != nil {
			return nil, ce
		}
		b = []byte(token)
	} else if err != nil {
		return nil, err
	}
	if len(b) != 32 {
		return nil, ErrIdentity
	}
	m := &Manager{options: options, locks: make(map[string]chan struct{}), active: make(map[string]bool), token: string(b)}
	if options.Endpoint != "" {
		m.engine, err = newEngine(options.Endpoint, options.ComposeTLSCertPath)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}
func randomID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}
func (m *Manager) Close() error {
	if m.engine != nil {
		m.engine.client.CloseIdleConnections()
	}
	return nil
}
func (m *Manager) authorize(ctx context.Context, actor string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if actor == "" || m.options.Authorize == nil {
		return ErrForbidden
	}
	return m.options.Authorize(ctx, actor)
}
func (m *Manager) requireEngine() error {
	if m.engine == nil {
		return fmt.Errorf("no Engine endpoint configured: %w", ErrCapability)
	}
	return nil
}
func (m *Manager) lock(ctx context.Context, key string) (func(), error) {
	m.mu.Lock()
	ch := m.locks[key]
	if ch == nil {
		ch = make(chan struct{}, 1)
		m.locks[key] = ch
	}
	m.mu.Unlock()
	select {
	case ch <- struct{}{}:
		return func() { <-ch }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
func atomicJSON(path string, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return atomicFile(path, b)
}
func atomicFile(path string, b []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".pending-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(b); err == nil {
		err = f.Sync()
	}
	ce := f.Close()
	if err != nil {
		return err
	}
	if ce != nil {
		return ce
	}
	if err = os.Rename(f.Name(), path); err != nil {
		return err
	}
	return syncDir(filepath.Dir(path))
}
func (m *Manager) operationPath(id string) string {
	return filepath.Join(m.options.StateRoot, "operations", id+".json")
}
func (m *Manager) loadOperation(id string) (journal, error) {
	var j journal
	if !safeID.MatchString(id) {
		return j, ErrIdentity
	}
	b, err := os.ReadFile(m.operationPath(id))
	if err != nil {
		return j, err
	}
	if len(b) > 8<<20 {
		return j, ErrIdentity
	}
	err = json.Unmarshal(b, &j)
	if err == nil && j.Result.TaskID != id {
		err = ErrIdentity
	}
	return j, err
}
func (x *execution) save() error {
	b, err := json.Marshal(x.entry)
	if err != nil {
		return err
	}
	if len(b) > 8<<20 {
		return errors.New("operation journal exceeds 8 MiB")
	}
	return atomicFile(x.manager.operationPath(x.entry.Result.TaskID), b)
}
func (x *execution) emit(phase, message string) error {
	return x.emitProgress(Progress{Phase: phase, Message: message})
}

func (x *execution) emitProgress(p Progress) error {
	if err := x.manager.authorize(x.ctx, x.actor); err != nil {
		return err
	}
	r := &x.entry.Result
	seq := uint64(1)
	if len(r.Progress) > 0 {
		seq = r.Progress[len(r.Progress)-1].Sequence + 1
	}
	if len(p.Message) > 2048 {
		p.Message = p.Message[:2048]
	}
	if len(p.Layer) > 128 {
		p.Layer = p.Layer[:128]
	}
	// Invalid or imprecise counters must not become a fabricated percentage.
	if p.Current < 0 || p.Total < 0 || p.Current > 9007199254740991 || p.Total > 9007199254740991 {
		p.Current = 0
		p.Total = 0
	}
	p.Sequence = seq
	p.At = time.Now().UTC()
	r.Phase = p.Phase
	r.Progress = append(r.Progress, p)
	if len(r.Progress) > 256 {
		r.Progress = r.Progress[len(r.Progress)-256:]
	}
	if err := x.save(); err != nil {
		return err
	}
	if x.notify != nil {
		return x.notify(p)
	}
	return nil
}
func (m *Manager) labels(taskID string) map[string]string {
	labels := map[string]string{}
	for k, v := range m.options.Labels {
		labels[k] = v
	}
	labels[labelPrefix+"owner"] = m.token
	labels[labelPrefix+"operation"] = taskID
	return labels
}

// Query requires node host.manage even for lists; ordinary instance permissions
// do not grant a view of unrelated containers, images, mounts or Compose source.
func (m *Manager) Query(ctx context.Context, actor string, q Query) (Snapshot, error) {
	s := Snapshot{ObservedAt: time.Now().UTC()}
	if err := m.authorize(ctx, actor); err != nil {
		return s, err
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	switch q.Kind {
	case "capabilities":
		s.Capabilities = m.capabilities(ctx)
		return s, nil
	case "projects":
		p, err := m.projects()
		s.Projects = p
		return s, err
	case "project":
		p, err := m.loadProject(q.ProjectID, false)
		if err != nil {
			return s, err
		}
		s.Project = &p
		if m.engine != nil {
			s.Items, s.Truncated, err = m.engine.list(ctx, "containers", map[string]string{"com.docker.compose.project": p.EngineName}, q.Limit)
		}
		return s, err
	case "operation":
		j, err := m.loadOperation(q.TaskID)
		if err != nil {
			return s, err
		}
		s.Operation = &j.Result
		m.mu.Lock()
		active := m.active[q.TaskID]
		m.mu.Unlock()
		if j.Result.State == "RUNNING" && !active {
			s.Operation.State = "INTERRUPTED"
			s.Operation.Unknown = true
			s.Operation.Diagnostic = "operation was interrupted; inspect actual resources before a new attempt"
		}
		return s, nil
	case "operation-output":
		if q.OutputOffset < 0 {
			return s, errors.New("invalid output offset")
		}
		j, err := m.loadOperation(q.TaskID)
		if err != nil {
			return s, err
		}
		s.NextOutput = -1
		if q.OutputOffset < len(j.Outputs) {
			s.Output = &j.Outputs[q.OutputOffset]
			if q.OutputOffset+1 < len(j.Outputs) {
				s.NextOutput = q.OutputOffset + 1
			}
		}
		return s, nil
	case "containers", "images", "volumes", "networks":
		if err := m.requireEngine(); err != nil {
			return s, err
		}
		items, truncated, err := m.engine.list(ctx, q.Kind, nil, q.Limit)
		s.Items = items
		s.Truncated = truncated
		return s, err
	case "container", "image", "volume", "network":
		if err := m.requireEngine(); err != nil {
			return s, err
		}
		if q.Target.Kind != q.Kind {
			return s, ErrIdentity
		}
		o, err := m.engine.inspect(ctx, q.Target, q.Details)
		if err == nil {
			s.Items = []Object{o}
		}
		return s, err
	default:
		return s, errors.New("unsupported container query")
	}
}

// Execute is called by the persistent daemon worker, never a browser handler.
// Its stable task ID is claimed before side effects. Repeating an interrupted
// operation never replays a create/delete/apply command.
func (m *Manager) Execute(ctx context.Context, actor string, op Operation, notify func(Progress) error) (result Result, returnErr error) {
	if err := m.authorize(ctx, actor); err != nil {
		return result, err
	}
	if !safeID.MatchString(op.TaskID) {
		return result, ErrIdentity
	}
	if len(op.Config) > 2<<20 {
		return result, errors.New("create configuration exceeds 2 MiB")
	}
	ctx, cancel := context.WithTimeout(ctx, m.options.OperationTimeout)
	defer cancel()
	// Compose may mutate several containers, networks and named volumes in one
	// invocation. Serialize privileged Engine mutations together so a direct
	// resource action cannot race that same resource through another view.
	// Ordinary isolated-instance control uses its separate runtime adapter.
	unlock, err := m.lock(ctx, "engine-mutations")
	if err != nil {
		return result, err
	}
	defer unlock()
	digest := fingerprint(struct {
		Actor     string
		Operation Operation
	}{actor, op})
	m.mu.Lock()
	cliRecoveryRequired := m.cliRecoveryRequired
	m.mu.Unlock()
	if cliRecoveryRequired {
		return result, fmt.Errorf("Compose CLI exit remains unconfirmed; restart recovery is required before Engine mutations: %w", ErrUnknown)
	}
	if prior, e := m.loadOperation(op.TaskID); e == nil {
		if prior.Digest != digest || prior.ActorID != actor {
			return prior.Result, ErrConflict
		}
		if prior.Result.State == "SUCCEEDED" {
			return prior.Result, nil
		}
		if prior.Result.State == "RUNNING" || prior.Result.Unknown {
			return prior.Result, ErrUnknown
		}
		return prior.Result, errors.New(prior.Result.Diagnostic)
	} else if !errors.Is(e, os.ErrNotExist) {
		return result, e
	}
	entries, err := os.ReadDir(filepath.Join(m.options.StateRoot, "operations"))
	if err != nil {
		return result, err
	}
	if len(entries) >= m.options.MaxOperations {
		return result, errors.New("container operation journal budget reached")
	}
	r := Result{TaskID: op.TaskID, Action: op.Action, State: "RUNNING", Phase: "accepted", StartedAt: time.Now().UTC(), Progress: []Progress{}}
	x := &execution{manager: m, actor: actor, ctx: ctx, notify: notify, entry: journal{ActorID: actor, Digest: digest, Operation: op, Result: r}}
	f, err := os.OpenFile(m.operationPath(op.TaskID), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return result, ErrConflict
	}
	b, _ := json.Marshal(x.entry)
	_, err = f.Write(b)
	if err == nil {
		err = f.Sync()
	}
	ce := f.Close()
	if err != nil {
		return result, err
	}
	if ce != nil {
		return result, ce
	}
	if err = syncDir(filepath.Dir(m.operationPath(op.TaskID))); err != nil {
		return result, err
	}
	m.mu.Lock()
	m.active[op.TaskID] = true
	m.mu.Unlock()
	defer func() { m.mu.Lock(); delete(m.active, op.TaskID); m.mu.Unlock() }()
	defer func() {
		r := &x.entry.Result
		r.FinishedAt = time.Now().UTC()
		if returnErr == nil {
			r.State = "SUCCEEDED"
			r.Phase = "complete"
		} else {
			r.Diagnostic = returnErr.Error()
			if len(r.Diagnostic) > 4096 {
				r.Diagnostic = r.Diagnostic[:4096]
			}
			r.State = "FAILED"
			if r.Unknown {
				r.State = "INTERRUPTED"
			} else if ctx.Err() != nil {
				r.State = "CANCELLED"
			}
		}
		if e := x.save(); e != nil {
			r.Unknown = true
			r.State = "INTERRUPTED"
			returnErr = errors.Join(returnErr, e)
		}
		result = *r
	}()
	if err = x.emit("validate", "Validate fixed target and operation"); err != nil {
		return result, err
	}
	if err = m.requireEngine(); err != nil {
		return result, err
	}
	if strings.HasPrefix(op.Action, "compose.") {
		return result, x.compose(ctx)
	}
	return result, x.engineOperation(ctx)
}

// Package runtime owns operating-system running units. A Record identifies one
// run, never just a PID. Resource authorization and serialization belong to the
// daemon; this package persists launch intent before creating any processes.
package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"blora.dev/panel/internal/model"
)

var (
	ErrUnknown     = errors.New("running unit identity cannot be confirmed")
	ErrStopTimeout = errors.New("running unit did not exit before deadline")
	ErrCapability  = errors.New("required runtime capability unavailable")
	idPattern      = regexp.MustCompile(`^[A-Za-z0-9_-]{1,96}$`)
)

// windowsCPUQuotaRate converts the common cgroup/Docker quota unit
// (microseconds of CPU time per 100ms period) to the Windows Job Object
// CpuRate value (1/10000 of the whole machine). A quota of 100000 per logical
// CPU therefore maps to 10000, regardless of the host's CPU count.
func windowsCPUQuotaRate(quota, cores int64) (uint32, error) {
	if quota <= 0 {
		return 0, nil
	}
	if cores < 1 {
		cores = 1
	}
	if quota > 100000*cores {
		return 0, fmt.Errorf("CPU quota exceeds available machine processors: %w", ErrCapability)
	}
	rate := quota / (10 * cores)
	if rate < 1 {
		rate = 1
	}
	if rate > 10000 {
		rate = 10000
	}
	return uint32(rate), nil
}

type Options struct {
	StateRoot         string
	CgroupRoot        string                                             // Existing delegated cgroup v2 subtree, configured by administrator.
	AllowPGIDFallback bool                                               // Trusted native processes only; never a sandbox.
	DockerEndpoint    string                                             // unix:///var/run/docker.sock or verified HTTPS.
	InstanceRoot      string                                             // Container bind sources are exactly InstanceRoot/instanceID.
	JobKeeperCommand  []string                                           // Windows only; administrator executable prefix, default current executable + --job-keeper.
	NativeInput       func(context.Context, Record, []byte) (int, error) // Optional independent stdin holder; called only after this run is observed.
}

type IO struct {
	Stdin, Stdout, Stderr *os.File
	Input                 *os.File // Parent write end, distinct from the child's Stdin read end.
}

type Record struct {
	RunID           string    `json:"runId"`
	InstanceID      string    `json:"instanceId"`
	Backend         string    `json:"backend"`
	Token           string    `json:"token"` // Internal ownership marker; not a user credential.
	PID             int       `json:"pid,omitempty"`
	PGID            int       `json:"pgid,omitempty"`
	StartTicks      uint64    `json:"startTicks,omitempty"`
	BootID          string    `json:"bootId,omitempty"`
	Unit            string    `json:"unit,omitempty"`
	UnitIdentity    uint64    `json:"unitIdentity,omitempty"`
	ContainerID     string    `json:"containerId,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	Phase           string    `json:"phase"`
	WeakContainment bool      `json:"weakContainment"`
}

type Process struct {
	PID        int    `json:"pid"`
	StartTicks uint64 `json:"startTicks"`
	State      string `json:"state"`
}
type Observation struct {
	State      string    `json:"state"`
	Exited     bool      `json:"exited"`
	Processes  []Process `json:"processes,omitempty"`
	Diagnostic string    `json:"diagnostic,omitempty"`
	ObservedAt time.Time `json:"observedAt"`
}
type StopPolicy struct {
	Grace, KillWait time.Duration
	Escalate, Force bool
	Input           string
}

func Policy(c model.InstanceConfig) StopPolicy {
	g, k := c.StopSeconds, c.KillSeconds
	if g <= 0 {
		g = 30
	}
	if k <= 0 {
		k = 10
	}
	return StopPolicy{Grace: time.Duration(g) * time.Second, KillWait: time.Duration(k) * time.Second, Escalate: c.Escalate, Input: c.StopInput}
}

type Manager struct {
	options Options
	mu      sync.Mutex
	live    map[string]*liveRun
	docker  *dockerClient
}
type liveRun struct {
	stdin    *os.File
	platform any
	inputMu  sync.Mutex
}

func New(options Options) (*Manager, error) {
	if options.StateRoot == "" {
		return nil, errors.New("runtime StateRoot is required")
	}
	root, err := filepath.Abs(options.StateRoot)
	if err != nil {
		return nil, err
	}
	options.StateRoot = root
	if err = os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	if options.InstanceRoot != "" {
		options.InstanceRoot, err = filepath.Abs(options.InstanceRoot)
		if err != nil {
			return nil, err
		}
		if err = os.MkdirAll(options.InstanceRoot, 0700); err != nil {
			return nil, err
		}
	}
	m := &Manager{options: options, live: make(map[string]*liveRun)}
	if options.DockerEndpoint != "" {
		m.docker, err = newDockerClient(options.DockerEndpoint)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (m *Manager) save(r Record) error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(m.options.StateRoot, ".run-*")
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
	if err = os.Rename(name, filepath.Join(m.options.StateRoot, r.RunID+".json")); err != nil {
		return err
	}
	return syncDirectory(m.options.StateRoot)
}

func (m *Manager) Record(runID string) (Record, error) {
	var r Record
	if !idPattern.MatchString(runID) {
		return r, errors.New("invalid run ID")
	}
	b, err := os.ReadFile(filepath.Join(m.options.StateRoot, runID+".json"))
	if err != nil {
		return r, err
	}
	err = json.Unmarshal(b, &r)
	if err == nil && (r.RunID != runID || !idPattern.MatchString(r.InstanceID)) {
		err = ErrUnknown
	}
	return r, err
}

// Recover reconciles launch intent as well as completed launches. It never starts
// a replacement run or replays application input.
func (m *Manager) Recover(ctx context.Context, runID string) (Record, Observation, error) {
	r, err := m.Record(runID)
	if err != nil {
		return r, Observation{}, err
	}
	o, err := m.Observe(ctx, r)
	return r, o, err
}

func (m *Manager) Start(ctx context.Context, runID, instanceID string, c model.InstanceConfig, streams IO) (Record, error) {
	r := Record{RunID: runID, InstanceID: instanceID, Token: model.ID(), CreatedAt: time.Now().UTC(), Phase: "intent"}
	if !idPattern.MatchString(runID) || !idPattern.MatchString(instanceID) {
		return r, errors.New("invalid runtime identity")
	}
	if len(c.Command) == 0 || c.Command[0] == "" {
		return r, errors.New("command is required")
	}
	for k, v := range c.Environment {
		if k == "" || strings.ContainsAny(k, "=\x00") || strings.ContainsRune(v, 0) {
			return r, errors.New("invalid environment variable")
		}
	}
	if c.MemoryBytes < 0 || c.CPUQuota < 0 || c.PidsLimit < 0 {
		return r, errors.New("negative resource limit")
	}
	if err := ctx.Err(); err != nil {
		return r, err
	}
	if c.Mode == "container" || c.Mode == "isolated" {
		r.Backend = "docker"
	} else if c.Mode == "native" || c.Mode == "" {
		r.Backend = nativeBackend()
	} else {
		return r, errors.New("unsupported runtime mode")
	}
	// O_EXCL reserves the ID across both goroutines and daemon processes. The
	// caller must still serialize different run IDs for the same resource.
	claim, err := os.OpenFile(filepath.Join(m.options.StateRoot, runID+".json"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return r, fmt.Errorf("run ID exists or cannot be reserved; reconcile original run: %w", err)
	}
	initial, _ := json.Marshal(r)
	_, writeErr := claim.Write(initial)
	if writeErr == nil {
		writeErr = claim.Sync()
	}
	closeErr := claim.Close()
	if writeErr != nil {
		return r, writeErr
	}
	if closeErr != nil {
		return r, closeErr
	}
	if err := m.save(r); err != nil {
		return r, err
	}
	if r.Backend == "docker" {
		r, err = m.startDocker(ctx, r, c)
	} else {
		r, err = m.startNative(ctx, r, c, streams)
	}
	if err != nil {
		if r.Unit == "" && r.PID == 0 && r.ContainerID == "" {
			r.Phase = "not_started"
			if saveErr := m.save(r); saveErr != nil {
				return r, errors.Join(err, saveErr)
			}
		}
		return r, err
	}
	r.Phase = "running"
	if err = m.save(r); err != nil {
		return r, fmt.Errorf("launch succeeded but recording failed; reconcile run %s: %w", runID, err)
	}
	return r, nil
}

func (m *Manager) Observe(ctx context.Context, r Record) (Observation, error) {
	if err := ctx.Err(); err != nil {
		return Observation{State: "UNKNOWN"}, err
	}
	if !idPattern.MatchString(r.RunID) || !idPattern.MatchString(r.InstanceID) || !idPattern.MatchString(r.Token) {
		return Observation{State: "UNKNOWN"}, ErrUnknown
	}
	if r.Phase == "not_started" || r.Phase == "exited" || r.Phase == "cleaned" {
		return Observation{State: "STOPPED", Exited: true, ObservedAt: time.Now().UTC(), Diagnostic: "durable runtime phase: " + r.Phase}, nil
	}
	if r.Phase == "intent" && r.Unit == "" && r.BootID == "" && r.PID == 0 && r.ContainerID == "" {
		return Observation{State: "STOPPED", Exited: true, ObservedAt: time.Now().UTC(), Diagnostic: "launch did not pass the durable pre-execution barrier"}, nil
	}
	if r.Backend == "docker" {
		return m.observeDocker(ctx, r)
	}
	return m.observeNative(ctx, r)
}

// Stop has bounded phases. Callback persistence failure aborts before the next
// side effect. Cancellation never means the requested OS signal was rolled back.
func (m *Manager) Stop(ctx context.Context, r Record, p StopPolicy, phase func(string) error) (Observation, error) {
	if p.Grace <= 0 {
		p.Grace = 30 * time.Second
	}
	if p.KillWait <= 0 {
		p.KillWait = 10 * time.Second
	}
	notify := func(s string) error {
		if phase != nil {
			return phase(s)
		}
		return nil
	}
	o, err := m.Observe(ctx, r)
	if err != nil || o.Exited {
		return o, err
	}
	if !p.Force {
		if err = notify("STOPPING"); err != nil {
			return o, err
		}
		if err = m.signal(ctx, r, false, p.Input); err != nil {
			return o, err
		}
		o, err = m.waitExit(ctx, r, p.Grace)
		if err == nil {
			return o, nil
		}
		if !errors.Is(err, ErrStopTimeout) || !p.Escalate {
			return o, err
		}
	}
	if err = notify("KILLING"); err != nil {
		return o, err
	}
	if err = m.signal(ctx, r, true, ""); err != nil {
		return o, err
	}
	if err = notify("VERIFYING_EXIT"); err != nil {
		return o, err
	}
	return m.waitExit(ctx, r, p.KillWait)
}
func (m *Manager) waitExit(ctx context.Context, r Record, limit time.Duration) (Observation, error) {
	ctx, cancel := context.WithTimeout(ctx, limit)
	defer cancel()
	ticker := time.NewTicker(40 * time.Millisecond)
	defer ticker.Stop()
	var o Observation
	for {
		var err error
		o, err = m.Observe(ctx, r)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			return o, err
		}
		if o.Exited {
			return o, nil
		}
		select {
		case <-ctx.Done():
			goto done
		case <-ticker.C:
		}
	}
done:
	o.State = "STOP_FAILED"
	o.Diagnostic = "exit confirmation deadline reached; replacement launch must remain blocked"
	if errors.Is(ctx.Err(), context.Canceled) {
		return o, ctx.Err()
	}
	return o, ErrStopTimeout
}
func (m *Manager) signal(ctx context.Context, r Record, force bool, input string) error {
	if r.Backend == "docker" {
		return m.signalDocker(ctx, r, force, input)
	}
	if input != "" && !force {
		_, err := m.WriteInput(ctx, r, []byte(input))
		return err
	}
	return m.signalNative(ctx, r, force)
}

// WriteInput is an at-most-once attempt. A partial/error return must never be
// retried automatically because the application may already have consumed it.
func (m *Manager) WriteInput(ctx context.Context, r Record, data []byte) (int, error) {
	if len(data) > 64<<10 {
		return 0, errors.New("input exceeds 64 KiB budget")
	}
	o, err := m.Observe(ctx, r)
	if err != nil {
		return 0, err
	}
	if o.Exited {
		return 0, errors.New("running unit has exited")
	}
	if o.State != "RUNNING" {
		return 0, fmt.Errorf("input requires a confirmed running unit: %w", ErrUnknown)
	}
	if r.Backend == "docker" {
		return m.writeDockerInput(ctx, r, data)
	}
	if m.options.NativeInput != nil {
		return m.options.NativeInput(ctx, r, data)
	}
	m.mu.Lock()
	live := m.live[r.RunID]
	m.mu.Unlock()
	if live == nil || live.stdin == nil {
		return 0, fmt.Errorf("stdin is no longer attached: %w", ErrCapability)
	}
	if !live.inputMu.TryLock() {
		return 0, errors.New("another input write is in progress")
	}
	defer live.inputMu.Unlock()
	deadline := time.Now().Add(time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if err = ctx.Err(); err != nil {
		return 0, err
	}
	if err = live.stdin.SetWriteDeadline(deadline); err != nil {
		return 0, fmt.Errorf("stdin does not support bounded writes: %w", err)
	}
	return live.stdin.Write(data)
}

// Cleanup only releases this run's metadata and empty OS unit. Callers retain the
// persisted Record for audit. It never deletes instance files or persistent volumes.
func (m *Manager) Cleanup(ctx context.Context, r Record) error {
	if saved, err := m.Record(r.RunID); err == nil && saved.Token == r.Token && (saved.Phase == "cleaned" || saved.Phase == "exited" || saved.Phase == "not_started") {
		r = saved
	}
	if r.Phase == "cleaned" || r.Phase == "not_started" {
		return nil
	}
	o, err := m.Observe(ctx, r)
	if err != nil {
		return err
	}
	if !o.Exited {
		return ErrUnknown
	}
	// Preserve authoritative exit evidence before deleting the OS ownership
	// object; an absent cgroup alone cannot distinguish cleanup from tampering.
	r.Phase = "exited"
	if err = m.save(r); err != nil {
		return err
	}
	if r.Backend == "docker" {
		err = m.cleanupDocker(ctx, r)
	} else {
		err = m.cleanupNative(r)
	}
	if err != nil {
		return err
	}
	r.Phase = "cleaned"
	return m.save(r)
}

func environment(c model.InstanceConfig, token string) []string {
	// Do not inherit daemon credentials or host library injection settings.
	env := map[string]string{"PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"}
	for k, v := range c.Environment {
		if !strings.ContainsAny(k, "=\x00") && !strings.ContainsRune(v, 0) {
			env[k] = v
		}
	}
	env["BLORA_RUN_TOKEN"] = token
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := make([]string, 0, len(keys))
	for _, k := range keys {
		result = append(result, k+"="+env[k])
	}
	return result
}

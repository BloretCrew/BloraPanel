package daemon

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"blora.dev/panel/internal/backup"
	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/containers"
	"blora.dev/panel/internal/containerterm"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/monitor"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/runlog"
	run "blora.dev/panel/internal/runtime"
	"blora.dev/panel/internal/storage"
	"blora.dev/panel/internal/terminal"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Config struct {
	StateDir           string   `json:"stateDir"`
	MasterURL          string   `json:"masterUrl"`
	CAFile             string   `json:"caFile"`
	EnrollmentFile     string   `json:"enrollmentFile,omitempty"`
	RotationFile       string   `json:"rotationFile,omitempty"`
	CgroupRoot         string   `json:"cgroupRoot,omitempty"`
	AllowPGIDFallback  bool     `json:"allowPGIDFallback"`
	DockerEndpoint     string   `json:"dockerEndpoint,omitempty"`
	LogHelperCommand   []string `json:"logHelperCommand,omitempty"`
	ComposeCommand     []string `json:"composeCommand,omitempty"`
	ComposeTLSCertPath string   `json:"composeTlsCertPath,omitempty"`
	BackupRoot         string   `json:"backupRoot,omitempty"`
	// BackupHookCommands is an administrator-owned map from a consistency hook
	// reference to a fixed argv. Commands are executed without a shell and only
	// when a backup explicitly selects the matching reference.
	BackupHookCommands map[string][]string `json:"backupHookCommands,omitempty"`
}
type identity struct {
	NodeID     string `json:"nodeId"`
	PrivateKey []byte `json:"privateKey"`
}
type Daemon struct {
	config            Config
	store             *storage.Store
	runtime           *run.Manager
	identity          identity
	client            *http.Client
	mu                sync.Mutex
	conn              *protocol.Conn
	locks             map[string]*sync.Mutex
	active            map[string]context.CancelFunc
	slots             chan struct{}
	fileSlots         chan struct{}
	terminalSlots     chan struct{}
	hostSlots         chan struct{}
	links             map[protocol.Channel]*bridge.Link
	files             map[string]*fileHandle
	startupID         string
	terminals         *terminal.Manager
	containers        *containers.Manager
	hostExecutor      *run.HostDockerExecutor
	backups           *backup.Manager
	metrics           *monitor.Collector
	wg                sync.WaitGroup
	closing, running  bool
	runCancel         context.CancelFunc
	closedDone        chan struct{}
	closeErr          error
	capabilities      map[string]string
	firewallMu        sync.Mutex
	firewallLeases    map[string]*firewallLease
	firewallCurrent   func(context.Context) ([]string, error)
	firewallApply     func(context.Context, []string) error
	firewallRollback  func(context.Context, []string, []string) error
	firewallAuthorize func(context.Context, string) error
}

func New(c Config) (*Daemon, error) {
	if c.StateDir == "" {
		return nil, errors.New("stateDir is required")
	}
	u, err := url.Parse(c.MasterURL)
	if err != nil || u.Host == "" || u.Scheme != "https" && u.Scheme != "http" {
		return nil, errors.New("masterUrl must use HTTPS, or HTTP for loopback only")
	}
	if u.Scheme == "http" {
		ip := net.ParseIP(u.Hostname())
		if u.Hostname() != "localhost" && (ip == nil || !ip.IsLoopback()) {
			return nil, errors.New("HTTP masterUrl is allowed only on loopback")
		}
	}
	if err := os.MkdirAll(c.StateDir, 0700); err != nil {
		return nil, err
	}
	pool, err := x509.SystemCertPool()
	if err != nil {
		pool = x509.NewCertPool()
	}
	if c.CAFile != "" {
		pem, err := os.ReadFile(c.CAFile)
		if err != nil {
			return nil, err
		}
		if !pool.AppendCertsFromPEM(pem) {
			return nil, errors.New("invalid CA certificate")
		}
	}
	dialer := &net.Dialer{Timeout: 20 * time.Second, KeepAlive: 30 * time.Second}
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12},
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			conn, err := dialer.DialContext(ctx, network, address)
			if err == nil && u.Scheme == "http" {
				host, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
				if ip := net.ParseIP(host); ip == nil || !ip.IsLoopback() {
					_ = conn.Close()
					return nil, errors.New("HTTP management connection resolved outside loopback")
				}
			}
			return conn, err
		}}, Timeout: 20 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return errors.New("management redirects are not allowed")
	}}
	s, err := storage.Open(filepath.Join(c.StateDir, "daemon.db"))
	if err != nil {
		return nil, err
	}
	rt, err := run.New(run.Options{StateRoot: filepath.Join(c.StateDir, "runs"), InstanceRoot: filepath.Join(c.StateDir, "instances"), CgroupRoot: c.CgroupRoot, AllowPGIDFallback: c.AllowPGIDFallback, DockerEndpoint: c.DockerEndpoint, NativeInput: func(ctx context.Context, r run.Record, data []byte) (int, error) {
		capture, err := runlog.Open(filepath.Join(c.StateDir, "logs"), r.RunID)
		if err != nil {
			return 0, err
		}
		return capture.WriteInput(ctx, data)
	}})
	if err != nil {
		s.Close()
		return nil, err
	}
	d := &Daemon{config: c, store: s, runtime: rt, client: client, startupID: model.ID(), locks: map[string]*sync.Mutex{}, active: map[string]context.CancelFunc{}, slots: make(chan struct{}, 4), fileSlots: make(chan struct{}, 8), terminalSlots: make(chan struct{}, 4), links: map[protocol.Channel]*bridge.Link{}, files: map[string]*fileHandle{}, metrics: monitor.New(), firewallLeases: map[string]*firewallLease{}}
	if err := d.loadIdentity(); err != nil {
		s.Close()
		return nil, err
	}
	if c.RotationFile != "" {
		if err := d.rotateIdentity(context.Background(), c.RotationFile); err != nil {
			s.Close()
			return nil, err
		}
	}
	backends := map[string]terminal.Backend{"docker": containerterm.Adapter{Runtime: rt, Resolve: d.resolveTerminalRun}}
	if c.DockerEndpoint != "" {
		d.hostExecutor, err = run.NewHostDockerExecutor(run.HostDockerOptions{StateRoot: filepath.Join(c.StateDir, "host-exec"), Endpoint: c.DockerEndpoint, TLSCertPath: c.ComposeTLSCertPath})
		if err != nil {
			s.Close()
			return nil, err
		}
		backends["docker-host"] = containerterm.HostAdapter{Executor: d.hostExecutor, Resolve: d.resolveHostTerminal}
	}
	d.terminals, err = terminal.New(terminal.Options{Root: filepath.Join(c.StateDir, "terminals"), AllowNative: true, Authorize: d.authorizeTerminal, Backends: backends})
	if err != nil {
		if d.hostExecutor != nil {
			_ = d.hostExecutor.Close()
		}
		s.Close()
		return nil, err
	}
	d.hostSlots = make(chan struct{}, 2)
	d.containers, err = containers.New(containers.Options{Endpoint: c.DockerEndpoint, StateRoot: filepath.Join(c.StateDir, "containers"), ComposeCommand: c.ComposeCommand, ComposeTLSCertPath: c.ComposeTLSCertPath, Authorize: func(ctx context.Context, actor string) error {
		return d.authorizeTerminal(ctx, actor, model.ResourceRef{Kind: "node", ID: d.identity.NodeID, NodeID: d.identity.NodeID}, "host.manage")
	}})
	if err != nil {
		d.terminals.Close()
		if d.hostExecutor != nil {
			_ = d.hostExecutor.Close()
		}
		s.Close()
		return nil, err
	}
	if err := d.initBackups(); err != nil {
		d.containers.Close()
		d.terminals.Close()
		if d.hostExecutor != nil {
			_ = d.hostExecutor.Close()
		}
		s.Close()
		return nil, err
	}
	return d, nil
}
func (d *Daemon) loadIdentity() error {
	path := filepath.Join(d.config.StateDir, "identity.json")
	b, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(b, &d.identity); err != nil {
			return err
		}
		if len(d.identity.PrivateKey) != ed25519.PrivateKeySize {
			return errors.New("invalid node identity key")
		}
		if d.identity.NodeID != "" {
			return nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	} else {
		_, key, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return err
		}
		d.identity.PrivateKey = key
		b, _ = json.Marshal(d.identity)
		if err := privateWrite(path, b); err != nil {
			return err
		}
	}
	if d.config.EnrollmentFile == "" {
		return errors.New("new node requires an enrollmentFile")
	}
	token, err := os.ReadFile(d.config.EnrollmentFile)
	if err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]any{"token": strings.TrimSpace(string(token)), "publicKey": ed25519.PrivateKey(d.identity.PrivateKey).Public().(ed25519.PublicKey)})
	req, err := http.NewRequest("POST", strings.TrimRight(d.config.MasterURL, "/")+"/api/v1/enroll", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return fmt.Errorf("node enrollment rejected: HTTP %d", resp.StatusCode)
	}
	var result struct {
		Node model.Node `json:"node"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 16384)).Decode(&result); err != nil {
		return err
	}
	if result.Node.ID == "" {
		return errors.New("missing enrolled node identity")
	}
	d.identity.NodeID = result.Node.ID
	b, _ = json.Marshal(d.identity)
	return privateWrite(path, b)
}
func privateWrite(path string, b []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".state-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err := f.Chmod(0600); err != nil {
		f.Close()
		return err
	}
	if _, err = f.Write(b); err == nil {
		err = f.Sync()
	}
	cerr := f.Close()
	if err != nil {
		return err
	}
	if cerr != nil {
		return cerr
	}
	return os.Rename(f.Name(), path)
}
func (d *Daemon) NodeID() string { return d.identity.NodeID }
func (d *Daemon) Close() error {
	d.mu.Lock()
	if d.closing {
		done := d.closedDone
		d.mu.Unlock()
		<-done
		return d.closeErr
	}
	d.closing = true
	d.closedDone = make(chan struct{})
	if d.runCancel != nil {
		d.runCancel()
	}
	c := d.conn
	links := make([]*bridge.Link, 0, len(d.links))
	for _, link := range d.links {
		links = append(links, link)
	}
	for _, cancel := range d.active {
		cancel()
	}
	d.mu.Unlock()
	if c != nil {
		_ = c.Close()
	}
	for _, link := range links {
		_ = link.Close()
	}
	d.wg.Wait()
	for _, link := range links {
		link.Wait()
	}
	if d.terminals != nil {
		d.closeErr = errors.Join(d.closeErr, d.terminals.Close())
	}
	if d.containers != nil {
		d.closeErr = errors.Join(d.closeErr, d.containers.Close())
	}
	if d.hostExecutor != nil {
		d.closeErr = errors.Join(d.closeErr, d.hostExecutor.Close())
	}
	if d.backups != nil {
		d.closeErr = errors.Join(d.closeErr, d.backups.Close())
	}
	for _, service := range d.files {
		d.closeErr = errors.Join(d.closeErr, service.service.Close())
	}
	d.closeErr = errors.Join(d.closeErr, d.store.Close())
	close(d.closedDone)
	return d.closeErr
}

func (d *Daemon) background(work func()) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closing {
		return false
	}
	d.wg.Add(1)
	go func() { defer d.wg.Done(); work() }()
	return true
}

// Run reconnects management independently of running resources and accepted jobs.
func (d *Daemon) Run(ctx context.Context) error {
	d.mu.Lock()
	if d.closing || d.running {
		d.mu.Unlock()
		return errors.New("daemon is already running or closed")
	}
	ctx, cancel := context.WithCancel(ctx)
	d.running = true
	d.runCancel = cancel
	d.wg.Add(1)
	d.mu.Unlock()
	defer d.wg.Done()
	defer cancel()
	if err := d.recover(ctx); err != nil {
		return err
	}
	d.capabilities = d.runtime.Capabilities(ctx)
	d.capabilities["file"] = "available"
	d.capabilities["terminal.native"] = "requires host.manage"
	d.background(func() { d.observe(ctx) })
	d.background(func() { d.maintainFiles(ctx) })
	backoff := time.Second
	for {
		err := d.connect(ctx)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		_ = err
		var jitter [1]byte
		_, _ = rand.Read(jitter[:])
		timer := time.NewTimer(backoff + time.Duration(jitter[0])*time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		if backoff < 15*time.Second {
			backoff *= 2
		}
	}
}
func (d *Daemon) connect(ctx context.Context) error {
	endpoint := managementWebSocketURL(d.config.MasterURL) + "/api/v1/agent/control?nodeId=" + url.QueryEscape(d.identity.NodeID)
	ws, _, err := websocket.Dial(ctx, endpoint, &websocket.DialOptions{HTTPClient: d.client})
	if err != nil {
		return err
	}
	defer ws.CloseNow()
	ws.SetReadLimit(16384)
	handshake, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var challenge protocol.Challenge
	if err := wsjson.Read(handshake, ws, &challenge); err != nil {
		return err
	}
	if challenge.NodeID != d.identity.NodeID || challenge.Channel != protocol.ChannelControl {
		return errors.New("invalid challenge scope")
	}
	signature, err := protocol.SignChallenge(ed25519.PrivateKey(d.identity.PrivateKey), challenge)
	if err != nil {
		return err
	}
	hello := model.AgentHello{Platform: runtime.GOOS + "/" + runtime.GOARCH, Capabilities: d.capabilities, StartupID: d.startupID}
	if err := wsjson.Write(handshake, ws, map[string]any{"signature": signature, "hello": hello}); err != nil {
		return err
	}
	var ack struct {
		Generation      uint64 `json:"generation"`
		ProtocolVersion uint32 `json:"protocolVersion"`
	}
	if err := wsjson.Read(handshake, ws, &ack); err != nil {
		return err
	}
	if ack.Generation != challenge.Generation || ack.ProtocolVersion != protocol.Version {
		return errors.New("handshake version mismatch")
	}
	c, err := protocol.NewConn(ctx, ws, protocol.Options{Generation: ack.Generation, Channel: protocol.ChannelControl})
	if err != nil {
		return err
	}
	defer c.Close()
	d.mu.Lock()
	d.conn = c
	d.mu.Unlock()
	defer func() {
		d.mu.Lock()
		if d.conn == c {
			d.conn = nil
		}
		d.mu.Unlock()
	}()
	connectionCtx, stop := context.WithCancel(ctx)
	defer stop()
	d.background(func() { d.dataLoop(connectionCtx, protocol.ChannelInteractive, ack.Generation) })
	d.background(func() { d.dataLoop(connectionCtx, protocol.ChannelBulk, ack.Generation) })
	if err := d.startupSnapshot(connectionCtx, c); err != nil {
		return err
	}
	d.background(func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-connectionCtx.Done():
				return
			case <-ticker.C:
				if err := c.Send(connectionCtx, protocol.Envelope{Type: protocol.TypePing}); err != nil {
					_ = c.Close()
					return
				}
			}
		}
	})
	for {
		e, err := c.Read(connectionCtx)
		if err != nil {
			return err
		}
		switch e.Type {
		case protocol.TypePong:
			continue
		case protocol.TypeOpen:
			var command model.AgentCommand
			if err := json.Unmarshal(e.Payload, &command); err != nil {
				return err
			}
			if command.Task.Resource.NodeID != d.identity.NodeID || (command.Task.Resource.Kind != "instance" && (command.Task.Resource.Kind != "node" || command.Task.Resource.ID != d.identity.NodeID || !model.IsNodeTaskAction(command.Task.Action))) {
				return errors.New("task scope mismatch")
			}
			if command.Kind == "reconcile" || command.Kind == "cancel" {
				existing, err := d.store.Task(ctx, command.Task.ID)
				if errors.Is(err, sql.ErrNoRows) {
					receipt, _ := json.Marshal(model.AgentReceipt{TaskID: command.Task.ID, RequestID: command.Task.RequestID, Missing: true})
					if err := c.Send(ctx, protocol.Envelope{Type: protocol.TypeOpenAck, RequestID: command.Task.RequestID, Payload: receipt}); err != nil {
						return err
					}
					continue
				}
				if err != nil {
					return err
				}
				if existing.RequestID != command.Task.RequestID || existing.Digest != command.Task.Digest {
					return errors.New("task reconciliation identity mismatch")
				}
				if command.Kind == "cancel" {
					d.cancelTask(ctx, existing.ID)
				} else {
					d.report(ctx, existing)
					if existing.State == model.Queued {
						d.schedule(connectionCtx, existing)
					}
				}
				continue
			}
			if command.Kind != "task" {
				return errors.New("unknown command")
			}
			accepted, fresh, err := d.store.Accept(ctx, command.Task)
			if err != nil {
				return err
			}
			d.report(ctx, accepted)
			if fresh || accepted.State == model.Queued {
				d.schedule(ctx, accepted)
			}
		default:
			return errors.New("unexpected control message")
		}
	}
}
func (d *Daemon) schedule(ctx context.Context, t model.Task) {
	d.mu.Lock()
	if d.closing {
		d.mu.Unlock()
		return
	}
	if _, ok := d.active[t.ID]; ok {
		d.mu.Unlock()
		return
	}
	jobCtx, cancel := context.WithCancel(ctx)
	d.active[t.ID] = cancel
	lock := d.locks[t.Resource.ID]
	if lock == nil {
		lock = &sync.Mutex{}
		d.locks[t.Resource.ID] = lock
	}
	d.mu.Unlock()
	if !d.background(func() {
		defer func() { cancel(); d.mu.Lock(); delete(d.active, t.ID); d.mu.Unlock() }()
		budget := d.slots
		if model.IsContainerAction(t.Action) {
			budget = d.hostSlots
		} else if strings.HasPrefix(t.Action, "file.") || strings.HasPrefix(t.Action, "backup.") {
			budget = d.fileSlots
		} else if strings.HasPrefix(t.Action, "terminal.") {
			budget = d.terminalSlots
		} else {
			lock.Lock()
			defer lock.Unlock()
		}
		select {
		case budget <- struct{}{}:
			defer func() { <-budget }()
		case <-jobCtx.Done():
			return
		}
		d.execute(jobCtx, t.ID)
	}) {
		cancel()
		d.mu.Lock()
		delete(d.active, t.ID)
		d.mu.Unlock()
	}
}
func (d *Daemon) transition(ctx context.Context, id string, state model.TaskState, phase string, result json.RawMessage, failure string) (model.Task, error) {
	for range 3 {
		t, err := d.store.Task(ctx, id)
		if err != nil {
			return t, err
		}
		if t.State.Terminal() {
			return t, storage.ErrTransition
		}
		updated, err := d.store.UpdateTask(ctx, id, t.Revision, state, phase, result, failure)
		if errors.Is(err, storage.ErrConflict) {
			continue
		}
		if err == nil {
			d.report(ctx, updated)
		}
		return updated, err
	}
	return model.Task{}, storage.ErrConflict
}
func (d *Daemon) report(ctx context.Context, t model.Task) {
	d.mu.Lock()
	c := d.conn
	d.mu.Unlock()
	if c == nil {
		return
	}
	event := model.AgentEvent{Task: t}
	var i model.Instance
	if _, err := d.store.Record(ctx, "instance", t.Resource.ID, &i); err == nil {
		event.Instance = &i
	}
	b, err := json.Marshal(event)
	if err != nil {
		return
	}
	sendCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	_ = c.Send(sendCtx, protocol.Envelope{Type: protocol.TypeTaskEvent, RequestID: t.RequestID, Payload: b})
}
func (d *Daemon) cancelTask(ctx context.Context, id string) {
	t, err := d.store.Task(ctx, id)
	if err != nil || t.State.Terminal() {
		return
	}
	state := model.CancelRequested
	if t.State == model.Queued {
		state = model.Cancelled
	}
	if _, err := d.transition(ctx, id, state, "cancel_requested", nil, ""); err != nil {
		return
	}
	d.mu.Lock()
	cancel := d.active[id]
	d.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}
func (d *Daemon) saveInstance(ctx context.Context, i *model.Instance) error {
	var old model.Instance
	rev, err := d.store.Record(ctx, "instance", i.ID, &old)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	i.Revision = old.Revision + 1
	_, err = d.store.PutRecord(ctx, "instance", i.ID, rev, i)
	return err
}

func (d *Daemon) execute(ctx context.Context, id string) {
	t, err := d.store.Task(ctx, id)
	if err != nil || t.State.Terminal() {
		return
	}
	if err := d.authorizeExecution(ctx, t); err != nil {
		if errors.Is(err, context.Canceled) {
			_, _ = d.transition(context.Background(), id, model.Cancelled, "cancelled_before_execution", nil, "")
			return
		}
		_, _ = d.transition(context.Background(), id, model.Failed, "authorization_not_confirmed", nil, "未执行副作用："+err.Error())
		return
	}
	t, err = d.transition(ctx, id, model.Running, "preparing", nil, "")
	if err != nil {
		return
	}
	if d.executeAdapter(ctx, t) {
		return
	}
	var config model.InstanceConfig
	if err = json.Unmarshal(t.Payload, &config); err != nil {
		_, _ = d.transition(context.Background(), id, model.Failed, "invalid_configuration", nil, err.Error())
		return
	}
	var i model.Instance
	_, err = d.store.Record(ctx, "instance", t.Resource.ID, &i)
	if errors.Is(err, sql.ErrNoRows) {
		i = model.Instance{ID: t.Resource.ID, NodeID: d.identity.NodeID, State: "STOPPED", Config: config}
	} else if err != nil {
		_, _ = d.transition(context.Background(), id, model.Failed, "instance_record_unreadable", nil, err.Error())
		return
	}
	i.Config = config
	finish := func(err error) {
		state := model.Succeeded
		phase := "completed"
		failure := ""
		if err != nil {
			state = model.Failed
			phase = "failed"
			failure = err.Error()
			if errors.Is(err, context.Canceled) {
				state = model.Cancelled
				phase = "cancelled_remaining_work"
			}
		}
		b, _ := json.Marshal(map[string]any{"state": i.State, "runId": i.RunID})
		_, _ = d.transition(context.Background(), id, state, phase, b, failure)
	}
	if t.Action != "instance.start" && t.Action != "instance.stop" && t.Action != "instance.restart" && t.Action != "instance.kill" {
		finish(errors.New("unsupported daemon action"))
		return
	}
	if i.RunID != "" {
		record, observation, observeErr := d.runtime.Recover(ctx, i.RunID)
		if observeErr != nil {
			i.State = "UNKNOWN"
			_ = d.saveInstance(context.Background(), &i)
			finish(observeErr)
			return
		}
		if !observation.Exited {
			if t.Action == "instance.start" {
				finish(errors.New("existing running unit has not exited"))
				return
			}
			policy := run.Policy(config)
			policy.Force = t.Action == "instance.kill"
			observation, err = d.runtime.Stop(ctx, record, policy, func(phase string) error {
				i.State = phase
				if phase == "VERIFYING_EXIT" {
					i.State = "KILLING"
				}
				if err := d.saveInstance(ctx, &i); err != nil {
					return err
				}
				_, err := d.transition(ctx, id, model.Running, phase, nil, "")
				return err
			})
			if err != nil {
				i.State = "STOP_FAILED"
				_ = d.saveInstance(context.Background(), &i)
				finish(err)
				return
			}
		}
		if !observation.Exited {
			finish(run.ErrUnknown)
			return
		}
		d.finishLog(i.RunID)
		if err := d.runtime.Cleanup(ctx, record); err != nil {
			finish(err)
			return
		}
		i.State = "STOPPED"
		if err := d.saveInstance(ctx, &i); err != nil {
			finish(err)
			return
		}
	}
	if t.Action == "instance.stop" || t.Action == "instance.kill" {
		i.State = "STOPPED"
		finish(d.saveInstance(ctx, &i))
		return
	}
	if err := ctx.Err(); err != nil {
		finish(err)
		return
	}
	i.RunID = model.ID()
	i.State = "STARTING"
	if err := d.saveInstance(ctx, &i); err != nil {
		finish(err)
		return
	}
	if _, err := d.transition(ctx, id, model.Running, "STARTING", nil, ""); err != nil {
		finish(err)
		return
	}
	directory := filepath.Join(d.config.StateDir, "instances", i.ID)
	if config.Mode != "container" {
		if err := os.MkdirAll(directory, 0700); err != nil {
			finish(err)
			return
		}
		if config.Directory == "" {
			config.Directory = directory
		}
	}
	var capture *runlog.Capture
	var streams run.IO
	if config.Mode != "container" {
		capture, err = d.startLog(ctx, i)
		if err != nil {
			finish(err)
			return
		}
		streams.Stdout, streams.Stderr = capture.Stdout(), capture.Stdout()
		streams.Stdin = capture.Stdin()
	}
	_, err = d.runtime.Start(ctx, i.RunID, i.ID, config, streams)
	if capture != nil {
		_ = capture.ReleaseWriter()
		_ = capture.ReleaseInputReader()
	}
	if err != nil {
		_, observation, recoveryErr := d.runtime.Recover(context.Background(), i.RunID)
		i.State = "UNKNOWN"
		if recoveryErr == nil {
			if observation.Exited {
				d.finishLog(i.RunID)
				i.State = "START_FAILED"
			} else {
				i.State = observation.State
			}
		}
		_ = d.saveInstance(context.Background(), &i)
		finish(err)
		return
	}
	i.State = "RUNNING"
	if err := d.saveInstance(ctx, &i); err != nil {
		finish(err)
		return
	}
	finish(nil)
}
func (d *Daemon) recover(ctx context.Context) error {
	if err := d.containers.RecoverCLI(ctx); err != nil {
		return fmt.Errorf("recover Compose CLI ownership: %w", err)
	}
	if err := d.recoverFirewallLeases(ctx); err != nil {
		return err
	}
	pending, err := d.store.Pending(ctx)
	if err != nil {
		return err
	}
	for _, t := range pending {
		if t.State == model.Queued {
			// Wait for the current Master connection and a fresh authority
			// check before resuming an accepted but unexecuted operation.
			continue
		}
		if _, err := d.transition(ctx, t.ID, model.Interrupted, "daemon_restarted_reconcile_required", nil, "Daemon 重启，副作用不自动重放"); err != nil {
			return err
		}
	}
	return nil
}
func (d *Daemon) observe(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		ids, err := d.store.RecordIDs(ctx, "instance")
		if err != nil {
			continue
		}
		items := []model.Instance{}
		for _, id := range ids {
			d.mu.Lock()
			lock := d.locks[id]
			if lock == nil {
				lock = &sync.Mutex{}
				d.locks[id] = lock
			}
			d.mu.Unlock()
			if !lock.TryLock() {
				continue
			}
			var i model.Instance
			if _, err := d.store.Record(ctx, "instance", id, &i); err != nil {
				lock.Unlock()
				continue
			}
			if i.RunID != "" {
				_, o, err := d.runtime.Recover(ctx, i.RunID)
				state := i.State
				if err != nil {
					state = "UNKNOWN"
				} else if o.Exited {
					state = "STOPPED"
					if i.State != "STOPPED" {
						d.finishLog(i.RunID)
					}
				}
				if state != i.State {
					i.State = state
					_ = d.saveInstance(ctx, &i)
				}
			}
			items = append(items, i)
			lock.Unlock()
		}
		d.mu.Lock()
		c := d.conn
		d.mu.Unlock()
		if c != nil {
			for start := 0; start < len(items); start += 32 {
				end := start + 32
				if end > len(items) {
					end = len(items)
				}
				b, _ := json.Marshal(map[string]any{"instances": items[start:end]})
				_ = c.Send(ctx, protocol.Envelope{Type: protocol.TypeSnapshot, Payload: b})
			}
		}
	}
}

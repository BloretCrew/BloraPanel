package runtime

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"blora.dev/panel/internal/containerterm/helper"
	"blora.dev/panel/internal/model"
)

type DockerExecOptions struct {
	SessionID   string
	Command     []string
	Directory   string
	Environment map[string]string
	Cols, Rows  uint16
}
type DockerExecInfo struct {
	ID          string `json:"ID"`
	ContainerID string `json:"ContainerID"`
	Running     bool   `json:"Running"`
	ExitCode    int    `json:"ExitCode"` // Supervisor cleanup status; child failures are preserved in terminal output.
	PID         int    `json:"Pid"`      // Engine host PID, never used for local signalling.
}
type dockerExecRecord struct {
	SessionID          string            `json:"sessionId"`
	ExecID             string            `json:"execId"`
	Run                Record            `json:"run"`
	ContainerStartedAt string            `json:"containerStartedAt"`
	Identity           helper.Identity   `json:"identity"`
	Phase              string            `json:"phase"`
	HostTarget         *HostDockerTarget `json:"hostTarget,omitempty"`
}
type DockerExec struct {
	manager   *Manager
	scope     *dockerExecScope
	record    dockerExecRecord
	stream    io.ReadWriteCloser
	reader    *bufio.Reader
	mu        sync.Mutex
	writeMu   sync.Mutex
	closeOnce sync.Once
	closed    chan struct{}
	closeErr  error
}

func (m *Manager) execContainer(ctx context.Context, r Record) (containerInspect, error) {
	if r.Backend != "docker" || !idPattern.MatchString(r.RunID) || !idPattern.MatchString(r.InstanceID) || !idPattern.MatchString(r.Token) || r.ContainerID == "" {
		return containerInspect{}, ErrUnknown
	}
	i, err := m.inspectDocker(ctx, r)
	if err != nil {
		return i, err
	}
	if !i.State.Running || i.State.Paused || i.State.Restarting || i.State.Dead {
		return i, fmt.Errorf("container is not ready for exec: %w", ErrCapability)
	}
	var security struct {
		ID         string `json:"Id"`
		Config     struct{ User string }
		HostConfig struct {
			Privileged  bool
			PidMode     string
			CapDrop     []string
			SecurityOpt []string
		}
	}
	if err = m.docker.api(ctx, "GET", "/containers/"+url.PathEscape(i.ID)+"/json", nil, &security); err != nil {
		return i, err
	}
	user := strings.Split(security.Config.User, ":")
	if len(user) != 2 {
		return i, fmt.Errorf("exec requires the numeric non-root instance identity: %w", ErrCapability)
	}
	uid, e1 := strconv.ParseUint(user[0], 10, 32)
	gid, e2 := strconv.ParseUint(user[1], 10, 32)
	if e1 != nil || e2 != nil || uid == 0 || gid == 0 || security.ID != i.ID || security.HostConfig.Privileged || security.HostConfig.PidMode != "" {
		return i, fmt.Errorf("container isolation contract differs: %w", ErrUnknown)
	}
	allCapsDropped, noNewPrivileges := false, false
	for _, value := range security.HostConfig.CapDrop {
		if strings.EqualFold(value, "ALL") {
			allCapsDropped = true
		}
	}
	for _, value := range security.HostConfig.SecurityOpt {
		if value == "no-new-privileges" || value == "no-new-privileges:true" {
			noNewPrivileges = true
		}
	}
	if !allCapsDropped || !noNewPrivileges {
		return i, fmt.Errorf("container execution privilege constraints missing: %w", ErrUnknown)
	}
	return i, nil
}

func (m *Manager) execRecordPath(session string) string {
	return filepath.Join(m.options.StateRoot, "exec", session+".json")
}
func (m *Manager) saveExec(r dockerExecRecord) error {
	dir := filepath.Dir(m.execRecordPath(r.SessionID))
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".exec-*")
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
	if err = os.Rename(name, m.execRecordPath(r.SessionID)); err != nil {
		return err
	}
	return syncDirectory(dir)
}

func (s *dockerExecScope) createExec(ctx context.Context, command, env []string, dir string, tty, stdin bool, cols, rows uint16) (string, error) {
	m, r := s.manager, s.run
	if err := s.allow(ctx); err != nil {
		return "", err
	}
	if _, err := s.container(ctx, true); err != nil {
		return "", err
	}
	body := map[string]any{"AttachStdin": stdin, "AttachStdout": true, "AttachStderr": true, "Tty": tty, "Privileged": false, "Cmd": command, "Env": env, "WorkingDir": dir}
	if tty {
		body["ConsoleSize"] = []uint16{rows, cols}
	}
	var result struct {
		ID string `json:"Id"`
	}
	if err := m.docker.api(ctx, "POST", "/containers/"+url.PathEscape(r.ContainerID)+"/exec", body, &result); err != nil {
		return "", err
	}
	if !idPattern.MatchString(result.ID) {
		return "", errors.New("Engine returned invalid exec identity")
	}
	return result.ID, nil
}

// upgradeExec attaches while starting an existing exec, exactly once. Docker
// does not provide a general reattach endpoint for a running exec; browser
// reattachment uses the terminal Manager's existing object and output archive.
func (m *Manager) upgradeExec(ctx context.Context, id string, tty bool, cols, rows uint16) (io.ReadWriteCloser, error) {
	body := map[string]any{"Detach": false, "Tty": tty}
	if tty {
		body["ConsoleSize"] = []uint16{rows, cols}
	}
	b, _ := json.Marshal(body)
	streamCtx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(streamCtx, "POST", m.docker.base+"/v"+dockerAPIVersion+"/exec/"+url.PathEscape(id)+"/start", bytes.NewReader(b))
	if err != nil {
		cancel()
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "tcp")
	client := *m.docker.client
	client.Timeout = 0
	headerDone := make(chan struct{})
	watchDone := make(chan struct{})
	go func() {
		defer close(watchDone)
		select {
		case <-ctx.Done():
			cancel()
		case <-headerDone:
		}
	}()
	resp, err := client.Do(req)
	close(headerDone)
	<-watchDone
	if err != nil {
		cancel()
		return nil, err
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		defer resp.Body.Close()
		defer cancel()
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, &engineError{resp.StatusCode, "exec attach upgrade rejected: " + string(raw)}
	}
	stream, ok := resp.Body.(io.ReadWriteCloser)
	if !ok {
		resp.Body.Close()
		cancel()
		return nil, fmt.Errorf("Docker transport does not provide duplex exec: %w", ErrCapability)
	}
	return &execStream{ReadWriteCloser: stream, cancel: cancel}, nil
}

type execStream struct {
	io.ReadWriteCloser
	cancel context.CancelFunc
}

func (s *execStream) Close() error { s.cancel(); return s.ReadWriteCloser.Close() }

// inspectExec validates both resource labels and Engine's exec-container link.
func (m *Manager) inspectExec(ctx context.Context, r Record, id string) (DockerExecInfo, error) {
	return (&dockerExecScope{manager: m, run: r}).inspectExec(ctx, id)
}
func (s *dockerExecScope) inspectExec(ctx context.Context, id string) (DockerExecInfo, error) {
	m, r := s.manager, s.run
	if !idPattern.MatchString(id) {
		return DockerExecInfo{}, ErrUnknown
	}
	if _, err := s.container(ctx, false); err != nil {
		return DockerExecInfo{}, err
	}
	var info DockerExecInfo
	if err := m.docker.api(ctx, "GET", "/exec/"+url.PathEscape(id)+"/json", nil, &info); err != nil {
		return info, err
	}
	if info.ID != id || info.ContainerID != r.ContainerID {
		return info, ErrUnknown
	}
	return info, nil
}
func (m *Manager) InspectDockerExec(ctx context.Context, r Record, id string) (DockerExecInfo, error) {
	return m.inspectExec(ctx, r, id)
}

// helperCommand only invokes the fixed public helper; it cannot accept a user
// command or privileged/user/mount override. Its result is checked by exec ID.
func (s *dockerExecScope) helperCommand(ctx context.Context, args ...string) ([]byte, error) {
	m := s.manager
	id, err := s.createExec(ctx, append([]string{helper.Path}, args...), nil, "/", false, false, 0, 0)
	if err != nil {
		return nil, err
	}
	stream, err := m.upgradeExec(ctx, id, false, 0, 0)
	if err != nil {
		return nil, err
	}
	defer stream.Close()
	finished := make(chan struct{})
	defer close(finished)
	go func() {
		select {
		case <-ctx.Done():
			stream.Close()
		case <-finished:
		}
	}()
	var out bytes.Buffer
	for {
		var header [8]byte
		_, err = io.ReadFull(stream, header[:])
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		n := binary.BigEndian.Uint32(header[4:])
		if (header[0] != 1 && header[0] != 2) || n > 64<<10 || out.Len()+int(n) > 64<<10 {
			return nil, errors.New("invalid or oversized helper response")
		}
		if _, err = io.CopyN(&out, stream, int64(n)); err != nil {
			return nil, err
		}
	}
	for {
		info, err := s.inspectExec(ctx, id)
		if err != nil {
			return nil, err
		}
		if !info.Running {
			if info.ExitCode != 0 {
				return nil, fmt.Errorf("container helper exited with code %d: %s: %w", info.ExitCode, strings.TrimSpace(out.String()), ErrCapability)
			}
			return out.Bytes(), nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(20 * time.Millisecond):
		}
	}
}

func (m *Manager) StartDockerExec(ctx context.Context, r Record, o DockerExecOptions) (*DockerExec, error) {
	return m.startDockerExec(ctx, &dockerExecScope{manager: m, run: r}, o)
}
func (m *Manager) startDockerExec(ctx context.Context, scope *dockerExecScope, o DockerExecOptions) (*DockerExec, error) {
	r := scope.run
	if !idPattern.MatchString(o.SessionID) || len(o.Command) == 0 || len(o.Command) > 128 || o.Command[0] == "" {
		return nil, errors.New("exec session and command are required")
	}
	if o.Cols == 0 {
		o.Cols = 100
	}
	if o.Rows == 0 {
		o.Rows = 28
	}
	if o.Cols > 1000 || o.Rows > 1000 {
		return nil, errors.New("exec terminal size exceeds limit")
	}
	if o.Directory == "" {
		o.Directory = "/workspace"
		if scope.host != nil {
			o.Directory = "/"
		}
	}
	if path.Clean(o.Directory) != o.Directory || !strings.HasPrefix(o.Directory, "/") || strings.ContainsRune(o.Directory, 0) || (scope.host == nil && o.Directory != "/workspace" && !strings.HasPrefix(o.Directory, "/workspace/")) {
		return nil, errors.New("exec directory must remain within /workspace")
	}
	budget := 0
	for _, arg := range o.Command {
		if strings.ContainsRune(arg, 0) {
			return nil, errors.New("invalid exec command")
		}
		budget += len(arg)
	}
	envValues := map[string]string{"TERM": "xterm-256color"}
	for key, value := range o.Environment {
		if key == "" || strings.ContainsAny(key, "=\x00") || strings.ContainsRune(value, 0) || key == helper.TokenEnvironment {
			return nil, errors.New("invalid or reserved exec environment")
		}
		envValues[key] = value
		budget += len(key) + len(value)
	}
	if budget > 32<<10 {
		return nil, errors.New("exec arguments/environment exceed limit")
	}
	if err := scope.allow(ctx); err != nil {
		return nil, err
	}
	i, err := scope.container(ctx, true)
	if err != nil {
		return nil, err
	}
	probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	raw, err := scope.helperCommand(probeCtx, "probe")
	cancel()
	if err != nil {
		return nil, fmt.Errorf("container exec helper unavailable: %w", err)
	}
	var probe helper.Probe
	if err = json.Unmarshal(raw, &probe); err != nil || probe.Version != helper.Version || !probe.PIDFD {
		return nil, fmt.Errorf("incompatible exec helper or missing pidfd: %w", ErrCapability)
	}
	if scope.host != nil && !probe.ProcessTree {
		return nil, fmt.Errorf("host container exec requires exec-helper v1 with processTree/subreaper capability: %w", ErrCapability)
	}
	token := model.ID()
	envValues[helper.TokenEnvironment] = token
	keys := make([]string, 0, len(envValues))
	for key := range envValues {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	env := make([]string, 0, len(keys))
	for _, key := range keys {
		env = append(env, key+"="+envValues[key])
	}
	record := dockerExecRecord{SessionID: o.SessionID, Run: r, HostTarget: scope.host, ContainerStartedAt: i.State.StartedAt, Phase: "intent"}
	if err = os.MkdirAll(filepath.Dir(m.execRecordPath(o.SessionID)), 0700); err != nil {
		return nil, err
	}
	claim, err := os.OpenFile(m.execRecordPath(o.SessionID), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, fmt.Errorf("exec session already claimed; reconcile original exec: %w", err)
	}
	claim.Close()
	if err = m.saveExec(record); err != nil {
		return nil, err
	}
	id, err := scope.createExec(ctx, append([]string{helper.Path, "serve", "--"}, o.Command...), env, o.Directory, true, true, o.Cols, o.Rows)
	if err != nil {
		return nil, err
	}
	record.ExecID = id
	record.Phase = "created"
	if err = m.saveExec(record); err != nil {
		return nil, err
	}
	startCtx, startCancel := context.WithTimeout(ctx, 8*time.Second)
	defer startCancel()
	stream, err := m.upgradeExec(startCtx, id, true, o.Cols, o.Rows)
	if err != nil {
		return nil, fmt.Errorf("exec start result unknown for %s: %w", id, err)
	}
	p := &DockerExec{manager: m, scope: scope, record: record, stream: stream, reader: bufio.NewReaderSize(stream, 4096), closed: make(chan struct{})}
	ready := make(chan struct{})
	readyWatcherDone := make(chan struct{})
	go func() {
		defer close(readyWatcherDone)
		select {
		case <-startCtx.Done():
			stream.Close()
		case <-ready:
		}
	}()
	lineBytes, readErr := p.reader.ReadSlice('\n')
	close(ready)
	<-readyWatcherDone
	line := string(lineBytes)
	if readErr != nil || len(line) > 4096 || !strings.HasPrefix(line, helper.Prefix) {
		stream.Close()
		return nil, fmt.Errorf("exec helper handshake failed; original exec %s must be reconciled: %w", id, ErrUnknown)
	}
	if err = json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(line, helper.Prefix))), &p.record.Identity); err != nil || p.record.Identity.Version != helper.Version || p.record.Identity.Token != token || p.record.Identity.PID <= 1 || p.record.Identity.SessionID != p.record.Identity.PID || p.record.Identity.StartTicks == 0 {
		stream.Close()
		return nil, fmt.Errorf("exec helper identity mismatch: %w", ErrUnknown)
	}
	p.record.Phase = "ready"
	if err = m.saveExec(p.record); err != nil {
		stream.Close()
		return nil, err
	}
	info, err := p.Inspect(startCtx)
	if err != nil || !info.Running {
		stream.Close()
		return nil, fmt.Errorf("exec vanished before launch permission: %w", ErrUnknown)
	}
	if _, err = p.Write([]byte("BLORA-EXEC-GO " + token + "\n")); err != nil {
		cleanupErr := p.Close()
		return nil, errors.Join(fmt.Errorf("exec launch acknowledgement uncertain; do not replay: %w", err), cleanupErr)
	}
	p.record.Phase = "running"
	if err = m.saveExec(p.record); err != nil {
		_ = p.Close()
		return nil, err
	}
	return p, nil
}

func (p *DockerExec) Read(b []byte) (int, error) { return p.reader.Read(b) }
func (p *DockerExec) Write(b []byte) (int, error) {
	if len(b) > 32<<10 {
		return 0, errors.New("exec input frame exceeds 32 KiB")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := p.scope.allow(ctx); err != nil {
		return 0, err
	}
	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	select {
	case <-p.closed:
		return 0, os.ErrClosed
	default:
	}
	finished := make(chan struct{})
	timer := time.AfterFunc(2*time.Second, func() { _ = p.stream.Close(); close(finished) })
	n, err := p.stream.Write(b)
	if !timer.Stop() {
		<-finished
	}
	return n, err
}
func (p *DockerExec) Inspect(ctx context.Context) (DockerExecInfo, error) {
	return p.inspect(ctx, p.scope)
}
func (p *DockerExec) inspect(ctx context.Context, scope *dockerExecScope) (DockerExecInfo, error) {
	i, err := scope.container(ctx, false)
	if err != nil {
		return DockerExecInfo{}, err
	}
	if i.State.StartedAt != p.record.ContainerStartedAt {
		return DockerExecInfo{}, fmt.Errorf("container run birth changed: %w", ErrUnknown)
	}
	return scope.inspectExec(ctx, p.record.ExecID)
}
func (p *DockerExec) Resize(cols, rows uint16) error {
	if cols == 0 || rows == 0 || cols > 1000 || rows > 1000 {
		return errors.New("invalid exec dimensions")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := p.scope.allow(ctx); err != nil {
		return err
	}
	i, err := p.Inspect(ctx)
	if err != nil {
		return err
	}
	if !i.Running {
		return errors.New("exec already exited")
	}
	return p.manager.docker.api(ctx, "POST", fmt.Sprintf("/exec/%s/resize?h=%d&w=%d", url.PathEscape(p.record.ExecID), rows, cols), nil, nil)
}
func (p *DockerExec) Wait() error {
	for {
		select {
		case <-p.closed:
			p.mu.Lock()
			defer p.mu.Unlock()
			return p.closeErr
		default:
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		info, err := p.Inspect(ctx)
		cancel()
		if err != nil {
			return err
		}
		if !info.Running {
			if info.ExitCode != 0 {
				return fmt.Errorf("container exec exited with code %d", info.ExitCode)
			}
			return nil
		}
		select {
		case <-p.closed:
			p.mu.Lock()
			defer p.mu.Unlock()
			return p.closeErr
		case <-time.After(150 * time.Millisecond):
		}
	}
}
func (p *DockerExec) Close() error {
	p.closeOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()
		defer p.stream.Close()
		cleanup := *p.scope
		cleanup.cleanup = true // Already-authorized session cleanup cannot be blocked by later user revocation.
		info, err := p.inspect(ctx, &cleanup)
		if err == nil && info.Running {
			id := p.record.Identity
			_, err = cleanup.helperCommand(ctx, "close", strconv.Itoa(id.PID), strconv.FormatUint(id.StartTicks, 10), id.Token)
			if err == nil {
				info, err = p.inspect(ctx, &cleanup)
				if err == nil && (info.Running || info.ExitCode != 0) {
					err = fmt.Errorf("exec or owned session cleanup was not confirmed: %w", ErrUnknown)
				}
			}
		} else if err == nil && info.ExitCode != 0 {
			err = fmt.Errorf("exec helper ended with code %d; owned-session cleanup cannot be confirmed: %w", info.ExitCode, ErrUnknown)
		}
		p.mu.Lock()
		p.closeErr = err
		if err == nil {
			p.record.Phase = "exited"
		} else {
			p.record.Phase = "close_unknown"
		}
		if saveErr := p.manager.saveExec(p.record); saveErr != nil {
			p.closeErr = errors.Join(p.closeErr, saveErr)
		}
		p.mu.Unlock()
		close(p.closed)
	})
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.closeErr
}

// Capabilities probes the actual configured backends. Native probing uses one
// labelled, short-lived run of this manager, then verifies/cleans that run only.
func (m *Manager) Capabilities(ctx context.Context) map[string]string {
	result := map[string]string{"native": "unavailable", "container": "unavailable"}
	if m.docker == nil {
		result["container.reason"] = "Docker endpoint is not configured"
	} else {
		probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		var version struct {
			APIVersion string `json:"ApiVersion"`
			Minimum    string `json:"MinAPIVersion"`
			OSType     string `json:"Os"`
		}
		err := m.docker.request(probeCtx, "GET", "/version", nil, &version)
		if err == nil && (version.OSType != "linux" || versionNumber(version.APIVersion) < versionNumber(dockerAPIVersion) || versionNumber(version.Minimum) > versionNumber(dockerAPIVersion)) {
			err = ErrCapability
		}
		if err == nil {
			var info struct {
				OSType                                         string
				MemoryLimit, SwapLimit, CPUCfsQuota, PidsLimit bool
			}
			err = m.docker.request(probeCtx, "GET", "/v"+dockerAPIVersion+"/info", nil, &info)
			if err == nil && (!info.MemoryLimit || !info.SwapLimit || !info.CPUCfsQuota || !info.PidsLimit || info.OSType != "linux") {
				err = errors.New("Engine cannot enforce required isolated instance resource limits")
			}
		}
		cancel()
		if err == nil {
			result["container"] = "available"
			result["container.api"] = dockerAPIVersion
			result["container.terminal"] = "requires exec-helper v1 in instance image"
		} else {
			result["container.reason"] = err.Error()
		}
	}
	if nativeBackend() == "unsupported" {
		result["native.reason"] = "native platform backend unavailable"
		return result
	}
	if nativeBackend() == "linux" && m.options.CgroupRoot == "" && !m.options.AllowPGIDFallback {
		result["native.reason"] = "delegated cgroup is not configured and trusted PGID fallback is disabled"
		return result
	}
	command := []string{"/bin/sh", "-c", "exit 0"}
	if nativeBackend() == "windows" {
		command = []string{"cmd.exe", "/D", "/C", "exit 0"}
	}
	probeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	r, err := m.Start(probeCtx, "cap-"+model.ID(), "cap-"+model.ID(), model.InstanceConfig{Mode: "native", Command: command}, IO{})
	cancel()
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cleanupCancel()
	if err == nil {
		_, err = m.Stop(cleanupCtx, r, StopPolicy{Force: true, KillWait: time.Second}, nil)
	}
	cleanupErr := m.Cleanup(cleanupCtx, r)
	if cleanupErr == nil {
		_ = os.Remove(filepath.Join(m.options.StateRoot, r.RunID+".json"))
	} else {
		err = errors.Join(err, fmt.Errorf("capability probe %s cleanup: %w", r.RunID, cleanupErr))
	}
	if err != nil {
		result["native.reason"] = err.Error()
		return result
	}
	result["native"] = "available"
	result["native.mode"] = nativeBackend()
	if r.WeakContainment {
		result["native.mode"] = "pgid_trusted"
		result["native.limits"] = "unavailable"
	} else if nativeBackend() == "linux" {
		result["native.mode"] = "cgroup_v2"
	}
	return result
}

package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"blora.dev/panel/internal/model"
)

const dockerAPIVersion = "1.45"

type dockerClient struct {
	client  *http.Client
	base    string
	mu      sync.Mutex
	checked bool
}
type engineError struct {
	status  int
	message string
}

func (e *engineError) Error() string {
	return fmt.Sprintf("Docker Engine status %d: %s", e.status, e.message)
}
func newDockerClient(endpoint string) (*dockerClient, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	base := "http://docker"
	switch u.Scheme {
	case "unix":
		if u.Path == "" {
			return nil, errors.New("Docker unix socket path missing")
		}
		transport.DialContext = func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", u.Path)
		}
	case "https":
		if u.Host == "" || u.User != nil || u.RawQuery != "" {
			return nil, errors.New("invalid Docker HTTPS endpoint")
		}
		base = strings.TrimSuffix(endpoint, "/")
	default:
		return nil, errors.New("Docker endpoint must be a local unix socket or verified HTTPS")
	}
	return &dockerClient{client: &http.Client{Transport: transport, Timeout: 15 * time.Second}, base: base}, nil
}
func (d *dockerClient) request(ctx context.Context, method, path string, body, out any) error {
	var data io.Reader
	if body != nil {
		b, e := json.Marshal(body)
		if e != nil {
			return e
		}
		data = bytes.NewReader(b)
	}
	r, e := http.NewRequestWithContext(ctx, method, d.base+path, data)
	if e != nil {
		return e
	}
	if body != nil {
		r.Header.Set("Content-Type", "application/json")
	}
	resp, e := d.client.Do(r)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return &engineError{resp.StatusCode, string(b)}
	}
	if out != nil {
		return json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(out)
	}
	_, e = io.Copy(io.Discard, io.LimitReader(resp.Body, 8<<20))
	return e
}
func versionNumber(v string) int {
	parts := strings.Split(v, ".")
	if len(parts) != 2 {
		return 0
	}
	a, _ := strconv.Atoi(parts[0])
	b, _ := strconv.Atoi(parts[1])
	return a*1000 + b
}
func (d *dockerClient) check(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.checked {
		return nil
	}
	var v struct {
		APIVersion string `json:"ApiVersion"`
		Minimum    string `json:"MinAPIVersion"`
		OSType     string `json:"Os"`
	}
	if err := d.request(ctx, "GET", "/version", nil, &v); err != nil {
		return err
	}
	if versionNumber(v.APIVersion) < versionNumber(dockerAPIVersion) || versionNumber(v.Minimum) > versionNumber(dockerAPIVersion) {
		return fmt.Errorf("Docker API %s outside Engine range %s..%s: %w", dockerAPIVersion, v.Minimum, v.APIVersion, ErrCapability)
	}
	if v.OSType != "linux" {
		return fmt.Errorf("isolated adapter requires a Linux Docker Engine: %w", ErrCapability)
	}
	d.checked = true
	return nil
}
func (d *dockerClient) api(ctx context.Context, method, path string, body, out any) error {
	if err := d.check(ctx); err != nil {
		return err
	}
	return d.request(ctx, method, "/v"+dockerAPIVersion+path, body, out)
}

type containerInspect struct {
	ID     string `json:"Id"`
	Config struct{ Labels map[string]string }
	State  struct {
		Running, Paused, Restarting, Dead bool
		Status                            string
		ExitCode                          int
		Error                             string
		StartedAt                         string
		FinishedAt                        string
		Pid                               int
	}
}

func (m *Manager) inspectDocker(ctx context.Context, r Record) (containerInspect, error) {
	var result containerInspect
	if m.docker == nil {
		return result, ErrCapability
	}
	id := r.ContainerID
	if id == "" {
		id = "blora-run-" + r.RunID
	}
	err := m.docker.api(ctx, "GET", "/containers/"+url.PathEscape(id)+"/json", nil, &result)
	if err != nil {
		return result, err
	}
	labels := result.Config.Labels
	if labels["dev.blora.run"] != r.RunID || labels["dev.blora.instance"] != r.InstanceID || labels["dev.blora.owner"] != r.Token {
		return result, ErrUnknown
	}
	if r.ContainerID != "" && result.ID != r.ContainerID {
		return result, ErrUnknown
	}
	return result, nil
}
func (m *Manager) startDocker(ctx context.Context, r Record, c model.InstanceConfig) (Record, error) {
	if m.docker == nil {
		return r, fmt.Errorf("Docker endpoint is not configured: %w", ErrCapability)
	}
	if c.Image == "" {
		return r, errors.New("container image is required; pull it using the image task first")
	}
	if m.options.InstanceRoot == "" {
		return r, errors.New("container instance root is not configured")
	}
	root, err := filepath.EvalSymlinks(m.options.InstanceRoot)
	if err != nil {
		return r, err
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return r, err
	}
	source := filepath.Join(root, r.InstanceID)
	uid, gid := uint32(65534), uint32(65534)
	if c.UID != nil {
		uid = *c.UID
	}
	if c.GID != nil {
		gid = *c.GID
	}
	if uid == 0 || gid == 0 {
		return r, errors.New("isolated container identity must be non-root")
	}
	if err = prepareContainerDirectory(source, uid, gid); err != nil {
		return r, err
	}
	actual, err := filepath.EvalSymlinks(source)
	if err != nil {
		return r, err
	}
	if actual != source {
		return r, errors.New("container root must not be a symbolic link")
	}
	working := c.Directory
	if working == "" {
		working = "/workspace"
	}
	if working != "/workspace" && !strings.HasPrefix(working, "/workspace/") {
		return r, errors.New("container working directory must be within /workspace")
	}
	if strings.Contains(working, "..") {
		return r, errors.New("invalid container working directory")
	}
	memory, pids, cpu := c.MemoryBytes, c.PidsLimit, c.CPUQuota
	if memory == 0 {
		memory = 512 << 20
	}
	if pids == 0 {
		pids = 256
	}
	if cpu == 0 {
		cpu = 100000
	}
	// This allowlist deliberately has no user-supplied HostConfig, mounts,
	// devices, namespaces, capabilities, socket, privileged or restart policy.
	body := map[string]any{
		"Image": c.Image, "Cmd": c.Command, "Entrypoint": []string{}, "Env": environment(c, r.Token), "User": fmt.Sprintf("%d:%d", uid, gid), "WorkingDir": working,
		"Labels":       map[string]string{"dev.blora.run": r.RunID, "dev.blora.instance": r.InstanceID, "dev.blora.owner": r.Token},
		"AttachStdout": true, "AttachStderr": true, "OpenStdin": true, "StdinOnce": false, "StopSignal": "SIGTERM",
		"HostConfig": map[string]any{
			"Privileged": false, "ReadonlyRootfs": true, "CapDrop": []string{"ALL"}, "SecurityOpt": []string{"no-new-privileges:true"},
			"NetworkMode": "bridge", "IpcMode": "private", "PidMode": "", "CgroupnsMode": "private",
			"Memory": memory, "MemorySwap": memory, "PidsLimit": pids, "CPUPeriod": 100000, "CPUQuota": cpu,
			"Mounts":        []map[string]any{{"Type": "bind", "Source": source, "Target": "/workspace", "ReadOnly": false, "BindOptions": map[string]any{"Propagation": "rprivate", "NonRecursive": true}}},
			"Tmpfs":         map[string]string{"/tmp": "rw,nosuid,nodev,noexec,size=67108864"},
			"RestartPolicy": map[string]string{"Name": "no"}, "LogConfig": map[string]any{"Type": "local", "Config": map[string]string{"max-size": "10m", "max-file": "3"}},
		},
	}
	r.Unit = "blora-run-" + r.RunID
	if err = m.save(r); err != nil {
		return r, err
	}
	var created struct {
		ID string `json:"Id"`
	}
	if err = m.docker.api(ctx, "POST", "/containers/create?name="+url.QueryEscape(r.Unit), body, &created); err != nil {
		return r, err
	}
	r.ContainerID = created.ID
	if err = m.save(r); err != nil {
		return r, err
	}
	if _, err = m.inspectDocker(ctx, r); err != nil {
		return r, err
	}
	err = m.docker.api(ctx, "POST", "/containers/"+url.PathEscape(r.ContainerID)+"/start", nil, nil)
	return r, err
}
func (m *Manager) observeDocker(ctx context.Context, r Record) (Observation, error) {
	o := Observation{State: "UNKNOWN", ObservedAt: time.Now().UTC()}
	i, err := m.inspectDocker(ctx, r)
	if err != nil {
		// A lost create response may leave no container. A durable intent plus
		// a successful Engine lookup of the exact unique name confirms absence.
		var e *engineError
		if errors.As(err, &e) && e.status == 404 {
			o.State = "STOPPED"
			o.Exited = true
			o.Diagnostic = "Engine confirms this run's container is absent"
			return o, nil
		}
		return o, err
	}
	if i.State.Running || i.State.Paused || i.State.Restarting {
		o.State = "RUNNING"
		if i.State.Pid > 0 {
			o.Processes = []Process{{PID: i.State.Pid}}
		}
		return o, nil
	}
	switch i.State.Status {
	case "created", "exited", "dead":
		o.Exited = true
		o.State = "STOPPED"
		o.Diagnostic = fmt.Sprintf("Engine state=%s exitCode=%d %s", i.State.Status, i.State.ExitCode, i.State.Error)
		return o, nil
	default:
		return o, ErrUnknown
	}
}
func (m *Manager) signalDocker(ctx context.Context, r Record, force bool, input string) error {
	if input != "" && !force {
		_, err := m.WriteInput(ctx, r, []byte(input))
		return err
	}
	i, err := m.inspectDocker(ctx, r)
	if err != nil {
		return err
	}
	if !i.State.Running && !i.State.Paused && !i.State.Restarting {
		return nil
	}
	signal := "SIGTERM"
	if force {
		signal = "SIGKILL"
	}
	return m.docker.api(ctx, "POST", "/containers/"+url.PathEscape(i.ID)+"/kill?signal="+signal, nil, nil)
}

func (m *Manager) writeDockerInput(ctx context.Context, r Record, data []byte) (int, error) {
	i, err := m.inspectDocker(ctx, r)
	if err != nil {
		return 0, err
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", m.docker.base+"/v"+dockerAPIVersion+"/containers/"+url.PathEscape(i.ID)+"/attach?stream=true&stdin=true&stdout=false&stderr=false", nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "tcp")
	resp, err := m.docker.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSwitchingProtocols {
		return 0, &engineError{resp.StatusCode, "stdin attach upgrade rejected"}
	}
	stream, ok := resp.Body.(io.ReadWriteCloser)
	if !ok {
		return 0, fmt.Errorf("Engine transport does not expose duplex stdin: %w", ErrCapability)
	}
	// Once upgraded, net/http may stop applying the request context to I/O.
	// Closing the connection enforces a bounded write even if the container is
	// not reading its stdin. Never replay a partially acknowledged write.
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			_ = stream.Close()
		case <-done:
		}
	}()
	return stream.Write(data)
}
func (m *Manager) cleanupDocker(ctx context.Context, r Record) error {
	i, err := m.inspectDocker(ctx, r)
	if err != nil {
		var e *engineError
		if errors.As(err, &e) && e.status == 404 {
			return nil
		}
		return err
	}
	// v=false is intentional: deleting a stopped running unit must preserve data.
	return m.docker.api(ctx, "DELETE", "/containers/"+url.PathEscape(i.ID)+"?v=false&force=false", nil, nil)
}

// ContainerLogs exposes only this run's bounded Engine log archive. It is a
// streaming reader, including Docker's multiplex headers; caller handles stream
// framing, consumer budgets and archive gaps. Closing it cancels the HTTP stream.
func (m *Manager) ContainerLogs(ctx context.Context, r Record, since time.Time, follow bool) (io.ReadCloser, error) {
	i, err := m.inspectDocker(ctx, r)
	if err != nil {
		return nil, err
	}
	query := url.Values{"stdout": {"true"}, "stderr": {"true"}, "timestamps": {"true"}, "tail": {"10000"}, "follow": {strconv.FormatBool(follow)}}
	if !since.IsZero() {
		query.Set("since", strconv.FormatInt(since.Unix(), 10))
	}
	req, err := http.NewRequestWithContext(ctx, "GET", m.docker.base+"/v"+dockerAPIVersion+"/containers/"+url.PathEscape(i.ID)+"/logs?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	client := *m.docker.client
	client.Timeout = 0
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, &engineError{resp.StatusCode, "log stream rejected"}
	}
	return resp.Body, nil
}

// Kept separate from source selection so container configuration never accepts
// arbitrary host paths. The parent directory is administrator controlled.
func newContainerDirectory(path string) (bool, error) {
	err := os.Mkdir(path, 0750)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrExist) {
		info, e := os.Lstat(path)
		if e != nil {
			return false, e
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return false, errors.New("container source is not a directory")
		}
		return false, nil
	}
	return false, err
}

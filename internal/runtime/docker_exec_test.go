package runtime

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"blora.dev/panel/internal/containerterm/helper"
	"blora.dev/panel/internal/model"
)

type execFixtureItem struct {
	command, env []string
	running      bool
	exit         int
	token        string
}
type execFixture struct {
	t                                                *testing.T
	mu                                               sync.Mutex
	r                                                Record
	server                                           *httptest.Server
	execs                                            map[string]*execFixtureItem
	connections                                      []net.Conn
	commands                                         int
	inputs                                           []string
	resized                                          bool
	probeFail, closeFail, foreignExec, missingLimits bool
	started                                          string
}

func newExecFixture(t *testing.T) (*Manager, *execFixture) {
	t.Helper()
	f := &execFixture{t: t, r: Record{RunID: model.ID(), InstanceID: model.ID(), Token: model.ID(), Backend: "docker", ContainerID: "owned-container", Phase: "running"}, execs: map[string]*execFixtureItem{}, started: "2026-09-09T00:00:00Z"}
	f.server = httptest.NewServer(http.HandlerFunc(f.handle))
	t.Cleanup(func() {
		f.mu.Lock()
		for _, conn := range f.connections {
			_ = conn.Close()
		}
		f.mu.Unlock()
		f.server.Close()
	})
	m, err := New(Options{StateRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	m.docker = &dockerClient{client: f.server.Client(), base: f.server.URL}
	return m, f
}
func (f *execFixture) handle(w http.ResponseWriter, q *http.Request) {
	if q.URL.Path == "/version" {
		_ = json.NewEncoder(w).Encode(map[string]string{"ApiVersion": "1.54", "MinAPIVersion": "1.44", "Os": "linux"})
		return
	}
	if q.URL.Path == "/v1.45/info" {
		f.mu.Lock()
		limits := !f.missingLimits
		f.mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{"OSType": "linux", "MemoryLimit": limits, "SwapLimit": true, "CpuCfsQuota": true, "PidsLimit": true})
		return
	}
	if q.URL.Path == "/v1.45/containers/"+f.r.ContainerID+"/json" {
		f.mu.Lock()
		started := f.started
		f.mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{"Id": f.r.ContainerID, "Config": map[string]any{"Labels": map[string]string{"dev.blora.run": f.r.RunID, "dev.blora.instance": f.r.InstanceID, "dev.blora.owner": f.r.Token}, "User": "65534:65534"}, "HostConfig": map[string]any{"Privileged": false, "PidMode": "", "CapDrop": []string{"ALL"}, "SecurityOpt": []string{"no-new-privileges:true"}}, "State": map[string]any{"Running": true, "Status": "running", "StartedAt": started}})
		return
	}
	if q.URL.Path == "/v1.45/containers/"+f.r.ContainerID+"/exec" {
		var body struct {
			Cmd, Env                     []string
			User                         string
			Privileged, AttachStdin, Tty bool
			ConsoleSize                  []int
		}
		if err := json.NewDecoder(q.Body).Decode(&body); err != nil {
			f.t.Error(err)
		}
		if body.User != "" || body.Privileged || len(body.Cmd) < 2 || body.Cmd[0] != helper.Path {
			f.t.Error("exec gained arbitrary privilege/user/helper selection")
		}
		item := &execFixtureItem{command: body.Cmd, env: body.Env}
		for _, env := range item.env {
			if strings.HasPrefix(env, helper.TokenEnvironment+"=") {
				item.token = strings.TrimPrefix(env, helper.TokenEnvironment+"=")
			}
		}
		f.mu.Lock()
		id := fmt.Sprintf("exec-%d", len(f.execs)+1)
		f.execs[id] = item
		if body.Cmd[1] == "serve" {
			f.commands++
			if !body.Tty || !body.AttachStdin {
				f.t.Error("PTY create did not request TTY/stdin")
			}
		}
		f.mu.Unlock()
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(map[string]string{"Id": id})
		return
	}
	if strings.HasPrefix(q.URL.Path, "/v1.45/exec/") {
		parts := strings.Split(strings.TrimPrefix(q.URL.Path, "/v1.45/exec/"), "/")
		if len(parts) != 2 {
			http.Error(w, "invalid", 400)
			return
		}
		id, action := parts[0], parts[1]
		f.mu.Lock()
		item := f.execs[id]
		f.mu.Unlock()
		if item == nil {
			http.Error(w, "missing", 404)
			return
		}
		switch action {
		case "json":
			f.mu.Lock()
			info := DockerExecInfo{ID: id, ContainerID: f.r.ContainerID, Running: item.running, ExitCode: item.exit, PID: 5678}
			if f.foreignExec {
				info.ContainerID = "foreign-container"
			}
			f.mu.Unlock()
			_ = json.NewEncoder(w).Encode(info)
			return
		case "resize":
			if q.URL.Query().Get("h") != "42" || q.URL.Query().Get("w") != "132" {
				f.t.Error("resize targeted wrong dimensions")
			}
			f.mu.Lock()
			f.resized = true
			f.mu.Unlock()
			w.WriteHeader(200)
			return
		case "start":
			f.start(w, q, item)
			return
		}
	}
	f.t.Errorf("unexpected or unsafe Engine API operation: %s %s", q.Method, q.URL.Path)
	http.Error(w, "unexpected", 500)
}
func (f *execFixture) start(w http.ResponseWriter, q *http.Request, item *execFixtureItem) {
	if q.Header.Get("Upgrade") != "tcp" {
		f.t.Error("exec stream was not upgraded")
	}
	var body struct{ Detach, Tty bool }
	if err := json.NewDecoder(q.Body).Decode(&body); err != nil {
		f.t.Error(err)
	}
	if body.Detach {
		f.t.Error("detached exec cannot supply terminal I/O")
	}
	conn, rw, err := w.(http.Hijacker).Hijack()
	if err != nil {
		f.t.Error(err)
		return
	}
	defer conn.Close()
	f.mu.Lock()
	f.connections = append(f.connections, conn)
	item.running = true
	f.mu.Unlock()
	_, _ = rw.WriteString("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n\r\n")
	_ = rw.Flush()
	if item.command[1] == "serve" {
		identity := helper.Identity{Version: 1, PID: 345, StartTicks: 67890, SessionID: 345, Token: item.token}
		raw, _ := json.Marshal(identity)
		_, _ = rw.WriteString(helper.Prefix + string(raw) + "\r\n")
		_ = rw.Flush()
		ack, err := rw.ReadString('\n')
		if err != nil {
			return
		}
		if ack != "BLORA-EXEC-GO "+item.token+"\n" {
			f.t.Error("launch acknowledgement mismatch")
			return
		}
		_, _ = rw.WriteString("\x1b[32m容器-ready\x1b[0m\r\n")
		_ = rw.Flush()
		for {
			line, err := rw.ReadString('\n')
			if err != nil {
				return
			}
			f.mu.Lock()
			f.inputs = append(f.inputs, line)
			f.mu.Unlock()
			_, _ = rw.WriteString("output:" + line)
			_ = rw.Flush()
		}
	}
	var raw []byte
	f.mu.Lock()
	switch item.command[1] {
	case "probe":
		if f.probeFail {
			item.exit = 127
			raw = []byte("helper missing")
		} else {
			raw, _ = json.Marshal(helper.Probe{Version: 1, PIDFD: true})
		}
	case "close":
		if f.closeFail {
			item.exit = 125
			raw = []byte("birth identity changed")
		} else {
			if len(item.command) != 5 || item.command[2] != "345" || item.command[3] != "67890" {
				f.t.Error("close did not bind helper PID and birth")
			}
			for _, other := range f.execs {
				if other.command[1] == "serve" {
					if item.command[4] != other.token {
						f.t.Error("close ownership token mismatch")
					}
					other.running = false
				}
			}
			raw = []byte(`{"closed":true}`)
		}
	default:
		f.t.Error("unknown helper operation")
	}
	item.running = false
	f.mu.Unlock()
	var header [8]byte
	header[0] = 1
	binary.BigEndian.PutUint32(header[4:], uint32(len(raw)))
	_, _ = rw.Write(header[:])
	_, _ = rw.Write(raw)
	_ = rw.Flush()
}

func TestDockerExecProtocolOwnershipPTYAndClose(t *testing.T) {
	m, f := newExecFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	p, err := m.StartDockerExec(ctx, f.r, DockerExecOptions{SessionID: model.ID(), Command: []string{"/bin/sh"}, Cols: 80, Rows: 24})
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	defer p.Close()
	reader := bufio.NewReader(p)
	line, err := reader.ReadString('\n')
	if err != nil || !strings.Contains(line, "容器-ready") {
		t.Fatalf("raw output missing: %v", err)
	}
	if _, err = p.Write([]byte("hello\n")); err != nil {
		t.Fatal(err)
	}
	line, err = reader.ReadString('\n')
	if err != nil || line != "output:hello\n" {
		t.Fatalf("write was not delivered once: %v", err)
	}
	if err = p.Resize(132, 42); err != nil {
		t.Fatal(err)
	}
	if err = p.Close(); err != nil {
		t.Fatal(err)
	}
	if err = p.Wait(); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.commands != 1 || !f.resized || len(f.inputs) != 1 {
		t.Fatal("exec side effects duplicated or resize missing")
	}
}

func TestDockerExecFailsClosedOnForeignIdentityAndMissingHelper(t *testing.T) {
	m, f := newExecFixture(t)
	r := f.r
	r.Token = model.ID()
	if _, err := m.StartDockerExec(context.Background(), r, DockerExecOptions{SessionID: model.ID(), Command: []string{"/bin/sh"}}); !errors.Is(err, ErrUnknown) {
		t.Fatalf("foreign labels accepted: %v", err)
	}
	f.mu.Lock()
	f.probeFail = true
	f.mu.Unlock()
	if _, err := m.StartDockerExec(context.Background(), f.r, DockerExecOptions{SessionID: model.ID(), Command: []string{"/bin/sh"}}); !errors.Is(err, ErrCapability) {
		t.Fatalf("missing helper accepted: %v", err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.commands != 0 {
		t.Fatal("user shell started before helper verification")
	}
}

func TestDockerExecUnconfirmedCloseAndContainerRestart(t *testing.T) {
	m, f := newExecFixture(t)
	p, err := m.StartDockerExec(context.Background(), f.r, DockerExecOptions{SessionID: model.ID(), Command: []string{"/bin/sh"}})
	if err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	f.closeFail = true
	f.mu.Unlock()
	if err = p.Close(); err == nil {
		t.Fatal("unconfirmed close reported success")
	}
	i, err := p.Inspect(context.Background())
	if err != nil || !i.Running {
		t.Fatal("close failure changed real exec state")
	}
	f.mu.Lock()
	f.started = "2026-09-09T01:00:00Z"
	f.mu.Unlock()
	if _, err = p.Inspect(context.Background()); !errors.Is(err, ErrUnknown) {
		t.Fatal("container birth change accepted")
	}
}

func TestCapabilitiesProbeEngineAndNativeConfiguration(t *testing.T) {
	m, f := newExecFixture(t)
	c := m.Capabilities(context.Background())
	if c["container"] != "available" || (nativeBackend() == "linux" && (c["native"] != "unavailable" || c["native.reason"] == "")) {
		t.Fatalf("capability probe differs: %+v", c)
	}
	f.mu.Lock()
	f.missingLimits = true
	f.mu.Unlock()
	c = m.Capabilities(context.Background())
	if c["container"] != "unavailable" || c["container.reason"] == "" {
		t.Fatal("Engine probe reused stale available flag")
	}
}

func TestDockerExecRealIsolatedSession(t *testing.T) {
	endpoint, image := os.Getenv("BLORA_TEST_DOCKER_ENDPOINT"), os.Getenv("BLORA_TEST_DOCKER_IMAGE")
	if endpoint == "" || image == "" {
		t.Skip("environment missing: authorized Engine and image with exec-helper required")
	}
	root := t.TempDir()
	if base := os.Getenv("BLORA_TEST_DOCKER_INSTANCE_ROOT"); base != "" {
		var err error
		root, err = os.MkdirTemp(base, "blora-exec-test-")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.RemoveAll(root) })
	}
	m, err := New(Options{StateRoot: t.TempDir(), InstanceRoot: root, DockerEndpoint: endpoint})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	r, err := m.Start(ctx, model.ID(), model.ID(), model.InstanceConfig{Mode: "isolated", Image: image, Command: []string{"/bin/sh", "-c", "while :; do sleep 30; done"}}, IO{})
	t.Cleanup(func() {
		cleanupCtx, c := context.WithTimeout(context.Background(), 10*time.Second)
		defer c()
		_, _ = m.Stop(cleanupCtx, r, StopPolicy{Force: true, KillWait: 3 * time.Second}, nil)
		if e := m.Cleanup(cleanupCtx, r); e != nil {
			t.Errorf("own run %s requires cleanup: %v", r.RunID, e)
		} else if r.ContainerID != "" {
			_, e := m.inspectDocker(cleanupCtx, r)
			var engine *engineError
			if !errors.As(e, &engine) || engine.status != 404 {
				t.Errorf("own container absence not confirmed: %v", e)
			} else {
				t.Logf("removed own test run %s container %s", r.RunID, r.ContainerID)
			}
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	p, err := m.StartDockerExec(ctx, r, DockerExecOptions{SessionID: model.ID(), Command: []string{"/bin/sh", "-c", `stty -echo; printf '\033[32m中文-ready\033[0m\n'; while IFS= read -r line; do if [ "$line" = size ]; then stty size; else printf 'received:%s\n' "$line"; fi; done`}})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	reader := bufio.NewReader(p)
	line, err := reader.ReadString('\n')
	if err != nil || !strings.Contains(line, "中文-ready") {
		t.Fatalf("container PTY output missing: %v", err)
	}
	if err = p.Resize(132, 42); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Write([]byte("size\n")); err != nil {
		t.Fatal(err)
	}
	line, err = reader.ReadString('\n')
	if err != nil || !strings.Contains(line, "42 132") {
		t.Fatalf("actual Docker PTY dimensions differ: %v", err)
	}
	if _, err = p.Write([]byte("once\n")); err != nil {
		t.Fatal(err)
	}
	line, err = reader.ReadString('\n')
	if err != nil || !strings.Contains(line, "received:once") {
		t.Fatalf("container input absent: %v", err)
	}
	second, err := m.StartDockerExec(ctx, r, DockerExecOptions{SessionID: model.ID(), Command: []string{"/bin/sh", "-c", `stty -echo; printf 'second-ready\n'; while IFS= read -r line; do printf 'second:%s\n' "$line"; done`}})
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	secondReader := bufio.NewReader(second)
	line, err = secondReader.ReadString('\n')
	if err != nil || !strings.Contains(line, "second-ready") {
		t.Fatal("second exec did not start")
	}
	if err = p.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = second.Write([]byte("still-alive\n")); err != nil {
		t.Fatal(err)
	}
	line, err = secondReader.ReadString('\n')
	if err != nil || !strings.Contains(line, "second:still-alive") {
		t.Fatal("closing first exec affected another terminal")
	}
	o, err := m.Observe(ctx, r)
	if err != nil || o.Exited || o.State != "RUNNING" {
		t.Fatalf("closing exec affected the instance: %+v %v", o, err)
	}
	if _, err = os.Stat(filepath.Join(root, r.InstanceID)); err != nil {
		t.Fatal("instance files affected")
	}
}

func TestDockerCapabilitiesRealEngine(t *testing.T) {
	endpoint := os.Getenv("BLORA_TEST_DOCKER_ENDPOINT")
	if endpoint == "" {
		t.Skip("environment missing: authorized Engine endpoint required")
	}
	m, err := New(Options{StateRoot: t.TempDir(), DockerEndpoint: endpoint})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	capabilities := m.Capabilities(ctx)
	if capabilities["container"] != "available" {
		t.Fatalf("actual Engine isolation capability probe failed: %s", capabilities["container.reason"])
	}
	t.Logf("actual Engine supports isolated runtime API %s", capabilities["container.api"])
}

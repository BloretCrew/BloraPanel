package master

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"blora.dev/panel/internal/daemon"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
	"github.com/coder/websocket"
	"golang.org/x/crypto/bcrypt"
)

type testClient struct {
	t          *testing.T
	client     *http.Client
	base, csrf string
}

func (c *testClient) request(method, path string, body any, key string, status int) map[string]json.RawMessage {
	c.t.Helper()
	var data []byte
	if body != nil {
		var err error
		data, err = json.Marshal(body)
		if err != nil {
			c.t.Fatal(err)
		}
	}
	req, err := http.NewRequest(method, c.base+"/api/v1"+path, bytes.NewReader(data))
	if err != nil {
		c.t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", c.csrf)
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		c.t.Fatal(err)
	}
	if resp.StatusCode != status {
		c.t.Fatalf("%s %s status %d expected %d: %s", method, path, resp.StatusCode, status, b)
	}
	var result map[string]json.RawMessage
	if err := json.Unmarshal(b, &result); err != nil {
		c.t.Fatal(err)
	}
	return result
}
func (c *testClient) raw(method, path string, body any, key string, status int) []byte {
	c.t.Helper()
	var data []byte
	if body != nil {
		data, _ = json.Marshal(body)
	}
	req, err := http.NewRequest(method, c.base+"/api/v1"+path, bytes.NewReader(data))
	if err != nil {
		c.t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", c.csrf)
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer resp.Body.Close()
	b, readErr := io.ReadAll(io.LimitReader(resp.Body, (16<<20)+1))
	if readErr != nil || len(b) > 16<<20 {
		c.t.Fatalf("raw response exceeds test bound or failed: %v", readErr)
	}
	if resp.StatusCode != status {
		c.t.Fatalf("%s %s status %d expected %d: %s", method, path, resp.StatusCode, status, b)
	}
	return b
}
func (c *testClient) login(name, password string) {
	result := c.request("POST", "/login", map[string]string{"name": name, "password": password}, "", 200)
	if err := json.Unmarshal(result["csrfToken"], &c.csrf); err != nil {
		c.t.Fatal(err)
	}
}
func eventually(t *testing.T, timeout time.Duration, check func() bool) {
	t.Helper()
	end := time.Now().Add(timeout)
	for time.Now().Before(end) {
		if check() {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatal("condition did not become true before deadline")
}

func TestFailedStopBlocksReplacementButOtherNodeRemainsUsable(t *testing.T) {
	f := newFileFixture(t)
	create := func(index int, mode string) model.Instance {
		value := f.admin.request("POST", "/instances", map[string]any{"nodeId": f.instances[index].NodeID, "name": "stop-boundary", "config": model.InstanceConfig{Mode: "native", Directory: f.roots[index], Command: nativeTestCommand(t, mode), Environment: nativeTestEnvironment(), StopSeconds: 1, KillSeconds: 1, Escalate: false}}, model.ID(), 201)
		var instance model.Instance
		if err := json.Unmarshal(value["instance"], &instance); err != nil {
			t.Fatal(err)
		}
		return instance
	}
	stubborn := create(0, "stubborn")
	other := create(1, "idle")
	action := func(instance model.Instance, name string) model.Task {
		return parseFileTask(t, f.admin.request("POST", "/instances/"+instance.ID+"/actions", map[string]string{"action": name}, model.ID(), 202))
	}
	f.awaitFile(t, action(stubborn, "start"), model.Succeeded)
	t.Cleanup(func() { f.awaitFile(t, action(stubborn, "kill"), model.Succeeded) })
	eventually(t, 3*time.Second, func() bool { _, err := os.Stat(filepath.Join(f.roots[0], "ready")); return err == nil })
	before, err := f.store.Instance(context.Background(), stubborn.ID)
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	stop := action(stubborn, "stop")
	f.awaitFile(t, action(other, "start"), model.Succeeded)
	t.Cleanup(func() { f.awaitFile(t, action(other, "kill"), model.Succeeded) })
	f.awaitFile(t, stop, model.Failed)
	if time.Since(started) > 5*time.Second {
		t.Fatal("failed stop exceeded finite budget")
	}
	f.awaitFile(t, action(stubborn, "start"), model.Failed)
	after, err := f.store.Instance(context.Background(), stubborn.ID)
	if err != nil || after.RunID != before.RunID {
		t.Fatal("failed stop allowed replacement run", err)
	}
	body, err := os.ReadFile(filepath.Join(f.roots[0], "starts"))
	if err != nil || string(body) != "x" {
		t.Fatalf("duplicate process started: %q %v", body, err)
	}
	f.awaitFile(t, action(other, "stop"), model.Succeeded)
}

func TestTaskCancellationReplayAfterTerminalKeepsReceipt(t *testing.T) {
	f := newFileFixture(t)
	f.stopNode(0)
	task, _, err := f.store.Accept(context.Background(), model.Task{
		ActorID:   f.adminUser.ID,
		RequestID: model.ID(),
		Resource:  instanceRef(f.instances[0]),
		Action:    "instance.start",
		Payload: func() []byte {
			b, _ := json.Marshal(f.instances[0].Config)
			return b
		}(),
	})
	if err != nil {
		t.Fatal(err)
	}
	first := parseFileTask(t, f.admin.request("POST", "/tasks/"+task.ID+"/cancel", nil, "cancel-http", 202))
	if first.State != model.Cancelled || first.CancellationRequestID != "cancel-http" {
		t.Fatalf("unexpected cancellation receipt: %+v", first)
	}
	replayed := parseFileTask(t, f.admin.request("POST", "/tasks/"+task.ID+"/cancel", nil, "cancel-http-retry", 202))
	if replayed.State != model.Cancelled || replayed.Revision != first.Revision || replayed.CancellationRequestID != first.CancellationRequestID {
		t.Fatalf("terminal cancellation replay changed receipt: first=%+v replay=%+v", first, replayed)
	}
}

func TestRestartTaskSurvivesMasterAndDatabaseReopen(t *testing.T) {
	f, reopen := newRestartTransferFixture(t)
	config := model.InstanceConfig{Mode: "native", Directory: f.roots[0], Command: nativeTestCommand(t, "generations"), Environment: nativeTestEnvironment(), StopSeconds: 2, KillSeconds: 2, Escalate: true}
	created := f.admin.request("POST", "/instances", map[string]any{"nodeId": f.instances[0].NodeID, "name": "master-reopen-run", "config": config}, model.ID(), 201)
	var instance model.Instance
	if err := json.Unmarshal(created["instance"], &instance); err != nil {
		t.Fatal(err)
	}
	action := func(name, key string) model.Task {
		return parseFileTask(t, f.admin.request("POST", "/instances/"+instance.ID+"/actions", map[string]string{"action": name}, key, 202))
	}
	f.awaitFile(t, action("start", model.ID()), model.Succeeded)
	t.Cleanup(func() { f.awaitFile(t, action("stop", model.ID()), model.Succeeded) })
	path := filepath.Join(f.roots[0], "generations")
	eventually(t, 3*time.Second, func() bool { body, _ := os.ReadFile(path); return string(body) == "x" })
	key := model.ID()
	request := action("restart", key)
	eventually(t, 3*time.Second, func() bool {
		return parseFileTask(t, f.admin.request("GET", "/tasks/"+request.ID, nil, "", 200)).State == model.Running
	})
	reopen()
	if retry := action("restart", key); retry.ID != request.ID {
		t.Fatal("database reopen lost request identity")
	}
	f.awaitFile(t, request, model.Succeeded)
	eventually(t, 3*time.Second, func() bool { body, _ := os.ReadFile(path); return string(body) == "xx" })
	before, err := f.store.Instance(context.Background(), instance.ID)
	if err != nil || before.RunID == "" {
		t.Fatal("missing recovered run identity", err)
	}
	reopen()
	if retry := action("restart", key); retry.ID != request.ID {
		t.Fatal("completed request duplicated after second reopen")
	}
	after, err := f.store.Instance(context.Background(), instance.ID)
	if err != nil || after.RunID != before.RunID {
		t.Fatal("completed restart changed run after reopen", err)
	}
	body, err := os.ReadFile(path)
	if err != nil || string(body) != "xx" {
		t.Fatalf("process starts=%q err=%v", body, err)
	}
}

func TestTwoNodesAuthorizationAndDurableLifecycle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	root := t.TempDir()
	store, err := storage.Open(filepath.Join(root, "master.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	password := model.ID()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	admin, err := store.CreateUser(ctx, model.User{Name: "admin", Admin: true}, hash)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := store.CreateUser(ctx, model.User{Name: "reader"}, hash)
	if err != nil {
		t.Fatal(err)
	}
	var app *Server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { app.ServeHTTP(w, r) }))
	defer server.Close()
	app = New(Options{Store: store, Origin: server.URL})
	defer app.Close()
	ca := filepath.Join(root, "ca.crt")
	if err := os.WriteFile(ca, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0600); err != nil {
		t.Fatal(err)
	}
	newClient := func() *testClient {
		jar, _ := cookiejar.New(nil)
		client := *server.Client()
		client.Jar = jar
		client.Timeout = 10 * time.Second
		return &testClient{t: t, client: &client, base: server.URL}
	}
	a := newClient()
	a.login(admin.Name, password)
	b := newClient()
	b.login(reader.Name, password)
	var daemons []*daemon.Daemon
	var wg sync.WaitGroup
	defer func() {
		cancel()
		for _, d := range daemons {
			_ = d.Close()
		}
		wg.Wait()
	}()
	for _, name := range []string{"node-a", "node-b"} {
		result := a.request("POST", "/nodes/enrollments", map[string]string{"name": name}, model.ID(), 201)
		var token string
		if err := json.Unmarshal(result["token"], &token); err != nil {
			t.Fatal(err)
		}
		file := filepath.Join(root, name+".enrollment")
		if err := os.WriteFile(file, []byte(token), 0600); err != nil {
			t.Fatal(err)
		}
		d, err := daemon.New(daemon.Config{StateDir: filepath.Join(root, name), MasterURL: server.URL, CAFile: ca, EnrollmentFile: file, AllowPGIDFallback: true})
		if err != nil {
			t.Fatal(err)
		}
		daemons = append(daemons, d)
		wg.Add(1)
		go func() { defer wg.Done(); _ = d.Run(ctx) }()
	}
	eventually(t, 5*time.Second, func() bool { app.mu.Lock(); defer app.mu.Unlock(); return len(app.peers) == 2 })
	b.request("GET", "/nodes", nil, "", 200)
	var instances []model.Instance
	for index, d := range daemons {
		config := model.InstanceConfig{Mode: "native", Command: nativeTestCommand(t, "stubborn-idle"), Environment: nativeTestEnvironment(), StopSeconds: 1, KillSeconds: 2, Escalate: true}
		result := a.request("POST", "/instances", map[string]any{"nodeId": d.NodeID(), "name": []string{"instance-a", "instance-b"}[index], "config": config}, model.ID(), 201)
		var instance model.Instance
		if err := json.Unmarshal(result["instance"], &instance); err != nil {
			t.Fatal(err)
		}
		instances = append(instances, instance)
		a.request("POST", "/grants", model.Grant{UserID: reader.ID, Resource: instanceRef(instance), Action: "instance.read"}, model.ID(), 200)
	}
	visible := b.request("GET", "/instances", nil, "", 200)
	var list []model.Instance
	if err := json.Unmarshal(visible["items"], &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("authorized cross-node aggregate = %d", len(list))
	}
	creatable := b.request("GET", "/nodes/creatable", nil, "", 200)
	if string(creatable["items"]) != "[]" {
		t.Fatal("read permission leaked create authority")
	}
	b.request("POST", "/instances/"+instances[0].ID+"/actions", map[string]string{"action": "start"}, model.ID(), 403)
	a.request("POST", "/grants", model.Grant{UserID: reader.ID, Resource: model.ResourceRef{Kind: "node", ID: daemons[1].NodeID()}, Action: "instance.create"}, model.ID(), 200)
	b.request("POST", "/instances", map[string]any{"nodeId": daemons[1].NodeID(), "name": "cannot-escape", "config": instances[0].Config}, model.ID(), 403)
	getTask := func(id string) model.Task {
		result := a.request("GET", "/tasks/"+id, nil, "", 200)
		var task model.Task
		if err := json.Unmarshal(result["task"], &task); err != nil {
			t.Fatal(err)
		}
		return task
	}
	action := func(index int, action, key string) model.Task {
		result := a.request("POST", "/instances/"+instances[index].ID+"/actions", map[string]string{"action": action}, key, 202)
		var task model.Task
		if err := json.Unmarshal(result["task"], &task); err != nil {
			t.Fatal(err)
		}
		return task
	}
	await := func(task model.Task) model.Task {
		eventually(t, 12*time.Second, func() bool { task = getTask(task.ID); return task.State.Terminal() })
		if task.State != model.Succeeded {
			t.Fatalf("task %s %s: %s", task.Action, task.State, task.Error)
		}
		return task
	}
	start := action(0, "start", model.ID())
	await(start)
	key := model.ID()
	first := action(0, "restart", key)
	second := action(0, "restart", key)
	if first.ID != second.ID {
		t.Fatal("same request created multiple restart tasks")
	}
	// Lose the management response while the actual restart is still stopping
	// the TERM-ignoring old run. The accepted request must survive reconnection.
	eventually(t, 5*time.Second, func() bool { return getTask(first.ID).State == model.Running })
	app.mu.Lock()
	restartingPeer := app.peers[daemons[0].NodeID()]
	app.mu.Unlock()
	if restartingPeer == nil {
		t.Fatal("restart node disappeared before disconnect injection")
	}
	if err := restartingPeer.conn.Close(); err != nil {
		t.Fatal(err)
	}
	eventually(t, 8*time.Second, func() bool {
		app.mu.Lock()
		defer app.mu.Unlock()
		peer := app.peers[daemons[0].NodeID()]
		return peer != nil && peer.generation > restartingPeer.generation
	})
	if retry := action(0, "restart", key); retry.ID != first.ID {
		t.Fatal("reconnect retry created a replacement restart task")
	}
	await(first)
	// Duplicate after completion must still resolve to the original task and run.
	if again := action(0, "restart", key); again.ID != first.ID {
		t.Fatal("completed retry duplicated")
	}
	var restarted map[string]string
	if err := json.Unmarshal(getTask(first.ID).Result, &restarted); err != nil {
		t.Fatal(err)
	}
	runID := restarted["runId"]
	if runID == "" {
		t.Fatal("missing actual run identity")
	}
	app.mu.Lock()
	peer := app.peers[daemons[0].NodeID()]
	app.mu.Unlock()
	_ = peer.conn.Close()
	eventually(t, 6*time.Second, func() bool {
		app.mu.Lock()
		defer app.mu.Unlock()
		p := app.peers[daemons[0].NodeID()]
		return p != nil && p.generation > peer.generation
	})
	current, err := store.Instance(ctx, instances[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if current.RunID != runID {
		t.Fatal("management reconnect restarted instance")
	}
	await(action(0, "stop", model.ID()))
	// Existing browser streams are cancelled at grant revocation.
	cookieURL, _ := http.NewRequest("GET", server.URL, nil)
	header := http.Header{"Origin": []string{server.URL}}
	for _, cookie := range b.client.Jar.Cookies(cookieURL.URL) {
		header.Add("Cookie", cookie.String())
	}
	ws, _, err := websocket.Dial(ctx, "wss"+server.URL[5:]+"/api/v1/events", &websocket.DialOptions{HTTPClient: b.client, HTTPHeader: header})
	if err != nil {
		t.Fatal(err)
	}
	defer ws.CloseNow()
	eventually(t, time.Second, func() bool { app.mu.Lock(); defer app.mu.Unlock(); return len(app.userStreams[reader.ID]) == 1 })
	a.request("DELETE", "/grants", model.Grant{UserID: reader.ID, Resource: instanceRef(instances[0]), Action: "instance.read"}, model.ID(), 200)
	readCtx, stop := context.WithTimeout(ctx, 2*time.Second)
	defer stop()
	for {
		_, _, err = ws.Read(readCtx)
		if err != nil {
			break
		}
	}
	if readCtx.Err() != nil {
		t.Fatal("revoked stream did not terminate")
	}
	b.request("GET", "/instances/"+instances[0].ID, nil, "", 403)
	t.Log("real TLS Master + two signed Daemons; scoped aggregate, native isolation denial, duplicate restart, reconnect without restart, process stop, stream revocation passed")
}

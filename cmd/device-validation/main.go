// device-validation exercises task-owned real Master/Daemon processes. Reports
// contain static checks only; credentials, process logs and state stay private.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"blora.dev/panel/internal/backup"
	"blora.dev/panel/internal/extensions"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/monitor"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/terminal"
	"github.com/coder/websocket"
)

type check struct {
	Name    string  `json:"name"`
	Status  string  `json:"status"`
	Seconds float64 `json:"seconds"`
}
type report struct {
	Platform    string   `json:"platform"`
	Started     string   `json:"startedAt"`
	Finished    string   `json:"finishedAt"`
	Outcome     string   `json:"outcome"`
	Checks      []check  `json:"checks"`
	Limitations []string `json:"limitations"`
}
type child struct {
	cmd     *exec.Cmd
	done    chan error
	log     *os.File
	stopped bool
}
type fixture struct {
	ctx                                                context.Context
	root, master, daemon, helper, static, origin, csrf string
	http                                               *http.Client
	children                                           []*child
	masterChild, daemonChild                           *child
	instance, session                                  string
	retainedTask                                       string
	config                                             string
	browserScript, node                                string
	stateParent                                        string
	r                                                  report
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--owned-instance" {
		fmt.Println("BLORA_OWNED_INSTANCE_READY")
		deadline := time.Now().Add(9 * time.Minute)
		for time.Now().Before(deadline) {
			time.Sleep(time.Second)
		}
		return
	}
	master := flag.String("master", "", "built Master executable")
	daemon := flag.String("daemon", "", "built Daemon executable")
	static := flag.String("static", "web/dist", "compiled frontend")
	output := flag.String("report", "device-report.json", "sanitized JSON output")
	browser := flag.String("browser-script", "", "optional real Chromium UI checker; omission is recorded")
	node := flag.String("node", "node", "Node.js executable")
	stateParent := flag.String("state-parent", "", "parent of a fresh private temporary fixture")
	flag.Parse()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	ctx, timeout := context.WithTimeout(ctx, 8*time.Minute)
	defer timeout()
	f := &fixture{ctx: ctx, master: *master, daemon: *daemon, static: *static, r: report{Platform: runtime.GOOS + "/" + runtime.GOARCH, Started: time.Now().UTC().Format(time.RFC3339Nano), Checks: []check{}, Limitations: []string{"Single-node real backend API/terminal WSS and optional Chromium UI smoke; not native IME/notification-center or performance acceptance.", "No services/tasks/firewall changes, Docker, release packaging/upgrade/rollback, disk exhaustion or power loss.", "Restart is only after the owned instance and terminal have confirmed stopped; no live-run crash claim."}}}
	f.browserScript, f.node = *browser, *node
	f.stateParent = *stateParent
	if *browser == "" {
		f.r.Limitations = append(f.r.Limitations, "Browser UI checker omitted; API/WSS passes cannot certify frontend refresh.")
	}
	err := f.run()
	err = errors.Join(err, f.cleanup())
	f.r.Finished = time.Now().UTC().Format(time.RFC3339Nano)
	f.r.Outcome = "CHECKS_PASSED_WITH_SCOPE_LIMITS"
	if err != nil {
		f.r.Outcome = "FAILED"
	}
	b, _ := json.MarshalIndent(f.r, "", "  ")
	if e := os.WriteFile(*output, b, 0600); e != nil {
		fmt.Fprintln(os.Stderr, "FAIL: cannot write sanitized report")
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "FAIL: "+err.Error())
		os.Exit(1)
	}
	fmt.Println("PASS: real device backend checks; scope limits recorded")
}
func (f *fixture) stage(name string, fn func() error) error {
	fmt.Println("RUN: " + name)
	at := time.Now()
	err := fn()
	s := "PASS"
	if err != nil {
		s = "FAIL"
	}
	f.r.Checks = append(f.r.Checks, check{name, s, time.Since(at).Seconds()})
	fmt.Println(s + ": " + name)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}
func (f *fixture) start(binary string, args ...string) (*child, error) {
	log, err := os.OpenFile(filepath.Join(f.root, fmt.Sprintf("private-process-%d.log", len(f.children))), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, errors.New("private log creation failed")
	}
	cmd := exec.Command(binary, args...)
	cmd.Stdout = log
	cmd.Stderr = log
	c := &child{cmd: cmd, log: log, done: make(chan error, 1)}
	if err = cmd.Start(); err != nil {
		log.Close()
		return nil, errors.New("owned process startup failed")
	}
	f.children = append(f.children, c)
	go func() { c.done <- cmd.Wait() }()
	return c, nil
}
func (f *fixture) stop(c *child) error {
	if c == nil || c.stopped {
		return nil
	}
	select {
	case <-c.done:
		c.stopped = true
		c.log.Close()
		return nil
	default:
	}
	if err := c.cmd.Process.Kill(); err != nil {
		return errors.New("owned process stop failed")
	}
	select {
	case <-c.done:
		c.stopped = true
		c.log.Close()
		return nil
	case <-time.After(15 * time.Second):
		return errors.New("owned process exit unconfirmed")
	}
}
func (f *fixture) call(method, path string, value, out any, status int) error {
	return f.callKey(method, path, value, out, status, model.ID())
}
func (f *fixture) callKey(method, path string, value, out any, status int, key string) error {
	var body io.Reader
	if value != nil {
		b, e := json.Marshal(value)
		if e != nil {
			return e
		}
		body = bytes.NewReader(b)
	}
	req, e := http.NewRequestWithContext(f.ctx, method, f.origin+path, body)
	if e != nil {
		return e
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", f.origin)
	req.Header.Set("X-CSRF-Token", f.csrf)
	req.Header.Set("Idempotency-Key", key)
	res, e := f.http.Do(req)
	if e != nil {
		return errors.New("management connection unavailable")
	}
	defer res.Body.Close()
	if res.StatusCode != status {
		return fmt.Errorf("unexpected HTTP status %d (wanted %d)", res.StatusCode, status)
	}
	if out != nil {
		return json.NewDecoder(io.LimitReader(res.Body, 4<<20)).Decode(out)
	}
	return nil
}
func (f *fixture) until(fn func() bool) error {
	timer := time.NewTimer(35 * time.Second)
	defer timer.Stop()
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		if fn() {
			return nil
		}
		select {
		case <-f.ctx.Done():
			return errors.New("device validation cancelled")
		case <-timer.C:
			return errors.New("bounded readiness deadline exceeded")
		case <-tick.C:
		}
	}
}
func (f *fixture) task(method, path string, value any) (model.Task, error) {
	var out struct {
		Task model.Task `json:"task"`
	}
	if e := f.call(method, path, value, &out, 202); e != nil {
		return out.Task, e
	}
	id := out.Task.ID
	e := f.until(func() bool {
		if f.call("GET", "/api/v1/tasks/"+id, nil, &out, 200) != nil {
			return false
		}
		return out.Task.State == model.Succeeded || out.Task.State == model.Failed || out.Task.State == model.Interrupted || out.Task.State == model.Cancelled
	})
	if e != nil {
		return out.Task, e
	}
	if out.Task.State != model.Succeeded {
		return out.Task, errors.New("real backend task did not succeed")
	}
	f.retainedTask = out.Task.ID
	return out.Task, nil
}
func (f *fixture) login(password string) error {
	var out struct {
		CSRF string `json:"csrfToken"`
	}
	f.csrf = ""
	if e := f.call("POST", "/api/v1/login", map[string]string{"name": "admin", "password": password}, &out, 200); e != nil {
		return e
	}
	f.csrf = out.CSRF
	return nil
}
func (f *fixture) run() error {
	if f.master == "" || f.daemon == "" {
		return errors.New("Master and Daemon executables required")
	}
	var e error
	f.helper, e = os.Executable()
	if e != nil {
		return e
	}
	if f.stateParent != "" {
		if e = os.MkdirAll(f.stateParent, 0700); e != nil {
			return errors.New("private parent creation failed")
		}
	}
	f.root, e = os.MkdirTemp(f.stateParent, "blora-device-")
	if e != nil {
		return errors.New("temporary root creation failed")
	}
	password := model.ID() + model.ID()
	pass := filepath.Join(f.root, "private-password")
	if e = os.WriteFile(pass, []byte(password), 0600); e != nil {
		return e
	}
	state := filepath.Join(f.root, "master")
	if e = f.stage("initialize-private-master", func() error {
		cmd := exec.CommandContext(f.ctx, f.master, "--state-dir", state, "--init", "--password-file", pass)
		if cmd.Run() != nil {
			return errors.New("initialization failed")
		}
		return nil
	}); e != nil {
		return e
	}
	_ = os.Remove(pass)
	listener, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		return e
	}
	addr := listener.Addr().String()
	listener.Close()
	f.origin = "https://" + addr
	pem, e := os.ReadFile(filepath.Join(state, "tls.crt"))
	if e != nil {
		return e
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return errors.New("local certificate invalid")
	}
	jar, _ := cookiejar.New(nil)
	f.http = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}}, Jar: jar, Timeout: 5 * time.Second}
	startMaster := func() error {
		var e error
		f.masterChild, e = f.start(f.master, "--state-dir", state, "--listen", addr, "--origin", f.origin, "--static-dir", f.static, "--extensions-dir", filepath.Join(f.root, "extensions"))
		if e != nil {
			return e
		}
		return f.until(func() bool { return f.call("GET", "/healthz", nil, nil, 200) == nil })
	}
	if e = f.stage("real-tls-static-login", func() error {
		if e := startMaster(); e != nil {
			return e
		}
		if e := f.call("GET", "/", nil, nil, 200); e != nil {
			return e
		}
		return f.login(password)
	}); e != nil {
		return e
	}
	var node string
	if e = f.stage("enroll-real-daemon", func() error {
		var out struct {
			Token string `json:"token"`
		}
		if e := f.call("POST", "/api/v1/nodes/enrollments", map[string]string{"name": "owned-device-node"}, &out, 201); e != nil {
			return e
		}
		ticket := filepath.Join(f.root, "private-ticket")
		if e := os.WriteFile(ticket, []byte(out.Token), 0600); e != nil {
			return e
		}
		f.config = filepath.Join(f.root, "private-daemon.json")
		b, _ := json.Marshal(map[string]any{"stateDir": filepath.Join(f.root, "daemon"), "masterUrl": f.origin, "caFile": filepath.Join(state, "tls.crt"), "enrollmentFile": ticket, "allowPGIDFallback": runtime.GOOS == "linux"})
		if e := os.WriteFile(f.config, b, 0600); e != nil {
			return e
		}
		var e error
		f.daemonChild, e = f.start(f.daemon, "--config", f.config)
		if e != nil {
			return e
		}
		return f.until(func() bool {
			var out struct {
				Items []model.Node `json:"items"`
			}
			if f.call("GET", "/api/v1/nodes", nil, &out, 200) != nil || len(out.Items) != 1 {
				return false
			}
			node = out.Items[0].ID
			return out.Items[0].State == "ONLINE"
		})
	}); e != nil {
		return e
	}
	var out struct {
		Instance model.Instance `json:"instance"`
	}
	if e = f.stage("create-stopped-owned-instance", func() error {
		dir := filepath.Join(f.root, "resource")
		if e := os.Mkdir(dir, 0700); e != nil {
			return e
		}
		config := model.InstanceConfig{Mode: "native", Directory: dir, Command: []string{f.helper, "--owned-instance"}, StopSeconds: 1, KillSeconds: 3, Escalate: true}
		if e := f.call("POST", "/api/v1/instances", map[string]any{"nodeId": node, "name": "owned-device-instance", "config": config}, &out, 201); e != nil {
			return e
		}
		f.instance = out.Instance.ID
		if out.Instance.State != "STOPPED" {
			return errors.New("new instance unexpectedly running")
		}
		return nil
	}); e != nil {
		return e
	}
	base := "/api/v1/instances/" + f.instance
	const content = "Blora real device UTF-8: 中文\n"
	if e = f.stage("real-user-permission-denial", func() error {
		memberPassword := model.ID() + model.ID()
		if e := f.call("POST", "/api/v1/users", map[string]any{"name": "owned-member", "password": memberPassword, "admin": false}, nil, 201); e != nil {
			return e
		}
		oldClient, oldCSRF := f.http, f.csrf
		jar, _ := cookiejar.New(nil)
		f.http = &http.Client{Transport: oldClient.Transport, Jar: jar, Timeout: 5 * time.Second}
		f.csrf = ""
		defer func() { f.http = oldClient; f.csrf = oldCSRF }()
		var result struct {
			CSRF string `json:"csrfToken"`
		}
		if e := f.call("POST", "/api/v1/login", map[string]string{"name": "owned-member", "password": memberPassword}, &result, 200); e != nil {
			return e
		}
		f.csrf = result.CSRF
		if e := f.call("GET", base+"/files/content?path=device.txt", nil, nil, 403); e != nil {
			return e
		}
		return f.call("POST", base+"/actions", map[string]string{"action": "start"}, nil, 403)
	}); e != nil {
		return e
	}
	if e = f.stage("real-file-save-and-read", func() error {
		_, e := f.task("PUT", base+"/files/content", map[string]string{"path": "device.txt", "text": content, "version": "missing"})
		if e != nil {
			return e
		}
		return f.verifyText(base, content)
	}); e != nil {
		return e
	}
	if e = f.stage("real-backup-restore", func() error { return f.backupCheck(base, content) }); e != nil {
		return e
	}
	if e = f.stage("real-monitor-sampling", func() error {
		return f.until(func() bool {
			var p monitor.Point
			return f.call("GET", "/api/v1/nodes/"+node+"/metrics", nil, &p, 200) == nil && !p.Stale && !p.ObservedAt.IsZero() && p.MemoryTotal > 0 && p.DiskTotal > 0
		})
	}); e != nil {
		return e
	}
	if e = f.stage("real-extension-install-upgrade-remove", func() error { return f.extensionCheck() }); e != nil {
		return e
	}
	if e = f.stage("real-instance-start-stop", func() error {
		if _, e := f.task("POST", base+"/actions", map[string]string{"action": "start"}); e != nil {
			return e
		}
		if e := f.call("GET", base, nil, &out, 200); e != nil {
			return e
		}
		if out.Instance.State != "RUNNING" {
			return errors.New("instance not actually running")
		}
		_, e := f.task("POST", base+"/actions", map[string]string{"action": "stop"})
		return e
	}); e != nil {
		return e
	}
	if e = f.stage("real-terminal-wss-input-resume-close", func() error { return f.terminalCheck(base) }); e != nil {
		return e
	}
	if f.browserScript != "" {
		if e = f.stage("real-chromium-editor-refresh-save", func() error {
			file := filepath.Join(f.root, "private-browser.json")
			b, _ := json.Marshal(map[string]string{"url": f.origin, "name": "admin", "password": password, "instance": f.instance})
			if e := os.WriteFile(file, b, 0600); e != nil {
				return e
			}
			defer os.Remove(file)
			ctx, cancel := context.WithTimeout(f.ctx, 110*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, f.node, f.browserScript, file)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if cmd.Run() != nil {
				return errors.New("real Chromium editor/refresh/save failed; no credential output captured")
			}
			return nil
		}); e != nil {
			return e
		}
	}
	if e = f.stage("stopped-state-process-restart", func() error {
		retainedTask := f.retainedTask
		if e := f.call("GET", base, nil, &out, 200); e != nil {
			return e
		}
		if out.Instance.State != "STOPPED" {
			return errors.New("refusing restart with a running resource")
		}
		if e := f.stop(f.daemonChild); e != nil {
			return e
		}
		if e := f.stop(f.masterChild); e != nil {
			return e
		}
		if e := startMaster(); e != nil {
			return e
		}
		if e := f.login(password); e != nil {
			return e
		}
		var e error
		f.daemonChild, e = f.start(f.daemon, "--config", f.config)
		if e != nil {
			return e
		}
		if e := f.until(func() bool {
			var nodes struct {
				Items []model.Node `json:"items"`
			}
			return f.call("GET", "/api/v1/nodes", nil, &nodes, 200) == nil && len(nodes.Items) == 1 && nodes.Items[0].ID == node && nodes.Items[0].State == "ONLINE"
		}); e != nil {
			return e
		}
		if e := f.call("GET", base, nil, &out, 200); e != nil {
			return e
		}
		if out.Instance.ID != f.instance || out.Instance.State != "STOPPED" {
			return errors.New("instance identity or stopped state changed")
		}
		var receipt struct {
			Task model.Task `json:"task"`
		}
		if retainedTask == "" || f.call("GET", "/api/v1/tasks/"+retainedTask, nil, &receipt, 200) != nil || receipt.Task.ID != retainedTask || receipt.Task.State != model.Succeeded {
			return errors.New("confirmed task receipt did not survive restart")
		}
		return f.verifyText(base, content)
	}); e != nil {
		return e
	}
	return nil
}
func (f *fixture) verifyText(base, text string) error {
	var out struct {
		Text string `json:"text"`
	}
	if e := f.call("GET", base+"/files/content?path=device.txt", nil, &out, 200); e != nil {
		return e
	}
	if out.Text != text {
		return errors.New("resource text mismatch")
	}
	return nil
}
func (f *fixture) terminalCheck(base string) error {
	t, e := f.task("POST", base+"/terminals", map[string]int{"cols": 90, "rows": 24})
	if e != nil {
		return e
	}
	var out struct {
		Session terminal.Session `json:"session"`
	}
	if e = json.Unmarshal(t.Result, &out); e != nil || out.Session.ID == "" {
		return errors.New("terminal identity missing")
	}
	f.session = out.Session.ID
	ctx, cancel := context.WithTimeout(f.ctx, 25*time.Second)
	defer cancel()
	var cursor uint64
	var output strings.Builder
	connect := func() (*protocol.Conn, error) {
		ws, _, e := websocket.Dial(ctx, strings.Replace(f.origin, "https:", "wss:", 1)+fmt.Sprintf("/api/v1/terminals/%s/stream?viewId=owned-device-view&sequence=%d", f.session, cursor), &websocket.DialOptions{HTTPClient: f.http, HTTPHeader: http.Header{"Origin": []string{f.origin}}})
		if e != nil {
			return nil, errors.New("terminal WSS unavailable")
		}
		return protocol.NewConn(ctx, ws, protocol.Options{Generation: 1, Channel: protocol.ChannelInteractive, InitialStreams: []string{f.session}})
	}
	read := func(c *protocol.Conn) (protocol.Type, error) {
		en, e := c.Read(ctx)
		if e != nil {
			return 0, e
		}
		if en.Type == protocol.TypeError {
			return 0, errors.New("terminal stream rejected")
		}
		if en.Type == protocol.TypeData {
			var ev terminal.Event
			if e = json.Unmarshal(en.Payload, &ev); e != nil {
				return 0, e
			}
			if ev.Sequence != cursor+1 {
				return 0, errors.New("terminal archive gap or replay duplicate")
			}
			cursor = ev.Sequence
			if ev.Kind == "output" {
				output.Write(ev.Data)
			}
			if e = c.Consume(ctx, f.session, en.Sequence); e != nil {
				return 0, e
			}
		}
		return en.Type, nil
	}
	c, e := connect()
	if e != nil {
		return e
	}
	defer c.Close()
	for {
		typ, e := read(c)
		if e != nil {
			return e
		}
		if typ == protocol.TypeResume {
			break
		}
	}
	input := "echo BLORA_DEVICE_WSS_MARKER > terminal-marker.txt\n"
	if runtime.GOOS == "windows" {
		input = "echo BLORA_DEVICE_WSS_MARKER>terminal-marker.txt\r"
	}
	if e = c.Send(ctx, protocol.Envelope{Type: protocol.TypeData, StreamID: f.session, Sequence: uint64(len(input)), Payload: []byte(input)}); e != nil {
		return e
	}
	for !strings.Contains(output.String(), "BLORA_DEVICE_WSS_MARKER") {
		if _, e := read(c); e != nil {
			return e
		}
	}
	if e := f.until(func() bool {
		var text struct {
			Text string `json:"text"`
		}
		return f.call("GET", base+"/files/content?path=terminal-marker.txt", nil, &text, 200) == nil && strings.TrimSpace(text.Text) == "BLORA_DEVICE_WSS_MARKER"
	}); e != nil {
		return errors.New("terminal input did not execute its owned file write")
	}
	if e = c.Close(); e != nil {
		return e
	}
	c, e = connect()
	if e != nil {
		return e
	}
	defer c.Close()
	for {
		typ, e := read(c)
		if e != nil {
			return e
		}
		if typ == protocol.TypeResume {
			break
		}
	}
	c.Close()
	_, e = f.task("POST", "/api/v1/terminals/"+f.session+"/close", map[string]any{})
	if e == nil {
		f.session = ""
	}
	return e
}
func (f *fixture) backupCheck(base, text string) error {
	var source struct {
		Version string `json:"version"`
	}
	if e := f.call("GET", base+"/files/stat?path=device.txt", nil, &source, 200); e != nil {
		return e
	}
	t, e := f.task("POST", base+"/backups", backup.CreateSpec{Path: "device.txt", Version: source.Version, Compression: "deflate", Consistency: backup.Consistency{Mode: "files"}})
	if e != nil {
		return e
	}
	var result backup.Result
	if e = json.Unmarshal(t.Result, &result); e != nil || result.Snapshot == nil {
		return errors.New("backup receipt missing")
	}
	if _, e = f.task("PUT", base+"/files/content", map[string]string{"path": "device.txt", "text": "owned changed content\n", "version": source.Version}); e != nil {
		return e
	}
	if e = f.call("GET", base+"/files/stat?path=device.txt", nil, &source, 200); e != nil {
		return e
	}
	var plan struct {
		Plan backup.RestorePlan `json:"plan"`
	}
	if e = f.call("POST", base+"/restore-plans", backup.RestoreRequest{BackupID: result.Snapshot.ID, Path: "device.txt", Version: source.Version}, &plan, 201); e != nil {
		return e
	}
	if plan.Plan.Hash == "" || plan.Plan.Request.ID == "" {
		return errors.New("restore plan identity missing")
	}
	_, e = f.task("POST", base+"/restores", backup.RestoreSpec{PlanID: plan.Plan.Request.ID, PlanHash: plan.Plan.Hash, OverwritePlanHash: plan.Plan.Hash})
	if e != nil {
		return e
	}
	return f.verifyText(base, text)
}
func (f *fixture) extensionCheck() error {
	for index, version := range []string{"1.0.0", "1.1.0"} {
		payload := []byte("export const start=async()=>{document.body.textContent='Owned device extension'}")
		sum := sha256.Sum256(payload)
		pkg := extensions.Package{Manifest: extensions.Manifest{AppID: "fixture.device", PackageVersion: version, HostAPIVersion: 1, Title: "Owned device extension", Icon: "D", Color: "#6688aa", Entrypoints: []string{"overview"}, Capabilities: []string{"window.open"}, Dependencies: map[string]string{}, WindowPolicy: "multiple", TabPolicy: &extensions.TabPolicy{Types: []string{"overview"}, Movable: true}, StateSchemaVersion: 1}, SHA256: hex.EncodeToString(sum[:]), Payload: payload}
		path := "/api/v1/extensions/install-package"
		status := 201
		if index > 0 {
			path = "/api/v1/extensions/fixture.device/upgrade"
			status = 200
		}
		wire := map[string]any{"manifest": pkg.Manifest, "sha256": pkg.SHA256, "payload": pkg.Payload}
		if e := f.call("POST", path, wire, nil, status); e != nil {
			return e
		}
	}
	return f.call("DELETE", "/api/v1/extensions/fixture.device?cleanup=true", nil, nil, 204)
}
func (f *fixture) cleanup() error {
	var cleanupErr error
	if f.http != nil && f.csrf != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		old := f.ctx
		f.ctx = ctx
		if f.session != "" {
			_, e := f.task("POST", "/api/v1/terminals/"+f.session+"/close", map[string]any{})
			cleanupErr = errors.Join(cleanupErr, e)
		}
		if f.instance != "" {
			_, e := f.task("POST", "/api/v1/instances/"+f.instance+"/actions", map[string]string{"action": "stop"})
			cleanupErr = errors.Join(cleanupErr, e)
			var state struct {
				Instance model.Instance `json:"instance"`
			}
			if f.call("GET", "/api/v1/instances/"+f.instance, nil, &state, 200) != nil || state.Instance.State != "STOPPED" {
				cleanupErr = errors.Join(cleanupErr, errors.New("owned resource exit unconfirmed"))
			}
		}
		f.ctx = old
		cancel()
	}
	all := true
	for n := len(f.children) - 1; n >= 0; n-- {
		if f.stop(f.children[n]) != nil {
			all = false
		}
	}
	if all && cleanupErr == nil && f.root != "" {
		if os.RemoveAll(f.root) != nil {
			cleanupErr = errors.New("private state removal failed")
		}
	}
	if !all || cleanupErr != nil {
		f.r.Checks = append(f.r.Checks, check{"cleanup-owned-resources", "FAIL", 0})
		f.r.Outcome = "FAILED"
		cleanupErr = errors.Join(cleanupErr, errors.New("owned cleanup unconfirmed; private state retained locally"))
	} else {
		f.r.Checks = append(f.r.Checks, check{"cleanup-owned-resources", "PASS", 0})
	}
	return cleanupErr
}

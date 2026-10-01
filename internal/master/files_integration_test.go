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
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"blora.dev/panel/internal/daemon"
	"blora.dev/panel/internal/filesystem"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type fileFixture struct {
	app                   *Server
	admin, reader         *testClient
	adminUser, readerUser model.User
	instances             []model.Instance
	roots                 []string
	store                 *storage.Store
	newClient             func() *testClient
	password              string
	restartNode           func(int)
	stopNode              func(int)
}

func newFileFixture(t *testing.T) *fileFixture {
	return newFileFixtureWithConfig(t, nil)
}

func newFileFixtureWithConfig(t *testing.T, configure func(*daemon.Config)) *fileFixture {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	base := t.TempDir()
	store, e := storage.Open(filepath.Join(base, "master.db"))
	if e != nil {
		t.Fatal(e)
	}
	password := model.ID()
	hash, e := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if e != nil {
		t.Fatal(e)
	}
	admin, e := store.CreateUser(ctx, model.User{Name: "file-admin", Admin: true}, hash)
	if e != nil {
		t.Fatal(e)
	}
	reader, e := store.CreateUser(ctx, model.User{Name: "file-reader"}, hash)
	if e != nil {
		t.Fatal(e)
	}
	var app *Server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { app.ServeHTTP(w, r) }))
	app = New(Options{Store: store, Origin: server.URL})
	ca := filepath.Join(base, "ca.crt")
	if e = os.WriteFile(ca, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0600); e != nil {
		t.Fatal(e)
	}
	newClient := func() *testClient {
		jar, _ := cookiejar.New(nil)
		client := *server.Client()
		client.Jar = jar
		client.Timeout = 20 * time.Second
		return &testClient{t: t, client: &client, base: server.URL}
	}
	f := &fileFixture{app: app, store: store, adminUser: admin, readerUser: reader, newClient: newClient, password: password}
	f.admin = newClient()
	f.admin.login(admin.Name, password)
	f.reader = newClient()
	f.reader.login(reader.Name, password)
	var nodes []*daemon.Daemon
	var nodeConfigs []daemon.Config
	var wg sync.WaitGroup
	t.Cleanup(func() {
		cancel()
		app.Close()
		for _, d := range nodes {
			d.Close()
		}
		wg.Wait()
		server.Close()
		store.Close()
	})
	for _, name := range []string{"files-node-a", "files-node-b"} {
		res := f.admin.request("POST", "/nodes/enrollments", map[string]string{"name": name}, model.ID(), 201)
		var token string
		if e = json.Unmarshal(res["token"], &token); e != nil {
			t.Fatal(e)
		}
		enrollment := filepath.Join(base, name+".enrollment")
		if e = os.WriteFile(enrollment, []byte(token), 0600); e != nil {
			t.Fatal(e)
		}
		config := daemon.Config{StateDir: filepath.Join(base, name), MasterURL: server.URL, CAFile: ca, EnrollmentFile: enrollment, AllowPGIDFallback: true}
		if configure != nil {
			configure(&config)
		}
		nodeConfigs = append(nodeConfigs, config)
		d, e := daemon.New(config)
		if e != nil {
			t.Fatal(e)
		}
		nodes = append(nodes, d)
		wg.Add(1)
		go func() { defer wg.Done(); _ = d.Run(ctx) }()
	}
	f.stopNode = func(index int) {
		id := nodes[index].NodeID()
		if err := nodes[index].Close(); err != nil {
			t.Fatal(err)
		}
		eventually(t, 8*time.Second, func() bool {
			app.mu.Lock()
			defer app.mu.Unlock()
			return app.peers[id] == nil && app.links[dataKey(id, protocol.ChannelBulk)] == nil
		})
	}
	f.restartNode = func(index int) {
		id := nodes[index].NodeID()
		before, _, err := store.Node(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if err := nodes[index].Close(); err != nil {
			t.Fatal(err)
		}
		d, err := daemon.New(nodeConfigs[index])
		if err != nil {
			t.Fatal(err)
		}
		nodes[index] = d
		wg.Add(1)
		go func() { defer wg.Done(); _ = d.Run(ctx) }()
		eventually(t, 8*time.Second, func() bool {
			n, _, err := store.Node(ctx, id)
			return err == nil && n.StartupID != "" && n.StartupID != before.StartupID
		})
		// StartupID is durable node identity, but the request path also needs
		// the freshly negotiated bulk channel. Wait for both before issuing
		// archive/file calls so a fast daemon restart cannot race the first API
		// request with channel establishment.
		eventually(t, 8*time.Second, func() bool {
			app.mu.Lock()
			defer app.mu.Unlock()
			return app.links[dataKey(id, protocol.ChannelBulk)] != nil
		})
	}
	eventually(t, 8*time.Second, func() bool {
		app.mu.Lock()
		defer app.mu.Unlock()
		return len(app.peers) == 2 && app.links[dataKey(nodes[0].NodeID(), protocol.ChannelBulk)] != nil && app.links[dataKey(nodes[1].NodeID(), protocol.ChannelBulk)] != nil
	})
	for index, d := range nodes {
		root := filepath.Join(base, []string{"resource-a", "resource-b"}[index])
		if e = os.Mkdir(root, 0750); e != nil {
			t.Fatal(e)
		}
		f.roots = append(f.roots, root)
		result := f.admin.request("POST", "/instances", map[string]any{"nodeId": d.NodeID(), "name": "files-" + d.NodeID(), "config": model.InstanceConfig{Mode: "native", Directory: root, Command: []string{"unused-file-fixture"}, StopSeconds: 1, KillSeconds: 1}}, model.ID(), 201)
		var i model.Instance
		if e = json.Unmarshal(result["instance"], &i); e != nil {
			t.Fatal(e)
		}
		f.instances = append(f.instances, i)
		for _, action := range []string{"instance.read", "file.read"} {
			f.admin.request("POST", "/grants", model.Grant{UserID: reader.ID, Resource: instanceRef(i), Action: action}, model.ID(), 200)
		}
	}
	return f
}
func fileURL(i model.Instance) string { return "/instances/" + i.ID + "/files" }
func parseFileTask(t *testing.T, result map[string]json.RawMessage) model.Task {
	t.Helper()
	var task model.Task
	if e := json.Unmarshal(result["task"], &task); e != nil {
		t.Fatal(e)
	}
	return task
}
func (f *fileFixture) awaitFile(t *testing.T, task model.Task, want model.TaskState) model.Task {
	t.Helper()
	eventually(t, 10*time.Second, func() bool {
		task = parseFileTask(t, f.admin.request("GET", "/tasks/"+task.ID, nil, "", 200))
		return task.State.Terminal()
	})
	if task.State != want {
		t.Fatalf("%s state=%s phase=%s error=%s result=%s", task.Action, task.State, task.Phase, task.Error, task.Result)
	}
	return task
}
func (f *fileFixture) save(t *testing.T, index int, p, text, v, key string) model.Task {
	t.Helper()
	return parseFileTask(t, f.admin.request("PUT", fileURL(f.instances[index])+"/content", map[string]string{"path": p, "text": text, "version": v}, key, 202))
}
func (f *fileFixture) stat(t *testing.T, index int, p string) filesystem.Entry {
	t.Helper()
	res := f.admin.request("GET", fileURL(f.instances[index])+"/stat?path="+p, nil, "", 200)
	b, _ := json.Marshal(res)
	var entry filesystem.Entry
	if e := json.Unmarshal(b, &entry); e != nil {
		t.Fatal(e)
	}
	return entry
}
func assertDisk(t *testing.T, root, p, expected string) {
	t.Helper()
	b, e := os.ReadFile(filepath.Join(root, p))
	if e != nil || string(b) != expected {
		t.Fatalf("%s bytes=%d err=%v", p, len(b), e)
	}
}

func TestFileAPIRealNodesTasksAndConditionalSaves(t *testing.T) {
	f := newFileFixture(t)
	ctx := context.Background()
	adminAccess := f.admin.request("GET", fileURL(f.instances[0])+"/access", nil, "", 200)
	readerAccess := f.reader.request("GET", fileURL(f.instances[0])+"/access", nil, "", 200)
	var adminPermissions, readerPermissions struct {
		CanRead  bool   `json:"canRead"`
		CanWrite bool   `json:"canWrite"`
		NodeID   string `json:"nodeId"`
		NodeName string `json:"nodeName"`
	}
	if json.Unmarshal(adminAccess["canRead"], &adminPermissions.CanRead) != nil || json.Unmarshal(adminAccess["canWrite"], &adminPermissions.CanWrite) != nil || json.Unmarshal(adminAccess["nodeId"], &adminPermissions.NodeID) != nil || json.Unmarshal(adminAccess["nodeName"], &adminPermissions.NodeName) != nil {
		t.Fatal("administrator file access response is incomplete")
	}
	if json.Unmarshal(readerAccess["canRead"], &readerPermissions.CanRead) != nil || json.Unmarshal(readerAccess["canWrite"], &readerPermissions.CanWrite) != nil || json.Unmarshal(readerAccess["nodeId"], &readerPermissions.NodeID) != nil || json.Unmarshal(readerAccess["nodeName"], &readerPermissions.NodeName) != nil {
		t.Fatal("reader file access response is incomplete")
	}
	if !adminPermissions.CanRead || !adminPermissions.CanWrite || !readerPermissions.CanRead || readerPermissions.CanWrite || adminPermissions.NodeID != f.instances[0].NodeID || readerPermissions.NodeID != f.instances[0].NodeID || adminPermissions.NodeName == "" || readerPermissions.NodeName != adminPermissions.NodeName {
		t.Fatalf("file access was not scoped to the current user and resource: admin=%+v reader=%+v", adminPermissions, readerPermissions)
	}
	readGrant := model.Grant{UserID: f.readerUser.ID, Resource: instanceRef(f.instances[0]), Action: "file.read"}
	f.admin.request("DELETE", "/grants", readGrant, model.ID(), 200)
	f.reader.request("GET", fileURL(f.instances[0])+"/access", nil, "", 403)
	f.admin.request("POST", "/grants", readGrant, model.ID(), 200)
	body := "\ufeff" + string(bytes.Repeat([]byte("中文\r\nconfig = true\n"), 12000))
	f.admin.request("PUT", fileURL(f.instances[0])+"/content", map[string]string{"path": "missing-key.txt", "text": "ignored", "version": filesystem.MissingVersion}, " ", 400)
	key := model.ID()
	task := f.save(t, 0, "config.toml", body, filesystem.MissingVersion, key)
	if task.State == model.Succeeded {
		t.Fatal("HTTP accepted save was presented as immediate commit")
	}
	if bytes.Contains(task.Payload, []byte("config = true")) {
		t.Fatal("file content entered the control-task payload")
	}
	task = f.awaitFile(t, task, model.Succeeded)
	var committed filesystem.Upload
	if e := json.Unmarshal(task.Result, &committed); e != nil || committed.Version != fileHash([]byte(body)) {
		t.Fatalf("%+v %v", committed, e)
	}
	assertDisk(t, f.roots[0], "config.toml", body)
	read := f.admin.request("GET", fileURL(f.instances[0])+"/content?path=config.toml", nil, "", 200)
	var text, v, encoding, newline string
	var maxBytes int64
	json.Unmarshal(read["text"], &text)
	json.Unmarshal(read["version"], &v)
	json.Unmarshal(read["encoding"], &encoding)
	json.Unmarshal(read["newline"], &newline)
	json.Unmarshal(read["maxBytes"], &maxBytes)
	if text != body || v != committed.Version || encoding != "UTF-8 BOM" || newline != "CRLF" || maxBytes != maxFileText {
		t.Fatalf("bulk text/format roundtrip mismatch: encoding=%q newline=%q maxBytes=%d version=%q", encoding, newline, maxBytes, v)
	}
	if again := f.save(t, 0, "config.toml", body, filesystem.MissingVersion, key); again.ID != task.ID {
		t.Fatal("retry duplicated save")
	}
	f.admin.request("PUT", fileURL(f.instances[0])+"/content", map[string]string{"path": "config.toml", "text": "changed-body", "version": v}, key, 409)
	f.reader.request("PUT", fileURL(f.instances[0])+"/content", map[string]string{"path": "config.toml", "text": "denied", "version": v}, model.ID(), 403)
	f.reader.request("GET", fileURL(f.instances[0])+"/content?path=../outside", nil, "", 400)
	if e := os.WriteFile(filepath.Join(f.roots[0], "config.toml"), []byte("external"), 0640); e != nil {
		t.Fatal(e)
	}
	failed := f.save(t, 0, "config.toml", "stale", v, model.ID())
	failed = f.awaitFile(t, failed, model.Failed)
	if failed.Phase != "file_conflict" {
		t.Fatalf("phase %s", failed.Phase)
	}
	assertDisk(t, f.roots[0], "config.toml", "external")
	// Both nodes use independent authoritative instance roots.
	f.awaitFile(t, f.save(t, 1, "config.toml", "second-node", filesystem.MissingVersion, model.ID()), model.Succeeded)
	assertDisk(t, f.roots[1], "config.toml", "second-node")
	assertDisk(t, f.roots[0], "config.toml", "external")
	list := f.admin.request("GET", fileURL(f.instances[0])+"?path=.&limit=1", nil, "", 200)
	var rows []map[string]any
	if e := json.Unmarshal(list["items"], &rows); e != nil || len(rows) != 1 || rows[0]["isDir"] != false {
		t.Fatalf("%s %v", list["items"], e)
	}
	for name, text := range map[string]string{"match-a": "1", "match-z": "12345"} {
		if e := os.WriteFile(filepath.Join(f.roots[0], name), []byte(text), 0640); e != nil {
			t.Fatal(e)
		}
	}
	sorted := f.admin.request("GET", fileURL(f.instances[0])+"?path=.&search=match&sort=size&order=desc&limit=1", nil, "", 200)
	if e := json.Unmarshal(sorted["items"], &rows); e != nil || len(rows) != 1 || rows[0]["name"] != "match-z" || string(sorted["total"]) != "2" {
		t.Fatalf("full-directory filtering/sorting: %s %v", sorted["items"], e)
	}
	// A real browser download receives only node-verified chunks.
	req, _ := http.NewRequestWithContext(ctx, "GET", f.admin.base+"/api/v1"+fileURL(f.instances[1])+"/download?path=config.toml", nil)
	resp, e := f.admin.client.Do(req)
	if e != nil {
		t.Fatal(e)
	}
	download, e := io.ReadAll(resp.Body)
	resp.Body.Close()
	if e != nil || resp.StatusCode != 200 || string(download) != "second-node" || resp.Header.Get("Content-Disposition") == "" {
		t.Fatalf("download status=%d bytes=%q err=%v", resp.StatusCode, download, e)
	}
	largeBytes := bytes.Repeat([]byte{'L'}, maxFileText+1)
	atLimitBytes := bytes.Repeat([]byte{'M'}, maxFileText)
	binaryBytes := []byte{0x00, 0x01, 0x7f, 0x80, 0xff}
	if e = os.WriteFile(filepath.Join(f.roots[0], "at-limit.txt"), atLimitBytes, 0640); e != nil {
		t.Fatal(e)
	}
	limitRequest, e := http.NewRequestWithContext(ctx, "GET", f.admin.base+"/api/v1"+fileURL(f.instances[0])+"/content?path=at-limit.txt", nil)
	if e != nil {
		t.Fatal(e)
	}
	limitResponse, e := f.admin.client.Do(limitRequest)
	if e != nil {
		t.Fatal(e)
	}
	limitBody, readErr := io.ReadAll(limitResponse.Body)
	limitResponse.Body.Close()
	var limitContent struct {
		Text     string `json:"text"`
		MaxBytes int64  `json:"maxBytes"`
	}
	if readErr != nil || limitResponse.StatusCode != http.StatusOK || json.Unmarshal(limitBody, &limitContent) != nil || len(limitContent.Text) != maxFileText || limitContent.Text != string(atLimitBytes) || limitContent.MaxBytes != maxFileText {
		t.Fatalf("at-limit content status=%d text=%d limit=%d error=%v", limitResponse.StatusCode, len(limitContent.Text), limitContent.MaxBytes, readErr)
	}
	if e = os.WriteFile(filepath.Join(f.roots[0], "large.txt"), largeBytes, 0640); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(f.roots[0], "image.bin"), binaryBytes, 0640); e != nil {
		t.Fatal(e)
	}
	for _, boundary := range []struct {
		path, code string
		status     int
	}{{"large.txt", "TEXT_LIMIT", 413}, {"image.bin", "UNSUPPORTED_ENCODING", 415}} {
		result := f.admin.request("GET", fileURL(f.instances[0])+"/content?path="+boundary.path, nil, "", boundary.status)
		var apiError struct {
			Code string `json:"code"`
		}
		if e = json.Unmarshal(result["error"], &apiError); e != nil || apiError.Code != boundary.code {
			t.Fatalf("%s boundary error=%+v decode=%v", boundary.path, apiError, e)
		}
	}
	preview := f.admin.request("GET", fileURL(f.instances[0])+"/preview?path=large.txt&limit=1024", nil, "", 200)
	var previewPage struct {
		Text       string `json:"text"`
		Version    string `json:"version"`
		Offset     int64  `json:"offset"`
		NextOffset int64  `json:"nextOffset"`
		Total      int64  `json:"total"`
		HasMore    bool   `json:"hasMore"`
	}
	if e = json.Unmarshal(preview["text"], &previewPage.Text); e != nil {
		t.Fatal(e)
	}
	for key, target := range map[string]any{"version": &previewPage.Version, "offset": &previewPage.Offset, "nextOffset": &previewPage.NextOffset, "total": &previewPage.Total, "hasMore": &previewPage.HasMore} {
		if e = json.Unmarshal(preview[key], target); e != nil {
			t.Fatalf("decode preview %s: %v", key, e)
		}
	}
	if previewPage.Text != strings.Repeat("L", 1024) || previewPage.Offset != 0 || previewPage.NextOffset != 1024 || previewPage.Total != int64(len(largeBytes)) || !previewPage.HasMore || previewPage.Version == "" {
		t.Fatalf("unexpected first preview: %+v", previewPage)
	}
	next := f.admin.request("GET", fileURL(f.instances[0])+"/preview?path=large.txt&version="+previewPage.Version+"&offset="+strconv.FormatInt(previewPage.NextOffset, 10)+"&limit=1024", nil, "", 200)
	var nextText string
	if e = json.Unmarshal(next["text"], &nextText); e != nil || nextText != strings.Repeat("L", 1024) {
		t.Fatalf("unexpected second preview text=%q error=%v", nextText, e)
	}
	f.admin.request("GET", fileURL(f.instances[0])+"/preview?path=image.bin", nil, "", 415)
	for path, expected := range map[string][]byte{"large.txt": largeBytes, "image.bin": binaryBytes} {
		req, e := http.NewRequestWithContext(ctx, "GET", f.admin.base+"/api/v1"+fileURL(f.instances[0])+"/download?path="+path, nil)
		if e != nil {
			t.Fatal(e)
		}
		response, e := f.admin.client.Do(req)
		if e != nil {
			t.Fatal(e)
		}
		data, readErr := io.ReadAll(response.Body)
		response.Body.Close()
		if readErr != nil || response.StatusCode != http.StatusOK || !bytes.Equal(data, expected) || response.ContentLength != int64(len(expected)) {
			t.Fatalf("download %s status=%d bytes=%d expected=%d content-length=%d error=%v", path, response.StatusCode, len(data), len(expected), response.ContentLength, readErr)
		}
	}
	t.Log("real TLS Master plus two Daemons: file roots, bulk text, durable commit, retry, conflict, read/write authorization, bounded text rejection and exact large/binary streaming downloads passed")
}

func TestFileActionsArchiveAndRecycleThroughDurableTasks(t *testing.T) {
	f := newFileFixture(t)
	base := fileURL(f.instances[0])
	f.admin.request("POST", base+"/actions", map[string]any{"action": "mkdir", "path": "missing-key"}, strings.Repeat("x", 129), 400)
	action := func(in map[string]any) model.Task {
		return f.awaitFile(t, parseFileTask(t, f.admin.request("POST", base+"/actions", in, model.ID(), 202)), model.Succeeded)
	}
	action(map[string]any{"action": "mkdir", "path": "dir"})
	f.awaitFile(t, f.save(t, 0, "dir/a", "one", filesystem.MissingVersion, model.ID()), model.Succeeded)
	action(map[string]any{"action": "copy", "path": "dir", "target": "copied", "version": f.stat(t, 0, "dir").Version, "targetVersion": filesystem.MissingVersion})
	assertDisk(t, f.roots[0], "copied/a", "one")
	action(map[string]any{"action": "move", "path": "copied", "target": "moved", "version": f.stat(t, 0, "copied").Version, "targetVersion": filesystem.MissingVersion})
	action(map[string]any{"action": "compress", "path": "moved", "target": "archive.zip", "version": f.stat(t, 0, "moved").Version, "targetVersion": filesystem.MissingVersion})
	action(map[string]any{"action": "extract", "path": "archive.zip", "target": "expanded", "version": f.stat(t, 0, "archive.zip").Version, "targetVersion": filesystem.MissingVersion})
	assertDisk(t, f.roots[0], "expanded/moved/a", "one")
	deleted := action(map[string]any{"action": "delete", "path": "moved", "version": f.stat(t, 0, "moved").Version})
	var trash filesystem.TrashItem
	if e := json.Unmarshal(deleted.Result, &trash); e != nil {
		t.Fatal(e)
	}
	trashList := f.admin.request("GET", base+"/trash", nil, "", 200)
	var recycled []filesystem.TrashItem
	if e := json.Unmarshal(trashList["items"], &recycled); e != nil || len(recycled) != 1 || recycled[0].ID != trash.ID {
		t.Fatalf("trash API %s %v", trashList["items"], e)
	}
	action(map[string]any{"action": "restore", "trashId": trash.ID, "target": "restored", "targetVersion": filesystem.MissingVersion})
	assertDisk(t, f.roots[0], "restored/a", "one")
	f.admin.request("POST", base+"/actions", map[string]any{"action": "delete", "path": "restored", "version": f.stat(t, 0, "restored").Version, "config": map[string]string{"directory": "/"}}, model.ID(), 400)
	t.Log("node worker tasks performed mkdir/copy/move/ZIP/extract/recycle/restore against actual instance files")
}

func TestBrowserUploadCheckpointReconnectAndCancellation(t *testing.T) {
	f := newFileFixture(t)
	base := fileURL(f.instances[0])
	data := bytes.Repeat([]byte("source"), 23000)
	key := model.ID()
	chunkKey := func(offset int64) string { return key + ":chunk:" + strconv.FormatInt(offset, 10) }
	completeKey := key + ":complete"
	begin := map[string]any{"path": "uploaded.bin", "total": len(data), "hash": fileHash(data), "version": filesystem.MissingVersion, "sourceName": "local.bin", "sourceModified": 123, "sourceFingerprint": fileHash(data)}
	res := f.admin.request("POST", base+"/uploads", begin, key, 202)
	task := parseFileTask(t, res)
	var checkpoint filesystem.Upload
	if e := json.Unmarshal(res["upload"], &checkpoint); e != nil {
		t.Fatal(e)
	}
	chunk := data[:fileChunkSize]
	f.admin.request("PUT", base+"/uploads/"+task.ID+"/chunks", map[string]any{"offset": 0, "data": chunk, "hash": fileHash(chunk)}, "", 400)
	f.admin.request("POST", base+"/uploads/"+task.ID+"/complete", map[string]any{}, "", 400)
	res = f.admin.request("PUT", base+"/uploads/"+task.ID+"/chunks", map[string]any{"offset": 0, "data": chunk, "hash": fileHash(chunk)}, chunkKey(0), 200)
	if e := json.Unmarshal(res["upload"], &checkpoint); e != nil || checkpoint.Offset != fileChunkSize {
		t.Fatalf("%+v %v", checkpoint, e)
	}
	f.admin.request("POST", base+"/uploads/"+task.ID+"/complete", map[string]any{}, completeKey, 409)
	// Close the management connection after one acknowledged chunk, then
	// re-establish a new browser session and resume using persisted identities.
	f.app.mu.Lock()
	old := f.app.peers[f.instances[0].NodeID]
	f.app.mu.Unlock()
	old.conn.Close()
	eventually(t, 8*time.Second, func() bool {
		f.app.mu.Lock()
		defer f.app.mu.Unlock()
		p := f.app.peers[f.instances[0].NodeID]
		return p != nil && p.generation > old.generation && f.app.links[dataKey(f.instances[0].NodeID, protocol.ChannelBulk)] != nil
	})
	stored, e := f.store.Task(context.Background(), task.ID)
	if e != nil || stored.State != model.WaitingClient {
		t.Fatalf("disconnected partial upload state=%s err=%v", stored.State, e)
	}
	if _, e = os.Stat(filepath.Join(f.roots[0], "uploaded.bin")); !os.IsNotExist(e) {
		t.Fatal("partial upload published as completed file")
	}
	resumed := f.newClient()
	resumed.login(f.adminUser.Name, f.password)
	res = resumed.request("GET", base+"/uploads/"+task.ID, nil, "", 200)
	if e = json.Unmarshal(res["upload"], &checkpoint); e != nil || checkpoint.Offset != fileChunkSize {
		t.Fatalf("%+v %v", checkpoint, e)
	}
	changed := map[string]any{}
	for k, v := range begin {
		changed[k] = v
	}
	changed["sourceFingerprint"] = "different-source"
	resumed.request("POST", base+"/uploads", changed, key, 409)
	f.reader.request("GET", base+"/uploads/"+task.ID, nil, "", 403)
	for off := checkpoint.Offset; off < int64(len(data)); {
		part := data[off:min(off+fileChunkSize, int64(len(data)))]
		res = resumed.request("PUT", base+"/uploads/"+task.ID+"/chunks", map[string]any{"offset": off, "data": part, "hash": fileHash(part)}, chunkKey(off), 200)
		if e = json.Unmarshal(res["upload"], &checkpoint); e != nil {
			t.Fatal(e)
		}
		off = checkpoint.Offset
	}
	task = parseFileTask(t, resumed.request("POST", base+"/uploads/"+task.ID+"/complete", map[string]any{}, completeKey, 202))
	f.awaitFile(t, task, model.Succeeded)
	assertDisk(t, f.roots[0], "uploaded.bin", string(data))
	begin["path"] = "cancelled.bin"
	res = resumed.request("POST", base+"/uploads", begin, model.ID(), 202)
	cancelled := parseFileTask(t, res)
	resumed.request("PUT", base+"/uploads/"+cancelled.ID+"/chunks", map[string]any{"offset": 0, "data": chunk, "hash": fileHash(chunk)}, key+":cancel-chunk:0", 200)
	cancelled = parseFileTask(t, resumed.request("DELETE", base+"/uploads/"+cancelled.ID, nil, model.ID(), 202))
	if cancelled.State != model.Cancelled {
		t.Fatalf("cancel state %s", cancelled.State)
	}
	if _, e = os.Stat(filepath.Join(f.roots[0], "cancelled.bin")); !os.IsNotExist(e) {
		t.Fatal("cancelled destination created")
	}
	assertDisk(t, f.roots[0], "uploaded.bin", string(data))
	t.Log("browser upload acknowledged node checkpoints; reconnect retained WAITING_CLIENT; same source resumed; cancellation cleaned only its own staging")
}

func TestDownloadRevocationTerminatesExistingStream(t *testing.T) {
	f := newFileFixture(t)
	base := fileURL(f.instances[0])
	file, e := os.Create(filepath.Join(f.roots[0], "large.bin"))
	if e != nil {
		t.Fatal(e)
	}
	if e = file.Truncate(8 << 20); e != nil {
		t.Fatal(e)
	}
	file.Close()
	resp, e := f.reader.client.Get(f.reader.base + "/api/v1" + base + "/download?path=large.bin")
	if e != nil {
		t.Fatal(e)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	first := make([]byte, 1024)
	n, e := io.ReadFull(resp.Body, first)
	if e != nil {
		t.Fatal(e)
	}
	f.admin.request("DELETE", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: instanceRef(f.instances[0]), Action: "file.read"}, model.ID(), 200)
	rest, readErr := io.Copy(io.Discard, resp.Body)
	if readErr == nil || int64(n)+rest >= 8<<20 {
		t.Fatalf("revoked stream completed: %d bytes, err=%v", int64(n)+rest, readErr)
	}
	f.reader.request("GET", base+"/stat?path=large.bin", nil, "", 403)
	f.admin.request("GET", base+"/stat?path=large.bin", nil, "", 200)
	t.Log("grant revocation cut an active real HTTPS file stream and refused subsequent reads while admin access remained available")
}

func TestCancelAcceptedUploadBeforeNodePreparation(t *testing.T) {
	f := newFileFixture(t)
	i := f.instances[0]
	key := model.ID()
	spec := newUploadSpec(f.adminUser, key, "never-started", filesystem.MissingVersion, fileHash([]byte("x")), "local", "fingerprint", 1, 0)
	payload, _ := json.Marshal(uploadPayload(i, spec))
	task, _, e := f.store.Accept(context.Background(), model.Task{ActorID: f.adminUser.ID, RequestID: key, Resource: instanceRef(i), Action: "file.upload", Payload: payload, State: model.WaitingClient})
	if e != nil {
		t.Fatal(e)
	}
	// Simulate a lost cancel response before bulk cleanup. Keep the real
	// dispatcher running through several ticks: a client upload has no
	// daemon task receipt yet, so generic cancellation must not be sent.
	cancelKey := model.ID()
	task, e = f.store.RequestCancel(context.Background(), task.ID, cancelKey)
	if e != nil {
		t.Fatal(e)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		current, err := f.store.Task(context.Background(), task.ID)
		if err != nil || current.State != model.CancelRequested || !current.DispatchedAt.IsZero() {
			t.Fatalf("client cleanup escaped to generic dispatcher: state=%s dispatched=%v err=%v", current.State, current.DispatchedAt, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	result := f.admin.request("DELETE", fileURL(i)+"/uploads/"+task.ID, nil, cancelKey, 202)
	task = parseFileTask(t, result)
	if task.State != model.Cancelled || task.CancellationRequestID != cancelKey || !task.DispatchedAt.IsZero() {
		t.Fatalf("unprepared cancellation %+v", task)
	}
	if _, e = os.Stat(filepath.Join(f.roots[0], "never-started")); !os.IsNotExist(e) {
		t.Fatalf("unexpected destination: %v", e)
	}
}

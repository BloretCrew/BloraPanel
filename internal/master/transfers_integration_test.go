package master

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"blora.dev/panel/internal/daemon"
	"blora.dev/panel/internal/filesystem"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

func startTransfer(t *testing.T, f *fileFixture, src, dest string, move bool, key string) model.Task {
	t.Helper()
	source := f.stat(t, 0, src)
	request := model.TransferRequest{Source: model.TransferEndpoint{InstanceID: f.instances[0].ID, Path: src, Version: source.Version}, Target: model.TransferEndpoint{InstanceID: f.instances[1].ID, Path: dest, Version: filesystem.MissingVersion}, Move: move}
	return parseFileTask(t, f.admin.request("POST", "/transfers", request, key, 202))
}
func transferSnapshot(t *testing.T, f *fileFixture, id string) (model.Task, model.TransferResult) {
	t.Helper()
	response := f.admin.request("GET", "/transfers/"+id, nil, "", 200)
	var p model.TransferResult
	if err := json.Unmarshal(response["transfer"], &p); err != nil {
		t.Fatal(err)
	}
	return parseFileTask(t, response), p
}
func awaitTransfer(t *testing.T, f *fileFixture, task model.Task, want model.TaskState) model.TransferResult {
	t.Helper()
	var p model.TransferResult
	eventually(t, 40*time.Second, func() bool { task, p = transferSnapshot(t, f, task.ID); return task.State.Terminal() })
	if task.State != want {
		t.Fatalf("transfer state=%s phase=%s error=%s result=%+v", task.State, task.Phase, task.Error, p)
	}
	return p
}
func transferWrite(t *testing.T, root, name string, body []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filepath.Join(root, name)), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, name), body, 0640); err != nil {
		t.Fatal(err)
	}
}

func TestTransferRealDirectoryCopyMoveAndMetadata(t *testing.T) {
	f := newFileFixture(t)
	transferWrite(t, f.roots[0], "tree/config.toml", []byte("source config\n"))
	transferWrite(t, f.roots[0], "tree/nested/binary.dat", bytes.Repeat([]byte{0, 0xff, 3, 4}, 40000))
	if err := os.Mkdir(filepath.Join(f.roots[0], "tree/empty"), 0710); err != nil {
		t.Fatal(err)
	}
	stamp := time.Unix(1700000000, 123456000)
	for _, name := range []string{"tree", "tree/config.toml", "tree/nested", "tree/nested/binary.dat", "tree/empty"} {
		if err := os.Chtimes(filepath.Join(f.roots[0], name), stamp, stamp); err != nil {
			t.Fatal(err)
		}
	}
	key := model.ID()
	task := startTransfer(t, f, "tree", "copied", false, key)
	active := f.admin.request("GET", "/transfers", nil, "", 200)
	var activeTasks []model.Task
	if err := json.Unmarshal(active["items"], &activeTasks); err != nil || len(activeTasks) != 1 || activeTasks[0].ID != task.ID {
		t.Fatalf("active transfer listing omitted accepted transfer: %s %v", active["items"], err)
	}
	if again := startTransfer(t, f, "tree", "copied", false, key); again.ID != task.ID {
		t.Fatal("retry duplicated parent")
	}
	p := awaitTransfer(t, f, task, model.Succeeded)
	if p.Completed != 5 || !p.DestinationVerified || p.SourceDeleted || p.Partial {
		t.Fatalf("%+v", p)
	}
	for _, name := range []string{"", "config.toml", "nested", "nested/binary.dat", "empty"} {
		source, err := os.Stat(filepath.Join(f.roots[0], "tree", name))
		if err != nil {
			t.Fatal(err)
		}
		target, err := os.Stat(filepath.Join(f.roots[1], "copied", name))
		if err != nil {
			t.Fatal(err)
		}
		if source.Mode().Perm() != target.Mode().Perm() || !source.ModTime().Equal(target.ModTime()) {
			t.Fatalf("metadata %s: source mode=%v modified=%v target mode=%v modified=%v", name, source.Mode(), source.ModTime(), target.Mode(), target.ModTime())
		}
		if !source.IsDir() {
			a, _ := os.ReadFile(filepath.Join(f.roots[0], "tree", name))
			b, _ := os.ReadFile(filepath.Join(f.roots[1], "copied", name))
			if !bytes.Equal(a, b) {
				t.Fatal("content mismatch", name)
			}
		}
	}
	entries := f.admin.request("GET", "/transfers/"+task.ID+"/entries?limit=2", nil, "", 200)
	var next int
	json.Unmarshal(entries["nextOffset"], &next)
	if next != 2 {
		t.Fatalf("manifest pagination %s", entries["nextOffset"])
	}
	move := startTransfer(t, f, "tree", "moved", true, model.ID())
	p = awaitTransfer(t, f, move, model.Succeeded)
	if !p.SourceDeleted || !p.DestinationVerified {
		t.Fatalf("%+v", p)
	}
	if _, err := os.Stat(filepath.Join(f.roots[0], "tree")); !os.IsNotExist(err) {
		t.Fatalf("move source remains: %v", err)
	}
	assertDisk(t, f.roots[1], "moved/config.toml", "source config\n")
	// Source deletion itself remains a durable, independently inspectable task.
	var progress transferProgress
	if _, err := f.store.Record(context.Background(), transferProgressNS, move.ID, &progress); err != nil {
		t.Fatal(err)
	}
	deleted, err := f.store.Task(context.Background(), progress.DeleteID)
	if err != nil || deleted.Action != "file.delete" || deleted.State != model.Succeeded {
		t.Fatalf("%+v %v", deleted, err)
	}
}

func TestTransferSourceChangesAndCancellationKeepPartialResults(t *testing.T) {
	f := newFileFixture(t)
	body := bytes.Repeat([]byte("abcdefgh"), 1<<20)
	transferWrite(t, f.roots[0], "changing.bin", body)
	task := startTransfer(t, f, "changing.bin", "changing-copy.bin", true, model.ID())
	eventually(t, 10*time.Second, func() bool { _, p := transferSnapshot(t, f, task.ID); return p.CurrentOffset > 0 })
	transferWrite(t, f.roots[0], "changing.bin", []byte("new source generation"))
	p := awaitTransfer(t, f, task, model.Failed)
	if p.SourceDeleted || p.CleanupPending {
		t.Fatalf("%+v", p)
	}
	assertDisk(t, f.roots[0], "changing.bin", "new source generation")
	if _, err := os.Stat(filepath.Join(f.roots[1], "changing-copy.bin")); !os.IsNotExist(err) {
		t.Fatalf("mixed upload became destination: %v", err)
	}

	transferWrite(t, f.roots[0], "partial/a.txt", []byte("committed first"))
	transferWrite(t, f.roots[0], "partial/z.bin", body)
	task = startTransfer(t, f, "partial", "partial-copy", true, model.ID())
	eventually(t, 15*time.Second, func() bool { _, p := transferSnapshot(t, f, task.ID); return p.Completed >= 2 && p.CurrentOffset > 0 })
	f.admin.request("POST", "/tasks/"+task.ID+"/cancel", nil, "", 400)
	f.admin.request("POST", "/tasks/"+task.ID+"/cancel", nil, model.ID(), 202)
	p = awaitTransfer(t, f, task, model.Cancelled)
	if p.Completed < 2 || !p.Partial || p.SourceDeleted || p.CleanupPending {
		t.Fatalf("%+v", p)
	}
	assertDisk(t, f.roots[1], "partial-copy/a.txt", "committed first")
	assertDisk(t, f.roots[0], "partial/a.txt", "committed first")
	entries, err := os.ReadDir(filepath.Join(f.roots[1], "partial-copy"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "a.txt" {
		t.Fatalf("cancel did not acknowledge its own staging cleanup: %v", entries)
	}
}

func TestTransferChangedReadPrefixCannotPublish(t *testing.T) {
	f := newFileFixture(t)
	body := bytes.Repeat([]byte("proof-source-bytes"), 500000)
	transferWrite(t, f.roots[0], "source.bin", body)
	baseline := f.stat(t, 0, "source.bin")
	task := startTransfer(t, f, "source.bin", "unpublished.bin", false, model.ID())
	eventually(t, 10*time.Second, func() bool { _, p := transferSnapshot(t, f, task.ID); return p.CurrentOffset >= fileChunkSize })
	// Pause only the real coordinator between acknowledged checkpoints. Daemon
	// files, proof cache, upload state and TLS RPCs remain real and running.
	f.app.closeTransfers()
	_, before := transferSnapshot(t, f, task.ID)
	if before.CurrentOffset >= int64(len(body)) {
		t.Fatal("test missed the mid-transfer boundary")
	}
	p := filepath.Join(f.roots[0], "source.bin")
	file, err := os.OpenFile(p, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = file.WriteAt([]byte("X"), 0)
	closeErr := file.Close()
	if err != nil || closeErr != nil {
		t.Fatal(err, closeErr)
	}
	if err = os.Chtimes(p, baseline.Modified, baseline.Modified); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	c := &transferCoordinator{s: f.app, ctx: ctx, cancel: cancel, active: map[string]context.CancelFunc{}}
	f.app.transferJobs = c
	c.wg.Add(1)
	go c.loop()
	result := awaitTransfer(t, f, task, model.Failed)
	finished, _ := transferSnapshot(t, f, task.ID)
	if !strings.Contains(finished.Error, "发布目标前改变") {
		t.Fatalf("did not exercise final full source validation: %s", finished.Error)
	}
	if result.DestinationVerified || result.SourceDeleted || result.CleanupPending {
		t.Fatalf("unexpected publication: %+v", result)
	}
	if _, err = os.Stat(filepath.Join(f.roots[1], "unpublished.bin")); !os.IsNotExist(err) {
		t.Fatalf("changed source published: %v", err)
	}
}

func TestTransferSourceDaemonRestartRebuildsProof(t *testing.T) {
	f := newFileFixture(t)
	body := bytes.Repeat([]byte("rebuild-source-proof"), 400000)
	transferWrite(t, f.roots[0], "source.bin", body)
	task := startTransfer(t, f, "source.bin", "rebuilt.bin", false, model.ID())
	eventually(t, 10*time.Second, func() bool { _, p := transferSnapshot(t, f, task.ID); return p.CurrentOffset >= fileChunkSize })
	f.restartNode(0)
	result := awaitTransfer(t, f, task, model.Succeeded)
	actual, err := os.ReadFile(filepath.Join(f.roots[1], "rebuilt.bin"))
	if err != nil || !bytes.Equal(actual, body) || !result.DestinationVerified {
		t.Fatalf("cold proof recovery: %v %+v", err, result)
	}
}

func TestTransferAuthorizationManifestBudgetAndTargetConflict(t *testing.T) {
	f := newFileFixture(t)
	transferWrite(t, f.roots[0], "source.txt", []byte("original"))
	version := f.stat(t, 0, "source.txt").Version
	request := model.TransferRequest{Source: model.TransferEndpoint{InstanceID: f.instances[0].ID, Path: "source.txt", Version: version}, Target: model.TransferEndpoint{InstanceID: f.instances[1].ID, Path: "target.txt", Version: filesystem.MissingVersion}}
	f.reader.request("POST", "/transfers", request, model.ID(), 403)
	transferWrite(t, f.roots[1], "target.txt", []byte("existing target"))
	task := parseFileTask(t, f.admin.request("POST", "/transfers", request, model.ID(), 202))
	p := awaitTransfer(t, f, task, model.Failed)
	if p.Completed != 0 {
		t.Fatalf("preflight had side effects: %+v", p)
	}
	assertDisk(t, f.roots[1], "target.txt", "existing target")
	if err := os.Mkdir(filepath.Join(f.roots[0], "many"), 0750); err != nil {
		t.Fatal(err)
	}
	for n := 0; n < transferMaxEntries; n++ {
		transferWrite(t, f.roots[0], "many/"+strconv.Itoa(n), nil)
	}
	task = startTransfer(t, f, "many", "many-copy", false, model.ID())
	p = awaitTransfer(t, f, task, model.Failed)
	if p.Completed != 0 {
		t.Fatal("over-budget manifest created destination")
	}
	if _, err := os.Stat(filepath.Join(f.roots[1], "many-copy")); !os.IsNotExist(err) {
		t.Fatalf("budget preflight mutation: %v", err)
	}
}

// This fixture restarts the real Master and closes/reopens its SQLite database.
// Nodes retain their own durable stores and reconnect to the unchanged TLS URL.
func newRestartTransferFixture(t *testing.T) (*fileFixture, func()) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	base := t.TempDir()
	database := filepath.Join(base, "master.db")
	db, err := storage.Open(database)
	if err != nil {
		t.Fatal(err)
	}
	password := model.ID()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	u, err := db.CreateUser(ctx, model.User{Name: "transfer-admin", Admin: true}, hash)
	if err != nil {
		t.Fatal(err)
	}
	var active atomic.Pointer[Server]
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app := active.Load(); app != nil {
			app.ServeHTTP(w, r)
		} else {
			http.Error(w, "Master restarting", 503)
		}
	}))
	app := New(Options{Store: db, Origin: server.URL})
	active.Store(app)
	ca := filepath.Join(base, "ca.crt")
	if err := os.WriteFile(ca, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0600); err != nil {
		t.Fatal(err)
	}
	jar, _ := cookiejar.New(nil)
	client := *server.Client()
	client.Jar = jar
	client.Timeout = 20 * time.Second
	f := &fileFixture{app: app, store: db, adminUser: u, admin: &testClient{t: t, client: &client, base: server.URL}}
	f.admin.login(u.Name, password)
	var nodes []*daemon.Daemon
	var wg sync.WaitGroup
	t.Cleanup(func() {
		cancel()
		active.Store(nil)
		app.Close()
		for _, d := range nodes {
			d.Close()
		}
		wg.Wait()
		server.Close()
		db.Close()
	})
	for n := 0; n < 2; n++ {
		result := f.admin.request("POST", "/nodes/enrollments", map[string]string{"name": "transfer-node-" + strconv.Itoa(n)}, model.ID(), 201)
		var token string
		json.Unmarshal(result["token"], &token)
		enrollment := filepath.Join(base, "enrollment-"+strconv.Itoa(n))
		if err = os.WriteFile(enrollment, []byte(token), 0600); err != nil {
			t.Fatal(err)
		}
		d, err := daemon.New(daemon.Config{StateDir: filepath.Join(base, "daemon-"+strconv.Itoa(n)), MasterURL: server.URL, CAFile: ca, EnrollmentFile: enrollment, AllowPGIDFallback: true})
		if err != nil {
			t.Fatal(err)
		}
		nodes = append(nodes, d)
		wg.Add(1)
		go func() { defer wg.Done(); _ = d.Run(ctx) }()
	}
	waitNodes := func() {
		eventually(t, 15*time.Second, func() bool {
			app.mu.Lock()
			defer app.mu.Unlock()
			return len(app.peers) == 2 && app.links[dataKey(nodes[0].NodeID(), protocol.ChannelBulk)] != nil && app.links[dataKey(nodes[1].NodeID(), protocol.ChannelBulk)] != nil
		})
	}
	waitNodes()
	for n, d := range nodes {
		root := filepath.Join(base, "root-"+strconv.Itoa(n))
		if err = os.Mkdir(root, 0750); err != nil {
			t.Fatal(err)
		}
		f.roots = append(f.roots, root)
		result := f.admin.request("POST", "/instances", map[string]any{"nodeId": d.NodeID(), "name": "transfer-resource", "config": model.InstanceConfig{Mode: "native", Directory: root, Command: []string{"unused-transfer-fixture"}, StopSeconds: 1, KillSeconds: 1}}, model.ID(), 201)
		var i model.Instance
		if err = json.Unmarshal(result["instance"], &i); err != nil {
			t.Fatal(err)
		}
		f.instances = append(f.instances, i)
	}
	restart := func() {
		active.Store(nil)
		app.Close()
		if err = db.Close(); err != nil {
			t.Fatal(err)
		}
		db, err = storage.Open(database)
		if err != nil {
			t.Fatal(err)
		}
		app = New(Options{Store: db, Origin: server.URL})
		f.app, f.store = app, db
		active.Store(app)
		waitNodes()
	}
	return f, restart
}

func TestTransferMasterRestartUsesNodeAcknowledgedCheckpoint(t *testing.T) {
	f, restart := newRestartTransferFixture(t)
	body := bytes.Repeat([]byte("persistent transfer bytes\x00"), 220000)
	transferWrite(t, f.roots[0], "restart.bin", body)
	task := startTransfer(t, f, "restart.bin", "restart-copy.bin", false, model.ID())
	var before model.TransferResult
	eventually(t, 12*time.Second, func() bool {
		_, before = transferSnapshot(t, f, task.ID)
		return before.CurrentOffset >= 2*fileChunkSize && before.CurrentOffset < int64(len(body))
	})
	restart()
	p := awaitTransfer(t, f, task, model.Succeeded)
	if p.CommittedBytes != int64(len(body)) || !p.DestinationVerified {
		t.Fatalf("%+v", p)
	}
	actual, err := os.ReadFile(filepath.Join(f.roots[1], "restart-copy.bin"))
	if err != nil || !bytes.Equal(actual, body) {
		t.Fatalf("restart content bytes=%d err=%v", len(actual), err)
	}
	var m transferManifest
	if _, err = f.store.Record(context.Background(), transferManifestNS, task.ID, &m); err != nil || len(m.Entries) != 1 {
		t.Fatalf("manifest %v", err)
	}
	children, err := f.store.Tasks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, child := range children {
		if child.RequestID == transferChildKey(task, 0, 0, "file.upload") {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("restart duplicated durable commit: %d", count)
	}
}

func TestTransferDisconnectAndMidstreamPermissionRevocation(t *testing.T) {
	f := newFileFixture(t)
	body := bytes.Repeat([]byte("bounded relay bytes"), 350000)
	transferWrite(t, f.roots[0], "disconnect.bin", body)
	task := startTransfer(t, f, "disconnect.bin", "after-reconnect.bin", false, model.ID())
	eventually(t, 10*time.Second, func() bool {
		current, progress := transferSnapshot(t, f, task.ID)
		return current.State == model.Running && progress.CurrentOffset > 0 && progress.CurrentOffset < int64(len(body))
	})
	f.app.mu.Lock()
	link := f.app.links[dataKey(f.instances[1].NodeID, protocol.ChannelBulk)]
	f.app.mu.Unlock()
	if link == nil {
		t.Fatal("expected actual target bulk connection")
	}
	if err := link.Close(); err != nil {
		t.Fatal(err)
	}
	p := awaitTransfer(t, f, task, model.Succeeded)
	if p.CommittedBytes != int64(len(body)) || !p.DestinationVerified {
		t.Fatalf("%+v", p)
	}
	actual, err := os.ReadFile(filepath.Join(f.roots[1], "after-reconnect.bin"))
	if err != nil || !bytes.Equal(actual, body) {
		t.Fatal("reconnected transfer content mismatch", err)
	}
	children, err := f.store.Tasks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	commits := 0
	for _, child := range children {
		if child.RequestID == transferChildKey(task, 0, 0, "file.upload") {
			commits++
		}
	}
	if commits != 1 {
		t.Fatalf("reconnect duplicated upload commit: %d", commits)
	}
	grant := model.Grant{UserID: f.readerUser.ID, Resource: instanceRef(f.instances[1]), Action: "file.write"}
	f.admin.request("POST", "/grants", grant, model.ID(), 200)
	request := model.TransferRequest{Source: model.TransferEndpoint{InstanceID: f.instances[0].ID, Path: "disconnect.bin", Version: f.stat(t, 0, "disconnect.bin").Version}, Target: model.TransferEndpoint{InstanceID: f.instances[1].ID, Path: "revoked.bin", Version: filesystem.MissingVersion}}
	task = parseFileTask(t, f.reader.request("POST", "/transfers", request, model.ID(), 202))
	eventually(t, 10*time.Second, func() bool { _, p := transferSnapshot(t, f, task.ID); return p.CurrentOffset > 0 })
	f.admin.request("DELETE", "/grants", grant, model.ID(), 200)
	p = awaitTransfer(t, f, task, model.Failed)
	if p.SourceDeleted || p.CleanupPending {
		t.Fatalf("%+v", p)
	}
	if _, err = os.Stat(filepath.Join(f.roots[1], "revoked.bin")); !os.IsNotExist(err) {
		t.Fatalf("revoked transfer committed: %v", err)
	}
	// The original owner can still inspect cleanup confirmation without gaining
	// access to manifest names after source/target authority was withdrawn.
	f.reader.request("GET", "/transfers/"+task.ID, nil, "", 200)
	f.reader.request("GET", "/transfers/"+task.ID+"/entries", nil, "", 403)
}

func TestTransferMoveRechecksSourceAfterDestinationConfirmation(t *testing.T) {
	f := newFileFixture(t)
	for _, remove := range []bool{false, true} {
		name := "late-" + strconv.FormatBool(remove) + ".txt"
		transferWrite(t, f.roots[0], name, []byte("captured source"))
		task := startTransfer(t, f, name, "copy-"+name, true, model.ID())
		eventually(t, 10*time.Second, func() bool { _, p := transferSnapshot(t, f, task.ID); return p.Stage == "verify_source" })
		if remove {
			if err := os.Remove(filepath.Join(f.roots[0], name)); err != nil {
				t.Fatal(err)
			}
		} else {
			transferWrite(t, f.roots[0], name, []byte("new content must survive"))
		}
		p := awaitTransfer(t, f, task, model.Failed)
		if !p.DestinationVerified || p.SourceDeleted || !p.Partial {
			t.Fatalf("%+v", p)
		}
		assertDisk(t, f.roots[1], "copy-"+name, "captured source")
		if !remove {
			assertDisk(t, f.roots[0], name, "new content must survive")
		}
	}
}

func TestTransferDirectoryMergeChecksFileBaselinesAndKeepsExtraFiles(t *testing.T) {
	f := newFileFixture(t)
	transferWrite(t, f.roots[0], "source/config.txt", []byte("replacement"))
	transferWrite(t, f.roots[0], "source/nested/value.txt", []byte("nested value"))
	transferWrite(t, f.roots[1], "existing/config.txt", []byte("old configuration"))
	transferWrite(t, f.roots[1], "existing/nested/keep.txt", []byte("not part of source"))
	if err := os.Chmod(filepath.Join(f.roots[1], "existing"), 0700); err != nil {
		t.Fatal(err)
	}
	request := model.TransferRequest{Source: model.TransferEndpoint{InstanceID: f.instances[0].ID, Path: "source", Version: f.stat(t, 0, "source").Version}, Target: model.TransferEndpoint{InstanceID: f.instances[1].ID, Path: "existing", Version: f.stat(t, 1, "existing").Version}}
	task := parseFileTask(t, f.admin.request("POST", "/transfers", request, model.ID(), 202))
	awaitTransfer(t, f, task, model.Succeeded)
	assertDisk(t, f.roots[1], "existing/config.txt", "replacement")
	assertDisk(t, f.roots[1], "existing/nested/value.txt", "nested value")
	assertDisk(t, f.roots[1], "existing/nested/keep.txt", "not part of source")
	info, err := os.Stat(filepath.Join(f.roots[1], "existing"))
	if err != nil || info.Mode().Perm() != 0700 {
		t.Fatalf("existing directory metadata replaced: %v", err)
	}
}

func TestTransferUnknownSourceDeleteReceiptIsNeverReplayed(t *testing.T) {
	f, restart := newRestartTransferFixture(t)
	transferWrite(t, f.roots[0], "unknown-delete.txt", []byte("source must remain"))
	task := startTransfer(t, f, "unknown-delete.txt", "verified-copy.txt", true, model.ID())
	eventually(t, 10*time.Second, func() bool { _, p := transferSnapshot(t, f, task.ID); return p.Stage == "verify_source" })
	f.app.closeTransfers()
	// Inject the persisted state of a dispatched deletion whose node receipt
	// could not be recovered. WaitingClient prevents the test setup itself
	// from dispatching the mutation; the production coordinator must treat the
	// later Interrupted receipt exactly like a real reconciliation failure.
	var payload model.TransferTaskPayload
	if err := json.Unmarshal(task.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	var manifest transferManifest
	if _, err := f.store.Record(context.Background(), transferManifestNS, task.ID, &manifest); err != nil {
		t.Fatal(err)
	}
	body, _ := transferChildData(model.FileTaskPayload{Config: payload.SourceConfig, Path: payload.Source.Path, Version: payload.Source.Version, ExpectedObjectID: manifest.SourceRelation.ObjectID}, task.ID)
	child, _, err := f.store.Accept(context.Background(), model.Task{ActorID: task.ActorID, RequestID: transferChildKey(task, 1, 0, "file.delete"), Resource: payload.SourceResource, Action: "file.delete", Payload: body, State: model.WaitingClient})
	if err != nil {
		t.Fatal(err)
	}
	child, err = f.store.MarkDispatch(context.Background(), child.ID, child.Revision)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.store.UpdateTask(context.Background(), child.ID, child.Revision, model.Interrupted, "node_acceptance_missing", nil, "injected lost durable receipt"); err != nil {
		t.Fatal(err)
	}
	restart()
	p := awaitTransfer(t, f, task, model.Interrupted)
	if !p.SourceOutcomeUnknown || p.SourceDeleted || !p.DestinationVerified {
		t.Fatalf("%+v", p)
	}
	assertDisk(t, f.roots[0], "unknown-delete.txt", "source must remain")
	assertDisk(t, f.roots[1], "verified-copy.txt", "source must remain")
	children, err := f.store.Tasks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, candidate := range children {
		if candidate.Action == "file.delete" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("unknown deletion replayed %d times", count)
	}
}

func transferInstance(t *testing.T, f *fileFixture, nodeIndex int, root string) model.Instance {
	t.Helper()
	if err := os.MkdirAll(root, 0750); err != nil {
		t.Fatal(err)
	}
	response := f.admin.request("POST", "/instances", map[string]any{"nodeId": f.instances[nodeIndex].NodeID, "name": "related-resource-" + model.ID(), "config": model.InstanceConfig{Mode: "native", Directory: root, Command: []string{"unused-relation-fixture"}, StopSeconds: 1, KillSeconds: 1}}, model.ID(), 201)
	var instance model.Instance
	if err := json.Unmarshal(response["instance"], &instance); err != nil {
		t.Fatal(err)
	}
	// Initialize the actual bound root before collecting source versions.
	f.admin.request("GET", fileURL(instance)+"/stat?path=.", nil, "", 200)
	return instance
}

func TestTransferSameNodeSeparateInstancesAndActualOverlapRejection(t *testing.T) {
	f := newFileFixture(t)
	independentRoot := filepath.Join(t.TempDir(), "independent")
	independent := transferInstance(t, f, 0, independentRoot)
	transferWrite(t, f.roots[0], "same-node/a.txt", []byte("same node content"))
	request := model.TransferRequest{Source: model.TransferEndpoint{InstanceID: f.instances[0].ID, Path: "same-node", Version: f.stat(t, 0, "same-node").Version}, Target: model.TransferEndpoint{InstanceID: independent.ID, Path: "moved-directory", Version: filesystem.MissingVersion}, Move: true}
	task := parseFileTask(t, f.admin.request("POST", "/transfers", request, model.ID(), 202))
	p := awaitTransfer(t, f, task, model.Succeeded)
	if !p.SourceDeleted {
		t.Fatalf("%+v", p)
	}
	assertDisk(t, independentRoot, "moved-directory/a.txt", "same node content")
	transferWrite(t, f.roots[0], "shared-root/source.txt", []byte("do not delete"))
	// Two distinct signed Daemons here run on this actual host. A different
	// node ID must not make their overlapping native roots appear independent.
	alias := transferInstance(t, f, 1, filepath.Join(f.roots[0], "shared-root"))
	for _, targetPath := range []string{".", "nested-copy"} {
		version := filesystem.MissingVersion
		if targetPath == "." {
			version = f.stat(t, 0, "shared-root").Version
		}
		request = model.TransferRequest{Source: model.TransferEndpoint{InstanceID: f.instances[0].ID, Path: "shared-root", Version: f.stat(t, 0, "shared-root").Version}, Target: model.TransferEndpoint{InstanceID: alias.ID, Path: targetPath, Version: version}, Move: true}
		task = parseFileTask(t, f.admin.request("POST", "/transfers", request, model.ID(), 202))
		p = awaitTransfer(t, f, task, model.Failed)
		if p.Completed != 0 || p.SourceDeleted {
			t.Fatalf("overlap mutated resources: %+v", p)
		}
		assertDisk(t, f.roots[0], "shared-root/source.txt", "do not delete")
	}
	if err := os.Link(filepath.Join(f.roots[0], "shared-root/source.txt"), filepath.Join(independentRoot, "hardlink.txt")); err != nil {
		t.Fatal(err)
	}
	sourceVersion := f.stat(t, 0, "shared-root/source.txt").Version
	request = model.TransferRequest{Source: model.TransferEndpoint{InstanceID: f.instances[0].ID, Path: "shared-root/source.txt", Version: sourceVersion}, Target: model.TransferEndpoint{InstanceID: independent.ID, Path: "hardlink.txt", Version: sourceVersion}, Move: true}
	task = parseFileTask(t, f.admin.request("POST", "/transfers", request, model.ID(), 202))
	awaitTransfer(t, f, task, model.Failed)
	assertDisk(t, independentRoot, "hardlink.txt", "do not delete")
}

func TestTransferSameContentSourceObjectReplacementIsNotDeleted(t *testing.T) {
	f := newFileFixture(t)
	transferWrite(t, f.roots[0], "identity.txt", []byte("same bytes"))
	before, err := os.Stat(filepath.Join(f.roots[0], "identity.txt"))
	if err != nil {
		t.Fatal(err)
	}
	task := startTransfer(t, f, "identity.txt", "identity-copy.txt", true, model.ID())
	eventually(t, 10*time.Second, func() bool { _, p := transferSnapshot(t, f, task.ID); return p.Stage == "verify_source" })
	if err = os.Rename(filepath.Join(f.roots[0], "identity.txt"), filepath.Join(f.roots[0], "original-identity.txt")); err != nil {
		t.Fatal(err)
	}
	transferWrite(t, f.roots[0], "identity.txt", []byte("same bytes"))
	if err = os.Chtimes(filepath.Join(f.roots[0], "identity.txt"), before.ModTime(), before.ModTime()); err != nil {
		t.Fatal(err)
	}
	p := awaitTransfer(t, f, task, model.Failed)
	if p.SourceDeleted || !p.DestinationVerified {
		t.Fatalf("%+v", p)
	}
	assertDisk(t, f.roots[0], "identity.txt", "same bytes")
	assertDisk(t, f.roots[1], "identity-copy.txt", "same bytes")
}

func TestTransferGuessedChildRequestIDCannotCertifyAnotherDeletion(t *testing.T) {
	f, restart := newRestartTransferFixture(t)
	transferWrite(t, f.roots[0], "intended-source.txt", []byte("must remain"))
	transferWrite(t, f.roots[0], "unrelated.txt", []byte("authorized separate deletion"))
	f.app.closeTransfers()
	parent := startTransfer(t, f, "intended-source.txt", "target-copy.txt", true, model.ID())
	// A browser chooses its own Idempotency-Key. This actual, independently
	// authorized deletion must never serve as the future transfer's receipt.
	foreign := parseFileTask(t, f.admin.request("POST", fileURL(f.instances[0])+"/actions", map[string]any{"action": "delete", "path": "unrelated.txt", "version": f.stat(t, 0, "unrelated.txt").Version}, transferChildKey(parent, 1, 0, "file.delete"), 202))
	f.awaitFile(t, foreign, model.Succeeded)
	restart()
	p := awaitTransfer(t, f, parent, model.Failed)
	if p.SourceDeleted || p.CleanupPending {
		t.Fatalf("foreign child certified deletion: %+v", p)
	}
	assertDisk(t, f.roots[0], "intended-source.txt", "must remain")
	assertDisk(t, f.roots[1], "target-copy.txt", "must remain")
}

package master

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"blora.dev/panel/internal/containers"
	"blora.dev/panel/internal/daemon"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"github.com/coder/websocket"
)

func TestComposeProjectSaveAcrossDaemonRestartAndExplicitApply(t *testing.T) {
	image := os.Getenv("BLORA_CONTAINER_E2E_IMAGE")
	if image == "" {
		t.Skip("requires explicit task-owned Docker test image; missing environment is not a pass")
	}
	f := newFileFixtureWithConfig(t, func(c *daemon.Config) { c.DockerEndpoint = "unix:///var/run/docker.sock" })
	url := "/nodes/" + f.instances[0].NodeID + "/docker"
	saveID, projectID := model.ID(), "api-project-"+model.ID()
	// This source deliberately exceeds the control and single RPC frame
	// limits; only acknowledged 64KiB chunks cross the bulk connection.
	source := []byte(strings.Repeat("# 中文-staged-source\n", 20000) + "services:\n  worker:\n    image: " + image + "\n    network_mode: none\n    command: [sh, -c, \"trap 'exit 0' TERM; while :; do sleep .1; done\"]\n")
	prepare := model.ComposeSaveRequest{SaveID: saveID, ProjectID: projectID, Total: int64(len(source)), SHA256: strings.TrimPrefix(fileHash(source), "sha256:")}
	f.reader.request("POST", url+"/project-saves", prepare, "", 403)
	f.admin.request("POST", url+"/project-saves", prepare, "", 400)
	prepareKey := model.ID()
	f.admin.request("POST", url+"/project-saves", prepare, prepareKey, 201)
	uploadURL := url + "/project-saves/" + saveID
	restarted := false
	for offset := 0; offset < len(source); {
		end := min(offset+containers.ProjectChunkBytes, len(source))
		data := source[offset:end]
		chunk := model.ComposeSaveRequest{Offset: int64(offset), Data: data, SHA256: strings.TrimPrefix(fileHash(data), "sha256:")}
		chunkKey := model.ID()
		f.admin.request("PUT", uploadURL+"/chunks", chunk, chunkKey, 200)
		if !restarted {
			f.restartNode(0)
			restarted = true
			// A lost chunk ACK can be retried verbatim after process restart.
			f.admin.request("PUT", uploadURL+"/chunks", chunk, chunkKey, 200)
		}
		offset = end
	}
	key := model.ID()
	saved := parseFileTask(t, f.admin.request("POST", uploadURL+"/commit", nil, key, 202))
	if strings.Contains(string(saved.Payload), "staged-source") {
		t.Fatal("Compose body entered the control task")
	}
	saved = f.awaitFile(t, saved, model.Succeeded)
	if again := parseFileTask(t, f.admin.request("POST", uploadURL+"/commit", nil, key, 202)); again.ID != saved.ID {
		t.Fatal("save commit duplicated")
	}
	var receipt containers.ProjectSave
	if err := json.Unmarshal(saved.Result, &receipt); err != nil || receipt.State != "COMMITTED" || receipt.Project == nil || receipt.Project.Revision != 1 {
		t.Fatal("project save receipt missing")
	}
	var restored []byte
	for offset := 0; offset < len(source); {
		response := f.admin.request("GET", url+"/projects/"+projectID+"/content?revision=1&offset="+strconv.Itoa(offset), nil, "", 200)
		var b []byte
		if err := json.Unmarshal(response["data"], &b); err != nil || len(b) == 0 {
			t.Fatal("project content chunk missing")
		}
		restored = append(restored, b...)
		offset += len(b)
	}
	if fileHash(restored) != fileHash(source) {
		t.Fatal("project source changed across staged save")
	}
	query := func() containers.Snapshot {
		r := f.admin.request("POST", url+"/query", containers.Query{Kind: "project", ProjectID: projectID}, "", 200)
		b, _ := json.Marshal(r)
		var snapshot containers.Snapshot
		if err := json.Unmarshal(b, &snapshot); err != nil {
			t.Fatal(err)
		}
		return snapshot
	}
	if snapshot := query(); snapshot.Project == nil || snapshot.Project.AppliedRevision != 0 || len(snapshot.Items) != 0 {
		t.Fatal("saving source deployed containers")
	}
	call := func(op containers.Operation) model.Task {
		return parseFileTask(t, f.admin.request("POST", url+"/actions", op, model.ID(), 202))
	}
	removed := false
	t.Cleanup(func() {
		if !removed {
			f.awaitFile(t, call(containers.Operation{Action: "compose.delete", ProjectID: projectID, Revision: 1, StopSeconds: 1}), model.Succeeded)
		}
	})
	f.awaitFile(t, call(containers.Operation{Action: "compose.apply", ProjectID: projectID, Revision: 1, PullPolicy: "never", HealthSeconds: 5}), model.Succeeded)
	if snapshot := query(); snapshot.Project.AppliedRevision != 1 || len(snapshot.Items) != 1 {
		t.Fatal("explicit Compose apply did not create requested service")
	}
	f.awaitFile(t, call(containers.Operation{Action: "compose.delete", ProjectID: projectID, Revision: 1, StopSeconds: 1}), model.Succeeded)
	removed = true
}

func TestContainerNodeAPIRequiresHostAuthorityAndReportsMissingEngine(t *testing.T) {
	f := newFileFixture(t)
	url := "/nodes/" + f.instances[0].NodeID + "/docker"
	f.reader.request("POST", url+"/query", containers.Query{Kind: "containers"}, "", 403)
	f.reader.request("POST", url+"/query", containers.Query{Kind: "operation-output", TaskID: "unrelated"}, "", 403)
	f.admin.request("POST", url+"/query", containers.Query{Kind: "operation-output", TaskID: "unrelated", OutputOffset: -1}, "", 400)
	f.reader.request("POST", url+"/actions", containers.Operation{Action: "container.create"}, model.ID(), 403)
	// Resource authorization precedes request-key/body/query validation on every
	// staged source route, including requests that never reach an Engine.
	f.reader.request("POST", url+"/project-saves", nil, "", 403)
	f.reader.request("PUT", url+"/project-saves/unowned/chunks", nil, "", 403)
	f.reader.request("POST", url+"/project-saves/unowned/commit", nil, "", 403)
	f.reader.request("GET", url+"/projects/unowned/content?revision=invalid", nil, "", 403)
	f.admin.request("POST", url+"/project-saves", nil, "", 400)
	f.admin.request("PUT", url+"/project-saves/unowned/chunks", nil, "", 400)
	response := f.admin.request("POST", url+"/query", containers.Query{Kind: "capabilities"}, "", 200)
	var capabilities map[string]string
	if err := json.Unmarshal(response["capabilities"], &capabilities); err != nil || len(capabilities) == 0 {
		t.Fatal("missing capability response")
	}
	key := model.ID()
	op := containers.Operation{Action: "image.pull", Image: "unused-no-engine:local"}
	task := parseFileTask(t, f.admin.request("POST", url+"/actions", op, key, 202))
	if task.Resource != nodeRef(f.instances[0].NodeID) {
		t.Fatal("host task has incorrect resource scope")
	}
	task = f.awaitFile(t, task, model.Failed)
	if !strings.Contains(task.Error, "no Engine endpoint configured") {
		t.Fatalf("dependency failure not reported: %s", task.Error)
	}
	if again := parseFileTask(t, f.admin.request("POST", url+"/actions", op, key, 202)); again.ID != task.ID {
		t.Fatal("host task retry created new task")
	}
	f.reader.request("GET", "/tasks/"+task.ID, nil, "", 404)
}

func TestContainerNodeAPIRealEngineLifecycle(t *testing.T) {
	image := os.Getenv("BLORA_CONTAINER_E2E_IMAGE")
	if image == "" {
		t.Skip("requires explicit task-owned Docker test image; missing environment is not a pass")
	}
	f := newFileFixtureWithConfig(t, func(c *daemon.Config) { c.DockerEndpoint = "unix:///var/run/docker.sock" })
	url := "/nodes/" + f.instances[0].NodeID + "/docker"
	call := func(op containers.Operation) model.Task {
		return parseFileTask(t, f.admin.request("POST", url+"/actions", op, model.ID(), 202))
	}
	config, _ := json.Marshal(map[string]any{"Image": image, "Cmd": []string{"/bin/sh", "-c", "printf '中文-Docker-logs\\n'; trap 'exit 0' TERM; while :; do sleep .1; done"}, "HostConfig": map[string]any{"NetworkMode": "none", "ReadonlyRootfs": true, "CapDrop": []string{"ALL"}, "SecurityOpt": []string{"no-new-privileges:true"}}})
	created := f.awaitFile(t, call(containers.Operation{Action: "container.create", Name: "blora-api-test-" + model.ID(), Config: config}), model.Succeeded)
	var result struct {
		Targets []containers.Target `json:"targets"`
	}
	if err := json.Unmarshal(created.Result, &result); err != nil || len(result.Targets) != 1 {
		t.Fatal("real container identity missing")
	}
	target := result.Targets[0]
	removed := false
	t.Cleanup(func() {
		if !removed {
			f.awaitFile(t, call(containers.Operation{Action: "container.delete", Target: target, Force: true}), model.Succeeded)
		}
	})
	f.awaitFile(t, call(containers.Operation{Action: "container.start", Target: target}), model.Succeeded)
	response := f.admin.request("POST", url+"/query", containers.Query{Kind: "container", Target: target}, "", 200)
	var objects []containers.Object
	if err := json.Unmarshal(response["items"], &objects); err != nil || len(objects) != 1 || objects[0].Target.ID != target.ID || objects[0].State != "running" {
		t.Fatal("Engine did not observe the requested container running")
	}
	grant := model.Grant{UserID: f.readerUser.ID, Resource: nodeRef(f.instances[0].NodeID), Action: "host.manage"}
	historyPath := url + "/containers/" + target.ID + "/logs/history?limit=1"
	f.reader.request("GET", historyPath, nil, "", 403)
	f.admin.request("POST", "/grants", grant, model.ID(), 200)
	managed := f.reader.request("GET", "/nodes/managed", nil, "", 200)
	var nodes []model.Node
	if err := json.Unmarshal(managed["items"], &nodes); err != nil || len(nodes) != 1 {
		t.Fatal("host-only permission cannot discover its managed node")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	endpoint := strings.Replace(f.reader.base, "https://", "wss://", 1) + "/api/v1" + url + "/containers/" + target.ID + "/logs"
	ws, _, err := websocket.Dial(ctx, endpoint, &websocket.DialOptions{HTTPClient: f.reader.client, HTTPHeader: http.Header{"Origin": []string{f.reader.base}}})
	if err != nil {
		t.Fatal(err)
	}
	stream, err := protocol.NewConn(ctx, ws, protocol.Options{Generation: 1, Channel: protocol.ChannelInteractive, InitialStreams: []string{target.ID}})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	for {
		e, err := stream.Read(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if e.Type == protocol.TypeError {
			t.Fatalf("Docker logs: %s", e.Payload)
		}
		if e.Type != protocol.TypeData {
			continue
		}
		var window model.ContainerLogWindow
		if err := json.Unmarshal(e.Payload, &window); err != nil {
			t.Fatal(err)
		}
		text := ""
		for _, frame := range window.Frames {
			text += string(frame.Data)
		}
		if !strings.Contains(text, "中文-Docker-logs") || !window.PossibleGap {
			t.Fatal("real Docker window missing output or honest retention semantics")
		}
		break // Deliberately retain byte credit while stopping the container.
	}
	history := f.admin.request("GET", historyPath, nil, "", 200)
	var archived []model.ContainerLogWindow
	if err := json.Unmarshal(history["items"], &archived); err != nil || len(archived) != 1 || archived[0].ContainerID != target.ID || !archived[0].PossibleGap {
		t.Fatalf("real Docker history endpoint missing persisted window: %s", history["items"])
	}
	f.admin.request("GET", url+"/containers/"+target.ID+"/logs/history?limit=101", nil, "", 400)
	f.awaitFile(t, call(containers.Operation{Action: "container.stop", Target: target, StopSeconds: 1}), model.Succeeded)
	f.admin.request("DELETE", "/grants", grant, model.ID(), 200)
	select {
	case <-stream.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("revoked Docker log stream remained connected")
	}
	f.awaitFile(t, call(containers.Operation{Action: "container.delete", Target: target}), model.Succeeded)
	removed = true
}

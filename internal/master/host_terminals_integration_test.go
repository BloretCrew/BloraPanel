package master

import (
	"blora.dev/panel/internal/containers"
	"blora.dev/panel/internal/containerterm/helper"
	"blora.dev/panel/internal/daemon"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/terminal"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestHostDockerTerminalAPIActorBindingAndRefresh(t *testing.T) {
	image, binary := os.Getenv("BLORA_CONTAINER_E2E_IMAGE"), os.Getenv("BLORA_HOST_EXEC_HELPER_BINARY")
	if image == "" || binary == "" {
		t.Skip("requires explicit owned Docker image and new processTree helper")
	}
	f := newFileFixtureWithConfig(t, func(c *daemon.Config) { c.DockerEndpoint = "unix:///var/run/docker.sock" })
	node := f.instances[0].NodeID
	base := "/nodes/" + node
	action := func(op containers.Operation) model.Task {
		return f.awaitFile(t, parseFileTask(t, f.admin.request("POST", base+"/docker/actions", op, model.ID(), 202)), model.Succeeded)
	}
	label := model.ID()
	config, _ := json.Marshal(map[string]any{"Image": image, "Cmd": []string{"/bin/sh", "-c", "exec sleep 300"}, "Labels": map[string]string{"dev.blora.host-terminal-test": label}, "HostConfig": map[string]any{"NetworkMode": "none"}})
	created := action(containers.Operation{Action: "container.create", Name: "blora-host-terminal-" + label, Config: config})
	var result struct {
		Targets []containers.Target `json:"targets"`
	}
	if err := json.Unmarshal(created.Result, &result); err != nil || len(result.Targets) != 1 {
		t.Fatal("container identity missing", err)
	}
	target := result.Targets[0]
	t.Cleanup(func() { action(containers.Operation{Action: "container.delete", Target: target, Force: true}) })
	action(containers.Operation{Action: "container.start", Target: target})
	endpoint := base + "/docker/containers/" + target.ID + "/terminals"
	input := map[string]any{"cols": 90, "rows": 24, "createdAt": target.CreatedAt}
	f.reader.request("POST", endpoint, input, model.ID(), 403)
	// Old helpers cannot safely account for escaped descendants. No shell may
	// be created until the stronger helper capability is actually installed.
	unsupported := f.awaitFile(t, parseFileTask(t, f.admin.request("POST", endpoint, input, model.ID(), 202)), model.Failed)
	if !strings.Contains(unsupported.Error, "capability") && !strings.Contains(unsupported.Error, "helper") {
		t.Fatal("missing dependency not explicit", unsupported.Error)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if output, err := exec.CommandContext(ctx, "docker", "--host", "unix:///var/run/docker.sock", "cp", binary, target.ID+":"+helper.Path).CombinedOutput(); err != nil {
		t.Fatal(err, string(output))
	}
	grant := model.Grant{UserID: f.readerUser.ID, Resource: nodeRef(node), Action: "host.manage"}
	f.admin.request("POST", "/grants", grant, model.ID(), 200)
	create := func(c *testClient) terminal.Session {
		key := model.ID()
		task := parseFileTask(t, c.request("POST", endpoint, input, key, 202))
		duplicate := parseFileTask(t, c.request("POST", endpoint, input, key, 202))
		if task.ID != duplicate.ID {
			t.Fatal("duplicate shell task")
		}
		task = f.awaitFile(t, task, model.Succeeded)
		var out struct {
			Session terminal.Session `json:"session"`
		}
		if err := json.Unmarshal(task.Result, &out); err != nil {
			t.Fatal(err)
		}
		if out.Session.Backend != "docker-host" || out.Session.Resource != nodeRef(node) {
			t.Fatal("incorrect scope")
		}
		return out.Session
	}
	admin, member := create(f.admin), create(f.reader)
	if admin.RunID != member.RunID || admin.ID == member.ID || admin.OwnerID == member.OwnerID {
		t.Fatal("host actor/run binding conflated")
	}
	for _, pair := range []struct {
		client  *testClient
		session terminal.Session
	}{{f.admin, admin}, {f.reader, member}} {
		list := pair.client.request("GET", base+"/terminals", nil, "", 200)
		var sessions []terminal.Session
		if err := json.Unmarshal(list["items"], &sessions); err != nil {
			t.Fatal(err)
		}
		found := false
		for _, s := range sessions {
			if s.OwnerID != pair.session.OwnerID {
				t.Fatal("other actor session listed")
			}
			found = found || s.ID == pair.session.ID
		}
		if !found {
			t.Fatal("own session missing")
		}
	}
	f.reader.request("POST", "/terminals/"+admin.ID+"/close", map[string]any{}, model.ID(), 403)
	first := openTerminal(t, f.reader, member.ID, "host-primary", 0)
	first.resume()
	first.input("printf '\\033[32m\\344\\270\\255\\346\\226\\207-host-OK\\033[0m\\n'\n")
	for !strings.Contains(first.output.String(), "中文-host-OK") {
		first.read()
	}
	if !strings.Contains(first.output.String(), "\x1b[32m") {
		t.Fatal("ANSI lost")
	}
	checkpoint := first.cursor
	first.conn.Close()
	restored := openTerminal(t, f.reader, member.ID, "host-primary", checkpoint)
	restored.resume()
	f.admin.request("DELETE", "/grants", grant, model.ID(), 200)
	eventually(t, 3*time.Second, func() bool {
		select {
		case <-restored.conn.Done():
			return true
		default:
			return false
		}
	})
	f.reader.request("GET", base+"/terminals", nil, "", 403)
	closed := parseFileTask(t, f.admin.request("POST", "/terminals/"+admin.ID+"/close", map[string]any{}, model.ID(), 202))
	f.awaitFile(t, closed, model.Succeeded)
	query := f.admin.request("POST", base+"/docker/query", containers.Query{Kind: "container", Target: target}, "", 200)
	var objects []containers.Object
	if err := json.Unmarshal(query["items"], &objects); err != nil || len(objects) != 1 || objects[0].State != "running" {
		t.Fatal("closing shell changed container lifecycle", err)
	}
}

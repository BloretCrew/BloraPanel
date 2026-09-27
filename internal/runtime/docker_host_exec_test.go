package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHostDockerScopeBindsFullBirthAndKeepsManagedInstanceBoundary(t *testing.T) {
	id := strings.Repeat("a", 64)
	target := HostDockerTarget{ContainerID: id, CreatedAt: "2026-09-09T00:00:00Z", StartedAt: "2026-09-09T01:00:00Z"}
	created, started, pidMode := target.CreatedAt, target.StartedAt, ""
	labels := map[string]string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/version" {
			json.NewEncoder(w).Encode(map[string]string{"ApiVersion": "1.54", "MinAPIVersion": "1.40", "Os": "linux"})
			return
		}
		if r.URL.Path != "/v1.45/containers/"+id+"/json" {
			t.Errorf("request did not use full fixed container ID: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"Id": id, "Created": created, "Config": map[string]any{"User": "0:0", "Labels": labels}, "State": map[string]any{"Running": true, "StartedAt": started}, "HostConfig": map[string]any{"PidMode": pidMode}})
	}))
	defer server.Close()
	m, err := New(Options{StateRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	m.docker = &dockerClient{client: server.Client(), base: server.URL}
	scope := &dockerExecScope{manager: m, host: &target, run: Record{ContainerID: id, Backend: "docker-host", RunID: target.Reference()}, authorize: func(context.Context) error { return nil }}
	if _, err = scope.container(context.Background(), true); err != nil {
		t.Fatal("authorized host container rejected", err)
	}
	created = "changed"
	if _, err = scope.container(context.Background(), true); !errors.Is(err, ErrUnknown) {
		t.Fatal("creation birth mismatch accepted", err)
	}
	created = target.CreatedAt
	started = "changed"
	if _, err = scope.container(context.Background(), true); !errors.Is(err, ErrUnknown) {
		t.Fatal("container restart birth mismatch accepted", err)
	}
	started = target.StartedAt
	pidMode = "host"
	if _, err = scope.container(context.Background(), true); !errors.Is(err, ErrCapability) {
		t.Fatal("host PID namespace accepted", err)
	}
	pidMode = ""
	labels["dev.blora.run"] = "run"
	labels["dev.blora.instance"] = "instance"
	if _, err = scope.container(context.Background(), true); err == nil {
		t.Fatal("host terminal bypassed managed instance boundary")
	}
	if scope.host.Reference() == (HostDockerTarget{ContainerID: id, CreatedAt: target.CreatedAt, StartedAt: "other"}).Reference() {
		t.Fatal("run reference ignored birth")
	}
}

package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
)

func TestDockerConfigurationAndOwnershipContract(t *testing.T) {
	var mu sync.Mutex
	var creation map[string]any
	state := "created"
	var mutations []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if q.URL.Path == "/version" {
			json.NewEncoder(w).Encode(map[string]string{"ApiVersion": "1.54", "MinAPIVersion": "1.44", "Os": "linux"})
			return
		}
		if q.Method != "GET" {
			mutations = append(mutations, q.Method+" "+q.URL.RequestURI())
		}
		switch {
		case q.URL.Path == "/v1.45/containers/create":
			if err := json.NewDecoder(q.Body).Decode(&creation); err != nil {
				t.Error(err)
			}
			w.WriteHeader(201)
			json.NewEncoder(w).Encode(map[string]string{"Id": "test-container-id"})
		case strings.HasSuffix(q.URL.Path, "/json"):
			if creation == nil {
				http.Error(w, "missing", 404)
				return
			}
			json.NewEncoder(w).Encode(map[string]any{"Id": "test-container-id", "Config": map[string]any{"Labels": creation["Labels"]}, "State": map[string]any{"Running": state == "running", "Status": state, "ExitCode": 0}})
		case strings.HasSuffix(q.URL.Path, "/start"):
			state = "running"
			w.WriteHeader(204)
		case strings.HasSuffix(q.URL.Path, "/kill"):
			state = "exited"
			w.WriteHeader(204)
		case q.Method == "DELETE":
			if q.URL.Query().Get("v") != "false" || q.URL.Query().Get("force") != "false" {
				t.Error("cleanup requested destructive volume or forced deletion")
			}
			w.WriteHeader(204)
		default:
			http.Error(w, "unexpected request", 500)
		}
	}))
	defer server.Close()
	root := t.TempDir()
	id := model.ID()
	if err := os.Mkdir(filepath.Join(root, id), 0750); err != nil {
		t.Fatal(err)
	}
	m, err := New(Options{StateRoot: t.TempDir(), InstanceRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	m.docker = &dockerClient{client: server.Client(), base: server.URL}
	r, err := m.Start(context.Background(), model.ID(), id, model.InstanceConfig{Mode: "isolated", Image: "example@sha256:test", Command: []string{"/bin/sh", "-c", "sleep 30"}}, IO{})
	if err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	host := creation["HostConfig"].(map[string]any)
	if host["Privileged"] != false || host["ReadonlyRootfs"] != true || host["NetworkMode"] != "bridge" || host["PidMode"] != "" {
		t.Fatalf("unsafe HostConfig: %+v", host)
	}
	if creation["User"] != "65534:65534" || host["Memory"].(float64) <= 0 || host["PidsLimit"].(float64) <= 0 || host["CPUQuota"].(float64) <= 0 {
		t.Fatal("missing identity or resource constraints")
	}
	if host["CapDrop"].([]any)[0] != "ALL" || host["SecurityOpt"].([]any)[0] != "no-new-privileges:true" {
		t.Fatal("capability constraints missing")
	}
	mounts := host["Mounts"].([]any)
	if len(mounts) != 1 || mounts[0].(map[string]any)["Source"] != filepath.Join(root, id) {
		t.Fatal("unexpected host mount")
	}
	mu.Unlock()
	forged := r
	forged.Token = model.ID()
	if _, err = m.Stop(context.Background(), forged, StopPolicy{Force: true}, nil); !errors.Is(err, ErrUnknown) {
		t.Fatalf("foreign labels accepted: %v", err)
	}
	o, err := m.Stop(context.Background(), r, StopPolicy{Force: true, KillWait: time.Second}, nil)
	if err != nil || !o.Exited {
		t.Fatalf("stop not confirmed %+v %v", o, err)
	}
	if err = m.Cleanup(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(mutations) != 4 {
		t.Fatalf("unexpected side effects: %v", mutations)
	}
}

func TestDockerVersionMinimumAndInsecureEndpointRejected(t *testing.T) {
	if _, err := newDockerClient("http://127.0.0.1:2375"); err == nil {
		t.Fatal("unverified remote Engine accepted")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"ApiVersion": "1.55", "MinAPIVersion": "1.50", "Os": "linux"})
	}))
	defer server.Close()
	d := dockerClient{client: server.Client(), base: server.URL}
	if err := d.check(context.Background()); !errors.Is(err, ErrCapability) {
		t.Fatalf("unsupported API minimum accepted: %v", err)
	}
}

func TestDockerStdinAttachSendsOnce(t *testing.T) {
	r := Record{RunID: model.ID(), InstanceID: model.ID(), Token: model.ID(), Backend: "docker", ContainerID: "stdin-test", Phase: "running"}
	received := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) {
		switch {
		case q.URL.Path == "/version":
			json.NewEncoder(w).Encode(map[string]string{"ApiVersion": "1.54", "MinAPIVersion": "1.44", "Os": "linux"})
		case strings.HasSuffix(q.URL.Path, "/json"):
			json.NewEncoder(w).Encode(map[string]any{"Id": r.ContainerID, "Config": map[string]any{"Labels": map[string]string{"dev.blora.run": r.RunID, "dev.blora.instance": r.InstanceID, "dev.blora.owner": r.Token}}, "State": map[string]any{"Running": true, "Status": "running"}})
		case strings.HasSuffix(q.URL.Path, "/attach"):
			if q.Header.Get("Upgrade") != "tcp" || q.URL.Query().Get("stdin") != "true" {
				t.Error("missing stdin upgrade")
			}
			conn, rw, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			defer conn.Close()
			if _, err = rw.WriteString("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n\r\n"); err != nil {
				t.Error(err)
				return
			}
			if err = rw.Flush(); err != nil {
				t.Error(err)
				return
			}
			b := make([]byte, 5)
			if _, err = io.ReadFull(rw, b); err != nil {
				t.Error(err)
				return
			}
			received <- string(b)
		default:
			http.Error(w, "unexpected request", 500)
		}
	}))
	defer server.Close()
	m, err := New(Options{StateRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	m.docker = &dockerClient{client: server.Client(), base: server.URL}
	n, err := m.WriteInput(context.Background(), r, []byte("stop\n"))
	if err != nil || n != 5 {
		t.Fatalf("stdin attach failed: %d %v", n, err)
	}
	select {
	case got := <-received:
		if got != "stop\n" {
			t.Fatalf("wrong input %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("stdin did not reach Engine attachment")
	}
}

func TestDockerRealIsolatedLifecycle(t *testing.T) {
	endpoint, image := os.Getenv("BLORA_TEST_DOCKER_ENDPOINT"), os.Getenv("BLORA_TEST_DOCKER_IMAGE")
	if endpoint == "" || image == "" {
		t.Skip("environment missing: set BLORA_TEST_DOCKER_ENDPOINT and BLORA_TEST_DOCKER_IMAGE to an authorized Engine and existing test image")
	}
	m, err := New(Options{StateRoot: t.TempDir(), InstanceRoot: filepath.Join(t.TempDir(), "new-instance-root"), DockerEndpoint: endpoint})
	if err != nil {
		t.Fatal(err)
	}
	r, err := m.Start(context.Background(), model.ID(), model.ID(), model.InstanceConfig{Mode: "isolated", Image: image, Command: []string{"/bin/sh", "-c", "trap '' TERM; sleep 30 & wait"}}, IO{})
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_, _ = m.Stop(ctx, r, StopPolicy{Force: true, KillWait: 5 * time.Second}, nil)
		if e := m.Cleanup(ctx, r); e != nil {
			t.Logf("cleanup requires follow-up for own run %s: %v", r.RunID, e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	o, err := m.Observe(context.Background(), r)
	if err != nil || o.Exited {
		t.Fatalf("container did not run: %+v %v", o, err)
	}
	point, err := m.ContainerMetrics(context.Background(), r)
	if err != nil || point.RunID != r.RunID || point.ObservedAt.IsZero() || point.MemoryUsed <= 0 || point.MemoryTotal <= 0 {
		t.Fatalf("real container metrics unavailable: %+v %v", point, err)
	}
	for _, unavailable := range point.Unavailable {
		if unavailable == "cpu" {
			t.Fatal("real two-sample CPU unavailable")
		}
	}
	o, err = m.Stop(context.Background(), r, StopPolicy{Grace: 150 * time.Millisecond, Escalate: true, KillWait: 5 * time.Second}, nil)
	if err != nil || !o.Exited {
		t.Fatalf("container exit unconfirmed: %+v %v", o, err)
	}
}

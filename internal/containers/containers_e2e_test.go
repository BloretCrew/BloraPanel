package containers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

// This opt-in test touches only random, labelled objects it creates. It never
// prunes or removes the prebuilt fixture image or pre-existing host resources.
func TestRealDockerComposeLifecycle(t *testing.T) {
	if os.Getenv("BLORA_CONTAINER_E2E") != "1" {
		t.Skip("set BLORA_CONTAINER_E2E=1 for dedicated labelled Docker objects")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	run := "blora-e2e-" + randomID()[:16]
	image := os.Getenv("BLORA_CONTAINER_E2E_IMAGE")
	if image == "" {
		image = "blora-isolated-e2e:20260909"
	}
	endpoint := os.Getenv("BLORA_CONTAINER_E2E_ENDPOINT")
	tlsDirectory := os.Getenv("BLORA_CONTAINER_E2E_TLS_DIR")
	if os.Getenv("BLORA_CONTAINER_E2E_REMOTE_TLS") == "1" {
		if endpoint != "" {
			t.Fatal("BLORA_CONTAINER_E2E_ENDPOINT cannot be combined with BLORA_CONTAINER_E2E_REMOTE_TLS")
		}
		endpoint, tlsDirectory = startDockerSocketHTTPSProxy(t, os.Getenv("BLORA_CONTAINER_E2E_SOCKET"))
	} else if endpoint == "" {
		endpoint = "unix:///var/run/docker.sock"
	}
	m, err := New(Options{StateRoot: t.TempDir(), Endpoint: endpoint, ComposeTLSCertPath: tlsDirectory, Labels: map[string]string{"blora.dev/e2e": run}, OperationTimeout: 90 * time.Second, Authorize: func(_ context.Context, actor string) error {
		if actor == "admin" {
			return nil
		}
		return ErrForbidden
	}})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	var fixture struct {
		ID string `json:"Id"`
	}
	if err = m.engine.api(ctx, "GET", "/images/"+url.PathEscape(image)+"/json", nil, &fixture); err != nil {
		t.Fatal("prebuilt isolated fixture image required", err)
	}
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		for _, kind := range []string{"containers", "networks", "volumes", "images"} {
			objects, truncated, err := m.engine.list(cleanup, kind, map[string]string{"blora.dev/e2e": run}, 1000)
			if err != nil || truncated {
				t.Errorf("cleanup list %s: %v truncated=%v", kind, err, truncated)
				continue
			}
			for _, o := range objects {
				if o.Labels["blora.dev/e2e"] != run {
					t.Errorf("cleanup identity mismatch")
					continue
				}
				var path string
				switch kind {
				case "containers":
					if err = m.engine.api(cleanup, "POST", "/containers/"+o.Target.ID+"/stop?t=1", nil, nil); err != nil && !notFound(err) {
						t.Errorf("cleanup stop: %v", err)
					}
					path = "/containers/" + o.Target.ID + "?force=false&v=false"
				case "networks":
					path = "/networks/" + o.Target.ID
				case "volumes":
					path = "/volumes/" + url.PathEscape(o.Target.Name) + "?force=false"
				case "images":
					if o.Target.ID == fixture.ID {
						t.Errorf("fixture image unexpectedly matched test label")
						continue
					}
					path = "/images/" + o.Target.ID + "?force=false&noprune=true"
				}
				if err = m.engine.api(cleanup, "DELETE", path, nil, nil); err != nil && !notFound(err) {
					t.Errorf("cleanup %s: %v", kind, err)
				}
			}
		}
	})
	sequence := 0
	execute := func(op Operation) Result {
		t.Helper()
		sequence++
		op.TaskID = fmt.Sprintf("%s_%d", run, sequence)
		r, err := m.Execute(ctx, "admin", op, nil)
		if err != nil || r.State != "SUCCEEDED" || r.Unknown {
			t.Fatalf("%s result=%+v error=%v", op.Action, r, err)
		}
		return r
	}
	caps, err := m.Query(ctx, "admin", Query{Kind: "capabilities"})
	if err != nil || caps.Capabilities["compose"] != "available" {
		t.Fatal(caps, err)
	}
	t.Logf("capabilities: %v", caps.Capabilities)
	if _, err = m.Query(ctx, "ordinary-user", Query{Kind: "containers"}); !errors.Is(err, ErrForbidden) {
		t.Fatal("host permission bypass", err)
	}
	volume := execute(Operation{Action: "volume.create", Name: run + "-data"}).Facts[0]
	network := execute(Operation{Action: "network.create", Name: run + "-net", Config: json.RawMessage(`{"Internal":true}`)}).Facts[0]
	config, _ := json.Marshal(map[string]any{"Image": image, "User": "0:0", "Cmd": []string{"/bin/sh", "-c", "echo stdout-marker; echo stderr-marker >&2; echo persistent-marker >/data/proof; exec /bin/sleep 300"}, "HostConfig": map[string]any{"NetworkMode": "none", "Mounts": []map[string]any{{"Type": "volume", "Source": volume.Name, "Target": "/data"}}}})
	container := execute(Operation{Action: "container.create", Name: run + "-container", Config: config}).Facts[0]
	execute(Operation{Action: "container.start", Target: container.Target})
	var output strings.Builder
	deadline := time.Now().Add(5 * time.Second)
	for {
		output.Reset()
		err = m.Logs(ctx, "admin", container.Target, LogOptions{Tail: 100}, func(f LogFrame) error { output.WriteString(f.Stream + ":" + string(f.Data)); return nil })
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(output.String(), "stdout-marker") && strings.Contains(output.String(), "stderr-marker") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("logs absent", output.String())
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Log("real timestamped stdout/stderr log frames received")
	execute(Operation{Action: "container.restart", Target: container.Target, StopSeconds: 1})
	execute(Operation{Action: "container.stop", Target: container.Target, StopSeconds: 1})
	// Create an independently labelled derivative image so image deletion never
	// touches the shared prebuilt fixture or a user image.
	var derived struct {
		ID string `json:"Id"`
	}
	if err = m.engine.api(ctx, "POST", "/commit?container="+container.Target.ID+"&repo="+run+"&tag=test", map[string]any{"Labels": map[string]string{"blora.dev/e2e": run}}, &derived); err != nil {
		t.Fatal(err)
	}
	execute(Operation{Action: "container.delete", Target: container.Target})
	if _, err = m.Query(ctx, "admin", Query{Kind: "volume", Target: volume.Target}); err != nil {
		t.Fatal("container deletion removed persistent data", err)
	}
	// Read the retained bytes with another owned container, without host mounts.
	checkConfig, _ := json.Marshal(map[string]any{"Image": image, "Cmd": []string{"/bin/sh", "-c", "cat /data/proof; exec /bin/sleep 300"}, "HostConfig": map[string]any{"NetworkMode": "none", "Mounts": []map[string]any{{"Type": "volume", "Source": volume.Name, "Target": "/data", "ReadOnly": true}}}})
	check := execute(Operation{Action: "container.create", Name: run + "-check", Config: checkConfig}).Facts[0]
	execute(Operation{Action: "container.start", Target: check.Target})
	output.Reset()
	deadline = time.Now().Add(5 * time.Second)
	for {
		err = m.Logs(ctx, "admin", check.Target, LogOptions{Tail: 100}, func(f LogFrame) error { output.Write(f.Data); return nil })
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(output.String(), "persistent-marker") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("retained volume bytes absent")
		}
		time.Sleep(25 * time.Millisecond)
	}
	execute(Operation{Action: "container.stop", Target: check.Target, StopSeconds: 1})
	execute(Operation{Action: "container.delete", Target: check.Target})
	execute(Operation{Action: "image.delete", Target: Target{Kind: "image", ID: derived.ID}})
	execute(Operation{Action: "network.delete", Target: network.Target})
	execute(Operation{Action: "volume.delete", Target: volume.Target})
	// A loopback registry serves a new, labelled zero-layer OCI image. This
	// verifies a real Engine pull without reaching an external registry.
	registryConfig, _ := json.Marshal(map[string]any{"architecture": "amd64", "os": "linux", "config": map[string]any{"Labels": map[string]string{"blora.dev/e2e": run}}, "rootfs": map[string]any{"type": "layers", "diff_ids": []string{}}, "history": []any{}})
	configDigest := "sha256:" + sourceHash(registryConfig)
	manifest, _ := json.Marshal(map[string]any{"schemaVersion": 2, "mediaType": "application/vnd.docker.distribution.manifest.v2+json", "config": map[string]any{"mediaType": "application/vnd.docker.container.image.v1+json", "size": len(registryConfig), "digest": configDigest}, "layers": []any{}})
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Docker-Distribution-Api-Version", "registry/2.0")
		switch {
		case r.URL.Path == "/v2/":
			w.Write([]byte("{}"))
		case strings.Contains(r.URL.Path, "/manifests/"):
			w.Header().Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
			w.Header().Set("Docker-Content-Digest", "sha256:"+sourceHash(manifest))
			w.Write(manifest)
		case strings.HasSuffix(r.URL.Path, "/blobs/"+configDigest):
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(registryConfig)
		default:
			http.NotFound(w, r)
		}
	}))
	defer registry.Close()
	pulled := execute(Operation{Action: "image.pull", Image: strings.TrimPrefix(registry.URL, "http://") + "/" + run + ":test"})
	if len(pulled.Facts) != 1 || pulled.Facts[0].Labels["blora.dev/e2e"] != run {
		t.Fatal("pulled image identity", pulled)
	}
	execute(Operation{Action: "image.delete", Target: pulled.Facts[0].Target})
	sequence++
	failed, err := m.Execute(ctx, "admin", Operation{TaskID: fmt.Sprintf("%s_%d", run, sequence), Action: "image.pull", Image: "127.0.0.1:1/" + run + ":missing"}, nil)
	if err == nil || failed.State == "SUCCEEDED" {
		t.Fatal("pull failure reported success", failed, err)
	}
	t.Logf("real unavailable registry failure: state=%s unknown=%v", failed.State, failed.Unknown)
	source := func(marker string) string {
		return fmt.Sprintf("services:\n  app:\n    image: %s\n    user: '0:0'\n    command: ['/bin/sh', '-c', 'echo %s; echo data-proof > /data/proof; exec /bin/sleep 300']\n    healthcheck:\n      test: ['CMD', '/bin/test', '-f', '/data/proof']\n      interval: 1s\n      timeout: 1s\n      retries: 3\n    volumes:\n      - data:/data\nvolumes:\n  data: {}\n", image, marker)
	}
	p, err := m.SaveProject(ctx, "admin", run, 0, source("compose-one"))
	if err != nil {
		t.Fatal(err)
	}
	apply := execute(Operation{Action: "compose.apply", ProjectID: p.ID, Revision: p.Revision, PullPolicy: "never", HealthSeconds: 10})
	if len(apply.Facts) == 0 {
		t.Fatal("no project containers")
	}
	if len(apply.Facts) != 1 || len(apply.Targets) != 1 {
		t.Fatal("Compose stage observations duplicated the same real resource", len(apply.Facts), len(apply.Targets))
	}
	for _, phase := range []string{"validate-config", "create", "start", "health"} {
		found := false
		for _, p := range apply.Progress {
			found = found || p.Phase == phase
		}
		if !found {
			t.Fatal("missing Compose phase", phase)
		}
	}
	firstID := apply.Facts[0].Target.ID
	p, err = m.SaveProject(ctx, "admin", run, p.Revision, source("compose-two"))
	if err != nil {
		t.Fatal(err)
	}
	apply = execute(Operation{Action: "compose.apply", ProjectID: p.ID, Revision: p.Revision, PullPolicy: "never", HealthSeconds: 10})
	if apply.Facts[0].Target.ID == firstID {
		t.Fatal("updated Compose service was not recreated")
	}
	volumes, truncated, err := m.engine.list(ctx, "volumes", map[string]string{"blora.dev/e2e": run}, 100)
	if err != nil || truncated || len(volumes) != 1 {
		t.Fatal("project data volume", volumes, truncated, err)
	}
	execute(Operation{Action: "compose.delete", ProjectID: p.ID, Revision: p.Revision, StopSeconds: 1})
	for _, v := range volumes {
		if _, err = m.engine.inspect(ctx, v.Target, false); err != nil {
			t.Fatal("Compose delete did not preserve volume", err)
		}
		execute(Operation{Action: "volume.delete", Target: v.Target})
	}
	// Real health failure retains the created service and its stage facts.
	badSource := strings.ReplaceAll(source("compose-unhealthy"), "['CMD', '/bin/test', '-f', '/data/proof']", "['CMD', '/bin/false']")
	bad, err := m.SaveProject(ctx, "admin", run+"-bad", 0, badSource)
	if err != nil {
		t.Fatal(err)
	}
	sequence++
	failed, err = m.Execute(ctx, "admin", Operation{TaskID: fmt.Sprintf("%s_%d", run, sequence), Action: "compose.apply", ProjectID: bad.ID, Revision: bad.Revision, PullPolicy: "never", HealthSeconds: 2}, nil)
	if err == nil || failed.State == "SUCCEEDED" || failed.Phase != "health" || len(failed.Facts) == 0 || !failed.Changed {
		t.Fatal("health failure lost partial result", failed, err)
	}
	t.Logf("real Compose health failure: state=%s phase=%s retainedContainers=%d", failed.State, failed.Phase, len(failed.Facts))
	execute(Operation{Action: "compose.delete", ProjectID: bad.ID, Revision: bad.Revision, StopSeconds: 1})
	// Force a real kernel ENOSPC inside an isolated container tmpfs. This keeps
	// the host and Engine storage untouched while exercising Compose's persisted
	// failure/partial-resource path against a real Engine and Compose CLI.
	pressureSource := fmt.Sprintf(`services:
  app:
    image: %s
    user: '0:0'
    command: ['/bin/sh', '-c', 'set -e; dd if=/dev/zero of=/data/full bs=65536 count=32 conv=fsync; echo unexpected-write-succeeded']
    tmpfs: ['/data:rw,size=1048576']
    healthcheck:
      test: ['CMD', '/bin/test', '-f', '/data/proof']
      interval: 1s
      timeout: 1s
      retries: 2
`, image)
	pressure, err := m.SaveProject(ctx, "admin", run+"-enospc", 0, pressureSource)
	if err != nil {
		t.Fatal("save ENOSPC Compose project", err)
	}
	sequence++
	pressureTaskID := fmt.Sprintf("%s_%d", run, sequence)
	pressureFailure, pressureErr := m.Execute(ctx, "admin", Operation{TaskID: pressureTaskID, Action: "compose.apply", ProjectID: pressure.ID, Revision: pressure.Revision, PullPolicy: "never", HealthSeconds: 5}, nil)
	if pressureErr == nil || pressureFailure.State != "FAILED" || pressureFailure.Phase != "health" || pressureFailure.Unknown || !pressureFailure.Changed || len(pressureFailure.Facts) != 1 || pressureFailure.Facts[0].State != "exited" {
		t.Fatalf("real ENOSPC was not reported as a confirmed Compose failure with retained container facts: result=%+v err=%v", pressureFailure, pressureErr)
	}
	observed, err := m.Query(ctx, "admin", Query{Kind: "container", Target: pressureFailure.Facts[0].Target, Details: true})
	if err != nil || len(observed.Items) != 1 {
		t.Fatal("inspect ENOSPC container", observed, err)
	}
	var pressureInspect containerInspect
	if err = json.Unmarshal(observed.Items[0].Details, &pressureInspect); err != nil || pressureInspect.State.ExitCode == 0 {
		t.Fatalf("ENOSPC container exit status was not preserved: %+v err=%v", pressureInspect.State, err)
	}
	var pressureLogs strings.Builder
	if err = m.Logs(ctx, "admin", pressureFailure.Facts[0].Target, LogOptions{Tail: 100}, func(f LogFrame) error {
		pressureLogs.Write(f.Data)
		return nil
	}); err != nil {
		t.Fatal("read real ENOSPC diagnostics", err)
	}
	if !strings.Contains(strings.ToLower(pressureLogs.String()), "no space left on device") || strings.Contains(pressureLogs.String(), "unexpected-write-succeeded") {
		t.Fatalf("real kernel ENOSPC diagnostic missing or write unexpectedly completed: %q", pressureLogs.String())
	}
	t.Logf("real Compose tmpfs ENOSPC: state=%s phase=%s exit=%d retainedContainers=%d diagnostic=%q", pressureFailure.State, pressureFailure.Phase, pressureInspect.State.ExitCode, len(pressureFailure.Facts), pressureLogs.String())
	execute(Operation{Action: "compose.delete", ProjectID: pressure.ID, Revision: pressure.Revision, StopSeconds: 1})
	t.Log("real Docker container/image/volume/network and Compose apply/update/delete with explicit volume preservation verified")
}

package containerterm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"blora.dev/panel/internal/containers"
	"blora.dev/panel/internal/containerterm/helper"
	"blora.dev/panel/internal/model"
	run "blora.dev/panel/internal/runtime"
	"blora.dev/panel/internal/terminal"
)

func TestRealHostDockerTerminalSessionLeaseBirthAndHelperCapabilities(t *testing.T) {
	if os.Getenv("BLORA_CONTAINER_E2E") != "1" {
		t.Skip("set BLORA_CONTAINER_E2E=1 for owned Docker terminal objects")
	}
	helperBinary := os.Getenv("BLORA_HOST_EXEC_HELPER_BINARY")
	if helperBinary == "" {
		t.Fatal("BLORA_HOST_EXEC_HELPER_BINARY must name the newly built processTree helper")
	}
	if _, err := os.Stat(helperBinary); err != nil {
		t.Fatal(err)
	}
	endpoint := os.Getenv("BLORA_CONTAINER_E2E_ENDPOINT")
	if endpoint == "" {
		endpoint = "unix:///var/run/docker.sock"
	}
	image := os.Getenv("BLORA_CONTAINER_E2E_IMAGE")
	if image == "" {
		image = "blora-isolated-e2e:20260909"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	identity := "blora-host-exec-" + model.ID()[:16]
	m, err := containers.New(containers.Options{StateRoot: t.TempDir(), Endpoint: endpoint, Labels: map[string]string{"blora.dev/hostexec-e2e": identity}, Authorize: func(_ context.Context, actor string) error {
		if actor == "admin" {
			return nil
		}
		return containers.ErrForbidden
	}})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	var targets []containers.Target
	sequence := 0
	action := func(op containers.Operation) containers.Result {
		t.Helper()
		sequence++
		op.TaskID = fmt.Sprintf("%s_%d", identity, sequence)
		result, err := m.Execute(ctx, "admin", op, nil)
		if err != nil {
			t.Fatal(op.Action, err)
		}
		return result
	}
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		for _, target := range targets {
			q, err := m.Query(cleanup, "admin", containers.Query{Kind: "container", Target: target})
			if err != nil {
				t.Errorf("cleanup inspect: %v", err)
				continue
			}
			if len(q.Items) != 1 || q.Items[0].Labels["blora.dev/hostexec-e2e"] != identity {
				t.Error("cleanup target ownership differs")
				continue
			}
			for _, name := range []string{"container.stop", "container.delete"} {
				sequence++
				result, err := m.Execute(cleanup, "admin", containers.Operation{TaskID: fmt.Sprintf("%s_%d", identity, sequence), Action: name, Target: target, StopSeconds: 1}, nil)
				if err != nil || result.Unknown {
					t.Errorf("own container cleanup %s: %v", name, err)
				}
			}
		}
	})
	create := func(name string, hideHelper bool) containers.Target {
		config := map[string]any{"Image": image, "Cmd": []string{"/bin/sh", "-c", "exec /bin/sleep 300"}, "HostConfig": map[string]any{"NetworkMode": "none"}}
		if hideHelper {
			config["HostConfig"].(map[string]any)["Tmpfs"] = map[string]string{"/usr/local/lib/blora": "rw,noexec,nosuid,size=1m"}
		}
		b, _ := json.Marshal(config)
		result := action(containers.Operation{Action: "container.create", Name: identity + "-" + name, Config: b})
		target := result.Facts[0].Target
		targets = append(targets, target)
		action(containers.Operation{Action: "container.start", Target: target})
		return target
	}
	target := create("ready", false)
	copy := exec.CommandContext(ctx, "docker", "--host", endpoint, "cp", helperBinary, target.ID+":"+helper.Path)
	if output, err := copy.CombinedOutput(); err != nil {
		t.Fatalf("copy new helper only into owned container: %v %s", err, output)
	}
	execTarget, err := m.InspectExecTarget(ctx, "admin", target)
	if err != nil {
		t.Fatal(err)
	}
	fixed := run.HostDockerTarget{ContainerID: target.ID, CreatedAt: execTarget.Target.CreatedAt, StartedAt: execTarget.StartedAt}
	executor, err := run.NewHostDockerExecutor(run.HostDockerOptions{StateRoot: t.TempDir(), Endpoint: endpoint})
	if err != nil {
		t.Fatal(err)
	}
	defer executor.Close()
	var permitted atomic.Bool
	permitted.Store(true)
	authorize := func(context.Context) error {
		if permitted.Load() {
			return nil
		}
		return terminal.ErrForbidden
	}
	ref := model.ResourceRef{Kind: "node", ID: "test-node", NodeID: "test-node"}
	adapter := HostAdapter{Executor: executor, Resolve: func(_ context.Context, actor string, resource model.ResourceRef, reference string) (HostBinding, error) {
		if actor != "admin" || resource != ref || reference != fixed.Reference() {
			return HostBinding{}, terminal.ErrForbidden
		}
		return HostBinding{Target: fixed, Authorize: authorize}, nil
	}}
	terminals, err := terminal.New(terminal.Options{Root: t.TempDir(), Backends: map[string]terminal.Backend{"docker-host": adapter}, Authorize: func(ctx context.Context, actor string, resource model.ResourceRef, _ string) error {
		if actor != "admin" || resource != ref {
			return terminal.ErrForbidden
		}
		return authorize(ctx)
	}})
	if err != nil {
		t.Fatal(err)
	}
	defer terminals.Close()
	request := terminal.CreateRequest{OwnerID: "admin", Resource: ref, RunID: fixed.Reference(), Backend: "docker-host", Directory: "/", Command: []string{"/bin/sh", "-c", `stty -echo; printf 'host-ready\n'; while IFS= read -r line; do if [ "$line" = size ]; then stty size; else printf 'host:%s\n' "$line"; fi; done`}}
	session, err := terminals.Create(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	for _, view := range []string{"one", "two"} {
		if _, err = terminals.Attach(ctx, session.ID, "admin", view, 0, 0); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = terminals.AcquireLease(session.ID, "admin", "one", time.Minute, false); err != nil {
		t.Fatal(err)
	}
	if err = terminals.WriteInput(ctx, session.ID, "admin", "two", []byte("must-not-run\n")); !errors.Is(err, terminal.ErrLease) {
		t.Fatal("observer obtained input", err)
	}
	readUntil := func(id, view, marker string) string {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for {
			batch, err := terminals.Read(ctx, id, "admin", view, 0, 0)
			if err != nil {
				t.Fatal(err)
			}
			var output strings.Builder
			for _, event := range batch.Events {
				if event.Kind == "output" {
					output.Write(event.Data)
				}
			}
			if strings.Contains(output.String(), marker) {
				return output.String()
			}
			if time.Now().After(deadline) {
				t.Fatalf("missing output %q: %s", marker, output.String())
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
	readUntil(session.ID, "one", "host-ready")
	if err = terminals.Resize(ctx, session.ID, "admin", "one", 132, 42); err != nil {
		t.Fatal(err)
	}
	if err = terminals.WriteInput(ctx, session.ID, "admin", "one", []byte("size\n")); err != nil {
		t.Fatal(err)
	}
	readUntil(session.ID, "one", "42 132")
	if err = terminals.WriteInput(ctx, session.ID, "admin", "one", []byte("once\n")); err != nil {
		t.Fatal(err)
	}
	readUntil(session.ID, "one", "host:once")
	if err = terminals.Detach(session.ID, "admin", "one"); err != nil {
		t.Fatal(err)
	}
	if _, err = terminals.Attach(ctx, session.ID, "admin", "one", 0, 0); err != nil {
		t.Fatal(err)
	}
	output := readUntil(session.ID, "one", "host:once")
	if strings.Count(output, "host:once") != 1 || strings.Contains(output, "must-not-run") {
		t.Fatal("input replay or observer input", output)
	}
	second, err := terminals.Create(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = terminals.Attach(ctx, second.ID, "admin", "second", 0, 0); err != nil {
		t.Fatal(err)
	}
	readUntil(second.ID, "second", "host-ready")
	if err = terminals.CloseSession(ctx, session.ID, "admin"); err != nil {
		t.Fatal(err)
	}
	if _, err = terminals.AcquireLease(second.ID, "admin", "second", time.Minute, false); err != nil {
		t.Fatal(err)
	}
	if err = terminals.WriteInput(ctx, second.ID, "admin", "second", []byte("still-alive\n")); err != nil {
		t.Fatal(err)
	}
	readUntil(second.ID, "second", "host:still-alive")
	permitted.Store(false)
	if err = terminals.WriteInput(ctx, second.ID, "admin", "second", []byte("revoked\n")); !errors.Is(err, terminal.ErrForbidden) {
		t.Fatal("revoked input accepted", err)
	}
	// Maintenance retains authority to close only its previously owned helper.
	if err = terminals.Close(); err != nil {
		t.Fatal("revocation prevented owned-session cleanup", err)
	}
	permitted.Store(true)
	if snapshot, err := m.Query(ctx, "admin", containers.Query{Kind: "container", Target: target}); err != nil || snapshot.Items[0].State != "running" {
		t.Fatal("closing terminal stopped host container", err)
	}
	changed := fixed
	changed.StartedAt = "2000-01-01T00:00:00Z"
	if _, err = executor.Start(ctx, changed, run.DockerExecOptions{SessionID: model.ID(), Command: []string{"/bin/sh"}}, authorize); !errors.Is(err, run.ErrUnknown) {
		t.Fatal("changed container birth accepted", err)
	}
	for _, item := range []struct {
		name string
		hide bool
	}{{"old-helper", false}, {"no-helper", true}} {
		missing := create(item.name, item.hide)
		checked, err := m.InspectExecTarget(ctx, "admin", missing)
		if err != nil {
			t.Fatal(err)
		}
		bound := run.HostDockerTarget{ContainerID: missing.ID, CreatedAt: checked.Target.CreatedAt, StartedAt: checked.StartedAt}
		if _, err = executor.Start(ctx, bound, run.DockerExecOptions{SessionID: model.ID(), Command: []string{"/bin/sh", "-c", "echo must-not-launch"}}, authorize); err == nil {
			t.Fatal("missing required helper capability accepted", item.name)
		} else {
			t.Logf("%s rejected: %v", item.name, err)
		}
	}
	t.Log("real host Docker session, lease, resize, archive reattach, isolated close, birth and helper dependencies verified")
}

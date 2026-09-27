//go:build linux

package containers

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestComposeCheckpointCrashHelper(t *testing.T) {
	root := os.Getenv("BLORA_TEST_COMPOSE_CRASH_ROOT")
	if root == "" {
		return
	}
	m := saveTestManager(t, root)
	m.options.Endpoint = "unix:///tmp/blora-output-test-unused.sock"
	m.options.ComposeCommand = []string{os.Args[0], "-test.run=^TestComposeOutputProcessHelper$", "--"}
	p := Project{ID: "crash-project", EngineName: "crash-fixture"}
	dir, err := m.projectDir(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	op := Operation{TaskID: "compose_crash", Action: "compose.apply", ProjectID: p.ID}
	digest := fingerprint(struct {
		Actor     string
		Operation Operation
	}{"admin", op})
	x := execution{manager: m, actor: "admin", ctx: context.Background(), entry: journal{ActorID: "admin", Digest: digest, Operation: op, Result: Result{TaskID: op.TaskID, State: "RUNNING"}}}
	if err := x.composeStage(context.Background(), p, "unused.yml", "build", []string{"identity-output=" + filepath.Join(root, "cli-identity"), "wait-output"}); err != nil {
		t.Fatal(err)
	}
}

func TestComposeOutputSurvivesExecutorSIGKILL(t *testing.T) {
	t.Run("session-recorded", func(t *testing.T) { testComposeExecutorCrash(t, false) })
	t.Run("before-session-recorded", func(t *testing.T) { testComposeExecutorCrash(t, true) })
}

func testComposeExecutorCrash(t *testing.T, beforeSession bool) {
	root := t.TempDir()
	reader := saveTestManager(t, root)
	cmd := exec.Command(os.Args[0], "-test.run=^TestComposeCheckpointCrashHelper$")
	cmd.Env = append(os.Environ(), "BLORA_TEST_COMPOSE_CRASH_ROOT="+root)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	waited := false
	defer func() {
		if !waited {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()
	// The production CLI session marker lets cleanup target only this test's CLI,
	// even if the executor has already died. Never kill a bare reused PID.
	defer func() {
		b, err := os.ReadFile(filepath.Join(root, "cli-identity"))
		if err != nil {
			t.Error(err)
			return
		}
		var identity struct {
			PID    int
			Marker string
		}
		if err := json.Unmarshal(b, &identity); err != nil {
			t.Error(err)
			return
		}
		if identity.PID <= 1 || identity.Marker == "" {
			t.Error("missing CLI identity")
			return
		}
		if err := terminateCLISession(identity.PID, identity.Marker); err != nil {
			t.Error(err)
		}
	}()
	deadline := time.Now().Add(10 * time.Second)
	var before CommandOutput
	for {
		j, err := reader.loadOperation("compose_crash")
		if err == nil && len(j.Outputs) == 1 && string(j.Outputs[0].Stderr) == "diagnostic 中文\n" {
			before = j.Outputs[0]
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("no durable checkpoint before crash")
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err == nil {
		t.Fatal("executor did not fail")
	}
	waited = true
	reader = saveTestManager(t, root)
	if beforeSession {
		entries, err := os.ReadDir(filepath.Join(root, "cli-runs"))
		if err != nil || len(entries) != 1 {
			t.Fatal("missing ownership record")
		}
		marker := entries[0].Name()
		if err := atomicJSON(filepath.Join(root, "cli-runs", marker), cliRunRecord{Marker: marker}); err != nil {
			t.Fatal(err)
		}
	}
	if err := reader.RecoverCLI(context.Background()); err != nil {
		t.Fatal(err)
	}
	identityBytes, err := os.ReadFile(filepath.Join(root, "cli-identity"))
	if err != nil {
		t.Fatal(err)
	}
	var identity struct {
		PID    int
		Marker string
	}
	if err := json.Unmarshal(identityBytes, &identity); err != nil {
		t.Fatal(err)
	}
	_, _, state, birthErr := cliBirth(identity.PID)
	if birthErr == nil && state != 'Z' && state != 'X' {
		t.Fatal("recovered CLI still alive")
	}
	records, err := os.ReadDir(filepath.Join(root, "cli-runs"))
	if err != nil || len(records) != 0 {
		t.Fatal("confirmed recovery did not clear ownership records")
	}
	if err := reader.RecoverCLI(context.Background()); err != nil {
		t.Fatal("recovery not idempotent", err)
	}
	page, err := reader.Query(context.Background(), "admin", Query{Kind: "operation-output", TaskID: "compose_crash"})
	if err != nil || page.Output == nil {
		t.Fatalf("checkpoint lost: %v", err)
	}
	a, _ := json.Marshal(before)
	b, _ := json.Marshal(page.Output)
	if string(a) != string(b) || page.Output.Completed {
		t.Fatal("crash changed checkpoint or falsely completed output")
	}
	operation, err := reader.Query(context.Background(), "admin", Query{Kind: "operation", TaskID: "compose_crash"})
	if err != nil || operation.Operation.State != "INTERRUPTED" || !operation.Operation.Unknown {
		t.Fatal("crash not presented as interrupted and unknown")
	}
	_, err = reader.Execute(context.Background(), "admin", Operation{TaskID: "compose_crash", Action: "compose.apply", ProjectID: "crash-project"}, nil)
	if !errors.Is(err, ErrUnknown) {
		t.Fatalf("interrupted operation replay not refused: %v", err)
	}
}

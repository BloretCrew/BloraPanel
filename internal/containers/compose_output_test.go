package containers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestComposeOutputProcessHelper(t *testing.T) {
	enabled := false
	failed := false
	for _, arg := range os.Args {
		if arg == "--" {
			enabled = true
		}
		if arg == "fail-output" {
			failed = true
		}
	}
	if !enabled {
		return
	}
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "identity-output=") {
			b, _ := json.Marshal(struct {
				PID    int
				Marker string
			}{os.Getpid(), os.Getenv("BLORA_CONTAINER_CLI_TOKEN")})
			if err := os.WriteFile(strings.TrimPrefix(arg, "identity-output="), b, 0600); err != nil {
				os.Exit(9)
			}
		}
	}
	os.Stdout.Write([]byte(strings.Repeat("x", 40<<10)))
	if failed {
		os.Stderr.Write([]byte(strings.Repeat("\x00", 40<<10)))
		os.Exit(7)
	}
	os.Stderr.Write([]byte("diagnostic 中文\n"))
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "release-output=") {
			for {
				if _, err := os.Stat(strings.TrimPrefix(arg, "release-output=")); err == nil {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
			os.Stderr.Write([]byte("released\n"))
			for {
				time.Sleep(time.Second)
			}
		}
		if arg == "wait-output" {
			for {
				time.Sleep(time.Second)
			}
		}
	}
	os.Exit(0)
}

func TestComposeOutputCheckpointFailureStopsCLI(t *testing.T) {
	m := saveTestManager(t, t.TempDir())
	m.options.Endpoint = "unix:///tmp/blora-output-test-unused.sock"
	m.options.ComposeCommand = []string{os.Args[0], "-test.run=^TestComposeOutputProcessHelper$", "--"}
	p := Project{ID: "checkpoint-failure", EngineName: "checkpoint-failure"}
	dir, err := m.projectDir(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	release := filepath.Join(dir, "release")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	x := execution{manager: m, actor: "admin", ctx: ctx, entry: journal{Result: Result{TaskID: "checkpoint_failure", State: "RUNNING"}}}
	done := make(chan error, 1)
	go func() { done <- x.composeStage(ctx, p, "unused.yml", "build", []string{"release-output=" + release}) }()
	defer func() { cancel(); <-done }()
	for {
		j, err := m.loadOperation("checkpoint_failure")
		if err == nil && len(j.Outputs) == 1 && len(j.Outputs[0].Stderr) != 0 {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("checkpoint not observed")
		case <-time.After(20 * time.Millisecond):
		}
	}
	// Keep the last durable record intact, but make subsequent atomic writes fail.
	operations := filepath.Join(m.options.StateRoot, "operations")
	backup := operations + ".saved"
	if err := os.Rename(operations, backup); err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.Remove(operations)
		if err := os.Rename(backup, operations); err != nil {
			t.Error(err)
		}
	}()
	if err := os.WriteFile(operations, []byte("injected failure"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(release, nil, 0600); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		done <- err
		if err == nil || !strings.Contains(err.Error(), "output checkpoint failed") {
			t.Fatalf("wrong failure: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("CLI not stopped after checkpoint failure")
	}
	b, err := os.ReadFile(filepath.Join(backup, "checkpoint_failure.json"))
	if err != nil {
		t.Fatal(err)
	}
	var j journal
	if err := json.Unmarshal(b, &j); err != nil {
		t.Fatal(err)
	}
	if !j.Result.Unknown || j.Outputs[0].Completed || string(j.Outputs[0].Stderr) != "diagnostic 中文\n" {
		t.Fatal("last durable checkpoint corrupted or falsely completed")
	}
}

func TestComposeOutputCheckpointBeforeExit(t *testing.T) {
	m := saveTestManager(t, t.TempDir())
	m.options.Endpoint = "unix:///tmp/blora-output-test-unused.sock"
	m.options.ComposeCommand = []string{os.Args[0], "-test.run=^TestComposeOutputProcessHelper$", "--"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/version" {
			json.NewEncoder(w).Encode(map[string]string{"ApiVersion": "1.54", "MinAPIVersion": "1.40", "Os": "linux"})
			return
		}
		json.NewEncoder(w).Encode([]any{})
	}))
	defer server.Close()
	m.engine = &engine{client: server.Client(), base: server.URL}
	p := Project{ID: "checkpoint-project", EngineName: "checkpoint-fixture"}
	dir, err := m.projectDir(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	x := execution{manager: m, actor: "admin", ctx: ctx, entry: journal{Result: Result{TaskID: "compose_checkpoint", State: "RUNNING"}}}
	done := make(chan error, 1)
	go func() { done <- x.composeStage(ctx, p, "unused.yml", "build", []string{"wait-output"}) }()
	defer func() { cancel(); <-done }()
	// Read through a separate manager while the original CLI is still blocked.
	reader := saveTestManager(t, m.options.StateRoot)
	for {
		page, err := reader.Query(context.Background(), "admin", Query{Kind: "operation-output", TaskID: "compose_checkpoint"})
		if err == nil && page.Output != nil && len(page.Output.Stdout) == 32<<10 && string(page.Output.Stderr) == "diagnostic 中文\n" {
			if page.Output.Completed || !page.Output.StdoutTruncated {
				t.Fatal("checkpoint falsely completed or unbounded")
			}
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("no output checkpoint before CLI exit")
		case <-time.After(20 * time.Millisecond):
		}
	}
	select {
	case err := <-done:
		done <- err
		t.Fatal("helper exited before cancellation")
	default:
	}
	cancel()
	// Deferred cleanup receives completion; inspect the final record after it.
	t.Cleanup(func() {
		j, err := reader.loadOperation("compose_checkpoint")
		if err != nil || !j.Result.Unknown || len(j.Outputs) != 1 || !j.Outputs[0].Completed || j.Outputs[0].CommandError == "" {
			t.Fatalf("cancelled output lost uncertainty or completion: %v", err)
		}
	})
}

func TestComposeCommandOutputPersistsOnSuccessAndFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("starts controlled CLI helper")
	}
	m := saveTestManager(t, t.TempDir())
	m.options.Endpoint = "unix:///tmp/blora-output-test-unused.sock"
	m.options.ComposeCommand = []string{os.Args[0], "-test.run=^TestComposeOutputProcessHelper$", "--"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/version" {
			json.NewEncoder(w).Encode(map[string]string{"ApiVersion": "1.54", "MinAPIVersion": "1.40", "Os": "linux"})
			return
		}
		json.NewEncoder(w).Encode([]any{})
	}))
	defer server.Close()
	m.engine = &engine{client: server.Client(), base: server.URL}
	p := Project{ID: "output-project", EngineName: "output-fixture"}
	dir, err := m.projectDir(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	x := execution{manager: m, actor: "admin", ctx: context.Background(), entry: journal{Result: Result{TaskID: "compose_output", State: "RUNNING"}}}
	if err := x.composeStage(context.Background(), p, "unused.yml", "build", []string{"success-output"}); err != nil {
		t.Fatal(err)
	}
	if err := x.composeStage(context.Background(), p, "unused.yml", "start", []string{"fail-output"}); err == nil {
		t.Fatal("failed CLI reported success")
	}
	// A fresh manager reads persisted bytes without relying on executor memory.
	m = saveTestManager(t, m.options.StateRoot)
	for index := 0; index < 2; index++ {
		result, err := m.Query(context.Background(), "admin", Query{Kind: "operation-output", TaskID: "compose_output", OutputOffset: index})
		if err != nil {
			t.Fatal(err)
		}
		output := result.Output
		if output == nil || !output.Completed || len(output.Stdout) != 32<<10 || !output.StdoutTruncated {
			t.Fatal("incorrect persisted output bounds")
		}
		if index == 0 && (output.StderrTruncated || string(output.Stderr) != "diagnostic 中文\n") {
			t.Fatal("short stderr changed")
		}
		if index == 1 && (!output.StderrTruncated || len(output.Stderr) != 32<<10) {
			t.Fatal("binary stderr budget not enforced")
		}
		if (index == 1) != (output.CommandError != "") {
			t.Fatal("CLI status not retained")
		}
		encoded, err := json.Marshal(result)
		if err != nil || len(encoded) > 96<<10 {
			t.Fatal("query exceeds transport budget")
		}
		if index == 0 && result.NextOutput != 1 || index == 1 && result.NextOutput != -1 {
			t.Fatal("incorrect output pagination")
		}
	}
	if _, err := m.Query(context.Background(), "denied", Query{Kind: "operation-output", TaskID: "compose_output"}); err == nil {
		t.Fatal("output authorization bypass")
	}
	if _, err := m.Query(context.Background(), "admin", Query{Kind: "operation-output", TaskID: "compose_output", OutputOffset: -1}); err == nil {
		t.Fatal("negative page accepted")
	}
	record, err := m.loadOperation("compose_output")
	if err != nil || len(record.Outputs) != 2 {
		t.Fatal("output not durable")
	}
}

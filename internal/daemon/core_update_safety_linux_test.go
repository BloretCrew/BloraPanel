//go:build linux

package daemon

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/runlog"
	run "blora.dev/panel/internal/runtime"
)

func TestDaemonCoreUpdateLogHelperProcess(t *testing.T) {
	for index, arg := range os.Args {
		if arg == "--runlog-root" {
			if err := runlog.RunHelper(os.Args[index:]); err != nil {
				os.Exit(81)
			}
			os.Exit(0)
		}
	}
}

func TestCoreRestartPreservesNativeBirthAndIndependentInputOutput(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	d := coreUpdateTestDaemon(t)
	config := d.config
	d.config.LogHelperCommand = []string{os.Args[0], "-test.run=^TestDaemonCoreUpdateLogHelperProcess$", "--"}
	i := model.Instance{ID: model.ID(), NodeID: d.identity.NodeID, RunID: model.ID(), State: "RUNNING", Config: model.InstanceConfig{Mode: "native", Command: []string{"/bin/sh", "-c", "while IFS= read -r line; do printf '%s\\n' \"$line\"; done"}}}
	capture, err := d.startLog(ctx, i)
	if err != nil {
		t.Fatal(err)
	}
	r, err := d.runtime.Start(ctx, i.RunID, i.ID, i.Config, run.IO{Stdin: capture.Stdin(), Stdout: capture.Stdout(), Stderr: capture.Stdout()})
	_ = capture.ReleaseWriter()
	_ = capture.ReleaseInputReader()
	if err != nil {
		_ = capture.Finish(ctx)
		t.Fatal(err)
	}
	manager := d.runtime
	t.Cleanup(func() {
		cleanup, stop := context.WithTimeout(context.Background(), 8*time.Second)
		defer stop()
		observation, err := manager.Stop(cleanup, r, run.StopPolicy{Force: true, KillWait: 4 * time.Second}, nil)
		if err != nil || !observation.Exited {
			t.Errorf("cleanup scoped run: %+v %v", observation, err)
			return
		}
		if err := capture.Finish(cleanup); err != nil {
			t.Error(err)
		}
		if err := manager.Cleanup(cleanup, r); err != nil {
			t.Error(err)
		}
	})
	if err := d.saveInstance(ctx, &i); err != nil {
		t.Fatal(err)
	}
	if err := d.PrepareCoreRestart(ctx); err != nil {
		t.Fatal(err)
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
	replacement, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := replacement.Close(); err != nil {
			t.Error(err)
		}
	})
	if replacement.identity.NodeID != i.NodeID {
		t.Fatal("restart changed node identity")
	}
	recovered, observed, err := replacement.runtime.Recover(ctx, i.RunID)
	if err != nil || observed.State != "RUNNING" || recovered.PID != r.PID || recovered.StartTicks != r.StartTicks || recovered.PGID != r.PGID || recovered.Token != r.Token {
		t.Fatalf("replacement did not recover same original business process: %+v %+v %v", recovered, observed, err)
	}
	if _, err := replacement.runtime.WriteInput(ctx, recovered, []byte("still-alive-after-core-restart\n")); err != nil {
		t.Fatal(err)
	}
	c, err := runlog.Open(filepath.Join(config.StateDir, "logs"), i.RunID)
	if err != nil {
		t.Fatal(err)
	}
	for {
		batch, err := c.Read(ctx, 0, 64<<10)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, chunk := range batch.Events {
			if strings.Contains(string(chunk.Data), "still-alive-after-core-restart") {
				found = true
			}
		}
		if found {
			break
		}
		if ctx.Err() != nil {
			t.Fatal("live output was not preserved after replacement")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := replacement.PrepareCoreRestart(ctx); err != nil {
		t.Fatal(err)
	}
	// Suppressing the helper's durable identity must prevent a subsequent
	// restart; checking only the instance PID would incorrectly accept it.
	replacement.AbortCoreRestart()
	if err := os.Rename(filepath.Join(config.StateDir, "logs", i.RunID, "capture.json"), filepath.Join(config.StateDir, "logs", i.RunID, "capture.hidden")); err != nil {
		t.Fatal(err)
	}
	err = replacement.PrepareCoreRestart(ctx)
	if restore := os.Rename(filepath.Join(config.StateDir, "logs", i.RunID, "capture.hidden"), filepath.Join(config.StateDir, "logs", i.RunID, "capture.json")); restore != nil {
		t.Fatal(restore)
	}
	if err == nil || !strings.Contains(err.Error(), "helper") {
		t.Fatalf("accepted unverified helper: %v", err)
	}
}

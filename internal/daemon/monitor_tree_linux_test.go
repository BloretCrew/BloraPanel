package daemon

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/monitor"
	run "blora.dev/panel/internal/runtime"
	"golang.org/x/sys/unix"
)

func TestNativeMetricTreeRejectsChangedMembershipAndRetries(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	root := t.TempDir()
	fifo := filepath.Join(root, "add-child")
	if err := unix.Mkfifo(fifo, 0600); err != nil {
		t.Fatal(err)
	}
	m, err := run.New(run.Options{StateRoot: filepath.Join(root, "runs"), AllowPGIDFallback: true})
	if err != nil {
		t.Fatal(err)
	}
	r, err := m.Start(ctx, model.ID(), model.ID(), model.InstanceConfig{Mode: "native", Command: []string{"/bin/sh", "-c", "sleep 30 & read signal < \"$1\"; sleep 30 & wait", "test", fifo}}, run.IO{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		if _, err := m.Stop(cleanup, r, run.StopPolicy{Force: true, KillWait: 5 * time.Second}, nil); err != nil {
			t.Error(err)
		}
		if err := m.Cleanup(cleanup, r); err != nil {
			t.Error(err)
		}
	})
	var before run.Observation
	for {
		before, err = m.Observe(ctx, r)
		if err != nil {
			t.Fatal(err)
		}
		if len(before.Processes) == 2 {
			break
		}
		if ctx.Err() != nil {
			t.Fatal(ctx.Err())
		}
		time.Sleep(10 * time.Millisecond)
	}
	trigger, err := os.OpenFile(fifo, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = trigger.WriteString("go\n")
	trigger.Close()
	if err != nil {
		t.Fatal(err)
	}
	for {
		after, err := m.Observe(ctx, r)
		if err != nil {
			t.Fatal(err)
		}
		if len(after.Processes) == 3 {
			break
		}
		if ctx.Err() != nil {
			t.Fatal(ctx.Err())
		}
		time.Sleep(10 * time.Millisecond)
	}
	d := &Daemon{runtime: m, metrics: monitor.New()}
	if _, err := d.sampleNativeRunMetrics(ctx, r, before); !errors.Is(err, run.ErrUnknown) {
		t.Fatalf("stale member snapshot accepted: %v", err)
	}
	point, err := d.nativeRunMetrics(ctx, r, before)
	if err != nil || point.ProcessCount != 3 || point.Scope != "process-tree" {
		t.Fatalf("retry failed: %+v %v", point, err)
	}
}

package daemon

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
)

func coreUpdateTestDaemon(t *testing.T) *Daemon {
	t.Helper()
	root := t.TempDir()
	_, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(identity{NodeID: model.ID(), PrivateKey: private})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "identity.json"), encoded, 0600); err != nil {
		t.Fatal(err)
	}
	d, err := New(Config{StateDir: root, MasterURL: "http://127.0.0.1:37861", AllowPGIDFallback: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := d.Close(); err != nil {
			t.Error(err)
		}
	})
	return d
}

func TestPrepareCoreRestartDrainsWithoutCancellingOperations(t *testing.T) {
	d := coreUpdateTestDaemon(t)
	var cancelled atomic.Bool
	d.active["existing-instance-stop"] = func() { cancelled.Store(true) }
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- d.PrepareCoreRestart(ctx) }()
	deadline := time.Now().Add(time.Second)
	for {
		d.mu.Lock()
		updating := d.coreUpdating
		d.mu.Unlock()
		if updating {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("preparation did not close acceptance gate")
		}
		time.Sleep(time.Millisecond)
	}
	select {
	case err := <-result:
		t.Fatalf("did not drain active operation: %v", err)
	default:
	}
	if err := d.PrepareCoreRestart(ctx); err == nil {
		t.Fatal("second concurrent preparation accepted")
	}
	d.mu.Lock()
	delete(d.active, "existing-instance-stop")
	d.mu.Unlock()
	if err := <-result; err != nil {
		t.Fatal(err)
	}
	if cancelled.Load() {
		t.Fatal("update cancelled a pre-existing operation")
	}
	d.mu.Lock()
	updating := d.coreUpdating
	d.mu.Unlock()
	if !updating {
		t.Fatal("successful preparation released gate before shutdown")
	}
	d.AbortCoreRestart()
}

func TestPrepareCoreRestartTimeoutResumesAndNeverCancels(t *testing.T) {
	d := coreUpdateTestDaemon(t)
	var cancelled atomic.Bool
	d.active["existing-start"] = func() { cancelled.Store(true) }
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := d.PrepareCoreRestart(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timeout: %v", err)
	}
	d.mu.Lock()
	updating := d.coreUpdating
	delete(d.active, "existing-start")
	d.mu.Unlock()
	if updating || cancelled.Load() {
		t.Fatal("timeout froze acceptance or cancelled existing operation")
	}
}

func TestPrepareCoreRestartBlocksUnreconciledExecutedTask(t *testing.T) {
	d := coreUpdateTestDaemon(t)
	ctx := context.Background()
	task := model.Task{ID: model.ID(), RequestID: model.ID(), ActorID: "actor", Resource: model.ResourceRef{Kind: "instance", ID: model.ID(), NodeID: d.identity.NodeID}, Action: "instance.restart", Payload: []byte("{}"), Digest: "fixed-digest", Deadline: time.Now().Add(time.Minute)}
	accepted, _, err := d.store.Accept(ctx, task)
	if err != nil {
		t.Fatal(err)
	}
	// An accepted queued task is safe to reconcile after reconnecting.
	if err := d.PrepareCoreRestart(ctx); err != nil {
		t.Fatal(err)
	}
	d.schedule(ctx, accepted)
	d.mu.Lock()
	active := len(d.active)
	d.mu.Unlock()
	stillQueued, err := d.store.Task(ctx, task.ID)
	if err != nil || active != 0 || stillQueued.State != model.Queued {
		t.Fatalf("restart gate executed accepted operation: active=%d task=%+v err=%v", active, stillQueued, err)
	}
	d.AbortCoreRestart()
	if _, err := d.store.UpdateTask(ctx, task.ID, accepted.Revision, model.Running, "STOPPING", nil, ""); err != nil {
		t.Fatal(err)
	}
	if err := d.PrepareCoreRestart(ctx); err == nil || !strings.Contains(err.Error(), "unreconciled") {
		t.Fatalf("accepted lost task outcome: %v", err)
	}
	d.mu.Lock()
	updating := d.coreUpdating
	d.mu.Unlock()
	if updating {
		t.Fatal("failed check retained acceptance gate")
	}
}

func TestPrepareCoreRestartRejectsMissingRunIdentity(t *testing.T) {
	d := coreUpdateTestDaemon(t)
	i := model.Instance{ID: model.ID(), NodeID: d.identity.NodeID, State: "RUNNING"}
	if err := d.saveInstance(context.Background(), &i); err != nil {
		t.Fatal(err)
	}
	if err := d.PrepareCoreRestart(context.Background()); err == nil || !strings.Contains(err.Error(), "no recoverable run identity") {
		t.Fatalf("accepted unowned running instance: %v", err)
	}
}

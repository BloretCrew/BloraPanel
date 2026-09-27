package containerterm

import (
	"context"
	"errors"
	"testing"

	"blora.dev/panel/internal/model"
	run "blora.dev/panel/internal/runtime"
	"blora.dev/panel/internal/terminal"
)

func TestAdapterRejectsUnboundContainerAndChangedRun(t *testing.T) {
	m, err := run.New(run.Options{StateRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	called := false
	a := Adapter{Runtime: m, Resolve: func(context.Context, model.ResourceRef, string) (run.Record, error) {
		called = true
		return run.Record{InstanceID: "other", RunID: "run", Backend: "docker", ContainerID: "container"}, nil
	}}
	if _, err = a.Spawn(context.Background(), terminal.SpawnSpec{Resource: model.ResourceRef{Kind: "container", ID: "arbitrary", NodeID: "node"}, RunID: "run"}); !errors.Is(err, terminal.ErrForbidden) || called {
		t.Fatal("browser-selected container bypassed resolver")
	}
	if _, err = a.Spawn(context.Background(), terminal.SpawnSpec{Resource: model.ResourceRef{Kind: "instance", ID: "expected", NodeID: "node"}, RunID: "run"}); !errors.Is(err, terminal.ErrForbidden) || !called {
		t.Fatal("changed instance run accepted")
	}
}

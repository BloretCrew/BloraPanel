package daemon

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/runlog"
	"blora.dev/panel/internal/terminal"
)

func (d *Daemon) startLog(ctx context.Context, i model.Instance) (*runlog.Capture, error) {
	command := d.config.LogHelperCommand
	if len(command) == 0 {
		executable, err := os.Executable()
		if err != nil {
			return nil, err
		}
		command = []string{executable, "--log-helper"}
	}
	info := model.RunLog{RunID: i.RunID, Resource: model.ResourceRef{Kind: "instance", ID: i.ID, NodeID: i.NodeID}, Backend: "native", StartedAt: time.Now().UTC()}
	if _, err := d.store.PutRecord(ctx, "run_log", i.RunID, 0, info); err != nil {
		return nil, err
	}
	return runlog.Start(ctx, runlog.Options{Root: filepath.Join(d.config.StateDir, "logs"), RunID: i.RunID, Command: command, HoldInput: true, Archive: terminal.ArchiveOptions{MaxBytes: 16 << 20, SegmentBytes: 1 << 20}})
}

func (d *Daemon) finishLog(runID string) {
	var info model.RunLog
	rev, err := d.store.Record(context.Background(), "run_log", runID, &info)
	if errors.Is(err, sql.ErrNoRows) {
		return
	}
	if err != nil {
		return
	}
	capture, err := runlog.Open(filepath.Join(d.config.StateDir, "logs"), runID)
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		err = capture.Finish(ctx)
		cancel()
	}
	if err != nil {
		info.Diagnostic = err.Error()
		_, _ = d.store.PutRecord(context.Background(), "run_log", runID, rev, info)
	}
}

func (d *Daemon) logRPC(ctx context.Context, r bridge.Request) (any, error) {
	if r.Resource.Kind != "instance" || r.Resource.NodeID != d.identity.NodeID {
		return nil, terminal.ErrForbidden
	}
	if r.Method == "log.list" {
		ids, err := d.store.RecordIDs(ctx, "run_log")
		if err != nil {
			return nil, err
		}
		items := []model.RunLog{}
		for n := len(ids) - 1; n >= 0; n-- {
			var info model.RunLog
			if _, err := d.store.Record(ctx, "run_log", ids[n], &info); err != nil {
				return nil, err
			}
			if info.Resource == r.Resource {
				items = append(items, info)
				sort.Slice(items, func(i, j int) bool { return items[i].StartedAt.After(items[j].StartedAt) })
				if len(items) > 100 {
					items = items[:100]
				}
			}
		}
		return map[string]any{"items": items}, nil
	}
	var args struct {
		RunID    string `json:"runId"`
		After    uint64 `json:"after"`
		MaxBytes int    `json:"maxBytes"`
	}
	if err := json.Unmarshal(r.Args, &args); err != nil {
		return nil, err
	}
	var info model.RunLog
	if _, err := d.store.Record(ctx, "run_log", args.RunID, &info); err != nil {
		return nil, err
	}
	if info.Resource != r.Resource {
		return nil, terminal.ErrForbidden
	}
	capture, err := runlog.Open(filepath.Join(d.config.StateDir, "logs"), args.RunID)
	if err != nil {
		return nil, err
	}
	switch r.Method {
	case "log.status":
		return capture.Status(ctx)
	case "log.read":
		return capture.Read(ctx, args.After, min(args.MaxBytes, 64<<10))
	default:
		return nil, errors.New("unsupported log RPC")
	}
}

func (d *Daemon) consoleInput(ctx context.Context, t model.Task) (any, error) {
	var input model.ConsoleInput
	if err := json.Unmarshal(t.Payload, &input); err != nil {
		return nil, err
	}
	if input.RunID == "" || len(input.Data) == 0 || len(input.Data) > 8<<10 {
		return nil, errors.New("console input requires 1–8192 bytes and a fixed run identity")
	}
	if err := d.authorizeTerminal(ctx, t.ActorID, t.Resource, "terminal.input"); err != nil {
		return nil, err
	}
	var i model.Instance
	if _, err := d.store.Record(ctx, "instance", t.Resource.ID, &i); err != nil {
		return nil, err
	}
	if i.RunID != input.RunID || i.NodeID != t.Resource.NodeID {
		return nil, errors.New("instance run changed; input was not sent")
	}
	r, observation, err := d.runtime.Recover(ctx, input.RunID)
	if err != nil {
		return nil, err
	}
	if observation.Exited {
		return nil, terminal.ErrInactive
	}
	n, err := d.runtime.WriteInput(ctx, r, []byte(input.Data))
	result := map[string]any{"runId": input.RunID, "bytesWritten": n, "deliveryOnly": true}
	if err != nil {
		return result, errors.New("console input outcome may be unknown; never automatically resend: " + err.Error())
	}
	return result, nil
}

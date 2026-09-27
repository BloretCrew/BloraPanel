package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"slices"
	"time"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/monitor"
	run "blora.dev/panel/internal/runtime"
	"blora.dev/panel/internal/terminal"
)

func (d *Daemon) monitorRPC(ctx context.Context, r bridge.Request) (any, error) {
	switch r.Method {
	case "monitor.node":
		if r.Resource.Kind != "node" || r.Resource.ID != d.identity.NodeID {
			return nil, terminal.ErrForbidden
		}
		return d.metrics.Node(d.config.StateDir)
	case "monitor.history":
		if r.Resource.Kind != "node" || r.Resource.ID != d.identity.NodeID {
			return nil, terminal.ErrForbidden
		}
		return map[string]any{"items": d.metrics.History(), "retention": 120}, nil
	case "monitor.processes":
		if r.Resource.Kind != "node" || r.Resource.ID != d.identity.NodeID {
			return nil, terminal.ErrForbidden
		}
		var a monitor.ProcessQuery
		if err := json.Unmarshal(r.Args, &a); err != nil {
			return nil, err
		}
		x, err := monitor.ListProcessPage(a)
		if err != nil {
			return nil, err
		}
		return x, nil
	case "monitor.instance":
		if r.Resource.Kind != "instance" {
			return nil, terminal.ErrForbidden
		}
		var args struct {
			RunID string `json:"runId"`
		}
		if err := json.Unmarshal(r.Args, &args); err != nil {
			return nil, err
		}
		var i model.Instance
		if _, err := d.store.Record(ctx, "instance", r.Resource.ID, &i); err != nil {
			return nil, err
		}
		if i.NodeID != d.identity.NodeID || i.RunID == "" || i.RunID != args.RunID {
			return nil, terminal.ErrInactive
		}
		record, observed, err := d.runtime.Recover(ctx, i.RunID)
		if err != nil {
			return nil, err
		}
		if observed.Exited {
			return nil, terminal.ErrInactive
		}
		if record.Backend == "docker" {
			return d.runtime.ContainerMetrics(ctx, record)
		}
		return d.nativeRunMetrics(ctx, record, observed)
	default:
		return nil, errors.New("unsupported monitor RPC")
	}
}

func (d *Daemon) nativeRunMetrics(ctx context.Context, record run.Record, before run.Observation) (monitor.ProcessPoint, error) {
	var point monitor.ProcessPoint
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		point, err = d.sampleNativeRunMetrics(ctx, record, before)
		if err == nil || ctx.Err() != nil {
			return point, err
		}
		if attempt < 2 {
			before, err = d.runtime.Observe(ctx, record)
			if err != nil {
				return point, err
			}
		}
	}
	return point, err
}

func (d *Daemon) sampleNativeRunMetrics(ctx context.Context, record run.Record, before run.Observation) (monitor.ProcessPoint, error) {
	p := monitor.ProcessPoint{RunID: record.RunID, PID: record.PID, Scope: "process-tree", Point: monitor.Point{ObservedAt: time.Now().UTC(), CPUBasis: "host-total", Unavailable: []string{"network", "disk"}}}
	if len(before.Processes) == 0 || len(before.Processes) > 1024 {
		return p, errors.New("owned process metric budget unavailable")
	}
	members := map[int]uint64{}
	cpuAvailable, memoryAvailable := true, true
	for _, process := range before.Processes {
		if err := ctx.Err(); err != nil {
			return p, err
		}
		if process.PID <= 0 || process.StartTicks == 0 || members[process.PID] != 0 {
			return p, run.ErrUnknown
		}
		members[process.PID] = process.StartTicks
		sample, err := d.metrics.Process(process.PID, record.RunID, process.StartTicks)
		if err != nil {
			return p, err
		}
		cpuAvailable = cpuAvailable && !slices.Contains(sample.Unavailable, "cpu")
		memoryAvailable = memoryAvailable && !slices.Contains(sample.Unavailable, "memory")
		if sample.RSS < 0 || sample.RSS > 9007199254740991-p.RSS || math.IsNaN(sample.CPUPercent) || math.IsInf(sample.CPUPercent, 0) || sample.CPUPercent < 0 {
			return p, run.ErrUnknown
		}
		p.RSS += sample.RSS
		p.CPUPercent += sample.CPUPercent
	}
	after, err := d.runtime.Observe(ctx, record)
	if err != nil {
		return p, err
	}
	if after.Exited || len(after.Processes) != len(members) {
		return p, run.ErrUnknown
	}
	seen := map[int]bool{}
	for _, process := range after.Processes {
		if seen[process.PID] || members[process.PID] != process.StartTicks || process.StartTicks == 0 {
			return p, run.ErrUnknown
		}
		seen[process.PID] = true
	}
	p.ProcessCount = len(members)
	if !cpuAvailable {
		p.CPUPercent = 0
		p.Unavailable = append(p.Unavailable, "cpu")
	}
	if !memoryAvailable {
		p.RSS = 0
		p.Unavailable = append(p.Unavailable, "memory")
	}
	return p, nil
}

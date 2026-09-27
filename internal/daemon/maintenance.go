package daemon

import (
	"context"
	"log/slog"
	"time"

	"blora.dev/panel/internal/model"
	run "blora.dev/panel/internal/runtime"
	"blora.dev/panel/internal/terminal"
)

func (d *Daemon) resolveTerminalRun(ctx context.Context, ref model.ResourceRef, runID string) (run.Record, error) {
	if ref.Kind != "instance" || ref.NodeID != d.identity.NodeID {
		return run.Record{}, terminal.ErrForbidden
	}
	var i model.Instance
	if _, err := d.store.Record(ctx, "instance", ref.ID, &i); err != nil {
		return run.Record{}, err
	}
	if i.ID != ref.ID || i.NodeID != ref.NodeID || i.RunID != runID || i.Config.Mode != "container" {
		return run.Record{}, terminal.ErrForbidden
	}
	r, o, err := d.runtime.Recover(ctx, runID)
	if err != nil {
		return r, err
	}
	if o.Exited || o.State != "RUNNING" {
		return r, terminal.ErrInactive
	}
	return r, nil
}

func (d *Daemon) maintainFiles(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			d.mu.Lock()
			services := make([]*fileHandle, 0, len(d.files))
			for _, service := range d.files {
				service.users++
				services = append(services, service)
			}
			d.mu.Unlock()
			for _, service := range services {
				if _, err := service.service.PurgeExpired(ctx, now); err != nil {
					slog.Warn("file trash maintenance failed", "error", err)
				}
				if _, err := service.service.PruneRecords(ctx, now.Add(-30*24*time.Hour)); err != nil {
					slog.Warn("file record maintenance failed", "error", err)
				}
				d.mu.Lock()
				service.users--
				d.mu.Unlock()
			}
		}
	}
}

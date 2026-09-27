package backup

import (
	"context"

	"blora.dev/panel/internal/model"
)

func (m *Manager) authorizeSnapshot(ctx context.Context, id, actor string) (Snapshot, error) {
	var snapshot Snapshot
	if err := ctx.Err(); err != nil {
		return snapshot, err
	}
	if !validID(id) || actor == "" {
		return snapshot, ErrInvalid
	}
	if err := m.readRecord("snapshots/"+id+".json", &snapshot); err != nil {
		return snapshot, err
	}
	if snapshot.Manifest.ID != id {
		return snapshot, ErrCorrupt
	}
	if m.opts.AuthorizeSnapshotRead != nil {
		return snapshot, m.opts.AuthorizeSnapshotRead(ctx, actor, snapshot.Manifest.Spec.Source)
	}
	if actor != snapshot.Manifest.Spec.OwnerID {
		return snapshot, ErrConflict
	}
	return snapshot, nil
}

func (m *Manager) authorizeTarget(ctx context.Context, actor string, target model.ResourceRef) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.opts.AuthorizeRestoreWrite != nil {
		return m.opts.AuthorizeRestoreWrite(ctx, actor, target)
	}
	return nil
}

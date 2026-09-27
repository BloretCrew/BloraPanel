package backup

import (
	"context"

	"blora.dev/panel/internal/filesystem"
)

// CleanupCreate observes only this actor's already accepted backup and its
// stable compensation IDs. It never starts another capture or publishes a
// pending archive. It remains usable after read permission was revoked so a
// previously accepted pause/stop can be compensated through its original hook.
func (m *Manager) CleanupCreate(ctx context.Context, id, owner string) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if !validID(id) || owner == "" {
		return Result{}, ErrInvalid
	}
	release, err := m.acquire(ctx, "repository")
	if err != nil {
		return Result{}, err
	}
	defer release()
	var cp createCheckpoint
	if err = m.readRecord("create/"+id+".json", &cp); err != nil {
		return Result{}, err
	}
	if cp.Spec.OwnerID != owner {
		return Result{}, ErrConflict
	}
	if cp.Result.Stage == "succeeded" {
		return cp.Result, nil
	}
	if cp.Result.Stage == "committing" {
		relation, e := filesystem.InspectRootRelation(ctx, m.root, archivePath(id))
		if e != nil {
			return cp.Result, e
		}
		if relation.Exists && relation.ObjectID == cp.ArchiveObjectID {
			if e = m.checkObject(ctx, archivePath(id), cp.ArchiveObjectID, cp.ArchiveHash, cp.ArchiveBytes); e != nil {
				return cp.Result, e
			}
			cp.Result.Stage = "after_hook"
		}
	}
	if cp.Result.Stage == "after_hook" {
		return m.finishCreate(ctx, &cp)
	}
	return m.failCreate(ctx, &cp, ErrInterrupted)
}

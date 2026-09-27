package backup

import (
	"context"
	"errors"
	"os"
	"sort"
	"time"

	"blora.dev/panel/internal/filesystem"
)

type RetentionResult struct {
	Removed []string `json:"removed"`
	Bytes   int64    `json:"bytes"`
	Kept    int      `json:"kept"`
	Busy    []string `json:"busy,omitempty"`
}

// Prune records deletion intent before removing an owned archive. Unrecognized
// files and another owner's snapshots are never inferred to be disposable.
func (m *Manager) Prune(ctx context.Context, policy Retention, now time.Time) (RetentionResult, error) {
	result := RetentionResult{Removed: []string{}}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if policy.OwnerID == "" || !validResource(policy.Source) || filesystem.ValidatePath(policy.Path) != nil || policy.KeepLast < 0 || policy.MaxAge < 0 || policy.MaxBytes < 0 || (policy.KeepLast == 0 && policy.MaxAge == 0 && policy.MaxBytes == 0) {
		return result, ErrInvalid
	}
	release, err := m.acquire(ctx, "repository")
	if err != nil {
		return result, err
	}
	defer release()
	ids, err := m.recordIDs("snapshots")
	if err != nil {
		return result, err
	}
	var snapshots []Snapshot
	var total int64
	for _, id := range ids {
		if err = ctx.Err(); err != nil {
			return result, err
		}
		var snap Snapshot
		if err = m.readRecord("snapshots/"+id+".json", &snap); err != nil {
			return result, err
		}
		if snap.Manifest.ID != id {
			return result, ErrCorrupt
		}
		if snap.Manifest.Spec.OwnerID != policy.OwnerID || snap.Manifest.Spec.Source != policy.Source || snap.Manifest.Spec.Path != policy.Path || snap.State == "pruned" {
			continue
		}
		if snap.State != "ready" && snap.State != "pruning" {
			return result, ErrCorrupt
		}
		snapshots = append(snapshots, snap)
		total += snap.ArchiveBytes
	}
	sort.Slice(snapshots, func(i, j int) bool { return snapshots[i].Manifest.Created.After(snapshots[j].Manifest.Created) })
	result.Kept = len(snapshots)
	// Oldest first for byte/age eviction, retaining the explicit newest floor.
	for i := len(snapshots) - 1; i >= 0; i-- {
		snap := snapshots[i]
		remove := snap.State == "pruning"
		if i >= policy.KeepLast {
			remove = remove || (policy.MaxAge > 0 && snap.Manifest.Created.Before(now.Add(-policy.MaxAge))) || (policy.MaxBytes > 0 && total > policy.MaxBytes) || (policy.MaxAge == 0 && policy.MaxBytes == 0)
		}
		if !remove {
			continue
		}
		id := snap.Manifest.ID
		m.mu.Lock()
		busy := m.readers[id] > 0
		if !busy {
			m.active["prune:"+id] = true
		}
		m.mu.Unlock()
		if busy {
			result.Busy = append(result.Busy, id)
			continue
		}
		err = m.pruneOne(ctx, &snap)
		m.mu.Lock()
		delete(m.active, "prune:"+id)
		m.mu.Unlock()
		if err != nil {
			return result, err
		}
		result.Removed = append(result.Removed, id)
		result.Bytes += snap.ArchiveBytes
		total -= snap.ArchiveBytes
		result.Kept--
	}
	return result, nil
}
func (m *Manager) pruneOne(ctx context.Context, snapshot *Snapshot) error {
	p := archivePath(snapshot.Manifest.ID)
	if snapshot.State == "ready" {
		if err := m.checkObject(ctx, p, snapshot.ArchiveObjectID, snapshot.ArchiveHash, snapshot.ArchiveBytes); err != nil {
			return err
		}
		snapshot.State = "pruning"
		if err := m.writeRecord("snapshots/"+snapshot.Manifest.ID+".json", snapshot); err != nil {
			return err
		}
	}
	relation, err := filesystem.InspectRootRelation(ctx, m.root, p)
	if err != nil {
		return err
	}
	if relation.Exists {
		if relation.ObjectID != snapshot.ArchiveObjectID {
			return ErrConflict
		}
		if err = m.checkObject(ctx, p, snapshot.ArchiveObjectID, snapshot.ArchiveHash, snapshot.ArchiveBytes); err != nil {
			return err
		}
		if err = ctx.Err(); err != nil {
			return err
		}
		if err = m.root.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err = syncDirectory(m.root, "objects"); err != nil {
			return err
		}
	}
	snapshot.State = "pruned"
	return m.writeRecord("snapshots/"+snapshot.Manifest.ID+".json", snapshot)
}

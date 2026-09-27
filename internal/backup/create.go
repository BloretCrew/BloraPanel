package backup

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"sort"
	"time"

	"blora.dev/panel/internal/filesystem"
)

func (m *Manager) Create(ctx context.Context, source *filesystem.Service, spec CreateSpec) (Result, error) {
	if spec.Compression == "" {
		spec.Compression = "deflate"
	}
	if spec.Consistency.Mode == "" {
		spec.Consistency.Mode = "files"
	}
	if !validID(spec.ID) || spec.OwnerID == "" || !validResource(spec.Source) || !validHash(spec.Version) || filesystem.ValidatePath(spec.Path) != nil || (spec.Compression != "store" && spec.Compression != "deflate") {
		return Result{}, ErrInvalid
	}
	switch spec.Consistency.Mode {
	case "files":
		if spec.Consistency.Before != "" || spec.Consistency.After != "" {
			return Result{}, ErrInvalid
		}
	case "save", "pause", "stop", "hooks":
		if m.opts.Hooks == nil {
			return Result{}, ErrUnsupported
		}
	default:
		return Result{}, ErrInvalid
	}
	// Archive writers and retention share this lock: free-space accounting never
	// grants the same repository budget to concurrent snapshots.
	release, err := m.acquire(ctx, "repository")
	if err != nil {
		return Result{}, err
	}
	defer release()
	name := "create/" + spec.ID + ".json"
	var cp createCheckpoint
	err = m.readRecord(name, &cp)
	if err == nil {
		if cp.Spec != spec {
			return cp.Result, ErrConflict
		}
		switch cp.Result.Stage {
		case "succeeded":
			if m.opts.AuthorizeSnapshotRead != nil {
				if err = m.opts.AuthorizeSnapshotRead(ctx, spec.OwnerID, spec.Source); err != nil {
					return Result{}, err
				}
			}
			var snap Snapshot
			if err = m.readRecord("snapshots/"+spec.ID+".json", &snap); err != nil {
				return cp.Result, err
			}
			if snap.State != "ready" {
				return cp.Result, ErrConflict
			}
			if err = m.checkObject(ctx, archivePath(spec.ID), snap.ArchiveObjectID, snap.ArchiveHash, snap.ArchiveBytes); err != nil {
				return cp.Result, err
			}
			return cp.Result, nil
		case "failed", "interrupted":
			if cp.Result.CleanupPending || cp.After != nil && cp.After.State != "succeeded" {
				return m.failCreate(ctx, &cp, ErrInterrupted)
			}
			return cp.Result, ErrInterrupted
		case "packing":
			return m.failCreate(ctx, &cp, ErrInterrupted)
		case "committing", "after_hook":
			if m.opts.AuthorizeSnapshotRead != nil {
				if err = m.opts.AuthorizeSnapshotRead(ctx, spec.OwnerID, spec.Source); err != nil {
					return m.failCreate(ctx, &cp, err)
				}
			}
			return m.finishCreate(ctx, &cp)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return Result{}, err
	} else {
		if m.opts.AuthorizeSnapshotRead != nil {
			if err = m.opts.AuthorizeSnapshotRead(ctx, spec.OwnerID, spec.Source); err != nil {
				return Result{}, err
			}
		}
		relation, e := m.separate(ctx, source, spec.Path)
		if e != nil {
			return Result{}, e
		}
		entry, e := source.Stat(ctx, spec.Path)
		if e != nil {
			return Result{}, e
		}
		if entry.Version != spec.Version {
			return Result{}, ErrConflict
		}
		cp = createCheckpoint{Spec: spec, Result: Result{ID: spec.ID, Stage: "accepted"}, SourceRelation: relation}
		if e = m.writeRecord(name, cp); e != nil {
			return cp.Result, e
		}
	}
	if m.opts.AuthorizeSnapshotRead != nil {
		if err = m.opts.AuthorizeSnapshotRead(ctx, spec.OwnerID, spec.Source); err != nil {
			return m.failCreate(ctx, &cp, err)
		}
	}
	if spec.Consistency.Mode != "files" {
		if err = m.runHook(ctx, &cp, "before"); err != nil {
			return m.failCreate(ctx, &cp, err)
		}
	}
	relation, err := m.separate(ctx, source, spec.Path)
	if err != nil {
		return m.failCreate(ctx, &cp, err)
	}
	if relation.RootID != cp.SourceRelation.RootID || (spec.Consistency.Mode == "files" && relation.ObjectID != cp.SourceRelation.ObjectID) {
		return m.failCreate(ctx, &cp, ErrConflict)
	}
	manifest, err := m.capture(ctx, source, spec, relation)
	if err != nil {
		return m.failCreate(ctx, &cp, err)
	}
	if spec.Consistency.Mode == "files" && manifest.CapturedVersion != spec.Version {
		return m.failCreate(ctx, &cp, ErrConflict)
	}
	cp.Manifest = &manifest
	limit, err := m.repositoryBudget(ctx)
	if err != nil {
		return m.failCreate(ctx, &cp, err)
	}
	cp.Result.Stage = "packing"
	if err = m.writeRecord(name, cp); err != nil {
		return cp.Result, err
	}
	f, err := m.root.OpenFile(pendingPath(spec.ID), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return m.failCreate(ctx, &cp, err)
	}
	rel, err := filesystem.InspectRootRelation(ctx, m.root, pendingPath(spec.ID))
	if err == nil {
		cp.ArchiveObjectID = rel.ObjectID
		err = m.writeRecord(name, cp)
	}
	if err == nil {
		err = m.pack(ctx, source, f, &cp, min(limit, m.opts.MaxArchiveBytes))
	}
	if err == nil {
		err = f.Sync()
	}
	if err == nil {
		cp.ArchiveHash, cp.ArchiveBytes, err = digestFile(ctx, f, m.opts.MaxArchiveBytes)
	}
	if e := f.Close(); err == nil {
		err = e
	}
	if err != nil {
		return m.failCreate(ctx, &cp, err)
	}
	last, err := source.Stat(ctx, spec.Path)
	if err == nil && last.Version != manifest.CapturedVersion {
		err = ErrConflict
	}
	finalRelation, e := m.separate(ctx, source, spec.Path)
	if err == nil {
		err = e
	}
	if err == nil && (finalRelation.ObjectID != relation.ObjectID || finalRelation.RootID != relation.RootID) {
		err = ErrConflict
	}
	if err != nil {
		return m.failCreate(ctx, &cp, err)
	}
	cp.Result.Stage = "committing"
	if err = m.writeRecord(name, cp); err != nil {
		return cp.Result, err
	}
	return m.finishCreate(ctx, &cp)
}

func (m *Manager) capture(ctx context.Context, source *filesystem.Service, spec CreateSpec, relation filesystem.RelationFacts) (Manifest, error) {
	first, err := source.Stat(ctx, spec.Path)
	if err != nil {
		return Manifest{}, err
	}
	out := Manifest{Format: "blora-backup-v1", ID: spec.ID, Spec: spec, CapturedVersion: first.Version, SourceRootID: relation.RootID, SourceObjectID: relation.ObjectID, Created: time.Now().UTC()}
	var visit func(string, string) error
	visit = func(p, relative string) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if m.opts.AuthorizeSnapshotRead != nil {
			if err := m.opts.AuthorizeSnapshotRead(ctx, spec.OwnerID, spec.Source); err != nil {
				return err
			}
		}
		if len(out.Entries) >= m.opts.MaxEntries {
			return ErrLimit
		}
		item, err := source.Stat(ctx, p)
		if err != nil {
			return err
		}
		entry := Entry{Path: relative, Kind: item.Kind, Mode: item.Mode & 0777, Modified: item.Modified}
		if item.Kind == "file" {
			entry.Size = item.Size
			entry.Hash = item.Version
			if item.Size > m.opts.MaxFileBytes || item.Size < 0 || out.Total > m.opts.MaxTotalBytes-item.Size {
				return ErrLimit
			}
			out.Total += item.Size
		} else if item.Kind != "directory" {
			return ErrUnsupported
		}
		out.Entries = append(out.Entries, entry)
		if item.Kind != "directory" {
			return nil
		}
		offset := 0
		version := ""
		for {
			page, err := source.List(ctx, p, filesystem.ListOptions{Offset: offset, Limit: 200, Version: version})
			if err != nil {
				return err
			}
			version = page.Version
			for _, child := range page.Items {
				if err = visit(path.Join(p, child.Name), path.Join(relative, child.Name)); err != nil {
					return err
				}
			}
			if page.NextOffset < 0 {
				break
			}
			offset = page.NextOffset
		}
		return nil
	}
	if err = visit(spec.Path, "."); err != nil {
		return Manifest{}, err
	}
	last, err := source.Stat(ctx, spec.Path)
	if err != nil {
		return Manifest{}, err
	}
	if first.Version != last.Version {
		return Manifest{}, ErrConflict
	}
	b, err := json.Marshal(out)
	if err != nil {
		return Manifest{}, err
	}
	if len(b) > manifestLimit {
		return Manifest{}, ErrLimit
	}
	return out, nil
}

type budgetWriter struct {
	io.Writer
	remaining int64
}

func (w *budgetWriter) Write(b []byte) (int, error) {
	if int64(len(b)) > w.remaining {
		return 0, ErrLimit
	}
	n, e := w.Writer.Write(b)
	w.remaining -= int64(n)
	return n, e
}
func archiveEntryName(e Entry) string {
	if e.Path == "." {
		if e.Kind == "directory" {
			return "data/"
		}
		return "data/content"
	}
	n := "data/" + e.Path
	if e.Kind == "directory" {
		n += "/"
	}
	return n
}
func (m *Manager) pack(ctx context.Context, source *filesystem.Service, f *os.File, cp *createCheckpoint, limit int64) error {
	w := zip.NewWriter(&budgetWriter{Writer: f, remaining: limit})
	for _, entry := range cp.Manifest.Entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		if m.opts.AuthorizeSnapshotRead != nil {
			if err := m.opts.AuthorizeSnapshotRead(ctx, cp.Spec.OwnerID, cp.Spec.Source); err != nil {
				return err
			}
		}
		method := uint16(zip.Store)
		if cp.Spec.Compression == "deflate" && entry.Kind == "file" {
			method = zip.Deflate
		}
		h := &zip.FileHeader{Name: archiveEntryName(entry), Method: method, Modified: entry.Modified}
		mode := os.FileMode(entry.Mode)
		if entry.Kind == "directory" {
			mode |= os.ModeDir
		}
		h.SetMode(mode)
		writer, err := w.CreateHeader(h)
		if err != nil {
			return err
		}
		if entry.Kind == "file" {
			for offset := int64(0); offset < entry.Size; {
				if m.opts.AuthorizeSnapshotRead != nil {
					if err := m.opts.AuthorizeSnapshotRead(ctx, cp.Spec.OwnerID, cp.Spec.Source); err != nil {
						return err
					}
				}
				chunk, err := source.ReadChunk(ctx, path.Join(cp.Spec.Path, entry.Path), entry.Hash, offset, chunkBytes)
				if err != nil {
					return err
				}
				if chunk.Offset != offset || chunk.Total != entry.Size || chunk.Version != entry.Hash || chunk.Hash != hashBytes(chunk.Data) || len(chunk.Data) == 0 {
					return ErrConflict
				}
				if _, err = writer.Write(chunk.Data); err != nil {
					return err
				}
				offset += int64(len(chunk.Data))
				cp.Result.Bytes += int64(len(chunk.Data))
			}
		}
		cp.Result.Completed++
		cp.Result.CurrentPath = entry.Path
		if err := w.Flush(); err != nil {
			return err
		}
		if err := f.Sync(); err != nil {
			return err
		}
		if err := m.writeRecord("create/"+cp.Spec.ID+".json", cp); err != nil {
			return err
		}
	}
	b, err := json.Marshal(cp.Manifest)
	if err != nil {
		return err
	}
	manifest, err := w.CreateHeader(&zip.FileHeader{Name: "manifest.json", Method: zip.Store})
	if err != nil {
		return err
	}
	if _, err = manifest.Write(b); err != nil {
		return err
	}
	return w.Close()
}

func (m *Manager) runHook(ctx context.Context, cp *createCheckpoint, phase string) error {
	if m.opts.Hooks == nil {
		return ErrUnsupported
	}
	ptr := &cp.Before
	if phase == "after" {
		ptr = &cp.After
	}
	call := HookCall{ID: cp.Spec.ID + ":" + phase, BackupID: cp.Spec.ID, OwnerID: cp.Spec.OwnerID, Resource: cp.Spec.Source, Strategy: cp.Spec.Consistency, Phase: phase}
	if *ptr != nil && (*ptr).State == "succeeded" {
		return nil
	}
	reconcile := *ptr != nil
	previousPending := cp.Result.CleanupPending
	if !reconcile {
		*ptr = &HookReceipt{ID: call.ID, State: "running", At: time.Now().UTC()}
		if phase == "after" {
			cp.Result.CleanupPending = true
		}
		if err := m.writeRecord("create/"+cp.Spec.ID+".json", cp); err != nil {
			return err
		}
	}
	var receipt HookReceipt
	var err error
	if reconcile {
		receipt, err = m.opts.Hooks.Reconcile(ctx, call)
	} else {
		receipt, err = m.opts.Hooks.Run(ctx, call)
	}
	if receipt.ID != call.ID || receipt.State == "" {
		receipt = HookReceipt{ID: call.ID, State: "unknown", At: time.Now().UTC()}
		if err == nil {
			err = ErrInterrupted
		}
	}
	*ptr = &receipt
	if phase == "after" {
		if receipt.State == "succeeded" {
			cp.Result.CleanupPending = previousPending
		} else {
			cp.Result.CleanupPending = true
		}
	}
	if e := m.writeRecord("create/"+cp.Spec.ID+".json", cp); e != nil {
		return e
	}
	if err != nil {
		return err
	}
	if receipt.State != "succeeded" {
		return ErrInterrupted
	}
	return nil
}

func (m *Manager) finishCreate(ctx context.Context, cp *createCheckpoint) (Result, error) {
	if cp.Manifest == nil || !validHash(cp.ArchiveHash) || cp.ArchiveObjectID == "" {
		return m.failCreate(ctx, cp, ErrInterrupted)
	}
	if cp.Result.Stage == "committing" {
		if err := ctx.Err(); err != nil {
			return m.failCreate(ctx, cp, err)
		}
		final, err := filesystem.InspectRootRelation(ctx, m.root, archivePath(cp.Spec.ID))
		if err != nil {
			return cp.Result, err
		}
		if !final.Exists {
			if err = m.checkObject(ctx, pendingPath(cp.Spec.ID), cp.ArchiveObjectID, cp.ArchiveHash, cp.ArchiveBytes); err != nil {
				return m.failCreate(ctx, cp, err)
			}
			if err = m.root.Rename(pendingPath(cp.Spec.ID), archivePath(cp.Spec.ID)); err != nil {
				return cp.Result, err
			}
			if err = syncDirectory(m.root, "objects"); err != nil {
				return cp.Result, err
			}
		}
		if err = m.checkObject(ctx, archivePath(cp.Spec.ID), cp.ArchiveObjectID, cp.ArchiveHash, cp.ArchiveBytes); err != nil {
			return m.failCreate(ctx, cp, err)
		}
		cp.Result.Stage = "after_hook"
		if err = m.writeRecord("create/"+cp.Spec.ID+".json", cp); err != nil {
			return cp.Result, err
		}
	}
	if cp.Spec.Consistency.Mode != "files" {
		captured := snapshotInfo(Snapshot{Manifest: *cp.Manifest, ArchiveHash: cp.ArchiveHash, ArchiveObjectID: cp.ArchiveObjectID, ArchiveBytes: cp.ArchiveBytes, State: "cleanup_pending", Before: cp.Before, After: cp.After})
		cp.Result.Snapshot = &captured
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		if err := m.runHook(cleanup, cp, "after"); err != nil {
			cp.Result.CleanupPending = true
			cp.Result.Unknown = true
			cp.Result.Error = err.Error()
			_ = m.writeRecord("create/"+cp.Spec.ID+".json", cp)
			return cp.Result, err
		}
	}
	snap := Snapshot{Manifest: *cp.Manifest, ArchiveHash: cp.ArchiveHash, ArchiveObjectID: cp.ArchiveObjectID, ArchiveBytes: cp.ArchiveBytes, State: "ready", Before: cp.Before, After: cp.After}
	if err := m.writeRecord("snapshots/"+cp.Spec.ID+".json", snap); err != nil {
		return cp.Result, err
	}
	info := snapshotInfo(snap)
	cp.Result.Stage = "succeeded"
	cp.Result.Snapshot = &info
	cp.Result.Partial = false
	cp.Result.Unknown = false
	cp.Result.CleanupPending = false
	cp.Result.Error = ""
	return cp.Result, m.writeRecord("create/"+cp.Spec.ID+".json", cp)
}

func (m *Manager) failCreate(ctx context.Context, cp *createCheckpoint, cause error) (Result, error) {
	cp.Result.CleanupPending = false
	cp.Result.Stage = "failed"
	cp.Result.Error = cause.Error()
	cp.Result.Partial = cp.Result.Completed > 0
	if errors.Is(cause, ErrInterrupted) {
		cp.Result.Stage = "interrupted"
		cp.Result.Unknown = true
	}
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	if cp.ArchiveObjectID != "" {
		relation, err := filesystem.InspectRootRelation(cleanup, m.root, pendingPath(cp.Spec.ID))
		if err == nil && relation.Exists && relation.ObjectID == cp.ArchiveObjectID {
			err = m.root.Remove(pendingPath(cp.Spec.ID))
			if err == nil {
				err = syncDirectory(m.root, "pending")
			}
		}
		if err != nil || relation.Exists && relation.ObjectID != cp.ArchiveObjectID {
			cp.Result.CleanupPending = true
		}
	}
	if cp.Spec.Consistency.Mode != "files" && cp.Before != nil {
		if err := m.runHook(cleanup, cp, "after"); err != nil {
			cp.Result.CleanupPending = true
			cp.Result.Unknown = true
			cause = errors.Join(cause, err)
		}
	}
	return cp.Result, errors.Join(cause, m.writeRecord("create/"+cp.Spec.ID+".json", cp))
}

func (m *Manager) repositoryBudget(ctx context.Context) (int64, error) {
	ids, err := m.recordIDs("snapshots")
	if err != nil {
		return 0, err
	}
	ready := 0
	for _, id := range ids {
		var s Snapshot
		if err = m.readRecord("snapshots/"+id+".json", &s); err != nil {
			return 0, err
		}
		if s.State != "pruned" {
			ready++
		}
	}
	if ready >= m.opts.MaxSnapshots {
		return 0, ErrLimit
	}
	var total int64
	for _, dir := range []string{"objects", "pending"} {
		f, err := m.root.Open(dir)
		if err != nil {
			return 0, err
		}
		count := 0
		for {
			if err = ctx.Err(); err != nil {
				f.Close()
				return 0, err
			}
			items, e := f.ReadDir(128)
			for _, item := range items {
				count++
				if count > m.opts.MaxSnapshots*4 {
					f.Close()
					return 0, ErrLimit
				}
				info, e := item.Info()
				if e != nil {
					f.Close()
					return 0, e
				}
				if !info.Mode().IsRegular() {
					f.Close()
					return 0, ErrCorrupt
				}
				if info.Size() > m.opts.MaxRepositoryBytes-total {
					f.Close()
					return 0, ErrLimit
				}
				total += info.Size()
			}
			if e == io.EOF {
				break
			}
			if e != nil {
				f.Close()
				return 0, e
			}
		}
		if err = f.Close(); err != nil {
			return 0, err
		}
	}
	if total >= m.opts.MaxRepositoryBytes {
		return 0, ErrLimit
	}
	return m.opts.MaxRepositoryBytes - total, nil
}

func (m *Manager) Inspect(ctx context.Context, id string) (SnapshotInfo, error) {
	snapshot, _, close, err := m.openSnapshot(ctx, id)
	if err != nil {
		return SnapshotInfo{}, err
	}
	defer close()
	return snapshotInfo(snapshot), nil
}
func (m *Manager) List(ctx context.Context) ([]SnapshotInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	ids, err := m.recordIDs("snapshots")
	if err != nil {
		return nil, err
	}
	items := make([]SnapshotInfo, 0, len(ids))
	for _, id := range ids {
		if err = ctx.Err(); err != nil {
			return nil, err
		}
		var s Snapshot
		if err = m.readRecord("snapshots/"+id+".json", &s); err != nil {
			return nil, err
		}
		if s.State != "pruned" {
			items = append(items, snapshotInfo(s))
			if len(items) > m.opts.MaxSnapshots {
				return nil, ErrLimit
			}
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Created.After(items[j].Created) })
	return items, nil
}

func sameJSON(a, b any) bool {
	left, e := json.Marshal(a)
	if e != nil {
		return false
	}
	right, e := json.Marshal(b)
	return e == nil && bytes.Equal(left, right)
}

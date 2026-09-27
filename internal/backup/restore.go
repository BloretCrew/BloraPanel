package backup

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"blora.dev/panel/internal/filesystem"
)

func (m *Manager) PlanRestore(ctx context.Context, target *filesystem.Service, request RestoreRequest) (RestorePlan, error) {
	if !validID(request.ID) || !validID(request.BackupID) || request.OwnerID == "" || !validResource(request.Target) || !validVersion(request.Version) || filesystem.ValidatePath(request.Path) != nil {
		return RestorePlan{}, ErrInvalid
	}
	if _, err := m.authorizeSnapshot(ctx, request.BackupID, request.OwnerID); err != nil {
		return RestorePlan{}, err
	}
	if err := m.authorizeTarget(ctx, request.OwnerID, request.Target); err != nil {
		return RestorePlan{}, err
	}
	release, err := m.acquire(ctx, "plan:"+request.ID)
	if err != nil {
		return RestorePlan{}, err
	}
	defer release()
	var old RestorePlan
	if old, err = m.GetRestorePlan(ctx, request.ID); err == nil {
		if old.Request != request {
			return old, ErrConflict
		}
		return old, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return old, err
	}
	ids, err := m.recordIDs("plans")
	if err != nil {
		return old, err
	}
	if len(ids) >= m.opts.MaxSnapshots {
		return old, ErrLimit
	}
	snapshot, _, close, err := m.openSnapshot(ctx, request.BackupID)
	if err != nil {
		return old, err
	}
	defer close()
	relation, err := m.separate(ctx, target, request.Path)
	if err != nil {
		return old, err
	}
	version, _, err := targetVersion(ctx, target, request.Path)
	if err != nil {
		return old, err
	}
	if version != request.Version {
		return old, ErrConflict
	}
	plan := RestorePlan{Request: request, ArchiveHash: snapshot.ArchiveHash, ArchiveObjectID: snapshot.ArchiveObjectID, TargetRootID: relation.RootID, TargetObjectID: relation.ObjectID, Total: snapshot.Manifest.Total, Created: time.Now().UTC()}
	for _, entry := range snapshot.Manifest.Entries {
		if _, err = m.authorizeSnapshot(ctx, request.BackupID, request.OwnerID); err != nil {
			return old, err
		}
		if err = m.authorizeTarget(ctx, request.OwnerID, request.Target); err != nil {
			return old, err
		}
		p := path.Join(request.Path, entry.Path)
		version, kind, err := targetVersion(ctx, target, p)
		if err != nil {
			return old, err
		}
		if version != filesystem.MissingVersion && kind != entry.Kind {
			return old, ErrConflict
		}
		r, err := target.RelationFacts(ctx, p)
		if err != nil {
			return old, err
		}
		plan.Entries = append(plan.Entries, RestoreEntry{Entry: entry, TargetPath: p, TargetVersion: version, TargetObjectID: r.ObjectID})
	}
	final, err := target.RelationFacts(ctx, request.Path)
	if err != nil {
		return old, err
	}
	version, _, err = targetVersion(ctx, target, request.Path)
	if err != nil {
		return old, err
	}
	if final.RootID != relation.RootID || final.ObjectID != relation.ObjectID || version != request.Version {
		return old, ErrConflict
	}
	if _, err = m.authorizeSnapshot(ctx, request.BackupID, request.OwnerID); err != nil {
		return old, err
	}
	if err = m.authorizeTarget(ctx, request.OwnerID, request.Target); err != nil {
		return old, err
	}
	b, err := json.Marshal(plan)
	if err != nil {
		return old, err
	}
	if len(b) > manifestLimit {
		return old, ErrLimit
	}
	plan.Hash = hashBytes(b)
	return plan, m.writeRecord("plans/"+request.ID+".json", plan)
}

func targetVersion(ctx context.Context, target *filesystem.Service, p string) (string, string, error) {
	entry, err := target.Stat(ctx, p)
	if errors.Is(err, os.ErrNotExist) {
		return filesystem.MissingVersion, "", nil
	}
	return entry.Version, entry.Kind, err
}

func (m *Manager) GetRestorePlan(ctx context.Context, id string) (RestorePlan, error) {
	if err := ctx.Err(); err != nil {
		return RestorePlan{}, err
	}
	if !validID(id) {
		return RestorePlan{}, ErrInvalid
	}
	var plan RestorePlan
	err := m.readRecord("plans/"+id+".json", &plan)
	if err != nil {
		return plan, err
	}
	if plan.Request.ID != id || !validHash(plan.Hash) || len(plan.Entries) == 0 || len(plan.Entries) > m.opts.MaxEntries {
		return plan, ErrCorrupt
	}
	expected := plan.Hash
	plan.Hash = ""
	body, e := json.Marshal(plan)
	plan.Hash = expected
	if e != nil || len(body) > manifestLimit || hashBytes(body) != expected {
		return plan, ErrCorrupt
	}
	return plan, nil
}

func (m *Manager) Restore(ctx context.Context, target *filesystem.Service, spec RestoreSpec) (Result, error) {
	if !validID(spec.ID) || !validID(spec.PlanID) || spec.OwnerID == "" || !validHash(spec.PlanHash) {
		return Result{}, ErrInvalid
	}
	plan, err := m.GetRestorePlan(ctx, spec.PlanID)
	if err != nil {
		return Result{}, err
	}
	if plan.Hash != spec.PlanHash || plan.Request.OwnerID != spec.OwnerID {
		return Result{}, ErrConflict
	}
	if _, err = m.authorizeSnapshot(ctx, plan.Request.BackupID, spec.OwnerID); err != nil {
		return Result{}, err
	}
	if err = m.authorizeTarget(ctx, spec.OwnerID, plan.Request.Target); err != nil {
		return Result{}, err
	}
	confirmed := 0
	confirmPlan := spec.OverwritePlanHash != ""
	if confirmPlan && (spec.OverwritePlanHash != plan.Hash || len(spec.Overwrite) != 0) {
		return Result{}, ErrConflict
	}
	for _, entry := range plan.Entries {
		if entry.Kind == "file" && entry.TargetVersion != filesystem.MissingVersion {
			if !confirmPlan && spec.Overwrite[entry.TargetPath] != entry.TargetVersion {
				return Result{}, ErrConflict
			}
			confirmed++
		}
	}
	if !confirmPlan && len(spec.Overwrite) != confirmed {
		return Result{}, ErrInvalid
	}
	release, err := m.acquire(ctx, "restore:"+plan.Request.Target.NodeID+":"+plan.Request.Target.ID)
	if err != nil {
		return Result{}, err
	}
	defer release()
	name := "restore/" + spec.ID + ".json"
	var cp restoreCheckpoint
	if err = m.readRecord(name, &cp); err == nil {
		if !sameJSON(cp.Spec, spec) {
			return cp.Result, ErrConflict
		}
		if cp.Result.Stage == "succeeded" {
			return cp.Result, nil
		}
		// A previous invocation may have changed the tree without persisting the
		// receipt. Observe the one recoverable upload, but never continue the
		// next overwrite merely because the caller repeated the request.
		return m.interruptRestore(ctx, target, &cp, ErrInterrupted)
	} else if !errors.Is(err, os.ErrNotExist) {
		return Result{}, err
	}
	snapshot, reader, close, err := m.openSnapshot(ctx, plan.Request.BackupID)
	if err != nil {
		return Result{}, err
	}
	defer close()
	if snapshot.ArchiveHash != plan.ArchiveHash || snapshot.ArchiveObjectID != plan.ArchiveObjectID {
		return Result{}, ErrConflict
	}
	if len(plan.Entries) != len(snapshot.Manifest.Entries) || plan.Total != snapshot.Manifest.Total {
		return Result{}, ErrCorrupt
	}
	for i, entry := range plan.Entries {
		if !sameJSON(entry.Entry, snapshot.Manifest.Entries[i]) || entry.TargetPath != path.Join(plan.Request.Path, entry.Path) || !validVersion(entry.TargetVersion) {
			return Result{}, ErrCorrupt
		}
	}
	relation, err := m.separate(ctx, target, plan.Request.Path)
	if err != nil {
		return Result{}, err
	}
	version, _, err := targetVersion(ctx, target, plan.Request.Path)
	if err != nil {
		return Result{}, err
	}
	if relation.RootID != plan.TargetRootID || relation.ObjectID != plan.TargetObjectID || version != plan.Request.Version {
		return Result{}, ErrConflict
	}
	cp = restoreCheckpoint{Spec: spec, Result: Result{ID: spec.ID, Stage: "restoring"}, DirectoryIDs: map[string]string{}, FileIDs: map[string]string{}}
	for _, entry := range plan.Entries {
		if entry.Kind == "directory" && entry.TargetVersion != filesystem.MissingVersion {
			cp.DirectoryIDs[entry.TargetPath] = entry.TargetObjectID
		}
	}
	if err = m.writeRecord(name, cp); err != nil {
		return cp.Result, err
	}
	index := make(map[string]*zip.File, len(reader.File))
	for _, f := range reader.File {
		index[f.Name] = f
	}
	for i, entry := range plan.Entries {
		if _, err = m.authorizeSnapshot(ctx, plan.Request.BackupID, spec.OwnerID); err != nil {
			return m.interruptRestore(ctx, target, &cp, err)
		}
		if err = m.authorizeTarget(ctx, spec.OwnerID, plan.Request.Target); err != nil {
			return m.interruptRestore(ctx, target, &cp, err)
		}
		if err = m.archiveBinding(ctx, snapshot); err != nil {
			return m.interruptRestore(ctx, target, &cp, err)
		}
		if err = m.checkRestoreParents(ctx, target, plan, &cp, entry.TargetPath); err != nil {
			return m.interruptRestore(ctx, target, &cp, err)
		}
		cp.Result.CurrentPath = entry.TargetPath
		current, kind, err := targetVersion(ctx, target, entry.TargetPath)
		if err != nil {
			return m.interruptRestore(ctx, target, &cp, err)
		}
		r, err := target.RelationFacts(ctx, entry.TargetPath)
		if err != nil {
			return m.interruptRestore(ctx, target, &cp, err)
		}
		if current != entry.TargetVersion || r.ObjectID != entry.TargetObjectID {
			return m.interruptRestore(ctx, target, &cp, ErrConflict)
		}
		if entry.Kind == "directory" {
			if current == filesystem.MissingVersion {
				cp.Operation = "mkdir"
				if err = m.writeRecord(name, cp); err != nil {
					return cp.Result, err
				}
				made, e := target.Mkdir(ctx, entry.TargetPath)
				if e != nil {
					return m.interruptRestore(ctx, target, &cp, e)
				}
				cp.LastReceipt, _ = json.Marshal(made)
				r, e = target.RelationFacts(ctx, entry.TargetPath)
				if e != nil {
					return m.interruptRestore(ctx, target, &cp, e)
				}
				cp.DirectoryIDs[entry.TargetPath] = r.ObjectID
			} else if kind != "directory" {
				return m.interruptRestore(ctx, target, &cp, ErrConflict)
			}
		} else {
			uploadID := hashBytes([]byte(spec.ID + ":" + fmt.Sprint(i)))[7:39]
			upload := filesystem.UploadSpec{ID: uploadID, OwnerID: spec.OwnerID, Path: entry.TargetPath, Total: entry.Size, Hash: entry.Hash, ExpectedVersion: entry.TargetVersion, ExpectedObjectID: entry.TargetObjectID, SourceName: path.Base(entry.Path), SourceModified: entry.Modified.UnixMilli(), SourceFingerprint: plan.ArchiveHash + ":" + plan.ArchiveObjectID, SourceResourceID: plan.Request.BackupID, SourcePath: entry.Path, SourceVersion: plan.ArchiveHash, PreserveMetadata: true, SourceMode: entry.Mode, SourceModifiedNano: entry.Modified.UnixNano()}
			cp.Upload = &upload
			cp.Operation = "upload"
			if err = m.writeRecord(name, cp); err != nil {
				return cp.Result, err
			}
			status, e := target.BeginUpload(ctx, upload)
			if e != nil {
				return m.interruptRestore(ctx, target, &cp, e)
			}
			if status.Spec != upload || status.Stage != "receiving" || status.Offset != 0 {
				return m.interruptRestore(ctx, target, &cp, ErrInterrupted)
			}
			writer := &restoreWriter{ctx: ctx, target: target, id: uploadID, bytes: min(chunkBytes, status.ChunkBytes), authorize: func() error {
				if _, e := m.authorizeSnapshot(ctx, plan.Request.BackupID, spec.OwnerID); e != nil {
					return e
				}
				return m.authorizeTarget(ctx, spec.OwnerID, plan.Request.Target)
			}}
			if e = verifyZIPEntry(ctx, index[archiveEntryName(entry.Entry)], entry.Entry, writer); e != nil {
				return m.interruptRestore(ctx, target, &cp, e)
			}
			if e = writer.Flush(); e != nil {
				return m.interruptRestore(ctx, target, &cp, e)
			}
			if e = m.archiveBinding(ctx, snapshot); e != nil {
				return m.interruptRestore(ctx, target, &cp, e)
			}
			if e = m.checkRestoreParents(ctx, target, plan, &cp, entry.TargetPath); e != nil {
				return m.interruptRestore(ctx, target, &cp, e)
			}
			status, e = target.CommitUpload(ctx, uploadID)
			if e != nil {
				return m.interruptRestore(ctx, target, &cp, e)
			}
			if status.Stage != "committed" || status.Version != entry.Hash {
				return m.interruptRestore(ctx, target, &cp, ErrInterrupted)
			}
			cp.LastReceipt, _ = json.Marshal(status)
			cp.FileIDs[entry.TargetPath] = status.CommitObjectID
			cp.Result.Bytes += entry.Size
		}
		cp.Cursor = i + 1
		cp.Result.Completed = cp.Cursor
		cp.Operation = ""
		cp.Upload = nil
		if err = m.writeRecord(name, cp); err != nil {
			return cp.Result, err
		}
	}
	// Directory timestamps are restored after descendants. Existing directories
	// retain their metadata and any entries outside this manifest.
	for i := len(plan.Entries) - 1; i >= 0; i-- {
		entry := plan.Entries[i]
		if entry.Kind != "directory" || entry.TargetVersion != filesystem.MissingVersion {
			continue
		}
		if _, err = m.authorizeSnapshot(ctx, plan.Request.BackupID, spec.OwnerID); err != nil {
			return m.interruptRestore(ctx, target, &cp, err)
		}
		if err = m.authorizeTarget(ctx, spec.OwnerID, plan.Request.Target); err != nil {
			return m.interruptRestore(ctx, target, &cp, err)
		}
		if err = m.archiveBinding(ctx, snapshot); err != nil {
			return m.interruptRestore(ctx, target, &cp, err)
		}
		if err = m.checkRestoreParents(ctx, target, plan, &cp, entry.TargetPath); err != nil {
			return m.interruptRestore(ctx, target, &cp, err)
		}
		r, e := target.RelationFacts(ctx, entry.TargetPath)
		if e != nil {
			return m.interruptRestore(ctx, target, &cp, e)
		}
		if r.ObjectID != cp.DirectoryIDs[entry.TargetPath] {
			return m.interruptRestore(ctx, target, &cp, ErrConflict)
		}
		current, e := target.Stat(ctx, entry.TargetPath)
		if e != nil {
			return m.interruptRestore(ctx, target, &cp, e)
		}
		cp.Operation = "metadata"
		cp.MetadataCursor = i
		cp.Result.CurrentPath = entry.TargetPath
		if err = m.writeRecord(name, cp); err != nil {
			return cp.Result, err
		}
		receipt, e := target.Metadata(ctx, entry.TargetPath, current.Version, entry.Mode, entry.Modified)
		if e != nil {
			return m.interruptRestore(ctx, target, &cp, e)
		}
		cp.LastReceipt, _ = json.Marshal(receipt)
		cp.Operation = ""
		if err = m.writeRecord(name, cp); err != nil {
			return cp.Result, err
		}
	}
	for _, entry := range plan.Entries {
		if _, err = m.authorizeSnapshot(ctx, plan.Request.BackupID, spec.OwnerID); err != nil {
			return m.interruptRestore(ctx, target, &cp, err)
		}
		if err = m.authorizeTarget(ctx, spec.OwnerID, plan.Request.Target); err != nil {
			return m.interruptRestore(ctx, target, &cp, err)
		}
		relation, e := target.RelationFacts(ctx, entry.TargetPath)
		if e != nil {
			return m.interruptRestore(ctx, target, &cp, e)
		}
		id := cp.DirectoryIDs[entry.TargetPath]
		if entry.Kind == "file" {
			id = cp.FileIDs[entry.TargetPath]
		}
		if id == "" || relation.ObjectID != id {
			return m.interruptRestore(ctx, target, &cp, ErrConflict)
		}
		if entry.Kind == "file" {
			current, e := target.Stat(ctx, entry.TargetPath)
			if e != nil {
				return m.interruptRestore(ctx, target, &cp, e)
			}
			if current.Version != entry.Hash || current.Size != entry.Size {
				return m.interruptRestore(ctx, target, &cp, ErrConflict)
			}
		}
	}
	if err = m.checkObject(ctx, archivePath(plan.Request.BackupID), plan.ArchiveObjectID, plan.ArchiveHash, snapshot.ArchiveBytes); err != nil {
		return m.interruptRestore(ctx, target, &cp, err)
	}
	cp.Result.Stage = "succeeded"
	cp.Result.Partial = false
	cp.Result.CurrentPath = ""
	return cp.Result, m.writeRecord(name, cp)
}

func (m *Manager) checkRestoreParents(ctx context.Context, target *filesystem.Service, plan RestorePlan, cp *restoreCheckpoint, p string) error {
	root, err := target.RelationFacts(ctx, ".")
	if err != nil {
		return err
	}
	if root.RootID != plan.TargetRootID {
		return ErrConflict
	}
	for ancestor := path.Dir(p); ; ancestor = path.Dir(ancestor) {
		if id, ok := cp.DirectoryIDs[ancestor]; ok {
			relation, err := target.RelationFacts(ctx, ancestor)
			if err != nil {
				return err
			}
			if relation.ObjectID != id {
				return ErrConflict
			}
		}
		if ancestor == "." {
			break
		}
	}
	return nil
}

type restoreWriter struct {
	ctx       context.Context
	target    *filesystem.Service
	id        string
	offset    int64
	bytes     int
	buffer    []byte
	authorize func() error
}

func (w *restoreWriter) Write(data []byte) (int, error) {
	written := 0
	for len(data) > 0 {
		n := min(len(data), w.bytes-len(w.buffer))
		if n <= 0 {
			return written, ErrLimit
		}
		w.buffer = append(w.buffer, data[:n]...)
		written += n
		data = data[n:]
		if len(w.buffer) == w.bytes {
			if err := w.Flush(); err != nil {
				return written, err
			}
		}
	}
	return written, nil
}

func (w *restoreWriter) Flush() error {
	if len(w.buffer) == 0 {
		return nil
	}
	if w.authorize != nil {
		if err := w.authorize(); err != nil {
			return err
		}
	}
	status, err := w.target.UploadChunk(w.ctx, w.id, w.offset, w.buffer, hashBytes(w.buffer))
	if err != nil {
		return err
	}
	if status.Offset != w.offset+int64(len(w.buffer)) {
		return ErrInterrupted
	}
	w.offset += int64(len(w.buffer))
	w.buffer = w.buffer[:0]
	return nil
}

func (m *Manager) interruptRestore(ctx context.Context, target *filesystem.Service, cp *restoreCheckpoint, cause error) (Result, error) {
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	cp.Result.Stage = "interrupted"
	cp.Result.Partial = cp.Result.Completed > 0
	cp.Result.Unknown = cp.Operation != ""
	cp.Result.Error = cause.Error()
	if cp.Upload != nil {
		status, err := target.UploadStatus(cleanup, cp.Upload.ID)
		if err == nil && status.Spec == *cp.Upload {
			if status.Stage == "committed" {
				cp.Cursor++
				cp.Result.Completed = cp.Cursor
				cp.Result.Bytes += status.Spec.Total
				cp.Result.Partial = true
				cp.Result.Unknown = false
				cp.LastReceipt, _ = json.Marshal(status)
				cp.Operation = ""
				cp.Upload = nil
			} else {
				status, err = target.CancelUpload(cleanup, cp.Upload.ID)
				if err == nil && status.Stage == "cancelled" {
					cp.Result.Unknown = false
					cp.Operation = ""
					cp.Upload = nil
				} else {
					cp.Result.CleanupPending = true
				}
			}
		} else if errors.Is(err, os.ErrNotExist) {
			cp.Result.Unknown = false
			cp.Operation = ""
			cp.Upload = nil
		} else {
			cp.Result.CleanupPending = true
		}
	}
	return cp.Result, errors.Join(cause, m.writeRecord("restore/"+cp.Spec.ID+".json", cp))
}

func (m *Manager) CreateStatus(ctx context.Context, id string) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if !validID(id) {
		return Result{}, ErrInvalid
	}
	var cp createCheckpoint
	err := m.readRecord("create/"+id+".json", &cp)
	return cp.Result, err
}
func (m *Manager) RestoreStatus(ctx context.Context, id string) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if !validID(id) {
		return Result{}, ErrInvalid
	}
	var cp restoreCheckpoint
	err := m.readRecord("restore/"+id+".json", &cp)
	return cp.Result, err
}

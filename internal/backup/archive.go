package backup

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"strings"

	"blora.dev/panel/internal/filesystem"
)

func (m *Manager) checkObject(ctx context.Context, p, objectID, hash string, size int64) error {
	info, err := m.root.Lstat(p)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() != size {
		return ErrCorrupt
	}
	relation, err := filesystem.InspectRootRelation(ctx, m.root, p)
	if err != nil {
		return err
	}
	if relation.ObjectID != objectID {
		return ErrConflict
	}
	f, err := m.root.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil {
		return err
	}
	if !os.SameFile(info, opened) {
		return ErrConflict
	}
	actual, total, err := digestFile(ctx, f, m.opts.MaxArchiveBytes)
	if err != nil {
		return err
	}
	if actual != hash || total != size {
		return ErrCorrupt
	}
	final, err := filesystem.InspectRootRelation(ctx, m.root, p)
	if err != nil {
		return err
	}
	if final.ObjectID != objectID {
		return ErrConflict
	}
	return nil
}

// openSnapshot verifies the complete package before returning any entry. The
// caller keeps this handle open throughout a restore and rechecks its binding.
func (m *Manager) openSnapshot(ctx context.Context, id string) (Snapshot, *zip.Reader, func(), error) {
	var snapshot Snapshot
	if !validID(id) {
		return snapshot, nil, nil, ErrInvalid
	}
	m.mu.Lock()
	if m.active["prune:"+id] {
		m.mu.Unlock()
		return snapshot, nil, nil, ErrConflict
	}
	m.readers[id]++
	m.mu.Unlock()
	pinned := true
	unpin := func() {
		m.mu.Lock()
		m.readers[id]--
		if m.readers[id] == 0 {
			delete(m.readers, id)
		}
		m.mu.Unlock()
	}
	defer func() {
		if pinned {
			unpin()
		}
	}()
	if err := m.readRecord("snapshots/"+id+".json", &snapshot); err != nil {
		return snapshot, nil, nil, err
	}
	if snapshot.State != "ready" || snapshot.Manifest.ID != id || snapshot.Manifest.Spec.ID != id || !validHash(snapshot.ArchiveHash) || snapshot.ArchiveObjectID == "" || snapshot.ArchiveBytes < 0 || snapshot.ArchiveBytes > m.opts.MaxArchiveBytes {
		return snapshot, nil, nil, ErrCorrupt
	}
	if err := m.checkObject(ctx, archivePath(id), snapshot.ArchiveObjectID, snapshot.ArchiveHash, snapshot.ArchiveBytes); err != nil {
		return snapshot, nil, nil, err
	}
	f, err := m.root.Open(archivePath(id))
	if err != nil {
		return snapshot, nil, nil, err
	}
	close := func() { _ = f.Close() }
	if err = filesystem.CheckZIPIndex(f, snapshot.ArchiveBytes, m.opts.MaxEntries+1); err != nil {
		close()
		return snapshot, nil, nil, err
	}
	reader, err := zip.NewReader(f, snapshot.ArchiveBytes)
	if err != nil {
		close()
		return snapshot, nil, nil, err
	}
	if err = m.validateArchive(ctx, &snapshot, reader); err != nil {
		close()
		return snapshot, nil, nil, err
	}
	// Hash the opened handle too; a pathname replacement between check and open
	// cannot substitute an archive with a coincidentally matching file size.
	actual, total, err := digestFile(ctx, f, m.opts.MaxArchiveBytes)
	if err == nil && (actual != snapshot.ArchiveHash || total != snapshot.ArchiveBytes) {
		err = ErrCorrupt
	}
	if err == nil {
		err = m.archiveBinding(ctx, snapshot)
	}
	if err != nil {
		close()
		return snapshot, nil, nil, err
	}
	pinned = false
	return snapshot, reader, func() { close(); unpin() }, nil
}

func (m *Manager) archiveBinding(ctx context.Context, s Snapshot) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	relation, err := filesystem.InspectRootRelation(ctx, m.root, archivePath(s.Manifest.ID))
	if err != nil {
		return err
	}
	if !relation.Exists || relation.ObjectID != s.ArchiveObjectID {
		return ErrConflict
	}
	info, err := m.root.Lstat(archivePath(s.Manifest.ID))
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() != s.ArchiveBytes {
		return ErrCorrupt
	}
	return nil
}

func (m *Manager) validateArchive(ctx context.Context, snapshot *Snapshot, reader *zip.Reader) error {
	manifest := snapshot.Manifest
	if manifest.Format != "blora-backup-v1" || len(manifest.Entries) == 0 || len(manifest.Entries) > m.opts.MaxEntries || len(reader.File) != len(manifest.Entries)+1 || manifest.Total < 0 || manifest.Total > m.opts.MaxTotalBytes {
		return ErrCorrupt
	}
	b, err := json.Marshal(manifest)
	if err != nil || len(b) > manifestLimit {
		return ErrCorrupt
	}
	index := make(map[string]*zip.File, len(reader.File))
	for _, f := range reader.File {
		if err = ctx.Err(); err != nil {
			return err
		}
		if _, ok := index[f.Name]; ok {
			return ErrCorrupt
		}
		index[f.Name] = f
		if f.Flags&1 != 0 || (f.Method != zip.Store && f.Method != zip.Deflate) || f.Mode()&(os.ModeSymlink|os.ModeDevice|os.ModeNamedPipe|os.ModeSocket|os.ModeCharDevice) != 0 {
			return ErrCorrupt
		}
		if f.Name != "manifest.json" {
			if !strings.HasPrefix(f.Name, "data/") {
				return ErrCorrupt
			}
			rel := strings.TrimSuffix(strings.TrimPrefix(f.Name, "data/"), "/")
			if rel != "" && filesystem.ValidatePath(rel) != nil {
				return ErrCorrupt
			}
		}
	}
	mf, ok := index["manifest.json"]
	if !ok || mf.UncompressedSize64 > manifestLimit || !mf.Mode().IsRegular() {
		return ErrCorrupt
	}
	r, err := mf.Open()
	if err != nil {
		return err
	}
	body, err := io.ReadAll(io.LimitReader(r, manifestLimit+1))
	closeErr := r.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	var embedded Manifest
	if len(body) > manifestLimit || json.Unmarshal(body, &embedded) != nil || !sameJSON(manifest, embedded) {
		return ErrCorrupt
	}
	paths := make(map[string]Entry, len(manifest.Entries))
	var total int64
	for i, e := range manifest.Entries {
		if filesystem.ValidatePath(e.Path) != nil || e.Mode > 0777 || e.Modified.UnixNano() <= 0 || e.Size < 0 || e.Size > m.opts.MaxFileBytes {
			return ErrCorrupt
		}
		if _, exists := paths[e.Path]; exists {
			return ErrCorrupt
		}
		if i == 0 {
			if e.Path != "." {
				return ErrCorrupt
			}
		} else {
			if e.Path == "." || paths[path.Dir(e.Path)].Kind != "directory" {
				return ErrCorrupt
			}
		}
		paths[e.Path] = e
		file, exists := index[archiveEntryName(e)]
		if !exists {
			return ErrCorrupt
		}
		switch e.Kind {
		case "directory":
			if e.Size != 0 || e.Hash != "" || !file.Mode().IsDir() || file.UncompressedSize64 != 0 {
				return ErrCorrupt
			}
		case "file":
			if !validHash(e.Hash) || !file.Mode().IsRegular() || file.UncompressedSize64 != uint64(e.Size) || total > m.opts.MaxTotalBytes-e.Size {
				return ErrCorrupt
			}
			total += e.Size
			if err = verifyZIPEntry(ctx, file, e, io.Discard); err != nil {
				return err
			}
		default:
			return ErrCorrupt
		}
	}
	if total != manifest.Total {
		return ErrCorrupt
	}
	return nil
}

func verifyZIPEntry(ctx context.Context, file *zip.File, entry Entry, dest io.Writer) error {
	r, err := file.Open()
	if err != nil {
		return err
	}
	defer r.Close()
	h := sha256.New()
	writer := io.MultiWriter(dest, h)
	buffer := make([]byte, chunkBytes)
	var total int64
	for {
		if err = ctx.Err(); err != nil {
			return err
		}
		n, e := r.Read(buffer)
		if n > 0 {
			if total > entry.Size-int64(n) {
				return ErrLimit
			}
			if _, err = writer.Write(buffer[:n]); err != nil {
				return err
			}
			total += int64(n)
		}
		if e == io.EOF {
			break
		}
		if e != nil {
			return errors.Join(ErrCorrupt, e)
		}
	}
	if total != entry.Size || "sha256:"+hex.EncodeToString(h.Sum(nil)) != entry.Hash {
		return ErrCorrupt
	}
	return nil
}

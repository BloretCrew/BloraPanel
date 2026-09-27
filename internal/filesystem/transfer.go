package filesystem

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"
)

type UploadSpec struct {
	ID                 string `json:"id"`
	OwnerID            string `json:"ownerId"`
	Path               string `json:"path"`
	Total              int64  `json:"total"`
	Hash               string `json:"hash"`
	ExpectedVersion    string `json:"expectedVersion"`
	ExpectedObjectID   string `json:"expectedObjectId,omitempty"`
	SourceName         string `json:"sourceName"`
	SourceModified     int64  `json:"sourceModified"`
	SourceFingerprint  string `json:"sourceFingerprint"`
	SourceNodeID       string `json:"sourceNodeId,omitempty"`
	SourceResourceID   string `json:"sourceResourceId,omitempty"`
	SourcePath         string `json:"sourcePath,omitempty"`
	SourceVersion      string `json:"sourceVersion,omitempty"`
	PreserveMetadata   bool   `json:"preserveMetadata,omitempty"`
	SourceMode         uint32 `json:"sourceMode,omitempty"`
	SourceModifiedNano int64  `json:"sourceModifiedNano,omitempty"`
}
type Upload struct {
	Spec           UploadSpec `json:"spec"`
	Offset         int64      `json:"offset"`
	ChunkBytes     int        `json:"chunkBytes"`
	LastChunkHash  string     `json:"lastChunkHash,omitempty"`
	Stage          string     `json:"stage"`
	Version        string     `json:"version,omitempty"`
	CommitObjectID string     `json:"commitObjectId,omitempty"`
	Updated        time.Time  `json:"updated"`
}
type Chunk struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	Hash    string `json:"hash"`
	Offset  int64  `json:"offset"`
	Total   int64  `json:"total"`
	Data    []byte `json:"data"`
	EOF     bool   `json:"eof"`
}

func uploadPart(u Upload) string    { return path.Join(path.Dir(u.Spec.Path), ".blora-upload-"+u.Spec.ID) }
func uploadRecord(id string) string { return "uploads/" + id + ".json" }

func (s *Service) loadUpload(id string) (Upload, error) {
	var u Upload
	if !validID(id) {
		return u, ErrTransfer
	}
	if e := s.readState(uploadRecord(id), &u); e != nil {
		return u, e
	}
	if u.Spec.ID != id || u.ChunkBytes <= 0 || u.ChunkBytes > 256<<10 || u.Offset < 0 || u.Offset > u.Spec.Total || u.Spec.Total > s.opts.MaxFileBytes {
		return u, ErrTransfer
	}
	if e := ValidatePath(u.Spec.Path); e != nil {
		return u, e
	}
	return u, nil
}
func (s *Service) BeginUpload(ctx context.Context, spec UploadSpec) (Upload, error) {
	if spec.PreserveMetadata && (spec.SourceMode > 0777 || spec.SourceModifiedNano <= 0) {
		return Upload{}, ErrTransfer
	}
	if !validID(spec.ID) || spec.OwnerID == "" || spec.SourceFingerprint == "" || len(spec.SourceFingerprint) > 512 || len(spec.OwnerID) > 256 || len(spec.SourceName) > 512 || len(spec.SourcePath) > 4096 || len(spec.SourceResourceID) > 256 || len(spec.SourceNodeID) > 256 || !validHash(spec.Hash) {
		return Upload{}, ErrTransfer
	}
	if spec.Total < 0 || spec.Total > s.opts.MaxFileBytes {
		return Upload{}, ErrLimit
	}
	if e := s.validate(spec.Path, false); e != nil {
		return Upload{}, e
	}
	done, e := s.acquire(ctx, true, spec.Path, "@uploads")
	if e != nil {
		return Upload{}, e
	}
	defer done()
	if old, err := s.loadUpload(spec.ID); err == nil {
		if old.Spec != spec {
			return old, ErrTransfer
		}
		return s.reconcileUpload(ctx, old)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return Upload{}, err
	}
	if e = s.expected(ctx, spec.Path, spec.ExpectedVersion); e != nil {
		return Upload{}, e
	}
	if e = s.expectedObject(spec.Path, spec.ExpectedObjectID); e != nil {
		return Upload{}, e
	}
	f, e := s.state.Open("uploads")
	if e != nil {
		return Upload{}, e
	}
	defer f.Close()
	count := 0
	recordCount := 0
	var reserved int64
	for {
		if e = ctx.Err(); e != nil {
			return Upload{}, e
		}
		names, err := f.Readdirnames(256)
		for _, n := range names {
			if !strings.HasSuffix(n, ".json") {
				continue
			}
			recordCount++
			if recordCount >= s.opts.MaxEntries {
				return Upload{}, ErrLimit
			}
			old, err := s.loadUpload(strings.TrimSuffix(n, ".json"))
			if err != nil {
				return Upload{}, err
			}
			if old.Stage != "committed" && old.Stage != "cancelled" {
				count++
				reserved += old.Spec.Total
			}
		}
		if count >= s.opts.MaxTransfers || reserved > s.opts.MaxTotalBytes-spec.Total {
			return Upload{}, ErrLimit
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return Upload{}, err
		}
	}
	u := Upload{Spec: spec, ChunkBytes: s.opts.ChunkBytes, Stage: "receiving", Updated: time.Now().UTC()}
	// Intent precedes creation. A crash before creation resumes an empty file;
	// only the identifier bound in this checkpoint may be cleaned up.
	if e = s.writeState(uploadRecord(spec.ID), u); e != nil {
		return u, e
	}
	fpart, e := s.root.OpenFile(uploadPart(u), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return u, e
	}
	e = fpart.Sync()
	if ce := fpart.Close(); e == nil {
		e = ce
	}
	if e == nil {
		e = syncDir(s.root, path.Dir(spec.Path))
	}
	return u, e
}

func (s *Service) UploadStatus(ctx context.Context, id string) (Upload, error) {
	if !validID(id) {
		return Upload{}, ErrTransfer
	}
	done, e := s.acquire(ctx, true, "@uploads")
	if e != nil {
		return Upload{}, e
	}
	defer done()
	u, e := s.loadUpload(id)
	if e != nil {
		return u, e
	}
	return s.reconcileUpload(ctx, u)
}
func (s *Service) reconcileUpload(ctx context.Context, u Upload) (Upload, error) {
	if e := ctx.Err(); e != nil {
		return u, e
	}
	if u.Stage == "cancelling" {
		if _, e := s.root.Lstat(uploadPart(u)); errors.Is(e, fs.ErrNotExist) {
			u.Stage = "cancelled"
			u.Updated = time.Now().UTC()
			return u, s.writeState(uploadRecord(u.Spec.ID), u)
		} else if e != nil {
			return u, e
		}
		return u, nil
	}
	if u.Stage == "committing" {
		if _, e := s.root.Lstat(uploadPart(u)); errors.Is(e, fs.ErrNotExist) {
			if u.CommitObjectID == "" {
				return u, ErrTransfer
			}
			if e = s.expectedObject(u.Spec.Path, u.CommitObjectID); e != nil {
				return u, e
			}
			v, fi, e := s.fileVersion(ctx, u.Spec.Path)
			if e != nil {
				return u, e
			}
			if v != u.Spec.Hash || fi.Size() != u.Spec.Total {
				return u, ErrTransfer
			}
			u.Stage = "committed"
			u.Version = v
			u.Updated = time.Now().UTC()
			return u, s.writeState(uploadRecord(u.Spec.ID), u)
		} else if e != nil {
			return u, e
		}
	}
	if u.Stage == "cancelled" || u.Stage == "committed" {
		return u, nil
	}
	f, e := s.root.OpenFile(uploadPart(u), os.O_RDWR|readNonblock, 0600)
	if errors.Is(e, fs.ErrNotExist) && u.Offset == 0 && u.Stage == "receiving" {
		f, e = s.root.OpenFile(uploadPart(u), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	}
	if e != nil {
		return u, e
	}
	defer f.Close()
	fi, e := f.Stat()
	if e != nil {
		return u, e
	}
	if !fi.Mode().IsRegular() || fi.Size() < u.Offset {
		return u, ErrTransfer
	}
	if u.Offset > 0 {
		start := ((u.Offset - 1) / int64(u.ChunkBytes)) * int64(u.ChunkBytes)
		b := make([]byte, u.Offset-start)
		if _, e = f.ReadAt(b, start); e != nil {
			return u, e
		}
		if hashBytes(b) != u.LastChunkHash {
			return u, ErrTransfer
		}
	}
	return u, nil
}

// UploadChunk only acknowledges a contiguous complete chunk after the node's
// data file and atomic checkpoint have both been synced. Retries of an already
// acknowledged chunk must match its durable bytes exactly.
func (s *Service) UploadChunk(ctx context.Context, id string, offset int64, data []byte, hash string) (Upload, error) {
	if !validID(id) || !validHash(hash) || hashBytes(data) != hash {
		return Upload{}, ErrTransfer
	}
	if len(data) > s.opts.ChunkBytes {
		return Upload{}, ErrLimit
	}
	done, e := s.acquire(ctx, true, "@uploads")
	if e != nil {
		return Upload{}, e
	}
	defer done()
	u, e := s.loadUpload(id)
	if e != nil {
		return u, e
	}
	if u.Stage != "receiving" {
		return u, ErrTransfer
	}
	u, e = s.reconcileUpload(ctx, u)
	if e != nil {
		return u, e
	}
	if offset < 0 || offset%int64(u.ChunkBytes) != 0 || offset >= u.Spec.Total || int64(len(data)) != min(int64(u.ChunkBytes), u.Spec.Total-offset) || offset > u.Offset {
		return u, ErrTransfer
	}
	f, e := s.root.OpenFile(uploadPart(u), os.O_RDWR|readNonblock, 0600)
	if e != nil {
		return u, e
	}
	defer f.Close()
	fi, e := f.Stat()
	if e != nil {
		return u, e
	}
	if !fi.Mode().IsRegular() {
		return u, ErrUnsupported
	}
	if offset < u.Offset {
		b := make([]byte, len(data))
		if _, e = f.ReadAt(b, offset); e != nil {
			return u, e
		}
		if hashBytes(b) != hash {
			return u, ErrTransfer
		}
		return u, nil
	}
	if e = ctx.Err(); e != nil {
		return u, e
	}
	if e = f.Truncate(u.Offset); e != nil {
		return u, e
	}
	if _, e = f.WriteAt(data, offset); e != nil {
		return u, e
	}
	if e = f.Sync(); e != nil {
		return u, e
	}
	next := u
	next.Offset += int64(len(data))
	next.LastChunkHash = hash
	next.Updated = time.Now().UTC()
	if e = s.writeState(uploadRecord(id), next); e != nil {
		return u, e
	}
	return next, nil
}

func (s *Service) CommitUpload(ctx context.Context, id string) (Upload, error) {
	if !validID(id) {
		return Upload{}, ErrTransfer
	}
	initial, e := s.loadUpload(id)
	if e != nil {
		return initial, e
	}
	done, e := s.acquire(ctx, true, "@uploads", initial.Spec.Path)
	if e != nil {
		return initial, e
	}
	defer done()
	u, e := s.loadUpload(id)
	if e != nil {
		return u, e
	}
	u, e = s.reconcileUpload(ctx, u)
	if e != nil {
		return u, e
	}
	if u.Stage == "committed" {
		return u, nil
	}
	if u.Stage != "receiving" && u.Stage != "committing" {
		return u, ErrTransfer
	}
	if u.Offset != u.Spec.Total {
		return u, ErrTransfer
	}
	version, fi, e := s.fileVersion(ctx, uploadPart(u))
	if e != nil {
		return u, e
	}
	if version != u.Spec.Hash || fi.Size() != u.Spec.Total {
		return u, ErrTransfer
	}
	if e = s.expected(ctx, u.Spec.Path, u.Spec.ExpectedVersion); e != nil {
		return u, e
	}
	if e = s.expectedObject(u.Spec.Path, u.Spec.ExpectedObjectID); e != nil {
		return u, e
	}
	old, oldErr := s.root.Lstat(u.Spec.Path)
	if u.Spec.PreserveMetadata || oldErr == nil {
		part, err := s.root.OpenFile(uploadPart(u), os.O_WRONLY, 0)
		if err != nil {
			return u, err
		}
		if u.Spec.PreserveMetadata {
			err = part.Chmod(os.FileMode(u.Spec.SourceMode))
			if err == nil {
				err = setFileTimes(part, time.Unix(0, u.Spec.SourceModifiedNano))
			}
		} else {
			err = part.Chmod(old.Mode().Perm())
		}
		if err == nil {
			err = part.Sync()
		}
		ce := part.Close()
		if err == nil {
			err = ce
		}
		if err != nil {
			return u, err
		}
	}
	if e = ctx.Err(); e != nil {
		return u, e
	}
	u.Stage = "committing"
	part, e := s.root.Open(uploadPart(u))
	if e != nil {
		return u, e
	}
	u.CommitObjectID, e = relationObjectID(part)
	closeErr := part.Close()
	if e == nil {
		e = closeErr
	}
	if e != nil {
		return u, e
	}
	u.Updated = time.Now().UTC()
	if e = s.writeState(uploadRecord(id), u); e != nil {
		return u, e
	}
	if e = s.expected(ctx, u.Spec.Path, u.Spec.ExpectedVersion); e != nil {
		return u, e
	}
	if e = ctx.Err(); e != nil {
		return u, e
	}
	if e = s.expectedObject(u.Spec.Path, u.Spec.ExpectedObjectID); e != nil {
		return u, e
	}
	if e = s.root.Rename(uploadPart(u), u.Spec.Path); e != nil {
		return u, e
	}
	if e = syncDir(s.root, path.Dir(u.Spec.Path)); e != nil {
		return u, fmt.Errorf("upload renamed; durability requires reconciliation: %w", e)
	}
	u.Stage = "committed"
	u.Version = version
	u.Updated = time.Now().UTC()
	return u, s.writeState(uploadRecord(id), u)
}
func (s *Service) CancelUpload(ctx context.Context, id string) (Upload, error) {
	if !validID(id) {
		return Upload{}, ErrTransfer
	}
	done, e := s.acquire(ctx, true, "@uploads")
	if e != nil {
		return Upload{}, e
	}
	defer done()
	u, e := s.loadUpload(id)
	if e != nil {
		return u, e
	}
	u, e = s.reconcileUpload(ctx, u)
	if e != nil {
		return u, e
	}
	if u.Stage == "committed" {
		return u, ErrConflict
	}
	if u.Stage == "cancelled" {
		return u, nil
	}
	if e = ctx.Err(); e != nil {
		return u, e
	}
	u.Stage = "cancelling"
	u.Updated = time.Now().UTC()
	if e = s.writeState(uploadRecord(id), u); e != nil {
		return u, e
	}
	if e = s.root.Remove(uploadPart(u)); e != nil && !errors.Is(e, fs.ErrNotExist) {
		return u, e
	}
	if e = syncDir(s.root, path.Dir(u.Spec.Path)); e != nil {
		return u, e
	}
	u.Stage = "cancelled"
	u.Updated = time.Now().UTC()
	return u, s.writeState(uploadRecord(id), u)
}

// ReadChunk rechecks the whole-file version before and after reading. Master
// must use the returned version/total and verify Hash on the independent bulk
// stream. A changed source fails instead of silently mixing file generations.
func (s *Service) ReadChunk(ctx context.Context, p, version string, offset int64, limit int) (Chunk, error) {
	out := Chunk{Path: p, Offset: offset}
	if e := s.validate(p, false); e != nil {
		return out, e
	}
	if !validHash(version) {
		return out, ErrConflict
	}
	if offset < 0 || limit <= 0 || limit > s.opts.ChunkBytes {
		return out, ErrLimit
	}
	done, e := s.acquire(ctx, false, p)
	if e != nil {
		return out, e
	}
	defer done()
	v, fi, e := s.fileVersion(ctx, p)
	if e != nil {
		return out, e
	}
	if v != version {
		return out, ErrConflict
	}
	if offset > fi.Size() {
		return out, ErrTransfer
	}
	f, _, e := s.openRegular(p)
	if e != nil {
		return out, e
	}
	defer f.Close()
	out.Data = make([]byte, min(int64(limit), fi.Size()-offset))
	if len(out.Data) > 0 {
		if _, e = f.ReadAt(out.Data, offset); e != nil {
			return out, e
		}
	}
	v, after, e := s.fileVersion(ctx, p)
	if e != nil {
		return out, e
	}
	if v != version || after.Size() != fi.Size() {
		return out, ErrConflict
	}
	out.Version = v
	out.Total = fi.Size()
	out.Hash = hashBytes(out.Data)
	out.EOF = offset+int64(len(out.Data)) == fi.Size()
	return out, ctx.Err()
}

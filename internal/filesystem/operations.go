package filesystem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

// tree hashes all regular-file contents and directory metadata within bounded
// limits. It refuses links and special files, including nested ones.
func (s *Service) tree(ctx context.Context, p string) ([]Entry, string, error) {
	var items []Entry
	var bytes int64
	var walk func(string, int) error
	walk = func(p string, depth int) error {
		if e := ctx.Err(); e != nil {
			return e
		}
		if depth > 128 {
			return ErrLimit
		}
		fi, e := s.root.Lstat(p)
		if e != nil {
			return e
		}
		out := entry(p, fi)
		if fi.IsDir() {
			items = append(items, out)
			if len(items) > s.opts.MaxEntries {
				return ErrLimit
			}
			children, _, e := s.directory(ctx, p)
			if e != nil {
				return e
			}
			for _, child := range children {
				if e = ValidatePath(child.Name); e != nil {
					return e
				}
				if e = walk(child.Path, depth+1); e != nil {
					return e
				}
			}
		} else {
			if !fi.Mode().IsRegular() {
				return ErrUnsupported
			}
			v, _, e := s.fileVersion(ctx, p)
			if e != nil {
				return e
			}
			out.Version = v
			bytes += fi.Size()
			if bytes > s.opts.MaxTotalBytes {
				return ErrLimit
			}
			items = append(items, out)
			if len(items) > s.opts.MaxEntries {
				return ErrLimit
			}
		}
		return nil
	}
	if e := walk(p, 0); e != nil {
		return items, "", e
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
	// Relative names make the content fingerprint stable across a tree rename.
	fingerprint := make([]Entry, len(items))
	copy(fingerprint, items)
	for i := range fingerprint {
		fingerprint[i].Path = strings.TrimPrefix(strings.TrimPrefix(fingerprint[i].Path, p), "/")
		fingerprint[i].Name = path.Base(fingerprint[i].Path)
	}
	b, _ := json.Marshal(fingerprint)
	return items, hashBytes(b), nil
}

func (s *Service) Copy(ctx context.Context, source, target, sourceVersion, targetVersion string) (Result, error) {
	if e := s.validatePair(source, target); e != nil {
		return Result{}, e
	}
	done, e := s.acquire(ctx, true, source, target)
	if e != nil {
		return Result{}, e
	}
	defer done()
	return s.copy(ctx, source, target, sourceVersion, targetVersion)
}
func (s *Service) validatePair(source, target string) error {
	if e := s.validate(source, false); e != nil {
		return e
	}
	if e := s.validate(target, false); e != nil {
		return e
	}
	if related(source, target) {
		return ErrPath
	}
	return nil
}
func (s *Service) copy(ctx context.Context, source, target, sourceVersion, targetVersion string) (result Result, err error) {
	result = Result{Source: source, Target: target, Stage: "validating", Completed: []string{}}
	if err = s.expected(ctx, source, sourceVersion); err != nil {
		return
	}
	if err = s.expected(ctx, target, targetVersion); err != nil {
		return
	}
	items, _, e := s.tree(ctx, source)
	if e != nil {
		err = e
		return
	}
	if fi, e := s.root.Lstat(target); e == nil && fi.IsDir() {
		err = fmt.Errorf("directory replacement requires an empty new destination: %w", ErrConflict)
		return
	}
	tmp := path.Join(path.Dir(target), ".blora-copy-"+randomID())
	result.Cleanup = []string{tmp}
	defer func() {
		if e := s.root.RemoveAll(tmp); e != nil {
			err = errors.Join(err, fmt.Errorf("cleanup %s: %w", tmp, e))
		} else {
			result.Cleanup = nil
		}
	}()
	result.Stage = "copying"
	for _, item := range items {
		if err = ctx.Err(); err != nil {
			return
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(item.Path, source), "/")
		dest := tmp
		if rel != "" {
			dest = path.Join(tmp, rel)
		}
		if item.Kind == "directory" {
			err = s.root.Mkdir(dest, 0750)
			if err != nil {
				return
			}
			continue
		}
		var n int64
		n, err = s.copyFile(ctx, item.Path, dest, item.Version, os.FileMode(item.Mode))
		result.Bytes += n
		if err != nil {
			return
		}
	}
	// Restore directory metadata after children and sync before publication.
	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		rel := strings.TrimPrefix(strings.TrimPrefix(item.Path, source), "/")
		dest := tmp
		if rel != "" {
			dest = path.Join(tmp, rel)
		}
		if err = s.applyMetadata(dest, item); err != nil {
			return
		}
	}
	result.Stage = "verifying"
	if err = s.expected(ctx, source, sourceVersion); err != nil {
		return
	}
	if err = s.expected(ctx, target, targetVersion); err != nil {
		return
	}
	if err = ctx.Err(); err != nil {
		return
	}
	if err = s.root.Rename(tmp, target); err != nil {
		return
	}
	result.Stage = "committed"
	result.Completed = []string{target}
	result.Cleanup = nil
	if err = syncDir(s.root, path.Dir(target)); err != nil {
		return
	}
	var out Entry
	out, err = s.stat(ctx, target)
	result.Version = out.Version
	return
}
func (s *Service) applyMetadata(p string, item Entry) error {
	return applyPathMetadata(s.root, p, item)
}

func (s *Service) syncDirectories(ctx context.Context, p string) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	f, e := s.root.OpenFile(p, os.O_RDONLY|readNonblock, 0)
	if e != nil {
		return e
	}
	defer f.Close()
	fi, e := f.Stat()
	if e != nil {
		return e
	}
	if !fi.IsDir() {
		return ErrUnsupported
	}
	for {
		items, err := f.ReadDir(128)
		for _, item := range items {
			if item.IsDir() {
				if e = s.syncDirectories(ctx, path.Join(p, item.Name())); e != nil {
					return e
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return syncDir(s.root, p)
}

func (s *Service) mkdirReported(ctx context.Context, p string, result *Result) error {
	if p == "." {
		return nil
	}
	parts := strings.Split(p, "/")
	for i := range parts {
		if e := ctx.Err(); e != nil {
			return e
		}
		dir := strings.Join(parts[:i+1], "/")
		fi, e := s.root.Lstat(dir)
		if e == nil {
			if !fi.IsDir() {
				return ErrUnsupported
			}
			continue
		}
		if !errors.Is(e, fs.ErrNotExist) {
			return e
		}
		if e = s.root.Mkdir(dir, 0750); e != nil {
			return e
		}
		result.Completed = append(result.Completed, dir)
		if e = syncDir(s.root, path.Dir(dir)); e != nil {
			return e
		}
	}
	return nil
}

func (s *Service) copyFile(ctx context.Context, src, dest, version string, mode os.FileMode) (int64, error) {
	f, _, e := s.openRegular(src)
	if e != nil {
		return 0, e
	}
	defer f.Close()
	out, e := s.root.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return 0, e
	}
	n, e := copyBounded(ctx, out, f, s.opts.MaxFileBytes)
	if e == nil {
		e = out.Chmod(mode.Perm())
	}
	if e == nil {
		e = out.Sync()
	}
	if ce := out.Close(); e == nil {
		e = ce
	}
	if e != nil {
		return n, e
	}
	v, _, e := s.fileVersion(ctx, dest)
	if e != nil {
		return n, e
	}
	if v != version {
		return n, ErrConflict
	}
	return n, nil
}

// Move tries the filesystem rename first. A cross-filesystem fallback exposes
// the copy/verify/source-delete stages and never deletes a changed source.
func (s *Service) Move(ctx context.Context, source, target, sourceVersion, targetVersion string) (Result, error) {
	r := Result{Source: source, Target: target, Stage: "validating", Completed: []string{}}
	if e := s.validatePair(source, target); e != nil {
		return r, e
	}
	done, e := s.acquire(ctx, true, source, target, "@trash", "@uploads")
	if e != nil {
		return r, e
	}
	defer done()
	if e = s.checkPinned(ctx, source); e != nil {
		return r, e
	}
	if e = s.expected(ctx, source, sourceVersion); e != nil {
		return r, e
	}
	if e = s.expected(ctx, target, targetVersion); e != nil {
		return r, e
	}
	if e = ctx.Err(); e != nil {
		return r, e
	}
	if e = s.root.Rename(source, target); e == nil {
		r.Stage = "committed"
		r.Completed = []string{target}
		r.SourceDeleted = true
		r.Version = sourceVersion
		return r, errors.Join(syncDir(s.root, path.Dir(source)), syncDir(s.root, path.Dir(target)))
	} else if !crossDevice(e) {
		return r, e
	}
	r, e = s.copy(ctx, source, target, sourceVersion, targetVersion)
	if e != nil {
		return r, e
	}
	r.Stage = "verifying_source"
	if e = s.expected(ctx, source, sourceVersion); e != nil {
		return r, e
	}
	if e = ctx.Err(); e != nil {
		return r, e
	}
	r.Stage = "deleting_source"
	_, e = s.delete(ctx, source, sourceVersion)
	if e != nil {
		return r, e
	}
	r.SourceDeleted = true
	r.Stage = "committed"
	return r, nil
}

type TrashItem struct {
	ID           string    `json:"id"`
	OriginalPath string    `json:"originalPath"`
	Version      string    `json:"version"`
	Size         int64     `json:"size"`
	DeletedAt    time.Time `json:"deletedAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
	Stage        string    `json:"stage"`
	RestoredPath string    `json:"restoredPath,omitempty"`
	// StoragePath is daemon generated. Cross-filesystem recycling uses a
	// hidden sibling on the source filesystem so deletion remains a rename.
	StoragePath string `json:"storagePath"`
}

func trashPath(t TrashItem) (string, error) {
	regular := privateDir + "/trash/" + t.ID
	if !validID(t.ID) || ValidatePath(t.OriginalPath) != nil {
		return "", ErrTransfer
	}
	if t.StoragePath == "" || t.StoragePath == regular {
		return regular, nil
	}
	if t.StoragePath != path.Join(path.Dir(t.OriginalPath), ".blora-trash-"+t.ID) {
		return "", ErrTransfer
	}
	return t.StoragePath, nil
}

// An unfinished transfer or a recycle object on a separate filesystem pins
// its parent path. Moving that ancestor would orphan its durable checkpoint.
func (s *Service) checkPinned(ctx context.Context, p string) error {
	for _, dir := range []string{"uploads", "trash"} {
		f, e := s.state.Open(dir)
		if e != nil {
			return e
		}
		count := 0
		for {
			if e = ctx.Err(); e != nil {
				f.Close()
				return e
			}
			names, err := f.Readdirnames(256)
			for _, name := range names {
				if !strings.HasSuffix(name, ".json") {
					continue
				}
				count++
				if count > s.opts.MaxEntries {
					f.Close()
					return ErrLimit
				}
				var pinned string
				if dir == "uploads" {
					u, e := s.loadUpload(strings.TrimSuffix(name, ".json"))
					if e != nil {
						f.Close()
						return e
					}
					if u.Stage != "committed" && u.Stage != "cancelled" {
						pinned = u.Spec.Path
					}
				} else {
					var t TrashItem
					if e = s.readState(dir+"/"+name, &t); e != nil {
						f.Close()
						return e
					}
					if t.Stage == "prepared" || t.Stage == "trashed" {
						pinned, e = trashPath(t)
						if e != nil {
							f.Close()
							return e
						}
					}
				}
				if pinned != "" && (p == pinned || strings.HasPrefix(pinned, p+"/")) {
					f.Close()
					return fmt.Errorf("path has a resumable transfer or recycle object: %w", ErrConflict)
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				f.Close()
				return err
			}
		}
		f.Close()
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, p, expectedVersion string) (TrashItem, error) {
	return s.DeleteChecked(ctx, p, expectedVersion, "")
}

// DeleteChecked additionally binds the source filesystem object for relay
// moves; identical content at a replaced source path is not sufficient.
func (s *Service) DeleteChecked(ctx context.Context, p, expectedVersion, expectedObjectID string) (TrashItem, error) {
	if e := s.validate(p, false); e != nil {
		return TrashItem{}, e
	}
	done, e := s.acquire(ctx, true, p, "@trash", "@uploads")
	if e != nil {
		return TrashItem{}, e
	}
	defer done()
	if e = s.checkPinned(ctx, p); e != nil {
		return TrashItem{}, e
	}
	return s.delete(ctx, p, expectedVersion, expectedObjectID)
}
func (s *Service) delete(ctx context.Context, p, version string, objectIDs ...string) (TrashItem, error) {
	objectID := ""
	if len(objectIDs) > 0 {
		objectID = objectIDs[0]
	}
	if err := s.expectedObject(p, objectID); err != nil {
		return TrashItem{}, err
	}
	if e := s.expected(ctx, p, version); e != nil {
		return TrashItem{}, e
	}
	items, _, e := s.tree(ctx, p)
	if e != nil {
		return TrashItem{}, e
	}
	var size int64
	for _, i := range items {
		if i.Kind == "file" {
			size += i.Size
		}
	}
	trash, e := s.trash(ctx)
	if e != nil {
		return TrashItem{}, e
	}
	usage := size
	for _, t := range trash {
		if t.Stage == "trashed" || t.Stage == "prepared" {
			usage += t.Size
		}
	}
	if usage > s.opts.TrashBytes || len(trash) >= s.opts.MaxEntries {
		return TrashItem{}, ErrLimit
	}
	now := time.Now().UTC()
	t := TrashItem{ID: randomID(), OriginalPath: p, Version: version, Size: size, DeletedAt: now, ExpiresAt: now.Add(s.opts.TrashRetention), Stage: "prepared"}
	t.StoragePath = privateDir + "/trash/" + t.ID
	if e = s.writeState("trash/"+t.ID+".json", t); e != nil {
		return t, e
	}
	if e = s.expected(ctx, p, version); e != nil {
		return t, e
	}
	if e = ctx.Err(); e != nil {
		return t, e
	}
	dest := t.StoragePath
	if e = s.expectedObject(p, objectID); e != nil {
		return t, e
	}
	if e = s.root.Rename(p, dest); e != nil {
		// If the daemon cannot create/use the root-level private trash directory,
		// retain the recycle object beside its source on the same filesystem.
		if !crossDevice(e) && !errors.Is(e, fs.ErrNotExist) && !errors.Is(e, fs.ErrPermission) {
			return t, e
		}
		t.StoragePath = path.Join(path.Dir(p), ".blora-trash-"+t.ID)
		dest = t.StoragePath
		if e = s.writeState("trash/"+t.ID+".json", t); e != nil {
			return t, e
		}
		if e = s.expected(ctx, p, version); e != nil {
			return t, e
		}
		if e = ctx.Err(); e != nil {
			return t, e
		}
		if e = s.expectedObject(p, objectID); e != nil {
			return t, e
		}
		if e = s.root.Rename(p, dest); e != nil {
			return t, e
		}
	}
	if e = s.expectedObject(dest, objectID); e != nil {
		return t, e
	}
	t.Stage = "trashed"
	if e = errors.Join(syncDir(s.root, path.Dir(p)), syncDir(s.root, path.Dir(dest))); e != nil {
		return t, e
	}
	return t, s.writeState("trash/"+t.ID+".json", t)
}

func (s *Service) Trash(ctx context.Context) ([]TrashItem, error) {
	done, e := s.acquire(ctx, false, "@trash")
	if e != nil {
		return nil, e
	}
	defer done()
	return s.trash(ctx)
}

type TrashPage struct {
	Items      []TrashItem `json:"items"`
	Total      int         `json:"total"`
	NextOffset int         `json:"nextOffset"`
}

func (s *Service) ListTrash(ctx context.Context, offset, limit int) (TrashPage, error) {
	if offset < 0 || limit < 0 || limit > 1000 {
		return TrashPage{}, ErrLimit
	}
	if limit == 0 {
		limit = 100
	}
	items, e := s.Trash(ctx)
	if e != nil {
		return TrashPage{}, e
	}
	if offset > len(items) {
		return TrashPage{}, ErrConflict
	}
	end := min(offset+limit, len(items))
	next := -1
	if end < len(items) {
		next = end
	}
	return TrashPage{items[offset:end], len(items), next}, nil
}
func (s *Service) trash(ctx context.Context) ([]TrashItem, error) {
	f, e := s.state.Open("trash")
	if e != nil {
		return nil, e
	}
	defer f.Close()
	out := []TrashItem{}
	for {
		if e = ctx.Err(); e != nil {
			return out, e
		}
		names, err := f.Readdirnames(256)
		for _, name := range names {
			if !strings.HasSuffix(name, ".json") {
				continue
			}
			id := strings.TrimSuffix(name, ".json")
			if !validID(id) {
				return out, ErrTransfer
			}
			var t TrashItem
			if e = s.readState("trash/"+name, &t); e != nil {
				return out, e
			}
			if t.ID != id {
				return out, ErrTransfer
			}
			stored, e := trashPath(t)
			if e != nil {
				return out, e
			}
			if t.Stage == "prepared" {
				if _, err := s.root.Lstat(stored); err == nil {
					t.Stage = "trashed"
				} else if errors.Is(err, fs.ErrNotExist) {
					continue
				} else {
					return out, err
				}
			}
			if t.Stage == "trashed" {
				out = append(out, t)
			}
			if len(out) > s.opts.MaxEntries {
				return out, ErrLimit
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return out, err
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DeletedAt.Before(out[j].DeletedAt) })
	return out, nil
}
func (s *Service) Restore(ctx context.Context, id, target, targetVersion string) (Result, error) {
	r := Result{Target: target, Stage: "validating", Completed: []string{}}
	if !validID(id) {
		return r, ErrTransfer
	}
	if e := s.validate(target, false); e != nil {
		return r, e
	}
	done, e := s.acquire(ctx, true, target, "@trash")
	if e != nil {
		return r, e
	}
	defer done()
	var t TrashItem
	if e = s.readState("trash/"+id+".json", &t); e != nil {
		return r, e
	}
	if t.ID != id || (t.Stage != "trashed" && t.Stage != "prepared") {
		return r, ErrTransfer
	}
	src, e := trashPath(t)
	if e != nil {
		return r, e
	}
	if e = s.expected(ctx, src, t.Version); e != nil {
		return r, e
	}
	if e = s.expected(ctx, target, targetVersion); e != nil {
		return r, e
	}
	if e = ctx.Err(); e != nil {
		return r, e
	}
	if e = s.root.Rename(src, target); e != nil {
		return r, e
	}
	r.Source = src
	r.Stage = "committed"
	r.Completed = []string{target}
	r.Version = t.Version
	r.Bytes = t.Size
	t.Stage = "restored"
	t.RestoredPath = target
	if e = errors.Join(syncDir(s.root, path.Dir(target)), syncDir(s.root, path.Dir(src))); e != nil {
		return r, e
	}
	return r, s.writeState("trash/"+id+".json", t)
}

// PurgeExpired only removes expired objects named by this service's own trash
// manifests. It reports each completed removal if cancellation or I/O interrupts.
func (s *Service) PurgeExpired(ctx context.Context, now time.Time) (Result, error) {
	r := Result{Stage: "purging", Completed: []string{}}
	done, e := s.acquire(ctx, true, "@trash")
	if e != nil {
		return r, e
	}
	defer done()
	items, e := s.trash(ctx)
	if e != nil {
		return r, e
	}
	for _, t := range items {
		if t.ExpiresAt.After(now) {
			continue
		}
		if e = ctx.Err(); e != nil {
			return r, e
		}
		stored, e := trashPath(t)
		if e != nil {
			return r, e
		}
		if e = s.expected(ctx, stored, t.Version); e != nil {
			return r, e
		}
		if e = s.removeTree(ctx, stored); e != nil {
			return r, e
		}
		r.Completed = append(r.Completed, t.ID)
		r.Bytes += t.Size
		if e = syncDir(s.root, path.Dir(stored)); e != nil {
			return r, e
		}
		t.Stage = "purged"
		if e = s.writeState("trash/"+t.ID+".json", t); e != nil {
			return r, e
		}
	}
	r.Stage = "committed"
	return r, syncDir(s.state, "trash")
}
func (s *Service) removeTree(ctx context.Context, p string) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	fi, e := s.root.Lstat(p)
	if e != nil {
		return e
	}
	if fi.IsDir() {
		f, e := s.root.OpenFile(p, os.O_RDONLY|readNonblock, 0)
		if e != nil {
			return e
		}
		for {
			names, err := f.Readdirnames(128)
			for _, n := range names {
				if e = s.removeTree(ctx, path.Join(p, n)); e != nil {
					f.Close()
					return e
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				f.Close()
				return err
			}
		}
		if e = f.Close(); e != nil {
			return e
		}
	}
	return s.root.Remove(p)
}

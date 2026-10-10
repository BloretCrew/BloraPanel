package filesystem

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
)

type limitedWriter struct {
	w    io.Writer
	left int64
}

// checkZipIndex limits allocation before archive/zip parses the central index.
// This service accepts single-disk, non-ZIP64 archives with at most 60,000
// records. ZIP64 and self-extracting archives require a separate capability.
func checkZipIndex(f *os.File, size int64, maxEntries int) error {
	if size < 22 {
		return zip.ErrFormat
	}
	tail := make([]byte, min(int64(65557), size))
	if _, e := f.ReadAt(tail, size-int64(len(tail))); e != nil {
		return e
	}
	end := -1
	for i := len(tail) - 22; i >= 0; i-- {
		if binary.LittleEndian.Uint32(tail[i:]) == 0x06054b50 && i+22+int(binary.LittleEndian.Uint16(tail[i+20:])) == len(tail) {
			end = i
			break
		}
	}
	if end < 0 {
		return zip.ErrFormat
	}
	b := tail[end:]
	if binary.LittleEndian.Uint16(b[4:]) != 0 || binary.LittleEndian.Uint16(b[6:]) != 0 {
		return ErrUnsupported
	}
	count := int(binary.LittleEndian.Uint16(b[10:]))
	dirSize := int64(binary.LittleEndian.Uint32(b[12:]))
	off := int64(binary.LittleEndian.Uint32(b[16:]))
	eocd := size - int64(len(tail)) + int64(end)
	if count == 65535 || dirSize == 0xffffffff || off == 0xffffffff {
		return ErrUnsupported
	}
	if count > min(maxEntries, 60000) || dirSize > 32<<20 {
		return ErrLimit
	}
	if binary.LittleEndian.Uint16(b[8:]) != uint16(count) || off+dirSize != eocd {
		return zip.ErrFormat
	}
	var hdr [46]byte
	pos := off
	actual := 0
	for pos < off+dirSize {
		if actual >= min(maxEntries, 60000) {
			return ErrLimit
		}
		if _, e := f.ReadAt(hdr[:], pos); e != nil {
			return e
		}
		if binary.LittleEndian.Uint32(hdr[:]) != 0x02014b50 {
			return zip.ErrFormat
		}
		record := int64(46) + int64(binary.LittleEndian.Uint16(hdr[28:])) + int64(binary.LittleEndian.Uint16(hdr[30:])) + int64(binary.LittleEndian.Uint16(hdr[32:]))
		pos += record
		actual++
	}
	if pos != off+dirSize || actual != count {
		return zip.ErrFormat
	}
	return nil
}

// CheckZIPIndex applies the service's allocation/index budget to a backup's
// already-open archive before archive/zip allocates its central directory.
func CheckZIPIndex(f *os.File, size int64, maxEntries int) error {
	return checkZipIndex(f, size, maxEntries)
}

func (w *limitedWriter) Write(b []byte) (int, error) {
	if int64(len(b)) > w.left {
		return 0, ErrLimit
	}
	n, e := w.w.Write(b)
	w.left -= int64(n)
	return n, e
}

// Compress creates a ZIP containing source (including its base name), checks
// every file's streamed digest, and atomically replaces the destination.
func (s *Service) Compress(ctx context.Context, source, target, sourceVersion, targetVersion string) (result Result, err error) {
	result = Result{Source: source, Target: target, Stage: "validating", Completed: []string{}}
	if err = s.validatePair(source, target); err != nil {
		return
	}
	done, e := s.acquire(ctx, true, source, target)
	if e != nil {
		err = e
		return
	}
	defer done()
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
	if len(items) > 60000 {
		err = ErrLimit
		return
	}
	tmp := path.Join(path.Dir(target), ".blora-zip-"+randomID())
	out, e := s.root.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		err = e
		return
	}
	result.Cleanup = []string{tmp}
	defer func() {
		out.Close()
		if e := s.root.Remove(tmp); e != nil && !errors.Is(e, fs.ErrNotExist) {
			err = errors.Join(err, e)
		} else {
			result.Cleanup = nil
		}
	}()
	zw := zip.NewWriter(&limitedWriter{w: out, left: s.opts.MaxFileBytes})
	result.Stage = "compressing"
	for _, item := range items {
		if err = ctx.Err(); err != nil {
			zw.Close()
			return
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(item.Path, source), "/")
		name := path.Base(source)
		if rel != "" {
			name = path.Join(name, rel)
		}
		h := &zip.FileHeader{Name: name, Method: zip.Deflate, Modified: item.Modified}
		h.SetMode(os.FileMode(item.Mode))
		if item.Kind == "directory" {
			h.Name += "/"
			h.SetMode(os.ModeDir | os.FileMode(item.Mode))
		}
		var w io.Writer
		w, err = zw.CreateHeader(h)
		if err != nil {
			zw.Close()
			return
		}
		if item.Kind == "directory" {
			continue
		}
		var f *os.File
		f, _, err = s.openRegular(item.Path)
		if err != nil {
			zw.Close()
			return
		}
		digest := sha256.New()
		var n int64
		n, err = copyBounded(ctx, io.MultiWriter(w, digest), f, s.opts.MaxFileBytes)
		f.Close()
		result.Bytes += n
		if err != nil {
			zw.Close()
			return
		}
		if "sha256:"+hex.EncodeToString(digest.Sum(nil)) != item.Version {
			err = ErrConflict
			zw.Close()
			return
		}
	}
	if err = zw.Close(); err != nil {
		return
	}
	if err = out.Sync(); err != nil {
		return
	}
	if err = out.Close(); err != nil {
		return
	}
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
	result.Completed = []string{target}
	result.Stage = "committed"
	if err = syncDir(s.root, path.Dir(target)); err != nil {
		return
	}
	var ent Entry
	ent, err = s.stat(ctx, target)
	result.Version = ent.Version
	return
}

type ExtractOptions struct {
	// TargetVersion binds the initial destination directory (or "missing").
	TargetVersion string `json:"targetVersion"`
	// Overwrite names are archive-relative and each value is the explicit
	// current version. Missing entries are always conflict, never overwrite.
	Overwrite map[string]string `json:"overwrite,omitempty"`
}
type archiveItem struct {
	name      string
	file      *zip.File
	directory bool
}

func (s *Service) Extract(ctx context.Context, archive, target, archiveVersion string, opts ExtractOptions) (result Result, err error) {
	result = Result{Source: archive, Target: target, Stage: "validating", Completed: []string{}}
	if err = s.validate(archive, false); err != nil {
		return
	}
	if err = s.validate(target, true); err != nil {
		return
	}
	done, e := s.acquire(ctx, true, archive, target)
	if e != nil {
		err = e
		return
	}
	defer done()
	if err = s.expected(ctx, archive, archiveVersion); err != nil {
		return
	}
	if err = s.expected(ctx, target, opts.TargetVersion); err != nil {
		return
	}
	if fi, e := s.root.Lstat(target); e == nil && !fi.IsDir() {
		err = ErrUnsupported
		return
	}
	f, info, e := s.openRegular(archive)
	if e != nil {
		err = e
		return
	}
	defer f.Close()
	if err = checkZipIndex(f, info.Size(), s.opts.MaxEntries); err != nil {
		return
	}
	zr, e := zip.NewReader(f, info.Size())
	if e != nil {
		err = e
		return
	}
	if len(zr.File) > s.opts.MaxEntries {
		err = ErrLimit
		return
	}
	items := make([]archiveItem, 0, len(zr.File))
	seen := map[string]bool{}
	allPaths := map[string]string{}
	var total int64
	for _, zf := range zr.File {
		if err = ctx.Err(); err != nil {
			return
		}
		name := strings.TrimSuffix(zf.Name, "/")
		if err = ValidatePath(name); err != nil {
			return
		}
		if name == "." || len(strings.Split(name, "/")) > 128 {
			err = ErrPath
			return
		}
		key := strings.ToLower(name)
		for p := name; p != "."; p = path.Dir(p) {
			folded := strings.ToLower(p)
			if existing, ok := allPaths[folded]; ok && existing != p {
				err = ErrConflict
				return
			}
			allPaths[folded] = p
			if len(allPaths) > s.opts.MaxEntries {
				err = ErrLimit
				return
			}
		}
		if _, ok := seen[key]; ok {
			err = fmt.Errorf("duplicate archive path %q: %w", name, ErrConflict)
			return
		}
		directory := zf.FileInfo().IsDir()
		seen[key] = directory
		if (directory && !strings.HasSuffix(zf.Name, "/")) || (!directory && !zf.Mode().IsRegular()) || (zf.Mode().Type() != 0 && zf.Mode().Type() != os.ModeDir) {
			err = ErrUnsupported
			return
		}
		if zf.UncompressedSize64 > uint64(s.opts.MaxFileBytes) || zf.UncompressedSize64 > uint64(s.opts.MaxTotalBytes-total) {
			err = ErrLimit
			return
		}
		total += int64(zf.UncompressedSize64)
		if zf.UncompressedSize64 > 1<<20 && (zf.CompressedSize64 == 0 || zf.UncompressedSize64/zf.CompressedSize64 > 1000) {
			err = fmt.Errorf("archive expansion ratio exceeds 1000: %w", ErrLimit)
			return
		}
		dest := path.Join(target, name)
		if dest == archive {
			err = ErrPath
			return
		}
		if err = s.validate(dest, false); err != nil {
			return
		}
		if fi, e := s.root.Lstat(dest); e == nil {
			if directory {
				if !fi.IsDir() {
					err = ErrConflict
					return
				}
			} else {
				if err = s.expected(ctx, dest, opts.Overwrite[name]); err != nil {
					return
				}
			}
		} else if !errors.Is(e, fs.ErrNotExist) {
			err = e
			return
		}
		items = append(items, archiveItem{name, zf, directory})
	}
	for name := range seen {
		parent := path.Dir(name)
		for parent != "." {
			if isdir, ok := seen[parent]; ok && !isdir {
				err = ErrConflict
				return
			}
			parent = path.Dir(parent)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].name < items[j].name })
	stage := path.Join(path.Dir(target), ".blora-extract-"+randomID())
	if err = s.root.Mkdir(stage, 0700); err != nil {
		return
	}
	result.Cleanup = []string{stage}
	defer func() {
		if e := s.root.RemoveAll(stage); e != nil {
			err = errors.Join(err, e)
		} else {
			result.Cleanup = nil
		}
	}()
	result.Stage = "extracting"
	for _, item := range items {
		if err = ctx.Err(); err != nil {
			return
		}
		dest := path.Join(stage, item.name)
		if err = s.root.MkdirAll(path.Dir(dest), 0750); err != nil {
			return
		}
		if item.directory {
			if err = s.root.MkdirAll(dest, 0750); err != nil {
				return
			}
			continue
		}
		var reader io.ReadCloser
		reader, err = item.file.Open()
		if err != nil {
			return
		}
		var out *os.File
		out, err = s.root.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			reader.Close()
			return
		}
		var n int64
		n, err = copyBounded(ctx, out, reader, int64(item.file.UncompressedSize64))
		result.Bytes += n
		reader.Close()
		if err == nil && uint64(n) != item.file.UncompressedSize64 {
			err = ErrTransfer
		}
		if err == nil {
			err = out.Chmod(item.file.Mode().Perm() & 0777)
		}
		if err == nil && !item.file.Modified.IsZero() {
			err = setModified(out, item.file.Modified)
		}
		if err == nil {
			err = out.Sync()
		}
		if e := out.Close(); err == nil {
			err = e
		}
		if err != nil {
			return
		}
	}
	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		if !item.directory {
			continue
		}
		if !item.file.Modified.IsZero() {
			if err = s.applyMetadata(path.Join(stage, item.name), Entry{Mode: uint32(item.file.Mode().Perm()), Modified: item.file.Modified}); err != nil {
				return
			}
		}
	}
	if err = s.syncDirectories(ctx, stage); err != nil {
		return
	}
	if err = s.expected(ctx, archive, archiveVersion); err != nil {
		return
	}
	if err = s.expected(ctx, target, opts.TargetVersion); err != nil {
		return
	}
	result.Stage = "committing"
	if _, e := s.root.Lstat(target); errors.Is(e, fs.ErrNotExist) {
		if err = ctx.Err(); err != nil {
			return
		}
		if err = s.root.Rename(stage, target); err != nil {
			return
		}
		result.Completed = []string{target}
		if err = syncDir(s.root, path.Dir(target)); err != nil {
			return
		}
	} else {
		for _, item := range items {
			if err = ctx.Err(); err != nil {
				return
			}
			dest := path.Join(target, item.name)
			if item.directory {
				if err = s.mkdirReported(ctx, dest, &result); err != nil {
					return
				}
				continue
			}
			if err = s.mkdirReported(ctx, path.Dir(dest), &result); err != nil {
				return
			}
			expected := opts.Overwrite[item.name]
			if expected == "" {
				expected = MissingVersion
			}
			if err = s.expected(ctx, dest, expected); err != nil {
				return
			}
			if err = s.root.Rename(path.Join(stage, item.name), dest); err != nil {
				return
			}
			result.Completed = append(result.Completed, dest)
			if err = syncDir(s.root, path.Dir(dest)); err != nil {
				return
			}
		}
	}
	result.Stage = "committed"
	var ent Entry
	ent, err = s.stat(ctx, target)
	result.Version = ent.Version
	return
}

package filesystem

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func testService(t *testing.T, options Options) (*Service, string, string) {
	t.Helper()
	base := t.TempDir()
	root := filepath.Join(base, "root")
	state := filepath.Join(base, "private")
	if e := os.Mkdir(root, 0700); e != nil {
		t.Fatal(e)
	}
	options.StateDir = state
	s, e := New(root, options)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { s.Close() })
	return s, root, state
}
func put(t *testing.T, root, p, content string) {
	t.Helper()
	name := filepath.Join(root, filepath.FromSlash(p))
	if e := os.MkdirAll(filepath.Dir(name), 0750); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(name, []byte(content), 0640); e != nil {
		t.Fatal(e)
	}
}
func version(t *testing.T, s *Service, p string) string {
	t.Helper()
	v, e := s.Stat(context.Background(), p)
	if e != nil {
		t.Fatal(e)
	}
	return v.Version
}
func content(t *testing.T, root, p, want string) {
	t.Helper()
	b, e := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
	if e != nil || string(b) != want {
		t.Fatalf("content %s = %q, %v; want %q", p, b, e, want)
	}
}
func missing(t *testing.T, root, p string) {
	t.Helper()
	if _, e := os.Lstat(filepath.Join(root, filepath.FromSlash(p))); !errors.Is(e, fs.ErrNotExist) {
		t.Fatalf("%s exists/error %v", p, e)
	}
}

func TestPathAndPrivateStateBoundary(t *testing.T) {
	s, root, _ := testService(t, Options{})
	ctx := context.Background()
	for _, p := range []string{"", "/etc/passwd", "..", "a/../b", "a//b", "a/./b", "C:/Windows", "C:relative", `a\b`, `\\server\share`, "file:stream", "NUL", "CON.txt", "COM1", "LPT²", "name.", "name ", "a\x00b", "a/.blora-files/secret", ".BLORA-upload-other", "a?b"} {
		t.Run(strings.ReplaceAll(p, "/", "_"), func(t *testing.T) {
			if e := ValidatePath(p); !errors.Is(e, ErrPath) {
				t.Fatalf("%q accepted: %v", p, e)
			}
			if _, e := s.ReadText(ctx, p); e == nil {
				t.Fatal("unsafe read accepted")
			}
		})
	}
	for _, p := range []string{".", "中文目录/配置.toml", "a-b_2.txt", ".config/file"} {
		if e := ValidatePath(p); e != nil {
			t.Fatalf("valid %q: %v", p, e)
		}
	}
	if _, e := New(root, Options{StateDir: filepath.Join(root, "unsafe-state")}); e == nil {
		t.Fatal("served state directory accepted")
	}
}

func TestTextVersionConflictAndBoundaries(t *testing.T) {
	s, root, _ := testService(t, Options{MaxTextBytes: 100, MaxFileBytes: 1024})
	ctx := context.Background()
	x, e := s.WriteText(ctx, "a.txt", "\ufeff你好\r\n", MissingVersion)
	if e != nil {
		t.Fatal(e)
	}
	got, e := s.ReadText(ctx, "a.txt")
	if e != nil || got.Content != "\ufeff你好\r\n" || got.Encoding != "UTF-8 BOM" || got.Newline != "CRLF" {
		t.Fatalf("%+v %v", got, e)
	}
	put(t, root, "a.txt", "external")
	if _, e = s.WriteText(ctx, "a.txt", "lost", x.Version); !errors.Is(e, ErrConflict) {
		t.Fatalf("expected conflict: %v", e)
	}
	content(t, root, "a.txt", "external")
	if _, e = s.WriteText(ctx, "a.txt", "lost", ""); !errors.Is(e, ErrConflict) {
		t.Fatal(e)
	}
	put(t, root, "binary", "a\x00b")
	if _, e = s.ReadText(ctx, "binary"); !errors.Is(e, ErrUnsupported) {
		t.Fatal(e)
	}
	put(t, root, "invalid", "\xff")
	if _, e = s.ReadText(ctx, "invalid"); !errors.Is(e, ErrUnsupported) {
		t.Fatal(e)
	}
	put(t, root, "large", strings.Repeat("x", 101))
	if _, e = s.ReadText(ctx, "large"); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	if _, e = s.WriteText(ctx, "new", strings.Repeat("x", 101), MissingVersion); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	missing(t, root, "new")
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	if _, e = s.WriteText(ctx, "cancelled", "x", MissingVersion); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	missing(t, root, "cancelled")
}

func TestConcurrentConditionalWritesAndDirectoryPages(t *testing.T) {
	s, root, _ := testService(t, Options{})
	ctx := context.Background()
	put(t, root, "shared", "first")
	v := version(t, s, "shared")
	var successes atomic.Int64
	var wg sync.WaitGroup
	errs := make(chan error, 12)
	for i := range 12 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, e := s.WriteText(ctx, "shared", strings.Repeat("z", i+1), v)
			if e == nil {
				successes.Add(1)
			} else if !errors.Is(e, ErrConflict) {
				errs <- e
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
	if successes.Load() != 1 {
		t.Fatalf("%d writers passed one baseline", successes.Load())
	}
	put(t, root, "a", "a")
	put(t, root, "b", "b")
	page, e := s.List(ctx, ".", ListOptions{Limit: 1})
	if e != nil || len(page.Items) != 1 || page.Total != 3 || page.NextOffset != 1 {
		t.Fatalf("%+v %v", page, e)
	}
	next, e := s.List(ctx, ".", ListOptions{Offset: 1, Limit: 2, Version: page.Version})
	if e != nil || len(next.Items) != 2 || next.NextOffset != -1 {
		t.Fatalf("%+v %v", next, e)
	}
	put(t, root, "c", "c")
	if _, e = s.List(ctx, ".", ListOptions{Offset: 1, Version: page.Version}); !errors.Is(e, ErrConflict) {
		t.Fatal(e)
	}
	for _, i := range next.Items {
		if strings.HasPrefix(i.Name, ".blora-") {
			t.Fatal("private staging was listed")
		}
	}
}

func TestRecursiveCopyMoveTrashRestoreAndSourceChange(t *testing.T) {
	s, root, _ := testService(t, Options{})
	ctx := context.Background()
	put(t, root, "source/sub/a", "one")
	put(t, root, "source/b", "two")
	v := version(t, s, "source")
	r, e := s.Copy(ctx, "source", "copy", v, MissingVersion)
	if e != nil || r.Stage != "committed" || r.Bytes != 6 || len(r.Cleanup) != 0 {
		t.Fatalf("%+v %v", r, e)
	}
	content(t, root, "copy/sub/a", "one")
	if _, e = s.Copy(ctx, "source", "source/child", v, MissingVersion); !errors.Is(e, ErrPath) {
		t.Fatal(e)
	}
	r, e = s.Move(ctx, "copy", "moved", r.Version, MissingVersion)
	if e != nil || !r.SourceDeleted {
		t.Fatalf("%+v %v", r, e)
	}
	missing(t, root, "copy")
	trash, e := s.Delete(ctx, "moved", r.Version)
	if e != nil || trash.Stage != "trashed" {
		t.Fatalf("%+v %v", trash, e)
	}
	missing(t, root, "moved")
	items, e := s.Trash(ctx)
	if e != nil || len(items) != 1 {
		t.Fatalf("%+v %v", items, e)
	}
	_, e = s.Restore(ctx, trash.ID, "restored", MissingVersion)
	if e != nil {
		t.Fatal(e)
	}
	content(t, root, "restored/sub/a", "one")
	put(t, root, "source/sub/a", "changed")
	if _, e = s.Delete(ctx, "source", v); !errors.Is(e, ErrConflict) {
		t.Fatal(e)
	}
	content(t, root, "source/sub/a", "changed")
}

func TestQuotaAndOwnedTrashRetention(t *testing.T) {
	s, root, _ := testService(t, Options{TrashBytes: 8})
	ctx := context.Background()
	put(t, root, "one", "1234")
	put(t, root, "two", "12345")
	t1, e := s.Delete(ctx, "one", version(t, s, "one"))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Delete(ctx, "two", version(t, s, "two")); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	content(t, root, "two", "12345")
	put(t, root, privateDir+"/trash/unowned", "keep")
	r, e := s.PurgeExpired(ctx, t1.ExpiresAt.Add(time.Second))
	if e != nil || len(r.Completed) != 1 || r.Completed[0] != t1.ID {
		t.Fatalf("%+v %v", r, e)
	}
	content(t, root, privateDir+"/trash/unowned", "keep")
	small, smallroot, _ := testService(t, Options{MaxEntries: 2, MaxTotalBytes: 10, MaxFileBytes: 10})
	put(t, smallroot, "a", "one")
	put(t, smallroot, "b", "two")
	put(t, smallroot, "c", "three")
	if _, e = small.List(ctx, ".", ListOptions{}); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
}

func makeZip(t *testing.T, root, name string, headers []zip.FileHeader, bodies []string) {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for i, h := range headers {
		out, e := w.CreateHeader(&h)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = out.Write([]byte(bodies[i])); e != nil {
			t.Fatal(e)
		}
	}
	if e := w.Close(); e != nil {
		t.Fatal(e)
	}
	put(t, root, name, buf.String())
}
func TestZipRoundTripAndOverwriteConflict(t *testing.T) {
	s, root, _ := testService(t, Options{})
	ctx := context.Background()
	put(t, root, "folder/sub/a", "configuration")
	put(t, root, "folder/b", "other")
	r, e := s.Compress(ctx, "folder", "archive.zip", version(t, s, "folder"), MissingVersion)
	if e != nil {
		t.Fatal(e)
	}
	r, e = s.Extract(ctx, "archive.zip", "expanded", r.Version, ExtractOptions{TargetVersion: MissingVersion})
	if e != nil || r.Stage != "committed" {
		t.Fatalf("%+v %v", r, e)
	}
	content(t, root, "expanded/folder/sub/a", "configuration")
	archiveVersion := version(t, s, "archive.zip")
	targetVersion := version(t, s, "expanded")
	if _, e = s.Extract(ctx, "archive.zip", "expanded", archiveVersion, ExtractOptions{TargetVersion: targetVersion}); !errors.Is(e, ErrConflict) {
		t.Fatal(e)
	}
	r, e = s.Extract(ctx, "archive.zip", "expanded", archiveVersion, ExtractOptions{TargetVersion: targetVersion, Overwrite: map[string]string{"folder/sub/a": version(t, s, "expanded/folder/sub/a"), "folder/b": version(t, s, "expanded/folder/b")}})
	if e != nil {
		t.Fatalf("%+v %v", r, e)
	}
}
func TestArchiveRejectsTraversalLinksDuplicatesAndBombs(t *testing.T) {
	for _, name := range []string{"../escape", "/absolute", "C:/bad", `a\b`, "file:ads", ".blora-files/x", "NUL"} {
		t.Run(name, func(t *testing.T) {
			s, root, _ := testService(t, Options{})
			makeZip(t, root, "bad.zip", []zip.FileHeader{{Name: name, Method: zip.Store}}, []string{"evil"})
			if _, e := s.Extract(context.Background(), "bad.zip", "out", version(t, s, "bad.zip"), ExtractOptions{TargetVersion: MissingVersion}); e == nil {
				t.Fatal("malicious archive accepted")
			}
			missing(t, root, "out")
		})
	}
	s, root, _ := testService(t, Options{})
	h := zip.FileHeader{Name: "link", Method: zip.Store}
	h.SetMode(os.ModeSymlink | 0777)
	makeZip(t, root, "link.zip", []zip.FileHeader{h}, []string{"/outside"})
	if _, e := s.Extract(context.Background(), "link.zip", "out", version(t, s, "link.zip"), ExtractOptions{TargetVersion: MissingVersion}); !errors.Is(e, ErrUnsupported) {
		t.Fatal(e)
	}
	makeZip(t, root, "duplicate.zip", []zip.FileHeader{{Name: "A"}, {Name: "a"}}, []string{"x", "y"})
	if _, e := s.Extract(context.Background(), "duplicate.zip", "out", version(t, s, "duplicate.zip"), ExtractOptions{TargetVersion: MissingVersion}); !errors.Is(e, ErrConflict) {
		t.Fatal(e)
	}
	makeZip(t, root, "bomb.zip", []zip.FileHeader{{Name: "large", Method: zip.Deflate}}, []string{strings.Repeat("0", 2<<20)})
	if _, e := s.Extract(context.Background(), "bomb.zip", "out", version(t, s, "bomb.zip"), ExtractOptions{TargetVersion: MissingVersion}); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	missing(t, root, "out")
	makeZip(t, root, "crc.zip", []zip.FileHeader{{Name: "a", Method: zip.Store}}, []string{"unique-payload"})
	b, e := os.ReadFile(filepath.Join(root, "crc.zip"))
	if e != nil {
		t.Fatal(e)
	}
	at := bytes.Index(b, []byte("unique-payload"))
	if at < 0 {
		t.Fatal("payload missing")
	}
	b[at] = 'X'
	put(t, root, "crc.zip", string(b))
	if _, e = s.Extract(context.Background(), "crc.zip", "out", version(t, s, "crc.zip"), ExtractOptions{TargetVersion: MissingVersion}); !errors.Is(e, zip.ErrChecksum) {
		t.Fatalf("bad CRC = %v", e)
	}
	missing(t, root, "out")
}

type cancelWhenFile struct {
	context.Context
	cancel context.CancelFunc
	file   string
}

func (c cancelWhenFile) Err() error {
	if _, e := os.Stat(c.file); e == nil {
		c.cancel()
	}
	return c.Context.Err()
}
func TestExtractionCancellationReportsCommittedSubset(t *testing.T) {
	s, root, _ := testService(t, Options{})
	makeZip(t, root, "archive.zip", []zip.FileHeader{{Name: "a"}, {Name: "b"}}, []string{"one", "two"})
	if _, e := s.Mkdir(context.Background(), "out"); e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	observed := cancelWhenFile{ctx, cancel, filepath.Join(root, "out/a")}
	r, e := s.Extract(observed, "archive.zip", "out", version(t, s, "archive.zip"), ExtractOptions{TargetVersion: version(t, s, "out")})
	if !errors.Is(e, context.Canceled) || r.Stage != "committing" || len(r.Completed) != 1 || r.Completed[0] != "out/a" {
		t.Fatalf("%+v %v", r, e)
	}
	content(t, root, "out/a", "one")
	missing(t, root, "out/b")
	if len(r.Cleanup) != 0 {
		t.Fatalf("unremoved staging %+v", r.Cleanup)
	}
}

func uploadSpec(data []byte, p string) UploadSpec {
	return UploadSpec{ID: randomID(), OwnerID: "user-a", Path: p, Total: int64(len(data)), Hash: hashBytes(data), ExpectedVersion: MissingVersion, SourceName: p, SourceModified: 123, SourceFingerprint: hashBytes(data)}
}
func sendAll(t *testing.T, s *Service, spec UploadSpec, data []byte) Upload {
	t.Helper()
	u, e := s.BeginUpload(context.Background(), spec)
	if e != nil {
		t.Fatal(e)
	}
	for off := u.Offset; off < int64(len(data)); {
		b := data[off:min(off+int64(u.ChunkBytes), int64(len(data)))]
		u, e = s.UploadChunk(context.Background(), spec.ID, off, b, hashBytes(b))
		if e != nil {
			t.Fatal(e)
		}
		off = u.Offset
	}
	return u
}

func TestDurableUploadResumeSourceIdentityAndCommit(t *testing.T) {
	s, root, state := testService(t, Options{})
	ctx := context.Background()
	data := bytes.Repeat([]byte("abc123"), 30000)
	spec := uploadSpec(data, "uploaded")
	u, e := s.BeginUpload(ctx, spec)
	if e != nil {
		t.Fatal(e)
	}
	first := data[:u.ChunkBytes]
	u, e = s.UploadChunk(ctx, spec.ID, 0, first, hashBytes(first))
	if e != nil || u.Offset != int64(len(first)) {
		t.Fatalf("%+v %v", u, e)
	}
	// Simulate bytes written and synced before the next checkpoint, then close
	// and reopen every service handle to exercise disk-backed recovery.
	f, e := os.OpenFile(filepath.Join(root, uploadPart(u)), os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = f.Write([]byte("unacknowledged-tail")); e != nil {
		t.Fatal(e)
	}
	f.Sync()
	f.Close()
	s.Close()
	s, e = New(root, Options{StateDir: state})
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	u, e = s.UploadStatus(ctx, spec.ID)
	if e != nil || u.Offset != int64(len(first)) {
		t.Fatalf("%+v %v", u, e)
	}
	changed := spec
	changed.SourceFingerprint = "another source"
	if _, e = s.BeginUpload(ctx, changed); !errors.Is(e, ErrTransfer) {
		t.Fatal(e)
	}
	if _, e = s.UploadChunk(ctx, spec.ID, 0, first, hashBytes(first)); e != nil {
		t.Fatal(e)
	}
	wrong := bytes.Repeat([]byte("x"), len(first))
	if _, e = s.UploadChunk(ctx, spec.ID, 0, wrong, hashBytes(wrong)); !errors.Is(e, ErrTransfer) {
		t.Fatal(e)
	}
	for off := u.Offset; off < int64(len(data)); {
		b := data[off:min(off+int64(u.ChunkBytes), int64(len(data)))]
		u, e = s.UploadChunk(ctx, spec.ID, off, b, hashBytes(b))
		if e != nil {
			t.Fatal(e)
		}
		off = u.Offset
	}
	u, e = s.CommitUpload(ctx, spec.ID)
	if e != nil || u.Stage != "committed" || u.Version != spec.Hash {
		t.Fatalf("%+v %v", u, e)
	}
	content(t, root, "uploaded", string(data))
	missing(t, root, uploadPart(u))
	if _, e = s.CommitUpload(ctx, spec.ID); e != nil {
		t.Fatal(e)
	}
}

func TestUploadConflictHashCancelAndReconciliation(t *testing.T) {
	s, root, _ := testService(t, Options{})
	ctx := context.Background()
	data := []byte("payload")
	spec := uploadSpec(data, "dest")
	u := sendAll(t, s, spec, data)
	put(t, root, "dest", "external")
	if _, e := s.CommitUpload(ctx, spec.ID); !errors.Is(e, ErrConflict) {
		t.Fatal(e)
	}
	content(t, root, "dest", "external")
	other := uploadSpec(data, "other")
	sendAll(t, s, other, data)
	u, e := s.CancelUpload(ctx, spec.ID)
	if e != nil || u.Stage != "cancelled" {
		t.Fatalf("%+v %v", u, e)
	}
	missing(t, root, uploadPart(u))
	content(t, root, "dest", "external")
	if _, e = s.UploadStatus(ctx, other.ID); e != nil {
		t.Fatal(e)
	}
	bad := uploadSpec(data, "bad")
	bad.Hash = hashBytes([]byte("different"))
	sendAll(t, s, bad, data)
	if _, e = s.CommitUpload(ctx, bad.ID); !errors.Is(e, ErrTransfer) {
		t.Fatal(e)
	}
	missing(t, root, "bad")
	// Simulate crash between target rename and the final committed checkpoint.
	u, e = s.UploadStatus(ctx, other.ID)
	if e != nil {
		t.Fatal(e)
	}
	u.Stage = "committing"
	part, e := s.root.Open(uploadPart(u))
	if e != nil {
		t.Fatal(e)
	}
	u.CommitObjectID, e = relationObjectID(part)
	part.Close()
	if e != nil {
		t.Fatal(e)
	}
	if e = s.writeState(uploadRecord(other.ID), u); e != nil {
		t.Fatal(e)
	}
	if e = s.root.Rename(uploadPart(u), other.Path); e != nil {
		t.Fatal(e)
	}
	u, e = s.UploadStatus(ctx, other.ID)
	if e != nil || u.Stage != "committed" {
		t.Fatalf("%+v %v", u, e)
	}
	// Simulate cancellation after deleting the part but before its final record.
	u, e = s.UploadStatus(ctx, bad.ID)
	if e != nil {
		t.Fatal(e)
	}
	u.Stage = "cancelling"
	if e = s.writeState(uploadRecord(bad.ID), u); e != nil {
		t.Fatal(e)
	}
	if e = s.root.Remove(uploadPart(u)); e != nil {
		t.Fatal(e)
	}
	u, e = s.UploadStatus(ctx, bad.ID)
	if e != nil || u.Stage != "cancelled" {
		t.Fatalf("%+v %v", u, e)
	}
	empty := uploadSpec(nil, "empty")
	sendAll(t, s, empty, nil)
	if _, e = s.CommitUpload(ctx, empty.ID); e != nil {
		t.Fatal(e)
	}
	content(t, root, "empty", "")
}

func TestUploadQuotaAndReadChunksBindSourceVersion(t *testing.T) {
	s, root, _ := testService(t, Options{MaxTotalBytes: 10, ChunkBytes: 4})
	ctx := context.Background()
	spec := uploadSpec([]byte("123456"), "one")
	if _, e := s.BeginUpload(ctx, spec); e != nil {
		t.Fatal(e)
	}
	if _, e := s.BeginUpload(ctx, uploadSpec([]byte("123456"), "two")); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	put(t, root, "source", "original")
	v := version(t, s, "source")
	c, e := s.ReadChunk(ctx, "source", v, 0, 4)
	if e != nil || string(c.Data) != "orig" || c.Total != 8 || c.Hash != hashBytes(c.Data) || c.Version != v {
		t.Fatalf("%+v %v", c, e)
	}
	put(t, root, "source", "modified")
	if _, e = s.ReadChunk(ctx, "source", v, 4, 4); !errors.Is(e, ErrConflict) {
		t.Fatal(e)
	}
	if _, e = s.Delete(ctx, "source", v); !errors.Is(e, ErrConflict) {
		t.Fatal(e)
	}
	content(t, root, "source", "modified")
}

func TestPendingUploadPinsParentAndImplicitArchiveDirectoryQuota(t *testing.T) {
	s, root, _ := testService(t, Options{})
	ctx := context.Background()
	put(t, root, "dir/existing", "x")
	spec := uploadSpec([]byte("pending"), "dir/file")
	if _, e := s.BeginUpload(ctx, spec); e != nil {
		t.Fatal(e)
	}
	if _, e := s.Move(ctx, "dir", "moved", version(t, s, "dir"), MissingVersion); !errors.Is(e, ErrConflict) {
		t.Fatal(e)
	}
	if _, e := s.Delete(ctx, "dir", version(t, s, "dir")); !errors.Is(e, ErrConflict) {
		t.Fatal(e)
	}
	if _, e := s.CancelUpload(ctx, spec.ID); e != nil {
		t.Fatal(e)
	}
	if _, e := s.Move(ctx, "dir", "moved", version(t, s, "dir"), MissingVersion); e != nil {
		t.Fatal(e)
	}
	small, smallroot, _ := testService(t, Options{MaxEntries: 2})
	makeZip(t, smallroot, "bad.zip", []zip.FileHeader{{Name: "a/b/c"}}, []string{"x"})
	if _, e := small.Extract(ctx, "bad.zip", "out", version(t, small, "bad.zip"), ExtractOptions{TargetVersion: MissingVersion}); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	missing(t, smallroot, "out")
	h := zip.FileHeader{Name: "link/"}
	h.SetMode(os.ModeSymlink | 0777)
	makeZip(t, root, "linkdir.zip", []zip.FileHeader{h}, []string{""})
	if _, e := s.Extract(ctx, "linkdir.zip", "out", version(t, s, "linkdir.zip"), ExtractOptions{TargetVersion: MissingVersion}); !errors.Is(e, ErrUnsupported) {
		t.Fatal(e)
	}
}

func TestDirectorySearchAndSortPrecedePagination(t *testing.T) {
	s, root, _ := testService(t, Options{})
	put(t, root, "match-a", "1")
	put(t, root, "match-b", "123")
	put(t, root, "unrelated", "12345")
	p, e := s.List(context.Background(), ".", ListOptions{Limit: 1, Search: "match", Sort: "size", Order: "desc"})
	if e != nil || p.Total != 2 || len(p.Items) != 1 || p.Items[0].Name != "match-b" || p.NextOffset != 1 {
		t.Fatalf("%+v %v", p, e)
	}
	second, e := s.List(context.Background(), ".", ListOptions{Offset: 1, Limit: 1, Version: p.Version, Search: "match", Sort: "size", Order: "desc"})
	if e != nil || len(second.Items) != 1 || second.Items[0].Name != "match-a" {
		t.Fatalf("%+v %v", second, e)
	}
}

func TestCheckpointRetentionKeepsActiveTransfersAndUserFiles(t *testing.T) {
	s, root, _ := testService(t, Options{})
	ctx := context.Background()
	data := []byte("contents")
	finished := uploadSpec(data, "done")
	sendAll(t, s, finished, data)
	if _, e := s.CommitUpload(ctx, finished.ID); e != nil {
		t.Fatal(e)
	}
	active := uploadSpec(data, "pending")
	if _, e := s.BeginUpload(ctx, active); e != nil {
		t.Fatal(e)
	}
	r, e := s.PruneRecords(ctx, time.Now().Add(time.Second))
	if e != nil || len(r.Completed) != 1 {
		t.Fatalf("%+v %v", r, e)
	}
	assert := func(p string) { content(t, root, p, "contents") }
	assert("done")
	if _, e = s.UploadStatus(ctx, active.ID); e != nil {
		t.Fatal(e)
	}
	if _, e = s.UploadStatus(ctx, finished.ID); !errors.Is(e, fs.ErrNotExist) {
		t.Fatal(e)
	}
}

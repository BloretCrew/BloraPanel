// Package filesystem provides bounded file operations under a daemon-selected
// directory handle. The caller must authorize each operation and keep one
// Service per resource root; client supplied host paths must never reach New.
package filesystem

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const MissingVersion = "missing"
const privateDir = ".blora-files"

var (
	ErrPath        = errors.New("unsafe relative path")
	ErrConflict    = errors.New("file version conflict")
	ErrLimit       = errors.New("file operation limit exceeded")
	ErrUnsupported = errors.New("unsupported file type or encoding")
	ErrTransfer    = errors.New("transfer identity or checkpoint mismatch")
)

type Options struct {
	// StateDir is a private directory owned by the daemon, outside the served
	// root and its workload mounts. It stores no whole-file upload copies.
	StateDir       string
	MaxFileBytes   int64
	MaxTextBytes   int64
	MaxTotalBytes  int64
	MaxEntries     int
	MaxConcurrent  int
	MaxTransfers   int
	ChunkBytes     int
	TrashBytes     int64
	TrashRetention time.Duration
}

type Entry struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Kind     string    `json:"kind"`
	Size     int64     `json:"size"`
	Mode     uint32    `json:"mode"`
	Modified time.Time `json:"modified"`
	Version  string    `json:"version,omitempty"`
}
type ListOptions struct {
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
	Version string `json:"version,omitempty"`
	Search  string `json:"search,omitempty"`
	Sort    string `json:"sort,omitempty"`
	Order   string `json:"order,omitempty"`
}
type Page struct {
	Items      []Entry `json:"items"`
	Version    string  `json:"version"`
	Total      int     `json:"total"`
	NextOffset int     `json:"nextOffset"`
}
type TextFile struct {
	Entry    Entry  `json:"entry"`
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
	Newline  string `json:"newline"`
	MaxBytes int64  `json:"maxBytes"`
}
type Result struct {
	Source        string   `json:"source,omitempty"`
	Target        string   `json:"target,omitempty"`
	Version       string   `json:"version,omitempty"`
	Completed     []string `json:"completed"`
	Bytes         int64    `json:"bytes"`
	SourceDeleted bool     `json:"sourceDeleted"`
	Stage         string   `json:"stage"`
	// A nonempty cleanup list identifies this operation's remaining objects.
	Cleanup []string `json:"cleanup,omitempty"`
}

type claim struct {
	paths []string
	write bool
}
type Service struct {
	root       *os.Root
	state      *os.Root
	opts       Options
	sem        chan struct{}
	mu         sync.Mutex
	claims     map[*claim]struct{}
	changed    chan struct{}
	proofMu    sync.Mutex
	proofs     map[string]*readProof
	proofBytes int
	proofClock uint64
}

func New(root string, options Options) (*Service, error) {
	if options.StateDir == "" {
		return nil, errors.New("filesystem: private StateDir is required")
	}
	if options.MaxFileBytes <= 0 {
		options.MaxFileBytes = 1 << 30
	}
	if options.MaxTextBytes <= 0 {
		options.MaxTextBytes = min(4<<20, options.MaxFileBytes)
	}
	if options.MaxTotalBytes <= 0 {
		options.MaxTotalBytes = 4 << 30
	}
	if options.MaxEntries <= 0 {
		options.MaxEntries = 100000
	}
	if options.MaxConcurrent <= 0 {
		options.MaxConcurrent = 4
	}
	if options.MaxTransfers <= 0 {
		options.MaxTransfers = 16
	}
	if options.ChunkBytes <= 0 {
		options.ChunkBytes = 64 << 10
	}
	if options.ChunkBytes > 256<<10 || options.MaxTextBytes > options.MaxFileBytes {
		return nil, ErrLimit
	}
	if options.TrashBytes <= 0 {
		options.TrashBytes = options.MaxTotalBytes
	}
	if options.TrashRetention <= 0 {
		options.TrashRetention = 30 * 24 * time.Hour
	}
	// This is an administrative configuration check, not the traversal guard.
	absRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, err
	}
	absRoot, err = filepath.Abs(absRoot)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(options.StateDir, 0700); err != nil {
		return nil, err
	}
	absState, err := filepath.EvalSymlinks(options.StateDir)
	if err != nil {
		return nil, err
	}
	absState, err = filepath.Abs(absState)
	if err != nil {
		return nil, err
	}
	rel, err := filepath.Rel(absRoot, absState)
	if err != nil || rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
		return nil, errors.New("filesystem: StateDir must be outside the served root")
	}
	r, err := os.OpenRoot(absRoot)
	if err != nil {
		return nil, err
	}
	st, err := os.OpenRoot(absState)
	if err != nil {
		r.Close()
		return nil, err
	}
	s := &Service{root: r, state: st, opts: options, sem: make(chan struct{}, options.MaxConcurrent), claims: make(map[*claim]struct{}), changed: make(chan struct{})}
	for _, p := range []string{privateDir, privateDir + "/work", privateDir + "/uploads", privateDir + "/trash"} {
		if info, e := r.Lstat(p); e == nil && (!info.IsDir() || info.Mode()&os.ModeSymlink != 0) {
			s.Close()
			return nil, ErrPath
		}
		if e := r.MkdirAll(p, 0700); e != nil {
			s.Close()
			return nil, e
		}
	}
	for _, p := range []string{"uploads", "trash"} {
		if e := st.MkdirAll(p, 0700); e != nil {
			s.Close()
			return nil, e
		}
	}
	return s, nil
}

// Close must be called after active operations have been joined by the daemon.
func (s *Service) Close() error    { return errors.Join(s.root.Close(), s.state.Close()) }
func (s *Service) Limits() Options { o := s.opts; o.StateDir = ""; return o }

// ValidatePath rejects alternative Windows path syntax on every platform so
// saved requests cannot gain a different interpretation after node migration.
func ValidatePath(p string) error {
	if p == "." {
		return nil
	}
	if p == "" || len(p) > 4096 || !utf8.ValidString(p) || strings.ContainsAny(p, "\\:\x00") || strings.HasPrefix(p, "/") {
		return ErrPath
	}
	for _, part := range strings.Split(p, "/") {
		if part == "" || part == "." || part == ".." || len(part) > 255 || strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") || strings.HasPrefix(strings.ToLower(part), ".blora-") {
			return ErrPath
		}
		for _, r := range part {
			if r < 32 || strings.ContainsRune("<>\"|?*", r) {
				return ErrPath
			}
		}
		base := strings.ToUpper(strings.SplitN(part, ".", 2)[0])
		if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" || base == "CONIN$" || base == "CONOUT$" || ((strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && len([]rune(base)) == 4 && strings.ContainsRune("123456789¹²³", []rune(base)[3])) {
			return ErrPath
		}
	}
	return nil
}

func (s *Service) validate(p string, allowRoot bool) error {
	if err := ValidatePath(p); err != nil {
		return err
	}
	if p == "." && !allowRoot {
		return ErrPath
	}
	parts := strings.Split(p, "/")
	for i := range parts {
		info, err := s.root.Lstat(strings.Join(parts[:i+1], "/"))
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 || (!info.Mode().IsRegular() && !info.IsDir()) {
			return ErrUnsupported
		}
	}
	return nil
}
func related(a, b string) bool {
	return a == "." || b == "." || a == b || strings.HasPrefix(a, b+"/") || strings.HasPrefix(b, a+"/")
}
func (s *Service) acquire(ctx context.Context, write bool, paths ...string) (func(), error) {
	select {
	case s.sem <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	c := &claim{paths: paths, write: write}
	for {
		if err := ctx.Err(); err != nil {
			<-s.sem
			return nil, err
		}
		s.mu.Lock()
		conflict := false
		for existing := range s.claims {
			if !write && !existing.write {
				continue
			}
			for _, a := range paths {
				for _, b := range existing.paths {
					if related(a, b) {
						conflict = true
					}
				}
			}
		}
		if !conflict {
			s.claims[c] = struct{}{}
			s.mu.Unlock()
			return func() {
				s.mu.Lock()
				delete(s.claims, c)
				close(s.changed)
				s.changed = make(chan struct{})
				s.mu.Unlock()
				<-s.sem
			}, nil
		}
		ch := s.changed
		s.mu.Unlock()
		select {
		case <-ch:
		case <-ctx.Done():
			<-s.sem
			return nil, ctx.Err()
		}
	}
}
func randomID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}
func hashBytes(b []byte) string { h := sha256.Sum256(b); return "sha256:" + hex.EncodeToString(h[:]) }
func validHash(s string) bool {
	if !strings.HasPrefix(s, "sha256:") || len(s) != 71 {
		return false
	}
	_, e := hex.DecodeString(s[7:])
	return e == nil
}
func validID(id string) bool {
	if len(id) != 32 {
		return false
	}
	_, err := hex.DecodeString(id)
	return err == nil && strings.ToLower(id) == id
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r contextReader) Read(b []byte) (int, error) {
	if e := r.ctx.Err(); e != nil {
		return 0, e
	}
	return r.r.Read(b)
}
func copyBounded(ctx context.Context, w io.Writer, r io.Reader, limit int64) (int64, error) {
	n, e := io.CopyBuffer(w, io.LimitReader(contextReader{ctx, r}, limit+1), make([]byte, 64<<10))
	if e != nil {
		return n, e
	}
	if n > limit {
		return n, ErrLimit
	}
	return n, ctx.Err()
}
func (s *Service) openRegular(p string) (*os.File, fs.FileInfo, error) {
	fi, e := s.root.Lstat(p)
	if e != nil {
		return nil, nil, e
	}
	if !fi.Mode().IsRegular() {
		return nil, nil, ErrUnsupported
	}
	f, e := s.root.OpenFile(p, os.O_RDONLY|readNonblock, 0)
	if e != nil {
		return nil, nil, e
	}
	fi, e = f.Stat()
	if e != nil || !fi.Mode().IsRegular() {
		f.Close()
		if e != nil {
			return nil, nil, e
		}
		return nil, nil, ErrUnsupported
	}
	if fi.Size() > s.opts.MaxFileBytes {
		f.Close()
		return nil, nil, ErrLimit
	}
	return f, fi, nil
}
func (s *Service) fileVersion(ctx context.Context, p string) (string, fs.FileInfo, error) {
	f, info, e := s.openRegular(p)
	if e != nil {
		return "", nil, e
	}
	defer f.Close()
	h := sha256.New()
	n, e := copyBounded(ctx, h, f, s.opts.MaxFileBytes)
	if e != nil {
		return "", nil, e
	}
	after, e := f.Stat()
	if e != nil {
		return "", nil, e
	}
	current, e := s.root.Lstat(p)
	if e != nil {
		return "", nil, e
	}
	if n != info.Size() || after.Size() != info.Size() || !after.ModTime().Equal(info.ModTime()) || !os.SameFile(info, current) {
		return "", nil, ErrConflict
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), info, nil
}
func entry(p string, fi fs.FileInfo) Entry {
	kind := "file"
	if fi.IsDir() {
		kind = "directory"
	} else if !fi.Mode().IsRegular() {
		kind = "unsupported"
	}
	return Entry{Name: path.Base(p), Path: p, Kind: kind, Size: fi.Size(), Mode: uint32(fi.Mode().Perm()), Modified: fi.ModTime()}
}

func (s *Service) directory(ctx context.Context, p string) ([]Entry, string, error) {
	f, e := s.root.OpenFile(p, os.O_RDONLY|readNonblock, 0)
	if e != nil {
		return nil, "", e
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil {
		return nil, "", e
	}
	if !info.IsDir() {
		return nil, "", ErrUnsupported
	}
	items := make([]Entry, 0)
	for {
		if e = ctx.Err(); e != nil {
			return nil, "", e
		}
		list, err := f.ReadDir(256)
		for _, de := range list {
			if strings.HasPrefix(strings.ToLower(de.Name()), ".blora-") {
				continue
			}
			fi, err := de.Info()
			if err != nil {
				return nil, "", err
			}
			items = append(items, entry(path.Join(p, de.Name()), fi))
			if len(items) > s.opts.MaxEntries {
				return nil, "", ErrLimit
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", err
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	raw, _ := json.Marshal(items)
	return items, hashBytes(raw), nil
}

func (s *Service) List(ctx context.Context, p string, o ListOptions) (Page, error) {
	if e := s.validate(p, true); e != nil {
		return Page{}, e
	}
	done, e := s.acquire(ctx, false, p)
	if e != nil {
		return Page{}, e
	}
	defer done()
	if o.Offset < 0 || o.Limit < 0 || o.Limit > 1000 {
		return Page{}, ErrLimit
	}
	if o.Limit == 0 {
		o.Limit = 200
	}
	items, v, e := s.directory(ctx, p)
	if e != nil {
		return Page{}, e
	}
	if o.Version != "" && o.Version != v {
		return Page{}, ErrConflict
	}
	if len(o.Search) > 256 || (o.Order != "" && o.Order != "asc" && o.Order != "desc") {
		return Page{}, ErrLimit
	}
	if o.Search != "" {
		filtered := items[:0]
		for _, item := range items {
			if strings.Contains(strings.ToLower(item.Name), strings.ToLower(o.Search)) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	if o.Sort == "" {
		o.Sort = "name"
	}
	if o.Sort != "name" && o.Sort != "size" && o.Sort != "modified" && o.Sort != "kind" {
		return Page{}, ErrUnsupported
	}
	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i], items[j]
		if o.Order == "desc" {
			a, b = b, a
		}
		less := a.Name < b.Name
		switch o.Sort {
		case "size":
			if a.Size != b.Size {
				less = a.Size < b.Size
			}
		case "modified":
			if !a.Modified.Equal(b.Modified) {
				less = a.Modified.Before(b.Modified)
			}
		case "kind":
			if a.Kind != b.Kind {
				less = a.Kind < b.Kind
			}
		}
		return less
	})
	if o.Offset > len(items) {
		return Page{}, ErrConflict
	}
	end := min(o.Offset+o.Limit, len(items))
	next := -1
	if end < len(items) {
		next = end
	}
	return Page{items[o.Offset:end], v, len(items), next}, nil
}
func (s *Service) Stat(ctx context.Context, p string) (Entry, error) {
	if e := s.validate(p, true); e != nil {
		return Entry{}, e
	}
	done, e := s.acquire(ctx, false, p)
	if e != nil {
		return Entry{}, e
	}
	defer done()
	return s.stat(ctx, p)
}
func (s *Service) stat(ctx context.Context, p string) (Entry, error) {
	fi, e := s.root.Lstat(p)
	if e != nil {
		return Entry{}, e
	}
	out := entry(p, fi)
	if fi.IsDir() {
		_, v, e := s.tree(ctx, p)
		out.Version = v
		return out, e
	}
	v, _, e := s.fileVersion(ctx, p)
	out.Version = v
	return out, e
}
func (s *Service) expected(ctx context.Context, p, v string) error {
	if v == "" {
		return ErrConflict
	}
	got, e := s.stat(ctx, p)
	if errors.Is(e, fs.ErrNotExist) && v == MissingVersion {
		return nil
	}
	if e != nil {
		return e
	}
	if v == MissingVersion || got.Version != v {
		return ErrConflict
	}
	return nil
}

func (s *Service) ReadText(ctx context.Context, p string) (TextFile, error) {
	if e := s.validate(p, false); e != nil {
		return TextFile{}, e
	}
	done, e := s.acquire(ctx, false, p)
	if e != nil {
		return TextFile{}, e
	}
	defer done()
	f, fi, e := s.openRegular(p)
	if e != nil {
		return TextFile{}, e
	}
	defer f.Close()
	if fi.Size() > s.opts.MaxTextBytes {
		return TextFile{}, ErrLimit
	}
	b, e := io.ReadAll(io.LimitReader(contextReader{ctx, f}, s.opts.MaxTextBytes+1))
	if e != nil {
		return TextFile{}, e
	}
	if int64(len(b)) > s.opts.MaxTextBytes {
		return TextFile{}, ErrLimit
	}
	if !utf8.Valid(b) || strings.ContainsRune(string(b), '\x00') {
		return TextFile{}, ErrUnsupported
	}
	v, _, e := s.fileVersion(ctx, p)
	if e != nil {
		return TextFile{}, e
	}
	if v != hashBytes(b) {
		return TextFile{}, ErrConflict
	}
	out := entry(p, fi)
	out.Version = v
	nl := "LF"
	if strings.Contains(string(b), "\r\n") {
		nl = "CRLF"
	}
	encoding := "UTF-8"
	if strings.HasPrefix(string(b), "\ufeff") {
		encoding = "UTF-8 BOM"
	}
	return TextFile{out, string(b), encoding, nl, s.opts.MaxTextBytes}, nil
}

func (s *Service) WriteText(ctx context.Context, p, content, expected string) (Entry, error) {
	if int64(len(content)) > s.opts.MaxTextBytes {
		return Entry{}, ErrLimit
	}
	if !utf8.ValidString(content) || strings.ContainsRune(content, '\x00') {
		return Entry{}, ErrUnsupported
	}
	if e := s.validate(p, false); e != nil {
		return Entry{}, e
	}
	done, e := s.acquire(ctx, true, p)
	if e != nil {
		return Entry{}, e
	}
	defer done()
	if e = s.expected(ctx, p, expected); e != nil {
		return Entry{}, e
	}
	tmp := path.Join(path.Dir(p), ".blora-write-"+randomID())
	f, e := s.root.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return Entry{}, e
	}
	defer s.root.Remove(tmp)
	_, e = copyBounded(ctx, f, strings.NewReader(content), s.opts.MaxTextBytes)
	if e == nil {
		e = f.Sync()
	}
	if ce := f.Close(); e == nil {
		e = ce
	}
	if e != nil {
		return Entry{}, e
	}
	if fi, err := s.root.Lstat(p); err == nil {
		f, e = s.root.OpenFile(tmp, os.O_WRONLY, 0)
		if e != nil {
			return Entry{}, e
		}
		e = f.Chmod(fi.Mode().Perm())
		ce := f.Close()
		if e == nil {
			e = ce
		}
		if e != nil {
			return Entry{}, e
		}
	}
	if e = s.expected(ctx, p, expected); e != nil {
		return Entry{}, e
	}
	if e = ctx.Err(); e != nil {
		return Entry{}, e
	}
	if e = s.root.Rename(tmp, p); e != nil {
		return Entry{}, e
	}
	if e = syncDir(s.root, path.Dir(p)); e != nil {
		return Entry{}, fmt.Errorf("file replaced; directory durability uncertain: %w", e)
	}
	return s.stat(ctx, p)
}
func (s *Service) Mkdir(ctx context.Context, p string) (Entry, error) {
	if e := s.validate(p, false); e != nil {
		return Entry{}, e
	}
	done, e := s.acquire(ctx, true, p)
	if e != nil {
		return Entry{}, e
	}
	defer done()
	if e = ctx.Err(); e != nil {
		return Entry{}, e
	}
	if e = s.root.Mkdir(p, 0750); e != nil {
		return Entry{}, e
	}
	if e = syncDir(s.root, path.Dir(p)); e != nil {
		return Entry{}, e
	}
	return s.stat(ctx, p)
}

// writeState uses an atomic checkpoint and syncs data before acknowledging it.
func (s *Service) writeState(p string, value any) error {
	b, e := json.Marshal(value)
	if e != nil {
		return e
	}
	tmp := path.Join(path.Dir(p), ".checkpoint-"+randomID())
	f, e := s.state.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer s.state.Remove(tmp)
	_, e = f.Write(b)
	if e == nil {
		e = f.Sync()
	}
	if ce := f.Close(); e == nil {
		e = ce
	}
	if e != nil {
		return e
	}
	if e = s.state.Rename(tmp, p); e != nil {
		return e
	}
	return syncDir(s.state, path.Dir(p))
}
func (s *Service) readState(p string, value any) error {
	f, e := s.state.Open(p)
	if e != nil {
		return e
	}
	defer f.Close()
	d := json.NewDecoder(io.LimitReader(f, 1<<20))
	d.DisallowUnknownFields()
	return d.Decode(value)
}

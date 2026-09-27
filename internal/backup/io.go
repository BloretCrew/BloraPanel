package backup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"blora.dev/panel/internal/filesystem"
	"blora.dev/panel/internal/model"
)

const manifestLimit = 4 << 20
const recordLimit = 8 << 20
const chunkBytes = 64 << 10

type Manager struct {
	root, state *os.Root
	opts        Options
	slots       chan struct{}
	mu          sync.Mutex
	active      map[string]bool
	readers     map[string]int
	changed     chan struct{}
}

func New(repositoryRoot string, opts Options) (*Manager, error) {
	if opts.StateDir == "" {
		return nil, ErrInvalid
	}
	if opts.MaxEntries <= 0 {
		opts.MaxEntries = 4096
	}
	if opts.MaxFileBytes <= 0 {
		opts.MaxFileBytes = 1 << 30
	}
	if opts.MaxTotalBytes <= 0 {
		opts.MaxTotalBytes = 4 << 30
	}
	if opts.MaxArchiveBytes <= 0 {
		opts.MaxArchiveBytes = (4 << 30) - 1
	}
	if opts.MaxRepositoryBytes <= 0 {
		opts.MaxRepositoryBytes = 16 << 30
	}
	if opts.MaxSnapshots <= 0 {
		opts.MaxSnapshots = 1000
	}
	if opts.MaxConcurrent <= 0 {
		opts.MaxConcurrent = 2
	}
	if opts.MaxEntries > 59999 || opts.MaxFileBytes > 1<<30 || opts.MaxTotalBytes > 4<<30 || opts.MaxArchiveBytes >= 4<<30 || opts.MaxConcurrent > 8 {
		return nil, ErrLimit
	}
	if err := os.MkdirAll(repositoryRoot, 0700); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(opts.StateDir, 0700); err != nil {
		return nil, err
	}
	repo, err := filepath.EvalSymlinks(repositoryRoot)
	if err != nil {
		return nil, err
	}
	statePath, err := filepath.EvalSymlinks(opts.StateDir)
	if err != nil {
		return nil, err
	}
	repo, err = filepath.Abs(repo)
	if err != nil {
		return nil, err
	}
	statePath, err = filepath.Abs(statePath)
	if err != nil {
		return nil, err
	}
	for _, pair := range [][2]string{{repo, statePath}, {statePath, repo}} {
		rel, err := filepath.Rel(pair[0], pair[1])
		if err != nil || rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
			return nil, ErrInvalid
		}
	}
	r, err := os.OpenRoot(repo)
	if err != nil {
		return nil, err
	}
	s, err := os.OpenRoot(statePath)
	if err != nil {
		r.Close()
		return nil, err
	}
	m := &Manager{root: r, state: s, opts: opts, slots: make(chan struct{}, opts.MaxConcurrent), active: map[string]bool{}, readers: map[string]int{}, changed: make(chan struct{})}
	for _, name := range []string{"objects", "pending"} {
		if err = privateDirectory(r, name); err != nil {
			m.Close()
			return nil, err
		}
	}
	for _, name := range []string{"snapshots", "create", "plans", "restore"} {
		if err = privateDirectory(s, name); err != nil {
			m.Close()
			return nil, err
		}
	}
	return m, nil
}
func privateDirectory(root *os.Root, name string) error {
	if err := root.Mkdir(name, 0700); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	info, err := root.Lstat(name)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return ErrInvalid
	}
	return nil
}

// The caller stops and joins its durable workers before closing the manager.
func (m *Manager) Close() error { return errors.Join(m.root.Close(), m.state.Close()) }
func (m *Manager) acquire(ctx context.Context, key string) (func(), error) {
	select {
	case m.slots <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	for {
		m.mu.Lock()
		if !m.active[key] {
			m.active[key] = true
			m.mu.Unlock()
			return func() {
				m.mu.Lock()
				delete(m.active, key)
				close(m.changed)
				m.changed = make(chan struct{})
				m.mu.Unlock()
				<-m.slots
			}, nil
		}
		changed := m.changed
		m.mu.Unlock()
		select {
		case <-changed:
		case <-ctx.Done():
			<-m.slots
			return nil, ctx.Err()
		}
	}
}
func validID(id string) bool {
	if len(id) != 32 {
		return false
	}
	_, err := hex.DecodeString(id)
	return err == nil
}
func hashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func validHash(value string) bool {
	if len(value) != 71 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(value[7:])
	return err == nil
}
func validVersion(value string) bool         { return value == filesystem.MissingVersion || validHash(value) }
func validResource(r model.ResourceRef) bool { return r.Kind != "" && r.ID != "" && r.NodeID != "" }
func syncDirectory(root *os.Root, name string) error {
	if runtime.GOOS == "windows" {
		return nil
	} // same explicit process-crash guarantee as filesystem.syncDir
	f, err := root.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
func (m *Manager) writeRecord(name string, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if len(b) > recordLimit {
		return ErrLimit
	}
	tmp := name + "." + model.ID() + ".tmp"
	f, err := m.state.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		_ = m.state.Remove(tmp)
		return err
	}
	if err = m.state.Rename(tmp, name); err != nil {
		_ = m.state.Remove(tmp)
		return err
	}
	return syncDirectory(m.state, filepath.ToSlash(filepath.Dir(name)))
}
func (m *Manager) readRecord(name string, value any) error {
	f, err := m.state.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() > recordLimit {
		return ErrCorrupt
	}
	b, err := io.ReadAll(io.LimitReader(f, recordLimit+1))
	if err != nil {
		return err
	}
	if len(b) > recordLimit {
		return ErrLimit
	}
	return json.Unmarshal(b, value)
}
func (m *Manager) recordIDs(namespace string) ([]string, error) {
	f, err := m.state.Open(namespace)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var ids []string
	count := 0
	for {
		names, err := f.Readdirnames(128)
		for _, name := range names {
			count++
			if count > m.opts.MaxSnapshots*4 {
				return nil, ErrLimit
			}
			if strings.HasSuffix(name, ".json") && validID(strings.TrimSuffix(name, ".json")) {
				ids = append(ids, strings.TrimSuffix(name, ".json"))
			}
		}
		if err == io.EOF {
			return ids, nil
		}
		if err != nil {
			return nil, err
		}
	}
}
func archivePath(id string) string { return "objects/" + id + ".zip" }
func pendingPath(id string) string { return "pending/" + id + ".zip" }
func snapshotInfo(s Snapshot) SnapshotInfo {
	return SnapshotInfo{ID: s.Manifest.ID, OwnerID: s.Manifest.Spec.OwnerID, Source: s.Manifest.Spec.Source, Path: s.Manifest.Spec.Path, CapturedVersion: s.Manifest.CapturedVersion, Created: s.Manifest.Created, Entries: len(s.Manifest.Entries), Total: s.Manifest.Total, ArchiveHash: s.ArchiveHash, ArchiveBytes: s.ArchiveBytes, State: s.State, Consistency: s.Manifest.Spec.Consistency, Before: s.Before, After: s.After}
}

func digestFile(ctx context.Context, f *os.File, max int64) (string, int64, error) {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", 0, err
	}
	h := sha256.New()
	buffer := make([]byte, chunkBytes)
	var total int64
	for {
		if err := ctx.Err(); err != nil {
			return "", total, err
		}
		n, err := f.Read(buffer)
		if n > 0 {
			if total > max-int64(n) {
				return "", total, ErrLimit
			}
			h.Write(buffer[:n])
			total += int64(n)
		}
		if err == io.EOF {
			return "sha256:" + hex.EncodeToString(h.Sum(nil)), total, nil
		}
		if err != nil {
			return "", total, err
		}
	}
}

func (m *Manager) separate(ctx context.Context, service *filesystem.Service, p string) (filesystem.RelationFacts, error) {
	source, err := service.RelationFacts(ctx, p)
	if err != nil {
		return source, err
	}
	repository, err := filesystem.InspectRootRelation(ctx, m.root, ".")
	if err != nil {
		return source, err
	}
	state, err := filesystem.InspectRootRelation(ctx, m.state, ".")
	if err != nil {
		return source, err
	}
	if filesystem.OverlappingRelations(source, repository) || filesystem.OverlappingRelations(source, state) {
		return source, ErrConflict
	}
	return source, nil
}

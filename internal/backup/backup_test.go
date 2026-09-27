package backup

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"blora.dev/panel/internal/filesystem"
	"blora.dev/panel/internal/model"
)

type backupFixture struct {
	manager                                   *Manager
	source, target                            *filesystem.Service
	base, sourceRoot, targetRoot, repo, state string
	options                                   Options
}

func newBackupFixture(t *testing.T, opts Options) *backupFixture {
	t.Helper()
	base := t.TempDir()
	f := &backupFixture{base: base, sourceRoot: filepath.Join(base, "source"), targetRoot: filepath.Join(base, "target"), repo: filepath.Join(base, "repository"), state: filepath.Join(base, "backup-state")}
	for _, p := range []string{f.sourceRoot, f.targetRoot} {
		if err := os.Mkdir(p, 0700); err != nil {
			t.Fatal(err)
		}
	}
	var err error
	f.source, err = filesystem.New(f.sourceRoot, filesystem.Options{StateDir: filepath.Join(base, "source-state")})
	if err != nil {
		t.Fatal(err)
	}
	f.target, err = filesystem.New(f.targetRoot, filesystem.Options{StateDir: filepath.Join(base, "target-state")})
	if err != nil {
		t.Fatal(err)
	}
	opts.StateDir = f.state
	f.options = opts
	f.manager, err = New(f.repo, opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.manager.Close(); f.source.Close(); f.target.Close() })
	return f
}
func (f *backupFixture) reopen(t *testing.T) {
	t.Helper()
	if err := f.manager.Close(); err != nil {
		t.Fatal(err)
	}
	var err error
	f.manager, err = New(f.repo, f.options)
	if err != nil {
		t.Fatal(err)
	}
}
func backupRef(id string) model.ResourceRef {
	return model.ResourceRef{Kind: "instance", ID: id, NodeID: "node"}
}
func (f *backupFixture) spec(t *testing.T, p string) CreateSpec {
	t.Helper()
	entry, err := f.source.Stat(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	return CreateSpec{ID: model.ID(), OwnerID: "owner", Source: backupRef("source"), Path: p, Version: entry.Version, Compression: "deflate", Consistency: Consistency{Mode: "files"}}
}
func diskWrite(t *testing.T, p string, b []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, b, 0640); err != nil {
		t.Fatal(err)
	}
}
func (f *backupFixture) plan(t *testing.T, id, p string) RestorePlan {
	t.Helper()
	version, _, err := targetVersion(context.Background(), f.target, p)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := f.manager.PlanRestore(context.Background(), f.target, RestoreRequest{ID: model.ID(), OwnerID: "owner", BackupID: id, Target: backupRef("target"), Path: p, Version: version})
	if err != nil {
		t.Fatal(err)
	}
	return plan
}
func restoreSpec(plan RestorePlan) RestoreSpec {
	spec := RestoreSpec{ID: model.ID(), OwnerID: "owner", PlanID: plan.Request.ID, PlanHash: plan.Hash, Overwrite: map[string]string{}}
	for _, e := range plan.Entries {
		if e.Kind == "file" && e.TargetVersion != filesystem.MissingVersion {
			spec.Overwrite[e.TargetPath] = e.TargetVersion
		}
	}
	return spec
}

func TestBackupRoundTripMetadataAndExplicitOverwrite(t *testing.T) {
	f := newBackupFixture(t, Options{})
	ctx := context.Background()
	content := bytes.Repeat([]byte("真实备份-data\x00\n"), 17000)
	filePath := filepath.Join(f.sourceRoot, "world", "nested", "state.bin")
	diskWrite(t, filePath, content)
	if err := os.Chmod(filePath, 0640); err != nil {
		t.Fatal(err)
	}
	diskWrite(t, filepath.Join(f.sourceRoot, "world", "empty"), nil)
	stamp := time.Unix(1700000000, 123456700)
	if err := os.Chtimes(filePath, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(filepath.Join(f.sourceRoot, "world", "nested"), stamp, stamp); err != nil {
		t.Fatal(err)
	}
	spec := f.spec(t, "world")
	result, err := f.manager.Create(ctx, f.source, spec)
	if err != nil {
		t.Fatal(err)
	}
	if result.Stage != "succeeded" || result.Completed != 4 || result.Bytes != int64(len(content)) || result.Snapshot == nil {
		t.Fatalf("snapshot receipt: %+v", result)
	}
	if info, err := f.manager.Inspect(ctx, spec.ID); err != nil || info.ArchiveHash != result.Snapshot.ArchiveHash {
		t.Fatalf("verify: %+v %v", info, err)
	}
	f.reopen(t)
	again, err := f.manager.Create(ctx, f.source, spec)
	if err != nil || again.Snapshot.ArchiveHash != result.Snapshot.ArchiveHash {
		t.Fatalf("same intent: %+v %v", again, err)
	}
	changed := spec
	changed.Path = "world/nested"
	if _, err = f.manager.Create(ctx, f.source, changed); !errors.Is(err, ErrConflict) {
		t.Fatalf("spec rebound: %v", err)
	}
	plan := f.plan(t, spec.ID, "restored")
	restored, err := f.manager.Restore(ctx, f.target, restoreSpec(plan))
	if err != nil {
		t.Fatal(err)
	}
	if restored.Stage != "succeeded" || restored.Completed != 4 {
		t.Fatalf("restore receipt: %+v", restored)
	}
	got, err := os.ReadFile(filepath.Join(f.targetRoot, "restored", "nested", "state.bin"))
	if err != nil || !bytes.Equal(got, content) {
		t.Fatalf("restore content %d %v", len(got), err)
	}
	for _, p := range []string{"restored/nested", "restored/nested/state.bin"} {
		info, err := os.Stat(filepath.Join(f.targetRoot, filepath.FromSlash(p)))
		if err != nil {
			t.Fatal(err)
		}
		if !info.ModTime().Equal(stamp) {
			t.Fatalf("mtime %s: %v", p, info.ModTime())
		}
		if runtime.GOOS != "windows" && p == "restored/nested/state.bin" && info.Mode().Perm() != 0640 {
			t.Fatalf("mode %s: %v", p, info.Mode())
		}
	}
	diskWrite(t, filepath.Join(f.targetRoot, "merge", "nested", "state.bin"), []byte("old"))
	diskWrite(t, filepath.Join(f.targetRoot, "merge", "extra"), []byte("retained"))
	plan = f.plan(t, spec.ID, "merge")
	confirm := restoreSpec(plan)
	missing := confirm
	missing.Overwrite = nil
	if _, err = f.manager.Restore(ctx, f.target, missing); !errors.Is(err, ErrConflict) {
		t.Fatalf("overwrite without confirmation: %v", err)
	}
	confirm.Overwrite = nil
	confirm.OverwritePlanHash = plan.Hash
	if _, err = f.manager.Restore(ctx, f.target, confirm); err != nil {
		t.Fatal(err)
	}
	if got, err = os.ReadFile(filepath.Join(f.targetRoot, "merge", "extra")); err != nil || string(got) != "retained" {
		t.Fatalf("merge removed extra %q %v", got, err)
	}
}

func TestBackupSourceAndRestoreIdentityConflicts(t *testing.T) {
	f := newBackupFixture(t, Options{})
	ctx := context.Background()
	diskWrite(t, filepath.Join(f.sourceRoot, "file"), []byte("source"))
	spec := f.spec(t, "file")
	diskWrite(t, filepath.Join(f.sourceRoot, "file"), []byte("changed"))
	if _, err := f.manager.Create(ctx, f.source, spec); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale source: %v", err)
	}
	spec = f.spec(t, "file")
	if _, err := f.manager.Create(ctx, f.source, spec); err != nil {
		t.Fatal(err)
	}
	diskWrite(t, filepath.Join(f.targetRoot, "dest"), []byte("old"))
	plan := f.plan(t, spec.ID, "dest")
	old, err := os.Stat(filepath.Join(f.targetRoot, "dest"))
	if err != nil {
		t.Fatal(err)
	}
	diskWrite(t, filepath.Join(f.targetRoot, "replacement"), []byte("old"))
	if err = os.Chtimes(filepath.Join(f.targetRoot, "replacement"), old.ModTime(), old.ModTime()); err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(filepath.Join(f.targetRoot, "replacement"), filepath.Join(f.targetRoot, "dest")); err != nil {
		t.Fatal(err)
	}
	if _, err = f.manager.Restore(ctx, f.target, restoreSpec(plan)); !errors.Is(err, ErrConflict) {
		t.Fatalf("identical-content replaced object accepted: %v", err)
	}
	plan = f.plan(t, spec.ID, "new-target")
	archive := filepath.Join(f.repo, filepath.FromSlash(archivePath(spec.ID)))
	body, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	diskWrite(t, filepath.Join(f.repo, "replacement.zip"), body)
	if err = os.Rename(filepath.Join(f.repo, "replacement.zip"), archive); err != nil {
		t.Fatal(err)
	}
	if _, err = f.manager.Restore(ctx, f.target, restoreSpec(plan)); !errors.Is(err, ErrConflict) {
		t.Fatalf("same hash replaced archive accepted: %v", err)
	}
	if _, err = os.Stat(filepath.Join(f.targetRoot, "new-target")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("target touched before archive check: %v", err)
	}
}

func TestBackupRejectsSymlinksOverlapCancellationAndBudgets(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name string
		opts Options
	}{{"file", Options{MaxFileBytes: 4}}, {"total", Options{MaxTotalBytes: 4}}, {"entries", Options{MaxEntries: 1}}, {"archive", Options{MaxArchiveBytes: 32}}, {"repository", Options{MaxRepositoryBytes: 32}}} {
		t.Run(tc.name, func(t *testing.T) {
			f := newBackupFixture(t, tc.opts)
			diskWrite(t, filepath.Join(f.sourceRoot, "dir", "file"), []byte("over budget"))
			if _, err := f.manager.Create(ctx, f.source, f.spec(t, "dir")); !errors.Is(err, ErrLimit) {
				t.Fatalf("limit not applied: %v", err)
			}
			items, err := f.manager.List(ctx)
			if err != nil || len(items) != 0 {
				t.Fatalf("failed backup listed: %v %v", items, err)
			}
		})
	}
	f := newBackupFixture(t, Options{})
	diskWrite(t, filepath.Join(f.sourceRoot, "file"), []byte("data"))
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := f.manager.Create(cancelled, f.source, f.spec(t, "file")); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel %v", err)
	}
	if runtime.GOOS != "windows" {
		outside := filepath.Join(f.base, "outside")
		diskWrite(t, outside, []byte("private"))
		if err := os.Symlink(outside, filepath.Join(f.sourceRoot, "link")); err != nil {
			t.Fatal(err)
		}
		spec := f.spec(t, "file")
		spec.Path = "link"
		if _, err := f.manager.Create(ctx, f.source, spec); err == nil {
			t.Fatal("source symlink followed")
		}
	}
	wide, err := filesystem.New(f.base, filesystem.Options{StateDir: filepath.Join(t.TempDir(), "state")})
	if err != nil {
		t.Fatal(err)
	}
	defer wide.Close()
	if _, err = f.manager.separate(ctx, wide, "."); !errors.Is(err, ErrConflict) {
		t.Fatalf("repository/source overlap: %v", err)
	}
	for _, p := range []string{"../escape", "/tmp/escape", `C:\escape`, `name:stream`} {
		spec := f.spec(t, "file")
		spec.Path = p
		if _, err = f.manager.Create(ctx, f.source, spec); err == nil {
			t.Fatalf("unsafe source %s", p)
		}
	}
}

func (f *backupFixture) forgeArchive(t *testing.T, entries []Entry, names []string, modes []os.FileMode, payloads [][]byte) string {
	t.Helper()
	id := model.ID()
	spec := CreateSpec{ID: id, OwnerID: "owner", Source: backupRef("source"), Path: ".", Version: hashBytes(nil), Compression: "store", Consistency: Consistency{Mode: "files"}}
	manifest := Manifest{Format: "blora-backup-v1", ID: id, Spec: spec, CapturedVersion: spec.Version, Created: time.Now().UTC(), Entries: entries}
	for _, e := range entries {
		manifest.Total += e.Size
	}
	file, err := f.manager.root.OpenFile(archivePath(id), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(file)
	for i, name := range names {
		h := &zip.FileHeader{Name: name, Method: zip.Deflate}
		h.SetMode(modes[i])
		dst, err := w.CreateHeader(h)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = dst.Write(payloads[i]); err != nil {
			t.Fatal(err)
		}
	}
	dst, err := w.Create("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(manifest)
	if _, err = dst.Write(b); err != nil {
		t.Fatal(err)
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	if err = file.Sync(); err != nil {
		t.Fatal(err)
	}
	hash, size, err := digestFile(context.Background(), file, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	file.Close()
	relation, err := filesystem.InspectRootRelation(context.Background(), f.manager.root, archivePath(id))
	if err != nil {
		t.Fatal(err)
	}
	if err = f.manager.writeRecord("snapshots/"+id+".json", Snapshot{Manifest: manifest, ArchiveHash: hash, ArchiveBytes: size, ArchiveObjectID: relation.ObjectID, State: "ready"}); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestBackupRejectsForgedArchivePathsLinksBombsAndContent(t *testing.T) {
	stamp := time.Now().UTC()
	file := Entry{Path: ".", Kind: "file", Size: 1, Hash: hashBytes([]byte("x")), Mode: 0600, Modified: stamp}
	cases := []struct {
		name    string
		entries []Entry
		names   []string
		modes   []os.FileMode
		data    [][]byte
	}{
		{"traversal", []Entry{file}, []string{"data/../../escape"}, []os.FileMode{0600}, [][]byte{[]byte("x")}},
		{"windows", []Entry{file}, []string{`data/C:\escape`}, []os.FileMode{0600}, [][]byte{[]byte("x")}},
		{"symlink", []Entry{file}, []string{"data/content"}, []os.FileMode{os.ModeSymlink | 0777}, [][]byte{[]byte("x")}},
		{"bomb", []Entry{file}, []string{"data/content"}, []os.FileMode{0600}, [][]byte{bytes.Repeat([]byte("x"), 100000)}},
		{"hash", []Entry{file}, []string{"data/content"}, []os.FileMode{0600}, [][]byte{[]byte("y")}},
		{"duplicate", []Entry{file, file}, []string{"data/content", "data/content"}, []os.FileMode{0600, 0600}, [][]byte{[]byte("x"), []byte("x")}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newBackupFixture(t, Options{})
			id := f.forgeArchive(t, tc.entries, tc.names, tc.modes, tc.data)
			_, err := f.manager.PlanRestore(context.Background(), f.target, RestoreRequest{ID: model.ID(), OwnerID: "owner", BackupID: id, Target: backupRef("target"), Path: "destination", Version: filesystem.MissingVersion})
			if err == nil {
				t.Fatal("malicious archive accepted")
			}
			items, err := f.target.List(context.Background(), ".", filesystem.ListOptions{Limit: 200})
			if err != nil || len(items.Items) != 0 {
				t.Fatalf("malicious package changed target: %v %v", items, err)
			}
		})
	}
}

func TestBackupRetentionOwnedScopePinnedReaderAndDeleteRecovery(t *testing.T) {
	f := newBackupFixture(t, Options{})
	ctx := context.Background()
	diskWrite(t, filepath.Join(f.sourceRoot, "file"), []byte("content"))
	var ids []string
	for i := 0; i < 3; i++ {
		spec := f.spec(t, "file")
		if _, err := f.manager.Create(ctx, f.source, spec); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, spec.ID)
	}
	_, _, close, err := f.manager.openSnapshot(ctx, ids[0])
	if err != nil {
		t.Fatal(err)
	}
	diskWrite(t, filepath.Join(f.repo, "objects", "unrecognized.zip"), []byte("untouched"))
	policy := Retention{OwnerID: "owner", Source: backupRef("source"), Path: "file", KeepLast: 1}
	result, err := f.manager.Prune(ctx, policy, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Removed) != 1 || result.Removed[0] != ids[1] || len(result.Busy) != 1 {
		t.Fatalf("retention %v", result)
	}
	close()
	var snap Snapshot
	if err = f.manager.readRecord("snapshots/"+ids[0]+".json", &snap); err != nil {
		t.Fatal(err)
	}
	snap.State = "pruning"
	if err = f.manager.writeRecord("snapshots/"+ids[0]+".json", snap); err != nil {
		t.Fatal(err)
	}
	if err = f.manager.root.Remove(archivePath(ids[0])); err != nil {
		t.Fatal(err)
	}
	f.reopen(t)
	result, err = f.manager.Prune(ctx, policy, time.Now())
	if err != nil || len(result.Removed) != 1 || result.Removed[0] != ids[0] {
		t.Fatalf("delete receipt recovery %v %v", result, err)
	}
	if body, err := os.ReadFile(filepath.Join(f.repo, "objects", "unrecognized.zip")); err != nil || string(body) != "untouched" {
		t.Fatalf("unowned file removed %q %v", body, err)
	}
	items, err := f.manager.List(ctx)
	if err != nil || len(items) != 1 || items[0].ID != ids[2] {
		t.Fatalf("retained snapshots %+v %v", items, err)
	}
	policy.OwnerID = "another-owner"
	result, err = f.manager.Prune(ctx, policy, time.Now())
	if err != nil || len(result.Removed) != 0 {
		t.Fatalf("owner scope %v %v", result, err)
	}
}

func TestBackupAndRestoreInterruptedReceiptsNeverReplay(t *testing.T) {
	f := newBackupFixture(t, Options{})
	ctx := context.Background()
	diskWrite(t, filepath.Join(f.sourceRoot, "file"), []byte("archive"))
	spec := f.spec(t, "file")
	created, err := f.manager.Create(ctx, f.source, spec)
	if err != nil {
		t.Fatal(err)
	}
	// Persist the actual archive commit with its final success checkpoint
	// deliberately rewound to the pre-rename boundary, then close/reopen roots.
	var create createCheckpoint
	if err = f.manager.readRecord("create/"+spec.ID+".json", &create); err != nil {
		t.Fatal(err)
	}
	create.Result.Stage = "committing"
	create.Result.Snapshot = nil
	if err = f.manager.writeRecord("create/"+spec.ID+".json", create); err != nil {
		t.Fatal(err)
	}
	f.reopen(t)
	got, err := f.manager.Create(ctx, f.source, spec)
	if err != nil || got.Snapshot.ArchiveHash != created.Snapshot.ArchiveHash {
		t.Fatalf("archive commit reconciliation %+v %v", got, err)
	}
	plan := f.plan(t, spec.ID, "restore")
	restore := restoreSpec(plan)
	upload := filesystem.UploadSpec{ID: model.ID(), OwnerID: "owner", Path: "restore", Total: 7, Hash: hashBytes([]byte("archive")), ExpectedVersion: filesystem.MissingVersion, SourceName: "file", SourceFingerprint: plan.ArchiveHash + ":" + plan.ArchiveObjectID}
	if _, err = f.target.BeginUpload(ctx, upload); err != nil {
		t.Fatal(err)
	}
	if _, err = f.target.UploadChunk(ctx, upload.ID, 0, []byte("archive"), upload.Hash); err != nil {
		t.Fatal(err)
	}
	if _, err = f.target.CommitUpload(ctx, upload.ID); err != nil {
		t.Fatal(err)
	}
	cp := restoreCheckpoint{Spec: restore, Result: Result{ID: restore.ID, Stage: "restoring"}, Operation: "upload", Upload: &upload}
	if err = f.manager.writeRecord("restore/"+restore.ID+".json", cp); err != nil {
		t.Fatal(err)
	}
	f.reopen(t)
	partial, err := f.manager.Restore(ctx, f.target, restore)
	if !errors.Is(err, ErrInterrupted) || partial.Completed != 1 || partial.Unknown || !partial.Partial {
		t.Fatalf("commit receipt %+v %v", partial, err)
	}
	diskWrite(t, filepath.Join(f.targetRoot, "restore"), []byte("user edit"))
	again, err := f.manager.Restore(ctx, f.target, restore)
	if !errors.Is(err, ErrInterrupted) || again.Completed != 1 {
		t.Fatalf("partial replay %+v %v", again, err)
	}
	body, err := os.ReadFile(filepath.Join(f.targetRoot, "restore"))
	if err != nil || string(body) != "user edit" {
		t.Fatalf("repeat overwrote user edit %q %v", body, err)
	}
	plan = f.plan(t, spec.ID, "unknown")
	restore = restoreSpec(plan)
	cp = restoreCheckpoint{Spec: restore, Result: Result{ID: restore.ID, Stage: "restoring"}, Operation: "mkdir"}
	if err = f.manager.writeRecord("restore/"+restore.ID+".json", cp); err != nil {
		t.Fatal(err)
	}
	partial, err = f.manager.Restore(ctx, f.target, restore)
	if !errors.Is(err, ErrInterrupted) || !partial.Unknown {
		t.Fatalf("unknown metadata/mkdir replay %+v %v", partial, err)
	}
}

type durableTestHooks struct {
	root    string
	unknown bool
}

func (h *durableTestHooks) Run(ctx context.Context, call HookCall) (HookReceipt, error) {
	if err := ctx.Err(); err != nil {
		return HookReceipt{}, err
	}
	p := filepath.Join(h.root, strings.ReplaceAll(call.ID, ":", "_"))
	file, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return HookReceipt{}, err
	}
	_, err = io.WriteString(file, call.Phase)
	if err == nil {
		err = file.Sync()
	}
	if e := file.Close(); err == nil {
		err = e
	}
	receipt := HookReceipt{ID: call.ID, State: "succeeded", At: time.Now().UTC()}
	if h.unknown && call.Phase == "before" {
		receipt.State = "unknown"
	}
	return receipt, err
}
func (h *durableTestHooks) Reconcile(ctx context.Context, call HookCall) (HookReceipt, error) {
	if err := ctx.Err(); err != nil {
		return HookReceipt{}, err
	}
	body, err := os.ReadFile(filepath.Join(h.root, strings.ReplaceAll(call.ID, ":", "_")))
	if err != nil {
		return HookReceipt{ID: call.ID, State: "unknown"}, err
	}
	state := "succeeded"
	if string(body) != call.Phase || h.unknown && call.Phase == "before" {
		state = "unknown"
	}
	return HookReceipt{ID: call.ID, State: state, At: time.Now().UTC()}, nil
}

func TestBackupHooksNeedAdapterAndUnknownDoesNotRepeat(t *testing.T) {
	f := newBackupFixture(t, Options{})
	ctx := context.Background()
	diskWrite(t, filepath.Join(f.sourceRoot, "file"), []byte("data"))
	spec := f.spec(t, "file")
	spec.Consistency.Mode = "pause"
	if _, err := f.manager.Create(ctx, f.source, spec); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("missing adapter pretended consistency: %v", err)
	}
	hooks := &durableTestHooks{root: t.TempDir(), unknown: true}
	f.manager.opts.Hooks = hooks
	f.options.Hooks = hooks
	result, err := f.manager.Create(ctx, f.source, spec)
	if err == nil || !result.Unknown {
		t.Fatalf("unknown hook accepted %+v %v", result, err)
	}
	f.reopen(t)
	if _, err = f.manager.Create(ctx, f.source, spec); !errors.Is(err, ErrInterrupted) {
		t.Fatalf("unknown hook replay %v", err)
	}
	entries, err := os.ReadDir(hooks.root)
	if err != nil || len(entries) != 2 {
		t.Fatalf("hook effects %v %v", entries, err)
	}
	hooks.unknown = false
	spec = f.spec(t, "file")
	spec.Consistency.Mode = "hooks"
	result, err = f.manager.Create(ctx, f.source, spec)
	if err != nil || result.Snapshot.Before.State != "succeeded" || result.Snapshot.After.State != "succeeded" {
		t.Fatalf("durable custom hook %+v %v", result, err)
	}
}

func TestBackupFailedOperationReconcilesCompensationReceipt(t *testing.T) {
	f := newBackupFixture(t, Options{})
	ctx := context.Background()
	diskWrite(t, filepath.Join(f.sourceRoot, "file"), []byte("data"))
	spec := f.spec(t, "file")
	spec.Consistency.Mode = "hooks"
	hooks := &durableTestHooks{root: t.TempDir()}
	f.manager.opts.Hooks = hooks
	f.options.Hooks = hooks
	call := HookCall{ID: spec.ID + ":before", BackupID: spec.ID, OwnerID: spec.OwnerID, Resource: spec.Source, Strategy: spec.Consistency, Phase: "before"}
	before, err := hooks.Run(ctx, call)
	if err != nil {
		t.Fatal(err)
	}
	call.ID = spec.ID + ":after"
	call.Phase = "after"
	after, err := hooks.Run(ctx, call)
	if err != nil {
		t.Fatal(err)
	}
	after.State = "running"
	cp := createCheckpoint{Spec: spec, Result: Result{ID: spec.ID, Stage: "failed", Error: "original operation failed"}, Before: &before, After: &after}
	if err = f.manager.writeRecord("create/"+spec.ID+".json", cp); err != nil {
		t.Fatal(err)
	}
	f.reopen(t)
	result, err := f.manager.Create(ctx, f.source, spec)
	if !errors.Is(err, ErrInterrupted) || result.CleanupPending {
		t.Fatalf("compensation receipt not reconciled %+v %v", result, err)
	}
	if err = f.manager.readRecord("create/"+spec.ID+".json", &cp); err != nil {
		t.Fatal(err)
	}
	if cp.After.State != "succeeded" {
		t.Fatalf("original compensation not observed %+v", cp.After)
	}
	entries, err := os.ReadDir(hooks.root)
	if err != nil || len(entries) != 2 {
		t.Fatalf("compensation repeated %v %v", entries, err)
	}
}

package containers

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

func saveTestManager(t *testing.T, root string) *Manager {
	t.Helper()
	m, err := New(Options{StateRoot: root, Authorize: func(_ context.Context, actor string) error {
		if actor == "admin" || actor == "other" {
			return nil
		}
		return ErrForbidden
	}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.Close() })
	return m
}
func uploadProject(t *testing.T, m *Manager, id string, expected uint64, source []byte) ProjectSave {
	t.Helper()
	ctx := context.Background()
	s, err := m.PrepareProjectSave(ctx, "admin", id, "demo", expected, int64(len(source)), sourceHash(source))
	if err != nil {
		t.Fatal(err)
	}
	for offset := 0; offset < len(source); {
		end := min(offset+ProjectChunkBytes, len(source))
		data := source[offset:end]
		s, err = m.WriteProjectChunk(ctx, "admin", id, int64(offset), data, sourceHash(data))
		if err != nil {
			t.Fatal(err)
		}
		offset = end
	}
	return s
}
func TestProjectSaveDurableChunksAndCommitRecovery(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	m := saveTestManager(t, root)
	data := []byte(strings.Repeat("# Compose comment testing UTF8 中文\n", 20000))
	s := uploadProject(t, m, "save_one", 0, data)
	if s.Offset != int64(len(data)) {
		t.Fatal(s)
	}
	first := data[:ProjectChunkBytes]
	if _, err := m.WriteProjectChunk(ctx, "admin", s.ID, 0, first, sourceHash(first)); err != nil {
		t.Fatal("lost ACK retry", err)
	}
	if _, err := m.PrepareProjectSave(ctx, "admin", s.ID, "demo", 0, int64(len(data)), sourceHash([]byte("changed"))); !errors.Is(err, ErrConflict) {
		t.Fatal("rebound save id", err)
	}
	if _, err := m.ProjectSaveStatus(ctx, "other", s.ID); !errors.Is(err, ErrForbidden) {
		t.Fatal("actor isolation", err)
	}
	if _, err := m.CommitProjectSave(ctx, "denied", s.ID); !errors.Is(err, ErrForbidden) {
		t.Fatal("permission revoked", err)
	}
	m = saveTestManager(t, root)
	s, err := m.CommitProjectSave(ctx, "admin", s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if s.State != "COMMITTED" || s.Project.Revision != 1 || s.Project.AppliedRevision != 0 || s.Project.Source != "" {
		t.Fatal(s)
	}
	var restored []byte
	for offset := int64(0); offset < int64(len(data)); {
		chunk, err := m.ReadProjectChunk(ctx, "admin", "demo", 1, offset, ProjectChunkBytes)
		if err != nil {
			t.Fatal(err)
		}
		if chunk.SHA256 != sourceHash(data) {
			t.Fatal("source checksum")
		}
		restored = append(restored, chunk.Data...)
		offset += int64(len(chunk.Data))
	}
	if !bytes.Equal(restored, data) {
		t.Fatal("source bytes differ")
	}
	// Reproduce a crash after public project index commit but before save ACK.
	s.State = "COMMITTING"
	s.Project = nil
	if err = m.saveProjectSave("admin", s); err != nil {
		t.Fatal(err)
	}
	// Subsequent revision must not destroy reconciliation for the earlier one.
	if _, err = m.SaveProject(ctx, "admin", "demo", 1, "services: {}\n"); err != nil {
		t.Fatal(err)
	}
	m = saveTestManager(t, root)
	s, err = m.CommitProjectSave(ctx, "admin", s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if s.State != "COMMITTED" || s.Project.Revision != 1 {
		t.Fatal(s)
	}
	q, err := m.Query(ctx, "admin", Query{Kind: "project", ProjectID: "demo", Details: true})
	if err != nil {
		t.Fatal(err)
	}
	if q.Project.Source != "" || q.Project.Revision != 2 {
		t.Fatal("query emitted source or replayed commit", q)
	}
}
func TestProjectSavePartialWriteAndConflicts(t *testing.T) {
	ctx := context.Background()
	m := saveTestManager(t, t.TempDir())
	data := []byte("services: {}\n")
	s, err := m.PrepareProjectSave(ctx, "admin", "partial", "demo", 0, int64(len(data)), sourceHash(data))
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(m.projectPartPath(s.ID), []byte("unacknowledged bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = m.CommitProjectSave(ctx, "admin", s.ID); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	if _, err = m.WriteProjectChunk(ctx, "admin", s.ID, 1, data, sourceHash(data)); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	if _, err = m.WriteProjectChunk(ctx, "admin", s.ID, 0, data, sourceHash([]byte("wrong"))); !errors.Is(err, ErrIdentity) {
		t.Fatal(err)
	}
	if _, err = m.WriteProjectChunk(ctx, "admin", s.ID, 0, data, sourceHash(data)); err != nil {
		t.Fatal(err)
	}
	if _, err = m.SaveProject(ctx, "admin", "demo", 0, "different source"); err != nil {
		t.Fatal(err)
	}
	if _, err = m.CommitProjectSave(ctx, "admin", s.ID); !errors.Is(err, ErrConflict) {
		t.Fatal("stale revision accepted", err)
	}
	p, err := m.loadProject("demo", true)
	if err != nil || p.Source != "different source" {
		t.Fatal("source overwritten", p, err)
	}
}

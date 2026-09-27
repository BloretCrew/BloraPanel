package containers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const ProjectChunkBytes = 64 << 10

type ProjectSave struct {
	ID               string    `json:"id"`
	ProjectID        string    `json:"projectId"`
	ExpectedRevision uint64    `json:"expectedRevision"`
	Total            int64     `json:"total"`
	SHA256           string    `json:"sha256"`
	Offset           int64     `json:"offset"`
	State            string    `json:"state"`
	Project          *Project  `json:"project,omitempty"`
	UpdatedAt        time.Time `json:"updatedAt"`
}
type projectSaveRecord struct {
	Actor string      `json:"actor"`
	Save  ProjectSave `json:"save"`
}
type revisionCommit struct {
	SaveID  string  `json:"saveId"`
	SHA256  string  `json:"sha256"`
	Project Project `json:"project"`
}
type ProjectChunk struct {
	ProjectID string `json:"projectId"`
	Revision  uint64 `json:"revision"`
	Offset    int64  `json:"offset"`
	Total     int64  `json:"total"`
	SHA256    string `json:"sha256"`
	Data      []byte `json:"data"`
}

func sourceHash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
func (m *Manager) projectSavePath(id string) string {
	return filepath.Join(m.options.StateRoot, "project-saves", id+".json")
}
func (m *Manager) projectPartPath(id string) string {
	return filepath.Join(m.options.StateRoot, "project-saves", id+".part")
}
func (m *Manager) loadProjectSave(actor, id string) (ProjectSave, error) {
	var r projectSaveRecord
	if !safeID.MatchString(id) {
		return r.Save, ErrIdentity
	}
	b, err := os.ReadFile(m.projectSavePath(id))
	if err != nil {
		return r.Save, err
	}
	if len(b) > 16<<10 {
		return r.Save, ErrIdentity
	}
	if err = json.Unmarshal(b, &r); err != nil {
		return r.Save, err
	}
	if r.Actor != actor {
		return ProjectSave{}, ErrForbidden
	}
	s := r.Save
	if s.ID != id || !safeID.MatchString(s.ProjectID) || s.Total < 1 || s.Total > 1<<20 || s.Offset < 0 || s.Offset > s.Total || !fullID.MatchString(s.SHA256) {
		return ProjectSave{}, ErrIdentity
	}
	return s, nil
}
func (m *Manager) saveProjectSave(actor string, s ProjectSave) error {
	return atomicJSON(m.projectSavePath(s.ID), projectSaveRecord{Actor: actor, Save: s})
}

// PrepareProjectSave binds the upload ID to actor, project, base revision and
// source digest. All acknowledged bytes have reached an on-node fsync checkpoint.
func (m *Manager) PrepareProjectSave(ctx context.Context, actor, saveID, projectID string, expectedRevision uint64, total int64, sha256 string) (ProjectSave, error) {
	if err := m.authorize(ctx, actor); err != nil {
		return ProjectSave{}, err
	}
	if !safeID.MatchString(saveID) || !safeID.MatchString(projectID) || total < 1 || total > 1<<20 || !fullID.MatchString(sha256) || expectedRevision >= 128 {
		return ProjectSave{}, ErrIdentity
	}
	done, err := m.lock(ctx, "project-saves")
	if err != nil {
		return ProjectSave{}, err
	}
	defer done()
	if s, e := m.loadProjectSave(actor, saveID); e == nil {
		if s.ProjectID != projectID || s.ExpectedRevision != expectedRevision || s.Total != total || s.SHA256 != sha256 {
			return s, ErrConflict
		}
		return s, nil
	} else if !errors.Is(e, os.ErrNotExist) {
		return ProjectSave{}, e
	}
	p, e := m.loadProject(projectID, false)
	if errors.Is(e, os.ErrNotExist) {
		if expectedRevision != 0 {
			return ProjectSave{}, ErrConflict
		}
	} else if e != nil {
		return ProjectSave{}, e
	} else if p.Revision != expectedRevision {
		return ProjectSave{}, ErrConflict
	}
	entries, e := os.ReadDir(filepath.Join(m.options.StateRoot, "project-saves"))
	if e != nil {
		return ProjectSave{}, e
	}
	count := 0
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".json" {
			count++
		}
	}
	if count >= m.options.MaxProjectSaves {
		return ProjectSave{}, errors.New("Compose save journal budget reached")
	}
	s := ProjectSave{ID: saveID, ProjectID: projectID, ExpectedRevision: expectedRevision, Total: total, SHA256: sha256, State: "PREPARED", UpdatedAt: time.Now().UTC()}
	return s, m.saveProjectSave(actor, s)
}
func (m *Manager) WriteProjectChunk(ctx context.Context, actor, saveID string, offset int64, data []byte, sha256 string) (ProjectSave, error) {
	if err := m.authorize(ctx, actor); err != nil {
		return ProjectSave{}, err
	}
	if len(data) == 0 || len(data) > ProjectChunkBytes || offset < 0 || !fullID.MatchString(sha256) || sourceHash(data) != sha256 {
		return ProjectSave{}, ErrIdentity
	}
	done, err := m.lock(ctx, "project-save:"+saveID)
	if err != nil {
		return ProjectSave{}, err
	}
	defer done()
	s, err := m.loadProjectSave(actor, saveID)
	if err != nil {
		return s, err
	}
	if offset > s.Total-int64(len(data)) || offset > s.Offset {
		return s, ErrConflict
	}
	if s.State != "PREPARED" {
		return s, ErrConflict
	}
	f, err := os.OpenFile(m.projectPartPath(saveID), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return s, err
	}
	defer f.Close()
	if offset < s.Offset {
		if offset+int64(len(data)) > s.Offset {
			return s, ErrConflict
		}
		old := make([]byte, len(data))
		if _, err = f.ReadAt(old, offset); err != nil {
			return s, err
		}
		if !bytes.Equal(old, data) {
			return s, ErrConflict
		}
		return s, nil
	}
	// Unacknowledged bytes from an interrupted write are overwritten; the
	// persisted offset is the only checkpoint exposed to clients.
	if err = f.Truncate(s.Offset); err != nil {
		return s, err
	}
	if _, err = f.WriteAt(data, offset); err != nil {
		return s, err
	}
	if err = f.Sync(); err != nil {
		return s, err
	}
	s.Offset += int64(len(data))
	s.UpdatedAt = time.Now().UTC()
	return s, m.saveProjectSave(actor, s)
}
func (m *Manager) reconcileProjectSave(actor string, s ProjectSave) (ProjectSave, error) {
	if s.State != "COMMITTING" {
		return s, nil
	}
	p, err := m.loadProject(s.ProjectID, false)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	if p.Revision < s.ExpectedRevision+1 {
		return s, nil
	}
	dir, _ := m.projectDir(s.ProjectID)
	b, err := os.ReadFile(filepath.Join(dir, "revisions", fmt.Sprintf("%020d.commit.json", s.ExpectedRevision+1)))
	if err != nil {
		return s, ErrConflict
	}
	var commit revisionCommit
	if json.Unmarshal(b, &commit) != nil || commit.SaveID != s.ID || commit.SHA256 != s.SHA256 {
		return s, ErrConflict
	}
	b, err = os.ReadFile(filepath.Join(dir, "revisions", fmt.Sprintf("%020d.yaml", s.ExpectedRevision+1)))
	if err != nil {
		return s, err
	}
	if int64(len(b)) != s.Total || sourceHash(b) != s.SHA256 {
		return s, ErrIdentity
	}
	s.State = "COMMITTED"
	s.Project = &commit.Project
	s.UpdatedAt = time.Now().UTC()
	if err = m.saveProjectSave(actor, s); err != nil {
		return s, err
	}
	if err = os.Remove(m.projectPartPath(s.ID)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return s, err
	}
	return s, nil
}
func (m *Manager) ProjectSaveStatus(ctx context.Context, actor, saveID string) (ProjectSave, error) {
	if err := m.authorize(ctx, actor); err != nil {
		return ProjectSave{}, err
	}
	done, err := m.lock(ctx, "project-save:"+saveID)
	if err != nil {
		return ProjectSave{}, err
	}
	defer done()
	s, err := m.loadProjectSave(actor, saveID)
	if err != nil {
		return s, err
	}
	return m.reconcileProjectSave(actor, s)
}
func (m *Manager) CommitProjectSave(ctx context.Context, actor, saveID string) (ProjectSave, error) {
	if err := m.authorize(ctx, actor); err != nil {
		return ProjectSave{}, err
	}
	done, err := m.lock(ctx, "project-save:"+saveID)
	if err != nil {
		return ProjectSave{}, err
	}
	defer done()
	s, err := m.loadProjectSave(actor, saveID)
	if err != nil {
		return s, err
	}
	unlock, err := m.lock(ctx, "project-config:"+s.ProjectID)
	if err != nil {
		return s, err
	}
	defer unlock()
	s, err = m.reconcileProjectSave(actor, s)
	if err != nil || s.State == "COMMITTED" {
		return s, err
	}
	if s.Offset != s.Total {
		return s, ErrConflict
	}
	b, err := os.ReadFile(m.projectPartPath(saveID))
	if err != nil {
		return s, err
	}
	if int64(len(b)) != s.Total || sourceHash(b) != s.SHA256 {
		return s, ErrIdentity
	}
	s.State = "COMMITTING"
	s.UpdatedAt = time.Now().UTC()
	if err = m.saveProjectSave(actor, s); err != nil {
		return s, err
	}
	if _, err = m.saveProjectRevision(ctx, actor, s.ProjectID, s.ExpectedRevision, string(b), s.ID); err != nil {
		return s, err
	}
	return m.reconcileProjectSave(actor, s)
}
func (m *Manager) ReadProjectChunk(ctx context.Context, actor, projectID string, revision uint64, offset int64, length int) (ProjectChunk, error) {
	var out ProjectChunk
	if err := m.authorize(ctx, actor); err != nil {
		return out, err
	}
	if length < 1 || length > ProjectChunkBytes || offset < 0 {
		return out, ErrIdentity
	}
	p, err := m.loadProject(projectID, false)
	if err != nil {
		return out, err
	}
	if revision == 0 || revision > p.Revision {
		return out, ErrConflict
	}
	dir, _ := m.projectDir(projectID)
	f, err := os.Open(filepath.Join(dir, "revisions", fmt.Sprintf("%020d.yaml", revision)))
	if err != nil {
		return out, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return out, err
	}
	if info.Size() > 1<<20 || offset > info.Size() {
		return out, ErrIdentity
	}
	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return out, err
	}
	data := make([]byte, min(int64(length), info.Size()-offset))
	if len(data) > 0 {
		if _, err = f.ReadAt(data, offset); err != nil {
			return out, err
		}
	}
	return ProjectChunk{ProjectID: projectID, Revision: revision, Offset: offset, Total: info.Size(), SHA256: hex.EncodeToString(h.Sum(nil)), Data: data}, nil
}

package master

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
)

const workspaceNamespace = "workspace_cloud"

type workspaceCloudRecord struct {
	WorkspaceID    string          `json:"workspaceId"`
	UserID         string          `json:"userId"`
	DeviceID       string          `json:"deviceId"`
	Title          string          `json:"title"`
	Revision       int64           `json:"revision"`
	SchemaVersion  int             `json:"schemaVersion"`
	IncludeContent bool            `json:"includeContent"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	State          json.RawMessage `json:"state"`
}

type workspaceMetadata struct {
	WorkspaceID    string    `json:"workspaceId"`
	Title          string    `json:"title"`
	DeviceID       string    `json:"deviceId"`
	Revision       int64     `json:"revision"`
	SchemaVersion  int       `json:"schemaVersion"`
	IncludeContent bool      `json:"includeContent"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func workspaceRecordID(userID, workspaceID string) string { return userID + ":" + workspaceID }

func validWorkspaceID(id string) bool {
	if id == "" || len(id) > 128 || strings.ContainsAny(id, "/\\:\x00\r\n") {
		return false
	}
	return true
}

func (s *Server) registerWorkspaces() {
	s.mux.HandleFunc("GET /api/v1/workspaces", s.auth(s.workspaceList))
	s.mux.HandleFunc("GET /api/v1/workspaces/{id}", s.auth(s.workspaceGet))
	s.mux.HandleFunc("PUT /api/v1/workspaces/{id}", s.auth(s.workspacePut))
	s.mux.HandleFunc("DELETE /api/v1/workspaces/{id}", s.auth(s.workspaceDelete))
}

func workspaceMeta(record workspaceCloudRecord) workspaceMetadata {
	return workspaceMetadata{WorkspaceID: record.WorkspaceID, Title: record.Title, DeviceID: record.DeviceID, Revision: record.Revision, SchemaVersion: record.SchemaVersion, IncludeContent: record.IncludeContent, UpdatedAt: record.UpdatedAt}
}

func (s *Server) workspaceList(w http.ResponseWriter, r *http.Request, u model.User) {
	ids, err := s.store.RecordIDs(r.Context(), workspaceNamespace)
	if err != nil {
		fail(w, 500, "WORKSPACE_STORE_FAILED", err.Error())
		return
	}
	prefix := u.ID + ":"
	items := make([]workspaceMetadata, 0)
	for _, id := range ids {
		if !strings.HasPrefix(id, prefix) {
			continue
		}
		var record workspaceCloudRecord
		if _, err := s.store.Record(r.Context(), workspaceNamespace, id, &record); err != nil || record.UserID != u.ID {
			continue
		}
		items = append(items, workspaceMeta(record))
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].UpdatedAt.Equal(items[j].UpdatedAt) {
			return items[i].WorkspaceID < items[j].WorkspaceID
		}
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
	reply(w, 200, map[string]any{"items": items})
}

func (s *Server) workspaceGet(w http.ResponseWriter, r *http.Request, u model.User) {
	id := r.PathValue("id")
	if !validWorkspaceID(id) {
		fail(w, 400, "INVALID_WORKSPACE", "工作区标识无效")
		return
	}
	var record workspaceCloudRecord
	if _, err := s.store.Record(r.Context(), workspaceNamespace, workspaceRecordID(u.ID, id), &record); err != nil || record.UserID != u.ID {
		fail(w, 404, "WORKSPACE_NOT_FOUND", "工作区副本不存在")
		return
	}
	reply(w, 200, map[string]any{"metadata": workspaceMeta(record), "state": json.RawMessage(record.State)})
}

func decodeWorkspaceBody(w http.ResponseWriter, r *http.Request, value any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(value); err != nil {
		fail(w, 400, "INVALID_WORKSPACE", "工作区副本内容无效")
		return false
	}
	var trailing any
	if err := d.Decode(&trailing); !errors.Is(err, io.EOF) {
		fail(w, 400, "INVALID_WORKSPACE", "仅允许一个工作区 JSON 对象")
		return false
	}
	return true
}

func stripWorkspaceContent(raw json.RawMessage) (json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return nil, errors.New("工作区状态必须是 JSON 对象")
	}
	// Layout, filters and resource references remain available on another
	// device; drafts, terminal screens and browser upload handles are opt-in
	// content and must never be uploaded by the default sync action.
	object["drafts"] = json.RawMessage(`{}`)
	object["terminals"] = json.RawMessage(`{}`)
	object["uploads"] = json.RawMessage(`{}`)
	if raw, ok := object["preferences"]; ok {
		var preferences map[string]json.RawMessage
		if err := json.Unmarshal(raw, &preferences); err != nil {
			return nil, errors.New("工作区偏好必须是对象")
		}
		delete(preferences, "notifications")
		clean, err := json.Marshal(preferences)
		if err != nil {
			return nil, err
		}
		object["preferences"] = clean
	}
	return json.Marshal(object)
}

func (s *Server) workspacePut(w http.ResponseWriter, r *http.Request, u model.User) {
	id := r.PathValue("id")
	if !validWorkspaceID(id) {
		fail(w, 400, "INVALID_WORKSPACE", "工作区标识无效")
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var in struct {
		Title          string          `json:"title"`
		DeviceID       string          `json:"deviceId"`
		BaseRevision   int64           `json:"baseRevision"`
		SchemaVersion  int             `json:"schemaVersion"`
		IncludeContent bool            `json:"includeContent"`
		State          json.RawMessage `json:"state"`
	}
	if !decodeWorkspaceBody(w, r, &in) {
		return
	}
	if in.DeviceID == "" || len(in.DeviceID) > 128 || len(in.Title) > 160 || in.SchemaVersion < 1 || len(in.State) == 0 || len(in.State) > 7<<20 {
		fail(w, 400, "INVALID_WORKSPACE", "工作区元数据或状态超出范围")
		return
	}
	var stateObject map[string]json.RawMessage
	if err := json.Unmarshal(in.State, &stateObject); err != nil || stateObject == nil {
		fail(w, 400, "INVALID_WORKSPACE", "工作区状态必须是 JSON 对象")
		return
	}
	if raw, ok := stateObject["userId"]; ok {
		var owner string
		if json.Unmarshal(raw, &owner) != nil || owner != u.ID {
			fail(w, 403, "WORKSPACE_OWNER_MISMATCH", "工作区状态不属于当前账号")
			return
		}
	} else {
		fail(w, 400, "INVALID_WORKSPACE", "工作区状态缺少所属账号")
		return
	}
	if raw, ok := stateObject["workspaceId"]; ok {
		var stateID string
		if json.Unmarshal(raw, &stateID) != nil || stateID != id {
			fail(w, 400, "INVALID_WORKSPACE", "工作区状态标识与请求路径不一致")
			return
		}
	} else {
		fail(w, 400, "INVALID_WORKSPACE", "工作区状态缺少工作区标识")
		return
	}
	state := in.State
	if !in.IncludeContent {
		var err error
		state, err = stripWorkspaceContent(state)
		if err != nil {
			fail(w, 400, "INVALID_WORKSPACE", err.Error())
			return
		}
	}
	key := workspaceRecordID(u.ID, id)
	data, err := s.store.UserMetadataMutation(r.Context(), u.ID, r.Header.Get("Idempotency-Key"), "workspace.put", struct {
		WorkspaceID    string          `json:"workspaceId"`
		Title          string          `json:"title"`
		DeviceID       string          `json:"deviceId"`
		BaseRevision   int64           `json:"baseRevision"`
		SchemaVersion  int             `json:"schemaVersion"`
		IncludeContent bool            `json:"includeContent"`
		State          json.RawMessage `json:"state"`
	}{id, strings.TrimSpace(in.Title), in.DeviceID, in.BaseRevision, in.SchemaVersion, in.IncludeContent, state}, func(tx *sql.Tx) (any, error) {
		var old workspaceCloudRecord
		var oldBytes []byte
		var storedRevision int64
		err := tx.QueryRowContext(r.Context(), "SELECT revision,document FROM records WHERE namespace=? AND id=?", workspaceNamespace, key).Scan(&storedRevision, &oldBytes)
		if err == nil {
			if err := json.Unmarshal(oldBytes, &old); err != nil {
				return nil, err
			}
			if old.UserID != u.ID || old.Revision != in.BaseRevision {
				return nil, errors.New("workspace revision conflict")
			}
		} else if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		} else if in.BaseRevision != 0 {
			return nil, errors.New("workspace revision conflict")
		}
		record := workspaceCloudRecord{WorkspaceID: id, UserID: u.ID, DeviceID: in.DeviceID, Title: strings.TrimSpace(in.Title), Revision: old.Revision + 1, SchemaVersion: in.SchemaVersion, IncludeContent: in.IncludeContent, UpdatedAt: time.Now().UTC(), State: state}
		b, err := json.Marshal(record)
		if err != nil {
			return nil, err
		}
		if storedRevision == 0 {
			if _, err = tx.ExecContext(r.Context(), "INSERT INTO records(namespace,id,revision,document) VALUES(?,?,1,?)", workspaceNamespace, key, b); err != nil {
				return nil, err
			}
		} else {
			result, updateErr := tx.ExecContext(r.Context(), "UPDATE records SET revision=revision+1,document=? WHERE namespace=? AND id=? AND revision=?", b, workspaceNamespace, key, storedRevision)
			if updateErr != nil {
				return nil, updateErr
			}
			count, _ := result.RowsAffected()
			if count != 1 {
				return nil, errors.New("workspace revision conflict")
			}
		}
		return map[string]any{"metadata": workspaceMeta(record)}, nil
	})
	if err != nil {
		if strings.Contains(err.Error(), "workspace revision conflict") || errors.Is(err, storage.ErrConflict) {
			fail(w, 409, "WORKSPACE_CONFLICT", "云端工作区版本已变化")
		} else {
			fail(w, 409, "WORKSPACE_STORE_FAILED", err.Error())
		}
		return
	}
	reply(w, 200, data)
}

func (s *Server) workspaceDelete(w http.ResponseWriter, r *http.Request, u model.User) {
	id := r.PathValue("id")
	if !validWorkspaceID(id) {
		fail(w, 400, "INVALID_WORKSPACE", "工作区标识无效")
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	_, err := s.store.UserMetadataMutation(r.Context(), u.ID, r.Header.Get("Idempotency-Key"), "workspace.delete", struct {
		WorkspaceID string `json:"workspaceId"`
	}{id}, func(tx *sql.Tx) (any, error) {
		if _, err := tx.ExecContext(r.Context(), "DELETE FROM records WHERE namespace=? AND id=?", workspaceNamespace, workspaceRecordID(u.ID, id)); err != nil {
			return nil, err
		}
		return map[string]bool{"deleted": true}, nil
	})
	if err != nil {
		fail(w, 500, "WORKSPACE_STORE_FAILED", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

package master

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) registerAdministration() {
	s.mux.HandleFunc("GET /api/v1/audit", s.auth(s.audit))
	s.mux.HandleFunc("PATCH /api/v1/nodes/{id}", s.auth(s.nodeSettings))
	s.mux.HandleFunc("POST /api/v1/nodes/{id}/rotation", s.auth(s.issueNodeRotation))
	s.mux.HandleFunc("POST /api/v1/agent/rotate", s.rotateNodeKey)
	s.mux.HandleFunc("PATCH /api/v1/users/{id}", s.auth(s.updateUser))
	s.mux.HandleFunc("POST /api/v1/users/{id}/password", s.auth(s.changePassword))
	s.mux.HandleFunc("GET /api/v1/roles", s.auth(s.roles))
	s.mux.HandleFunc("PUT /api/v1/roles/{id}", s.auth(s.putRole))
	s.mux.HandleFunc("POST /api/v1/roles/{id}/apply", s.auth(s.applyRole))
}
func administrationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrForbidden):
		fail(w, 403, "FORBIDDEN", err.Error())
	case errors.Is(err, sql.ErrNoRows):
		fail(w, 404, "NOT_FOUND", err.Error())
	case errors.Is(err, storage.ErrRequestMismatch):
		fail(w, 409, "IDEMPOTENCY_CONFLICT", err.Error())
	case errors.Is(err, storage.ErrConflict), errors.Is(err, storage.ErrLastAdmin):
		fail(w, 409, "CONFIG_CONFLICT", err.Error())
	default:
		fail(w, 400, "INVALID_REQUEST", err.Error())
	}
}
func (s *Server) updateUser(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.admin(w, u) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var input model.UserPatch
	if !decode(w, r, &input) {
		return
	}
	if input.Name != nil && (len(strings.TrimSpace(*input.Name)) < 1 || len(*input.Name) > 80) {
		fail(w, 400, "INVALID_NAME", "用户名必须为1～80字节")
		return
	}
	updated, err := s.store.UpdateUser(r.Context(), u.ID, r.PathValue("id"), r.Header.Get("Idempotency-Key"), input)
	if err != nil {
		administrationError(w, err)
		return
	}
	s.revokeStreams(updated.ID)
	reply(w, 200, map[string]any{"user": updated})
}
func (s *Server) changePassword(w http.ResponseWriter, r *http.Request, u model.User) {
	id := r.PathValue("id")
	if id != u.ID && !s.admin(w, u) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var input struct {
		Password        string `json:"password"`
		CurrentPassword string `json:"currentPassword"`
		Revision        int64  `json:"revision"`
	}
	if !decode(w, r, &input) {
		return
	}
	if len(input.Password) < 12 || len(input.Password) > 72 {
		fail(w, 400, "INVALID_PASSWORD", "新密码必须为12～72字节")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err == nil {
		err = s.store.ChangePassword(r.Context(), u.ID, id, r.Header.Get("Idempotency-Key"), input.Revision, hash, storage.Hash([]byte(input.Password)), input.CurrentPassword)
	}
	if err != nil {
		administrationError(w, err)
		return
	}
	s.revokeStreams(id)
	reply(w, 200, map[string]bool{"changed": true, "sessionsRevoked": true})
}
func (s *Server) role(r *http.Request, id string) (model.RoleTemplate, error) {
	for _, role := range model.BuiltinRoles() {
		if role.ID == id {
			return role, nil
		}
	}
	var role model.RoleTemplate
	_, err := s.store.Record(r.Context(), "role_template", id, &role)
	return role, err
}
func (s *Server) roles(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.admin(w, u) {
		return
	}
	items := model.BuiltinRoles()
	ids, err := s.store.RecordIDs(r.Context(), "role_template")
	if err != nil {
		administrationError(w, err)
		return
	}
	for _, id := range ids {
		role, err := s.role(r, id)
		if err != nil {
			administrationError(w, err)
			return
		}
		items = append(items, role)
	}
	reply(w, 200, map[string]any{"items": items, "semantics": "templates-expand-to-explicit-resource-grants"})
}
func (s *Server) putRole(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.admin(w, u) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var input model.RoleTemplate
	if !decode(w, r, &input) {
		return
	}
	input.ID = r.PathValue("id")
	for _, builtin := range model.BuiltinRoles() {
		if input.ID == builtin.ID {
			fail(w, 409, "BUILTIN_ROLE", "内置模板不可修改；可另存为自定义模板")
			return
		}
	}
	if input.ID == "" || len(input.ID) > 64 || len(input.Name) < 1 || len(input.Name) > 80 || len(input.Actions) < 1 || len(input.Actions) > len(actions) {
		fail(w, 400, "INVALID_ROLE", "角色模板字段无效")
		return
	}
	input.Builtin = false
	seen := map[string]bool{}
	for _, action := range input.Actions {
		if !actions[action] || seen[action] {
			fail(w, 400, "INVALID_ACTION", "权限动作未知或重复")
			return
		}
		if action == "instance.create" || action == "host.manage" || action == "node.read" {
			input.NodeOnly = true
		}
		seen[action] = true
	}
	data, err := s.store.MetadataMutation(r.Context(), u.ID, r.Header.Get("Idempotency-Key"), "role.update", input, func(tx *sql.Tx) (any, error) {
		var revision int64
		err := tx.QueryRowContext(r.Context(), "SELECT revision FROM records WHERE namespace='role_template' AND id=?", input.ID).Scan(&revision)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if revision != input.Revision {
			return nil, storage.ErrConflict
		}
		if revision == 0 {
			var count int
			if err := tx.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM records WHERE namespace='role_template'").Scan(&count); err != nil {
				return nil, err
			}
			if count >= 100 {
				return nil, errors.New("custom role template quota exceeded")
			}
		}
		input.Revision++
		b, err := json.Marshal(input)
		if err != nil {
			return nil, err
		}
		_, err = tx.ExecContext(r.Context(), "INSERT INTO records(namespace,id,revision,document) VALUES('role_template',?,?,?) ON CONFLICT(namespace,id) DO UPDATE SET revision=excluded.revision,document=excluded.document", input.ID, input.Revision, b)
		return map[string]any{"role": input}, err
	})
	if err != nil {
		administrationError(w, err)
		return
	}
	reply(w, 200, data)
}
func (s *Server) applyRole(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.admin(w, u) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var input struct {
		UserID   string            `json:"userId"`
		Resource model.ResourceRef `json:"resource"`
		Revision int64             `json:"revision"`
		Revoke   bool              `json:"revoke"`
	}
	if !decode(w, r, &input) {
		return
	}
	role, err := s.role(r, r.PathValue("id"))
	if err != nil {
		administrationError(w, err)
		return
	}
	if role.Revision != input.Revision {
		administrationError(w, storage.ErrConflict)
		return
	}
	if (input.Resource.Kind != "node" && input.Resource.Kind != "instance") || (role.NodeOnly && input.Resource.Kind != "node") {
		fail(w, 400, "INVALID_SCOPE", "此角色模板不适用于指定资源范围")
		return
	}
	if err = s.store.ApplyRole(r.Context(), u.ID, r.Header.Get("Idempotency-Key"), input.UserID, input.Resource, role, input.Revoke); err != nil {
		administrationError(w, err)
		return
	}
	s.revokeStreams(input.UserID)
	reply(w, 200, map[string]any{"applied": true, "role": role})
}

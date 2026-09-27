package master

import (
	"encoding/json"
	"net/http"
	"strconv"

	"blora.dev/panel/internal/backup"
	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/storage"
)

func (s *Server) registerBackups() {
	s.mux.HandleFunc("GET /api/v1/instances/{id}/backups", s.auth(s.backupList))
	s.mux.HandleFunc("POST /api/v1/instances/{id}/backups", s.auth(s.backupCreate))
	s.mux.HandleFunc("POST /api/v1/instances/{id}/backups/prune", s.auth(s.backupPrune))
	s.mux.HandleFunc("POST /api/v1/instances/{id}/restore-plans", s.auth(s.backupPlan))
	s.mux.HandleFunc("GET /api/v1/instances/{id}/restore-plans/{planId}", s.auth(s.backupPlanGet))
	s.mux.HandleFunc("POST /api/v1/instances/{id}/restores", s.auth(s.backupRestore))
}
func (s *Server) backupInstance(w http.ResponseWriter, r *http.Request, u model.User, action string) (model.Instance, bool) {
	i, ok := s.fileInstance(w, r, u, action)
	if !ok {
		return i, false
	}
	permission := "file.read"
	if action == "backup.restore" {
		permission = "file.write"
	}
	return i, s.allowed(w, r, u, instanceRef(i), permission)
}
func (s *Server) backupCall(w http.ResponseWriter, r *http.Request, u model.User, i model.Instance, method string, args any, status int) {
	b, _ := json.Marshal(args)
	var out json.RawMessage
	if err := s.nodeCall(r.Context(), i.NodeID, protocol.ChannelBulk, bridge.Request{Method: method, ActorID: u.ID, Resource: instanceRef(i), Config: &i.Config, Args: b}, &out); err != nil {
		bridgeError(w, err)
		return
	}
	reply(w, status, out)
}
func (s *Server) backupList(w http.ResponseWriter, r *http.Request, u model.User) {
	i, ok := s.backupInstance(w, r, u, "file.read")
	if !ok {
		return
	}
	offset := 0
	if raw := r.URL.Query().Get("offset"); raw != "" {
		var err error
		offset, err = strconv.Atoi(raw)
		if err != nil || offset < 0 {
			fail(w, 400, "INVALID_PAGE", "备份分页偏移无效")
			return
		}
	}
	s.backupCall(w, r, u, i, "backup.list", map[string]any{"offset": offset, "limit": 50}, 200)
}
func backupOperationID(u model.User, r *http.Request, kind string) string {
	return storage.Hash([]byte(u.ID + ":" + kind + ":" + r.Header.Get("Idempotency-Key")))[:32]
}
func (s *Server) acceptBackup(w http.ResponseWriter, r *http.Request, u model.User, i model.Instance, action string, p backup.TaskPayload) {
	p.Config = i.Config
	b, _ := json.Marshal(p)
	if len(b) > 20<<10 {
		fail(w, 413, "CONFIG_LIMIT", "备份任务配置超过20 KiB")
		return
	}
	t, _, err := s.store.Accept(r.Context(), model.Task{ActorID: u.ID, RequestID: r.Header.Get("Idempotency-Key"), Resource: instanceRef(i), Action: action, Payload: b})
	if err != nil {
		taskError(w, err)
		return
	}
	reply(w, 202, map[string]any{"task": s.presentAcceptedTask(r.Context(), t)})
}
func (s *Server) backupCreate(w http.ResponseWriter, r *http.Request, u model.User) {
	i, ok := s.backupInstance(w, r, u, "backup.create")
	if !ok {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var p backup.CreateSpec
	if !decode(w, r, &p) {
		return
	}
	if p.ID != "" || p.OwnerID != "" || p.Source.ID != "" {
		fail(w, 400, "INVALID_SCOPE", "身份与资源由服务端绑定")
		return
	}
	if !filePath(w, p.Path, true) || !validFileHash(p.Version) {
		fail(w, 400, "INVALID_VERSION", "需要当前源版本")
		return
	}
	p.ID, p.OwnerID, p.Source = backupOperationID(u, r, "create"), u.ID, instanceRef(i)
	s.acceptBackup(w, r, u, i, "backup.create", backup.TaskPayload{Create: &p})
}
func (s *Server) backupPrune(w http.ResponseWriter, r *http.Request, u model.User) {
	i, ok := s.backupInstance(w, r, u, "backup.create")
	if !ok {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var p backup.Retention
	if !decode(w, r, &p) {
		return
	}
	if p.OwnerID != "" || p.Source.ID != "" {
		fail(w, 400, "INVALID_SCOPE", "身份与资源由服务端绑定")
		return
	}
	p.OwnerID, p.Source = u.ID, instanceRef(i)
	s.acceptBackup(w, r, u, i, "backup.prune", backup.TaskPayload{Retention: &p})
}
func (s *Server) backupPlan(w http.ResponseWriter, r *http.Request, u model.User) {
	i, ok := s.backupInstance(w, r, u, "backup.restore")
	if !ok {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var p backup.RestoreRequest
	if !decode(w, r, &p) {
		return
	}
	if p.ID != "" || p.OwnerID != "" || p.Target.ID != "" {
		fail(w, 400, "INVALID_SCOPE", "需要请求键，身份与资源由服务端绑定")
		return
	}
	p.ID, p.OwnerID, p.Target = backupOperationID(u, r, "plan"), u.ID, instanceRef(i)
	s.backupCall(w, r, u, i, "backup.plan", map[string]any{"request": p, "limit": 100}, 201)
}
func (s *Server) backupPlanGet(w http.ResponseWriter, r *http.Request, u model.User) {
	i, ok := s.backupInstance(w, r, u, "backup.restore")
	if !ok {
		return
	}
	offset := 0
	if raw := r.URL.Query().Get("offset"); raw != "" {
		var err error
		offset, err = strconv.Atoi(raw)
		if err != nil || offset < 0 {
			fail(w, 400, "INVALID_PAGE", "恢复计划分页偏移无效")
			return
		}
	}
	s.backupCall(w, r, u, i, "backup.plan.get", map[string]any{"id": r.PathValue("planId"), "offset": offset, "limit": 100}, 200)
}
func (s *Server) backupRestore(w http.ResponseWriter, r *http.Request, u model.User) {
	i, ok := s.backupInstance(w, r, u, "backup.restore")
	if !ok {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var p backup.RestoreSpec
	if !decode(w, r, &p) {
		return
	}
	if p.ID != "" || p.OwnerID != "" {
		fail(w, 400, "INVALID_SCOPE", "身份由服务端绑定")
		return
	}
	p.ID, p.OwnerID = backupOperationID(u, r, "restore"), u.ID
	s.acceptBackup(w, r, u, i, "backup.restore", backup.TaskPayload{Restore: &p})
}

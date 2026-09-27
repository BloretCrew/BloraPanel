package master

import (
	"encoding/json"
	"net/http"
	"strconv"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/containers"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
)

func (s *Server) registerComposeSaves() {
	s.mux.HandleFunc("POST /api/v1/nodes/{id}/docker/project-saves", s.auth(s.composePrepare))
	s.mux.HandleFunc("GET /api/v1/nodes/{id}/docker/project-saves/{saveId}", s.auth(s.composeStatus))
	s.mux.HandleFunc("PUT /api/v1/nodes/{id}/docker/project-saves/{saveId}/chunks", s.auth(s.composeChunk))
	s.mux.HandleFunc("POST /api/v1/nodes/{id}/docker/project-saves/{saveId}/commit", s.auth(s.composeCommit))
	s.mux.HandleFunc("GET /api/v1/nodes/{id}/docker/projects/{projectId}/content", s.auth(s.composeContent))
}

func (s *Server) composeAuthority(w http.ResponseWriter, r *http.Request, u model.User) bool {
	ref := nodeRef(r.PathValue("id"))
	if _, _, err := s.store.Node(r.Context(), ref.ID); err != nil {
		fail(w, 404, "NOT_FOUND", "节点不存在")
		return false
	}
	if !s.allowed(w, r, u, ref, "host.manage") {
		return false
	}
	return true
}

func (s *Server) composeCall(w http.ResponseWriter, r *http.Request, u model.User, method string, input model.ComposeSaveRequest, out any) bool {
	if !s.composeAuthority(w, r, u) {
		return false
	}
	ref := nodeRef(r.PathValue("id"))
	b, _ := json.Marshal(input)
	if len(b) > 96<<10 {
		fail(w, 413, "CHUNK_LIMIT", "项目保存分片超出上限")
		return false
	}
	if err := s.nodeCall(r.Context(), ref.ID, protocol.ChannelBulk, bridge.Request{Method: method, ActorID: u.ID, Resource: ref, Args: b}, out); err != nil {
		bridgeError(w, err)
		return false
	}
	return true
}
func (s *Server) composePrepare(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.composeAuthority(w, r, u) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var input model.ComposeSaveRequest
	if !decode(w, r, &input) {
		return
	}
	input.Data = nil
	var result containers.ProjectSave
	if !s.composeCall(w, r, u, "container.project.prepare", input, &result) {
		return
	}
	reply(w, 201, result)
}
func (s *Server) composeStatus(w http.ResponseWriter, r *http.Request, u model.User) {
	var result containers.ProjectSave
	if !s.composeCall(w, r, u, "container.project.status", model.ComposeSaveRequest{SaveID: r.PathValue("saveId")}, &result) {
		return
	}
	reply(w, 200, result)
}
func (s *Server) composeChunk(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.composeAuthority(w, r, u) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var input model.ComposeSaveRequest
	if !decode(w, r, &input) {
		return
	}
	input.SaveID = r.PathValue("saveId")
	if len(input.Data) > containers.ProjectChunkBytes {
		fail(w, 413, "CHUNK_LIMIT", "项目保存分片不能超过64 KiB")
		return
	}
	var result containers.ProjectSave
	if !s.composeCall(w, r, u, "container.project.chunk", input, &result) {
		return
	}
	reply(w, 200, result)
}
func (s *Server) composeCommit(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.composeAuthority(w, r, u) {
		return
	}
	ref := nodeRef(r.PathValue("id"))
	if !requireRequestID(w, r) {
		return
	}
	input := model.ComposeSaveRequest{SaveID: r.PathValue("saveId")}
	// Source ownership, checksum and checkpoint are rechecked by the node
	// worker; the control payload contains no Compose source text.
	b, _ := json.Marshal(input)
	task, _, err := s.store.Accept(r.Context(), model.Task{ActorID: u.ID, RequestID: r.Header.Get("Idempotency-Key"), Resource: ref, Action: "compose.save", Payload: b})
	if err != nil {
		taskError(w, err)
		return
	}
	task = s.presentAcceptedTask(r.Context(), task)
	reply(w, 202, map[string]any{"task": task})
}
func (s *Server) composeContent(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.composeAuthority(w, r, u) {
		return
	}
	q := r.URL.Query()
	revision, err := strconv.ParseUint(q.Get("revision"), 10, 64)
	if err != nil || revision == 0 {
		fail(w, 400, "INVALID_REVISION", "需要固定项目版本")
		return
	}
	offset := int64(0)
	if q.Get("offset") != "" {
		offset, err = strconv.ParseInt(q.Get("offset"), 10, 64)
	}
	if err != nil || offset < 0 {
		fail(w, 400, "INVALID_OFFSET", "无效读取位置")
		return
	}
	length := containers.ProjectChunkBytes
	if q.Get("length") != "" {
		length, err = strconv.Atoi(q.Get("length"))
	}
	if err != nil || length < 1 || length > containers.ProjectChunkBytes {
		fail(w, 400, "CHUNK_LIMIT", "读取长度需要1～65536字节")
		return
	}
	var result containers.ProjectChunk
	if !s.composeCall(w, r, u, "container.project.read", model.ComposeSaveRequest{ProjectID: r.PathValue("projectId"), Revision: revision, Offset: offset, Length: length}, &result) {
		return
	}
	reply(w, 200, result)
}

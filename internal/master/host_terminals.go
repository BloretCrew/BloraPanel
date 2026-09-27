package master

import (
	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/containers"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/terminal"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
)

func (s *Server) registerHostTerminals() {
	s.mux.HandleFunc("POST /api/v1/nodes/{id}/docker/containers/{containerId}/terminals", s.auth(s.hostTerminalCreate))
	s.mux.HandleFunc("GET /api/v1/nodes/{id}/terminals", s.auth(s.hostTerminalList))
}

var hostTerminalContainerID = regexp.MustCompile(`^[a-f0-9]{64}$`)

func (s *Server) hostTerminalCreate(w http.ResponseWriter, r *http.Request, u model.User) {
	ref := nodeRef(r.PathValue("id"))
	if _, _, err := s.store.Node(r.Context(), ref.ID); err != nil {
		fail(w, 404, "NOT_FOUND", "节点不存在")
		return
	}
	if !s.allowed(w, r, u, ref, "host.manage") {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var input struct {
		Cols      uint16 `json:"cols"`
		Rows      uint16 `json:"rows"`
		CreatedAt string `json:"createdAt"`
	}
	if !decode(w, r, &input) {
		return
	}
	id := r.PathValue("containerId")
	if !hostTerminalContainerID.MatchString(id) || input.Cols > 1000 || input.Rows > 1000 {
		fail(w, 400, "INVALID_REQUEST", "需要完整容器ID及有效终端尺寸")
		return
	}
	if input.Cols == 0 {
		input.Cols = 100
	}
	if input.Rows == 0 {
		input.Rows = 28
	}
	b, _ := json.Marshal(containers.Target{Kind: "container", ID: id, CreatedAt: input.CreatedAt})
	var target model.HostTerminalTarget
	if err := s.nodeCall(r.Context(), ref.ID, protocol.ChannelInteractive, bridge.Request{Method: "container.exec-target", ActorID: u.ID, Resource: ref, Args: b}, &target); err != nil {
		bridgeError(w, err)
		return
	}
	if target.ContainerID != id || target.CreatedAt == "" || target.StartedAt == "" {
		fail(w, 502, "RESOURCE_MISMATCH", "容器身份未确认")
		return
	}
	payload, _ := json.Marshal(model.TerminalTaskPayload{Host: &target, Cols: input.Cols, Rows: input.Rows})
	task, _, err := s.store.Accept(r.Context(), model.Task{ActorID: u.ID, RequestID: r.Header.Get("Idempotency-Key"), Resource: ref, Action: "terminal.create", Payload: payload})
	if err != nil {
		taskError(w, err)
		return
	}
	reply(w, 202, map[string]any{"task": s.presentAcceptedTask(r.Context(), task)})
}

func (s *Server) hostTerminalList(w http.ResponseWriter, r *http.Request, u model.User) {
	ref := nodeRef(r.PathValue("id"))
	if !s.allowed(w, r, u, ref, "host.manage") {
		return
	}
	var result struct {
		Items []terminal.Session `json:"items"`
	}
	if err := s.nodeCall(r.Context(), ref.ID, protocol.ChannelInteractive, bridge.Request{Method: "terminal.list", ActorID: u.ID, Resource: ref, Args: []byte(`{}`)}, &result); err != nil {
		bridgeError(w, err)
		return
	}
	for _, session := range result.Items {
		if session.Resource != ref || session.OwnerID != u.ID || session.Backend != "docker-host" {
			fail(w, 502, "RESOURCE_MISMATCH", "节点返回不同主体会话")
			return
		}
		var old terminal.Session
		rev, err := s.store.Record(r.Context(), "terminal", session.ID, &old)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			bridgeError(w, err)
			return
		}
		if _, err = s.store.PutRecord(r.Context(), "terminal", session.ID, rev, session); err != nil {
			bridgeError(w, err)
			return
		}
	}
	reply(w, 200, result)
}

package master

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/systeminfo"
)

func (s *Server) registerSystem() {
	s.mux.HandleFunc("GET /api/v1/nodes/{id}/system/capabilities", s.auth(s.systemCapabilities))
	s.mux.HandleFunc("GET /api/v1/nodes/{id}/system/services", s.auth(s.systemServices))
	s.mux.HandleFunc("GET /api/v1/nodes/{id}/system/firewall", s.auth(s.systemFirewall))
	s.mux.HandleFunc("GET /api/v1/nodes/{id}/system/tasks", s.auth(s.systemTasks))
	s.mux.HandleFunc("POST /api/v1/nodes/{id}/system/firewall/preview", s.auth(s.systemFirewallPreview))
	s.mux.HandleFunc("POST /api/v1/nodes/{id}/system/firewall/confirm", s.auth(s.systemFirewallConfirm))
	s.mux.HandleFunc("POST /api/v1/nodes/{id}/system/services/actions", s.auth(s.systemServiceAction))
	s.mux.HandleFunc("POST /api/v1/nodes/{id}/system/tasks/actions", s.auth(s.systemTaskAction))
	s.mux.HandleFunc("POST /api/v1/nodes/{id}/system/firewall/apply", s.auth(s.systemFirewallApply))
}
func (s *Server) systemFirewallPreview(w http.ResponseWriter, r *http.Request, u model.User) {
	ref := nodeRef(r.PathValue("id"))
	if !s.allowed(w, r, u, ref, "host.manage") {
		return
	}
	var in struct {
		Desired []string `json:"desired"`
	}
	if !decode(w, r, &in) {
		return
	}
	b, _ := json.Marshal(in)
	var out map[string]any
	if err := s.nodeCall(r.Context(), ref.ID, protocol.ChannelBulk, bridge.Request{Method: "system.firewall.preview", ActorID: u.ID, Resource: ref, Args: b}, &out); err != nil {
		bridgeError(w, err)
		return
	}
	reply(w, 200, out)
}
func (s *Server) systemTasks(w http.ResponseWriter, r *http.Request, u model.User) {
	ref := nodeRef(r.PathValue("id"))
	if !s.allowed(w, r, u, ref, "host.manage") {
		return
	}
	var out map[string]any
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 200 {
			fail(w, 400, "LIMIT", "计划任务数量限制为1～200")
			return
		}
		limit = value
	}
	after := r.URL.Query().Get("after")
	if after != "" && !systeminfo.ValidTaskName(after) {
		fail(w, 400, "INVALID_CURSOR", "计划任务分页标识无效")
		return
	}
	args, _ := json.Marshal(map[string]any{"limit": limit, "after": after})
	if err := s.nodeCall(r.Context(), ref.ID, protocol.ChannelBulk, bridge.Request{Method: "system.tasks", ActorID: u.ID, Resource: ref, Args: args}, &out); err != nil {
		bridgeError(w, err)
		return
	}
	reply(w, 200, out)
}
func (s *Server) systemServiceAction(w http.ResponseWriter, r *http.Request, u model.User) {
	ref := nodeRef(r.PathValue("id"))
	if !s.allowed(w, r, u, ref, "host.manage") {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var in struct {
		Name   string `json:"name"`
		Action string `json:"action"`
	}
	if !decode(w, r, &in) {
		return
	}
	if in.Action != "start" && in.Action != "stop" && in.Action != "restart" {
		fail(w, 400, "INVALID_ACTION", "仅支持 start、stop、restart")
		return
	}
	if !systeminfo.ValidServiceName(in.Name) {
		fail(w, 400, "INVALID_SERVICE", "服务标识无效")
		return
	}
	b, _ := json.Marshal(map[string]string{"name": in.Name})
	task, _, err := s.store.Accept(r.Context(), model.Task{ActorID: u.ID, RequestID: r.Header.Get("Idempotency-Key"), Resource: ref, Action: "system.service." + in.Action, Payload: b})
	if err != nil {
		taskError(w, err)
		return
	}
	reply(w, 202, map[string]any{"task": s.presentAcceptedTask(r.Context(), task)})
}
func (s *Server) systemTaskAction(w http.ResponseWriter, r *http.Request, u model.User) {
	ref := nodeRef(r.PathValue("id"))
	if !s.allowed(w, r, u, ref, "host.manage") {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var in struct {
		Name   string `json:"name"`
		Action string `json:"action"`
	}
	if !decode(w, r, &in) {
		return
	}
	if in.Action != "enable" && in.Action != "disable" {
		fail(w, 400, "INVALID_ACTION", "仅支持 enable、disable")
		return
	}
	if !systeminfo.ValidTaskName(in.Name) {
		fail(w, 400, "INVALID_TASK", "计划任务标识无效")
		return
	}
	b, _ := json.Marshal(map[string]string{"name": in.Name})
	task, _, err := s.store.Accept(r.Context(), model.Task{ActorID: u.ID, RequestID: r.Header.Get("Idempotency-Key"), Resource: ref, Action: "system.task." + in.Action, Payload: b})
	if err != nil {
		taskError(w, err)
		return
	}
	reply(w, 202, map[string]any{"task": s.presentAcceptedTask(r.Context(), task)})
}
func (s *Server) systemFirewallApply(w http.ResponseWriter, r *http.Request, u model.User) {
	ref := nodeRef(r.PathValue("id"))
	if !s.allowed(w, r, u, ref, "host.manage") {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var in struct {
		Desired      []string `json:"desired"`
		ConfirmUntil string   `json:"confirmUntil"`
		PlanHash     string   `json:"planHash"`
	}
	if !decode(w, r, &in) {
		return
	}
	if len(in.Desired) > 256 {
		fail(w, 400, "RULE_LIMIT", "防火墙规则数量不能超过256")
		return
	}
	if len(in.PlanHash) != 64 {
		fail(w, 400, "PREVIEW_REQUIRED", "需要有效防火墙预览标识")
		return
	}
	deadline, err := time.Parse(time.RFC3339, in.ConfirmUntil)
	if err != nil || deadline.Before(time.Now().UTC()) || deadline.After(time.Now().UTC().Add(60*time.Second)) {
		fail(w, 400, "CONFIRMATION_EXPIRED", "需要60秒内有效的确认期限")
		return
	}
	b, _ := json.Marshal(map[string]any{"desired": in.Desired, "confirmUntil": deadline.UTC(), "planHash": in.PlanHash})
	task, _, err := s.store.Accept(r.Context(), model.Task{ActorID: u.ID, RequestID: r.Header.Get("Idempotency-Key"), Resource: ref, Action: "system.firewall.apply", Payload: b})
	if err != nil {
		taskError(w, err)
		return
	}
	reply(w, 202, map[string]any{"task": s.presentAcceptedTask(r.Context(), task)})
}

// systemFirewallConfirm is the second phase of a firewall change. The task's
// actor is carried through to the Daemon so an administrator can rescue a
// change without allowing an arbitrary task ID to confirm another actor's
// lease. The node performs the durable lease check again before committing.
func (s *Server) systemFirewallConfirm(w http.ResponseWriter, r *http.Request, u model.User) {
	ref := nodeRef(r.PathValue("id"))
	if !s.allowed(w, r, u, ref, "host.manage") {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var in struct {
		TaskID string `json:"taskId"`
	}
	if !decode(w, r, &in) {
		return
	}
	t, err := s.store.Task(r.Context(), in.TaskID)
	if err != nil || t.Action != "system.firewall.apply" || t.Resource != ref {
		fail(w, 404, "FIREWALL_TASK_NOT_FOUND", "防火墙租约任务不存在或目标不匹配")
		return
	}
	if !u.Admin && t.ActorID != u.ID {
		fail(w, 403, "FORBIDDEN", "只能确认自己发起的防火墙租约")
		return
	}
	if t.State.Terminal() {
		// The confirmation request may have succeeded before the browser saw its
		// response. The Daemon persists the request key in the task result so a
		// retry after lease cleanup can replay the same success without touching
		// the host again.
		if t.State == model.Succeeded && t.Phase == "confirmed" {
			var result struct {
				ConfirmRequestID string `json:"confirmRequestId"`
			}
			if json.Unmarshal(t.Result, &result) == nil && result.ConfirmRequestID == r.Header.Get("Idempotency-Key") {
				reply(w, 200, map[string]any{"taskId": t.ID, "confirmed": true})
				return
			}
		}
		fail(w, 409, "FIREWALL_LEASE_CLOSED", "防火墙租约已结束")
		return
	}
	b, _ := json.Marshal(map[string]string{"taskId": t.ID, "taskActorId": t.ActorID, "requestId": r.Header.Get("Idempotency-Key")})
	var out map[string]any
	if err := s.nodeCall(r.Context(), ref.ID, protocol.ChannelBulk, bridge.Request{Method: "system.firewall.confirm", ActorID: u.ID, Resource: ref, Args: b}, &out); err != nil {
		bridgeError(w, err)
		return
	}
	reply(w, 200, out)
}
func (s *Server) systemFirewall(w http.ResponseWriter, r *http.Request, u model.User) {
	ref := nodeRef(r.PathValue("id"))
	if !s.allowed(w, r, u, ref, "host.manage") {
		return
	}
	var out map[string]any
	if err := s.nodeCall(r.Context(), ref.ID, protocol.ChannelBulk, bridge.Request{Method: "system.firewall", ActorID: u.ID, Resource: ref, Args: []byte(`{}`)}, &out); err != nil {
		bridgeError(w, err)
		return
	}
	reply(w, 200, out)
}
func (s *Server) systemServices(w http.ResponseWriter, r *http.Request, u model.User) {
	ref := nodeRef(r.PathValue("id"))
	if !s.allowed(w, r, u, ref, "host.manage") {
		return
	}
	limit := 50
	if q := r.URL.Query().Get("limit"); q != "" {
		if n, e := strconv.Atoi(q); e != nil || n < 1 || n > 200 {
			fail(w, 400, "LIMIT", "服务数量限制为1～200")
			return
		} else {
			limit = n
		}
	}
	after := r.URL.Query().Get("after")
	if after != "" && !systeminfo.ValidServiceName(after) {
		fail(w, 400, "INVALID_CURSOR", "服务分页标识无效")
		return
	}
	b, _ := json.Marshal(map[string]any{"limit": limit, "after": after})
	var out map[string]any
	if err := s.nodeCall(r.Context(), ref.ID, protocol.ChannelBulk, bridge.Request{Method: "system.services", ActorID: u.ID, Resource: ref, Args: b}, &out); err != nil {
		bridgeError(w, err)
		return
	}
	reply(w, 200, out)
}
func (s *Server) systemCapabilities(w http.ResponseWriter, r *http.Request, u model.User) {
	ref := nodeRef(r.PathValue("id"))
	if !s.allowed(w, r, u, ref, "host.manage") {
		return
	}
	var out map[string]any
	if err := s.nodeCall(r.Context(), ref.ID, protocol.ChannelBulk, bridge.Request{Method: "system.capabilities", ActorID: u.ID, Resource: ref, Args: []byte(`{}`)}, &out); err != nil {
		bridgeError(w, err)
		return
	}
	reply(w, 200, out)
}

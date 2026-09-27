package master

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/containers"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
)

func nodeRef(id string) model.ResourceRef { return model.ResourceRef{Kind: "node", ID: id, NodeID: id} }

func (s *Server) registerContainers() {
	s.registerComposeSaves()
	s.mux.HandleFunc("GET /api/v1/nodes/{id}/docker/containers/{containerId}/logs", s.auth(s.containerLogs))
	s.mux.HandleFunc("GET /api/v1/nodes/{id}/docker/containers/{containerId}/logs/history", s.auth(s.containerLogHistory))
	s.mux.HandleFunc("POST /api/v1/nodes/{id}/docker/query", s.auth(s.containerQuery))
	s.mux.HandleFunc("POST /api/v1/nodes/{id}/docker/actions", s.auth(s.containerAction))
}

func (s *Server) containerLogHistory(w http.ResponseWriter, r *http.Request, u model.User) {
	ref := nodeRef(r.PathValue("id"))
	if !s.allowed(w, r, u, ref, "host.manage") {
		return
	}
	id := r.PathValue("containerId")
	if len(id) != 64 {
		fail(w, 400, "INVALID_TARGET", "需要容器完整ID")
		return
	}
	limit := 20
	if q := r.URL.Query().Get("limit"); q != "" {
		n, e := strconv.Atoi(q)
		if e != nil || n < 1 || n > 100 {
			fail(w, 400, "LIMIT", "历史窗口限制为1～100")
			return
		}
		limit = n
	}
	b, _ := json.Marshal(map[string]any{"containerId": id, "limit": limit})
	var out map[string]any
	if err := s.nodeCall(r.Context(), ref.ID, protocol.ChannelBulk, bridge.Request{Method: "container.logs.history", ActorID: u.ID, Resource: ref, Args: b}, &out); err != nil {
		bridgeError(w, err)
		return
	}
	reply(w, 200, out)
}

func (s *Server) containerQuery(w http.ResponseWriter, r *http.Request, u model.User) {
	ref := nodeRef(r.PathValue("id"))
	if _, _, err := s.store.Node(r.Context(), ref.ID); err != nil {
		fail(w, 404, "NOT_FOUND", "节点不存在")
		return
	}
	if !s.allowed(w, r, u, ref, "host.manage") {
		return
	}
	var query containers.Query
	if !decode(w, r, &query) {
		return
	}
	if query.Kind == "operation-output" && query.OutputOffset < 0 {
		fail(w, 400, "INVALID_PAGE", "命令输出页码无效")
		return
	}
	if query.Limit < 1 {
		query.Limit = 50
	}
	if query.Limit > 100 {
		query.Limit = 100
	}
	b, _ := json.Marshal(query)
	var snapshot containers.Snapshot
	if err := s.nodeCall(r.Context(), ref.ID, protocol.ChannelBulk, bridge.Request{Method: "container.query", ActorID: u.ID, Resource: ref, Args: b}, &snapshot); err != nil {
		bridgeError(w, err)
		return
	}
	reply(w, 200, snapshot)
}

func (s *Server) containerAction(w http.ResponseWriter, r *http.Request, u model.User) {
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
	var op containers.Operation
	if !decode(w, r, &op) {
		return
	}
	if !model.IsContainerAction(op.Action) || op.Action == "compose.save" || op.TaskID != "" {
		fail(w, 400, "INVALID_ACTION", "不支持此Docker操作或由客户端指定任务ID")
		return
	}
	b, _ := json.Marshal(op)
	if len(b) > 20<<10 {
		fail(w, 413, "CONFIG_LIMIT", "容器操作配置不能超过20 KiB；Compose正文使用项目保存接口")
		return
	}
	task, _, err := s.store.Accept(r.Context(), model.Task{ActorID: u.ID, RequestID: r.Header.Get("Idempotency-Key"), Resource: ref, Action: op.Action, Payload: b})
	if err != nil {
		taskError(w, err)
		return
	}
	task = s.presentAcceptedTask(r.Context(), task)
	reply(w, 202, map[string]any{"task": task})
}

func (s *Server) presentAcceptedTask(ctx context.Context, t model.Task) model.Task {
	if t.State.Terminal() || t.State == model.WaitingNode || t.State == model.WaitingClient {
		return t
	}
	s.mu.Lock()
	online := s.peers[t.Resource.NodeID] != nil
	s.mu.Unlock()
	if !online {
		if updated, err := s.store.WaitingNode(ctx, t.ID, t.Revision); err == nil {
			return updated
		}
	}
	return t
}

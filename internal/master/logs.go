package master

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/runlog"
	"blora.dev/panel/internal/terminal"
	"github.com/coder/websocket"
)

func (s *Server) registerLogs() {
	s.mux.HandleFunc("GET /api/v1/instances/{id}/logs", s.auth(s.logs))
	s.mux.HandleFunc("GET /api/v1/instances/{id}/logs/{runId}/stream", s.auth(s.logStream))
	s.mux.HandleFunc("POST /api/v1/instances/{id}/logs/{runId}/input", s.auth(s.consoleInput))
}

func (s *Server) consoleInput(w http.ResponseWriter, r *http.Request, u model.User) {
	i, err := s.store.Instance(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, 404, "NOT_FOUND", "实例不存在")
		return
	}
	if !s.allowed(w, r, u, instanceRef(i), "terminal.input") {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var input model.ConsoleInput
	if !decode(w, r, &input) {
		return
	}
	input.RunID = r.PathValue("runId")
	if len(input.Data) == 0 || len(input.Data) > 8<<10 {
		fail(w, 400, "INPUT_LIMIT", "独立命令限1～8192字节")
		return
	}
	if i.RunID != input.RunID {
		fail(w, 409, "RUN_CHANGED", "实例运行代次已变更，未发送命令")
		return
	}
	b, _ := json.Marshal(input)
	task, _, err := s.store.Accept(r.Context(), model.Task{ActorID: u.ID, RequestID: r.Header.Get("Idempotency-Key"), Resource: instanceRef(i), Action: "console.input", Payload: b})
	if err != nil {
		taskError(w, err)
		return
	}
	reply(w, 202, map[string]any{"task": task})
}
func (s *Server) logs(w http.ResponseWriter, r *http.Request, u model.User) {
	i, err := s.store.Instance(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, 404, "NOT_FOUND", "实例不存在")
		return
	}
	if !s.allowed(w, r, u, instanceRef(i), "instance.read") {
		return
	}
	var result struct {
		Items []model.RunLog `json:"items"`
	}
	err = s.nodeCall(r.Context(), i.NodeID, protocol.ChannelInteractive, bridge.Request{Method: "log.list", ActorID: u.ID, Resource: instanceRef(i), Args: []byte(`{}`)}, &result)
	if err != nil {
		bridgeError(w, err)
		return
	}
	for _, info := range result.Items {
		if info.Resource != instanceRef(i) {
			fail(w, 502, "RESOURCE_MISMATCH", "节点返回日志资源不匹配")
			return
		}
	}
	reply(w, 200, result)
}
func (s *Server) logStream(w http.ResponseWriter, r *http.Request, u model.User) {
	if r.TLS == nil || r.Header.Get("Origin") != s.requestOrigin(r) {
		fail(w, 403, "ORIGIN_DENIED", "日志流需要可信TLS来源")
		return
	}
	i, err := s.store.Instance(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, 404, "NOT_FOUND", "实例不存在")
		return
	}
	if !s.allowed(w, r, u, instanceRef(i), "instance.read") {
		return
	}
	after := uint64(0)
	if q := r.URL.Query().Get("sequence"); q != "" {
		after, err = strconv.ParseUint(q, 10, 64)
		if err != nil {
			fail(w, 400, "INVALID_CURSOR", "无效日志游标")
			return
		}
	}
	runID := r.PathValue("runId")
	call := func(ctx context.Context, method string, args map[string]any, out any) error {
		current, err := s.store.User(ctx, u.ID)
		if err != nil || !s.store.Allowed(ctx, current, instanceRef(i), "instance.read") {
			return terminal.ErrForbidden
		}
		if args == nil {
			args = map[string]any{}
		}
		args["runId"] = runID
		b, _ := json.Marshal(args)
		return s.nodeCall(ctx, i.NodeID, protocol.ChannelInteractive, bridge.Request{Method: method, ActorID: u.ID, Resource: instanceRef(i), Args: b}, out)
	}
	var status runlog.Status
	if err := call(r.Context(), "log.status", nil, &status); err != nil {
		bridgeError(w, err)
		return
	}
	ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	defer ws.CloseNow()
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	key := model.ID()
	s.mu.Lock()
	if len(s.userStreams[u.ID]) >= 8 {
		s.mu.Unlock()
		return
	}
	if s.userStreams[u.ID] == nil {
		s.userStreams[u.ID] = map[string]context.CancelFunc{}
	}
	s.userStreams[u.ID][key] = cancel
	s.mu.Unlock()
	defer func() { s.mu.Lock(); delete(s.userStreams[u.ID], key); s.mu.Unlock() }()
	c, err := protocol.NewConn(ctx, ws, protocol.Options{Generation: 1, Channel: protocol.ChannelInteractive, InitialStreams: []string{runID}})
	if err != nil {
		return
	}
	defer c.Close()
	go func() {
		defer cancel()
		for {
			e, err := c.Read(ctx)
			if err != nil {
				return
			}
			if e.Type != protocol.TypePing {
				return
			}
			if err := c.Send(ctx, protocol.Envelope{Type: protocol.TypePong}); err != nil {
				return
			}
		}
	}()
	send := func(kind protocol.Type, payload any) error {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		return c.Send(ctx, protocol.Envelope{Type: kind, StreamID: runID, Payload: b})
	}
	if err := send(protocol.TypeOpenAck, map[string]any{"runId": runID, "status": status, "inputAvailable": status.InputAvailable && i.RunID == runID && s.store.Allowed(ctx, u, instanceRef(i), "terminal.input")}); err != nil {
		return
	}
	sent := uint64(0)
	for {
		var batch terminal.Batch
		if err := call(ctx, "log.read", map[string]any{"after": after, "maxBytes": 64 << 10}, &batch); err != nil {
			_ = send(protocol.TypeError, model.APIError{Code: "LOG_UNAVAILABLE", Message: err.Error()})
			return
		}
		if batch.Gap {
			if err := send(protocol.TypeReset, map[string]any{"code": "OUTPUT_GAP", "earliest": batch.Earliest, "latest": batch.Latest, "message": "较早日志已超出保留范围"}); err != nil {
				return
			}
			after = batch.Earliest - 1
			continue
		}
		for _, event := range batch.Events {
			b, err := json.Marshal(event)
			if err != nil {
				return
			}
			sent += uint64(len(b))
			if err := c.Send(ctx, protocol.Envelope{Type: protocol.TypeData, StreamID: runID, Sequence: sent, Payload: b}); err != nil {
				return
			}
			after = event.Sequence
		}
		deadline := time.NewTimer(45 * time.Second)
		for c.Flow().Stats().SentUnacknowledged != 0 {
			select {
			case <-ctx.Done():
				deadline.Stop()
				return
			case <-c.Done():
				deadline.Stop()
				return
			case <-deadline.C:
				// TypeError is queued with priority, but closing the websocket in the
				// same scheduling turn can race the client's reader and turn a
				// diagnostic into a bare EOF. Give the writer a short, bounded grace
				// period after a successful send so the error frame is observable.
				if err := send(protocol.TypeError, model.APIError{Code: "SLOW_CONSUMER", Message: "日志消费超过期限，可从最后确认的游标重新读取"}); err == nil {
					select {
					case <-ctx.Done():
					case <-time.After(100 * time.Millisecond):
					}
				}
				return
			case <-time.After(10 * time.Millisecond):
			}
		}
		deadline.Stop()
		if err := call(ctx, "log.status", nil, &status); err != nil {
			_ = send(protocol.TypeError, model.APIError{Code: "LOG_UNAVAILABLE", Message: err.Error()})
			return
		}
		if status.Phase == "complete" || status.Phase == "failed" || status.Phase == "invalid" {
			if after >= status.Latest {
				_ = send(protocol.TypeClose, status)
				return
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-c.Done():
			return
		case <-time.After(100 * time.Millisecond):
		}
	}
}

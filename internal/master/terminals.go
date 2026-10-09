package master

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/storage"
	"blora.dev/panel/internal/terminal"
	"github.com/coder/websocket"
)

type terminalView struct {
	cancel context.CancelFunc
	done   chan struct{}
}

// Refresh may open a replacement socket before its previous handler detaches.
// Serialize those attachments so the old handler cannot release the new lease.
func (s *Server) claimTerminalView(ctx context.Context, cancel context.CancelFunc, key string) (func(), error) {
	v := &terminalView{cancel: cancel, done: make(chan struct{})}
	for {
		s.mu.Lock()
		if s.terminalViews == nil {
			s.terminalViews = map[string]*terminalView{}
		}
		previous := s.terminalViews[key]
		if previous == nil {
			s.terminalViews[key] = v
			s.mu.Unlock()
			return func() {
				s.mu.Lock()
				if s.terminalViews[key] == v {
					delete(s.terminalViews, key)
				}
				close(v.done)
				s.mu.Unlock()
			}, nil
		}
		s.mu.Unlock()
		previous.cancel()
		select {
		case <-previous.done:
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(4 * time.Second):
			return nil, terminal.ErrLease
		}
	}
}

func (s *Server) registerTerminals() {
	s.registerHostTerminals()
	s.mux.HandleFunc("GET /api/v1/instances/{id}/terminals", s.auth(s.terminalList))
	s.mux.HandleFunc("POST /api/v1/instances/{id}/terminals", s.auth(s.terminalCreate))
	s.mux.HandleFunc("POST /api/v1/terminals/{id}/close", s.auth(s.terminalClose))
	s.mux.HandleFunc("GET /api/v1/terminals/{id}/stream", s.auth(s.terminalStream))
}
func (s *Server) terminalPermission(ctx context.Context, u model.User, ref model.ResourceRef, backend, action string) bool {
	if ref.Kind == "node" {
		if backend != "docker-host" || ref.ID != ref.NodeID {
			return false
		}
		action = "host.manage"
	}
	current, err := s.store.User(ctx, u.ID)
	if err != nil || !s.store.Allowed(ctx, current, ref, action) {
		return false
	}
	return backend != "native" || s.store.Allowed(ctx, current, model.ResourceRef{Kind: "node", ID: ref.NodeID}, "host.manage")
}
func (s *Server) terminalList(w http.ResponseWriter, r *http.Request, u model.User) {
	i, err := s.store.Instance(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, 404, "NOT_FOUND", "实例不存在")
		return
	}
	backend := "native"
	if i.Config.Mode == "container" {
		backend = "docker"
	}
	if !s.terminalPermission(r.Context(), u, instanceRef(i), backend, "terminal.read") {
		fail(w, 403, "FORBIDDEN", "没有此终端范围权限；原生Shell额外需要host.manage")
		return
	}
	var result struct {
		Items []terminal.Session `json:"items"`
	}
	err = s.nodeCall(r.Context(), i.NodeID, protocol.ChannelInteractive, bridge.Request{Method: "terminal.list", ActorID: u.ID, Resource: instanceRef(i), Args: []byte(`{}`)}, &result)
	if err != nil {
		bridgeError(w, err)
		return
	}
	for _, session := range result.Items {
		if session.Resource != instanceRef(i) {
			fail(w, 502, "RESOURCE_MISMATCH", "节点返回了不同资源")
			return
		}
		var old terminal.Session
		rev, err := s.store.Record(r.Context(), "terminal", session.ID, &old)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			bridgeError(w, err)
			return
		}
		if _, err := s.store.PutRecord(r.Context(), "terminal", session.ID, rev, session); err != nil {
			bridgeError(w, err)
			return
		}
	}
	reply(w, 200, result)
}
func (s *Server) terminalCreate(w http.ResponseWriter, r *http.Request, u model.User) {
	i, err := s.store.Instance(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, 404, "NOT_FOUND", "实例不存在")
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var input struct {
		Cols uint16 `json:"cols"`
		Rows uint16 `json:"rows"`
	}
	if !decode(w, r, &input) {
		return
	}
	backend := "native"
	if i.Config.Mode == "container" {
		backend = "docker"
	}
	if !s.terminalPermission(r.Context(), u, instanceRef(i), backend, "terminal.input") {
		fail(w, 403, "FORBIDDEN", "原生Shell需要host.manage；普通实例终端需要隔离运行后端")
		return
	}
	if input.Cols == 0 {
		input.Cols = 100
	}
	if input.Rows == 0 {
		input.Rows = 28
	}
	if input.Cols > 1000 || input.Rows > 1000 {
		fail(w, 400, "INVALID_SIZE", "终端尺寸超过范围")
		return
	}
	payload, _ := json.Marshal(model.TerminalTaskPayload{Config: i.Config, Cols: input.Cols, Rows: input.Rows})
	task, _, err := s.store.Accept(r.Context(), model.Task{ActorID: u.ID, RequestID: r.Header.Get("Idempotency-Key"), Resource: instanceRef(i), Action: "terminal.create", Payload: payload})
	if err != nil {
		taskError(w, err)
		return
	}
	reply(w, 202, map[string]any{"task": task})
}
func taskError(w http.ResponseWriter, err error) {
	status := 400
	code := "INVALID_REQUEST"
	if errors.Is(err, storage.ErrRequestMismatch) {
		status = 409
		code = "IDEMPOTENCY_CONFLICT"
	}
	if errors.Is(err, storage.ErrQueueLimit) {
		status = 429
		code = "QUEUE_LIMIT"
	}
	fail(w, status, code, err.Error())
}
func (s *Server) terminalClose(w http.ResponseWriter, r *http.Request, u model.User) {
	var session terminal.Session
	if _, err := s.store.Record(r.Context(), "terminal", r.PathValue("id"), &session); err != nil {
		fail(w, 404, "NOT_FOUND", "终端会话不存在")
		return
	}
	if !s.terminalPermission(r.Context(), u, session.Resource, session.Backend, "terminal.input") {
		fail(w, 403, "FORBIDDEN", "没有结束会话权限")
		return
	}
	if session.Backend == "docker-host" && session.OwnerID != u.ID {
		fail(w, 403, "FORBIDDEN", "会话属于另一用户")
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	p := model.TerminalTaskPayload{SessionID: session.ID}
	if session.Resource.Kind == "instance" {
		i, err := s.store.Instance(r.Context(), session.Resource.ID)
		if err != nil {
			fail(w, 404, "NOT_FOUND", "实例不存在")
			return
		}
		p.Config = i.Config
	}
	payload, _ := json.Marshal(p)
	task, _, err := s.store.Accept(r.Context(), model.Task{ActorID: u.ID, RequestID: r.Header.Get("Idempotency-Key"), Resource: session.Resource, Action: "terminal.close", Payload: payload})
	if err != nil {
		taskError(w, err)
		return
	}
	reply(w, 202, map[string]any{"task": task})
}

func (s *Server) terminalStream(w http.ResponseWriter, r *http.Request, u model.User) {
	if !managementTransport(r) || r.Header.Get("Origin") != s.requestOrigin(r) {
		fail(w, 403, "ORIGIN_DENIED", "终端WSS需要可信TLS来源")
		return
	}
	var session terminal.Session
	if _, err := s.store.Record(r.Context(), "terminal", r.PathValue("id"), &session); err != nil {
		fail(w, 404, "NOT_FOUND", "终端会话不存在")
		return
	}
	if !s.terminalPermission(r.Context(), u, session.Resource, session.Backend, "terminal.read") {
		fail(w, 403, "FORBIDDEN", "没有查看终端权限")
		return
	}
	if session.Backend == "docker-host" && session.OwnerID != u.ID {
		fail(w, 403, "FORBIDDEN", "会话属于另一用户")
		return
	}
	view := r.URL.Query().Get("viewId")
	if view == "" || len(view) > 128 {
		fail(w, 400, "INVALID_VIEW", "需要独立viewTabId")
		return
	}
	after, err := strconv.ParseUint(r.URL.Query().Get("sequence"), 10, 64)
	if err != nil && r.URL.Query().Get("sequence") != "" {
		fail(w, 400, "INVALID_CURSOR", "无效输出游标")
		return
	}
	encoding := r.URL.Query().Get("encoding")
	if encoding != "" && encoding != "events-v1" {
		fail(w, 400, "INVALID_ENCODING", "不支持的终端输出编码")
		return
	}
	ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	defer ws.CloseNow()
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	releaseView, err := s.claimTerminalView(ctx, cancel, session.ID+":"+u.ID+":"+view)
	if err != nil {
		return
	}
	defer releaseView()
	streamKey := model.ID()
	s.mu.Lock()
	if len(s.userStreams[u.ID]) >= 8 {
		s.mu.Unlock()
		return
	}
	if s.userStreams[u.ID] == nil {
		s.userStreams[u.ID] = map[string]context.CancelFunc{}
	}
	s.userStreams[u.ID][streamKey] = cancel
	s.mu.Unlock()
	defer func() { s.mu.Lock(); delete(s.userStreams[u.ID], streamKey); s.mu.Unlock() }()
	c, err := protocol.NewConn(ctx, ws, protocol.Options{Generation: 1, Channel: protocol.ChannelInteractive, InitialStreams: []string{session.ID}})
	if err != nil {
		return
	}
	defer c.Close()
	call := func(ctx context.Context, method string, args map[string]any, out any) error {
		if args == nil {
			args = map[string]any{}
		}
		args["sessionId"] = session.ID
		args["viewId"] = view
		b, _ := json.Marshal(args)
		return s.nodeCall(ctx, session.Resource.NodeID, protocol.ChannelInteractive, bridge.Request{Method: method, ActorID: u.ID, Resource: session.Resource, Args: b}, out)
	}
	sendError := func(code, message string) {
		b, _ := json.Marshal(model.APIError{Code: code, Message: message})
		_ = c.Send(ctx, protocol.Envelope{Type: protocol.TypeError, StreamID: session.ID, Payload: b})
	}
	var batch terminal.Batch
	if err := call(ctx, "terminal.attach", map[string]any{"after": after, "maxBytes": 64 << 10}, &batch); err != nil {
		sendError("ATTACH_FAILED", err.Error())
		return
	}
	defer func() {
		detachCtx, stop := context.WithTimeout(context.Background(), 2*time.Second)
		defer stop()
		_ = call(detachCtx, "terminal.detach", nil, nil)
	}()
	if batch.Gap {
		sendError("OUTPUT_GAP", "终端归档已超出检查点保留范围，无法完整恢复原画面")
		return
	}
	var writable, replayed atomic.Bool
	acquire := func(takeover bool) error {
		if !s.terminalPermission(ctx, u, session.Resource, session.Backend, "terminal.input") {
			return terminal.ErrForbidden
		}
		var lease terminal.Lease
		err := call(ctx, "terminal.lease", map[string]any{"takeover": takeover}, &lease)
		writable.Store(err == nil)
		return err
	}
	_ = acquire(false)
	initialEarliest, initialLatest := batch.Earliest, batch.Latest
	ready := func() error {
		b, _ := json.Marshal(map[string]any{"sessionId": session.ID, "writable": writable.Load(), "earliest": initialEarliest, "latest": initialLatest, "state": session.State})
		return c.Send(ctx, protocol.Envelope{Type: protocol.TypeOpenAck, StreamID: session.ID, Payload: b})
	}
	if err := ready(); err != nil {
		return
	}
	go func() {
		defer cancel()
		for {
			e, err := c.Read(ctx)
			if err != nil {
				return
			}
			if e.StreamID != "" && e.StreamID != session.ID {
				return
			}
			switch e.Type {
			case protocol.TypeData:
				if !replayed.Load() || !writable.Load() || len(e.Payload) > 64<<10 || !s.terminalPermission(ctx, u, session.Resource, session.Backend, "terminal.input") {
					sendError("INPUT_DENIED", "当前视图尚未取得终端输入权")
					return
				}
				if err := call(ctx, "terminal.input", map[string]any{"data": e.Payload}, nil); err != nil {
					sendError("INPUT_OUTCOME_UNKNOWN", err.Error()+"；不自动重发按键")
					return
				}
				if err := c.Consume(ctx, session.ID, e.Sequence); err != nil {
					return
				}
			case protocol.TypeResize:
				if !replayed.Load() || !writable.Load() {
					continue
				}
				var size struct {
					Cols uint16 `json:"cols"`
					Rows uint16 `json:"rows"`
				}
				if err := json.Unmarshal(e.Payload, &size); err != nil {
					return
				}
				if err := call(ctx, "terminal.resize", map[string]any{"cols": size.Cols, "rows": size.Rows}, nil); err != nil {
					sendError("RESIZE_DENIED", err.Error())
					return
				}
			case protocol.TypeOpen:
				var request struct {
					Takeover bool `json:"takeover"`
				}
				if err := json.Unmarshal(e.Payload, &request); err != nil {
					return
				}
				if err := acquire(request.Takeover); err != nil {
					sendError("LEASE_DENIED", err.Error())
					continue
				}
				if err := ready(); err != nil {
					return
				}
			case protocol.TypePing:
				if err := c.Send(ctx, protocol.Envelope{Type: protocol.TypePong}); err != nil {
					return
				}
			default:
				return
			}
		}
	}()
	replayEnd := batch.Latest
	cursor := after
	var sent uint64
	lastLease := time.Now()
	for {
		if !s.terminalPermission(ctx, u, session.Resource, session.Backend, "terminal.read") {
			sendError("AUTHORIZATION_REVOKED", "终端权限已撤销")
			return
		}
		frames, err := terminalDataFrames(batch.Events, cursor, encoding == "events-v1")
		if err != nil {
			sendError("OUTPUT_UNAVAILABLE", "终端输出无法编码")
			return
		}
		for _, frame := range frames {
			payload := frame.payload
			sent += uint64(len(payload))
			if err := c.Send(ctx, protocol.Envelope{Type: protocol.TypeData, StreamID: session.ID, Sequence: sent, Payload: payload}); err != nil {
				return
			}
			cursor = frame.cursor
		}
		// Do not read another archived batch until xterm has parsed this one.
		ackDeadline := time.Now().Add(45 * time.Second)
		for c.Flow().Stats().SentUnacknowledged != 0 {
			if time.Now().After(ackDeadline) {
				sendError("SLOW_CONSUMER", "输出处理超过期限；可从已保护检查点重新挂载")
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-c.Done():
				return
			case <-time.After(10 * time.Millisecond):
			}
		}
		if !replayed.Load() && cursor >= replayEnd {
			replayed.Store(true)
			if err := c.Send(ctx, protocol.Envelope{Type: protocol.TypeResume, StreamID: session.ID}); err != nil {
				return
			}
		}
		if writable.Load() && time.Since(lastLease) > 10*time.Second {
			if err := acquire(false); err != nil {
				_ = ready()
			}
			lastLease = time.Now()
		}
		select {
		case <-ctx.Done():
			return
		case <-c.Done():
			return
		case <-time.After(75 * time.Millisecond):
		}
		if err := call(ctx, "terminal.read", map[string]any{"after": cursor, "maxBytes": 64 << 10}, &batch); err != nil {
			sendError("OUTPUT_UNAVAILABLE", err.Error())
			return
		}
		if batch.Gap {
			sendError("OUTPUT_GAP", "输出归档发生缺口，停止拼接不完整画面")
			return
		}
	}
}

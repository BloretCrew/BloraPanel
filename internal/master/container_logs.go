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
	"github.com/coder/websocket"
)

func (s *Server) containerLogs(w http.ResponseWriter, r *http.Request, u model.User) {
	ref := nodeRef(r.PathValue("id"))
	if r.TLS == nil || r.Header.Get("Origin") != s.origin {
		fail(w, 403, "ORIGIN_DENIED", "日志流需要可信TLS来源")
		return
	}
	if !s.allowed(w, r, u, ref, "host.manage") {
		return
	}
	containerID := r.PathValue("containerId")
	if len(containerID) != 64 {
		fail(w, 400, "INVALID_TARGET", "需要容器完整ID")
		return
	}
	tail := 200
	if q := r.URL.Query().Get("tail"); q != "" {
		n, err := strconv.Atoi(q)
		if err != nil || n < 1 || n > 1000 {
			fail(w, 400, "TAIL_LIMIT", "日志尾部行数需要1～1000")
			return
		}
		tail = n
	}
	args, _ := json.Marshal(map[string]any{"containerId": containerID, "tail": tail})
	read := func(ctx context.Context) (model.ContainerLogWindow, error) {
		var result model.ContainerLogWindow
		current, err := s.store.User(ctx, u.ID)
		if err != nil || !s.store.Allowed(ctx, current, ref, "host.manage") {
			return result, &model.APIError{Code: "FORBIDDEN", Message: "主机管理权限已撤销"}
		}
		err = s.nodeCall(ctx, ref.ID, protocol.ChannelInteractive, bridge.Request{Method: "container.logs", ActorID: u.ID, Resource: ref, Args: args}, &result)
		return result, err
	}
	first, err := read(r.Context())
	if err != nil {
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
	c, err := protocol.NewConn(ctx, ws, protocol.Options{Generation: 1, Channel: protocol.ChannelInteractive, InitialStreams: []string{containerID}})
	if err != nil {
		return
	}
	defer c.Close()
	go func() {
		defer cancel()
		for {
			e, err := c.Read(ctx)
			if err != nil || e.Type != protocol.TypePing {
				return
			}
			if c.Send(ctx, protocol.Envelope{Type: protocol.TypePong}) != nil {
				return
			}
		}
	}()
	b, _ := json.Marshal(map[string]any{"containerId": containerID, "mode": "timestamp-window", "possibleGap": true, "message": "Docker保留日志窗口，每次替换显示；窗口可能重叠或存在保留缺口，不提供连续事件游标"})
	if err := c.Send(ctx, protocol.Envelope{Type: protocol.TypeOpenAck, StreamID: containerID, Payload: b}); err != nil {
		return
	}
	sent := uint64(0)
	window := first
	for {
		b, _ := json.Marshal(window)
		sent += uint64(len(b))
		if err := c.Send(ctx, protocol.Envelope{Type: protocol.TypeData, StreamID: containerID, Sequence: sent, Payload: b}); err != nil {
			return
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
				return
			case <-time.After(20 * time.Millisecond):
			}
		}
		deadline.Stop()
		select {
		case <-ctx.Done():
			return
		case <-c.Done():
			return
		case <-time.After(time.Second):
		}
		window, err = read(ctx)
		if err != nil {
			b, _ := json.Marshal(model.APIError{Code: "CONTAINER_LOG_UNAVAILABLE", Message: err.Error()})
			_ = c.Send(ctx, protocol.Envelope{Type: protocol.TypeError, StreamID: containerID, Payload: b})
			return
		}
	}
}

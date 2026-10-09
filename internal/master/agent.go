package master

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func (s *Server) enroll(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Token     string `json:"token"`
		PublicKey []byte `json:"publicKey"`
	}
	if !decode(w, r, &in) {
		return
	}
	if len(in.PublicKey) != ed25519.PublicKeySize {
		fail(w, 400, "INVALID_KEY", "无效节点公钥")
		return
	}
	n, err := s.store.Enroll(r.Context(), in.Token, in.PublicKey)
	if err != nil {
		fail(w, 403, "ENROLLMENT_DENIED", "登记凭据失效或已使用")
		return
	}
	reply(w, 201, map[string]any{"node": n})
}

func (s *Server) agent(w http.ResponseWriter, r *http.Request) {
	if r.TLS == nil {
		fail(w, 400, "TLS_REQUIRED", "节点连接必须使用TLS")
		return
	}
	id := r.URL.Query().Get("nodeId")
	node, key, err := s.store.Node(r.Context(), id)
	if err != nil {
		fail(w, 403, "NODE_DENIED", "节点未登记或已吊销")
		return
	}
	generation, err := s.store.Generation(r.Context(), id)
	if err != nil {
		fail(w, 403, "NODE_DENIED", "节点已吊销")
		return
	}
	ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	defer ws.CloseNow()
	ws.SetReadLimit(16384)
	handshake, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	challenge, err := protocol.NewChallenge(id, model.ID(), generation+1, protocol.ChannelControl, 30*time.Second)
	if err != nil {
		return
	}
	if err := wsjson.Write(handshake, ws, challenge); err != nil {
		return
	}
	var auth struct {
		Signature []byte           `json:"signature"`
		Hello     model.AgentHello `json:"hello"`
	}
	if err := wsjson.Read(handshake, ws, &auth); err != nil {
		return
	}
	if err := protocol.VerifyChallenge(ed25519.PublicKey(key), challenge, auth.Signature); err != nil {
		return
	}
	// One response on this socket consumes the challenge; generation CAS also rejects competing handshakes.
	newGeneration, err := s.store.AdvanceGeneration(handshake, id, generation)
	if err != nil {
		return
	}
	if err := wsjson.Write(handshake, ws, map[string]any{"generation": newGeneration, "protocolVersion": protocol.Version}); err != nil {
		return
	}
	c, err := protocol.NewConn(r.Context(), ws, protocol.Options{Generation: newGeneration, Channel: protocol.ChannelControl, CurrentGeneration: func() uint64 {
		g, e := s.store.Generation(context.Background(), id)
		if e != nil {
			return 0
		}
		return g
	}})
	if err != nil {
		return
	}
	defer c.Close()
	p := &peer{conn: c, generation: newGeneration}
	s.mu.Lock()
	old := s.peers[id]
	s.peers[id] = p
	s.mu.Unlock()
	if old != nil {
		_ = old.conn.Close()
	}
	node.Generation = newGeneration
	node.Platform = auth.Hello.Platform
	node.Capabilities = auth.Hello.Capabilities
	node.StartupID = auth.Hello.StartupID
	node.State = "ONLINE"
	node.LastSeen = time.Now().UTC()
	node.Revision++
	if err := s.store.PutNode(r.Context(), node); err != nil {
		return
	}
	ctx, stop := context.WithCancel(r.Context())
	defer stop()
	go s.dispatch(ctx, id, p)
	defer func() {
		s.mu.Lock()
		current := s.peers[id] == p
		if current {
			delete(s.peers, id)
		}
		s.mu.Unlock()
		if current {
			node.State = "OFFLINE"
			node.Revision++
			_ = s.store.PutNode(context.Background(), node)
			s.markNodeWaiting(context.Background(), id)
		}
	}()
	for {
		readCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
		e, err := c.Read(readCtx)
		cancel()
		if err != nil {
			return
		}
		node.LastSeen = time.Now().UTC()
		switch e.Type {
		case protocol.TypePing:
			if err := c.Send(ctx, protocol.Envelope{Type: protocol.TypePong}); err != nil {
				return
			}
			if err := s.store.PutNode(ctx, node); err != nil {
				return
			}
		case protocol.TypeTaskEvent:
			var event model.AgentEvent
			if err := json.Unmarshal(e.Payload, &event); err != nil {
				return
			}
			t, err := s.store.ReconcileTask(ctx, id, event.Task)
			if err != nil {
				return
			}
			if event.Instance != nil && event.Instance.ID == t.Resource.ID {
				if err := s.store.ReconcileInstance(ctx, id, *event.Instance); err != nil {
					return
				}
			}
		case protocol.TypeOpenAck:
			var receipt model.AgentReceipt
			if err := json.Unmarshal(e.Payload, &receipt); err != nil {
				return
			}
			t, err := s.store.Task(ctx, receipt.TaskID)
			if err != nil {
				return
			}
			if t.Resource.NodeID != id || t.RequestID != receipt.RequestID {
				return
			}
			if receipt.Missing && !t.State.Terminal() {
				_, err = s.store.UpdateTask(ctx, t.ID, t.Revision, model.Interrupted, "node_acceptance_missing", nil, "节点未找到原请求记录；不自动重复副作用，请核对资源后显式重试")
				if err != nil {
					return
				}
			}
		case protocol.TypeSnapshot:
			var snapshot struct {
				Instances    []model.Instance `json:"instances"`
				StartupReady bool             `json:"startupReady"`
				StartupID    string           `json:"startupId"`
			}
			if err := json.Unmarshal(e.Payload, &snapshot); err != nil {
				return
			}
			if len(snapshot.Instances) > 32 {
				return
			}
			for _, instance := range snapshot.Instances {
				if err := s.store.ReconcileInstance(ctx, id, instance); err != nil {
					return
				}
			}
			if snapshot.StartupReady {
				if snapshot.StartupID != node.StartupID || len(snapshot.StartupID) != 32 {
					return
				}
				if err := s.enqueueAutostart(ctx, node); err != nil {
					return
				}
			}
		default:
			return
		}
	}
}

func (s *Server) dispatch(ctx context.Context, node string, p *peer) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	sent := map[string]time.Time{}
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.conn.Done():
			return
		case <-ticker.C:
		}
		tasks, err := s.store.Pending(ctx)
		if err != nil {
			continue
		}
		for _, t := range tasks {
			if t.Resource.NodeID != node || isLocalTask(t.Action) {
				continue
			}
			if t.State == model.WaitingClient {
				continue
			}
			// Client uploads are prepared through the bulk RPC, not a daemon
			// task receipt. Their cancel route must confirm staging cleanup;
			// dispatching a generic cancel first races that cleanup with a
			// misleading missing-task receipt. Already dispatched commits
			// still require the normal daemon cancellation/reconciliation.
			if t.CancellationRequested && t.DispatchedAt.IsZero() && (t.Action == "file.upload" || t.Action == "file.save") {
				continue
			}
			if t.DispatchedAt.IsZero() && (t.Action == "instance.start" || t.Action == "instance.restart") {
				n, _, err := s.store.Node(ctx, node)
				if err != nil || n.Maintenance {
					continue
				}
			}
			if time.Since(sent[t.ID+string(t.State)]) < 3*time.Second {
				continue
			}
			u, err := s.store.User(ctx, t.ActorID)
			if (err != nil || !s.authorizeTask(ctx, u, t)) && t.DispatchedAt.IsZero() {
				if t.State == model.Queued || t.State == model.WaitingNode {
					_, _ = s.store.UpdateTask(ctx, t.ID, t.Revision, model.Failed, "authorization_revoked", nil, "权限已撤销，未下发任务")
				}
				continue
			}
			kind := "task"
			if t.CancellationRequested {
				kind = "cancel"
			}
			if !t.DispatchedAt.IsZero() && !t.CancellationRequested {
				kind = "reconcile"
			}
			t, err = s.store.MarkDispatch(ctx, t.ID, t.Revision)
			if err != nil {
				continue
			}
			data, err := json.Marshal(model.AgentCommand{Kind: kind, Task: t})
			if err != nil {
				continue
			}
			if err := p.conn.Send(ctx, protocol.Envelope{Type: protocol.TypeOpen, RequestID: t.RequestID, Payload: data}); err != nil {
				return
			}
			sent[t.ID+string(t.State)] = time.Now()
		}
		if len(sent) > 20000 {
			sent = map[string]time.Time{}
		}
	}
}

func (s *Server) markNodeWaiting(ctx context.Context, nodeID string) {
	tasks, err := s.store.Pending(ctx)
	if err != nil {
		return
	}
	for _, t := range tasks {
		if t.Resource.NodeID == nodeID && t.State != model.WaitingClient && !isLocalTask(t.Action) {
			_, _ = s.store.WaitingNode(ctx, t.ID, t.Revision)
		}
	}
}

func (s *Server) events(w http.ResponseWriter, r *http.Request, u model.User) {
	if r.Header.Get("Origin") != s.requestOrigin(r) {
		fail(w, 403, "ORIGIN_DENIED", "WSS需要可信Origin")
		return
	}
	// Notification subscriptions start at the current durable boundary once.
	// Subsequent connections use the browser's atomically persisted cursor.
	notifications := r.URL.Query().Get("notifications") == "1"
	var cursor int64
	if notifications {
		latest, err := s.store.LatestEventSequence(r.Context())
		if err != nil {
			fail(w, 500, "STORAGE_ERROR", "任务事件读取失败")
			return
		}
		cursor = latest
		if raw := r.URL.Query().Get("after"); raw != "" {
			cursor, err = strconv.ParseInt(raw, 10, 64)
			if err != nil || cursor < 0 || cursor > latest || cursor > 9007199254740991 {
				fail(w, 400, "INVALID_CURSOR", "任务通知游标无效")
				return
			}
		}
	}
	ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	defer ws.CloseNow()
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	id := model.ID()
	s.mu.Lock()
	if s.userStreams[u.ID] == nil {
		s.userStreams[u.ID] = map[string]context.CancelFunc{}
	}
	if len(s.userStreams[u.ID]) >= 8 {
		s.mu.Unlock()
		return
	}
	s.userStreams[u.ID][id] = cancel
	s.mu.Unlock()
	defer func() { s.mu.Lock(); delete(s.userStreams[u.ID], id); s.mu.Unlock() }()
	c, err := protocol.NewConn(ctx, ws, protocol.Options{Generation: 1, Channel: protocol.ChannelControl})
	if err != nil {
		return
	}
	defer c.Close()
	go func() {
		for {
			e, err := c.Read(ctx)
			if err != nil {
				cancel()
				return
			}
			if e.Type == protocol.TypePing {
				_ = c.Send(ctx, protocol.Envelope{Type: protocol.TypePong})
			}
		}
	}()
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	if notifications {
		if err := c.Send(ctx, protocol.Envelope{Type: protocol.TypeSnapshot, Sequence: uint64(cursor)}); err != nil {
			return
		}
	}
	for {
		events, err := s.store.Events(ctx, cursor, 128)
		if err != nil {
			return
		}
		for _, event := range events {
			cursor = event.Sequence
			var child struct {
				ParentID string `json:"transferParentId"`
			}
			if json.Unmarshal(event.Task.Payload, &child) == nil && child.ParentID != "" {
				continue
			}
			if !s.canTask(ctx, u, event.Task) {
				continue
			}
			var value any = event
			if notifications {
				if !event.Task.State.Terminal() {
					continue
				}
				// Keep control messages small and omit payload/configuration.
				value = map[string]any{"sequence": event.Sequence, "task": map[string]any{"taskId": event.Task.ID, "state": event.Task.State, "action": event.Task.Action, "resource": event.Task.Resource}}
			}
			b, err := json.Marshal(value)
			if err != nil {
				return
			}
			if err := c.Send(ctx, protocol.Envelope{Type: protocol.TypeTaskEvent, Sequence: uint64(event.Sequence), Payload: b}); err != nil {
				return
			}
		}
		if notifications && len(events) > 0 {
			if err := c.Send(ctx, protocol.Envelope{Type: protocol.TypeSnapshot, Sequence: uint64(cursor)}); err != nil {
				return
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-c.Done():
			return
		case <-ticker.C:
		}
	}
}

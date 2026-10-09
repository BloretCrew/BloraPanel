package master

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func (s *Server) dataAgent(w http.ResponseWriter, r *http.Request) {
	if !managementTransport(r) {
		fail(w, 400, "TLS_REQUIRED", "节点数据连接必须使用TLS")
		return
	}
	channel, err := protocol.ParseChannel(r.PathValue("channel"))
	if err != nil || channel == protocol.ChannelControl {
		fail(w, 400, "INVALID_CHANNEL", "无效数据通道")
		return
	}
	id := r.URL.Query().Get("nodeId")
	_, key, err := s.store.Node(r.Context(), id)
	if err != nil {
		fail(w, 403, "NODE_DENIED", "节点未登记")
		return
	}
	s.mu.Lock()
	control := s.peers[id]
	s.mu.Unlock()
	if control == nil {
		fail(w, 409, "CONTROL_REQUIRED", "需要有效控制连接")
		return
	}
	generation := control.generation
	ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	defer ws.CloseNow()
	ws.SetReadLimit(16384)
	handshake, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	challenge, err := protocol.NewChallenge(id, model.ID(), generation, channel, 30*time.Second)
	if err != nil {
		return
	}
	if err := wsjson.Write(handshake, ws, challenge); err != nil {
		return
	}
	var auth struct {
		Signature []byte `json:"signature"`
	}
	if err := wsjson.Read(handshake, ws, &auth); err != nil {
		return
	}
	if err := protocol.VerifyChallenge(ed25519.PublicKey(key), challenge, auth.Signature); err != nil {
		return
	}
	if now, err := s.store.Generation(handshake, id); err != nil || now != generation {
		return
	}
	if err := wsjson.Write(handshake, ws, map[string]any{"generation": generation, "protocolVersion": protocol.Version}); err != nil {
		return
	}
	c, err := protocol.NewConn(r.Context(), ws, protocol.Options{Generation: generation, Channel: channel, InitialStreams: []string{bridge.StreamID}, CurrentGeneration: func() uint64 {
		g, err := s.store.Generation(context.Background(), id)
		if err != nil {
			return 0
		}
		return g
	}})
	if err != nil {
		return
	}
	link := bridge.New(r.Context(), c, func(ctx context.Context, request bridge.Request) (any, error) {
		return s.nodeAuthority(ctx, id, request)
	})
	defer link.Close()
	keyID := dataKey(id, channel)
	s.mu.Lock()
	old := s.links[keyID]
	s.links[keyID] = link
	s.mu.Unlock()
	if old != nil {
		_ = old.Close()
	}
	defer func() {
		s.mu.Lock()
		if s.links[keyID] == link {
			delete(s.links, keyID)
		}
		s.mu.Unlock()
	}()
	select {
	case <-r.Context().Done():
	case <-control.conn.Done():
	case <-link.Done():
	}
}
func dataKey(node string, channel protocol.Channel) string { return node + "/" + channel.String() }
func (s *Server) nodeCall(ctx context.Context, node string, channel protocol.Channel, request bridge.Request, out any) error {
	// A control connection can be ready a few milliseconds before its bulk and
	// interactive data links. This happens most visibly when a browser reloads
	// while the old terminal stream is detaching. Wait briefly for the expected
	// data link instead of turning that normal reconnection window into a
	// permanent NODE_UNREACHABLE view error. If the control peer is gone, fail
	// immediately; the caller still gets a distinct offline result.
	waitCtx, cancelWait := context.WithTimeout(ctx, 2*time.Second)
	defer cancelWait()
	var link *bridge.Link
	for {
		s.mu.Lock()
		link = s.links[dataKey(node, channel)]
		control := s.peers[node]
		s.mu.Unlock()
		if link != nil {
			break
		}
		if control == nil {
			return &model.APIError{Code: "NODE_UNREACHABLE", Message: "节点数据通道尚未连接"}
		}
		select {
		case <-waitCtx.Done():
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return &model.APIError{Code: "NODE_UNREACHABLE", Message: "节点数据通道尚未连接"}
		case <-time.After(50 * time.Millisecond):
		}
	}
	callCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	return link.Call(callCtx, request, out)
}
func bridgeError(w http.ResponseWriter, err error) {
	code := "NODE_OPERATION_FAILED"
	status := 502
	message := err.Error()
	if api, ok := err.(*model.APIError); ok {
		code = api.Code
		message = api.Message
		if code == "FORBIDDEN" {
			status = 403
		}
		if code == "NODE_UNREACHABLE" {
			status = 503
		}
		if strings.Contains(code, "CONFLICT") {
			status = 409
		}
		switch code {
		case "NOT_FOUND":
			status = 404
		case "RESOURCE_LIMIT":
			status = 413
		case "FILE_TRANSFER_MISMATCH":
			status = 409
		case "FILE_UNSUPPORTED":
			status = 415
		}
	}
	fail(w, status, code, message)
}

func (s *Server) nodeAuthority(ctx context.Context, node string, r bridge.Request) (any, error) {
	if (r.Method != "authority.check" && r.Method != "authority.task" && r.Method != "extension.module") || r.Resource.NodeID != node {
		return nil, &model.APIError{Code: "FORBIDDEN", Message: "节点权限查询范围无效"}
	}
	u, err := s.store.User(ctx, r.ActorID)
	if err != nil {
		return map[string]bool{"allowed": false}, nil
	}
	var args struct {
		Action    string `json:"action"`
		TaskID    string `json:"taskId"`
		Digest    string `json:"digest"`
		RequestID string `json:"requestId"`
		Offset    int    `json:"offset"`
	}
	if err := json.Unmarshal(r.Args, &args); err != nil {
		return nil, err
	}
	if r.Method == "authority.task" || r.Method == "extension.module" {
		t, err := s.store.Task(ctx, args.TaskID)
		matched := err == nil && t.ActorID == u.ID && t.Resource == r.Resource && t.Digest == args.Digest && t.RequestID == args.RequestID
		if matched && t.CancellationRequested {
			if r.Method == "extension.module" {
				return nil, &model.APIError{Code: "TASK_CANCELLED", Message: "任务已请求取消"}
			}
			return map[string]bool{"allowed": false, "cancelled": true}, nil
		}
		allowed := err == nil && t.ActorID == u.ID && t.Resource == r.Resource && t.Digest == args.Digest && t.RequestID == args.RequestID && !t.State.Terminal() && !t.CancellationRequested && s.authorizeTask(ctx, u, t)
		if allowed && t.Action == "extension.task" {
			var p model.ExtensionTaskPayload
			_ = json.Unmarshal(t.Payload, &p)
			x, getErr := s.extensions.Get(p.ExtensionID)
			allowed = getErr == nil && x.SHA256 == p.PackageHash
		}
		if r.Method == "extension.module" {
			if !allowed || t.Action != "extension.task" || args.Offset < 0 {
				return nil, &model.APIError{Code: "FORBIDDEN", Message: "扩展模块读取未获授权"}
			}
			var p model.ExtensionTaskPayload
			if err := json.Unmarshal(t.Payload, &p); err != nil {
				return nil, err
			}
			module, err := s.extensions.Backend(p.ExtensionID, p.PackageHash)
			if err != nil {
				return nil, err
			}
			if args.Offset >= len(module) {
				return nil, &model.APIError{Code: "INVALID_OFFSET", Message: "无效模块偏移"}
			}
			end := min(args.Offset+32*1024, len(module))
			return map[string]any{"data": module[args.Offset:end], "total": len(module)}, nil
		}
		return map[string]bool{"allowed": allowed}, nil
	}
	if !actions[args.Action] {
		return map[string]bool{"allowed": false}, nil
	}
	resource := r.Resource
	if resource.Kind == "instance" {
		i, err := s.store.Instance(ctx, resource.ID)
		if err != nil || i.NodeID != node {
			return map[string]bool{"allowed": false}, nil
		}
		if args.Action == "host.manage" {
			resource = model.ResourceRef{Kind: "node", ID: node}
		}
	} else if resource.Kind != "node" || resource.ID != node {
		return map[string]bool{"allowed": false}, nil
	}
	return map[string]bool{"allowed": s.store.Allowed(ctx, u, resource, args.Action)}, nil
}

package master

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/coreupdate"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
)

// Core updates execute trusted release code on the host. They deliberately do
// not inherit app.use, host.manage or extension permissions.
func (s *Server) registerCoreUpdates() {
	s.mux.HandleFunc("GET /api/v1/system/updates", s.auth(s.coreUpdateStatus))
	s.mux.HandleFunc("POST /api/v1/system/updates/check", s.auth(s.coreUpdateCheck))
	s.mux.HandleFunc("POST /api/v1/system/updates/apply", s.auth(s.coreUpdateApply))
	s.mux.HandleFunc("PUT /api/v1/system/updates/source", s.auth(s.coreUpdateSource))
	for _, method := range []string{"status", "check", "apply", "source"} {
		verb := "POST"
		if method == "status" {
			verb = "GET"
		}
		if method == "source" {
			verb = "PUT"
		}
		s.mux.HandleFunc(verb+" /api/v1/nodes/{id}/updates/"+method, s.auth(s.nodeCoreUpdate))
	}
}

func (s *Server) coreUpdateAvailable(w http.ResponseWriter, u model.User) bool {
	if !s.admin(w, u) {
		return false
	}
	if s.coreUpdater == nil {
		fail(w, http.StatusServiceUnavailable, "UPDATES_UNAVAILABLE", "此运行方式尚未配置后台更新服务")
		return false
	}
	return true
}

func (s *Server) coreUpdateStatus(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.coreUpdateAvailable(w, u) {
		return
	}
	status := s.coreUpdater.Status()
	if status.Preview != nil {
		preview := *status.Preview
		if reason := s.coreUpdateFleetReason(r, preview); reason != "" {
			preview.Compatible = false
			preview.Reason = reason
		}
		status.Preview = &preview
	}
	reply(w, http.StatusOK, status)
}

func (s *Server) coreUpdateCheck(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.coreUpdateAvailable(w, u) {
		return
	}
	status, err := s.coreUpdater.BeginCheck()
	if err != nil {
		fail(w, http.StatusConflict, "UPDATE_CHECK_FAILED", err.Error())
		return
	}
	reply(w, http.StatusAccepted, status)
}

func (s *Server) coreUpdateApply(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.coreUpdateAvailable(w, u) || !requireRequestID(w, r) {
		return
	}
	var in struct {
		Revision string `json:"revision"`
	}
	if !decode(w, r, &in) {
		return
	}
	requestID := u.ID + ":" + r.Header.Get("Idempotency-Key")
	if old, found, err := s.coreUpdater.Lookup(requestID, in.Revision); err != nil {
		fail(w, http.StatusConflict, "IDEMPOTENCY_CONFLICT", err.Error())
		return
	} else if found {
		reply(w, http.StatusAccepted, old)
		return
	}
	preview := s.coreUpdater.Status().Preview
	if preview == nil || preview.Revision != in.Revision {
		fail(w, http.StatusConflict, "UPDATE_CHECK_REQUIRED", "请先检查更新并确认指定提交")
		return
	}
	if reason := s.coreUpdateFleetReason(r, *preview); reason != "" {
		fail(w, http.StatusConflict, "UPDATE_PEER_INCOMPATIBLE", reason)
		return
	}
	job, err := s.coreUpdater.Start(requestID, in.Revision)
	if err != nil {
		fail(w, http.StatusConflict, "UPDATE_REJECTED", err.Error())
		return
	}
	_ = s.store.Audit(r.Context(), u.ID, model.ResourceRef{Kind: "master", ID: "local"}, "core.update.apply", r.Header.Get("Idempotency-Key"), job.ID)
	reply(w, http.StatusAccepted, job)
}

func (s *Server) coreUpdateSource(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.coreUpdateAvailable(w, u) || !requireRequestID(w, r) {
		return
	}
	var in coreupdate.Config
	if !decode(w, r, &in) {
		return
	}
	status, err := s.coreUpdater.SetSource(r.Context(), in)
	if err != nil {
		fail(w, http.StatusConflict, "UPDATE_SOURCE_REJECTED", err.Error())
		return
	}
	_ = s.store.Audit(r.Context(), u.ID, model.ResourceRef{Kind: "master", ID: "local"}, "core.update.source", r.Header.Get("Idempotency-Key"), "updated")
	reply(w, http.StatusOK, status)
}

func (s *Server) nodeCoreUpdate(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.admin(w, u) {
		return
	}
	ref := nodeRef(r.PathValue("id"))
	if _, _, err := s.store.Node(r.Context(), ref.ID); err != nil {
		fail(w, 404, "NODE_NOT_FOUND", "节点不存在")
		return
	}
	method := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	args := []byte(`{}`)
	if method == "apply" {
		if !requireRequestID(w, r) {
			return
		}
		var in struct {
			Revision string `json:"revision"`
		}
		if !decode(w, r, &in) {
			return
		}
		args, _ = json.Marshal(map[string]any{"revision": in.Revision, "requestId": u.ID + ":" + ref.ID + ":" + r.Header.Get("Idempotency-Key"), "peerProtocol": protocol.Version})
	} else if method == "source" {
		if !requireRequestID(w, r) {
			return
		}
		var in coreupdate.Config
		if !decode(w, r, &in) {
			return
		}
		args, _ = json.Marshal(in)
	}
	var out json.RawMessage
	if err := s.nodeCall(r.Context(), ref.ID, protocol.ChannelBulk, bridge.Request{Method: "core.update." + method, ActorID: u.ID, Resource: ref, Args: args}, &out); err != nil {
		bridgeError(w, err)
		return
	}
	if method == "apply" || method == "source" {
		_ = s.store.Audit(r.Context(), u.ID, ref, "core.update."+method, r.Header.Get("Idempotency-Key"), "accepted")
	}
	status := http.StatusOK
	if method == "apply" {
		status = http.StatusAccepted
	}
	reply(w, status, out)
}

// Offline nodes retain their last known protocol and remain in the gate. A
// missing capability means the existing protocol-v1 enrollment contract.
func (s *Server) coreUpdateFleetReason(r *http.Request, preview coreupdate.Preview) string {
	if err := s.ValidateCoreUpdatePeers(r.Context(), preview.Manifest); err != nil {
		return err.Error()
	}
	return ""
}

// ValidateCoreUpdatePeers is also called by the activation callback after
// preparation, because a node's capability may change during the download.
func (s *Server) ValidateCoreUpdatePeers(ctx context.Context, manifest coreupdate.Manifest) error {
	nodes, err := s.store.Nodes(ctx)
	if err != nil {
		return errors.New("无法读取节点兼容性，更新未获准")
	}
	for _, node := range nodes {
		version := protocol.Version
		if raw := node.Capabilities["core.protocol"]; raw != "" {
			parsed, err := strconv.ParseUint(raw, 10, 32)
			if err != nil || parsed == 0 {
				return errors.New("节点“" + node.Name + "”的协议版本无效，更新未获准")
			}
			version = uint32(parsed)
		}
		if !manifest.SupportsPeer(version) {
			return errors.New("更新与节点“" + node.Name + "”的协议不兼容；离线节点也必须纳入适配")
		}
	}
	return nil
}

// ValidateCoreUpdateActor rechecks the account at activation, instead of
// treating the session that admitted a long download as perpetual authority.
func (s *Server) ValidateCoreUpdateActor(ctx context.Context, requestID string) error {
	actor, _, ok := strings.Cut(requestID, ":")
	if !ok || actor == "" {
		return errors.New("更新请求缺少账号身份")
	}
	u, err := s.store.User(ctx, actor)
	if err != nil || !u.Admin || u.Disabled {
		return errors.New("更新账号已失去管理员权限，禁止切换版本")
	}
	return nil
}

package master

import (
	"crypto/ed25519"
	"errors"
	"net/http"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/storage"
)

func (s *Server) issueNodeRotation(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.admin(w, u) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var input struct {
		Reactivate bool `json:"reactivate"`
	}
	if !decode(w, r, &input) {
		return
	}
	ticket, err := s.store.NewRotationTicketMutation(r.Context(), u.ID, r.Header.Get("Idempotency-Key"), r.PathValue("id"), input.Reactivate)
	if errors.Is(err, storage.ErrRequestMismatch) {
		fail(w, 409, "IDEMPOTENCY_CONFLICT", err.Error())
		return
	}
	if err != nil {
		administrationError(w, err)
		return
	}
	reply(w, 201, map[string]any{"nodeId": r.PathValue("id"), "ticket": ticket, "expiresIn": 600})
}
func (s *Server) rotateNodeKey(w http.ResponseWriter, r *http.Request) {
	if r.TLS == nil {
		fail(w, 400, "TLS_REQUIRED", "节点换钥需要TLS")
		return
	}
	var input struct {
		NodeID    string `json:"nodeId"`
		Ticket    string `json:"ticket"`
		PublicKey []byte `json:"publicKey"`
		Signature []byte `json:"signature"`
	}
	if !decode(w, r, &input) {
		return
	}
	if len(input.PublicKey) != ed25519.PublicKeySize || len(input.Signature) != ed25519.SignatureSize || !ed25519.Verify(input.PublicKey, model.RotationProof(input.NodeID, input.Ticket, input.PublicKey), input.Signature) {
		fail(w, 403, "ROTATION_DENIED", "换钥证明无效")
		return
	}
	generation, err := s.store.RotateNodeKey(r.Context(), input.NodeID, input.Ticket, input.PublicKey)
	if err != nil {
		fail(w, 403, "ROTATION_DENIED", "换钥票据无效、过期或已用于不同密钥")
		return
	}
	s.disconnectNodeBefore(input.NodeID, generation)
	reply(w, 200, map[string]any{"nodeId": input.NodeID, "generation": generation, "rotated": true})
}

func (s *Server) disconnectNode(id string) {
	s.disconnectNodeBefore(id, ^uint64(0))
}
func (s *Server) disconnectNodeBefore(id string, generation uint64) {
	s.mu.Lock()
	p := s.peers[id]
	if p != nil && p.generation >= generation {
		s.mu.Unlock()
		return
	}
	delete(s.peers, id)
	var links []*bridge.Link
	for _, channel := range []protocol.Channel{protocol.ChannelInteractive, protocol.ChannelBulk} {
		key := dataKey(id, channel)
		if link := s.links[key]; link != nil {
			links = append(links, link)
			delete(s.links, key)
		}
	}
	s.mu.Unlock()
	if p != nil {
		_ = p.conn.Close()
	}
	for _, link := range links {
		_ = link.Close()
	}
}

package master

import (
	"net/http"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
)

func (s *Server) nodeSettings(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.admin(w, u) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var input storage.NodeSettings
	if !decode(w, r, &input) {
		return
	}
	n, err := s.store.SetNodeSettings(r.Context(), u.ID, r.PathValue("id"), r.Header.Get("Idempotency-Key"), input)
	if err != nil {
		administrationError(w, err)
		return
	}
	s.mu.Lock()
	connected := s.peers[n.ID] != nil
	s.mu.Unlock()
	if !n.Maintenance && connected {
		n.State = "ONLINE"
	}
	reply(w, 200, map[string]any{"node": n})
}

package master

import (
	"net/http"

	"blora.dev/panel/internal/model"
)

func (s *Server) registerInstanceSettings() {
	s.mux.HandleFunc("PATCH /api/v1/instances/{id}", s.auth(s.instanceSettings))
	s.mux.HandleFunc("GET /api/v1/instance-templates", s.auth(func(w http.ResponseWriter, r *http.Request, u model.User) {
		reply(w, 200, map[string]any{"items": model.InstanceTemplates()})
	}))
}

func (s *Server) instanceSettings(w http.ResponseWriter, r *http.Request, u model.User) {
	i, err := s.store.Instance(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, 404, "NOT_FOUND", "实例不存在")
		return
	}
	if !s.allowed(w, r, u, instanceRef(i), "instance.configure") {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var input model.InstanceSettings
	if !decode(w, r, &input) {
		return
	}
	updated, err := s.store.SetInstanceSettings(r.Context(), u.ID, i.ID, r.Header.Get("Idempotency-Key"), input)
	if err != nil {
		administrationError(w, err)
		return
	}
	s.nodePresentation(r.Context(), &updated)
	reply(w, 200, map[string]any{"instance": updated})
}

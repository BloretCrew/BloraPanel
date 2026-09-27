package master

import (
	"net/http"
	"strings"

	"blora.dev/panel/internal/model"
)

// Authorize an in-workspace notification. The browser persists and displays it;
// this endpoint does not send system notifications or contact other users.
func (s *Server) extensionNotification(w http.ResponseWriter, r *http.Request, u model.User) {
	id := r.PathValue("id")
	if s.extensions == nil {
		fail(w, 503, "EXTENSIONS_UNAVAILABLE", "扩展注册源未配置")
		return
	}
	if err := s.extensions.AuthorizeCapability(id, "notification.publish"); err != nil {
		fail(w, 403, "EXTENSION_CAPABILITY_DENIED", "扩展未启用或未声明通知能力")
		return
	}
	if !s.allowed(w, r, u, model.ResourceRef{Kind: "extension", ID: id}, "app.use") {
		return
	}
	var input struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	if !decode(w, r, &input) {
		return
	}
	if strings.TrimSpace(input.Title) == "" || len(input.Title) > 256 || len(input.Message) > 4096 {
		fail(w, 400, "INVALID_NOTIFICATION", "通知需要标题，标题限256字节、正文限4096字节")
		return
	}
	reply(w, 200, map[string]any{"appId": id, "title": input.Title, "message": input.Message})
}

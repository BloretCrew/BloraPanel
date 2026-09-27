package master

import (
	"net/http"

	"blora.dev/panel/internal/model"
)

// Only metadata crosses this bridge. Runtime configuration and host execution
// must not be obtained by declaring resource.write.
func (s *Server) extensionResourceUpdate(w http.ResponseWriter, r *http.Request, u model.User) {
	id := r.PathValue("id")
	if s.extensions == nil {
		fail(w, 503, "EXTENSIONS_UNAVAILABLE", "扩展注册源未配置")
		return
	}
	if err := s.extensions.AuthorizeCapability(id, "resource.write"); err != nil {
		fail(w, 403, "EXTENSION_CAPABILITY_DENIED", "扩展未启用或未声明资源写入能力")
		return
	}
	if !s.allowed(w, r, u, model.ResourceRef{Kind: "extension", ID: id}, "app.use") {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var in struct {
		Resource model.ResourceRef `json:"resource"`
		Revision *int64            `json:"revision"`
		Name     *string           `json:"name,omitempty"`
		Group    *string           `json:"group,omitempty"`
		Tags     *[]string         `json:"tags,omitempty"`
	}
	if !decode(w, r, &in) {
		return
	}
	if in.Revision == nil || *in.Revision < 0 {
		fail(w, 400, "REVISION_REQUIRED", "需要实例配置修订号")
		return
	}
	if in.Resource.Kind != "instance" || in.Resource.ID == "" {
		fail(w, 400, "INVALID_RESOURCE", "仅支持实例元数据")
		return
	}
	i, err := s.store.Instance(r.Context(), in.Resource.ID)
	if err != nil || (in.Resource.NodeID != "" && in.Resource.NodeID != i.NodeID) {
		fail(w, 404, "NOT_FOUND", "实例不存在")
		return
	}
	if !s.allowed(w, r, u, instanceRef(i), "instance.configure") {
		return
	}
	updated, err := s.store.SetInstanceSettings(r.Context(), u.ID, i.ID, r.Header.Get("Idempotency-Key"), model.InstanceSettings{Revision: *in.Revision, Name: in.Name, Group: in.Group, Tags: in.Tags})
	if err != nil {
		administrationError(w, err)
		return
	}
	reply(w, 200, map[string]any{"resource": instanceRef(updated), "name": updated.Name, "group": updated.Group, "tags": updated.Tags, "configRevision": updated.ConfigRevision, "revision": updated.Revision})
}

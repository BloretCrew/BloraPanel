package master

import (
	"encoding/json"
	"errors"
	"net/http"

	"blora.dev/panel/internal/extensions"
	"blora.dev/panel/internal/model"
)

func (s *Server) extensionData(w http.ResponseWriter, r *http.Request, u model.User) {
	id := r.PathValue("id")
	capability := "data.read"
	if r.Method == http.MethodPut {
		capability = "data.write"
	}
	if s.extensions == nil {
		fail(w, 503, "EXTENSIONS_UNAVAILABLE", "扩展服务不可用")
		return
	}
	if !s.store.Allowed(r.Context(), u, model.ResourceRef{Kind: "extension", ID: id}, "app.use") {
		fail(w, 403, "FORBIDDEN", "未获准使用此扩展")
		return
	}
	if err := s.extensions.AuthorizeCapability(id, capability); err != nil {
		fail(w, 403, "EXTENSION_CAPABILITY_DENIED", err.Error())
		return
	}
	var data extensions.UserData
	var err error
	if r.Method == http.MethodGet {
		data, err = s.extensions.ReadUserData(id, u.ID)
	} else {
		if !requireRequestID(w, r) {
			return
		}
		var in struct {
			SchemaVersion    int             `json:"schemaVersion"`
			ExpectedRevision *int64          `json:"expectedRevision"`
			Data             json.RawMessage `json:"data"`
		}
		if !decode(w, r, &in) {
			return
		}
		if in.ExpectedRevision == nil {
			fail(w, 400, "VERSION_REQUIRED", "需要 expectedRevision")
			return
		}
		if len(in.Data) > extensions.MaxUserData {
			fail(w, 413, "EXTENSION_DATA_TOO_LARGE", "数据超过 16 KiB")
			return
		}
		data, err = s.extensions.WriteUserData(id, u.ID, r.Header.Get("Idempotency-Key"), *in.ExpectedRevision, in.SchemaVersion, in.Data)
	}
	if err != nil {
		code := "EXTENSION_DATA_REJECTED"
		if errors.Is(err, extensions.ErrDataConflict) {
			code = "VERSION_CONFLICT"
		}
		fail(w, 409, code, err.Error())
		return
	}
	reply(w, 200, data)
}

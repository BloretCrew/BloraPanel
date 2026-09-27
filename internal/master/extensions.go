package master

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	"blora.dev/panel/internal/extensions"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
)

func (s *Server) registerExtensions() {
	s.mux.HandleFunc("POST /api/v1/extensions/{id}/notifications", s.auth(s.extensionNotification))
	s.mux.HandleFunc("GET /api/v1/extensions/{id}/data", s.auth(s.extensionData))
	s.mux.HandleFunc("PUT /api/v1/extensions/{id}/data", s.auth(s.extensionData))
	s.mux.HandleFunc("GET /api/v1/extensions", s.auth(s.extensionList))
	s.mux.HandleFunc("GET /api/v1/extensions/catalog", s.auth(s.extensionCatalogList))
	s.mux.HandleFunc("GET /api/v1/extensions/catalog/{appId}/{version}", s.auth(s.extensionCatalogPackage))
	s.mux.HandleFunc("POST /api/v1/extensions/catalog/install", s.auth(s.extensionCatalogInstall))
	s.mux.HandleFunc("GET /api/v1/extensions/{id}/manifest", s.auth(s.extensionManifest))
	s.mux.HandleFunc("GET /api/v1/extensions/{id}/bundle", s.auth(s.extensionBundle))
	s.mux.HandleFunc("GET /api/v1/extensions/{id}/package", s.auth(s.extensionPackage))
	s.mux.HandleFunc("POST /api/v1/extensions/install", s.auth(s.extensionInstall))
	s.mux.HandleFunc("POST /api/v1/extensions/install-package", s.auth(s.extensionInstallPackage))
	s.mux.HandleFunc("POST /api/v1/extensions/{id}/upgrade", s.auth(s.extensionUpgrade))
	s.mux.HandleFunc("POST /api/v1/extensions/{id}/rollback", s.auth(s.extensionRollback))
	s.mux.HandleFunc("POST /api/v1/extensions/{id}/enabled", s.auth(s.extensionEnabled))
	s.mux.HandleFunc("DELETE /api/v1/extensions/{id}", s.auth(s.extensionUninstall))
	s.mux.HandleFunc("POST /api/v1/extensions/{id}/tasks", s.auth(s.extensionTask))
	s.mux.HandleFunc("GET /api/v1/extensions/{id}/tasks/{taskId}/status", s.auth(s.extensionTaskRead))
	s.mux.HandleFunc("POST /api/v1/extensions/{id}/tasks/{taskId}/cancel", s.auth(s.extensionTaskCancel))
	s.mux.HandleFunc("GET /api/v1/extensions/{id}/resource", s.auth(s.extensionResource))
	s.mux.HandleFunc("PATCH /api/v1/extensions/{id}/resource", s.auth(s.extensionResourceUpdate))
}

func (s *Server) scopedExtensionTask(w http.ResponseWriter, r *http.Request, u model.User) (model.Task, bool) {
	t, err := s.store.Task(r.Context(), r.PathValue("taskId"))
	var p model.ExtensionTaskPayload
	if err != nil || t.Action != "extension.task" || json.Unmarshal(t.Payload, &p) != nil || p.ExtensionID != r.PathValue("id") || !s.canTask(r.Context(), u, t) {
		fail(w, 404, "NOT_FOUND", "扩展任务不存在或未授权")
		return t, false
	}
	return t, true
}
func (s *Server) extensionTaskRead(w http.ResponseWriter, r *http.Request, u model.User) {
	t, ok := s.scopedExtensionTask(w, r, u)
	if !ok {
		return
	}
	reply(w, 200, map[string]any{"task": t})
}
func (s *Server) extensionTaskCancel(w http.ResponseWriter, r *http.Request, u model.User) {
	t, ok := s.scopedExtensionTask(w, r, u)
	if !ok {
		return
	}
	r.SetPathValue("id", t.ID)
	s.cancelTask(w, r, u)
}

// extensionResource exposes resource summaries, never execution configuration
// or node credentials. Both the installed capability and caller grants apply.
func (s *Server) extensionResource(w http.ResponseWriter, r *http.Request, u model.User) {
	id := r.PathValue("id")
	if s.extensions == nil {
		fail(w, 503, "EXTENSIONS_UNAVAILABLE", "扩展注册源未配置")
		return
	}
	if err := s.extensions.AuthorizeCapability(id, "resource.read"); err != nil {
		fail(w, 403, "EXTENSION_CAPABILITY_DENIED", "扩展未启用或未声明资源读取能力")
		return
	}
	if !s.allowed(w, r, u, model.ResourceRef{Kind: "extension", ID: id}, "app.use") {
		return
	}
	ref := model.ResourceRef{Kind: r.URL.Query().Get("kind"), ID: r.URL.Query().Get("id"), NodeID: r.URL.Query().Get("nodeId")}
	if ref.ID == "" || (ref.Kind != "node" && ref.Kind != "instance") {
		fail(w, 400, "INVALID_RESOURCE", "需要节点或实例资源标识")
		return
	}
	if ref.Kind == "instance" {
		i, err := s.store.Instance(r.Context(), ref.ID)
		if err != nil || (ref.NodeID != "" && ref.NodeID != i.NodeID) {
			fail(w, 404, "NOT_FOUND", "资源不存在")
			return
		}
		if !s.allowed(w, r, u, instanceRef(i), "instance.read") {
			return
		}
		s.nodePresentation(r.Context(), &i)
		reply(w, 200, map[string]any{"resource": instanceRef(i), "name": i.Name, "group": i.Group, "tags": i.Tags, "configRevision": i.ConfigRevision, "state": i.State, "nodeState": i.NodeState, "revision": i.Revision})
		return
	}
	if ref.NodeID != "" && ref.NodeID != ref.ID {
		fail(w, 404, "NOT_FOUND", "资源不存在")
		return
	}
	ref.NodeID = ref.ID
	if !s.allowed(w, r, u, ref, "node.read") {
		return
	}
	n, _, err := s.store.Node(r.Context(), ref.ID)
	if err != nil {
		fail(w, 404, "NOT_FOUND", "资源不存在")
		return
	}
	// Use the same online/maintenance presentation as instance summaries.
	i := model.Instance{NodeID: n.ID}
	s.nodePresentation(r.Context(), &i)
	reply(w, 200, map[string]any{"resource": ref, "name": n.Name, "state": i.NodeState, "platform": n.Platform, "revision": n.Revision})
}

func (s *Server) extensionCatalogManager(w http.ResponseWriter, u model.User) bool {
	if !s.admin(w, u) {
		return false
	}
	if s.extensionCatalogErr != nil {
		fail(w, 503, "EXTENSION_CATALOG_INVALID", s.extensionCatalogErr.Error())
		return false
	}
	if s.extensionCatalog == nil {
		fail(w, 503, "EXTENSION_CATALOG_UNAVAILABLE", "扩展注册源未配置")
		return false
	}
	return true
}

func (s *Server) extensionCatalogList(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.extensionCatalogManager(w, u) {
		return
	}
	items, err := s.extensionCatalog.List()
	if err != nil {
		fail(w, 500, "EXTENSION_CATALOG_FAILED", err.Error())
		return
	}
	reply(w, 200, map[string]any{"items": items})
}

func (s *Server) extensionCatalogPackage(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.extensionCatalogManager(w, u) {
		return
	}
	p, err := s.extensionCatalog.Load(r.PathValue("appId"), r.PathValue("version"))
	if err != nil {
		fail(w, 404, "EXTENSION_CATALOG_NOT_FOUND", err.Error())
		return
	}
	reply(w, 200, map[string]any{"manifest": p.Manifest, "sha256": p.SHA256, "payload": p.Payload, "signature": p.Signature})
}

func (s *Server) extensionCatalogInstall(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.extensionCatalogManager(w, u) {
		return
	}
	if s.extensions == nil {
		fail(w, 503, "EXTENSIONS_UNAVAILABLE", "扩展安装目录未配置")
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var in struct {
		AppID   string `json:"appId"`
		Version string `json:"version"`
	}
	if !decode(w, r, &in) {
		return
	}
	p, err := s.extensionCatalog.Load(in.AppID, in.Version)
	if err != nil {
		fail(w, 404, "EXTENSION_CATALOG_NOT_FOUND", err.Error())
		return
	}
	status := http.StatusCreated
	var installed extensions.Installed
	if _, readErr := s.extensions.Get(in.AppID); readErr == nil {
		status = http.StatusOK
		installed, err = s.extensions.Upgrade(p)
	} else if errors.Is(readErr, os.ErrNotExist) {
		installed, err = s.extensions.Install(p)
	} else {
		err = readErr
	}
	if err != nil {
		fail(w, 409, "EXTENSION_REJECTED", err.Error())
		return
	}
	reply(w, status, installed)
}
func (s *Server) extensionManager(w http.ResponseWriter, u model.User) bool {
	if !s.admin(w, u) {
		return false
	}
	if s.extensions == nil {
		fail(w, 503, "EXTENSIONS_UNAVAILABLE", "扩展注册源未配置")
		return false
	}
	return true
}
func (s *Server) extensionList(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.extensionManager(w, u) {
		return
	}
	items, err := s.extensions.List()
	if err != nil {
		fail(w, 500, "EXTENSION_STORE_FAILED", err.Error())
		return
	}
	reply(w, 200, map[string]any{"items": items})
}
func (s *Server) extensionPackage(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.extensionManager(w, u) {
		return
	}
	b, err := s.extensions.Payload(r.PathValue("id"))
	if err != nil {
		fail(w, 404, "EXTENSION_NOT_FOUND", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=package.pkg")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}
func (s *Server) extensionManifest(w http.ResponseWriter, r *http.Request, u model.User) {
	if s.extensions == nil {
		fail(w, 503, "EXTENSIONS_UNAVAILABLE", "扩展注册源未配置")
		return
	}
	x, err := s.extensions.Get(r.PathValue("id"))
	if err != nil {
		fail(w, 404, "EXTENSION_NOT_FOUND", err.Error())
		return
	}
	if !x.Enabled {
		fail(w, 409, "EXTENSION_DISABLED", "扩展已禁用")
		return
	}
	reply(w, 200, x.Manifest)
}

// extensionBundle serves only the verified bytes of an enabled extension to
// an authenticated browser loader. Execution remains the loader's sandbox
// responsibility; this endpoint never grants a package-management privilege.
func (s *Server) extensionBundle(w http.ResponseWriter, r *http.Request, u model.User) {
	if s.extensions == nil {
		fail(w, 503, "EXTENSIONS_UNAVAILABLE", "扩展注册源未配置")
		return
	}
	x, b, err := s.extensions.Snapshot(r.PathValue("id"))
	if err != nil {
		fail(w, 404, "EXTENSION_NOT_FOUND", err.Error())
		return
	}
	if !x.Enabled {
		fail(w, 409, "EXTENSION_DISABLED", "扩展已禁用")
		return
	}
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "no-store")
	bundle, err := extensions.DecodeBundle(b)
	if err != nil {
		fail(w, 409, "EXTENSION_BUNDLE_INVALID", err.Error())
		return
	}
	_, _ = w.Write([]byte(bundle.Frontend))
}
func decodeExtension(w http.ResponseWriter, r *http.Request) (extensions.Package, bool) {
	var in struct {
		Manifest  extensions.Manifest `json:"manifest"`
		SHA256    string              `json:"sha256"`
		Payload   []byte              `json:"payload"`
		Signature []byte              `json:"signature"`
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<20))
	d.DisallowUnknownFields()
	if err := d.Decode(&in); err != nil {
		fail(w, 400, "EXTENSION_PACKAGE_INVALID", "无效扩展包")
		return extensions.Package{}, false
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		fail(w, 400, "EXTENSION_PACKAGE_INVALID", "包文件包含多余 JSON")
		return extensions.Package{}, false
	}
	return extensions.Package{Manifest: in.Manifest, SHA256: in.SHA256, Payload: in.Payload, Signature: in.Signature}, true
}
func (s *Server) extensionInstall(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.extensionManager(w, u) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	p, ok := decodeExtension(w, r)
	if !ok {
		return
	}
	x, err := s.extensions.Install(p)
	if err != nil {
		fail(w, 409, "EXTENSION_REJECTED", err.Error())
		return
	}
	reply(w, 201, x)
}

// extensionInstallPackage accepts the SDK's standalone JSON package envelope.
func (s *Server) extensionInstallPackage(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.extensionManager(w, u) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	b, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 16<<20))
	if err != nil {
		fail(w, 413, "EXTENSION_PACKAGE_TOO_LARGE", err.Error())
		return
	}
	var in struct {
		Manifest  extensions.Manifest `json:"manifest"`
		SHA256    string              `json:"sha256"`
		Payload   []byte              `json:"payload"`
		Signature []byte              `json:"signature"`
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err = dec.Decode(&in); err != nil {
		fail(w, 400, "EXTENSION_PACKAGE_INVALID", "包文件不是有效 JSON")
		return
	}
	var trailing any
	if err = dec.Decode(&trailing); !errors.Is(err, io.EOF) {
		fail(w, 400, "EXTENSION_PACKAGE_INVALID", "包文件包含多余 JSON")
		return
	}
	p := extensions.Package{Manifest: in.Manifest, SHA256: in.SHA256, Payload: in.Payload, Signature: in.Signature}
	x, err := s.extensions.Install(p)
	if err != nil {
		fail(w, 409, "EXTENSION_REJECTED", err.Error())
		return
	}
	reply(w, 201, x)
}
func (s *Server) extensionUpgrade(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.extensionManager(w, u) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	p, ok := decodeExtension(w, r)
	if !ok {
		return
	}
	if p.Manifest.AppID != r.PathValue("id") {
		fail(w, 400, "EXTENSION_ID_MISMATCH", "扩展标识不一致")
		return
	}
	x, err := s.extensions.Upgrade(p)
	if err != nil {
		fail(w, 409, "EXTENSION_REJECTED", err.Error())
		return
	}
	reply(w, 200, x)
}
func (s *Server) extensionRollback(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.extensionManager(w, u) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	x, err := s.extensions.Rollback(r.PathValue("id"))
	if err != nil {
		fail(w, 409, "EXTENSION_ROLLBACK_FAILED", err.Error())
		return
	}
	reply(w, 200, x)
}
func (s *Server) extensionEnabled(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.extensionManager(w, u) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var in struct {
		Enabled bool `json:"enabled"`
	}
	if !decode(w, r, &in) {
		return
	}
	x, err := s.extensions.SetEnabled(r.PathValue("id"), in.Enabled)
	if err != nil {
		fail(w, 404, "EXTENSION_NOT_FOUND", err.Error())
		return
	}
	reply(w, 200, x)
}
func (s *Server) extensionUninstall(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.extensionManager(w, u) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	if err := s.extensions.Uninstall(r.PathValue("id")); err != nil {
		fail(w, 404, "EXTENSION_NOT_FOUND", err.Error())
		return
	}
	if r.URL.Query().Get("cleanup") == "true" {
		if err := s.extensions.CleanupUserData(r.PathValue("id")); err != nil {
			fail(w, 500, "EXTENSION_DATA_CLEANUP_FAILED", err.Error())
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// extensionTask binds a WASI execution to the verified installed package.
func (s *Server) extensionTask(w http.ResponseWriter, r *http.Request, u model.User) {
	if s.extensions == nil {
		fail(w, 503, "EXTENSIONS_UNAVAILABLE", "扩展注册源未配置")
		return
	}
	id := r.PathValue("id")
	x, err := s.extensions.Get(id)
	if err != nil {
		fail(w, 404, "EXTENSION_NOT_FOUND", err.Error())
		return
	}
	if !x.Enabled {
		fail(w, 409, "EXTENSION_DISABLED", "扩展已禁用")
		return
	}
	if err := s.extensions.AuthorizeCapability(id, "task.create"); err != nil {
		fail(w, 403, "EXTENSION_CAPABILITY_DENIED", err.Error())
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var in struct {
		NodeID  string          `json:"nodeId"`
		Payload json.RawMessage `json:"payload"`
	}
	if !decode(w, r, &in) {
		return
	}
	if len(in.Payload) > extensions.MaxBackendIO {
		fail(w, 413, "EXTENSION_TASK_TOO_LARGE", "扩展任务数据超过 32 KiB")
		return
	}
	if !json.Valid(in.Payload) {
		fail(w, 400, "EXTENSION_INPUT_REQUIRED", "扩展任务需要 JSON 输入")
		return
	}
	if in.NodeID == "" {
		fail(w, 400, "NODE_REQUIRED", "扩展任务必须指定目标节点")
		return
	}
	ref := model.ResourceRef{Kind: "node", ID: in.NodeID, NodeID: in.NodeID}
	extRef := model.ResourceRef{Kind: "extension", ID: id}
	if !s.store.Allowed(r.Context(), u, extRef, "app.use") || !s.store.Allowed(r.Context(), u, ref, "node.read") {
		fail(w, 403, "FORBIDDEN", "无权在目标节点创建扩展任务")
		return
	}
	bound := model.ExtensionTaskPayload{ExtensionID: id, PackageHash: x.SHA256, Payload: in.Payload}
	if old, lookupErr := s.store.TaskByRequest(r.Context(), u.ID, r.Header.Get("Idempotency-Key")); lookupErr == nil && old.Action == "extension.task" && old.Resource == ref {
		var previous model.ExtensionTaskPayload
		if json.Unmarshal(old.Payload, &previous) == nil && previous.ExtensionID == id {
			bound.PackageHash, bound.ModuleHash = previous.PackageHash, previous.ModuleHash
		}
	}
	if bound.ModuleHash == "" {
		module, err := s.extensions.Backend(id, x.SHA256)
		if err != nil {
			fail(w, 409, "EXTENSION_BACKEND_UNAVAILABLE", err.Error())
			return
		}
		bound.ModuleHash = extensions.ModuleHash(module)
	}
	payload, _ := json.Marshal(bound)
	task, _, err := s.store.Accept(r.Context(), model.Task{ActorID: u.ID, RequestID: r.Header.Get("Idempotency-Key"), Resource: ref, Action: "extension.task", Payload: payload})
	if err != nil {
		// A key reused with different content stays distinguishable from a
		// capacity rejection, matching the other durable write entrypoints.
		if errors.Is(err, storage.ErrRequestMismatch) {
			fail(w, 409, "IDEMPOTENCY_CONFLICT", err.Error())
			return
		}
		fail(w, 409, "TASK_REJECTED", err.Error())
		return
	}
	reply(w, 202, map[string]any{"task": task})
}

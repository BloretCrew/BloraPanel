package master

import (
	"context"
	"crypto/ed25519"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/extensions"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type Options struct {
	Store               *storage.Store
	Origin              string
	StaticDir           string
	ExtensionRoot       string
	ExtensionCatalogDir string
	ExtensionCatalogURL string
	// ExtensionCatalogHTTPClient is optional and is primarily useful when a
	// deployment supplies a pinned CA/transport. The remote source still
	// replaces its redirect policy with a no-redirect policy.
	ExtensionCatalogHTTPClient *http.Client
	ExtensionTrustedKeys       []ed25519.PublicKey
}
type peer struct {
	conn       *protocol.Conn
	generation uint64
}
type Server struct {
	store               *storage.Store
	origin              string
	static              string
	mux                 *http.ServeMux
	mu                  sync.Mutex
	peers               map[string]*peer
	links               map[string]*bridge.Link
	userStreams         map[string]map[string]context.CancelFunc
	terminalViews       map[string]*terminalView
	transferJobs        *transferCoordinator
	schedules           *scheduleService
	attempts            map[string]loginAttempt
	extensions          *extensions.Manager
	extensionCatalog    extensions.CatalogSource
	extensionCatalogErr error
}
type loginAttempt struct {
	Since time.Time
	Count int
}

func New(opts Options) *Server {
	s := &Server{store: opts.Store, origin: strings.TrimRight(opts.Origin, "/"), static: opts.StaticDir, mux: http.NewServeMux(), peers: map[string]*peer{}, links: map[string]*bridge.Link{}, userStreams: map[string]map[string]context.CancelFunc{}, attempts: map[string]loginAttempt{}}
	if opts.ExtensionRoot != "" {
		s.extensions, _ = extensions.New(opts.ExtensionRoot, []string{"window.open", "window.move", "window.close", "shortcut.create", "data.read", "data.write", "resource.read", "resource.write", "notification.publish", "task.create"})
		if s.extensions != nil && len(opts.ExtensionTrustedKeys) > 0 {
			s.extensions.SetTrustedKeys(opts.ExtensionTrustedKeys)
		}
	}
	if opts.ExtensionCatalogDir != "" && opts.ExtensionCatalogURL != "" {
		s.extensionCatalogErr = errors.New("extension catalog directory and URL are mutually exclusive")
	} else if opts.ExtensionCatalogDir != "" {
		s.extensionCatalog, s.extensionCatalogErr = extensions.NewCatalog(opts.ExtensionCatalogDir)
	} else if opts.ExtensionCatalogURL != "" {
		s.extensionCatalog, s.extensionCatalogErr = extensions.NewRemoteCatalog(opts.ExtensionCatalogURL, opts.ExtensionCatalogHTTPClient)
	}
	s.mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := s.store.DB.PingContext(r.Context()); err != nil {
			fail(w, 503, "STORAGE_UNAVAILABLE", err.Error())
			return
		}
		reply(w, 200, map[string]string{"status": "ok"})
	})
	s.mux.HandleFunc("POST /api/v1/login", s.login)
	s.mux.HandleFunc("POST /api/v1/enroll", s.enroll)
	s.mux.HandleFunc("GET /api/v1/agent/control", s.agent)
	s.mux.HandleFunc("GET /api/v1/agent/data/{channel}", s.dataAgent)
	s.mux.HandleFunc("GET /api/v1/session", s.auth(s.session))
	s.mux.HandleFunc("POST /api/v1/logout", s.auth(s.logout))
	s.mux.HandleFunc("GET /api/v1/users", s.auth(s.users))
	s.mux.HandleFunc("POST /api/v1/users", s.auth(s.createUser))
	s.mux.HandleFunc("GET /api/v1/grants", s.auth(s.grants))
	s.mux.HandleFunc("POST /api/v1/grants", s.auth(s.grant))
	s.mux.HandleFunc("DELETE /api/v1/grants", s.auth(s.grant))
	s.mux.HandleFunc("GET /api/v1/nodes", s.auth(s.nodes))
	s.mux.HandleFunc("GET /api/v1/nodes/creatable", s.auth(s.nodes))
	s.mux.HandleFunc("GET /api/v1/nodes/managed", s.auth(s.nodes))
	s.mux.HandleFunc("POST /api/v1/nodes/enrollments", s.auth(s.newEnrollment))
	s.mux.HandleFunc("POST /api/v1/nodes/{id}/revoke", s.auth(s.revokeNode))
	s.mux.HandleFunc("GET /api/v1/instances", s.auth(s.instances))
	s.mux.HandleFunc("POST /api/v1/instances", s.auth(s.createInstance))
	s.mux.HandleFunc("GET /api/v1/instances/{id}", s.auth(s.instance))
	s.mux.HandleFunc("DELETE /api/v1/instances/{id}", s.auth(s.deleteInstance))
	s.mux.HandleFunc("POST /api/v1/instances/{id}/actions", s.auth(s.instanceAction))
	s.mux.HandleFunc("GET /api/v1/tasks", s.auth(s.tasks))
	s.mux.HandleFunc("GET /api/v1/tasks/summary", s.auth(s.taskSummary))
	s.mux.HandleFunc("GET /api/v1/tasks/{id}", s.auth(s.task))
	s.mux.HandleFunc("GET /api/v1/tasks/{id}/events", s.auth(s.taskStages))
	s.mux.HandleFunc("POST /api/v1/tasks/{id}/cancel", s.auth(s.cancelTask))
	s.mux.HandleFunc("POST /api/v1/tasks/{id}/retry", s.auth(s.retryTask))
	s.mux.HandleFunc("GET /api/v1/events", s.auth(s.events))
	s.registerTerminals()
	s.registerFiles()
	s.registerAdministration()
	s.registerTransfers()
	s.registerLogs()
	s.registerInstanceSettings()
	s.registerContainers()
	s.registerBackups()
	s.registerSchedules()
	s.registerMonitoring()
	s.registerSystem()
	s.registerExtensions()
	s.registerWorkspaces()
	if opts.StaticDir != "" {
		s.mux.Handle("/", http.FileServer(http.Dir(filepath.Clean(opts.StaticDir))))
	}
	return s
}
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "same-origin")
	// External extension bootstrap code is loaded from a Blob URL inside an
	// opaque sandbox iframe. Inline scripts remain forbidden; allowing blob:
	// here is required for the sandbox bootstrap and existing worker URLs.
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' blob:; style-src 'self' 'unsafe-inline'; worker-src 'self' blob:; img-src 'self' data: blob:; connect-src 'self' wss:; frame-src 'self'; frame-ancestors 'none'; object-src 'none'")
	if strings.HasPrefix(r.URL.Path, "/api/") {
		w.Header().Set("Cache-Control", "no-store")
	}
	if origin := r.Header.Get("Origin"); origin != "" && origin != s.origin {
		fail(w, 403, "ORIGIN_DENIED", "来源不受信任")
		return
	}
	if r.Header.Get("Origin") == s.origin && s.origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", s.origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Vary", "Origin")
	}
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token, Idempotency-Key")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.WriteHeader(204)
		return
	}
	s.mux.ServeHTTP(w, r)
}
func reply(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func fail(w http.ResponseWriter, status int, code, message string) {
	reply(w, status, map[string]any{"error": model.APIError{Code: code, Message: message}})
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		fail(w, 400, "INVALID_REQUEST", "无效请求内容")
		return false
	}
	var trailing any
	if err := d.Decode(&trailing); !errors.Is(err, io.EOF) {
		fail(w, 400, "INVALID_REQUEST", "仅允许一个 JSON 对象")
		return false
	}
	return true
}

type authed func(http.ResponseWriter, *http.Request, model.User)

func (s *Server) auth(next authed) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("blora_session")
		if err != nil {
			fail(w, 401, "UNAUTHENTICATED", "请登录")
			return
		}
		u, csrf, err := s.store.Session(r.Context(), cookie.Value)
		if err != nil {
			fail(w, 401, "UNAUTHENTICATED", "会话已失效")
			return
		}
		if r.Method != "GET" && r.Method != "HEAD" {
			if subtle.ConstantTimeCompare([]byte(csrf), []byte(r.Header.Get("X-CSRF-Token"))) != 1 {
				fail(w, 403, "CSRF_DENIED", "请求校验失效")
				return
			}
		}
		next(w, r, u)
	}
}
func (s *Server) allowed(w http.ResponseWriter, r *http.Request, u model.User, ref model.ResourceRef, action string) bool {
	if !s.store.Allowed(r.Context(), u, ref, action) {
		fail(w, 403, "FORBIDDEN", "需要权限："+action)
		return false
	}
	return true
}
func (s *Server) admin(w http.ResponseWriter, u model.User) bool {
	if !u.Admin {
		fail(w, 403, "FORBIDDEN", "需要平台管理员权限")
		return false
	}
	return true
}
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	s.mu.Lock()
	a := s.attempts[ip]
	if time.Since(a.Since) > time.Minute {
		a = loginAttempt{Since: time.Now()}
	}
	a.Count++
	if len(s.attempts) > 1024 {
		for key, v := range s.attempts {
			if time.Since(v.Since) > time.Minute {
				delete(s.attempts, key)
			}
		}
	}
	s.attempts[ip] = a
	s.mu.Unlock()
	if a.Count > 12 {
		fail(w, 429, "RATE_LIMIT", "登录尝试过多，请稍后重试")
		return
	}
	var in struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if !decode(w, r, &in) {
		return
	}
	u, hash, err := s.store.UserByName(r.Context(), in.Name)
	if err != nil {
		hash = []byte("$2a$12$C6UzMDM.H6dfI/f/IKcEe.2TGaYBRYWsLMxnLo4eZkS48oJcFkGEm")
	}
	check := bcrypt.CompareHashAndPassword(hash, []byte(in.Password))
	if err != nil || check != nil || u.Disabled {
		fail(w, 401, "LOGIN_FAILED", "账号或密码不正确")
		return
	}
	token, csrf, err := s.store.NewSession(r.Context(), u.ID)
	if err != nil {
		fail(w, 500, "STORAGE_ERROR", "无法建立会话")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "blora_session", Value: token, Path: "/", HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode, MaxAge: 43200})
	reply(w, 200, map[string]any{"user": u, "csrfToken": csrf})
}
func (s *Server) session(w http.ResponseWriter, r *http.Request, u model.User) {
	cookie, _ := r.Cookie("blora_session")
	_, csrf, err := s.store.Session(r.Context(), cookie.Value)
	if err != nil {
		fail(w, 401, "UNAUTHENTICATED", "会话已失效")
		return
	}
	reply(w, 200, map[string]any{"user": u, "csrfToken": csrf})
}
func (s *Server) logout(w http.ResponseWriter, r *http.Request, u model.User) {
	cookie, _ := r.Cookie("blora_session")
	if err := s.store.DropSession(r.Context(), cookie.Value); err != nil {
		fail(w, 500, "STORAGE_ERROR", err.Error())
		return
	}
	s.revokeStreams(u.ID)
	http.SetCookie(w, &http.Cookie{Name: "blora_session", Value: "", Path: "/", HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode, MaxAge: -1})
	reply(w, 200, map[string]bool{"loggedOut": true})
}
func (s *Server) users(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.admin(w, u) {
		return
	}
	items, err := s.store.Users(r.Context())
	if err != nil {
		fail(w, 500, "STORAGE_ERROR", err.Error())
		return
	}
	reply(w, 200, map[string]any{"items": items})
}
func (s *Server) createUser(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.admin(w, u) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var in struct {
		Name     string `json:"name"`
		Password string `json:"password"`
		Admin    bool   `json:"admin"`
	}
	if !decode(w, r, &in) {
		return
	}
	if len(in.Name) < 1 || len(in.Name) > 80 || len(in.Password) < 12 || len(in.Password) > 72 {
		fail(w, 400, "INVALID_REQUEST", "账号长度1～80，密码长度12～72字节")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 12)
	if err != nil {
		fail(w, 400, "INVALID_REQUEST", err.Error())
		return
	}
	created, err := s.store.CreateUserMutation(r.Context(), u.ID, r.Header.Get("Idempotency-Key"), model.User{Name: in.Name, Admin: in.Admin}, hash, storage.Hash([]byte(in.Password)))
	if err != nil {
		fail(w, 409, "USER_CONFLICT", "账号已存在或无法创建")
		return
	}
	reply(w, 201, map[string]any{"user": created})
}
func (s *Server) grants(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.admin(w, u) {
		return
	}
	items, err := s.store.Grants(r.Context())
	if err != nil {
		fail(w, 500, "STORAGE_ERROR", err.Error())
		return
	}
	reply(w, 200, map[string]any{"items": items})
}

var actions = map[string]bool{"instance.read": true, "instance.create": true, "instance.start": true, "instance.stop": true, "instance.restart": true, "instance.kill": true, "file.read": true, "file.write": true, "terminal.read": true, "terminal.input": true, "node.read": true, "host.manage": true, "backup.create": true, "backup.restore": true, "schedule.create": true, "process.terminate": true, "system.service.start": true, "system.service.stop": true, "system.service.restart": true, "system.task.enable": true, "system.task.disable": true, "system.firewall.apply": true, "app.use": true}

func init() { actions["instance.configure"] = true }

func (s *Server) grant(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.admin(w, u) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var in model.Grant
	if !decode(w, r, &in) {
		return
	}
	if !actions[in.Action] || (in.Resource.Kind != "node" && in.Resource.Kind != "instance" && !(in.Resource.Kind == "extension" && in.Action == "app.use")) || in.Resource.ID == "" {
		fail(w, 400, "INVALID_GRANT", "授权动作或资源无效")
		return
	}
	if (in.Action == "instance.create" || in.Action == "host.manage") && in.Resource.Kind != "node" {
		fail(w, 400, "INVALID_GRANT", "创建与主机权限必须授予节点范围")
		return
	}
	if err := s.store.SetGrantMutation(r.Context(), u.ID, r.Header.Get("Idempotency-Key"), in, r.Method == "DELETE"); err != nil {
		if errors.Is(err, storage.ErrRequestMismatch) {
			fail(w, 409, "IDEMPOTENCY_CONFLICT", err.Error())
			return
		}
		fail(w, 400, "INVALID_GRANT", err.Error())
		return
	}
	s.revokeStreams(in.UserID)
	reply(w, 200, map[string]bool{"updated": true})
}
func (s *Server) revokeStreams(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, cancel := range s.userStreams[userID] {
		cancel()
	}
	delete(s.userStreams, userID)
}
func (s *Server) nodes(w http.ResponseWriter, r *http.Request, u model.User) {
	items, err := s.store.Nodes(r.Context())
	if err != nil {
		fail(w, 500, "STORAGE_ERROR", err.Error())
		return
	}
	out := []model.Node{}
	create := strings.HasSuffix(r.URL.Path, "/creatable")
	for _, n := range items {
		action := "node.read"
		if create {
			action = "instance.create"
		}
		if strings.HasSuffix(r.URL.Path, "/managed") {
			action = "host.manage"
		}
		if s.store.Allowed(r.Context(), u, model.ResourceRef{Kind: "node", ID: n.ID}, action) {
			s.mu.Lock()
			_, connected := s.peers[n.ID]
			s.mu.Unlock()
			if !connected && n.State != "MAINTENANCE" {
				n.State = "OFFLINE"
			} else if connected && !n.Maintenance {
				n.State = "ONLINE"
			}
			if create && (!connected || n.State == "MAINTENANCE") {
				continue
			}
			out = append(out, n)
		}
	}
	reply(w, 200, map[string]any{"items": out})
}
func (s *Server) newEnrollment(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.admin(w, u) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var in struct {
		Name string `json:"name"`
	}
	if !decode(w, r, &in) {
		return
	}
	if len(in.Name) < 1 || len(in.Name) > 80 {
		fail(w, 400, "INVALID_REQUEST", "请输入节点名称")
		return
	}
	token, err := s.store.EnrollmentMutation(r.Context(), u.ID, r.Header.Get("Idempotency-Key"), in.Name)
	if errors.Is(err, storage.ErrRequestMismatch) {
		fail(w, 409, "IDEMPOTENCY_CONFLICT", err.Error())
		return
	}
	if err != nil {
		fail(w, 500, "STORAGE_ERROR", err.Error())
		return
	}
	reply(w, 201, map[string]any{"token": token, "expiresIn": 600})
}
func (s *Server) revokeNode(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.admin(w, u) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	id := r.PathValue("id")
	node, err := s.store.RevokeNode(r.Context(), u.ID, id, r.Header.Get("Idempotency-Key"))
	if err != nil {
		administrationError(w, err)
		return
	}
	s.disconnectNode(id)
	reply(w, 200, map[string]any{"revoked": true, "nodeId": id, "generation": node.Generation})
}
func instanceRef(i model.Instance) model.ResourceRef {
	return model.ResourceRef{Kind: "instance", ID: i.ID, NodeID: i.NodeID}
}
func (s *Server) instances(w http.ResponseWriter, r *http.Request, u model.User) {
	items, err := s.store.Instances(r.Context())
	if err != nil {
		fail(w, 500, "STORAGE_ERROR", err.Error())
		return
	}
	out := []model.Instance{}
	stats := map[string]int{}
	for _, i := range items {
		if !s.store.Allowed(r.Context(), u, instanceRef(i), "instance.read") {
			continue
		}
		q := r.URL.Query()
		if node := q.Get("nodeId"); node != "" && node != i.NodeID {
			continue
		}
		if text := q.Get("search"); text != "" && !strings.Contains(strings.ToLower(i.Name), strings.ToLower(text)) {
			continue
		}
		if state := q.Get("state"); state != "" && state != i.State {
			continue
		}
		if group := q.Get("group"); group != "" && group != i.Group {
			continue
		}
		if tag := q.Get("tag"); tag != "" {
			matched := false
			for _, item := range i.Tags {
				if item == tag {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		s.nodePresentation(r.Context(), &i)
		out = append(out, i)
		stats[i.State]++
	}
	reply(w, 200, map[string]any{"items": out, "statistics": stats})
}
func (s *Server) instance(w http.ResponseWriter, r *http.Request, u model.User) {
	i, err := s.store.Instance(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, 404, "NOT_FOUND", "实例不存在")
		return
	}
	if !s.allowed(w, r, u, instanceRef(i), "instance.read") {
		return
	}
	s.nodePresentation(r.Context(), &i)
	reply(w, 200, map[string]any{"instance": i})
}

func (s *Server) deleteInstance(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.admin(w, u) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	deleted, err := s.store.DeleteInstance(r.Context(), u.ID, r.PathValue("id"), r.Header.Get("Idempotency-Key"))
	if err != nil {
		administrationError(w, err)
		return
	}
	reply(w, 200, map[string]any{"deleted": true, "instanceId": deleted.ID})
}
func (s *Server) createInstance(w http.ResponseWriter, r *http.Request, u model.User) {
	if !requireRequestID(w, r) {
		return
	}
	var in struct {
		NodeID string               `json:"nodeId"`
		Name   string               `json:"name"`
		Config model.InstanceConfig `json:"config"`
	}
	in.Config.Escalate = true
	if !decode(w, r, &in) {
		return
	}
	if len(in.Name) < 1 || len(in.Name) > 120 || len(in.Config.Command) == 0 || (in.Config.Mode != "native" && in.Config.Mode != "container") {
		fail(w, 400, "INVALID_CONFIG", "需要名称、命令和运行模式 native/container")
		return
	}
	if err := model.NormalizeInstanceConfig(&in.Config); err != nil {
		fail(w, 400, "INVALID_CONFIG", err.Error())
		return
	}
	if !s.allowed(w, r, u, model.ResourceRef{Kind: "node", ID: in.NodeID}, "instance.create") {
		return
	}
	n, _, err := s.store.Node(r.Context(), in.NodeID)
	if err != nil {
		fail(w, 404, "NODE_NOT_FOUND", "节点不存在")
		return
	}
	s.mu.Lock()
	_, online := s.peers[in.NodeID]
	s.mu.Unlock()
	if !online {
		fail(w, 409, "NODE_UNREACHABLE", "节点离线，无法核验运行能力")
		return
	}
	if in.Config.Mode == "container" && n.Capabilities["container"] != "available" {
		fail(w, 409, "CAPABILITY_UNAVAILABLE", "节点没有可用容器运行时")
		return
	}
	if in.Config.StopSeconds < 1 {
		in.Config.StopSeconds = 30
	}
	if in.Config.KillSeconds < 1 {
		in.Config.KillSeconds = 10
	}
	if in.Config.StopSeconds > 600 || in.Config.KillSeconds > 120 {
		fail(w, 400, "INVALID_CONFIG", "停止期限超出上限")
		return
	}
	i, err := s.store.CreateInstance(r.Context(), u, model.Instance{NodeID: in.NodeID, Name: in.Name, Config: in.Config, Tags: []string{}}, r.Header.Get("Idempotency-Key"))
	if errors.Is(err, storage.ErrQuotaExceeded) || errors.Is(err, storage.ErrNodeUnavailable) {
		fail(w, 409, "NODE_CAPACITY_CONFLICT", err.Error())
		return
	}
	if errors.Is(err, storage.ErrRequestMismatch) {
		fail(w, 409, "IDEMPOTENCY_CONFLICT", err.Error())
		return
	}
	if err != nil {
		fail(w, 403, "CREATE_DENIED", err.Error())
		return
	}
	reply(w, 201, map[string]any{"instance": i})
}
func (s *Server) instanceAction(w http.ResponseWriter, r *http.Request, u model.User) {
	i, err := s.store.Instance(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, 404, "NOT_FOUND", "实例不存在")
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var in struct {
		Action string `json:"action"`
	}
	if !decode(w, r, &in) {
		return
	}
	if in.Action != "start" && in.Action != "stop" && in.Action != "restart" && in.Action != "kill" {
		fail(w, 400, "INVALID_ACTION", "未知实例操作")
		return
	}
	action := "instance." + in.Action
	if !s.allowed(w, r, u, instanceRef(i), action) {
		return
	}
	if in.Action == "start" || in.Action == "restart" {
		n, _, err := s.store.Node(r.Context(), i.NodeID)
		if err != nil || n.Maintenance {
			fail(w, 409, "NODE_MAINTENANCE", "节点不可接受新启动；停止及已有运行不受影响")
			return
		}
	}
	request := r.Header.Get("Idempotency-Key")
	payload, _ := json.Marshal(i.Config)
	task, _, err := s.store.Accept(r.Context(), model.Task{ActorID: u.ID, RequestID: request, Resource: instanceRef(i), Action: action, Payload: payload})
	if errors.Is(err, storage.ErrQueueLimit) {
		fail(w, 429, "QUEUE_LIMIT", err.Error())
		return
	}
	if errors.Is(err, storage.ErrConflict) {
		fail(w, 409, "TASK_CONFLICT", "实例配置或身份已变更，未接受新的控制任务")
		return
	}
	if errors.Is(err, storage.ErrRequestMismatch) {
		fail(w, 409, "IDEMPOTENCY_CONFLICT", err.Error())
		return
	}
	if err != nil {
		fail(w, 500, "STORAGE_ERROR", err.Error())
		return
	}
	s.mu.Lock()
	_, online := s.peers[i.NodeID]
	s.mu.Unlock()
	if !online && !task.State.Terminal() && task.State != model.WaitingNode {
		if updated, err := s.store.WaitingNode(r.Context(), task.ID, task.Revision); err == nil {
			task = updated
		}
	}
	reply(w, 202, map[string]any{"task": task})
}

func (s *Server) nodePresentation(ctx context.Context, i *model.Instance) {
	n, _, err := s.store.Node(ctx, i.NodeID)
	if err != nil {
		i.NodeState = "REMOVED"
		return
	}
	i.NodeName = n.Name
	s.mu.Lock()
	_, online := s.peers[i.NodeID]
	s.mu.Unlock()
	i.NodeState = n.State
	if n.Maintenance {
		i.NodeState = "MAINTENANCE"
	} else if !online {
		i.NodeState = "OFFLINE"
	} else {
		i.NodeState = "ONLINE"
	}
}
func (s *Server) canTask(ctx context.Context, u model.User, t model.Task) bool {
	if t.Action == "extension.task" {
		var p model.ExtensionTaskPayload
		if json.Unmarshal(t.Payload, &p) != nil || p.ExtensionID == "" {
			return false
		}
		return (u.Admin || t.ActorID == u.ID) && s.store.Allowed(ctx, u, model.ResourceRef{Kind: "extension", ID: p.ExtensionID}, "app.use") && s.store.Allowed(ctx, u, t.Resource, "node.read")
	}
	if t.Resource.Kind == "node" {
		return (u.Admin || t.ActorID == u.ID) && s.store.Allowed(ctx, u, t.Resource, "host.manage")
	}
	return (u.Admin || t.ActorID == u.ID) && s.store.Allowed(ctx, u, t.Resource, "instance.read")
}
func (s *Server) tasks(w http.ResponseWriter, r *http.Request, u model.User) {
	if r.URL.Query().Has("before") {
		s.taskHistory(w, r, u)
		return
	}
	items, err := s.store.RootTasks(r.Context())
	if err != nil {
		fail(w, 500, "STORAGE_ERROR", err.Error())
		return
	}
	out := []model.Task{}
	for _, t := range items {
		if s.canTask(r.Context(), u, t) {
			out = append(out, t)
		}
	}
	offset, limit := 0, 1000
	if raw := r.URL.Query().Get("offset"); raw != "" {
		offset, err = strconv.Atoi(raw)
		if err != nil || offset < 0 {
			fail(w, 400, "INVALID_PAGE", "任务分页偏移无效")
			return
		}
	}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 1000 {
			fail(w, 400, "INVALID_PAGE", "任务分页大小应为1～1000")
			return
		}
	}
	if offset > len(out) {
		offset = len(out)
	}
	end := min(offset+limit, len(out))
	next := -1
	if end < len(out) {
		next = end
	}
	reply(w, 200, map[string]any{"items": out[offset:end], "nextOffset": next})
}
func (s *Server) task(w http.ResponseWriter, r *http.Request, u model.User) {
	t, err := s.store.Task(r.Context(), r.PathValue("id"))
	if err != nil || !s.canTask(r.Context(), u, t) {
		fail(w, 404, "NOT_FOUND", "任务不存在或未授权")
		return
	}
	reply(w, 200, map[string]any{"task": t})
}
func (s *Server) cancelTask(w http.ResponseWriter, r *http.Request, u model.User) {
	t, err := s.store.Task(r.Context(), r.PathValue("id"))
	if err != nil || !s.canTask(r.Context(), u, t) {
		fail(w, 404, "NOT_FOUND", "任务不存在或未授权")
		return
	}
	if t.Action != "extension.task" && !s.allowed(w, r, u, t.Resource, taskPermission(t.Action)) {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	if t.Action == "extension.task" && t.CancellationRequested {
		reply(w, 202, map[string]any{"task": t})
		return
	}
	if t.Action == "file.save" || t.Action == "file.upload" {
		s.cancelUploadTask(w, r, u, t)
		return
	}
	if isLocalTask(t.Action) {
		s.cancelTransferTask(w, r, u, t)
		return
	}
	t, err = s.store.RequestCancel(r.Context(), t.ID, r.Header.Get("Idempotency-Key"))
	if err != nil {
		fail(w, 409, "TASK_CONFLICT", err.Error())
		return
	}
	reply(w, 202, map[string]any{"task": t})
}

// retryTask creates a related attempt only for instance lifecycle actions.
// Those actions rebuild their payload from the current authoritative instance
// configuration, so a retry cannot replay stale credentials or a previous run
// identity. File, terminal-input, transfer, backup and host operations keep
// their action-specific retry/compensation contracts and are deliberately not
// replayed by the generic task center.
func (s *Server) retryTask(w http.ResponseWriter, r *http.Request, u model.User) {
	requestID := r.Header.Get("Idempotency-Key")
	original, err := s.store.Task(r.Context(), r.PathValue("id"))
	if err != nil || !s.canTask(r.Context(), u, original) {
		fail(w, 404, "NOT_FOUND", "任务不存在或未授权")
		return
	}
	if original.State != model.Failed && original.State != model.Interrupted {
		fail(w, 409, "TASK_CONFLICT", "只有失败或中断的任务可以重试")
		return
	}
	switch original.Action {
	case "instance.start", "instance.stop", "instance.restart", "instance.kill":
		// supported below
	default:
		fail(w, 400, "CAPABILITY_UNAVAILABLE", "此任务由所属应用处理重试")
		return
	}
	i, err := s.store.Instance(r.Context(), original.Resource.ID)
	if err != nil || instanceRef(i) != original.Resource {
		fail(w, 409, "TASK_CONFLICT", "实例已删除或资源身份已改变")
		return
	}
	if !s.allowed(w, r, u, instanceRef(i), original.Action) {
		return
	}
	if original.Action == "instance.start" || original.Action == "instance.restart" {
		n, _, nodeErr := s.store.Node(r.Context(), i.NodeID)
		if nodeErr != nil || n.Maintenance {
			fail(w, 409, "NODE_MAINTENANCE", "节点不可接受新启动或重启")
			return
		}
	}
	payload, err := json.Marshal(i.Config)
	if err != nil {
		fail(w, 500, "STORAGE_ERROR", err.Error())
		return
	}
	retry := model.Task{ActorID: u.ID, RetryOf: original.ID, RequestID: requestID, Resource: instanceRef(i), Action: original.Action, Payload: payload}
	if !s.authorizeTask(r.Context(), u, retry) {
		fail(w, 403, "FORBIDDEN", "当前账号不能重试此实例操作")
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	task, created, err := s.store.Accept(r.Context(), retry)
	if errors.Is(err, storage.ErrRequestMismatch) {
		fail(w, 409, "REQUEST_ID_MISMATCH", err.Error())
		return
	}
	if err != nil {
		taskError(w, err)
		return
	}
	if !created && task.RetryOf != original.ID {
		fail(w, 409, "REQUEST_ID_MISMATCH", "请求键已用于其他任务")
		return
	}
	_ = s.store.Audit(r.Context(), u.ID, instanceRef(i), "task.retry", requestID, original.ID)
	status := http.StatusAccepted
	if !created {
		status = http.StatusOK
	}
	reply(w, status, map[string]any{"task": task})
}
func (s *Server) Close() error {
	s.closeSchedules()
	s.closeTransfers()
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.peers {
		_ = p.conn.Close()
	}
	for _, link := range s.links {
		_ = link.Close()
	}
	for _, streams := range s.userStreams {
		for _, cancel := range streams {
			cancel()
		}
	}
	return nil
}

var _ = sql.ErrNoRows

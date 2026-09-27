package master

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"path"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/filesystem"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/storage"
)

const transferManifestNS = "transfer_manifest_v1"
const transferProgressNS = "transfer_progress_v1"
const transferMaxEntries = 4096
const transferMaxManifest = 4 << 20
const transferMaxBytes int64 = 4 << 30
const transferMaxFile int64 = 1 << 30
const transferChunksPerStep = 16

type transferEntry struct {
	Relative      string    `json:"relative"`
	Kind          string    `json:"kind"`
	Size          int64     `json:"size"`
	Version       string    `json:"version"`
	TargetVersion string    `json:"targetVersion"`
	Mode          uint32    `json:"mode"`
	Modified      time.Time `json:"modified"`
}
type transferManifest struct {
	Entries        []transferEntry          `json:"entries"`
	Total          int64                    `json:"total"`
	SourceRelation filesystem.RelationFacts `json:"sourceRelation"`
	TargetRelation filesystem.RelationFacts `json:"targetRelation"`
}
type transferProgress struct {
	model.TransferResult
	Cursor          int             `json:"cursor"`
	Verify          int             `json:"verify"`
	Attempt         int             `json:"attempt"`
	FinalState      model.TaskState `json:"finalState,omitempty"`
	RetryAfter      time.Time       `json:"retryAfter,omitempty"`
	UploadID        string          `json:"uploadId,omitempty"`
	DeleteID        string          `json:"deleteId,omitempty"`
	ChildAction     string          `json:"childAction,omitempty"`
	Metadata        int             `json:"metadata"`
	MetadataVersion string          `json:"metadataVersion,omitempty"`
}

// Only these tasks are executed by Master. The normal node dispatcher and its
// disconnect handler must skip them, including during node reconciliation.
func isLocalTask(action string) bool { return action == "transfer.copy" || action == "transfer.move" }

type transferCoordinator struct {
	s      *Server
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
	active map[string]context.CancelFunc
	wg     sync.WaitGroup
	next   int
}

func (s *Server) registerTransfers() {
	s.mux.HandleFunc("POST /api/v1/transfers", s.auth(s.createTransfer))
	s.mux.HandleFunc("GET /api/v1/transfers", s.auth(s.listTransfers))
	s.mux.HandleFunc("GET /api/v1/transfers/{id}", s.auth(s.getTransfer))
	s.mux.HandleFunc("GET /api/v1/transfers/{id}/entries", s.auth(s.getTransferEntries))
	s.mux.HandleFunc("POST /api/v1/transfers/{id}/cancel", s.auth(s.cancelTransfer))
	ctx, cancel := context.WithCancel(context.Background())
	c := &transferCoordinator{s: s, ctx: ctx, cancel: cancel, active: map[string]context.CancelFunc{}}
	s.transferJobs = c
	c.wg.Add(1)
	go c.loop()
}

// listTransfers returns transfer roots visible to the caller. Source and
// destination permissions are checked independently for every row; the
// target resource on a transfer task alone is not enough to authorize its
// manifest. Terminal rows remain available for reconciliation/history.
func (s *Server) listTransfers(w http.ResponseWriter, r *http.Request, u model.User) {
	tasks, err := s.store.RootTasks(r.Context())
	if err != nil {
		fail(w, 500, "STORAGE_UNAVAILABLE", err.Error())
		return
	}
	items := make([]model.Task, 0)
	for _, task := range tasks {
		if !isLocalTask(task.Action) || !s.canTask(r.Context(), u, task) {
			continue
		}
		var payload model.TransferTaskPayload
		if json.Unmarshal(task.Payload, &payload) != nil {
			continue
		}
		if !u.Admin && (!s.store.Allowed(r.Context(), u, payload.SourceResource, "file.read") || !s.store.Allowed(r.Context(), u, payload.TargetResource, "file.write")) {
			continue
		}
		items = append(items, task)
	}
	reply(w, 200, map[string]any{"items": items})
}

// Call before taking Server.mu: workers can be waiting for nodeCall's lock.
// Shutdown preserves accepted work and node checkpoints; it is not cancellation.
func (s *Server) closeTransfers() {
	if c := s.transferJobs; c != nil {
		c.cancel()
		c.wg.Wait()
	}
}
func (c *transferCoordinator) loop() {
	defer c.wg.Done()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
		}
		tasks, err := c.s.store.Pending(c.ctx)
		if err != nil || len(tasks) == 0 {
			continue
		}
		// Rotate visits so a long directory or an offline node cannot monopolize
		// both workers. Each visit transfers a bounded number of 64 KiB chunks;
		// the checkpoint and cancellation check still run after every chunk.
		start := c.next % len(tasks)
		for n := 0; n < len(tasks); n++ {
			t := tasks[(start+n)%len(tasks)]
			if !isLocalTask(t.Action) {
				continue
			}
			c.mu.Lock()
			if len(c.active) >= 2 {
				c.mu.Unlock()
				break
			}
			if c.active[t.ID] != nil {
				c.mu.Unlock()
				continue
			}
			ctx, cancel := context.WithCancel(c.ctx)
			c.active[t.ID] = cancel
			c.wg.Add(1)
			c.mu.Unlock()
			go func(t model.Task) {
				defer c.wg.Done()
				defer func() { cancel(); c.mu.Lock(); delete(c.active, t.ID); c.mu.Unlock() }()
				c.step(ctx, t)
			}(t)
			c.next = (start + n + 1) % len(tasks)
		}
	}
}

func transferFailure(code, detail string) error { return &model.APIError{Code: code, Message: detail} }
func transferCode(err error, code string) bool {
	var api *model.APIError
	return errors.As(err, &api) && api.Code == code
}
func transferOffline(err error) bool {
	return transferCode(err, "NODE_UNREACHABLE") || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, protocol.ErrClosed)
}

func (s *Server) transferAuthorize(ctx context.Context, actor string, p model.TransferTaskPayload) error {
	u, err := s.store.User(ctx, actor)
	if err != nil || u.Disabled || !s.store.Allowed(ctx, u, p.SourceResource, "file.read") || !s.store.Allowed(ctx, u, p.TargetResource, "file.write") || (p.Move && !s.store.Allowed(ctx, u, p.SourceResource, "file.write")) {
		return transferFailure("FORBIDDEN", "传输源或目标的文件权限已撤销")
	}
	for _, binding := range []struct {
		ref    model.ResourceRef
		config model.InstanceConfig
	}{{p.SourceResource, p.SourceConfig}, {p.TargetResource, p.TargetConfig}} {
		i, err := s.store.Instance(ctx, binding.ref.ID)
		if err != nil || instanceRef(i) != binding.ref || !reflect.DeepEqual(i.Config, binding.config) {
			return transferFailure("CONFLICT", "传输资源或授权根配置已改变")
		}
	}
	return nil
}

func (s *Server) transferCall(ctx context.Context, t model.Task, p model.TransferTaskPayload, source bool, method string, args, out any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	current, err := s.store.Task(ctx, t.ID)
	if err != nil {
		return err
	}
	if current.CancellationRequested {
		return context.Canceled
	}
	if err = s.transferAuthorize(ctx, t.ActorID, p); err != nil {
		return err
	}
	ref, config := p.TargetResource, p.TargetConfig
	if source {
		ref, config = p.SourceResource, p.SourceConfig
	}
	data, err := json.Marshal(args)
	if err != nil {
		return err
	}
	return s.nodeCall(ctx, ref.NodeID, protocol.ChannelBulk, bridge.Request{Method: method, ActorID: t.ActorID, Resource: ref, Config: &config, Args: data}, out)
}
func (s *Server) transferStat(ctx context.Context, t model.Task, p model.TransferTaskPayload, source bool, name string) (filesystem.Entry, error) {
	var e filesystem.Entry
	err := s.transferCall(ctx, t, p, source, "file.stat", map[string]string{"path": name}, &e)
	if !source && transferCode(err, "NOT_FOUND") {
		return filesystem.Entry{Path: name, Version: filesystem.MissingVersion}, nil
	}
	return e, err
}

func (s *Server) createTransfer(w http.ResponseWriter, r *http.Request, u model.User) {
	if !requireRequestID(w, r) {
		return
	}
	var request model.TransferRequest
	if !decode(w, r, &request) {
		return
	}
	if !filePath(w, request.Source.Path, !request.Move) || !filePath(w, request.Target.Path, true) {
		return
	}
	if !validFileHash(request.Source.Version) || !fileVersionValid(request.Target.Version) {
		fail(w, 400, "VERSION_REQUIRED", "源与目标必须绑定提交时的版本")
		return
	}
	source, err := s.store.Instance(r.Context(), request.Source.InstanceID)
	if err != nil {
		fail(w, 404, "NOT_FOUND", "源实例不存在")
		return
	}
	target, err := s.store.Instance(r.Context(), request.Target.InstanceID)
	if err != nil {
		fail(w, 404, "NOT_FOUND", "目标实例不存在")
		return
	}
	p := model.TransferTaskPayload{TransferRequest: request, SourceResource: instanceRef(source), TargetResource: instanceRef(target), SourceConfig: source.Config, TargetConfig: target.Config}
	if err = s.transferAuthorize(r.Context(), u.ID, p); err != nil {
		bridgeError(w, err)
		return
	}
	payload, _ := json.Marshal(p)
	action := "transfer.copy"
	if request.Move {
		action = "transfer.move"
	}
	t, _, err := s.store.Accept(r.Context(), model.Task{ActorID: u.ID, RequestID: r.Header.Get("Idempotency-Key"), Resource: p.TargetResource, Action: action, Payload: payload})
	if err != nil {
		if errors.Is(err, storage.ErrRequestMismatch) {
			fail(w, 409, "REQUEST_ID_MISMATCH", err.Error())
		} else {
			fail(w, 400, "TRANSFER_REJECTED", err.Error())
		}
		return
	}
	reply(w, http.StatusAccepted, map[string]any{"task": t})
}

func (s *Server) transferForRequest(w http.ResponseWriter, r *http.Request, u model.User, cancel bool) (model.Task, bool) {
	t, err := s.store.Task(r.Context(), r.PathValue("id"))
	if err != nil || !isLocalTask(t.Action) {
		fail(w, 404, "NOT_FOUND", "传输任务不存在")
		return t, false
	}
	if t.ActorID != u.ID && !u.Admin {
		fail(w, 403, "FORBIDDEN", "不能访问其他用户的传输")
		return t, false
	}
	var p model.TransferTaskPayload
	if json.Unmarshal(t.Payload, &p) != nil {
		fail(w, 500, "INVALID_TASK", "传输记录损坏")
		return t, false
	}
	// The owner can cancel and inspect its bounded result after revocation so
	// it can see cleanup confirmation. Manifest names still require read access.
	if !cancel && (!s.store.Allowed(r.Context(), u, p.SourceResource, "file.read") || !s.store.Allowed(r.Context(), u, p.TargetResource, "file.write")) {
		fail(w, 403, "FORBIDDEN", "没有传输源或目标权限")
		return t, false
	}
	return t, true
}
func (s *Server) getTransfer(w http.ResponseWriter, r *http.Request, u model.User) {
	t, ok := s.transferForRequest(w, r, u, true)
	if !ok {
		return
	}
	var p transferProgress
	_, err := s.store.Record(r.Context(), transferProgressNS, t.ID, &p)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		fail(w, 500, "STORAGE_UNAVAILABLE", err.Error())
		return
	}
	if p.Stage == "" {
		p.Stage = "preflight"
	}
	reply(w, 200, map[string]any{"task": t, "transfer": p.TransferResult})
}
func (s *Server) getTransferEntries(w http.ResponseWriter, r *http.Request, u model.User) {
	t, ok := s.transferForRequest(w, r, u, false)
	if !ok {
		return
	}
	offset, limit := 0, 100
	var err error
	if q := r.URL.Query().Get("offset"); q != "" {
		offset, err = strconv.Atoi(q)
	}
	if err != nil || offset < 0 {
		fail(w, 400, "INVALID_PAGE", "偏移无效")
		return
	}
	if q := r.URL.Query().Get("limit"); q != "" {
		limit, err = strconv.Atoi(q)
	}
	if err != nil || limit < 1 || limit > 200 {
		fail(w, 400, "INVALID_PAGE", "分页大小应为1～200")
		return
	}
	var m transferManifest
	_, err = s.store.Record(r.Context(), transferManifestNS, t.ID, &m)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		fail(w, 500, "STORAGE_UNAVAILABLE", err.Error())
		return
	}
	var p transferProgress
	_, err = s.store.Record(r.Context(), transferProgressNS, t.ID, &p)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		fail(w, 500, "STORAGE_UNAVAILABLE", err.Error())
		return
	}
	type entryStatus struct {
		transferEntry
		Status string `json:"status"`
	}
	items := make([]entryStatus, 0, limit)
	for n := min(offset, len(m.Entries)); n < min(offset+limit, len(m.Entries)); n++ {
		state := "pending"
		if n < p.Cursor {
			state = "committed"
			if m.Entries[n].Kind == "directory" && m.Entries[n].TargetVersion != filesystem.MissingVersion {
				state = "existing"
			}
		} else if n == p.Cursor && p.CurrentPath != "" {
			state = p.Stage
		}
		items = append(items, entryStatus{m.Entries[n], state})
	}
	next := min(offset+len(items), len(m.Entries))
	if next == len(m.Entries) {
		next = -1
	}
	reply(w, 200, map[string]any{"items": items, "total": len(m.Entries), "nextOffset": next})
}
func (s *Server) cancelTransfer(w http.ResponseWriter, r *http.Request, u model.User) {
	t, ok := s.transferForRequest(w, r, u, true)
	if !ok {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	s.cancelTransferTask(w, r, u, t)
}
func (s *Server) cancelTransferTask(w http.ResponseWriter, r *http.Request, u model.User, t model.Task) {
	if !isLocalTask(t.Action) || (t.ActorID != u.ID && !u.Admin) {
		fail(w, 403, "FORBIDDEN", "不能取消此传输")
		return
	}
	var err error
	// Local receipt prevents generic cancellation from declaring completion
	// before the destination has acknowledged staging cleanup.
	if !t.State.Terminal() && t.DispatchedAt.IsZero() {
		t, err = s.store.MarkDispatch(r.Context(), t.ID, t.Revision)
		if errors.Is(err, storage.ErrConflict) {
			t, err = s.store.Task(r.Context(), t.ID)
		}
	}
	if err == nil {
		t, err = s.store.RequestCancel(r.Context(), t.ID, r.Header.Get("Idempotency-Key"))
	}
	if err != nil {
		fail(w, 409, "TASK_CONFLICT", err.Error())
		return
	}
	if c := s.transferJobs; c != nil {
		c.mu.Lock()
		if cancel := c.active[t.ID]; cancel != nil {
			cancel()
		}
		c.mu.Unlock()
	}
	reply(w, 202, map[string]any{"task": t})
}

func transferName(root, relative string) string {
	if relative == "." {
		return root
	}
	return path.Join(root, relative)
}

func (c *transferCoordinator) relations(ctx context.Context, t model.Task, p model.TransferTaskPayload, baseline *transferManifest) (filesystem.RelationFacts, filesystem.RelationFacts, error) {
	var source, target filesystem.RelationFacts
	if err := c.s.transferCall(ctx, t, p, true, "file.relation", map[string]string{"path": p.Source.Path}, &source); err != nil {
		return source, target, err
	}
	if err := c.s.transferCall(ctx, t, p, false, "file.relation", map[string]string{"path": p.Target.Path}, &target); err != nil {
		return source, target, err
	}
	for _, facts := range []filesystem.RelationFacts{source, target} {
		if !validFileHash(facts.RootID) || !validFileHash(facts.MachineID) || !validFileHash(facts.PathFingerprint) || len(facts.Ancestors) > 256 {
			return source, target, transferFailure("FILE_TRANSFER_MISMATCH", "节点未提供有效的文件系统关系证明")
		}
	}
	if !source.Exists || !validFileHash(source.ObjectID) {
		return source, target, transferFailure("NOT_FOUND", "传输来源对象已不存在")
	}
	if filesystem.OverlappingRelations(source, target) {
		return source, target, transferFailure("CONFLICT", "源目标是相同文件或具有目录包含关系，不能传输")
	}
	if baseline != nil {
		if baseline.SourceRelation.RootID != source.RootID || baseline.SourceRelation.ObjectID != source.ObjectID || baseline.TargetRelation.RootID != target.RootID {
			return source, target, transferFailure("CONFLICT", "授权根或来源文件系统对象已被替换")
		}
		if baseline.TargetRelation.Exists && len(baseline.Entries) > 0 && baseline.Entries[0].Kind == "directory" && baseline.TargetRelation.ObjectID != target.ObjectID {
			return source, target, transferFailure("CONFLICT", "已有目标对象已被替换")
		}
	}
	return source, target, nil
}

func (c *transferCoordinator) manifest(ctx context.Context, t model.Task, p model.TransferTaskPayload) (transferManifest, error) {
	m := transferManifest{}
	var err error
	m.SourceRelation, m.TargetRelation, err = c.relations(ctx, t, p, nil)
	if err != nil {
		return m, err
	}
	root, err := c.s.transferStat(ctx, t, p, true, p.Source.Path)
	if err != nil {
		return m, err
	}
	if root.Version != p.Source.Version {
		return m, transferFailure("CONFLICT", "来源版本已改变")
	}
	target, err := c.s.transferStat(ctx, t, p, false, p.Target.Path)
	if err != nil {
		return m, err
	}
	if target.Version != p.Target.Version {
		return m, transferFailure("CONFLICT", "目标版本已改变")
	}
	queue := []filesystem.Entry{root}
	for len(queue) > 0 {
		if err = ctx.Err(); err != nil {
			return m, err
		}
		e := queue[0]
		queue = queue[1:]
		rel := "."
		if e.Path != p.Source.Path {
			prefix := p.Source.Path + "/"
			if p.Source.Path == "." {
				prefix = ""
			}
			if !strings.HasPrefix(e.Path, prefix) {
				return m, transferFailure("INVALID_PATH", "节点返回越界目录项")
			}
			rel = strings.TrimPrefix(e.Path, prefix)
		}
		if filesystem.ValidatePath(rel) != nil || strings.Count(rel, "/") > 128 || len(m.Entries)+len(queue) >= transferMaxEntries {
			return m, transferFailure("RESOURCE_LIMIT", "目录清单超过路径或4096项预算")
		}
		if e.Kind != "file" && e.Kind != "directory" {
			return m, transferFailure("FILE_UNSUPPORTED", "传输拒绝链接或特殊文件")
		}
		if e.Kind == "file" && e.Version == "" {
			e, err = c.s.transferStat(ctx, t, p, true, e.Path)
			if err != nil {
				return m, err
			}
		}
		if e.Kind == "file" && (e.Size < 0 || e.Size > transferMaxFile || m.Total > transferMaxBytes-e.Size) {
			return m, transferFailure("RESOURCE_LIMIT", "传输超过单文件1GiB或总计4GiB预算")
		}
		if e.Kind == "file" {
			m.Total += e.Size
		}
		dest := target
		if rel != "." {
			dest, err = c.s.transferStat(ctx, t, p, false, transferName(p.Target.Path, rel))
			if err != nil {
				return m, err
			}
		}
		if dest.Version != filesystem.MissingVersion && dest.Kind != e.Kind {
			return m, transferFailure("CONFLICT", "目标文件与目录类型冲突")
		}
		m.Entries = append(m.Entries, transferEntry{Relative: rel, Kind: e.Kind, Size: e.Size, Version: e.Version, TargetVersion: dest.Version, Mode: e.Mode, Modified: e.Modified})
		if e.Kind == "directory" {
			offset, version := 0, ""
			for {
				var page filesystem.Page
				limit := min(16, max(1, (bridge.MaxFrameBytes-8192)/(6*(len(e.Path)+256)+1024)))
				err = c.s.transferCall(ctx, t, p, true, "file.list", map[string]any{"path": e.Path, "offset": offset, "limit": limit, "version": version, "sort": "name", "order": "asc"}, &page)
				if err != nil {
					return m, err
				}
				if page.Total-offset > transferMaxEntries-len(m.Entries)-len(queue) {
					return m, transferFailure("RESOURCE_LIMIT", "目录清单超过4096项预算")
				}
				if len(m.Entries)+len(queue)+len(page.Items) > transferMaxEntries {
					return m, transferFailure("RESOURCE_LIMIT", "目录清单超过4096项预算")
				}
				for _, child := range page.Items {
					if path.Dir(child.Path) != e.Path {
						return m, transferFailure("INVALID_PATH", "目录页返回非直接子项")
					}
					queue = append(queue, child)
				}
				if queued, encodeErr := json.Marshal(queue); encodeErr != nil || len(queued) > transferMaxManifest {
					return m, transferFailure("RESOURCE_LIMIT", "待处理目录清单超过4MiB预算")
				}
				if page.NextOffset < 0 {
					break
				}
				if page.NextOffset <= offset || len(page.Items) == 0 {
					return m, transferFailure("FILE_TRANSFER_MISMATCH", "目录分页未取得进展")
				}
				offset, version = page.NextOffset, page.Version
			}
		}
		// Names, escaped JSON and recorded metadata count against the manifest
		// budget, not just content bytes. No remote writes precede this check.
		if b, marshalErr := json.Marshal(m); marshalErr != nil || len(b) > transferMaxManifest {
			return m, transferFailure("RESOURCE_LIMIT", "目录清单超过4MiB预算")
		}
	}
	root, err = c.s.transferStat(ctx, t, p, true, p.Source.Path)
	if err != nil {
		return m, err
	}
	if root.Version != p.Source.Version {
		return m, transferFailure("CONFLICT", "枚举期间来源已改变")
	}
	target, err = c.s.transferStat(ctx, t, p, false, p.Target.Path)
	if err != nil {
		return m, err
	}
	if target.Version != p.Target.Version {
		return m, transferFailure("CONFLICT", "枚举期间目标已改变")
	}
	if _, _, err = c.relations(ctx, t, p, &m); err != nil {
		return m, err
	}
	return m, nil
}

func (c *transferCoordinator) persist(ctx context.Context, t model.Task, p *transferProgress, revision *int64, state model.TaskState) error {
	p.Completed, p.Partial = p.Cursor, p.Cursor > 0 && state != model.Succeeded
	rev, err := c.s.store.PutRecord(ctx, transferProgressNS, t.ID, *revision, p)
	if err != nil {
		return err
	}
	*revision = rev
	// A cancellation request can race with the checkpoint. Preserve its flag
	// and state; the next visit reconciles children before terminal cancellation.
	for attempts := 0; attempts < 3; attempts++ {
		current, err := c.s.store.Task(ctx, t.ID)
		if err != nil || current.State.Terminal() {
			return err
		}
		if current.CancellationRequested && state == model.Running {
			state = model.CancelRequested
		}
		data, _ := json.Marshal(p.TransferResult)
		_, err = c.s.store.UpdateTask(ctx, current.ID, current.Revision, state, p.Stage, data, p.Error)
		if errors.Is(err, storage.ErrConflict) {
			continue
		}
		return err
	}
	return storage.ErrConflict
}

func (c *transferCoordinator) step(ctx context.Context, t model.Task) {
	var payload model.TransferTaskPayload
	if json.Unmarshal(t.Payload, &payload) != nil {
		return
	}
	var p transferProgress
	revision, err := c.s.store.Record(ctx, transferProgressNS, t.ID, &p)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return
	}
	if p.Stage == "" {
		p.Stage = "preflight"
	}
	if time.Now().Before(p.RetryAfter) && !t.CancellationRequested {
		return
	}
	p.RetryAfter = time.Time{}
	if t.DispatchedAt.IsZero() {
		t, err = c.s.store.MarkDispatch(ctx, t.ID, t.Revision)
		if err != nil {
			return
		}
	}
	if t.CancellationRequested && p.FinalState == "" {
		p.FinalState, p.Error = model.Cancelled, "用户请求取消；已提交的目标保留"
	}
	var m transferManifest
	_, err = c.s.store.Record(ctx, transferManifestNS, t.ID, &m)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return
	}
	if p.FinalState != "" {
		err = c.cleanup(ctx, t, payload, &p)
	} else if errors.Is(err, sql.ErrNoRows) {
		if err = c.persist(ctx, t, &p, &revision, model.Running); err != nil {
			return
		}
		m, err = c.manifest(ctx, t, payload)
		if err == nil {
			_, err = c.s.store.PutRecord(ctx, transferManifestNS, t.ID, 0, m)
			if err == nil {
				p.Entries, p.Total, p.Stage = len(m.Entries), m.Total, "copy"
			}
		}
	} else {
		p.Entries, p.Total = len(m.Entries), m.Total
		if p.Stage == "preflight" {
			p.Stage = "copy"
		}
		if err = c.persist(ctx, t, &p, &revision, model.Running); err != nil {
			return
		}
		err = c.advance(ctx, t, payload, m, &p, &revision)
	}
	if ctx.Err() != nil {
		return
	} // Close/cancel leaves the last confirmed checkpoint.
	state := model.Running
	if err != nil {
		if transferOffline(err) {
			state, p.RetryAfter = model.WaitingNode, time.Now().Add(time.Second)
			p.Error = "节点不可达；保留已确认检查点，等待重连"
		} else if errors.Is(err, context.Canceled) {
			return // next visit reloads the durable cancellation flag
		} else {
			p.FinalState, p.Error, p.Stage, p.CleanupPending = model.Failed, err.Error(), "cleanup", true
			if transferCode(err, "OUTCOME_UNKNOWN") {
				p.FinalState = model.Interrupted
			}
		}
	} else if p.Stage == "complete" {
		state, p.Partial, p.Error = model.Succeeded, false, ""
	} else if p.Stage == "cleaned" {
		state = p.FinalState
	}
	_ = c.persist(ctx, t, &p, &revision, state)
}

func transferChildKey(t model.Task, cursor, attempt int, action string) string {
	return "transfer:" + t.ID + ":" + strconv.Itoa(cursor) + ":" + action + ":" + strconv.Itoa(attempt)
}
func (c *transferCoordinator) child(ctx context.Context, t model.Task, payload model.TransferTaskPayload, p *transferProgress, action string, body any, source bool) (model.Task, error) {
	cursor := p.Cursor
	if action == "file.metadata" {
		cursor = -1 - p.Metadata
	}
	key := transferChildKey(t, cursor, p.Attempt, action)
	ref := payload.TargetResource
	if source {
		ref = payload.SourceResource
	}
	data, err := transferChildData(body, t.ID)
	if err != nil {
		return model.Task{}, err
	}
	if old, err := c.s.store.TaskByRequest(ctx, t.ActorID, key); err == nil {
		if old.Action != action || old.Resource != ref || !bytes.Equal(old.Payload, data) {
			return model.Task{}, storage.ErrRequestMismatch
		}
		return old, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return old, err
	}
	if err := c.s.transferAuthorize(ctx, t.ActorID, payload); err != nil {
		return model.Task{}, err
	}
	current, err := c.s.store.Task(ctx, t.ID)
	if err != nil {
		return model.Task{}, err
	}
	if current.CancellationRequested {
		return model.Task{}, context.Canceled
	}
	child, _, err := c.s.store.Accept(ctx, model.Task{ActorID: t.ActorID, RequestID: key, Resource: ref, Action: action, Payload: data})
	return child, err
}

func transferChildData(body any, parentID string) (json.RawMessage, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	var envelope map[string]json.RawMessage
	if err = json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}
	envelope["transferParentId"], _ = json.Marshal(parentID)
	data, err = json.Marshal(envelope)
	return data, err
}

func transferOwnsChild(parent model.Task, child model.Task) bool {
	var binding struct {
		ParentID string `json:"transferParentId"`
	}
	return child.ActorID == parent.ActorID && strings.HasPrefix(child.RequestID, "transfer:"+parent.ID+":") && json.Unmarshal(child.Payload, &binding) == nil && binding.ParentID == parent.ID
}
func transferChildResult(child model.Task) error {
	switch child.State {
	case model.Succeeded:
		return nil
	case model.Interrupted:
		return transferFailure("OUTCOME_UNKNOWN", "子任务结果不明："+child.ID+" "+child.Error)
	case model.Failed, model.Cancelled:
		return transferFailure("CHILD_FAILED", "子任务未成功："+child.ID+" "+child.Error)
	}
	return nil
}

func (c *transferCoordinator) advance(ctx context.Context, t model.Task, payload model.TransferTaskPayload, m transferManifest, p *transferProgress, revision *int64) error {
	if err := c.s.transferAuthorize(ctx, t.ActorID, payload); err != nil {
		return err
	}
	if p.Stage != "delete_source" {
		if _, _, err := c.relations(ctx, t, payload, &m); err != nil {
			return err
		}
	}
	if p.Stage == "copy" && p.Cursor < len(m.Entries) {
		e := m.Entries[p.Cursor]
		p.CurrentPath = transferName(payload.Target.Path, e.Relative)
		if e.Kind == "directory" {
			if e.TargetVersion != filesystem.MissingVersion {
				// Existing directory merge is an explicit baseline. Descendant
				// files still carry their individual overwrite preconditions.
				actual, err := c.s.transferStat(ctx, t, payload, false, p.CurrentPath)
				if err != nil {
					return err
				}
				if actual.Kind != "directory" || actual.Version != e.TargetVersion {
					return transferFailure("CONFLICT", "目标目录基线已改变")
				}
				p.Cursor++
				p.CurrentPath = ""
				return nil
			}
			p.ChildAction = "file.mkdir"
			if err := c.persist(ctx, t, p, revision, model.Running); err != nil {
				return err
			}
			child, err := c.child(ctx, t, payload, p, "file.mkdir", model.FileTaskPayload{Config: payload.TargetConfig, Path: p.CurrentPath, Version: filesystem.MissingVersion}, false)
			if err != nil {
				return err
			}
			if transferOwnsChild(t, child) {
				p.ChildTaskID = child.ID
			}
			if child.State.Terminal() {
				if err = transferChildResult(child); err != nil {
					return err
				}
				p.Cursor++
				p.ChildTaskID, p.ChildAction, p.CurrentPath = "", "", ""
			}
			return nil
		}
		spec := filesystem.UploadSpec{ID: fileHash([]byte(t.ID + ":" + strconv.Itoa(p.Cursor)))[7:39], OwnerID: t.ActorID, Path: p.CurrentPath, Total: e.Size, Hash: e.Version, ExpectedVersion: e.TargetVersion, SourceName: path.Base(transferName(payload.Source.Path, e.Relative)), SourceModified: e.Modified.UnixMilli(), SourceFingerprint: e.Version, SourceNodeID: payload.SourceResource.NodeID, SourceResourceID: payload.SourceResource.ID, SourcePath: transferName(payload.Source.Path, e.Relative), SourceVersion: e.Version}
		spec.PreserveMetadata, spec.SourceMode, spec.SourceModifiedNano = true, e.Mode, e.Modified.UnixNano()
		var source filesystem.Entry
		err := c.s.transferCall(ctx, t, payload, true, "file.transfer.stat", map[string]any{"path": spec.SourcePath, "version": e.Version}, &source)
		if err != nil {
			return err
		}
		if source.Kind != "file" || source.Size != e.Size || source.Version != e.Version || source.Mode != e.Mode || !source.Modified.Equal(e.Modified) {
			return transferFailure("CONFLICT", "来源内容或元信息已改变")
		}
		p.UploadID = spec.ID
		p.ChildAction = "file.upload"
		// Persist ownership before Begin: even a lost Begin response or Master
		// crash leaves a precise upload ID to reconcile/cancel after restart.
		if err := c.persist(ctx, t, p, revision, model.Running); err != nil {
			return err
		}
		var upload filesystem.Upload
		if err := c.s.transferCall(ctx, t, payload, false, "file.upload.begin", map[string]any{"spec": spec}, &upload); err != nil {
			return err
		}
		if upload.Spec != spec || upload.Offset < 0 || upload.Offset > spec.Total || upload.ChunkBytes != fileChunkSize || upload.Stage == "cancelled" {
			return transferFailure("FILE_TRANSFER_MISMATCH", "目标检查点不匹配传输身份")
		}
		p.CurrentOffset = upload.Offset
		for chunks := 0; upload.Offset < spec.Total && chunks < transferChunksPerStep; chunks++ {
			var chunk filesystem.Chunk
			if err := c.s.transferCall(ctx, t, payload, true, "file.transfer.chunk", map[string]any{"path": spec.SourcePath, "version": e.Version, "offset": upload.Offset, "limit": fileChunkSize}, &chunk); err != nil {
				return err
			}
			expectedLength := min(int64(fileChunkSize), spec.Total-upload.Offset)
			if chunk.Path != spec.SourcePath || chunk.Version != e.Version || chunk.Offset != upload.Offset || chunk.Total != spec.Total || int64(len(chunk.Data)) != expectedLength || chunk.Hash != fileHash(chunk.Data) {
				return transferFailure("FILE_TRANSFER_MISMATCH", "源分片身份、长度或校验和不匹配")
			}
			var ack filesystem.Upload
			if err := c.s.transferCall(ctx, t, payload, false, "file.upload.chunk", map[string]any{"id": spec.ID, "offset": chunk.Offset, "data": chunk.Data, "hash": chunk.Hash}, &ack); err != nil {
				return err
			}
			if ack.Spec != spec || ack.Offset != upload.Offset+expectedLength || ack.LastChunkHash != chunk.Hash {
				return transferFailure("FILE_TRANSFER_MISMATCH", "节点分片确认不匹配")
			}
			upload = ack
			p.CurrentOffset, p.Error = ack.Offset, ""
			if err := c.persist(ctx, t, p, revision, model.Running); err != nil {
				return err
			}
		}
		if upload.Offset < spec.Total {
			return nil
		}
		// The fast block proof guarantees version-bound bytes. Revalidate the
		// complete live source before publishing the verified destination, so a
		// change outside the last requested block cannot be reported as success.
		source, err = c.s.transferStat(ctx, t, payload, true, spec.SourcePath)
		if err != nil {
			return err
		}
		if source.Kind != "file" || source.Size != e.Size || source.Version != e.Version || source.Mode != e.Mode || !source.Modified.Equal(e.Modified) {
			return transferFailure("CONFLICT", "来源在发布目标前改变")
		}
		body := filePayload{FileTaskPayload: model.FileTaskPayload{Config: payload.TargetConfig, Path: spec.Path, Version: spec.ExpectedVersion, UploadID: spec.ID, Hash: spec.Hash, Total: spec.Total}, UploadSpec: &spec}
		child, err := c.child(ctx, t, payload, p, "file.upload", body, false)
		if err != nil {
			return err
		}
		p.ChildTaskID = child.ID
		if child.State == model.Interrupted {
			// CommitUpload reconciles its durable intent and is idempotent. A new
			// explicit durable attempt certifies it after a Daemon crash. No such
			// replay is permitted for mkdir or source deletion.
			if p.Attempt >= 16 {
				return transferFailure("OUTCOME_UNKNOWN", "上传提交重试预算耗尽")
			}
			p.Attempt++
			p.ChildTaskID = ""
			return nil
		}
		if child.State.Terminal() {
			if err = transferChildResult(child); err != nil {
				return err
			}
			var result filesystem.Upload
			if json.Unmarshal(child.Result, &result) != nil || result.Stage != "committed" || result.Spec != spec || result.Version != e.Version {
				return transferFailure("FILE_TRANSFER_MISMATCH", "持久提交任务的结果不匹配")
			}
			p.Cursor++
			p.CommittedBytes += e.Size
			p.ChildTaskID, p.UploadID, p.CurrentPath, p.CurrentOffset, p.Attempt = "", "", "", 0, 0
			p.ChildAction = ""
		}
		return nil
	}
	if p.Stage == "copy" {
		p.Stage, p.CurrentPath = "verify_destination", ""
	}
	if p.Stage == "verify_destination" {
		for checks := 0; p.Verify < len(m.Entries) && checks < 8; checks++ {
			e := m.Entries[p.Verify]
			actual, err := c.s.transferStat(ctx, t, payload, false, transferName(payload.Target.Path, e.Relative))
			if err != nil {
				return err
			}
			if actual.Kind != e.Kind || (e.Kind == "file" && (actual.Version != e.Version || actual.Size != e.Size)) {
				return transferFailure("CONFLICT", "目标提交后的校验失败；保留来源")
			}
			p.Verify++
		}
		if p.Verify < len(m.Entries) {
			return nil
		}
		p.DestinationVerified, p.Stage = true, "metadata"
		return nil
	}
	if p.Stage == "metadata" {
		// Reverse breadth-first order leaves parent timestamps until all child
		// creation is complete. Existing target directories retain their metadata.
		for p.Metadata < len(m.Entries) {
			e := m.Entries[len(m.Entries)-1-p.Metadata]
			if e.Kind != "directory" || e.TargetVersion != filesystem.MissingVersion {
				p.Metadata++
				continue
			}
			p.CurrentPath = transferName(payload.Target.Path, e.Relative)
			if p.MetadataVersion == "" {
				actual, err := c.s.transferStat(ctx, t, payload, false, p.CurrentPath)
				if err != nil {
					return err
				}
				if actual.Kind != "directory" {
					return transferFailure("CONFLICT", "元信息提交前目录已改变")
				}
				p.MetadataVersion = actual.Version
			}
			p.ChildAction = "file.metadata"
			if err := c.persist(ctx, t, p, revision, model.Running); err != nil {
				return err
			}
			body := struct {
				model.FileTaskPayload
				Mode     uint32    `json:"mode"`
				Modified time.Time `json:"modified"`
			}{model.FileTaskPayload{Config: payload.TargetConfig, Path: p.CurrentPath, Version: p.MetadataVersion}, e.Mode, e.Modified}
			child, err := c.child(ctx, t, payload, p, "file.metadata", body, false)
			if err != nil {
				return err
			}
			p.ChildTaskID = child.ID
			if child.State.Terminal() {
				if err = transferChildResult(child); err != nil {
					return err
				}
				p.Metadata++
				p.ChildTaskID, p.ChildAction, p.MetadataVersion, p.CurrentPath = "", "", "", ""
			}
			return nil
		}
		p.Stage = "verify_source"
		return nil
	}
	if p.Stage == "verify_source" {
		source, err := c.s.transferStat(ctx, t, payload, true, payload.Source.Path)
		if err != nil {
			return err
		}
		if source.Version != payload.Source.Version {
			return transferFailure("CONFLICT", "来源在传输期间改变；目标保留且不删除来源")
		}
		if len(m.Entries) == 1 && m.Entries[0].Kind == "file" {
			e := m.Entries[0]
			if source.Size != e.Size || source.Mode != e.Mode || !source.Modified.Equal(e.Modified) {
				return transferFailure("CONFLICT", "来源元信息在传输期间改变；保留来源")
			}
		}
		if !payload.Move {
			p.Stage = "complete"
			return nil
		}
		p.Stage = "delete_source"
		// Persist destination verification before creating the deletion task.
		if err = c.persist(ctx, t, p, revision, model.Running); err != nil {
			return err
		}
	}
	if p.Stage == "delete_source" {
		if _, err := c.s.store.TaskByRequest(ctx, t.ActorID, transferChildKey(t, p.Cursor, p.Attempt, "file.delete")); errors.Is(err, sql.ErrNoRows) {
			if _, _, err = c.relations(ctx, t, payload, &m); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		p.ChildAction = "file.delete"
		if err := c.persist(ctx, t, p, revision, model.Running); err != nil {
			return err
		}
		body := struct {
			model.FileTaskPayload
			ExpectedObjectID string `json:"expectedObjectId"`
		}{model.FileTaskPayload{Config: payload.SourceConfig, Path: payload.Source.Path, Version: payload.Source.Version}, m.SourceRelation.ObjectID}
		child, err := c.child(ctx, t, payload, p, "file.delete", body, true)
		if err != nil {
			return err
		}
		p.ChildTaskID, p.DeleteID = child.ID, child.ID
		if child.State.Terminal() {
			if err = transferChildResult(child); err != nil {
				p.SourceOutcomeUnknown = child.State == model.Interrupted
				var recycle filesystem.TrashItem
				if json.Unmarshal(child.Result, &recycle) == nil && recycle.ID != "" {
					p.SourceOutcomeUnknown = true
					return transferFailure("OUTCOME_UNKNOWN", "源回收子任务留下未确认记录："+child.ID)
				}
				return err
			}
			p.SourceDeleted, p.Stage = true, "complete"
		}
	}
	return nil
}

func (c *transferCoordinator) cleanup(ctx context.Context, t model.Task, payload model.TransferTaskPayload, p *transferProgress) error {
	p.Stage, p.CleanupPending = "cleanup", true
	// Recover an accepted child even if Master crashed before storing its ID.
	if p.ChildTaskID == "" && p.ChildAction != "" {
		cursor := p.Cursor
		if p.ChildAction == "file.metadata" {
			cursor = -1 - p.Metadata
		}
		if child, err := c.s.store.TaskByRequest(ctx, t.ActorID, transferChildKey(t, cursor, p.Attempt, p.ChildAction)); err == nil {
			p.ChildTaskID = child.ID
		} else if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}
	if p.ChildTaskID != "" {
		child, err := c.s.store.Task(ctx, p.ChildTaskID)
		if err != nil {
			return err
		}
		if !transferOwnsChild(t, child) {
			// A client can choose request IDs, but cannot turn a different task
			// into our completion receipt or cleanup target by guessing our key.
			p.ChildTaskID, p.ChildAction = "", ""
			return nil
		}
		if !child.State.Terminal() {
			if !child.CancellationRequested {
				_, err = c.s.store.RequestCancel(ctx, child.ID)
			}
			return err // wait for the durable node result; cancellation is not rollback
		}
		if child.State == model.Succeeded {
			if child.Action == "file.delete" {
				p.SourceDeleted = true
			} else if child.Action == "file.mkdir" {
				p.Cursor++
			} else if child.Action == "file.upload" {
				var result filesystem.Upload
				if json.Unmarshal(child.Result, &result) != nil || result.Stage != "committed" {
					return transferFailure("OUTCOME_UNKNOWN", "提交子任务结果损坏")
				}
				p.Cursor++
				p.CommittedBytes += result.Spec.Total
				p.UploadID = ""
			}
		} else if child.State == model.Interrupted && child.Action != "file.upload" {
			p.SourceOutcomeUnknown = child.Action == "file.delete"
			p.FinalState = model.Interrupted
			p.Error = "子任务结果不明：" + child.ID + "；已知目标保留"
		}
		p.ChildTaskID = ""
		p.ChildAction = ""
	}
	if p.UploadID != "" {
		// This is strictly cleanup of this task's bound upload ID. Permission
		// revocation must not strand staging, nor authorize any replacement write.
		var status filesystem.Upload
		args, _ := json.Marshal(map[string]string{"id": p.UploadID})
		request := bridge.Request{Method: "file.upload.status", ActorID: t.ActorID, Resource: payload.TargetResource, Config: &payload.TargetConfig, Args: args}
		err := c.s.nodeCall(ctx, payload.TargetResource.NodeID, protocol.ChannelBulk, request, &status)
		if err != nil && !transferCode(err, "NOT_FOUND") {
			return err
		}
		if err == nil {
			if status.Spec.ID != p.UploadID || status.Spec.OwnerID != t.ActorID || status.Spec.Path != p.CurrentPath {
				return transferFailure("FILE_TRANSFER_MISMATCH", "清理检查点不属于此任务")
			}
			if status.Stage == "committed" {
				p.Cursor++
				p.CommittedBytes += status.Spec.Total
			} else if status.Stage != "cancelled" {
				request.Method = "file.upload.cancel"
				if err = c.s.nodeCall(ctx, payload.TargetResource.NodeID, protocol.ChannelBulk, request, &status); err != nil {
					return err
				}
				if status.Stage != "cancelled" {
					return transferFailure("OUTCOME_UNKNOWN", "目标尚未确认清理")
				}
			}
		}
		p.UploadID = ""
	}
	p.Stage, p.CleanupPending, p.CurrentOffset = "cleaned", false, 0
	return nil
}

// The common dispatcher invokes this for every child dispatch/reconcile. An
// already-accepted child cannot retain stale source authority while queued.
func (s *Server) authorizeTransferChild(ctx context.Context, u model.User, t model.Task) bool {
	var binding struct {
		ParentID string `json:"transferParentId"`
	}
	if json.Unmarshal(t.Payload, &binding) != nil {
		return false
	}
	if binding.ParentID == "" {
		return true
	}
	parent, err := s.store.Task(ctx, binding.ParentID)
	if err != nil || !isLocalTask(parent.Action) || parent.ActorID != t.ActorID || !strings.HasPrefix(t.RequestID, "transfer:"+parent.ID+":") {
		return false
	}
	var p model.TransferTaskPayload
	if json.Unmarshal(parent.Payload, &p) != nil || s.transferAuthorize(ctx, u.ID, p) != nil {
		return false
	}
	ref, config := p.TargetResource, p.TargetConfig
	if t.Action == "file.delete" {
		ref, config = p.SourceResource, p.SourceConfig
	}
	var child model.FileTaskPayload
	return t.Resource == ref && json.Unmarshal(t.Payload, &child) == nil && reflect.DeepEqual(child.Config, config)
}

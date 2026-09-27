package master

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/filesystem"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/storage"
)

const maxFileText = 4 << 20
const fileChunkSize = 64 << 10

// Preview chunks stay below the daemon transport frame limit while leaving
// room for a UTF-8 boundary overlap at either side of a requested range.
const maxFilePreview = 60 << 10

type filePayload struct {
	model.FileTaskPayload
	UploadSpec *filesystem.UploadSpec `json:"uploadSpec,omitempty"`
}

func (s *Server) registerFiles() {
	// The budget is per Master, including buffered text. Streaming downloads
	// occupy one slot and use bounded chunks rather than a whole-file buffer.
	slots := make(chan struct{}, 4)
	register := func(pattern string, handler authed) {
		s.mux.HandleFunc(pattern, s.auth(func(w http.ResponseWriter, r *http.Request, u model.User) {
			select {
			case slots <- struct{}{}:
				defer func() { <-slots }()
			case <-r.Context().Done():
				return
			}
			ctx, cancel := context.WithCancel(r.Context())
			id := model.ID()
			s.mu.Lock()
			if s.userStreams[u.ID] == nil {
				s.userStreams[u.ID] = map[string]context.CancelFunc{}
			}
			s.userStreams[u.ID][id] = cancel
			s.mu.Unlock()
			defer func() { cancel(); s.mu.Lock(); delete(s.userStreams[u.ID], id); s.mu.Unlock() }()
			handler(w, r.WithContext(ctx), u)
		}))
	}
	register("GET /api/v1/instances/{id}/files", s.fileList)
	register("GET /api/v1/instances/{id}/files/access", s.fileAccess)
	register("GET /api/v1/instances/{id}/files/stat", s.fileStat)
	register("GET /api/v1/instances/{id}/files/trash", s.fileTrash)
	register("GET /api/v1/instances/{id}/files/preview", s.filePreview)
	register("GET /api/v1/instances/{id}/files/content", s.fileContent)
	register("PUT /api/v1/instances/{id}/files/content", s.fileSave)
	register("GET /api/v1/instances/{id}/files/download", s.fileDownload)
	register("POST /api/v1/instances/{id}/files/actions", s.fileAction)
	register("POST /api/v1/instances/{id}/files/uploads", s.fileUploadBegin)
	register("GET /api/v1/instances/{id}/files/uploads/{taskId}", s.fileUploadStatus)
	register("PUT /api/v1/instances/{id}/files/uploads/{taskId}/chunks", s.fileUploadChunk)
	register("POST /api/v1/instances/{id}/files/uploads/{taskId}/complete", s.fileUploadComplete)
	register("DELETE /api/v1/instances/{id}/files/uploads/{taskId}", s.fileUploadCancel)
}

func fileHash(b []byte) string {
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func validFileHash(value string) bool {
	if len(value) != 71 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(value[7:])
	return err == nil
}
func fileVersionValid(v string) bool { return v == filesystem.MissingVersion || validFileHash(v) }
func (s *Server) fileInstance(w http.ResponseWriter, r *http.Request, u model.User, action string) (model.Instance, bool) {
	i, e := s.store.Instance(r.Context(), r.PathValue("id"))
	if e != nil {
		fail(w, 404, "NOT_FOUND", "实例不存在")
		return i, false
	}
	if !s.allowed(w, r, u, instanceRef(i), action) {
		return i, false
	}
	return i, true
}
func (s *Server) callFile(ctx context.Context, u model.User, i model.Instance, method string, args, out any) error {
	action := "file.read"
	if strings.HasPrefix(method, "file.upload.") {
		action = "file.write"
	}
	current, e := s.store.User(ctx, u.ID)
	if e != nil || !s.store.Allowed(ctx, current, instanceRef(i), action) {
		return &model.APIError{Code: "FORBIDDEN", Message: "文件权限已撤销"}
	}
	b, e := json.Marshal(args)
	if e != nil {
		return e
	}
	return s.nodeCall(ctx, i.NodeID, protocol.ChannelBulk, bridge.Request{Method: method, ActorID: u.ID, Resource: instanceRef(i), Config: &i.Config, Args: b}, out)
}
func filePath(w http.ResponseWriter, p string, allowRoot bool) bool {
	if filesystem.ValidatePath(p) != nil || (!allowRoot && p == ".") {
		fail(w, 400, "INVALID_PATH", "需要授权根内的安全相对路径")
		return false
	}
	return true
}

func (s *Server) fileList(w http.ResponseWriter, r *http.Request, u model.User) {
	i, ok := s.fileInstance(w, r, u, "file.read")
	if !ok {
		return
	}
	p := r.URL.Query().Get("path")
	if p == "" {
		p = "."
	}
	if !filePath(w, p, true) {
		return
	}
	offset, e := strconv.Atoi(r.URL.Query().Get("offset"))
	if r.URL.Query().Get("offset") == "" {
		offset = 0
		e = nil
	}
	if e != nil || offset < 0 {
		fail(w, 400, "INVALID_PAGE", "目录分页偏移无效")
		return
	}
	limit := 200
	if text := r.URL.Query().Get("limit"); text != "" {
		limit, e = strconv.Atoi(text)
		if e != nil || limit < 1 || limit > 1000 {
			fail(w, 400, "INVALID_PAGE", "目录分页大小应为1～1000")
			return
		}
	}
	// Names can require JSON escaping. This cap keeps the node response below
	// the independent data channel frame budget even for long parent paths.
	limit = min(limit, max(1, (bridge.MaxFrameBytes-8192)/(len(p)+2048)))
	var page filesystem.Page
	if e = s.callFile(r.Context(), u, i, "file.list", map[string]any{"path": p, "offset": offset, "limit": limit, "version": r.URL.Query().Get("version"), "search": r.URL.Query().Get("search"), "sort": r.URL.Query().Get("sort"), "order": r.URL.Query().Get("order")}, &page); e != nil {
		bridgeError(w, e)
		return
	}
	type row struct {
		filesystem.Entry
		IsDir bool `json:"isDir"`
	}
	items := make([]row, 0, len(page.Items))
	for _, entry := range page.Items {
		items = append(items, row{entry, entry.Kind == "directory"})
	}
	reply(w, 200, map[string]any{"items": items, "version": page.Version, "total": page.Total, "nextOffset": page.NextOffset})
}
func (s *Server) fileStat(w http.ResponseWriter, r *http.Request, u model.User) {
	i, ok := s.fileInstance(w, r, u, "file.read")
	if !ok {
		return
	}
	p := r.URL.Query().Get("path")
	if !filePath(w, p, true) {
		return
	}
	var entry filesystem.Entry
	if e := s.callFile(r.Context(), u, i, "file.stat", map[string]string{"path": p}, &entry); e != nil {
		bridgeError(w, e)
		return
	}
	reply(w, 200, entry)
}

func (s *Server) fileAccess(w http.ResponseWriter, r *http.Request, u model.User) {
	i, ok := s.fileInstance(w, r, u, "file.read")
	if !ok {
		return
	}
	s.nodePresentation(r.Context(), &i)
	reply(w, 200, map[string]any{"instanceId": i.ID, "nodeId": i.NodeID, "nodeName": i.NodeName, "nodeState": i.NodeState, "canRead": true, "canWrite": s.store.Allowed(r.Context(), u, instanceRef(i), "file.write")})
}

func (s *Server) fileTrash(w http.ResponseWriter, r *http.Request, u model.User) {
	i, ok := s.fileInstance(w, r, u, "file.read")
	if !ok {
		return
	}
	offset, e := strconv.Atoi(r.URL.Query().Get("offset"))
	if r.URL.Query().Get("offset") == "" {
		offset = 0
		e = nil
	}
	if e != nil || offset < 0 {
		fail(w, 400, "INVALID_PAGE", "回收列表偏移无效")
		return
	}
	limit := 10
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, e = strconv.Atoi(raw)
		if e != nil || limit < 1 || limit > 1000 {
			fail(w, 400, "INVALID_PAGE", "回收分页大小无效")
			return
		}
	}
	limit = min(limit, 10)
	var page filesystem.TrashPage
	if e = s.callFile(r.Context(), u, i, "file.trash", map[string]int{"offset": offset, "limit": limit}, &page); e != nil {
		bridgeError(w, e)
		return
	}
	reply(w, 200, page)
}
func (s *Server) readFileChunks(ctx context.Context, u model.User, i model.Instance, p, v string, total int64, consume func([]byte) error) error {
	for offset := int64(0); offset < total; {
		var chunk filesystem.Chunk
		if e := s.callFile(ctx, u, i, "file.chunk", map[string]any{"path": p, "version": v, "offset": offset, "limit": fileChunkSize}, &chunk); e != nil {
			return e
		}
		if chunk.Path != p || chunk.Version != v || chunk.Total != total || chunk.Offset != offset || len(chunk.Data) == 0 || len(chunk.Data) > fileChunkSize || offset+int64(len(chunk.Data)) > total || chunk.Hash != fileHash(chunk.Data) || chunk.EOF != (offset+int64(len(chunk.Data)) == total) {
			return &model.APIError{Code: "FILE_TRANSFER_MISMATCH", Message: "节点分片身份或校验不匹配"}
		}
		if e := consume(chunk.Data); e != nil {
			return e
		}
		offset += int64(len(chunk.Data))
	}
	return ctx.Err()
}
func (s *Server) fileContent(w http.ResponseWriter, r *http.Request, u model.User) {
	i, ok := s.fileInstance(w, r, u, "file.read")
	if !ok {
		return
	}
	p := r.URL.Query().Get("path")
	if !filePath(w, p, false) {
		return
	}
	var entry filesystem.Entry
	if e := s.callFile(r.Context(), u, i, "file.stat", map[string]string{"path": p}, &entry); e != nil {
		bridgeError(w, e)
		return
	}
	if entry.Kind != "file" || entry.Size > maxFileText || entry.Size < 0 {
		fail(w, 413, "TEXT_LIMIT", "编辑器仅支持至多4 MiB的UTF-8文本；请使用下载")
		return
	}
	var buf bytes.Buffer
	buf.Grow(int(entry.Size))
	if e := s.readFileChunks(r.Context(), u, i, p, entry.Version, entry.Size, func(b []byte) error { _, e := buf.Write(b); return e }); e != nil {
		bridgeError(w, e)
		return
	}
	b := buf.Bytes()
	if !utf8.Valid(b) || bytes.IndexByte(b, 0) >= 0 {
		fail(w, 415, "UNSUPPORTED_ENCODING", "此文件不是可编辑UTF-8文本；请下载或使用专用工具")
		return
	}
	if fileHash(b) != entry.Version {
		fail(w, 409, "FILE_CONFLICT", "读取期间文件发生变化")
		return
	}
	encoding := "UTF-8"
	if bytes.HasPrefix(b, []byte{0xef, 0xbb, 0xbf}) {
		encoding = "UTF-8 BOM"
	}
	newline := "LF"
	if bytes.Contains(b, []byte("\r\n")) {
		newline = "CRLF"
	}
	reply(w, 200, map[string]any{"text": string(b), "version": entry.Version, "encoding": encoding, "newline": newline, "maxBytes": maxFileText})
}

// filePreview serves one bounded, read-only text window for files which are
// too large for the editor document model. It deliberately uses the same
// version-bound daemon chunks as downloads, and never buffers the whole file
// in the Master or browser. The small overlap keeps a UTF-8 code point from
// being split between adjacent preview pages.
func (s *Server) filePreview(w http.ResponseWriter, r *http.Request, u model.User) {
	i, ok := s.fileInstance(w, r, u, "file.read")
	if !ok {
		return
	}
	p := r.URL.Query().Get("path")
	if !filePath(w, p, false) {
		return
	}
	var entry filesystem.Entry
	if e := s.callFile(r.Context(), u, i, "file.stat", map[string]string{"path": p}, &entry); e != nil {
		bridgeError(w, e)
		return
	}
	if entry.Kind != "file" || entry.Size < 0 {
		fail(w, 400, "NOT_A_FILE", "预览目标必须是普通文件")
		return
	}
	if requested := r.URL.Query().Get("version"); requested != "" && requested != entry.Version {
		fail(w, 409, "FILE_CONFLICT", "预览来源版本已变化")
		return
	}
	offset := int64(0)
	if raw := r.URL.Query().Get("offset"); raw != "" {
		parsed, e := strconv.ParseInt(raw, 10, 64)
		if e != nil {
			fail(w, 400, "INVALID_PREVIEW", "预览偏移无效")
			return
		}
		offset = parsed
	}
	limit := int64(48 << 10)
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, e := strconv.ParseInt(raw, 10, 64)
		if e != nil || parsed < 1 || parsed > maxFilePreview {
			fail(w, 400, "INVALID_PREVIEW", "预览分段大小应为1～60 KiB")
			return
		}
		limit = parsed
	}
	if offset < 0 || offset > entry.Size {
		fail(w, 416, "INVALID_PREVIEW", "预览偏移超出文件范围")
		return
	}
	if offset == entry.Size {
		reply(w, 200, map[string]any{"text": "", "version": entry.Version, "offset": offset, "nextOffset": offset, "total": entry.Size, "hasMore": false, "encoding": "UTF-8", "newline": "LF", "maxBytes": maxFileText})
		return
	}
	// Read up to three bytes before and after the requested range. A valid
	// preview page always starts at a rune boundary and advances at least one
	// complete rune, so callers can safely use nextOffset for the next page.
	readStart := offset
	if readStart > 3 {
		readStart -= 3
	} else {
		readStart = 0
	}
	readLimit := (offset - readStart) + limit + 3
	if readLimit > fileChunkSize {
		readLimit = fileChunkSize
	}
	var chunk filesystem.Chunk
	if e := s.callFile(r.Context(), u, i, "file.chunk", map[string]any{"path": p, "version": entry.Version, "offset": readStart, "limit": int(readLimit)}, &chunk); e != nil {
		bridgeError(w, e)
		return
	}
	if chunk.Path != p || chunk.Version != entry.Version || chunk.Total != entry.Size || chunk.Offset != readStart || len(chunk.Data) == 0 || len(chunk.Data) > fileChunkSize || readStart+int64(len(chunk.Data)) > entry.Size || chunk.Hash != fileHash(chunk.Data) || chunk.EOF != (readStart+int64(len(chunk.Data)) == entry.Size) {
		fail(w, 409, "FILE_TRANSFER_MISMATCH", "节点分片身份或校验不匹配")
		return
	}
	start := int(offset - readStart)
	for start < len(chunk.Data) && !utf8.RuneStart(chunk.Data[start]) {
		start++
	}
	if start >= len(chunk.Data) {
		fail(w, 415, "UNSUPPORTED_ENCODING", "此文件不是可预览UTF-8文本；请下载或使用专用工具")
		return
	}
	end := min(len(chunk.Data), start+int(limit))
	textBytes := chunk.Data[start:end]
	consumed := 0
	for consumed < len(textBytes) {
		runeValue, size := utf8.DecodeRune(textBytes[consumed:])
		if runeValue == utf8.RuneError && size == 1 {
			if !utf8.FullRune(textBytes[consumed:]) {
				break
			}
			fail(w, 415, "UNSUPPORTED_ENCODING", "此文件不是可预览UTF-8文本；请下载或使用专用工具")
			return
		}
		if textBytes[consumed] == 0 {
			fail(w, 415, "UNSUPPORTED_ENCODING", "此文件不是可预览UTF-8文本；请下载或使用专用工具")
			return
		}
		consumed += size
	}
	if consumed == 0 {
		fail(w, 415, "UNSUPPORTED_ENCODING", "此文件不是可预览UTF-8文本；请下载或使用专用工具")
		return
	}
	textBytes = textBytes[:consumed]
	visibleOffset := readStart + int64(start)
	encoding := "UTF-8"
	if visibleOffset == 0 && bytes.HasPrefix(textBytes, []byte{0xef, 0xbb, 0xbf}) {
		encoding = "UTF-8 BOM"
	}
	newline := "LF"
	if bytes.Contains(textBytes, []byte("\r\n")) {
		newline = "CRLF"
	}
	next := visibleOffset + int64(len(textBytes))
	reply(w, 200, map[string]any{"text": string(textBytes), "version": entry.Version, "offset": visibleOffset, "nextOffset": next, "total": entry.Size, "hasMore": next < entry.Size, "encoding": encoding, "newline": newline, "maxBytes": maxFileText})
}

func (s *Server) fileDownload(w http.ResponseWriter, r *http.Request, u model.User) {
	defer http.NewResponseController(w).SetWriteDeadline(time.Time{})
	i, ok := s.fileInstance(w, r, u, "file.read")
	if !ok {
		return
	}
	p := r.URL.Query().Get("path")
	if !filePath(w, p, false) {
		return
	}
	var entry filesystem.Entry
	if e := s.callFile(r.Context(), u, i, "file.stat", map[string]string{"path": p}, &entry); e != nil {
		bridgeError(w, e)
		return
	}
	if entry.Kind != "file" || entry.Size < 0 {
		fail(w, 400, "NOT_A_FILE", "下载目标必须是普通文件")
		return
	}
	if v := r.URL.Query().Get("version"); v != "" && v != entry.Version {
		fail(w, 409, "FILE_CONFLICT", "下载来源版本已变化")
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": path.Base(p)}))
	w.Header().Set("Content-Length", strconv.FormatInt(entry.Size, 10))
	w.Header().Set("ETag", strconv.Quote(entry.Version))
	w.Header().Set("X-Content-SHA256", entry.Version)
	started := false
	e := s.readFileChunks(r.Context(), u, i, p, entry.Version, entry.Size, func(b []byte) error {
		if e := r.Context().Err(); e != nil {
			return e
		}
		_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(15 * time.Second))
		started = true
		_, e := w.Write(b)
		return e
	})
	if e != nil {
		if !started {
			w.Header().Del("Content-Length")
			w.Header().Del("Content-Disposition")
			bridgeError(w, e)
		} else { // Content-Length plus connection abort makes an incomplete download explicit.
			panic(http.ErrAbortHandler)
		}
		return
	}
	if !started {
		w.WriteHeader(200)
	}
}

func decodeFileText(w http.ResponseWriter, r *http.Request, v any) bool {
	controller := http.NewResponseController(w)
	_ = controller.SetReadDeadline(time.Now().Add(30 * time.Second))
	defer controller.SetReadDeadline(time.Time{})
	r.Body = http.MaxBytesReader(w, r.Body, 6*maxFileText+65536)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		fail(w, 400, "INVALID_REQUEST", "文件保存请求无效或过大")
		return false
	}
	var extra any
	if e := d.Decode(&extra); !errors.Is(e, io.EOF) {
		fail(w, 400, "INVALID_REQUEST", "仅允许一个JSON对象")
		return false
	}
	return true
}
func (s *Server) acceptFile(w http.ResponseWriter, r *http.Request, u model.User, i model.Instance, action string, p filePayload, state model.TaskState) (model.Task, bool) {
	key := r.Header.Get("Idempotency-Key")
	b, e := json.Marshal(p)
	if e != nil {
		fail(w, 400, "INVALID_REQUEST", e.Error())
		return model.Task{}, false
	}
	t, _, e := s.store.Accept(r.Context(), model.Task{ActorID: u.ID, RequestID: key, Resource: instanceRef(i), Action: action, Payload: b, State: state})
	if errors.Is(e, storage.ErrRequestMismatch) {
		fail(w, 409, "IDEMPOTENCY_CONFLICT", "同一请求ID绑定了不同文件操作或内容")
		return t, false
	}
	if errors.Is(e, storage.ErrQueueLimit) {
		fail(w, 429, "QUEUE_LIMIT", e.Error())
		return t, false
	}
	if e != nil {
		fail(w, 500, "STORAGE_ERROR", e.Error())
		return t, false
	}
	return t, true
}
func newUploadSpec(u model.User, key, p, version, hash, name, fingerprint string, total, modified int64) filesystem.UploadSpec {
	digest := sha256.Sum256([]byte(u.ID + "\x00" + key))
	return filesystem.UploadSpec{ID: hex.EncodeToString(digest[:16]), OwnerID: u.ID, Path: p, Total: total, Hash: hash, ExpectedVersion: version, SourceName: name, SourceModified: modified, SourceFingerprint: fingerprint}
}
func uploadPayload(i model.Instance, spec filesystem.UploadSpec) filePayload {
	return filePayload{FileTaskPayload: model.FileTaskPayload{Config: i.Config, Path: spec.Path, Version: spec.ExpectedVersion, UploadID: spec.ID, Hash: spec.Hash, Total: spec.Total}, UploadSpec: &spec}
}
func (s *Server) checkpointUpload(ctx context.Context, t model.Task, u filesystem.Upload, phase string, failure string) (model.Task, error) {
	b, _ := json.Marshal(u)
	for range 4 {
		current, e := s.store.Task(ctx, t.ID)
		if e != nil {
			return current, e
		}
		if current.State != model.WaitingClient {
			return current, nil
		}
		updated, e := s.store.UpdateTask(ctx, t.ID, current.Revision, model.WaitingClient, phase, b, failure)
		if errors.Is(e, storage.ErrConflict) {
			continue
		}
		return updated, e
	}
	return t, storage.ErrConflict
}
func (s *Server) uploadFailure(w http.ResponseWriter, r *http.Request, t model.Task, upload filesystem.Upload, e error) {
	api := &model.APIError{Code: "NODE_UNREACHABLE", Message: e.Error()}
	var from *model.APIError
	if errors.As(e, &from) {
		api = from
	}
	state := model.WaitingClient
	phase := "waiting_client_or_node"
	if api.Code == "FILE_CONFLICT" || api.Code == "FILE_TRANSFER_MISMATCH" || api.Code == "FORBIDDEN" || api.Code == "RESOURCE_LIMIT" || api.Code == "NOT_FOUND" || api.Code == "NODE_OPERATION_FAILED" {
		state = model.Failed
		phase = "file_failed"
	}
	if api.Code == "FILE_CONFLICT" {
		phase = "file_conflict"
	}
	// The request context can be cancelled after a node accepted a chunk. A
	// short independent journal write preserves the task without replaying data.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	current, err := s.store.Task(ctx, t.ID)
	if err == nil {
		t = current
	}
	if err == nil && current.State == model.WaitingClient {
		result, _ := json.Marshal(map[string]any{"upload": upload, "error": api})
		if updated, err := s.store.UpdateTask(ctx, current.ID, current.Revision, state, phase, result, api.Message); err == nil {
			t = updated
		} else {
			t = current
		}
	}
	reply(w, 202, map[string]any{"task": t, "upload": upload, "waiting": api})
}

func (s *Server) fileSave(w http.ResponseWriter, r *http.Request, u model.User) {
	i, ok := s.fileInstance(w, r, u, "file.write")
	if !ok {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var in struct {
		Path    string `json:"path"`
		Text    string `json:"text"`
		Version string `json:"version"`
	}
	if !decodeFileText(w, r, &in) {
		return
	}
	if !filePath(w, in.Path, false) {
		return
	}
	if len(in.Text) > maxFileText || !utf8.ValidString(in.Text) || strings.ContainsRune(in.Text, 0) {
		fail(w, 413, "TEXT_LIMIT", "只能保存至多4 MiB的UTF-8文本")
		return
	}
	if !fileVersionValid(in.Version) {
		fail(w, 400, "VERSION_REQUIRED", "保存必须提供基线版本或missing")
		return
	}
	hash := fileHash([]byte(in.Text))
	spec := newUploadSpec(u, r.Header.Get("Idempotency-Key"), in.Path, in.Version, hash, path.Base(in.Path), hash, int64(len(in.Text)), 0)
	t, ok := s.acceptFile(w, r, u, i, "file.save", uploadPayload(i, spec), model.WaitingClient)
	if !ok {
		return
	}
	if t.State != model.WaitingClient {
		reply(w, 202, map[string]any{"task": t})
		return
	}
	var upload filesystem.Upload
	if e := s.callFile(r.Context(), u, i, "file.upload.begin", map[string]any{"spec": spec}, &upload); e != nil {
		s.uploadFailure(w, r, t, upload, e)
		return
	}
	if upload.Spec != spec || upload.ChunkBytes != fileChunkSize || upload.Offset < 0 || upload.Offset > spec.Total {
		s.uploadFailure(w, r, t, upload, &model.APIError{Code: "NODE_OPERATION_FAILED", Message: "上传检查点身份不一致"})
		return
	}
	for offset := upload.Offset; offset < spec.Total; {
		data := []byte(in.Text[offset:min(offset+int64(fileChunkSize), spec.Total)])
		var next filesystem.Upload
		if e := s.callFile(r.Context(), u, i, "file.upload.chunk", map[string]any{"id": spec.ID, "offset": offset, "data": data, "hash": fileHash(data)}, &next); e != nil {
			s.uploadFailure(w, r, t, upload, e)
			return
		}
		if next.Spec != spec || next.Offset < offset+int64(len(data)) || next.Offset > spec.Total {
			s.uploadFailure(w, r, t, upload, &model.APIError{Code: "NODE_OPERATION_FAILED", Message: "节点未确认预期分片"})
			return
		}
		upload = next
		offset = upload.Offset
		var e error
		t, e = s.checkpointUpload(r.Context(), t, upload, "receiving_file", "")
		if e != nil {
			s.uploadFailure(w, r, t, upload, e)
			return
		}
		if t.State != model.WaitingClient {
			reply(w, 202, map[string]any{"task": t, "upload": upload})
			return
		}
	}
	t, e := s.queueUpload(r.Context(), u, i, t, upload)
	if e != nil {
		s.uploadFailure(w, r, t, upload, e)
		return
	}
	reply(w, 202, map[string]any{"task": t, "upload": upload})
}
func (s *Server) queueUpload(ctx context.Context, u model.User, i model.Instance, t model.Task, upload filesystem.Upload) (model.Task, error) {
	latest, err := s.store.Task(ctx, t.ID)
	if err != nil {
		return t, err
	}
	if latest.State != model.WaitingClient {
		return latest, nil
	}
	t = latest
	if upload.Offset != upload.Spec.Total || upload.Stage != "receiving" {
		return t, &model.APIError{Code: "FILE_TRANSFER_MISMATCH", Message: "上传尚未准备完成"}
	}
	current, e := s.store.User(ctx, u.ID)
	if e != nil || !s.store.Allowed(ctx, current, instanceRef(i), "file.write") {
		return t, &model.APIError{Code: "FORBIDDEN", Message: "提交时写入权限已撤销"}
	}
	for range 4 {
		t, e = s.store.Task(ctx, t.ID)
		if e != nil {
			return t, e
		}
		if t.State != model.WaitingClient {
			return t, nil
		}
		b, _ := json.Marshal(upload)
		next, e := s.store.UpdateTask(ctx, t.ID, t.Revision, model.Queued, "node_data_ready", b, "")
		if errors.Is(e, storage.ErrConflict) {
			continue
		}
		return next, e
	}
	return t, storage.ErrConflict
}

func (s *Server) fileAction(w http.ResponseWriter, r *http.Request, u model.User) {
	i, ok := s.fileInstance(w, r, u, "file.write")
	if !ok {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var in struct {
		Action        string            `json:"action"`
		Path          string            `json:"path"`
		Target        string            `json:"target"`
		Version       string            `json:"version"`
		TargetVersion string            `json:"targetVersion"`
		TrashID       string            `json:"trashId"`
		Overwrite     map[string]string `json:"overwrite,omitempty"`
	}
	if !decode(w, r, &in) {
		return
	}
	allowed := map[string]bool{"mkdir": true, "copy": true, "move": true, "delete": true, "restore": true, "compress": true, "extract": true}
	if !allowed[in.Action] {
		fail(w, 400, "INVALID_ACTION", "未知文件操作")
		return
	}
	if in.Action != "restore" && !filePath(w, in.Path, false) {
		return
	}
	if in.Action != "mkdir" && in.Action != "restore" && !validFileHash(in.Version) {
		fail(w, 400, "VERSION_REQUIRED", "源版本必填")
		return
	}
	if in.Action == "copy" || in.Action == "move" || in.Action == "compress" || in.Action == "extract" {
		if !s.allowed(w, r, u, instanceRef(i), "file.read") {
			return
		}
	}
	if in.Action != "mkdir" && in.Action != "delete" {
		if !filePath(w, in.Target, in.Action == "extract") {
			return
		}
		if !fileVersionValid(in.TargetVersion) {
			fail(w, 400, "VERSION_REQUIRED", "目标基线版本必填")
			return
		}
	}
	if in.Action == "restore" && (len(in.TrashID) != 32 || strings.ContainsAny(in.TrashID, "/\\.")) {
		fail(w, 400, "INVALID_TRASH", "回收项ID无效")
		return
	}
	if len(in.Overwrite) > 100 {
		fail(w, 413, "REQUEST_LIMIT", "一次覆盖清单最多100项")
		return
	}
	for name, v := range in.Overwrite {
		if !filePath(w, name, false) {
			return
		}
		if !validFileHash(v) {
			fail(w, 400, "VERSION_REQUIRED", "覆盖项必须提供现有版本")
			return
		}
	}
	p := filePayload{FileTaskPayload: model.FileTaskPayload{Config: i.Config, Path: in.Path, Target: in.Target, Version: in.Version, TargetVersion: in.TargetVersion, TrashID: in.TrashID}}
	// Overwrite is included in the persisted digest; node execution consumes
	// the same field after the shared model declares it.
	var body map[string]any
	b, _ := json.Marshal(p)
	_ = json.Unmarshal(b, &body)
	if len(in.Overwrite) > 0 {
		body["overwrite"] = in.Overwrite
	}
	raw, _ := json.Marshal(body)
	key := r.Header.Get("Idempotency-Key")
	task, _, e := s.store.Accept(r.Context(), model.Task{ActorID: u.ID, RequestID: key, Resource: instanceRef(i), Action: "file." + in.Action, Payload: raw})
	if errors.Is(e, storage.ErrRequestMismatch) {
		fail(w, 409, "IDEMPOTENCY_CONFLICT", e.Error())
		return
	}
	if errors.Is(e, storage.ErrQueueLimit) {
		fail(w, 429, "QUEUE_LIMIT", e.Error())
		return
	}
	if e != nil {
		fail(w, 500, "STORAGE_ERROR", e.Error())
		return
	}
	reply(w, 202, map[string]any{"task": task})
}

func (s *Server) fileUploadBegin(w http.ResponseWriter, r *http.Request, u model.User) {
	i, ok := s.fileInstance(w, r, u, "file.write")
	if !ok {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	var in struct {
		Path              string `json:"path"`
		Total             int64  `json:"total"`
		Hash              string `json:"hash"`
		Version           string `json:"version"`
		SourceName        string `json:"sourceName"`
		SourceModified    int64  `json:"sourceModified"`
		SourceFingerprint string `json:"sourceFingerprint"`
	}
	if !decode(w, r, &in) {
		return
	}
	if !filePath(w, in.Path, false) {
		return
	}
	if in.Total < 0 || in.Total > 1<<30 || !validFileHash(in.Hash) || !fileVersionValid(in.Version) || in.SourceName == "" || len(in.SourceName) > 512 || in.SourceFingerprint == "" || len(in.SourceFingerprint) > 512 {
		fail(w, 400, "INVALID_TRANSFER", "需要大小、SHA256、来源名称/时间/指纹和目标基线")
		return
	}
	spec := newUploadSpec(u, r.Header.Get("Idempotency-Key"), in.Path, in.Version, in.Hash, in.SourceName, in.SourceFingerprint, in.Total, in.SourceModified)
	t, ok := s.acceptFile(w, r, u, i, "file.upload", uploadPayload(i, spec), model.WaitingClient)
	if !ok {
		return
	}
	if t.State != model.WaitingClient {
		reply(w, 202, map[string]any{"task": t})
		return
	}
	var upload filesystem.Upload
	if e := s.callFile(r.Context(), u, i, "file.upload.begin", map[string]any{"spec": spec}, &upload); e != nil {
		s.uploadFailure(w, r, t, upload, e)
		return
	}
	if upload.Spec != spec {
		s.uploadFailure(w, r, t, upload, &model.APIError{Code: "NODE_OPERATION_FAILED", Message: "上传身份不一致"})
		return
	}
	t, e := s.checkpointUpload(r.Context(), t, upload, "waiting_client_data", "")
	if e != nil {
		s.uploadFailure(w, r, t, upload, e)
		return
	}
	reply(w, 202, map[string]any{"task": t, "upload": upload})
}
func (s *Server) uploadTask(w http.ResponseWriter, r *http.Request, u model.User) (model.Instance, model.Task, filePayload, bool) {
	i, ok := s.fileInstance(w, r, u, "file.write")
	if !ok {
		return i, model.Task{}, filePayload{}, false
	}
	t, e := s.store.Task(r.Context(), r.PathValue("taskId"))
	var p filePayload
	if e != nil || t.ActorID != u.ID || t.Resource.Key() != instanceRef(i).Key() || (t.Action != "file.upload" && t.Action != "file.save") || json.Unmarshal(t.Payload, &p) != nil || p.UploadSpec == nil || p.UploadID != p.UploadSpec.ID || p.UploadSpec.OwnerID != u.ID {
		fail(w, 404, "NOT_FOUND", "上传任务不存在或不属于当前账号和实例")
		return i, t, p, false
	}
	return i, t, p, true
}
func (s *Server) fileUploadStatus(w http.ResponseWriter, r *http.Request, u model.User) {
	i, t, p, ok := s.uploadTask(w, r, u)
	if !ok {
		return
	}
	var upload filesystem.Upload
	if e := s.callFile(r.Context(), u, i, "file.upload.status", map[string]string{"id": p.UploadID}, &upload); e != nil {
		bridgeError(w, e)
		return
	}
	if upload.Spec != *p.UploadSpec {
		fail(w, 409, "FILE_TRANSFER_MISMATCH", "检查点来源不匹配")
		return
	}
	reply(w, 200, map[string]any{"task": t, "upload": upload})
}
func (s *Server) fileUploadChunk(w http.ResponseWriter, r *http.Request, u model.User) {
	i, t, p, ok := s.uploadTask(w, r, u)
	if !ok {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	if t.State != model.WaitingClient {
		fail(w, 409, "TASK_CONFLICT", "当前任务不接收上传数据")
		return
	}
	var in struct {
		Offset int64  `json:"offset"`
		Data   []byte `json:"data"`
		Hash   string `json:"hash"`
	}
	if !decode(w, r, &in) {
		return
	}
	if in.Offset < 0 || len(in.Data) == 0 || len(in.Data) > fileChunkSize || in.Hash != fileHash(in.Data) {
		fail(w, 400, "INVALID_CHUNK", "分片长度或SHA256无效")
		return
	}
	var upload filesystem.Upload
	if e := s.callFile(r.Context(), u, i, "file.upload.chunk", map[string]any{"id": p.UploadID, "offset": in.Offset, "data": in.Data, "hash": in.Hash}, &upload); e != nil {
		bridgeError(w, e)
		return
	}
	if upload.Spec != *p.UploadSpec || upload.Offset < in.Offset+int64(len(in.Data)) {
		fail(w, 409, "FILE_TRANSFER_MISMATCH", "节点检查点未确认当前来源分片")
		return
	}
	t, e := s.checkpointUpload(r.Context(), t, upload, "waiting_client_data", "")
	if e != nil {
		bridgeError(w, e)
		return
	}
	reply(w, 200, map[string]any{"task": t, "upload": upload})
}
func (s *Server) fileUploadComplete(w http.ResponseWriter, r *http.Request, u model.User) {
	i, t, p, ok := s.uploadTask(w, r, u)
	if !ok {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	if t.State != model.WaitingClient {
		reply(w, 202, map[string]any{"task": t})
		return
	}
	var upload filesystem.Upload
	if e := s.callFile(r.Context(), u, i, "file.upload.status", map[string]string{"id": p.UploadID}, &upload); e != nil {
		s.uploadFailure(w, r, t, upload, e)
		return
	}
	if upload.Spec != *p.UploadSpec {
		fail(w, 409, "FILE_TRANSFER_MISMATCH", "上传来源不匹配")
		return
	}
	t, e := s.queueUpload(r.Context(), u, i, t, upload)
	if e != nil {
		bridgeError(w, e)
		return
	}
	reply(w, 202, map[string]any{"task": t, "upload": upload})
}
func (s *Server) fileUploadCancel(w http.ResponseWriter, r *http.Request, u model.User) {
	_, t, _, ok := s.uploadTask(w, r, u)
	if !ok {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	s.cancelUploadTask(w, r, u, t)
}

// cancelUploadTask also serves the generic task cancel route. No source or
// committed destination is removed, and cleanup is acknowledged by the node.
func (s *Server) cancelUploadTask(w http.ResponseWriter, r *http.Request, u model.User, t model.Task) {
	var p filePayload
	if json.Unmarshal(t.Payload, &p) != nil || p.UploadSpec == nil || (t.ActorID != u.ID && !u.Admin) {
		fail(w, 403, "FORBIDDEN", "上传任务不属于当前账号")
		return
	}
	i, e := s.store.Instance(r.Context(), t.Resource.ID)
	if e != nil {
		fail(w, 404, "NOT_FOUND", "上传任务的实例不存在")
		return
	}
	if !s.allowed(w, r, u, t.Resource, "file.write") {
		return
	}
	if t.State == model.Succeeded {
		fail(w, 409, "TASK_CONFLICT", "已提交文件不能通过取消上传删除")
		return
	}
	if !t.State.Terminal() {
		t, e = s.store.RequestCancel(r.Context(), t.ID, r.Header.Get("Idempotency-Key"))
		if e != nil {
			bridgeError(w, e)
			return
		}
	}
	var upload filesystem.Upload
	// An authorized administrator can clean the original actor's staging even
	// after that actor lost write permission. The actor is taken only from the
	// persisted task, and the node still checks upload ownership.
	args, _ := json.Marshal(map[string]string{"id": p.UploadID})
	e = s.nodeCall(r.Context(), i.NodeID, protocol.ChannelBulk, bridge.Request{Method: "file.upload.cancel", ActorID: t.ActorID, Resource: instanceRef(i), Config: &i.Config, Args: args}, &upload)
	if e != nil {
		var api *model.APIError
		if errors.As(e, &api) && api.Code == "NOT_FOUND" {
			upload = filesystem.Upload{Spec: *p.UploadSpec, Stage: "cancelled", Updated: time.Now().UTC()}
		} else {
			reply(w, 202, map[string]any{"task": t, "waiting": model.APIError{Code: "CLEANUP_PENDING", Message: e.Error()}})
			return
		}
	}
	if upload.Stage != "cancelled" {
		fail(w, 409, "TASK_CONFLICT", "节点尚未确认取消")
		return
	}
	for range 4 {
		t, e = s.store.Task(r.Context(), t.ID)
		if e != nil {
			bridgeError(w, e)
			return
		}
		if t.State.Terminal() {
			break
		}
		result, _ := json.Marshal(upload)
		next, err := s.store.UpdateTask(r.Context(), t.ID, t.Revision, model.Cancelled, "node_upload_cancelled", result, "")
		if errors.Is(err, storage.ErrConflict) {
			continue
		}
		if err != nil {
			bridgeError(w, err)
			return
		}
		t = next
		break
	}
	reply(w, 202, map[string]any{"task": t, "upload": upload})
}

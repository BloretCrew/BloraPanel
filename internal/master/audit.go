package master

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
)

func (s *Server) audit(w http.ResponseWriter, r *http.Request, u model.User) {
	if !s.admin(w, u) {
		return
	}
	q := r.URL.Query()
	value := func(key, alternate string) string {
		if v := strings.TrimSpace(q.Get(key)); v != "" {
			return v
		}
		return strings.TrimSpace(q.Get(alternate))
	}
	filter := storage.AuditFilter{
		ActorID:     value("userId", "actorId"),
		NodeID:      value("nodeId", ""),
		ResourceKey: value("resource", "resourceKey"),
		Action:      value("action", ""),
		Limit:       100,
	}
	for name, field := range map[string]*string{
		"userId": &filter.ActorID, "actorId": &filter.ActorID,
		"nodeId": &filter.NodeID, "resource": &filter.ResourceKey,
		"resourceKey": &filter.ResourceKey, "action": &filter.Action,
	} {
		if len(*field) > 160 {
			fail(w, http.StatusBadRequest, "INVALID_QUERY", name+" 过长")
			return
		}
	}
	var err error
	if raw := strings.TrimSpace(q.Get("from")); raw != "" {
		filter.From, err = parseAuditTime(raw)
		if err != nil {
			fail(w, http.StatusBadRequest, "INVALID_QUERY", "from 必须是 RFC3339 时间")
			return
		}
	}
	if raw := strings.TrimSpace(q.Get("to")); raw != "" {
		filter.To, err = parseAuditTime(raw)
		if err != nil {
			fail(w, http.StatusBadRequest, "INVALID_QUERY", "to 必须是 RFC3339 时间")
			return
		}
	}
	if !filter.From.IsZero() && !filter.To.IsZero() && filter.From.After(filter.To) {
		fail(w, http.StatusBadRequest, "INVALID_QUERY", "from 不能晚于 to")
		return
	}
	if raw := strings.TrimSpace(q.Get("limit")); raw != "" {
		filter.Limit, err = strconv.Atoi(raw)
		if err != nil || filter.Limit < 1 || filter.Limit > 200 {
			fail(w, http.StatusBadRequest, "INVALID_QUERY", "limit 必须在1～200之间")
			return
		}
	}
	if raw := strings.TrimSpace(q.Get("offset")); raw != "" {
		filter.Offset, err = strconv.Atoi(raw)
		if err != nil || filter.Offset < 0 || filter.Offset > 1000000 {
			fail(w, http.StatusBadRequest, "INVALID_QUERY", "offset 超出范围")
			return
		}
	}
	items, err := s.store.Audits(r.Context(), filter)
	if err != nil {
		fail(w, http.StatusInternalServerError, "STORAGE_ERROR", err.Error())
		return
	}
	next := 0
	if len(items) == filter.Limit {
		next = filter.Offset + len(items)
	}
	reply(w, http.StatusOK, map[string]any{"items": items, "nextOffset": next})
}

func parseAuditTime(value string) (time.Time, error) {
	if len(value) > 64 {
		return time.Time{}, strconv.ErrSyntax
	}
	return time.Parse(time.RFC3339Nano, value)
}

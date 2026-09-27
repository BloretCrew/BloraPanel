package master

import (
	"blora.dev/panel/internal/model"
	"net/http"
	"strconv"
)

func (s *Server) taskStages(w http.ResponseWriter, r *http.Request, u model.User) {
	task, err := s.store.Task(r.Context(), r.PathValue("id"))
	if err != nil || !s.canTask(r.Context(), u, task) {
		fail(w, 404, "NOT_FOUND", "任务不存在或未授权")
		return
	}
	before := int64(0)
	limit := 50
	if raw := r.URL.Query().Get("before"); raw != "" {
		before, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || before < 0 || before > 9007199254740991 {
			fail(w, 400, "INVALID_PAGE", "阶段记录游标无效")
			return
		}
	}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 100 {
			fail(w, 400, "INVALID_PAGE", "阶段记录每页1～100条")
			return
		}
	}
	items, more, err := s.store.TaskStagePage(r.Context(), task.ID, before, limit)
	if err != nil {
		fail(w, 500, "STORAGE_ERROR", "阶段记录读取失败")
		return
	}
	next := int64(-1)
	if more && len(items) > 0 {
		next = items[len(items)-1].Revision
	}
	reply(w, 200, map[string]any{"items": items, "nextBefore": next})
}

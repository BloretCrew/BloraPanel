package master

import (
	"blora.dev/panel/internal/model"
	"context"
	"net/http"
	"strconv"
	"time"
)

func (s *Server) taskHistory(w http.ResponseWriter, r *http.Request, u model.User) {
	before, err := strconv.ParseInt(r.URL.Query().Get("before"), 10, 64)
	if err != nil || before < 0 || before > 9007199254740991 || r.URL.Query().Has("offset") {
		fail(w, 400, "INVALID_PAGE", "任务历史游标无效")
		return
	}
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 1000 {
			fail(w, 400, "INVALID_PAGE", "任务分页大小应为1～1000")
			return
		}
	}
	state := model.TaskState(r.URL.Query().Get("state"))
	if state != "" && state != model.Queued && state != model.Running && state != model.WaitingNode && state != model.WaitingClient && state != model.CancelRequested && !state.Terminal() {
		fail(w, 400, "INVALID_STATE", "任务状态无效")
		return
	}
	entries, more, err := s.store.RootTaskPageFiltered(r.Context(), before, limit, state, false)
	if err != nil {
		fail(w, 500, "STORAGE_ERROR", "任务历史读取失败")
		return
	}
	items := []model.Task{}
	for _, entry := range entries {
		if s.canTask(r.Context(), u, entry.Task) {
			items = append(items, entry.Task)
		}
	}
	next := int64(-1)
	if more && len(entries) > 0 {
		next = entries[len(entries)-1].Cursor
	}
	reply(w, 200, map[string]any{"items": items, "nextBefore": next})
}

func (s *Server) taskSummary(w http.ResponseWriter, r *http.Request, u model.User) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	before := int64(0)
	total := 0
	states := map[model.TaskState]int{}
	for {
		entries, more, err := s.store.RootTaskPageFiltered(ctx, before, 500, "", true)
		if err != nil {
			fail(w, 503, "TASK_SUMMARY_UNAVAILABLE", "后台任务统计暂不可用")
			return
		}
		for _, entry := range entries {
			before = entry.Cursor
			if s.canTask(ctx, u, entry.Task) {
				total++
				states[entry.Task.State]++
			}
		}
		if ctx.Err() != nil {
			fail(w, 503, "TASK_SUMMARY_UNAVAILABLE", "后台任务统计超时")
			return
		}
		if !more {
			break
		}
	}
	reply(w, 200, map[string]any{"active": total, "states": states})
}

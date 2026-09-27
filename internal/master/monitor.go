package master

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/monitor"
	"blora.dev/panel/internal/protocol"
)

const monitorHistoryNS = "monitor_points_v1"
const instanceMetricHistoryNS = "monitor_instance_points_v1"

func (s *Server) cacheMetric(ctx context.Context, node string, point monitor.Point) {
	id := node + "-" + point.ObservedAt.UTC().Format("20060102150405.000000000")
	if _, err := s.store.PutRecord(ctx, monitorHistoryNS, id, 0, point); err != nil {
		return
	}
	ids, err := s.store.RecordIDs(ctx, monitorHistoryNS)
	if err != nil {
		return
	}
	// IDs contain the node and UTC timestamp; prune only this node and retain
	// the newest bounded window, never deleting another node's history.
	owned := make([]string, 0, len(ids))
	for _, v := range ids {
		if len(v) > len(node)+1 && v[:len(node)+1] == node+"-" {
			owned = append(owned, v)
		}
	}
	for len(owned) > 120 {
		id := owned[0]
		owned = owned[1:]
		var p monitor.Point
		rev, e := s.store.Record(ctx, monitorHistoryNS, id, &p)
		if e == nil {
			_, _ = s.store.DB.ExecContext(ctx, "DELETE FROM records WHERE namespace=? AND id=? AND revision=?", monitorHistoryNS, id, rev)
		}
	}
}
func (s *Server) cachedMetrics(ctx context.Context, node string) ([]monitor.Point, error) {
	ids, err := s.store.RecordIDs(ctx, monitorHistoryNS)
	if err != nil {
		return nil, err
	}
	out := []monitor.Point{}
	for _, id := range ids {
		if len(id) <= len(node)+1 || id[:len(node)+1] != node+"-" {
			continue
		}
		var p monitor.Point
		if _, e := s.store.Record(ctx, monitorHistoryNS, id, &p); e != nil {
			continue
		}
		p.Stale = true
		out = append(out, p)
	}
	if len(out) > 120 {
		out = out[len(out)-120:]
	}
	return out, nil
}

func instanceMetricKey(instance string, observed time.Time) string {
	return instance + "-" + observed.UTC().Format("20060102150405.000000000")
}

func (s *Server) cacheInstanceMetric(ctx context.Context, instance string, point monitor.ProcessPoint) {
	if instance == "" || point.ObservedAt.IsZero() {
		return
	}
	id := instanceMetricKey(instance, point.ObservedAt)
	if _, err := s.store.PutRecord(ctx, instanceMetricHistoryNS, id, 0, point); err != nil {
		return
	}
	ids, err := s.store.RecordIDs(ctx, instanceMetricHistoryNS)
	if err != nil {
		return
	}
	owned := make([]string, 0, len(ids))
	for _, candidate := range ids {
		if strings.HasPrefix(candidate, instance+"-") {
			owned = append(owned, candidate)
		}
	}
	sort.Strings(owned)
	for len(owned) > 120 {
		id := owned[0]
		owned = owned[1:]
		var p monitor.ProcessPoint
		rev, e := s.store.Record(ctx, instanceMetricHistoryNS, id, &p)
		if e == nil {
			_, _ = s.store.DB.ExecContext(ctx, "DELETE FROM records WHERE namespace=? AND id=? AND revision=?", instanceMetricHistoryNS, id, rev)
		}
	}
}

func (s *Server) cachedInstanceMetric(ctx context.Context, instance, runID string) (monitor.ProcessPoint, error) {
	var out monitor.ProcessPoint
	ids, err := s.store.RecordIDs(ctx, instanceMetricHistoryNS)
	if err != nil {
		return out, err
	}
	owned := make([]string, 0, len(ids))
	for _, candidate := range ids {
		if strings.HasPrefix(candidate, instance+"-") {
			owned = append(owned, candidate)
		}
	}
	if len(owned) == 0 {
		return out, errors.New("instance metric history unavailable")
	}
	sort.Strings(owned)
	for index := len(owned) - 1; index >= 0; index-- {
		if _, err = s.store.Record(ctx, instanceMetricHistoryNS, owned[index], &out); err != nil {
			return out, err
		}
		if runID != "" && out.RunID == runID {
			out.Stale = true
			return out, nil
		}
	}
	return monitor.ProcessPoint{}, errors.New("current run metric history unavailable")
}

func (s *Server) registerMonitoring() {
	s.mux.HandleFunc("GET /api/v1/nodes/{id}/metrics", s.auth(s.nodeMetrics))
	s.mux.HandleFunc("GET /api/v1/nodes/{id}/metrics/history", s.auth(s.nodeMetricsHistory))
	s.mux.HandleFunc("GET /api/v1/instances/{id}/metrics", s.auth(s.instanceMetrics))
	s.mux.HandleFunc("GET /api/v1/nodes/{id}/processes", s.auth(s.processList))
	s.mux.HandleFunc("POST /api/v1/nodes/{id}/processes/{pid}/terminate", s.auth(s.processTerminate))
}
func (s *Server) processList(w http.ResponseWriter, r *http.Request, u model.User) {
	ref := nodeRef(r.PathValue("id"))
	if !s.allowed(w, r, u, ref, "host.manage") {
		return
	}
	var out map[string]any
	q := monitor.ProcessQuery{Limit: 100, Search: r.URL.Query().Get("search")}
	for key, target := range map[string]*int{"limit": &q.Limit, "afterPid": &q.AfterPID} {
		if raw := r.URL.Query().Get(key); raw != "" {
			value, err := strconv.Atoi(raw)
			if err != nil {
				fail(w, 400, "INVALID_PAGE", "无效进程分页参数")
				return
			}
			*target = value
		}
	}
	if q.Limit < 1 || q.Limit > 100 || q.AfterPID < 0 || len(q.Search) > 256 {
		fail(w, 400, "INVALID_PAGE", "无效进程分页参数")
		return
	}
	args, _ := json.Marshal(q)
	if err := s.nodeCall(r.Context(), ref.ID, protocol.ChannelBulk, bridge.Request{Method: "monitor.processes", ActorID: u.ID, Resource: ref, Args: args}, &out); err != nil {
		bridgeError(w, err)
		return
	}
	reply(w, 200, out)
}
func (s *Server) processTerminate(w http.ResponseWriter, r *http.Request, u model.User) {
	ref := nodeRef(r.PathValue("id"))
	if !s.allowed(w, r, u, ref, "host.manage") {
		return
	}
	if !requireRequestID(w, r) {
		return
	}
	pid, err := strconv.Atoi(r.PathValue("pid"))
	if err != nil || pid <= 0 {
		fail(w, 400, "INVALID_PID", "无效进程号")
		return
	}
	var in struct {
		StartTicks uint64 `json:"startTicks"`
	}
	if !decode(w, r, &in) || in.StartTicks == 0 {
		fail(w, 400, "PROCESS_IDENTITY_REQUIRED", "需要进程出生标识")
		return
	}
	b, _ := json.Marshal(map[string]any{"pid": pid, "startTicks": in.StartTicks})
	task, _, err := s.store.Accept(r.Context(), model.Task{ActorID: u.ID, RequestID: r.Header.Get("Idempotency-Key"), Resource: ref, Action: "process.terminate", Payload: b})
	if err != nil {
		taskError(w, err)
		return
	}
	reply(w, 202, map[string]any{"task": s.presentAcceptedTask(r.Context(), task)})
}
func (s *Server) nodeMetricsHistory(w http.ResponseWriter, r *http.Request, u model.User) {
	ref := nodeRef(r.PathValue("id"))
	if _, _, err := s.store.Node(r.Context(), ref.ID); err != nil {
		fail(w, 404, "NOT_FOUND", "节点不存在")
		return
	}
	if !s.allowed(w, r, u, ref, "node.read") {
		return
	}
	var out map[string]any
	if err := s.nodeCall(r.Context(), ref.ID, protocol.ChannelBulk, bridge.Request{Method: "monitor.history", ActorID: u.ID, Resource: ref, Args: []byte(`{}`)}, &out); err != nil {
		items, e := s.cachedMetrics(r.Context(), ref.ID)
		if e != nil || len(items) == 0 {
			bridgeError(w, err)
			return
		}
		reply(w, 200, map[string]any{"items": items, "retention": 120, "stale": true, "diagnostic": "节点当前无法连接，显示最近保存的采样"})
		return
	}
	reply(w, 200, out)
}
func (s *Server) nodeMetrics(w http.ResponseWriter, r *http.Request, u model.User) {
	ref := nodeRef(r.PathValue("id"))
	if _, _, err := s.store.Node(r.Context(), ref.ID); err != nil {
		fail(w, 404, "NOT_FOUND", "节点不存在")
		return
	}
	if !s.allowed(w, r, u, ref, "node.read") {
		return
	}
	var point monitor.Point
	if err := s.nodeCall(r.Context(), ref.ID, protocol.ChannelBulk, bridge.Request{Method: "monitor.node", ActorID: u.ID, Resource: ref, Args: []byte(`{}`)}, &point); err != nil {
		if cached, e := s.cachedMetrics(r.Context(), ref.ID); e == nil && len(cached) > 0 {
			last := cached[len(cached)-1]
			last.Diagnostic = "节点当前无法连接，显示最近保存的采样"
			reply(w, 200, last)
			return
		}
		bridgeError(w, err)
		return
	}
	s.cacheMetric(r.Context(), ref.ID, point)
	reply(w, 200, point)
}
func (s *Server) instanceMetrics(w http.ResponseWriter, r *http.Request, u model.User) {
	i, err := s.store.Instance(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, 404, "NOT_FOUND", "实例不存在")
		return
	}
	if !s.allowed(w, r, u, instanceRef(i), "instance.read") {
		return
	}
	args, _ := json.Marshal(map[string]string{"runId": i.RunID})
	var point monitor.ProcessPoint
	sampleErr := s.nodeCall(r.Context(), i.NodeID, protocol.ChannelBulk, bridge.Request{Method: "monitor.instance", ActorID: u.ID, Resource: instanceRef(i), Args: args}, &point)
	latest, err := s.store.Instance(r.Context(), i.ID)
	if err != nil {
		fail(w, 404, "NOT_FOUND", "实例不存在")
		return
	}
	if !s.allowed(w, r, u, instanceRef(latest), "instance.read") {
		return
	}
	if latest.RunID != i.RunID || latest.NodeID != i.NodeID {
		fail(w, 409, "RUN_CHANGED", "采样期间运行代次已变化，请重新读取指标")
		return
	}
	if sampleErr != nil {
		if cached, cacheErr := s.cachedInstanceMetric(r.Context(), i.ID, i.RunID); cacheErr == nil {
			cached.Diagnostic = "实例节点当前无法连接，显示该运行代次的最近采样"
			reply(w, 200, cached)
			return
		}
		bridgeError(w, sampleErr)
		return
	}
	if i.RunID == "" || point.RunID != i.RunID {
		fail(w, 409, "RUN_CHANGED", "节点返回的指标不属于当前运行代次")
		return
	}
	s.cacheInstanceMetric(r.Context(), i.ID, point)
	reply(w, 200, point)
}

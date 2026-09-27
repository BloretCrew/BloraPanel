package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"testing"
	"time"
)

func TestContainerMetricsChecksIdentityAndRunBirth(t *testing.T) {
	for _, mode := range []string{"valid", "foreign-stats", "restart", "old-sample"} {
		t.Run(mode, func(t *testing.T) {
			m, f := newExecFixture(t)
			f.server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) {
				if q.URL.Path != "/v1.45/containers/"+f.r.ContainerID+"/stats" {
					f.handle(w, q)
					return
				}
				if q.URL.Query().Get("stream") != "false" {
					t.Error("unbounded stats stream requested")
				}
				s := metricStatsFixture(f.r.ContainerID)
				if mode == "foreign-stats" {
					s.ID = "foreign"
				}
				if mode == "old-sample" {
					s.Read = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
				}
				if mode == "restart" {
					f.mu.Lock()
					f.started = "2026-09-09T00:01:00Z"
					f.mu.Unlock()
				}
				json.NewEncoder(w).Encode(s)
			})
			p, err := m.ContainerMetrics(context.Background(), f.r)
			if mode != "valid" {
				if !errors.Is(err, ErrUnknown) {
					t.Fatalf("unsafe sample accepted: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if p.RunID != f.r.RunID || p.CPUPercent != 50 || p.MemoryUsed != 1024 || p.MemoryTotal != 4096 || p.NetworkRx != 30 || p.NetworkTx != 60 {
				t.Fatalf("wrong metrics: %+v", p)
			}
		})
	}
}

func metricStatsFixture(id string) dockerStats {
	s := dockerStats{ID: id, Read: time.Date(2026, 9, 9, 0, 0, 2, 0, time.UTC), PreRead: time.Date(2026, 9, 9, 0, 0, 1, 0, time.UTC)}
	s.CPU.Usage.Total, s.PreviousCPU.Usage.Total = 150, 100
	s.CPU.System, s.PreviousCPU.System, s.CPU.Online = 1200, 1000, 2
	u, l := uint64(1024), uint64(4096)
	s.Memory.Usage, s.Memory.Limit = &u, &l
	json.Unmarshal([]byte(`{"eth0":{"rx_bytes":10,"tx_bytes":20},"eth1":{"rx_bytes":20,"tx_bytes":40}}`), &s.Networks)
	return s
}

func TestDockerMetricsUnavailableAndCounterBounds(t *testing.T) {
	s := metricStatsFixture("owned")
	s.CPU.Usage.Total = 1
	s.Memory.Usage = nil
	s.Networks = nil
	p := dockerMetricPoint("run", s)
	for _, field := range []string{"cpu", "memory", "network", "rss", "disk"} {
		if !slices.Contains(p.Unavailable, field) {
			t.Fatal("missing unavailable flag", field)
		}
	}
	if p.CPUPercent != 0 || p.MemoryUsed != 0 || p.NetworkRx != 0 {
		t.Fatal("invalid counters rendered as observations")
	}
	s = metricStatsFixture("owned")
	s.CPU.Online = 0
	s.CPU.Usage.PerCPU = []uint64{1, 2}
	if p := dockerMetricPoint("run", s); p.CPUPercent != 50 {
		t.Fatal("legacy CPU count fallback lost")
	}
	json.Unmarshal([]byte(`{"eth0":{"rx_bytes":9007199254740991,"tx_bytes":0},"eth1":{"rx_bytes":1,"tx_bytes":0}}`), &s.Networks)
	if p := dockerMetricPoint("run", s); !slices.Contains(p.Unavailable, "network") || p.NetworkRx != 0 {
		t.Fatal("inexact network sum exposed")
	}
}

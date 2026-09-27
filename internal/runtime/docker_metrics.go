package runtime

import (
	"context"
	"net/url"
	"time"

	"blora.dev/panel/internal/monitor"
)

type dockerCPUStats struct {
	Usage struct {
		Total  uint64   `json:"total_usage"`
		PerCPU []uint64 `json:"percpu_usage"`
	} `json:"cpu_usage"`
	System uint64 `json:"system_cpu_usage"`
	Online uint64 `json:"online_cpus"`
}
type dockerStats struct {
	ID          string         `json:"id"`
	Read        time.Time      `json:"read"`
	PreRead     time.Time      `json:"preread"`
	CPU         dockerCPUStats `json:"cpu_stats"`
	PreviousCPU dockerCPUStats `json:"precpu_stats"`
	Memory      struct {
		Usage *uint64 `json:"usage"`
		Limit *uint64 `json:"limit"`
	} `json:"memory_stats"`
	Networks map[string]struct {
		Rx uint64 `json:"rx_bytes"`
		Tx uint64 `json:"tx_bytes"`
	} `json:"networks"`
}

// ContainerMetrics samples the exact owned container and rejects a restart
// during sampling. stream=false waits for two CPU samples without a live stream.
func (m *Manager) ContainerMetrics(ctx context.Context, r Record) (monitor.ProcessPoint, error) {
	var point monitor.ProcessPoint
	if r.Backend != "docker" || r.ContainerID == "" {
		return point, ErrUnknown
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	before, err := m.inspectDocker(ctx, r)
	if err != nil {
		return point, err
	}
	if !before.State.Running || before.State.Restarting || before.State.Dead {
		return point, ErrUnknown
	}
	birth, err := time.Parse(time.RFC3339Nano, before.State.StartedAt)
	if err != nil || birth.IsZero() {
		return point, ErrUnknown
	}
	var stats dockerStats
	if err := m.docker.api(ctx, "GET", "/containers/"+url.PathEscape(r.ContainerID)+"/stats?stream=false", nil, &stats); err != nil {
		return point, err
	}
	after, err := m.inspectDocker(ctx, r)
	if err != nil {
		return point, err
	}
	if !after.State.Running || after.State.Restarting || after.State.Dead || before.State.StartedAt != after.State.StartedAt || stats.ID != r.ContainerID || stats.Read.Before(birth) || stats.Read.IsZero() {
		return point, ErrUnknown
	}
	if stats.PreRead.Before(birth) {
		stats.PreRead = time.Time{}
	}
	return dockerMetricPoint(r.RunID, stats), nil
}

func dockerMetricPoint(runID string, s dockerStats) monitor.ProcessPoint {
	p := monitor.ProcessPoint{RunID: runID, Point: monitor.Point{ObservedAt: s.Read, CPUBasis: "one-core", Unavailable: []string{"disk", "rss"}}}
	cpus := s.CPU.Online
	if cpus == 0 {
		cpus = uint64(len(s.CPU.Usage.PerCPU))
	}
	if cpus > 0 && cpus <= 1<<20 && !s.PreRead.IsZero() && s.Read.After(s.PreRead) && s.CPU.System > s.PreviousCPU.System && s.CPU.Usage.Total >= s.PreviousCPU.Usage.Total {
		p.CPUPercent = float64(s.CPU.Usage.Total-s.PreviousCPU.Usage.Total) / float64(s.CPU.System-s.PreviousCPU.System) * float64(cpus) * 100
	} else {
		p.Unavailable = append(p.Unavailable, "cpu")
	}
	const maxExact = uint64(9007199254740991)
	if s.Memory.Usage != nil && s.Memory.Limit != nil && *s.Memory.Usage <= maxExact && *s.Memory.Limit > 0 && *s.Memory.Limit <= maxExact {
		p.MemoryUsed, p.MemoryTotal = int64(*s.Memory.Usage), int64(*s.Memory.Limit)
	} else {
		p.Unavailable = append(p.Unavailable, "memory")
	}
	validNetwork := len(s.Networks) > 0
	for _, network := range s.Networks {
		if network.Rx > maxExact-p.NetworkRx || network.Tx > maxExact-p.NetworkTx {
			validNetwork = false
			break
		}
		p.NetworkRx += network.Rx
		p.NetworkTx += network.Tx
	}
	if !validNetwork {
		p.NetworkRx, p.NetworkTx = 0, 0
		p.Unavailable = append(p.Unavailable, "network")
	}
	return p
}

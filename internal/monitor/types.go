// Package monitor collects bounded node and owned-process observations.
package monitor

import "time"

type Point struct {
	ObservedAt  time.Time `json:"observedAt"`
	CPUPercent  float64   `json:"cpuPercent,omitempty"`
	CPUBasis    string    `json:"cpuBasis,omitempty"`
	MemoryUsed  int64     `json:"memoryUsedBytes,omitempty"`
	MemoryTotal int64     `json:"memoryTotalBytes,omitempty"`
	DiskUsed    int64     `json:"diskUsedBytes,omitempty"`
	DiskTotal   int64     `json:"diskTotalBytes,omitempty"`
	NetworkRx   uint64    `json:"networkRxBytes"`
	NetworkTx   uint64    `json:"networkTxBytes"`
	Stale       bool      `json:"stale"`
	Diagnostic  string    `json:"diagnostic,omitempty"`
	Unavailable []string  `json:"unavailable,omitempty"`
}

type ProcessPoint struct {
	Point
	PID          int    `json:"pid"`
	RunID        string `json:"runId"`
	RSS          int64  `json:"rssBytes,omitempty"`
	Scope        string `json:"scope,omitempty"`
	ProcessCount int    `json:"processCount,omitempty"`
}

type SystemProcess struct {
	PID         int      `json:"pid"`
	StartTicks  uint64   `json:"startTicks"`
	Command     string   `json:"command"`
	UID         string   `json:"uid,omitempty"`
	RSS         int64    `json:"rssBytes,omitempty"`
	Unavailable []string `json:"unavailable,omitempty"`
}

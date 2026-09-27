//go:build !linux && !windows

package monitor

import (
	"errors"
	"time"
)

func (c *Collector) Node(root string) (Point, error) {
	return c.add(Point{ObservedAt: time.Now().UTC(), Unavailable: []string{"cpu", "memory", "disk", "network"}}), nil
}
func (c *Collector) Process(pid int, runID string, expectedBirth ...uint64) (ProcessPoint, error) {
	return ProcessPoint{Point: Point{ObservedAt: time.Now().UTC(), Unavailable: []string{"cpu", "memory"}}, PID: pid, RunID: runID}, errors.New("owned process metrics unavailable on this platform")
}

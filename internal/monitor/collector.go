package monitor

import (
	"sync"
	"time"
)

type Collector struct {
	mu              sync.Mutex
	previous        cpuTicks
	history         []Point
	processPrevious map[int]processTicks
}
type processTicks struct {
	birth, total uint64
	system       uint64
	runID        string
	at           time.Time
}
type cpuTicks struct{ total, idle uint64 }

func New() *Collector { return &Collector{history: make([]Point, 0, 120)} }

func (c *Collector) add(p Point) Point {
	c.mu.Lock()
	defer c.mu.Unlock()
	p.Stale = false
	c.history = append(c.history, p)
	if len(c.history) > 120 {
		c.history = c.history[len(c.history)-120:]
	}
	return p
}
func (c *Collector) History() []Point {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Point, len(c.history))
	copy(out, c.history)
	return out
}

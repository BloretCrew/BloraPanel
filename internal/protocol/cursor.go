package protocol

import "sync"

// Cursor must be instantiated per resource (or durable event log). Once a gap
// is detected no later event is accepted until a fresh snapshot is installed.
type Cursor struct {
	mu               sync.Mutex
	revision         uint64
	initialized, gap bool
}

func (c *Cursor) Snapshot(revision uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.revision = revision
	c.initialized = true
	c.gap = false
}

// Event reports whether this event is new. Replayed older events are ignored.
func (c *Cursor) Event(revision uint64) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.initialized || c.gap {
		return false, ErrGap
	}
	if revision <= c.revision {
		return false, nil
	}
	if revision-c.revision != 1 {
		c.gap = true
		return false, ErrGap
	}
	c.revision = revision
	return true, nil
}

func (c *Cursor) Revision() (revision uint64, needsSnapshot bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.revision, !c.initialized || c.gap
}

// ResumeAvailable validates a byte checkpoint against the archive retention
// interval [firstRetained, end]. It never invents missing history.
func ResumeAvailable(checkpoint, firstRetained, end uint64) error {
	if firstRetained > end || checkpoint < firstRetained || checkpoint > end {
		return ErrGap
	}
	return nil
}

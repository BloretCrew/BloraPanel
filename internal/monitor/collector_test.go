package monitor

import (
	"testing"
	"time"
)

func TestNodeHistoryIsBoundedAndTimestamped(t *testing.T) {
	c := New()
	for i := 0; i < 140; i++ {
		p, err := c.Node("/tmp")
		if err != nil {
			t.Fatal(err)
		}
		if p.ObservedAt.IsZero() {
			t.Fatal("missing timestamp")
		}
		time.Sleep(time.Microsecond)
	}
	h := c.History()
	if len(h) != 120 {
		t.Fatalf("history=%d", len(h))
	}
	if !h[0].ObservedAt.Before(h[len(h)-1].ObservedAt) || h[0].Stale {
		t.Fatal("history ordering/age invalid")
	}
}

//go:build windows

package terminal

import (
	"bytes"
	"context"
	"testing"
	"time"
)

// Run on an actual supported Windows host; cross compilation is not execution.
func TestWindowsConPTYRealCommandResizeAndJobClose(t *testing.T) {
	ctx := context.Background()
	m, err := New(Options{Root: t.TempDir(), AllowNative: true})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	r := testRequest()
	r.Backend = "native"
	r.Command = []string{"cmd.exe", "/Q", "/D"}
	s, err := m.Create(ctx, r)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.Attach(ctx, s.ID, "owner", "view", 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err = m.AcquireLease(s.ID, "owner", "view", time.Minute, false); err != nil {
		t.Fatal(err)
	}
	if err = m.Resize(ctx, s.ID, "owner", "view", 132, 42); err != nil {
		t.Fatal(err)
	}
	if err = m.WriteInput(ctx, s.ID, "owner", "view", []byte("chcp 65001\r\necho 中文-ConPTY-marker\r\n")); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(8 * time.Second)
	var out []byte
	var after uint64
	for time.Now().Before(deadline) {
		b, err := m.Read(ctx, s.ID, "owner", "view", after, 0)
		if err != nil {
			t.Fatal(err)
		}
		after = b.Next
		for _, e := range b.Events {
			out = append(out, e.Data...)
		}
		if bytes.Contains(out, []byte("中文-ConPTY-marker")) {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !bytes.Contains(out, []byte("中文-ConPTY-marker")) {
		t.Fatalf("ConPTY Unicode output missing: %q", out)
	}
	closeCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	if err = m.CloseSession(closeCtx, s.ID, "owner"); err != nil {
		t.Fatal(err)
	}
	got, err := m.Get(ctx, s.ID, "owner")
	if err != nil || got.State != "exited" {
		t.Fatalf("Job exit not confirmed: %+v %v", got, err)
	}
}

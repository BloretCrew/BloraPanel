//go:build linux || windows

package runlog

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"blora.dev/panel/internal/terminal"
)

func TestRunlogHelperProcess(t *testing.T) {
	for i, arg := range os.Args {
		if arg == "--runlog-root" {
			if err := RunHelper(os.Args[i:]); err != nil {
				os.Exit(31)
			}
			os.Exit(0)
		}
	}
}

func helperCommand() []string {
	return []string{os.Args[0], "-test.run=^TestRunlogHelperProcess$", "--"}
}

func startCapture(t *testing.T, size int64) *Capture {
	t.Helper()
	c, err := Start(context.Background(), Options{Root: t.TempDir(), RunID: "run-test", Command: helperCommand(), Archive: terminal.ArchiveOptions{MaxBytes: size, SegmentBytes: size / 4}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = c.ReleaseWriter()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = c.Finish(ctx)
	})
	return c
}

func waitLatest(t *testing.T, c *Capture, atLeast uint64) Status {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		s, err := c.Status(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if s.Latest >= atLeast {
			return s
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("output did not reach durable archive")
	return Status{}
}

func TestPersistentRawOutputAndPrivateIPC(t *testing.T) {
	c := startCapture(t, 128<<10)
	r, err := loadRecord(c.root, c.runID)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.Get("http://" + r.Address + "/status")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status %d", res.StatusCode)
	}
	raw := []byte("中文\x1b[31m红色\x1b[0m\r\n\x00raw")
	if _, err = c.Stdout().Write(raw); err != nil {
		t.Fatal(err)
	}
	waitLatest(t, c, 1)
	status, err := c.Status(context.Background())
	if err != nil || status.InputAvailable {
		t.Fatal("stdout-only capture advertises input")
	}
	recovered, err := Open(filepath.Dir(c.root), c.runID)
	if err != nil {
		t.Fatal(err)
	}
	b, err := recovered.Read(context.Background(), 0, 4096)
	if err != nil {
		t.Fatal(err)
	}
	var all []byte
	for _, event := range b.Events {
		all = append(all, event.Data...)
	}
	if !bytes.Equal(all, raw) || b.Gap {
		t.Fatalf("raw replay mismatch: %q", all)
	}
	if _, err = recovered.Read(context.Background(), b.Latest+1, 4096); !errors.Is(err, terminal.ErrCursor) {
		t.Fatalf("future cursor: %v", err)
	}
	if err = c.ReleaseWriter(); err != nil {
		t.Fatal(err)
	}
	if err = c.Finish(context.Background()); err != nil {
		t.Fatal(err)
	}
	recovered, err = Open(filepath.Dir(c.root), c.runID)
	if err != nil {
		t.Fatal(err)
	}
	b, err = recovered.Read(context.Background(), 0, 4096)
	if err != nil || len(b.Events) == 0 {
		t.Fatalf("finished archive: %v", err)
	}
	s, err := recovered.Status(context.Background())
	if err != nil || s.Phase != "complete" {
		t.Fatalf("completion: %+v %v", s, err)
	}
}

func TestRotationNoReaderAndExclusiveRun(t *testing.T) {
	c := startCapture(t, 64<<10)
	if _, err := Start(context.Background(), Options{Root: filepath.Dir(c.root), RunID: c.runID, Command: helperCommand()}); !errors.Is(err, os.ErrExist) {
		t.Fatalf("duplicate run accepted: %v", err)
	}
	// No IPC/browser reader is attached during a megabyte of real pipe output.
	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(c.Stdout(), strings.NewReader(strings.Repeat("bounded-output\n", 80000)))
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("stdout blocked without browser")
	}
	if err := c.ReleaseWriter(); err != nil {
		t.Fatal(err)
	}
	if err := c.Finish(context.Background()); err != nil {
		t.Fatal(err)
	}
	s, err := c.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if s.Bytes > 64<<10 || s.Earliest <= 1 {
		t.Fatalf("retention: %+v", s)
	}
	b, err := c.Read(context.Background(), 0, 64<<10)
	if err != nil || !b.Gap || b.Earliest != s.Earliest {
		t.Fatalf("gap %+v %v", b, err)
	}
	if b.Next <= b.Earliest || b.Next > b.Latest {
		t.Fatalf("cursor %+v", b)
	}
	t.Logf("produced 1,200,000 bytes without a browser; retained %d; event bounds %d..%d", s.Bytes, s.Earliest, s.Latest)
}

func TestBirthGuardAndIncompleteDrain(t *testing.T) {
	c := startCapture(t, 64<<10)
	r, err := loadRecord(c.root, c.runID)
	if err != nil {
		t.Fatal(err)
	}
	wrong := r.Identity
	wrong.Birth++
	if err = terminateIdentity(wrong); err != nil {
		t.Fatal(err)
	}
	alive, err := identityAlive(r.Identity)
	if err != nil || !alive {
		t.Fatalf("mismatched birth killed helper: %v", err)
	}
	// This duplicate simulates a leaked writer AFTER the actual run exited.
	// Finish must report incomplete capture and target only the helper.
	duplicate, err := duplicateFile(c.Stdout())
	if err != nil {
		t.Fatal(err)
	}
	defer duplicate.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err = c.Finish(ctx); !errors.Is(err, ErrIncomplete) {
		t.Fatalf("unclosed pipe counted as drained: %v", err)
	}
	alive, err = identityAlive(r.Identity)
	if err != nil || alive {
		t.Fatalf("helper termination unconfirmed: %v", err)
	}
	recovered, err := Open(filepath.Dir(c.root), c.runID)
	if err != nil {
		t.Fatal(err)
	}
	s, err := recovered.Status(context.Background())
	if err != nil || s.Phase != "invalid" || s.Diagnostic == "" {
		t.Fatalf("lost helper appeared healthy: %+v %v", s, err)
	}
	if _, err = recovered.Read(context.Background(), 0, 4096); err != nil {
		t.Fatalf("confirmed-dead helper historical archive unavailable: %v", err)
	}
}

package filesystem

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
)

func TestSymlinkFIFOAndDirectorySwapStayConfined(t *testing.T) {
	s, root, _ := testService(t, Options{})
	ctx := context.Background()
	outside := t.TempDir()
	put(t, outside, "secret", "outside-secret")
	put(t, root, "inside", "inside-data")
	if e := os.Symlink(filepath.Join(outside, "secret"), filepath.Join(root, "link")); e != nil {
		t.Fatal(e)
	}
	if _, e := s.ReadText(ctx, "link"); !errors.Is(e, ErrUnsupported) {
		t.Fatal(e)
	}
	if e := os.Symlink("inside", filepath.Join(root, "internal-link")); e != nil {
		t.Fatal(e)
	}
	if _, e := s.ReadText(ctx, "internal-link"); !errors.Is(e, ErrUnsupported) {
		t.Fatal(e)
	}
	if e := syscall.Mkfifo(filepath.Join(root, "fifo"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := s.ReadText(ctx, "fifo"); !errors.Is(e, ErrUnsupported) {
		t.Fatal(e)
	}
	put(t, root, "racy/secret", "inside-data")
	var wg sync.WaitGroup
	stop := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			if os.Rename(filepath.Join(root, "racy"), filepath.Join(root, "parked")) == nil {
				os.Symlink(outside, filepath.Join(root, "racy"))
				os.Remove(filepath.Join(root, "racy"))
				os.Rename(filepath.Join(root, "parked"), filepath.Join(root, "racy"))
			}
		}
	}()
	for range 200 {
		got, e := s.ReadText(ctx, "racy/secret")
		if e == nil && got.Content != "inside-data" {
			t.Errorf("escaped root: %q", got.Content)
		}
		_, _ = s.WriteText(ctx, "racy/new", "local", MissingVersion)
	}
	close(stop)
	wg.Wait()
	content(t, outside, "secret", "outside-secret")
	missing(t, outside, "new")
}

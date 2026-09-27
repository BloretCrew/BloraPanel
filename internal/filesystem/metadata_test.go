package filesystem

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMetadataVersionCancellationAndLinkBoundary(t *testing.T) {
	s, root, _ := testService(t, Options{})
	put(t, root, "dir/file", "content")
	ctx := context.Background()
	stamp := time.Unix(1700000000, 123456000)
	v := version(t, s, "dir")
	if _, err := s.Metadata(ctx, "dir", v, 0710, stamp); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(root, "dir"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0710 || !info.ModTime().Equal(stamp) {
		t.Fatalf("%v %v", info.Mode(), info.ModTime())
	}
	if _, err = s.Metadata(ctx, "dir", v, 0777, stamp); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale metadata: %v", err)
	}
	v = version(t, s, "dir/file")
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err = s.Metadata(cancelled, "dir/file", v, 0777, stamp); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
	if _, err = s.Metadata(ctx, "dir/file", v, 04755, stamp); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("special bits: %v", err)
	}
	outside := t.TempDir()
	if err = os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Metadata(ctx, "link", v, 0777, stamp); err == nil {
		t.Fatal("metadata followed link")
	}
}

func TestUploadMetadataIsBoundToDurableIdentity(t *testing.T) {
	s, root, state := testService(t, Options{})
	body := []byte("metadata content")
	spec := uploadSpec(body, "program.sh")
	spec.PreserveMetadata = true
	spec.SourceMode = 0750
	spec.SourceModifiedNano = time.Unix(1700000000, 123456000).UnixNano()
	sendAll(t, s, spec, body)
	changed := spec
	changed.SourceMode = 0777
	if _, err := s.BeginUpload(context.Background(), changed); !errors.Is(err, ErrTransfer) {
		t.Fatalf("metadata source changed: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := New(root, Options{StateDir: state})
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if _, err = reopened.CommitUpload(context.Background(), spec.ID); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(root, spec.Path))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0750 || info.ModTime().UnixNano() != spec.SourceModifiedNano {
		t.Fatalf("committed metadata %v %v", info.Mode(), info.ModTime())
	}
}

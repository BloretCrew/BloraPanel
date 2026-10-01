//go:build windows

package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWindowsMetadataAttributesAndTimes(t *testing.T) {
	s, root, _ := testService(t, Options{})
	put(t, root, "dir/file", "retained content")
	for _, path := range []string{"dir", "dir/file"} {
		t.Run(path, func(t *testing.T) {
			for _, mode := range []uint32{0444, 0666} {
				stamp := time.Unix(1700000000+int64(mode), 123456000)
				before := relation(t, s, path)
				if _, err := s.Metadata(context.Background(), path, version(t, s, path), mode, stamp); err != nil {
					t.Fatal(err)
				}
				info, err := os.Stat(filepath.Join(root, filepath.FromSlash(path)))
				if err != nil {
					t.Fatal(err)
				}
				want := os.FileMode(mode)
				if info.IsDir() {
					want |= 0111
				}
				if info.Mode().Perm() != want || !info.ModTime().Equal(stamp) {
					t.Fatalf("mode/time = %v %s, want %v %s", info.Mode(), info.ModTime(), want, stamp)
				}
				if after := relation(t, s, path); after.ObjectID != before.ObjectID {
					t.Fatal("metadata changed object identity")
				}
			}
		})
	}
	content(t, root, "dir/file", "retained content")
}

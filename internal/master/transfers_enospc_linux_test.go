//go:build linux

package master

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
	"golang.org/x/sys/unix"
)

// Opt in because the host must allow private user/mount namespaces. The child
// verifies namespace separation before mounting only a fixture-owned directory.
func privateENOSPCNamespace(t *testing.T) bool {
	t.Helper()
	if os.Getenv("BLORA_TEST_ENOSPC") != "1" {
		t.Skip("requires BLORA_TEST_ENOSPC=1 and private Linux user/mount namespaces")
	}
	namespace, err := os.Readlink("/proc/self/ns/mnt")
	if err != nil {
		t.Fatal(err)
	}
	parent := os.Getenv("BLORA_ENOSPC_PARENT_NAMESPACE")
	if parent == "" {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "unshare", "--user", "--map-root-user", "--mount", os.Args[0], "-test.run=^"+regexp.QuoteMeta(t.Name())+"$", "-test.v")
		cmd.Env = append(os.Environ(), "BLORA_ENOSPC_PARENT_NAMESPACE="+namespace)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("private ENOSPC test: %v\n%s", err, output)
		}
		t.Logf("%s", output)
		return false
	}
	if namespace == parent {
		t.Fatal("refusing to mount in the parent namespace")
	}
	if err := unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
		t.Fatal(err)
	}
	return true
}

func TestTransferTargetENOSPCKeepsMoveSource(t *testing.T) {
	if !privateENOSPCNamespace(t) {
		return
	}
	f := newFileFixture(t)
	if err := unix.Mount("blora-enospc-test", f.roots[1], "tmpfs", unix.MS_NODEV|unix.MS_NOSUID|unix.MS_NOEXEC, "size=1048576,mode=0750"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := unix.Unmount(f.roots[1], unix.MNT_DETACH); err != nil {
			t.Errorf("unmount private test filesystem: %v", err)
		}
	})
	body := bytes.Repeat([]byte("actual ENOSPC source must survive\n"), 65536)
	transferWrite(t, f.roots[0], "source.bin", body)
	transferWrite(t, f.roots[1], "keep.txt", []byte("existing target data"))
	task := startTransfer(t, f, "source.bin", "destination.bin", true, model.ID())
	result := awaitTransfer(t, f, task, model.Failed)
	if result.DestinationVerified || result.SourceDeleted || result.CleanupPending {
		t.Fatalf("unsafe failed move result: %+v", result)
	}
	if !strings.Contains(strings.ToLower(result.Error), "no space left") {
		t.Fatalf("expected actual ENOSPC diagnostic, got %+v", result)
	}
	got, err := os.ReadFile(filepath.Join(f.roots[0], "source.bin"))
	if err != nil || !bytes.Equal(got, body) {
		t.Fatalf("source changed after target ENOSPC: %v", err)
	}
	if _, err := os.Stat(filepath.Join(f.roots[1], "destination.bin")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("failed destination was published: %v", err)
	}
	assertDisk(t, f.roots[1], "keep.txt", "existing target data")
	entries, err := os.ReadDir(f.roots[1])
	if err != nil || len(entries) != 2 || entries[0].Name() != ".blora-files" || entries[1].Name() != "keep.txt" {
		t.Fatalf("target staging data not cleaned: %v %v", entries, err)
	}
	for _, name := range []string{"work", "uploads", "trash"} {
		entries, err := os.ReadDir(filepath.Join(f.roots[1], ".blora-files", name))
		if err != nil || len(entries) != 0 {
			t.Fatalf("private %s data not cleaned: %v %v", name, entries, err)
		}
	}
}

package extensions

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBackendSandboxExecutesAndBoundsGuest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "backend.wasm")
	build := exec.Command("go", "build", "-buildvcs=false", "-o", path, "./testdata/backend")
	build.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm", "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build WASI guest: %v: %s", err, out)
	}
	module, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	out, err := RunBackend(context.Background(), module, json.RawMessage(`{"value":7}`))
	if err != nil || !strings.Contains(string(out), `"squared":49`) {
		t.Fatalf("compute: %s %v", out, err)
	}
	out, err = RunBackend(context.Background(), module, json.RawMessage(`{"mode":"isolation"}`))
	if err != nil || !strings.Contains(string(out), `"fileDenied":true`) || !strings.Contains(string(out), `"environment":0`) {
		t.Fatalf("isolation: %s %v", out, err)
	}
	if _, err := RunBackend(context.Background(), module, json.RawMessage(`{"mode":"overflow"}`)); err == nil || !strings.Contains(err.Error(), "output limit") {
		t.Fatalf("overflow: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := RunBackend(ctx, module, json.RawMessage(`{"mode":"loop"}`)); err == nil || ctx.Err() == nil {
		t.Fatalf("loop cancellation: %v", err)
	}
}

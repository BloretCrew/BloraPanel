package extensions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
)

const MaxBackendModule = 8 << 20

// Leave room for task metadata inside the 64 KiB control-message contract.
const MaxBackendIO = 32 << 10

type boundedOutput struct {
	bytes.Buffer
	overflow bool
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	if len(p) > MaxBackendIO-b.Len() {
		b.overflow = true
		return 0, errors.New("extension output limit exceeded")
	}
	return b.Buffer.Write(p)
}

// RunBackend executes a WASI command using only JSON stdin/stdout. No host
// filesystem, environment, sockets, process execution or custom imports are
// configured. Every invocation has fresh linear memory and a hard deadline.
func RunBackend(ctx context.Context, module []byte, input json.RawMessage) (json.RawMessage, error) {
	if len(input) > MaxBackendIO || !json.Valid(input) {
		return nil, errors.New("invalid extension backend input")
	}
	runner, err := compileBackend(ctx, module)
	if err != nil {
		return nil, err
	}
	defer runner.close()
	return runner.run(ctx, input)
}

type backendRunner struct {
	runtime  wazero.Runtime
	compiled wazero.CompiledModule
}

func (r *backendRunner) close() { _ = r.runtime.Close(context.Background()) }
func compileBackend(ctx context.Context, module []byte) (*backendRunner, error) {
	if len(module) == 0 || len(module) > MaxBackendModule {
		return nil, errors.New("invalid extension backend module size")
	}
	compileCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithMemoryLimitPages(1024).WithCloseOnContextDone(true))
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		_ = r.Close(context.Background())
		return nil, err
	}
	compiled, err := r.CompileModule(compileCtx, module)
	if err != nil {
		_ = r.Close(context.Background())
		return nil, fmt.Errorf("extension compilation failed: %w", err)
	}
	return &backendRunner{runtime: r, compiled: compiled}, nil
}
func (runner *backendRunner) run(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	if len(input) > MaxBackendIO || !json.Valid(input) {
		return nil, errors.New("invalid extension backend input")
	}
	executionCtx, stop := context.WithTimeout(ctx, 5*time.Second)
	defer stop()
	var out, diagnostics boundedOutput
	config := wazero.NewModuleConfig().WithName("").WithStdin(bytes.NewReader(input)).WithStdout(&out).WithStderr(&diagnostics)
	instance, err := runner.runtime.InstantiateModule(executionCtx, runner.compiled, config)
	if instance != nil {
		defer instance.Close(context.Background())
	}
	if executionCtx.Err() != nil {
		return nil, fmt.Errorf("extension execution stopped: %w", executionCtx.Err())
	}
	if out.overflow || diagnostics.overflow {
		return nil, errors.New("extension output limit exceeded")
	}
	if err != nil {
		var exit *sys.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 0 {
			return nil, fmt.Errorf("extension execution failed: %w", err)
		}
	}
	if !json.Valid(out.Bytes()) {
		return nil, errors.New("extension backend must return one JSON value")
	}
	return append(json.RawMessage(nil), out.Bytes()...), nil
}

package runlog

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"blora.dev/panel/internal/terminal"
)

// RunHelper implements the service's optional hidden --log-helper command.
// Its caller must immediately exit after return, before opening daemon stores
// or connections. Arguments are exactly --runlog-root PATH.
func RunHelper(args []string) error {
	if len(args) != 2 || args[0] != "--runlog-root" || !filepath.IsAbs(args[1]) {
		return errors.New("invalid log helper arguments")
	}
	root := filepath.Clean(args[1])
	runID := filepath.Base(root)
	r, err := loadRecord(root, runID)
	if err != nil {
		return err
	}
	token := os.Getenv("BLORA_RUNLOG_TOKEN")
	if len(token) != 64 || subtle.ConstantTimeCompare([]byte(token), []byte(r.Token)) != 1 || r.Phase != "intent" {
		return ErrUnknown
	}
	if err = validateIndependent(); err != nil {
		return err
	}
	r.Identity, err = selfIdentity()
	if err != nil {
		return err
	}
	a, err := terminal.OpenArchive(filepath.Join(root, "archive"), r.Archive)
	if err != nil {
		return err
	}
	defer a.Close()
	var input *heldInput
	if r.HoldInput {
		input, err = newHeldInput()
		if err != nil {
			return err
		}
		defer input.close()
		r.Input, err = input.descriptor()
		if err != nil {
			return err
		}
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer listener.Close()
	r.Address = listener.Addr().String()
	r.Phase = "running"
	if err = saveRecord(root, r); err != nil {
		return err
	}
	var mu sync.Mutex
	limit := make(chan struct{}, 8)
	inputGate := make(chan struct{}, 1)
	var inputAvailable atomic.Bool
	inputAvailable.Store(input != nil)
	finishRequested := make(chan struct{})
	var finishOnce sync.Once
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if subtle.ConstantTimeCompare([]byte(req.Header.Get("Authorization")), []byte("Bearer "+token)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		select {
		case limit <- struct{}{}:
			defer func() { <-limit }()
		default:
			http.Error(w, "busy", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if ((req.URL.Path == "/read" || req.URL.Path == "/status") && req.Method != http.MethodGet) || ((req.URL.Path == "/input" || req.URL.Path == "/finish") && req.Method != http.MethodPost) {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		switch req.URL.Path {
		case "/input":
			if input == nil || !inputAvailable.Load() {
				http.Error(w, "input unavailable", http.StatusConflict)
				return
			}
			data, e := io.ReadAll(http.MaxBytesReader(w, req.Body, 32<<10))
			if e != nil {
				http.Error(w, "input budget exceeded", http.StatusRequestEntityTooLarge)
				return
			}
			ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
			defer cancel()
			select {
			case inputGate <- struct{}{}:
				defer func() { <-inputGate }()
			case <-ctx.Done():
				http.Error(w, "input busy", http.StatusServiceUnavailable)
				return
			}
			select {
			case <-finishRequested:
				http.Error(w, "capture finishing", http.StatusGone)
				return
			default:
			}
			n, e := input.write(ctx, data)
			if errors.Is(e, ErrIncomplete) || errors.Is(e, os.ErrClosed) {
				inputAvailable.Store(false)
			}
			result := inputResult{Written: n}
			if e != nil {
				result.Error = e.Error()
			}
			_ = json.NewEncoder(w).Encode(result)
		case "/finish":
			if input == nil {
				http.Error(w, "not an input holder", http.StatusConflict)
				return
			}
			finishOnce.Do(func() { close(finishRequested) })
			inputAvailable.Store(false)
			_ = json.NewEncoder(w).Encode(struct {
				Accepted bool `json:"accepted"`
			}{true})
		case "/read":
			after, e := strconv.ParseUint(req.URL.Query().Get("after"), 10, 64)
			if e != nil {
				http.Error(w, "invalid cursor", http.StatusBadRequest)
				return
			}
			budget, e := strconv.Atoi(req.URL.Query().Get("maxBytes"))
			if e != nil {
				http.Error(w, "invalid budget", http.StatusBadRequest)
				return
			}
			batch, e := a.Read(after, budget)
			if e != nil {
				code := http.StatusInternalServerError
				if errors.Is(e, terminal.ErrCursor) {
					code = http.StatusConflict
				}
				http.Error(w, "archive unavailable", code)
				return
			}
			_ = json.NewEncoder(w).Encode(batch)
		case "/status":
			earliest, latest := a.Bounds()
			bytes := a.Bytes()
			mu.Lock()
			s := Status{RunID: runID, Phase: r.Phase, Earliest: earliest, Latest: latest, Bytes: bytes, Diagnostic: r.Diagnostic, InputAvailable: inputAvailable.Load()}
			mu.Unlock()
			_ = json.NewEncoder(w).Encode(s)
		default:
			http.NotFound(w, req)
		}
	})
	server := &http.Server{Handler: handler, ReadHeaderTimeout: time.Second, ReadTimeout: 2 * time.Second, WriteTimeout: 3 * time.Second, IdleTimeout: time.Second, MaxHeaderBytes: 4096}
	go func() { _ = server.Serve(listener) }()
	defer server.Close()
	// This is the only write to the startup pipe. Losing the parent after
	// readiness cannot cause the helper to write to a dead daemon descriptor.
	if _, err = io.WriteString(os.Stdout, "BLORA-RUNLOG/1 READY\n"); err != nil {
		return err
	}
	_ = os.Stdout.Close()
	buffer := make([]byte, terminal.MaxEventBytes)
	var captureErr error
	for {
		n, readErr := os.Stdin.Read(buffer)
		if n > 0 && captureErr == nil {
			captureErr = a.AppendOutput(buffer[:n])
			if captureErr != nil {
				mu.Lock()
				r.Phase = "degraded"
				r.Diagnostic = "archive write failed; remaining output is being drained: " + captureErr.Error()
				_ = saveRecord(root, r)
				mu.Unlock()
			}
		}
		// Keep draining after a storage failure, so losing logging does not
		// deliberately SIGPIPE the still-running business process.
		if readErr != nil {
			if !errors.Is(readErr, io.EOF) && captureErr == nil {
				captureErr = readErr
			}
			break
		}
	}
	if input != nil {
		mu.Lock()
		r.Phase = "output_complete"
		if captureErr != nil {
			r.Phase = "output_failed"
			r.Diagnostic = captureErr.Error()
		}
		err = saveRecord(root, r)
		mu.Unlock()
		// Closing stdout is not a business exit. Preserve stdin even with no
		// output until the runtime explicitly confirms Finish.
		<-finishRequested
		inputGate <- struct{}{}
		input.close()
		<-inputGate
		if err != nil {
			return err
		}
	}
	_ = a.Close()
	mu.Lock()
	r.Phase = "complete"
	if captureErr != nil {
		r.Phase = "failed"
		r.Diagnostic = captureErr.Error()
	}
	err = saveRecord(root, r)
	mu.Unlock()
	if err != nil {
		return err
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdown)
	if captureErr != nil {
		return fmt.Errorf("log capture: %w", captureErr)
	}
	return nil
}

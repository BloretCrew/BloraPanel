// Package runlog keeps native stdout capture in a separate operating-system
// process. Losing the daemon must not close a running instance's pipe reader.
// Each run owns one helper and one bounded terminal archive.
package runlog

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"blora.dev/panel/internal/terminal"
)

var (
	ErrUnknown    = errors.New("log helper identity cannot be confirmed")
	ErrIncomplete = errors.New("log capture ended without confirmed complete drain")
	validID       = regexp.MustCompile(`^[A-Za-z0-9_-]{1,96}$`)
)

type Options struct {
	Root  string
	RunID string
	// Command is an administrator-selected executable and optional arguments,
	// e.g. []string{daemonPath, "--log-helper"}. Never accept it from an instance.
	// An empty command looks up blora-log-helper on the service PATH.
	Command        []string
	Archive        terminal.ArchiveOptions
	StartupTimeout time.Duration
	HoldInput      bool // Keep the business stdin writer alive independently until confirmed Finish.
}

type inputDescriptor struct {
	Handle uint64 `json:"handle"`
	Device uint64 `json:"device,omitempty"`
	Inode  uint64 `json:"inode,omitempty"`
}

type Identity struct {
	PID    int    `json:"pid"`
	Birth  uint64 `json:"birth"`
	BootID string `json:"bootId,omitempty"`
}

type record struct {
	Version    int                     `json:"version"`
	RunID      string                  `json:"runId"`
	Token      string                  `json:"token"`
	Identity   Identity                `json:"identity"`
	Address    string                  `json:"address,omitempty"`
	Archive    terminal.ArchiveOptions `json:"archive"`
	Phase      string                  `json:"phase"`
	Diagnostic string                  `json:"diagnostic,omitempty"`
	HoldInput  bool                    `json:"holdInput,omitempty"`
	Input      inputDescriptor         `json:"input,omitempty"`
}

type Status struct {
	RunID          string `json:"runId"`
	Phase          string `json:"phase"`
	Earliest       uint64 `json:"earliest"`
	Latest         uint64 `json:"latest"`
	Bytes          int64  `json:"bytes"`
	Diagnostic     string `json:"diagnostic,omitempty"`
	InputAvailable bool   `json:"inputAvailable"`
}

type Capture struct {
	root        string
	runID       string
	mu          sync.Mutex
	writer      *os.File
	inputReader *os.File
}

func runRoot(root, runID string) (string, error) {
	if root == "" || !validID.MatchString(runID) {
		return "", errors.New("log root and valid run ID are required")
	}
	base, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	return filepath.Join(base, runID), nil
}

func loadRecord(root, runID string) (record, error) {
	var r record
	f, err := os.Open(filepath.Join(root, "capture.json"))
	if err != nil {
		return r, err
	}
	defer f.Close()
	err = json.NewDecoder(io.LimitReader(f, 8192)).Decode(&r)
	if err == nil && (r.Version != 1 || r.RunID != runID || len(r.Token) != 64) {
		err = ErrUnknown
	}
	return r, err
}

func readableRecord(root, runID string) (record, error) {
	r, err := loadRecord(root, runID)
	if err != nil || finished(r.Phase) {
		return r, err
	}
	alive, err := identityAlive(r.Identity)
	if err != nil {
		return r, err
	}
	if !alive {
		// The original writer is positively gone. Historical complete records
		// may be recovered, but no new pipe/helper can restore live capture.
		r.Phase = "invalid"
		r.Diagnostic = "log helper exited without a durable complete drain; historical output may be incomplete"
	}
	return r, nil
}

func saveRecord(root string, r record) error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(root, ".capture-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(b); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err = os.Rename(f.Name(), filepath.Join(root, "capture.json")); err != nil {
		return err
	}
	return syncDirectory(root)
}

// Start launches an independent helper BEFORE launching the business process.
// Pass Stdout() as BOTH runtime.IO.Stdout and runtime.IO.Stderr, then call
// ReleaseWriter after runtime.Start returns. Command/context lifetime is not
// attached to the helper after this function returns.
func Start(ctx context.Context, options Options) (*Capture, error) {
	root, err := runRoot(options.Root, options.RunID)
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	if err = os.MkdirAll(filepath.Dir(root), 0700); err != nil {
		return nil, err
	}
	if err = os.Mkdir(root, 0700); err != nil {
		return nil, fmt.Errorf("claim run log directory (never replace an existing capture): %w", err)
	}
	var token [32]byte
	if _, err = rand.Read(token[:]); err != nil {
		return nil, err
	}
	r := record{Version: 1, RunID: options.RunID, Token: hex.EncodeToString(token[:]), Archive: options.Archive, Phase: "intent", HoldInput: options.HoldInput}
	if err = saveRecord(root, r); err != nil {
		return nil, err
	}
	command := options.Command
	if len(command) == 0 {
		command = []string{"blora-log-helper"}
	}
	read, write, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer read.Close()
	readyRead, readyWrite, err := os.Pipe()
	if err != nil {
		write.Close()
		return nil, err
	}
	defer readyRead.Close()
	defer readyWrite.Close()
	args := append(append([]string{}, command[1:]...), "--runlog-root", root)
	cmd := exec.Command(command[0], args...)
	cmd.Stdin, cmd.Stdout = read, readyWrite
	cmd.Env = helperEnvironment(r.Token)
	if err = independentProcess(cmd); err != nil {
		write.Close()
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		write.Close()
		return nil, fmt.Errorf("start independent log helper: %w", err)
	}
	helperPID := cmd.Process.Pid
	started := false
	defer func() {
		if !started {
			// No business process can own this pipe until Start succeeds. The
			// os.Process handle belongs to the exact child created above.
			_ = cmd.Process.Kill()
		}
	}()
	readyWrite.Close()
	// Wait reaps only our helper; it does not read/copy business output and does
	// not bind helper lifetime to this daemon's lifetime.
	go func() { _ = cmd.Wait() }()
	ready := make(chan error, 1)
	go func() {
		line, e := bufio.NewReaderSize(readyRead, 1024).ReadSlice('\n')
		if e == nil && string(line) != "BLORA-RUNLOG/1 READY\n" {
			e = ErrUnknown
		}
		ready <- e
	}()
	duration := options.StartupTimeout
	if duration <= 0 || duration > 30*time.Second {
		duration = 5 * time.Second
	}
	deadline, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	select {
	case err = <-ready:
	case <-deadline.Done():
		err = deadline.Err()
	}
	if err != nil {
		write.Close()
		// No instance has received the writer yet. Closing it lets an already
		// initialized helper drain to EOF. A pending startup has a finite bound.
		return nil, fmt.Errorf("log helper readiness: %w", err)
	}
	current, err := loadRecord(root, options.RunID)
	if err == nil && (current.Token != r.Token || current.Identity.PID != helperPID || current.Phase != "running") {
		err = ErrUnknown
	}
	if err == nil {
		var alive bool
		alive, err = identityAlive(current.Identity)
		if err == nil && !alive {
			err = ErrUnknown
		}
	}
	if err != nil {
		write.Close()
		return nil, err
	}
	var inputReader *os.File
	if current.HoldInput {
		inputReader, err = duplicateInput(current)
		if err != nil {
			write.Close()
			return nil, err
		}
	}
	started = true
	return &Capture{root: root, runID: options.RunID, writer: write, inputReader: inputReader}, nil
}

// Open reconnects to the existing helper by persisted OS birth identity and
// secret IPC token. It never creates a replacement process or output pipe.
func Open(root, runID string) (*Capture, error) {
	dir, err := runRoot(root, runID)
	if err != nil {
		return nil, err
	}
	_, err = readableRecord(dir, runID)
	if err != nil {
		return nil, err
	}
	return &Capture{root: dir, runID: runID}, nil
}

func finished(phase string) bool {
	return phase == "complete" || phase == "failed" || phase == "invalid"
}

func (c *Capture) Stdout() *os.File { c.mu.Lock(); defer c.mu.Unlock(); return c.writer }

// Stdin is the birth-verified read-only copy prepared by Start(HoldInput:true).
// Open does not duplicate descriptors merely to read historical logs.
func (c *Capture) Stdin() *os.File { c.mu.Lock(); defer c.mu.Unlock(); return c.inputReader }

func (c *Capture) ReleaseInputReader() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.inputReader == nil {
		return nil
	}
	err := c.inputReader.Close()
	c.inputReader = nil
	return err
}

// ReleaseWriter closes only this daemon's duplicate. A daemon shutdown should
// call this, never Finish for a still-running instance.
func (c *Capture) ReleaseWriter() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.writer == nil {
		return nil
	}
	err := c.writer.Close()
	c.writer = nil
	return err
}

func (c *Capture) request(ctx context.Context, r record, path string, out any) error {
	return c.requestBody(ctx, r, http.MethodGet, path, nil, out)
}

func (c *Capture) requestBody(ctx context.Context, r record, method, path string, body io.Reader, out any) error {
	alive, err := identityAlive(r.Identity)
	if err != nil {
		return err
	}
	if !alive {
		return ErrIncomplete
	}
	host, port, err := net.SplitHostPort(r.Address)
	if err != nil || host != "127.0.0.1" {
		return ErrUnknown
	}
	if p, e := strconv.ParseUint(port, 10, 16); e != nil || p == 0 {
		return ErrUnknown
	}
	transport := &http.Transport{Proxy: nil, DialContext: (&net.Dialer{Timeout: time.Second}).DialContext, DisableKeepAlives: true, ResponseHeaderTimeout: 2 * time.Second}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	req, err := http.NewRequestWithContext(ctx, method, "http://"+r.Address+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+r.Token)
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusConflict {
		return terminal.ErrCursor
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("log helper HTTP status %d", res.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(res.Body, 2*terminal.MaxBatchBytes+8192)).Decode(out)
}

type inputResult struct {
	Written int    `json:"written"`
	Error   string `json:"error,omitempty"`
}

// WriteInput makes exactly one authenticated attempt. Any error or missing
// response is indeterminate and MUST NOT be automatically replayed.
func (c *Capture) WriteInput(ctx context.Context, data []byte) (int, error) {
	if len(data) > 32<<10 {
		return 0, errors.New("input exceeds 32 KiB budget")
	}
	r, err := loadRecord(c.root, c.runID)
	if err != nil {
		return 0, err
	}
	if !r.HoldInput || finished(r.Phase) {
		return 0, os.ErrClosed
	}
	var result inputResult
	err = c.requestBody(ctx, r, http.MethodPost, "/input", strings.NewReader(string(data)), &result)
	if err != nil {
		return result.Written, err
	}
	if result.Written < 0 || result.Written > len(data) {
		return 0, ErrUnknown
	}
	if result.Error != "" {
		return result.Written, errors.New(result.Error)
	}
	if result.Written != len(data) {
		return result.Written, io.ErrShortWrite
	}
	return result.Written, nil
}

func (c *Capture) Read(ctx context.Context, after uint64, maxBytes int) (terminal.Batch, error) {
	r, err := readableRecord(c.root, c.runID)
	if err != nil {
		return terminal.Batch{}, err
	}
	if !finished(r.Phase) {
		var batch terminal.Batch
		err = c.request(ctx, r, "/read?after="+strconv.FormatUint(after, 10)+"&maxBytes="+strconv.Itoa(maxBytes), &batch)
		if err == nil || errors.Is(err, terminal.ErrCursor) {
			return batch, err
		}
		// EOF completion can race the last IPC request. Only a durable finished
		// record permits opening this archive in a second process.
		r, _ = readableRecord(c.root, c.runID)
		if !finished(r.Phase) {
			return batch, err
		}
	}
	a, err := terminal.OpenArchive(filepath.Join(c.root, "archive"), r.Archive)
	if err != nil {
		return terminal.Batch{}, err
	}
	defer a.Close()
	return a.Read(after, maxBytes)
}

func (c *Capture) Status(ctx context.Context) (Status, error) {
	r, err := readableRecord(c.root, c.runID)
	if err != nil {
		return Status{}, err
	}
	if !finished(r.Phase) {
		var s Status
		err = c.request(ctx, r, "/status", &s)
		if err == nil {
			return s, nil
		}
		r, _ = readableRecord(c.root, c.runID)
		if !finished(r.Phase) {
			return s, err
		}
	}
	a, err := terminal.OpenArchive(filepath.Join(c.root, "archive"), r.Archive)
	if err != nil {
		return Status{}, err
	}
	defer a.Close()
	earliest, latest := a.Bounds()
	return Status{RunID: c.runID, Phase: r.Phase, Earliest: earliest, Latest: latest, Bytes: a.Bytes(), Diagnostic: r.Diagnostic}, nil
}

// Finish may ONLY be called after the runtime has positively confirmed that
// this entire run exited. It first waits for normal pipe EOF and archive sync.
// A remaining inherited writer or inaccessible helper never counts as success.
// The bounded fallback terminates only the birth-verified log helper, and
// returns ErrIncomplete; it never targets a business PID or process group.
func (c *Capture) Finish(ctx context.Context) error {
	_ = c.ReleaseWriter()
	_ = c.ReleaseInputReader()
	deadline, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	r, err := loadRecord(c.root, c.runID)
	if err != nil {
		return err
	}
	if r.HoldInput && !finished(r.Phase) {
		var accepted struct {
			Accepted bool `json:"accepted"`
		}
		// This operation only ends capture after the caller has confirmed that
		// the runtime exited. It does not execute a business input command.
		_ = c.requestBody(deadline, r, http.MethodPost, "/finish", nil, &accepted)
	}
	for {
		r, err := loadRecord(c.root, c.runID)
		if err != nil {
			return err
		}
		if finished(r.Phase) {
			if r.Phase == "failed" {
				return fmt.Errorf("%s: %w", r.Diagnostic, ErrIncomplete)
			}
			return nil
		}
		alive, err := identityAlive(r.Identity)
		if err != nil {
			return err
		}
		if !alive {
			return ErrIncomplete
		}
		select {
		case <-deadline.Done():
			if err = terminateIdentity(r.Identity); err != nil {
				return fmt.Errorf("log drain deadline; helper cleanup unconfirmed: %w", err)
			}
			return fmt.Errorf("log drain deadline exceeded: %w", ErrIncomplete)
		case <-time.After(25 * time.Millisecond):
		}
	}
}

func helperEnvironment(token string) []string {
	var env []string
	for _, value := range os.Environ() {
		key, _, _ := strings.Cut(value, "=")
		// The helper needs no daemon credentials. Only OS executable loading and
		// test race-runtime settings are inherited.
		switch strings.ToUpper(key) {
		case "PATH", "SYSTEMROOT", "WINDIR", "TMP", "TEMP", "GORACE":
			env = append(env, value)
		}
	}
	return append(env, "BLORA_RUNLOG_TOKEN="+token)
}

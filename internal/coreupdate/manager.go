package coreupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"blora.dev/panel/internal/storage"
)

type Manager struct {
	mu          sync.Mutex
	op          sync.Mutex
	options     Options
	source      Config
	job         *Job
	preview     *Preview
	selected    *selection
	busy        bool
	closed      bool
	cancel      context.CancelFunc
	checking    bool
	checkError  string
	checkCancel context.CancelFunc
	client      *http.Client
	probe       func(context.Context, string) (runtimeInfo, error)
}

func normalizeConfig(c Config) (Config, error) {
	if c.Repository == "" {
		c.Repository = DefaultRepository
	}
	if c.Channel == "" {
		c.Channel = DefaultChannel
	}
	if c.Channel != "stable" && c.Channel != "beta" {
		return Config{}, errors.New("update channel must be stable or beta")
	}
	c.Repository = strings.TrimSuffix(strings.TrimRight(c.Repository, "/"), ".git")
	u, err := secureURL(c.Repository)
	if err != nil {
		return Config{}, errors.New("release repository must be an HTTPS URL without credentials")
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || strings.ContainsAny(parts[0]+parts[1], " %?#\\") {
		return Config{}, errors.New("repository URL must identify an owner and repository")
	}
	if c.APIURL == "" {
		if u.Hostname() != "github.com" || u.Port() != "" {
			return Config{}, errors.New("a mirror repository requires its GitHub-compatible API URL")
		}
		c.APIURL = "https://api.github.com/repos/" + parts[0] + "/" + parts[1]
	}
	c.APIURL = strings.TrimRight(c.APIURL, "/")
	if _, err := secureURL(c.APIURL); err != nil {
		return Config{}, errors.New("release API URL must use HTTPS without credentials, query or fragment")
	}
	return c, nil
}

func secureURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || len(raw) > 2048 || strings.ContainsAny(raw, "\r\n\x00") {
		return nil, errors.New("invalid HTTPS URL")
	}
	return u, nil
}

func New(options Options) (*Manager, error) {
	if options.Component != "master" && options.Component != "daemon" {
		return nil, errors.New("invalid update component")
	}
	if options.Root == "" {
		return nil, errors.New("private update state directory is required")
	}
	root, err := filepath.Abs(options.Root)
	if err != nil {
		return nil, err
	}
	options.Root = root
	if options.CurrentRevision == "" {
		options.CurrentRevision = BuildRevision()
	}
	if options.CurrentRevision != "" && !revisionPattern.MatchString(options.CurrentRevision) {
		return nil, errors.New("invalid installed source revision")
	}
	if options.CurrentVersion == "" {
		options.CurrentVersion = Version
	}
	if err := privateDirectory(root); err != nil {
		return nil, err
	}
	c, err := normalizeConfig(options.Config)
	if err != nil {
		return nil, err
	}
	if err := readJSON(filepath.Join(root, "source.json"), &c); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read update source: %w", err)
	}
	c, err = normalizeConfig(c)
	if err != nil {
		return nil, err
	}
	m := &Manager{options: options, source: c, probe: probeExecutable, client: &http.Client{Timeout: 5 * time.Minute, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) > 8 {
			return errors.New("too many update redirects")
		}
		if req.URL.Scheme != "https" || req.URL.User != nil {
			return errors.New("update redirect must use HTTPS without credentials")
		}
		return nil
	}}}
	var check struct {
		Checking bool    `json:"checking"`
		Error    string  `json:"error,omitempty"`
		Preview  Preview `json:"preview,omitempty"`
	}
	if err := readJSON(filepath.Join(root, "check.json"), &check); err == nil {
		m.checkError = check.Error
		if check.Checking {
			m.checkError = "The previous update check was interrupted; check again."
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read update check: %w", err)
	}
	var job Job
	if err := readJSON(filepath.Join(root, "job.json"), &job); err == nil {
		m.job = &job
		if job.State == "running" && job.Phase != "activating" {
			m.job.State, m.job.Phase, m.job.Detail = "interrupted", "interrupted", "The previous preparation was interrupted; check and retry."
			m.job.UpdatedAt = time.Now().UTC()
			if err := m.persistJobLocked(); err != nil {
				return nil, err
			}
		}
		if job.State == "running" && job.Phase == "activating" {
			var failure struct {
				Revision string `json:"revision"`
				Error    string `json:"error"`
			}
			if readJSON(filepath.Join(root, "activation-error.json"), &failure) == nil && failure.Revision == job.Revision {
				m.job.State, m.job.Phase, m.job.Detail = "failed", "failed", "Replacement could not become ready; the previous installation remains active."
				m.job.UpdatedAt = time.Now().UTC()
				if err := m.persistJobLocked(); err != nil {
					return nil, err
				}
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read update job: %w", err)
	}
	return m, nil
}

func privateDirectory(path string) error {
	if info, err := os.Lstat(path); err == nil && (info.Mode()&os.ModeSymlink != 0 || !info.IsDir()) {
		return errors.New("update directory must be a real directory")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(path, 0700); err != nil {
		return err
	}
	return os.Chmod(path, 0700)
}

func readJSON(path string, value any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	d := json.NewDecoder(io.LimitReader(f, 1<<20))
	d.DisallowUnknownFields()
	if err := d.Decode(value); err != nil {
		return err
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return errors.New("update state must contain one JSON object")
	}
	return nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".update-state-")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(append(data, '\n'))
	}
	if err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return replaceFile(name, path)
}

func (m *Manager) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := Status{Component: m.options.Component, Version: m.options.CurrentVersion, Revision: m.options.CurrentRevision, Source: m.source, Platform: runtime.GOOS + "/" + runtime.GOARCH, Checking: m.checking, CheckError: m.checkError}
	if m.job != nil {
		j := *m.job
		if j.Prepared != nil {
			p := *j.Prepared
			j.Prepared = &p
		}
		s.Job = &j
	}
	if m.preview != nil {
		p := *m.preview
		s.Preview = &p
	}
	return s
}

func (m *Manager) SetSource(ctx context.Context, source Config) (Status, error) {
	if err := ctx.Err(); err != nil {
		return Status{}, err
	}
	c, err := normalizeConfig(source)
	if err != nil {
		return Status{}, err
	}
	if !m.op.TryLock() {
		return Status{}, ErrBusy
	}
	defer m.op.Unlock()
	m.mu.Lock()
	if m.busy || m.closed {
		m.mu.Unlock()
		return Status{}, ErrBusy
	}
	if err := writeJSON(filepath.Join(m.options.Root, "source.json"), c); err != nil {
		m.mu.Unlock()
		return Status{}, err
	}
	m.source, m.preview, m.selected, m.checkError = c, nil, nil, ""
	m.mu.Unlock()
	return m.Status(), nil
}

func (m *Manager) Check(ctx context.Context) (Preview, error) {
	if !m.op.TryLock() {
		return Preview{}, ErrBusy
	}
	defer m.op.Unlock()
	m.mu.Lock()
	if m.busy || m.closed {
		m.mu.Unlock()
		return Preview{}, ErrBusy
	}
	source := m.source
	m.mu.Unlock()
	return m.checkSource(ctx, source)
}

// BeginCheck makes network progress independent of browser/node request timeouts.
func (m *Manager) BeginCheck() (Status, error) {
	if !m.op.TryLock() {
		return Status{}, ErrBusy
	}
	m.mu.Lock()
	if m.busy || m.closed || m.checking {
		m.mu.Unlock()
		m.op.Unlock()
		return Status{}, ErrBusy
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	m.checking, m.checkError, m.checkCancel, m.preview, m.selected = true, "", cancel, nil, nil
	source := m.source
	if err := writeJSON(filepath.Join(m.options.Root, "check.json"), map[string]any{"checking": true}); err != nil {
		cancel()
		m.checking = false
		m.checkCancel = nil
		m.mu.Unlock()
		m.op.Unlock()
		return Status{}, err
	}
	m.mu.Unlock()
	go func() {
		defer m.op.Unlock()
		defer cancel()
		preview, err := m.checkSource(ctx, source)
		m.mu.Lock()
		defer m.mu.Unlock()
		m.checking, m.checkCancel = false, nil
		if err != nil {
			m.checkError = err.Error()
		}
		_ = writeJSON(filepath.Join(m.options.Root, "check.json"), map[string]any{"checking": false, "error": m.checkError, "preview": preview})
	}()
	return m.Status(), nil
}

func (m *Manager) checkSource(ctx context.Context, source Config) (Preview, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	selected, err := m.discover(ctx, source)
	if err != nil {
		return Preview{}, err
	}
	preview := selected.preview
	m.mu.Lock()
	m.preview = &preview
	m.selected = selected
	m.mu.Unlock()
	return preview, nil
}

func (m *Manager) Start(requestID, revision string) (Job, error) {
	if requestID == "" || len(requestID) > 256 || strings.ContainsAny(requestID, "\r\n\x00") {
		return Job{}, errors.New("a bounded update request ID is required")
	}
	if !revisionPattern.MatchString(revision) {
		return Job{}, errors.New("a full release source revision is required")
	}
	if !m.op.TryLock() {
		return Job{}, ErrBusy
	}
	defer m.op.Unlock()
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return Job{}, errors.New("updater is closed")
	}
	if job, found, err := m.lookupLocked(requestID, revision); found || err != nil {
		return job, err
	}
	if m.busy || (m.job != nil && m.job.State == "running" && m.job.Phase == "activating") {
		return Job{}, ErrBusy
	}
	if m.preview == nil || m.selected == nil || m.preview.Revision != revision {
		return Job{}, ErrPreviewRequired
	}
	if !m.preview.Compatible {
		return Job{}, ErrIncompatible
	}
	if m.preview.UpToDate {
		return Job{}, errors.New("this release is already installed")
	}
	if m.options.Ready == nil {
		return Job{}, errors.New("resource-preserving activation is unavailable in this installation")
	}
	// The prior supervisor failure belongs to the previous attempt. Leaving it
	// behind would incorrectly fail a healthy retry of the same source revision.
	if err := os.Remove(filepath.Join(m.options.Root, "activation-error.json")); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Job{}, errors.New("could not clear the previous activation failure")
	}
	now := time.Now().UTC()
	m.job = &Job{ID: requestID, Revision: revision, State: "running", Phase: "downloading", CreatedAt: now, UpdatedAt: now, TotalBytes: m.selected.asset.Size}
	if err := m.persistJobLocked(); err != nil {
		m.job = nil
		return Job{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	m.cancel, m.busy = cancel, true
	job := *m.job
	selected := *m.selected
	go m.prepare(ctx, selected)
	return job, nil
}

func (m *Manager) persistJobLocked() error {
	if err := privateDirectory(filepath.Join(m.options.Root, "jobs")); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(m.options.Root, "jobs", RequestKey("", m.job.ID)+".json"), m.job); err != nil {
		return err
	}
	return writeJSON(filepath.Join(m.options.Root, "job.json"), m.job)
}

func (m *Manager) Lookup(requestID, revision string) (Job, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lookupLocked(requestID, revision)
}
func (m *Manager) lookupLocked(requestID, revision string) (Job, bool, error) {
	var job Job
	if m.job != nil && m.job.ID == requestID {
		job = *m.job
	} else {
		err := readJSON(filepath.Join(m.options.Root, "jobs", RequestKey("", requestID)+".json"), &job)
		if errors.Is(err, os.ErrNotExist) {
			return Job{}, false, nil
		}
		if err != nil {
			return Job{}, false, err
		}
	}
	if job.Revision != revision {
		return Job{}, true, errors.New("update request ID was already used for a different revision")
	}
	return job, true, nil
}

func (m *Manager) transition(phase, detail string, prepared *Prepared) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.job.Phase, m.job.Detail, m.job.UpdatedAt = phase, detail, time.Now().UTC()
	if prepared != nil {
		p := *prepared
		m.job.Prepared = &p
	}
	return m.persistJobLocked()
}

func (m *Manager) downloadProgress(n int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.job.DownloadedBytes = n
	m.job.UpdatedAt = time.Now().UTC()
	return m.persistJobLocked()
}

func (m *Manager) prepare(ctx context.Context, selected selection) {
	err := m.prepareRelease(ctx, selected)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.busy = false
	if err != nil {
		m.job.State, m.job.Phase, m.job.Detail = "failed", "failed", err.Error()
		m.job.UpdatedAt = time.Now().UTC()
		_ = m.persistJobLocked()
	}
}

func (m *Manager) prepareRelease(ctx context.Context, selected selection) error {
	stage, err := os.MkdirTemp(m.options.Root, "stage-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	archive := filepath.Join(stage, "download")
	if err := m.downloadArchive(ctx, selected.asset, selected.checksum, archive); err != nil {
		return err
	}
	if err := m.transition("verifying", "Download verified; checking every installation file.", nil); err != nil {
		return err
	}
	directory := filepath.Join(stage, "installation")
	manifest, err := extractVerified(archive, selected.asset.Name, directory, m.options.Component, selected.preview.Version, runtime.GOOS+"-"+runtime.GOARCH)
	if err != nil {
		return err
	}
	_ = manifest
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	binary := filepath.Join(directory, "blora-"+m.options.Component+suffix)
	if err := os.Chmod(binary, 0700); err != nil {
		return errors.New("installation does not contain the expected executable")
	}
	if err := m.transition("preflight", "Checking replacement runtime compatibility.", nil); err != nil {
		return err
	}
	info, err := m.probe(ctx, binary)
	if err != nil {
		return errors.New("release does not support safe managed updates; use a maintenance installation")
	}
	if err := validateRuntime(info, m.options.Component, selected.preview.Manifest); err != nil {
		return err
	}
	if err := validateSourceNotice(filepath.Join(directory, "SOURCE-REVISION.txt"), selected.preview.Revision); err != nil {
		return err
	}
	if m.options.Component == "master" {
		if stat, err := os.Stat(filepath.Join(directory, "web", "dist", "index.html")); err != nil || !stat.Mode().IsRegular() {
			return errors.New("Master release is missing the frontend")
		}
	}
	if err := privateDirectory(filepath.Join(m.options.Root, "revisions")); err != nil {
		return err
	}
	// Each attempt has a distinct immutable slot. A failed activation can be
	// retried without deleting a slot that the supervisor may still be using.
	final := filepath.Join(m.options.Root, "revisions", selected.preview.Revision+"-"+filepath.Base(stage))
	if err := os.Rename(directory, final); err != nil {
		return err
	}
	prepared := Prepared{Revision: selected.preview.Revision, Binary: filepath.Join(final, filepath.Base(binary)), Directory: final, Manifest: selected.preview.Manifest}
	if m.options.Component == "master" {
		prepared.WebDirectory = filepath.Join(final, "web", "dist")
	}
	if err := writeJSON(filepath.Join(final, "prepared.json"), prepared); err != nil {
		return err
	}
	if err := m.transition("activating", "Verified release is ready; replacing the service without stopping instances.", &prepared); err != nil {
		return err
	}
	return m.options.Ready(ctx, prepared)
}

// ConfirmActivation is called after the replacement listener or node connection
// is ready. A downloaded installation alone is never reported as complete.
func (m *Manager) ConfirmActivation() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.job == nil || m.job.Revision != m.options.CurrentRevision || m.job.Phase != "activating" {
		return nil
	}
	m.job.State, m.job.Phase, m.job.Detail = "completed", "complete", "Replacement service is ready; instances were preserved."
	m.job.UpdatedAt = time.Now().UTC()
	return m.persistJobLocked()
}

func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	if m.cancel != nil {
		m.cancel()
	}
	if m.checkCancel != nil {
		m.checkCancel()
	}
}

func RequestKey(scope, requestID string) string {
	h := sha256.Sum256([]byte(scope + "\x00" + requestID))
	return hex.EncodeToString(h[:])
}

func compatibleManifest(manifest Manifest) error {
	if manifest.FormatVersion != 1 || manifest.ProtocolVersion != 1 || manifest.SchemaVersion != storage.SchemaVersion() || manifest.SchemaFingerprint != storage.SchemaFingerprint() || !manifest.SupportsPeer(1) || !manifest.PreserveInstances || !revisionPattern.MatchString(manifest.Revision) {
		return ErrIncompatible
	}
	return nil
}

package containers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type projectRecord struct {
	Project Project `json:"project"`
	Token   string  `json:"token"`
}

func (m *Manager) projectDir(id string) (string, error) {
	if !safeID.MatchString(id) {
		return "", ErrIdentity
	}
	return filepath.Join(m.options.StateRoot, "projects", id), nil
}
func (m *Manager) loadProject(id string, source bool) (Project, error) {
	var p projectRecord
	dir, err := m.projectDir(id)
	if err != nil {
		return p.Project, err
	}
	b, err := os.ReadFile(filepath.Join(dir, "project.json"))
	if err != nil {
		return p.Project, err
	}
	if err = json.Unmarshal(b, &p); err != nil {
		return p.Project, err
	}
	if p.Project.ID != id || len(p.Token) != 32 {
		return p.Project, ErrIdentity
	}
	p.Project.Token = p.Token
	if source {
		b, err = os.ReadFile(filepath.Join(dir, "revisions", fmt.Sprintf("%020d.yaml", p.Project.Revision)))
		if err != nil {
			return p.Project, err
		}
		p.Project.Source = string(b)
	}
	return p.Project, nil
}
func (m *Manager) saveProject(p Project) error {
	dir, err := m.projectDir(p.ID)
	if err != nil {
		return err
	}
	token := p.Token
	p.Source = ""
	return atomicJSON(filepath.Join(dir, "project.json"), projectRecord{Project: p, Token: token})
}
func (m *Manager) projects() ([]Project, error) {
	entries, err := os.ReadDir(filepath.Join(m.options.StateRoot, "projects"))
	if err != nil {
		return nil, err
	}
	var out []Project
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p, err := m.loadProject(e.Name(), false)
		if errors.Is(err, os.ErrNotExist) {
			continue // A revision commit may have stopped before the project index.
		}
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// SaveProject commits an immutable configuration revision, without invoking
// Compose, pulling an image, creating a container, or changing running resources.
func (m *Manager) SaveProject(ctx context.Context, actor, id string, expected uint64, source string) (Project, error) {
	if err := m.authorize(ctx, actor); err != nil {
		return Project{}, err
	}
	if len(source) == 0 || len(source) > 1<<20 {
		return Project{}, errors.New("Compose source must be 1 byte to 1 MiB")
	}
	if _, err := m.projectDir(id); err != nil {
		return Project{}, err
	}
	unlock, err := m.lock(ctx, "project-config:"+id)
	if err != nil {
		return Project{}, err
	}
	defer unlock()
	return m.saveProjectRevision(ctx, actor, id, expected, source, "")
}

// Caller holds project-config:id. Revision ownership is recorded before the
// public index so a lost commit acknowledgement remains reconcilable later.
func (m *Manager) saveProjectRevision(ctx context.Context, actor, id string, expected uint64, source, saveID string) (Project, error) {
	unlock, err := m.lock(ctx, "project-index")
	if err != nil {
		return Project{}, err
	}
	defer unlock()
	if err := m.authorize(ctx, actor); err != nil {
		return Project{}, err
	}
	dir, err := m.projectDir(id)
	if err != nil {
		return Project{}, err
	}
	p, err := m.loadProject(id, false)
	if errors.Is(err, os.ErrNotExist) {
		if expected != 0 {
			return p, ErrConflict
		}
		all, e := m.projects()
		if e != nil {
			return p, e
		}
		if len(all) >= m.options.MaxProjects {
			return p, errors.New("Compose project budget reached")
		}
		p = Project{ID: id, EngineName: "blora-project-" + fingerprint(m.token + ":" + id)[:20], Token: randomID()}
		if err = os.MkdirAll(dir, 0700); err != nil {
			return p, err
		}
		if err = os.MkdirAll(filepath.Join(dir, "revisions"), 0700); err != nil {
			return p, err
		}
	} else if err != nil {
		return p, err
	}
	if p.Revision != expected {
		return p, ErrConflict
	}
	if p.Revision >= 128 {
		return p, errors.New("Compose revision budget reached; archive old project history before saving more")
	}
	p.Revision++
	p.SourceSHA256 = sourceHash([]byte(source))
	p.UpdatedAt = time.Now().UTC()
	if err = atomicFile(filepath.Join(dir, "revisions", fmt.Sprintf("%020d.yaml", p.Revision)), []byte(source)); err != nil {
		return p, err
	}
	if err = atomicJSON(filepath.Join(dir, "revisions", fmt.Sprintf("%020d.commit.json", p.Revision)), revisionCommit{SaveID: saveID, SHA256: p.SourceSHA256, Project: p}); err != nil {
		return p, err
	}
	if err = m.saveProject(p); err != nil {
		return p, err
	}
	p.Source = source
	return p, nil
}

type boundedOutput struct {
	mu       sync.Mutex
	data     []byte
	max      int
	overflow bool
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n := len(p)
	remaining := b.max - len(b.data)
	if len(p) > remaining {
		b.overflow = true
		p = p[:max(0, remaining)]
	}
	b.data = append(b.data, p...)
	return n, nil
}
func (b *boundedOutput) bytes() ([]byte, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]byte{}, b.data...), b.overflow
}

func (m *Manager) composeEnvironment() ([]string, error) {
	values := map[string]string{"DOCKER_CONFIG": filepath.Join(m.options.StateRoot, "docker-client"), "COMPOSE_ANSI": "never", "COMPOSE_PROGRESS": "plain", "COMPOSE_PARALLEL_LIMIT": "4", "COMPOSE_MENU": "false"}
	for _, v := range os.Environ() {
		k, value, _ := strings.Cut(v, "=")
		switch strings.ToUpper(k) {
		case "PATH", "SYSTEMROOT", "WINDIR", "TMP", "TEMP":
			values[k] = value
		}
	}
	u, err := url.Parse(m.options.Endpoint)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "unix":
		values["DOCKER_HOST"] = m.options.Endpoint
	case "https":
		if m.options.ComposeTLSCertPath == "" {
			return nil, fmt.Errorf("Compose HTTPS requires administrator Docker TLS certificate directory: %w", ErrCapability)
		}
		values["DOCKER_HOST"] = "tcp://" + u.Host
		values["DOCKER_TLS_VERIFY"] = "1"
		values["DOCKER_CERT_PATH"] = m.options.ComposeTLSCertPath
	default:
		return nil, ErrCapability
	}
	var env []string
	for k, v := range values {
		env = append(env, k+"="+v)
	}
	sort.Strings(env)
	return env, nil
}
func (m *Manager) composeCommand() []string {
	if len(m.options.ComposeCommand) > 0 {
		return append([]string{}, m.options.ComposeCommand...)
	}
	return []string{"docker", "compose"}
}
func (m *Manager) invoke(ctx context.Context, p Project, file string, args []string, stdout, stderr io.Writer) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	env, err := m.composeEnvironment()
	if err != nil {
		return err
	}
	dir, err := m.projectDir(p.ID)
	if err != nil {
		return err
	}
	command := m.composeCommand()
	command = append(command, "--ansi", "never", "--progress", "plain", "--parallel", "4", "--project-directory", dir, "--project-name", p.EngineName, "--file", file)
	command = append(command, args...)
	return m.runCLI(ctx, command, dir, env, stdout, stderr)
}
func (m *Manager) composeVersion(ctx context.Context) (string, error) {
	out := &boundedOutput{max: 4096}
	errout := &boundedOutput{max: 4096}
	command := append(m.composeCommand(), "version", "--short")
	env, err := m.composeEnvironment()
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err = m.runCLI(ctx, command, m.options.StateRoot, env, out, errout); err != nil {
		return "", fmt.Errorf("Compose CLI unavailable: %w", err)
	}
	b, _ := out.bytes()
	v := strings.TrimPrefix(strings.TrimSpace(string(b)), "v")
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		return v, ErrCapability
	}
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	if major < 2 || (major == 2 && minor < 24) {
		return v, fmt.Errorf("Compose >=2.24 required: %w", ErrCapability)
	}
	return v, nil
}
func (m *Manager) capabilities(ctx context.Context) map[string]string {
	caps := map[string]string{"docker": "unavailable", "compose": "unavailable"}
	if m.engine == nil {
		caps["docker.reason"] = "no Engine endpoint configured"
		return caps
	}
	if err := m.engine.check(ctx); err != nil {
		caps["docker.reason"] = err.Error()
		return caps
	}
	caps["docker"] = "available"
	caps["docker.api"] = apiVersion
	caps["docker.terminal"] = "requires exec-helper v1 with processTree/pidfd and a private PID namespace"
	if v, err := m.composeVersion(ctx); err != nil {
		caps["compose.reason"] = err.Error()
	} else {
		caps["compose"] = "available"
		caps["compose.version"] = v
	}
	return caps
}

func (x *execution) compose(ctx context.Context) error {
	op := x.entry.Operation
	p, err := x.manager.loadProject(op.ProjectID, false)
	if err != nil {
		return err
	}
	if op.Revision == 0 || op.Revision > p.Revision {
		return ErrConflict
	}
	if op.Action == "compose.delete" {
		return x.deleteProject(ctx, p)
	}
	if op.Action != "compose.apply" {
		return errors.New("unsupported Compose operation")
	}
	seconds := op.HealthSeconds
	if seconds <= 0 {
		seconds = 60
	}
	if seconds > 600 {
		return errors.New("health timeout exceeds 600 seconds")
	}
	if _, err = x.manager.composeVersion(ctx); err != nil {
		return err
	}
	dir, _ := x.manager.projectDir(p.ID)
	source := filepath.Join(dir, "revisions", fmt.Sprintf("%020d.yaml", op.Revision))
	if err = x.emit("validate-config", "Validate the selected immutable Compose revision; saving did not deploy it"); err != nil {
		return err
	}
	out := &boundedOutput{max: 8 << 20}
	stderr := &boundedOutput{max: 64 << 10}
	if err = x.manager.invoke(ctx, p, source, []string{"config", "--format", "json"}, out, stderr); err != nil {
		b, _ := stderr.bytes()
		return fmt.Errorf("Compose configuration failed: %w: %s", err, string(b))
	}
	b, overflow := out.bytes()
	if overflow {
		return errors.New("resolved Compose configuration exceeds 8 MiB")
	}
	var doc map[string]any
	if err = json.Unmarshal(b, &doc); err != nil {
		return fmt.Errorf("Compose normalized model: %w", err)
	}
	if err = x.bindProject(ctx, p, doc); err != nil {
		return err
	}
	effective := filepath.Join(dir, "revisions", fmt.Sprintf("%020d.effective.json", op.Revision))
	if err = atomicJSON(effective, doc); err != nil {
		return err
	}
	policy := op.PullPolicy
	if policy == "" {
		policy = "missing"
	}
	if policy != "missing" && policy != "always" && policy != "never" {
		return errors.New("invalid pull policy")
	}
	if policy != "never" {
		if err = x.composeStage(ctx, p, effective, "pull", []string{"pull", "--policy", policy, "--ignore-buildable"}); err != nil {
			return err
		}
	}
	services, _ := doc["services"].(map[string]any)
	build := false
	for _, v := range services {
		service, _ := v.(map[string]any)
		if service["build"] != nil {
			build = true
		}
	}
	if build {
		if err = x.composeStage(ctx, p, effective, "build", []string{"build"}); err != nil {
			return err
		}
	}
	if err = x.composeStage(ctx, p, effective, "create", []string{"up", "--no-start", "--no-build", "--pull", "never"}); err != nil {
		return err
	}
	if err = x.composeStage(ctx, p, effective, "start", []string{"start"}); err != nil {
		return err
	}
	healthDeadline := time.Now().Add(time.Duration(seconds) * time.Second)
	if err = x.composeStage(ctx, p, effective, "health", []string{"start", "--wait", "--wait-timeout", strconv.Itoa(seconds)}); err != nil {
		return err
	}
	facts, err := x.waitProjectHealth(ctx, p, doc, healthDeadline)
	if err != nil {
		return err
	}
	if err = x.confirmed(facts...); err != nil {
		return err
	}
	unlock, err := x.manager.lock(ctx, "project-config:"+p.ID)
	if err != nil {
		return err
	}
	defer unlock()
	current, err := x.manager.loadProject(p.ID, false)
	if err != nil {
		return err
	}
	current.AppliedRevision = op.Revision
	current.LastTaskID = op.TaskID
	current.Deleted = false
	current.UpdatedAt = time.Now().UTC()
	return x.manager.saveProject(current)
}

// CLI exit status is not sufficient evidence that the selected project's
// requested services reached running/healthy. Some Compose versions return 0
// on a wait timeout; inspect actual Engine facts within our own deadline.
func (x *execution) waitProjectHealth(ctx context.Context, p Project, doc map[string]any, deadline time.Time) ([]Object, error) {
	services, _ := doc["services"].(map[string]any)
	expected := map[string]int{}
	completed := map[string]bool{}
	for name, raw := range services {
		service, _ := raw.(map[string]any)
		profiles, _ := service["profiles"].([]any)
		if len(profiles) > 0 {
			continue
		}
		replicas := 1
		if scale, ok := service["scale"].(float64); ok {
			replicas = int(scale)
		}
		if deploy, ok := service["deploy"].(map[string]any); ok {
			if count, ok := deploy["replicas"].(float64); ok {
				replicas = int(count)
			}
		}
		expected[name] = replicas
		dependencies, _ := service["depends_on"].(map[string]any)
		for dependency, raw := range dependencies {
			value, _ := raw.(map[string]any)
			if value["condition"] == "service_completed_successfully" {
				completed[dependency] = true
			}
		}
	}
	for {
		if err := x.manager.authorize(ctx, x.actor); err != nil {
			return nil, err
		}
		facts, err := x.projectFacts(ctx, p)
		if err != nil {
			x.entry.Result.Unknown = true
			return nil, err
		}
		x.entry.Result.Facts = facts
		seen := map[string]int{}
		pending := ""
		for _, o := range facts {
			service := o.Labels["com.docker.compose.service"]
			if _, wanted := expected[service]; !wanted {
				continue
			}
			seen[service]++
			if o.State == "exited" && completed[service] {
				actual, err := x.manager.engine.inspect(ctx, o.Target, true)
				if err != nil {
					return nil, err
				}
				var v containerInspect
				if err = json.Unmarshal(actual.Details, &v); err != nil {
					return nil, err
				}
				if v.State.ExitCode == 0 {
					continue
				}
			}
			if o.State != "running" || (o.Health != "" && o.Health != "healthy") {
				pending = fmt.Sprintf("service %s container %s is %s (health=%s)", service, o.Target.ID, o.State, o.Health)
			}
		}
		for service, count := range expected {
			if seen[service] < count {
				pending = fmt.Sprintf("service %s has %d of %d requested containers", service, seen[service], count)
			}
		}
		if pending == "" {
			return facts, nil
		}
		if !time.Now().Before(deadline) {
			return facts, fmt.Errorf("Compose health deadline elapsed; created resources retained: %s", pending)
		}
		timer := time.NewTimer(min(250*time.Millisecond, time.Until(deadline)))
		select {
		case <-ctx.Done():
			timer.Stop()
			return facts, ctx.Err()
		case <-timer.C:
		}
	}
}
func (x *execution) composeStage(ctx context.Context, p Project, file, stage string, args []string) error {
	if len(x.entry.Outputs) >= 16 {
		return errors.New("Compose output stage budget reached")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := x.emit(stage, "Compose "+stage+" for the fixed project revision"); err != nil {
		return err
	}
	x.entry.Result.Unknown = true
	outputIndex := len(x.entry.Outputs)
	x.entry.Outputs = append(x.entry.Outputs, CommandOutput{Phase: stage, RecordedAt: time.Now().UTC()})
	if err := x.save(); err != nil {
		return err
	}
	out := &boundedOutput{max: 64 << 10}
	stderr := &boundedOutput{max: 64 << 10}
	commandCtx, stopCommand := context.WithCancel(ctx)
	defer stopCommand()
	done := make(chan error, 1)
	go func() { done <- x.manager.invoke(commandCtx, p, file, args, out, stderr) }()
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	snapshot := func() CommandOutput {
		stdoutBytes, stdoutOverflow := out.bytes()
		stderrBytes, stderrOverflow := stderr.bytes()
		output := CommandOutput{Phase: stage, RecordedAt: time.Now().UTC(), Stdout: stdoutBytes, Stderr: stderrBytes, StdoutTruncated: stdoutOverflow, StderrTruncated: stderrOverflow}
		if len(output.Stdout) > 32<<10 {
			output.Stdout = output.Stdout[:32<<10]
			output.StdoutTruncated = true
		}
		if len(output.Stderr) > 32<<10 {
			output.Stderr = output.Stderr[:32<<10]
			output.StderrTruncated = true
		}
		return output
	}
	var err error
running:
	for {
		select {
		case err = <-done:
			break running
		case <-ticker.C:
			output := snapshot()
			previous := x.entry.Outputs[outputIndex]
			if bytes.Equal(previous.Stdout, output.Stdout) && bytes.Equal(previous.Stderr, output.Stderr) && previous.StdoutTruncated == output.StdoutTruncated && previous.StderrTruncated == output.StderrTruncated {
				continue
			}
			x.entry.Outputs[outputIndex] = output
			if saveErr := x.save(); saveErr != nil {
				stopCommand()
				<-done // Confirm the owned CLI exited before releasing this execution.
				return fmt.Errorf("Compose output checkpoint failed; remote result remains unknown: %w", saveErr)
			}
		}
	}
	output := snapshot()
	output.Completed = true
	if err != nil {
		output.CommandError = err.Error()
		if len(output.CommandError) > 512 {
			output.CommandError = output.CommandError[:512]
		}
	}
	x.entry.Outputs[outputIndex] = output
	if saveErr := x.save(); saveErr != nil {
		return fmt.Errorf("Compose command exited but output persistence failed: %w", saveErr)
	}
	factsCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	facts, factErr := x.projectFacts(factsCtx, p)
	cancel()
	if factErr == nil {
		x.entry.Result.Facts = facts
		for _, o := range facts {
			x.trackTarget(o.Target)
		}
		if len(facts) > 0 {
			x.entry.Result.Changed = true
		}
	}
	if err != nil {
		b, _ := stderr.bytes()
		if len(b) == 0 {
			b, _ = out.bytes()
		}
		if factErr == nil && ctx.Err() == nil && !errors.Is(err, ErrUnknown) {
			x.entry.Result.Unknown = false
		}
		return fmt.Errorf("Compose %s failed (already-created resources are retained): %w: %s", stage, err, string(b))
	}
	if factErr != nil {
		return factErr
	}
	x.entry.Result.Unknown = false
	x.entry.Result.Changed = true
	return x.save()
}
func (x *execution) projectFacts(ctx context.Context, p Project) ([]Object, error) {
	items, truncated, err := x.manager.engine.list(ctx, "containers", map[string]string{"com.docker.compose.project": p.EngineName}, 1000)
	if err != nil {
		return nil, err
	}
	if truncated {
		return nil, errors.New("project exceeds 1000-container inspection budget")
	}
	var facts []Object
	for _, o := range items {
		if o.Labels[labelPrefix+"project"] != p.Token {
			return nil, ErrIdentity
		}
		v, err := x.manager.engine.inspect(ctx, o.Target, false)
		if err != nil {
			return nil, err
		}
		facts = append(facts, v)
	}
	return facts, nil
}
func (x *execution) bindProject(ctx context.Context, p Project, doc map[string]any) error {
	if _, err := x.projectFacts(ctx, p); err != nil {
		return err
	}
	labels := x.manager.labels(x.entry.Operation.TaskID)
	labels[labelPrefix+"project"] = p.Token
	services, ok := doc["services"].(map[string]any)
	if !ok || len(services) == 0 || len(services) > 128 {
		return errors.New("Compose requires 1 to 128 services")
	}
	for _, raw := range services {
		service, ok := raw.(map[string]any)
		if !ok {
			return ErrIdentity
		}
		putComposeLabels(service, labels)
		if name, ok := service["container_name"].(string); ok {
			var existing containerInspect
			err := x.manager.engine.api(ctx, "GET", "/containers/"+url.PathEscape(name)+"/json", nil, &existing)
			if err == nil && existing.Config.Labels[labelPrefix+"project"] != p.Token {
				return ErrIdentity
			}
			if err != nil && !notFound(err) {
				return err
			}
		}
	}
	for _, kind := range []string{"volumes", "networks"} {
		objects, _ := doc[kind].(map[string]any)
		for _, raw := range objects {
			value, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			external, _ := value["external"].(bool)
			if external {
				continue
			}
			putComposeLabels(value, labels)
			name, _ := value["name"].(string)
			if name == "" {
				continue
			}
			if kind == "volumes" {
				o, err := x.manager.engine.inspect(ctx, Target{Kind: "volume", Name: name}, false)
				if err == nil && o.Labels[labelPrefix+"project"] != p.Token {
					return ErrIdentity
				}
				if err != nil && !notFound(err) {
					return err
				}
			} else {
				var existing networkInspect
				err := x.manager.engine.api(ctx, "GET", "/networks/"+url.PathEscape(name), nil, &existing)
				if err == nil && existing.Labels[labelPrefix+"project"] != p.Token {
					return ErrIdentity
				}
				if err != nil && !notFound(err) {
					return err
				}
			}
		}
	}
	return nil
}
func putComposeLabels(doc map[string]any, values map[string]string) {
	labels, _ := doc["labels"].(map[string]any)
	if labels == nil {
		labels = map[string]any{}
	}
	for k, v := range values {
		labels[k] = v
	}
	doc["labels"] = labels
}

func (x *execution) deleteProject(ctx context.Context, p Project) error {
	facts, err := x.projectFacts(ctx, p)
	if err != nil {
		return err
	}
	networks, truncated, err := x.manager.engine.list(ctx, "networks", map[string]string{"com.docker.compose.project": p.EngineName}, 1000)
	if err != nil {
		return err
	}
	if truncated {
		return ErrIdentity
	}
	for _, o := range networks {
		if o.Labels[labelPrefix+"project"] != p.Token {
			return ErrIdentity
		}
	}
	// Capture concrete immutable IDs BEFORE any deletion. Never call compose
	// down -v, global prune, or dynamically retarget a reused container name.
	for _, o := range facts {
		if err = ctx.Err(); err != nil {
			return err
		}
		if _, err = x.manager.engine.inspect(ctx, o.Target, false); err != nil {
			return err
		}
		seconds := x.entry.Operation.StopSeconds
		if seconds <= 0 {
			seconds = 30
		}
		if seconds > 300 {
			return errors.New("invalid stop timeout")
		}
		if o.State == "running" || o.State == "restarting" {
			if err = x.mutate(ctx, "stop", "POST", "/containers/"+o.Target.ID+"/stop?t="+strconv.Itoa(seconds), nil, nil); err != nil {
				return err
			}
		}
		if err = x.mutate(ctx, "delete-containers", "DELETE", "/containers/"+o.Target.ID+"?v=false&force=false", nil, nil); err != nil {
			return err
		}
		if _, err = x.manager.engine.inspect(ctx, o.Target, false); !notFound(err) {
			return ErrUnknown
		}
		x.entry.Result.Targets = append(x.entry.Result.Targets, o.Target)
		if err = x.confirmed(); err != nil {
			return err
		}
	}
	for _, o := range networks {
		if err = x.mutate(ctx, "delete-networks", "DELETE", "/networks/"+o.Target.ID, nil, nil); err != nil {
			return err
		}
		if _, err = x.manager.engine.inspect(ctx, o.Target, false); !notFound(err) {
			return ErrUnknown
		}
		x.entry.Result.Targets = append(x.entry.Result.Targets, o.Target)
		if err = x.confirmed(); err != nil {
			return err
		}
	}
	for _, t := range x.entry.Operation.DeleteVolumes {
		if err = x.deleteVolume(ctx, t); err != nil {
			return err
		}
	}
	remaining, err := x.projectFacts(ctx, p)
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		return ErrUnknown
	}
	unlock, err := x.manager.lock(ctx, "project-config:"+p.ID)
	if err != nil {
		return err
	}
	defer unlock()
	current, err := x.manager.loadProject(p.ID, false)
	if err != nil {
		return err
	}
	current.Deleted = true
	current.LastTaskID = x.entry.Operation.TaskID
	current.UpdatedAt = time.Now().UTC()
	return x.manager.saveProject(current)
}

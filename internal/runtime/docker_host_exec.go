package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"blora.dev/panel/internal/dockerapi"
)

// HostDockerTarget is a privileged caller's fixed container identity. It is
// deliberately separate from a platform-owned instance Record and its labels.
type HostDockerTarget struct {
	ContainerID string `json:"containerId"`
	CreatedAt   string `json:"createdAt"`
	StartedAt   string `json:"startedAt"`
}

func (t HostDockerTarget) Reference() string {
	sum := sha256.Sum256([]byte(t.ContainerID + "\x00" + t.CreatedAt + "\x00" + t.StartedAt))
	return hex.EncodeToString(sum[:])
}

type HostDockerOptions struct {
	StateRoot   string
	Endpoint    string
	TLSCertPath string
}
type HostDockerExecutor struct{ manager *Manager }

func NewHostDockerExecutor(options HostDockerOptions) (*HostDockerExecutor, error) {
	u, err := url.Parse(options.Endpoint)
	if err != nil {
		return nil, err
	}
	if u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme == "https" && u.Path != "") || (u.Scheme == "unix" && u.Host != "") {
		return nil, ErrCapability
	}
	m, err := New(Options{StateRoot: options.StateRoot, DockerEndpoint: options.Endpoint})
	if err != nil {
		return nil, err
	}
	if m.docker == nil {
		return nil, ErrCapability
	}
	transport := m.docker.client.Transport.(*http.Transport)
	transport.Proxy = nil
	if u.Scheme == "https" {
		transport.TLSClientConfig, err = dockerapi.TLSConfig(options.TLSCertPath)
		if err != nil {
			return nil, err
		}
	}
	return &HostDockerExecutor{manager: m}, nil
}
func (h *HostDockerExecutor) Close() error {
	if h != nil && h.manager != nil && h.manager.docker != nil {
		h.manager.docker.client.CloseIdleConnections()
	}
	return nil
}
func (h *HostDockerExecutor) Start(ctx context.Context, target HostDockerTarget, options DockerExecOptions, authorize func(context.Context) error) (*DockerExec, error) {
	if h == nil || h.manager == nil || authorize == nil {
		return nil, ErrCapability
	}
	if !hostContainerID.MatchString(target.ContainerID) || target.CreatedAt == "" || target.StartedAt == "" {
		return nil, ErrUnknown
	}
	if err := authorize(ctx); err != nil {
		return nil, err
	}
	scope := &dockerExecScope{manager: h.manager, host: &target, authorize: authorize, run: Record{RunID: target.Reference(), Backend: "docker-host", ContainerID: target.ContainerID}}
	return h.manager.startDockerExec(ctx, scope, options)
}

var hostContainerID = regexp.MustCompile(`^[a-f0-9]{64}$`)

type dockerExecScope struct {
	manager   *Manager
	run       Record
	host      *HostDockerTarget
	authorize func(context.Context) error
	cleanup   bool
}

func (s *dockerExecScope) allow(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.host == nil || s.cleanup {
		return nil
	}
	if s.authorize == nil {
		return ErrCapability
	}
	return s.authorize(ctx)
}
func (s *dockerExecScope) container(ctx context.Context, ready bool) (containerInspect, error) {
	if s.host == nil {
		if ready {
			return s.manager.execContainer(ctx, s.run)
		}
		return s.manager.inspectDocker(ctx, s.run)
	}
	var actual struct {
		containerInspect
		Created    string
		HostConfig struct{ PidMode string }
	}
	target := s.host
	if !hostContainerID.MatchString(target.ContainerID) || target.CreatedAt == "" || target.StartedAt == "" {
		return actual.containerInspect, ErrUnknown
	}
	if err := s.manager.docker.api(ctx, "GET", "/containers/"+target.ContainerID+"/json", nil, &actual); err != nil {
		return actual.containerInspect, err
	}
	if actual.ID != target.ContainerID || actual.Created != target.CreatedAt || actual.State.StartedAt != target.StartedAt {
		return actual.containerInspect, fmt.Errorf("host container identity or run birth changed: %w", ErrUnknown)
	}
	// The helper's PID namespace must remain the selected container's. Sharing
	// the host/another container PID namespace prevents a trustworthy helper
	// ownership boundary, even for a host administrator.
	if actual.HostConfig.PidMode != "" && actual.HostConfig.PidMode != "private" {
		return actual.containerInspect, fmt.Errorf("exec helper requires the container's own PID namespace: %w", ErrCapability)
	}
	if ready && (!actual.State.Running || actual.State.Paused || actual.State.Restarting || actual.State.Dead) {
		return actual.containerInspect, fmt.Errorf("container is not ready for exec: %w", ErrCapability)
	}
	if actual.Config.Labels["dev.blora.run"] != "" && actual.Config.Labels["dev.blora.instance"] != "" {
		return actual.containerInspect, errors.New("platform-managed container requires its instance terminal")
	}
	return actual.containerInspect, nil
}

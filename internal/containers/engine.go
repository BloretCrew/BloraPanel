package containers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"blora.dev/panel/internal/dockerapi"
)

const apiVersion = "1.45"
const labelPrefix = "blora.dev/panel."

var fullID = regexp.MustCompile(`^[a-f0-9]{64}$`)
var safeID = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,95}$`)
var volumeName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{1,127}$`)

type engine struct {
	client  *http.Client
	base    string
	mu      sync.Mutex
	checked bool
}
type engineError struct {
	Status  int
	Message string
}

func (e *engineError) Error() string {
	return fmt.Sprintf("Docker Engine status %d: %s", e.Status, e.Message)
}

func newEngine(endpoint, tlsDirectory string) (*engine, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.Proxy = nil
	// Docker's stop endpoint keeps the HTTP response open until the container
	// exits (up to the requested stop grace period, which is bounded at
	// 300 seconds by the operation validator). A ten-second transport limit
	// turns a valid default stop into an unknown operation. Keep a transport
	// ceiling above the protocol's maximum and rely on each query/operation
	// context for the tighter request deadline.
	tr.ResponseHeaderTimeout = 5 * time.Minute
	base := ""
	switch u.Scheme {
	case "unix":
		if u.Path == "" || u.Host != "" {
			return nil, ErrCapability
		}
		base = "http://docker"
		tr.DialContext = func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "unix", u.Path)
		}
	case "https":
		if u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" {
			return nil, ErrCapability
		}
		base = strings.TrimSuffix(endpoint, "/")
		tr.TLSClientConfig, err = dockerapi.TLSConfig(tlsDirectory)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Engine endpoint requires local unix or verified HTTPS: %w", ErrCapability)
	}
	return &engine{client: &http.Client{Transport: tr}, base: base}, nil
}
func version(v string) int {
	p := strings.Split(v, ".")
	if len(p) != 2 {
		return 0
	}
	a, _ := strconv.Atoi(p[0])
	b, _ := strconv.Atoi(p[1])
	return a*1000 + b
}
func (e *engine) check(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.checked {
		return nil
	}
	var v struct {
		APIVersion    string `json:"ApiVersion"`
		MinAPIVersion string `json:"MinAPIVersion"`
		OS            string `json:"Os"`
	}
	if err := e.request(ctx, "GET", "/version", nil, &v); err != nil {
		return err
	}
	if v.OS != "linux" && v.OS != "windows" {
		return fmt.Errorf("unsupported Engine OS: %w", ErrCapability)
	}
	if version(v.APIVersion) < version(apiVersion) || version(v.MinAPIVersion) > version(apiVersion) {
		return fmt.Errorf("Engine does not support API %s: %w", apiVersion, ErrCapability)
	}
	e.checked = true
	return nil
}
func (e *engine) api(ctx context.Context, method, path string, body, out any) error {
	if err := e.check(ctx); err != nil {
		return err
	}
	return e.request(ctx, method, "/v"+apiVersion+path, body, out)
}
func (e *engine) response(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var data io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		data = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, e.base+path, data)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		defer res.Body.Close()
		b, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return nil, &engineError{res.StatusCode, string(b)}
	}
	return res, nil
}
func (e *engine) request(ctx context.Context, method, path string, body, out any) error {
	res, err := e.response(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if out == nil {
		_, err = io.Copy(io.Discard, io.LimitReader(res.Body, 1<<20))
		return err
	}
	return json.NewDecoder(io.LimitReader(res.Body, 16<<20)).Decode(out)
}
func fingerprint(v any) string {
	b, _ := json.Marshal(v)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
func validateTarget(t Target, kind string) error {
	if t.Kind != kind {
		return ErrIdentity
	}
	switch kind {
	case "container", "network":
		if !fullID.MatchString(t.ID) {
			return ErrIdentity
		}
	case "image":
		if !strings.HasPrefix(t.ID, "sha256:") || !fullID.MatchString(strings.TrimPrefix(t.ID, "sha256:")) {
			return ErrIdentity
		}
	case "volume":
		if !volumeName.MatchString(t.Name) {
			return ErrIdentity
		}
	default:
		return ErrIdentity
	}
	return nil
}

type containerInspect struct {
	ID      string `json:"Id"`
	Name    string
	Created string
	Image   string
	Config  struct {
		Labels map[string]string
		Image  string
		Tty    bool
	}
	State struct {
		Status     string
		Running    bool
		Paused     bool
		Restarting bool
		Dead       bool
		ExitCode   int
		StartedAt  string
		Health     struct{ Status string }
	}
	Mounts []struct {
		Type        string
		Name        string
		Destination string
		Source      string
	}
}
type volumeInspect struct {
	Name       string
	CreatedAt  string
	Driver     string
	Scope      string
	Mountpoint string
	Labels     map[string]string
	Options    map[string]string
}
type networkInspect struct {
	ID         string `json:"Id"`
	Name       string
	Created    string
	Driver     string
	Labels     map[string]string
	Containers map[string]json.RawMessage
}

func (e *engine) inspect(ctx context.Context, t Target, details bool) (Object, error) {
	if err := validateTarget(t, t.Kind); err != nil {
		return Object{}, err
	}
	var raw json.RawMessage
	var path string
	switch t.Kind {
	case "container":
		path = "/containers/" + t.ID + "/json"
	case "image":
		path = "/images/" + t.ID + "/json"
	case "network":
		path = "/networks/" + t.ID
	case "volume":
		path = "/volumes/" + url.PathEscape(t.Name)
	}
	if err := e.api(ctx, "GET", path, nil, &raw); err != nil {
		return Object{}, err
	}
	o := Object{Target: t}
	switch t.Kind {
	case "container":
		var v containerInspect
		if err := json.Unmarshal(raw, &v); err != nil {
			return o, err
		}
		if v.ID != t.ID {
			return o, ErrIdentity
		}
		o.Name = strings.TrimPrefix(v.Name, "/")
		o.Target.CreatedAt = v.Created
		o.State = v.State.Status
		o.Health = v.State.Health.Status
		o.Image = v.Config.Image
		o.Labels = v.Config.Labels
	case "network":
		var v networkInspect
		if err := json.Unmarshal(raw, &v); err != nil {
			return o, err
		}
		if v.ID != t.ID {
			return o, ErrIdentity
		}
		o.Name = v.Name
		o.Target.CreatedAt = v.Created
		o.Labels = v.Labels
	case "volume":
		var v volumeInspect
		if err := json.Unmarshal(raw, &v); err != nil {
			return o, err
		}
		if v.Name != t.Name {
			return o, ErrIdentity
		}
		o.Name = v.Name
		o.Target.CreatedAt = v.CreatedAt
		o.Target.Fingerprint = fingerprint(v)
		o.Labels = v.Labels
	case "image":
		var v struct {
			ID       string `json:"Id"`
			Created  string
			RepoTags []string
			Config   struct{ Labels map[string]string }
		}
		if err := json.Unmarshal(raw, &v); err != nil {
			return o, err
		}
		if v.ID != t.ID {
			return o, ErrIdentity
		}
		o.Target.CreatedAt = v.Created
		o.Name = strings.Join(v.RepoTags, ", ")
		o.Labels = v.Config.Labels
	}
	if t.CreatedAt != "" && t.CreatedAt != o.Target.CreatedAt {
		return o, ErrIdentity
	}
	if t.Fingerprint != "" && t.Fingerprint != o.Target.Fingerprint {
		return o, ErrIdentity
	}
	if details {
		o.Details = raw
	}
	return o, nil
}
func labelFilter(labels map[string]string) string {
	var values []string
	for k, v := range labels {
		values = append(values, k+"="+v)
	}
	sort.Strings(values)
	b, _ := json.Marshal(map[string][]string{"label": values})
	return url.QueryEscape(string(b))
}

func (e *engine) list(ctx context.Context, kind string, labels map[string]string, limit int) ([]Object, bool, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	var path string
	switch kind {
	case "containers":
		path = "/containers/json?all=true"
	case "images":
		path = "/images/json?all=true"
	case "networks":
		path = "/networks?"
	case "volumes":
		path = "/volumes?"
	default:
		return nil, false, ErrIdentity
	}
	if len(labels) > 0 {
		path += "&filters=" + labelFilter(labels)
	}
	var raw json.RawMessage
	if err := e.api(ctx, "GET", path, nil, &raw); err != nil {
		return nil, false, err
	}
	var entries []json.RawMessage
	if kind == "volumes" {
		var x struct{ Volumes []json.RawMessage }
		if err := json.Unmarshal(raw, &x); err != nil {
			return nil, false, err
		}
		entries = x.Volumes
	} else if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, false, err
	}
	truncated := len(entries) > limit
	if truncated {
		entries = entries[:limit]
	}
	objects := make([]Object, 0, len(entries))
	for _, b := range entries {
		var v struct {
			ID        string `json:"Id"`
			Name      string
			Names     []string
			State     string
			Status    string
			Image     string
			Labels    map[string]string
			CreatedAt string
			Created   json.RawMessage
			RepoTags  []string
		}
		if err := json.Unmarshal(b, &v); err != nil {
			return nil, false, err
		}
		o := Object{Name: v.Name, State: v.State, Image: v.Image, Labels: v.Labels}
		switch kind {
		case "containers":
			o.Target = Target{Kind: "container", ID: v.ID}
			if len(v.Names) > 0 {
				o.Name = strings.TrimPrefix(v.Names[0], "/")
			}
		case "images":
			o.Target = Target{Kind: "image", ID: v.ID}
			o.Name = strings.Join(v.RepoTags, ", ")
		case "networks":
			o.Target = Target{Kind: "network", ID: v.ID}
		case "volumes":
			var vol volumeInspect
			if err := json.Unmarshal(b, &vol); err != nil {
				return nil, false, err
			}
			o.Target = Target{Kind: "volume", Name: vol.Name, CreatedAt: vol.CreatedAt, Fingerprint: fingerprint(vol)}
		}
		objects = append(objects, o)
	}
	return objects, truncated, nil
}

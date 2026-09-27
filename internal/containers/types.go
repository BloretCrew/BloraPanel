// Package containers implements privileged node-wide Docker and Compose
// management. It is separate from ordinary users' isolated instance runtime.
package containers

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrForbidden  = errors.New("host management permission required")
	ErrCapability = errors.New("container management capability unavailable")
	ErrIdentity   = errors.New("container resource identity changed or is incomplete")
	ErrConflict   = errors.New("resource or operation revision conflict")
	ErrUnknown    = errors.New("operation outcome is not fully confirmed; inspect before a new attempt")
)

type Options struct {
	Endpoint           string
	StateRoot          string
	ComposeCommand     []string                            // Administrator executable prefix; default docker compose.
	ComposeTLSCertPath string                              // Administrator-only shared Engine/Compose TLS directory (ca.pem, cert.pem, key.pem).
	Authorize          func(context.Context, string) error // Bind to host.manage for THIS node; nil denies all API calls.
	Labels             map[string]string                   // Administrator metadata added to objects created by this adapter.
	OperationTimeout   time.Duration
	MaxProjects        int
	MaxOperations      int
	MaxProjectSaves    int
}

type Target struct {
	Kind        string `json:"kind"`
	ID          string `json:"id,omitempty"`   // Full immutable container/network/image ID.
	Name        string `json:"name,omitempty"` // Volumes have no immutable API ID.
	CreatedAt   string `json:"createdAt,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

type Object struct {
	Target  Target            `json:"target"`
	Name    string            `json:"name"`
	State   string            `json:"state,omitempty"`
	Health  string            `json:"health,omitempty"`
	Image   string            `json:"image,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
	Details json.RawMessage   `json:"details,omitempty"`
}

type Query struct {
	Kind         string `json:"kind"` // capabilities, containers/container, images/image, volumes/volume, networks/network, projects/project, operation.
	Target       Target `json:"target,omitempty"`
	ProjectID    string `json:"projectId,omitempty"`
	TaskID       string `json:"taskId,omitempty"`
	Limit        int    `json:"limit,omitempty"`
	Details      bool   `json:"details,omitempty"`
	OutputOffset int    `json:"outputOffset,omitempty"`
}

type Snapshot struct {
	ObservedAt   time.Time         `json:"observedAt"`
	Items        []Object          `json:"items,omitempty"`
	Projects     []Project         `json:"projects,omitempty"`
	Project      *Project          `json:"project,omitempty"`
	Operation    *Result           `json:"operation,omitempty"`
	Capabilities map[string]string `json:"capabilities,omitempty"`
	Truncated    bool              `json:"truncated"`
	Output       *CommandOutput    `json:"output,omitempty"`
	NextOutput   int               `json:"nextOutput,omitempty"`
}

type CommandOutput struct {
	Completed       bool      `json:"completed"`
	Phase           string    `json:"phase"`
	RecordedAt      time.Time `json:"recordedAt"`
	Stdout          []byte    `json:"stdout"`
	Stderr          []byte    `json:"stderr"`
	StdoutTruncated bool      `json:"stdoutTruncated"`
	StderrTruncated bool      `json:"stderrTruncated"`
	CommandError    string    `json:"commandError,omitempty"`
}

type Operation struct {
	TaskID        string          `json:"taskId"`
	Action        string          `json:"action"` // container.create/start/stop/restart/delete, image.pull/delete, volume.create/delete, network.create/delete, compose.apply/delete.
	Target        Target          `json:"target,omitempty"`
	Name          string          `json:"name,omitempty"`
	Config        json.RawMessage `json:"config,omitempty"` // Engine create document; accepted only from host.manage.
	Image         string          `json:"image,omitempty"`
	ProjectID     string          `json:"projectId,omitempty"`
	Revision      uint64          `json:"revision,omitempty"`
	PullPolicy    string          `json:"pullPolicy,omitempty"` // missing(default), always, never.
	StopSeconds   int             `json:"stopSeconds,omitempty"`
	HealthSeconds int             `json:"healthSeconds,omitempty"`
	Force         bool            `json:"force,omitempty"`
	DeleteVolumes []Target        `json:"deleteVolumes,omitempty"` // Explicit separately presented volume identities. No implicit -v.
}

type Progress struct {
	Sequence uint64    `json:"sequence"`
	Phase    string    `json:"phase"`
	Message  string    `json:"message"`
	At       time.Time `json:"at"`
	Current  int64     `json:"current,omitempty"`
	Total    int64     `json:"total,omitempty"`
	Layer    string    `json:"layer,omitempty"`
}

type Result struct {
	TaskID     string     `json:"taskId"`
	Action     string     `json:"action"`
	State      string     `json:"state"` // RUNNING, SUCCEEDED, FAILED, CANCELLED, INTERRUPTED.
	Phase      string     `json:"phase"`
	Changed    bool       `json:"changed"`
	Unknown    bool       `json:"unknown"`
	Targets    []Target   `json:"targets,omitempty"`
	Facts      []Object   `json:"facts,omitempty"`
	Progress   []Progress `json:"progress"`
	Diagnostic string     `json:"diagnostic,omitempty"`
	StartedAt  time.Time  `json:"startedAt"`
	FinishedAt time.Time  `json:"finishedAt,omitempty"`
}

type Project struct {
	ID              string    `json:"id"`
	EngineName      string    `json:"engineName"`
	Revision        uint64    `json:"revision"`
	AppliedRevision uint64    `json:"appliedRevision"`
	Source          string    `json:"source,omitempty"`
	SourceSHA256    string    `json:"sourceSHA256,omitempty"`
	LastTaskID      string    `json:"lastTaskId,omitempty"`
	UpdatedAt       time.Time `json:"updatedAt"`
	Deleted         bool      `json:"deleted"`
	Token           string    `json:"-"`
}

type LogOptions struct {
	Since  time.Time
	Tail   int
	Follow bool
}

// LogFrame is Docker's original timestamped stdout/stderr payload, not a
// terminal event cursor. since/tail may overlap or have retention gaps.
type LogFrame struct {
	Stream string
	Data   []byte
}

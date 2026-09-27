// Package model defines identities and persisted management contracts.
package model

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"
)

func ID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}

type ResourceRef struct {
	Kind   string `json:"kind"`
	ID     string `json:"id"`
	NodeID string `json:"nodeId,omitempty"`
}

func (r ResourceRef) Key() string { return r.Kind + ":" + r.ID }

type User struct {
	ID       string `json:"userId"`
	Name     string `json:"name"`
	Admin    bool   `json:"admin"`
	Disabled bool   `json:"disabled"`
	Revision int64  `json:"revision"`
}
type Grant struct {
	UserID   string      `json:"userId"`
	Resource ResourceRef `json:"resource"`
	Action   string      `json:"action"`
}

// Audit is an append-only record of an authenticated management operation.
// RecordedAt is serialized as an RFC3339 timestamp so browser consumers do
// not need to know the storage epoch unit.
type Audit struct {
	Sequence    int64     `json:"sequence"`
	ActorID     string    `json:"actorId"`
	NodeID      string    `json:"nodeId,omitempty"`
	ResourceKey string    `json:"resourceKey"`
	Action      string    `json:"action"`
	RequestID   string    `json:"requestId,omitempty"`
	Result      string    `json:"result"`
	RecordedAt  time.Time `json:"recordedAt"`
}
type Node struct {
	ID             string            `json:"nodeId"`
	Name           string            `json:"name"`
	Group          string            `json:"group"`
	Tags           []string          `json:"tags"`
	Platform       string            `json:"platform"`
	State          string            `json:"state"`
	Generation     uint64            `json:"generation"`
	LastSeen       time.Time         `json:"lastSeen"`
	Capabilities   map[string]string `json:"capabilities"`
	Revision       int64             `json:"revision"`
	Maintenance    bool              `json:"maintenance"`
	Quota          int               `json:"quota"`
	ConfigRevision int64             `json:"configRevision"`
	StartupID      string            `json:"startupId,omitempty"`
}
type InstanceConfig struct {
	Command     []string          `json:"command"`
	Directory   string            `json:"directory"`
	Environment map[string]string `json:"environment,omitempty"`
	Mode        string            `json:"mode"`
	Image       string            `json:"image,omitempty"`
	UID         *uint32           `json:"uid,omitempty"`
	GID         *uint32           `json:"gid,omitempty"`
	MemoryBytes int64             `json:"memoryBytes,omitempty"`
	CPUQuota    int64             `json:"cpuQuota,omitempty"`
	PidsLimit   int64             `json:"pidsLimit,omitempty"`
	StopSeconds int               `json:"stopSeconds"`
	KillSeconds int               `json:"killSeconds"`
	Escalate    bool              `json:"escalate"`
	StopInput   string            `json:"stopInput,omitempty"`
	Autostart   bool              `json:"autostart"`
}
type Instance struct {
	ID             string         `json:"instanceId"`
	NodeID         string         `json:"nodeId"`
	NodeName       string         `json:"nodeName,omitempty"`
	NodeState      string         `json:"nodeState,omitempty"`
	Name           string         `json:"name"`
	Group          string         `json:"group"`
	Tags           []string       `json:"tags"`
	Config         InstanceConfig `json:"config"`
	State          string         `json:"state"`
	RunID          string         `json:"runId,omitempty"`
	Revision       int64          `json:"revision"`
	ConfigRevision int64          `json:"configRevision"`
	AutostartActor string         `json:"autostartActor,omitempty"`
	AutostartAfter string         `json:"autostartAfter,omitempty"`
	// Deleted is a durable tombstone.  The row remains so an old resourceRef
	// can never resolve to a newly-created instance with the same display name.
	Deleted bool `json:"deleted,omitempty"`
}
type TaskState string

const (
	Queued          TaskState = "QUEUED"
	Running         TaskState = "RUNNING"
	WaitingNode     TaskState = "WAITING_NODE"
	WaitingClient   TaskState = "WAITING_CLIENT"
	CancelRequested TaskState = "CANCEL_REQUESTED"
	Succeeded       TaskState = "SUCCEEDED"
	Failed          TaskState = "FAILED"
	Cancelled       TaskState = "CANCELLED"
	Interrupted     TaskState = "INTERRUPTED"
)

func (s TaskState) Terminal() bool {
	return s == Succeeded || s == Failed || s == Cancelled || s == Interrupted
}

type Task struct {
	ID                    string          `json:"taskId"`
	RetryOf               string          `json:"retryOf,omitempty"`
	RequestID             string          `json:"requestId"`
	ActorID               string          `json:"actorId"`
	Resource              ResourceRef     `json:"resource"`
	Action                string          `json:"action"`
	Payload               json.RawMessage `json:"payload"`
	Digest                string          `json:"digest"`
	State                 TaskState       `json:"state"`
	Phase                 string          `json:"phase"`
	Revision              int64           `json:"revision"`
	CreatedAt             time.Time       `json:"createdAt"`
	UpdatedAt             time.Time       `json:"updatedAt"`
	Deadline              time.Time       `json:"deadline"`
	DispatchedAt          time.Time       `json:"dispatchedAt,omitempty"`
	CancellationRequested bool            `json:"cancellationRequested"`
	CancellationRequestID string          `json:"cancellationRequestId,omitempty"`
	Result                json.RawMessage `json:"result,omitempty"`
	Error                 string          `json:"error,omitempty"`
}
type APIError struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	RequestID string       `json:"requestId,omitempty"`
	Resource  *ResourceRef `json:"resource,omitempty"`
}

func (e *APIError) Error() string { return e.Message }

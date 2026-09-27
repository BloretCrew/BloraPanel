package model

import "time"

type FileTaskPayload struct {
	Config           InstanceConfig    `json:"config"`
	Path             string            `json:"path"`
	Target           string            `json:"target,omitempty"`
	Version          string            `json:"version"`
	TargetVersion    string            `json:"targetVersion,omitempty"`
	UploadID         string            `json:"uploadId,omitempty"`
	Hash             string            `json:"hash,omitempty"`
	Total            int64             `json:"total,omitempty"`
	TrashID          string            `json:"trashId,omitempty"`
	Overwrite        map[string]string `json:"overwrite,omitempty"`
	Mode             uint32            `json:"mode,omitempty"`
	Modified         time.Time         `json:"modified,omitempty"`
	ExpectedObjectID string            `json:"expectedObjectId,omitempty"`
}

type TerminalTaskPayload struct {
	Host      *HostTerminalTarget `json:"host,omitempty"`
	Config    InstanceConfig      `json:"config"`
	SessionID string              `json:"sessionId,omitempty"`
	Cols      uint16              `json:"cols"`
	Rows      uint16              `json:"rows"`
}

type HostTerminalTarget struct {
	ContainerID string `json:"containerId"`
	CreatedAt   string `json:"createdAt"`
	StartedAt   string `json:"startedAt"`
}

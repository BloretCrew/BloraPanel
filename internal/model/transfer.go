package model

// TransferEndpoint is a browser-selected resource and version, never a host path.
type TransferEndpoint struct {
	InstanceID string `json:"instanceId"`
	Path       string `json:"path"`
	Version    string `json:"version"`
}

type TransferRequest struct {
	Source TransferEndpoint `json:"source"`
	Target TransferEndpoint `json:"target"`
	Move   bool             `json:"move"`
}

// TransferTaskPayload freezes the authoritative resource/root binding before
// any node side effect. Configurations are supplied by Master, not the client.
type TransferTaskPayload struct {
	TransferRequest
	SourceResource ResourceRef    `json:"sourceResource"`
	TargetResource ResourceRef    `json:"targetResource"`
	SourceConfig   InstanceConfig `json:"sourceConfig"`
	TargetConfig   InstanceConfig `json:"targetConfig"`
}

// TransferResult is deliberately bounded. Individual entries are paginated via
// the transfer endpoint, rather than embedded in task/control-channel events.
type TransferResult struct {
	Stage                string `json:"stage"`
	Entries              int    `json:"entries"`
	Completed            int    `json:"completed"`
	Total                int64  `json:"total"`
	CommittedBytes       int64  `json:"committedBytes"`
	CurrentPath          string `json:"currentPath,omitempty"`
	CurrentOffset        int64  `json:"currentOffset"`
	ChildTaskID          string `json:"childTaskId,omitempty"`
	DestinationVerified  bool   `json:"destinationVerified"`
	SourceDeleted        bool   `json:"sourceDeleted"`
	SourceOutcomeUnknown bool   `json:"sourceOutcomeUnknown"`
	Partial              bool   `json:"partial"`
	CleanupPending       bool   `json:"cleanupPending"`
	Error                string `json:"error,omitempty"`
}

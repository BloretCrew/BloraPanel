package model

type AgentHello struct {
	Platform     string            `json:"platform"`
	Capabilities map[string]string `json:"capabilities"`
	StartupID    string            `json:"startupId"`
}
type AgentCommand struct {
	Kind string `json:"kind"`
	Task Task   `json:"task"`
}
type AgentEvent struct {
	Task     Task      `json:"task"`
	Instance *Instance `json:"instance,omitempty"`
}

type AgentReceipt struct {
	TaskID    string `json:"taskId"`
	RequestID string `json:"requestId"`
	Missing   bool   `json:"missing"`
}

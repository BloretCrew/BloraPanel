package model

import "time"

type RunLog struct {
	RunID      string      `json:"runId"`
	Resource   ResourceRef `json:"resource"`
	Backend    string      `json:"backend"`
	StartedAt  time.Time   `json:"startedAt"`
	Diagnostic string      `json:"diagnostic,omitempty"`
}

type ConsoleInput struct {
	RunID string `json:"runId"`
	Data  string `json:"data"`
}

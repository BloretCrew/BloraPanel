package model

import "time"

type ContainerLogFrame struct {
	Stream string `json:"stream"`
	Data   []byte `json:"data"`
}

// Docker exposes retained time/tail windows, not stable byte/event offsets.
// Consumers replace this bounded window, preserving their own search/scroll.
type ContainerLogWindow struct {
	ContainerID string              `json:"containerId"`
	ObservedAt  time.Time           `json:"observedAt"`
	Frames      []ContainerLogFrame `json:"frames"`
	Truncated   bool                `json:"truncated"`
	PossibleGap bool                `json:"possibleGap"`
}

package model

import "encoding/json"

type ExtensionTaskPayload struct {
	ExtensionID string          `json:"extensionId"`
	PackageHash string          `json:"packageHash"`
	ModuleHash  string          `json:"moduleHash"`
	Payload     json.RawMessage `json:"payload"`
}

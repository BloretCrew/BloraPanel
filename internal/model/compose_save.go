package model

type ComposeSaveRequest struct {
	SaveID           string `json:"saveId"`
	ProjectID        string `json:"projectId"`
	ExpectedRevision uint64 `json:"expectedRevision"`
	Total            int64  `json:"total"`
	SHA256           string `json:"sha256"`
	Offset           int64  `json:"offset,omitempty"`
	Data             []byte `json:"data,omitempty"`
	Revision         uint64 `json:"revision,omitempty"`
	Length           int    `json:"length,omitempty"`
}

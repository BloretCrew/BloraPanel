package master

import (
	"net/http"
	"strings"
)

// requireRequestID makes the idempotency contract explicit at every HTTP
// boundary that accepts a durable task. Storage also validates this field, but
// checking it here gives callers a stable error before remote inspection or
// task admission work starts.
func requireRequestID(w http.ResponseWriter, r *http.Request) bool {
	key := r.Header.Get("Idempotency-Key")
	if strings.TrimSpace(key) == "" {
		fail(w, http.StatusBadRequest, "REQUEST_ID_REQUIRED", "需要 Idempotency-Key")
		return false
	}
	if len(key) > 128 {
		fail(w, http.StatusBadRequest, "REQUEST_ID_INVALID", "Idempotency-Key 最长 128 字节")
		return false
	}
	return true
}

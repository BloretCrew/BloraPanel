package extensions

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// SignatureMessage binds every supported manifest field and the payload hash.
// Normalization matches installation defaults; encoding/json sorts dependency
// map keys. Signers should use this function rather than serialize their own JSON.
// The domain prefix intentionally rejects the old payload-only signatures.
func SignatureMessage(p Package) []byte {
	h := sha256.Sum256(p.Payload)
	data, _ := json.Marshal(struct {
		Manifest Manifest `json:"manifest"`
		SHA256   string   `json:"sha256"`
	}{normalizeManifest(p.Manifest), hex.EncodeToString(h[:])})
	return append([]byte("BLORA-EXTENSION-SIGNATURE-2\n"), data...)
}

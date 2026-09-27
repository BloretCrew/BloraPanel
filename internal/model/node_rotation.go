package model

import (
	"crypto/sha256"
	"encoding/hex"
)

// RotationProof binds possession of the replacement private key to this exact
// single-use ticket and durable node identity.
func RotationProof(nodeID, ticket string, publicKey []byte) []byte {
	sum := sha256.Sum256([]byte(ticket))
	return []byte("blora.node-key-rotation.v1\n" + nodeID + "\n" + hex.EncodeToString(sum[:]) + "\n" + hex.EncodeToString(publicKey))
}

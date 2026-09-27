package protocol

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

var ErrChallenge = errors.New("invalid or expired node authentication challenge")

// Challenge is generated and durably consumed once by Master. Nonce binds the
// signature to this handshake; ConnectionID binds it to the pending WSS session.
// Verification alone does not prevent replay: Master must atomically consume
// the nonce, check revocation, and replace the old generation before accepting.
type Challenge struct {
	NodeID          string  `json:"nodeId"`
	Nonce           string  `json:"nonce"`
	ConnectionID    string  `json:"connectionId"`
	Generation      uint64  `json:"generation"`
	Channel         Channel `json:"channel"`
	ProtocolVersion uint32  `json:"protocolVersion"`
	ExpiresAt       int64   `json:"expiresAt"` // Unix seconds, UTC.
}

func NewChallenge(nodeID, connectionID string, generation uint64, channel Channel, ttl time.Duration) (Challenge, error) {
	var c Challenge
	if ttl <= 0 || ttl > 5*time.Minute {
		return c, ErrChallenge
	}
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return c, err
	}
	c = Challenge{NodeID: nodeID, Nonce: base64.RawURLEncoding.EncodeToString(nonce), ConnectionID: connectionID, Generation: generation, Channel: channel, ProtocolVersion: Version, ExpiresAt: time.Now().Add(ttl).Unix()}
	return c, c.validate(time.Now())
}

func (c Challenge) validate(now time.Time) error {
	nonce, err := base64.RawURLEncoding.DecodeString(c.Nonce)
	if err != nil || len(nonce) != 32 || c.NodeID == "" || len(c.NodeID) > MaxIdentifierSize || c.ConnectionID == "" || len(c.ConnectionID) > MaxIdentifierSize || c.Generation == 0 || !c.Channel.Valid() || c.ProtocolVersion != Version || c.ExpiresAt <= now.Unix() || c.ExpiresAt > now.Add(5*time.Minute).Unix() {
		return ErrChallenge
	}
	return nil
}

func challengeBytes(c Challenge) []byte {
	b, _ := json.Marshal(c) // Fixed struct order and JSON escaping remove concatenation ambiguity.
	return append([]byte("blora-node-auth-v1\x00"), b...)
}

func SignChallenge(private ed25519.PrivateKey, c Challenge) ([]byte, error) {
	if len(private) != ed25519.PrivateKeySize {
		return nil, ErrChallenge
	}
	if err := c.validate(time.Now()); err != nil {
		return nil, err
	}
	return ed25519.Sign(private, challengeBytes(c)), nil
}

func VerifyChallenge(public ed25519.PublicKey, c Challenge, signature []byte) error {
	if len(public) != ed25519.PublicKeySize || len(signature) != ed25519.SignatureSize {
		return ErrChallenge
	}
	if err := c.validate(time.Now()); err != nil {
		return err
	}
	if !ed25519.Verify(public, challengeBytes(c), signature) {
		return ErrChallenge
	}
	return nil
}

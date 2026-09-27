package daemon

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
)

type pendingRotation struct {
	TicketHash string `json:"ticketHash"`
	PrivateKey []byte `json:"privateKey"`
}

// Rotation is a startup operation. The pending replacement key is protected on
// disk before sending the request, so an ambiguous response never creates a new
// key on retry. Managed instance identities remain unchanged.
func (d *Daemon) rotateIdentity(ctx context.Context, file string) error {
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	if len(b) > 4096 {
		return errors.New("rotation ticket file exceeds size limit")
	}
	var ticket struct {
		NodeID string `json:"nodeId"`
		Ticket string `json:"ticket"`
	}
	if err := json.Unmarshal(b, &ticket); err != nil {
		return errors.New("rotation file must contain nodeId and ticket JSON")
	}
	if ticket.NodeID != d.identity.NodeID || len(ticket.Ticket) != 64 {
		return errors.New("rotation ticket belongs to another node or is invalid")
	}
	path := filepath.Join(d.config.StateDir, "rotation-pending.json")
	var pending pendingRotation
	b, err = os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(b, &pending); err != nil {
			return err
		}
		if len(pending.PrivateKey) != ed25519.PrivateKeySize {
			return errors.New("invalid pending replacement key")
		}
		if pending.TicketHash != storage.Hash([]byte(ticket.Ticket)) && !bytes.Equal(pending.PrivateKey, d.identity.PrivateKey) {
			return errors.New("a different pending rotation exists; reconcile its outcome before replacing it")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if errors.Is(err, os.ErrNotExist) || pending.TicketHash != storage.Hash([]byte(ticket.Ticket)) {
		_, key, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return err
		}
		pending = pendingRotation{TicketHash: storage.Hash([]byte(ticket.Ticket)), PrivateKey: key}
		b, err = json.Marshal(pending)
		if err != nil {
			return err
		}
		if err := privateWrite(path, b); err != nil {
			return err
		}
	}
	key := ed25519.PrivateKey(pending.PrivateKey)
	public := key.Public().(ed25519.PublicKey)
	body, err := json.Marshal(map[string]any{"nodeId": ticket.NodeID, "ticket": ticket.Ticket, "publicKey": public, "signature": ed25519.Sign(key, model.RotationProof(ticket.NodeID, ticket.Ticket, public))})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(d.config.MasterURL, "/")+"/api/v1/agent/rotate", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("node key rotation rejected; pending replacement key retained for reconciliation")
	}
	var result struct {
		NodeID  string `json:"nodeId"`
		Rotated bool   `json:"rotated"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&result); err != nil {
		return err
	}
	if result.NodeID != d.identity.NodeID || !result.Rotated {
		return errors.New("invalid key rotation confirmation")
	}
	next := identity{NodeID: d.identity.NodeID, PrivateKey: pending.PrivateKey}
	b, err = json.Marshal(next)
	if err != nil {
		return err
	}
	if err := privateWrite(filepath.Join(d.config.StateDir, "identity.json"), b); err != nil {
		return err
	}
	d.identity = next
	// Keeping the committed pending record allows the same configured ticket to
	// be replayed after another restart without generating a replacement key.
	return nil
}

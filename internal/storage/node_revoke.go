package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"blora.dev/panel/internal/model"
)

// RevokeNode durably fences a node identity and records an idempotent receipt.
// Existing instances remain in storage so their resource references and
// recovery records are not silently deleted; the node simply cannot reconnect
// until an administrator issues an explicit rotation ticket.
func (s *Store) RevokeNode(ctx context.Context, actor, id, requestID string) (model.Node, error) {
	if actor == "" || id == "" || requestID == "" || len(requestID) > 128 {
		return model.Node{}, errors.New("actor, node and request ID required")
	}
	input, err := json.Marshal(struct {
		ID string `json:"nodeId"`
	}{id})
	if err != nil {
		return model.Node{}, err
	}
	digest := Hash(append([]byte("node.revoke\n"), input...))
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return model.Node{}, err
	}
	defer tx.Rollback()
	if err := requireAdmin(ctx, tx, actor); err != nil {
		return model.Node{}, err
	}
	receiptKey := actor + ":" + requestID
	var saved []byte
	err = tx.QueryRowContext(ctx, "SELECT document FROM records WHERE namespace='metadata_request' AND id=?", receiptKey).Scan(&saved)
	if err == nil {
		var receipt struct {
			Digest string
			Result json.RawMessage
		}
		if err := json.Unmarshal(saved, &receipt); err != nil {
			return model.Node{}, err
		}
		if receipt.Digest != digest {
			return model.Node{}, ErrRequestMismatch
		}
		var node model.Node
		if err := json.Unmarshal(receipt.Result, &node); err != nil {
			return model.Node{}, err
		}
		return node, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return model.Node{}, err
	}
	var document []byte
	var revoked bool
	var generation uint64
	if err := tx.QueryRowContext(ctx, "SELECT document,revoked,generation FROM nodes WHERE id=?", id).Scan(&document, &revoked, &generation); err != nil {
		return model.Node{}, err
	}
	var node model.Node
	if err := json.Unmarshal(document, &node); err != nil {
		return model.Node{}, err
	}
	if !revoked {
		generation++
		if _, err := tx.ExecContext(ctx, "UPDATE nodes SET revoked=1,generation=? WHERE id=? AND revoked=0", generation, id); err != nil {
			return model.Node{}, err
		}
	}
	node.Generation = generation
	node.State = "REVOKED"
	result, err := json.Marshal(node)
	if err != nil {
		return model.Node{}, err
	}
	receipt, err := json.Marshal(struct {
		Digest string
		Result json.RawMessage
	}{digest, result})
	if err != nil {
		return model.Node{}, err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO records(namespace,id,revision,document) VALUES('metadata_request',?,1,?)", receiptKey, receipt); err != nil {
		return model.Node{}, err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO audit(actor_id,node_id,resource_key,action,request_id,result,recorded_at) VALUES(?,?,?, ?,?,?,?)", actor, id, "node:"+id, "node.revoke", requestID, "applied", time.Now().UnixMilli()); err != nil {
		return model.Node{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.Node{}, err
	}
	return node, nil
}

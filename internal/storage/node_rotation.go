package storage

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"blora.dev/panel/internal/model"
)

type rotationTicket struct {
	Hash       string
	Expires    int64
	Used       bool
	PublicKey  []byte
	Reactivate bool
}

func (s *Store) NewRotationTicket(ctx context.Context, actor, id string, reactivate bool) (string, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	if err := requireAdmin(ctx, tx, actor); err != nil {
		return "", err
	}
	var revoked bool
	if err := tx.QueryRowContext(ctx, "SELECT revoked FROM nodes WHERE id=?", id).Scan(&revoked); err != nil {
		return "", err
	}
	if revoked && !reactivate {
		return "", errors.New("reactivation of a revoked identity must be explicitly requested")
	}
	ticket := model.ID() + model.ID()
	b, err := json.Marshal(rotationTicket{Hash: Hash([]byte(ticket)), Expires: time.Now().Add(10 * time.Minute).Unix(), Reactivate: reactivate})
	if err != nil {
		return "", err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO records(namespace,id,revision,document) VALUES('node_rotation',?,1,?) ON CONFLICT(namespace,id) DO UPDATE SET revision=revision+1,document=excluded.document", id, b); err != nil {
		return "", err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO audit(actor_id,node_id,resource_key,action,request_id,result,recorded_at) VALUES(?,?,?,?,?,?,?)", actor, id, "node:"+id, "node.rotation_ticket", "", "issued", time.Now().UnixMilli()); err != nil {
		return "", err
	}
	return ticket, tx.Commit()
}

// NewRotationTicketMutation issues one rotation ticket per administrator
// request. A lost-response retry with the same request ID returns the
// original ticket; a different payload under that ID is rejected by the
// durable metadata receipt.
func (s *Store) NewRotationTicketMutation(ctx context.Context, actor, requestID, id string, reactivate bool) (string, error) {
	ticket := model.ID() + model.ID()
	data, err := s.MetadataMutation(ctx, actor, requestID, "node.rotation_ticket", struct {
		ID         string `json:"ID"`
		Reactivate bool   `json:"reactivate"`
	}{ID: id, Reactivate: reactivate}, func(tx *sql.Tx) (any, error) {
		var revoked bool
		if err := tx.QueryRowContext(ctx, "SELECT revoked FROM nodes WHERE id=?", id).Scan(&revoked); err != nil {
			return nil, err
		}
		if revoked && !reactivate {
			return nil, errors.New("reactivation of a revoked identity must be explicitly requested")
		}
		b, err := json.Marshal(rotationTicket{Hash: Hash([]byte(ticket)), Expires: time.Now().Add(10 * time.Minute).Unix(), Reactivate: reactivate})
		if err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO records(namespace,id,revision,document) VALUES('node_rotation',?,1,?) ON CONFLICT(namespace,id) DO UPDATE SET revision=revision+1,document=excluded.document", id, b); err != nil {
			return nil, err
		}
		return struct {
			Ticket string `json:"ticket"`
		}{Ticket: ticket}, nil
	})
	if err != nil {
		return "", err
	}
	var result struct {
		Ticket string `json:"ticket"`
	}
	if err := json.Unmarshal(data, &result); err != nil || result.Ticket == "" {
		if err == nil {
			err = errors.New("rotation ticket receipt missing ticket")
		}
		return "", err
	}
	return result.Ticket, nil
}

func (s *Store) RotateNodeKey(ctx context.Context, id, ticket string, key []byte) (uint64, error) {
	if len(key) != 32 || len(ticket) != 64 {
		return 0, ErrForbidden
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var b, oldKey []byte
	if err := tx.QueryRowContext(ctx, "SELECT document FROM records WHERE namespace='node_rotation' AND id=?", id).Scan(&b); err != nil {
		return 0, ErrForbidden
	}
	var r rotationTicket
	if err := json.Unmarshal(b, &r); err != nil {
		return 0, err
	}
	if r.Hash != Hash([]byte(ticket)) {
		return 0, ErrForbidden
	}
	var generation uint64
	var revoked bool
	if err := tx.QueryRowContext(ctx, "SELECT public_key,generation,revoked FROM nodes WHERE id=?", id).Scan(&oldKey, &generation, &revoked); err != nil {
		return 0, err
	}
	if r.Used {
		if bytes.Equal(r.PublicKey, key) && bytes.Equal(oldKey, key) && !revoked {
			return generation, nil
		}
		return 0, ErrRequestMismatch
	}
	if r.Expires < time.Now().Unix() || revoked && !r.Reactivate {
		return 0, ErrForbidden
	}
	if bytes.Equal(oldKey, key) {
		return 0, errors.New("replacement key must differ from current key")
	}
	r.Used = true
	r.PublicKey = append([]byte(nil), key...)
	b, err = json.Marshal(r)
	if err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE nodes SET public_key=?,generation=generation+1,revoked=0 WHERE id=?", key, id); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE records SET revision=revision+1,document=? WHERE namespace='node_rotation' AND id=?", b, id); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO audit(actor_id,node_id,resource_key,action,request_id,result,recorded_at) VALUES(?,?,?,?,?,?,?)", "node:"+id, id, "node:"+id, "node.key_rotate", "", "old_generation_fenced", time.Now().UnixMilli()); err != nil {
		return 0, err
	}
	return generation + 1, tx.Commit()
}

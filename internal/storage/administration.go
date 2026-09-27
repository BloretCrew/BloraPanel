package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"blora.dev/panel/internal/model"
	"golang.org/x/crypto/bcrypt"
)

var ErrForbidden = errors.New("current administrator permission required")
var ErrLastAdmin = errors.New("at least one enabled platform administrator must remain")

func requireAdmin(ctx context.Context, tx *sql.Tx, actor string) error {
	var admin, disabled bool
	if err := tx.QueryRowContext(ctx, "SELECT admin,disabled FROM users WHERE id=?", actor).Scan(&admin, &disabled); err != nil {
		return err
	}
	if !admin || disabled {
		return ErrForbidden
	}
	return nil
}

// MetadataMutation atomically records a non-secret metadata change and its
// receipt. Callers must never supply a password/token in input or result.
func (s *Store) MetadataMutation(ctx context.Context, actor, requestID, action string, input any, change func(*sql.Tx) (any, error)) (json.RawMessage, error) {
	return s.metadataMutation(ctx, actor, requestID, action, input, func(tx *sql.Tx) error { return requireAdmin(ctx, tx, actor) }, change)
}

// UserMetadataMutation records an authenticated user's durable metadata
// change and its receipt. Unlike MetadataMutation it permits a non-admin
// actor; callers still perform any resource-specific authorization in change.
func (s *Store) UserMetadataMutation(ctx context.Context, actor, requestID, action string, input any, change func(*sql.Tx) (any, error)) (json.RawMessage, error) {
	return s.metadataMutation(ctx, actor, requestID, action, input, func(tx *sql.Tx) error {
		var disabled bool
		if err := tx.QueryRowContext(ctx, "SELECT disabled FROM users WHERE id=?", actor).Scan(&disabled); err != nil {
			return err
		}
		if disabled {
			return ErrForbidden
		}
		return nil
	}, change)
}

func (s *Store) metadataMutation(ctx context.Context, actor, requestID, action string, input any, authorize func(*sql.Tx) error, change func(*sql.Tx) (any, error)) (json.RawMessage, error) {
	if requestID == "" || len(requestID) > 128 {
		return nil, errors.New("request ID required")
	}
	data, err := json.Marshal(struct {
		Action string
		Input  any
	}{action, input})
	if err != nil {
		return nil, err
	}
	digest := Hash(data)
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := authorize(tx); err != nil {
		return nil, err
	}
	var saved []byte
	err = tx.QueryRowContext(ctx, "SELECT document FROM records WHERE namespace='metadata_request' AND id=?", actor+":"+requestID).Scan(&saved)
	if err == nil {
		var receipt struct {
			Digest string
			Result json.RawMessage
		}
		if err = json.Unmarshal(saved, &receipt); err != nil {
			return nil, err
		}
		if receipt.Digest != digest {
			return nil, ErrRequestMismatch
		}
		return receipt.Result, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	value, err := change(tx)
	if err != nil {
		return nil, err
	}
	result, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	receipt, err := json.Marshal(struct {
		Digest string
		Result json.RawMessage
	}{digest, result})
	if err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO records(namespace,id,revision,document) VALUES('metadata_request',?,1,?)", actor+":"+requestID, receipt); err != nil {
		return nil, err
	}
	resourceKey, nodeID := "platform:administration", ""
	var identity struct {
		ID string `json:"ID"`
	}
	if raw, marshalErr := json.Marshal(input); marshalErr == nil && json.Unmarshal(raw, &identity) == nil {
		switch action {
		case "user.create":
			if identity.ID != "" {
				resourceKey = "user:" + identity.ID
			}
		case "user.password":
			var passwordInput struct {
				Target string `json:"target"`
			}
			if json.Unmarshal(raw, &passwordInput) == nil && passwordInput.Target != "" {
				resourceKey = "user:" + passwordInput.Target
			}
		case "workspace.put", "workspace.delete":
			var workspaceInput struct {
				WorkspaceID string `json:"workspaceId"`
			}
			if json.Unmarshal(raw, &workspaceInput) == nil && workspaceInput.WorkspaceID != "" {
				resourceKey = "workspace:" + workspaceInput.WorkspaceID
			}
		case "node.settings":
			nodeID, resourceKey = identity.ID, "node:"+identity.ID
		case "node.rotation_ticket":
			nodeID, resourceKey = identity.ID, "node:"+identity.ID
		case "instance.configure":
			resourceKey = "instance:" + identity.ID
			_ = tx.QueryRowContext(ctx, "SELECT node_id FROM instances WHERE id=?", identity.ID).Scan(&nodeID)
		case "instance.delete":
			resourceKey = "instance:" + identity.ID
			_ = tx.QueryRowContext(ctx, "SELECT node_id FROM instances WHERE id=?", identity.ID).Scan(&nodeID)
		case "schedule.save":
			var scheduleInput struct {
				ID   string `json:"id"`
				Spec struct {
					Resource struct {
						NodeID string `json:"nodeId"`
					} `json:"resource"`
				} `json:"spec"`
			}
			if json.Unmarshal(raw, &scheduleInput) == nil && scheduleInput.ID != "" {
				resourceKey = "schedule:" + scheduleInput.ID
				nodeID = scheduleInput.Spec.Resource.NodeID
			}
		case "schedule.delete":
			if identity.ID != "" {
				resourceKey = "schedule:" + identity.ID
			}
		case "grant.update":
			var grant struct {
				Grant model.Grant `json:"grant"`
			}
			if json.Unmarshal(raw, &grant) == nil && grant.Grant.Resource.ID != "" {
				resourceKey = grant.Grant.Resource.Key()
				switch grant.Grant.Resource.Kind {
				case "node":
					nodeID = grant.Grant.Resource.ID
				case "instance":
					_ = tx.QueryRowContext(ctx, "SELECT node_id FROM instances WHERE id=?", grant.Grant.Resource.ID).Scan(&nodeID)
				}
			}
		}
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO audit(actor_id,node_id,resource_key,action,request_id,result,recorded_at) VALUES(?,?,?,?,?,?,?)", actor, nodeID, resourceKey, action, requestID, "applied", time.Now().UnixMilli()); err != nil {
		return nil, err
	}
	return result, tx.Commit()
}

func (s *Store) UpdateUser(ctx context.Context, actor, target, requestID string, p model.UserPatch) (model.User, error) {
	data, err := s.MetadataMutation(ctx, actor, requestID, "user.update", struct {
		Target string
		Patch  model.UserPatch
	}{target, p}, func(tx *sql.Tx) (any, error) {
		var u model.User
		if err := tx.QueryRowContext(ctx, "SELECT id,name,admin,disabled,revision FROM users WHERE id=?", target).Scan(&u.ID, &u.Name, &u.Admin, &u.Disabled, &u.Revision); err != nil {
			return nil, err
		}
		if u.Revision != p.Revision {
			return nil, ErrConflict
		}
		if p.Name != nil {
			u.Name = *p.Name
		}
		if p.Admin != nil {
			u.Admin = *p.Admin
		}
		if p.Disabled != nil {
			u.Disabled = *p.Disabled
		}
		if !u.Admin || u.Disabled {
			var others int
			if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE admin=1 AND disabled=0 AND id<>?", target).Scan(&others); err != nil {
				return nil, err
			}
			if others == 0 {
				return nil, ErrLastAdmin
			}
		}
		u.Revision++
		if _, err := tx.ExecContext(ctx, "UPDATE users SET name=?,admin=?,disabled=?,revision=? WHERE id=?", u.Name, u.Admin, u.Disabled, u.Revision, u.ID); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM sessions WHERE user_id=?", u.ID); err != nil {
			return nil, err
		}
		return u, nil
	})
	var u model.User
	if err == nil {
		err = json.Unmarshal(data, &u)
	}
	return u, err
}

func (s *Store) ChangePassword(ctx context.Context, actor, target, requestID string, revision int64, hash []byte, passwordDigest, currentPassword string) error {
	_, err := s.metadataMutation(ctx, actor, requestID, "user.password", struct {
		Target         string `json:"target"`
		Revision       int64  `json:"revision"`
		PasswordDigest string `json:"passwordDigest"`
	}{target, revision, passwordDigest}, func(tx *sql.Tx) error {
		var disabled bool
		if err := tx.QueryRowContext(ctx, "SELECT disabled FROM users WHERE id=?", actor).Scan(&disabled); err != nil {
			return err
		}
		if disabled {
			return ErrForbidden
		}
		if actor != target {
			return requireAdmin(ctx, tx, actor)
		}
		return nil
	}, func(tx *sql.Tx) (any, error) {
		var currentHash []byte
		if actor == target {
			if err := tx.QueryRowContext(ctx, "SELECT password_hash FROM users WHERE id=?", target).Scan(&currentHash); err != nil {
				return nil, err
			}
			if bcrypt.CompareHashAndPassword(currentHash, []byte(currentPassword)) != nil {
				return nil, ErrForbidden
			}
		}
		r, err := tx.ExecContext(ctx, "UPDATE users SET password_hash=?,revision=revision+1 WHERE id=? AND revision=? AND disabled=0", hash, target, revision)
		if err != nil {
			return nil, err
		}
		count, _ := r.RowsAffected()
		if count != 1 {
			return nil, ErrConflict
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM sessions WHERE user_id=?", target); err != nil {
			return nil, err
		}
		return map[string]bool{"changed": true, "sessionsRevoked": true}, nil
	})
	return err
}

func (s *Store) ApplyRole(ctx context.Context, actor, requestID, user string, ref model.ResourceRef, role model.RoleTemplate, revoke bool) error {
	_, err := s.MetadataMutation(ctx, actor, requestID, "role.apply", struct {
		User   string
		Ref    model.ResourceRef
		Role   model.RoleTemplate
		Revoke bool
	}{user, ref, role, revoke}, func(tx *sql.Tx) (any, error) {
		if !role.Builtin {
			var revision int64
			if err := tx.QueryRowContext(ctx, "SELECT revision FROM records WHERE namespace='role_template' AND id=?", role.ID).Scan(&revision); err != nil {
				return nil, err
			}
			if revision != role.Revision {
				return nil, ErrConflict
			}
		}
		var count int
		if ref.Kind == "node" {
			if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM nodes WHERE id=? AND revoked=0", ref.ID).Scan(&count); err != nil {
				return nil, err
			}
		} else if ref.Kind == "instance" {
			if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM instances WHERE id=?", ref.ID).Scan(&count); err != nil {
				return nil, err
			}
		}
		if count != 1 {
			return nil, sql.ErrNoRows
		}
		for _, action := range role.Actions {
			var err error
			if revoke {
				_, err = tx.ExecContext(ctx, "DELETE FROM grants WHERE user_id=? AND kind=? AND resource_id=? AND action=?", user, ref.Kind, ref.ID, action)
			} else {
				_, err = tx.ExecContext(ctx, "INSERT OR IGNORE INTO grants(user_id,kind,resource_id,action) VALUES(?,?,?,?)", user, ref.Kind, ref.ID, action)
			}
			if err != nil {
				return nil, err
			}
		}
		return map[string]bool{"applied": true}, nil
	})
	return err
}

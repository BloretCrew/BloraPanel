package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"blora.dev/panel/internal/model"
)

var ErrQuotaExceeded = errors.New("node instance quota exceeded")
var ErrNodeUnavailable = errors.New("node unavailable")

func (s *Store) CreateUser(ctx context.Context, u model.User, passwordHash []byte) (model.User, error) {
	if u.ID == "" {
		u.ID = model.ID()
	}
	u.Revision = 1
	_, err := s.DB.ExecContext(ctx, "INSERT INTO users(id,name,password_hash,admin,disabled,revision) VALUES(?,?,?,?,?,1)", u.ID, u.Name, passwordHash, u.Admin, u.Disabled)
	return u, err
}

// CreateUserMutation creates an account with a durable idempotency receipt.
// The password digest is used only for request identity; the plaintext and
// bcrypt hash never enter the receipt or audit record.
func (s *Store) CreateUserMutation(ctx context.Context, actor, requestID string, u model.User, passwordHash []byte, passwordDigest string) (model.User, error) {
	if u.ID == "" {
		u.ID = model.ID()
	}
	u.Revision = 1
	data, err := s.MetadataMutation(ctx, actor, requestID, "user.create", struct {
		Name           string `json:"name"`
		Admin          bool   `json:"admin"`
		PasswordDigest string `json:"passwordDigest"`
	}{u.Name, u.Admin, passwordDigest}, func(tx *sql.Tx) (any, error) {
		if _, err := tx.ExecContext(ctx, "INSERT INTO users(id,name,password_hash,admin,disabled,revision) VALUES(?,?,?,?,?,1)", u.ID, u.Name, passwordHash, u.Admin, u.Disabled); err != nil {
			return nil, err
		}
		return u, nil
	})
	if err != nil {
		return model.User{}, err
	}
	var created model.User
	if err := json.Unmarshal(data, &created); err != nil {
		return model.User{}, err
	}
	return created, nil
}
func (s *Store) UserByName(ctx context.Context, name string) (model.User, []byte, error) {
	var u model.User
	var hash []byte
	err := s.DB.QueryRowContext(ctx, "SELECT id,name,admin,disabled,revision,password_hash FROM users WHERE name=?", name).Scan(&u.ID, &u.Name, &u.Admin, &u.Disabled, &u.Revision, &hash)
	return u, hash, err
}
func (s *Store) User(ctx context.Context, id string) (model.User, error) {
	var u model.User
	err := s.DB.QueryRowContext(ctx, "SELECT id,name,admin,disabled,revision FROM users WHERE id=?", id).Scan(&u.ID, &u.Name, &u.Admin, &u.Disabled, &u.Revision)
	return u, err
}
func (s *Store) Users(ctx context.Context) ([]model.User, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT id,name,admin,disabled,revision FROM users ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Admin, &u.Disabled, &u.Revision); err != nil {
			return nil, err
		}
		items = append(items, u)
	}
	return items, rows.Err()
}
func (s *Store) Session(ctx context.Context, token string) (model.User, string, error) {
	var uid, csrf string
	err := s.DB.QueryRowContext(ctx, "SELECT user_id,csrf FROM sessions WHERE token_hash=? AND expires_at>?", Hash([]byte(token)), time.Now().Unix()).Scan(&uid, &csrf)
	if err != nil {
		return model.User{}, "", err
	}
	u, err := s.User(ctx, uid)
	if u.Disabled {
		return u, "", errors.New("account disabled")
	}
	return u, csrf, err
}
func (s *Store) NewSession(ctx context.Context, user string) (string, string, error) {
	token := model.ID() + model.ID()
	csrf := model.ID() + model.ID()
	_, err := s.DB.ExecContext(ctx, "INSERT INTO sessions(token_hash,user_id,csrf,expires_at) VALUES(?,?,?,?)", Hash([]byte(token)), user, csrf, time.Now().Add(12*time.Hour).Unix())
	return token, csrf, err
}
func (s *Store) DropSession(ctx context.Context, token string) error {
	_, err := s.DB.ExecContext(ctx, "DELETE FROM sessions WHERE token_hash=?", Hash([]byte(token)))
	return err
}
func (s *Store) SetGrant(ctx context.Context, g model.Grant, revoke bool) error {
	if revoke {
		_, err := s.DB.ExecContext(ctx, "DELETE FROM grants WHERE user_id=? AND kind=? AND resource_id=? AND action=?", g.UserID, g.Resource.Kind, g.Resource.ID, g.Action)
		return err
	}
	_, err := s.DB.ExecContext(ctx, "INSERT OR IGNORE INTO grants(user_id,kind,resource_id,action) VALUES(?,?,?,?)", g.UserID, g.Resource.Kind, g.Resource.ID, g.Action)
	return err
}

// SetGrantMutation applies an administrator grant change together with a
// durable idempotency receipt. A lost response can therefore be retried with
// the same request key without silently changing a different grant.
func (s *Store) SetGrantMutation(ctx context.Context, actor, requestID string, g model.Grant, revoke bool) error {
	_, err := s.MetadataMutation(ctx, actor, requestID, "grant.update", struct {
		Grant  model.Grant `json:"grant"`
		Revoke bool        `json:"revoke"`
	}{Grant: g, Revoke: revoke}, func(tx *sql.Tx) (any, error) {
		if revoke {
			_, err := tx.ExecContext(ctx, "DELETE FROM grants WHERE user_id=? AND kind=? AND resource_id=? AND action=?", g.UserID, g.Resource.Kind, g.Resource.ID, g.Action)
			if err != nil {
				return nil, err
			}
		} else {
			_, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO grants(user_id,kind,resource_id,action) VALUES(?,?,?,?)", g.UserID, g.Resource.Kind, g.Resource.ID, g.Action)
			if err != nil {
				return nil, err
			}
		}
		return map[string]bool{"updated": true}, nil
	})
	return err
}
func (s *Store) Grants(ctx context.Context) ([]model.Grant, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT user_id,kind,resource_id,action FROM grants ORDER BY user_id,kind,resource_id,action")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []model.Grant{}
	for rows.Next() {
		var g model.Grant
		if err := rows.Scan(&g.UserID, &g.Resource.Kind, &g.Resource.ID, &g.Action); err != nil {
			return nil, err
		}
		items = append(items, g)
	}
	return items, rows.Err()
}
func (s *Store) Allowed(ctx context.Context, u model.User, resource model.ResourceRef, action string) bool {
	if u.Disabled {
		return false
	}
	if u.Admin {
		return true
	}
	var count int
	err := s.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM grants WHERE user_id=? AND action=? AND ((kind=? AND resource_id=?) OR (kind='node' AND resource_id=?))", u.ID, action, resource.Kind, resource.ID, resource.NodeID).Scan(&count)
	return err == nil && count > 0
}
func (s *Store) Instances(ctx context.Context) ([]model.Instance, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT document FROM instances ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []model.Instance{}
	for rows.Next() {
		var b []byte
		var item model.Instance
		if err := rows.Scan(&b); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(b, &item); err != nil {
			return nil, err
		}
		if item.Deleted {
			continue
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (s *Store) Instance(ctx context.Context, id string) (model.Instance, error) {
	var i model.Instance
	var b []byte
	err := s.DB.QueryRowContext(ctx, "SELECT document FROM instances WHERE id=?", id).Scan(&b)
	if err != nil {
		return i, err
	}
	if err := json.Unmarshal(b, &i); err != nil {
		return i, err
	}
	if i.Deleted {
		return i, sql.ErrNoRows
	}
	return i, nil
}

// DeleteInstance records an administrator-only tombstone after checking the
// authoritative lifecycle state.  It deliberately keeps the row and its
// identity so a stale desktop shortcut cannot resolve to a later same-named
// resource.  No daemon side effect is needed: only STOPPED instances with no
// active task may be tombstoned, and the daemon ignores the tombstone on its
// next startup snapshot.
func (s *Store) DeleteInstance(ctx context.Context, actor, id, requestID string) (model.Instance, error) {
	data, err := s.MetadataMutation(ctx, actor, requestID, "instance.delete", struct {
		ID string `json:"id"`
	}{ID: id}, func(tx *sql.Tx) (any, error) {
		var b []byte
		if err := tx.QueryRowContext(ctx, "SELECT document FROM instances WHERE id=?", id).Scan(&b); err != nil {
			return nil, err
		}
		var i model.Instance
		if err := json.Unmarshal(b, &i); err != nil {
			return nil, err
		}
		if i.Deleted {
			return nil, sql.ErrNoRows
		}
		if i.State != "STOPPED" {
			return nil, ErrConflict
		}
		var active int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM tasks
			WHERE state NOT IN ('SUCCEEDED','FAILED','CANCELLED','INTERRUPTED')
			AND (resource_key=? OR json_extract(document,'$.resource.id')=?
			OR json_extract(document,'$.payload.sourceResource.id')=?
			OR json_extract(document,'$.payload.targetResource.id')=?)`, "instance:"+id, id, id, id).Scan(&active); err != nil {
			return nil, err
		}
		if active != 0 {
			return nil, ErrConflict
		}
		i.Deleted = true
		i.Revision++
		encoded, err := json.Marshal(i)
		if err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, "UPDATE instances SET document=? WHERE id=?", encoded, id); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM grants WHERE kind='instance' AND resource_id=?", id); err != nil {
			return nil, err
		}
		return i, nil
	})
	if err != nil {
		return model.Instance{}, err
	}
	var deleted model.Instance
	if err := json.Unmarshal(data, &deleted); err != nil {
		return model.Instance{}, err
	}
	return deleted, nil
}

// CreateInstance repeats authorization and quota checks in the insertion transaction.
func (s *Store) CreateInstance(ctx context.Context, u model.User, i model.Instance, requestID string) (model.Instance, error) {
	if requestID == "" || len(requestID) > 128 {
		return i, errors.New("request ID is required (at most 128 bytes)")
	}
	requestBytes, err := json.Marshal(i)
	if err != nil {
		return i, err
	}
	digest := Hash(requestBytes)
	receiptKey := u.ID + ":" + requestID
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return i, err
	}
	defer tx.Rollback()
	var disabled, admin bool
	if err := tx.QueryRowContext(ctx, "SELECT disabled,admin FROM users WHERE id=?", u.ID).Scan(&disabled, &admin); err != nil {
		return i, err
	}
	if disabled {
		return i, errors.New("forbidden")
	}
	if !admin {
		var count int
		if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM grants WHERE user_id=? AND kind='node' AND resource_id=? AND action='instance.create'", u.ID, i.NodeID).Scan(&count); err != nil {
			return i, err
		}
		if count == 0 {
			return i, errors.New("forbidden")
		}
	}
	var receiptBytes []byte
	err = tx.QueryRowContext(ctx, "SELECT document FROM records WHERE namespace='create_instance' AND id=?", receiptKey).Scan(&receiptBytes)
	if err == nil {
		var receipt struct {
			Digest   string         `json:"digest"`
			Instance model.Instance `json:"instance"`
		}
		if err := json.Unmarshal(receiptBytes, &receipt); err != nil {
			return i, err
		}
		if receipt.Digest != digest {
			return i, ErrRequestMismatch
		}
		return receipt.Instance, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return i, err
	}
	var revoked, maintenance bool
	var quota, count int
	if err := tx.QueryRowContext(ctx, "SELECT revoked,maintenance,quota FROM nodes WHERE id=?", i.NodeID).Scan(&revoked, &maintenance, &quota); err != nil {
		return i, err
	}
	if revoked || maintenance {
		return i, ErrNodeUnavailable
	}
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM instances WHERE node_id=? AND COALESCE(json_extract(document,'$.deleted'),0)=0", i.NodeID).Scan(&count); err != nil {
		return i, err
	}
	if count >= quota {
		return i, ErrQuotaExceeded
	}
	// Native creation additionally requires host authority. Ordinary grants never unlock host execution.
	if i.Config.Mode != "container" && !admin {
		var count int
		if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM grants WHERE user_id=? AND kind='node' AND resource_id=? AND action='host.manage'", u.ID, i.NodeID).Scan(&count); err != nil {
			return i, err
		}
		if count == 0 {
			return i, errors.New("native mode requires host.manage")
		}
	}
	i.ID = model.ID()
	i.Revision = 1
	i.ConfigRevision = 1
	if i.Config.Autostart {
		var nodeBytes []byte
		if err := tx.QueryRowContext(ctx, "SELECT document FROM nodes WHERE id=?", i.NodeID).Scan(&nodeBytes); err != nil {
			return i, err
		}
		var node model.Node
		if err := json.Unmarshal(nodeBytes, &node); err != nil {
			return i, err
		}
		i.AutostartActor, i.AutostartAfter = u.ID, node.StartupID
	}
	i.State = "STOPPED"
	b, err := json.Marshal(i)
	if err != nil {
		return i, err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO instances(id,node_id,document) VALUES(?,?,?)", i.ID, i.NodeID, b); err != nil {
		return i, err
	}
	for _, action := range []string{"instance.read", "instance.start", "instance.stop", "instance.restart", "instance.kill", "instance.configure", "file.read", "file.write", "terminal.read", "terminal.input"} {
		if _, err := tx.ExecContext(ctx, "INSERT INTO grants(user_id,kind,resource_id,action) VALUES(?,'instance',?,?)", u.ID, i.ID, action); err != nil {
			return i, err
		}
	}
	receiptBytes, err = json.Marshal(map[string]any{"digest": digest, "instance": i})
	if err != nil {
		return i, err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO records(namespace,id,revision,document) VALUES('create_instance',?,1,?)", receiptKey, receiptBytes); err != nil {
		return i, err
	}
	return i, tx.Commit()
}
func (s *Store) PutInstance(ctx context.Context, i model.Instance) error {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx, "UPDATE instances SET document=? WHERE id=? AND node_id=?", b, i.ID, i.NodeID)
	return err
}
func (s *Store) Node(ctx context.Context, id string) (model.Node, []byte, error) {
	var n model.Node
	var key, b []byte
	var revoked bool
	var name string
	var maintenance bool
	var quota int
	var configRevision int64
	err := s.DB.QueryRowContext(ctx, "SELECT document,public_key,revoked,name,maintenance,quota,settings_revision FROM nodes WHERE id=?", id).Scan(&b, &key, &revoked, &name, &maintenance, &quota, &configRevision)
	if err != nil {
		return n, nil, err
	}
	if revoked {
		return n, nil, sql.ErrNoRows
	}
	err = json.Unmarshal(b, &n)
	if n.Tags == nil {
		n.Tags = []string{}
	}
	n.Name, n.Maintenance, n.Quota, n.ConfigRevision = name, maintenance, quota, configRevision
	if maintenance {
		n.State = "MAINTENANCE"
	}
	return n, key, err
}
func (s *Store) Nodes(ctx context.Context) ([]model.Node, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT document,name,maintenance,quota,settings_revision FROM nodes WHERE revoked=0 ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []model.Node{}
	for rows.Next() {
		var b []byte
		var n model.Node
		var name string
		var maintenance bool
		var quota int
		var revision int64
		if err := rows.Scan(&b, &name, &maintenance, &quota, &revision); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(b, &n); err != nil {
			return nil, err
		}
		if n.Tags == nil {
			n.Tags = []string{}
		}
		n.Name, n.Maintenance, n.Quota, n.ConfigRevision = name, maintenance, quota, revision
		if maintenance {
			n.State = "MAINTENANCE"
		}
		items = append(items, n)
	}
	return items, rows.Err()
}
func (s *Store) PutNode(ctx context.Context, n model.Node) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var oldBytes []byte
	var generation uint64
	if err := tx.QueryRowContext(ctx, "SELECT document,name,maintenance,quota,settings_revision,generation FROM nodes WHERE id=? AND revoked=0", n.ID).Scan(&oldBytes, &n.Name, &n.Maintenance, &n.Quota, &n.ConfigRevision, &generation); err != nil {
		return err
	}
	if n.Generation != generation {
		return ErrConflict
	}
	var old model.Node
	if err := json.Unmarshal(oldBytes, &old); err != nil {
		return err
	}
	// Daemon heartbeats carry runtime facts and may not know administrator
	// metadata. Preserve the latter while accepting the fresh observation.
	n.Group, n.Tags = old.Group, append([]string{}, old.Tags...)
	n.Revision = old.Revision + 1
	if n.Maintenance {
		n.State = "MAINTENANCE"
	}
	b, err := json.Marshal(n)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "UPDATE nodes SET document=? WHERE id=?", b, n.ID)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (s *Store) Enrollment(ctx context.Context, name string) (string, error) {
	token := model.ID() + model.ID()
	_, err := s.DB.ExecContext(ctx, "INSERT INTO enrollments(token_hash,name,expires_at) VALUES(?,?,?)", Hash([]byte(token)), name, time.Now().Add(10*time.Minute).Unix())
	return token, err
}

// EnrollmentMutation issues one enrollment ticket per administrator request.
// The token is returned from the durable receipt on a lost-response retry;
// the receipt lives in the protected state database and is never written to
// the audit result.
func (s *Store) EnrollmentMutation(ctx context.Context, actor, requestID, name string) (string, error) {
	token := model.ID() + model.ID()
	data, err := s.MetadataMutation(ctx, actor, requestID, "node.enrollment", struct {
		Name string `json:"name"`
	}{Name: name}, func(tx *sql.Tx) (any, error) {
		if _, err := tx.ExecContext(ctx, "INSERT INTO enrollments(token_hash,name,expires_at) VALUES(?,?,?)", Hash([]byte(token)), name, time.Now().Add(10*time.Minute).Unix()); err != nil {
			return nil, err
		}
		return struct {
			Token string `json:"token"`
		}{Token: token}, nil
	})
	if err != nil {
		return "", err
	}
	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(data, &result); err != nil || result.Token == "" {
		if err == nil {
			err = errors.New("enrollment receipt missing token")
		}
		return "", err
	}
	return result.Token, nil
}
func (s *Store) Enroll(ctx context.Context, token string, key []byte) (model.Node, error) {
	var n model.Node
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return n, err
	}
	defer tx.Rollback()
	var name string
	err = tx.QueryRowContext(ctx, "SELECT name FROM enrollments WHERE token_hash=? AND used=0 AND expires_at>?", Hash([]byte(token)), time.Now().Unix()).Scan(&name)
	if err != nil {
		return n, err
	}
	n = model.Node{ID: model.ID(), Name: name, State: "OFFLINE", Revision: 1, Capabilities: map[string]string{}, Tags: []string{}}
	b, _ := json.Marshal(n)
	if _, err = tx.ExecContext(ctx, "INSERT INTO nodes(id,name,public_key,document) VALUES(?,?,?,?)", n.ID, n.Name, key, b); err != nil {
		return n, err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE enrollments SET used=1 WHERE token_hash=?", Hash([]byte(token))); err != nil {
		return n, err
	}
	return n, tx.Commit()
}
func (s *Store) AdvanceGeneration(ctx context.Context, id string, old uint64) (uint64, error) {
	r, err := s.DB.ExecContext(ctx, "UPDATE nodes SET generation=generation+1 WHERE id=? AND generation=? AND revoked=0", id, old)
	if err != nil {
		return 0, err
	}
	count, _ := r.RowsAffected()
	if count != 1 {
		return 0, ErrConflict
	}
	return old + 1, nil
}
func (s *Store) Generation(ctx context.Context, id string) (uint64, error) {
	var generation uint64
	err := s.DB.QueryRowContext(ctx, "SELECT generation FROM nodes WHERE id=? AND revoked=0", id).Scan(&generation)
	return generation, err
}

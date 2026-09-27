package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"blora.dev/panel/internal/model"
)

type NodeSettings struct {
	Name        *string   `json:"name,omitempty"`
	Group       *string   `json:"group,omitempty"`
	Tags        *[]string `json:"tags,omitempty"`
	Maintenance *bool     `json:"maintenance,omitempty"`
	Quota       *int      `json:"quota,omitempty"`
	Revision    int64     `json:"revision"`
}

func (s *Store) SetNodeSettings(ctx context.Context, actor, id, requestID string, p NodeSettings) (model.Node, error) {
	data, err := s.MetadataMutation(ctx, actor, requestID, "node.settings", struct {
		ID    string
		Patch NodeSettings
	}{id, p}, func(tx *sql.Tx) (any, error) {
		var b []byte
		var name string
		var maintenance bool
		var quota int
		var revision int64
		if err := tx.QueryRowContext(ctx, "SELECT document,name,maintenance,quota,settings_revision FROM nodes WHERE id=? AND revoked=0", id).Scan(&b, &name, &maintenance, &quota, &revision); err != nil {
			return nil, err
		}
		if revision != p.Revision {
			return nil, ErrConflict
		}
		if p.Name != nil {
			name = *p.Name
		}
		var n model.Node
		if err := json.Unmarshal(b, &n); err != nil {
			return nil, err
		}
		if n.Tags == nil {
			n.Tags = []string{}
		}
		if p.Group != nil {
			if len(*p.Group) > 120 {
				return nil, errors.New("node group outside supported bounds")
			}
			n.Group = *p.Group
		}
		if p.Tags != nil {
			if len(*p.Tags) > 32 {
				return nil, errors.New("node tags exceed 32 items")
			}
			seen := map[string]bool{}
			for _, tag := range *p.Tags {
				if strings.TrimSpace(tag) == "" || len(tag) > 80 || seen[tag] {
					return nil, errors.New("node tags must be non-empty, unique and at most 80 bytes")
				}
				seen[tag] = true
			}
			n.Tags = append([]string(nil), (*p.Tags)...)
		}
		if p.Maintenance != nil {
			maintenance = *p.Maintenance
		}
		if p.Quota != nil {
			quota = *p.Quota
		}
		if name == "" || len(name) > 80 || quota < 1 || quota > 10000 {
			return nil, errors.New("node name/quota outside supported bounds")
		}
		n.Name, n.Maintenance, n.Quota, n.ConfigRevision = name, maintenance, quota, revision+1
		n.Revision++
		if maintenance {
			n.State = "MAINTENANCE"
		} else if n.State == "MAINTENANCE" {
			n.State = "OFFLINE"
		}
		b, err := json.Marshal(n)
		if err != nil {
			return nil, err
		}
		_, err = tx.ExecContext(ctx, "UPDATE nodes SET name=?,maintenance=?,quota=?,settings_revision=?,document=? WHERE id=?", name, maintenance, quota, n.ConfigRevision, b, id)
		return n, err
	})
	var n model.Node
	if err == nil {
		err = json.Unmarshal(data, &n)
	}
	return n, err
}

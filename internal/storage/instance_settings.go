package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"blora.dev/panel/internal/model"
)

func requireResource(ctx context.Context, tx *sql.Tx, actor string, resource model.ResourceRef, action string) error {
	var admin, disabled bool
	if err := tx.QueryRowContext(ctx, "SELECT admin,disabled FROM users WHERE id=?", actor).Scan(&admin, &disabled); err != nil {
		return err
	}
	if disabled {
		return ErrForbidden
	}
	if admin {
		return nil
	}
	var count int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM grants WHERE user_id=? AND action=? AND ((kind=? AND resource_id=?) OR (kind='node' AND resource_id=?))", actor, action, resource.Kind, resource.ID, resource.NodeID).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return ErrForbidden
	}
	return nil
}

func (s *Store) SetInstanceSettings(ctx context.Context, actor, id, requestID string, p model.InstanceSettings) (model.Instance, error) {
	var current model.Instance
	data, err := s.metadataMutation(ctx, actor, requestID, "instance.configure", struct {
		ID    string
		Patch model.InstanceSettings
	}{id, p}, func(tx *sql.Tx) error {
		var b []byte
		if err := tx.QueryRowContext(ctx, "SELECT document FROM instances WHERE id=?", id).Scan(&b); err != nil {
			return err
		}
		if err := json.Unmarshal(b, &current); err != nil {
			return err
		}
		ref := model.ResourceRef{Kind: "instance", ID: id, NodeID: current.NodeID}
		if err := requireResource(ctx, tx, actor, ref, "instance.configure"); err != nil {
			return err
		}
		if p.Config != nil && (current.Config.Mode == "native" || p.Config.Mode == "native") {
			return requireResource(ctx, tx, actor, model.ResourceRef{Kind: "node", ID: current.NodeID}, "host.manage")
		}
		return nil
	}, func(tx *sql.Tx) (any, error) {
		if p.Revision != current.ConfigRevision {
			return nil, ErrConflict
		}
		if p.Name == nil && p.Group == nil && p.Tags == nil && p.Config == nil {
			return nil, errors.New("配置补丁为空")
		}
		if p.Name != nil {
			if strings.TrimSpace(*p.Name) == "" || len(*p.Name) > 120 {
				return nil, errors.New("名称需要1～120字节")
			}
			current.Name = *p.Name
		}
		if p.Group != nil {
			if len(*p.Group) > 120 {
				return nil, errors.New("分组不能超过120字节")
			}
			current.Group = *p.Group
		}
		if p.Tags != nil {
			if len(*p.Tags) > 32 {
				return nil, errors.New("标签不能超过32项")
			}
			seen := map[string]bool{}
			for _, tag := range *p.Tags {
				if strings.TrimSpace(tag) == "" || len(tag) > 80 || seen[tag] {
					return nil, errors.New("标签不能为空、重复或超过80字节")
				}
				seen[tag] = true
			}
			current.Tags = *p.Tags
		}
		if p.Config != nil {
			if err := model.NormalizeInstanceConfig(p.Config); err != nil {
				return nil, err
			}
			if model.LaunchConfiguration(current.Config) != model.LaunchConfiguration(*p.Config) {
				if current.State != "STOPPED" && current.State != "START_FAILED" {
					return nil, ErrConflict
				}
				var count int
				if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE state NOT IN ('SUCCEEDED','FAILED','CANCELLED','INTERRUPTED') AND (resource_key=? OR json_extract(document,'$.payload.sourceResource.id')=?)", "instance:"+id, id).Scan(&count); err != nil {
					return nil, err
				}
				if count > 0 {
					return nil, ErrConflict
				}
			}
			if p.Config.Autostart {
				ref := model.ResourceRef{Kind: "instance", ID: id, NodeID: current.NodeID}
				if err := requireResource(ctx, tx, actor, ref, "instance.start"); err != nil {
					return nil, err
				}
				if !current.Config.Autostart || current.AutostartActor != actor {
					var nodeBytes []byte
					if err := tx.QueryRowContext(ctx, "SELECT document FROM nodes WHERE id=? AND revoked=0", current.NodeID).Scan(&nodeBytes); err != nil {
						return nil, err
					}
					var n model.Node
					if err := json.Unmarshal(nodeBytes, &n); err != nil {
						return nil, err
					}
					current.AutostartActor, current.AutostartAfter = actor, n.StartupID
				}
			} else {
				current.AutostartActor, current.AutostartAfter = "", ""
			}
			current.Config = *p.Config
		}
		current.ConfigRevision++
		current.Revision++
		b, err := json.Marshal(current)
		if err != nil {
			return nil, err
		}
		if _, err = tx.ExecContext(ctx, "UPDATE instances SET document=? WHERE id=?", b, id); err != nil {
			return nil, err
		}
		return current, nil
	})
	if err != nil {
		return current, err
	}
	err = json.Unmarshal(data, &current)
	return current, err
}

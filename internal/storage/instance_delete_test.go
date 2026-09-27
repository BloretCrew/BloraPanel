package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"blora.dev/panel/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func TestDeleteInstanceLeavesIdentityTombstoneAndReleasesQuota(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	hash, err := bcrypt.GenerateFromPassword([]byte("delete-instance-test-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	admin, err := s.CreateUser(ctx, model.User{Name: "delete-admin", Admin: true}, hash)
	if err != nil {
		t.Fatal(err)
	}
	node := model.Node{ID: "delete-node", Name: "delete-node", State: "ONLINE", StartupID: model.ID(), Capabilities: map[string]string{}}
	nodeBytes, _ := json.Marshal(node)
	if _, err := s.DB.ExecContext(ctx, `INSERT INTO nodes(id,name,public_key,document) VALUES(?,?,?,?)`, node.ID, node.Name, []byte("key"), nodeBytes); err != nil {
		t.Fatal(err)
	}
	i := model.Instance{ID: "delete-instance", NodeID: node.ID, Name: "same-name", Config: model.InstanceConfig{Mode: "native", Command: []string{"/bin/sh"}}, State: "STOPPED", Revision: 1, ConfigRevision: 1}
	instanceBytes, _ := json.Marshal(i)
	if _, err := s.DB.ExecContext(ctx, `INSERT INTO instances(id,node_id,document) VALUES(?,?,?)`, i.ID, i.NodeID, instanceBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB.ExecContext(ctx, `INSERT INTO grants(user_id,kind,resource_id,action) VALUES(?,?,?,?)`, admin.ID, "instance", i.ID, "instance.read"); err != nil {
		t.Fatal(err)
	}

	deleted, err := s.DeleteInstance(ctx, admin.ID, i.ID, "delete-once")
	if err != nil || !deleted.Deleted {
		t.Fatalf("delete result=%+v err=%v", deleted, err)
	}
	if replay, err := s.DeleteInstance(ctx, admin.ID, i.ID, "delete-once"); err != nil || !replay.Deleted {
		t.Fatalf("idempotent delete replay failed: %+v %v", replay, err)
	}
	if _, err := s.DeleteInstance(ctx, admin.ID, i.ID, "delete-again"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("deleted resource accepted with a new key: %v", err)
	}
	if _, err := s.Instance(ctx, i.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("deleted resource still readable: %v", err)
	}
	items, err := s.Instances(ctx)
	if err != nil || len(items) != 0 {
		t.Fatalf("deleted resource listed: %d %v", len(items), err)
	}
	var grants int
	if err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM grants WHERE resource_id=?`, i.ID).Scan(&grants); err != nil || grants != 0 {
		t.Fatalf("grants retained: %d %v", grants, err)
	}
	var audits int
	if err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM audit WHERE action='instance.delete' AND resource_key=?`, "instance:"+i.ID).Scan(&audits); err != nil || audits != 1 {
		t.Fatalf("delete audit missing: %d %v", audits, err)
	}
}

func TestDeleteInstanceRejectsRunningOrActiveResource(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	hash, _ := bcrypt.GenerateFromPassword([]byte("delete-instance-test-password"), bcrypt.MinCost)
	admin, err := s.CreateUser(ctx, model.User{Name: "delete-admin", Admin: true}, hash)
	if err != nil {
		t.Fatal(err)
	}
	nodeBytes, _ := json.Marshal(model.Node{ID: "delete-node", Name: "delete-node", StartupID: model.ID(), Capabilities: map[string]string{}})
	if _, err := s.DB.ExecContext(ctx, `INSERT INTO nodes(id,name,public_key,document) VALUES(?,?,?,?)`, "delete-node", "delete-node", []byte("key"), nodeBytes); err != nil {
		t.Fatal(err)
	}
	for _, state := range []string{"RUNNING", "STOPPED"} {
		i := model.Instance{ID: "delete-" + state, NodeID: "delete-node", Name: state, Config: model.InstanceConfig{Mode: "native", Command: []string{"/bin/sh"}}, State: state, Revision: 1, ConfigRevision: 1}
		b, _ := json.Marshal(i)
		if _, err := s.DB.ExecContext(ctx, `INSERT INTO instances(id,node_id,document) VALUES(?,?,?)`, i.ID, i.NodeID, b); err != nil {
			t.Fatal(err)
		}
		if state == "STOPPED" {
			if _, _, err := s.Accept(ctx, model.Task{ActorID: admin.ID, RequestID: "active-delete", Resource: model.ResourceRef{Kind: "instance", ID: i.ID, NodeID: i.NodeID}, Action: "metadata.check", Payload: []byte(`{}`)}); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := s.DeleteInstance(ctx, admin.ID, i.ID, "delete-"+state); !errors.Is(err, ErrConflict) {
			t.Fatalf("%s delete did not conflict: %v", state, err)
		}
	}
}

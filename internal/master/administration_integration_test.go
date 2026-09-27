package master

import (
	"context"
	"encoding/json"
	"testing"

	"blora.dev/panel/internal/model"
)

func TestRolesRemainScopedAndAccountChangesRevokeSessions(t *testing.T) {
	f := newFileFixture(t)
	f.reader.request("GET", "/roles", nil, "", 403)
	roles := f.admin.request("GET", "/roles", nil, "", 200)
	var items []model.RoleTemplate
	if err := json.Unmarshal(roles["items"], &items); err != nil || len(items) < 3 {
		t.Fatalf("default roles: %v", err)
	}
	createKey := model.ID()
	created := f.admin.request("POST", "/users", map[string]any{"name": "idempotent-user", "password": "a-strong-password-123", "admin": false}, createKey, 201)
	recreated := f.admin.request("POST", "/users", map[string]any{"name": "idempotent-user", "password": "a-strong-password-123", "admin": false}, createKey, 201)
	var firstUser, replayUser model.User
	if err := json.Unmarshal(created["user"], &firstUser); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(recreated["user"], &replayUser); err != nil || replayUser.ID != firstUser.ID {
		t.Fatalf("create user replay changed identity: %+v %+v (%v)", firstUser, replayUser, err)
	}
	f.admin.request("POST", "/users", map[string]any{"name": "idempotent-user", "password": "different-password-123", "admin": false}, createKey, 409)
	enrollmentKey := model.ID()
	enrollment := f.admin.request("POST", "/nodes/enrollments", map[string]string{"name": "idempotent-enrollment"}, enrollmentKey, 201)
	enrollmentReplay := f.admin.request("POST", "/nodes/enrollments", map[string]string{"name": "idempotent-enrollment"}, enrollmentKey, 201)
	var firstTicket, replayTicket string
	if err := json.Unmarshal(enrollment["token"], &firstTicket); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(enrollmentReplay["token"], &replayTicket); err != nil || replayTicket != firstTicket {
		t.Fatalf("enrollment replay changed ticket: %q %q (%v)", firstTicket, replayTicket, err)
	}
	f.admin.request("POST", "/nodes/enrollments", map[string]string{"name": "different-enrollment"}, enrollmentKey, 409)
	ref := instanceRef(f.instances[0])
	grant := model.Grant{UserID: f.readerUser.ID, Resource: ref, Action: "instance.start"}
	grantKey := model.ID()
	f.admin.request("POST", "/grants", grant, grantKey, 200)
	f.admin.request("POST", "/grants", grant, grantKey, 200)
	f.admin.request("POST", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: ref, Action: "instance.stop"}, grantKey, 409)
	revokeKey := model.ID()
	f.admin.request("DELETE", "/grants", grant, revokeKey, 200)
	f.admin.request("POST", "/grants", grant, grantKey, 200)
	if f.store.Allowed(context.Background(), f.readerUser, ref, "instance.start") {
		t.Fatal("replaying an old grant restored revoked authority")
	}
	f.admin.request("POST", "/grants", grant, model.ID(), 200)
	f.admin.request("DELETE", "/grants", grant, revokeKey, 200)
	if !f.store.Allowed(context.Background(), f.readerUser, ref, "instance.start") {
		t.Fatal("replaying an old revocation removed newly granted authority")
	}
	key := model.ID()
	payload := map[string]any{"userId": f.readerUser.ID, "resource": ref, "revision": 1}
	f.admin.request("POST", "/roles/operator/apply", payload, key, 200)
	f.admin.request("POST", "/roles/operator/apply", payload, key, 200)
	if !f.store.Allowed(context.Background(), f.readerUser, ref, "instance.start") || !f.store.Allowed(context.Background(), f.readerUser, ref, "file.write") {
		t.Fatal("operator role did not expand to explicit grants")
	}
	if f.store.Allowed(context.Background(), f.readerUser, instanceRef(f.instances[1]), "file.write") || f.store.Allowed(context.Background(), f.readerUser, model.ResourceRef{Kind: "node", ID: ref.NodeID}, "host.manage") {
		t.Fatal("role scope leaked to another resource or host")
	}
	f.admin.request("POST", "/roles/node-admin/apply", payload, model.ID(), 400)
	custom := model.RoleTemplate{ID: "custom-reader", Name: "自定义只读", Actions: []string{"file.read"}}
	f.admin.request("PUT", "/roles/"+custom.ID, custom, model.ID(), 200)
	f.admin.request("PUT", "/roles/"+custom.ID, custom, model.ID(), 409)
	payload["revoke"] = true
	f.admin.request("POST", "/roles/operator/apply", payload, model.ID(), 200)
	if f.store.Allowed(context.Background(), f.readerUser, ref, "file.write") {
		t.Fatal("explicit template revocation retained authority")
	}
	f.admin.request("PATCH", "/users/"+f.readerUser.ID, map[string]any{"disabled": true, "revision": f.readerUser.Revision}, model.ID(), 200)
	f.admin.request("POST", "/users", map[string]any{"name": "missing-key", "password": model.ID() + model.ID()}, "", 400)
	f.admin.request("POST", "/nodes/enrollments", map[string]any{"name": "missing-key-node"}, "", 400)
	f.reader.request("GET", "/session", nil, "", 401)
	f.admin.request("PATCH", "/users/"+f.adminUser.ID, map[string]any{"admin": false, "revision": f.adminUser.Revision}, model.ID(), 409)
	f.admin.request("PATCH", "/users/"+f.readerUser.ID, map[string]any{"disabled": false, "revision": 2}, model.ID(), 200)
	f.reader.login(f.readerUser.Name, f.password)
	newPassword := model.ID()
	f.reader.request("POST", "/users/"+f.readerUser.ID+"/password", map[string]any{"password": newPassword, "currentPassword": f.password, "revision": 3}, "", 400)
	passwordKey := model.ID()
	f.reader.request("POST", "/users/"+f.readerUser.ID+"/password", map[string]any{"password": newPassword, "currentPassword": "incorrect", "revision": 3}, passwordKey, 403)
	f.reader.request("POST", "/users/"+f.readerUser.ID+"/password", map[string]any{"password": newPassword, "currentPassword": f.password, "revision": 3}, passwordKey, 200)
	f.reader.request("GET", "/session", nil, "", 401)
	f.reader.login(f.readerUser.Name, newPassword)
	f.reader.request("GET", "/session", nil, "", 200)
	// The first successful request revoked the original session. After logging
	// in with the new password, retrying its key must replay the durable receipt
	// before checking the now-stale current password.
	f.reader.request("POST", "/users/"+f.readerUser.ID+"/password", map[string]any{"password": newPassword, "currentPassword": f.password, "revision": 3}, passwordKey, 200)
	f.reader.request("POST", "/users/"+f.readerUser.ID+"/password", map[string]any{"password": "different-strong-password-123", "currentPassword": f.password, "revision": 3}, passwordKey, 409)
}

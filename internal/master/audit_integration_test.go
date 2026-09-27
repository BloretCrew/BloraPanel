package master

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

func TestAuditEndpointFiltersAndNodeRevokeIsIdempotent(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open(filepath.Join(t.TempDir(), "master.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	password := model.ID()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	admin, err := store.CreateUser(ctx, model.User{Name: "audit-admin", Admin: true}, hash)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := store.CreateUser(ctx, model.User{Name: "audit-reader"}, hash)
	if err != nil {
		t.Fatal(err)
	}
	var app *Server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { app.ServeHTTP(w, r) }))
	defer server.Close()
	app = New(Options{Store: store, Origin: server.URL})
	defer app.Close()
	newClient := func() *testClient {
		jar, _ := cookiejar.New(nil)
		client := *server.Client()
		client.Jar = jar
		client.Timeout = 10 * time.Second
		return &testClient{t: t, client: &client, base: server.URL}
	}
	adminClient := newClient()
	adminClient.login(admin.Name, password)
	readerClient := newClient()
	readerClient.login(reader.Name, password)

	ref := model.ResourceRef{Kind: "node", ID: "audit-node", NodeID: "audit-node"}
	if err := store.Audit(ctx, admin.ID, ref, "node.settings", "audit-request", "applied"); err != nil {
		t.Fatal(err)
	}
	result := adminClient.request("GET", "/audit?userId="+admin.ID+"&nodeId=audit-node&action=node.settings&limit=1", nil, "", 200)
	var items []model.Audit
	if err := json.Unmarshal(result["items"], &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].RequestID != "audit-request" || items[0].NodeID != "audit-node" {
		t.Fatalf("audit response = %+v", items)
	}
	readerClient.request("GET", "/audit", nil, "", 403)

	token, err := store.Enrollment(ctx, "revocable-node")
	if err != nil {
		t.Fatal(err)
	}
	node, err := store.Enroll(ctx, token, make([]byte, ed25519.PublicKeySize))
	if err != nil {
		t.Fatal(err)
	}
	key := model.ID()
	revoked := adminClient.request("POST", "/nodes/"+node.ID+"/revoke", nil, key, 200)
	var generation uint64
	if err := json.Unmarshal(revoked["generation"], &generation); err != nil || generation != 1 {
		t.Fatalf("revoke response generation=%d err=%v", generation, err)
	}
	retry := adminClient.request("POST", "/nodes/"+node.ID+"/revoke", nil, key, 200)
	if err := json.Unmarshal(retry["generation"], &generation); err != nil || generation != 1 {
		t.Fatalf("idempotent revoke generation=%d err=%v", generation, err)
	}
	filtered := adminClient.request("GET", "/audit?nodeId="+node.ID+"&action=node.revoke", nil, "", 200)
	if err := json.Unmarshal(filtered["items"], &items); err != nil || len(items) != 1 || items[0].ActorID != admin.ID {
		t.Fatalf("node revoke audit = %+v err=%v", items, err)
	}
}

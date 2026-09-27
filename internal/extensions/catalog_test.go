package extensions

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalogPublishListLoadAndIdentityGuards(t *testing.T) {
	c, err := NewCatalog(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte("reference-v1")
	h := sha256.Sum256(payload)
	p := Package{Manifest: Manifest{AppID: "reference.app", PackageVersion: "1.0.0", HostAPIVersion: 1, Title: "Reference", Capabilities: []string{"window.open"}}, SHA256: hex.EncodeToString(h[:]), Payload: payload}
	if err := c.Publish(p); err != nil {
		t.Fatal(err)
	}
	if err := c.Publish(p); err == nil {
		t.Fatal("catalog replaced an existing version")
	}
	p2 := p
	p2.Manifest.PackageVersion = "2.0.0"
	p2.Payload = []byte("reference-v2")
	h = sha256.Sum256(p2.Payload)
	p2.SHA256 = hex.EncodeToString(h[:])
	if err := c.Publish(p2); err != nil {
		t.Fatal(err)
	}
	entries, err := c.List()
	if err != nil || len(entries) != 2 {
		t.Fatalf("catalog list: %+v %v", entries, err)
	}
	if entries[0].Manifest.PackageVersion != "2.0.0" {
		t.Fatalf("catalog versions not newest first: %+v", entries)
	}
	got, err := c.Load("reference.app", "1.0.0")
	if err != nil || string(got.Payload) != string(payload) {
		t.Fatalf("catalog load: %q %v", got.Payload, err)
	}
	path := filepath.Join(c.root, "reference.app", "1.0.0.json")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(original, '\n', '{'), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Load("reference.app", "1.0.0"); err == nil {
		t.Fatal("catalog accepted trailing or malformed JSON")
	}
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	var envelope catalogEnvelope
	if err := json.Unmarshal(original, &envelope); err != nil {
		t.Fatal(err)
	}
	envelope.Payload = []byte("tampered")
	tampered, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, tampered, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Load("reference.app", "1.0.0"); err == nil {
		t.Fatal("catalog accepted payload digest mismatch")
	}
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Load("reference.app", "../1.0.0"); err == nil {
		t.Fatal("path traversal package accepted")
	}
	if err := os.Symlink(filepath.Join(c.root, "reference.app", "1.0.0.json"), filepath.Join(c.root, "reference.app", "1.0.1.json")); err == nil {
		if _, err := c.Load("reference.app", "1.0.1"); err == nil {
			t.Fatal("symlink catalog entry accepted")
		}
	}
}

func TestRemoteCatalogUsesBoundedHTTPSSourceAndRejectsRedirects(t *testing.T) {
	payload := []byte("remote-reference-v1")
	digest := sha256.Sum256(payload)
	manifest := Manifest{AppID: "remote.app", PackageVersion: "1.0.0", HostAPIVersion: 1, Title: "Remote", Capabilities: []string{"window.open"}}
	entry := CatalogEntry{Manifest: manifest, SHA256: hex.EncodeToString(digest[:]), SignaturePresent: false, Size: int64(len(payload))}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/registry/index.json":
			_ = json.NewEncoder(w).Encode(map[string]any{"items": []CatalogEntry{entry}})
		case "/registry/remote.app/1.0.0.json":
			_ = json.NewEncoder(w).Encode(catalogEnvelope{Manifest: manifest, SHA256: entry.SHA256, Payload: payload})
		case "/registry/remote.app/1.0.1.json":
			http.Redirect(w, r, "/registry/remote.app/1.0.0.json", http.StatusFound)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	remote, err := NewRemoteCatalog(server.URL+"/registry/", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	items, err := remote.List()
	if err != nil || len(items) != 1 || items[0].Manifest.AppID != manifest.AppID {
		t.Fatalf("remote catalog list: %+v %v", items, err)
	}
	got, err := remote.Load(manifest.AppID, manifest.PackageVersion)
	if err != nil || string(got.Payload) != string(payload) {
		t.Fatalf("remote catalog package: %q %v", got.Payload, err)
	}
	if _, err := remote.Load(manifest.AppID, "1.0.1"); err == nil || !strings.Contains(err.Error(), "redirect") {
		t.Fatalf("redirect was followed or reported without the safety reason: %v", err)
	}
	if _, err := NewRemoteCatalog("http://catalog.invalid/", nil); err == nil {
		t.Fatal("non-HTTPS catalog accepted")
	}
	if _, err := NewRemoteCatalog(server.URL+"/?token=secret", server.Client()); err == nil {
		t.Fatal("catalog URL with query accepted")
	}
}

func TestRemoteCatalogRejectsMalformedIndexBeforeInstall(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		payload := []CatalogEntry{{Manifest: Manifest{AppID: "remote.app", PackageVersion: "1.0.0"}, SHA256: strings.Repeat("0", 64), Size: -1}}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": payload})
	}))
	defer server.Close()
	remote, err := NewRemoteCatalog(server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := remote.List(); err == nil {
		t.Fatal("malformed remote index accepted")
	}
}

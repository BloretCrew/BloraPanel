package master

import (
	"blora.dev/panel/internal/extensions"
	"blora.dev/panel/internal/model"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestExtensionNotificationAuthorizationAndBounds(t *testing.T) {
	f := newFileFixture(t)
	f.app.extensions, _ = extensions.New(t.TempDir(), []string{"notification.publish"})
	payload := []byte("export function start(){}")
	h := sha256.Sum256(payload)
	manifest := extensions.Manifest{AppID: "example.notify", Title: "Notify", PackageVersion: "1.0.0", HostAPIVersion: 1, Capabilities: []string{"notification.publish"}}
	f.admin.request("POST", "/extensions/install-package", map[string]any{"manifest": manifest, "payload": payload, "sha256": hex.EncodeToString(h[:])}, model.ID(), 201)
	path := "/extensions/example.notify/notifications"
	input := map[string]string{"title": "Notice", "message": "Hello"}
	f.reader.request("POST", path, input, "", 403)
	f.admin.request("POST", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: model.ResourceRef{Kind: "extension", ID: "example.notify"}, Action: "app.use"}, model.ID(), 200)
	got := f.reader.request("POST", path, input, "", 200)
	if string(got["appId"]) != `"example.notify"` {
		t.Fatal("notification source not bound")
	}
	f.reader.request("POST", path, map[string]string{"title": strings.Repeat("a", 257)}, "", 400)
	f.reader.request("POST", path, map[string]string{"title": "Notice", "appId": "blora.admin"}, "", 400)
	f.admin.request("POST", "/extensions/example.notify/enabled", map[string]bool{"enabled": false}, model.ID(), 200)
	f.reader.request("POST", path, input, "", 403)
}

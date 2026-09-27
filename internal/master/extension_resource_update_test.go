package master

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"blora.dev/panel/internal/extensions"
	"blora.dev/panel/internal/model"
)

func TestExtensionMetadataWritePermissionRevisionAndReplay(t *testing.T) {
	f := newFileFixture(t)
	var err error
	f.app.extensions, err = extensions.New(t.TempDir(), []string{"resource.read", "resource.write"})
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte("export function start(){}")
	h := sha256.Sum256(payload)
	pkg := extensions.Package{Manifest: extensions.Manifest{AppID: "example.metadata", Title: "Metadata", PackageVersion: "1.0.0", HostAPIVersion: 1, Capabilities: []string{"resource.read", "resource.write"}}, Payload: payload, SHA256: hex.EncodeToString(h[:])}
	f.admin.request("POST", "/extensions/install-package", map[string]any{"manifest": pkg.Manifest, "sha256": pkg.SHA256, "payload": payload}, model.ID(), 201)
	url := "/extensions/example.metadata/resource"
	instance := f.instances[0]
	input := map[string]any{"resource": instanceRef(instance), "revision": instance.ConfigRevision, "name": "Renamed by extension"}
	f.reader.request("PATCH", url, input, model.ID(), 403)
	f.admin.request("POST", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: model.ResourceRef{Kind: "extension", ID: "example.metadata"}, Action: "app.use"}, model.ID(), 200)
	f.reader.request("PATCH", url, input, model.ID(), 403)
	f.admin.request("POST", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: instanceRef(instance), Action: "instance.configure"}, model.ID(), 200)
	f.reader.request("PATCH", url, map[string]any{"resource": instanceRef(instance), "name": "missing revision"}, model.ID(), 400)
	key := model.ID()
	got := f.reader.request("PATCH", url, input, key, 200)
	if string(got["name"]) != `"Renamed by extension"` || got["config"] != nil {
		t.Fatalf("incorrect metadata response %v", got)
	}
	replay := f.reader.request("PATCH", url, input, key, 200)
	if string(replay["configRevision"]) != string(got["configRevision"]) {
		t.Fatal("replay changed revision")
	}
	f.reader.request("PATCH", url, input, model.ID(), 409)
	input["config"] = map[string]any{"command": []string{"forbidden"}}
	f.reader.request("PATCH", url, input, model.ID(), 400)
	f.admin.request("POST", "/extensions/example.metadata/enabled", map[string]bool{"enabled": false}, model.ID(), 200)
	delete(input, "config")
	f.reader.request("PATCH", url, input, model.ID(), 403)
}

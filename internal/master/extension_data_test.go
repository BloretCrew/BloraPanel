package master

import (
	"encoding/json"
	"testing"

	"blora.dev/panel/internal/extensions"
	"blora.dev/panel/internal/model"
)

func TestExtensionDataAPIUsesAuthenticatedOwnerAndCurrentGrant(t *testing.T) {
	f := newFileFixture(t)
	var err error
	f.app.extensions, err = extensions.New(t.TempDir(), []string{"data.read", "data.write"})
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte("export function start(){}")
	p := extensions.Package{Manifest: extensions.Manifest{AppID: "data.example", PackageVersion: "1.0.0", HostAPIVersion: 1, Capabilities: []string{"data.read", "data.write"}}, Payload: payload, SHA256: extensions.ModuleHash(payload)}
	if _, err := f.app.extensions.Install(p); err != nil {
		t.Fatal(err)
	}
	path := "/extensions/data.example/data"
	f.reader.request("GET", path, nil, "", 403)
	grant := model.Grant{UserID: f.readerUser.ID, Resource: model.ResourceRef{Kind: "extension", ID: p.Manifest.AppID}, Action: "app.use"}
	f.admin.request("POST", "/grants", grant, model.ID(), 200)
	read := f.reader.request("GET", path, nil, "", 200)
	if string(read["revision"]) != "0" {
		t.Fatal("initial revision")
	}
	input := map[string]any{"expectedRevision": 0, "schemaVersion": 1, "data": map[string]string{"note": "reader private"}}
	key := model.ID()
	f.reader.request("PUT", path, input, key, 200)
	f.reader.request("PUT", path, input, key, 200)
	f.reader.request("PUT", path, input, model.ID(), 409)
	admin := f.admin.request("GET", path, nil, "", 200)
	if string(admin["data"]) != "{}" {
		t.Fatal("administrator implicitly accessed other user's data")
	}
	input["userId"] = "somebody-else"
	f.reader.request("PUT", path, input, model.ID(), 400)
	delete(input, "userId")
	f.admin.request("DELETE", "/grants", grant, model.ID(), 200)
	f.reader.request("GET", path, nil, "", 403)
	f.reader.request("PUT", path, input, key, 403)
	f.admin.request("POST", "/grants", grant, model.ID(), 200)
	read = f.reader.request("GET", path, nil, "", 200)
	var data map[string]string
	if err := json.Unmarshal(read["data"], &data); err != nil || data["note"] != "reader private" {
		t.Fatal("stored document changed")
	}
	f.admin.request("POST", "/extensions/data.example/enabled", map[string]bool{"enabled": false}, model.ID(), 200)
	f.reader.request("GET", path, nil, "", 403)
}

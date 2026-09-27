package extensions

import (
	"encoding/json"
	"os"
	"testing"
)

func TestBundleIntegrityAndVersionBinding(t *testing.T) {
	m, err := New(t.TempDir(), []string{"task.create"})
	if err != nil {
		t.Fatal(err)
	}
	wasm := []byte{0, 97, 115, 109, 1, 0, 0, 0}
	encoded, _ := json.Marshal(Bundle{Frontend: "export function start(){}", Backend: wasm})
	payload := append([]byte(BundlePrefix), encoded...)
	p := Package{Manifest: Manifest{AppID: "example.wasm", PackageVersion: "1.0.0", HostAPIVersion: 1, Capabilities: []string{"task.create"}}, Payload: payload, SHA256: ModuleHash(payload)}
	if _, err := m.Install(p); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Backend(p.Manifest.AppID, "wrong-hash"); err == nil {
		t.Fatal("package identity not enforced")
	}
	got, err := m.Backend(p.Manifest.AppID, p.SHA256)
	if err != nil || string(got) != string(wasm) {
		t.Fatalf("backend: %v", err)
	}
	got[0] = 99
	fresh, err := m.Backend(p.Manifest.AppID, p.SHA256)
	if err != nil || fresh[0] != 0 {
		t.Fatal("caller changed cached module")
	}
	if err := os.WriteFile(m.payloadPath(p.Manifest.AppID), []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Backend(p.Manifest.AppID, p.SHA256); err == nil {
		t.Fatal("cache bypassed package integrity")
	}
	if err := os.WriteFile(m.payloadPath(p.Manifest.AppID), p.Payload, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := m.SetEnabled(p.Manifest.AppID, false); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Backend(p.Manifest.AppID, p.SHA256); err == nil {
		t.Fatal("disabled module returned")
	}
	for _, bad := range []string{BundlePrefix + `{"frontend":"ok","backend":"bm90IHdhc20="}`, BundlePrefix + `{"frontend":"ok","surprise":1}`, BundlePrefix + `{"frontend":"ok"} {}`} {
		if _, err := DecodeBundle([]byte(bad)); err == nil {
			t.Fatalf("invalid bundle accepted: %q", bad)
		}
	}
}

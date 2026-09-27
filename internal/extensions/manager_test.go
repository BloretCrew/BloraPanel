package extensions

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func pkg(id, version string, payload []byte) Package {
	h := sha256.Sum256(payload)
	return Package{Manifest: Manifest{AppID: id, PackageVersion: version, HostAPIVersion: 1, Title: id, Capabilities: []string{"window.open"}}, SHA256: hex.EncodeToString(h[:]), Payload: payload}
}

func TestLifecycleIntegrityAndCapabilityPolicy(t *testing.T) {
	m, err := New(t.TempDir(), []string{"window.open"})
	if err != nil {
		t.Fatal(err)
	}
	p := pkg("example.app", "1.0.0", []byte("v1"))
	x, err := m.Install(p)
	if err != nil || !x.Enabled {
		t.Fatalf("install: %+v %v", x, err)
	}
	if got, err := m.Payload("example.app"); err != nil || string(got) != "v1" {
		t.Fatalf("payload persistence: %q %v", got, err)
	}
	if err := os.WriteFile(filepath.Join(m.root, "example.app.pkg"), []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Get("example.app"); err == nil {
		t.Fatal("tampered installed payload was exposed")
	}
	if _, err := m.Payload("example.app"); err == nil {
		t.Fatal("tampered installed payload was served")
	}
	if err := os.WriteFile(filepath.Join(m.root, "example.app.pkg"), p.Payload, 0600); err != nil {
		t.Fatal(err)
	}
	dataDir, err := m.UserDataDir("example.app")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(dataDir+"/settings.json", []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = m.Install(p); err == nil {
		t.Fatal("duplicate install accepted")
	}
	p2 := pkg("example.app", "2.0.0", []byte("v2"))
	x, err = m.Upgrade(p2)
	if err != nil || x.Manifest.PackageVersion != "2.0.0" {
		t.Fatalf("upgrade: %+v %v", x, err)
	}
	if installed, err := m.List(); err != nil || len(installed) != 1 || installed[0].Manifest.PackageVersion != "2.0.0" {
		t.Fatalf("rollback metadata leaked into installed list: %+v %v", installed, err)
	}
	if x, err = m.Rollback("example.app"); err != nil || x.Manifest.PackageVersion != "1.0.0" {
		t.Fatalf("rollback: %+v %v", x, err)
	}
	if got, err := m.Payload("example.app"); err != nil || string(got) != "v1" {
		t.Fatalf("rollback payload: %q %v", got, err)
	}
	// Reapply v2 so downgrade checks continue to exercise the upgraded state.
	if _, err = m.Upgrade(p2); err != nil {
		t.Fatalf("re-upgrade: %v", err)
	}
	if _, err := m.Upgrade(pkg("example.app", "1.5.0", []byte("old"))); err == nil {
		t.Fatal("downgrade accepted")
	}
	if _, err = m.SetEnabled("example.app", false); err != nil {
		t.Fatal(err)
	}
	if err = m.AuthorizeCapability("example.app", "window.open"); err == nil {
		t.Fatal("disabled extension accepted a new capability request")
	}
	if x, err = m.Get("example.app"); err != nil || x.Enabled {
		t.Fatalf("disable not persisted: %+v %v", x, err)
	}
	bad := p
	bad.SHA256 = "00"
	if _, err = m.Upgrade(bad); err == nil {
		t.Fatal("tampered package accepted")
	}
	forbidden := pkg("other.app", "1.0.0", []byte("x"))
	forbidden.Manifest.Capabilities = []string{"host.exec"}
	if _, err = m.Install(forbidden); err == nil {
		t.Fatal("forbidden capability accepted")
	}
	if err = m.Uninstall("example.app"); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(dataDir + "/settings.json"); err != nil {
		t.Fatalf("user data not retained: %v", err)
	}
	if err = m.CleanupUserData("example.app"); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(dataDir); !os.IsNotExist(err) {
		t.Fatalf("user data not cleaned: %v", err)
	}
	if _, err = m.Get("example.app"); err == nil {
		t.Fatal("uninstall left extension")
	}
}

func TestTrustedPackageSignature(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	m, err := New(t.TempDir(), []string{"window.open"})
	if err != nil {
		t.Fatal(err)
	}
	m.SetTrustedKeys([]ed25519.PublicKey{pub})
	p := pkg("signed.app", "1.0.0", []byte("payload"))
	p.Signature = ed25519.Sign(priv, SignatureMessage(p))
	if _, err = m.Install(p); err != nil {
		t.Fatal(err)
	}
	bad := pkg("bad.app", "1.0.0", []byte("payload"))
	if _, err = m.Install(bad); err == nil {
		t.Fatal("unsigned package accepted")
	}
	if _, err = m.Get(p.Manifest.AppID); err != nil {
		t.Fatalf("normalized installed signature: %v", err)
	}
	for name, change := range map[string]func(*Package){
		"identity":    func(p *Package) { p.Manifest.AppID = "other.app" },
		"version":     func(p *Package) { p.Manifest.PackageVersion = "2.0.0" },
		"capability":  func(p *Package) { p.Manifest.Capabilities = nil },
		"permissions": func(p *Package) { p.Manifest.Permissions = []string{"host.manage"} },
		"state":       func(p *Package) { p.Manifest.StateSchemaVersion = 2 },
		"data state":  func(p *Package) { p.Manifest.DataSchemaVersion = 2 },
		"entry":       func(p *Package) { p.Manifest.Entrypoints = []string{"other"} },
		"resources":   func(p *Package) { p.Manifest.ResourceHandlers = []string{"node"} },
		"dependency":  func(p *Package) { p.Manifest.Dependencies = map[string]string{"other.app": "1.0.0"} },
		"payload": func(p *Package) {
			p.Payload = []byte("changed")
			h := sha256.Sum256(p.Payload)
			p.SHA256 = hex.EncodeToString(h[:])
		},
	} {
		t.Run(name, func(t *testing.T) {
			altered := p
			change(&altered)
			if err := m.Validate(altered); err == nil {
				t.Fatal("modified signed package accepted")
			}
		})
	}
	legacy := p
	h := sha256.Sum256(legacy.Payload)
	legacy.Signature = ed25519.Sign(priv, h[:])
	if err := m.Validate(legacy); err == nil {
		t.Fatal("legacy unbound signature accepted")
	}
}

func TestVersionFormatRequired(t *testing.T) {
	m, _ := New(t.TempDir(), []string{"window.open"})
	p := pkg("bad.version", "latest", []byte("x"))
	if _, err := m.Install(p); err == nil {
		t.Fatal("non-semver version accepted")
	}
}

func TestExternalManifestReservedIDsAndDependencies(t *testing.T) {
	m, err := New(t.TempDir(), []string{"window.open"})
	if err != nil {
		t.Fatal(err)
	}
	reserved := pkg("blora.instances", "1.0.0", []byte("x"))
	if _, err := m.Install(reserved); err == nil {
		t.Fatal("reserved built-in app id accepted")
	}
	dependent := pkg("example.dependent", "1.0.0", []byte("x"))
	dependent.Manifest.Dependencies = map[string]string{"example.base": "1.0.0"}
	if _, err := m.Install(dependent); err == nil {
		t.Fatal("missing dependency accepted")
	}
	base := pkg("example.base", "1.0.0", []byte("base"))
	if _, err := m.Install(base); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Install(dependent); err != nil {
		t.Fatalf("satisfied dependency rejected: %v", err)
	}
	badVersion := pkg("example.baddep", "1.0.0", []byte("x"))
	badVersion.Manifest.Dependencies = map[string]string{"example.base": "2.0.0"}
	if _, err := m.Install(badVersion); err == nil {
		t.Fatal("incompatible dependency accepted")
	}
}

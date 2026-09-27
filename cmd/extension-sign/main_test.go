package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"blora.dev/panel/internal/extensions"
)

func TestSignedPackageInstallsAndRejectsMetadataTampering(t *testing.T) {
	root := t.TempDir()
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	key, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(root, "key.pem")
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key}), 0600); err != nil {
		t.Fatal(err)
	}
	payload := []byte("export function start() {}")
	h := sha256.Sum256(payload)
	pkg := map[string]any{"manifest": extensions.Manifest{AppID: "signed.example", PackageVersion: "1.0.0", HostAPIVersion: 1, Title: "Signed", Capabilities: []string{"window.open"}}, "payload": payload, "sha256": hex.EncodeToString(h[:])}
	input := filepath.Join(root, "input.json")
	data, _ := json.Marshal(pkg)
	if err := os.WriteFile(input, data, 0600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(root, "signed.json")
	if err := run(input, keyPath, output); err != nil {
		t.Fatal(err)
	}
	if err := run(input, keyPath, output); err == nil {
		t.Fatal("overwrote existing file")
	}
	data, err = os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	var signed struct {
		Manifest  extensions.Manifest
		Payload   []byte
		SHA256    string
		Signature []byte
	}
	if err := json.Unmarshal(data, &signed); err != nil {
		t.Fatal(err)
	}
	m, err := extensions.New(filepath.Join(root, "registry"), []string{"window.open"})
	if err != nil {
		t.Fatal(err)
	}
	m.SetTrustedKeys([]ed25519.PublicKey{pub})
	p := extensions.Package{Manifest: signed.Manifest, Payload: signed.Payload, SHA256: signed.SHA256, Signature: signed.Signature}
	if _, err := m.Install(p); err != nil {
		t.Fatal(err)
	}
	reopened, err := extensions.New(filepath.Join(root, "registry"), []string{"window.open"})
	if err != nil {
		t.Fatal(err)
	}
	reopened.SetTrustedKeys([]ed25519.PublicKey{pub})
	if _, err := reopened.Get(p.Manifest.AppID); err != nil {
		t.Fatal(err)
	}
	metadataPath := filepath.Join(root, "registry", p.Manifest.AppID+".json")
	data, err = os.ReadFile(metadataPath)
	if err != nil {
		t.Fatal(err)
	}
	var installed extensions.Installed
	if err := json.Unmarshal(data, &installed); err != nil {
		t.Fatal(err)
	}
	installed.Manifest.Title = "Changed without publisher signature"
	data, _ = json.Marshal(installed)
	if err := os.WriteFile(metadataPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := reopened.Get(p.Manifest.AppID); err == nil {
		t.Fatal("tampered installed metadata accepted")
	}
}

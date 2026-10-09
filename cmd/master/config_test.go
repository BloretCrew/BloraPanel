package main

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func configFile(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "master.json")
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestConfigDefaultsAndPaths(t *testing.T) {
	path := configFile(t, `{"passwordFile":"initial-password","extensionsDir":"state/extensions","trustedProxies":["127.0.0.1/32","::1/128"]}`)
	c, initialize, err := parseConfigAt([]string{"--config", path, "--init"}, io.Discard, filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if !initialize || c.Listen != "127.0.0.1:37861" || c.AdminName != "admin" || c.Origin != "" {
		t.Fatalf("incorrect defaults: %+v", c)
	}
	base := filepath.Dir(path)
	if c.StateDir != filepath.Join(base, "state/master") || c.StaticDir != filepath.Join(base, "web/dist") ||
		c.PasswordFile != filepath.Join(base, "initial-password") || c.ExtensionsDir != filepath.Join(base, "state/extensions") {
		t.Fatalf("paths must follow executable directory: %+v", c)
	}
	if c.TLSCert != "" || c.TLSKey != "" || len(c.TrustedProxies) != 2 {
		t.Fatalf("incorrect TLS/proxy settings: %+v", c)
	}
}

func TestConfigExplicitOverrides(t *testing.T) {
	path := configFile(t, `{"stateDir":"private","listen":"127.0.0.1:45678","staticDir":"","origin":"https://localhost:45678","trustedProxies":["127.0.0.1/32"],"adminName":"owner"}`)
	c, _, err := parseConfigAt([]string{"--config", path}, io.Discard, filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if c.Listen != "127.0.0.1:45678" || c.StaticDir != "" || c.AdminName != "owner" {
		t.Fatalf("implicit CLI defaults overwrote file: %+v", c)
	}
	c, _, err = parseConfigAt([]string{"--config", path, "--listen", defaultListen, "--state-dir", "legacy-state", "--origin", "", "--trusted-proxies", ""}, io.Discard, filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if c.Listen != defaultListen || c.StateDir != "legacy-state" || c.Origin != "" || strings.Join(c.TrustedProxies, ",") != "" {
		t.Fatalf("explicit legacy overrides not respected: %+v", c)
	}
}

func TestConfigRejectsInvalidBeforeState(t *testing.T) {
	for name, body := range map[string]string{
		"unknown":            `{"password":"do-not-put-credentials-here"}`,
		"malformed":          `{"listen":`,
		"null":               "null",
		"array":              "[]",
		"trailing":           "{} {}",
		"wrong-type":         `{"trustedProxies":"127.0.0.1/32"}`,
		"empty-state":        `{"stateDir":""}`,
		"empty-listen":       `{"listen":""}`,
		"zero-port":          `{"listen":"127.0.0.1:0"}`,
		"large-port":         `{"listen":"127.0.0.1:65536"}`,
		"bad-proxy":          `{"trustedProxies":["not-an-address"]}`,
		"unpaired-tls":       `{"tlsCert":"cert.pem"}`,
		"strict-without-ips": `{"verifyProxyIPs":true}`,
		"strict-empty-ip":    `{"verifyProxyIPs":true,"trustedProxies":[""]}`,
		"invalid-key":        `{"extensionsPublicKey":"abcd"}`,
		"oversize":           `{"staticDir":"` + strings.Repeat("a", 1<<20) + `"}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := configFile(t, body)
			_, _, err := parseConfigAt([]string{"--config", path}, io.Discard, filepath.Dir(path))
			if err == nil {
				t.Fatal("invalid configuration accepted")
			}
			if strings.Contains(err.Error(), "do-not-put-credentials-here") {
				t.Fatal("credential contents leaked in error")
			}
			if _, err := os.Stat(filepath.Join(filepath.Dir(path), "state")); !os.IsNotExist(err) {
				t.Fatal("configuration parsing created state")
			}
		})
	}
}

func TestConfigMissingExplicitFileAndHelp(t *testing.T) {
	_, _, err := parseConfigAt([]string{"--config", filepath.Join(t.TempDir(), "missing.json")}, io.Discard, t.TempDir())
	if err == nil {
		t.Fatal("explicit missing configuration accepted")
	}
	_, _, err = parseConfigAt([]string{"--help"}, io.Discard, t.TempDir())
	if err != flag.ErrHelp {
		t.Fatalf("help: %v", err)
	}
	_, _, err = parseConfigAt([]string{"unexpected"}, io.Discard, t.TempDir())
	if err == nil {
		t.Fatal("positional argument accepted")
	}
}

func TestConfigAutomaticFileAndLegacyWithoutFile(t *testing.T) {
	dir := t.TempDir()
	c, _, err := parseConfigAt([]string{"--state-dir", "fixture-state"}, io.Discard, dir)
	if err != nil || c.StateDir != "fixture-state" || c.Listen != defaultListen || !c.HTTPS || !c.legacy {
		t.Fatalf("legacy command without file: %+v, %v", c, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "master.json"), []byte(`{"listen":"[::1]:37862","adminName":"owner"}`), 0600); err != nil {
		t.Fatal(err)
	}
	c, _, err = parseConfigAt(nil, io.Discard, dir)
	if err != nil || c.Listen != "[::1]:37862" || c.AdminName != "owner" || c.HTTPS || c.VerifyProxyIPs ||
		c.StateDir != filepath.Join(dir, "state/master") || c.PasswordFile != filepath.Join(dir, "initial-password") {
		t.Fatalf("sibling master.json defaults: %+v, %v", c, err)
	}
}

func TestConfigExternalFileStillUsesExecutablePaths(t *testing.T) {
	base := t.TempDir()
	path := configFile(t, `{"stateDir":"data","passwordFile":"password","staticDir":"site","verifyProxyIPs":true,"trustedProxies":["127.0.0.1"]}`)
	c, _, err := parseConfigAt([]string{"--config", path}, io.Discard, base)
	if err != nil || c.StateDir != filepath.Join(base, "data") || c.StaticDir != filepath.Join(base, "site") ||
		c.PasswordFile != filepath.Join(base, "password") || !c.VerifyProxyIPs {
		t.Fatalf("paths must use executable directory even with an external configuration: %+v, %v", c, err)
	}
}

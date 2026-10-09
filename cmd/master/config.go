package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"blora.dev/panel/internal/bootstrap"
	"blora.dev/panel/internal/master"
)

const defaultListen = "127.0.0.1:37861"

type masterConfig struct {
	StateDir             string   `json:"stateDir"`
	Listen               string   `json:"listen"`
	Origin               string   `json:"origin"`
	TrustedProxies       []string `json:"trustedProxies"`
	StaticDir            string   `json:"staticDir"`
	ExtensionsDir        string   `json:"extensionsDir"`
	ExtensionsCatalog    string   `json:"extensionsCatalog"`
	ExtensionsCatalogURL string   `json:"extensionsCatalogUrl"`
	ExtensionsPublicKey  string   `json:"extensionsPublicKey"`
	AdminName            string   `json:"adminName"`
	PasswordFile         string   `json:"passwordFile"`
	TLSCert              string   `json:"tlsCert"`
	TLSKey               string   `json:"tlsKey"`
	HTTPS                bool     `json:"https"`
	VerifyProxyIPs       bool     `json:"verifyProxyIPs"`
	legacy               bool
}

func defaultConfig() masterConfig {
	return masterConfig{StateDir: "state/master", Listen: defaultListen, StaticDir: "web/dist", AdminName: "admin", PasswordFile: "initial-password"}
}

// Configuration and data paths default to the executable directory. Legacy CLI
// paths remain relative to cwd so existing test/deployment scripts keep working.
func parseConfig(args []string, output io.Writer) (masterConfig, bool, error) {
	base, err := bootstrap.ExecutableDir()
	if err != nil {
		return masterConfig{}, false, err
	}
	return parseConfigAt(args, output, base)
}

func parseConfigAt(args []string, output io.Writer, base string) (masterConfig, bool, error) {
	overrides := defaultConfig()
	flags := flag.NewFlagSet("blora-master", flag.ContinueOnError)
	flags.SetOutput(output)
	configPath := flags.String("config", filepath.Join(base, "master.json"), "JSON configuration file; defaults beside the executable")
	initialize := flags.Bool("init", false, "initialize administrator (and optional direct TLS), then exit")
	fields := map[string]*string{
		"state-dir": &overrides.StateDir, "listen": &overrides.Listen,
		"origin": &overrides.Origin, "static-dir": &overrides.StaticDir,
		"extensions-dir": &overrides.ExtensionsDir, "extensions-catalog": &overrides.ExtensionsCatalog,
		"extensions-catalog-url": &overrides.ExtensionsCatalogURL, "extensions-public-key": &overrides.ExtensionsPublicKey,
		"admin-name": &overrides.AdminName, "password-file": &overrides.PasswordFile,
		"tls-cert": &overrides.TLSCert, "tls-key": &overrides.TLSKey,
	}
	for name, field := range fields {
		flags.StringVar(field, name, *field, "legacy override; prefer the JSON configuration")
	}
	proxyList := flags.String("trusted-proxies", "", "legacy comma-separated override; prefer trustedProxies in JSON")
	if err := flags.Parse(args); err != nil {
		return masterConfig{}, false, err
	}
	if flags.NArg() != 0 {
		return masterConfig{}, false, fmt.Errorf("unexpected positional arguments")
	}
	explicit := map[string]bool{}
	flags.Visit(func(f *flag.Flag) { explicit[f.Name] = true })
	legacy := false
	if !explicit["config"] {
		for name := range explicit {
			if name != "init" {
				legacy = true
			}
		}
	}
	config := defaultConfig()
	// Legacy flag-based automation does not inherit a user's sibling config.
	var f *os.File
	var err error
	if !legacy {
		f, err = os.Open(bootstrap.ResolvePath(base, *configPath))
	} else {
		config.HTTPS, config.legacy = true, true
	}
	if err != nil {
		return masterConfig{}, false, fmt.Errorf("open Master configuration: %w", err)
	} else if f != nil {
		b, readErr := io.ReadAll(io.LimitReader(f, (1<<20)+1))
		_ = f.Close()
		if readErr != nil {
			return masterConfig{}, false, fmt.Errorf("read Master configuration: %w", readErr)
		}
		if len(b) > 1<<20 {
			return masterConfig{}, false, fmt.Errorf("Master configuration exceeds 1 MiB")
		}
		b = bytes.TrimSpace(b)
		if len(b) == 0 || b[0] != '{' {
			return masterConfig{}, false, fmt.Errorf("Master configuration must be a JSON object")
		}
		decoder := json.NewDecoder(bytes.NewReader(b))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&config); err != nil {
			return masterConfig{}, false, fmt.Errorf("decode Master configuration: %w", err)
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			return masterConfig{}, false, fmt.Errorf("Master configuration must contain exactly one JSON object")
		}
		for _, path := range []*string{&config.StateDir, &config.StaticDir, &config.ExtensionsDir, &config.ExtensionsCatalog, &config.PasswordFile, &config.TLSCert, &config.TLSKey} {
			*path = bootstrap.ResolvePath(base, *path)
		}
	}
	targets := map[string]*string{
		"state-dir": &config.StateDir, "listen": &config.Listen,
		"origin": &config.Origin, "static-dir": &config.StaticDir,
		"extensions-dir": &config.ExtensionsDir, "extensions-catalog": &config.ExtensionsCatalog,
		"extensions-catalog-url": &config.ExtensionsCatalogURL, "extensions-public-key": &config.ExtensionsPublicKey,
		"admin-name": &config.AdminName, "password-file": &config.PasswordFile,
		"tls-cert": &config.TLSCert, "tls-key": &config.TLSKey,
	}
	for name, target := range targets {
		if explicit[name] {
			*target = *fields[name]
		}
	}
	if explicit["trusted-proxies"] {
		config.TrustedProxies = strings.Split(*proxyList, ",")
		config.legacy = true
	}
	if err := config.validate(); err != nil {
		return masterConfig{}, false, err
	}
	return config, *initialize, nil
}

func (c masterConfig) validate() error {
	if strings.TrimSpace(c.StateDir) == "" {
		return fmt.Errorf("stateDir must not be empty")
	}
	_, port, err := net.SplitHostPort(c.Listen)
	n, portErr := strconv.Atoi(port)
	if err != nil || portErr != nil || n < 1 || n > 65535 {
		return fmt.Errorf("listen must be a host:port address with port 1–65535")
	}
	proxies, err := master.ParseTrustedProxies(strings.Join(c.TrustedProxies, ","))
	if err != nil {
		return fmt.Errorf("trustedProxies: %w", err)
	}
	if c.VerifyProxyIPs && len(proxies) == 0 {
		return fmt.Errorf("trustedProxies must list IPs/CIDRs when verifyProxyIPs is enabled")
	}
	if (c.TLSCert == "") != (c.TLSKey == "") {
		return fmt.Errorf("tlsCert and tlsKey must be configured together")
	}
	if !c.HTTPS && c.TLSCert != "" {
		return fmt.Errorf("tlsCert/tlsKey require https to be enabled")
	}
	if c.ExtensionsPublicKey != "" {
		b, err := hex.DecodeString(c.ExtensionsPublicKey)
		if err != nil || len(b) != 32 {
			return fmt.Errorf("extensionsPublicKey must be 32-byte hex")
		}
	}
	return nil
}

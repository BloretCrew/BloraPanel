// Package extensions implements the local, administrator-controlled extension registry.
package extensions

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var appIDPattern = regexp.MustCompile(`^[-a-z0-9]+\.[-a-z0-9.]+$`)
var versionPattern = regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+$`)

type Manifest struct {
	AppID              string            `json:"appId"`
	PackageVersion     string            `json:"packageVersion"`
	HostAPIVersion     int               `json:"hostApiVersion"`
	Title              string            `json:"title"`
	Icon               string            `json:"icon,omitempty"`
	Color              string            `json:"color,omitempty"`
	Permissions        []string          `json:"permissions,omitempty"`
	Protected          bool              `json:"protected,omitempty"`
	Capabilities       []string          `json:"capabilities"`
	Entrypoints        []string          `json:"entrypoints,omitempty"`
	ResourceHandlers   []string          `json:"resourceHandlers,omitempty"`
	Dependencies       map[string]string `json:"dependencies,omitempty"`
	WindowPolicy       string            `json:"windowPolicy,omitempty"`
	TabPolicy          *TabPolicy        `json:"tabPolicy,omitempty"`
	StateSchemaVersion int               `json:"stateSchemaVersion,omitempty"`
	DataSchemaVersion  int               `json:"dataSchemaVersion,omitempty"`
}

type TabPolicy struct {
	Types   []string `json:"types"`
	Movable bool     `json:"movable"`
}

func normalizeManifest(m Manifest) Manifest {
	if m.Icon == "" {
		m.Icon = "✦"
	}
	if m.Color == "" {
		m.Color = "#8ec5ff"
	}
	if m.WindowPolicy == "" {
		m.WindowPolicy = "multiple"
	}
	if m.TabPolicy == nil {
		m.TabPolicy = &TabPolicy{Movable: true}
	}
	return m
}

type Package struct {
	Manifest  Manifest `json:"manifest"`
	SHA256    string   `json:"sha256"`
	Payload   []byte   `json:"-"`
	Signature []byte   `json:"signature,omitempty"`
}

type Installed struct {
	Manifest Manifest `json:"manifest"`
	SHA256   string   `json:"sha256"`
	Enabled  bool     `json:"enabled"`
	// Keep the authenticity proof with the installed metadata so a restart
	// enforces the same trusted-key policy as the original installation.
	Signature []byte `json:"signature,omitempty"`
}

type Manager struct {
	mu              sync.RWMutex
	migrationActive bool
	root            string
	allow           map[string]bool
	trusted         []ed25519.PublicKey
	backendMu       sync.Mutex
	backendHash     string
	backendModule   []byte
}

// SetTrustedKeys makes signatures mandatory for every package accepted by this registry.
func (m *Manager) SetTrustedKeys(keys []ed25519.PublicKey) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.trusted = make([]ed25519.PublicKey, len(keys))
	for i, key := range keys {
		m.trusted[i] = append(ed25519.PublicKey(nil), key...)
	}
}

func New(root string, allowedCapabilities []string) (*Manager, error) {
	if root == "" {
		return nil, errors.New("extension root required")
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	a := map[string]bool{}
	for _, c := range allowedCapabilities {
		a[c] = true
	}
	m := &Manager{root: root, allow: a}
	if err := m.recoverTransactions(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Manager) Validate(p Package) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.validate(p)
}

func (m *Manager) validate(p Package) error {
	if p.Manifest.Protected {
		return errors.New("external extension cannot be protected")
	}
	if p.Manifest.DataSchemaVersion < 0 || p.Manifest.DataSchemaVersion > 1_000_000 {
		return errors.New("invalid data schema version")
	}
	if !appIDPattern.MatchString(p.Manifest.AppID) || len(p.Manifest.AppID) > 128 {
		return errors.New("invalid app id")
	}
	if strings.HasPrefix(p.Manifest.AppID, "blora.") {
		return errors.New("external extension app id is reserved")
	}
	if strings.HasSuffix(p.Manifest.AppID, ".previous") {
		return errors.New("extension app id uses a reserved suffix")
	}
	if !versionPattern.MatchString(p.Manifest.PackageVersion) || len(p.Manifest.PackageVersion) > 64 {
		return errors.New("invalid package version")
	}
	if p.Manifest.HostAPIVersion != 1 {
		return errors.New("unsupported host api version")
	}
	if len(p.Manifest.Entrypoints) > 64 || len(p.Manifest.ResourceHandlers) > 64 {
		return errors.New("extension manifest too large")
	}
	if p.Manifest.WindowPolicy != "" && p.Manifest.WindowPolicy != "multiple" {
		return errors.New("unsupported window policy")
	}
	if p.Manifest.TabPolicy != nil && !p.Manifest.TabPolicy.Movable {
		return errors.New("non-movable tabs are not supported")
	}
	for _, c := range p.Manifest.Capabilities {
		if !m.allow[c] {
			return fmt.Errorf("capability %q is not allowed", c)
		}
	}
	for _, permission := range p.Manifest.Permissions {
		if permission == "" || len(permission) > 128 {
			return errors.New("invalid extension permission")
		}
	}
	if len(p.Manifest.Dependencies) > 64 {
		return errors.New("extension dependency list too large")
	}
	for dependency, required := range p.Manifest.Dependencies {
		if dependency == p.Manifest.AppID || !appIDPattern.MatchString(dependency) || len(dependency) > 128 {
			return errors.New("invalid extension dependency id")
		}
		if strings.HasPrefix(dependency, "blora.") || !versionPattern.MatchString(required) || len(required) > 64 {
			return errors.New("invalid extension dependency version")
		}
	}
	h := sha256.Sum256(p.Payload)
	if !strings.EqualFold(hex.EncodeToString(h[:]), p.SHA256) {
		return errors.New("package integrity check failed")
	}
	if len(m.trusted) > 0 {
		if len(p.Signature) != ed25519.SignatureSize {
			return errors.New("package signature required")
		}
		valid := false
		message := SignatureMessage(p)
		for _, key := range m.trusted {
			if len(key) == ed25519.PublicKeySize && ed25519.Verify(key, message, p.Signature) {
				valid = true
				break
			}
		}
		if !valid {
			return errors.New("package signature rejected")
		}
	}
	return nil
}

func (m *Manager) path(id string) string         { return filepath.Join(m.root, id+".json") }
func (m *Manager) payloadPath(id string) string  { return filepath.Join(m.root, id+".pkg") }
func (m *Manager) previousPath(id string) string { return filepath.Join(m.root, id+".previous.json") }
func (m *Manager) previousPayloadPath(id string) string {
	return filepath.Join(m.root, id+".previous.pkg")
}
func (m *Manager) dataPath(id string) string { return filepath.Join(m.root, id+".data") }
func (m *Manager) read(id string) (Installed, error) {
	var x Installed
	if !appIDPattern.MatchString(id) || len(id) > 128 {
		return x, errors.New("invalid app id")
	}
	if _, err := os.Stat(m.transactionPath(id)); err == nil {
		return x, errors.New("extension transaction requires recovery")
	} else if !errors.Is(err, os.ErrNotExist) {
		return x, err
	}
	b, e := os.ReadFile(m.path(id))
	if e != nil {
		return x, e
	}
	e = json.Unmarshal(b, &x)
	if e == nil && x.Manifest.AppID != id {
		return Installed{}, errors.New("installed manifest identity mismatch")
	}
	return x, e
}

// readVerified is used whenever an installed package is exposed or granted a
// capability. A private registry directory can still be damaged or restored
// partially; never execute payload bytes without rechecking their digest and
// the configured manifest/signature policy.
func (m *Manager) readVerified(id string) (Installed, []byte, error) {
	x, err := m.read(id)
	if err != nil {
		return x, nil, err
	}
	payload, err := os.ReadFile(m.payloadPath(id))
	if err != nil {
		return x, nil, err
	}
	if err := m.validate(Package{Manifest: x.Manifest, SHA256: x.SHA256, Payload: payload, Signature: x.Signature}); err != nil {
		return x, nil, err
	}
	return x, payload, nil
}
func (m *Manager) write(id string, x Installed) error {
	b, e := json.MarshalIndent(x, "", "  ")
	if e != nil {
		return e
	}
	f, err := os.CreateTemp(m.root, ".extension-")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if err = f.Chmod(0600); err != nil {
		f.Close()
		return err
	}
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err := replaceRegistryFile(name, m.path(id)); err != nil {
		return err
	}
	return syncDir(m.root)
}
func (m *Manager) writePayload(id string, payload []byte) error {
	tmp, err := os.CreateTemp(m.root, ".package-")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err = tmp.Chmod(0600); err == nil {
		_, err = tmp.Write(payload)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if e := tmp.Close(); err == nil {
		err = e
	}
	if err != nil {
		return err
	}
	if err := replaceRegistryFile(name, m.payloadPath(id)); err != nil {
		return err
	}
	return syncDir(m.root)
}

func syncDir(path string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	d, err := os.Open(path)
	if err != nil {
		return err
	}
	err = d.Sync()
	closeErr := d.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func (m *Manager) Install(p Package) (Installed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, err := DecodeBundle(p.Payload); err != nil {
		return Installed{}, err
	}
	p.Manifest = normalizeManifest(p.Manifest)
	if err := m.validate(p); err != nil {
		return Installed{}, err
	}
	if _, err := m.read(p.Manifest.AppID); err == nil {
		return Installed{}, errors.New("extension already installed")
	} else if !errors.Is(err, os.ErrNotExist) {
		return Installed{}, err
	}
	if err := m.checkDependencies(p.Manifest); err != nil {
		return Installed{}, err
	}
	x := Installed{Manifest: p.Manifest, SHA256: p.SHA256, Enabled: true, Signature: append([]byte(nil), p.Signature...)}
	if err := os.MkdirAll(m.dataPath(p.Manifest.AppID), 0700); err != nil {
		return Installed{}, err
	}
	migratedData, err := m.prepareDataMigration(p.Manifest.AppID, p.Manifest, p.Payload, p.Signature)
	if err != nil {
		return Installed{}, err
	}
	if err := m.transact(p.Manifest.AppID, func() error {
		if migratedData != nil {
			if err := atomicRegistryFile(m.dataPath(p.Manifest.AppID), m.dataFile(p.Manifest.AppID), migratedData); err != nil {
				return err
			}
		}
		if err := m.writePayload(p.Manifest.AppID, p.Payload); err != nil {
			return err
		}
		if err := m.write(p.Manifest.AppID, x); err != nil {
			return err
		}
		for _, path := range []string{m.previousPath(p.Manifest.AppID), m.previousPayloadPath(p.Manifest.AppID)} {
			if err := m.removeRegistryFile(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		return nil
	}); err != nil {
		return Installed{}, err
	}
	return x, nil
}
func (m *Manager) Upgrade(p Package) (Installed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, err := DecodeBundle(p.Payload); err != nil {
		return Installed{}, err
	}
	p.Manifest = normalizeManifest(p.Manifest)
	if err := m.validate(p); err != nil {
		return Installed{}, err
	}
	old, oldPayload, err := m.readVerified(p.Manifest.AppID)
	if err != nil {
		return Installed{}, err
	}
	if old.Manifest.PackageVersion == p.Manifest.PackageVersion {
		return Installed{}, errors.New("same extension version")
	}
	if compareVersion(p.Manifest.PackageVersion, old.Manifest.PackageVersion) <= 0 {
		return Installed{}, errors.New("extension downgrade rejected")
	}
	if err := m.checkDependencies(p.Manifest); err != nil {
		return Installed{}, err
	}
	x := Installed{Manifest: p.Manifest, SHA256: p.SHA256, Enabled: old.Enabled, Signature: append([]byte(nil), p.Signature...)}
	oldMeta, err := json.Marshal(old)
	if err != nil {
		return Installed{}, err
	}
	migratedData, err := m.prepareDataMigration(p.Manifest.AppID, p.Manifest, p.Payload, p.Signature)
	if err != nil {
		return Installed{}, err
	}
	if err := m.transact(p.Manifest.AppID, func() error {
		if migratedData != nil {
			if err := atomicRegistryFile(m.dataPath(p.Manifest.AppID), m.dataFile(p.Manifest.AppID), migratedData); err != nil {
				return err
			}
		}
		if err := atomicRegistryFile(m.root, m.previousPath(p.Manifest.AppID), oldMeta); err != nil {
			return err
		}
		if err := atomicRegistryFile(m.root, m.previousPayloadPath(p.Manifest.AppID), oldPayload); err != nil {
			return err
		}
		if err := m.writePayload(p.Manifest.AppID, p.Payload); err != nil {
			return err
		}
		return m.write(p.Manifest.AppID, x)
	}); err != nil {
		return Installed{}, err
	}
	return x, nil
}

func (m *Manager) checkDependencies(manifest Manifest) error {
	for dependency, required := range manifest.Dependencies {
		installed, _, err := m.readVerified(dependency)
		if err != nil {
			return fmt.Errorf("extension dependency %s is not installed or verified: %w", dependency, err)
		}
		if compareVersion(installed.Manifest.PackageVersion, required) != 0 {
			return fmt.Errorf("extension dependency %s requires version %s", dependency, required)
		}
		if !installed.Enabled {
			return fmt.Errorf("extension dependency %s is disabled", dependency)
		}
	}
	return nil
}
func compareVersion(a, b string) int {
	parse := func(s string) []int {
		parts := strings.Split(strings.TrimPrefix(s, "v"), ".")
		out := make([]int, 3)
		for i := 0; i < len(parts) && i < 3; i++ {
			n, _ := strconv.Atoi(parts[i])
			out[i] = n
		}
		return out
	}
	x, y := parse(a), parse(b)
	for i := range x {
		if x[i] < y[i] {
			return -1
		}
		if x[i] > y[i] {
			return 1
		}
	}
	return 0
}
func (m *Manager) SetEnabled(id string, enabled bool) (Installed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	x, _, e := m.readVerified(id)
	if e != nil {
		return x, e
	}
	if enabled {
		if err := m.checkDependencies(x.Manifest); err != nil {
			return x, err
		}
	}
	x.Enabled = enabled
	e = m.write(id, x)
	return x, e
}

// AuthorizeCapability is the server-side gate used before an extension starts
// any host operation. It re-reads current metadata so disabling an extension
// takes effect for new requests without trusting a stale UI manifest.
func (m *Manager) AuthorizeCapability(id, capability string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	x, _, err := m.readVerified(id)
	if err != nil {
		return err
	}
	if !x.Enabled {
		return errors.New("extension disabled")
	}
	if err := m.checkDependencies(x.Manifest); err != nil {
		return err
	}
	for _, c := range x.Manifest.Capabilities {
		if c == capability {
			return nil
		}
	}
	return fmt.Errorf("capability %q not granted", capability)
}
func (m *Manager) Uninstall(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, e := m.read(id); e != nil {
		return e
	}
	return m.transact(id, func() error {
		for _, path := range m.transactionFiles(id) {
			if err := m.removeRegistryFile(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		return syncDir(m.root)
	})
}

// CleanupUserData explicitly removes data retained by Uninstall.
func (m *Manager) CleanupUserData(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !appIDPattern.MatchString(id) || len(id) > 128 || strings.HasPrefix(id, "blora.") {
		return errors.New("invalid app id")
	}
	return os.RemoveAll(m.dataPath(id))
}
func (m *Manager) UserDataDir(id string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, err := m.read(id); err != nil {
		return "", err
	}
	if err := os.MkdirAll(m.dataPath(id), 0700); err != nil {
		return "", err
	}
	return m.dataPath(id), nil
}
func (m *Manager) Get(id string) (Installed, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	x, _, err := m.readVerified(id)
	return x, err
}
func (m *Manager) Payload(id string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, payload, err := m.readVerified(id)
	return payload, err
}

// Snapshot keeps metadata and verified bytes from the same registry revision.
func (m *Manager) Snapshot(id string) (Installed, []byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.readVerified(id)
}
func (m *Manager) Rollback(id string) (Installed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	currentMeta, err := m.read(id)
	if err != nil {
		return Installed{}, err
	}
	b, err := os.ReadFile(m.previousPath(id))
	if err != nil {
		return Installed{}, err
	}
	var x Installed
	if err = json.Unmarshal(b, &x); err != nil {
		return x, err
	}
	p, err := os.ReadFile(m.previousPayloadPath(id))
	if err != nil {
		return x, err
	}
	if x.Manifest.AppID != id {
		return Installed{}, errors.New("rollback manifest identity mismatch")
	}
	if err := m.validate(Package{Manifest: x.Manifest, SHA256: x.SHA256, Payload: p, Signature: x.Signature}); err != nil {
		return Installed{}, err
	}
	if _, err := DecodeBundle(p); err != nil {
		return Installed{}, err
	}
	if err := m.checkDependencies(x.Manifest); err != nil {
		return Installed{}, err
	}
	x.Enabled = currentMeta.Enabled
	migratedData, err := m.prepareDataMigration(id, x.Manifest, p, x.Signature)
	if err != nil {
		return Installed{}, err
	}
	if err = m.transact(id, func() error {
		if migratedData != nil {
			if err := atomicRegistryFile(m.dataPath(id), m.dataFile(id), migratedData); err != nil {
				return err
			}
		}
		if err := m.writePayload(id, p); err != nil {
			return err
		}
		return m.write(id, x)
	}); err != nil {
		return x, err
	}
	if _, _, err = m.readVerified(id); err != nil {
		return Installed{}, err
	}
	return x, nil
}
func (m *Manager) List() ([]Installed, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	es, e := os.ReadDir(m.root)
	if e != nil {
		return nil, e
	}
	out := []Installed{}
	for _, v := range es {
		if !strings.HasSuffix(v.Name(), ".json") || strings.HasSuffix(v.Name(), ".previous.json") {
			continue
		}
		x, _, e := m.readVerified(strings.TrimSuffix(v.Name(), ".json"))
		if e == nil {
			out = append(out, x)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Manifest.AppID < out[j].Manifest.AppID })
	return out, nil
}

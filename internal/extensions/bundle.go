package extensions

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
)

const BundlePrefix = "BLORA-BUNDLE-1\n"

type Bundle struct {
	Frontend string `json:"frontend"`
	Backend  []byte `json:"backend,omitempty"`
}

// DecodeBundle retains compatibility with frontend-only JavaScript packages.
// The entire envelope, including both modules, is covered by the package hash.
func DecodeBundle(payload []byte) (Bundle, error) {
	if !bytes.HasPrefix(payload, []byte(BundlePrefix)) {
		return Bundle{Frontend: string(payload)}, nil
	}
	var b Bundle
	d := json.NewDecoder(bytes.NewReader(payload[len(BundlePrefix):]))
	d.DisallowUnknownFields()
	if err := d.Decode(&b); err != nil {
		return b, err
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return b, errors.New("invalid bundle envelope")
	}
	if len(b.Frontend) == 0 || len(b.Frontend) > 4<<20 || len(b.Backend) > MaxBackendModule {
		return b, errors.New("invalid bundle module size")
	}
	if len(b.Backend) > 0 && (len(b.Backend) < 8 || !bytes.Equal(b.Backend[:8], []byte{0, 97, 115, 109, 1, 0, 0, 0})) {
		return b, errors.New("backend must be a WebAssembly module")
	}
	return b, nil
}

func (m *Manager) Backend(id, packageHash string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	x, p, err := m.readVerified(id)
	if err != nil {
		return nil, err
	}
	if !x.Enabled || x.SHA256 != packageHash {
		return nil, errors.New("extension package changed or disabled")
	}
	// Keep at most one decoded module. Every call above still reads and
	// verifies the current package bytes and enabled state before cache use.
	m.backendMu.Lock()
	defer m.backendMu.Unlock()
	if m.backendHash == x.SHA256 && len(m.backendModule) > 0 {
		return append([]byte(nil), m.backendModule...), nil
	}
	b, err := DecodeBundle(p)
	if err != nil {
		return nil, err
	}
	if len(b.Backend) == 0 {
		return nil, errors.New("extension has no backend module")
	}
	m.backendHash = x.SHA256
	m.backendModule = append([]byte(nil), b.Backend...)
	return b.Backend, nil
}

func ModuleHash(module []byte) string {
	h := sha256.Sum256(module)
	return hex.EncodeToString(h[:])
}

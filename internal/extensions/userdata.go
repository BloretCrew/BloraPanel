package extensions

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// prepareDataMigration is called with the registry write lock held. Only the
// expensive sandbox work runs without it; the caller receives the lock back
// after checking that the original package/data/transaction identity is intact.
func (m *Manager) prepareDataMigration(id string, target Manifest, payload, signature []byte) ([]byte, error) {
	d, err := m.readData(id)
	if err != nil {
		return nil, err
	}
	needed := false
	for _, record := range d.Users {
		if record.SchemaVersion != dataVersion(target) {
			needed = true
			break
		}
	}
	if !needed {
		return nil, nil
	}
	if m.migrationActive {
		return nil, errors.New("another extension data migration is running; retry after it completes")
	}
	before, err := m.migrationSnapshot(id)
	if err != nil {
		return nil, err
	}
	m.migrationActive = true
	defer func() { m.migrationActive = false }()
	var migrated []byte
	// Always reacquire before returning, including a panic from the worker.
	func() {
		m.mu.Unlock()
		defer m.mu.Lock()
		migrated, err = migrateUserDocuments(d, target, payload)
	}()
	if err != nil {
		return nil, err
	}
	after, err := m.migrationSnapshot(id)
	if err != nil {
		return nil, err
	}
	if before != after {
		return nil, ErrDataConflict
	}
	// Trust and dependency policy may have changed while compiling/running.
	if err := m.validate(Package{Manifest: target, Payload: payload, SHA256: ModuleHash(payload), Signature: signature}); err != nil {
		return nil, err
	}
	if err := m.checkDependencies(target); err != nil {
		return nil, err
	}
	return migrated, nil
}

func (m *Manager) migrationSnapshot(id string) ([6]string, error) {
	var snapshot [6]string
	if _, err := os.Stat(m.transactionPath(id)); err == nil {
		return snapshot, errors.New("extension transaction requires recovery")
	} else if !errors.Is(err, os.ErrNotExist) {
		return snapshot, err
	}
	paths := m.transactionFiles(id)
	all := []string{paths[0], paths[1], paths[2], paths[3], m.dataFile(id), filepath.Join(m.root, id+".committed")}
	for i, path := range all {
		data, err := readRegistryFile(path, 16<<20)
		if errors.Is(err, os.ErrNotExist) {
			snapshot[i] = "absent"
			continue
		}
		if err != nil {
			return snapshot, err
		}
		snapshot[i] = ModuleHash(data)
	}
	return snapshot, nil
}

// No filesystem access occurs while converting the captured documents. Each
// user gets fresh guest memory, while the compiled code is reused for the batch.
func migrateUserDocuments(d extensionData, target Manifest, payload []byte) ([]byte, error) {
	var err error
	var runner *backendRunner
	defer func() {
		if runner != nil {
			runner.close()
		}
	}()
	changed := false
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	for user, record := range d.Users {
		if record.SchemaVersion == dataVersion(target) {
			continue
		}
		if runner == nil {
			bundle, err := DecodeBundle(payload)
			if err != nil {
				return nil, err
			}
			if len(bundle.Backend) == 0 {
				return nil, errors.New("data migration requires a WASI backend")
			}
			runner, err = compileBackend(ctx, bundle.Backend)
			if err != nil {
				return nil, err
			}
		}
		input, _ := json.Marshal(map[string]any{"operation": "blora.data.migrate", "fromVersion": record.SchemaVersion, "toVersion": dataVersion(target), "data": record.Data})
		out, err := runner.run(ctx, input)
		if err != nil {
			return nil, err
		}
		var result struct {
			SchemaVersion int             `json:"schemaVersion"`
			Data          json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(out, &result); err != nil {
			return nil, err
		}
		if result.SchemaVersion != dataVersion(target) || len(result.Data) > MaxUserData || !json.Valid(result.Data) {
			return nil, errors.New("invalid migrated user data")
		}
		record.SchemaVersion = result.SchemaVersion
		record.Data = result.Data
		record.Revision++
		d.Users[user] = record
		changed = true
	}
	if !changed {
		return nil, nil
	}
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	if len(b) > MaxExtensionData {
		return nil, errors.New("migrated extension data exceeds quota")
	}
	return b, nil
}

const MaxUserData = 16 << 10
const MaxExtensionData = 4 << 20

var ErrDataConflict = errors.New("extension data revision or request conflict")

type UserData struct {
	SchemaVersion int             `json:"schemaVersion"`
	Revision      int64           `json:"revision"`
	Data          json.RawMessage `json:"data"`
}
type dataReceipt struct {
	Digest   string `json:"digest"`
	Revision int64  `json:"revision"`
}
type userRecord struct {
	UserData
	Requests map[string]dataReceipt `json:"requests"`
}
type extensionData struct {
	Version int                   `json:"version"`
	Users   map[string]userRecord `json:"users"`
}

func dataVersion(m Manifest) int {
	if m.DataSchemaVersion == 0 {
		return 1
	}
	return m.DataSchemaVersion
}
func (m *Manager) dataFile(id string) string { return filepath.Join(m.dataPath(id), "records.json") }
func (m *Manager) readData(id string) (extensionData, error) {
	var d extensionData
	b, err := readRegistryFile(m.dataFile(id), MaxExtensionData)
	if errors.Is(err, os.ErrNotExist) {
		return extensionData{Version: 1, Users: map[string]userRecord{}}, nil
	}
	if err != nil {
		return d, err
	}
	if err := json.Unmarshal(b, &d); err != nil {
		return d, err
	}
	if d.Version != 1 || d.Users == nil {
		return d, errors.New("invalid extension data file")
	}
	for owner, record := range d.Users {
		if owner == "" || len(owner) > 128 || record.Revision < 1 || record.SchemaVersion < 1 || record.SchemaVersion > 1_000_000 || len(record.Data) > MaxUserData || !json.Valid(record.Data) {
			return d, errors.New("invalid stored user data")
		}
		for request, receipt := range record.Requests {
			if request == "" || len(request) > 128 || len(receipt.Digest) != 64 || receipt.Revision < 1 || receipt.Revision > record.Revision {
				return d, errors.New("invalid stored data receipt")
			}
		}
	}
	return d, nil
}
func (m *Manager) dataAccess(id, user, capability string) (Installed, error) {
	if user == "" || len(user) > 128 {
		return Installed{}, errors.New("invalid data owner")
	}
	x, _, err := m.readVerified(id)
	if err != nil {
		return x, err
	}
	if !x.Enabled {
		return x, errors.New("extension disabled")
	}
	if err := m.checkDependencies(x.Manifest); err != nil {
		return x, err
	}
	for _, c := range x.Manifest.Capabilities {
		if c == capability {
			return x, nil
		}
	}
	return x, errors.New("extension data capability not declared")
}

// ReadUserData never accepts an owner from extension-controlled JSON. The
// authenticated API must provide its user ID and separately enforce app.use.
func (m *Manager) ReadUserData(id, user string) (UserData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	x, err := m.dataAccess(id, user, "data.read")
	if err != nil {
		return UserData{}, err
	}
	d, err := m.readData(id)
	if err != nil {
		return UserData{}, err
	}
	if record, ok := d.Users[user]; ok {
		if record.SchemaVersion != dataVersion(x.Manifest) {
			return UserData{}, errors.New("extension data migration required")
		}
		return record.UserData, nil
	}
	return UserData{SchemaVersion: dataVersion(x.Manifest), Data: json.RawMessage(`{}`)}, nil
}

// WriteUserData conditionally replaces only this user's document. Receipts are
// durable and never evicted silently; the shared bounded data quota applies.
func (m *Manager) WriteUserData(id, user, request string, expected int64, schema int, value json.RawMessage) (UserData, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	x, err := m.dataAccess(id, user, "data.write")
	if err != nil {
		return UserData{}, err
	}
	if request == "" || len(request) > 128 || expected < 0 || len(value) > MaxUserData || !json.Valid(value) {
		return UserData{}, errors.New("invalid extension data write")
	}
	if schema != dataVersion(x.Manifest) {
		return UserData{}, ErrDataConflict
	}
	d, err := m.readData(id)
	if err != nil {
		return UserData{}, err
	}
	record, ok := d.Users[user]
	if !ok {
		record = userRecord{UserData: UserData{SchemaVersion: schema, Data: json.RawMessage(`{}`)}, Requests: map[string]dataReceipt{}}
	}
	digestBytes, _ := json.Marshal(struct {
		Expected int64
		Schema   int
		Value    json.RawMessage
	}{expected, schema, value})
	digest := ModuleHash(digestBytes)
	if old, ok := record.Requests[request]; ok {
		if old.Digest != digest {
			return UserData{}, ErrDataConflict
		}
		// Return current data without overwriting a later successful write.
		return record.UserData, nil
	}
	if record.Revision != expected || record.SchemaVersion != schema {
		return UserData{}, ErrDataConflict
	}
	record.Revision++
	record.Data = append(json.RawMessage(nil), value...)
	if record.Requests == nil {
		record.Requests = map[string]dataReceipt{}
	}
	record.Requests[request] = dataReceipt{Digest: digest, Revision: record.Revision}
	d.Users[user] = record
	b, err := json.Marshal(d)
	if err != nil {
		return UserData{}, err
	}
	if len(b) > MaxExtensionData {
		return UserData{}, errors.New("extension data quota exceeded")
	}
	if err := atomicRegistryFile(m.dataPath(id), m.dataFile(id), b); err != nil {
		return UserData{}, err
	}
	return record.UserData, nil
}

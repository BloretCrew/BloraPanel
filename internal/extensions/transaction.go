package extensions

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const transactionSuffix = ".transaction"

type savedFile struct {
	Exists bool   `json:"exists"`
	Data   []byte `json:"data,omitempty"`
}
type registryTransaction struct {
	Version int          `json:"version"`
	ID      string       `json:"id"`
	Token   string       `json:"token"`
	Files   [4]savedFile `json:"files"`
	Data    *savedFile   `json:"userData,omitempty"`
}

func (m *Manager) transactionPath(id string) string {
	return filepath.Join(m.root, id+transactionSuffix)
}
func (m *Manager) transactionFiles(id string) [4]string {
	return [4]string{m.path(id), m.payloadPath(id), m.previousPath(id), m.previousPayloadPath(id)}
}

func atomicRegistryFile(root, path string, data []byte) error {
	f, err := os.CreateTemp(root, ".registry-")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(data)
	}
	if err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = replaceRegistryFile(name, path); err != nil {
		return err
	}
	return syncDir(root)
}

// Rename deletion out of the live namespace before unlinking. On Windows the
// rename uses WRITE_THROUGH; a leftover temporary file is never an installed
// package or a recovery record.
func (m *Manager) removeRegistryFile(path string) error {
	f, err := os.CreateTemp(m.root, ".registry-removed-")
	if err != nil {
		return err
	}
	tombstone := f.Name()
	defer os.Remove(tombstone)
	if err := f.Close(); err != nil {
		return err
	}
	return replaceRegistryFile(path, tombstone)
}

func (m *Manager) beginTransaction(id string) (registryTransaction, error) {
	tx := registryTransaction{Version: 1, ID: id}
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return tx, err
	}
	tx.Token = hex.EncodeToString(nonce[:])
	if !appIDPattern.MatchString(id) || len(id) > 128 || strings.HasPrefix(id, "blora.") {
		return tx, errors.New("invalid transaction app id")
	}
	if _, err := os.Stat(m.transactionPath(id)); err == nil {
		return tx, errors.New("extension transaction requires recovery")
	} else if !errors.Is(err, os.ErrNotExist) {
		return tx, err
	}
	for i, path := range m.transactionFiles(id) {
		data, err := readRegistryFile(path, 16<<20)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return tx, err
		}
		tx.Files[i] = savedFile{Exists: true, Data: data}
	}
	data, dataErr := readRegistryFile(m.dataFile(id), MaxExtensionData)
	if dataErr != nil && !errors.Is(dataErr, os.ErrNotExist) {
		return tx, dataErr
	}
	tx.Data = &savedFile{Exists: dataErr == nil, Data: data}
	body, err := json.Marshal(tx)
	if err != nil {
		return tx, err
	}
	envelope, err := json.Marshal(struct {
		Body   json.RawMessage `json:"body"`
		SHA256 string          `json:"sha256"`
	}{body, ModuleHash(body)})
	if err != nil {
		return tx, err
	}
	if len(envelope) > 64<<20 {
		return tx, errors.New("extension recovery record too large")
	}
	return tx, atomicRegistryFile(m.root, m.transactionPath(id), envelope)
}

func readRegistryFile(path string, limit int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	closeErr := f.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if int64(len(data)) > limit {
		return nil, errors.New("registry file exceeds size limit")
	}
	return data, nil
}

func (m *Manager) restoreTransaction(tx registryTransaction) error {
	if tx.Data != nil {
		if tx.Data.Exists {
			if err := os.MkdirAll(m.dataPath(tx.ID), 0700); err != nil {
				return err
			}
			if err := atomicRegistryFile(m.dataPath(tx.ID), m.dataFile(tx.ID), tx.Data.Data); err != nil {
				return err
			}
		} else if err := m.removeRegistryFile(m.dataFile(tx.ID)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	for i, path := range m.transactionFiles(tx.ID) {
		old := tx.Files[i]
		if old.Exists {
			if err := atomicRegistryFile(m.root, path, old.Data); err != nil {
				return err
			}
		} else if err := m.removeRegistryFile(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if err := syncDir(m.root); err != nil {
		return err
	}
	return m.finishTransaction(tx.ID)
}
func (m *Manager) finishTransaction(id string) error {
	if err := m.removeRegistryFile(m.transactionPath(id)); err != nil {
		return err
	}
	return syncDir(m.root)
}

func (m *Manager) commitTransaction(tx registryTransaction) error {
	// Keep a durable token even when deletion of the undo file is replayed
	// after a crash. Each new transaction has a fresh, unguessable token.
	if err := m.markCommitted(tx); err != nil {
		return err
	}
	return m.finishTransaction(tx.ID)
}

func (m *Manager) markCommitted(tx registryTransaction) error {
	return atomicRegistryFile(m.root, filepath.Join(m.root, tx.ID+".committed"), []byte(tx.Token))
}

// The undo record is durable before any package or rollback file changes.
// An interruption before commit restores all four files, including absence.
func (m *Manager) transact(id string, change func() error) error {
	tx, err := m.beginTransaction(id)
	if err != nil {
		return err
	}
	if err = change(); err != nil {
		if restoreErr := m.restoreTransaction(tx); restoreErr != nil {
			return fmt.Errorf("%w; recovery pending: %v", err, restoreErr)
		}
		return err
	}
	return m.commitTransaction(tx)
}

func (m *Manager) recoverTransactions() error {
	entries, err := os.ReadDir(m.root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), transactionSuffix) {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), transactionSuffix)
		if !appIDPattern.MatchString(id) || len(id) > 128 || strings.HasPrefix(id, "blora.") {
			return errors.New("invalid recovery record identity")
		}
		f, err := os.Open(m.transactionPath(id))
		if err != nil {
			return err
		}
		data, err := io.ReadAll(io.LimitReader(f, (64<<20)+1))
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		if len(data) > 64<<20 {
			return errors.New("recovery record too large")
		}
		var envelope struct {
			Body   json.RawMessage `json:"body"`
			SHA256 string          `json:"sha256"`
		}
		if err := json.Unmarshal(data, &envelope); err != nil {
			return err
		}
		if ModuleHash(envelope.Body) != envelope.SHA256 {
			return errors.New("recovery record integrity failed")
		}
		var tx registryTransaction
		if err := json.Unmarshal(envelope.Body, &tx); err != nil {
			return err
		}
		if tx.ID != id || tx.Version != 1 {
			return errors.New("recovery record identity or version mismatch")
		}
		if tx.Token != "" {
			committed, err := readRegistryFile(filepath.Join(m.root, id+".committed"), 128)
			if err == nil && string(committed) == tx.Token {
				if err := m.finishTransaction(id); err != nil {
					return err
				}
				continue
			}
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		if err := m.restoreTransaction(tx); err != nil {
			return fmt.Errorf("recover extension %s: %w", id, err)
		}
	}
	return nil
}

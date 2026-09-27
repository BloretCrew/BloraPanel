package extensions

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestUserDataConcurrentCASAndQuotaKeepCommittedDocument(t *testing.T) {
	m, err := New(t.TempDir(), []string{"data.read", "data.write"})
	if err != nil {
		t.Fatal(err)
	}
	p := pkg("data.concurrent", "1.0.0", []byte("export function start(){}"))
	p.Manifest.Capabilities = []string{"data.read", "data.write"}
	if _, err := m.Install(p); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := m.WriteUserData(p.Manifest.AppID, "alice", fmt.Sprint(i), 0, 1, json.RawMessage(`{"note":"kept"}`))
			results <- err
		}(i)
	}
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, ErrDataConflict) {
			conflict++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatal("concurrent writes did not use CAS")
	}
	if _, err := m.WriteUserData(p.Manifest.AppID, "alice", "oversize", 1, 1, json.RawMessage(`"`+strings.Repeat("x", MaxUserData)+`"`)); err == nil {
		t.Fatal("oversize document accepted")
	}
	d, err := m.readData(p.Manifest.AppID)
	if err != nil {
		t.Fatal(err)
	}
	record := d.Users["alice"]
	// Fill the persistent receipt budget with valid, fixed-size receipts.
	receipt := dataReceipt{Digest: strings.Repeat("a", 64), Revision: 1}
	entry, _ := json.Marshal(map[string]dataReceipt{"00000000": receipt})
	base, _ := json.Marshal(d)
	count := (MaxExtensionData - len(base)) / (len(entry) - 1)
	for i := 0; i < count; i++ {
		record.Requests[fmt.Sprintf("%08d", i)] = receipt
	}
	d.Users["alice"] = record
	filled, _ := json.Marshal(d)
	if len(filled) > MaxExtensionData {
		t.Fatal("invalid quota setup")
	}
	if err := atomicRegistryFile(m.dataPath(p.Manifest.AppID), m.dataFile(p.Manifest.AppID), filled); err != nil {
		t.Fatal(err)
	}
	_, err = m.WriteUserData(p.Manifest.AppID, "bob", "new-user", 0, 1, json.RawMessage(`"`+strings.Repeat("b", 1024)+`"`))
	if err == nil || !strings.Contains(err.Error(), "quota") {
		t.Fatalf("quota was not enforced: %v", err)
	}
	after, err := os.ReadFile(m.dataFile(p.Manifest.AppID))
	if err != nil || string(after) != string(filled) {
		t.Fatal("quota failure changed committed file")
	}
}

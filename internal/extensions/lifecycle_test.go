package extensions

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
)

func TestCorruptRollbackPreservesCurrentPackage(t *testing.T) {
	for _, corruption := range []string{"payload", "identity", "metadata"} {
		t.Run(corruption, func(t *testing.T) {
			m, _ := New(t.TempDir(), []string{"window.open"})
			if _, err := m.Install(pkg("example.app", "1.0.0", []byte("old"))); err != nil {
				t.Fatal(err)
			}
			if _, err := m.Upgrade(pkg("example.app", "2.0.0", []byte("current"))); err != nil {
				t.Fatal(err)
			}
			var err error
			switch corruption {
			case "payload":
				err = os.WriteFile(m.previousPayloadPath("example.app"), []byte("corrupt"), 0600)
			case "metadata":
				err = os.WriteFile(m.previousPath("example.app"), []byte("invalid"), 0600)
			case "identity":
				b, readErr := os.ReadFile(m.previousPath("example.app"))
				if readErr != nil {
					t.Fatal(readErr)
				}
				var old Installed
				if err = json.Unmarshal(b, &old); err != nil {
					t.Fatal(err)
				}
				old.Manifest.AppID = "another.app"
				b, err = json.Marshal(old)
				if err != nil {
					t.Fatal(err)
				}
				err = os.WriteFile(m.previousPath("example.app"), b, 0600)
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err := m.Rollback("example.app"); err == nil {
				t.Fatal("corrupt rollback accepted")
			}
			x, b, err := m.Snapshot("example.app")
			if err != nil || x.Manifest.PackageVersion != "2.0.0" || string(b) != "current" {
				t.Fatalf("rollback damaged current package: %+v %s %v", x, b, err)
			}
		})
	}
}

func TestConcurrentLifecycleReadersSeeVerifiedSnapshots(t *testing.T) {
	m, _ := New(t.TempDir(), []string{"window.open"})
	if _, err := m.Install(pkg("example.app", "1.0.0", []byte("v1"))); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 60 {
				x, b, err := m.Snapshot("example.app")
				if err != nil || ModuleHash(b) != x.SHA256 {
					t.Errorf("torn snapshot: %v", err)
					return
				}
				if _, err := m.Get("example.app"); err != nil {
					t.Error(err)
					return
				}
			}
		}()
	}
	for version := 2; version <= 20; version++ {
		if _, err := m.Upgrade(pkg("example.app", fmt.Sprintf("%d.0.0", version), []byte(fmt.Sprintf("v%d", version)))); err != nil {
			t.Fatal(err)
		}
		if _, err := m.SetEnabled("example.app", version%2 == 0); err != nil {
			t.Fatal(err)
		}
	}
	wg.Wait()
}

func TestRollbackKeepsDisabledAndDependencyGateTracksChanges(t *testing.T) {
	m, _ := New(t.TempDir(), []string{"window.open"})
	base := pkg("example.base", "1.0.0", []byte("base"))
	if _, err := m.Install(base); err != nil {
		t.Fatal(err)
	}
	dependent := pkg("example.dependent", "1.0.0", []byte("dependent"))
	dependent.Manifest.Dependencies = map[string]string{"example.base": "1.0.0"}
	if _, err := m.Install(dependent); err != nil {
		t.Fatal(err)
	}
	if _, err := m.SetEnabled("example.base", false); err != nil {
		t.Fatal(err)
	}
	if err := m.AuthorizeCapability("example.dependent", "window.open"); err == nil {
		t.Fatal("disabled dependency still usable")
	}
	if _, err := m.SetEnabled("example.dependent", true); err == nil {
		t.Fatal("enabled with unavailable dependency")
	}
	if _, err := m.Upgrade(pkg("example.base", "2.0.0", []byte("new"))); err != nil {
		t.Fatal(err)
	}
	x, err := m.Rollback("example.base")
	if err != nil || x.Enabled {
		t.Fatalf("rollback re-enabled package: %+v %v", x, err)
	}
}

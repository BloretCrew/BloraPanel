package extensions

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

func TestRegistryTransactionCrashChild(t *testing.T) {
	root := os.Getenv("BLORA_TRANSACTION_TEST_ROOT")
	if root == "" {
		return
	}
	phase, _ := strconv.Atoi(os.Getenv("BLORA_TRANSACTION_TEST_PHASE"))
	m, err := New(root, []string{"window.open"})
	if err != nil {
		t.Fatal(err)
	}
	p := pkg("example.app", "3.0.0", []byte("new"))
	x := Installed{Manifest: p.Manifest, SHA256: p.SHA256, Enabled: true}
	writeData := func() error {
		if err := os.MkdirAll(m.dataPath("example.app"), 0700); err != nil {
			return err
		}
		return atomicRegistryFile(m.dataPath("example.app"), m.dataFile("example.app"), []byte("new user data"))
	}
	if phase == 6 {
		tx, err := m.beginTransaction("example.app")
		if err != nil {
			t.Fatal(err)
		}
		if err := writeData(); err != nil {
			t.Fatal(err)
		}
		if err := m.writePayload("example.app", p.Payload); err != nil {
			t.Fatal(err)
		}
		if err := m.write("example.app", x); err != nil {
			t.Fatal(err)
		}
		if err := m.markCommitted(tx); err != nil {
			t.Fatal(err)
		}
		os.Exit(91)
	}
	err = m.transact("example.app", func() error {
		if phase == 0 {
			os.Exit(91)
		}
		if err := writeData(); err != nil {
			return err
		}
		if err := m.writePayload("example.app", p.Payload); err != nil {
			return err
		}
		if phase == 1 {
			os.Exit(91)
		}
		if err := m.write("example.app", x); err != nil {
			return err
		}
		if phase == 2 {
			os.Exit(91)
		}
		previous := pkg("example.app", "2.0.0", []byte("intermediate"))
		previousMeta, _ := json.Marshal(Installed{Manifest: previous.Manifest, SHA256: previous.SHA256, Enabled: true})
		if err := atomicRegistryFile(m.root, m.previousPath("example.app"), previousMeta); err != nil {
			return err
		}
		if err := atomicRegistryFile(m.root, m.previousPayloadPath("example.app"), previous.Payload); err != nil {
			return err
		}
		if phase == 3 {
			os.Exit(91)
		}
		if phase == 4 {
			for _, path := range m.transactionFiles("example.app") {
				if err := os.Remove(path); err != nil {
					return err
				}
			}
			os.Exit(91)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	os.Exit(91)
}

func TestRegistryRecoversInterruptedTransactions(t *testing.T) {
	for _, fresh := range []bool{false, true} {
		for phase := 0; phase <= 6; phase++ {
			t.Run(strconv.FormatBool(fresh)+"/"+strconv.Itoa(phase), func(t *testing.T) {
				root := t.TempDir()
				m, err := New(root, []string{"window.open"})
				if err != nil {
					t.Fatal(err)
				}
				if !fresh {
					if _, err := m.Install(pkg("example.app", "1.0.0", []byte("previous"))); err != nil {
						t.Fatal(err)
					}
					if _, err := m.Upgrade(pkg("example.app", "2.0.0", []byte("current"))); err != nil {
						t.Fatal(err)
					}
				}
				if !fresh {
					if err := os.WriteFile(m.dataFile("example.app"), []byte("old user data"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				child := exec.Command(os.Args[0], "-test.run=^TestRegistryTransactionCrashChild$")
				child.Env = append(os.Environ(), "BLORA_TRANSACTION_TEST_ROOT="+root, "BLORA_TRANSACTION_TEST_PHASE="+strconv.Itoa(phase))
				out, err := child.CombinedOutput()
				var exit *exec.ExitError
				if !errors.As(err, &exit) || exit.ExitCode() != 91 {
					t.Fatalf("child did not interrupt at phase: %v %s", err, out)
				}
				if phase < 5 {
					if _, err := m.Get("example.app"); err == nil {
						t.Fatal("pending transaction exposed as installed")
					}
				}
				recovered, err := New(root, []string{"window.open"})
				if err != nil {
					t.Fatal(err)
				}
				x, b, err := recovered.Snapshot("example.app")
				data, dataErr := os.ReadFile(recovered.dataFile("example.app"))
				if phase >= 5 {
					if dataErr != nil || string(data) != "new user data" {
						t.Fatal("committed user data lost")
					}
				} else if fresh {
					if !errors.Is(dataErr, os.ErrNotExist) {
						t.Fatal("unexpected data after recovery")
					}
				} else if dataErr != nil || string(data) != "old user data" {
					t.Fatal("original user data not restored")
				}
				if phase >= 5 {
					if err != nil || x.Manifest.PackageVersion != "3.0.0" || string(b) != "new" {
						t.Fatalf("committed version lost: %v", err)
					}
				} else if fresh {
					if !errors.Is(err, os.ErrNotExist) {
						t.Fatalf("interrupted install survived: %v", err)
					}
				} else {
					if err != nil || x.Manifest.PackageVersion != "2.0.0" || string(b) != "current" {
						t.Fatalf("current snapshot not restored: %+v %s %v", x, b, err)
					}
					previous, err := os.ReadFile(recovered.previousPayloadPath("example.app"))
					if err != nil || string(previous) != "previous" {
						t.Fatalf("rollback snapshot not restored: %s %v", previous, err)
					}
				}
				if _, err := New(root, []string{"window.open"}); err != nil {
					t.Fatalf("second recovery: %v", err)
				}
			})
		}
	}
}

func TestRegistryTransactionErrorRestoresAndCorruptionBlocksRecovery(t *testing.T) {
	m, _ := New(t.TempDir(), []string{"window.open"})
	if _, err := m.Install(pkg("example.app", "1.0.0", []byte("original"))); err != nil {
		t.Fatal(err)
	}
	err := m.transact("example.app", func() error {
		if err := m.writePayload("example.app", []byte("partial")); err != nil {
			return err
		}
		return errors.New("simulated write failure")
	})
	if err == nil {
		t.Fatal("write failure hidden")
	}
	if b, err := m.Payload("example.app"); err != nil || string(b) != "original" {
		t.Fatalf("failed write damaged package: %s %v", b, err)
	}
	if _, err := m.beginTransaction("example.app"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(m.transactionPath("example.app"), []byte(`{"body":{},"sha256":"invalid"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := New(m.root, []string{"window.open"}); err == nil {
		t.Fatal("corrupt recovery record accepted")
	}
	if b, err := os.ReadFile(m.payloadPath("example.app")); err != nil || string(b) != "original" {
		t.Fatal("corrupt journal changed package")
	}
}

func TestRollbackFilenameCannotBecomeAnInstalledApp(t *testing.T) {
	m, _ := New(t.TempDir(), []string{"window.open"})
	if _, err := m.Install(pkg("example.app.previous", "1.0.0", []byte("other app"))); err == nil {
		t.Fatal("app can collide with rollback files")
	}
	if _, err := m.Install(pkg("example.app", "1.0.0", []byte("original"))); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Upgrade(pkg("example.app", "2.0.0", []byte("new"))); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Rollback("example.app"); err != nil {
		t.Fatal(err)
	}
	if b, err := m.Payload("example.app"); err != nil || string(b) != "original" {
		t.Fatal("reserved suffix check broke normal rollback")
	}
}

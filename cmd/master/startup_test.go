//go:build !windows

package main

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"blora.dev/panel/internal/bootstrap"
)

// The launcher re-executes this test binary as a management child. Dispatch
// its private prefix before testing parses flags, just as the real executable
// dispatches it before configuration parsing.
func TestMain(m *testing.M) {
	if os.Getenv("BLORA_CONFIG_TEST_CHILD") == "1" && len(os.Args) > 1 && os.Args[1] == bootstrap.ManagedFlag {
		main()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestMasterConfigProcess(t *testing.T) {
	if os.Getenv("BLORA_CONFIG_TEST_CHILD") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{os.Args[0]}, os.Args[i+1:]...)
			main()
			os.Exit(0)
		}
	}
	os.Exit(2)
}

func TestMasterConfigStartupTLSAndRestart(t *testing.T) {
	dir := t.TempDir()
	passwordBytes := make([]byte, 24)
	if _, err := rand.Read(passwordBytes); err != nil {
		t.Fatal(err)
	}
	password := hex.EncodeToString(passwordBytes)
	if err := os.WriteFile(filepath.Join(dir, "initial-password"), []byte(password+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	_ = listener.Close()
	body, _ := json.Marshal(masterConfig{StateDir: filepath.Join(dir, "state"), Listen: address, PasswordFile: filepath.Join(dir, "initial-password"), AdminName: "owner", HTTPS: true})
	path := filepath.Join(dir, "master.json")
	if err := os.WriteFile(path, body, 0600); err != nil {
		t.Fatal(err)
	}
	command := func(args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestMasterConfigProcess$", "--", "--config", path}, args...)...)
		cmd.Env = append(os.Environ(), "BLORA_CONFIG_TEST_CHILD=1")
		cmd.Dir = dir
		return cmd
	}
	out, err := command("--init").CombinedOutput()
	if err != nil || strings.Contains(string(out), password) {
		t.Fatalf("configuration initialization failed or exposed password: %v", err)
	}
	certPath := filepath.Join(dir, "state", "tls.crt")
	cert, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatal(err)
	}
	out, err = command("--init").CombinedOutput()
	if err == nil || !strings.Contains(string(out), "already initialized") {
		t.Fatal("reinitialization did not refuse existing state")
	}
	after, _ := os.ReadFile(certPath)
	if string(after) != string(cert) {
		t.Fatal("reinitialization replaced certificate")
	}
	if err := os.Remove(filepath.Join(dir, "initial-password")); err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(cert) {
		t.Fatal("invalid generated certificate")
	}
	transport := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: time.Second}
	for attempt := 0; attempt < 2; attempt++ {
		cmd := command()
		cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = cmd.Process.Kill() })
		deadline := time.Now().Add(15 * time.Second)
		ok := false
		for time.Now().Before(deadline) {
			response, err := client.Get("https://" + address + "/healthz")
			if err == nil {
				_ = response.Body.Close()
				ok = response.StatusCode == http.StatusOK
				if ok {
					break
				}
			}
			time.Sleep(50 * time.Millisecond)
		}
		if !ok {
			t.Fatal("configuration-only process did not serve verified HTTPS health")
		}
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			t.Fatal(err)
		}
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("graceful shutdown: %v", err)
			}
		case <-time.After(15 * time.Second):
			t.Fatal("shutdown timed out")
		}
		transport.CloseIdleConnections()
	}
}

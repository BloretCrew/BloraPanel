package main

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"encoding/hex"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"blora.dev/panel/internal/bootstrap"
	"blora.dev/panel/internal/master"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if err := run(); err != nil {
		slog.Error("master stopped", "error", err)
		os.Exit(1)
	}
}
func run() error {
	state := flag.String("state-dir", ".local/master", "private state directory")
	listen := flag.String("listen", "127.0.0.1:8443", "HTTPS address")
	origin := flag.String("origin", "", "optional fixed browser origin; otherwise loopback or trusted proxy HTTPS origin")
	proxyList := flag.String("trusted-proxies", "", "comma-separated trusted proxy IPs/CIDRs; default trusts none")
	static := flag.String("static-dir", "web/dist", "built frontend directory")
	extensionsDir := flag.String("extensions-dir", "", "private extension registry directory (disabled when empty)")
	extensionsCatalog := flag.String("extensions-catalog", "", "local extension catalog directory (disabled when empty)")
	extensionsCatalogURL := flag.String("extensions-catalog-url", "", "HTTPS extension catalog base URL (index.json and package files; disabled when empty)")
	extensionsKey := flag.String("extensions-public-key", "", "hex Ed25519 public key; makes package signatures mandatory")
	initialize := flag.Bool("init", false, "initialize administrator and fresh local TLS, then exit")
	name := flag.String("admin-name", "admin", "initial administrator name")
	passwordFile := flag.String("password-file", "", "initial password file; contents never logged")
	cert := flag.String("tls-cert", "", "TLS certificate path")
	key := flag.String("tls-key", "", "TLS private key path")
	flag.Parse()
	proxies, err := master.ParseTrustedProxies(*proxyList)
	if err != nil {
		return err
	}
	s, err := storage.Open(filepath.Join(*state, "master.db"))
	if err != nil {
		return err
	}
	defer s.Close()
	if *initialize {
		users, err := s.Users(context.Background())
		if err != nil {
			return err
		}
		if len(users) > 0 {
			return fmt.Errorf("already initialized; refusing to replace accounts")
		}
		if *passwordFile == "" {
			return fmt.Errorf("--password-file is required for initialization")
		}
		b, err := os.ReadFile(*passwordFile)
		if err != nil {
			return err
		}
		password := strings.TrimSuffix(strings.TrimSuffix(string(b), "\n"), "\r")
		if len(password) < 12 || len(password) > 72 {
			return fmt.Errorf("password must contain 12–72 bytes")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			return err
		}
		if err := bootstrap.LocalTLS(*state); err != nil {
			return err
		}
		if _, err := s.CreateUser(context.Background(), model.User{Name: *name, Admin: true}, hash); err != nil {
			return err
		}
		slog.Info("initialized", "stateDir", *state, "adminName", *name)
		return nil
	}
	users, err := s.Users(context.Background())
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return fmt.Errorf("not initialized; run with --init and a private --password-file")
	}
	if *cert == "" {
		*cert = filepath.Join(*state, "tls.crt")
	}
	if *key == "" {
		*key = filepath.Join(*state, "tls.key")
	}
	var trusted []ed25519.PublicKey
	if *extensionsKey != "" {
		b, e := hex.DecodeString(*extensionsKey)
		if e != nil || len(b) != ed25519.PublicKeySize {
			return fmt.Errorf("--extensions-public-key must be 32-byte hex")
		}
		trusted = []ed25519.PublicKey{ed25519.PublicKey(b)}
	}
	app := master.New(master.Options{Store: s, Origin: *origin, TrustedProxies: proxies, StaticDir: *static, ExtensionRoot: *extensionsDir, ExtensionCatalogDir: *extensionsCatalog, ExtensionCatalogURL: *extensionsCatalogURL, ExtensionTrustedKeys: trusted})
	defer app.Close()
	server := &http.Server{Addr: *listen, Handler: app, ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 90 * time.Second, MaxHeaderBytes: 32768, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12}}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	go func() {
		<-ctx.Done()
		shutdown, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		_ = app.Close()
		_ = server.Shutdown(shutdown)
	}()
	slog.Info("HTTPS management listening", "address", *listen)
	err = server.ListenAndServeTLS(*cert, *key)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

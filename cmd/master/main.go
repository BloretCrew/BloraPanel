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
	if err := run(); err != nil && err != flag.ErrHelp {
		slog.Error("master stopped", "error", err)
		os.Exit(1)
	}
}
func run() error {
	config, initialize, err := parseConfig(os.Args[1:], os.Stderr)
	if err != nil {
		return err
	}
	return runConfig(config, initialize)
}

func runConfig(config masterConfig, initialize bool) error {
	proxies, err := master.ParseTrustedProxies(strings.Join(config.TrustedProxies, ","))
	if err != nil {
		return err
	}
	s, err := storage.Open(filepath.Join(config.StateDir, "master.db"))
	if err != nil {
		return err
	}
	defer s.Close()
	users, err := s.Users(context.Background())
	if err != nil {
		return err
	}
	if initialize || len(users) == 0 {
		if len(users) > 0 {
			return fmt.Errorf("already initialized; refusing to replace accounts")
		}
		if err := initializeMaster(s, config); err != nil {
			return err
		}
		if initialize {
			return nil
		}
	}
	if config.TLSCert == "" {
		config.TLSCert = filepath.Join(config.StateDir, "tls.crt")
	}
	if config.TLSKey == "" {
		config.TLSKey = filepath.Join(config.StateDir, "tls.key")
	}
	var trusted []ed25519.PublicKey
	if config.ExtensionsPublicKey != "" {
		b, e := hex.DecodeString(config.ExtensionsPublicKey)
		if e != nil || len(b) != ed25519.PublicKeySize {
			return fmt.Errorf("extensionsPublicKey must be 32-byte hex")
		}
		trusted = []ed25519.PublicKey{ed25519.PublicKey(b)}
	}
	app := master.New(master.Options{Store: s, Origin: config.Origin, TrustedProxies: proxies,
		TrustAllProxies: !config.VerifyProxyIPs && !config.legacy, AllowLoopbackHTTP: !config.HTTPS,
		StaticDir: config.StaticDir, ExtensionRoot: config.ExtensionsDir, ExtensionCatalogDir: config.ExtensionsCatalog,
		ExtensionCatalogURL: config.ExtensionsCatalogURL, ExtensionTrustedKeys: trusted})
	defer app.Close()
	server := &http.Server{Addr: config.Listen, Handler: app, ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 90 * time.Second, MaxHeaderBytes: 32768, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12}}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	go func() {
		<-ctx.Done()
		shutdown, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		_ = app.Close()
		_ = server.Shutdown(shutdown)
	}()
	slog.Info("management listening", "address", config.Listen, "https", config.HTTPS)
	if config.HTTPS {
		err = server.ListenAndServeTLS(config.TLSCert, config.TLSKey)
	} else {
		err = server.ListenAndServe()
	}
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func initializeMaster(s *storage.Store, config masterConfig) error {
	if config.PasswordFile == "" {
		return fmt.Errorf("passwordFile is required in the configuration for initialization")
	}
	b, err := os.ReadFile(config.PasswordFile)
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
	if config.HTTPS && config.TLSCert == "" {
		if err := bootstrap.LocalTLS(config.StateDir); err != nil {
			return err
		}
	}
	if _, err := s.CreateUser(context.Background(), model.User{Name: config.AdminName, Admin: true}, hash); err != nil {
		return err
	}
	slog.Info("initialized", "stateDir", config.StateDir, "adminName", config.AdminName)
	return nil
}

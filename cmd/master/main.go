package main

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"blora.dev/panel/internal/bootstrap"
	"blora.dev/panel/internal/coreupdate"
	"blora.dev/panel/internal/master"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--core-update-info" {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"formatVersion": 1, "component": "master", "version": coreupdate.Version, "revision": coreupdate.BuildRevision(), "protocolVersion": protocol.Version, "schemaVersion": storage.SchemaVersion(), "schemaFingerprint": storage.SchemaFingerprint(), "preserveInstances": true, "managedRestart": true})
		return
	}
	if err := run(); err != nil && err != flag.ErrHelp {
		slog.Error("master stopped", "error", err)
		os.Exit(1)
	}
}
func run() error {
	child, args, err := bootstrap.ParseServiceChild(os.Args[1:])
	if err != nil {
		return err
	}
	var config masterConfig
	var initialize bool
	if child == nil {
		config, initialize, err = parseConfig(args, os.Stderr)
	} else {
		config, initialize, err = parseConfigAt(args, os.Stderr, child.InstallRoot)
	}
	if err != nil {
		return err
	}
	if child == nil && !initialize && !config.legacy {
		executable, e := os.Executable()
		if e != nil {
			return e
		}
		executable, e = filepath.EvalSymlinks(executable)
		if e != nil {
			return e
		}
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		return bootstrap.Supervise(ctx, bootstrap.ServiceOptions{Executable: executable, InstallRoot: filepath.Dir(executable), UpdateRoot: filepath.Join(config.StateDir, "updates"), Args: args})
	}
	return runManagedConfig(config, initialize, child)
}

func runConfig(config masterConfig, initialize bool) error {
	return runManagedConfig(config, initialize, nil)
}

func runManagedConfig(config masterConfig, initialize bool, child *bootstrap.ServiceChild) error {
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
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	var updater *coreupdate.Manager
	var app *master.Server
	if child != nil {
		go child.WatchStop(ctx, cancel)
		updater, err = coreupdate.New(coreupdate.Options{Root: child.UpdateRoot, InstallRoot: child.InstallRoot, Component: "master", CurrentRevision: coreupdate.BuildRevision(), Config: config.Updates, Ready: func(ctx context.Context, p coreupdate.Prepared) error {
			job := updater.Status().Job
			if job == nil {
				return fmt.Errorf("update request is missing")
			}
			if err := app.ValidateCoreUpdatePeers(ctx, p.Manifest); err != nil {
				return err
			}
			if err := app.ValidateCoreUpdateActor(ctx, job.ID); err != nil {
				return err
			}
			if err := bootstrap.RequestActivation(child.UpdateRoot, bootstrap.Activation{Revision: p.Revision, Binary: p.Binary, Directory: p.Directory}); err != nil {
				return err
			}
			time.AfterFunc(300*time.Millisecond, cancel)
			return nil
		}})
		if err != nil {
			return err
		}
		defer updater.Close()
		// Only the default bundled frontend follows the selected release. An
		// independently configured static directory remains operator-managed.
		if config.StaticDir == filepath.Join(child.InstallRoot, "web", "dist") {
			base, e := bootstrap.ExecutableDir()
			if e != nil {
				return e
			}
			config.StaticDir = filepath.Join(base, "web", "dist")
		}
	}
	app = master.New(master.Options{Store: s, Origin: config.Origin, TrustedProxies: proxies,
		TrustAllProxies: !config.VerifyProxyIPs && !config.legacy, AllowHTTP: !config.HTTPS,
		StaticDir: config.StaticDir, ExtensionRoot: config.ExtensionsDir, ExtensionCatalogDir: config.ExtensionsCatalog,
		ExtensionCatalogURL: config.ExtensionsCatalogURL, ExtensionTrustedKeys: trusted, CoreUpdater: updater})
	defer app.Close()
	server := &http.Server{Addr: config.Listen, Handler: app, ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 90 * time.Second, MaxHeaderBytes: 32768, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12}}
	go func() {
		<-ctx.Done()
		shutdown, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		_ = app.Close()
		_ = server.Shutdown(shutdown)
	}()
	slog.Info("management listening", "address", config.Listen, "https", config.HTTPS)
	listener, err := net.Listen("tcp", config.Listen)
	if err != nil {
		return err
	}
	defer listener.Close()
	if config.HTTPS {
		if _, err := tls.LoadX509KeyPair(config.TLSCert, config.TLSKey); err != nil {
			return err
		}
	}
	if child != nil {
		if err := child.ReadyAndWait(ctx); err != nil {
			return err
		}
		if err := updater.ConfirmActivation(); err != nil {
			return err
		}
	}
	if config.HTTPS {
		err = server.ServeTLS(listener, config.TLSCert, config.TLSKey)
	} else {
		err = server.Serve(listener)
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

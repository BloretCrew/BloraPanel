package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"blora.dev/panel/internal/bootstrap"
	"blora.dev/panel/internal/coreupdate"
	"blora.dev/panel/internal/daemon"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/runlog"
	runtimeAPI "blora.dev/panel/internal/runtime"
	"blora.dev/panel/internal/storage"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--core-update-info" {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"formatVersion": 1, "component": "daemon", "version": coreupdate.Version, "revision": coreupdate.BuildRevision(), "protocolVersion": protocol.Version, "schemaVersion": storage.SchemaVersion(), "schemaFingerprint": storage.SchemaFingerprint(), "preserveInstances": true, "managedRestart": true})
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--job-keeper" {
		if err := runtimeAPI.RunJobKeeper(os.Args[2:]); err != nil {
			slog.Error("Job keeper stopped", "error", err)
			os.Exit(1)
		}
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--log-helper" {
		if err := runlog.RunHelper(os.Args[2:]); err != nil {
			slog.Error("log helper stopped", "error", err)
			os.Exit(1)
		}
		return
	}
	if err := run(); err != nil && err != flag.ErrHelp {
		slog.Error("daemon stopped", "error", err)
		os.Exit(1)
	}
}
func run() error {
	child, args, err := bootstrap.ParseServiceChild(os.Args[1:])
	if err != nil {
		return err
	}
	var config daemon.Config
	if child == nil {
		config, err = parseConfig(args, os.Stderr)
	} else {
		config, err = parseConfigAt(args, os.Stderr, child.InstallRoot)
	}
	if err != nil {
		return err
	}
	if child == nil {
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
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	go child.WatchStop(ctx, cancel)
	var d *daemon.Daemon
	var updater *coreupdate.Manager
	updater, err = coreupdate.New(coreupdate.Options{Root: child.UpdateRoot, Component: "daemon", InstallRoot: child.InstallRoot, CurrentRevision: coreupdate.BuildRevision(), Config: config.Updates, Ready: func(ctx context.Context, p coreupdate.Prepared) error {
		job := updater.Status().Job
		if job == nil {
			return errors.New("update request is missing")
		}
		if err := d.ValidateCoreUpdateActor(ctx, job.ID); err != nil {
			return err
		}
		if err := d.PrepareCoreRestart(ctx); err != nil {
			return err
		}
		if err := d.ValidateCoreUpdateActor(ctx, job.ID); err != nil {
			d.AbortCoreRestart()
			return err
		}
		if err := bootstrap.RequestActivation(child.UpdateRoot, bootstrap.Activation{Revision: p.Revision, Binary: p.Binary, Directory: p.Directory}); err != nil {
			d.AbortCoreRestart()
			return err
		}
		time.AfterFunc(300*time.Millisecond, cancel)
		return nil
	}})
	if err != nil {
		return err
	}
	defer updater.Close()
	config.CoreUpdater = updater
	var ready sync.Once
	config.OnReady = func() {
		ready.Do(func() {
			if err := child.ReadyAndWait(ctx); err != nil {
				slog.Error("readiness record failed", "error", err)
				cancel()
				return
			}
			if err := updater.ConfirmActivation(); err != nil {
				slog.Error("update activation record failed", "error", err)
				cancel()
			}
		})
	}
	d, err = daemon.New(config)
	if err != nil {
		return err
	}
	defer d.Close()
	slog.Info("daemon connecting", "nodeId", d.NodeID())
	err = d.Run(ctx)
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

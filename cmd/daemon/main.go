package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"blora.dev/panel/internal/daemon"
	"blora.dev/panel/internal/runlog"
	runtimeAPI "blora.dev/panel/internal/runtime"
)

func main() {
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
	config, err := parseConfig(os.Args[1:], os.Stderr)
	if err != nil {
		return err
	}
	d, err := daemon.New(config)
	if err != nil {
		return err
	}
	defer d.Close()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	slog.Info("daemon connecting", "nodeId", d.NodeID())
	err = d.Run(ctx)
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

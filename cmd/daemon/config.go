package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"blora.dev/panel/internal/bootstrap"
	"blora.dev/panel/internal/daemon"
)

func parseConfig(args []string, output io.Writer) (daemon.Config, error) {
	base, err := bootstrap.ExecutableDir()
	if err != nil {
		return daemon.Config{}, err
	}
	return parseConfigAt(args, output, base)
}

func parseConfigAt(args []string, output io.Writer, base string) (daemon.Config, error) {
	flags := flag.NewFlagSet("blora-daemon", flag.ContinueOnError)
	flags.SetOutput(output)
	path := flags.String("config", filepath.Join(base, "daemon.json"), "private JSON configuration; defaults beside the executable")
	if err := flags.Parse(args); err != nil {
		return daemon.Config{}, err
	}
	if flags.NArg() != 0 {
		return daemon.Config{}, fmt.Errorf("unexpected positional arguments")
	}
	file, err := os.Open(bootstrap.ResolvePath(base, *path))
	if err != nil {
		return daemon.Config{}, fmt.Errorf("open Daemon configuration: %w", err)
	}
	defer file.Close()
	b, err := io.ReadAll(io.LimitReader(file, (1<<20)+1))
	if err != nil {
		return daemon.Config{}, err
	}
	b = bytes.TrimSpace(b)
	if len(b) > 1<<20 || len(b) == 0 || b[0] != '{' {
		return daemon.Config{}, fmt.Errorf("Daemon configuration must be a JSON object no larger than 1 MiB")
	}
	c := daemon.Config{StateDir: "state/daemon", MasterURL: "http://127.0.0.1:37861"}
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&c); err != nil {
		return daemon.Config{}, fmt.Errorf("decode Daemon configuration: %w", err)
	}
	if decoder.Decode(new(any)) != io.EOF {
		return daemon.Config{}, fmt.Errorf("Daemon configuration must contain exactly one JSON object")
	}
	for _, path := range []*string{&c.StateDir, &c.CAFile, &c.EnrollmentFile, &c.RotationFile, &c.CgroupRoot, &c.ComposeTLSCertPath, &c.BackupRoot} {
		*path = bootstrap.ResolvePath(base, *path)
	}
	return c, nil
}

package daemon

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/containers"
	"blora.dev/panel/internal/model"
)

func (d *Daemon) containerLogs(ctx context.Context, r bridge.Request) (any, error) {
	if r.Resource.Kind != "node" || r.Resource.ID != d.identity.NodeID {
		return nil, containers.ErrForbidden
	}
	var args struct {
		ContainerID string `json:"containerId"`
		Tail        int    `json:"tail"`
	}
	if err := json.Unmarshal(r.Args, &args); err != nil {
		return nil, err
	}
	if args.Tail < 1 {
		args.Tail = 200
	}
	if args.Tail > 1000 {
		args.Tail = 1000
	}
	window := model.ContainerLogWindow{ContainerID: args.ContainerID, ObservedAt: time.Now().UTC(), Frames: []model.ContainerLogFrame{}, PossibleGap: true}
	total := 0
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	err := d.containers.Logs(ctx, r.ActorID, containers.Target{Kind: "container", ID: args.ContainerID}, containers.LogOptions{Tail: args.Tail}, func(frame containers.LogFrame) error {
		window.Frames = append(window.Frames, model.ContainerLogFrame{Stream: frame.Stream, Data: frame.Data})
		total += len(frame.Data)
		for total > 48<<10 || len(window.Frames) > 256 {
			total -= len(window.Frames[0].Data)
			window.Frames = window.Frames[1:]
			window.Truncated = true
		}
		return nil
	})
	if err == nil {
		_ = d.persistContainerLog(window)
	}
	return window, err
}

func (d *Daemon) persistContainerLog(window model.ContainerLogWindow) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.config.StateDir == "" || len(window.Frames) == 0 {
		return nil
	}
	root := filepath.Join(d.config.StateDir, "container-logs")
	if err := os.MkdirAll(root, 0700); err != nil {
		return err
	}
	name := safeContainerLogName(window.ContainerID)
	if name == "" {
		return os.ErrInvalid
	}
	path := filepath.Join(root, name+".json")
	var a struct {
		Windows []model.ContainerLogWindow `json:"windows"`
	}
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, &a)
	}
	a.Windows = append(a.Windows, window)
	if len(a.Windows) > 100 {
		a.Windows = a.Windows[len(a.Windows)-100:]
	}
	b, err := json.Marshal(a)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(root, ".log-")
	if err != nil {
		return err
	}
	n := tmp.Name()
	defer os.Remove(n)
	if err = tmp.Chmod(0600); err == nil {
		_, err = tmp.Write(b)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if e := tmp.Close(); err == nil {
		err = e
	}
	if err != nil {
		return err
	}
	if err := os.Rename(n, path); err != nil {
		return err
	}
	// Also persist the directory entry on platforms that support directory
	// sync. Windows guarantees the file flush above, not directory durability.
	return syncContainerLogDirectory(root)
}

func safeContainerLogName(id string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' || r >= 'a' && r <= 'f' {
			return r
		}
		return -1
	}, strings.ToLower(id))
}

func (d *Daemon) containerLogHistory(containerID string, limit int) ([]model.ContainerLogWindow, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	name := safeContainerLogName(containerID)
	if name == "" {
		return nil, os.ErrInvalid
	}
	b, err := os.ReadFile(filepath.Join(d.config.StateDir, "container-logs", name+".json"))
	if err != nil {
		return nil, err
	}
	var a struct {
		Windows []model.ContainerLogWindow `json:"windows"`
	}
	if err = json.Unmarshal(b, &a); err != nil {
		return nil, err
	}
	if len(a.Windows) > limit {
		a.Windows = a.Windows[len(a.Windows)-limit:]
	}
	for i := range a.Windows {
		a.Windows[i].PossibleGap = true
	}
	return a.Windows, nil
}

// purgeContainerLogArchive removes durable windows only after the Engine has
// confirmed deletion. A failed or unknown delete keeps the diagnostic history.
func (d *Daemon) purgeContainerLogArchive(containerID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	name := safeContainerLogName(containerID)
	if name == "" || d.config.StateDir == "" {
		return os.ErrInvalid
	}
	err := os.Remove(filepath.Join(d.config.StateDir, "container-logs", name+".json"))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

package daemon

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/containers"
	"blora.dev/panel/internal/filesystem"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/storage"
	"blora.dev/panel/internal/systeminfo"
	"blora.dev/panel/internal/terminal"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func (d *Daemon) dataLoop(ctx context.Context, channel protocol.Channel, generation uint64) {
	for {
		if ctx.Err() != nil {
			return
		}
		_ = d.connectData(ctx, channel, generation)
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}
func (d *Daemon) connectData(ctx context.Context, channel protocol.Channel, generation uint64) error {
	endpoint := managementWebSocketURL(d.config.MasterURL) + "/api/v1/agent/data/" + channel.String() + "?nodeId=" + url.QueryEscape(d.identity.NodeID)
	ws, _, err := websocket.Dial(ctx, endpoint, &websocket.DialOptions{HTTPClient: d.client})
	if err != nil {
		return err
	}
	defer ws.CloseNow()
	ws.SetReadLimit(16384)
	handshake, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var challenge protocol.Challenge
	if err := wsjson.Read(handshake, ws, &challenge); err != nil {
		return err
	}
	if challenge.NodeID != d.identity.NodeID || challenge.Channel != channel || challenge.Generation != generation {
		return errors.New("data challenge scope mismatch")
	}
	signature, err := protocol.SignChallenge(ed25519.PrivateKey(d.identity.PrivateKey), challenge)
	if err != nil {
		return err
	}
	if err := wsjson.Write(handshake, ws, map[string]any{"signature": signature}); err != nil {
		return err
	}
	var ack struct {
		Generation      uint64 `json:"generation"`
		ProtocolVersion uint32 `json:"protocolVersion"`
	}
	if err := wsjson.Read(handshake, ws, &ack); err != nil {
		return err
	}
	if ack.Generation != generation || ack.ProtocolVersion != protocol.Version {
		return errors.New("invalid data handshake acknowledgement")
	}
	c, err := protocol.NewConn(ctx, ws, protocol.Options{Generation: generation, Channel: channel, InitialStreams: []string{bridge.StreamID}})
	if err != nil {
		return err
	}
	link := bridge.New(ctx, c, func(ctx context.Context, r bridge.Request) (any, error) { return d.handleRPC(ctx, channel, r) })
	defer func() { _ = link.Close(); link.Wait() }()
	d.mu.Lock()
	d.links[channel] = link
	d.mu.Unlock()
	defer func() {
		d.mu.Lock()
		if d.links[channel] == link {
			delete(d.links, channel)
		}
		d.mu.Unlock()
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-link.Done():
		return protocol.ErrClosed
	}
}

// Terminal authorization asks the current Master on every operation, including
// lease renewal. A signed connection is delegation transport, not a stale grant.
func (d *Daemon) authorizeTerminal(ctx context.Context, actor string, resource model.ResourceRef, action string) error {
	if resource.Kind == "node" && (action == "terminal.read" || action == "terminal.input") {
		if resource.ID != d.identity.NodeID || resource.NodeID != resource.ID {
			return terminal.ErrForbidden
		}
		action = "host.manage"
	}
	d.mu.Lock()
	link := d.links[protocol.ChannelInteractive]
	d.mu.Unlock()
	if link == nil {
		return terminal.ErrForbidden
	}
	args, _ := json.Marshal(map[string]string{"action": action})
	var response struct {
		Allowed bool `json:"allowed"`
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := link.Call(ctx, bridge.Request{Method: "authority.check", ActorID: actor, Resource: resource, Args: args}, &response); err != nil {
		return err
	}
	if !response.Allowed {
		return terminal.ErrForbidden
	}
	return nil
}

type fileHandle struct {
	service *filesystem.Service
	root    string
	users   int
}

func (d *Daemon) filesFor(r bridge.Request) (*filesystem.Service, func(), error) {
	if r.Resource.NodeID != d.identity.NodeID || r.Resource.ID == "" {
		return nil, nil, filesystem.ErrPath
	}
	var root, cacheID, state string
	if r.Resource.Kind == "node" {
		if r.Resource.ID != d.identity.NodeID {
			return nil, nil, filesystem.ErrPath
		}
		root = string(os.PathSeparator)
		if volume := filepath.VolumeName(d.config.StateDir); volume != "" {
			root = volume + string(os.PathSeparator)
		}
		cacheID = "node:" + r.Resource.ID
		state = filepath.Join(d.config.StateDir, ".blora-files", "node-files", storage.Hash([]byte(cacheID)))
	} else if r.Resource.Kind == "instance" {
		if strings.ContainsAny(r.Resource.ID, "/\\.") || r.Config == nil {
			return nil, nil, filesystem.ErrPath
		}
		root = filepath.Join(d.config.StateDir, "instances", r.Resource.ID)
		if r.Config.Mode == "native" && r.Config.Directory != "" {
			root = r.Config.Directory
		}
		cacheID = r.Resource.ID
	} else {
		return nil, nil, filesystem.ErrPath
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, nil, err
	}
	if cacheID == r.Resource.ID {
		if err := os.MkdirAll(root, 0700); err != nil {
			return nil, nil, err
		}
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closing {
		return nil, nil, protocol.ErrClosed
	}
	if old := d.files[cacheID]; old != nil && old.root != root {
		if old.users != 0 {
			return nil, nil, filesystem.ErrConflict
		}
		if err := old.service.Close(); err != nil {
			return nil, nil, err
		}
		delete(d.files, cacheID)
	}
	if d.files[cacheID] == nil {
		if r.Resource.Kind == "instance" {
			// Preserve durable trash and upload records for configured instance roots.
			var binding struct {
				Root string `json:"root"`
			}
			_, err := d.store.Record(context.Background(), "file_root", r.Resource.ID, &binding)
			if errors.Is(err, sql.ErrNoRows) {
				binding.Root = root
				_, err = d.store.PutRecord(context.Background(), "file_root", r.Resource.ID, 0, binding)
			}
			if err != nil {
				return nil, nil, err
			}
			state = filepath.Join(d.config.StateDir, "file-state", r.Resource.ID)
			if binding.Root != root {
				state = filepath.Join(d.config.StateDir, "file-state-roots", r.Resource.ID, storage.Hash([]byte(root)))
			}
		}
		service, err := filesystem.New(root, filesystem.Options{StateDir: state, MaxTextBytes: 4 << 20, ChunkBytes: 64 << 10, SkipRootPrivate: r.Resource.Kind == "node"})
		if err != nil {
			return nil, nil, err
		}
		d.files[cacheID] = &fileHandle{service: service, root: root}
	}
	h := d.files[cacheID]
	h.users++
	return h.service, func() { d.mu.Lock(); h.users--; d.mu.Unlock() }, nil
}
func (d *Daemon) handleRPC(ctx context.Context, channel protocol.Channel, r bridge.Request) (any, error) {
	if r.Resource.NodeID != d.identity.NodeID || r.ActorID == "" {
		return nil, &model.APIError{Code: "FORBIDDEN", Message: "请求不属于本节点"}
	}
	if strings.HasPrefix(r.Method, "core.update.") {
		return d.coreUpdateRPC(ctx, channel, r)
	}
	if r.Method == "container.query" && channel == protocol.ChannelBulk {
		value, err := d.containerQuery(ctx, r)
		return value, nodeError(err)
	}
	if r.Method == "system.capabilities" && channel == protocol.ChannelBulk {
		return systeminfo.Detect(), nil
	}
	if r.Method == "system.services" && channel == protocol.ChannelBulk {
		var a struct {
			Limit int    `json:"limit"`
			After string `json:"after"`
		}
		if err := json.Unmarshal(r.Args, &a); err != nil {
			return nil, nodeError(err)
		}
		items, err := systeminfo.ListServicePage(ctx, a.Limit, a.After)
		return items, nodeError(err)
	}
	if r.Method == "system.firewall" && channel == protocol.ChannelBulk {
		return systeminfo.Firewall(ctx)
	}
	if r.Method == "system.tasks" && channel == protocol.ChannelBulk {
		var a struct {
			Limit int    `json:"limit"`
			After string `json:"after"`
		}
		if err := json.Unmarshal(r.Args, &a); err != nil {
			return nil, nodeError(err)
		}
		items, err := systeminfo.ListScheduledTaskPage(ctx, a.Limit, a.After)
		return items, nodeError(err)
	}
	if r.Method == "system.firewall.preview" && channel == protocol.ChannelBulk {
		var a struct {
			Desired []string `json:"desired"`
		}
		if err := json.Unmarshal(r.Args, &a); err != nil {
			return nil, nodeError(err)
		}
		return systeminfo.PreviewFirewall(ctx, a.Desired)
	}
	if r.Method == "system.firewall.confirm" && channel == protocol.ChannelBulk {
		var a struct {
			TaskID      string `json:"taskId"`
			TaskActorID string `json:"taskActorId"`
			RequestID   string `json:"requestId"`
		}
		if err := json.Unmarshal(r.Args, &a); err != nil {
			return nil, nodeError(err)
		}
		if err := d.authorizeTerminal(ctx, r.ActorID, model.ResourceRef{Kind: "node", ID: d.identity.NodeID, NodeID: d.identity.NodeID}, "host.manage"); err != nil {
			return nil, nodeError(err)
		}
		value, err := d.confirmFirewallLease(ctx, a.TaskID, a.TaskActorID, a.RequestID)
		return value, nodeError(err)
	}
	if r.Method == "system.task.action" && channel == protocol.ChannelBulk {
		var a struct {
			Name   string `json:"name"`
			Action string `json:"action"`
		}
		if err := json.Unmarshal(r.Args, &a); err != nil {
			return nil, nodeError(err)
		}
		if err := d.authorizeTerminal(ctx, r.ActorID, r.Resource, "host.manage"); err != nil {
			return nil, nodeError(err)
		}
		if err := systeminfo.TaskAction(ctx, a.Name, a.Action); err != nil {
			return nil, nodeError(err)
		}
		return map[string]any{"name": a.Name, "action": a.Action}, nil
	}
	if r.Method == "container.exec-target" && channel == protocol.ChannelInteractive {
		value, err := d.hostTerminalTarget(ctx, r)
		return value, nodeError(err)
	}
	if strings.HasPrefix(r.Method, "backup.") && channel == protocol.ChannelBulk {
		value, err := d.backupRPC(ctx, r)
		return value, nodeError(err)
	}
	if r.Method == "container.logs" && channel == protocol.ChannelInteractive {
		value, err := d.containerLogs(ctx, r)
		return value, nodeError(err)
	}
	if r.Method == "container.logs.history" && channel == protocol.ChannelBulk {
		var a struct {
			ContainerID string `json:"containerId"`
			Limit       int    `json:"limit"`
		}
		if err := json.Unmarshal(r.Args, &a); err != nil {
			return nil, nodeError(err)
		}
		items, err := d.containerLogHistory(a.ContainerID, a.Limit)
		return map[string]any{"items": items, "retention": 100, "possibleGap": true}, nodeError(err)
	}
	if strings.HasPrefix(r.Method, "monitor.") && channel == protocol.ChannelBulk {
		value, err := d.monitorRPC(ctx, r)
		return value, nodeError(err)
	}
	if strings.HasPrefix(r.Method, "container.project.") && channel == protocol.ChannelBulk {
		value, err := d.composeRPC(ctx, r)
		return value, nodeError(err)
	}
	if strings.HasPrefix(r.Method, "file.") && channel == protocol.ChannelBulk {
		value, err := d.fileRPC(ctx, r)
		return value, nodeError(err)
	}
	if strings.HasPrefix(r.Method, "terminal.") && channel == protocol.ChannelInteractive {
		value, err := d.terminalRPC(ctx, r)
		return value, nodeError(err)
	}
	if strings.HasPrefix(r.Method, "log.") && channel == protocol.ChannelInteractive {
		value, err := d.logRPC(ctx, r)
		return value, nodeError(err)
	}
	return nil, &model.APIError{Code: "CAPABILITY_UNAVAILABLE", Message: "此通道未开放请求的能力"}
}
func nodeError(err error) error {
	if err == nil {
		return nil
	}
	code := "NODE_OPERATION_FAILED"
	switch {
	case errors.Is(err, errFirewallLeaseConflict):
		code = "FIREWALL_LEASE_CONFLICT"
	case errors.Is(err, containers.ErrForbidden):
		code = "FORBIDDEN"
	case errors.Is(err, containers.ErrIdentity), errors.Is(err, containers.ErrConflict):
		code = "CONTAINER_CONFLICT"
	case errors.Is(err, containers.ErrCapability):
		code = "CAPABILITY_UNAVAILABLE"
	case errors.Is(err, filesystem.ErrTransfer):
		code = "FILE_TRANSFER_MISMATCH"
	case errors.Is(err, filesystem.ErrUnsupported):
		code = "FILE_UNSUPPORTED"
	case errors.Is(err, filesystem.ErrConflict):
		code = "FILE_CONFLICT"
	case errors.Is(err, filesystem.ErrPath), errors.Is(err, terminal.ErrForbidden):
		code = "FORBIDDEN"
	case errors.Is(err, filesystem.ErrLimit), errors.Is(err, terminal.ErrLimit):
		code = "RESOURCE_LIMIT"
	case errors.Is(err, terminal.ErrLease):
		code = "LEASE_CONFLICT"
	case errors.Is(err, terminal.ErrInactive):
		code = "SESSION_INACTIVE"
	case errors.Is(err, os.ErrNotExist), errors.Is(err, terminal.ErrNotFound), errors.Is(err, sql.ErrNoRows):
		code = "NOT_FOUND"
	}
	return &model.APIError{Code: code, Message: err.Error()}
}
func (d *Daemon) fileRPC(ctx context.Context, r bridge.Request) (any, error) {
	service, release, err := d.filesFor(r)
	if err != nil {
		return nil, err
	}
	defer release()
	var args struct {
		Path    string                `json:"path"`
		Version string                `json:"version"`
		Offset  int64                 `json:"offset"`
		Limit   int                   `json:"limit"`
		Search  string                `json:"search"`
		Sort    string                `json:"sort"`
		Order   string                `json:"order"`
		ID      string                `json:"id"`
		Data    []byte                `json:"data"`
		Hash    string                `json:"hash"`
		Spec    filesystem.UploadSpec `json:"spec"`
	}
	if err := json.Unmarshal(r.Args, &args); err != nil {
		return nil, err
	}
	switch r.Method {
	case "file.list":
		return service.List(ctx, args.Path, filesystem.ListOptions{Offset: int(args.Offset), Limit: args.Limit, Version: args.Version, Search: args.Search, Sort: args.Sort, Order: args.Order})
	case "file.trash":
		return service.ListTrash(ctx, int(args.Offset), args.Limit)
	case "file.stat":
		return service.Stat(ctx, args.Path)
	case "file.relation":
		return service.RelationFacts(ctx, args.Path)
	case "file.chunk":
		return service.ReadChunk(ctx, args.Path, args.Version, args.Offset, args.Limit)
	case "file.transfer.stat":
		return service.TransferStat(ctx, args.Path, args.Version)
	case "file.transfer.chunk":
		return service.ReadTransferChunk(ctx, args.Path, args.Version, args.Offset, args.Limit)
	case "file.upload.begin":
		if args.Spec.OwnerID != r.ActorID {
			return nil, filesystem.ErrTransfer
		}
		return service.BeginUpload(ctx, args.Spec)
	case "file.upload.status", "file.upload.chunk", "file.upload.cancel":
		status, err := service.UploadStatus(ctx, args.ID)
		if err != nil {
			return nil, err
		}
		if status.Spec.OwnerID != r.ActorID {
			return nil, filesystem.ErrTransfer
		}
		if r.Method == "file.upload.status" {
			return status, nil
		}
		if r.Method == "file.upload.chunk" {
			return service.UploadChunk(ctx, args.ID, args.Offset, args.Data, args.Hash)
		}
		return service.CancelUpload(ctx, args.ID)
	default:
		return nil, errors.New("unsupported file RPC")
	}
}
func (d *Daemon) terminalRPC(ctx context.Context, r bridge.Request) (any, error) {
	var args struct {
		SessionID string `json:"sessionId"`
		ViewID    string `json:"viewId"`
		After     uint64 `json:"after"`
		MaxBytes  int    `json:"maxBytes"`
		Cols      uint16 `json:"cols"`
		Rows      uint16 `json:"rows"`
		Data      []byte `json:"data"`
		Takeover  bool   `json:"takeover"`
		UserID    string `json:"userId"`
	}
	if err := json.Unmarshal(r.Args, &args); err != nil {
		return nil, err
	}
	if r.Method == "terminal.list" {
		items, err := d.terminals.List(ctx, r.ActorID)
		if err != nil {
			return nil, err
		}
		out := []terminal.Session{}
		for _, s := range items {
			if s.Resource.Key() == r.Resource.Key() && (s.Backend != "docker-host" || s.OwnerID == r.ActorID) {
				out = append(out, s)
			}
		}
		return map[string]any{"items": out}, nil
	}
	session, err := d.terminals.Get(ctx, args.SessionID, r.ActorID)
	if err != nil {
		return nil, err
	}
	if session.Resource.Key() != r.Resource.Key() || (session.Backend == "docker-host" && session.OwnerID != r.ActorID) {
		return nil, terminal.ErrForbidden
	}
	switch r.Method {
	case "terminal.get":
		return session, nil
	case "terminal.attach":
		return d.terminals.Attach(ctx, args.SessionID, r.ActorID, args.ViewID, args.After, args.MaxBytes)
	case "terminal.read":
		return d.terminals.Read(ctx, args.SessionID, r.ActorID, args.ViewID, args.After, args.MaxBytes)
	case "terminal.detach":
		return nil, d.terminals.Detach(args.SessionID, r.ActorID, args.ViewID)
	case "terminal.lease":
		return d.terminals.AcquireLease(args.SessionID, r.ActorID, args.ViewID, 30*time.Second, args.Takeover)
	case "terminal.input":
		return nil, d.terminals.WriteInput(ctx, args.SessionID, r.ActorID, args.ViewID, args.Data)
	case "terminal.resize":
		return nil, d.terminals.Resize(ctx, args.SessionID, r.ActorID, args.ViewID, args.Cols, args.Rows)
	default:
		return nil, errors.New("unsupported terminal RPC")
	}
}

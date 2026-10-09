package daemon

import (
	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/coreupdate"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"context"
	"encoding/json"
	"errors"
	"strings"
)

func (d *Daemon) coreUpdateRPC(ctx context.Context, channel protocol.Channel, r bridge.Request) (any, error) {
	ref := model.ResourceRef{Kind: "node", ID: d.identity.NodeID, NodeID: d.identity.NodeID}
	if channel != protocol.ChannelBulk || r.Resource != ref {
		return nil, nodeError(errors.New("invalid core update scope"))
	}
	if err := d.authorizeTerminal(ctx, r.ActorID, ref, "core.update"); err != nil {
		return nil, nodeError(err)
	}
	m := d.config.CoreUpdater
	if m == nil {
		return nil, &model.APIError{Code: "UPDATE_UNAVAILABLE", Message: "此启动方式不支持后台更新；请使用配置文件零参数启动"}
	}
	switch r.Method {
	case "core.update.status":
		return m.Status(), nil
	case "core.update.check":
		return m.BeginCheck()
	case "core.update.source":
		var c coreupdate.Config
		if err := json.Unmarshal(r.Args, &c); err != nil {
			return nil, nodeError(err)
		}
		return m.SetSource(ctx, c)
	case "core.update.apply":
		var args struct {
			Revision     string `json:"revision"`
			RequestID    string `json:"requestId"`
			PeerProtocol uint32 `json:"peerProtocol"`
		}
		if err := json.Unmarshal(r.Args, &args); err != nil {
			return nil, nodeError(err)
		}
		requestID := r.ActorID + ":" + args.RequestID
		if old, found, err := m.Lookup(requestID, args.Revision); err != nil {
			return nil, nodeError(err)
		} else if found {
			return old, nil
		}
		status := m.Status()
		if status.Preview == nil || status.Preview.Revision != args.Revision || !status.Preview.Manifest.SupportsPeer(args.PeerProtocol) || args.PeerProtocol != protocol.Version {
			return nil, &model.APIError{Code: "UPDATE_INCOMPATIBLE", Message: "更新与当前 Master 不兼容，或版本预览已失效"}
		}
		return m.Start(requestID, args.Revision)
	default:
		return nil, nodeError(errors.New("unknown core update action"))
	}
}

// ValidateCoreUpdateActor asks the connected Master immediately before the
// resource-preserving handoff. A missing authority connection vetoes updates.
func (d *Daemon) ValidateCoreUpdateActor(ctx context.Context, requestID string) error {
	actor, _, ok := strings.Cut(requestID, ":")
	if !ok || actor == "" {
		return errors.New("update request has no account identity")
	}
	ref := model.ResourceRef{Kind: "node", ID: d.identity.NodeID, NodeID: d.identity.NodeID}
	return d.authorizeTerminal(ctx, actor, ref, "core.update")
}

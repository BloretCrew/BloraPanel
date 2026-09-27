package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/extensions"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
)

func (d *Daemon) runExtension(ctx context.Context, t model.Task) (any, error) {
	var p model.ExtensionTaskPayload
	if err := json.Unmarshal(t.Payload, &p); err != nil || p.ExtensionID == "" || p.PackageHash == "" || p.ModuleHash == "" {
		return nil, errors.New("invalid extension task identity")
	}
	acquireCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	d.mu.Lock()
	link := d.links[protocol.ChannelBulk]
	d.mu.Unlock()
	if link == nil {
		return nil, errors.New("extension module channel unavailable")
	}
	var module []byte
	total := 0
	for {
		args, _ := json.Marshal(map[string]any{"taskId": t.ID, "requestId": t.RequestID, "digest": t.Digest, "offset": len(module)})
		var part struct {
			Data  []byte `json:"data"`
			Total int    `json:"total"`
		}
		if err := link.Call(acquireCtx, bridge.Request{Method: "extension.module", ActorID: t.ActorID, Resource: t.Resource, Args: args}, &part); err != nil {
			var apiErr *model.APIError
			if errors.As(err, &apiErr) && apiErr.Code == "TASK_CANCELLED" {
				return nil, context.Canceled
			}
			return nil, err
		}
		if part.Total <= 0 || part.Total > extensions.MaxBackendModule || (total != 0 && total != part.Total) || len(part.Data) == 0 || len(part.Data) > 32*1024 || len(module)+len(part.Data) > part.Total {
			return nil, errors.New("invalid extension module chunk")
		}
		total = part.Total
		module = append(module, part.Data...)
		if len(module) == total {
			break
		}
	}
	if extensions.ModuleHash(module) != p.ModuleHash {
		return nil, errors.New("extension module digest mismatch")
	}
	cancel()
	if err := d.authorizeExecution(ctx, t); err != nil {
		return nil, err
	}
	result, err := extensions.RunBackend(ctx, module, p.Payload)
	if err != nil {
		return nil, err
	}
	return map[string]any{"extensionId": p.ExtensionID, "packageHash": p.PackageHash, "result": result}, nil
}

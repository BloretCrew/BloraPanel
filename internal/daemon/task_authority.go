package daemon

import (
	"context"
	"encoding/json"
	"time"

	"blora.dev/panel/internal/bridge"
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
	"blora.dev/panel/internal/terminal"
)

func (d *Daemon) authorizeExecution(ctx context.Context, t model.Task) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var link *bridge.Link
	for link == nil {
		d.mu.Lock()
		link = d.links[protocol.ChannelInteractive]
		d.mu.Unlock()
		if link != nil {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(20 * time.Millisecond):
		}
	}
	args, _ := json.Marshal(map[string]string{"taskId": t.ID, "requestId": t.RequestID, "digest": t.Digest})
	var result struct {
		Allowed   bool `json:"allowed"`
		Cancelled bool `json:"cancelled"`
	}
	if err := link.Call(ctx, bridge.Request{Method: "authority.task", ActorID: t.ActorID, Resource: t.Resource, Args: args}, &result); err != nil {
		return err
	}
	if !result.Allowed {
		if result.Cancelled {
			return context.Canceled
		}
		return terminal.ErrForbidden
	}
	return nil
}

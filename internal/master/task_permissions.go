package master

import (
	"context"
	"encoding/json"
	"strings"

	"blora.dev/panel/internal/model"
)

func taskPermission(action string) string {
	if strings.HasPrefix(action, "process.") {
		return "host.manage"
	}
	if strings.HasPrefix(action, "system.service.") {
		return "host.manage"
	}
	if strings.HasPrefix(action, "system.task.") {
		return "host.manage"
	}
	if action == "system.firewall.apply" {
		return "host.manage"
	}
	if action == "backup.prune" {
		return "backup.create"
	}
	if model.IsContainerAction(action) {
		return "host.manage"
	}
	if action == "console.input" {
		return "terminal.input"
	}
	if strings.HasPrefix(action, "file.") || isLocalTask(action) {
		return "file.write"
	}
	if strings.HasPrefix(action, "terminal.") {
		return "terminal.input"
	}
	return action
}
func (s *Server) authorizeTask(ctx context.Context, u model.User, t model.Task) bool {
	if t.Action == "extension.task" {
		var p struct {
			ExtensionID string `json:"extensionId"`
		}
		if s.extensions == nil || json.Unmarshal(t.Payload, &p) != nil || p.ExtensionID == "" {
			return false
		}
		if err := s.extensions.AuthorizeCapability(p.ExtensionID, "task.create"); err != nil {
			return false
		}
		return s.store.Allowed(ctx, u, model.ResourceRef{Kind: "extension", ID: p.ExtensionID}, "app.use") && s.store.Allowed(ctx, u, t.Resource, "node.read")
	}
	if t.Resource.Kind == "node" && (t.Action == "terminal.create" || t.Action == "terminal.close") {
		return t.Resource.ID == t.Resource.NodeID && s.store.Allowed(ctx, u, t.Resource, "host.manage")
	}
	if !s.authorizeTransferChild(ctx, u, t) {
		return false
	}
	if !s.store.Allowed(ctx, u, t.Resource, taskPermission(t.Action)) {
		return false
	}
	switch t.Action {
	case "backup.create", "backup.prune":
		return s.store.Allowed(ctx, u, t.Resource, "file.read")
	case "backup.restore":
		return s.store.Allowed(ctx, u, t.Resource, "file.write")
	case "file.copy", "file.move", "file.compress", "file.extract":
		if !s.store.Allowed(ctx, u, t.Resource, "file.read") {
			return false
		}
	}
	if t.Action == "terminal.create" {
		var p model.TerminalTaskPayload
		if err := json.Unmarshal(t.Payload, &p); err != nil {
			return false
		}
		if p.Config.Mode != "container" {
			return s.store.Allowed(ctx, u, model.ResourceRef{Kind: "node", ID: t.Resource.NodeID}, "host.manage")
		}
	}
	return true
}

package containers

import (
	"context"
	"encoding/json"
	"fmt"
)

type ExecTarget struct {
	Target    Target `json:"target"`
	StartedAt string `json:"startedAt"`
}

// InspectExecTarget resolves the current birth before a daemon records a host
// terminal intent. Runtime rechecks the same complete identity at execution.
func (m *Manager) InspectExecTarget(ctx context.Context, actor string, target Target) (ExecTarget, error) {
	if err := m.authorize(ctx, actor); err != nil {
		return ExecTarget{}, err
	}
	if err := m.requireEngine(); err != nil {
		return ExecTarget{}, err
	}
	if err := validateTarget(target, "container"); err != nil {
		return ExecTarget{}, err
	}
	o, err := m.engine.inspect(ctx, target, true)
	if err != nil {
		return ExecTarget{}, err
	}
	var container containerInspect
	if err = json.Unmarshal(o.Details, &container); err != nil {
		return ExecTarget{}, err
	}
	if !container.State.Running || container.State.Paused || container.State.Restarting || container.State.Dead || container.Created == "" || container.State.StartedAt == "" {
		return ExecTarget{}, fmt.Errorf("container is not ready for exec: %w", ErrCapability)
	}
	if container.Config.Labels["dev.blora.run"] != "" && container.Config.Labels["dev.blora.instance"] != "" {
		return ExecTarget{}, fmt.Errorf("platform-managed container requires its instance terminal: %w", ErrConflict)
	}
	return ExecTarget{Target: o.Target, StartedAt: container.State.StartedAt}, nil
}

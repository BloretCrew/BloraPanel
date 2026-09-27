//go:build !linux && !windows

package terminal

import "context"

type nativeBackend struct{}

func (nativeBackend) Spawn(context.Context, SpawnSpec) (Process, error) { return nil, ErrCapability }

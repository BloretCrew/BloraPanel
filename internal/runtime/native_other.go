//go:build !linux && !windows

package runtime

import (
	"blora.dev/panel/internal/model"
	"context"
	"os"
)

func nativeBackend() string { return "unsupported" }
func syncDirectory(path string) error {
	f, e := os.Open(path)
	if e != nil {
		return e
	}
	defer f.Close()
	return f.Sync()
}
func prepareContainerDirectory(path string, _, _ uint32) error {
	_, e := newContainerDirectory(path)
	return e
}
func (m *Manager) startNative(_ context.Context, r Record, _ model.InstanceConfig, _ IO) (Record, error) {
	return r, ErrCapability
}
func (m *Manager) observeNative(context.Context, Record) (Observation, error) {
	return Observation{State: "UNKNOWN"}, ErrCapability
}
func (m *Manager) signalNative(context.Context, Record, bool) error { return ErrCapability }
func (m *Manager) cleanupNative(Record) error                       { return ErrCapability }

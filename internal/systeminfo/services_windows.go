//go:build windows

package systeminfo

import (
	"context"
	"fmt"
	"sort"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

func windowsServiceAction(ctx context.Context, name, action string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return err
	}
	h, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return err
	}
	defer windows.CloseServiceHandle(h)
	key, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return err
	}
	access := uint32(windows.SERVICE_QUERY_STATUS)
	if action != "stop" {
		access |= windows.SERVICE_START
	}
	if action != "start" {
		access |= windows.SERVICE_STOP
	}
	handle, err := windows.OpenService(h, key, access)
	if err != nil {
		return err
	}
	s := mgr.Service{Name: name, Handle: handle}
	defer s.Close()
	wait := func(target svc.State) error {
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			status, err := s.Query()
			if err != nil {
				return err
			}
			if status.State == target {
				return nil
			}
			if target == svc.Running && status.State == svc.Stopped {
				return fmt.Errorf("service stopped before reaching running: exit code %d", status.Win32ExitCode)
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
			}
		}
	}
	if action == "stop" || action == "restart" {
		status, err := s.Query()
		if err != nil {
			return err
		}
		if status.State != svc.Stopped && status.State != svc.StopPending {
			if err := ctx.Err(); err != nil {
				return err
			}
			if _, err := s.Control(svc.Stop); err != nil {
				return err
			}
		}
		if err := wait(svc.Stopped); err != nil {
			return err
		}
	}
	if action == "start" || action == "restart" {
		if err := ctx.Err(); err != nil {
			return err
		}
		status, err := s.Query()
		if err != nil {
			return err
		}
		if status.State != svc.Running && status.State != svc.StartPending {
			if err := s.Start(); err != nil {
				return err
			}
		}
		return wait(svc.Running)
	}
	return nil
}

func listWindowsServices(ctx context.Context, limit int, after string) ([]Service, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	h, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT|windows.SC_MANAGER_ENUMERATE_SERVICE)
	if err != nil {
		return nil, err
	}
	m := mgr.Mgr{Handle: h}
	defer m.Disconnect()
	names, err := m.ListServices()
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	out := make([]Service, 0, limit)
	for _, name := range names {
		if name <= after {
			continue
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		p := Service{Name: name, State: "unavailable"}
		key, err := windows.UTF16PtrFromString(name)
		if err != nil {
			continue
		}
		handle, err := windows.OpenService(h, key, windows.SERVICE_QUERY_STATUS)
		if err == nil {
			s := mgr.Service{Name: name, Handle: handle}
			status, err := s.Query()
			s.Close()
			if err == nil {
				states := map[svc.State]string{svc.Stopped: "stopped", svc.StartPending: "start_pending", svc.StopPending: "stop_pending", svc.Running: "running", svc.ContinuePending: "continue_pending", svc.PausePending: "pause_pending", svc.Paused: "paused"}
				p.State = states[status.State]
				if p.State == "" {
					p.State = "unknown"
				}
			}
		}
		out = append(out, p)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

//go:build !windows

package systeminfo

import (
	"context"
	"errors"
)

func listWindowsServices(context.Context, int, string) ([]Service, error) {
	return nil, errors.New("Windows service manager unavailable")
}

func windowsServiceAction(context.Context, string, string) error {
	return errors.New("Windows service manager unavailable")
}

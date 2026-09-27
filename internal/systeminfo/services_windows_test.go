//go:build windows

package systeminfo

import (
	"context"
	"testing"
	"time"
)

// Read-only real-machine entry; never starts, stops or installs a service.
func TestWindowsNativeServiceEnumeration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	items, err := ListServices(ctx, 200)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) == 0 {
		t.Fatal("SCM returned no services")
	}
	known := false
	for _, item := range items {
		if item.Name == "" || item.State == "" {
			t.Fatalf("invalid service %+v", item)
		}
		if item.State != "unavailable" && item.State != "unknown" {
			known = true
		}
	}
	if !known {
		t.Fatal("no actual service status available")
	}
}

func TestWindowsTaskSchedulerEnumeration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	items, err := ListScheduledTasks(ctx, 200)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) == 0 {
		t.Fatal("no scheduled tasks available for real-machine validation")
	}
	for _, item := range items {
		if !ValidTaskName(item.Name) || item.State == "" {
			t.Fatalf("invalid scheduler row %+v", item)
		}
	}
}

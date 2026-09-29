package systeminfo

import (
	"context"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func systemdOutput(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "systemctl", args...)
	cmd.Env = append(os.Environ(), "LC_ALL=C", "SYSTEMD_COLORS=0")
	return cmd.Output()
}

func systemdUnitNames(ctx context.Context, kind string, limit int, after string) ([]string, error) {
	// list-units alone loses stopped services when the manager garbage-collects
	// their loaded unit. Include installed units as well as transient instances.
	names := map[string]bool{}
	for _, args := range [][]string{
		{"list-units", "--type=" + kind, "--all", "--no-legend", "--no-pager", "--plain"},
		{"list-unit-files", "--type=" + kind, "--no-legend", "--no-pager", "--plain"},
	} {
		output, err := systemdOutput(ctx, args...)
		if err != nil {
			return nil, err
		}
		for _, line := range strings.Split(string(output), "\n") {
			fields := strings.Fields(line)
			if len(fields) == 0 {
				continue
			}
			name := fields[0]
			// A template without an instance is not a concrete runnable service.
			if strings.HasSuffix(name, "."+kind) && !strings.HasSuffix(name, "@."+kind) && name > after {
				names[name] = true
			}
		}
	}
	ordered := make([]string, 0, len(names))
	for name := range names {
		ordered = append(ordered, name)
	}
	sort.Strings(ordered)
	if len(ordered) > limit {
		ordered = ordered[:limit]
	}
	return ordered, nil
}

func systemdUnitProperties(ctx context.Context, ordered []string, properties string) (map[string]map[string]string, error) {
	details := map[string]map[string]string{}
	if len(ordered) == 0 {
		return details, nil
	}
	// Resolve this page in one manager query. Do not infer runtime state from
	// enablement, masking, or absence in the earlier loaded-unit snapshot.
	args := append([]string{"show", "--no-pager", "--property=Id,Names," + properties, "--"}, ordered...)
	output, err := systemdOutput(ctx, args...)
	if err != nil {
		return nil, err
	}
	for _, block := range strings.Split(strings.TrimSpace(string(output)), "\n\n") {
		properties := map[string]string{}
		for _, line := range strings.Split(block, "\n") {
			if key, value, ok := strings.Cut(line, "="); ok {
				properties[key] = value
			}
		}
		for _, name := range append(strings.Fields(properties["Names"]), properties["Id"]) {
			details[name] = properties
		}
	}
	return details, nil
}

func listSystemdServicesAfter(ctx context.Context, limit int, after string) ([]Service, error) {
	ordered, err := systemdUnitNames(ctx, "service", limit, after)
	if err != nil {
		return nil, err
	}
	details, err := systemdUnitProperties(ctx, ordered, "ActiveState,LoadState,Description")
	if err != nil {
		return nil, err
	}
	items := make([]Service, 0, len(ordered))
	for _, name := range ordered {
		properties := details[name]
		state := properties["ActiveState"]
		if properties["LoadState"] == "not-found" {
			state = "not-found"
		}
		if state == "" {
			state = "unknown"
		}
		items = append(items, Service{Name: name, State: state, Description: properties["Description"]})
	}
	return items, nil
}

func listSystemdTimersAfter(ctx context.Context, limit int, after string) ([]ScheduledTask, error) {
	ordered, err := systemdUnitNames(ctx, "timer", limit, after)
	if err != nil {
		return nil, err
	}
	details, err := systemdUnitProperties(ctx, ordered, "UnitFileState,LoadState,TimersCalendar,TimersMonotonic")
	if err != nil {
		return nil, err
	}
	output, err := systemdOutput(ctx, "list-timers", "--all", "--no-legend", "--no-pager")
	if err != nil {
		return nil, err
	}
	schedules := map[string]string{}
	for _, item := range parseLinuxTimers(string(output), 0) {
		schedules[item.Name] = item.Schedule
	}
	items := make([]ScheduledTask, 0, len(ordered))
	for _, name := range ordered {
		properties := details[name]
		state := properties["UnitFileState"]
		if properties["LoadState"] == "not-found" {
			state = "not-found"
		}
		if state == "" {
			state = "unknown"
		}
		schedule := schedules[name]
		if schedule == "" {
			schedule = strings.TrimSpace(properties["TimersCalendar"] + " " + properties["TimersMonotonic"])
		}
		items = append(items, ScheduledTask{Name: name, State: state, Schedule: schedule})
	}
	return items, nil
}

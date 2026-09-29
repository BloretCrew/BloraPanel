package systeminfo

import (
	"context"
	"errors"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"unicode"
)

var taskNamePattern = regexp.MustCompile(`^[A-Za-z0-9_.@:/\\-]{1,256}$`)

func ValidTaskName(name string) bool {
	if strings.HasPrefix(name, `\`) {
		if len(name) > 1024 || strings.ContainsAny(name, "/\"<>|*?:") {
			return false
		}
		for _, r := range name {
			if unicode.IsControl(r) {
				return false
			}
		}
		for _, part := range strings.Split(name[1:], `\`) {
			if part == "" || part == "." || part == ".." {
				return false
			}
		}
		return true
	}
	return taskNamePattern.MatchString(name) && !strings.Contains(name, "..")
}

// TaskAction changes only scheduling enablement for one explicitly named task.
func TaskAction(ctx context.Context, name, action string) error {
	if !ValidTaskName(name) {
		return errors.New("invalid task name")
	}
	if action != "enable" && action != "disable" {
		return errors.New("unsupported task action")
	}
	switch runtime.GOOS {
	case "linux":
		if !serviceNamePattern.MatchString(name) || !strings.HasSuffix(name, ".timer") {
			return errors.New("scheduled task must identify a timer unit")
		}
		return exec.CommandContext(ctx, "systemctl", action, "--", name).Run()
	case "windows":
		return windowsTaskAction(ctx, name, action)
	case "darwin":
		return exec.CommandContext(ctx, "launchctl", action, name).Run()
	default:
		return errors.New("task backend unavailable")
	}
}

type ScheduledTask struct {
	Name     string `json:"name"`
	State    string `json:"state"`
	Schedule string `json:"schedule,omitempty"`
}

func ListScheduledTasks(ctx context.Context, limit int) ([]ScheduledTask, error) {
	p, err := ListScheduledTaskPage(ctx, limit, "")
	return p.Items, err
}

type ScheduledTaskPage struct {
	Items     []ScheduledTask `json:"items"`
	NextAfter string          `json:"nextAfter,omitempty"`
}

func ListScheduledTaskPage(ctx context.Context, limit int, after string) (ScheduledTaskPage, error) {
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		return ScheduledTaskPage{}, errors.New("task limit exceeded")
	}
	if after != "" && !ValidTaskName(after) {
		return ScheduledTaskPage{}, errors.New("invalid task cursor")
	}
	items, err := listScheduledTasksAfter(ctx, limit+1, after)
	if err != nil {
		return ScheduledTaskPage{}, err
	}
	p := ScheduledTaskPage{Items: items}
	if len(items) > limit {
		p.Items = items[:limit]
		p.NextAfter = p.Items[limit-1].Name
	}
	return p, nil
}
func listScheduledTasksAfter(ctx context.Context, limit int, after string) ([]ScheduledTask, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		return listSystemdTimersAfter(ctx, limit, after)
	case "windows":
		return listWindowsTasks(ctx, limit, after)
	case "darwin":
		cmd = exec.CommandContext(ctx, "launchctl", "list")
	default:
		return nil, errors.New("task backend unavailable")
	}
	b, e := cmd.Output()
	if e != nil {
		return nil, e
	}
	out := []ScheduledTask{}
	for _, line := range strings.Split(string(b), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		x := ScheduledTask{Name: f[0], State: "reported"}
		out = append(out, x)
	}
	return taskRowsAfter(out, limit, after), nil
}

func taskRowsAfter(rows []ScheduledTask, limit int, after string) []ScheduledTask {
	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	out := []ScheduledTask{}
	for _, row := range rows {
		if row.Name > after {
			out = append(out, row)
			if len(out) >= limit {
				break
			}
		}
	}
	return out
}

func parseLinuxTimers(output string, limit int) []ScheduledTask {
	out := []ScheduledTask{}
	for _, line := range strings.Split(output, "\n") {
		f := strings.Fields(line)
		// list-timers ends with UNIT and ACTIVATES. Only UNIT is a timer
		// identity; timestamps before it have a variable number of fields.
		if len(f) < 3 || !strings.HasSuffix(f[len(f)-2], ".timer") {
			continue
		}
		out = append(out, ScheduledTask{Name: f[len(f)-2], State: "reported", Schedule: strings.Join(f[:len(f)-2], " ")})
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

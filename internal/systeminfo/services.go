package systeminfo

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"runtime"
	"sort"
	"strings"
)

type Service struct {
	Name        string `json:"name"`
	State       string `json:"state"`
	Description string `json:"description,omitempty"`
}

func ListServices(ctx context.Context, limit int) ([]Service, error) {
	p, err := ListServicePage(ctx, limit, "")
	return p.Items, err
}

type ServicePage struct {
	Items     []Service `json:"items"`
	NextAfter string    `json:"nextAfter,omitempty"`
}

func ListServicePage(ctx context.Context, limit int, after string) (ServicePage, error) {
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		return ServicePage{}, errors.New("service limit exceeded")
	}
	if after != "" && !ValidServiceName(after) {
		return ServicePage{}, errors.New("invalid service cursor")
	}
	items, err := listServicesAfter(ctx, limit+1, after)
	if err != nil {
		return ServicePage{}, err
	}
	p := ServicePage{Items: items}
	if len(items) > limit {
		p.Items = items[:limit]
		p.NextAfter = p.Items[limit-1].Name
	}
	return p, nil
}
func listServicesAfter(ctx context.Context, limit int, after string) ([]Service, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		return listSystemdServicesAfter(ctx, limit, after)
	case "windows":
		return listWindowsServices(ctx, limit, after)
	case "darwin":
		cmd = exec.CommandContext(ctx, "launchctl", "list")
	default:
		return nil, errors.New("service backend unavailable")
	}
	b, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	out := []Service{}
	for _, line := range strings.Split(string(b), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		s := Service{Name: f[0], State: "unknown"}
		if runtime.GOOS == "linux" && len(f) >= 4 {
			s.State = f[2]
			s.Description = strings.Join(f[4:], " ")
		}
		if s.Name > after {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func ServicesJSON(ctx context.Context, limit int) ([]byte, error) {
	x, e := ListServices(ctx, limit)
	if e != nil {
		return nil, e
	}
	return json.Marshal(x)
}

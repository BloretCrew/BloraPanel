//go:build linux

package master

import (
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/monitor"
)

func TestHostProcessTaskDispatchAndIdentity(t *testing.T) {
	f := newFileFixture(t)
	child := exec.Command("sleep", "60")
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- child.Wait() }()
	t.Cleanup(func() { _ = child.Process.Kill() })
	pid := strconv.Itoa(child.Process.Pid)
	listURL := "/nodes/" + f.instances[0].NodeID + "/processes"
	f.reader.request("GET", listURL, nil, "", 403)
	f.admin.request("GET", listURL+"?limit=101", nil, "", 400)
	var page monitor.ProcessPage
	if err := json.Unmarshal(f.admin.request("GET", listURL+"?search="+pid, nil, "", 200)["items"], &page.Items); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range page.Items {
		if p.PID == child.Process.Pid {
			found = true
		}
	}
	if !found {
		t.Fatalf("process search did not reach daemon: %+v", page)
	}
	stat, err := os.ReadFile("/proc/" + pid + "/stat")
	if err != nil {
		t.Fatal(err)
	}
	fields := strings.Fields(string(stat)[strings.LastIndex(string(stat), ")")+2:])
	start, err := strconv.ParseUint(fields[19], 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	url := "/nodes/" + f.instances[0].NodeID + "/processes/" + pid + "/terminate"
	f.reader.request("POST", url, map[string]any{"startTicks": start}, model.ID(), 403)
	bad := parseFileTask(t, f.admin.request("POST", url, map[string]any{"startTicks": start + 1}, model.ID(), 202))
	f.awaitFile(t, bad, model.Failed)
	select {
	case <-done:
		t.Fatal("identity mismatch terminated the process")
	default:
	}
	key := model.ID()
	result := f.admin.request("POST", url, map[string]any{"startTicks": start}, key, 202)
	task := f.awaitFile(t, parseFileTask(t, result), model.Succeeded)
	replay := parseFileTask(t, f.admin.request("POST", url, map[string]any{"startTicks": start}, key, 202))
	if replay.ID != task.ID {
		t.Fatal("duplicate termination created a second task")
	}
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("process exited without termination")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("successful task did not terminate its target")
	}
}

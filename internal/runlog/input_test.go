//go:build linux || windows

package runlog

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInputBusinessProcess(t *testing.T) {
	result := os.Getenv("BLORA_TEST_INPUT_RESULT")
	if result == "" {
		return
	}
	_ = os.Stdout.Close()
	_ = os.Stderr.Close()
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		_ = os.WriteFile(result, []byte("unexpected EOF"), 0600)
		os.Exit(81)
	}
	if err = os.WriteFile(result, []byte(line), 0600); err != nil {
		os.Exit(82)
	}
	os.Exit(0)
}

func TestInputDaemonProcess(t *testing.T) {
	root := os.Getenv("BLORA_TEST_INPUT_DAEMON")
	if root == "" {
		return
	}
	c, err := Start(context.Background(), Options{Root: root, RunID: "independent-input", HoldInput: true, Command: helperCommand()})
	if err != nil {
		os.Exit(83)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestInputBusinessProcess$")
	cmd.Env = append(os.Environ(), "BLORA_TEST_INPUT_RESULT="+filepath.Join(root, "business-result"), "GORACE=atexit_sleep_ms=0")
	cmd.Stdin = c.Stdin()
	cmd.Stdout = c.Stdout()
	cmd.Stderr = c.Stdout()
	if err = cmd.Start(); err != nil {
		os.Exit(84)
	}
	_ = c.ReleaseWriter()
	_ = c.ReleaseInputReader()
	_ = json.NewEncoder(os.Stdout).Encode(map[string]int{"businessPID": cmd.Process.Pid})
	os.Exit(0)
}

func TestStdinSurvivesDaemonAndOutputEOF(t *testing.T) {
	root := t.TempDir()
	daemon := exec.Command(os.Args[0], "-test.run=^TestInputDaemonProcess$")
	daemon.Env = append(os.Environ(), "BLORA_TEST_INPUT_DAEMON="+root, "GORACE=atexit_sleep_ms=0")
	out, err := daemon.Output()
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		BusinessPID int `json:"businessPID"`
	}
	if err = json.Unmarshal(out, &result); err != nil || result.BusinessPID <= 1 {
		t.Fatalf("daemon result %v", err)
	}
	c, err := Open(root, "independent-input")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Finish(context.Background()) })
	if c.Stdin() != nil {
		t.Fatal("historical Open duplicated a child stdin descriptor")
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		s, e := c.Status(context.Background())
		if e != nil {
			t.Fatal(e)
		}
		if s.Phase == "output_complete" {
			if !s.InputAvailable {
				t.Fatal("closed stdout incorrectly disabled independent stdin")
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("stdout EOF not observed: %+v", s)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err = os.Stat(filepath.Join(root, "business-result")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("business received EOF before management input")
	}
	line := "只发送一次的输入\n"
	if n, err := c.WriteInput(context.Background(), []byte(line)); err != nil || n != len(line) {
		t.Fatalf("independent input: %d %v", n, err)
	}
	// The result path can exist while WriteFile is still writing it. Wait for
	// the owned business process to finish before inspecting its completed
	// result, and before invoking the runtime-owned Finish action.
	if err = waitTestProcessExit(result.BusinessPID, 3*time.Second); err != nil {
		t.Fatal(err)
	}
	for {
		data, e := os.ReadFile(filepath.Join(root, "business-result"))
		if e == nil {
			if string(data) != line {
				t.Fatalf("business input %q", data)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("business did not receive input")
		}
		time.Sleep(10 * time.Millisecond)
	}
	batch, err := c.Read(context.Background(), 0, 4096)
	if err != nil || len(batch.Events) != 0 {
		t.Fatalf("stdin was incorrectly archived: %+v %v", batch, err)
	}
	if err = c.Finish(context.Background()); err != nil {
		t.Fatal(err)
	}
	status, err := c.Status(context.Background())
	if err != nil || status.InputAvailable {
		t.Fatalf("finished capture advertises stdin: %+v %v", status, err)
	}
	t.Log("daemon and business stdout exited, stdin stayed open; one authenticated command reached business and was not replayed")
}

func TestIndependentInputBudgetAndBoundedBackpressure(t *testing.T) {
	c, err := Start(context.Background(), Options{Root: t.TempDir(), RunID: "blocked-input", HoldInput: true, Command: helperCommand()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Finish(context.Background()) })
	if c.Stdin() == nil {
		t.Fatal("missing business read end")
	}
	if _, err = c.Stdin().Write([]byte("wrong direction")); err == nil {
		t.Fatal("child stdin copy is writable")
	}
	if _, err = c.WriteInput(context.Background(), make([]byte, (32<<10)+1)); err == nil {
		t.Fatal("oversized input accepted")
	}
	blocked := false
	for i := 0; i < 32; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		began := time.Now()
		n, e := c.WriteInput(ctx, []byte(strings.Repeat("x", 32<<10)))
		cancel()
		if time.Since(began) > time.Second {
			t.Fatal("input timeout was unbounded")
		}
		if e != nil && n == 0 {
			blocked = true
			break
		}
		if e != nil && !errors.Is(e, io.ErrShortWrite) {
			t.Fatalf("unexpected write result: %d %v", n, e)
		}
	}
	if !blocked {
		t.Fatal("test did not fill the bounded OS pipe")
	}
	if err = c.Finish(context.Background()); err != nil {
		t.Fatal(err)
	}
}

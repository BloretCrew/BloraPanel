package master

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"blora.dev/panel/internal/runlog"
	run "blora.dev/panel/internal/runtime"
)

func TestMain(m *testing.M) {
	if len(os.Args) > 2 && os.Args[1] == "--native-test-process" {
		if err := nativeTestProcess(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if len(os.Args) > 1 && os.Args[1] == "--job-keeper" {
		if err := run.RunJobKeeper(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if len(os.Args) > 1 && os.Args[1] == "--log-helper" {
		if err := runlog.RunHelper(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// The real child executable is portable. Tests exercise the same stdin, Job /
// process-tree, logging and durable run paths without requiring a Unix shell.
func nativeTestCommand(t *testing.T, mode string, args ...string) []string {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return append([]string{exe, "--native-test-process", mode}, args...)
}

func nativeTestEnvironment(existing ...map[string]string) map[string]string {
	env := map[string]string{}
	for _, values := range existing {
		for key, value := range values {
			env[key] = value
		}
	}
	// Race-instrumented test executables sleep for one second on normal exit.
	// Disable only that fixture delay, preserving the production stop deadline
	// and race detection in both the helper and the Master/Daemon processes.
	env["GORACE"] = "atexit_sleep_ms=0"
	return env
}

func nativeTestProcess(args []string) error {
	appendFile := func(name, value string) error {
		f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			return err
		}
		_, err = f.WriteString(value)
		return errorsJoinClose(err, f.Close())
	}
	mode := args[0]
	if mode == "stubborn" || mode == "stubborn-idle" || mode == "generations" {
		// Go handles CTRL_BREAK as os.Interrupt on Windows. Installing a handler
		// also keeps this fixture alive for the finite failed-stop assertion.
		ignored := make(chan os.Signal, 1)
		signal.Notify(ignored, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(ignored)
		if mode == "stubborn" {
			if err := appendFile("starts", "x"); err != nil {
				return err
			}
			if err := os.WriteFile("ready", nil, 0600); err != nil {
				return err
			}
		} else if mode == "generations" {
			if err := appendFile("generations", "x"); err != nil {
				return err
			}
		}
	}
	switch mode {
	case "idle", "stubborn", "stubborn-idle", "generations":
		time.Sleep(10 * time.Minute)
	case "log-markers":
		fmt.Println("first-marker")
		time.Sleep(2 * time.Second)
		fmt.Println("second-marker")
		time.Sleep(60 * time.Second)
	case "console", "stop-policy", "scheduled-input":
		fmt.Println("ready")
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			if mode == "scheduled-input" {
				if err := appendFile("scheduled-input", line+"\n"); err != nil {
					return err
				}
			} else if line == "quit" {
				file := "stop-policy"
				if mode == "console" {
					file = "graceful-stop"
				}
				return os.WriteFile(file, []byte("stopped"), 0600)
			} else if mode == "console" {
				if err := appendFile("input-count", "x"); err != nil {
					return err
				}
				fmt.Println("CONSOLE:" + line)
			}
		}
		return scanner.Err()
	case "continuous-log":
		delay, err := time.ParseDuration(args[1])
		if err != nil {
			return err
		}
		fmt.Println("中文日志-ready")
		line := strings.Repeat("0", 4096)
		for {
			fmt.Println(line)
			time.Sleep(delay)
		}
	case "descendants":
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		for _, childMode := range []string{"busy", "idle"} {
			child := exec.Command(exe, "--native-test-process", childMode)
			if err := child.Start(); err != nil {
				return err
			}
			defer child.Process.Kill()
		}
		time.Sleep(10 * time.Minute)
	case "busy":
		var n uint64
		for {
			atomic.AddUint64(&n, 1)
		}
	default:
		return fmt.Errorf("unknown native test mode %q", mode)
	}
	return nil
}

func errorsJoinClose(writeErr, closeErr error) error {
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

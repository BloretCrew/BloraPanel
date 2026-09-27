//go:build windows

package runtime

import (
	"bufio"
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var compareObjectHandles = windows.NewLazySystemDLL("kernelbase.dll").NewProc("CompareObjectHandles")

type keeperRecord struct {
	Version int    `json:"version"`
	RunID   string `json:"runId"`
	Token   string `json:"token"`
	Unit    string `json:"unit"`
	PID     int    `json:"pid"`
	Birth   uint64 `json:"birth"`
	Handle  uint64 `json:"handle"` // Internal handle in the exact keeper process.
	Phase   string `json:"phase"`
}

type startedKeeper struct {
	cmd  *exec.Cmd
	done chan error
}

func (k *startedKeeper) abort() error {
	_ = k.cmd.Process.Kill()
	select {
	case <-k.done:
		return nil
	case <-time.After(2 * time.Second):
		return fmt.Errorf("keeper startup cleanup unconfirmed: %w", ErrUnknown)
	}
}

func keeperPath(root, runID string) string { return filepath.Join(root, "keepers", runID+".json") }
func saveKeeper(root string, k keeperRecord) error {
	path := keeperPath(root, k.RunID)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	b, err := json.Marshal(k)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".keeper-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(b); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(f.Name(), path)
}
func loadKeeper(root string, r Record) (keeperRecord, error) {
	var k keeperRecord
	f, err := os.Open(keeperPath(root, r.RunID))
	if err != nil {
		return k, err
	}
	defer f.Close()
	err = json.NewDecoder(io.LimitReader(f, 8192)).Decode(&k)
	if err == nil && (k.Version != 1 || k.RunID != r.RunID || k.Token != r.Token || k.Unit != r.Unit || k.PID <= 0 || k.Birth == 0 || k.Handle == 0) {
		err = ErrUnknown
	}
	return k, err
}
func keeperProcess(k keeperRecord, access uint32) (windows.Handle, error) {
	h, err := windows.OpenProcess(access|windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.SYNCHRONIZE, false, uint32(k.PID))
	if err != nil {
		return 0, err
	}
	birth, err := windowsBirth(h)
	if err != nil || birth != k.Birth {
		windows.CloseHandle(h)
		return 0, ErrUnknown
	}
	state, err := windows.WaitForSingleObject(h, 0)
	if err != nil || state != uint32(windows.WAIT_TIMEOUT) {
		windows.CloseHandle(h)
		return 0, ErrUnknown
	}
	return h, nil
}

// validateKeeper proves the named Job is the SAME kernel object held by the
// original keeper. A matching PID or a recreated Job name alone is not proof.
func (m *Manager) validateKeeper(r Record, job windows.Handle) error {
	k, err := loadKeeper(m.options.StateRoot, r)
	if err != nil {
		return fmt.Errorf("Job keeper record: %w", ErrUnknown)
	}
	if k.Phase != "holding" {
		return ErrUnknown
	}
	h, err := keeperProcess(k, windows.PROCESS_DUP_HANDLE)
	if err != nil {
		return fmt.Errorf("Job keeper birth/liveness: %w", ErrUnknown)
	}
	defer windows.CloseHandle(h)
	var duplicate windows.Handle
	if err = windows.DuplicateHandle(h, windows.Handle(k.Handle), windows.CurrentProcess(), &duplicate, 0, false, windows.DUPLICATE_SAME_ACCESS); err != nil {
		return fmt.Errorf("Job keeper handle: %w", ErrUnknown)
	}
	defer windows.CloseHandle(duplicate)
	if err = compareObjectHandles.Find(); err != nil {
		return fmt.Errorf("CompareObjectHandles: %w", ErrCapability)
	}
	same, _, _ := compareObjectHandles.Call(uintptr(job), uintptr(duplicate))
	if same == 0 {
		return ErrUnknown
	}
	return nil
}

func (m *Manager) startJobKeeper(ctx context.Context, r Record, job windows.Handle) (*startedKeeper, error) {
	command := m.options.JobKeeperCommand
	if len(command) == 0 {
		executable, err := os.Executable()
		if err != nil {
			return nil, err
		}
		command = []string{executable, "--job-keeper"}
	}
	args := append(append([]string{}, command[1:]...), "--state-root", m.options.StateRoot, "--run-id", r.RunID)
	cmd := exec.Command(command[0], args...)
	flags := uint32(windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP)
	inside, err := processInJob(windows.CurrentProcess(), 0)
	if err != nil {
		return nil, err
	}
	if inside {
		flags |= windows.CREATE_BREAKAWAY_FROM_JOB
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: flags}
	for _, value := range os.Environ() {
		key, _, _ := strings.Cut(value, "=")
		switch strings.ToUpper(key) {
		case "SYSTEMROOT", "WINDIR", "PATH", "TMP", "TEMP", "GORACE":
			cmd.Env = append(cmd.Env, value)
		}
	}
	cmd.Env = append(cmd.Env, "BLORA_JOB_KEEPER_TOKEN="+r.Token)
	read, write, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer read.Close()
	defer write.Close()
	cmd.Stdout = write
	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("independent Job keeper launch: %w", err)
	}
	pid := cmd.Process.Pid
	write.Close()
	done := make(chan error, 1)
	go func() {
		line, e := bufio.NewReaderSize(read, 1024).ReadSlice('\n')
		if e == nil && string(line) != "BLORA-JOB-KEEPER/1 READY\n" {
			e = ErrUnknown
		}
		done <- e
	}()
	keeper := &startedKeeper{cmd: cmd, done: make(chan error, 1)}
	go func() { keeper.done <- cmd.Wait() }()
	deadline, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	select {
	case err = <-done:
	case <-deadline.Done():
		err = deadline.Err()
	}
	if err == nil {
		var k keeperRecord
		k, err = loadKeeper(m.options.StateRoot, r)
		if err == nil && k.PID != pid {
			err = ErrUnknown
		}
	}
	if err == nil {
		err = m.validateKeeper(r, job)
	}
	if err != nil {
		return nil, errors.Join(fmt.Errorf("Job keeper readiness: %w", err), keeper.abort())
	}
	return keeper, nil
}

// RunJobKeeper is called by the daemon's hidden --job-keeper mode, before normal
// daemon initialization. This process remains outside all instance/daemon Jobs
// and holds only the specific, persisted Job. It never listens on the network.
func RunJobKeeper(args []string) error {
	if len(args) != 4 || args[0] != "--state-root" || args[2] != "--run-id" || !filepath.IsAbs(args[1]) || !idPattern.MatchString(args[3]) {
		return errors.New("invalid Job keeper arguments")
	}
	root, runID := filepath.Clean(args[1]), args[3]
	f, err := os.Open(filepath.Join(root, runID+".json"))
	if err != nil {
		return err
	}
	var r Record
	err = json.NewDecoder(io.LimitReader(f, 64<<10)).Decode(&r)
	f.Close()
	if err != nil {
		return err
	}
	token := os.Getenv("BLORA_JOB_KEEPER_TOKEN")
	if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(r.Token)) != 1 || r.RunID != runID || r.Backend != "windows" || r.Unit != `Local\Blora-`+r.RunID+"-"+r.Token {
		return ErrUnknown
	}
	inside, err := processInJob(windows.CurrentProcess(), 0)
	if err != nil {
		return err
	}
	if inside {
		return fmt.Errorf("keeper remained inside enclosing Job: %w", ErrCapability)
	}
	job, err := openJob(r.Unit)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(job)
	birth, err := windowsBirth(windows.CurrentProcess())
	if err != nil {
		return err
	}
	k := keeperRecord{Version: 1, RunID: runID, Token: r.Token, Unit: r.Unit, PID: os.Getpid(), Birth: birth, Handle: uint64(job), Phase: "holding"}
	if err = saveKeeper(root, k); err != nil {
		return err
	}
	if _, err = io.WriteString(os.Stdout, "BLORA-JOB-KEEPER/1 READY\n"); err != nil {
		return err
	}
	_ = os.Stdout.Close()
	initialDeadline := time.Now().Add(10 * time.Second)
	for {
		var accounting jobAccounting
		if err = windows.QueryInformationJobObject(job, windows.JobObjectBasicAccountingInformation, uintptr(unsafe.Pointer(&accounting)), uint32(unsafe.Sizeof(accounting)), nil); err != nil {
			return err
		}
		if accounting.ActiveProcesses == 0 && (accounting.TotalProcesses > 0 || time.Now().After(initialDeadline)) {
			// Empty startup failure is bounded. Once ANY process was assigned,
			// there is no timeout: every descendant must actually exit first.
			k.Phase = "complete"
			return saveKeeper(root, k)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

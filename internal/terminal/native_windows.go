//go:build windows

package terminal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

type nativeBackend struct{}
type windowsProcess struct {
	input, output         *os.File
	console, job, process windows.Handle
	mu                    sync.Mutex
	closeOnce             sync.Once
	closeErr              error
	closed                bool
}
type jobAccounting struct {
	User, Kernel, PeriodUser, PeriodKernel int64
	Faults, Total, Active, Terminated      uint32
}

var conPTYKernel = windows.NewLazySystemDLL("kernel32.dll")
var updateConPTYAttribute = conPTYKernel.NewProc("UpdateProcThreadAttribute")

func (nativeBackend) Spawn(ctx context.Context, s SpawnSpec) (Process, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := conPTYKernel.NewProc("CreatePseudoConsole").Find(); err != nil {
		return nil, fmt.Errorf("ConPTY needs supported Windows 10/Server: %w", ErrCapability)
	}
	inRead, inWrite, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	outRead, outWrite, err := os.Pipe()
	if err != nil {
		inRead.Close()
		inWrite.Close()
		return nil, err
	}
	defer inRead.Close()
	defer outWrite.Close()
	p := &windowsProcess{input: inWrite, output: outRead}
	ok := false
	defer func() {
		if !ok {
			inWrite.Close()
			outRead.Close()
			if p.job != 0 {
				_ = windows.TerminateJobObject(p.job, 1)
				windows.CloseHandle(p.job)
			}
			if p.process != 0 {
				windows.CloseHandle(p.process)
			}
			if p.console != 0 {
				windows.ClosePseudoConsole(p.console)
			}
		}
	}()
	if err = windows.CreatePseudoConsole(windows.Coord{X: int16(s.Cols), Y: int16(s.Rows)}, windows.Handle(inRead.Fd()), windows.Handle(outWrite.Fd()), 0, &p.console); err != nil {
		return nil, fmt.Errorf("create ConPTY: %w", err)
	}
	if p.job, err = windows.CreateJobObject(nil, nil); err != nil {
		return nil, err
	}
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err = windows.SetInformationJobObject(p.job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits))); err != nil {
		return nil, err
	}
	attrs, err := windows.NewProcThreadAttributeList(2)
	if err != nil {
		return nil, err
	}
	defer attrs.Delete()
	// ConPTY's documented lpValue is the HPCON value itself, unlike normal
	// pointer-valued attributes. Pass it as uintptr to avoid an invalid Go
	// pointer retained by ProcThreadAttributeListContainer.Update.
	r, _, callErr := updateConPTYAttribute.Call(uintptr(unsafe.Pointer(attrs.List())), 0, windows.PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE, uintptr(p.console), unsafe.Sizeof(p.console), 0, 0)
	if r == 0 {
		return nil, fmt.Errorf("ConPTY startup attribute: %w", callErr)
	}
	// Atomic Job assignment closes the daemon-crash gap before ResumeThread.
	if err = attrs.Update(0x0002000d, unsafe.Pointer(&p.job), unsafe.Sizeof(p.job)); err != nil {
		return nil, fmt.Errorf("atomic Job assignment unavailable: %w", err)
	}
	si := windows.StartupInfoEx{}
	si.Cb = uint32(unsafe.Sizeof(si))
	si.ProcThreadAttributeList = attrs.List()
	// With redirected daemon/test stdio, Windows can otherwise copy the
	// parent's pipe handles despite HPCON. Explicit null standard handles make
	// the console subsystem connect stdin/stdout/stderr to this pseudoconsole.
	// https://github.com/microsoft/terminal/discussions/15814
	si.Flags = windows.STARTF_USESTDHANDLES
	executable, err := exec.LookPath(s.Command[0])
	if err != nil {
		return nil, err
	}
	app, err := windows.UTF16PtrFromString(executable)
	if err != nil {
		return nil, err
	}
	command, err := windows.UTF16PtrFromString(windows.ComposeCommandLine(append([]string{executable}, s.Command[1:]...)))
	if err != nil {
		return nil, err
	}
	var dir *uint16
	if s.Directory != "" {
		if dir, err = windows.UTF16PtrFromString(s.Directory); err != nil {
			return nil, err
		}
	}
	root, err := windows.GetWindowsDirectory()
	if err != nil {
		return nil, err
	}
	env := map[string]string{"SYSTEMROOT": root, "WINDIR": root, "PATH": root + `\System32;` + root, "TERM": "xterm-256color", "BLORA_TERMINAL_ID": s.SessionID}
	for k, v := range s.Environment {
		env[strings.ToUpper(k)] = v
	}
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var block []uint16
	for _, k := range keys {
		v, err := windows.UTF16FromString(k + "=" + env[k])
		if err != nil {
			return nil, err
		}
		block = append(block, v...)
	}
	block = append(block, 0)
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	var pi windows.ProcessInformation
	flags := uint32(windows.EXTENDED_STARTUPINFO_PRESENT | windows.CREATE_UNICODE_ENVIRONMENT | windows.CREATE_SUSPENDED)
	if err = windows.CreateProcess(app, command, nil, nil, false, flags, &block[0], dir, &si.StartupInfo, &pi); err != nil {
		return nil, fmt.Errorf("suspended ConPTY process in Job: %w", err)
	}
	p.process = pi.Process
	defer windows.CloseHandle(pi.Thread)
	// Already assigned atomically; explicitly validate the Job contains it.
	var accounting jobAccounting
	if err = windows.QueryInformationJobObject(p.job, windows.JobObjectBasicAccountingInformation, uintptr(unsafe.Pointer(&accounting)), uint32(unsafe.Sizeof(accounting)), nil); err != nil || accounting.Active != 1 {
		return nil, errors.New("ConPTY process Job membership not confirmed")
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	if _, err = windows.ResumeThread(pi.Thread); err != nil {
		return nil, err
	}
	ok = true
	return p, nil
}
func (p *windowsProcess) Read(b []byte) (int, error) {
	n, err := p.output.Read(b)
	if errors.Is(err, windows.ERROR_BROKEN_PIPE) {
		err = io.EOF
	}
	return n, err
}
func (p *windowsProcess) Write(b []byte) (int, error) {
	// os.Pipe uses synchronous Windows pipe handles. CancelIoEx cancels a
	// stalled write; bounded frames and one lease mean at most one in flight.
	finished := make(chan struct{})
	timer := time.AfterFunc(2*time.Second, func() { _ = windows.CancelIoEx(windows.Handle(p.input.Fd()), nil); close(finished) })
	n, err := p.input.Write(b)
	if !timer.Stop() {
		<-finished
	}
	return n, err
}
func (p *windowsProcess) Resize(cols, rows uint16) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return os.ErrClosed
	}
	return windows.ResizePseudoConsole(p.console, windows.Coord{X: int16(cols), Y: int16(rows)})
}
func (p *windowsProcess) Wait() error {
	for {
		p.mu.Lock()
		if p.closed {
			p.mu.Unlock()
			return nil
		}
		result, err := windows.WaitForSingleObject(p.process, 100)
		if err != nil {
			p.mu.Unlock()
			return err
		}
		if result == windows.WAIT_OBJECT_0 {
			var code uint32
			err = windows.GetExitCodeProcess(p.process, &code)
			p.mu.Unlock()
			if err != nil {
				return err
			}
			if code != 0 {
				return fmt.Errorf("terminal process exited with status %d", code)
			}
			return nil
		}
		p.mu.Unlock()
	}
}
func (p *windowsProcess) Close() error {
	p.closeOnce.Do(func() {
		p.mu.Lock()
		p.closed = true
		err := windows.TerminateJobObject(p.job, 1)
		p.mu.Unlock()
		if err != nil {
			p.closeErr = err
		}
		deadline := time.Now().Add(3 * time.Second)
		for {
			var a jobAccounting
			err = windows.QueryInformationJobObject(p.job, windows.JobObjectBasicAccountingInformation, uintptr(unsafe.Pointer(&a)), uint32(unsafe.Sizeof(a)), nil)
			if err != nil {
				p.closeErr = errors.Join(p.closeErr, err)
				break
			}
			if a.Active == 0 {
				break
			}
			if time.Now().After(deadline) {
				p.closeErr = errors.Join(p.closeErr, errors.New("ConPTY Job still has active processes"))
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
		_ = p.input.Close()
		// ClosePseudoConsole can produce final output and block until drained;
		// the manager's reader remains active throughout this call.
		closed := make(chan struct{})
		go func() { windows.ClosePseudoConsole(p.console); close(closed) }()
		select {
		case <-closed:
		case <-time.After(2 * time.Second):
			p.closeErr = errors.Join(p.closeErr, errors.New("ConPTY close output drain deadline exceeded"))
		}
		_ = p.output.Close()
		windows.CloseHandle(p.process)
		windows.CloseHandle(p.job)
	})
	return p.closeErr
}

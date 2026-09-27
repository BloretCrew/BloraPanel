//go:build windows

package runtime

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	goruntime "runtime"
	"sort"
	"strings"
	"time"
	"unsafe"

	"blora.dev/panel/internal/model"
	"golang.org/x/sys/windows"
)

const procThreadAttributeJobList = 0x0002000d
const jobAllAccess = 0x001f003f

var kernel = windows.NewLazySystemDLL("kernel32.dll")
var openJobProc = kernel.NewProc("OpenJobObjectW")
var inJobProc = kernel.NewProc("IsProcessInJob")

type windowsRun struct{ job windows.Handle }
type jobAccounting struct {
	TotalUserTime, TotalKernelTime, ThisPeriodTotalUserTime, ThisPeriodTotalKernelTime int64
	TotalPageFaultCount, TotalProcesses, ActiveProcesses, TotalTerminatedProcesses     uint32
}

func nativeBackend() string      { return "windows" }
func syncDirectory(string) error { return nil } // File contents use FlushFileBuffers via File.Sync; directory fsync is not available.
func prepareContainerDirectory(path string, _, _ uint32) error {
	_, err := newContainerDirectory(path)
	return err
}
func openJob(name string) (windows.Handle, error) {
	p, e := windows.UTF16PtrFromString(name)
	if e != nil {
		return 0, e
	}
	r, _, e := openJobProc.Call(jobAllAccess, 0, uintptr(unsafe.Pointer(p)))
	if r == 0 {
		return 0, e
	}
	return windows.Handle(r), nil
}
func processInJob(process, job windows.Handle) (bool, error) {
	var result int32
	r, _, e := inJobProc.Call(uintptr(process), uintptr(job), uintptr(unsafe.Pointer(&result)))
	if r == 0 {
		return false, e
	}
	return result != 0, nil
}
func windowsBirth(handle windows.Handle) (uint64, error) {
	var created, exited, kernel, user windows.Filetime
	err := windows.GetProcessTimes(handle, &created, &exited, &kernel, &user)
	return uint64(created.HighDateTime)<<32 | uint64(created.LowDateTime), err
}
func windowsEnvironment(c model.InstanceConfig, token string) ([]uint16, error) {
	root, err := windows.GetWindowsDirectory()
	if err != nil {
		return nil, err
	}
	values := map[string]string{"SYSTEMROOT": root, "WINDIR": root, "PATH": root + `\System32;` + root}
	for k, v := range c.Environment {
		if strings.ContainsAny(k, "=\x00") || strings.ContainsRune(v, 0) {
			return nil, errors.New("invalid environment variable")
		}
		values[strings.ToUpper(k)] = v
	}
	values["BLORA_RUN_TOKEN"] = token
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var result []uint16
	for _, k := range keys {
		v, e := windows.UTF16FromString(k + "=" + values[k])
		if e != nil {
			return nil, e
		}
		result = append(result, v...)
	}
	return append(result, 0), nil
}

func (m *Manager) startNative(ctx context.Context, r Record, c model.InstanceConfig, streams IO) (result Record, resultErr error) {
	if c.UID != nil || c.GID != nil {
		return r, fmt.Errorf("Unix UID/GID do not identify a Windows account: %w", ErrCapability)
	}
	r.Unit = `Local\Blora-` + r.RunID + "-" + r.Token
	name, err := windows.UTF16PtrFromString(r.Unit)
	if err != nil {
		return r, err
	}
	job, err := windows.CreateJobObject(nil, name)
	if err != nil {
		return r, fmt.Errorf("create Job Object: %w", err)
	}
	ok := false
	defer func() {
		if !ok {
			resultErr = errors.Join(resultErr, windows.CloseHandle(job))
		}
	}()
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	// No business BREAKAWAY or SILENT_BREAKAWAY. An independent keeper retains
	// the Job handle across daemon exits; losing a management handle is harmless.
	if c.MemoryBytes > 0 {
		limits.JobMemoryLimit = uintptr(c.MemoryBytes)
		limits.BasicLimitInformation.LimitFlags |= windows.JOB_OBJECT_LIMIT_JOB_MEMORY
	}
	if c.PidsLimit > 0 {
		if c.PidsLimit > 1<<32-1 {
			return r, errors.New("process limit exceeds Windows range")
		}
		limits.BasicLimitInformation.ActiveProcessLimit = uint32(c.PidsLimit)
		limits.BasicLimitInformation.LimitFlags |= windows.JOB_OBJECT_LIMIT_ACTIVE_PROCESS
	}
	if _, err = windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits))); err != nil {
		return r, err
	}
	if c.CPUQuota > 0 {
		cores := int64(goruntime.NumCPU())
		// Common quota units are microseconds per 100 ms (100000 = one CPU).
		// Windows rate is a fraction of the entire machine, in 1/10000 units.
		rate, rateErr := windowsCPUQuotaRate(c.CPUQuota, cores)
		if rateErr != nil {
			return r, rateErr
		}
		cpu := struct{ ControlFlags, CPURate uint32 }{ControlFlags: 1 | 4, CPURate: rate}
		if _, err = windows.SetInformationJobObject(job, windows.JobObjectCpuRateControlInformation, uintptr(unsafe.Pointer(&cpu)), uint32(unsafe.Sizeof(cpu))); err != nil {
			return r, err
		}
	}
	if err = m.save(r); err != nil {
		return r, err
	}
	attrs, err := windows.NewProcThreadAttributeList(2)
	if err != nil {
		return r, fmt.Errorf("process attributes: %w", err)
	}
	defer attrs.Delete()
	// Atomic Job assignment also closes the daemon-crash gap between suspended
	// creation and a separate AssignProcessToJobObject call (Windows 10+).
	if err = attrs.Update(procThreadAttributeJobList, unsafe.Pointer(&job), unsafe.Sizeof(job)); err != nil {
		return r, fmt.Errorf("atomic Job assignment unsupported: %w", err)
	}
	null, err := os.OpenFile(os.DevNull, os.O_RDWR, 0600)
	if err != nil {
		return r, err
	}
	defer null.Close()
	files := []*os.File{streams.Stdin, streams.Stdout, streams.Stderr}
	handles := make([]windows.Handle, 3)
	for index, f := range files {
		if f == nil {
			f = null
		}
		if err = windows.DuplicateHandle(windows.CurrentProcess(), windows.Handle(f.Fd()), windows.CurrentProcess(), &handles[index], 0, true, windows.DUPLICATE_SAME_ACCESS); err != nil {
			return r, err
		}
		defer windows.CloseHandle(handles[index])
	}
	if err = attrs.Update(windows.PROC_THREAD_ATTRIBUTE_HANDLE_LIST, unsafe.Pointer(&handles[0]), uintptr(len(handles))*unsafe.Sizeof(handles[0])); err != nil {
		return r, err
	}
	startup := windows.StartupInfoEx{}
	startup.Cb = uint32(unsafe.Sizeof(startup))
	startup.Flags = windows.STARTF_USESTDHANDLES
	startup.StdInput = handles[0]
	startup.StdOutput = handles[1]
	startup.StdErr = handles[2]
	startup.ProcThreadAttributeList = attrs.List()
	executable, err := exec.LookPath(c.Command[0])
	if err != nil {
		return r, err
	}
	app, err := windows.UTF16PtrFromString(executable)
	if err != nil {
		return r, err
	}
	command, err := windows.UTF16PtrFromString(windows.ComposeCommandLine(append([]string{executable}, c.Command[1:]...)))
	if err != nil {
		return r, err
	}
	var directory *uint16
	if c.Directory != "" {
		directory, err = windows.UTF16PtrFromString(c.Directory)
		if err != nil {
			return r, err
		}
	}
	env, err := windowsEnvironment(c, r.Token)
	if err != nil {
		return r, err
	}
	if err = ctx.Err(); err != nil {
		return r, err
	}
	keeper, err := m.startJobKeeper(ctx, r, job)
	if err != nil {
		return r, err
	}
	defer func() {
		if !ok {
			// A keeper must not preserve a half-created/suspended run after a
			// reported startup failure. These are handles created by this call.
			resultErr = errors.Join(resultErr, windows.TerminateJobObject(job, 1), keeper.abort())
		}
	}()
	var process windows.ProcessInformation
	flags := uint32(windows.CREATE_SUSPENDED | windows.CREATE_NEW_PROCESS_GROUP | windows.EXTENDED_STARTUPINFO_PRESENT | windows.CREATE_UNICODE_ENVIRONMENT)
	insideParentJob, err := processInJob(windows.CurrentProcess(), 0)
	if err != nil {
		return r, err
	}
	if insideParentJob {
		// The business must not remain in a daemon-owned outer Job either.
		// BREAKAWAY must be allowed there; otherwise creation fails explicitly.
		flags |= windows.CREATE_BREAKAWAY_FROM_JOB
	}
	if err = windows.CreateProcess(app, command, nil, nil, true, flags, &env[0], directory, &startup.StartupInfo, &process); err != nil {
		return r, fmt.Errorf("suspended creation in Job: %w", err)
	}
	defer windows.CloseHandle(process.Thread)
	defer windows.CloseHandle(process.Process)
	assigned, err := processInJob(process.Process, job)
	if err != nil || !assigned {
		_ = windows.TerminateProcess(process.Process, 1)
		return r, fmt.Errorf("process Job membership was not confirmed: %w", ErrUnknown)
	}
	r.PID = int(process.ProcessId)
	r.PGID = r.PID
	r.StartTicks, err = windowsBirth(process.Process)
	if err != nil {
		return r, err
	}
	r.Phase = "suspended"
	if err = m.save(r); err != nil {
		return r, err
	}
	if err = ctx.Err(); err != nil {
		return r, err
	}
	if err = m.validateKeeper(r, job); err != nil {
		return r, fmt.Errorf("keeper disappeared before execution: %w", err)
	}
	if _, err = windows.ResumeThread(process.Thread); err != nil {
		return r, err
	}
	m.mu.Lock()
	m.live[r.RunID] = &liveRun{stdin: streams.Input, platform: &windowsRun{job: job}}
	m.mu.Unlock()
	ok = true
	return r, nil
}
func (m *Manager) job(r Record) (windows.Handle, bool, error) {
	if r.Backend != "windows" || r.Unit != `Local\Blora-`+r.RunID+"-"+r.Token {
		return 0, false, ErrUnknown
	}
	h, err := openJob(r.Unit)
	if err != nil {
		return 0, false, err
	}
	var accounting jobAccounting
	if err = windows.QueryInformationJobObject(h, windows.JobObjectBasicAccountingInformation, uintptr(unsafe.Pointer(&accounting)), uint32(unsafe.Sizeof(accounting)), nil); err == nil && accounting.ActiveProcesses != 0 {
		err = m.validateKeeper(r, h)
	}
	if err != nil {
		windows.CloseHandle(h)
		return 0, false, err
	}
	return h, true, nil
}
func (m *Manager) observeNative(ctx context.Context, r Record) (Observation, error) {
	o := Observation{State: "UNKNOWN", ObservedAt: time.Now().UTC()}
	job, closeIt, err := m.job(r)
	if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
		// A named Job is destroyed only after the last handle is closed and its
		// associated processes have exited. The independent keeper normally
		// retains the final handle until it confirms ActiveProcesses == 0.
		o.State = "STOPPED"
		o.Exited = true
		o.Diagnostic = "named Job no longer exists; no associated processes remain"
		return o, nil
	}
	if err != nil {
		return o, err
	}
	if closeIt {
		defer windows.CloseHandle(job)
	}
	var accounting jobAccounting
	if err = windows.QueryInformationJobObject(job, windows.JobObjectBasicAccountingInformation, uintptr(unsafe.Pointer(&accounting)), uint32(unsafe.Sizeof(accounting)), nil); err != nil {
		return o, err
	}
	if accounting.ActiveProcesses == 0 {
		o.State = "STOPPED"
		o.Exited = true
		return o, nil
	}
	// Confirm the original birth identity when it still exists; process list and
	// ActiveProcesses remain authoritative when the original parent exits first.
	if r.PID > 0 {
		h, e := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(r.PID))
		if e == nil {
			birth, e := windowsBirth(h)
			same, e2 := processInJob(h, job)
			windows.CloseHandle(h)
			if e != nil || e2 != nil {
				return o, ErrUnknown
			}
			if same != (birth == r.StartTicks) {
				return o, ErrUnknown
			}
		} else if !errors.Is(e, windows.ERROR_INVALID_PARAMETER) {
			return o, e
		}
	}
	capacity := accounting.ActiveProcesses + 32
	if capacity > 1<<20 {
		return o, errors.New("Job process count exceeds inspection budget")
	}
	// Two DWORD counts followed by ULONG_PTR process IDs.
	buffer := make([]uintptr, int(capacity)+2)
	if err = windows.QueryInformationJobObject(job, windows.JobObjectBasicProcessIdList, uintptr(unsafe.Pointer(&buffer[0])), uint32(len(buffer))*uint32(unsafe.Sizeof(buffer[0])), nil); err != nil {
		return o, err
	}
	counts := (*[2]uint32)(unsafe.Pointer(&buffer[0]))
	count := counts[1]
	if count > capacity {
		return o, ErrUnknown
	}
	base := unsafe.Add(unsafe.Pointer(&buffer[0]), 8)
	for i := uint32(0); i < count; i++ {
		if err = ctx.Err(); err != nil {
			return o, err
		}
		pid := *(*uintptr)(unsafe.Add(base, uintptr(i)*unsafe.Sizeof(uintptr(0))))
		if pid == 0 || uint64(pid) > uint64(^uint32(0)) {
			return o, ErrUnknown
		}
		h, e := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
		if errors.Is(e, windows.ERROR_INVALID_PARAMETER) {
			continue
		}
		if e != nil {
			return o, e
		}
		birth, birthErr := windowsBirth(h)
		owned, ownedErr := processInJob(h, job)
		windows.CloseHandle(h)
		if birthErr != nil || ownedErr != nil || !owned {
			return o, ErrUnknown
		}
		o.Processes = append(o.Processes, Process{PID: int(pid), StartTicks: birth})
	}
	o.State = "RUNNING"
	if r.Phase == "suspended" {
		// A daemon may have crashed between the durable suspended checkpoint and
		// ResumeThread/running commit. Do not claim execution, or resume it again.
		// The verified Job can still be explicitly stopped to resolve the run.
		o.State = "UNKNOWN"
		o.Diagnostic = "launch was interrupted around the suspended execution barrier; stop this confirmed Job before another start"
	}
	return o, nil
}
func (m *Manager) signalNative(ctx context.Context, r Record, force bool) error {
	o, err := m.observeNative(ctx, r)
	if err != nil || o.Exited {
		return err
	}
	job, closeIt, err := m.job(r)
	if err != nil {
		return err
	}
	if closeIt {
		defer windows.CloseHandle(job)
	}
	if force {
		return windows.TerminateJobObject(job, 1)
	}
	// Console control requires a console shared with the daemon. Service-mode
	// programs should configure an application stop input. Failure is explicit.
	if err = windows.GenerateConsoleCtrlEvent(windows.CTRL_BREAK_EVENT, uint32(r.PGID)); err != nil {
		return fmt.Errorf("console stop unavailable; configure application stop input: %w", err)
	}
	return nil
}
func (m *Manager) cleanupNative(r Record) error {
	m.mu.Lock()
	live := m.live[r.RunID]
	delete(m.live, r.RunID)
	m.mu.Unlock()
	if live != nil {
		if p, ok := live.platform.(*windowsRun); ok {
			return windows.CloseHandle(p.job)
		}
	}
	return nil
}

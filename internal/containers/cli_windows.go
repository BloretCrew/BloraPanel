//go:build windows

package containers

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
	"io"
	"os"
	"os/exec"
	"sort"
	"sync"
	"time"
	"unsafe"
)

func syncDir(string) error                     { return nil }
func recoverCLI(context.Context, string) error { return nil } // Atomic Job membership and kill-on-close handle executor death.
func runCLI(ctx context.Context, command []string, dir, stateRoot string, env []string, stdout, stderr io.Writer) error {
	if len(command) == 0 {
		return ErrCapability
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(job)
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err = windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits))); err != nil {
		return err
	}
	attrs, err := windows.NewProcThreadAttributeList(2)
	if err != nil {
		return err
	}
	defer attrs.Delete()
	if err = attrs.Update(0x0002000d, unsafe.Pointer(&job), unsafe.Sizeof(job)); err != nil {
		return fmt.Errorf("atomic CLI Job assignment: %w", err)
	}
	outRead, outWrite, err := os.Pipe()
	if err != nil {
		return err
	}
	defer outRead.Close()
	defer outWrite.Close()
	errRead, errWrite, err := os.Pipe()
	if err != nil {
		return err
	}
	defer errRead.Close()
	defer errWrite.Close()
	null, err := os.OpenFile(os.DevNull, os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer null.Close()
	files := []*os.File{null, outWrite, errWrite}
	handles := make([]windows.Handle, 3)
	for i, file := range files {
		if err = windows.DuplicateHandle(windows.CurrentProcess(), windows.Handle(file.Fd()), windows.CurrentProcess(), &handles[i], 0, true, windows.DUPLICATE_SAME_ACCESS); err != nil {
			return err
		}
		defer windows.CloseHandle(handles[i])
	}
	if err = attrs.Update(windows.PROC_THREAD_ATTRIBUTE_HANDLE_LIST, unsafe.Pointer(&handles[0]), uintptr(len(handles))*unsafe.Sizeof(handles[0])); err != nil {
		return err
	}
	startup := windows.StartupInfoEx{}
	startup.Cb = uint32(unsafe.Sizeof(startup))
	startup.Flags = windows.STARTF_USESTDHANDLES
	startup.StdInput = handles[0]
	startup.StdOutput = handles[1]
	startup.StdErr = handles[2]
	startup.ProcThreadAttributeList = attrs.List()
	executable, err := exec.LookPath(command[0])
	if err != nil {
		return err
	}
	app, err := windows.UTF16PtrFromString(executable)
	if err != nil {
		return err
	}
	line, err := windows.UTF16PtrFromString(windows.ComposeCommandLine(append([]string{executable}, command[1:]...)))
	if err != nil {
		return err
	}
	directory, err := windows.UTF16PtrFromString(dir)
	if err != nil {
		return err
	}
	sort.Strings(env)
	var block []uint16
	for _, item := range env {
		value, e := windows.UTF16FromString(item)
		if e != nil {
			return e
		}
		block = append(block, value...)
	}
	block = append(block, 0)
	var process windows.ProcessInformation
	flags := uint32(windows.CREATE_SUSPENDED | windows.CREATE_NEW_PROCESS_GROUP | windows.EXTENDED_STARTUPINFO_PRESENT | windows.CREATE_UNICODE_ENVIRONMENT)
	if err = windows.CreateProcess(app, line, nil, nil, true, flags, &block[0], directory, &startup.StartupInfo, &process); err != nil {
		return err
	}
	defer windows.CloseHandle(process.Process)
	defer windows.CloseHandle(process.Thread)
	for i, h := range handles {
		windows.CloseHandle(h)
		handles[i] = 0
	}
	outWrite.Close()
	errWrite.Close()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); _, _ = io.Copy(stdout, outRead) }()
	go func() { defer wg.Done(); _, _ = io.Copy(stderr, errRead) }()
	defer func() { outRead.Close(); errRead.Close(); wg.Wait() }()
	if _, err = windows.ResumeThread(process.Thread); err != nil {
		return err
	}
	for {
		state, e := windows.WaitForSingleObject(process.Process, 50)
		if e != nil {
			return e
		}
		if ctx.Err() != nil {
			_ = windows.TerminateJobObject(job, 1)
			deadline := time.Now().Add(2 * time.Second)
			for time.Now().Before(deadline) {
				empty, e := cliJobEmpty(job)
				if e != nil {
					return errors.Join(ctx.Err(), e)
				}
				if empty {
					return ctx.Err()
				}
				time.Sleep(20 * time.Millisecond)
			}
			return errors.Join(ctx.Err(), ErrUnknown)
		}
		if state == windows.WAIT_OBJECT_0 {
			empty, e := cliJobEmpty(job)
			if e != nil {
				return e
			}
			if !empty {
				continue
			}
			var code uint32
			if err = windows.GetExitCodeProcess(process.Process, &code); err != nil {
				return err
			}
			if code != 0 {
				return fmt.Errorf("Compose CLI exit status %d", code)
			}
			return nil
		}
	}
}
func cliJobEmpty(job windows.Handle) (bool, error) {
	var accounting struct {
		Times                                 [4]int64
		PageFaults, Total, Active, Terminated uint32
	}
	err := windows.QueryInformationJobObject(job, windows.JobObjectBasicAccountingInformation, uintptr(unsafe.Pointer(&accounting)), uint32(unsafe.Sizeof(accounting)), nil)
	return accounting.Active == 0, err
}

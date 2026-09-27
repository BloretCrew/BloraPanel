//go:build windows

package monitor

import (
	"errors"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Unix microseconds stay exact in JavaScript numbers. As with Linux clock
// ticks, this is a process-creation token, not a PID or a runtime generation.
func processBirth(h windows.Handle) (uint64, error) {
	var birth, exit, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(h, &birth, &exit, &kernel, &user); err != nil {
		return 0, err
	}
	return uint64(birth.Nanoseconds() / 1000), nil
}
func readProcessDetails(pid int) (SystemProcess, error) {
	p := SystemProcess{PID: pid}
	if pid <= 0 || uint64(pid) > uint64(^uint32(0)) {
		return p, errors.New("invalid process identifier")
	}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return p, err
	}
	defer windows.CloseHandle(h)
	p.StartTicks, err = processBirth(h)
	if err != nil {
		return p, err
	}
	var path [32768]uint16
	size := uint32(len(path))
	if windows.QueryFullProcessImageName(h, 0, &path[0], &size) == nil {
		p.Command = windows.UTF16ToString(path[:size])
	}
	p.RSS, err = processRSS(h)
	if err != nil {
		p.Unavailable = append(p.Unavailable, "memory")
	}
	var token windows.Token
	if err := windows.OpenProcessToken(h, windows.TOKEN_QUERY, &token); err == nil {
		defer token.Close()
		if user, err := token.GetTokenUser(); err == nil {
			p.UID = user.User.Sid.String()
		}
	}
	if p.UID == "" {
		p.Unavailable = append(p.Unavailable, "user")
	}
	return p, nil
}
func visitProcesses(visit func(SystemProcess)) error {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(snapshot)
	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	err = windows.Process32First(snapshot, &entry)
	for err == nil {
		if entry.ProcessID > 0 {
			p, detailErr := readProcessDetails(int(entry.ProcessID))
			if detailErr != nil {
				p = SystemProcess{PID: int(entry.ProcessID), Unavailable: []string{"identity", "memory", "user"}}
			}
			if p.Command == "" {
				p.Command = windows.UTF16ToString(entry.ExeFile[:])
			}
			visit(p)
		}
		err = windows.Process32Next(snapshot, &entry)
	}
	if err != nil && !errors.Is(err, windows.ERROR_NO_MORE_FILES) {
		return err
	}
	return nil
}
func Terminate(pid int, start uint64) error {
	if pid <= 4 || uint64(pid) > uint64(^uint32(0)) || pid == os.Getpid() || start == 0 {
		return errors.New("invalid or protected process target")
	}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.PROCESS_TERMINATE|windows.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	birth, err := processBirth(h)
	if err != nil {
		return err
	}
	if birth != start {
		return errors.New("process identity changed")
	}
	if err := windows.TerminateProcess(h, 1); err != nil {
		return err
	}
	result, err := windows.WaitForSingleObject(h, 10000)
	if err != nil {
		return err
	}
	if result != windows.WAIT_OBJECT_0 {
		return errors.New("process exit not confirmed before deadline")
	}
	return nil
}

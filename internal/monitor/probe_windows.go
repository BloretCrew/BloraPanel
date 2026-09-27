//go:build windows

package monitor

import (
	"errors"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var kernelMonitor = windows.NewLazySystemDLL("kernel32.dll")
var systemTimes = kernelMonitor.NewProc("GetSystemTimes")
var memoryStatus = kernelMonitor.NewProc("GlobalMemoryStatusEx")
var processMemory = kernelMonitor.NewProc("K32GetProcessMemoryInfo")
var processorGroups = kernelMonitor.NewProc("GetActiveProcessorGroupCount")

func callMonitor(proc *windows.LazyProc, args ...uintptr) error {
	if err := proc.Find(); err != nil {
		return err
	}
	result, _, err := proc.Call(args...)
	if result != 0 {
		return nil
	}
	if err == windows.ERROR_SUCCESS {
		return errors.New("Windows monitoring API failed")
	}
	return err
}
func fileTicks(t windows.Filetime) uint64 { return uint64(t.HighDateTime)<<32 | uint64(t.LowDateTime) }

type memoryStatusEx struct {
	Length, Load                                                                                                                 uint32
	TotalPhysical, AvailablePhysical, TotalPageFile, AvailablePageFile, TotalVirtual, AvailableVirtual, AvailableExtendedVirtual uint64
}
type processMemoryCounters struct {
	Size, PageFaults                                                                                             uint32
	PeakWorkingSet, WorkingSet, PeakPagedPool, PagedPool, PeakNonPagedPool, NonPagedPool, PageFile, PeakPageFile uintptr
}

func processRSS(handle windows.Handle) (int64, error) {
	var counters processMemoryCounters
	counters.Size = uint32(unsafe.Sizeof(counters))
	err := callMonitor(processMemory, uintptr(handle), uintptr(unsafe.Pointer(&counters)), uintptr(counters.Size))
	return int64(counters.WorkingSet), err
}

func (c *Collector) Node(root string) (Point, error) {
	p := Point{ObservedAt: time.Now().UTC()}
	var idle, kernel, user windows.Filetime
	groups := uintptr(0)
	if processorGroups.Find() == nil {
		groups, _, _ = processorGroups.Call()
	}
	if groups > 1 {
		// GetSystemTimes only covers the calling processor group in this case.
		p.Unavailable = append(p.Unavailable, "cpu", "cpu_multiple_processor_groups")
	} else if err := callMonitor(systemTimes, uintptr(unsafe.Pointer(&idle)), uintptr(unsafe.Pointer(&kernel)), uintptr(unsafe.Pointer(&user))); err != nil {
		p.Unavailable = append(p.Unavailable, "cpu")
	} else {
		total, idleTicks := fileTicks(kernel)+fileTicks(user), fileTicks(idle)
		c.mu.Lock()
		previous := c.previous
		c.previous = cpuTicks{total: total, idle: idleTicks}
		c.mu.Unlock()
		if previous.total > 0 && total > previous.total && idleTicks >= previous.idle && idleTicks-previous.idle <= total-previous.total {
			p.CPUPercent = 100 * float64(total-previous.total-(idleTicks-previous.idle)) / float64(total-previous.total)
		} else {
			p.Unavailable = append(p.Unavailable, "cpu")
		}
	}
	var memory memoryStatusEx
	memory.Length = uint32(unsafe.Sizeof(memory))
	if err := callMonitor(memoryStatus, uintptr(unsafe.Pointer(&memory))); err != nil {
		p.Unavailable = append(p.Unavailable, "memory")
	} else {
		p.MemoryTotal = int64(memory.TotalPhysical)
		p.MemoryUsed = int64(memory.TotalPhysical - memory.AvailablePhysical)
	}
	path, err := windows.UTF16PtrFromString(root)
	var available, total, free uint64
	if err != nil || windows.GetDiskFreeSpaceEx(path, &available, &total, &free) != nil {
		p.Unavailable = append(p.Unavailable, "disk")
	} else {
		p.DiskTotal = int64(total)
		p.DiskUsed = int64(total - available)
	}
	var table *windows.MibIfTable2
	if err := windows.GetIfTable2Ex(windows.MibIfTableNormal, &table); err != nil {
		p.Unavailable = append(p.Unavailable, "network")
	} else if table != nil {
		defer windows.FreeMibTable(unsafe.Pointer(table))
		if table.NumEntries > 65536 {
			p.Unavailable = append(p.Unavailable, "network")
		} else {
			for _, row := range unsafe.Slice(&table.Table[0], int(table.NumEntries)) {
				if row.Type != 24 {
					p.NetworkRx += row.InOctets
					p.NetworkTx += row.OutOctets
				}
			}
		}
	} else {
		p.Unavailable = append(p.Unavailable, "network")
	}
	return c.add(p), nil
}

func (c *Collector) Process(pid int, runID string, expectedBirth ...uint64) (ProcessPoint, error) {
	p := ProcessPoint{Point: Point{ObservedAt: time.Now().UTC(), CPUBasis: "host-total", Unavailable: []string{"network", "disk"}}, PID: pid, RunID: runID}
	if pid <= 0 || uint64(pid) > uint64(^uint32(0)) {
		return p, errors.New("owned process has no host pid")
	}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return p, err
	}
	defer windows.CloseHandle(h)
	p.RSS, err = processRSS(h)
	if err != nil {
		p.Unavailable = append(p.Unavailable, "memory")
	}
	var birth, exit, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(h, &birth, &exit, &kernel, &user); err != nil {
		p.Unavailable = append(p.Unavailable, "cpu")
		return p, nil
	}
	if fileTicks(exit) != 0 {
		return p, errors.New("process has exited")
	}
	if len(expectedBirth) > 0 && expectedBirth[0] != 0 && fileTicks(birth) != expectedBirth[0] {
		return p, errors.New("process birth changed")
	}
	current := processTicks{birth: fileTicks(birth), total: fileTicks(kernel) + fileTicks(user), at: p.ObservedAt, runID: runID}
	c.mu.Lock()
	previous, exists := c.processPrevious[pid]
	if c.processPrevious == nil || (!exists && len(c.processPrevious) >= 1024) {
		c.processPrevious = map[int]processTicks{}
	}
	c.processPrevious[pid] = current
	c.mu.Unlock()
	if previous.birth == current.birth && previous.runID == runID && current.total >= previous.total && current.at.After(previous.at) {
		cpus := windows.GetActiveProcessorCount(0xffff)
		if cpus > 0 {
			p.CPUPercent = 100 * float64(current.total-previous.total) / 1e7 / current.at.Sub(previous.at).Seconds() / float64(cpus)
		} else {
			p.Unavailable = append(p.Unavailable, "cpu")
		}
	} else {
		p.Unavailable = append(p.Unavailable, "cpu")
	}
	return p, nil
}

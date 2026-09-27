//go:build linux

package monitor

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func (c *Collector) Node(root string) (Point, error) {
	p := Point{ObservedAt: time.Now().UTC()}
	unavailable := []string{}
	total, idle, err := readCPU()
	if err != nil {
		unavailable = append(unavailable, "cpu")
	} else {
		c.mu.Lock()
		previous := c.previous
		c.previous = cpuTicks{total, idle}
		c.mu.Unlock()
		if previous.total > 0 && total > previous.total {
			p.CPUPercent = 100 * float64((total-previous.total)-(idle-previous.idle)) / float64(total-previous.total)
		}
	}
	memTotal, memAvailable, err := readMem()
	if err != nil {
		unavailable = append(unavailable, "memory")
	} else {
		p.MemoryTotal = memTotal
		p.MemoryUsed = memTotal - memAvailable
	}
	var fs syscall.Statfs_t
	if err := syscall.Statfs(root, &fs); err != nil {
		unavailable = append(unavailable, "disk")
	} else {
		p.DiskTotal = int64(fs.Blocks) * int64(fs.Bsize)
		p.DiskUsed = (int64(fs.Blocks) - int64(fs.Bavail)) * int64(fs.Bsize)
	}
	rx, tx, err := readNet()
	if err != nil {
		unavailable = append(unavailable, "network")
	} else {
		p.NetworkRx = rx
		p.NetworkTx = tx
	}
	p.Unavailable = unavailable
	return c.add(p), nil
}
func (c *Collector) Process(pid int, runID string, expectedBirth ...uint64) (ProcessPoint, error) {
	p := ProcessPoint{Point: Point{ObservedAt: time.Now().UTC(), CPUBasis: "host-total", Unavailable: []string{"network", "disk"}}, PID: pid, RunID: runID}
	if pid <= 0 {
		return p, errors.New("owned process has no host pid")
	}
	before, err := readProcessCPU(pid)
	if err != nil {
		return p, err
	}
	if len(expectedBirth) > 0 && expectedBirth[0] != 0 && before.birth != expectedBirth[0] {
		return p, errors.New("process birth changed")
	}
	f, err := os.Open(filepath.Join("/proc", strconv.Itoa(pid), "status"))
	if err != nil {
		return p, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	foundMemory := false
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[0] == "VmRSS:" {
			kb, parseErr := strconv.ParseInt(fields[1], 10, 64)
			if parseErr == nil && kb >= 0 && kb <= 9007199254740991/1024 {
				p.RSS = kb << 10
				foundMemory = true
			}
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return p, err
	}
	if !foundMemory {
		p.Unavailable = append(p.Unavailable, "memory")
	}
	current, err := readProcessCPU(pid)
	if err != nil {
		return p, err
	}
	if current.birth != before.birth {
		return p, errors.New("process birth changed during sampling")
	}
	current.system, _, err = readCPU()
	if err != nil {
		p.Unavailable = append(p.Unavailable, "cpu")
		return p, nil
	}
	current.runID = runID
	c.mu.Lock()
	previous, exists := c.processPrevious[pid]
	if c.processPrevious == nil || (!exists && len(c.processPrevious) >= 1024) {
		c.processPrevious = map[int]processTicks{}
	}
	c.processPrevious[pid] = current
	c.mu.Unlock()
	if previous.birth == current.birth && previous.runID == runID && current.total >= previous.total && current.system > previous.system {
		p.CPUPercent = 100 * float64(current.total-previous.total) / float64(current.system-previous.system)
	} else {
		p.Unavailable = append(p.Unavailable, "cpu")
	}
	return p, nil
}
func readProcessCPU(pid int) (processTicks, error) {
	var p processTicks
	b, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
	if err != nil {
		return p, err
	}
	end := strings.LastIndexByte(string(b), ')')
	if end < 0 {
		return p, errors.New("invalid process stat")
	}
	f := strings.Fields(string(b[end+1:]))
	if len(f) < 20 || f[0] == "Z" || f[0] == "X" {
		return p, errors.New("process unavailable")
	}
	u, err := strconv.ParseUint(f[11], 10, 64)
	if err != nil {
		return p, err
	}
	s, err := strconv.ParseUint(f[12], 10, 64)
	if err != nil {
		return p, err
	}
	p.birth, err = strconv.ParseUint(f[19], 10, 64)
	if err != nil {
		return p, err
	}
	if u > ^uint64(0)-s {
		return p, errors.New("process CPU counter overflow")
	}
	p.total = u + s
	return p, nil
}
func readCPU() (uint64, uint64, error) {
	b, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	line := strings.SplitN(string(b), "\n", 2)[0]
	f := strings.Fields(line)
	if len(f) < 6 || f[0] != "cpu" {
		return 0, 0, errors.New("invalid proc stat")
	}
	var values []uint64
	for _, v := range f[1:] {
		n, e := strconv.ParseUint(v, 10, 64)
		if e != nil {
			return 0, 0, e
		}
		values = append(values, n)
	}
	var total uint64
	// guest and guest_nice are already included in user and nice.
	for _, n := range values[:min(8, len(values))] {
		if total > ^uint64(0)-n {
			return 0, 0, errors.New("CPU counter overflow")
		}
		total += n
	}
	return total, values[3] + values[4], nil
}
func readMem() (int64, int64, error) {
	b, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}
	var total, available int64
	for _, line := range strings.Split(string(b), "\n") {
		f := strings.Fields(line)
		if len(f) < 2 {
			continue
		}
		n, _ := strconv.ParseInt(f[1], 10, 64)
		switch f[0] {
		case "MemTotal:":
			total = n << 10
		case "MemAvailable:":
			available = n << 10
		}
	}
	if total == 0 {
		return 0, 0, errors.New("invalid meminfo")
	}
	return total, available, nil
}
func readNet() (uint64, uint64, error) {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	var rx, tx uint64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "lo" {
			continue
		}
		fields := strings.Fields(parts[1])
		if len(fields) < 9 {
			continue
		}
		a, e := strconv.ParseUint(fields[0], 10, 64)
		if e != nil {
			return 0, 0, e
		}
		b, e := strconv.ParseUint(fields[8], 10, 64)
		if e != nil {
			return 0, 0, e
		}
		rx += a
		tx += b
	}
	return rx, tx, scanner.Err()
}

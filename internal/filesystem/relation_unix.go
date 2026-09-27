//go:build !windows

package filesystem

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"strings"
	"sync"
)

var machineIdentityOnce sync.Once
var machineIdentity string

func relationMachineID() (string, error) {
	machineIdentityOnce.Do(func() {
		for _, name := range []string{"/etc/machine-id", "/var/lib/dbus/machine-id", "/proc/sys/kernel/random/boot_id"} {
			value, err := os.ReadFile(name)
			if err == nil && len(strings.TrimSpace(string(value))) >= 16 {
				machineIdentity = hashBytes([]byte(strings.TrimSpace(string(value))))
				break
			}
		}
	})
	if machineIdentity == "" {
		return "", ErrUnsupported
	}
	return machineIdentity, nil
}

func relationObjectID(f *os.File) (string, error) {
	var info unix.Stat_t
	if err := unix.Fstat(int(f.Fd()), &info); err != nil {
		return "", err
	}
	var volume unix.Statfs_t
	if err := unix.Fstatfs(int(f.Fd()), &volume); err != nil {
		return "", err
	}
	// Filesystem IDs survive bind-mount aliases and distinct Daemon processes.
	// Include the filesystem type so unrelated implementations do not share an
	// identity namespace. A zero FSID uses the machine-local device namespace.
	fsid := fmt.Sprintf("%v:%d", volume.Fsid, volume.Type)
	if volume.Fsid == (unix.Fsid{}) {
		machine, err := relationMachineID()
		if err != nil {
			return "", err
		}
		fsid = fmt.Sprintf("%s:%d:%d", machine, info.Dev, volume.Type)
	}
	return hashBytes([]byte(fmt.Sprintf("unix:%s:%d", fsid, info.Ino))), nil
}

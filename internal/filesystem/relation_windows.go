package filesystem

import (
	"fmt"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"os"
	"sync"
	"unsafe"
)

var machineIdentityOnce sync.Once
var machineIdentity string

func relationMachineID() (string, error) {
	machineIdentityOnce.Do(func() {
		key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Cryptography`, registry.QUERY_VALUE|registry.WOW64_64KEY)
		if err != nil {
			return
		}
		defer key.Close()
		value, _, err := key.GetStringValue("MachineGuid")
		if err == nil && value != "" {
			machineIdentity = hashBytes([]byte(value))
		}
	})
	if machineIdentity == "" {
		return "", ErrUnsupported
	}
	return machineIdentity, nil
}

func relationObjectID(f *os.File) (string, error) {
	var info struct {
		Volume uint64
		FileID [16]byte
	}
	if err := windows.GetFileInformationByHandleEx(windows.Handle(f.Fd()), windows.FileIdInfo, (*byte)(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info))); err != nil {
		return "", err
	}
	return hashBytes([]byte(fmt.Sprintf("windows:%016x:%x", info.Volume, info.FileID))), nil
}

//go:build windows

package core

import (
	"syscall"
	"unsafe"
)

var (
	kernel32Process = syscall.NewLazyDLL("kernel32.dll")
	openProcess     = kernel32Process.NewProc("OpenProcess")
	getExitCode     = kernel32Process.NewProc("GetExitCodeProcess")
	closeHandle     = kernel32Process.NewProc("CloseHandle")
)

const (
	processQueryLimitedInformation = 0x1000
	stillActive                    = 259
)

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	h, _, callErr := openProcess.Call(processQueryLimitedInformation, 0, uintptr(pid))
	if h == 0 {
		// Access denied does not mean the process is dead (for example when the
		// other instance is elevated). Be conservative and keep the lock.
		if errno, ok := callErr.(syscall.Errno); ok && errno == syscall.ERROR_ACCESS_DENIED {
			return true
		}
		return false
	}
	defer closeHandle.Call(h)
	var code uint32
	r, _, _ := getExitCode.Call(h, uintptr(unsafe.Pointer(&code)))
	return r != 0 && code == stillActive
}

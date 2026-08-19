//go:build windows

package core

import (
	"syscall"
	"unsafe"
)

var (
	kernel32Replace = syscall.NewLazyDLL("kernel32.dll")
	moveFileExW     = kernel32Replace.NewProc("MoveFileExW")
)

const (
	movefileReplaceExisting = 0x1
	movefileWriteThrough    = 0x8
)

func atomicReplace(src, dst string) error {
	srcp, err := syscall.UTF16PtrFromString(src)
	if err != nil {
		return err
	}
	dstp, err := syscall.UTF16PtrFromString(dst)
	if err != nil {
		return err
	}
	r, _, callErr := moveFileExW.Call(
		uintptr(unsafe.Pointer(srcp)),
		uintptr(unsafe.Pointer(dstp)),
		movefileReplaceExisting|movefileWriteThrough,
	)
	if r == 0 {
		return callErr
	}
	return nil
}

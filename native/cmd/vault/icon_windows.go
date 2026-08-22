//go:build windows

package main

import (
	_ "embed"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

//go:embed assets/zervyra.ico.b64
var embeddedIconBase64 string

var bigIcon, smallIcon uintptr

func installWindowIcon(h uintptr) {
	iconBase64 := strings.TrimSpace(embeddedIconBase64)
	if h == 0 || iconBase64 == "" {
		return
	}
	embeddedIcon, err := base64.StdEncoding.DecodeString(iconBase64)
	if err != nil || len(embeddedIcon) < 6 {
		logEvent("embedded icon decode failed: %v", err)
		return
	}
	if embeddedIcon[0] != 0 || embeddedIcon[1] != 0 || embeddedIcon[2] != 1 || embeddedIcon[3] != 0 {
		logEvent("embedded icon has invalid ICO header")
		return
	}
	dir := filepath.Join(dataDir(), "cache")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return
	}
	path := filepath.Join(dir, "zervyra.ico")
	// The icon is public branding, not vault data. Writing it here lets the
	// dependency-free binary use the same multi-size ICO without a sidecar file.
	if err := os.WriteFile(path, embeddedIcon, 0600); err != nil {
		return
	}
	p := uintptr(unsafe.Pointer(w(path)))
	bigIcon, _, _ = pLoadImage.Call(0, p, IMAGE_ICON, 32, 32, LR_LOADFROMFILE)
	smallIcon, _, _ = pLoadImage.Call(0, p, IMAGE_ICON, 16, 16, LR_LOADFROMFILE)
	if bigIcon != 0 {
		pSendMessage.Call(h, WM_SETICON, ICON_BIG, bigIcon)
	}
	if smallIcon != 0 {
		pSendMessage.Call(h, WM_SETICON, ICON_SMALL, smallIcon)
	}
}

func destroyWindowIcons() {
	if bigIcon != 0 {
		pDestroyIcon.Call(bigIcon)
		bigIcon = 0
	}
	if smallIcon != 0 {
		pDestroyIcon.Call(smallIcon)
		bigIcon = 0
		smallIcon = 0
	}
}

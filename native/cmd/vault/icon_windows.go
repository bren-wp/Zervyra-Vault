//go:build windows

package main

import (
	_ "embed"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unsafe"
)

//go:embed assets/zervyra.ico.b64.part01
var embeddedIconBase64Part01 string

//go:embed assets/zervyra.ico.b64.part02
var embeddedIconBase64Part02 string

//go:embed assets/zervyra.ico.b64.part03
var embeddedIconBase64Part03 string

//go:embed assets/zervyra.ico.b64.part04
var embeddedIconBase64Part04 string

//go:embed assets/zervyra.ico.b64.part05
var embeddedIconBase64Part05 string

//go:embed assets/zervyra.ico.b64.part06
var embeddedIconBase64Part06 string

//go:embed assets/zervyra.ico.b64.part07
var embeddedIconBase64Part07 string

var bigIcon, smallIcon uintptr

func normalizeEmbeddedBase64(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\uFEFF' || unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

func installWindowIcon(h uintptr) {
	embeddedIconBase64 := normalizeEmbeddedBase64(
		embeddedIconBase64Part01 + embeddedIconBase64Part02 + embeddedIconBase64Part03 +
			embeddedIconBase64Part04 + embeddedIconBase64Part05 + embeddedIconBase64Part06 + embeddedIconBase64Part07,
	)
	if h == 0 || len(embeddedIconBase64) == 0 {
		return
	}
	embeddedIcon, err := base64.StdEncoding.DecodeString(embeddedIconBase64)
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
		smallIcon = 0
	}
}

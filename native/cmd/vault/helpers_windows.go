//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

func w(s string) *uint16 { return syscall.StringToUTF16Ptr(s) }

func rgb(r, g, b byte) uintptr { return uintptr(r) | uintptr(g)<<8 | uintptr(b)<<16 }

func multiSz(parts ...string) []uint16 {
	out := make([]uint16, 0, 128)
	for _, p := range parts {
		u, _ := syscall.UTF16FromString(p)
		out = append(out, u...)
	}
	out = append(out, 0)
	return out
}

func logEvent(format string, args ...any) {
	defer func() { _ = recover() }()
	dir := dataDir()
	logDir := filepath.Join(dir, "logs")
	_ = os.MkdirAll(logDir, 0700)
	logPath := filepath.Join(logDir, "app.log")
	if st, err := os.Stat(logPath); err == nil && st.Size() > 1<<20 {
		_ = os.Remove(logPath + ".1")
		_ = os.Rename(logPath, logPath+".1")
	}
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = fmt.Fprintf(f, "%s "+format+"\n", append([]any{time.Now().Format(time.RFC3339)}, args...)...)
}

func rememberedVaultPath() string {
	b, err := os.ReadFile(filepath.Join(dataDir(), "last-vault.txt"))
	if err != nil || len(b) > 32768 {
		return ""
	}
	p := strings.TrimSpace(string(b))
	if p == "" || !filepath.IsAbs(p) {
		return ""
	}
	return filepath.Clean(p)
}

func rememberVaultPath(p string) {
	p = filepath.Clean(strings.TrimSpace(p))
	if p == "" || p == "." || !filepath.IsAbs(p) {
		return
	}
	dir := dataDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return
	}
	// This file contains only a path, never credentials. A direct replacement is
	// intentionally best-effort and must never block vault save/unlock operations.
	_ = os.WriteFile(filepath.Join(dir, "last-vault.txt"), []byte(p+"\n"), 0600)
}

func text(h uintptr) string {
	if h == 0 {
		return ""
	}
	n, _, _ := pGetWindowTextLen.Call(h)
	if n > 4<<20 {
		n = 4 << 20
	}
	buf := make([]uint16, n+1)
	if len(buf) == 0 {
		return ""
	}
	r, _, _ := pGetWindowText.Call(h, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf[:r])
}

func setText(h uintptr, s string) {
	if h != 0 {
		pSetWindowText.Call(h, uintptr(unsafe.Pointer(w(s))))
	}
}

func msgBox(s string, flags uintptr) int {
	r, _, _ := pMessageBox.Call(hwnd, uintptr(unsafe.Pointer(w(s))), uintptr(unsafe.Pointer(w("Zervyra Vault"))), flags)
	return int(r)
}

func info(s string) { msgBox(s, MB_ICONINFORMATION) }
func warn(s string) { msgBox(s, MB_ICONWARNING) }
func fail(s string) { logEvent("ERROR: %s", s); msgBox(s, MB_ICONERROR) }

func setStatus(s string) { setText(controls[ID_SECURITY+1000], s) }

func createControl(parent uintptr, class, caption string, exStyle, style uint32, id int) uintptr {
	originalClass := class
	if class == "BUTTON" {
		class = buttonClassName
	}
	h, _, err := pCreateWindowEx.Call(
		uintptr(exStyle), uintptr(unsafe.Pointer(w(class))), uintptr(unsafe.Pointer(w(caption))),
		uintptr(style|WS_CHILD|WS_VISIBLE), 0, 0, 10, 10, parent, uintptr(id), 0, 0,
	)
	if h == 0 {
		logEvent("CreateWindowExW failed for id=%d class=%s: %v", id, class, err)
		return 0
	}
	if guiFont != 0 {
		pSendMessage.Call(h, WM_SETFONT, guiFont, 1)
	}
	if originalClass == "EDIT" || originalClass == "LISTBOX" {
		_, _, _ = pSetWindowTheme.Call(h, uintptr(unsafe.Pointer(w("DarkMode_Explorer"))), 0)
	}
	if id != 0 {
		controls[id] = h
	}
	return h
}

func move(id int, x, y, cx, cy int32) {
	if h := controls[id]; h != 0 {
		if cx < 1 {
			cx = 1
		}
		if cy < 1 {
			cy = 1
		}
		pMoveWindow.Call(h, uintptr(x), uintptr(y), uintptr(cx), uintptr(cy), 1)
	}
}

func visible(id int, on bool) {
	if h := controls[id]; h != 0 {
		cmd := uintptr(SW_HIDE)
		if on {
			cmd = SW_SHOW
		}
		pShowWindow.Call(h, cmd)
	}
}

func enabled(id int, on bool) {
	v := uintptr(0)
	if on {
		v = 1
	}
	if h := controls[id]; h != 0 {
		pEnableWindow.Call(h, v)
	}
}

func checked(id int) bool {
	r, _, _ := pSendMessage.Call(controls[id], BM_GETCHECK, 0, 0)
	return r == BST_CHECKED
}

func setChecked(id int, on bool) {
	v := uintptr(0)
	if on {
		v = BST_CHECKED
	}
	pSendMessage.Call(controls[id], BM_SETCHECK, v, 0)
}

func setPasswordMask(id int, visible bool) {
	ch := uintptr('●')
	if visible {
		ch = 0
	}
	pSendMessage.Call(controls[id], EM_SETPASSWORDCHAR, ch, 0)
	pInvalidateRect.Call(controls[id], 0, 1)
}

func touchActivity() { lastActivity = time.Now() }

func markDirty() {
	dirty = true
	lastEditAt = time.Now()
	autosaveError = ""
}

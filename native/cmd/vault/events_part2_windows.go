//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

func wndProc(h uintptr, m uint32, wp, lp uintptr) (ret uintptr) {
	defer func() {
		if r := recover(); r != nil {
			logEvent("PANIC in wndProc message=0x%x: %v", m, r)
			msgBox(fmt.Sprintf("Neočekivana greška u sučelju.\n\n%v\n\nDetalji su zapisani u app.log.", r), MB_ICONERROR)
			ret = 0
		}
	}()
	switch m {
	case WM_PAINT:
		paintMain(h)
		return 0
	case WM_ERASEBKGND:
		return 1
	case WM_CTLCOLOREDIT, WM_CTLCOLORLISTBOX:
		pSetTextColor.Call(wp, rgb(236, 240, 247))
		pSetBkColor.Call(wp, rgb(25, 33, 46))
		return fieldBrush
	case WM_CTLCOLORSTATIC:
		idv, _, _ := pGetDlgCtrlID.Call(lp)
		id := int(idv)
		if id == ID_BRAND_TITLE || id == ID_BRAND_SUBTITLE || id == ID_SECURITY+1000 {
			pSetBkColor.Call(wp, rgb(10, 13, 19))
			if id == ID_BRAND_TITLE {
				pSetTextColor.Call(wp, rgb(247, 249, 253))
			} else {
				pSetTextColor.Call(wp, rgb(135, 148, 168))
			}
			pSetBkMode.Call(wp, TRANSPARENT)
			return bgBrush
		}
		pSetBkColor.Call(wp, rgb(17, 23, 33))
		if id == ID_LIST_HEADING || id == ID_DETAIL_HEADING || id == ID_LOCK_HEADING {
			pSetTextColor.Call(wp, rgb(237, 241, 248))
		} else if id == ID_LOCK_SUBTITLE {
			pSetTextColor.Call(wp, rgb(151, 163, 181))
		} else {
			pSetTextColor.Call(wp, rgb(159, 171, 188))
		}
		pSetBkMode.Call(wp, TRANSPARENT)
		return panelBrush
	case WM_CTLCOLORBTN:
		return panelBrush
	case WM_CREATE:
		installWindowIcon(h)
		createUI(h)
		pSetTimer.Call(h, 1, 1000, 0)
		layout()
		logEvent("application started version=%s portable=%t", version, isPortableMode())
		return 0
	case WM_SIZE:
		layout()
		if uint32(wp) == SIZE_MINIMIZED && master != "" {
			lockVault("minimizirano")
		}
		return 0
	case WM_COMMAND:
		handleCommand(wp)
		return 0
	case WM_TIMER:
		timerTick()
		return 0
	case WM_POWERBROADCAST:
		if wp == PBT_APMSUSPEND && master != "" {
			if !lockVault("Windows suspend") {
				clearClipboardIfOwned()
			}
		}
		return 1
	case WM_QUERYENDSESSION:
		if master != "" && dirty && !showTrash {
			if err := saveCurrentEntry(true); err != nil {
				logEvent("shutdown save failed: %v", err)
				return 0
			}
		}
		return 1
	case WM_ENDSESSION:
		if wp != 0 {
			clearClipboardIfOwned()
			if vaultLock != nil {
				vaultLock.Release()
				vaultLock = nil
			}
		}
		return 0
	case WM_CLOSE:
		if master != "" && dirty && !showTrash {
			if err := saveCurrentEntry(true); err != nil {
				r := msgBox("Promjene nije moguće spremiti:\n\n"+err.Error()+"\n\nZatvoriti aplikaciju bez spremanja?", MB_YESNO|MB_ICONWARNING)
				if r != IDYES {
					return 0
				}
			}
		}
		if vaultLock != nil {
			vaultLock.Release()
			vaultLock = nil
		}
	case WM_DESTROY:
		pKillTimer.Call(h, 1)
		clearClipboardIfOwned()
		if vaultLock != nil {
			vaultLock.Release()
			vaultLock = nil
		}
		destroyWindowIcons()
		destroyThemeResources()
		pPostQuit.Call(0)
		return 0
	}
	r, _, _ := pDefWindowProc.Call(h, uintptr(m), wp, lp)
	return r
}

func isPortableMode() bool {
	if strings.EqualFold(strings.TrimSpace(portableBuild), "true") {
		return true
	}
	ex, err := os.Executable()
	if err != nil {
		return false
	}
	dir := filepath.Dir(ex)
	name := strings.ToLower(filepath.Base(ex))
	if strings.Contains(name, "portable") {
		return true
	}
	if _, err := os.Stat(filepath.Join(dir, "portable.flag")); err == nil {
		return true
	}
	return false
}

func dataDir() string {
	ex, _ := os.Executable()
	if isPortableMode() {
		return filepath.Join(filepath.Dir(ex), "data")
	}
	if d := os.Getenv("LOCALAPPDATA"); d != "" {
		return filepath.Join(d, "Zervyra Vault")
	}
	return filepath.Dir(ex)
}

func defaultPath() string {
	if remembered := rememberedVaultPath(); remembered != "" {
		if _, err := os.Stat(remembered); err == nil {
			return remembered
		}
	}
	preferred := filepath.Join(dataDir(), "vault.bvault")
	if _, err := os.Stat(preferred); err == nil {
		return preferred
	}
	// Brand migration must never make an existing local vault appear "lost".
	if !isPortableMode() {
		if base := os.Getenv("LOCALAPPDATA"); base != "" {
			for _, legacy := range []string{"Velunox Vault", "Brendigo Vault"} {
				p := filepath.Join(base, legacy, "vault.bvault")
				if _, err := os.Stat(p); err == nil {
					return p
				}
			}
		}
	}
	return preferred
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			logEvent("FATAL startup/message-loop panic: %v", r)
			msgBox(fmt.Sprintf("Zervyra Vault je zaustavio neočekivanu grešku.\n\n%v\n\nDetalji su zapisani u app.log.", r), MB_ICONERROR)
		}
	}()
	// Win32 windows and their message loops are thread-affine. Keeping the whole UI
	// on one OS thread removes a class of intermittent crashes caused by goroutine migration.
	runtime.LockOSThread()
	// Best-effort Per-Monitor-V2 DPI awareness. Unsupported systems simply ignore it.
	_, _, _ = pSetProcessDPI.Call(^uintptr(3)) // DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 == -4
	initThemeResources()
	if !registerButtonClass() {
		msgBox("Zervyra Vault se ne može pokrenuti (UI inicijalizacija nije uspjela).", MB_ICONERROR)
		return
	}

	class := w("ZervyraVaultNativeV110")
	wc := WNDCLASSEX{
		CbSize:     uint32(unsafe.Sizeof(WNDCLASSEX{})),
		WndProc:    syscall.NewCallback(wndProc),
		ClassName:  uintptr(unsafe.Pointer(class)),
		Background: bgBrush,
	}
	atom, _, err := pRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 {
		logEvent("RegisterClassExW failed: %v", err)
		msgBox("Zervyra Vault se ne može pokrenuti (registracija prozora nije uspjela).", MB_ICONERROR)
		return
	}
	title := fmt.Sprintf("Zervyra Vault %s", version)
	h, _, err := pCreateWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(class)),
		uintptr(unsafe.Pointer(w(title))),
		WS_OVERLAPPEDWINDOW|WS_CLIPCHILDREN,
		80, 60, 1220, 820,
		0, 0, 0, 0,
	)
	if h == 0 {
		logEvent("CreateWindowExW main window failed: %v", err)
		msgBox("Zervyra Vault se ne može pokrenuti (kreiranje glavnog prozora nije uspjelo).", MB_ICONERROR)
		return
	}
	hwnd = h
	applyDarkWindow(h)
	pShowWindow.Call(h, SW_SHOW)
	pUpdateWindow.Call(h)

	var m MSG
	for {
		r, _, err := pGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) == -1 {
			logEvent("GetMessageW failed: %v", err)
			break
		}
		if r == 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
	}
}

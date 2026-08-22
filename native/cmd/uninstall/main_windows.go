//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var (
	user32      = syscall.NewLazyDLL("user32.dll")
	pMessageBox = user32.NewProc("MessageBoxW")
)

const (
	MB_ICONINFORMATION = 0x40
	MB_ICONWARNING     = 0x30
	MB_YESNO           = 0x4
	IDYES              = 6
)

func w(s string) *uint16 { return syscall.StringToUTF16Ptr(s) }

func msg(s string, flags uintptr) int {
	r, _, _ := pMessageBox.Call(0, uintptr(unsafe.Pointer(w(s))), uintptr(unsafe.Pointer(w("Zervyra Vault Uninstall"))), flags)
	return int(r)
}

func psQuote(s string) string { return strings.ReplaceAll(s, "'", "''") }

func main() {
	if msg("Deinstalirati Zervyra Vault?\n\nTvoj šifrirani vault i korisnički podaci neće biti obrisani.", MB_YESNO|MB_ICONWARNING) != IDYES {
		return
	}

	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		return
	}
	installDir := filepath.Join(base, "Programs", "Zervyra Vault")

	// Remove shortcuts using Windows known-folder APIs exposed through PowerShell.
	ps := `$ErrorActionPreference='SilentlyContinue';` +
		`$desktop=[Environment]::GetFolderPath('Desktop');Remove-Item -LiteralPath ([IO.Path]::Combine($desktop,'Zervyra Vault.lnk')) -Force;` +
		`$start=[IO.Path]::Combine($env:APPDATA,'Microsoft\Windows\Start Menu\Programs','Zervyra Vault.lnk');Remove-Item -LiteralPath $start -Force;`
	c := exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", ps)
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = c.Run()

	key := `HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\Zervyra Vault`
	c = exec.Command("reg.exe", "delete", key, "/f")
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = c.Run()

	msg("Zervyra Vault je deinstaliran.\n\nŠifrirani vault podaci su sačuvani u LOCALAPPDATA\\Zervyra Vault.", MB_ICONINFORMATION)

	// A running executable cannot delete itself. Use PowerShell -LiteralPath with
	// single-quote escaping instead of concatenating the path into cmd.exe syntax;
	// LOCALAPPDATA may legally contain shell metacharacters or apostrophes.
	cleanup := `Start-Sleep -Seconds 2; Remove-Item -LiteralPath '` + psQuote(installDir) + `' -Recurse -Force -ErrorAction SilentlyContinue`
	c = exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-WindowStyle", "Hidden", "-Command", cleanup)
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	c.Dir = os.TempDir()
	_ = c.Start()
}

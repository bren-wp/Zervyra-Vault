//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
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

	// A running executable cannot delete itself. Schedule directory removal after this process exits.
	script := `ping 127.0.0.1 -n 3 >nul & rmdir /S /Q "` + installDir + `"`
	c = exec.Command("cmd.exe", "/C", script)
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	// Do not keep the install directory as cmd.exe's working directory; Windows
	// otherwise may refuse to remove it after this uninstaller exits.
	c.Dir = os.TempDir()
	_ = c.Start()
}

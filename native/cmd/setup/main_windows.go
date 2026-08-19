//go:build windows && installer

package main

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

//go:embed Zervyra-Vault.exe
var app []byte

//go:embed Zervyra-Vault-Uninstall.exe
var uninstaller []byte

//go:embed Zervyra.ico
var icon []byte

var (
	version     = "1.1.0-dev"
	user32      = syscall.NewLazyDLL("user32.dll")
	kernel32    = syscall.NewLazyDLL("kernel32.dll")
	pMessageBox = user32.NewProc("MessageBoxW")
	pMoveFileEx = kernel32.NewProc("MoveFileExW")
)

const (
	MB_ICONINFORMATION = 0x40
	MB_ICONERROR       = 0x10
)

func w(s string) *uint16 { return syscall.StringToUTF16Ptr(s) }
func msg(s string, flags uintptr) {
	pMessageBox.Call(0, uintptr(unsafe.Pointer(w(s))), uintptr(unsafe.Pointer(w("Zervyra Vault Setup"))), flags)
}

func writeAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".zervyra-install-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	src, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return err
	}
	dst, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	r, _, callErr := pMoveFileEx.Call(uintptr(unsafe.Pointer(src)), uintptr(unsafe.Pointer(dst)), 0x1|0x8) // REPLACE_EXISTING | WRITE_THROUGH
	if r == 0 {
		return callErr
	}
	return nil
}

func verifyFile(path string, expected []byte) error {
	got, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	a := sha256.Sum256(got)
	b := sha256.Sum256(expected)
	if a != b {
		return fmt.Errorf("integrity verification failed for %s", filepath.Base(path))
	}
	return nil
}

func psQuote(s string) string { return strings.ReplaceAll(s, "'", "''") }

func hidden(name string, args ...string) error {
	c := exec.Command(name, args...)
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return c.Run()
}

func regAdd(key, name, typ, value string) {
	_ = hidden("reg.exe", "add", key, "/v", name, "/t", typ, "/d", value, "/f")
}

func main() {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		msg("LOCALAPPDATA nije dostupan. Instalacija nije moguća.", MB_ICONERROR)
		return
	}
	dir := filepath.Join(base, "Programs", "Zervyra Vault")
	exe := filepath.Join(dir, "Zervyra-Vault.exe")
	uninst := filepath.Join(dir, "Zervyra-Vault-Uninstall.exe")
	iconPath := filepath.Join(dir, "Zervyra.ico")

	if err := writeAtomic(exe, app); err != nil {
		msg("Instalacija aplikacije nije uspjela:\n\n"+err.Error(), MB_ICONERROR)
		return
	}
	if err := writeAtomic(uninst, uninstaller); err != nil {
		msg("Instalacija uninstallera nije uspjela:\n\n"+err.Error(), MB_ICONERROR)
		return
	}
	if err := writeAtomic(iconPath, icon); err != nil {
		msg("Instalacija ikone nije uspjela:\n\n"+err.Error(), MB_ICONERROR)
		return
	}
	if err := verifyFile(exe, app); err != nil {
		msg("Provjera instalirane aplikacije nije uspjela:\n\n"+err.Error(), MB_ICONERROR)
		return
	}
	if err := verifyFile(uninst, uninstaller); err != nil {
		msg("Provjera uninstallera nije uspjela:\n\n"+err.Error(), MB_ICONERROR)
		return
	}
	if err := verifyFile(iconPath, icon); err != nil {
		msg("Provjera ikone nije uspjela:\n\n"+err.Error(), MB_ICONERROR)
		return
	}

	ps := `$ErrorActionPreference='Stop';$ws=New-Object -ComObject WScript.Shell;` +
		`$desktop=[Environment]::GetFolderPath('Desktop');$s=$ws.CreateShortcut([IO.Path]::Combine($desktop,'Zervyra Vault.lnk'));` +
		`$s.TargetPath='` + psQuote(exe) + `';$s.WorkingDirectory='` + psQuote(dir) + `';$s.IconLocation='` + psQuote(iconPath) + `';$s.Save();` +
		`$sm=[IO.Path]::Combine($env:APPDATA,'Microsoft\Windows\Start Menu\Programs','Zervyra Vault.lnk');` +
		`$s2=$ws.CreateShortcut($sm);$s2.TargetPath='` + exe + `';$s2.WorkingDirectory='` + dir + `';$s2.IconLocation='` + iconPath + `';$s2.Save()`
	_ = hidden("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", ps)

	key := `HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\Zervyra Vault`
	regAdd(key, "DisplayName", "REG_SZ", "Zervyra Vault")
	regAdd(key, "DisplayVersion", "REG_SZ", version)
	regAdd(key, "Publisher", "REG_SZ", "Brendigo LTD.")
	regAdd(key, "InstallLocation", "REG_SZ", dir)
	regAdd(key, "DisplayIcon", "REG_SZ", iconPath)
	regAdd(key, "UninstallString", "REG_SZ", `"`+uninst+`"`)
	regAdd(key, "NoModify", "REG_DWORD", "1")
	regAdd(key, "NoRepair", "REG_DWORD", "1")

	msg(fmt.Sprintf("Zervyra Vault %s je uspješno instaliran.\n\nLokacija:\n%s", version, exe), MB_ICONINFORMATION)
	_ = exec.Command(exe).Start()
}

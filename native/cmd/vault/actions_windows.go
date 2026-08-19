//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"
	"unsafe"
	"zervyra-vault-native/internal/core"
)

type simpleError string

func (e simpleError) Error() string { return string(e) }
func errorsText(s string) error     { return simpleError(s) }

func commitVaultMutation(candidate core.Vault, successStatus string) bool {
	if err := core.Save(currentPath, master, candidate); err != nil {
		fail("Promjenu nije moguće sigurno spremiti. Izvorni podaci su zadržani u memoriji i na disku:\n\n" + err.Error())
		return false
	}
	vault = candidate
	dirty = false
	lastSavedAt = time.Now()
	autosaveError = ""
	selectedID = ""
	editingNew = false
	clearEditor()
	refreshList()
	if successStatus != "" {
		setStatus(successStatus + " • " + lastSavedAt.Format("15:04:05"))
	}
	return true
}

func moveSelectedToTrash() {
	if selectedID == "" || showTrash {
		return
	}
	if dirty {
		if err := saveCurrentEntry(true); err != nil {
			fail(err.Error())
			return
		}
	}
	if msgBox("Premjestiti odabrani zapis u koš?", MB_YESNO|MB_ICONWARNING) != IDYES {
		return
	}
	candidate := core.CloneVault(vault)
	for i, e := range candidate.Entries {
		if e.ID == selectedID {
			e.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
			candidate.Trash = append(candidate.Trash, e)
			candidate.Entries = append(candidate.Entries[:i], candidate.Entries[i+1:]...)
			commitVaultMutation(candidate, "Premješteno u koš")
			return
		}
	}
}

func restoreSelected() {
	if selectedID == "" || !showTrash {
		return
	}
	candidate := core.CloneVault(vault)
	for i, e := range candidate.Trash {
		if e.ID == selectedID {
			e.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
			candidate.Entries = append(candidate.Entries, e)
			candidate.Trash = append(candidate.Trash[:i], candidate.Trash[i+1:]...)
			if commitVaultMutation(candidate, "Zapis vraćen") {
				showTrash = false
				setChecked(ID_TRASHMODE, false)
				refreshList()
			}
			return
		}
	}
}

func restoreLastEdit() {
	if selectedID == "" || showTrash || master == "" {
		return
	}
	if dirty {
		if err := saveCurrentEntry(true); err != nil {
			fail(err.Error())
			return
		}
	}
	candidate := core.CloneVault(vault)
	for i := range candidate.Entries {
		if candidate.Entries[i].ID != selectedID {
			continue
		}
		if len(candidate.Entries[i].Revisions) == 0 {
			warn("Za ovaj zapis nema spremljene prethodne verzije.")
			return
		}
		if msgBox("Vratiti prethodnu verziju ovog zapisa?\n\nVraćaju se naziv, e-mail, korisničko ime, lozinka, URL, 2FA, tagovi i bilješke iz zadnje sigurne revizije.", MB_YESNO|MB_ICONWARNING) != IDYES {
			return
		}
		if !core.RestoreLastRevision(&candidate.Entries[i]) {
			return
		}
		id := selectedID
		if err := core.Save(currentPath, master, candidate); err != nil {
			fail("Prethodnu verziju nije moguće sigurno vratiti. Aktivni podaci nisu promijenjeni:\n\n" + err.Error())
			return
		}
		vault = candidate
		dirty = false
		lastSavedAt = time.Now()
		selectedID = id
		revisionCaptured = true
		refreshList()
		fillEditor(selectedEntry())
		revisionCaptured = true
		setStatus("Prethodna verzija zapisa vraćena • " + lastSavedAt.Format("15:04:05"))
		return
	}
}

func permanentlyDeleteSelected() {
	if selectedID == "" || !showTrash {
		return
	}
	if msgBox("Trajno obrisati odabrani zapis?\n\nOvu radnju nije moguće poništiti u aktivnom vaultu. Rotirajući šifrirani backupi i dalje čuvaju prethodne generacije.", MB_YESNO|MB_ICONWARNING) != IDYES {
		return
	}
	candidate := core.CloneVault(vault)
	for i, e := range candidate.Trash {
		if e.ID == selectedID {
			candidate.Trash = append(candidate.Trash[:i], candidate.Trash[i+1:]...)
			commitVaultMutation(candidate, "Zapis trajno uklonjen iz aktivnog vaulta")
			return
		}
	}
}

func restoreNewestBackup() {
	if master == "" || currentPath == "" {
		return
	}
	backup := currentPath + ".bak1"
	if _, err := os.Stat(backup); err != nil {
		warn("Nema dostupnog rotirajućeg backupa za ovaj vault.")
		return
	}
	if dirty && !showTrash {
		if err := saveCurrentEntry(true); err != nil {
			fail(err.Error())
			return
		}
	}
	if msgBox("Vratiti najnoviji backup (.bak1)?\n\nTrenutačno stanje će se prije vraćanja također sačuvati u backup rotaciji.", MB_YESNO|MB_ICONWARNING) != IDYES {
		return
	}
	candidate, err := core.Load(backup, master)
	if err != nil {
		fail("Backup nije moguće otvoriti:\n\n" + err.Error())
		return
	}
	if !commitVaultMutation(candidate, "Backup vraćen") {
		return
	}
	showTrash = false
	setChecked(ID_TRASHMODE, false)
	refreshList()
}

func updatePasswordScore() {
	p := text(controls[ID_PASSWORD])
	if p == "" {
		setText(controls[ID_SECURITY+1001], "Snaga lozinke: —")
		return
	}
	score := core.Score(p)
	label := "slaba"
	if score >= 80 {
		label = "odlična"
	} else if score >= 65 {
		label = "dobra"
	} else if score >= 45 {
		label = "srednja"
	}
	setText(controls[ID_SECURITY+1001], fmt.Sprintf("Snaga lozinke: %d/100 • %s", score, label))
}

func updateTOTPLabel() {
	s := strings.TrimSpace(text(controls[ID_TOTP]))
	if s == "" {
		setText(controls[ID_SECURITY+1002], "TOTP: —")
		return
	}
	code, remain, err := core.TOTPDetails(s, time.Now())
	if err != nil {
		setText(controls[ID_SECURITY+1002], "TOTP: neispravan secret")
		return
	}
	setText(controls[ID_SECURITY+1002], fmt.Sprintf("TOTP: %s • %ds", code, remain))
}

func securityReport() {
	weak := 0
	reusedEntries := 0
	missingTOTP := 0
	historyItems := 0
	revisionItems := 0
	seen := map[string]int{}
	for _, e := range vault.Entries {
		if core.Score(e.Password) < 65 {
			weak++
		}
		if e.Password != "" {
			seen[e.Password]++
		}
		if strings.TrimSpace(e.TOTP) == "" {
			missingTOTP++
		}
		historyItems += len(e.PasswordHistory)
		revisionItems += len(e.Revisions)
	}
	for _, e := range vault.Entries {
		if e.Password != "" && seen[e.Password] > 1 {
			reusedEntries++
		}
	}
	info(fmt.Sprintf("Sigurnosni pregled\n\nZapisa: %d\nSlabih/srednjih lozinki: %d\nZapisa s ponovljenom lozinkom: %d\nBez TOTP-a: %d\nStarih lozinki u povijesti: %d\nRevizija cijelih zapisa: %d\nU košu: %d\n\nPreporuka: jedinstvena lozinka od najmanje 16 znakova i 2FA gdje god je dostupno.", len(vault.Entries), weak, reusedEntries, missingTOTP, historyItems, revisionItems, len(vault.Trash)))
}

func chooseVaultFile() string {
	buf := make([]uint16, 32768)
	initial := strings.TrimSpace(text(controls[ID_PATH]))
	if initial != "" {
		u := syscall.StringToUTF16(initial)
		copy(buf, u)
	}
	filter := multiSz("Zervyra Vault (*.bvault)", "*.bvault", "Sve datoteke (*.*)", "*.*")
	of := OPENFILENAME{
		HwndOwner: hwnd, LpstrFilter: &filter[0], NFilterIndex: 1,
		LpstrFile: &buf[0], NMaxFile: uint32(len(buf)),
		LpstrTitle:  w("Otvori Zervyra Vault"),
		Flags:       OFN_EXPLORER | OFN_PATHMUSTEXIST | OFN_FILEMUSTEXIST,
		LpstrDefExt: w("bvault"),
	}
	of.LStructSize = uint32(unsafe.Sizeof(of))
	r, _, _ := pGetOpenFileName.Call(uintptr(unsafe.Pointer(&of)))
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

func chooseNewVaultDestination() string {
	buf := make([]uint16, 32768)
	base := "zervyra-vault.bvault"
	if current := strings.TrimSpace(text(controls[ID_PATH])); current != "" {
		if b := filepath.Base(current); b != "." && b != string(filepath.Separator) {
			base = b
		}
	}
	u := syscall.StringToUTF16(base)
	copy(buf, u)
	filter := multiSz("Zervyra Vault (*.bvault)", "*.bvault", "Sve datoteke (*.*)", "*.*")
	of := OPENFILENAME{
		HwndOwner: hwnd, LpstrFilter: &filter[0], NFilterIndex: 1,
		LpstrFile: &buf[0], NMaxFile: uint32(len(buf)),
		LpstrTitle:  w("Kreiraj novi Zervyra Vault"),
		Flags:       OFN_EXPLORER | OFN_PATHMUSTEXIST | OFN_OVERWRITEPROMPT,
		LpstrDefExt: w("bvault"),
	}
	of.LStructSize = uint32(unsafe.Sizeof(of))
	r, _, _ := pGetSaveFileName.Call(uintptr(unsafe.Pointer(&of)))
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

func chooseBackupDestination() string {
	buf := make([]uint16, 32768)
	base := "zervyra-backup-" + time.Now().Format("20060102-150405") + ".bvault"
	if currentPath != "" {
		base = strings.TrimSuffix(filepath.Base(currentPath), filepath.Ext(currentPath)) + "-backup-" + time.Now().Format("20060102-150405") + ".bvault"
	}
	u := syscall.StringToUTF16(base)
	copy(buf, u)
	filter := multiSz("Zervyra šifrirani backup (*.bvault)", "*.bvault", "Sve datoteke (*.*)", "*.*")
	of := OPENFILENAME{
		HwndOwner: hwnd, LpstrFilter: &filter[0], NFilterIndex: 1,
		LpstrFile: &buf[0], NMaxFile: uint32(len(buf)),
		LpstrTitle:  w("Spremi šifriranu sigurnosnu kopiju na drugi disk"),
		Flags:       OFN_EXPLORER | OFN_PATHMUSTEXIST | OFN_OVERWRITEPROMPT,
		LpstrDefExt: w("bvault"),
	}
	of.LStructSize = uint32(unsafe.Sizeof(of))
	r, _, _ := pGetSaveFileName.Call(uintptr(unsafe.Pointer(&of)))
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

func exportSafetyCopy() {
	if master == "" || currentPath == "" {
		return
	}
	if dirty && !showTrash {
		if err := saveCurrentEntry(true); err != nil {
			fail(err.Error())
			return
		}
	}
	dst := chooseBackupDestination()
	if dst == "" {
		return
	}
	if filepath.Ext(dst) == "" {
		dst += ".bvault"
	}
	if err := core.ExportVaultBackup(dst, master, vault); err != nil {
		fail("Sigurnosnu kopiju nije moguće napraviti:\n\n" + err.Error())
		return
	}
	setStatus("Šifrirana backup kopija provjerena i spremljena")
	info("Sigurnosna kopija je uspješno napravljena i provjerena.\n\nZa zaštitu od kvara diska čuvaj je na drugom fizičkom uređaju.")
}

func doOpen(create bool) {
	touchActivity()
	autosaveError = ""
	if master != "" && dirty && !showTrash {
		if err := saveCurrentEntry(true); err != nil {
			fail(err.Error())
			return
		}
	}
	p := strings.TrimSpace(text(controls[ID_PATH]))
	if create {
		if p == "" {
			p = defaultPath()
			setText(controls[ID_PATH], p)
		}
		if filepath.Ext(p) == "" {
			p += ".bvault"
			setText(controls[ID_PATH], p)
		}
	} else if p == "" || func() bool { _, err := os.Stat(p); return err != nil }() {
		p = chooseVaultFile()
		if p != "" {
			setText(controls[ID_PATH], p)
		}
	}
	if p == "" {
		warn("Odaberi vault datoteku.")
		return
	}
	if !filepath.IsAbs(p) {
		if ex, err := os.Executable(); err == nil {
			p = filepath.Join(filepath.Dir(ex), p)
		}
	}
	p = filepath.Clean(p)
	pw := text(controls[ID_MASTER])
	if pw == "" {
		warn("Upiši master lozinku.")
		return
	}
	if create && utf8.RuneCountInString(pw) < 12 {
		warn("Za novi trezor master lozinka mora imati najmanje 12 znakova. Preporuka je duga jedinstvena fraza.")
		return
	}
	if create {
		confirm := text(controls[ID_CONFIRM])
		if confirm != pw {
			warn("Master lozinke se ne podudaraju. Novi trezor nije kreiran.")
			pSetFocus.Call(controls[ID_CONFIRM])
			return
		}
		if _, statErr := os.Stat(p); statErr == nil {
			newPath := chooseNewVaultDestination()
			if newPath == "" {
				return
			}
			if filepath.Ext(newPath) == "" {
				newPath += ".bvault"
			}
			p = filepath.Clean(newPath)
			setText(controls[ID_PATH], p)
		}
	}

	sameLock := currentPath == p && vaultLock != nil
	var newLock *core.VaultLock
	var err error
	if sameLock {
		newLock = vaultLock
	} else {
		newLock, err = core.AcquireLock(p)
		if err != nil {
			fail(err.Error())
			return
		}
	}
	releaseNewOnError := !sameLock

	var newVault core.Vault
	if create {
		if _, statErr := os.Stat(p); statErr == nil {
			if releaseNewOnError {
				newLock.Release()
			}
			warn("Odabrana datoteka je u međuvremenu nastala. Zervyra je neće prepisati. Odaberi drugo ime za novi trezor.")
			return
		}
		newVault = core.NewVault()
		if err = core.CreateNew(p, pw, newVault); err != nil {
			if releaseNewOnError {
				newLock.Release()
			}
			fail(err.Error())
			return
		}
	} else {
		result, loadErr := core.LoadBest(p, pw)
		err = loadErr
		if err != nil {
			if releaseNewOnError {
				newLock.Release()
			}
			fail(err.Error())
			return
		}
		newVault = result.Vault
		if result.Recovered {
			if repairErr := core.Save(p, pw, newVault); repairErr != nil {
				logEvent("recovery opened from %s but main repair failed: %v", result.Source, repairErr)
				autosaveError = repairErr.Error()
			} else {
				logEvent("vault automatically recovered from %s", result.Source)
				autosaveError = "recovered:" + result.Source
			}
		}
	}

	if vaultLock != nil && vaultLock != newLock {
		vaultLock.Release()
	}
	vaultLock = newLock
	currentPath = p
	rememberVaultPath(p)
	master = pw
	vault = newVault
	selectedID = ""
	editingNew = false
	showTrash = false
	clearEditor()
	setText(controls[ID_PATH], p)
	suppressChange = true
	setText(controls[ID_MASTER], "")
	setText(controls[ID_CONFIRM], "")
	setChecked(ID_SHOWMASTER, false)
	setPasswordMask(ID_MASTER, false)
	suppressChange = false
	setUnlockedUI(true)
	refreshList()
	lastActivity = time.Now()
	lastLockTouch = time.Now()
	if strings.HasPrefix(autosaveError, "recovered:") {
		setStatus(fmt.Sprintf("Oporavljeno i otključano • %d zapisa", len(vault.Entries)))
		autosaveError = ""
	} else if autosaveError != "" {
		setStatus("Oporavljeno • glavni vault još nije moguće popraviti")
	} else {
		setStatus(fmt.Sprintf("Otključano • %d zapisa • automatsko spremanje uključeno", len(vault.Entries)))
	}
	lastSavedAt = time.Now()
	logEvent("vault opened path=%s entries=%d", p, len(vault.Entries))
}

func lockVault(reason string) bool {
	if master == "" {
		return true
	}
	if dirty && !showTrash {
		if err := saveCurrentEntry(true); err != nil {
			fail("Vault nije zaključan jer promjene nije moguće sigurno spremiti:\n\n" + err.Error())
			lastActivity = time.Now()
			return false
		}
	}
	clearClipboardIfOwned()
	if vaultLock != nil {
		vaultLock.Release()
		vaultLock = nil
	}
	master = ""
	vault = core.NewVault()
	runtime.GC()
	selectedID = ""
	editingNew = false
	showTrash = false
	suppressChange = true
	setText(controls[ID_MASTER], "")
	setText(controls[ID_CONFIRM], "")
	suppressChange = false
	clearEditor()
	refreshList()
	setUnlockedUI(false)
	updateLockedScreen(reason)
	logEvent("vault locked reason=%s", reason)
	return true
}

func openURL() {
	e := selectedEntry()
	if e == nil || !core.IsSafeHTTPURL(e.URL) {
		warn("Odabrani zapis nema valjan HTTP/HTTPS URL.")
		return
	}
	cmd := exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", e.URL)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Start(); err != nil {
		fail(err.Error())
	}
}

func setClipboard(s string) error {
	if s == "" {
		return errorsText("Nema sadržaja za kopiranje.")
	}
	opened := false
	for i := 0; i < 8; i++ {
		r, _, _ := pOpenClipboard.Call(hwnd)
		if r != 0 {
			opened = true
			break
		}
		time.Sleep(15 * time.Millisecond)
	}
	if !opened {
		return errorsText("Clipboard je trenutno zauzet. Pokušaj ponovno.")
	}
	defer pCloseClipboard.Call()
	pEmptyClipboard.Call()
	u := syscall.StringToUTF16(s)
	size := uintptr(len(u) * 2)
	h, _, err := pGlobalAlloc.Call(GMEM_MOVEABLE, size)
	if h == 0 {
		return err
	}
	ptr, _, err := pGlobalLock.Call(h)
	if ptr == 0 {
		pGlobalFree.Call(h)
		return err
	}
	pLstrcpyW.Call(ptr, uintptr(unsafe.Pointer(&u[0])))
	pGlobalUnlock.Call(h)
	r, _, setErr := pSetClipboardData.Call(CF_UNICODETEXT, h)
	if r == 0 {
		pGlobalFree.Call(h)
		return setErr
	}
	seq, _, _ := pGetClipboardSeq.Call()
	clipboardSeq = seq
	clipboardClearAt = time.Now().Add(20 * time.Second)
	return nil
}

func clearClipboardIfOwned() {
	if clipboardSeq == 0 {
		clipboardClearAt = time.Time{}
		return
	}
	cur, _, _ := pGetClipboardSeq.Call()
	if cur != clipboardSeq {
		clipboardSeq = 0
		clipboardClearAt = time.Time{}
		return
	}
	r, _, _ := pOpenClipboard.Call(hwnd)
	if r == 0 {
		clipboardClearAt = time.Now().Add(500 * time.Millisecond)
		return
	}
	pEmptyClipboard.Call()
	pCloseClipboard.Call()
	clipboardSeq = 0
	clipboardClearAt = time.Time{}
}

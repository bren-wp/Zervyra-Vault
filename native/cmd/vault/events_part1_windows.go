//go:build windows

package main

import (
	"syscall"
	"time"
	"unsafe"
	"zervyra-vault-native/internal/core"
)

func handleCommand(wp uintptr) {
	id := int(wp & 0xffff)
	notify := int((wp >> 16) & 0xffff)
	touchActivity()

	if notify == EN_CHANGE {
		if suppressChange {
			return
		}
		switch id {
		case ID_PATH:
			if master == "" {
				updateLockedScreen("")
			}
		case ID_SEARCH:
			if dirty && !showTrash {
				if err := saveCurrentEntry(true); err != nil {
					fail(err.Error())
					refreshList()
					return
				}
			}
			refreshList()
		case ID_PASSWORD:
			if master != "" && !showTrash {
				markDirty()
			}
			updatePasswordScore()
		case ID_TOTP:
			if master != "" && !showTrash {
				markDirty()
			}
			updateTOTPLabel()
		case ID_TITLE, ID_USERNAME, ID_EMAIL, ID_URL, ID_TAGS, ID_NOTES:
			if master != "" && !showTrash {
				markDirty()
			}
		}
		return
	}
	if notify == LBN_SELCHANGE && id == ID_LIST {
		selectFromList()
		return
	}
	if notify != BN_CLICKED {
		return
	}

	switch id {
	case ID_BROWSE:
		if p := chooseVaultFile(); p != "" {
			setText(controls[ID_PATH], p)
			updateLockedScreen("")
		}
	case ID_SHOWMASTER:
		visible := checked(ID_SHOWMASTER)
		setPasswordMask(ID_MASTER, visible)
		setPasswordMask(ID_CONFIRM, visible)
	case ID_OPEN:
		doOpen(false)
	case ID_CREATE:
		doOpen(true)
	case ID_TRASHMODE:
		if dirty && !showTrash {
			if err := saveCurrentEntry(true); err != nil {
				fail(err.Error())
				return
			}
		}
		showTrash = !showTrash
		selectedID = ""
		editingNew = false
		clearEditor()
		refreshList()
	case ID_NEW:
		if master == "" || showTrash {
			return
		}
		if dirty {
			if err := saveCurrentEntry(true); err != nil {
				fail(err.Error())
				return
			}
		}
		selectedID = ""
		editingNew = true
		clearEditor()
		setText(controls[ID_TITLE], "Novi zapis")
		markDirty()
		pSetFocus.Call(controls[ID_TITLE])
		updateModeButtons()
	case ID_SHOWPASS:
		setPasswordMask(ID_PASSWORD, checked(ID_SHOWPASS))
	case ID_FAVORITE:
		if master != "" && !showTrash {
			markDirty()
		}
	case ID_GENERATE:
		if showTrash {
			return
		}
		p := core.Generate(core.DefaultPasswordLen)
		if p == "" {
			fail("Generator lozinki nije uspio dobiti sigurne slučajne podatke.")
			return
		}
		setText(controls[ID_PASSWORD], p)
		markDirty()
		updatePasswordScore()
		if err := setClipboard(p); err != nil {
			warn(err.Error())
		} else {
			setStatus("Nova jaka lozinka generirana i kopirana • clipboard se čisti za 20 s")
		}
	case ID_SAVEENTRY:
		if err := saveCurrentEntry(true); err != nil {
			warn(err.Error())
		}
	case ID_DELETE:
		if showTrash {
			permanentlyDeleteSelected()
		} else {
			moveSelectedToTrash()
		}
	case ID_RESTORE:
		if showTrash {
			restoreSelected()
		} else {
			restoreLastEdit()
		}
	case ID_COPYUSER:
		if e := selectedEntry(); e != nil {
			if err := setClipboard(e.Username); err != nil {
				warn(err.Error())
			} else {
				setStatus("Korisničko ime kopirano")
			}
		}
	case ID_COPYEMAIL:
		if e := selectedEntry(); e != nil {
			if err := setClipboard(e.Email); err != nil {
				warn(err.Error())
			} else {
				setStatus("E-mail kopiran • clipboard se čisti za 20 s")
			}
		}
	case ID_COPYPASS:
		if e := selectedEntry(); e != nil {
			if err := setClipboard(e.Password); err != nil {
				warn(err.Error())
			} else {
				setStatus("Lozinka kopirana • clipboard se čisti za 20 s")
			}
		}
	case ID_COPYTOTP:
		if e := selectedEntry(); e != nil {
			code, _, err := core.TOTPDetails(e.TOTP, time.Now())
			if err != nil {
				warn("Odabrani zapis nema valjan TOTP secret.")
			} else if err := setClipboard(code); err != nil {
				warn(err.Error())
			} else {
				setStatus("TOTP kopiran")
			}
		}
	case ID_OPENURL:
		openURL()
	case ID_SECURITY:
		securityReport()
	case ID_SAVEVAULT:
		if dirty && !showTrash {
			if err := saveCurrentEntry(false); err != nil {
				warn(err.Error())
				return
			}
		}
		if master != "" {
			if err := core.Save(currentPath, master, vault); err != nil {
				dirty = true
				fail(err.Error())
			} else {
				dirty = false
				revisionCaptured = false
				lastSavedAt = time.Now()
				setStatus("Sigurno spremljeno • " + lastSavedAt.Format("15:04:05"))
			}
		}
	case ID_BACKUP:
		restoreNewestBackup()
	case ID_EXPORT:
		exportSafetyCopy()
	case ID_LOCK:
		lockVault("ručno")
	}
}

func timerTick() {
	// Clipboard operations stay on the Win32 UI thread; this avoids cross-thread
	// user32 access and data races from background goroutines.
	if !clipboardClearAt.IsZero() && !time.Now().Before(clipboardClearAt) {
		clearClipboardIfOwned()
	}
	if master == "" {
		return
	}
	if dirty && !showTrash && !lastEditAt.IsZero() && time.Since(lastEditAt) >= time.Second && editorHasContent() {
		if err := autosaveCurrentEntry(); err != nil {
			logEvent("autosave failed: %v", err)
			setStatus("Promjene su u memoriji • automatsko spremanje nije uspjelo")
		}
	}
	if time.Since(lastActivity) >= 5*time.Minute {
		lockVault("5 min neaktivnosti")
		return
	}
	if vaultLock != nil && time.Since(lastLockTouch) >= 30*time.Second {
		vaultLock.Touch()
		lastLockTouch = time.Now()
	}
	updateTOTPLabel()
}

func registerButtonClass() bool {
	class := w(buttonClassName)
	wc := WNDCLASSEX{
		CbSize:     uint32(unsafe.Sizeof(WNDCLASSEX{})),
		WndProc:    syscall.NewCallback(buttonWndProc),
		ClassName:  uintptr(unsafe.Pointer(class)),
		Background: 0,
	}
	atom, _, err := pRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 {
		logEvent("Register custom button class failed: %v", err)
		return false
	}
	return true
}

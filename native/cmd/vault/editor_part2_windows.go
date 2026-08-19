//go:build windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

func lockedPathExists() bool {
	p := strings.TrimSpace(text(controls[ID_PATH]))
	if p == "" {
		return false
	}
	if !filepath.IsAbs(p) {
		if ex, err := os.Executable(); err == nil {
			p = filepath.Join(filepath.Dir(ex), p)
		}
	}
	st, err := os.Stat(filepath.Clean(p))
	return err == nil && !st.IsDir()
}

func updateLockedScreen(reason string) {
	if master != "" {
		return
	}
	exists := lockedPathExists()
	if exists {
		setText(controls[ID_LOCK_HEADING], "Otključaj svoj trezor")
		setText(controls[ID_LOCK_SUBTITLE], "Trezor postoji. Upiši master lozinku za siguran pristup zapisima.")
		visible(ID_CONFIRM, false)
		enabled(ID_OPEN, true)
		enabled(ID_CREATE, true)
		status := "Trezor je sigurno zaključan"
		if reason != "" {
			status += " • " + reason
		}
		setStatus(status)
	} else {
		setText(controls[ID_LOCK_HEADING], "Dobro došli u Zervyra Vault")
		setText(controls[ID_LOCK_SUBTITLE], "Kreiraj novi šifrirani trezor ili odaberi postojeću .bvault datoteku.")
		visible(ID_CONFIRM, true)
		enabled(ID_OPEN, false)
		enabled(ID_CREATE, true)
		setStatus("Nema otvorenog trezora • podaci još nisu kreirani")
	}
	layout()
}

func layout() {
	var r RECT
	pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
	cw, ch := r.Right-r.Left, r.Bottom-r.Top
	if cw < 980 {
		cw = 980
	}
	if ch < 680 {
		ch = 680
	}
	margin := int32(18)

	move(ID_BRAND_TITLE, margin+48, 14, 260, 30)
	move(ID_BRAND_SUBTITLE, margin+50, 43, 520, 20)
	move(ID_SECURITY+1000, cw-470, 25, 440, 22)

	if master == "" {
		cardW := int32(650)
		if cardW > cw-2*margin {
			cardW = cw - 2*margin
		}
		x := (cw - cardW) / 2
		y := int32(132)
		move(ID_LOCK_HEADING, x+34, y+28, cardW-68, 32)
		move(ID_LOCK_SUBTITLE, x+36, y+68, cardW-72, 22)
		move(ID_PATH, x+34, y+116, cardW-34-34-104, 38)
		move(ID_BROWSE, x+cardW-34-96, y+116, 96, 38)
		move(ID_MASTER, x+34, y+170, cardW-34-34-94, 38)
		move(ID_SHOWMASTER, x+cardW-34-86, y+170, 86, 38)
		btnGap := int32(10)
		btnW := (cardW - 68 - btnGap) / 2
		if lockedPathExists() {
			move(ID_OPEN, x+34, y+232, btnW, 42)
			move(ID_CREATE, x+34+btnW+btnGap, y+232, btnW, 42)
		} else {
			move(ID_CONFIRM, x+34, y+224, cardW-68, 38)
			move(ID_OPEN, x+34, y+286, btnW, 42)
			move(ID_CREATE, x+34+btnW+btnGap, y+286, btnW, 42)
		}
		pInvalidateRect.Call(hwnd, 0, 0)
		return
	}

	top := int32(84)
	bottom := int32(70)
	leftW := cw * 31 / 100
	if leftW < 300 {
		leftW = 300
	}
	if leftW > 395 {
		leftW = 395
	}
	contentGap := int32(14)
	rightX := margin + leftW + contentGap
	rightW := cw - rightX - margin

	move(ID_LIST_HEADING, margin+18, top+16, leftW-36, 24)
	move(ID_SEARCH, margin+16, top+50, leftW-110, 36)
	move(ID_TRASHMODE, margin+leftW-88, top+50, 72, 36)
	move(ID_LIST, margin+16, top+98, leftW-32, ch-top-bottom-158)
	move(ID_NEW, margin+16, ch-bottom-50, leftW-32, 36)

	move(ID_DETAIL_HEADING, rightX+20, top+16, rightW-40, 24)
	labelW := int32(126)
	fieldX := rightX + 20 + labelW
	fieldW := rightW - 40 - labelW
	y := top + 52
	line := int32(40)
	move(ID_LABEL_TITLE, rightX+20, y+9, labelW-8, 20)
	move(ID_TITLE, fieldX, y, fieldW, 34)
	y += line
	move(ID_LABEL_USER, rightX+20, y+9, labelW-8, 20)
	move(ID_USERNAME, fieldX, y, fieldW, 34)
	y += line
	move(ID_LABEL_EMAIL, rightX+20, y+9, labelW-8, 20)
	move(ID_EMAIL, fieldX, y, fieldW, 34)
	y += line
	move(ID_LABEL_PASS, rightX+20, y+9, labelW-8, 20)
	move(ID_PASSWORD, fieldX, y, fieldW-90, 34)
	move(ID_SHOWPASS, fieldX+fieldW-82, y, 82, 34)
	y += 35
	move(ID_SECURITY+1001, fieldX, y, fieldW, 18)
	y += 23
	move(ID_LABEL_URL, rightX+20, y+9, labelW-8, 20)
	move(ID_URL, fieldX, y, fieldW-90, 34)
	move(ID_OPENURL, fieldX+fieldW-82, y, 82, 34)
	y += line
	move(ID_LABEL_TOTP, rightX+20, y+9, labelW-8, 20)
	move(ID_TOTP, fieldX, y, fieldW, 34)
	y += 35
	move(ID_SECURITY+1002, fieldX, y, fieldW, 18)
	y += 23
	move(ID_LABEL_TAGS, rightX+20, y+9, labelW-8, 20)
	move(ID_TAGS, fieldX, y, fieldW-114, 34)
	move(ID_FAVORITE, fieldX+fieldW-106, y, 106, 34)
	y += line
	move(ID_LABEL_NOTES, rightX+20, y+9, labelW-8, 20)
	notesH := ch - bottom - y - 98
	if notesH < 72 {
		notesH = 72
	}
	move(ID_NOTES, fieldX, y, fieldW, notesH)
	y += notesH + 10
	btnGap := int32(7)
	btnW := (fieldW - 3*btnGap) / 4
	move(ID_SAVEENTRY, fieldX, y, btnW, 34)
	move(ID_DELETE, fieldX+btnW+btnGap, y, btnW, 34)
	move(ID_RESTORE, fieldX+2*(btnW+btnGap), y, btnW, 34)
	move(ID_GENERATE, fieldX+3*(btnW+btnGap), y, btnW, 34)
	y += 41
	move(ID_COPYUSER, fieldX, y, btnW, 34)
	move(ID_COPYEMAIL, fieldX+btnW+btnGap, y, btnW, 34)
	move(ID_COPYPASS, fieldX+2*(btnW+btnGap), y, btnW, 34)
	move(ID_COPYTOTP, fieldX+3*(btnW+btnGap), y, btnW, 34)

	move(ID_SECURITY, margin, ch-52, 112, 34)
	move(ID_SAVEVAULT, margin+120, ch-52, 118, 34)
	move(ID_BACKUP, margin+246, ch-52, 106, 34)
	move(ID_EXPORT, margin+360, ch-52, 126, 34)
	move(ID_LOCK, margin+494, ch-52, 102, 34)
	pInvalidateRect.Call(hwnd, 0, 0)
}

func setUnlockedUI(on bool) {
	loginIDs := []int{ID_LOCK_HEADING, ID_LOCK_SUBTITLE, ID_PATH, ID_BROWSE, ID_MASTER, ID_CONFIRM, ID_SHOWMASTER, ID_OPEN, ID_CREATE}
	appIDs := []int{ID_LIST_HEADING, ID_DETAIL_HEADING, ID_SEARCH, ID_LIST, ID_TRASHMODE, ID_NEW, ID_LABEL_TITLE, ID_TITLE, ID_LABEL_USER, ID_USERNAME, ID_LABEL_EMAIL, ID_EMAIL, ID_COPYEMAIL, ID_LABEL_PASS, ID_PASSWORD, ID_SHOWPASS, ID_SECURITY + 1001, ID_LABEL_URL, ID_URL, ID_LABEL_TOTP, ID_TOTP, ID_SECURITY + 1002, ID_LABEL_TAGS, ID_TAGS, ID_FAVORITE, ID_LABEL_NOTES, ID_NOTES, ID_SAVEENTRY, ID_DELETE, ID_RESTORE, ID_COPYUSER, ID_COPYPASS, ID_GENERATE, ID_COPYTOTP, ID_OPENURL, ID_SECURITY, ID_SAVEVAULT, ID_BACKUP, ID_EXPORT, ID_LOCK}
	for _, id := range loginIDs {
		visible(id, !on)
	}
	for _, id := range appIDs {
		visible(id, on)
	}
	if on {
		updateModeButtons()
	}
	layout()
}

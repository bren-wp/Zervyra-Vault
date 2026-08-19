//go:build windows

package main

const (
	ID_BRAND_TITLE    = 2050
	ID_BRAND_SUBTITLE = 2051
	ID_LIST_HEADING   = 2052
	ID_DETAIL_HEADING = 2053
	ID_LOCK_HEADING   = 2060
	ID_LOCK_SUBTITLE  = 2061
	ID_LABEL_TITLE    = 2010
	ID_LABEL_USER     = 2011
	ID_LABEL_PASS     = 2012
	ID_LABEL_URL      = 2013
	ID_LABEL_TOTP     = 2014
	ID_LABEL_TAGS     = 2015
	ID_LABEL_NOTES    = 2016
	ID_LABEL_EMAIL    = 2017
)

func createUI(parent uintptr) {
	hwnd = parent
	applyDarkWindow(parent)

	createControl(parent, "STATIC", "ZERVYRA", 0, 0, ID_BRAND_TITLE)
	createControl(parent, "STATIC", "Vault • privatno, lokalno i šifrirano • v"+version, 0, 0, ID_BRAND_SUBTITLE)
	if headingFont != 0 {
		pSendMessage.Call(controls[ID_BRAND_TITLE], WM_SETFONT, headingFont, 1)
	}
	if smallFont != 0 {
		pSendMessage.Call(controls[ID_BRAND_SUBTITLE], WM_SETFONT, smallFont, 1)
	}

	createControl(parent, "STATIC", "Otključaj svoj trezor", 0, 0, ID_LOCK_HEADING)
	createControl(parent, "STATIC", "Jedna master lozinka. Svi zapisi ostaju lokalno šifrirani.", 0, 0, ID_LOCK_SUBTITLE)
	if headingFont != 0 {
		pSendMessage.Call(controls[ID_LOCK_HEADING], WM_SETFONT, headingFont, 1)
	}
	if smallFont != 0 {
		pSendMessage.Call(controls[ID_LOCK_SUBTITLE], WM_SETFONT, smallFont, 1)
	}

	createControl(parent, "EDIT", defaultPath(), 0, WS_TABSTOP|ES_AUTOHSCROLL, ID_PATH)
	createControl(parent, "BUTTON", "Odaberi", 0, WS_TABSTOP, ID_BROWSE)
	createControl(parent, "EDIT", "", 0, WS_TABSTOP|ES_AUTOHSCROLL|ES_PASSWORD, ID_MASTER)
	createControl(parent, "EDIT", "", 0, WS_TABSTOP|ES_AUTOHSCROLL|ES_PASSWORD, ID_CONFIRM)
	createControl(parent, "BUTTON", "Prikaži", 0, WS_TABSTOP, ID_SHOWMASTER)
	createControl(parent, "BUTTON", "Otključaj", 0, WS_TABSTOP, ID_OPEN)
	createControl(parent, "BUTTON", "Novi trezor", 0, WS_TABSTOP, ID_CREATE)
	setCue(ID_PATH, "Lokacija .bvault datoteke")
	setCue(ID_MASTER, "Master lozinka")
	setCue(ID_CONFIRM, "Ponovi master lozinku samo kod kreiranja")

	createControl(parent, "STATIC", "Zapisi", 0, 0, ID_LIST_HEADING)
	createControl(parent, "EDIT", "", 0, WS_TABSTOP|ES_AUTOHSCROLL, ID_SEARCH)
	setCue(ID_SEARCH, "Pretraži naziv, e-mail, korisnika ili tag…")
	createControl(parent, "LISTBOX", "", 0, WS_TABSTOP|WS_VSCROLL|LBS_NOTIFY|LBS_NOINTEGRALHEIGHT, ID_LIST)
	createControl(parent, "BUTTON", "Koš", 0, WS_TABSTOP, ID_TRASHMODE)
	createControl(parent, "BUTTON", "+ Novi zapis", 0, WS_TABSTOP, ID_NEW)

	createControl(parent, "STATIC", "Detalji zapisa", 0, 0, ID_DETAIL_HEADING)
	createControl(parent, "STATIC", "Naziv", 0, 0, ID_LABEL_TITLE)
	createControl(parent, "EDIT", "", 0, WS_TABSTOP|ES_AUTOHSCROLL, ID_TITLE)
	createControl(parent, "STATIC", "Korisničko ime", 0, 0, ID_LABEL_USER)
	createControl(parent, "EDIT", "", 0, WS_TABSTOP|ES_AUTOHSCROLL, ID_USERNAME)
	createControl(parent, "STATIC", "E-mail", 0, 0, ID_LABEL_EMAIL)
	createControl(parent, "EDIT", "", 0, WS_TABSTOP|ES_AUTOHSCROLL, ID_EMAIL)
	createControl(parent, "BUTTON", "Kopiraj e-mail", 0, WS_TABSTOP, ID_COPYEMAIL)
	createControl(parent, "STATIC", "Lozinka", 0, 0, ID_LABEL_PASS)
	createControl(parent, "EDIT", "", 0, WS_TABSTOP|ES_AUTOHSCROLL|ES_PASSWORD, ID_PASSWORD)
	createControl(parent, "BUTTON", "Prikaži", 0, WS_TABSTOP, ID_SHOWPASS)
	createControl(parent, "STATIC", "Web adresa", 0, 0, ID_LABEL_URL)
	createControl(parent, "EDIT", "", 0, WS_TABSTOP|ES_AUTOHSCROLL, ID_URL)
	createControl(parent, "STATIC", "2FA / TOTP", 0, 0, ID_LABEL_TOTP)
	createControl(parent, "EDIT", "", 0, WS_TABSTOP|ES_AUTOHSCROLL, ID_TOTP)
	createControl(parent, "STATIC", "Tagovi", 0, 0, ID_LABEL_TAGS)
	createControl(parent, "EDIT", "", 0, WS_TABSTOP|ES_AUTOHSCROLL, ID_TAGS)
	createControl(parent, "STATIC", "Bilješke", 0, 0, ID_LABEL_NOTES)
	createControl(parent, "EDIT", "", 0, WS_TABSTOP|WS_VSCROLL|ES_MULTILINE|ES_AUTOVSCROLL|ES_WANTRETURN, ID_NOTES)
	createControl(parent, "BUTTON", "☆ Favorit", 0, WS_TABSTOP, ID_FAVORITE)

	createControl(parent, "BUTTON", "Spremi", 0, WS_TABSTOP, ID_SAVEENTRY)
	createControl(parent, "BUTTON", "Premjesti u koš", 0, WS_TABSTOP, ID_DELETE)
	createControl(parent, "BUTTON", "Vrati", 0, WS_TABSTOP, ID_RESTORE)
	createControl(parent, "BUTTON", "Kopiraj korisnika", 0, WS_TABSTOP, ID_COPYUSER)
	createControl(parent, "BUTTON", "Kopiraj lozinku", 0, WS_TABSTOP, ID_COPYPASS)
	createControl(parent, "BUTTON", "Generiraj", 0, WS_TABSTOP, ID_GENERATE)
	createControl(parent, "BUTTON", "Kopiraj 2FA", 0, WS_TABSTOP, ID_COPYTOTP)
	createControl(parent, "BUTTON", "Otvori web", 0, WS_TABSTOP, ID_OPENURL)

	createControl(parent, "BUTTON", "Sigurnost", 0, WS_TABSTOP, ID_SECURITY)
	createControl(parent, "BUTTON", "Spremi sada", 0, WS_TABSTOP, ID_SAVEVAULT)
	createControl(parent, "BUTTON", "Oporavak", 0, WS_TABSTOP, ID_BACKUP)
	createControl(parent, "BUTTON", "Backup kopija", 0, WS_TABSTOP, ID_EXPORT)
	createControl(parent, "BUTTON", "Zaključaj", 0, WS_TABSTOP, ID_LOCK)
	createControl(parent, "STATIC", "Nema otvorenog trezora", 0, 0, ID_SECURITY+1000)
	createControl(parent, "STATIC", "Snaga lozinke: —", 0, 0, ID_SECURITY+1001)
	createControl(parent, "STATIC", "2FA: —", 0, 0, ID_SECURITY+1002)

	pSendMessage.Call(controls[ID_PATH], EM_SETLIMITTEXT, 32767, 0)
	pSendMessage.Call(controls[ID_MASTER], EM_SETLIMITTEXT, 1024, 0)
	pSendMessage.Call(controls[ID_CONFIRM], EM_SETLIMITTEXT, 1024, 0)
	pSendMessage.Call(controls[ID_TITLE], EM_SETLIMITTEXT, 4096, 0)
	pSendMessage.Call(controls[ID_USERNAME], EM_SETLIMITTEXT, 16384, 0)
	pSendMessage.Call(controls[ID_EMAIL], EM_SETLIMITTEXT, 16384, 0)
	pSendMessage.Call(controls[ID_PASSWORD], EM_SETLIMITTEXT, 65536, 0)
	pSendMessage.Call(controls[ID_URL], EM_SETLIMITTEXT, 32768, 0)
	pSendMessage.Call(controls[ID_TOTP], EM_SETLIMITTEXT, 8192, 0)
	pSendMessage.Call(controls[ID_TAGS], EM_SETLIMITTEXT, 32768, 0)
	pSendMessage.Call(controls[ID_NOTES], EM_SETLIMITTEXT, 4<<20, 0)
	setCue(ID_TITLE, "npr. Google, poslovni portal…")
	setCue(ID_USERNAME, "korisničko ime")
	setCue(ID_EMAIL, "ime@domena.hr")
	setCue(ID_PASSWORD, "lozinka")
	setCue(ID_URL, "https://…")
	setCue(ID_TOTP, "TOTP secret ili otpauth:// URI")
	setCue(ID_TAGS, "posao, privatno, financije…")

	setPasswordMask(ID_MASTER, false)
	setPasswordMask(ID_CONFIRM, false)
	setPasswordMask(ID_PASSWORD, false)
	setUnlockedUI(false)
	updateLockedScreen("")
	refreshList()
}

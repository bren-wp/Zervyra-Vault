//go:build windows

package main

import (
	"fmt"
	"sort"
	"strings"
	"unsafe"
	"zervyra-vault-native/internal/core"
)

func updateModeButtons() {
	unlocked := master != ""
	activeSelection := selectedID != "" && !showTrash
	trashSelection := selectedID != "" && showTrash
	editMode := unlocked && !showTrash && (activeSelection || editingNew)
	for _, id := range []int{ID_TITLE, ID_USERNAME, ID_EMAIL, ID_PASSWORD, ID_SHOWPASS, ID_URL, ID_TOTP, ID_TAGS, ID_NOTES, ID_FAVORITE} {
		enabled(id, editMode)
	}
	enabled(ID_SAVEENTRY, editMode)
	enabled(ID_DELETE, unlocked && (activeSelection || trashSelection))
	canUndo := false
	if activeSelection {
		if e := selectedEntry(); e != nil {
			canUndo = len(e.Revisions) > 0
		}
	}
	enabled(ID_RESTORE, unlocked && (trashSelection || canUndo))
	enabled(ID_COPYUSER, unlocked && selectedID != "")
	enabled(ID_COPYEMAIL, unlocked && selectedID != "")
	enabled(ID_COPYPASS, unlocked && selectedID != "")
	enabled(ID_COPYTOTP, unlocked && selectedID != "")
	enabled(ID_OPENURL, unlocked && selectedID != "")
	enabled(ID_NEW, unlocked && !showTrash)
	enabled(ID_GENERATE, editMode)
	enabled(ID_BACKUP, unlocked)
	if showTrash {
		setText(controls[ID_TRASHMODE], "Zapisi")
		setText(controls[ID_DELETE], "Trajno obriši")
		setText(controls[ID_RESTORE], "Vrati")
	} else {
		setText(controls[ID_TRASHMODE], fmt.Sprintf("Koš (%d)", len(vault.Trash)))
		setText(controls[ID_DELETE], "Premjesti u koš")
		setText(controls[ID_RESTORE], "Vrati izmjenu")
	}
}

func selectedEntry() *core.Entry {
	if selectedID == "" {
		return nil
	}
	list := &vault.Entries
	if showTrash {
		list = &vault.Trash
	}
	for i := range *list {
		if (*list)[i].ID == selectedID {
			return &(*list)[i]
		}
	}
	return nil
}

func clearEditor() {
	suppressChange = true
	defer func() { suppressChange = false }()
	for _, id := range []int{ID_TITLE, ID_USERNAME, ID_EMAIL, ID_PASSWORD, ID_URL, ID_TOTP, ID_TAGS, ID_NOTES} {
		setText(controls[id], "")
	}
	setChecked(ID_FAVORITE, false)
	setText(controls[ID_SECURITY+1001], "Snaga lozinke: —")
	setText(controls[ID_SECURITY+1002], "TOTP: —")
	dirty = false
	revisionCaptured = false
}

func fillEditor(e *core.Entry) {
	suppressChange = true
	defer func() { suppressChange = false }()
	if e == nil {
		clearEditor()
		return
	}
	setText(controls[ID_TITLE], e.Title)
	setText(controls[ID_USERNAME], e.Username)
	setText(controls[ID_EMAIL], e.Email)
	setText(controls[ID_PASSWORD], e.Password)
	setText(controls[ID_URL], e.URL)
	setText(controls[ID_TOTP], e.TOTP)
	setText(controls[ID_TAGS], e.Tags)
	setText(controls[ID_NOTES], e.Notes)
	setChecked(ID_FAVORITE, e.Favorite)
	updatePasswordScore()
	updateTOTPLabel()
	dirty = false
	revisionCaptured = false
}

func refreshList() {
	oldID := selectedID
	pSendMessage.Call(controls[ID_LIST], LB_RESETCONTENT, 0, 0)
	visibleIndices = visibleIndices[:0]
	query := strings.ToLower(strings.TrimSpace(text(controls[ID_SEARCH])))
	entries := vault.Entries
	if showTrash {
		entries = vault.Trash
	}
	for i, e := range entries {
		hay := strings.ToLower(strings.Join([]string{e.Title, e.Username, e.Email, e.URL, e.Tags, e.Notes}, "\n"))
		if query == "" || strings.Contains(hay, query) {
			visibleIndices = append(visibleIndices, i)
		}
	}
	sort.SliceStable(visibleIndices, func(i, j int) bool {
		a, b := entries[visibleIndices[i]], entries[visibleIndices[j]]
		if a.Favorite != b.Favorite {
			return a.Favorite
		}
		return strings.ToLower(a.Title) < strings.ToLower(b.Title)
	})
	selectedRow := -1
	for row, idx := range visibleIndices {
		e := entries[idx]
		label := e.Title
		if strings.TrimSpace(label) == "" {
			label = "(Bez naziva)"
		}
		if e.Favorite {
			label = "★ " + label
		}
		identity := e.Username
		if identity == "" {
			identity = e.Email
		}
		if identity != "" {
			label += "   —   " + identity
		}
		pSendMessage.Call(controls[ID_LIST], LB_ADDSTRING, 0, uintptr(unsafe.Pointer(w(label))))
		if e.ID == oldID {
			selectedRow = row
		}
	}
	if selectedRow >= 0 {
		pSendMessage.Call(controls[ID_LIST], LB_SETCURSEL, uintptr(selectedRow), 0)
	} else if oldID != "" {
		selectedID = ""
		if !editingNew {
			clearEditor()
		}
	}
	updateModeButtons()
}

func selectFromList() {
	if dirty && !showTrash {
		if err := saveCurrentEntry(true); err != nil {
			fail(err.Error())
			refreshList()
			return
		}
	}
	r, _, _ := pSendMessage.Call(controls[ID_LIST], LB_GETCURSEL, 0, 0)
	row := int(int32(r))
	if row < 0 || row >= len(visibleIndices) {
		selectedID = ""
		clearEditor()
		updateModeButtons()
		return
	}
	entries := vault.Entries
	if showTrash {
		entries = vault.Trash
	}
	idx := visibleIndices[row]
	if idx < 0 || idx >= len(entries) {
		return
	}
	selectedID = entries[idx].ID
	editingNew = false
	fillEditor(selectedEntry())
	updateModeButtons()
}

func editorHasContent() bool {
	for _, id := range []int{ID_TITLE, ID_USERNAME, ID_EMAIL, ID_PASSWORD, ID_URL, ID_TOTP, ID_TAGS, ID_NOTES} {
		if strings.TrimSpace(text(controls[id])) != "" {
			return true
		}
	}
	return checked(ID_FAVORITE)
}

func saveCurrentEntry(persist bool) error {
	return saveCurrentEntryMode(persist, false)
}

func autosaveCurrentEntry() error {
	return saveCurrentEntryMode(true, true)
}

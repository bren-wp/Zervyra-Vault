//go:build windows

package main

import (
	"strings"
	"time"
	"zervyra-vault-native/internal/core"
)

func saveCurrentEntryMode(persist, autosave bool) error {
	if master == "" || showTrash {
		return nil
	}
	title := strings.TrimSpace(text(controls[ID_TITLE]))
	if title == "" {
		if autosave && editorHasContent() {
			title = "Bez naziva"
			suppressChange = true
			setText(controls[ID_TITLE], title)
			suppressChange = false
		} else {
			return errorsText("Naziv zapisa je obavezan.")
		}
	}
	var e *core.Entry
	wasNew := editingNew || selectedID == ""
	if wasNew {
		ne := core.NewEntry()
		vault.Entries = append(vault.Entries, ne)
		e = &vault.Entries[len(vault.Entries)-1]
		selectedID = e.ID
		editingNew = false
		revisionCaptured = true
	} else {
		e = selectedEntry()
		if e == nil {
			return errorsText("Odabrani zapis više ne postoji.")
		}
	}
	var before core.EntryRevision
	trackRevision := !wasNew && !revisionCaptured
	if trackRevision {
		before = core.SnapshotEntry(*e)
	}

	e.Title = title
	e.Username = text(controls[ID_USERNAME])
	e.Email = strings.TrimSpace(text(controls[ID_EMAIL]))
	newPassword := text(controls[ID_PASSWORD])
	core.RecordPasswordChange(e, newPassword)
	e.URL = strings.TrimSpace(text(controls[ID_URL]))
	e.TOTP = strings.TrimSpace(text(controls[ID_TOTP]))
	e.Tags = strings.TrimSpace(text(controls[ID_TAGS]))
	e.Notes = text(controls[ID_NOTES])
	e.Favorite = checked(ID_FAVORITE)
	if trackRevision && !core.RevisionMatchesEntry(before, *e) {
		core.AppendRevision(e, before)
		revisionCaptured = true
	}
	e.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)

	if persist {
		var err error
		if autosave {
			err = core.SaveAutosave(currentPath, master, vault)
		} else {
			err = core.Save(currentPath, master, vault)
		}
		if err != nil {
			dirty = true
			autosaveError = err.Error()
			return err
		}
		dirty = false
		// A successful persistence point becomes the baseline for the next edit,
		// so each later autosave can preserve another full-record revision.
		revisionCaptured = false
		lastSavedAt = time.Now()
		autosaveError = ""
		if autosave {
			setStatus("Automatski spremljeno • " + lastSavedAt.Format("15:04:05"))
		} else {
			setStatus("Spremljeno • " + lastSavedAt.Format("15:04:05"))
		}
	} else {
		dirty = true
	}
	refreshList()
	return nil
}

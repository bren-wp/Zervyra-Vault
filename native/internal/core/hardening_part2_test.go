package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmailRoundTripAndSearchSafeSchema(t *testing.T) {
	p := filepath.Join(t.TempDir(), "vault.bvault")
	pw := "Very-Strong-Master-Password!"
	v := NewVault()
	v.Entries = []Entry{{ID: RandomID(), Title: "Mail", Username: "user1", Email: "user@example.com", Password: "Strong-Password-123!"}}
	if err := Save(p, pw, v); err != nil {
		t.Fatal(err)
	}
	got, err := Load(p, pw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Entries) != 1 || got.Entries[0].Email != "user@example.com" || got.Entries[0].Username != "user1" {
		t.Fatalf("identity fields lost: %#v", got.Entries)
	}
}

func TestEntryRevisionRestoresIdentityAndSecretFields(t *testing.T) {
	e := NewEntry()
	e.Title = "Original"
	e.Username = "user-old"
	e.Email = "old@example.com"
	e.Password = "Old-Strong-Password-123!"
	e.URL = "https://old.example.com"
	e.Notes = "old note"
	e.TOTP = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	before := SnapshotEntry(e)
	e.Title = "Changed"
	e.Username = "user-new"
	e.Email = "new@example.com"
	e.Password = "New-Strong-Password-456!"
	e.URL = "https://new.example.com"
	e.Notes = "new note"
	e.TOTP = ""
	AppendRevision(&e, before)
	if len(e.Revisions) != 1 {
		t.Fatalf("expected one revision, got %d", len(e.Revisions))
	}
	if !RestoreLastRevision(&e) {
		t.Fatal("restore failed")
	}
	if e.Title != "Original" || e.Username != "user-old" || e.Email != "old@example.com" || e.Password != "Old-Strong-Password-123!" || e.URL != "https://old.example.com" || e.Notes != "old note" || e.TOTP == "" {
		t.Fatalf("full record was not restored: %#v", e)
	}
}

func TestVeryLongMasterPasswordRoundTrip(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "long-master.bvault")
	pw := strings.Repeat("A", 4096) + "!9"
	v := NewVault()
	v.Entries = []Entry{{ID: RandomID(), Title: "Long master", Password: "Strong-Password-123!"}}
	if err := Save(p, pw, v); err != nil {
		t.Fatal(err)
	}
	got, err := Load(p, pw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Entries) != 1 || got.Entries[0].Title != "Long master" {
		t.Fatal("long master roundtrip failed")
	}
}

func TestDeadPIDLockIsRecovered(t *testing.T) {
	p := filepath.Join(t.TempDir(), "vault.bvault")
	lock := p + ".lock"
	if err := os.WriteFile(lock, []byte("pid=2147483646\ncreated=2020-01-01T00:00:00Z\n"), 0600); err != nil {
		t.Fatal(err)
	}
	l, err := AcquireLock(p)
	if err != nil {
		t.Fatal(err)
	}
	l.Release()
}

func TestExportVaultBackupUsesProvidedMemoryState(t *testing.T) {
	d := t.TempDir()
	dst := filepath.Join(d, "memory-backup.bvault")
	pw := "Very-Strong-Master-Password!"
	v := NewVault()
	v.Entries = []Entry{{ID: RandomID(), Title: "Newest in memory", Email: "latest@example.com", Password: "Strong-Password-123!"}}
	if err := ExportVaultBackup(dst, pw, v); err != nil {
		t.Fatal(err)
	}
	got, err := Load(dst, pw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Entries) != 1 || got.Entries[0].Email != "latest@example.com" {
		t.Fatalf("memory backup mismatch: %#v", got.Entries)
	}
}

func TestCreateNewNeverOverwritesExistingVault(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "existing.bvault")
	original := []byte("do-not-overwrite")
	if err := os.WriteFile(p, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := CreateNew(p, "Very-Strong-Master-Password!", NewVault()); err == nil {
		t.Fatal("CreateNew overwrote an existing destination")
	}
	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatal("existing file content changed")
	}
}

func TestPBKDF2LongKeyPrehashMatchesStandardVector(t *testing.T) {
	got := derive(strings.Repeat("A", 100), []byte("salt"), 2, 32)
	const want = "2b0cfa170ed5f039393188125a7240fe6dfc28038633a4c930b10ea61327ae14"
	if fmt.Sprintf("%x", got) != want {
		t.Fatalf("PBKDF2 compatibility mismatch: %x", got)
	}
}

func TestImmediateSnapshotsProtectRecentGenerations(t *testing.T) {
	p := filepath.Join(t.TempDir(), "vault.bvault")
	pw := "Very-Strong-Master-Password!"
	v := NewVault()
	v.Entries = []Entry{{ID: RandomID(), Title: "generation-1", Password: "Strong-Password-123!"}}
	if err := Save(p, pw, v); err != nil {
		t.Fatal(err)
	}
	v.Entries[0].Title = "generation-2"
	if err := SaveAutosave(p, pw, v); err != nil {
		t.Fatal(err)
	}
	v.Entries[0].Title = "generation-3"
	if err := SaveAutosave(p, pw, v); err != nil {
		t.Fatal(err)
	}
	prev1, err := Load(p+".prev1", pw)
	if err != nil {
		t.Fatal("prev1 missing or invalid:", err)
	}
	prev2, err := Load(p+".prev2", pw)
	if err != nil {
		t.Fatal("prev2 missing or invalid:", err)
	}
	if prev1.Entries[0].Title != "generation-2" || prev2.Entries[0].Title != "generation-1" {
		t.Fatalf("unexpected immediate generations: prev1=%q prev2=%q", prev1.Entries[0].Title, prev2.Entries[0].Title)
	}
}

func TestLoadBestFallsBackToImmediateSnapshot(t *testing.T) {
	p := filepath.Join(t.TempDir(), "vault.bvault")
	pw := "Very-Strong-Master-Password!"
	v := NewVault()
	v.Entries = []Entry{{ID: RandomID(), Title: "safe-old", Password: "Strong-Password-123!"}}
	if err := Save(p, pw, v); err != nil {
		t.Fatal(err)
	}
	v.Entries[0].Title = "new-main"
	if err := SaveAutosave(p, pw, v); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("corrupt-main"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p+".recovery", []byte("corrupt-recovery"), 0600); err != nil {
		t.Fatal(err)
	}
	result, err := LoadBest(p, pw)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Vault.Entries) != 1 || result.Vault.Entries[0].Title != "safe-old" {
		t.Fatalf("did not recover immediate previous generation: %#v", result)
	}
	if !strings.HasSuffix(result.Source, ".prev1") {
		t.Fatalf("expected prev1 recovery, got %s", result.Source)
	}
}

func TestSecureCharRejectsEmptyAlphabet(t *testing.T) {
	if _, err := secureChar(""); err == nil {
		t.Fatal("empty generator alphabet accepted")
	}
}

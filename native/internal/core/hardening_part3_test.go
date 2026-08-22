package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVaultLockReleaseDoesNotDeleteReplacementLock(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault.bvault")

	lock, err := AcquireLock(vaultPath)
	if err != nil {
		t.Fatal(err)
	}
	lockPath := vaultPath + ".lock"

	replacement := "pid=999999\ntoken=replacement-owner\ncreated=2099-01-01T00:00:00Z\n"
	if err := os.WriteFile(lockPath, []byte(replacement), 0600); err != nil {
		t.Fatal(err)
	}
	lock.Release()

	raw, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("replacement lock was deleted by previous owner: %v", err)
	}
	if string(raw) != replacement {
		t.Fatalf("replacement lock changed unexpectedly: %q", string(raw))
	}
}

func TestAcquireLockWritesOwnershipToken(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault.bvault")

	lock, err := AcquireLock(vaultPath)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Release()

	raw, err := os.ReadFile(vaultPath + ".lock")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "token=") || readLockToken(vaultPath+".lock") == "" {
		t.Fatalf("lock ownership token missing: %q", string(raw))
	}
}

func TestCreateNewPreservesVerifiedMainIfRecoveryCreationFails(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault.bvault")
	password := "correct horse battery staple"
	v := NewVault()
	v.Entries = append(v.Entries, Entry{ID: RandomID(), Title: "Example", Username: "user", Password: "secret"})

	// Force recovery creation to fail after the main file has already been written
	// and verified by occupying the recovery path with a directory.
	if err := os.Mkdir(vaultPath+".recovery", 0700); err != nil {
		t.Fatal(err)
	}

	err := CreateNew(vaultPath, password, v)
	if err == nil {
		t.Fatal("expected recovery creation failure")
	}
	if !strings.Contains(err.Error(), "main vault was preserved") {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, loadErr := Load(vaultPath, password)
	if loadErr != nil {
		t.Fatalf("verified main vault was not preserved: %v", loadErr)
	}
	if len(loaded.Entries) != 1 || loaded.Entries[0].Username != "user" {
		t.Fatalf("preserved main vault has wrong data: %+v", loaded)
	}
}

func TestEncryptedEnvelopeLimitExceedsPlaintextLimit(t *testing.T) {
	if MaxVaultFileSize <= MaxVaultPlaintextSize {
		t.Fatalf("on-disk envelope limit %d must exceed plaintext limit %d", MaxVaultFileSize, MaxVaultPlaintextSize)
	}
	// Base64 expands by 4/3 plus a small JSON envelope. 96 MiB leaves ample room
	// for the current 64 MiB plaintext policy without relaxing plaintext limits.
	minimum := MaxVaultPlaintextSize + MaxVaultPlaintextSize/3
	if MaxVaultFileSize <= minimum {
		t.Fatalf("on-disk envelope limit %d does not account for Base64 expansion", MaxVaultFileSize)
	}
}

package core

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMalformedNonceRejectedWithoutPanic(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "bad.bvault")
	env := envelope{
		Format: Format, Iterations: KDFIterations,
		Salt:    base64.StdEncoding.EncodeToString(make([]byte, 16)),
		Nonce:   base64.StdEncoding.EncodeToString([]byte{1, 2, 3}),
		Payload: base64.StdEncoding.EncodeToString(make([]byte, 32)),
	}
	raw, _ := json.Marshal(env)
	if err := os.WriteFile(p, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p, "Very-Strong-Master-Password!"); err == nil {
		t.Fatal("malformed nonce accepted")
	}
}

func TestVaultLockPreventsSecondWriter(t *testing.T) {
	p := filepath.Join(t.TempDir(), "nested", "vault.bvault")
	l1, err := AcquireLock(p)
	if err != nil {
		t.Fatal(err)
	}
	defer l1.Release()
	if _, err := AcquireLock(p); err == nil {
		t.Fatal("second lock unexpectedly succeeded")
	}
	l1.Release()
	l2, err := AcquireLock(p)
	if err != nil {
		t.Fatal(err)
	}
	l2.Release()
}

func TestBackupRotation(t *testing.T) {
	p := filepath.Join(t.TempDir(), "vault.bvault")
	pw := "Very-Strong-Master-Password!"
	v := NewVault()
	v.Entries = append(v.Entries, Entry{ID: RandomID(), Title: "one", Password: "A-Strong-Password-123!"})
	if err := Save(p, pw, v); err != nil {
		t.Fatal(err)
	}
	v.Entries[0].Title = "two"
	if err := Save(p, pw, v); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p + ".bak1"); err != nil {
		t.Fatal("bak1 missing", err)
	}
	v.Entries[0].Title = "three"
	if err := Save(p, pw, v); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p + ".bak2"); err != nil {
		t.Fatal("bak2 missing", err)
	}
}

func TestGeneratorContainsAllClasses(t *testing.T) {
	p := Generate(32)
	if len(p) != 32 {
		t.Fatalf("length=%d", len(p))
	}
	checks := []string{"abcdefghijklmnopqrstuvwxyz", "ABCDEFGHIJKLMNOPQRSTUVWXYZ", "0123456789", "!@#$%^&*()-_=+[]{}"}
	for _, set := range checks {
		if !strings.ContainsAny(p, set) {
			t.Fatalf("missing class %q in %q", set, p)
		}
	}
}

func TestTOTPURI(t *testing.T) {
	uri := "otpauth://totp/Test?secret=GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ&digits=8&period=30&algorithm=SHA1"
	code, remain, err := TOTPDetails(uri, time.Unix(59, 0))
	if err != nil {
		t.Fatal(err)
	}
	if code != "94287082" {
		t.Fatalf("got %s", code)
	}
	if remain != 1 {
		t.Fatalf("remaining=%d", remain)
	}
}

func TestPasswordHistory(t *testing.T) {
	e := NewEntry()
	e.Password = "old-password-123!"
	RecordPasswordChange(&e, "new-password-456!")
	if e.Password != "new-password-456!" {
		t.Fatal("password not changed")
	}
	if len(e.PasswordHistory) != 1 || e.PasswordHistory[0].Password != "old-password-123!" {
		t.Fatal("history not recorded")
	}
}

func TestSafeURL(t *testing.T) {
	if !IsSafeHTTPURL("https://example.com/path") {
		t.Fatal("https rejected")
	}
	for _, s := range []string{"javascript:alert(1)", "file:///c:/x", "data:text/html,x", "https:///missing-host"} {
		if IsSafeHTTPURL(s) {
			t.Fatalf("unsafe URL accepted: %s", s)
		}
	}
}

func TestLoadBestPrefersNewestValidRecovery(t *testing.T) {
	p := filepath.Join(t.TempDir(), "vault.bvault")
	pw := "Very-Strong-Master-Password!"
	v := NewVault()
	v.Entries = []Entry{{ID: RandomID(), Title: "old", Password: "Old-Strong-Password-123!"}}
	if err := Save(p, pw, v); err != nil {
		t.Fatal(err)
	}
	oldRaw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Millisecond)
	v.Entries[0].Title = "newest-autosave"
	if err := SaveAutosave(p, pw, v); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, oldRaw, 0600); err != nil {
		t.Fatal(err)
	}
	result, err := LoadBest(p, pw)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Vault.Entries) == 0 || result.Vault.Entries[0].Title != "newest-autosave" {
		t.Fatalf("did not choose newest valid recovery: %#v", result)
	}
	if !result.Recovered {
		t.Fatal("expected recovered=true")
	}
}

func TestLoadBestFallsBackWhenMainCorrupt(t *testing.T) {
	p := filepath.Join(t.TempDir(), "vault.bvault")
	pw := "Very-Strong-Master-Password!"
	v := NewVault()
	v.Entries = []Entry{{ID: RandomID(), Title: "survives", Password: "Strong-Password-123!"}}
	if err := Save(p, pw, v); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("corrupt"), 0600); err != nil {
		t.Fatal(err)
	}
	result, err := LoadBest(p, pw)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Vault.Entries) != 1 || result.Vault.Entries[0].Title != "survives" {
		t.Fatalf("recovery data mismatch: %#v", result.Vault)
	}
	if !result.Recovered {
		t.Fatal("expected recovery path")
	}
}

func TestAutosaveCreatesCheckpointButDoesNotRotateOnEveryEdit(t *testing.T) {
	p := filepath.Join(t.TempDir(), "vault.bvault")
	pw := "Very-Strong-Master-Password!"
	v := NewVault()
	v.Entries = []Entry{{ID: RandomID(), Title: "base", Password: "Strong-Password-123!"}}
	if err := Save(p, pw, v); err != nil {
		t.Fatal(err)
	}
	v.Entries[0].Title = "autosave-1"
	if err := SaveAutosave(p, pw, v); err != nil {
		t.Fatal(err)
	}
	st1, err := os.Stat(p + ".bak1")
	if err != nil {
		t.Fatal("autosave safety checkpoint missing", err)
	}
	time.Sleep(2 * time.Millisecond)
	v.Entries[0].Title = "autosave-2"
	if err := SaveAutosave(p, pw, v); err != nil {
		t.Fatal(err)
	}
	st2, err := os.Stat(p + ".bak1")
	if err != nil {
		t.Fatal(err)
	}
	if !st1.ModTime().Equal(st2.ModTime()) {
		t.Fatal("backup chain rotated again before checkpoint interval")
	}
}

func TestCloneVaultIsDeepCopy(t *testing.T) {
	v := NewVault()
	v.Entries = []Entry{{ID: "1", Title: "original", PasswordHistory: []PasswordHistoryItem{{Password: "old"}}}}
	c := CloneVault(v)
	c.Entries[0].Title = "changed"
	c.Entries[0].PasswordHistory[0].Password = "changed-old"
	if v.Entries[0].Title != "original" || v.Entries[0].PasswordHistory[0].Password != "old" {
		t.Fatal("clone mutated original vault")
	}
}

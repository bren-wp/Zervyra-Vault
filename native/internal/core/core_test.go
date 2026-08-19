package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRoundTrip(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "v.bvault")
	v := NewVault()
	v.Entries = append(v.Entries, Entry{ID: RandomID(), Title: "Test", Password: "Secret!123456789"})
	if err := Save(p, "Very-Strong-Master-Password!", v); err != nil {
		t.Fatal(err)
	}
	got, err := Load(p, "Very-Strong-Master-Password!")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Entries) != 1 || got.Entries[0].Title != "Test" {
		t.Fatal("bad roundtrip")
	}
	_, err = Load(p, "wrong-password-xxxx")
	if err == nil {
		t.Fatal("wrong password accepted")
	}
	_ = os.Remove(p)
}
func TestGenerate(t *testing.T) {
	if len(Generate(24)) != 24 {
		t.Fatal()
	}
}
func TestTOTPVector(t *testing.T) {
	code, err := TOTP("GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ", time.Unix(59, 0))
	if err != nil {
		t.Fatal(err)
	}
	if code != "287082" {
		t.Fatalf("got %s", code)
	}
}
func TestTamperRejected(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "v.bvault")
	v := NewVault()
	if err := Save(p, "Very-Strong-Master-Password!", v); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] ^= 1
			break
		}
	}
	_ = os.WriteFile(p, b, 0600)
	if _, err = Load(p, "Very-Strong-Master-Password!"); err == nil {
		t.Fatal("tamper accepted")
	}
}

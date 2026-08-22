package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"strings"
)

// PBKDF2-HMAC-SHA256. Passwords longer than SHA-256's HMAC block size are
// pre-hashed once. This is mathematically equivalent to HMAC's key handling and
// avoids re-hashing a very long master password on every PBKDF2 iteration.
func derive(password string, salt []byte, iter, keyLen int) []byte {
	rawKey := []byte(password)
	defer zeroBytes(rawKey)
	keyMaterial := rawKey
	var prehash [sha256.Size]byte
	if len(rawKey) > sha256.BlockSize {
		prehash = sha256.Sum256(rawKey)
		keyMaterial = prehash[:]
		defer zeroBytes(prehash[:])
	}

	hLen := sha256.Size
	blocks := (keyLen + hLen - 1) / hLen
	out := make([]byte, 0, blocks*hLen)
	for i := 1; i <= blocks; i++ {
		mac := hmac.New(sha256.New, keyMaterial)
		mac.Write(salt)
		var ib [4]byte
		binary.BigEndian.PutUint32(ib[:], uint32(i))
		mac.Write(ib[:])
		u := mac.Sum(nil)
		t := append([]byte(nil), u...)
		for j := 1; j < iter; j++ {
			mac = hmac.New(sha256.New, keyMaterial)
			mac.Write(u)
			u = mac.Sum(nil)
			for k := range t {
				t[k] ^= u[k]
			}
		}
		out = append(out, t...)
	}
	return out[:keyLen]
}

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func normalizeVault(v *Vault) error {
	if len(v.Entries) > MaxEntries || len(v.Trash) > MaxEntries {
		return errors.New("vault contains too many entries")
	}
	if strings.TrimSpace(v.Name) == "" {
		v.Name = "Zervyra Vault"
	}
	if len(v.Name) > 256 {
		return errors.New("vault name is too long")
	}
	seen := make(map[string]bool, len(v.Entries)+len(v.Trash))
	normalize := func(e *Entry) error {
		if len(e.Title) > 4096 || len(e.Username) > 16384 || len(e.Email) > 16384 || len(e.Password) > 65536 || len(e.URL) > 32768 || len(e.Notes) > 4<<20 || len(e.TOTP) > 8192 || len(e.Tags) > 32768 {
			return errors.New("entry contains an oversized field")
		}
		if e.ID == "" || seen[e.ID] {
			e.ID = RandomID()
		}
		seen[e.ID] = true
		if e.CreatedAt == "" {
			e.CreatedAt = nowRFC3339()
		}
		if e.UpdatedAt == "" {
			e.UpdatedAt = e.CreatedAt
		}
		if len(e.PasswordHistory) > 50 {
			e.PasswordHistory = append([]PasswordHistoryItem(nil), e.PasswordHistory[len(e.PasswordHistory)-50:]...)
		}
		if len(e.Revisions) > MaxEntryRevisions {
			e.Revisions = append([]EntryRevision(nil), e.Revisions[len(e.Revisions)-MaxEntryRevisions:]...)
		}
		for _, r := range e.Revisions {
			if len(r.Title) > 4096 || len(r.Username) > 16384 || len(r.Email) > 16384 || len(r.Password) > 65536 || len(r.URL) > 32768 || len(r.Notes) > 4<<20 || len(r.TOTP) > 8192 || len(r.Tags) > 32768 {
				return errors.New("entry revision contains an oversized field")
			}
		}
		return nil
	}
	for i := range v.Entries {
		if err := normalize(&v.Entries[i]); err != nil {
			return err
		}
	}
	for i := range v.Trash {
		if err := normalize(&v.Trash[i]); err != nil {
			return err
		}
	}
	if v.Entries == nil {
		v.Entries = []Entry{}
	}
	if v.Trash == nil {
		v.Trash = []Entry{}
	}
	return nil
}

func encodeVault(password string, v Vault) ([]byte, error) {
	if password == "" {
		return nil, errors.New("master password is empty")
	}
	if err := normalizeVault(&v); err != nil {
		return nil, err
	}
	v.UpdatedAt = nowRFC3339()
	plain, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(plain) > MaxVaultPlaintextSize {
		zeroBytes(plain)
		return nil, errors.New("vault is too large")
	}
	defer zeroBytes(plain)

	salt := make([]byte, 16)
	nonce := make([]byte, 12)
	if _, err = io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	key := derive(password, salt, KDFIterations, 32)
	defer zeroBytes(key)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ct := gcm.Seal(nil, nonce, plain, []byte(Format))
	env := envelope{
		Format: Format, Iterations: KDFIterations,
		Salt:    base64.StdEncoding.EncodeToString(salt),
		Nonce:   base64.StdEncoding.EncodeToString(nonce),
		Payload: base64.StdEncoding.EncodeToString(ct),
	}
	raw, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return nil, err
	}
	if len(raw)+1 > MaxVaultFileSize {
		return nil, errors.New("encrypted vault envelope is too large")
	}
	return append(raw, '\n'), nil
}

func decodeVault(raw []byte, password string) (Vault, error) {
	var v Vault
	if len(raw) <= 0 || len(raw) > MaxVaultFileSize {
		return v, errors.New("vault file size is invalid")
	}
	var e envelope
	if err := json.Unmarshal(raw, &e); err != nil {
		return v, errors.New("vault envelope is invalid")
	}
	if e.Format != Format || e.Iterations < MinKDFIterations || e.Iterations > MaxKDFIterations {
		return v, errors.New("unsupported vault format or KDF parameters")
	}
	salt, err := base64.StdEncoding.DecodeString(e.Salt)
	if err != nil || len(salt) != 16 {
		return v, errors.New("vault salt is invalid")
	}
	nonce, err := base64.StdEncoding.DecodeString(e.Nonce)
	if err != nil || len(nonce) != 12 {
		return v, errors.New("vault nonce is invalid")
	}
	ct, err := base64.StdEncoding.DecodeString(e.Payload)
	if err != nil || len(ct) < 16 || len(ct) > MaxVaultPlaintextSize+16 {
		return v, errors.New("vault payload is invalid")
	}
	key := derive(password, salt, e.Iterations, 32)
	defer zeroBytes(key)
	block, err := aes.NewCipher(key)
	if err != nil {
		return v, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return v, err
	}
	plain, err := gcm.Open(nil, nonce, ct, []byte(Format))
	if err != nil {
		return v, errors.New("wrong master password or modified vault")
	}
	defer zeroBytes(plain)
	if len(plain) > MaxVaultPlaintextSize {
		return v, errors.New("decrypted vault is too large")
	}
	if err := json.Unmarshal(plain, &v); err != nil {
		return v, errors.New("decrypted vault data is invalid")
	}
	if err := normalizeVault(&v); err != nil {
		return v, err
	}
	return v, nil
}

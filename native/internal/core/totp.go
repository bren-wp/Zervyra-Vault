package core

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func TOTP(secret string, now time.Time) (string, error) {
	code, _, err := TOTPDetails(secret, now)
	return code, err
}

func TOTPDetails(input string, now time.Time) (string, int, error) {
	secret := strings.TrimSpace(input)
	digits := 6
	period := 30
	algorithm := "SHA1"
	if strings.HasPrefix(strings.ToLower(secret), "otpauth://") {
		u, err := url.Parse(secret)
		if err != nil || strings.ToLower(u.Scheme) != "otpauth" || strings.ToLower(u.Host) != "totp" {
			return "", 0, errors.New("invalid otpauth URI")
		}
		q := u.Query()
		secret = q.Get("secret")
		if d, err := strconv.Atoi(q.Get("digits")); err == nil && (d == 6 || d == 8) {
			digits = d
		}
		if p, err := strconv.Atoi(q.Get("period")); err == nil && p >= 15 && p <= 120 {
			period = p
		}
		if a := strings.ToUpper(q.Get("algorithm")); a == "SHA1" || a == "SHA256" || a == "SHA512" {
			algorithm = a
		}
	}
	clean := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(secret, " ", ""), "-", ""))
	if clean == "" {
		return "", 0, errors.New("TOTP secret is empty")
	}
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.TrimRight(clean, "="))
	if err != nil || len(key) < 10 {
		return "", 0, errors.New("invalid TOTP secret")
	}
	counter := uint64(now.Unix() / int64(period))
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], counter)
	var hf func() hash.Hash
	switch algorithm {
	case "SHA256":
		hf = sha256.New
	case "SHA512":
		hf = sha512.New
	default:
		hf = sha1.New
	}
	mac := hmac.New(hf, key)
	mac.Write(b[:])
	sum := mac.Sum(nil)
	o := sum[len(sum)-1] & 15
	if int(o)+3 >= len(sum) {
		return "", 0, errors.New("invalid TOTP digest")
	}
	code := (uint32(sum[o])&127)<<24 | (uint32(sum[o+1]) << 16) | (uint32(sum[o+2]) << 8) | uint32(sum[o+3])
	mod := uint32(1000000)
	if digits == 8 {
		mod = 100000000
	}
	remaining := period - int(now.Unix()%int64(period))
	return fmt.Sprintf("%0*d", digits, code%mod), remaining, nil
}

func IsSafeHTTPURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

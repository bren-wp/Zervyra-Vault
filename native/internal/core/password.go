package core

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
)

func Generate(n int) string {
	if n < 16 {
		n = 16
	}
	if n > 128 {
		n = 128
	}
	lower := "abcdefghijkmnopqrstuvwxyz"
	upper := "ABCDEFGHJKLMNPQRSTUVWXYZ"
	digits := "23456789"
	symbols := "!@#$%^&*()-_=+[]{}"
	all := lower + upper + digits + symbols
	result := make([]byte, n)
	required := []string{lower, upper, digits, symbols}
	for i := 0; i < len(required); i++ {
		c, err := secureChar(required[i])
		if err != nil {
			return ""
		}
		result[i] = c
	}
	for i := len(required); i < n; i++ {
		c, err := secureChar(all)
		if err != nil {
			return ""
		}
		result[i] = c
	}
	for i := n - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return ""
		}
		j := int(jBig.Int64())
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}

func secureChar(chars string) (byte, error) {
	if len(chars) == 0 {
		return 0, errors.New("empty character set")
	}
	x, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
	if err != nil {
		return 0, err
	}
	return chars[x.Int64()], nil
}

func Score(p string) int {
	if p == "" {
		return 0
	}
	score := 0
	l := len([]rune(p))
	switch {
	case l >= 20:
		score += 45
	case l >= 16:
		score += 38
	case l >= 12:
		score += 30
	case l >= 8:
		score += 15
	default:
		score += l
	}
	classes := 0
	for _, set := range []string{"abcdefghijklmnopqrstuvwxyz", "ABCDEFGHIJKLMNOPQRSTUVWXYZ", "0123456789", "!@#$%^&*()-_=+[]{}~`|;:'\",.<>/?\\"} {
		if strings.ContainsAny(p, set) {
			classes++
		}
	}
	score += classes * 12
	unique := make(map[rune]struct{})
	for _, r := range p {
		unique[r] = struct{}{}
	}
	if len(unique) >= 10 {
		score += 7
	}
	lower := strings.ToLower(p)
	for _, bad := range []string{"password", "qwerty", "123456", "admin", "letmein", "welcome"} {
		if strings.Contains(lower, bad) {
			score -= 25
			break
		}
	}
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}

func RecordPasswordChange(e *Entry, newPassword string) {
	if e == nil || newPassword == "" || newPassword == e.Password {
		return
	}
	if e.Password != "" {
		e.PasswordHistory = append(e.PasswordHistory, PasswordHistoryItem{Password: e.Password, ChangedAt: nowRFC3339()})
		if len(e.PasswordHistory) > 20 {
			e.PasswordHistory = append([]PasswordHistoryItem(nil), e.PasswordHistory[len(e.PasswordHistory)-20:]...)
		}
	}
	e.Password = newPassword
	e.UpdatedAt = nowRFC3339()
}

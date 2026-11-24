// Shared auth policy for PLANZI and POPCORN. The rules, the blocklist file
// (common_passwords.txt) and testdata/auth_conformance.json must stay identical in both repos;
// see LINKING_SPEC.md "Auth parity v2".
package authpolicy

import (
	"crypto/rand"
	_ "embed"
	"net/mail"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const (
	CodeInvalidEmail      = "INVALID_EMAIL"
	CodeInvalidUsername   = "INVALID_USERNAME"
	CodeWeakPassword      = "WEAK_PASSWORD"
	CodePasswordTooLong   = "PASSWORD_TOO_LONG"
	MinPasswordBytes      = 8
	MaxPasswordBytes      = 72 // bcrypt silently truncates beyond this
	MaxEmailLength        = 254
	BcryptCost            = 12
	minContainedNameBytes = 3
)

// PolicyError carries the wire code returned to clients (HTTP 400).
type PolicyError struct {
	Code    string
	Message string
}

func (e *PolicyError) Error() string { return e.Message }

var (
	ErrInvalidEmail    = &PolicyError{CodeInvalidEmail, "invalid email address"}
	ErrInvalidUsername = &PolicyError{CodeInvalidUsername, "username must be 3-30 characters of a-z, 0-9 or _"}
	ErrWeakPassword    = &PolicyError{CodeWeakPassword, "password is too weak"}
	ErrPasswordTooLong = &PolicyError{CodePasswordTooLong, "password must be at most 72 bytes"}
)

//go:embed common_passwords.txt
var blocklistRaw string

var (
	blocklistOnce sync.Once
	blocklist     map[string]struct{}
)

func isBlocked(lowerPassword string) bool {
	blocklistOnce.Do(func() {
		lines := strings.Split(blocklistRaw, "\n")
		blocklist = make(map[string]struct{}, len(lines))
		for _, l := range lines {
			if l = strings.TrimSpace(l); l != "" {
				blocklist[l] = struct{}{}
			}
		}
	})
	_, ok := blocklist[lowerPassword]
	return ok
}

// NormalizeEmail is the single spelling used to store and look up addresses.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ValidateEmail normalizes then validates: <= 254 bytes, parseable by net/mail as a bare address,
// exactly one @, non-empty local part, and a dotted domain with no empty labels.
func ValidateEmail(email string) (string, error) {
	n := NormalizeEmail(email)
	if n == "" || len(n) > MaxEmailLength || strings.Count(n, "@") != 1 {
		return "", ErrInvalidEmail
	}
	addr, err := mail.ParseAddress(n)
	if err != nil || addr.Address != n {
		return "", ErrInvalidEmail
	}
	at := strings.IndexByte(n, '@')
	local, domain := n[:at], n[at+1:]
	if local == "" || !strings.Contains(domain, ".") {
		return "", ErrInvalidEmail
	}
	for _, label := range strings.Split(domain, ".") {
		if label == "" {
			return "", ErrInvalidEmail
		}
	}
	return n, nil
}

var usernameRe = regexp.MustCompile(`^[a-z0-9_]{3,30}$`)

func NormalizeUsername(u string) string {
	return strings.ToLower(strings.TrimSpace(u))
}

// ValidateUsername normalizes then validates against ^[a-z0-9_]{3,30}$.
func ValidateUsername(u string) (string, error) {
	n := NormalizeUsername(u)
	if !usernameRe.MatchString(n) {
		return "", ErrInvalidUsername
	}
	return n, nil
}

// ValidatePassword enforces 8..72 bytes, no blocklisted password, no email local part or
// username inside it (names shorter than 3 bytes are ignored), and not one repeated character.
// There are deliberately no composition rules. email and username may be empty.
func ValidatePassword(password, email, username string) error {
	if len(password) > MaxPasswordBytes {
		return ErrPasswordTooLong
	}
	if len(password) < MinPasswordBytes {
		return ErrWeakPassword
	}
	lower := strings.ToLower(password)
	if isBlocked(lower) {
		return ErrWeakPassword
	}
	if r, _ := utf8.DecodeRuneInString(password); strings.Count(password, string(r)) == utf8.RuneCountInString(password) {
		return ErrWeakPassword
	}
	if n := NormalizeEmail(email); n != "" {
		local := n
		if at := strings.IndexByte(n, '@'); at >= 0 {
			local = n[:at]
		}
		if len(local) >= minContainedNameBytes && strings.Contains(lower, local) {
			return ErrWeakPassword
		}
	}
	if u := NormalizeUsername(username); len(u) >= minContainedNameBytes && strings.Contains(lower, u) {
		return ErrWeakPassword
	}
	return nil
}

// HashPassword hashes at BcryptCost.
func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	return string(h), err
}

// NeedsRehash reports whether a stored hash was made below BcryptCost.
func NeedsRehash(hash string) bool {
	cost, err := bcrypt.Cost([]byte(hash))
	return err == nil && cost < BcryptCost
}

var (
	dummyOnce sync.Once
	dummyHash []byte
)

// BurnPassword runs a bcrypt comparison against a dummy hash so that "unknown user" takes as
// long as "wrong password". Always returns false.
func BurnPassword(password string) bool {
	dummyOnce.Do(func() {
		seed := make([]byte, 16)
		_, _ = rand.Read(seed)
		dummyHash, _ = bcrypt.GenerateFromPassword(seed, BcryptCost)
	})
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
	return false
}

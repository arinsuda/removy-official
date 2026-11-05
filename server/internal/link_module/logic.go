package link_module

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

type Decision int

const (
	DecisionLoginLinked Decision = iota
	DecisionProvision
	DecisionAutoLink
	DecisionConfirm
)

// Decide picks the redeem action. localExists/localVerified describe the local
// user matched by email (case-insensitive); alreadyLinked is a remote_user_id hit.
func Decide(alreadyLinked, localExists, localVerified, remoteVerified bool) Decision {
	switch {
	case alreadyLinked:
		return DecisionLoginLinked
	case !localExists:
		return DecisionProvision
	case localVerified && remoteVerified:
		return DecisionAutoLink
	default:
		return DecisionConfirm
	}
}

// ConfirmMethods lists methods usable to confirm a pending link. Password needs
// a local password and a verified local email; the email code is always allowed.
func ConfirmMethods(localHasPassword, localVerified bool) []string {
	var m []string
	if localHasPassword && localVerified {
		m = append(m, MethodPasswd)
	}
	return append(m, MethodEmail)
}

func sanitizeLocalPart(email string) string {
	local := email
	if i := strings.Index(email, "@"); i >= 0 {
		local = email[:i]
	}
	var b strings.Builder
	for _, r := range strings.ToLower(local) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	s := b.String()
	if len(s) > 30 {
		s = s[:30]
	}
	if len(s) < 3 {
		s += "user"
	}
	return s
}

// GenerateUsername derives a unique username (<=50 chars) from the email local
// part, appending a numeric suffix on collision.
func GenerateUsername(email string, exists func(string) bool) string {
	base := sanitizeLocalPart(email)
	if !exists(base) {
		return base
	}
	for i := 0; i < 20; i++ {
		cand := fmt.Sprintf("%s%04d", base, randInt(10000))
		if !exists(cand) {
			return cand
		}
	}
	return fmt.Sprintf("%s%010d", base, randInt(10000000000))
}

func randInt(max int64) int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0
	}
	return n.Int64()
}

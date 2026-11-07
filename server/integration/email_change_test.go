//go:build integration

package integration

import (
	"errors"
	"testing"

	"github.com/arinsuda/movie-hub/internal/user_module"
)

func (e *env) verifiedAt(uid uint) bool {
	_, _, verified, _ := e.userRow(uid)
	return verified
}

// Happy path: two codes, one to each mailbox; the account stays verified throughout.
func TestEmailChangeNeedsBothCodesAndStaysVerified(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "old@x.test", true, "pw-secret-12")
	e.db.Exec(`INSERT INTO password_reset_tokens (user_id, hashed_token, expires_at, created_at) VALUES (?, 'h', now() + interval '1 hour', now())`, uid)
	e.db.Exec(`INSERT INTO link_handoff_tokens (token_hash, user_id, target_platform, expires_at) VALUES (repeat('b',64), ?, 'planzi', now() + interval '1 minute')`, uid)

	if err := e.users.RequestEmailChange(uid, uid, "New@X.test"); err != nil {
		t.Fatal(err)
	}
	oldCode := e.mail.last(t, "old@x.test", reCode)
	newCode := e.mail.last(t, "new@x.test", reCode)
	if oldCode == newCode {
		t.Fatal("both addresses got the same code")
	}
	if !e.verifiedAt(uid) {
		t.Fatal("request unverified the account")
	}

	if _, err := e.users.VerifyEmailChange(uid, uid, 0, oldCode, newCode); err != nil {
		t.Fatal(err)
	}
	email, pw, verified, _ := e.userRow(uid)
	if email != "new@x.test" || !verified || pw == nil {
		t.Fatalf("email=%s verified=%v password kept=%v", email, verified, pw != nil)
	}
	if e.count(`SELECT count(*) FROM password_reset_tokens WHERE user_id = ?`, uid) != 0 || e.count(`SELECT count(*) FROM link_handoff_tokens WHERE user_id = ?`, uid) != 0 {
		t.Fatal("tokens minted for the old address survived")
	}
	if err := e.login("new@x.test", "pw-secret-12"); err != nil {
		t.Fatalf("login after change: %v", err)
	}
}

// T2b: an attacker who controls only their own mailbox cannot move the account to the victim's address.
func TestT2b_EmailChangeToVictimWithoutNewAddressCodeRejected(t *testing.T) {
	e := newEnv(t)
	seedUser(t, e.db, "Victim@Gmail.com", true, "victim-pass12") // legacy mixed-case row
	uid := seedUser(t, e.db, "attacker@x.test", true, "attacker-pass1")

	// the address is taken, so the request itself is refused, case-insensitively.
	if err := e.users.RequestEmailChange(uid, uid, "VICTIM@gmail.com"); !errors.Is(err, user_module.ErrEmailAlreadyInUse) {
		t.Fatalf("taken address (different case): %v", err)
	}

	// Free address the attacker does not control the mailbox of: only the old-address code is readable.
	if err := e.users.RequestEmailChange(uid, uid, "someone-else@x.test"); err != nil {
		t.Fatal(err)
	}
	oldCode := e.mail.last(t, "attacker@x.test", reCode)
	for name, newCode := range map[string]string{"guess": "000000", "reuse old code": oldCode} {
		if _, err := e.users.VerifyEmailChange(uid, uid, 0, oldCode, newCode); !errors.Is(err, user_module.ErrOTPInvalid) {
			t.Fatalf("%s: want invalid otp, got %v", name, err)
		}
	}
	if email, _, verified, _ := e.userRow(uid); email != "attacker@x.test" || !verified {
		t.Fatalf("account changed or unverified: %s %v", email, verified)
	}
	// failed attempts are counted, and five of them kill the request
	var attempts int
	e.db.Raw(`SELECT attempt_count FROM email_change_requests WHERE user_id = ?`, uid).Scan(&attempts)
	if attempts != 2 {
		t.Fatalf("attempt_count = %d, want 2", attempts)
	}
}

func TestEmailChangeLegacySingleCodeRequestRejected(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "old@x.test", true, "pw-secret-12")
	// a request created by the previous version: one code, no new_otp_hash
	if err := e.users.RequestEmailChange(uid, uid, "new@x.test"); err != nil {
		t.Fatal(err)
	}
	oldCode := e.mail.last(t, "old@x.test", reCode)
	newCode := e.mail.last(t, "new@x.test", reCode)
	e.db.Exec(`UPDATE email_change_requests SET new_otp_hash = NULL WHERE user_id = ?`, uid)
	if _, err := e.users.VerifyEmailChange(uid, uid, 0, oldCode, newCode); !errors.Is(err, user_module.ErrOTPNotFound) {
		t.Fatalf("legacy request must be rejected, got %v", err)
	}
	if email, _, _, _ := e.userRow(uid); email != "old@x.test" {
		t.Fatal("email changed from a legacy request")
	}
}

// A request that passed RequireAuth before a claim cannot finish the change.
func TestEmailChangeStaleSessionRejected(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "old@x.test", true, "pw-secret-12")
	if err := e.users.RequestEmailChange(uid, uid, "new@x.test"); err != nil {
		t.Fatal(err)
	}
	oldCode := e.mail.last(t, "old@x.test", reCode)
	newCode := e.mail.last(t, "new@x.test", reCode)
	e.db.Exec(`UPDATE users SET token_version = token_version + 1 WHERE id = ?`, uid)
	if _, err := e.users.VerifyEmailChange(uid, uid, 0, oldCode, newCode); !errors.Is(err, user_module.ErrSessionStale) {
		t.Fatalf("got %v", err)
	}
}

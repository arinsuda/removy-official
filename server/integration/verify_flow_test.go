//go:build integration

package integration

import (
	"testing"

	"github.com/arinsuda/movie-hub/internal/auth_module"
)

// T4b: the right password keeps the account intact; a wrong or missing one wipes it.
func TestT4b_VerifyWithCorrectPasswordKeepsEverything(t *testing.T) {
	e := newEnv(t)
	uid := e.squat("my-own-password1")
	e.seedGoogleIdentity(uid, "my-google", "mine@gmail.com")
	e.db.Exec(`INSERT INTO account_links (local_user_id, remote_platform, remote_user_id, linked_via) VALUES (?, 'planzi', 'p-mine', 'auto')`, uid)

	res, err := e.auth.VerifyEmail(e.mail.last(t, victim, reVerifyToken), ptr("my-own-password1"), false)
	if err != nil {
		t.Fatal(err)
	}
	if res.PasswordResetRequired {
		t.Fatal("password_reset_required must be false when the password matched")
	}
	_, pw, verified, ver := e.userRow(uid)
	if !verified || pw == nil || ver != 0 {
		t.Fatalf("verified=%v password kept=%v version=%d", verified, pw != nil, ver)
	}
	if e.count(`SELECT count(*) FROM user_identities WHERE user_id = ?`, uid) != 1 || e.count(`SELECT count(*) FROM account_links WHERE local_user_id = ?`, uid) != 1 {
		t.Fatal("identity or link wiped although the same person verified")
	}
	if err := e.login(victim, "my-own-password1"); err != nil {
		t.Fatalf("login with the kept password: %v", err)
	}
}

// Only an explicit skip wipes: the person said they do not know the password.
func TestT4b_VerifySkipWipesAndAsksForReset(t *testing.T) {
	e := newEnv(t)
	uid := e.squat("attacker-pass1")
	e.seedGoogleIdentity(uid, "atk-google", "atk@gmail.com")
	e.db.Exec(`INSERT INTO account_links (local_user_id, remote_platform, remote_user_id, linked_via) VALUES (?, 'planzi', 'p-atk', 'auto')`, uid)

	res, err := e.auth.VerifyEmail(e.mail.last(t, victim, reVerifyToken), nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if !res.PasswordResetRequired {
		t.Fatal("password_reset_required must be true")
	}
	_, stored, verified, ver := e.userRow(uid)
	if !verified || stored != nil || ver != 1 {
		t.Fatalf("verified=%v password=%v version=%d", verified, stored, ver)
	}
	if e.count(`SELECT count(*) FROM user_identities WHERE user_id = ?`, uid) != 0 || e.count(`SELECT count(*) FROM account_links WHERE local_user_id = ?`, uid) != 0 {
		t.Fatal("planted identity or link survived")
	}
}

// A typo is not consent to wipe: 4xx, nothing changes, the token still works.
func TestT4b_WrongPasswordChangesNothingAndKeepsToken(t *testing.T) {
	e := newEnv(t)
	uid := e.squat("my-own-password1")
	e.seedGoogleIdentity(uid, "g", "g@gmail.com")
	tok := e.mail.last(t, victim, reVerifyToken)

	code, body := e.do("POST", "/api/auth/verify-email", map[string]any{"token": tok, "password": "typo"}, "")
	if code != 400 || body["code"] != "WRONG_PASSWORD" {
		t.Fatalf("wrong password: %d %v", code, body)
	}
	_, pw, verified, ver := e.userRow(uid)
	if verified || pw == nil || ver != 0 || e.count(`SELECT count(*) FROM user_identities WHERE user_id = ?`, uid) != 1 {
		t.Fatalf("a wrong password changed the account: verified=%v pw=%v ver=%d", verified, pw != nil, ver)
	}
	if e.count(`SELECT count(*) FROM email_verifications WHERE token = ?`, auth_module.HashToken(tok)) != 1 {
		t.Fatal("a wrong password consumed the token")
	}
	if e.count(`SELECT count(*) FROM email_verifications WHERE token = ?`, tok) != 0 {
		t.Fatal("the raw token is stored in the database")
	}
	// neither password nor skip: refused without touching anything
	if code, body = e.do("POST", "/api/auth/verify-email", map[string]any{"token": tok}, ""); code != 400 || body["code"] != "PASSWORD_OR_SKIP_REQUIRED" {
		t.Fatalf("no choice: %d %v", code, body)
	}
	// retry with the right password still works
	if code, body = e.do("POST", "/api/auth/verify-email", map[string]any{"token": tok, "password": "my-own-password1"}, ""); code != 200 || body["password_reset_required"] != false {
		t.Fatalf("retry: %d %v", code, body)
	}
}

// Wrong guesses are capped per token and the count survives the rollback; past the cap
// even the right password is refused (no wipe) and the owner must resend.
func TestVerifyTokenAttemptCap(t *testing.T) {
	e := newEnv(t)
	uid := e.squat("attacker-pass1")
	tok := e.mail.last(t, victim, reVerifyToken)
	for i := 0; i < auth_module.MaxVerifyAttempts; i++ {
		if _, err := e.auth.VerifyEmail(tok, ptr("guess"), false); err != auth_module.ErrVerifyWrongPassword {
			t.Fatalf("guess %d: %v", i, err)
		}
	}
	if _, err := e.auth.VerifyEmail(tok, ptr("attacker-pass1"), false); err != auth_module.ErrVerifyAttemptsExceeded {
		t.Fatalf("past the cap: %v", err)
	}
	if _, pw, verified, ver := e.userRow(uid); verified || pw == nil || ver != 0 {
		t.Fatal("capped token changed the account")
	}
	if _, err := e.auth.VerifyEmail(tok, nil, true); err != auth_module.ErrInvalidToken {
		t.Fatalf("a capped token must be gone, even for skip: %v", err)
	}
	// resend gives a fresh token with a fresh counter
	if err := e.auth.ResendVerification(victim); err != nil {
		t.Fatal(err)
	}
	if _, err := e.auth.VerifyEmail(e.mail.last(t, victim, reVerifyToken), ptr("attacker-pass1"), false); err != nil {
		t.Fatalf("fresh token: %v", err)
	}
}

// GET changes nothing (mail scanners follow links); POST does the work.
func TestVerifyEmailGetIsReadOnlyAndPostVerifies(t *testing.T) {
	e := newEnv(t)
	uid := e.squat("attacker-pass1")
	tok := e.mail.last(t, victim, reVerifyToken)

	code, body := e.do("GET", "/api/auth/verify-email?token="+tok, nil, "")
	if code != 400 || body["code"] != "VERIFY_EMAIL_REQUIRES_POST" {
		t.Fatalf("GET: %d %v", code, body)
	}
	if _, pw, verified, ver := e.userRow(uid); verified || pw == nil || ver != 0 {
		t.Fatal("GET changed the account")
	}
	if e.count(`SELECT count(*) FROM email_verifications WHERE user_id = ?`, uid) != 1 {
		t.Fatal("GET consumed the token")
	}

	code, body = e.do("POST", "/api/auth/verify-email", map[string]string{"token": tok, "password": "attacker-pass1"}, "")
	if code != 200 || body["password_reset_required"] != false {
		t.Fatalf("POST with the right password: %d %v", code, body)
	}
	if _, _, verified, _ := e.userRow(uid); !verified {
		t.Fatal("POST did not verify")
	}
}

// A typed password on a passwordless account can never match: it is a wrong guess,
// not consent to wipe, so nothing is verified or wiped.
func TestVerifyPasswordlessAccountWithTypedPasswordIsWrong(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "gsuite@corp.test", false, "")
	e.seedGoogleIdentity(uid, "planted-sub", "planted@gmail.com")
	if err := e.auth.SendVerification(uid, "gsuite@corp.test"); err != nil {
		t.Fatal(err)
	}
	tok := e.mail.last(t, "gsuite@corp.test", reVerifyToken)
	if _, err := e.auth.VerifyEmail(tok, ptr("anything-123"), false); err != auth_module.ErrVerifyWrongPassword {
		t.Fatalf("typed password on a passwordless account: %v", err)
	}
	if _, _, verified, ver := e.userRow(uid); verified || ver != 0 {
		t.Fatal("wrong guess changed the account")
	}
	if e.count(`SELECT count(*) FROM user_identities WHERE user_id = ?`, uid) != 1 {
		t.Fatal("wrong guess wiped identities")
	}
}

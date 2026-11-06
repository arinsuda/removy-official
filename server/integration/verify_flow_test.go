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

	res, err := e.auth.VerifyEmail(e.mail.last(t, victim, reVerifyToken), ptr("my-own-password1"))
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

func TestT4b_VerifyWithWrongPasswordWipesAndAsksForReset(t *testing.T) {
	for name, pw := range map[string]*string{"wrong": ptr("not-the-password"), "none": nil} {
		t.Run(name, func(t *testing.T) {
			e := newEnv(t)
			uid := e.squat("attacker-pass1")
			e.seedGoogleIdentity(uid, "atk-google", "atk@gmail.com")
			e.db.Exec(`INSERT INTO account_links (local_user_id, remote_platform, remote_user_id, linked_via) VALUES (?, 'planzi', 'p-atk', 'auto')`, uid)

			res, err := e.auth.VerifyEmail(e.mail.last(t, victim, reVerifyToken), pw)
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
		})
	}
}

// A mistyped password burns the token: no second guess, so the link cannot be used to brute-force the password.
func TestVerifyTokenIsSingleAttempt(t *testing.T) {
	e := newEnv(t)
	e.squat("attacker-pass1")
	tok := e.mail.last(t, victim, reVerifyToken)
	if _, err := e.auth.VerifyEmail(tok, ptr("guess-1")); err != nil {
		t.Fatal(err)
	}
	if _, err := e.auth.VerifyEmail(tok, ptr("attacker-pass1")); err != auth_module.ErrInvalidToken {
		t.Fatalf("second attempt with the same token: %v", err)
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

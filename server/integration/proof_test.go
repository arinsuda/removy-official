//go:build integration

package integration

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/arinsuda/movie-hub/internal/auth_module"
	"github.com/arinsuda/movie-hub/internal/link_module"
	"github.com/arinsuda/movie-hub/internal/user_module"
)

const victim = "victim@gmail.com"

// squat registers victim@ the way an attacker would: password set, mailbox not theirs.
func (e *env) squat(password string) uint {
	e.t.Helper()
	u, err := e.auth.Register(auth_module.RegisterRequest{
		Username: "attacker" + randomSuffix()[:4], Email: victim, Password: password, ConfirmPassword: password,
	})
	if err != nil {
		e.t.Fatalf("register: %v", err)
	}
	return u.ID
}

func (e *env) login(identifier, password string) error {
	_, _, err := e.auth.Login(auth_module.LoginRequest{Identifier: identifier, Password: password}, "ua", "1.2.3.4")
	return err
}

func (e *env) forgotAndReset(uid uint, email, newPassword string) error {
	if err := e.auth.ForgotPassword(email); err != nil {
		e.t.Fatal(err)
	}
	tok := e.mail.last(e.t, email, reResetToken)
	return e.auth.ResetPassword(uid, tok, user_module.ResetPasswordRequest{NewPassword: newPassword, ConfirmPassword: newPassword})
}

// T4: the victim clicks the verification link; the attacker's password, planted
// Google identity, link and handoff token must not survive.
func TestT4_VictimVerifiesSquattedAccount_PlantedCredentialsWiped(t *testing.T) {
	e := newEnv(t)
	uid := e.squat("attacker-pass1")
	e.seedGoogleIdentity(uid, "attacker-google-sub", "attacker@gmail.com")
	e.seedRefreshToken(uid, "atk-refresh")
	e.db.Exec(`INSERT INTO account_links (local_user_id, remote_platform, remote_user_id, linked_via) VALUES (?, 'planzi', 'atk-planzi', 'auto')`, uid)
	e.db.Exec(`INSERT INTO link_handoff_tokens (token_hash, user_id, target_platform, expires_at) VALUES (repeat('a',64), ?, 'planzi', now() + interval '1 minute')`, uid)

	if err := e.login(victim, "attacker-pass1"); !errors.Is(err, auth_module.ErrEmailUnverified) {
		t.Fatalf("unverified login must stay blocked, got %v", err)
	}

	tok := e.mail.last(t, victim, reVerifyToken)
	if err := e.auth.VerifyEmail(tok, nil); err != nil {
		t.Fatalf("victim verify: %v", err)
	}

	_, pw, verified, ver := e.userRow(uid)
	if !verified || pw != nil || ver != 1 {
		t.Fatalf("after proof: verified=%v password=%v version=%d", verified, pw, ver)
	}
	if err := e.login(victim, "attacker-pass1"); !errors.Is(err, auth_module.ErrWrongPassword) {
		t.Fatalf("attacker password must fail, got %v", err)
	}
	for _, table := range []string{"user_identities", "account_links", "link_handoff_tokens", "refresh_tokens"} {
		col := "user_id"
		if table == "account_links" {
			col = "local_user_id"
		}
		if n := e.count(`SELECT count(*) FROM `+table+` WHERE `+col+` = ?`, uid); n != 0 {
			t.Errorf("%s still has %d rows", table, n)
		}
	}
}

// T2: reset tokens do not outlive the proof or the address they were minted for.
func TestT2_LeftoverResetTokenRejectedAfterClaim(t *testing.T) {
	t.Run("minted before the victim proves the address", func(t *testing.T) {
		e := newEnv(t)
		uid := e.squat("attacker-pass1")
		// Attacker obtained a reset token for this account while it was theirs.
		if err := e.auth.ForgotPassword(victim); err != nil {
			t.Fatal(err)
		}
		leftover := e.mail.last(t, victim, reResetToken)

		if err := e.auth.VerifyEmail(e.mail.last(t, victim, reVerifyToken), nil); err != nil {
			t.Fatal(err)
		}
		err := e.auth.ResetPassword(uid, leftover, user_module.ResetPasswordRequest{NewPassword: "attacker-new1", ConfirmPassword: "attacker-new1"})
		if err == nil {
			t.Fatal("leftover reset token still worked after the claim")
		}
		if _, pw, _, _ := e.userRow(uid); pw != nil {
			t.Fatal("password set through leftover token")
		}
	})

	t.Run("minted for a previous address", func(t *testing.T) {
		e := newEnv(t)
		uid := seedUser(t, e.db, "attacker@gmail.com", true, "attacker-pass1")
		if err := e.auth.ForgotPassword("attacker@gmail.com"); err != nil {
			t.Fatal(err)
		}
		leftover := e.mail.last(t, "attacker@gmail.com", reResetToken)
		// The account's address becomes the victim's (any path that changes it).
		e.db.Exec(`UPDATE users SET email = ? WHERE id = ?`, victim, uid)

		err := e.auth.ResetPassword(uid, leftover, user_module.ResetPasswordRequest{NewPassword: "attacker-new1", ConfirmPassword: "attacker-new1"})
		if !errors.Is(err, user_module.ErrPasswordResetTokenInvalid) {
			t.Fatalf("token bound to the old address must be invalid, got %v", err)
		}
	})
}

// T3: a link code mailed to the old address must not link after the email changed.
func TestT3_LinkCodeToOldEmailRejectedAfterEmailChange(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "attacker@gmail.com", false, "attacker-pass1")
	e.remote.mintToken("tok-1", link_module.RemoteUser{Platform: "planzi", UserID: "p-victim", Email: "attacker@gmail.com", EmailVerified: true})
	_, err := e.link.Redeem(context.Background(), "tok-1")
	var ce *link_module.ConfirmRequiredError
	if !errors.As(err, &ce) {
		t.Fatalf("expected confirm required, got %v", err)
	}
	if err := e.link.SendCode(ce.PendingID); err != nil {
		t.Fatal(err)
	}
	code := e.mail.last(t, "attacker@gmail.com", reCode)

	// Attacker (who reads that mailbox) now moves the account to the victim's address.
	e.db.Exec(`UPDATE users SET email = ? WHERE id = ?`, victim, uid)

	if _, err := e.link.Confirm(ce.PendingID, link_module.MethodEmail, "", code); !errors.Is(err, link_module.ErrBadCredentials) {
		t.Fatalf("code for the old address must be rejected, got %v", err)
	}
	if n := e.count(`SELECT count(*) FROM account_links WHERE local_user_id = ?`, uid); n != 0 {
		t.Fatalf("link created despite stale code")
	}
}

// The email code proves the mailbox, not the person: an unverified account loses its password at link time.
func TestLinkEmailCodeProvesMailboxOnlyAndWipesPlantedPassword(t *testing.T) {
	e := newEnv(t)
	uid := e.squat("attacker-pass1")
	e.seedGoogleIdentity(uid, "atk-sub", "attacker@gmail.com")
	e.remote.mintToken("tok-1", link_module.RemoteUser{Platform: "planzi", UserID: "p-victim", Email: victim, EmailVerified: true})
	_, err := e.link.Redeem(context.Background(), "tok-1")
	var ce *link_module.ConfirmRequiredError
	if !errors.As(err, &ce) {
		t.Fatalf("expected confirm required, got %v", err)
	}
	if err := e.link.SendCode(ce.PendingID); err != nil {
		t.Fatal(err)
	}
	if _, err := e.link.Confirm(ce.PendingID, link_module.MethodEmail, "", e.mail.last(t, victim, reCode)); err != nil {
		t.Fatal(err)
	}
	_, pw, verified, _ := e.userRow(uid)
	if !verified || pw != nil {
		t.Fatalf("verified=%v password=%v", verified, pw)
	}
	if n := e.count(`SELECT count(*) FROM user_identities WHERE user_id = ?`, uid); n != 0 {
		t.Fatal("planted google identity survived")
	}
	if n := e.count(`SELECT count(*) FROM account_links WHERE local_user_id = ? AND remote_user_id = 'p-victim'`, uid); n != 1 {
		t.Fatal("the confirmed link was not created")
	}
}

// T9: a passwordless account proves its address, cannot log in with a password, and recovers by reset.
func TestT9_PasswordlessUserProvesThenRecoversByReset(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "gsuite@corp.test", false, "")
	if err := e.auth.SendVerification(uid, "gsuite@corp.test"); err != nil {
		t.Fatal(err)
	}
	if err := e.auth.VerifyEmail(e.mail.last(t, "gsuite@corp.test", reVerifyToken), nil); err != nil {
		t.Fatal(err)
	}
	if err := e.login("gsuite@corp.test", "anything-123"); !errors.Is(err, auth_module.ErrWrongPassword) {
		t.Fatalf("passwordless login must fail, got %v", err)
	}
	if err := e.forgotAndReset(uid, "gsuite@corp.test", "brand-new-pw1"); err != nil {
		t.Fatalf("reset: %v", err)
	}
	if err := e.login("gsuite@corp.test", "brand-new-pw1"); err != nil {
		t.Fatalf("login after reset: %v", err)
	}
}

// N5: a token mailed for the old address must not verify the new one.
func TestN5_VerificationTokenBoundToItsAddress(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "old@x.test", false, "pw-oldoldold")
	// ResendVerification read the old address and mailed the token there.
	if err := e.auth.SendVerification(uid, "old@x.test"); err != nil {
		t.Fatal(err)
	}
	tok := e.mail.last(t, "old@x.test", reVerifyToken)
	// ...while the email changed to the victim's address.
	e.db.Exec(`UPDATE users SET email = ? WHERE id = ?`, victim, uid)

	if err := e.auth.VerifyEmail(tok, nil); !errors.Is(err, auth_module.ErrInvalidToken) {
		t.Fatalf("token for the old address must not verify the new one, got %v", err)
	}
	if _, _, verified, _ := e.userRow(uid); verified {
		t.Fatal("new address was marked verified")
	}
}

// Legacy token rows (no bound email) prove nothing.
func TestLegacyVerificationTokenWithoutEmailRejected(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "old@x.test", false, "pw-oldoldold")
	e.db.Exec(`INSERT INTO email_verifications (user_id, token, expires_at, created_at) VALUES (?, 'legacy', now() + interval '1 hour', now())`, uid)
	if err := e.auth.VerifyEmail("legacy", nil); !errors.Is(err, auth_module.ErrInvalidToken) {
		t.Fatalf("got %v", err)
	}
}

// Reset on a verified account replaces the password, ends sessions, keeps identities and links.
func TestResetOnVerifiedAccountKeepsIdentitiesAndLinks(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "real@x.test", true, "old-password1")
	e.seedGoogleIdentity(uid, "real-sub", "real@x.test")
	e.db.Exec(`INSERT INTO account_links (local_user_id, remote_platform, remote_user_id, linked_via) VALUES (?, 'planzi', 'p-1', 'auto')`, uid)
	e.seedRefreshToken(uid, "old-session")

	if err := e.forgotAndReset(uid, "real@x.test", "new-password1"); err != nil {
		t.Fatal(err)
	}
	if e.count(`SELECT count(*) FROM user_identities WHERE user_id = ?`, uid) != 1 || e.count(`SELECT count(*) FROM account_links WHERE local_user_id = ?`, uid) != 1 {
		t.Fatal("identity or link lost on a verified reset")
	}
	if e.count(`SELECT count(*) FROM refresh_tokens WHERE user_id = ?`, uid) != 0 {
		t.Fatal("sessions survived the reset")
	}
	if _, _, _, ver := e.userRow(uid); ver != 1 {
		t.Fatalf("version %d", ver)
	}
	if err := e.login("real@x.test", "new-password1"); err != nil {
		t.Fatal(err)
	}
}

// Reset on an unverified account proves the mailbox: planted identities go, the new password stays.
func TestResetOnUnverifiedAccountWipesPlantedCredentials(t *testing.T) {
	e := newEnv(t)
	uid := e.squat("attacker-pass1")
	e.seedGoogleIdentity(uid, "atk-sub", "attacker@gmail.com")
	if err := e.forgotAndReset(uid, victim, "victim-new-pw1"); err != nil {
		t.Fatal(err)
	}
	if e.count(`SELECT count(*) FROM user_identities WHERE user_id = ?`, uid) != 0 {
		t.Fatal("planted identity survived")
	}
	if err := e.login(victim, "attacker-pass1"); err == nil {
		t.Fatal("attacker password still works")
	}
	if err := e.login(victim, "victim-new-pw1"); err != nil {
		t.Fatalf("victim cannot log in: %v", err)
	}
}

// Google authoritative login on a squatted account: proof wipes the password, identity re-attached.
func TestGoogleAuthoritativeLoginWipesSquatterPassword(t *testing.T) {
	e := newEnv(t)
	uid := e.squat("attacker-pass1")
	e.seedGoogleIdentity(uid, "atk-sub", "attacker@gmail.com")
	e.google.set(auth_module.GoogleIdentityClaims{Subject: "victim-sub", Email: victim, EmailVerified: true})

	if _, _, err := e.googleLogin(); err != nil {
		t.Fatalf("google login: %v", err)
	}
	_, pw, verified, _ := e.userRow(uid)
	if !verified || pw != nil {
		t.Fatalf("verified=%v pw=%v", verified, pw)
	}
	var subs []string
	e.db.Raw(`SELECT provider_subject FROM user_identities WHERE user_id = ?`, uid).Scan(&subs)
	if len(subs) != 1 || subs[0] != "victim-sub" {
		t.Fatalf("identities = %v, want only the victim's", subs)
	}
	if err := e.login(victim, "attacker-pass1"); err == nil {
		t.Fatal("attacker password still works")
	}
}

// T10: two concurrent Google logins for one squatted account end with exactly one identity.
func TestT10_ConcurrentGoogleLoginSingleIdentity(t *testing.T) {
	e := newEnv(t)
	uid := e.squat("attacker-pass1")
	e.google.set(auth_module.GoogleIdentityClaims{Subject: "victim-sub", Email: victim, EmailVerified: true})

	done := make(chan error, 6)
	for i := 0; i < 6; i++ {
		go func() { _, _, err := e.googleLogin(); done <- err }()
	}
	okCount := 0
	for i := 0; i < 6; i++ {
		if err := <-done; err == nil {
			okCount++
		} else if !strings.Contains(err.Error(), "already") {
			t.Logf("concurrent login error: %v", err)
		}
	}
	if okCount == 0 {
		t.Fatal("no concurrent google login succeeded")
	}
	if n := e.count(`SELECT count(*) FROM user_identities WHERE user_id = ?`, uid); n != 1 {
		t.Fatalf("identities = %d, want 1", n)
	}
}

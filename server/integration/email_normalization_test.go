//go:build integration

package integration

import (
	"testing"

	"github.com/arinsuda/movie-hub/internal/auth_module"
)

// T4c: "Victim@" registered first, then Google login with "victim@": one account, not two.
func TestT4c_MixedCaseRegisterThenGoogleLoginReachesOneAccount(t *testing.T) {
	e := newEnv(t)
	u, err := e.auth.Register(auth_module.RegisterRequest{Username: "squatter", Email: "  Victim@Gmail.COM ", Password: "attacker-pass1", ConfirmPassword: "attacker-pass1"})
	if err != nil {
		t.Fatal(err)
	}
	if u.Email != victim {
		t.Fatalf("stored email %q, want normalized %q", u.Email, victim)
	}
	e.google.set(auth_module.GoogleIdentityClaims{Subject: "victim-sub", Email: victim, EmailVerified: true})
	if _, gu, err := e.googleLogin(); err != nil || gu.ID != u.ID {
		t.Fatalf("google login: user=%v err=%v", gu, err)
	}
	if n := e.count(`SELECT count(*) FROM users WHERE lower(email) = ?`, victim); n != 1 {
		t.Fatalf("%d accounts for one mailbox", n)
	}
	if _, pw, verified, _ := e.userRow(u.ID); !verified || pw != nil {
		t.Fatalf("verified=%v password=%v", verified, pw)
	}
}

// Legacy mixed-case rows are still reachable by every lookup.
func TestLegacyMixedCaseRowMatchedByGoogleLoginAndPasswordLogin(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "Legacy.User@Gmail.com", true, "pw-secret-12")
	if err := e.login("legacy.user@gmail.com", "pw-secret-12"); err != nil {
		t.Fatalf("password login with other case: %v", err)
	}
	if err := e.login("  LEGACY.USER@GMAIL.COM ", "pw-secret-12"); err != nil {
		t.Fatalf("password login with spaces/upper: %v", err)
	}
	e.google.set(auth_module.GoogleIdentityClaims{Subject: "legacy-sub", Email: "legacy.user@gmail.com", EmailVerified: true})
	if _, gu, err := e.googleLogin(); err != nil || gu.ID != uid {
		t.Fatalf("google login: user=%v err=%v", gu, err)
	}
	if n := e.count(`SELECT count(*) FROM users WHERE lower(email) = 'legacy.user@gmail.com'`); n != 1 {
		t.Fatalf("a second account was created: %d", n)
	}
}

// With case-variant duplicates, the verified row is the one that is found.
func TestVerifiedRowPreferredAmongCaseDuplicates(t *testing.T) {
	e := newEnv(t)
	seedUser(t, e.db, "dup@x.test", false, "squatter-pass1")
	good := seedUser(t, e.db, "DUP@x.test", true, "owner-pass-12")
	if err := e.login("dup@x.test", "owner-pass-12"); err != nil {
		t.Fatalf("owner must reach the verified row: %v", err)
	}
	e.google.set(auth_module.GoogleIdentityClaims{Subject: "dup-sub", Email: "dup@x.test", EmailVerified: true, HostedDomain: "x.test"})
	if _, gu, err := e.googleLogin(); err != nil || gu.ID != good {
		t.Fatalf("google login must pick the verified row: user=%v err=%v", gu, err)
	}
}

func TestResendAndForgotPasswordAreCaseInsensitive(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "mixed@x.test", false, "pw-secret-12")
	if err := e.auth.ResendVerification("  MIXED@X.TEST "); err != nil {
		t.Fatal(err)
	}
	e.mail.last(t, "mixed@x.test", reVerifyToken)
	if err := e.auth.ForgotPassword("MIXED@x.test"); err != nil {
		t.Fatal(err)
	}
	e.mail.last(t, "mixed@x.test", reResetToken)
	_ = uid
}

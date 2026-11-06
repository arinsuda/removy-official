//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/internal/auth_module"
	"github.com/arinsuda/movie-hub/internal/identity_policy"
	"github.com/arinsuda/movie-hub/internal/link_module"
)

func (e *env) mintHandoffRow(uid uint, token string) {
	e.t.Helper()
	// the same hash the service uses: sha256 hex of the raw token
	if err := e.db.Exec(`INSERT INTO link_handoff_tokens (token_hash, user_id, target_platform, expires_at)
		VALUES (encode(sha256(?::bytea), 'hex'), ?, 'planzi', now() + interval '1 minute')`, token, uid).Error; err != nil {
		e.t.Fatal(err)
	}
}

// T7: a handoff token minted before a claim is dead afterwards.
func TestT7_HandoffTokenMintedBeforeClaimRejected(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, victim, true, "attacker-pass1")
	url, _, err := e.link.CreateHandoff(uid, 0, "planzi")
	if err != nil {
		t.Fatal(err)
	}
	token := url[len("http://planzi.test/auth/handoff#t="):]

	e.claimByReset(uid, victim)
	time.Sleep(time.Second) // "+1s": still inside the 60s handoff lifetime

	if _, err := e.link.Introspect(token); !errors.Is(err, link_module.ErrInvalidToken) {
		t.Fatalf("pre-claim handoff token must be rejected, got %v", err)
	}
}

// Control: an untouched token of a verified account still introspects (and only once).
func TestHandoffTokenOfVerifiedAccountStillWorksOnce(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "me@x.test", true, "pw-secret-12")
	url, _, err := e.link.CreateHandoff(uid, 0, "planzi")
	if err != nil {
		t.Fatal(err)
	}
	token := url[len("http://planzi.test/auth/handoff#t="):]
	ru, err := e.link.Introspect(token)
	if err != nil || ru.Email != "me@x.test" || !ru.EmailVerified {
		t.Fatalf("introspect: %v %+v", err, ru)
	}
	if _, err := e.link.Introspect(token); err == nil {
		t.Fatal("handoff token worked twice")
	}
}

// The race itself: introspect waits for the claim in flight, then finds the token gone.
func TestT7_IntrospectDuringClaimFindsTokenGone(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, victim, true, "attacker-pass1")
	url, _, _ := e.link.CreateHandoff(uid, 0, "planzi")
	token := url[len("http://planzi.test/auth/handoff#t="):]

	tx := e.db.Begin()
	if _, err := identity_policy.ProveEmail(tx, identity_policy.Proof{UserID: uid, Email: victim, Kind: identity_policy.KindPasswordReset, NewPasswordHash: ptr("$2a$04$x")}); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := e.link.Introspect(token); done <- err }()
	time.Sleep(300 * time.Millisecond)
	if err := tx.Commit().Error; err != nil {
		t.Fatal(err)
	}
	if err := <-done; !errors.Is(err, link_module.ErrInvalidToken) {
		t.Fatalf("introspect racing a claim: %v", err)
	}
}

// I2 on the internal API: an unverified account is never handed to the other platform.
func TestIntrospectAndVerifyPasswordRefuseUnverifiedAccount(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "plain@x.test", false, "pw-secret-12")
	e.mintHandoffRow(uid, "pre-deploy-token")
	if _, err := e.link.Introspect("pre-deploy-token"); !errors.Is(err, link_module.ErrInvalidToken) {
		t.Fatalf("introspect unverified: %v", err)
	}
	if _, ok := e.link.VerifyPassword("plain@x.test", "pw-secret-12"); ok {
		t.Fatal("verify-password accepted an unverified account")
	}
	// control: verified account passes
	seedUser(t, e.db, "ok@x.test", true, "pw-secret-12")
	if ru, ok := e.link.VerifyPassword("ok@x.test", "pw-secret-12"); !ok || ru.Email != "ok@x.test" {
		t.Fatal("verified account must still verify")
	}
}

func TestIntrospectEmitsIdentityResetAt(t *testing.T) {
	e := newEnv(t)
	uid := e.squat("attacker-pass1")
	if err := e.verify(e.mail.last(t, victim, reVerifyToken), nil); err != nil { // wipe: stamps identity_reset_at
		t.Fatal(err)
	}
	url, _, err := e.link.CreateHandoff(uid, 1, "planzi")
	if err != nil {
		t.Fatal(err)
	}
	ru, err := e.link.Introspect(url[len("http://planzi.test/auth/handoff#t="):])
	if err != nil {
		t.Fatal(err)
	}
	if ru.IdentityResetAt == nil || time.Since(*ru.IdentityResetAt) > time.Minute {
		t.Fatalf("identity_reset_at = %v", ru.IdentityResetAt)
	}
}

// D1: an unverified PLANZI user gets no POPCORN account and no session.
func TestRedeemUnverifiedRemoteWithoutLocalProvisionsNothing(t *testing.T) {
	e := newEnv(t)
	e.remote.mintToken("tok", link_module.RemoteUser{Platform: "planzi", UserID: "p-1", Email: "Fresh@X.test", EmailVerified: false, PasswordHash: "$2a$04$abcdefghijklmnopqrstuuAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"})
	if _, err := e.link.Redeem(context.Background(), "tok"); !errors.Is(err, link_module.ErrRemoteUnverified) {
		t.Fatalf("got %v", err)
	}
	if n := e.count(`SELECT count(*) FROM users WHERE lower(email) = 'fresh@x.test'`); n != 0 {
		t.Fatal("a local user was provisioned from an unverified remote account")
	}
}

// Control: a verified remote user is provisioned, lower-cased, verified.
func TestRedeemVerifiedRemoteProvisionsNormalizedVerifiedUser(t *testing.T) {
	e := newEnv(t)
	e.remote.mintToken("tok", link_module.RemoteUser{Platform: "planzi", UserID: "p-2", Email: " Fresh@X.test ", EmailVerified: true})
	res, err := e.link.Redeem(context.Background(), "tok")
	if err != nil || !res.Provisioned {
		t.Fatalf("redeem: %v %+v", err, res)
	}
	if res.User.Email != "fresh@x.test" || res.User.VerifiedEmailAt == nil {
		t.Fatalf("email=%q verified=%v", res.User.Email, res.User.VerifiedEmailAt)
	}
}

// F7/N3: a link that points at an account that never proved its email must not log in.
func TestRedeemAlreadyLinkedUnverifiedLocalNeedsConfirmation(t *testing.T) {
	e := newEnv(t)
	uid := e.squat("attacker-pass1")
	e.db.Exec(`INSERT INTO account_links (local_user_id, remote_platform, remote_user_id, linked_via) VALUES (?, 'planzi', 'p-victim', 'auto')`, uid)
	e.remote.mintToken("tok", link_module.RemoteUser{Platform: "planzi", UserID: "p-victim", Email: victim, EmailVerified: true})

	res, err := e.link.Redeem(context.Background(), "tok")
	var ce *link_module.ConfirmRequiredError
	if res != nil || !errors.As(err, &ce) {
		t.Fatalf("expected confirm required with no session, got res=%v err=%v", res, err)
	}
	if len(ce.Methods) != 1 || ce.Methods[0] != link_module.MethodEmail {
		t.Fatalf("methods = %v, want only email code", ce.Methods)
	}
	// the victim confirms by mailbox: planted password goes, the link is re-created
	if err := e.link.SendCode(ce.PendingID); err != nil {
		t.Fatal(err)
	}
	if _, err := e.link.Confirm(ce.PendingID, link_module.MethodEmail, "", e.mail.last(t, victim, reCode)); err != nil {
		t.Fatal(err)
	}
	if _, pw, verified, _ := e.userRow(uid); !verified || pw != nil {
		t.Fatalf("verified=%v password=%v", verified, pw)
	}
	if n := e.count(`SELECT count(*) FROM account_links WHERE local_user_id = ? AND remote_user_id = 'p-victim'`, uid); n != 1 {
		t.Fatalf("links = %d", n)
	}
}

// Unverified remote + verified local with the same email: confirm, never auto-link.
func TestRedeemUnverifiedRemoteDoesNotAutoLinkVerifiedLocal(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "same@x.test", true, "pw-secret-12")
	e.remote.mintToken("tok", link_module.RemoteUser{Platform: "planzi", UserID: "p-3", Email: "same@x.test", EmailVerified: false})
	_, err := e.link.Redeem(context.Background(), "tok")
	var ce *link_module.ConfirmRequiredError
	if !errors.As(err, &ce) {
		t.Fatalf("got %v", err)
	}
	if e.count(`SELECT count(*) FROM account_links WHERE local_user_id = ?`, uid) != 0 {
		t.Fatal("auto-linked on an unverified remote email")
	}
}

// Login fallback: an unverified PLANZI account with the right password creates no user and no session.
func TestLoginFallbackRefusesUnverifiedRemote(t *testing.T) {
	e := newEnv(t)
	e.remote.addUser(link_module.RemoteUser{Platform: "planzi", UserID: "p-9", Email: "planzi-only@x.test", EmailVerified: false})
	_, _, err := e.auth.Login(auth_module.LoginRequest{Identifier: "planzi-only@x.test", Password: "remote-pw"}, "ua", "ip")
	if !errors.Is(err, auth_module.ErrUserNotFound) {
		t.Fatalf("got %v", err)
	}
	if n := e.count(`SELECT count(*) FROM users WHERE email = 'planzi-only@x.test'`); n != 0 {
		t.Fatal("login fallback provisioned an unverified remote user")
	}
	// control: a verified PLANZI account still gets in through the fallback
	e.remote.addUser(link_module.RemoteUser{Platform: "planzi", UserID: "p-10", Email: "verified-planzi@x.test", EmailVerified: true})
	if pair, _, err := e.auth.Login(auth_module.LoginRequest{Identifier: "verified-planzi@x.test", Password: "remote-pw"}, "ua", "ip"); err != nil || pair == nil {
		t.Fatalf("verified fallback: %v", err)
	}
}

//go:build integration

package integration

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/internal/auth_module"
	"github.com/arinsuda/movie-hub/internal/identity_policy"
	"github.com/arinsuda/movie-hub/internal/link_module"
	"github.com/arinsuda/movie-hub/internal/user_module"
	"github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// accessToken signs an access token the way the server does, for any epoch.
func (e *env) accessToken(uid uint, ver int) string {
	e.t.Helper()
	claims := middleware.Claims{
		UserID: uid, Role: "user", Type: middleware.AccessToken, TokenVersion: ver,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)), ID: randomSuffix()},
	}
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(e.cfg.JWT.AccessSecret))
	if err != nil {
		e.t.Fatal(err)
	}
	return s
}

// reset performs a password reset as the mailbox owner (the claim that bumps the epoch).
func (e *env) claimByReset(uid uint, email string) {
	e.t.Helper()
	if err := e.forgotAndReset(uid, email, "victim-new-pw1"); err != nil {
		e.t.Fatalf("claim: %v", err)
	}
}

// T1: a token minted before a claim stops working on every protected path afterwards.
func TestT1_AttackerJWTRejectedAfterClaim(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, victim, true, "attacker-pass1")
	atk := e.accessToken(uid, 0)

	if code, _ := e.do("GET", "/api/ping", nil, atk); code != 200 {
		t.Fatalf("token should work before the claim, got %d", code)
	}
	e.claimByReset(uid, victim)

	routes := []struct {
		name, method, path string
		body               any
	}{
		{"any protected route", "GET", "/api/ping", nil},
		{"password change", "PATCH", "/api/users/1/password", map[string]string{"old_password": "attacker-pass1", "new_password": "attacker-new11", "confirm_password": "attacker-new11"}},
		{"google link", "POST", "/api/auth/google/link", map[string]string{}},
		{"handoff", "POST", "/api/link/handoff", map[string]string{"target": "planzi"}},
		{"delete account", "DELETE", "/api/users/1", nil},
	}
	for _, r := range routes {
		if code, _ := e.do(r.method, r.path, r.body, atk); code != 401 {
			t.Errorf("%s: want 401 after claim, got %d", r.name, code)
		}
	}
	var active bool
	e.db.Raw(`SELECT is_active FROM users WHERE id = ?`, uid).Scan(&active)
	if !active {
		t.Error("stale token deleted the account")
	}
}

// The crucial shape: unverified account, token minted before the deploy (epoch 0).
// The victim's proof makes the account verified, so only the epoch tells the tokens apart.
func TestT1_PreDeployTokenDeadAfterVictimProof(t *testing.T) {
	e := newEnv(t)
	uid := e.squat("attacker-pass1")
	old := e.accessToken(uid, 0)
	if code, _ := e.do("GET", "/api/ping", nil, old); code != 401 {
		t.Fatalf("unverified account must not hold a session, got %d", code)
	}
	if err := e.verify(e.mail.last(t, victim, reVerifyToken), nil); err != nil {
		t.Fatal(err)
	}
	if code, _ := e.do("GET", "/api/ping", nil, old); code != 401 {
		t.Fatalf("pre-claim token must stay dead after the victim verified, got %d", code)
	}
	if code, _ := e.do("GET", "/api/ping", nil, e.accessToken(uid, 1)); code != 200 {
		t.Fatalf("a token of the new epoch must work, got %d", code)
	}
}

// A request that passed RequireAuth just before a claim must fail when it writes.
func TestT1_InFlightWriteDuringClaimFails(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, victim, true, "attacker-pass1")

	tx := e.db.Begin()
	if _, err := identity_policy.ProveEmail(tx, identity_policy.Proof{UserID: uid, Email: victim, Kind: identity_policy.KindPasswordReset, NewPasswordHash: ptr("$2a$04$x")}); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	var pwErr error
	var handoffErr error
	wg.Add(2)
	go func() {
		defer wg.Done()
		pwErr = e.users.ChangePassword(uid, uid, 0, user_module.ChangePasswordRequest{OldPassword: "attacker-pass1", NewPassword: "attacker-new11", ConfirmPassword: "attacker-new11"})
	}()
	go func() {
		defer wg.Done()
		_, _, handoffErr = e.link.CreateHandoff(uid, 0, "planzi")
	}()
	time.Sleep(300 * time.Millisecond) // both are now waiting on the user row
	if err := tx.Commit().Error; err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	if !errors.Is(pwErr, user_module.ErrSessionStale) {
		t.Errorf("password change in flight: %v", pwErr)
	}
	if !errors.Is(handoffErr, link_module.ErrSessionStale) {
		t.Errorf("handoff in flight: %v", handoffErr)
	}
	if n := e.count(`SELECT count(*) FROM link_handoff_tokens WHERE user_id = ?`, uid); n != 0 {
		t.Errorf("handoff token minted by a stale session: %d", n)
	}
}

func ptr[T any](v T) *T { return &v }

// T8: a refresh racing a claim yields no token; a double refresh yields exactly one.
func TestT8_RefreshRacingClaimYieldsNoToken(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, victim, true, "attacker-pass1")
	pair, _, err := e.auth.Login(auth_module.LoginRequest{Identifier: victim, Password: "attacker-pass1"}, "ua", "ip")
	if err != nil {
		t.Fatal(err)
	}

	tx := e.db.Begin()
	if _, err := identity_policy.ProveEmail(tx, identity_policy.Proof{UserID: uid, Email: victim, Kind: identity_policy.KindPasswordReset, NewPasswordHash: ptr("$2a$04$x")}); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		_, _, err := e.auth.Refresh(pair.RefreshToken, "ua", "ip")
		done <- err
	}()
	time.Sleep(300 * time.Millisecond)
	if err := tx.Commit().Error; err != nil {
		t.Fatal(err)
	}
	if err := <-done; !errors.Is(err, auth_module.ErrInvalidToken) {
		t.Fatalf("racing refresh must fail, got %v", err)
	}
	if n := e.count(`SELECT count(*) FROM refresh_tokens WHERE user_id = ?`, uid); n != 0 {
		t.Fatalf("a session survived the claim: %d refresh tokens", n)
	}
}

// Epoch layer on its own: the refresh token row survives, only the epoch moves.
func TestT8_RefreshAfterEpochBumpFailsEvenIfTokenRowSurvives(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, victim, true, "pw-secret-12")
	pair, _, err := e.auth.Login(auth_module.LoginRequest{Identifier: victim, Password: "pw-secret-12"}, "ua", "ip")
	if err != nil {
		t.Fatal(err)
	}
	tx := e.db.Begin()
	if err := tx.Exec(`UPDATE users SET token_version = token_version + 1 WHERE id = ?`, uid).Error; err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		_, _, err := e.auth.Refresh(pair.RefreshToken, "ua", "ip")
		done <- err
	}()
	time.Sleep(300 * time.Millisecond)
	if err := tx.Commit().Error; err != nil {
		t.Fatal(err)
	}
	if err := <-done; !errors.Is(err, auth_module.ErrInvalidToken) {
		t.Fatalf("refresh with a stale epoch must fail, got %v", err)
	}
	if n := e.count(`SELECT count(*) FROM refresh_tokens WHERE user_id = ?`, uid); n != 1 {
		t.Fatalf("refresh rows = %d: a failed refresh must not rotate or mint", n)
	}
}

// Rotation layer on its own: all refreshes pass the early lookup, then queue on the user row.
func TestT8_DoubleRefreshYieldsOneSession(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, victim, true, "pw-secret-12")
	pair, _, err := e.auth.Login(auth_module.LoginRequest{Identifier: victim, Password: "pw-secret-12"}, "ua", "ip")
	if err != nil {
		t.Fatal(err)
	}
	gate := e.db.Begin() // holds the user row so every refresh is parked inside its transaction
	if err := gate.Exec(`SELECT id FROM users WHERE id = ? FOR UPDATE`, uid).Error; err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	wins := 0
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, _, err := e.auth.Refresh(pair.RefreshToken, "ua", "ip"); err == nil {
				mu.Lock()
				wins++
				mu.Unlock()
			}
		}()
	}
	time.Sleep(400 * time.Millisecond)
	if err := gate.Commit().Error; err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	if wins != 1 {
		t.Fatalf("%d concurrent refreshes succeeded, want 1", wins)
	}
	if n := e.count(`SELECT count(*) FROM refresh_tokens WHERE user_id = ?`, uid); n != 1 {
		t.Fatalf("refresh rows = %d, want 1", n)
	}
}

// N1: a Google link started before a claim must not attach the attacker's Google afterwards.
func TestN1_GoogleLinkCallbackAfterClaimRejected(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, victim, true, "attacker-pass1")
	ctx := context.Background()
	_, state, err := e.auth.InitiateGoogleLink(ctx, uid, "/")
	if err != nil {
		t.Fatal(err)
	}
	e.claimByReset(uid, victim) // victim recovers the account, epoch moves

	e.google.set(auth_module.GoogleIdentityClaims{Subject: "attacker-google", Email: "attacker@gmail.com", EmailVerified: true})
	_, _, _, err = e.auth.HandleGoogleCallback(ctx, "code", state, state, "ua", "ip")
	if !errors.Is(err, identity_policy.ErrSessionStale) {
		t.Fatalf("link callback after claim: %v", err)
	}
	if n := e.count(`SELECT count(*) FROM user_identities WHERE user_id = ?`, uid); n != 0 {
		t.Fatal("attacker google identity attached after the claim")
	}
}

// Without a claim in between, the same link flow still works.
func TestGoogleLinkCallbackStillWorksForLiveSession(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "me@gmail.com", true, "my-password1")
	ctx := context.Background()
	_, state, err := e.auth.InitiateGoogleLink(ctx, uid, "/")
	if err != nil {
		t.Fatal(err)
	}
	e.google.set(auth_module.GoogleIdentityClaims{Subject: "my-google", Email: "me@gmail.com", EmailVerified: true})
	if _, _, _, err := e.auth.HandleGoogleCallback(ctx, "code", state, state, "ua", "ip"); err != nil {
		t.Fatal(err)
	}
	if n := e.count(`SELECT count(*) FROM user_identities WHERE user_id = ?`, uid); n != 1 {
		t.Fatalf("identities = %d", n)
	}
}

// I2: no proof, no session, on every issue path.
func TestUnverifiedAccountNeverGetsSession(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "plain@x.test", false, "pw-secret-12")
	if err := e.login("plain@x.test", "pw-secret-12"); !errors.Is(err, auth_module.ErrEmailUnverified) {
		t.Fatalf("login: %v", err)
	}
	if _, err := e.auth.IssueAndStoreTokens(&user_module.User{Model: modelWithID(uid), RoleID: 2}, "ua", "ip"); !errors.Is(err, auth_module.ErrEmailUnverified) {
		t.Fatalf("issue: %v", err)
	}
	if n := e.count(`SELECT count(*) FROM refresh_tokens WHERE user_id = ?`, uid); n != 0 {
		t.Fatal("refresh token stored for an unverified account")
	}
}

// Google path 3 with a non-authoritative email: account created, verification mailed, no session.
func TestGoogleNonAuthoritativeNewAccountGetsNoSession(t *testing.T) {
	e := newEnv(t)
	e.google.set(auth_module.GoogleIdentityClaims{Subject: "yahoo-sub", Email: "someone@yahoo.com", EmailVerified: false, Name: "Some One"})
	pair, _, err := e.googleLogin()
	if !errors.Is(err, auth_module.ErrGoogleEmailVerificationRequired) || pair != nil {
		t.Fatalf("pair=%v err=%v", pair, err)
	}
	if n := e.count(`SELECT count(*) FROM users WHERE email = 'someone@yahoo.com' AND verified_email_at IS NULL`); n != 1 {
		t.Fatal("account not created unverified")
	}
	e.mail.last(t, "someone@yahoo.com", reVerifyToken) // verification link was mailed
	if n := e.count(`SELECT count(*) FROM refresh_tokens`); n != 0 {
		t.Fatal("a session was stored")
	}
}

// A database failure must answer 503, never 401: the client would log the user out.
func TestRequireAuthDatabaseErrorIs503(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, victim, true, "pw-secret-12")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	app := fiber.New()
	app.Get("/x", middleware.NewAuthMiddleware(e.cfg, e.db.WithContext(ctx)).RequireAuth, func(c fiber.Ctx) error { return c.SendString("ok") })
	e2 := *e
	e2.app = app
	if code, _ := e2.do("GET", "/x", nil, e.accessToken(uid, 0)); code != 503 {
		t.Fatalf("want 503, got %d", code)
	}
}

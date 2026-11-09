//go:build integration

package integration

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/arinsuda/movie-hub/internal/admin_module"
	"github.com/arinsuda/movie-hub/internal/auth_module"
	"github.com/arinsuda/movie-hub/internal/user_module"
)

func (e *env) seedAdmin(email, password string) uint {
	e.t.Helper()
	uid := seedUser(e.t, e.db, email, true, password)
	if err := e.db.Exec(`UPDATE users SET role_id = (SELECT id FROM roles WHERE role_name = 'admin') WHERE id = ?`, uid).Error; err != nil {
		e.t.Fatal(err)
	}
	return uid
}

// Self-delete only: even an admin JWT cannot remove another user through /users/:id.
func TestDeleteUserIsSelfOnly(t *testing.T) {
	e := newEnv(t)
	e.seedAdmin("admin@x.test", "admin-password1")
	victimID := seedUser(t, e.db, "target@x.test", true, "target-password1")
	admin := e.loginPair("admin@x.test", "admin-password1")

	if code, _ := e.do("DELETE", fmt.Sprintf("/api/users/%d", victimID), nil, admin.AccessToken); code != 403 {
		t.Fatalf("admin deleting another user via /users: %d, want 403", code)
	}
	var active bool
	e.db.Raw(`SELECT is_active FROM users WHERE id = ?`, victimID).Scan(&active)
	if !active {
		t.Fatal("target was deactivated")
	}
}

// Self-delete ends the caller's sessions: both access and refresh tokens die.
func TestSelfDeleteRevokesSessions(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "me@x.test", true, "my-password1")
	p := e.loginPair("me@x.test", "my-password1")

	if code, body := e.do("DELETE", fmt.Sprintf("/api/users/%d", uid), map[string]string{"password": "my-password1"}, p.AccessToken); code != 204 {
		t.Fatalf("self delete: %d %v", code, body)
	}
	if code, _ := e.do("GET", "/api/ping", nil, p.AccessToken); code != 401 {
		t.Fatalf("access token survived self-delete: %d", code)
	}
	if _, _, err := e.auth.Refresh(p.RefreshToken, "ua", "ip"); err == nil {
		t.Fatal("refresh token survived self-delete")
	}
}

// The only active admin cannot remove itself; the response is 409 and nothing changes.
func TestLastAdminCannotSelfDelete(t *testing.T) {
	e := newEnv(t)
	uid := e.seedAdmin("admin@x.test", "admin-password1")
	p := e.loginPair("admin@x.test", "admin-password1")

	if code, _ := e.do("DELETE", fmt.Sprintf("/api/users/%d", uid), map[string]string{"password": "admin-password1"}, p.AccessToken); code != 409 {
		t.Fatalf("last admin self-delete: %d, want 409", code)
	}
	if code, _ := e.do("GET", "/api/ping", nil, p.AccessToken); code != 200 {
		t.Fatalf("a refused delete must leave the session alone: %d", code)
	}
}

// Admin deactivation (and deletion) kills the target's access and refresh tokens.
func TestAdminDeactivateAndDeleteRevokeTargetSessions(t *testing.T) {
	e := newEnv(t)
	e.seedAdmin("admin@x.test", "admin-password1")
	admin := e.loginPair("admin@x.test", "admin-password1")

	for _, deleteIt := range []bool{false, true} {
		email := fmt.Sprintf("t%v@x.test", deleteIt)
		uid := seedUser(t, e.db, email, true, "target-password1")
		p := e.loginPair(email, "target-password1")
		if code, _ := e.do("GET", "/api/ping", nil, p.AccessToken); code != 200 {
			t.Fatalf("precondition: %d", code)
		}

		var code int
		if deleteIt {
			code, _ = e.do("DELETE", fmt.Sprintf("/api/admin/users/%d", uid), map[string]string{}, admin.AccessToken)
		} else {
			code, _ = e.do("PATCH", fmt.Sprintf("/api/admin/users/%d/status", uid), map[string]any{"is_active": false}, admin.AccessToken)
		}
		if code != 200 && code != 204 {
			t.Fatalf("admin action (delete=%v): %d", deleteIt, code)
		}
		if code, _ := e.do("GET", "/api/ping", nil, p.AccessToken); code != 401 {
			t.Fatalf("access token survived (delete=%v): %d", deleteIt, code)
		}
		if _, _, err := e.auth.Refresh(p.RefreshToken, "ua", "ip"); err == nil {
			t.Fatalf("refresh token survived (delete=%v)", deleteIt)
		}
	}
}

// Resend answers identically for unknown, unverified and verified addresses; verified gets no mail.
func TestResendVerificationIsUniform(t *testing.T) {
	e := newEnv(t)
	seedUser(t, e.db, "unverified@x.test", false, "pw-unverified1")
	seedUser(t, e.db, "verified@x.test", true, "pw-verified1")

	var first []byte
	for _, email := range []string{"nobody@x.test", "unverified@x.test", "verified@x.test"} {
		code, body := e.do("POST", "/api/auth/resend-verification", map[string]string{"email": email}, "")
		raw, _ := json.Marshal(body)
		if code != 200 {
			t.Fatalf("%s: %d %s", email, code, raw)
		}
		if first == nil {
			first = raw
		} else if string(raw) != string(first) {
			t.Fatalf("%s response differs: %s vs %s", email, raw, first)
		}
	}
	e.mail.last(t, "unverified@x.test", reVerifyToken)
	if _, ok := e.mail.find("verified@x.test", reVerifyToken); ok {
		t.Fatal("verified account received a verification mail")
	}
}

// Two admins removing each other (or themselves) at the same moment must never leave zero
// active admins: exactly one removal wins, the other gets the final-admin error.
func TestConcurrentLastAdminRemovalKeepsOneAdmin(t *testing.T) {
	e := newEnv(t)
	repo := admin_module.NewRepository(e.db)
	isFinalAdmin := func(err error) bool {
		return errors.Is(err, user_module.ErrFinalAdmin) || errors.Is(err, admin_module.ErrFinalAdminProtection)
	}
	type op func() error
	adminPw := user_module.DeleteAccountRequest{Password: "admin-password1"}
	scenarios := map[string]func(a, b uint) (op, op){
		"both self-delete": func(a, b uint) (op, op) {
			return func() error { return e.users.DeleteUser(a, a, 0, adminPw) }, func() error { return e.users.DeleteUser(b, b, 0, adminPw) }
		},
		"deactivate each other": func(a, b uint) (op, op) {
			return func() error { return repo.UpdateUserStatus(a, b, false, nil) }, func() error { return repo.UpdateUserStatus(b, a, false, nil) }
		},
		"self-delete vs deactivate the other": func(a, b uint) (op, op) {
			return func() error { return e.users.DeleteUser(a, a, 0, adminPw) }, func() error { return repo.UpdateUserStatus(b, a, false, nil) }
		},
	}
	for name, build := range scenarios {
		for i := 0; i < 5; i++ {
			e.db.Exec(`UPDATE users SET is_active = false WHERE role_id = (SELECT id FROM roles WHERE role_name = 'admin')`)
			a := e.seedAdmin(fmt.Sprintf("a%d-%s@x.test", i, randomSuffix()), "admin-password1")
			b := e.seedAdmin(fmt.Sprintf("b%d-%s@x.test", i, randomSuffix()), "admin-password1")
			opA, opB := build(a, b)

			errs := make([]error, 2)
			var wg sync.WaitGroup
			start := make(chan struct{})
			for idx, f := range []op{opA, opB} {
				wg.Add(1)
				go func() { defer wg.Done(); <-start; errs[idx] = f() }()
			}
			close(start)
			wg.Wait()

			// "self-delete vs deactivate the other" targets the same account twice, so there the
			// loser may also fail on the already-inactive account instead of the admin guard.
			failures := 0
			for _, err := range errs {
				if err != nil {
					failures++
					if !isFinalAdmin(err) && name != "self-delete vs deactivate the other" {
						t.Fatalf("%s #%d: unexpected error %v", name, i, err)
					}
				}
			}
			if name != "self-delete vs deactivate the other" && failures != 1 {
				t.Fatalf("%s #%d: %d failures (%v), want exactly 1", name, i, failures, errs)
			}
			if n := e.count(`SELECT count(*) FROM users WHERE is_active = true AND role_id = (SELECT id FROM roles WHERE role_name = 'admin')`); n < 1 {
				t.Fatalf("%s #%d: no active admin left", name, i)
			}
		}
	}
}

// Accounts deactivated before sessions were revoked on deactivation still hold tokens;
// reactivating through the admin path must end them.
func TestAdminReactivationRevokesOldSessions(t *testing.T) {
	e := newEnv(t)
	e.seedAdmin("admin@x.test", "admin-password1")
	admin := e.loginPair("admin@x.test", "admin-password1")
	uid := seedUser(t, e.db, "old@x.test", true, "old-password1")
	p := e.loginPair("old@x.test", "old-password1")
	e.db.Exec(`UPDATE users SET is_active = false WHERE id = ?`, uid) // old-style deactivation: no epoch bump

	if code, body := e.do("PATCH", fmt.Sprintf("/api/admin/users/%d/status", uid), map[string]any{"is_active": true}, admin.AccessToken); code != 200 {
		t.Fatalf("reactivate: %d %v", code, body)
	}
	if code, _ := e.do("GET", "/api/ping", nil, p.AccessToken); code != 401 {
		t.Fatalf("old access token survived reactivation: %d", code)
	}
	if _, _, err := e.auth.Refresh(p.RefreshToken, "ua", "ip"); err == nil {
		t.Fatal("old refresh token survived reactivation")
	}
}

// A stored hash is not a token: submitting the hash itself must not verify the account,
// and a raw value stored by an old instance is no longer accepted.
func TestVerifyRejectsStoredValueAsToken(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "raw@x.test", false, "pw-rawraw1")
	stored := auth_module.HashToken("mailed-token")
	e.db.Exec(`INSERT INTO email_verifications (user_id, token, email, expires_at, created_at) VALUES (?, ?, 'raw@x.test', now() + interval '1 hour', now())`, uid, stored)
	if _, err := e.auth.VerifyEmail(stored, ptr("pw-rawraw1"), false); err == nil {
		t.Fatal("stored hash accepted as a token")
	}
	if _, _, verified, _ := e.userRow(uid); verified {
		t.Fatal("verified with the stored hash")
	}
	if _, err := e.auth.VerifyEmail("mailed-token", ptr("pw-rawraw1"), false); err != nil {
		t.Fatalf("mailed token: %v", err)
	}
}

// A wrong or missing password changes nothing; the right one deactivates.
func TestSelfDeleteRequiresPassword(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "pw@x.test", true, "my-password1")
	p := e.loginPair("pw@x.test", "my-password1")
	path := fmt.Sprintf("/api/users/%d", uid)

	for _, body := range []any{nil, map[string]string{"password": "wrong-password1"}, map[string]string{"confirm_username": "ignored"}} {
		code, out := e.do("DELETE", path, body, p.AccessToken)
		if code != 400 || out["code"] != "WRONG_PASSWORD" {
			t.Fatalf("body %v: %d %v, want 400 WRONG_PASSWORD", body, code, out)
		}
	}
	if code, _ := e.do("GET", "/api/ping", nil, p.AccessToken); code != 200 {
		t.Fatalf("a refused delete must keep the session: %d", code)
	}
	if n := e.count(`SELECT count(*) FROM users WHERE id = ? AND is_active = true`, uid); n != 1 {
		t.Fatal("account deactivated by a refused delete")
	}
	if code, out := e.do("DELETE", path, map[string]string{"password": "my-password1"}, p.AccessToken); code != 204 {
		t.Fatalf("correct password: %d %v", code, out)
	}
}

// Google-only accounts (no password) confirm by typing their username.
func TestSelfDeleteGoogleOnlyConfirmsUsername(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "g@x.test", true, "")
	e.seedGoogleIdentity(uid, "sub-g", "g@x.test")
	e.google.set(auth_module.GoogleIdentityClaims{Subject: "sub-g", Email: "g@x.test", EmailVerified: true})
	var username string
	e.db.Raw(`SELECT username FROM users WHERE id = ?`, uid).Scan(&username)
	p, _, err := e.googleLogin()
	if err != nil {
		t.Fatal(err)
	}
	path := fmt.Sprintf("/api/users/%d", uid)

	if code, out := e.do("DELETE", path, map[string]string{"confirm_username": "someone-else"}, p.AccessToken); code != 400 || out["code"] != "CONFIRM_MISMATCH" {
		t.Fatalf("mismatch: %d %v", code, out)
	}
	if code, out := e.do("DELETE", path, map[string]string{"confirm_username": username}, p.AccessToken); code != 204 {
		t.Fatalf("matching username: %d %v", code, out)
	}
}

// Failed confirmations are rate limited per user; success is not counted.
func TestSelfDeleteFailuresAreRateLimited(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "rl@x.test", true, "my-password1")
	p := e.loginPair("rl@x.test", "my-password1")
	path := fmt.Sprintf("/api/users/%d", uid)
	var last int
	for i := 0; i < 7; i++ {
		last, _ = e.do("DELETE", path, map[string]string{"password": "wrong-password1"}, p.AccessToken)
	}
	if last != 429 {
		t.Fatalf("after repeated failures: %d, want 429", last)
	}
}

// Saving or skipping favorite genres marks onboarding done once; only the owner sees it.
func TestSkipGenresSetsOnboardedAtOnce(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "ob@x.test", true, "my-password1")
	other := seedUser(t, e.db, "ob2@x.test", true, "my-password1")
	p := e.loginPair("ob@x.test", "my-password1")
	path := fmt.Sprintf("/api/users/%d", uid)

	_, out := e.do("GET", path, nil, p.AccessToken)
	if out["onboarded_at"] != nil {
		t.Fatalf("fresh account already onboarded: %v", out["onboarded_at"])
	}
	code, out := e.do("PATCH", path+"/genres", map[string]any{"favorite_genres": []int{}}, p.AccessToken)
	if code != 200 {
		t.Fatalf("skip: %d %v", code, out)
	}
	user, _ := out["user"].(map[string]any)
	first := user["onboarded_at"]
	if first == nil {
		t.Fatal("skip did not set onboarded_at")
	}
	_, out = e.do("PATCH", path+"/genres", map[string]any{"favorite_genres": []int{28}}, p.AccessToken)
	if user, _ := out["user"].(map[string]any); user["onboarded_at"] != first {
		t.Fatalf("onboarded_at moved: %v -> %v", first, user["onboarded_at"])
	}
	code, out = e.do("POST", "/api/auth/login", map[string]string{"identifier": "ob@x.test", "password": "my-password1"}, "")
	if u, _ := out["user"].(map[string]any); code != 200 || u["onboarded_at"] != first {
		t.Fatalf("login response: %d %v", code, out)
	}
	if _, out = e.do("GET", fmt.Sprintf("/api/users/%d", uid), nil, e.loginPair("ob2@x.test", "my-password1").AccessToken); out["onboarded_at"] != nil {
		t.Fatalf("other users must not see onboarded_at: %v", out["onboarded_at"])
	}
	_ = other
}

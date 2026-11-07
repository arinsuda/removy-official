//go:build integration

package integration

import (
	"testing"

	"github.com/arinsuda/movie-hub/internal/user_module"
)

// A password change ends every in-flight proof and session artifact the same way a
// verified reset does, and keeps identities and links.
func TestPasswordChangeRevokesSameScopeAsReset(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "owner@x.test", true, "old-password1")
	e.seedGoogleIdentity(uid, "owner-sub", "owner@x.test")
	e.db.Exec(`INSERT INTO account_links (local_user_id, remote_platform, remote_user_id, linked_via) VALUES (?, 'planzi', 'p-1', 'auto')`, uid)
	e.seedRefreshToken(uid, "old-session")
	e.mintHandoffRow(uid, "handoff-raw")
	if err := e.auth.ForgotPassword("owner@x.test"); err != nil {
		t.Fatal(err)
	}
	e.db.Exec(`INSERT INTO email_verifications (user_id, token, email, expires_at, created_at) VALUES (?, 'tok', 'owner@x.test', now() + interval '1 hour', now())`, uid)

	if err := e.users.ChangePassword(uid, uid, 0, user_module.ChangePasswordRequest{
		OldPassword: "old-password1", NewPassword: "new-password1", ConfirmPassword: "new-password1"}); err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{
		`SELECT count(*) FROM refresh_tokens WHERE user_id = ?`,
		`SELECT count(*) FROM link_handoff_tokens WHERE user_id = ?`,
		`SELECT count(*) FROM password_reset_tokens WHERE user_id = ?`,
		`SELECT count(*) FROM email_verifications WHERE user_id = ?`,
	} {
		if n := e.count(q, uid); n != 0 {
			t.Errorf("%s left %d rows", q, n)
		}
	}
	if e.count(`SELECT count(*) FROM user_identities WHERE user_id = ?`, uid) != 1 || e.count(`SELECT count(*) FROM account_links WHERE local_user_id = ?`, uid) != 1 {
		t.Fatal("identity or link lost on a password change")
	}
}

// If ending the sessions fails, the endpoint must say so and keep the cookies.
func TestLogoutAllReports500WhenRevokeFails(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "owner@x.test", true, "old-password1")
	tok := e.accessToken(uid, 0)
	if err := e.db.Exec(`CREATE FUNCTION block_user_update() RETURNS trigger AS $$ BEGIN RAISE EXCEPTION 'blocked'; END $$ LANGUAGE plpgsql;
		CREATE TRIGGER block_user_update BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION block_user_update()`).Error; err != nil {
		t.Fatal(err)
	}
	if code, body := e.do("POST", "/api/auth/logout-all", nil, tok); code != 500 {
		t.Fatalf("want 500 when the revoke fails, got %d %v", code, body)
	}
}

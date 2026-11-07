//go:build integration

package integration

import "testing"

// Logout must revoke the refresh token sent in the body when no cookie arrives, and only that one.
func TestLogoutRevokesBodyRefreshToken(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "out@x.test", true, "right-password1")
	login := func() string {
		code, body := e.do("POST", "/api/auth/login", map[string]string{"identifier": "out@x.test", "password": "right-password1"}, "")
		rt, _ := body["refresh_token"].(string)
		if code != 200 || rt == "" {
			t.Fatalf("login: %d %v", code, body)
		}
		return rt
	}
	first, second := login(), login()
	count := func() int64 { return e.count(`SELECT count(*) FROM refresh_tokens WHERE user_id = ?`, uid) }
	if n := count(); n != 2 {
		t.Fatalf("want 2 sessions, got %d", n)
	}

	if code, body := e.do("POST", "/api/auth/logout", map[string]string{"refresh_token": first}, ""); code != 200 {
		t.Fatalf("logout: %d %v", code, body)
	}
	if n := count(); n != 1 {
		t.Fatalf("logout must revoke exactly the body token, %d sessions left", n)
	}
	if code, _ := e.do("POST", "/api/auth/refresh", map[string]string{"refresh_token": first}, ""); code != 401 {
		t.Fatalf("revoked token must not refresh, got %d", code)
	}
	if code, _ := e.do("POST", "/api/auth/refresh", map[string]string{"refresh_token": second}, ""); code != 200 {
		t.Fatalf("other session must survive, got %d", code)
	}
}

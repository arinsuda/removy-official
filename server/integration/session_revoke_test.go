//go:build integration

package integration

import (
	"fmt"
	"testing"

	"github.com/arinsuda/movie-hub/internal/auth_module"
)

func (e *env) loginPair(email, password string) *auth_module.TokenPair {
	e.t.Helper()
	pair, _, err := e.auth.Login(auth_module.LoginRequest{Identifier: email, Password: password}, "ua", "1.2.3.4")
	if err != nil {
		e.t.Fatalf("login: %v", err)
	}
	return pair
}

// A password change ends every other session but keeps the caller signed in.
func TestPasswordChangeRevokesOthersAndReissuesCurrentSession(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "me@x.test", true, "old-password1")
	current := e.loginPair("me@x.test", "old-password1")
	other := e.loginPair("me@x.test", "old-password1") // another device

	code, body := e.do("PATCH", fmt.Sprintf("/api/users/%d/password", uid),
		map[string]string{"old_password": "old-password1", "new_password": "new-password1", "confirm_password": "new-password1"}, current.AccessToken)
	if code != 200 {
		t.Fatalf("change: %d %v", code, body)
	}
	newAccess, _ := body["access_token"].(string)
	newRefresh, _ := body["refresh_token"].(string)
	if newAccess == "" || newRefresh == "" {
		t.Fatalf("response must carry a reissued session: %v", body)
	}
	if _, _, _, ver := e.userRow(uid); ver != 1 {
		t.Fatalf("token_version = %d, want 1", ver)
	}
	if code, _ := e.do("GET", "/api/ping", nil, current.AccessToken); code != 401 {
		t.Fatalf("old access token still works: %d", code)
	}
	if code, _ := e.do("GET", "/api/ping", nil, newAccess); code != 200 {
		t.Fatalf("reissued access token rejected: %d", code)
	}
	for name, rt := range map[string]string{"current": current.RefreshToken, "other device": other.RefreshToken} {
		if _, _, err := e.auth.Refresh(rt, "ua", "ip"); err == nil {
			t.Fatalf("old refresh token of %s still works", name)
		}
	}
	if _, _, err := e.auth.Refresh(newRefresh, "ua", "ip"); err != nil {
		t.Fatalf("reissued refresh token: %v", err)
	}
}

// Logout-all must also kill access tokens, not only refresh tokens.
func TestLogoutAllBumpsEpochAndKillsAccessTokens(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "me@x.test", true, "old-password1")
	p := e.loginPair("me@x.test", "old-password1")
	if code, _ := e.do("POST", "/api/auth/logout-all", nil, p.AccessToken); code != 200 {
		t.Fatalf("logout-all: %d", code)
	}
	if _, _, _, ver := e.userRow(uid); ver != 1 {
		t.Fatalf("token_version = %d, want 1", ver)
	}
	if code, _ := e.do("GET", "/api/ping", nil, p.AccessToken); code != 401 {
		t.Fatalf("access token survived logout-all: %d", code)
	}
	if _, _, err := e.auth.Refresh(p.RefreshToken, "ua", "ip"); err == nil {
		t.Fatal("refresh token survived logout-all")
	}
}

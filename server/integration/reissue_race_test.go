//go:build integration

package integration

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/arinsuda/movie-hub/internal/identity_policy"
	"github.com/arinsuda/movie-hub/internal/user_module"
	"github.com/gofiber/fiber/v3"
)

// reissue calls ReissueSession from inside a request, as the password-change handler does.
func (e *env) reissue(uid uint, expectedVersion int) (access string, err error) {
	e.t.Helper()
	app := fiber.New()
	app.Get("/r", func(c fiber.Ctx) error {
		access, _, err = e.auth.ReissueSession(c, uid, expectedVersion)
		return c.SendStatus(200)
	})
	if _, terr := app.Test(httptest.NewRequest("GET", "/r", nil)); terr != nil {
		e.t.Fatal(terr)
	}
	return access, err
}

// A reset that commits between the password change and the reissue must win: the
// change's session is not handed out, and no refresh token is stored.
func TestReissueRefusedWhenResetCommitsAfterPasswordChange(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "owner@x.test", true, "old-password1")

	if err := e.users.ChangePassword(uid, uid, 0, user_module.ChangePasswordRequest{
		OldPassword: "old-password1", NewPassword: "new-password1", ConfirmPassword: "new-password1"}); err != nil {
		t.Fatal(err)
	}
	e.claimByReset(uid, "owner@x.test") // epoch 1 -> 2 before the handler reissues

	access, err := e.reissue(uid, 1)
	if !errors.Is(err, identity_policy.ErrSessionStale) || access != "" {
		t.Fatalf("want ErrSessionStale and no token, got %q %v", access, err)
	}
	if n := e.count(`SELECT count(*) FROM refresh_tokens WHERE user_id = ?`, uid); n != 0 {
		t.Fatalf("a refresh token was stored despite the stale epoch: %d", n)
	}
}

func TestReissueIssuesWhenEpochMatches(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "owner@x.test", true, "old-password1")
	if err := e.users.ChangePassword(uid, uid, 0, user_module.ChangePasswordRequest{
		OldPassword: "old-password1", NewPassword: "new-password1", ConfirmPassword: "new-password1"}); err != nil {
		t.Fatal(err)
	}
	access, err := e.reissue(uid, 1)
	if err != nil || access == "" {
		t.Fatalf("matching epoch must issue: %q %v", access, err)
	}
	if code, _ := e.do("GET", "/api/ping", nil, access); code != 200 {
		t.Fatalf("reissued token rejected: %d", code)
	}
}

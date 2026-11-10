//go:build integration

package integration

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/database"
	"github.com/arinsuda/movie-hub/internal/admin_module"
	"github.com/arinsuda/movie-hub/internal/auth_module"
)

const releaseMigration = "20261111_release_removed_user_links"

func (e *env) seedLinkRows(uid uint, remoteID string) {
	e.t.Helper()
	if err := e.db.Exec(`INSERT INTO account_links (local_user_id, remote_platform, remote_user_id, linked_via, linked_at) VALUES (?, 'planzi', ?, 'password', now())`, uid, remoteID).Error; err != nil {
		e.t.Fatal(err)
	}
	if err := e.db.Exec(`INSERT INTO link_handoff_tokens (token_hash, user_id, target_platform, expires_at, created_at) VALUES (?, ?, 'planzi', now() + interval '1 hour', now())`, fmt.Sprintf("h%d", uid), uid).Error; err != nil {
		e.t.Fatal(err)
	}
}

func (e *env) releasedRows(uid uint) int64 {
	return e.count(`SELECT (SELECT count(*) FROM user_identities WHERE user_id = ?)
		+ (SELECT count(*) FROM account_links WHERE local_user_id = ?)
		+ (SELECT count(*) FROM link_handoff_tokens WHERE user_id = ?)`, uid, uid, uid)
}

// Admin delete frees the Google subject, the PLANZI link and the email for a fresh signup.
func TestAdminDeleteReleasesIdentitiesAndLinks(t *testing.T) {
	e := newEnv(t)
	adminID := e.seedAdmin("admin@x.test", "admin-password1")
	uid := seedUser(t, e.db, "gone@gmail.com", true, "")
	e.seedGoogleIdentity(uid, "gone-sub", "gone@gmail.com")
	e.seedLinkRows(uid, "planzi-1")

	repo := admin_module.NewRepository(e.db)
	if err := repo.DeleteUser(adminID, uid, nil); err != nil {
		t.Fatal(err)
	}
	if n := e.releasedRows(uid); n != 0 {
		t.Fatalf("%d identity/link rows left", n)
	}

	e.google.set(auth_module.GoogleIdentityClaims{Subject: "gone-sub", Email: "gone@gmail.com", EmailVerified: true})
	_, nu, err := e.googleLogin()
	if err != nil || nu.ID == uid {
		t.Fatalf("google re-signup: user=%v err=%v", nu, err)
	}
	if n := e.count(`SELECT count(*) FROM user_identities WHERE user_id = ? AND provider_subject = 'gone-sub'`, nu.ID); n != 1 {
		t.Fatalf("new identity rows: %d", n)
	}
}

// Defensive path: a stale identity of a soft-deleted user does not block a new Google signup.
func TestGoogleLoginDropsStaleIdentityOfSoftDeletedUser(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "stale@gmail.com", true, "")
	e.seedGoogleIdentity(uid, "stale-sub", "stale@gmail.com")
	if err := e.db.Exec(`UPDATE users SET deleted_at = now(), is_active = false WHERE id = ?`, uid).Error; err != nil {
		t.Fatal(err)
	}
	e.google.set(auth_module.GoogleIdentityClaims{Subject: "stale-sub", Email: "stale@gmail.com", EmailVerified: true})
	_, nu, err := e.googleLogin()
	if err != nil || nu.ID == uid {
		t.Fatalf("user=%v err=%v", nu, err)
	}
	if n := e.count(`SELECT count(*) FROM user_identities WHERE user_id = ?`, uid); n != 0 {
		t.Fatalf("stale identity kept: %d", n)
	}
}

// Rule 8: self-deactivation keeps the identity and the email reserved.
func TestSelfDeactivatedUserStaysReserved(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "quiet@x.test", true, "")
	e.seedGoogleIdentity(uid, "quiet-sub", "quiet@x.test")
	if err := e.db.Exec(`UPDATE users SET is_active = false WHERE id = ?`, uid).Error; err != nil {
		t.Fatal(err)
	}
	e.google.set(auth_module.GoogleIdentityClaims{Subject: "quiet-sub", Email: "quiet@x.test", EmailVerified: true})
	if _, _, err := e.googleLogin(); !errors.Is(err, auth_module.ErrUserInactive) {
		t.Fatalf("want ErrUserInactive, got %v", err)
	}
	if n := e.count(`SELECT count(*) FROM user_identities WHERE user_id = ?`, uid); n != 1 {
		t.Fatalf("identity removed: %d", n)
	}
	if _, err := e.auth.Register(auth_module.RegisterRequest{Username: "squat", Email: "quiet@x.test", Password: "squat-pass1234", ConfirmPassword: "squat-pass1234"}); err == nil {
		t.Fatal("email of a deactivated user was reusable")
	}
}

// The backfill removes stale rows of already soft-deleted users only.
func TestBackfillReleasesRemovedUserLinks(t *testing.T) {
	e := newEnv(t)
	gone := seedUser(t, e.db, "old-gone@gmail.com", true, "")
	live := seedUser(t, e.db, "live@x.test", true, "")
	for i, uid := range []uint{gone, live} {
		e.seedGoogleIdentity(uid, fmt.Sprintf("sub-%d", i), "x@x.test")
		e.seedLinkRows(uid, fmt.Sprintf("planzi-%d", i))
	}
	if err := e.db.Exec(`UPDATE users SET deleted_at = now(), is_active = false WHERE id = ?`, gone).Error; err != nil {
		t.Fatal(err)
	}
	if err := e.db.Exec(`DELETE FROM migration_history WHERE version = ?`, releaseMigration).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(e.db, testConfig()); err != nil {
		t.Fatal(err)
	}
	if n := e.releasedRows(gone); n != 0 {
		t.Fatalf("removed user still has %d rows", n)
	}
	if n := e.releasedRows(live); n != 3 {
		t.Fatalf("live user rows = %d, want 3", n)
	}
}

// A failing step of the new delete loops surfaces as an error and rolls the delete back.
// Each table renamed here is touched only by DeleteUser's new code (RevokeSessionScope
// does not reach user_identities, account_links, email_change_requests or library_items).
func TestAdminDeleteSurfacesDBErrors(t *testing.T) {
	for _, table := range []string{"user_identities", "account_links", "email_change_requests", "library_items"} {
		t.Run(table, func(t *testing.T) {
			e := newEnv(t)
			adminID := e.seedAdmin("admin@x.test", "admin-password1")
			uid := seedUser(t, e.db, "keep@x.test", true, "")
			if err := e.db.Exec(`ALTER TABLE ` + table + ` RENAME TO renamed_x`).Error; err != nil {
				t.Fatal(err)
			}
			if err := admin_module.NewRepository(e.db).DeleteUser(adminID, uid, nil); err == nil {
				t.Fatal("expected error")
			}
			if n := e.count(`SELECT count(*) FROM users WHERE id = ? AND deleted_at IS NULL`, uid); n != 1 {
				t.Fatal("delete was not rolled back")
			}
		})
	}
}

// IntentLink: a Google identity left on a soft-deleted user does not block linking it to a live one.
func TestGoogleLinkDropsStaleIdentityOfSoftDeletedUser(t *testing.T) {
	e := newEnv(t)
	gone := seedUser(t, e.db, "gone2@x.test", true, "")
	e.seedGoogleIdentity(gone, "shared-sub", "gone2@gmail.com")
	if err := e.db.Exec(`UPDATE users SET deleted_at = now(), is_active = false WHERE id = ?`, gone).Error; err != nil {
		t.Fatal(err)
	}
	live := seedUser(t, e.db, "live2@gmail.com", true, "live-password1")
	ctx := context.Background()
	_, state, err := e.auth.InitiateGoogleLink(ctx, live, 0, "/")
	if err != nil {
		t.Fatal(err)
	}
	e.google.set(auth_module.GoogleIdentityClaims{Subject: "shared-sub", Email: "live2@gmail.com", EmailVerified: true})
	if _, _, _, err := e.auth.HandleGoogleCallback(ctx, "code", state, state, "ua", "ip"); err != nil {
		t.Fatalf("link: %v", err)
	}
	if n := e.count(`SELECT count(*) FROM user_identities WHERE user_id = ? AND provider_subject = 'shared-sub'`, live); n != 1 {
		t.Fatalf("live user identity rows: %d", n)
	}
	if n := e.count(`SELECT count(*) FROM user_identities WHERE user_id = ?`, gone); n != 0 {
		t.Fatal("stale identity kept")
	}
}

// Concurrent first-time Google logins for one subject: one account, and the loser gets a
// mapped conflict error (or the winner's account), never a raw database error.
func TestConcurrentGoogleSignupMapsUniqueViolation(t *testing.T) {
	e := newEnv(t)
	e.google.set(auth_module.GoogleIdentityClaims{Subject: "race-sub", Email: "race@gmail.com", EmailVerified: true})
	var wg sync.WaitGroup
	errs := make([]error, 6)
	for i := range errs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, errs[i] = e.googleLogin()
		}()
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil && !errors.Is(err, auth_module.ErrGoogleIdentityAlreadyConnected) {
			t.Fatalf("unmapped error: %v", err)
		}
	}
	if n := e.count(`SELECT count(*) FROM user_identities WHERE provider_subject = 'race-sub'`); n != 1 {
		t.Fatalf("identities = %d", n)
	}
}

// Prod case: the same email and username register again after an admin delete, and the
// new account verifies with its own token while the old user's tokens are gone.
func TestRegisterSameEmailAndUsernameAfterAdminDelete(t *testing.T) {
	e := newEnv(t)
	adminID := e.seedAdmin("admin@x.test", "admin-password1")
	const email, username = "returning@x.test", "returning_user"
	old, err := e.auth.Register(auth_module.RegisterRequest{Username: username, Email: email, Password: "first-pass1234", ConfirmPassword: "first-pass1234"})
	if err != nil {
		t.Fatal(err)
	}
	oldTok := e.mail.last(t, email, reVerifyToken) // Register mails asynchronously; wait until the old token exists
	if e.count(`SELECT count(*) FROM email_verifications WHERE user_id = ?`, old.ID) == 0 {
		t.Fatal("precondition: old user has a verification token")
	}
	if err := admin_module.NewRepository(e.db).DeleteUser(adminID, old.ID, nil); err != nil {
		t.Fatal(err)
	}
	if n := e.count(`SELECT count(*) FROM email_verifications WHERE user_id = ?`, old.ID); n != 0 {
		t.Fatalf("old verification tokens left: %d", n)
	}

	nu, err := e.auth.Register(auth_module.RegisterRequest{Username: username, Email: email, Password: "second-pass1234", ConfirmPassword: "second-pass1234"})
	if err != nil || nu.ID == old.ID {
		t.Fatalf("re-register: user=%v err=%v", nu, err)
	}
	var newTok string
	for deadline := time.Now().Add(3 * time.Second); time.Now().Before(deadline); time.Sleep(50 * time.Millisecond) {
		if tok := e.mail.last(t, email, reVerifyToken); tok != oldTok {
			newTok = tok
			break
		}
	}
	if newTok == "" {
		t.Fatal("no new verification mail for the re-registered account")
	}
	if n := e.count(`SELECT count(*) FROM email_verifications WHERE user_id = ?`, nu.ID); n != 1 {
		t.Fatalf("new user tokens: %d", n)
	}
	if err := e.verify(oldTok, ptr("second-pass1234")); err == nil {
		t.Fatal("old user's token still verifies")
	}
	if err := e.verify(newTok, ptr("second-pass1234")); err != nil {
		t.Fatalf("verify new account: %v", err)
	}
	if _, _, verified, _ := e.userRow(nu.ID); !verified {
		t.Fatal("new account not verified")
	}
}

//go:build integration

package integration

import (
	"testing"

	"github.com/arinsuda/movie-hub/database"
	"github.com/arinsuda/movie-hub/internal/auth_module"
	"github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/internal/user_module"
	"gorm.io/gorm"
)

var securityColumns = map[string][]string{
	"users":                 {"token_version", "identity_reset_at"},
	"email_verifications":   {"email"},
	"password_reset_tokens": {"email"},
	"link_pending":          {"code_email"},
	"email_change_requests": {"new_otp_hash"},
	"notifications":         {"params"},
}

func TestSecurityColumnsMigrated(t *testing.T) {
	db := newTestDB(t)
	for table, cols := range securityColumns {
		for _, col := range cols {
			if !db.Migrator().HasColumn(table, col) {
				t.Errorf("%s.%s missing after Migrate", table, col)
			}
		}
	}
	var n int64
	db.Raw(`SELECT count(*) FROM migration_history WHERE version = '20261107_account_security_columns'`).Scan(&n)
	if n != 1 {
		t.Fatalf("migration not recorded exactly once: %d", n)
	}
}

// AutoMigrate must not add the columns itself: it would do so without
// lock_timeout, and before the guarded migration runs.
func TestAutoMigrateDoesNotAddSecurityColumns(t *testing.T) {
	db := newBareDB(t)
	if err := db.AutoMigrate(&user_module.Role{}, &user_module.User{}, &user_module.EmailVerification{},
		&user_module.PasswordResetToken{}, &user_module.EmailChangeRequest{}, &notification_module.Notification{}); err != nil {
		t.Fatal(err)
	}
	for table, cols := range securityColumns {
		for _, col := range cols {
			if table == "link_pending" {
				continue // raw-SQL table, never AutoMigrated
			}
			if db.Migrator().HasColumn(table, col) {
				t.Errorf("AutoMigrate added %s.%s", table, col)
			}
		}
	}
}

// legacyVersions are the migrations that add the new columns/indexes; forgetting them and
// dropping the columns turns a freshly migrated database back into the pre-release one.
var legacyVersions = []string{"20261107_account_security_columns", "20261108_email_verification_attempts", "20261109_email_verification_token_hash", "20261110_notification_params"}

// makeLegacy restores the schema the old binary left behind: the new columns are gone and
// their migrations are unrecorded, so the next Migrate really runs them.
func makeLegacy(t *testing.T, db *gorm.DB) {
	t.Helper()
	for _, q := range []string{
		`ALTER TABLE users DROP COLUMN token_version, DROP COLUMN identity_reset_at`,
		`ALTER TABLE email_verifications DROP COLUMN email, DROP COLUMN attempts`,
		`ALTER TABLE password_reset_tokens DROP COLUMN email`,
		`ALTER TABLE link_pending DROP COLUMN code_email`,
		`ALTER TABLE email_change_requests DROP COLUMN new_otp_hash`,
		`ALTER TABLE notifications DROP COLUMN params`,
	} {
		if err := db.Exec(q).Error; err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	if err := db.Exec(`DELETE FROM migration_history WHERE version IN ?`, legacyVersions).Error; err != nil {
		t.Fatal(err)
	}
}

// Rows written by the old binary (no new columns) are seeded BEFORE the new migrations run,
// then stay readable and the defaults hold; the new model reads NULL email columns without error.
func TestLegacyRowsSurviveAdditiveMigration(t *testing.T) {
	db := newTestDB(t)
	makeLegacy(t, db)

	// legacy rows, inserted with the old column set only
	var uid uint
	if err := db.Raw(`INSERT INTO users (created_at, updated_at, username, email, password, role_id, is_active)
		VALUES (now(), now(), 'olduser', 'old@x.test', 'legacy-hash', 2, true) RETURNING id`).Scan(&uid).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO email_verifications (user_id, token, expires_at, created_at) VALUES (?, 'legacy-token', now() + interval '1 hour', now())`, uid).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO password_reset_tokens (user_id, hashed_token, expires_at, created_at) VALUES (?, 'legacy-hash', now() + interval '1 hour', now())`, uid).Error; err != nil {
		t.Fatal(err)
	}

	if err := db.Exec(`INSERT INTO notifications (created_at, updated_at, user_id, type, category, message, is_read)
		VALUES (now(), now(), ?, 'welcome', 'system', 'legacy text', false)`, uid).Error; err != nil {
		t.Fatal(err)
	}

	if err := database.Migrate(db, testConfig()); err != nil {
		t.Fatalf("migrate over legacy rows: %v", err)
	}

	if !db.Migrator().HasColumn("notifications", "params") {
		t.Fatal("notifications.params missing after migrate")
	}
	var legacyNotif notification_module.Notification
	if err := db.Where("user_id = ?", uid).First(&legacyNotif).Error; err != nil || legacyNotif.Params != nil || legacyNotif.Message != "legacy text" {
		t.Fatalf("legacy notification: err=%v params=%v msg=%q", err, legacyNotif.Params, legacyNotif.Message)
	}

	var u user_module.User
	if err := db.First(&u, uid).Error; err != nil {
		t.Fatal(err)
	}
	if u.TokenVersion != 0 || u.IdentityResetAt != nil || u.Email != "old@x.test" {
		t.Fatalf("legacy row changed: version=%d reset=%v email=%s", u.TokenVersion, u.IdentityResetAt, u.Email)
	}
	var ev user_module.EmailVerification
	if err := db.Where("token = ?", auth_module.HashToken("legacy-token")).First(&ev).Error; err != nil || ev.Email != nil {
		t.Fatalf("legacy verification: err=%v email=%v", err, ev.Email)
	}
	var attempts int
	db.Raw(`SELECT attempts FROM email_verifications WHERE token = ?`, auth_module.HashToken("legacy-token")).Scan(&attempts)
	if attempts != 0 {
		t.Fatalf("legacy verification attempts = %d, want 0", attempts)
	}
	var legacyResets int64
	db.Raw(`SELECT count(*) FROM password_reset_tokens WHERE user_id = ? AND email IS NULL`, uid).Scan(&legacyResets)
	if legacyResets != 1 {
		t.Fatalf("legacy reset token rows = %d, want 1 with NULL email", legacyResets)
	}
	// creating through the model must never write the read-only epoch columns
	nu := user_module.User{Username: "newu", Email: "new@x.test", RoleID: 2, IsActive: true}
	if err := db.Create(&nu).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&user_module.User{}).Where("id = ?", nu.ID).Update("bio", "x").Error; err != nil {
		t.Fatal(err)
	}
	db.Exec(`UPDATE users SET token_version = 5 WHERE id = ?`, nu.ID)
	db.Model(&user_module.User{}).Where("id = ?", nu.ID).Updates(map[string]any{"bio": "y"})
	var v int
	db.Raw(`SELECT token_version FROM users WHERE id = ?`, nu.ID).Scan(&v)
	if v != 5 {
		t.Fatalf("model write clobbered token_version: %d", v)
	}
}

// Unexpired raw tokens mailed before the hash change must still verify after the migration;
// expired rows are dropped and nothing raw is left behind.
func TestVerificationTokenHashMigration(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "hash@x.test", false, "pw-hashhash1")
	other := seedUser(t, e.db, "old@x.test", false, "pw-oldoldold")
	if err := e.db.Exec(`DELETE FROM migration_history WHERE version = '20261109_email_verification_token_hash'`).Error; err != nil {
		t.Fatal(err)
	}
	e.db.Exec(`INSERT INTO email_verifications (user_id, token, email, expires_at, created_at) VALUES (?, 'raw-live-token', 'hash@x.test', now() + interval '1 hour', now())`, uid)
	e.db.Exec(`INSERT INTO email_verifications (user_id, token, email, expires_at, created_at) VALUES (?, 'raw-expired-token', 'old@x.test', now() - interval '1 hour', now())`, other)

	if err := database.Migrate(e.db, testConfig()); err != nil {
		t.Fatal(err)
	}
	if e.count(`SELECT count(*) FROM email_verifications WHERE token LIKE 'raw-%'`) != 0 {
		t.Fatal("raw tokens remain after migration")
	}
	if e.count(`SELECT count(*) FROM email_verifications WHERE user_id = ?`, other) != 0 {
		t.Fatal("expired row survived")
	}
	if _, err := e.auth.VerifyEmail("raw-live-token", ptr("pw-hashhash1"), false); err != nil {
		t.Fatalf("legacy raw link no longer verifies: %v", err)
	}
}

// Existing accounts count as onboarded after the backfill; accounts created afterwards stay NULL.
func TestBackfillOnboardedAt(t *testing.T) {
	db := newTestDB(t)
	old := seedUser(t, db, "old@x.test", true, "pw-oldold1")
	db.Exec(`UPDATE users SET onboarded_at = NULL, first_login_at = NULL WHERE id = ?`, old)
	if err := db.Exec(`DELETE FROM migration_history WHERE version = '20261109_backfill_onboarded_at'`).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db, testConfig()); err != nil {
		t.Fatal(err)
	}
	var n int64
	db.Raw(`SELECT count(*) FROM users WHERE id = ? AND onboarded_at IS NOT NULL`, old).Scan(&n)
	if n != 1 {
		t.Fatal("existing user not backfilled")
	}
	fresh := seedUser(t, db, "new@x.test", true, "pw-newnew1")
	if err := database.Migrate(db, testConfig()); err != nil {
		t.Fatal(err)
	}
	db.Raw(`SELECT count(*) FROM users WHERE id = ? AND onboarded_at IS NULL`, fresh).Scan(&n)
	if n != 1 {
		t.Fatal("a rerun of Migrate must not onboard later users")
	}
}

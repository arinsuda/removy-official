//go:build integration

package integration

import (
	"testing"

	"github.com/arinsuda/movie-hub/internal/user_module"
)

var securityColumns = map[string][]string{
	"users":                 {"token_version", "identity_reset_at"},
	"email_verifications":   {"email"},
	"password_reset_tokens": {"email"},
	"link_pending":          {"code_email"},
	"email_change_requests": {"new_otp_hash"},
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
		&user_module.PasswordResetToken{}, &user_module.EmailChangeRequest{}); err != nil {
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

// Rows written by the old binary (no new columns) stay readable and the
// defaults hold; the new model reads NULL email columns without error.
func TestLegacyRowsSurviveAdditiveMigration(t *testing.T) {
	db := newTestDB(t)
	uid := seedUser(t, db, "old@x.test", false, "pw-oldoldold")
	if err := db.Exec(`INSERT INTO email_verifications (user_id, token, expires_at, created_at) VALUES (?, 'legacy-token', now() + interval '1 hour', now())`, uid).Error; err != nil {
		t.Fatal(err)
	}
	var u user_module.User
	if err := db.First(&u, uid).Error; err != nil {
		t.Fatal(err)
	}
	if u.TokenVersion != 0 || u.IdentityResetAt != nil {
		t.Fatalf("defaults wrong: %d %v", u.TokenVersion, u.IdentityResetAt)
	}
	var ev user_module.EmailVerification
	if err := db.Where("token = 'legacy-token'").First(&ev).Error; err != nil || ev.Email != nil {
		t.Fatalf("legacy verification: err=%v email=%v", err, ev.Email)
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

//go:build integration

package integration

import (
	"fmt"
	"strings"
	"testing"

	"github.com/arinsuda/movie-hub/database"
	"github.com/arinsuda/movie-hub/database/migrate"
	"github.com/arinsuda/movie-hub/internal/admin_module"
	"gorm.io/gorm"
)

const (
	fkNotValidMigration = "20261111_user_fk_cascade_not_valid"
	fkValidateMigration = "20261111_user_fk_cascade_validate"
)

type userFKRow struct {
	Tbl       string
	Col       string
	Conname   string
	Confdel   string
	Validated bool
}

// userFKRows lists every single-column FK that references users.
func userFKRows(t *testing.T, db *gorm.DB) []userFKRow {
	t.Helper()
	var rows []userFKRow
	err := db.Raw(`SELECT c.conrelid::regclass::text AS tbl, a.attname AS col, c.conname, c.confdeltype::text AS confdel, c.convalidated AS validated
		FROM pg_constraint c JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = c.conkey[1]
		WHERE c.contype = 'f' AND c.confrelid = 'users'::regclass AND array_length(c.conkey, 1) = 1
		ORDER BY 1, 2`).Scan(&rows).Error
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

// Every FK to users is validated and has the intended delete rule.
func TestUserForeignKeysValidatedWithExpectedRules(t *testing.T) {
	db := newTestDB(t)
	rows := userFKRows(t, db)
	if len(rows) == 0 {
		t.Fatal("no FK to users found")
	}
	for _, r := range rows {
		if !r.Validated {
			t.Errorf("%s.%s (%s) is NOT VALID", r.Tbl, r.Col, r.Conname)
		}
		want := "c"
		switch {
		case r.Tbl == "admin_audit_logs" && r.Col == "admin_id", r.Tbl == "activity_events" && r.Col == "target_user_id":
			want = "n"
		}
		if r.Confdel != want {
			t.Errorf("%s.%s delete rule %q, want %q", r.Tbl, r.Col, r.Confdel, want)
		}
	}
	for _, fk := range migrate.UserFKs {
		n := 0
		for _, r := range rows {
			if r.Tbl == fk.Table && r.Col == fk.Column {
				n++
			}
		}
		if n != 1 {
			t.Errorf("%s.%s has %d FKs to users, want exactly 1 (no duplicates)", fk.Table, fk.Column, n)
		}
	}
}

// Orphans in tables that never had an FK (prod state) are cleaned, then the FKs validate.
func TestUserFKMigrationCleansOrphans(t *testing.T) {
	db := newTestDB(t)
	live := seedUser(t, db, "live@x.test", true, "")
	const ghost = 987654
	// Prod state: no FK on user_lists/notifications/admin_audit_logs. library_items and media_likes
	// get their GORM-named FK back as NOT VALID so AutoMigrate (which runs first) leaves them alone.
	for _, fk := range migrate.UserFKs {
		for _, r := range userFKRows(t, db) {
			if r.Tbl == fk.Table && r.Col == fk.Column {
				if err := db.Exec(fmt.Sprintf(`ALTER TABLE %s DROP CONSTRAINT %s`, r.Tbl, r.Conname)).Error; err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	if err := db.Exec(`ALTER TABLE admin_audit_logs ALTER COLUMN admin_id SET NOT NULL`).Error; err != nil {
		t.Fatal(err)
	}
	seedUserRefs(t, db, live, live, "live")
	seedUserRefs(t, db, ghost, live, "ghost")
	for _, fk := range migrate.UserFKs {
		if fk.Table == "library_items" || fk.Table == "media_likes" {
			if err := db.Exec(fmt.Sprintf(`ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES users(id) ON DELETE NO ACTION NOT VALID`, fk.Table, fk.Constraint, fk.Column)).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := db.Exec(`DELETE FROM migration_history WHERE version IN (?, ?)`, fkNotValidMigration, fkValidateMigration).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db, testConfig()); err != nil {
		t.Fatal(err)
	}
	for _, fk := range migrate.UserFKs {
		var orphans int64
		db.Raw(fmt.Sprintf(`SELECT count(*) FROM %s WHERE %s = ?`, fk.Table, fk.Column), ghost).Scan(&orphans)
		var liveRows int64
		db.Raw(fmt.Sprintf(`SELECT count(*) FROM %s WHERE %s = ?`, fk.Table, fk.Column), live).Scan(&liveRows)
		if fk.OnDelete == migrate.OnDeleteSetNull {
			var nulls int64
			db.Raw(fmt.Sprintf(`SELECT count(*) FROM %s WHERE %s IS NULL`, fk.Table, fk.Column)).Scan(&nulls)
			if orphans != 0 || nulls != 1 || liveRows != 1 {
				t.Errorf("%s.%s: orphans=%d nulls=%d live=%d", fk.Table, fk.Column, orphans, nulls, liveRows)
			}
			continue
		}
		if orphans != 0 || liveRows != 1 {
			t.Errorf("%s.%s: orphans=%d live=%d", fk.Table, fk.Column, orphans, liveRows)
		}
	}
	for _, r := range userFKRows(t, db) {
		if !r.Validated {
			t.Errorf("%s.%s not validated", r.Tbl, r.Col)
		}
	}
}

// seedUserRefs inserts one row in every table that references users, for user uid
// (notifications go to recipient, which must exist).
func seedUserRefs(t *testing.T, db *gorm.DB, uid, recipient uint, tag string) {
	t.Helper()
	stmts := []string{
		`INSERT INTO library_items (created_at, updated_at, user_id, media_id, media_type, list_type) VALUES (now(), now(), ?, 1, 'movie', 'watchlist')`,
		`INSERT INTO media_likes (created_at, updated_at, user_id, media_id, media_type) VALUES (now(), now(), ?, 1, 'movie')`,
		`INSERT INTO user_lists (created_at, updated_at, user_id, name) VALUES (now(), now(), ?, ?)`,
		`INSERT INTO notifications (created_at, updated_at, user_id, actor_id, type, category, message) VALUES (now(), now(), ?, ?, 'follow', 'social', 'm')`,
		`INSERT INTO admin_audit_logs (admin_id, action, target_type, target_id, created_at) VALUES (?, 'USER_DELETED', 'user', 1, now())`,
	}
	for _, s := range stmts {
		args := []any{uid}
		if strings.Contains(s, "notifications") {
			args = []any{recipient, uid}
		}
		if strings.Contains(s, "user_lists") {
			args = append(args, "l-"+tag)
		}
		if err := db.Exec(s, args...).Error; err != nil {
			t.Fatalf("seed %q: %v", s, err)
		}
	}
}

// A hard DELETE of a fully seeded user leaves no row referencing it, and audit logs survive.
func TestUserHardDeleteCascades(t *testing.T) {
	e := newEnv(t)
	adminID := e.seedAdmin("admin@x.test", "admin-password1")
	other := seedUser(t, e.db, "other@x.test", true, "")
	uid := seedUser(t, e.db, "victim@x.test", true, "")
	seedUserRefs(t, e.db, uid, other, "victim")
	e.seedGoogleIdentity(uid, "victim-sub", "victim@x.test")
	e.seedLinkRows(uid, "planzi-victim")
	e.seedRefreshToken(uid, "victim-rt")
	e.seedActivity(uid, "public", nil)
	e.seedReview(uid, 4242)
	for _, pair := range [][2]uint{{uid, other}, {other, uid}} {
		if err := e.db.Exec(`INSERT INTO user_follows (follower_id, followee_id, created_at) VALUES (?, ?, now())`, pair[0], pair[1]).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := e.db.Exec(`INSERT INTO activity_events (actor_id, target_user_id, type, message, visibility, is_visible, created_at, updated_at)
		VALUES (?, ?, 'follow', 'm', 'public', true, now(), now())`, other, uid).Error; err != nil {
		t.Fatal(err)
	}
	// An audit log written by the admin we are about to delete.
	if err := e.db.Exec(`INSERT INTO admin_audit_logs (admin_id, action, target_type, target_id, created_at) VALUES (?, 'USER_DELETED', 'user', 1, now())`, adminID).Error; err != nil {
		t.Fatal(err)
	}

	if err := e.db.Exec(`DELETE FROM users WHERE id = ?`, uid).Error; err != nil {
		t.Fatalf("hard delete: %v", err)
	}
	for _, r := range userFKRows(t, e.db) {
		if r.Col == "target_user_id" {
			continue
		}
		if n := e.count(fmt.Sprintf(`SELECT count(*) FROM %s WHERE %s = ?`, r.Tbl, r.Col), uid); n != 0 {
			t.Errorf("%s.%s still has %d rows for the deleted user", r.Tbl, r.Col, n)
		}
	}
	if n := e.count(`SELECT count(*) FROM activity_events WHERE target_user_id IS NULL AND actor_id = ?`, other); n != 1 {
		t.Errorf("target_user_id not nulled: %d", n)
	}

	if err := e.db.Exec(`DELETE FROM users WHERE id = ?`, adminID).Error; err != nil {
		t.Fatalf("hard delete admin: %v", err)
	}
	rows, total, err := admin_module.NewRepository(e.db).ListAuditLogs(admin_module.AuditLogFilter{})
	if err != nil || total == 0 {
		t.Fatalf("list audit logs: total=%d err=%v", total, err)
	}
	for _, r := range rows {
		if r.AdminID != nil || r.AdminUsername != nil {
			t.Errorf("audit row %d still points at a deleted admin", r.ID)
		}
	}
}

// A row lock held on an orphan makes the validate migration hit lock_timeout; it defers: not recorded, FKs stay NOT VALID;
// the next run validates and records it.
func TestUserFKValidateDefersOnLockTimeout(t *testing.T) {
	db := newTestDB(t)
	live := seedUser(t, db, "live@x.test", true, "")
	const ghost = 987654
	if err := db.Exec(`ALTER TABLE user_lists DROP CONSTRAINT IF EXISTS fk_user_lists_user`).Error; err != nil {
		t.Fatal(err)
	}
	for _, uid := range []uint{live, ghost} {
		if err := db.Exec(`INSERT INTO user_lists (created_at, updated_at, user_id, name) VALUES (now(), now(), ?, ?)`, uid, fmt.Sprintf("l%d", uid)).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Exec(`ALTER TABLE user_lists ADD CONSTRAINT fk_user_lists_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE NOT VALID`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`DELETE FROM migration_history WHERE version = ?`, fkValidateMigration).Error; err != nil {
		t.Fatal(err)
	}

	cfg := testConfig()
	cfg.DB.MigrationLockTimeoutMs = 300
	holder := db.Begin()
	if err := holder.Exec(`SELECT id FROM user_lists WHERE user_id = ? FOR UPDATE`, ghost).Error; err != nil {
		t.Fatal(err)
	}
	err := database.Migrate(db, cfg)
	_ = holder.Rollback().Error
	if err != nil {
		t.Fatalf("deferred migration must not fail the boot: %v", err)
	}
	var recorded int64
	db.Raw(`SELECT count(*) FROM migration_history WHERE version = ?`, fkValidateMigration).Scan(&recorded)
	if recorded != 0 {
		t.Fatal("deferred migration was recorded")
	}
	var notValid int64
	db.Raw(`SELECT count(*) FROM pg_constraint WHERE conname = 'fk_user_lists_user' AND NOT convalidated`).Scan(&notValid)
	if notValid != 1 {
		t.Fatalf("fk_user_lists_user should still be NOT VALID, got %d", notValid)
	}

	if err := database.Migrate(db, testConfig()); err != nil {
		t.Fatal(err)
	}
	db.Raw(`SELECT count(*) FROM migration_history WHERE version = ?`, fkValidateMigration).Scan(&recorded)
	if recorded != 1 {
		t.Fatalf("retry did not record the migration: %d", recorded)
	}
	for _, r := range userFKRows(t, db) {
		if !r.Validated {
			t.Errorf("%s.%s not validated after retry", r.Tbl, r.Col)
		}
	}
	var orphans int64
	db.Raw(`SELECT count(*) FROM user_lists WHERE user_id = ?`, ghost).Scan(&orphans)
	if orphans != 0 {
		t.Fatalf("orphans left: %d", orphans)
	}
}

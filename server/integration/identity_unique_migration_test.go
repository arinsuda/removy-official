//go:build integration

package integration

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/arinsuda/movie-hub/database"
	"gorm.io/gorm"
)

const identityEnsureVersion = "20261107_user_identity_unique_ensure"

// dropIdentityUniques simulates a database where the raw-SQL migration was skipped and the
// ensure migration has not run yet.
func dropIdentityUniques(t *testing.T, db *gorm.DB) {
	t.Helper()
	for _, name := range []string{"uq_provider_subject", "uq_user_provider"} {
		db.Exec(`ALTER TABLE user_identities DROP CONSTRAINT IF EXISTS ` + name)
		db.Exec(`DROP INDEX IF EXISTS ` + name)
	}
	if err := db.Exec(`DELETE FROM migration_history WHERE version = ?`, identityEnsureVersion).Error; err != nil {
		t.Fatal(err)
	}
}

func indexCount(db *gorm.DB) int64 {
	var n int64
	db.Raw(`SELECT count(*) FROM pg_indexes WHERE schemaname = current_schema() AND indexname IN ('uq_provider_subject','uq_user_provider')`).Scan(&n)
	return n
}

func recorded(db *gorm.DB) int64 {
	var n int64
	db.Raw(`SELECT count(*) FROM migration_history WHERE version = ?`, identityEnsureVersion).Scan(&n)
	return n
}

func captureLog(t *testing.T) *bytes.Buffer {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	return &buf
}

// Duplicates must not fail the boot, delete anything or record the version; once they are
// resolved the next boot creates the indexes and records it.
func TestIdentityUniqueEnsureDefersOnDuplicates(t *testing.T) {
	db := newTestDB(t)
	a := seedUser(t, db, "a@x.test", true, "pw-aaaaaaaa")
	b := seedUser(t, db, "b@x.test", true, "pw-bbbbbbbb")
	dropIdentityUniques(t, db)
	ins := `INSERT INTO user_identities (user_id, provider, provider_subject, provider_email) VALUES (?, 'google', ?, 'x@x.test')`
	for _, args := range [][]any{{a, "same-sub"}, {b, "same-sub"}, {a, "other-sub"}} { // dup subject and dup (user, provider)
		if err := db.Exec(ins, args...).Error; err != nil {
			t.Fatal(err)
		}
	}
	buf := captureLog(t)

	if err := database.Migrate(db, testConfig()); err != nil {
		t.Fatalf("duplicates must not fail the boot: %v", err)
	}
	if !strings.Contains(buf.String(), "ERROR migration: uq_provider_subject") || !strings.Contains(buf.String(), "ERROR migration: uq_user_provider") {
		t.Fatalf("expected ERROR logs for both indexes, got:\n%s", buf.String())
	}
	if recorded(db) != 0 || indexCount(db) != 0 {
		t.Fatalf("recorded=%d indexes=%d, want none", recorded(db), indexCount(db))
	}
	if n := e2count(db, `SELECT count(*) FROM user_identities`); n != 3 {
		t.Fatalf("rows changed: %d", n)
	}

	// resolve by hand, reboot
	db.Exec(`DELETE FROM user_identities WHERE user_id = ? AND provider_subject = 'same-sub'`, b)
	db.Exec(`DELETE FROM user_identities WHERE provider_subject = 'other-sub'`)
	if err := database.Migrate(db, testConfig()); err != nil {
		t.Fatal(err)
	}
	if recorded(db) != 1 || indexCount(db) != 2 {
		t.Fatalf("after fixing: recorded=%d indexes=%d, want 1 and 2", recorded(db), indexCount(db))
	}
}

func e2count(db *gorm.DB, sql string) int64 {
	var n int64
	db.Raw(sql).Scan(&n)
	return n
}

// Indexes already present (fresh DB or the raw-SQL migration ran): nothing to do, version recorded.
func TestIdentityUniqueEnsureWithIndexesPresent(t *testing.T) {
	db := newTestDB(t)
	if indexCount(db) != 2 {
		t.Fatalf("fresh migrate left %d indexes", indexCount(db))
	}
	db.Exec(`DELETE FROM migration_history WHERE version = ?`, identityEnsureVersion)
	if err := database.Migrate(db, testConfig()); err != nil {
		t.Fatal(err)
	}
	if recorded(db) != 1 || indexCount(db) != 2 {
		t.Fatalf("recorded=%d indexes=%d", recorded(db), indexCount(db))
	}
}

// An index with the same name in another schema must not count as present.
func TestIdentityUniqueEnsureIgnoresOtherSchemas(t *testing.T) {
	db := newTestDB(t)
	dropIdentityUniques(t, db)
	db.Exec(`CREATE SCHEMA other_it`)
	db.Exec(`CREATE TABLE other_it.t (a int, b int)`)
	db.Exec(`CREATE UNIQUE INDEX uq_provider_subject ON other_it.t (a)`)
	db.Exec(`CREATE UNIQUE INDEX uq_user_provider ON other_it.t (b)`)
	if err := database.Migrate(db, testConfig()); err != nil {
		t.Fatal(err)
	}
	if indexCount(db) != 2 || recorded(db) != 1 {
		t.Fatalf("index in another schema hid the missing one: indexes=%d recorded=%d", indexCount(db), recorded(db))
	}
}

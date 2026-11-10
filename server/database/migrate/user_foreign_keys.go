package migrate

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// ON DELETE rules accepted by ReplaceUserFK.
const (
	OnDeleteCascade = "CASCADE"
	OnDeleteSetNull = "SET NULL"
)

// UserFK describes one column that must reference users(id).
type UserFK struct {
	Table, Column, Constraint, OnDelete string
}

// UserFKs is the set repaired by the user_fk_cascade migrations.
var UserFKs = []UserFK{
	{"library_items", "user_id", "fk_library_items_user", OnDeleteCascade},
	{"media_likes", "user_id", "fk_media_likes_user", OnDeleteCascade},
	{"user_lists", "user_id", "fk_user_lists_user", OnDeleteCascade},
	{"notifications", "actor_id", "fk_notifications_actor", OnDeleteCascade},
	{"admin_audit_logs", "admin_id", "fk_admin_audit_logs_admin", OnDeleteSetNull},
}

// userFKNames lists every FK on table.column that points at users, whatever its name.
func userFKNames(tx *gorm.DB, table, column string) ([]string, error) {
	var names []string
	err := tx.Raw(`SELECT c.conname FROM pg_constraint c
		JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = ANY (c.conkey)
		WHERE c.contype = 'f' AND c.conrelid = to_regclass(?) AND c.confrelid = to_regclass('users')
		AND array_length(c.conkey, 1) = 1 AND a.attname = ?`, table, column).Scan(&names).Error
	return names, err
}

// Orphan cleanup works in id ranges so no single statement runs long.
const (
	UserFKOrphanBatchSize = 50000
	// UserFKValidateTimeoutMs bounds VALIDATE CONSTRAINT, which only takes SHARE UPDATE EXCLUSIVE.
	UserFKValidateTimeoutMs = 600000
)

// RemoveUserOrphans clears rows whose user id has no users row: deleted, or the
// column set to NULL for SET NULL. Identifiers come from UserFKs, never from input.
// An error (statement timeout, lock timeout) is returned between batches, so the
// caller can defer and retry.
func RemoveUserOrphans(tx *gorm.DB, fk UserFK) error {
	var bounds struct{ Lo, Hi *int64 }
	if err := tx.Raw(fmt.Sprintf("SELECT min(id) AS lo, max(id) AS hi FROM %s", fk.Table)).Scan(&bounds).Error; err != nil {
		return err
	}
	if bounds.Lo == nil || bounds.Hi == nil {
		return nil
	}
	cond := fmt.Sprintf("id >= ? AND id < ? AND %s IS NOT NULL AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id = %s.%s)", fk.Column, fk.Table, fk.Column)
	for lo := *bounds.Lo; lo <= *bounds.Hi; lo += UserFKOrphanBatchSize {
		var stmt string
		if fk.OnDelete == OnDeleteSetNull {
			stmt = fmt.Sprintf("UPDATE %s SET %s = NULL WHERE %s", fk.Table, fk.Column, cond)
		} else {
			stmt = fmt.Sprintf("DELETE FROM %s WHERE %s", fk.Table, cond)
		}
		if err := tx.Exec(stmt, lo, lo+UserFKOrphanBatchSize).Error; err != nil {
			return err
		}
	}
	return nil
}

// ReplaceUserFK drops any existing FK from table.column to users and adds
// fk.Constraint as NOT VALID in one ALTER TABLE: no table scan under the lock, and
// it needs no clean data (new writes are checked at once). Orphans are removed and
// the constraint validated later by ValidateUserFK. A SET NULL rule also lifts
// NOT NULL on the column.
func ReplaceUserFK(tx *gorm.DB, fk UserFK) error {
	names, err := userFKNames(tx, fk.Table, fk.Column)
	if err != nil {
		return err
	}
	var actions []string
	for _, n := range names {
		actions = append(actions, "DROP CONSTRAINT "+quoteIdent(n))
	}
	if fk.OnDelete == OnDeleteSetNull {
		actions = append(actions, "ALTER COLUMN "+fk.Column+" DROP NOT NULL")
	}
	actions = append(actions, fmt.Sprintf("ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES users(id) ON DELETE %s NOT VALID",
		fk.Constraint, fk.Column, fk.OnDelete))
	return tx.Exec(fmt.Sprintf("ALTER TABLE %s %s", fk.Table, strings.Join(actions, ", "))).Error
}

// ValidateUserFK clears orphans in batches, then validates the constraint under
// UserFKValidateTimeoutMs.
func ValidateUserFK(tx *gorm.DB, fk UserFK) error {
	if err := RemoveUserOrphans(tx, fk); err != nil {
		return err
	}
	if err := tx.Exec(fmt.Sprintf("SET LOCAL statement_timeout = %d", UserFKValidateTimeoutMs)).Error; err != nil {
		return err
	}
	return tx.Exec(fmt.Sprintf("ALTER TABLE %s VALIDATE CONSTRAINT %s", fk.Table, fk.Constraint)).Error
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

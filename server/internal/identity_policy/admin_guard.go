package identity_policy

import "gorm.io/gorm"

// AdminGuardLockKey is the pg_advisory_xact_lock key every path that could remove the
// last active admin (role demotion, deactivation, deletion, self-delete) takes.
const AdminGuardLockKey int64 = 7_310_020_001

// LockAdminGuard serialises admin-removing transactions until commit. Call it first in the
// transaction, before locking any user row: row locks (and FK checks on audit rows) taken
// earlier would deadlock against another transaction holding this lock.
func LockAdminGuard(tx *gorm.DB) error {
	return tx.Exec("SELECT pg_advisory_xact_lock(?)", AdminGuardLockKey).Error
}

// CountActiveAdmins counts active admins; call it after LockAdminGuard. Without the lock two
// READ COMMITTED transactions each see the other admin and both proceed, leaving none.
func CountActiveAdmins(tx *gorm.DB, adminRole string) (int64, error) {
	var n int64
	err := tx.Table("users").
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.role_name = ? AND users.is_active = true AND users.deleted_at IS NULL", adminRole).
		Count(&n).Error
	return n, err
}

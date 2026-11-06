package database

import (
	"fmt"
	"hash/fnv"
	"log"

	"gorm.io/gorm"
)

func runMigrationWithHistoryTx(
	db *gorm.DB,
	migrationName string,
	lockTimeoutMs int,
	statementTimeoutMs int,
	migrate func(tx *gorm.DB) error,
) error {
	return runVersionedMigrationTx(db, migrationName, lockTimeoutMs, statementTimeoutMs, true, migrate)
}

// runSchemaMigrationTx is runMigrationWithHistoryTx without the reviews table
// lock, for migrations that touch other tables. lock_timeout makes it fail the
// boot instead of queueing behind a long-running transaction.
func runSchemaMigrationTx(
	db *gorm.DB,
	migrationName string,
	lockTimeoutMs int,
	statementTimeoutMs int,
	migrate func(tx *gorm.DB) error,
) error {
	return runVersionedMigrationTx(db, migrationName, lockTimeoutMs, statementTimeoutMs, false, migrate)
}

func runVersionedMigrationTx(
	db *gorm.DB,
	migrationName string,
	lockTimeoutMs int,
	statementTimeoutMs int,
	lockReviews bool,
	migrate func(tx *gorm.DB) error,
) error {
	var applied bool
	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. Session-level advisory lock using FNV-1a hash of the name
		if err := tx.Exec(`SELECT pg_advisory_xact_lock(?)`, hashString(migrationName)).Error; err != nil {
			return fmt.Errorf("advisory lock failed for %s: %w", migrationName, err)
		}

		// 2. Guard via migration_history
		var count int64
		if err := tx.Table("migration_history").Where("version = ?", migrationName).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}

		// 3. Apply transaction-local timeouts
		if err := tx.Exec(fmt.Sprintf("SET LOCAL lock_timeout = %d", lockTimeoutMs)).Error; err != nil {
			return fmt.Errorf("failed to set lock_timeout: %w", err)
		}
		if err := tx.Exec(fmt.Sprintf("SET LOCAL statement_timeout = %d", statementTimeoutMs)).Error; err != nil {
			return fmt.Errorf("failed to set statement_timeout: %w", err)
		}

		// 4. SHARE ROW EXCLUSIVE table lock on reviews (blocks writes, allows reads)
		if lockReviews {
			if err := tx.Exec(`LOCK TABLE reviews IN SHARE ROW EXCLUSIVE MODE`).Error; err != nil {
				return fmt.Errorf("failed to lock reviews table: %w", err)
			}
		}

		// 5. Run migration callback
		if err := migrate(tx); err != nil {
			return err
		}

		// 6. Record history in the same transaction
		if err := tx.Exec(`INSERT INTO migration_history (version) VALUES (?)`, migrationName).Error; err != nil {
			return fmt.Errorf("failed to record migration history for %s: %w", migrationName, err)
		}

		applied = true
		return nil
	})

	if err == nil {
		if applied {
			log.Printf("migration %s applied successfully", migrationName)
		} else {
			log.Printf("migration %s already applied", migrationName)
		}
	}
	return err
}

func hashString(s string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return int64(h.Sum64())
}

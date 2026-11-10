package config

import "time"

// Account erasure (PDPA). A self-requested erasure and an admin soft delete both wait this long
// before the purge worker removes the account for good.
const (
	// ErasureGracePeriod is how long a user who requested erasure stays recoverable by an admin.
	ErasureGracePeriod = 30 * 24 * time.Hour
	// AdminDeletePurgeAfter is the age at which an admin-deleted account is purged.
	AdminDeletePurgeAfter = 30 * 24 * time.Hour
	// ErasurePurgeInterval is the purge worker tick.
	ErasurePurgeInterval = time.Hour
	// ErasurePurgeBatch caps accounts erased per tick; the rest wait for the next one.
	ErasurePurgeBatch = 50
	// ErasureCleanupRetryBatch caps audit rows whose avatar cleanup is retried per tick.
	ErasureCleanupRetryBatch = 50
	// ErasureLockTimeout and ErasureStatementTimeout bound one account's erase transaction.
	ErasureLockTimeout      = 10 * time.Second
	ErasureStatementTimeout = 30 * time.Second
	// ErasureSupportEmail is where a user writes to cancel a pending erasure within the grace period.
	ErasureSupportEmail = "removy.official@gmail.com"
)

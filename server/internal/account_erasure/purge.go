package account_erasure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/identity_policy"
	"github.com/arinsuda/movie-hub/internal/shared/storage"
	"github.com/arinsuda/movie-hub/internal/user_module"
	"gorm.io/gorm"
)

const (
	// ActionUserErasedScheduled is the audit action of an erasure run by the purge worker.
	ActionUserErasedScheduled = "USER_ERASED_SCHEDULED"

	// Why the worker picked the account (audit meta "trigger").
	TriggerSelfRequest = "self_request"
	TriggerAdminDelete = "admin_delete_expired"

	metaTrigger        = "trigger"
	metaRole           = "target_role"
	metaCleanup        = "storage_cleanup"
	cleanupPending     = "pending"
	cleanupDone        = "done"
	storageCallTimeout = 15 * time.Second
)

// Storage is the object store the worker cleans avatars from after the commit.
type Storage interface {
	DeleteObject(ctx context.Context, key string) error
	DeletePrefix(ctx context.Context, prefix string) error
}

// Purger erases accounts whose grace period is over: self-requested erasures (users.erasure_due_at
// has passed) and admin-deleted accounts older than config.AdminDeletePurgeAfter.
type Purger struct {
	db    *gorm.DB
	store Storage // nil: avatars are left behind and the audit row stays storage_cleanup "pending"
	now   func() time.Time
}

func NewPurger(db *gorm.DB, store Storage) *Purger {
	return &Purger{db: db, store: store, now: time.Now}
}

// Run ticks every config.ErasurePurgeInterval until ctx is cancelled.
func (p *Purger) Run(ctx context.Context) {
	ticker := time.NewTicker(config.ErasurePurgeInterval)
	defer ticker.Stop()
	for {
		p.tick(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// tick runs one pass and never lets a panic kill the worker goroutine.
func (p *Purger) tick(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("erasure purge: recovered from panic: %v", r)
		}
	}()
	erased, err := p.RunOnce(ctx)
	retried := p.RetryCleanups(ctx)
	if err != nil {
		log.Printf("erasure purge: %v", err)
	}
	log.Printf("erasure purge: tick erased=%d cleanup_retried=%d", erased, retried)
}

// RunOnce erases up to config.ErasurePurgeBatch due accounts. It is safe to run on several
// instances at once: every account is claimed with LockAdminGuard + FOR UPDATE and its due
// condition is re-checked under the lock, so a second runner skips what the first one erased.
func (p *Purger) RunOnce(ctx context.Context) (erased int, err error) {
	now := p.now().UTC()
	var ids []uint
	if err := p.db.WithContext(ctx).Raw(`SELECT id FROM users
		WHERE (erasure_due_at <= ? AND is_active = false) OR (deleted_at IS NOT NULL AND deleted_at <= ?)
		ORDER BY id LIMIT ?`, now, now.Add(-config.AdminDeletePurgeAfter), config.ErasurePurgeBatch).Scan(&ids).Error; err != nil {
		return 0, err
	}
	for _, id := range ids {
		if ctx.Err() != nil {
			return erased, ctx.Err()
		}
		ok, err := p.purgeOne(ctx, id, now)
		if err != nil {
			// No PII: the id only.
			log.Printf("erasure purge: user %d: %v", id, err)
			continue
		}
		if ok {
			erased++
		}
	}
	return erased, nil
}

// RetryCleanups re-runs the avatar cleanup of recent erasures whose storage call failed. The
// prefixes cover every avatar the user ever owned, so the row's target id is enough.
func (p *Purger) RetryCleanups(ctx context.Context) int {
	if p.store == nil {
		return 0
	}
	var rows []struct{ ID, TargetID uint }
	if err := p.db.WithContext(ctx).Raw(`SELECT id, target_id FROM admin_audit_logs
		WHERE action = ? AND meta_data->>? = ? ORDER BY id LIMIT ?`,
		ActionUserErasedScheduled, metaCleanup, cleanupPending, config.ErasureCleanupRetryBatch).Scan(&rows).Error; err != nil {
		log.Printf("erasure purge: list pending cleanups: %v", err)
		return 0
	}
	done := 0
	for _, r := range rows {
		if p.cleanup(r.TargetID, r.ID, Result{AvatarPrefixes: storage.AvatarPrefixes(r.TargetID)}) {
			done++
		}
	}
	return done
}

var errSkipped = errors.New("skipped")

// purgeOne erases one account in its own transaction and cleans storage after the commit.
// It returns false when the account is no longer due (reactivated meanwhile) or erasing it
// would remove the last active admin.
func (p *Purger) purgeOne(ctx context.Context, id uint, now time.Time) (bool, error) {
	var res Result
	var auditID uint
	err := p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Bounded waits: a stuck lock must not hang the worker.
		if err := tx.Exec(fmt.Sprintf("SET LOCAL lock_timeout = %d", config.ErasureLockTimeout.Milliseconds())).Error; err != nil {
			return err
		}
		if err := tx.Exec(fmt.Sprintf("SET LOCAL statement_timeout = %d", config.ErasureStatementTimeout.Milliseconds())).Error; err != nil {
			return err
		}
		if err := identity_policy.LockAdminGuard(tx); err != nil {
			return err
		}
		var target struct {
			ID           uint
			IsActive     bool
			RoleName     string
			DeletedAt    *time.Time
			ErasureDueAt *time.Time
		}
		// Raw SQL so soft-deleted rows are visible; the due check is repeated under the row lock.
		if err := tx.Raw(`SELECT u.id, u.is_active, u.deleted_at, u.erasure_due_at, roles.role_name
			FROM users u JOIN roles ON roles.id = u.role_id
			WHERE u.id = ?
			AND ((u.erasure_due_at <= ? AND u.is_active = false) OR (u.deleted_at IS NOT NULL AND u.deleted_at <= ?))
			FOR UPDATE OF u`, id, now, now.Add(-config.AdminDeletePurgeAfter)).Scan(&target).Error; err != nil {
			return err
		}
		if target.ID == 0 {
			return errSkipped
		}
		if target.RoleName == string(user_module.RoleAdmin) && target.IsActive {
			n, err := identity_policy.CountActiveAdmins(tx, string(user_module.RoleAdmin))
			if err != nil {
				return err
			}
			if n <= 1 {
				log.Printf("erasure purge: user %d skipped: last active admin", id)
				return errSkipped
			}
		}

		trigger := TriggerSelfRequest
		if target.DeletedAt != nil {
			trigger = TriggerAdminDelete
		}
		var err error
		if res, err = Erase(tx, id); err != nil {
			return err
		}
		meta, _ := json.Marshal(map[string]any{
			metaTrigger: trigger,
			metaRole:    target.RoleName,
			metaCleanup: cleanupPending,
		})
		return tx.Raw(`INSERT INTO admin_audit_logs (admin_id, action, target_type, target_id, meta_data, created_at)
			VALUES (NULL, ?, ?, ?, ?::jsonb, ?) RETURNING id`,
			ActionUserErasedScheduled, auditTargetType, id, string(meta), now).Scan(&auditID).Error
	})
	if errors.Is(err, errSkipped) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	identity_policy.Outcome{SessionsRevoked: true}.Publish(id)
	p.cleanup(id, auditID, res)
	return true, nil
}

// cleanup deletes the avatars after the commit. A failure leaves storage_cleanup "pending".
func (p *Purger) cleanup(userID, auditID uint, res Result) bool {
	if p.store == nil {
		return false
	}
	ok := true
	for _, prefix := range res.AvatarPrefixes {
		ok = p.storageCall(p.store.DeletePrefix, prefix) && ok
	}
	if res.AvatarKey != "" {
		ok = p.storageCall(p.store.DeleteObject, res.AvatarKey) && ok
	}
	if !ok {
		return false
	}
	if err := p.db.Exec(`UPDATE admin_audit_logs SET meta_data = jsonb_set(meta_data, ARRAY[?], to_jsonb(?::text))
		WHERE id = ? AND action = ?`, metaCleanup, cleanupDone, auditID, ActionUserErasedScheduled).Error; err != nil {
		log.Printf("erasure purge: user %d: record cleanup failed: %v", userID, err)
		return false
	}
	return true
}

func (p *Purger) storageCall(del func(context.Context, string) error, key string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), storageCallTimeout)
	defer cancel()
	if err := del(ctx, key); err != nil {
		log.Printf("erasure purge: storage cleanup of %q failed: %v", key, err)
		return false
	}
	return true
}

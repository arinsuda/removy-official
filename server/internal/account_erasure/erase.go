// Package account_erasure permanently removes a user row and everything that must not outlive it.
// It is the explicit exception to soft delete (AGENTS.md hard rule 5): an admin runs it as the
// second step, after the soft delete.
package account_erasure

import (
	"fmt"

	"github.com/arinsuda/movie-hub/internal/shared/storage"
	"gorm.io/gorm"
)

const (
	// auditTargetType marks admin_audit_logs rows whose target is a user.
	auditTargetType = "user"
	// auditPIIKeys are the meta_data keys that identify the erased user.
	auditPIIKeyUsername = "target_username"
	auditPIIKeyEmail    = "target_email"
)

// Result tells the caller what to clean up outside the database after the commit.
type Result struct {
	// AvatarKey is the stored avatar object, empty when none (or an external URL).
	AvatarKey string
	// AvatarPrefixes cover every avatar this user ever owned (uploads and Google imports).
	AvatarPrefixes []string
}

// counterFixes lowers the denormalised counters on other people's reviews for the rows the
// erased user leaves behind. Likes and helpful votes have no soft delete; comments already
// soft-deleted were decremented when they were deleted, so only live ones count.
// media_stats.like_count is not maintained anywhere (stats read live COUNT(*) from
// media_likes), so there is nothing to decrement there.
var counterFixes = []struct{ counter, source, filter string }{
	{"like_count", "review_likes", ""},
	{"helpful_count", "review_helpfuls", ""},
	{"comment_count", "review_comments", " AND deleted_at IS NULL"},
}

// reviewChildren hang off reviews.id without a foreign key, so the cascade from users to reviews
// would leave them orphaned. Others' likes, helpful votes, comments and feed events about the
// erased user's reviews are removed with them.
var reviewChildren = []string{
	"review_likes", "review_helpfuls", "review_comments", "activity_events",
}

// Erase hard-deletes userID inside tx. The caller must hold identity_policy.LockAdminGuard and
// have validated the request. Rows owned by the user go via FK CASCADE; audit logs stay with
// admin_id SET NULL and the target's identifying meta removed.
func Erase(tx *gorm.DB, userID uint) (Result, error) {
	var res Result

	var avatar *string
	if err := tx.Raw(`SELECT avatar_url FROM users WHERE id = ?`, userID).Scan(&avatar).Error; err != nil {
		return res, fmt.Errorf("read avatar: %w", err)
	}
	if avatar != nil && storage.IsAvatarKey(*avatar) {
		res.AvatarKey = *avatar
	}
	res.AvatarPrefixes = storage.AvatarPrefixes(userID)

	for _, f := range counterFixes {
		// counter, source and filter are constants from counterFixes, never input.
		q := fmt.Sprintf(`UPDATE reviews r SET %[1]s = GREATEST(r.%[1]s - x.n, 0)
			FROM (SELECT review_id, count(*) AS n FROM %[2]s WHERE user_id = ?%[3]s GROUP BY review_id) x
			WHERE r.id = x.review_id`, f.counter, f.source, f.filter)
		if err := tx.Exec(q, userID).Error; err != nil {
			return res, fmt.Errorf("decrement reviews.%s: %w", f.counter, err)
		}
	}

	for _, table := range reviewChildren {
		// table comes from reviewChildren, never input. Includes soft-deleted reviews.
		q := fmt.Sprintf(`DELETE FROM %s WHERE review_id IN (SELECT id FROM reviews WHERE user_id = ?)`, table)
		if err := tx.Exec(q, userID).Error; err != nil {
			return res, fmt.Errorf("delete %s of user reviews: %w", table, err)
		}
	}

	// Their message names the user, and the FK would only null the target.
	if err := tx.Exec(`DELETE FROM activity_events WHERE target_user_id = ?`, userID).Error; err != nil {
		return res, fmt.Errorf("delete activity about user: %w", err)
	}

	if err := tx.Exec(`UPDATE admin_audit_logs SET meta_data = meta_data - ? - ?
		WHERE target_type = ? AND target_id = ? AND meta_data IS NOT NULL`,
		auditPIIKeyUsername, auditPIIKeyEmail, auditTargetType, userID).Error; err != nil {
		return res, fmt.Errorf("strip audit pii: %w", err)
	}
	// The reason is free text an admin typed and may name the person.
	if err := tx.Exec(`UPDATE admin_audit_logs SET reason = NULL
		WHERE target_type = ? AND target_id = ? AND reason IS NOT NULL`, auditTargetType, userID).Error; err != nil {
		return res, fmt.Errorf("redact audit reason: %w", err)
	}

	if err := tx.Exec(`DELETE FROM users WHERE id = ?`, userID).Error; err != nil {
		return res, fmt.Errorf("delete user: %w", err)
	}
	return res, nil
}

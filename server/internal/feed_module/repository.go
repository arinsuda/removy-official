package feed_module

import (
	"context"
	"errors"
	"time"

	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrActivityNotFound = errors.New("activity not found")
	ErrForbidden        = errors.New("forbidden")
	ErrUserNotFound     = errors.New("user not found")
)

type repository struct {
	db *gorm.DB
}

func newRepository(db *gorm.DB) *repository {
	return &repository{db: db}
}

func (r *repository) FindByID(ctx context.Context, id uint) (*ActivityEvent, error) {
	var e ActivityEvent
	err := r.db.WithContext(ctx).First(&e, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrActivityNotFound
	}
	return &e, err
}

func (r *repository) CreateOrRestore(ctx context.Context, event *ActivityEvent) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing ActivityEvent
		query := tx.Unscoped().Where("actor_id = ? AND type = ?", event.ActorID, string(event.Type))
		if event.MediaID != nil && event.MediaType != nil {
			query = query.Where("media_id = ? AND media_type = ?", event.MediaID, event.MediaType)
		} else if event.ReviewID != nil {
			query = query.Where("review_id = ?", event.ReviewID)
		} else if event.CommentID != nil {
			query = query.Where("comment_id = ?", event.CommentID)
		} else if event.AchievementID != nil {
			query = query.Where("achievement_id = ?", event.AchievementID)
		} else if event.TargetUserID != nil {
			query = query.Where("target_user_id = ?", event.TargetUserID)
		}

		err := query.Clauses(clause.Locking{Strength: "UPDATE"}).First(&existing).Error
		if err == nil {
			if !existing.DeletedAt.Valid {
				return nil
			}
			updates := map[string]any{
				"deleted_at": gorm.DeletedAt{Valid: false},
				"created_at": time.Now().UTC(),
				"visibility": string(event.Visibility),
				"is_visible": true,
				"message":    event.Message,
			}
			return tx.Unscoped().Model(&existing).Updates(updates).Error
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		event.CreatedAt = time.Now().UTC()
		return tx.Create(event).Error
	})
}

func (r *repository) UpdateOwnedActivityVisibility(ctx context.Context, activityID, requesterID uint, visibility ActivityVisibility) error {
	result := r.db.WithContext(ctx).Model(&ActivityEvent{}).
		Where("id = ? AND actor_id = ? AND deleted_at IS NULL", activityID, requesterID).
		Update("visibility", string(visibility))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrActivityNotFound
	}
	return nil
}

func (r *repository) DeleteOwnedActivity(ctx context.Context, activityID, requesterID uint) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND actor_id = ? AND deleted_at IS NULL", activityID, requesterID).
		Delete(&ActivityEvent{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrActivityNotFound
	}
	return nil
}

func (r *repository) DeleteReviewActivity(ctx context.Context, actorID uint, reviewID uint) error {
	return r.db.WithContext(ctx).
		Where("actor_id = ? AND review_id = ? AND deleted_at IS NULL", actorID, reviewID).
		Delete(&ActivityEvent{}).Error
}

func (r *repository) DeleteCommentActivity(ctx context.Context, actorID uint, commentID uint) error {
	return r.db.WithContext(ctx).
		Where("actor_id = ? AND comment_id = ? AND deleted_at IS NULL", actorID, commentID).
		Delete(&ActivityEvent{}).Error
}

func (r *repository) DeleteWatchLogActivity(ctx context.Context, actorID uint, watchLogID uint) error {
	return r.db.WithContext(ctx).
		Where("actor_id = ? AND watch_log_id = ? AND deleted_at IS NULL", actorID, watchLogID).
		Delete(&ActivityEvent{}).Error
}

func (r *repository) DeleteMediaActivity(ctx context.Context, actorID uint, activityType ActivityType, mediaID int, mediaType string) error {
	return r.db.WithContext(ctx).
		Where("actor_id = ? AND type = ? AND media_id = ? AND media_type = ? AND deleted_at IS NULL", actorID, string(activityType), mediaID, mediaType).
		Delete(&ActivityEvent{}).Error
}

func (r *repository) DeleteFollowActivity(ctx context.Context, followerID uint, followeeID uint) error {
	return r.db.WithContext(ctx).
		Where("actor_id = ? AND type = ? AND target_user_id = ? AND deleted_at IS NULL", followerID, string(ActivityUserFollowed), followeeID).
		Delete(&ActivityEvent{}).Error
}

type feedRow struct {
	ID   uint
	Type privacy_policy.ActivityType

	ActorID          uint
	ActorUsername    string
	ActorDisplayName *string
	ActorAvatarURL   *string

	MediaID   *int
	MediaType *string

	ReviewID      *uint
	CommentID     *uint
	AchievementID *uint
	LibraryItemID *uint

	TargetUserID          *uint
	TargetUserUsername    string
	TargetUserDisplayName *string
	TargetUserAvatarURL   *string

	Message    string
	Visibility privacy_policy.ActivityVisibility
	CreatedAt  time.Time
}

const feedSelectColumns = `
	ae.id, ae.type,
	ae.actor_id, u.username AS actor_username, u.display_name AS actor_display_name, u.avatar_url AS actor_avatar_url,
	ae.media_id, ae.media_type,
	ae.review_id, ae.comment_id, ae.achievement_id, ae.library_item_id,
	ae.target_user_id, tu.username AS target_user_username, tu.display_name AS target_user_display_name, tu.avatar_url AS target_user_avatar_url,
	ae.message, ae.visibility, ae.created_at
`

func (r *repository) FindFeed(ctx context.Context, userID uint, pq PaginationQuery) ([]feedRow, int64, error) {
	buildQuery := func() *gorm.DB {
		db := r.db.WithContext(ctx).Table("activity_events ae").
			Joins("JOIN users u ON u.id = ae.actor_id AND u.is_active = true").
			Joins("LEFT JOIN user_follows uf ON uf.followee_id = ae.actor_id AND uf.follower_id = ?", userID).
			Joins("LEFT JOIN activity_privacy_settings aps ON aps.user_id = ae.actor_id AND aps.activity_type = ae.type").
			Where("ae.deleted_at IS NULL").
			Where("ae.actor_id <> ?", userID).
			Where(`
				(uf.status = 'accepted' AND ae.is_visible = true)
				AND (
					` + activityEnabledSQL + `
				)
				AND (
					(u.is_private = false OR uf.status = 'accepted')
					AND (
						ae.visibility = 'public' 
						OR (ae.visibility = 'followers' AND uf.status = 'accepted')
						OR (ae.visibility = 'default' AND (u.is_private = false OR uf.status = 'accepted'))
					)
				)
			`)

		if pq.Category == "reviews" {
			db = db.Where("ae.type = 'review_created'")
		} else if pq.Category == "lists" {
			db = db.Where("ae.type IN ('watchlist_added', 'watched_added')")
		} else if pq.Category == "social" {
			db = db.Where("ae.type IN ('user_followed', 'achievement_unlocked', 'review_liked', 'review_commented', 'media_liked')")
		}

		return db
	}

	var total int64
	if err := buildQuery().Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []feedRow
	err := buildQuery().
		Select(feedSelectColumns).
		Joins("LEFT JOIN users tu ON tu.id = ae.target_user_id").
		Order("ae.created_at DESC, ae.id DESC").
		Limit(pq.Limit).Offset(pq.Offset()).
		Scan(&rows).Error

	return rows, total, err
}

func (r *repository) FindUserActivities(ctx context.Context, targetUserID, requesterID uint, pq PaginationQuery) ([]feedRow, int64, error) {
	buildQuery := func() *gorm.DB {
		return r.db.WithContext(ctx).Table("activity_events ae").
			Joins("JOIN users u ON u.id = ae.actor_id AND u.is_active = true").
			Joins("LEFT JOIN user_follows uf ON uf.followee_id = ae.actor_id AND uf.follower_id = ?", requesterID).
			Joins("LEFT JOIN activity_privacy_settings aps ON aps.user_id = ae.actor_id AND aps.activity_type = ae.type").
			Where("ae.actor_id = ? AND ae.deleted_at IS NULL", targetUserID).
			Where(`
				(ae.actor_id = ? OR (
					ae.is_visible = true
					AND (
						`+activityEnabledSQL+`
					)
					AND (u.is_private = false OR uf.status = 'accepted')
					AND (
						ae.visibility = 'public'
						OR (ae.visibility = 'followers' AND uf.status = 'accepted')
						OR (ae.visibility = 'default' AND (u.is_private = false OR uf.status = 'accepted'))
					)
				))
			`, requesterID)
	}

	var total int64
	if err := buildQuery().Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []feedRow
	err := buildQuery().
		Select(feedSelectColumns).
		Joins("LEFT JOIN users tu ON tu.id = ae.target_user_id").
		Order("ae.created_at DESC, ae.id DESC").
		Limit(pq.Limit).Offset(pq.Offset()).
		Scan(&rows).Error

	return rows, total, err
}

func (r *repository) FindSettings(ctx context.Context, userID uint) ([]ActivityPrivacySetting, error) {
	var rows []ActivityPrivacySetting
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error
	return rows, err
}

func (r *repository) IsEnabled(ctx context.Context, userID uint, t ActivityType) (bool, error) {
	var setting ActivityPrivacySetting
	err := r.db.WithContext(ctx).Where("user_id = ? AND activity_type = ?", userID, string(t)).First(&setting).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return privacy_policy.ActivityEnabledByDefault(t), nil
	}
	if err != nil {
		return false, err
	}
	return setting.Enabled, nil
}

func (r *repository) UpsertSetting(ctx context.Context, userID uint, t ActivityType, enabled bool) error {
	setting := ActivityPrivacySetting{UserID: userID, ActivityType: t, Enabled: enabled}
	return r.db.WithContext(ctx).
		Where(ActivityPrivacySetting{UserID: userID, ActivityType: t}).
		Assign(ActivityPrivacySetting{Enabled: enabled}).
		FirstOrCreate(&setting).Error
}

func (r *repository) CountNewFeedItems(ctx context.Context, userID uint, afterActivityID uint, category string) (int64, error) {
	var count int64
	db := r.db.WithContext(ctx).Table("activity_events ae").
		Joins("JOIN users u ON u.id = ae.actor_id AND u.is_active = true").
		Joins("LEFT JOIN user_follows uf ON uf.followee_id = ae.actor_id AND uf.follower_id = ?", userID).
		Joins("LEFT JOIN activity_privacy_settings aps ON aps.user_id = ae.actor_id AND aps.activity_type = ae.type").
		Where("ae.deleted_at IS NULL").
		Where("ae.id > ?", afterActivityID).
		Where("ae.actor_id <> ?", userID).
		Where(`
			(uf.status = 'accepted' AND ae.is_visible = true)
			AND (
				` + activityEnabledSQL + `
			)
			AND (
				(u.is_private = false OR uf.status = 'accepted')
				AND (
					ae.visibility = 'public' 
					OR (ae.visibility = 'followers' AND uf.status = 'accepted')
					OR (ae.visibility = 'default' AND (u.is_private = false OR uf.status = 'accepted'))
				)
			)
		`)

	if category == "reviews" {
		db = db.Where("ae.type = 'review_created'")
	} else if category == "lists" {
		db = db.Where("ae.type IN ('watchlist_added', 'watched_added')")
	} else if category == "social" {
		db = db.Where("ae.type IN ('user_followed', 'achievement_unlocked', 'review_liked', 'review_commented', 'media_liked')")
	}

	err := db.Count(&count).Error
	return count, err
}

// activityEnabledSQL is the preference-or-default predicate shared by the feed queries.
var activityEnabledSQL = privacy_policy.ActivityEnabledSQL("ae.type")

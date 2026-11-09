package privacy_policy

import (
	"context"
	"database/sql"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrForbidden = errors.New("forbidden")
	ErrNotFound  = errors.New("user not found")
)

type UserAccessPolicy interface {
	CanViewProfileSection(ctx context.Context, viewerID, ownerID uint, section ProfileSection) (bool, error)
	CanViewActivity(ctx context.Context, viewerID, actorID uint, activityType ActivityType, visibility ActivityVisibility) (bool, error)
}

type userAccessPolicy struct {
	db *gorm.DB
}

func NewUserAccessPolicy(db *gorm.DB) UserAccessPolicy {
	return &userAccessPolicy{db: db}
}

func (p *userAccessPolicy) CanViewProfileSection(ctx context.Context, viewerID, ownerID uint, section ProfileSection) (bool, error) {
	if viewerID == ownerID {
		return true, nil
	}

	var isPrivate bool
	err := p.db.WithContext(ctx).Table("users").Select("is_private").Where("id = ? AND is_active = true AND deleted_at IS NULL", ownerID).Row().Scan(&isPrivate)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
			return false, ErrNotFound
		}
		return false, err
	}

	if !isPrivate {
		return true, nil
	}

	// If private, viewer must be an accepted follower
	var count int64
	err = p.db.WithContext(ctx).Table("user_follows").
		Where("follower_id = ? AND followee_id = ? AND status = 'accepted'", viewerID, ownerID).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (p *userAccessPolicy) CanViewActivity(ctx context.Context, viewerID, actorID uint, activityType ActivityType, visibility ActivityVisibility) (bool, error) {
	if viewerID == actorID {
		return true, nil
	}

	// 1. Check account visibility
	canViewProfile, err := p.CanViewProfileSection(ctx, viewerID, actorID, SectionProfile)
	if err != nil || !canViewProfile {
		return false, err
	}

	// 2. Check activity preference settings of actor
	var dbEnabled bool
	err = p.db.WithContext(ctx).Table("activity_privacy_settings").
		Select("enabled").
		Where("user_id = ? AND activity_type = ?", actorID, string(activityType)).
		Row().Scan(&dbEnabled)

	enabled := true
	if err != nil {
		// Fallback to default settings
		defaults := map[ActivityType]bool{
			ActivityReviewCreated:       true,
			ActivityReviewCommented:     true,
			ActivityReviewLiked:         false,
			ActivityMediaLiked:          false,
			ActivityWatchlistAdded:      false,
			ActivityWatchedAdded:        false,
			ActivityAchievementUnlocked: true,
			ActivityUserFollowed:        false,
		}
		if val, exists := defaults[activityType]; exists {
			enabled = val
		}
	} else {
		enabled = dbEnabled
	}

	if !enabled {
		return false, nil
	}

	// 3. Check per-activity visibility
	switch visibility {
	case VisibilityPrivate:
		return false, nil
	case VisibilityFollowers:
		var count int64
		err = p.db.WithContext(ctx).Table("user_follows").
			Where("follower_id = ? AND followee_id = ? AND status = 'accepted'", viewerID, actorID).
			Count(&count).Error
		if err != nil {
			return false, err
		}
		return count > 0, nil
	case VisibilityPublic:
		return true, nil
	case VisibilityDefault:
		// If account is private, default is followers-only. Otherwise public.
		var isPrivate bool
		err := p.db.WithContext(ctx).Table("users").Select("is_private").Where("id = ? AND is_active = true", actorID).Row().Scan(&isPrivate)
		if err != nil {
			return false, err
		}
		if isPrivate {
			var count int64
			err = p.db.WithContext(ctx).Table("user_follows").
				Where("follower_id = ? AND followee_id = ? AND status = 'accepted'", viewerID, actorID).
				Count(&count).Error
			if err != nil {
				return false, err
			}
			return count > 0, nil
		}
		return true, nil
	default:
		return true, nil
	}
}

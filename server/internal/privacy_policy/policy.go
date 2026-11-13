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
	CanViewList(ctx context.Context, viewerID, ownerID uint, visibility ListVisibility) (bool, error)
}

type userAccessPolicy struct {
	db *gorm.DB
}

func NewUserAccessPolicy(db *gorm.DB) UserAccessPolicy {
	return &userAccessPolicy{db: db}
}

func (p *userAccessPolicy) CanViewProfileSection(ctx context.Context, viewerID, ownerID uint, section ProfileSection) (bool, error) {
	if ownerID == 0 {
		return false, ErrNotFound
	}
	if viewerID != 0 && viewerID == ownerID {
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
	if viewerID == 0 {
		return false, nil
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
	if actorID == 0 {
		return false, ErrNotFound
	}
	if viewerID != 0 && viewerID == actorID {
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
		enabled = ActivityEnabledByDefault(activityType)
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
		if viewerID == 0 {
			return false, nil
		}
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
		err := p.db.WithContext(ctx).Table("users").Select("is_private").Where("id = ? AND is_active = true AND deleted_at IS NULL", actorID).Row().Scan(&isPrivate)
		if err != nil {
			return false, err
		}
		if isPrivate {
			if viewerID == 0 {
				return false, nil
			}
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

// CanViewList decides in order: the owner always may; then the owner's library section
// (inactive owner -> ErrNotFound, locked private profile -> false); then the list's own
// visibility, where "followers" needs an accepted follow.
func (p *userAccessPolicy) CanViewList(ctx context.Context, viewerID, ownerID uint, visibility ListVisibility) (bool, error) {
	if ownerID == 0 {
		return false, ErrNotFound
	}
	if viewerID != 0 && viewerID == ownerID {
		return true, nil
	}
	canView, err := p.CanViewProfileSection(ctx, viewerID, ownerID, SectionLibrary)
	if err != nil || !canView {
		return false, err
	}
	switch visibility {
	case ListPublic:
		return true, nil
	case ListFollowers:
		if viewerID == 0 {
			return false, nil
		}
		var count int64
		err = p.db.WithContext(ctx).Table("user_follows").
			Where("follower_id = ? AND followee_id = ? AND status = 'accepted'", viewerID, ownerID).
			Count(&count).Error
		return count > 0, err
	default:
		return false, nil
	}
}

// VisibleAuthorScope restricts a query to rows whose author (userIDColumn, e.g. "reviews.user_id")
// the viewer may see content from: the viewer themselves, or an active, non-deleted account that
// is public or followed by the viewer (accepted). It is the set form of CanViewProfileSection, so
// list endpoints filter in the same query instead of asking the policy once per row.
// viewerID 0 is a guest and matches no row through the first branch.
func VisibleAuthorScope(viewerID uint, userIDColumn string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`(`+userIDColumn+` = ? OR EXISTS (
			SELECT 1 FROM users pu
			WHERE pu.id = `+userIDColumn+` AND pu.is_active = true AND pu.deleted_at IS NULL
			AND (pu.is_private = false OR EXISTS (
				SELECT 1 FROM user_follows pf
				WHERE pf.follower_id = ? AND pf.followee_id = pu.id AND pf.status = 'accepted'))))`,
			viewerID, viewerID)
	}
}

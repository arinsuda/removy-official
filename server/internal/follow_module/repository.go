package follow_module

import (
	"database/sql"
	"errors"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrAlreadyFollowing = errors.New("already following")
	ErrNotFollowing     = errors.New("not following")
	ErrForbidden        = errors.New("forbidden")
	ErrNotFound         = errors.New("follow request not found")
	ErrUserNotFound     = errors.New("user not found")
)

type repository struct {
	db *gorm.DB
}

func newRepository(db *gorm.DB) *repository {
	return &repository{db: db}
}

func (r *repository) isPrivate(db *gorm.DB, userID uint) (bool, error) {
	var isPrivate bool
	err := db.
		Table("users").
		Select("is_private").
		Where("id = ? AND is_active = true AND deleted_at IS NULL", userID).
		Row().
		Scan(&isPrivate)
	if errors.Is(err, sql.ErrNoRows) {
		return false, ErrUserNotFound
	}
	if err != nil {
		return false, err
	}
	return isPrivate, nil
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}

func (r *repository) Follow(followerID, followeeID uint) (*UserFollow, error) {
	if followerID == followeeID {
		return nil, ErrForbidden
	}

	var follow UserFollow
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var existing UserFollow
		err := tx.
			Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
			First(&existing).Error
		if err == nil {
			return ErrAlreadyFollowing
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		isPrivate, err := r.isPrivate(tx, followeeID)
		if err != nil {
			return err
		}

		status := StatusAccepted
		if isPrivate {
			status = StatusPending
		}

		follow = UserFollow{
			FollowerID: followerID,
			FolloweeID: followeeID,
			Status:     status,
		}
		if err := tx.Create(&follow).Error; err != nil {
			if isUniqueViolation(err) {
				return ErrAlreadyFollowing
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &follow, nil
}

func (r *repository) Unfollow(followerID, followeeID uint) error {
	result := r.db.
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Delete(&UserFollow{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFollowing
	}
	return nil
}

func (r *repository) AcceptFollow(followerID, followeeID uint) error {
	result := r.db.
		Model(&UserFollow{}).
		Where("follower_id = ? AND followee_id = ? AND status = ?",
			followerID, followeeID, StatusPending).
		Update("status", StatusAccepted)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) RejectFollow(followerID, followeeID uint) error {
	result := r.db.
		Where("follower_id = ? AND followee_id = ? AND status = ?",
			followerID, followeeID, StatusPending).
		Delete(&UserFollow{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) GetStatus(followerID, followeeID uint) (*UserFollow, error) {
	var follow UserFollow
	err := r.db.
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Take(&follow).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &follow, nil
}

type listRow struct {
	ID          uint
	Username    string
	DisplayName *string
	AvatarURL   *string
	Status      string
}

func (r *repository) GetFollowers(userID uint) ([]listRow, error) {
	var rows []listRow
	err := r.db.
		Table("user_follows uf").
		Select("u.id, u.username, u.display_name, u.avatar_url, uf.status").
		Joins("JOIN users u ON u.id = uf.follower_id AND u.is_active = true AND u.deleted_at IS NULL").
		Where("uf.followee_id = ? AND uf.status = ?", userID, StatusAccepted).
		Order("uf.created_at DESC").
		Scan(&rows).Error
	return rows, err
}

func (r *repository) GetFollowing(userID uint) ([]listRow, error) {
	var rows []listRow
	err := r.db.
		Table("user_follows uf").
		Select("u.id, u.username, u.display_name, u.avatar_url, uf.status").
		Joins("JOIN users u ON u.id = uf.followee_id AND u.is_active = true AND u.deleted_at IS NULL").
		Where("uf.follower_id = ? AND uf.status = ?", userID, StatusAccepted).
		Order("uf.created_at DESC").
		Scan(&rows).Error
	return rows, err
}

func (r *repository) GetPendingRequests(userID uint) ([]listRow, error) {
	var rows []listRow
	err := r.db.
		Table("user_follows uf").
		Select("u.id, u.username, u.display_name, u.avatar_url, uf.status").
		Joins("JOIN users u ON u.id = uf.follower_id AND u.is_active = true AND u.deleted_at IS NULL").
		Where("uf.followee_id = ? AND uf.status = ?", userID, StatusPending).
		Order("uf.created_at DESC").
		Scan(&rows).Error
	return rows, err
}

func (r *repository) CountFollowers(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&UserFollow{}).
		Where("followee_id = ? AND status = ?", userID, StatusAccepted).
		Count(&count).Error
	return count, err
}

func (r *repository) CountFollowing(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&UserFollow{}).
		Where("follower_id = ? AND status = ?", userID, StatusAccepted).
		Count(&count).Error
	return count, err
}

package like_module

import (
	"errors"
	"github.com/arinsuda/movie-hub/internal/movie_module"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

var (
	ErrAlreadyLiked = errors.New("already liked")
	ErrNotLiked     = errors.New("not liked")
)

type repository struct {
	db *gorm.DB
}

func newRepository(db *gorm.DB) *repository {
	return &repository{db: db}
}

func (r *repository) Create(userID uint, mediaID int, mediaType movie_module.MediaType) error {
	like := MediaLike{UserID: userID, MediaID: mediaID, MediaType: mediaType}
	err := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&like).Error
	if err != nil && isDuplicateError(err) {
		return ErrAlreadyLiked
	}
	return err
}

func (r *repository) Delete(userID uint, mediaID int, mediaType movie_module.MediaType) error {
	result := r.db.Where(
		"user_id = ? AND media_id = ? AND media_type = ?",
		userID, mediaID, mediaType,
	).Delete(&MediaLike{})
	if result.RowsAffected == 0 {
		return ErrNotLiked
	}
	return result.Error
}

func (r *repository) FindByUser(userID uint, limit int) ([]MediaLike, error) {
	var likes []MediaLike
	q := r.db.Where("user_id = ?", userID).Order("created_at DESC, id DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&likes).Error
	return likes, err
}

func isDuplicateError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "UNIQUE constraint")
}

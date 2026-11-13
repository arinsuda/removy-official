package library_module

import (
	"errors"
	"fmt"
	"strings"

	"github.com/arinsuda/movie-hub/internal/movie_module"
	"gorm.io/gorm"
)

var (
	ErrItemNotFound     = errors.New("item not found")
	ErrForbidden        = errors.New("forbidden")
	ErrDuplicate        = errors.New("item already in list")
	ErrInvalidWatchedAt = errors.New("invalid watched_at")
	ErrInvalidMediaType = errors.New("media_type must be 'movie' or 'tv'")
	ErrInvalidListType  = errors.New("list_type must be 'watchlist', 'favorite', or 'watched'")
	ErrInvalidMediaID   = errors.New("invalid media_id")
)

type repository struct {
	db *gorm.DB
}

func newRepository(db *gorm.DB) *repository {
	return &repository{db: db}
}

func (r *repository) Create(item *LibraryItem) error {
	err := r.db.Create(item).Error
	if err != nil && isDuplicateError(err) {
		return ErrDuplicate
	}
	return err
}

func (r *repository) FindByUser(userID uint, listType *movie_module.ListType, mediaType *movie_module.MediaType, limit int) ([]LibraryItem, error) {
	query := r.db.Where("user_id = ?", userID)
	if listType != nil {
		query = query.Where("list_type = ?", *listType)
	}
	if mediaType != nil {
		query = query.Where("media_type = ?", *mediaType)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	var items []LibraryItem
	err := query.Order("created_at DESC, id DESC").Find(&items).Error
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	uniqueItems := make([]LibraryItem, 0, len(items))
	for _, item := range items {
		key := fmt.Sprintf("%d_%s_%s", item.MediaID, item.MediaType, item.ListType)
		if !seen[key] {
			seen[key] = true
			uniqueItems = append(uniqueItems, item)
		}
	}
	return uniqueItems, nil
}

func (r *repository) FindOne(id, userID uint) (*LibraryItem, error) {
	var item LibraryItem
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrItemNotFound
	}
	return &item, err
}

// ดึงสถานะของ media นึงว่าอยู่ใน list ไหนบ้าง
func (r *repository) FindMediaStatus(userID uint, mediaID int, mediaType movie_module.MediaType) ([]LibraryItem, error) {
	var items []LibraryItem
	err := r.db.Where("user_id = ? AND media_id = ? AND media_type = ?", userID, mediaID, mediaType).
		Find(&items).Error
	return items, err
}

func (r *repository) Update(id uint, updates map[string]any) error {
	result := r.db.Model(&LibraryItem{}).Where("id = ?", id).Updates(updates)
	if result.RowsAffected == 0 {
		return ErrItemNotFound
	}
	return result.Error
}

func (r *repository) Delete(id uint) error {
	result := r.db.Delete(&LibraryItem{}, id)
	if result.RowsAffected == 0 {
		return ErrItemNotFound
	}
	return result.Error
}

func isDuplicateError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "UNIQUE constraint")
}

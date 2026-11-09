package bmol_module

import (
	"github.com/arinsuda/movie-hub/internal/movie_module"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func newRepository(db *gorm.DB) *repository {
	return &repository{db: db}
}

// Update applies only the provided fields.
func (r *repository) Update(id uint, rank *int, note *string) error {
	updates := map[string]any{}
	if rank != nil {
		updates["rank"] = *rank
	}
	if note != nil {
		updates["note"] = *note
	}
	return r.db.Model(&BMOLItem{}).Where("id = ?", id).Updates(updates).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&BMOLItem{}, id).Error
}

func (r *repository) FindOne(id uint) (*BMOLItem, error) {
	var item BMOLItem
	err := r.db.First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *repository) FindByUser(userID uint, mediaType *movie_module.MediaType) ([]BMOLItem, error) {
	var items []BMOLItem
	query := r.db.Where("user_id = ? AND deleted_at IS NULL", userID)
	if mediaType != nil && *mediaType != "" {
		query = query.Where("media_type = ?", *mediaType)
	}
	err := query.Order("rank asc").Order("created_at asc").Find(&items).Error
	return items, err
}

// FindAnyByMedia includes soft-deleted rows: the unique index covers them too.
func findAnyByMedia(tx *gorm.DB, userID uint, mediaID int, mediaType movie_module.MediaType) (*BMOLItem, error) {
	var item BMOLItem
	err := tx.Unscoped().Where("user_id = ? AND media_id = ? AND media_type = ?", userID, mediaID, mediaType).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

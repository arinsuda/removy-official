package reading_progress_module

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type repository struct {
	db *gorm.DB
}

func newRepository(db *gorm.DB) *repository {
	return &repository{db: db}
}

// Find returns the user's row for the title, or nil when there is none.
func (r *repository) Find(ctx context.Context, userID uint, t Target) (*ReadingProgress, error) {
	var row ReadingProgress
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND media_id = ? AND media_type = ?", userID, t.MediaID, t.MediaType).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// Upsert is one INSERT ... ON CONFLICT statement on the unique (user, media) index, so concurrent saves
// of the same title end as one row. RETURNING fills row with the stored values.
func (r *repository) Upsert(ctx context.Context, row *ReadingProgress) error {
	return r.db.WithContext(ctx).Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "media_id"}, {Name: "media_type"}},
			DoUpdates: clause.AssignmentColumns([]string{"chapter", "volume", "updated_at"}),
		},
		clause.Returning{},
	).Create(row).Error
}

// Delete removes the user's row for the title. Deleting nothing is not an error (idempotent).
func (r *repository) Delete(ctx context.Context, userID uint, t Target) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND media_id = ? AND media_type = ?", userID, t.MediaID, t.MediaType).
		Delete(&ReadingProgress{}).Error
}

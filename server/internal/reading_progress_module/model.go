package reading_progress_module

import (
	"time"

	"github.com/arinsuda/movie-hub/internal/shared"
	users "github.com/arinsuda/movie-hub/internal/user_module"
)

// ReadingProgress is the chapter and volume one user reached in one manga. There is no soft delete:
// DELETE removes the row. The unique index on (user, media) backs the single-row upsert.
type ReadingProgress struct {
	ID        uint             `gorm:"primarykey;autoIncrement"`
	UserID    uint             `gorm:"not null;uniqueIndex:idx_reading_progress_user_media"`
	User      users.User       `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	MediaID   int              `gorm:"not null;uniqueIndex:idx_reading_progress_user_media"`
	MediaType shared.MediaType `gorm:"type:varchar(10);not null;uniqueIndex:idx_reading_progress_user_media"`
	Chapter   int              `gorm:"not null;default:0"`
	Volume    int              `gorm:"not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName keeps the table name aligned with the chk_reading_progress_* constraints.
func (ReadingProgress) TableName() string {
	return "reading_progress"
}

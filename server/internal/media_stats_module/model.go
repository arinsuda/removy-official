package media_stats_module

import (
	"time"

	"github.com/arinsuda/movie-hub/internal/movie_module"
	users "github.com/arinsuda/movie-hub/internal/user_module"
	"gorm.io/gorm"
)

type MediaStat struct {
	gorm.Model
	MediaID   int                    `gorm:"not null;uniqueIndex:idx_media"`
	MediaType movie_module.MediaType `gorm:"not null;uniqueIndex:idx_media"`
	LikeCount int                    `gorm:"default:0"`
	ViewCount int                    `gorm:"default:0"`
}

// MediaViewDay records that a user viewed a title on one Bangkok calendar day. A repeat view that
// day adds no row. Monthly voting reads it; rows older than MediaViewRetention are purged.
type MediaViewDay struct {
	UserID    uint       `gorm:"primaryKey;not null"`
	User      users.User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	MediaType string     `gorm:"primaryKey;type:varchar(10);not null;index:idx_media_view_days_type_day,priority:1"`
	MediaID   int        `gorm:"primaryKey;not null"`
	// The single-column day index serves the purge (DELETE ... WHERE day < cutoff).
	Day time.Time `gorm:"primaryKey;type:date;not null;index:idx_media_view_days_type_day,priority:2;index:idx_media_view_days_day"`
}

func (MediaViewDay) TableName() string { return "media_view_days" }

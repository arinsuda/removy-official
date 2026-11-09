package bmol_module

import (
	"github.com/arinsuda/movie-hub/internal/movie_module"
	users "github.com/arinsuda/movie-hub/internal/user_module"
	"gorm.io/gorm"
)

type BMOLItem struct {
	gorm.Model
	UserID    uint                   `gorm:"not null;index;uniqueIndex:idx_user_media"`
	User      users.User             `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	MediaID   int                    `gorm:"not null;uniqueIndex:idx_user_media"`
	MediaType movie_module.MediaType `gorm:"type:varchar(10);not null;uniqueIndex:idx_user_media"`
	Rank      int                    `gorm:"not null"`
	Note      string                 `gorm:"type:varchar(280);not null;default:''"`
}

package like_module

import (
	"github.com/arinsuda/movie-hub/internal/movie_module"
	users "github.com/arinsuda/movie-hub/internal/user_module"
	"gorm.io/gorm"
)

type MediaLike struct {
	gorm.Model
	UserID    uint                   `gorm:"not null;index;uniqueIndex:idx_user_media_like"`
	User      users.User             `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	MediaID   int                    `gorm:"not null;index;uniqueIndex:idx_user_media_like"`
	MediaType movie_module.MediaType `gorm:"not null;uniqueIndex:idx_user_media_like"`
}

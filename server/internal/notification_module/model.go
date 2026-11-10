package notification_module

import (
	"time"

	users "github.com/arinsuda/movie-hub/internal/user_module"
	"gorm.io/gorm"
)

type NotificationType string

const (
	NotifWelcome NotificationType = "welcome"

	NotifFollowedYou NotificationType = "followed_you"

	NotifMovieNowPlaying NotificationType = "movie_now_playing"

	NotifFollowingReviewed       NotificationType = "following_reviewed"
	NotifFollowingLikedReview    NotificationType = "following_liked_review"
	NotifFollowingAddedWatchlist NotificationType = "following_added_watchlist"
	NotifFollowingAddedWatched   NotificationType = "following_added_watched"
	NotifFollowingMarkedHelpful  NotificationType = "following_marked_helpful"
	NotifFollowingCommented      NotificationType = "following_commented"
	NotifFollowingLikedMedia     NotificationType = "following_liked_media"

	NotifReviewLiked         NotificationType = "review_liked"
	NotifReviewCommented     NotificationType = "review_commented"
	NotifReviewMarkedHelpful NotificationType = "review_marked_helpful"

	NotifAchievementUnlocked NotificationType = "achievement_unlocked"

	NotifEmailVerified   NotificationType = "email_verified"
	NotifPasswordChanged NotificationType = "password_changed"

	NotifFollowRequested NotificationType = "follow_requested"
	NotifFollowAccepted  NotificationType = "follow_accepted"
)

type NotificationCategory string

const (
	CategorySystem      NotificationCategory = "system"
	CategorySocial      NotificationCategory = "social"
	CategoryMedia       NotificationCategory = "media"
	CategoryAchievement NotificationCategory = "achievement"
)

var categoryByType = map[NotificationType]NotificationCategory{
	NotifWelcome:                 CategorySystem,
	NotifFollowedYou:             CategorySocial,
	NotifFollowRequested:         CategorySocial,
	NotifFollowAccepted:          CategorySocial,
	NotifFollowingReviewed:       CategorySocial,
	NotifFollowingLikedReview:    CategorySocial,
	NotifFollowingAddedWatchlist: CategorySocial,
	NotifFollowingAddedWatched:   CategorySocial,
	NotifFollowingMarkedHelpful:  CategorySocial,
	NotifFollowingCommented:      CategorySocial,
	NotifFollowingLikedMedia:     CategorySocial,
	NotifReviewLiked:             CategorySocial,
	NotifReviewCommented:         CategorySocial,
	NotifReviewMarkedHelpful:     CategorySocial,

	NotifMovieNowPlaying: CategoryMedia,

	NotifAchievementUnlocked: CategoryAchievement,

	NotifEmailVerified:   CategorySystem,
	NotifPasswordChanged: CategorySystem,
}

type Notification struct {
	gorm.Model
	UserID    uint                 `gorm:"not null;index"`
	User      users.User           `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ActorID   *uint                `gorm:"index"`
	Type      NotificationType     `gorm:"type:varchar(50);not null;index"`
	Category  NotificationCategory `gorm:"type:varchar(20);not null;index"`
	TargetID  *uint
	TargetRef *string `gorm:"type:varchar(50)"`
	// Message is empty on new rows (clients build text from Type + Params); kept for legacy rows.
	Message string `gorm:"type:text;not null"`
	// -:migration keeps AutoMigrate from adding it; migration 20261110_notification_params does.
	Params *NotificationParams `gorm:"-:migration;type:jsonb;column:params"`
	IsRead bool                `gorm:"default:false;index"`
	ReadAt *time.Time
}

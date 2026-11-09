package feed_module

import (
	"time"

	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	users "github.com/arinsuda/movie-hub/internal/user_module"
	"gorm.io/gorm"
)

type ActivityType = privacy_policy.ActivityType
type ActivityVisibility = privacy_policy.ActivityVisibility

const (
	ActivityReviewCreated       ActivityType = privacy_policy.ActivityReviewCreated
	ActivityReviewCommented     ActivityType = privacy_policy.ActivityReviewCommented
	ActivityReviewLiked         ActivityType = privacy_policy.ActivityReviewLiked
	ActivityMediaLiked          ActivityType = privacy_policy.ActivityMediaLiked
	ActivityWatchlistAdded      ActivityType = privacy_policy.ActivityWatchlistAdded
	ActivityWatchedAdded        ActivityType = privacy_policy.ActivityWatchedAdded
	ActivityAchievementUnlocked ActivityType = privacy_policy.ActivityAchievementUnlocked
	ActivityUserFollowed        ActivityType = privacy_policy.ActivityUserFollowed
	ActivityWatchLogCreated     ActivityType = privacy_policy.ActivityWatchLogCreated
	ActivityRewatched           ActivityType = privacy_policy.ActivityRewatched
)

var AllActivityTypes = []ActivityType{
	ActivityReviewCreated,
	ActivityReviewCommented,
	ActivityReviewLiked,
	ActivityMediaLiked,
	ActivityWatchlistAdded,
	ActivityWatchedAdded,
	ActivityAchievementUnlocked,
	ActivityUserFollowed,
	ActivityWatchLogCreated,
	ActivityRewatched,
}

type ActivityEvent struct {
	ID uint `gorm:"primarykey;autoIncrement"`

	ActorID uint       `gorm:"not null;index"`
	Actor   users.User `gorm:"foreignKey:ActorID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Type ActivityType `gorm:"type:varchar(40);not null;index"`

	MediaID   *int    `gorm:"index"`
	MediaType *string `gorm:"type:varchar(10)"`

	ReviewID      *uint `gorm:"index"`
	CommentID     *uint `gorm:"index"`
	AchievementID *uint `gorm:"index"`
	LibraryItemID *uint `gorm:"index"`
	WatchLogID    *uint `gorm:"index"`

	TargetUserID *uint       `gorm:"index"`
	TargetUser   *users.User `gorm:"foreignKey:TargetUserID;references:ID;constraint:OnDelete:SET NULL;"`
	Message      string      `gorm:"type:text"`

	Visibility ActivityVisibility `gorm:"type:varchar(20);not null;default:'default';index"`

	IsVisible bool `gorm:"default:true;index"`

	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ActivityPrivacySetting struct {
	ID uint `gorm:"primarykey;autoIncrement"`

	UserID uint       `gorm:"not null;uniqueIndex:idx_user_activity_type"`
	User   users.User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	ActivityType ActivityType `gorm:"type:varchar(40);not null;uniqueIndex:idx_user_activity_type"`

	Enabled bool `gorm:"default:true"`

	UpdatedAt time.Time
	CreatedAt time.Time
}

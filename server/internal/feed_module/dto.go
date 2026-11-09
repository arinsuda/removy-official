package feed_module

import (
	"time"

	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/arinsuda/movie-hub/internal/shared"
)

type ActivityPayload struct {
	MediaID   *int
	MediaType *string

	ReviewID      *uint
	CommentID     *uint
	AchievementID *uint
	LibraryItemID *uint
	WatchLogID    *uint
	TargetUserID  *uint

	Message string
}

type PaginationQuery struct {
	Page     int    `query:"page"`
	Limit    int    `query:"limit"`
	Category string `query:"category"`
}

func (p *PaginationQuery) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 || p.Limit > 100 {
		p.Limit = 20
	}
}

func (p *PaginationQuery) Offset() int {
	return (p.Page - 1) * p.Limit
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func newPaginationMeta(page, limit int, total int64) PaginationMeta {
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}
	return PaginationMeta{Page: page, Limit: limit, Total: total, TotalPages: totalPages}
}

type ActorSummary struct {
	ID          uint    `json:"id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
}

type FeedItemResponse struct {
	ID uint `json:"id"`

	Type privacy_policy.ActivityType `json:"type"`

	Actor ActorSummary `json:"actor"`

	Media *shared.MediaSummary `json:"media,omitempty"`

	ReviewID      *uint `json:"review_id,omitempty"`
	CommentID     *uint `json:"comment_id,omitempty"`
	AchievementID *uint `json:"achievement_id,omitempty"`
	LibraryItemID *uint `json:"library_item_id,omitempty"`
	WatchLogID    *uint `json:"watch_log_id,omitempty"`

	TargetUser *ActorSummary `json:"target_user,omitempty"`

	Message string `json:"message"`

	Visibility privacy_policy.ActivityVisibility `json:"visibility"`

	CreatedAt time.Time `json:"created_at"`
}

type FeedListResponse struct {
	Items      []FeedItemResponse `json:"items"`
	Pagination PaginationMeta     `json:"pagination"`
}

type UpdateActivityVisibilityRequest struct {
	Visibility privacy_policy.ActivityVisibility `json:"visibility"`
}

type ActivitySettingsResponse struct {
	ReviewCreated       bool `json:"review_created"`
	ReviewCommented     bool `json:"review_commented"`
	ReviewLiked         bool `json:"review_liked"`
	MediaLiked          bool `json:"media_liked"`
	WatchlistAdded      bool `json:"watchlist_added"`
	WatchedAdded        bool `json:"watched_added"`
	AchievementUnlocked bool `json:"achievement_unlocked"`
	UserFollowed        bool `json:"user_followed"`
	WatchLogCreated     bool `json:"watch_log_created"`
	Rewatched           bool `json:"rewatched"`
}

type UpdateActivitySettingsRequest struct {
	ReviewCreated       *bool `json:"review_created"`
	ReviewCommented     *bool `json:"review_commented"`
	ReviewLiked         *bool `json:"review_liked"`
	MediaLiked          *bool `json:"media_liked"`
	WatchlistAdded      *bool `json:"watchlist_added"`
	WatchedAdded        *bool `json:"watched_added"`
	AchievementUnlocked *bool `json:"achievement_unlocked"`
	UserFollowed        *bool `json:"user_followed"`
	WatchLogCreated     *bool `json:"watch_log_created"`
	Rewatched           *bool `json:"rewatched"`
}

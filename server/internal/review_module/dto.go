package review_module

import (
	tmdb "github.com/arinsuda/movie-hub/internal/tmdb_module"
	users "github.com/arinsuda/movie-hub/internal/user_module"
	"time"
)

// ── Review ────────────────────────────────────────────────────────

type CreateReviewRequest struct {
	MediaID    int     `json:"media_id"`
	MediaType  string  `json:"media_type"` // "movie" | "tv"
	Rating     float32 `json:"rating"`     // 0.5 - 5.0 (increments of 0.5)
	Body       string  `json:"body"`
	Visibility string  `json:"visibility"`
	IsPublic   *bool   `json:"is_public,omitempty"` // DEPRECATED: expand-and-contract compat
}

type UpdateReviewRequest struct {
	Rating     *float32 `json:"rating"`
	Body       *string  `json:"body"`
	Visibility *string  `json:"visibility"`
	IsPublic   *bool    `json:"is_public,omitempty"` // DEPRECATED: expand-and-contract compat
}

type ReviewResponse struct {
	ID             uint                      `json:"id"`
	User           users.UserSummaryResponse `json:"user"`
	Media          tmdb.Media                `json:"media"`
	Rating         float32                   `json:"rating"`
	Body           string                    `json:"body"`
	Visibility     string                    `json:"visibility"`
	LikeCount      int                       `json:"like_count"`
	CommentCount   int                       `json:"comment_count"`
	IsLiked        bool                      `json:"is_liked"`
	HelpfulCount   int                       `json:"helpful_count"`    // NEW
	IsHelpfulVoted bool                      `json:"is_helpful_voted"` // NEW
	CreatedAt      time.Time                 `json:"created_at"`
	UpdatedAt      time.Time                 `json:"updated_at"`
}

// ── In-app Rating (aggregate) ─────────────────────────────────────

// RatingBucket holds the count of public reviews for one 0.5-step rating (0.5..5.0).
type RatingBucket struct {
	Rating float32 `json:"rating"`
	Count  int     `json:"count"`
}

// RatingResponse คือ aggregate rating ที่คำนวณจาก reviews ภายใน app เท่านั้น
// ไม่เกี่ยวกับ TMDB
type RatingResponse struct {
	MediaID       int            `json:"media_id"`
	MediaType     string         `json:"media_type"`
	AverageRating float32        `json:"average_rating"` // 0.0 ถ้าไม่มี review
	ReviewCount   int            `json:"review_count"`
	HasRating     bool           `json:"has_rating"` // false = ยังไม่มีใครรีวิว
	Distribution  []RatingBucket `json:"distribution"`
}

// ── Comment ───────────────────────────────────────────────────────

type CreateCommentRequest struct {
	Body string `json:"body"`
}

type UpdateCommentRequest struct {
	Body string `json:"body"`
}

type CommentResponse struct {
	ID        uint      `json:"id"`
	ReviewID  uint      `json:"review_id"`
	UserID    uint      `json:"user_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Media and user review list paging, sorting, and filtering constants.
const (
	DefaultMediaReviewsPageSize = 20
	MaxMediaReviewsPageSize     = 50

	ReviewSortNewest  = "newest"
	ReviewSortOldest  = "oldest"
	ReviewSortHighest = "highest"
	ReviewSortLowest  = "lowest"

	ReviewRatingFilterAll      = "all"
	ReviewRatingFilterHigh     = "high"
	ReviewRatingFilterLow      = "low"
	ReviewRatingFilterWithText = "withText"

	MaxReviewSearchLength = 100
	// MaxReviewsPage bounds the page number so offset arithmetic cannot overflow.
	MaxReviewsPage = 10000

	// Rating chip thresholds: "high" is rating >= HighRatingMin, "low" is rating <= LowRatingMax.
	HighRatingMin = 4.0
	LowRatingMax  = 2.5
)

// ReviewFilterCounts are the chip counts of a user's reviews, ignoring the active rating filter.
type ReviewFilterCounts struct {
	All      int `json:"all"`
	High     int `json:"high"`
	Low      int `json:"low"`
	WithText int `json:"withText"`
}

// UserReviewsSummary describes the whole privacy/visibility/month scope of a user's review list,
// independent of rating filter, search and paging.
type UserReviewsSummary struct {
	Count         int                `json:"count"`
	AverageRating float32            `json:"average_rating"` // 0 when there are no reviews
	FilterCounts  ReviewFilterCounts `json:"filter_counts"`
}

type Pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
	HasMore  bool  `json:"has_more"`
}

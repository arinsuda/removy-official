package admin_module

import (
	"github.com/arinsuda/movie-hub/internal/account_erasure"
	"gorm.io/datatypes"
)

type GrowthStatus string

const (
	GrowthStatusGrowth             GrowthStatus = "growth"
	GrowthStatusDecline            GrowthStatus = "decline"
	GrowthStatusNoChange           GrowthStatus = "no_change"
	GrowthStatusNoPreviousBaseline GrowthStatus = "no_previous_baseline"
)

type OverviewStats struct {
	TotalRegisteredUsers       int64        `json:"total_registered_users"`
	ActiveUsersCount           int64        `json:"active_users_count"`
	InactiveUsersCount         int64        `json:"inactive_users_count"`
	CurrentMonthRegistrations  int64        `json:"current_month_registrations"`
	PreviousMonthRegistrations int64        `json:"previous_month_registrations"`
	AbsoluteGrowth             int64        `json:"absolute_growth"`
	GrowthPercentage           *float64     `json:"growth_percentage"`
	GrowthStatus               GrowthStatus `json:"growth_status"`
	UniqueOnlineUsers          int          `json:"unique_online_users"`
	TotalActivityEvents        int64        `json:"total_activity_events"`
	ActivityEventsToday        int64        `json:"activity_events_today"`
	DauToday                   int64        `json:"dau_today"`
	Wau7d                      int64        `json:"wau_7d"`
	Mau30d                     int64        `json:"mau_30d"`
	TotalReviews               int64        `json:"total_reviews"`
	TotalMediaLikes            int64        `json:"total_media_likes"`
}

type GrowthPoint struct {
	Month     string `json:"month"`
	UserCount int64  `json:"user_count"`
}

type AdminUserRow struct {
	ID          uint    `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
	Role        string  `json:"role"`
	IsActive    bool    `json:"is_active"`
	CreatedAt   string  `json:"created_at"`
	DeletedAt   *string `json:"deleted_at"`
	// ErasureDueAt is set while the user's self-requested erasure is pending.
	ErasureDueAt *string `json:"erasure_due_at"`
	ReviewCount  int     `json:"review_count"`
}

type AdminReviewRow struct {
	ID         uint    `json:"id"`
	UserID     uint    `json:"user_id"`
	Username   string  `json:"username"`
	MediaID    int     `json:"media_id"`
	MediaType  string  `json:"media_type"`
	MediaTitle string  `json:"media_title,omitempty"`
	PosterURL  string  `json:"poster_url,omitempty"`
	Rating     float32 `json:"rating"`
	Body       string  `json:"body"`
	IsPublic   bool    `json:"is_public"`
	LikeCount  int     `json:"like_count"`
	CreatedAt  string  `json:"created_at"`
}

type AdminAuditLogRow struct {
	ID            uint           `json:"id"`
	AdminID       *uint          `json:"admin_id"`
	AdminUsername *string        `json:"admin_username"`
	Action        AuditAction    `json:"action"`
	TargetType    string         `json:"target_type"`
	TargetID      uint           `json:"target_id"`
	Reason        *string        `json:"reason"`
	MetaData      datatypes.JSON `json:"meta_data"`
	CreatedAt     string         `json:"created_at"`
}

type UpdateRoleRequest struct {
	Role   string  `json:"role"`
	Reason *string `json:"reason"`
}

type UpdateStatusRequest struct {
	IsActive bool    `json:"is_active"`
	Reason   *string `json:"reason"`
}

type DeleteUserRequest struct {
	Reason *string `json:"reason"`
}

// Storage cleanup states recorded in the USER_ERASED audit meta under auditCleanupKey.
const (
	auditCleanupKey = "storage_cleanup"
	CleanupPending  = "pending"
	CleanupDone     = "done"
)

// ErasureResult is what the repository hands the service after the commit.
type ErasureResult struct {
	account_erasure.Result
	AuditID uint
}

// EraseUserRequest confirms a permanent erasure by typing the username exactly.
type EraseUserRequest struct {
	ConfirmUsername string  `json:"confirm_username"`
	Reason          *string `json:"reason"`
}

type DeleteReviewRequest struct {
	Reason *string `json:"reason"`
}

// UserStatusDeleted lists soft-deleted users (the first step before erasure).
const UserStatusDeleted = "deleted"

type UserFilter struct {
	Page      int    `query:"page"`
	Limit     int    `query:"limit"`
	Search    string `query:"search"`
	Role      string `query:"role"`
	Status    string `query:"status"`
	SortBy    string `query:"sort_by"`
	SortOrder string `query:"sort_order"`
}

func (f *UserFilter) Normalize() {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	switch f.SortBy {
	case "username", "email", "review_count", "created_at":
	default:
		f.SortBy = "created_at"
	}
	if f.SortOrder != "asc" && f.SortOrder != "desc" {
		f.SortOrder = "desc"
	}
}

func (f *UserFilter) Offset() int {
	return (f.Page - 1) * f.Limit
}

type ReviewFilter struct {
	Page       int      `query:"page"`
	Limit      int      `query:"limit"`
	Search     string   `query:"search"`
	MediaType  string   `query:"media_type"`
	Visibility string   `query:"visibility"`
	MinRating  *float32 `query:"min_rating"`
	MaxRating  *float32 `query:"max_rating"`
	SortBy     string   `query:"sort_by"`
	SortOrder  string   `query:"sort_order"`
}

func (f *ReviewFilter) Normalize() {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	switch f.SortBy {
	case "rating", "like_count", "created_at":
	default:
		f.SortBy = "created_at"
	}
	if f.SortOrder != "asc" && f.SortOrder != "desc" {
		f.SortOrder = "desc"
	}
}

func (f *ReviewFilter) Offset() int {
	return (f.Page - 1) * f.Limit
}

type AuditLogFilter struct {
	Page       int    `query:"page"`
	Limit      int    `query:"limit"`
	Action     string `query:"action"`
	TargetType string `query:"target_type"`
}

func (f *AuditLogFilter) Normalize() {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
}

func (f *AuditLogFilter) Offset() int {
	return (f.Page - 1) * f.Limit
}

type PaginatedResponse[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

func NewPaginatedResponse[T any](items []T, total int64, page, limit int) PaginatedResponse[T] {
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}
	if items == nil {
		items = []T{}
	}
	return PaginatedResponse[T]{
		Items:      items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}

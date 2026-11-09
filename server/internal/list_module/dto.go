package list_module

import (
	"time"

	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/arinsuda/movie-hub/internal/shared"
)

type CreateListRequest struct {
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	Visibility  privacy_policy.ListVisibility `json:"visibility"` // empty -> public
}

// UpdateListRequest changes only the fields that are present.
// CoverItemID 0 clears the cover (the poster collage is used again).
type UpdateListRequest struct {
	Name        *string                        `json:"name"`
	Description *string                        `json:"description"`
	Visibility  *privacy_policy.ListVisibility `json:"visibility"`
	CoverItemID *uint                          `json:"cover_item_id"`
}

type AddItemRequest struct {
	MediaID   int              `json:"media_id"`
	MediaType shared.MediaType `json:"media_type"`
}

type ReorderRequest struct {
	ItemIDs []uint `json:"item_ids"`
}

// ListSummary is a list without its items. CoverPoster is the picked item's poster;
// CollagePosters (up to CollagePosterCount) is the fallback when no item is picked.
type ListSummary struct {
	ID             uint                          `json:"id"`
	UserID         uint                          `json:"user_id"`
	Name           string                        `json:"name"`
	Description    string                        `json:"description"`
	Visibility     privacy_policy.ListVisibility `json:"visibility"`
	ItemCount      int                           `json:"item_count"`
	CoverItemID    *uint                         `json:"cover_item_id"`
	CoverPoster    string                        `json:"cover_poster,omitempty"`
	CollagePosters []string                      `json:"collage_posters"`
	ContainsMedia  *bool                         `json:"contains_media,omitempty"`
	CreatedAt      time.Time                     `json:"created_at"`
	UpdatedAt      time.Time                     `json:"updated_at"`
}

type ListItemResponse struct {
	ID        uint                `json:"id"`
	Position  int                 `json:"position"`
	Media     shared.MediaSummary `json:"media"`
	MediaType shared.MediaType    `json:"media_type"`
	AddedAt   time.Time           `json:"added_at"`
}

type Pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
	HasMore  bool  `json:"has_more"`
}

type ListDetailResponse struct {
	List       ListSummary        `json:"list"`
	Items      []ListItemResponse `json:"items"`
	Pagination Pagination         `json:"pagination"`
}

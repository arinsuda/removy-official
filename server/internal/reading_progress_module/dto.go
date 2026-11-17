package reading_progress_module

import (
	"time"

	"github.com/arinsuda/movie-hub/internal/shared"
)

// UpsertRequest is a full replacement: both fields are always sent (a missing one reads as 0).
type UpsertRequest struct {
	Chapter int `json:"chapter"`
	Volume  int `json:"volume"`
}

// ProgressResponse is one user's progress in one manga.
type ProgressResponse struct {
	MediaID   int              `json:"media_id"`
	MediaType shared.MediaType `json:"media_type"`
	Chapter   int              `json:"chapter"`
	Volume    int              `json:"volume"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// Target names one title: a manga (AniList id) read or written for the signed-in user.
type Target struct {
	MediaID   int
	MediaType shared.MediaType
}

// toResponse maps a stored row; nil (no progress yet) stays nil so the JSON reads null.
func toResponse(p *ReadingProgress) *ProgressResponse {
	if p == nil {
		return nil
	}
	return &ProgressResponse{
		MediaID:   p.MediaID,
		MediaType: p.MediaType,
		Chapter:   p.Chapter,
		Volume:    p.Volume,
		UpdatedAt: p.UpdatedAt,
	}
}

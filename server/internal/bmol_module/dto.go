package bmol_module

import (
	"time"

	"github.com/arinsuda/movie-hub/internal/movie_module"
	shared "github.com/arinsuda/movie-hub/internal/shared"
)

// Validation is manual (see validate.go); the server has no struct validator.
type CreateBMOLRequest struct {
	MediaID   int                    `json:"media_id"`
	MediaType movie_module.MediaType `json:"media_type"`
	Rank      int                    `json:"rank"`
}

// UpdateBMOLRequest needs at least one of rank or note.
type UpdateBMOLRequest struct {
	Rank *int    `json:"rank"`
	Note *string `json:"note"`
}

type BMOLItemResponse struct {
	ID        uint                   `json:"id"`
	Media     shared.MediaSummary    `json:"media"`
	MediaType movie_module.MediaType `json:"media_type"`
	Rank      int                    `json:"rank"`
	Note      string                 `json:"note"`
	CreatedAt time.Time              `json:"created_at"`
}

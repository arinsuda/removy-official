package reading_progress_module

import (
	"errors"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/shared"
)

var (
	ErrInvalidMediaType = errors.New("reading progress is kept for manga only")
	ErrInvalidID        = errors.New("media id must be a positive integer")
	ErrOutOfRange       = errors.New("chapter or volume is out of range")
	ErrInvalidInput     = errors.New("chapter or volume must be at least 1")
)

// validateTarget accepts manga only: reading progress is not kept for movies or TV.
func validateTarget(t Target) error {
	if t.MediaType != shared.MediaTypeManga {
		return ErrInvalidMediaType
	}
	if t.MediaID <= 0 {
		return ErrInvalidID
	}
	return nil
}

// validateProgress checks the caps first, then rejects an entry that records nothing.
func validateProgress(req UpsertRequest) error {
	if req.Chapter < 0 || req.Chapter > config.ReadingProgressMaxChapter ||
		req.Volume < 0 || req.Volume > config.ReadingProgressMaxVolume {
		return ErrOutOfRange
	}
	if req.Chapter == 0 && req.Volume == 0 {
		return ErrInvalidInput
	}
	return nil
}

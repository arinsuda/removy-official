package reading_progress_module

import (
	"context"
	"time"

	"github.com/arinsuda/movie-hub/internal/shared"
	"gorm.io/gorm"
)

type Service struct {
	repo *repository
}

func NewService(db *gorm.DB) *Service {
	return &Service{repo: newRepository(db)}
}

// Get returns the viewer's progress, or nil when there is none. It is not age-gated: a user reads back
// only their own entry, and adult titles can only be written through Upsert, which is gated.
func (s *Service) Get(ctx context.Context, userID uint, t Target) (*ProgressResponse, error) {
	if err := validateTarget(t); err != nil {
		return nil, err
	}
	row, err := s.repo.Find(ctx, userID, t)
	if err != nil {
		return nil, err
	}
	return toResponse(row), nil
}

// Upsert replaces the viewer's progress for one manga. The age gate runs before any write.
func (s *Service) Upsert(ctx context.Context, userID uint, t Target, req UpsertRequest) (*ProgressResponse, error) {
	if err := validateTarget(t); err != nil {
		return nil, err
	}
	if err := validateProgress(req); err != nil {
		return nil, err
	}
	if err := shared.CheckMediaAccess(ctx, userID, t.MediaID, t.MediaType); err != nil {
		return nil, err
	}
	row := &ReadingProgress{
		UserID:    userID,
		MediaID:   t.MediaID,
		MediaType: t.MediaType,
		Chapter:   req.Chapter,
		Volume:    req.Volume,
		UpdatedAt: time.Now(),
	}
	if err := s.repo.Upsert(ctx, row); err != nil {
		return nil, err
	}
	return toResponse(row), nil
}

// Delete clears the viewer's progress for one manga. Missing progress is not an error.
func (s *Service) Delete(ctx context.Context, userID uint, t Target) error {
	if err := validateTarget(t); err != nil {
		return err
	}
	return s.repo.Delete(ctx, userID, t)
}

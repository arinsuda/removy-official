package bmol_module

import (
	"context"
	"errors"
	"time"

	"github.com/arinsuda/movie-hub/internal/movie_module"
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	shared "github.com/arinsuda/movie-hub/internal/shared"
	tmdbmodule "github.com/arinsuda/movie-hub/internal/tmdb_module"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

const tmdbConcurrency = 6

var (
	ErrDuplicate        = errors.New("media item is already ranked in your BOL")
	ErrForbidden        = errors.New("you do not have permission to modify this item")
	ErrInvalidMediaType = errors.New("invalid media type, must be movie or tv")
	ErrNotFound         = errors.New("item not found")
)

type Service struct {
	repo   *repository
	db     *gorm.DB
	policy privacy_policy.UserAccessPolicy
}

func NewService(db *gorm.DB, policy privacy_policy.UserAccessPolicy) *Service {
	return &Service{
		repo:   newRepository(db),
		db:     db,
		policy: policy,
	}
}

func (s *Service) AddItem(ctx context.Context, userID uint, req CreateBMOLRequest) (*BMOLItemResponse, error) {
	if err := validateCreate(req); err != nil {
		return nil, err
	}

	var item *BMOLItem
	now := time.Now()
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		existing, err := findAnyByMedia(tx, userID, req.MediaID, req.MediaType)
		switch {
		case err == nil && !existing.DeletedAt.Valid:
			return ErrDuplicate
		case err == nil:
			// Restore the soft-deleted row (unique index includes it). The deleted_at
			// guard makes a concurrent restore lose with RowsAffected == 0.
			res := tx.Unscoped().Model(&BMOLItem{}).
				Where("id = ? AND deleted_at IS NOT NULL", existing.ID).
				Updates(map[string]any{"deleted_at": nil, "rank": req.Rank, "note": "", "created_at": now})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return ErrDuplicate
			}
			existing.DeletedAt = gorm.DeletedAt{}
			existing.Rank = req.Rank
			existing.Note = ""
			existing.CreatedAt = now
			item = existing
			return nil
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return err
		}
		item = &BMOLItem{UserID: userID, MediaID: req.MediaID, MediaType: req.MediaType, Rank: req.Rank}
		return tx.Create(item).Error
	})
	if err != nil {
		if shared.IsPgUniqueViolation(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}

	return s.toResponse(item), nil
}

func (s *Service) UpdateItem(ctx context.Context, itemID, requesterID uint, req UpdateBMOLRequest) (*BMOLItemResponse, error) {
	note, err := validateUpdate(req)
	if err != nil {
		return nil, err
	}

	item, err := s.repo.FindOne(itemID)
	if err != nil {
		return nil, ErrNotFound
	}

	if item.UserID != requesterID {
		return nil, ErrForbidden
	}

	if err := s.repo.Update(itemID, req.Rank, note); err != nil {
		return nil, err
	}

	if req.Rank != nil {
		item.Rank = *req.Rank
	}
	if note != nil {
		item.Note = *note
	}
	return s.toResponse(item), nil
}

func (s *Service) RemoveItem(ctx context.Context, itemID, requesterID uint) error {
	item, err := s.repo.FindOne(itemID)
	if err != nil {
		return ErrNotFound
	}

	if item.UserID != requesterID {
		return ErrForbidden
	}

	return s.repo.Delete(itemID)
}

func (s *Service) GetUserBMOL(ctx context.Context, viewerID, userID uint, mediaType *movie_module.MediaType) ([]BMOLItemResponse, error) {
	canView, err := s.policy.CanViewProfileSection(ctx, viewerID, userID, privacy_policy.SectionLibrary)
	if err != nil {
		return nil, err
	}
	if !canView {
		return nil, ErrForbidden
	}

	items, err := s.repo.FindByUser(userID, mediaType)
	if err != nil {
		return nil, err
	}

	responses := make([]BMOLItemResponse, len(items))
	media := make([]*shared.MediaSummary, len(items))
	var g errgroup.Group
	g.SetLimit(tmdbConcurrency)
	for i := range items {
		g.Go(func() error {
			responses[i] = *s.toResponse(&items[i])
			return nil
		})
	}
	_ = g.Wait()
	for i := range responses {
		media[i] = &responses[i].Media
	}
	if err := shared.AttachRatings(ctx, s.db, media); err != nil {
		return nil, err
	}
	return responses, nil
}

func (s *Service) toResponse(item *BMOLItem) *BMOLItemResponse {
	media := shared.MediaSummary{
		ID:        item.MediaID,
		MediaType: shared.MediaType(item.MediaType),
	}
	switch item.MediaType {
	case movie_module.MediaMovie:
		if details, err := tmdbmodule.GetMovieByID(item.MediaID); err == nil && details != nil {
			media.Title = details.Title
			media.PosterURL = details.PosterPath
			media.Genres = details.Genres
			media.VoteAverage = float32(details.VoteAverage)
		}
	case movie_module.MediaSeries:
		if details, err := tmdbmodule.GetSeriesByID(item.MediaID); err == nil && details != nil {
			media.Title = details.Name
			media.PosterURL = details.PosterPath
			media.Genres = details.Genres
			media.VoteAverage = float32(details.VoteAverage)
		}
	}

	return &BMOLItemResponse{
		ID:        item.ID,
		Media:     media,
		MediaType: item.MediaType,
		Rank:      item.Rank,
		Note:      item.Note,
		CreatedAt: item.CreatedAt,
	}
}

package bmol_module

import (
	"context"
	"errors"

	"github.com/arinsuda/movie-hub/internal/movie_module"
	shared "github.com/arinsuda/movie-hub/internal/shared"
	tmdbmodule "github.com/arinsuda/movie-hub/internal/tmdb_module"
	"gorm.io/gorm"
)

var (
	ErrDuplicate        = errors.New("media item is already ranked in your BOL")
	ErrForbidden        = errors.New("you do not have permission to modify this item")
	ErrInvalidMediaType = errors.New("invalid media type, must be movie or tv")
	ErrNotFound         = errors.New("item not found")
)

type Service struct {
	repo *repository
	db   *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		repo: newRepository(db),
		db:   db,
	}
}

func (s *Service) AddItem(ctx context.Context, userID uint, req CreateBMOLRequest) (*BMOLItemResponse, error) {
	if req.MediaType != movie_module.MediaMovie && req.MediaType != movie_module.MediaSeries {
		return nil, ErrInvalidMediaType
	}

	// Check duplicates
	existing, err := s.repo.FindDuplicate(userID, req.MediaID, req.MediaType)
	if err == nil && existing != nil {
		return nil, ErrDuplicate
	}

	item := &BMOLItem{
		UserID:    userID,
		MediaID:   req.MediaID,
		MediaType: req.MediaType,
		Rank:      req.Rank,
	}

	if err := s.repo.Create(item); err != nil {
		return nil, err
	}

	return s.toResponse(item), nil
}

func (s *Service) UpdateItem(ctx context.Context, itemID, requesterID uint, rank int) (*BMOLItemResponse, error) {
	item, err := s.repo.FindOne(itemID)
	if err != nil {
		return nil, ErrNotFound
	}

	if item.UserID != requesterID {
		return nil, ErrForbidden
	}

	if err := s.repo.Update(itemID, rank); err != nil {
		return nil, err
	}

	item.Rank = rank
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

func (s *Service) GetUserBMOL(ctx context.Context, userID uint, mediaType *movie_module.MediaType) ([]BMOLItemResponse, error) {
	items, err := s.repo.FindByUser(userID, mediaType)
	if err != nil {
		return nil, err
	}

	responses := make([]BMOLItemResponse, len(items))
	media := make([]*shared.MediaSummary, len(items))
	for i, item := range items {
		responses[i] = *s.toResponse(&item)
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
		CreatedAt: item.CreatedAt,
	}
}

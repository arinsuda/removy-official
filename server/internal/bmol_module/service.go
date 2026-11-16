package bmol_module

import (
	"context"
	"errors"
	"time"

	"github.com/arinsuda/movie-hub/internal/movie_module"
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	shared "github.com/arinsuda/movie-hub/internal/shared"
	"gorm.io/gorm"
)

const tmdbConcurrency = 6

var (
	ErrDuplicate        = errors.New("media item is already ranked in your BOL")
	ErrForbidden        = errors.New("you do not have permission to modify this item")
	ErrInvalidMediaType = errors.New("invalid media type, must be movie, tv or manga")
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

	if err := shared.CheckMediaAccess(ctx, userID, req.MediaID, shared.MediaType(req.MediaType)); err != nil {
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
	vis, err := shared.NewMediaVisibility(ctx, viewerID)
	if err != nil {
		return nil, err
	}
	items, err = shared.KeepVisible(ctx, vis, items, func(i BMOLItem) shared.VisibleRef {
		return shared.VisibleRef{Media: shared.MediaKey{ID: i.MediaID, Type: shared.MediaType(i.MediaType)}, OwnerID: i.UserID}
	})
	if err != nil {
		return nil, err
	}

	keys := make([]shared.MediaKey, len(items))
	for i, item := range items {
		keys[i] = shared.MediaKey{ID: item.MediaID, Type: shared.MediaType(item.MediaType)}
	}
	summaryByKey, err := shared.MediaSummaries(ctx, keys)
	if err != nil {
		return nil, err
	}
	responses := make([]BMOLItemResponse, len(items))
	media := make([]*shared.MediaSummary, len(items))
	for i := range items {
		responses[i] = *toResponseWith(&items[i], summaryByKey[keys[i]])
		media[i] = &responses[i].Media
	}
	if err := shared.AttachRatings(ctx, s.db, media); err != nil {
		return nil, err
	}
	return responses, nil
}

func (s *Service) toResponse(item *BMOLItem) *BMOLItemResponse {
	return toResponseWith(item, shared.FetchMediaSummary(item.MediaID, shared.MediaType(item.MediaType)))
}

// toResponseWith builds the DTO from a summary the caller already read (lists batch them).
func toResponseWith(item *BMOLItem, media shared.MediaSummary) *BMOLItemResponse {
	return &BMOLItemResponse{
		ID:        item.ID,
		Media:     media,
		MediaType: item.MediaType,
		Rank:      item.Rank,
		Note:      item.Note,
		CreatedAt: item.CreatedAt,
	}
}

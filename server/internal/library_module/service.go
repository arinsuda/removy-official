package library_module

import (
	"context"
	"encoding/json"
	"time"

	achievementsmodule "github.com/arinsuda/movie-hub/internal/achievements_module"
	"github.com/arinsuda/movie-hub/internal/feed_module"
	"github.com/arinsuda/movie-hub/internal/movie_module"
	notification_module "github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	shared "github.com/arinsuda/movie-hub/internal/shared"
	tmdbmodule "github.com/arinsuda/movie-hub/internal/tmdb_module"
	users "github.com/arinsuda/movie-hub/internal/user_module"
	stats "github.com/arinsuda/movie-hub/internal/user_stats_module"
	"gorm.io/gorm"
)

type Service struct {
	repo       *repository
	db         *gorm.DB
	expPort    stats.ExpAdder
	achieveSvc achievementsmodule.Service
	notifSvc   *notification_module.Service
	feedSvc    feed_module.Service
	policy     privacy_policy.UserAccessPolicy
}

func NewService(
	db *gorm.DB,
	exp stats.ExpAdder,
	achieve achievementsmodule.Service,
	notif *notification_module.Service,
	feed feed_module.Service,
	policy privacy_policy.UserAccessPolicy,
) *Service {
	return &Service{
		repo:       newRepository(db),
		db:         db,
		expPort:    exp,
		achieveSvc: achieve,
		notifSvc:   notif,
		feedSvc:    feed,
		policy:     policy,
	}
}

func (s *Service) AddItem(ctx context.Context, userID uint, req AddItemRequest) (*OwnLibraryItemResponse, error) {
	if err := validateAddItemRequest(req); err != nil {
		return nil, err
	}

	existing, err := s.repo.FindMediaStatus(userID, req.MediaID, req.MediaType)
	if err == nil {
		for _, item := range existing {
			if item.ListType == req.ListType {
				return nil, ErrDuplicate
			}
		}
	}

	item := &LibraryItem{
		UserID:    userID,
		MediaID:   req.MediaID,
		MediaType: req.MediaType,
		ListType:  req.ListType,
		Note:      req.Note,
	}
	if req.WatchedAt != nil {
		t, err := time.Parse("2006-01-02", *req.WatchedAt)
		if err != nil {
			return nil, ErrInvalidWatchedAt
		}
		item.WatchedAt = &t
	}

	if err := s.repo.Create(item); err != nil {
		return nil, err
	}

	switch req.ListType {
	case movie_module.ListWatched:
		_ = s.expPort.AddExperience(userID, stats.ExpPerWatched)

		var watchedCount int64
		s.db.WithContext(ctx).Model(&LibraryItem{}).
			Where("user_id = ? AND list_type = 'watched' AND deleted_at IS NULL", userID).
			Count(&watchedCount)
		unlocked := shared.TrackAndNotify(ctx, s.achieveSvc, s.notifSvc, userID, "watched_count", int(watchedCount))
		s.createAchievementFeedActivities(ctx, userID, unlocked)

		s.trackLibraryTotal(ctx, userID)

		if s.notifSvc != nil {
			if actor, err := s.getUserSummary(userID); err == nil {
				title, _ := s.fetchTitle(req.MediaID, req.MediaType)
				_ = s.notifSvc.PushFollowingAddedWatched(
					ctx,
					userID,
					actor.Username,
					uint(req.MediaID),
					title,
				)
			}
		}

		if s.feedSvc != nil {
			mediaType := string(req.MediaType)
			_ = s.feedSvc.CreateActivity(ctx, userID, feed_module.ActivityWatchedAdded, feed_module.ActivityPayload{
				MediaID:       &req.MediaID,
				MediaType:     &mediaType,
				LibraryItemID: &item.ID,
			})
		}

	case movie_module.ListWatchlist:
		var watchlistCount int64
		s.db.WithContext(ctx).Model(&LibraryItem{}).
			Where("user_id = ? AND list_type = 'watchlist' AND deleted_at IS NULL", userID).
			Count(&watchlistCount)
		unlocked := shared.TrackAndNotify(ctx, s.achieveSvc, s.notifSvc, userID, "watchlist_count", int(watchlistCount))
		s.createAchievementFeedActivities(ctx, userID, unlocked)
		s.trackLibraryTotal(ctx, userID)

		if s.notifSvc != nil {
			if actor, err := s.getUserSummary(userID); err == nil {
				title, _ := s.fetchTitle(req.MediaID, req.MediaType)
				_ = s.notifSvc.PushFollowingAddedWatchlist(
					ctx,
					userID,
					actor.Username,
					uint(req.MediaID),
					title,
				)
			}
		}

		if s.feedSvc != nil {
			mediaType := string(req.MediaType)
			_ = s.feedSvc.CreateActivity(ctx, userID, feed_module.ActivityWatchlistAdded, feed_module.ActivityPayload{
				MediaID:       &req.MediaID,
				MediaType:     &mediaType,
				LibraryItemID: &item.ID,
			})
		}
	}

	return toOwnResponse(item), nil
}

func (s *Service) GetLibrary(ctx context.Context, userID, requesterID uint, listType *movie_module.ListType, mediaType *movie_module.MediaType) (any, error) {
	canView, err := s.policy.CanViewProfileSection(ctx, requesterID, userID, privacy_policy.SectionLibrary)
	if err != nil {
		return nil, err
	}
	if !canView {
		return nil, privacy_policy.ErrForbidden
	}

	items, err := s.repo.FindByUser(userID, listType, mediaType)
	if err != nil {
		return nil, err
	}

	if userID == requesterID {
		responses := make([]OwnLibraryItemResponse, len(items))
		media := make([]*shared.MediaSummary, len(items))
		for i, item := range items {
			responses[i] = *toOwnResponse(&item)
			media[i] = &responses[i].Media
		}
		if err := shared.AttachRatings(ctx, s.db, media); err != nil {
			return nil, err
		}
		return responses, nil
	} else {
		responses := make([]PublicLibraryItemResponse, len(items))
		media := make([]*shared.MediaSummary, len(items))
		for i, item := range items {
			responses[i] = *toPublicResponse(&item)
			media[i] = &responses[i].Media
		}
		if err := shared.AttachRatings(ctx, s.db, media); err != nil {
			return nil, err
		}
		return responses, nil
	}
}

func (s *Service) RemoveItem(ctx context.Context, itemID, requesterID uint) error {
	item, err := s.repo.FindOne(itemID, requesterID)
	if err != nil {
		return err
	}
	if item.UserID != requesterID {
		return ErrForbidden
	}

	if err := s.repo.Delete(itemID); err != nil {
		return err
	}

	if item.ListType == movie_module.ListWatched {
		_ = s.expPort.AddExperience(requesterID, -stats.ExpPerWatched)
	}

	if s.feedSvc != nil {
		activityType := feed_module.ActivityWatchlistAdded
		if item.ListType == movie_module.ListWatched {
			activityType = feed_module.ActivityWatchedAdded
		}
		mediaType := string(item.MediaType)
		_ = s.feedSvc.DeleteMediaActivity(ctx, requesterID, activityType, item.MediaID, mediaType)
	}

	return nil
}

func (s *Service) UpdateItem(ctx context.Context, itemID, requesterID uint, req UpdateItemRequest) (*OwnLibraryItemResponse, error) {
	item, err := s.repo.FindOne(itemID, requesterID)
	if err != nil {
		return nil, err
	}
	if item.UserID != requesterID {
		return nil, ErrForbidden
	}

	updates := map[string]any{}
	if req.WatchedAt != nil {
		t, err := time.Parse("2006-01-02", *req.WatchedAt)
		if err != nil {
			return nil, ErrInvalidWatchedAt
		}
		updates["watched_at"] = t
	}
	if req.Tags != nil {
		b, _ := json.Marshal(req.Tags)
		updates["tags"] = string(b)
	}
	if req.Note != nil {
		updates["note"] = req.Note
	}
	if len(updates) == 0 {
		return toOwnResponse(item), nil
	}

	if err := s.repo.Update(itemID, updates); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindOne(itemID, requesterID)
	if err != nil {
		return nil, err
	}
	return toOwnResponse(updated), nil
}

func (s *Service) GetMediaStatus(userID uint, mediaID int, mediaType movie_module.MediaType) (*MediaStatusResponse, error) {
	items, err := s.repo.FindMediaStatus(userID, mediaID, mediaType)
	if err != nil {
		return nil, err
	}

	inLists := make([]MediaItemStatus, len(items))
	for i, item := range items {
		inLists[i] = MediaItemStatus{ListType: item.ListType, ItemID: item.ID}
	}

	return &MediaStatusResponse{MediaID: mediaID, MediaType: mediaType, InLists: inLists}, nil
}

func (s *Service) trackLibraryTotal(ctx context.Context, userID uint) {
	var total int64
	s.db.WithContext(ctx).Model(&LibraryItem{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&total)
	unlocked := shared.TrackAndNotify(ctx, s.achieveSvc, s.notifSvc, userID, "library_total", int(total))
	s.createAchievementFeedActivities(ctx, userID, unlocked)
}

func (s *Service) createAchievementFeedActivities(ctx context.Context, userID uint, unlocked []shared.NeutralUnlocked) {
	if s.feedSvc == nil {
		return
	}
	for _, u := range unlocked {
		_ = s.feedSvc.CreateActivity(ctx, userID, feed_module.ActivityAchievementUnlocked, feed_module.ActivityPayload{
			AchievementID: &u.ID,
			Message:       u.Name,
		})
	}
}

func (s *Service) getUserSummary(userID uint) (*users.User, error) {
	var u users.User
	err := s.db.First(&u, userID).Error
	return &u, err
}

func (s *Service) fetchTitle(mediaID int, mediaType movie_module.MediaType) (string, error) {
	switch mediaType {
	case movie_module.MediaMovie:
		if d, err := tmdbmodule.GetMovieByID(mediaID); err == nil {
			return d.Title, nil
		}
	case movie_module.MediaSeries:
		if d, err := tmdbmodule.GetSeriesByID(mediaID); err == nil {
			return d.Name, nil
		}
	}
	return "", nil
}

func validateAddItemRequest(req AddItemRequest) error {
	if req.MediaType != movie_module.MediaMovie && req.MediaType != movie_module.MediaSeries {
		return ErrInvalidMediaType
	}
	if req.ListType != movie_module.ListWatchlist &&
		req.ListType != movie_module.ListWatched {
		return ErrInvalidListType
	}
	if req.MediaID <= 0 {
		return ErrInvalidMediaID
	}
	return nil
}

func fetchMediaSummaryHelper(mediaID int, mediaType movie_module.MediaType) shared.MediaSummary {
	media := shared.MediaSummary{
		ID:        mediaID,
		MediaType: shared.MediaType(mediaType),
	}
	switch mediaType {
	case movie_module.MediaMovie:
		if details, err := tmdbmodule.GetMovieByID(mediaID); err == nil && details != nil {
			media.Title = details.Title
			media.PosterURL = details.PosterPath
			media.Genres = details.Genres
			media.VoteAverage = float32(details.VoteAverage)
		}
	case movie_module.MediaSeries:
		if details, err := tmdbmodule.GetSeriesByID(mediaID); err == nil && details != nil {
			media.Title = details.Name
			media.PosterURL = details.PosterPath
			media.Genres = details.Genres
			media.VoteAverage = float32(details.VoteAverage)
		}
	}
	return media
}

func toOwnResponse(item *LibraryItem) *OwnLibraryItemResponse {
	tags := []string{}
	if item.Tags != "" {
		_ = json.Unmarshal([]byte(item.Tags), &tags)
	}
	media := fetchMediaSummaryHelper(item.MediaID, item.MediaType)

	return &OwnLibraryItemResponse{
		ID:        item.ID,
		Media:     media,
		ListType:  item.ListType,
		Tags:      tags,
		WatchedAt: item.WatchedAt,
		Note:      item.Note,
		CreatedAt: item.CreatedAt,
	}
}

func toPublicResponse(item *LibraryItem) *PublicLibraryItemResponse {
	media := fetchMediaSummaryHelper(item.MediaID, item.MediaType)

	return &PublicLibraryItemResponse{
		ID:        item.ID,
		Media:     media,
		ListType:  item.ListType,
		CreatedAt: item.CreatedAt,
	}
}

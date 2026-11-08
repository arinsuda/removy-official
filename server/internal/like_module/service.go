package like_module

import (
	"context"
	"errors"

	achievementsmodule "github.com/arinsuda/movie-hub/internal/achievements_module"
	"github.com/arinsuda/movie-hub/internal/feed_module"
	"github.com/arinsuda/movie-hub/internal/movie_module"
	"github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/arinsuda/movie-hub/internal/shared"
	tmdbmodule "github.com/arinsuda/movie-hub/internal/tmdb_module"
	"gorm.io/gorm"
)

var ErrForbidden = errors.New("forbidden")

type Service struct {
	repo       *repository
	db         *gorm.DB
	achieveSvc achievementsmodule.Service
	notifSvc   *notification_module.Service
	feedSvc    feed_module.Service
	policy     privacy_policy.UserAccessPolicy
}

func NewService(
	db *gorm.DB,
	achieve achievementsmodule.Service,
	notif *notification_module.Service,
	feed feed_module.Service,
	policy privacy_policy.UserAccessPolicy,
) *Service {
	return &Service{
		repo:       newRepository(db),
		db:         db,
		achieveSvc: achieve,
		notifSvc:   notif,
		feedSvc:    feed,
		policy:     policy,
	}
}

func (s *Service) Like(ctx context.Context, userID uint, mediaID int, mediaType movie_module.MediaType) error {
	if mediaType != movie_module.MediaMovie && mediaType != movie_module.MediaSeries {
		return errors.New("invalid media type")
	}
	if mediaID <= 0 {
		return errors.New("invalid media id")
	}

	if err := s.repo.Create(userID, mediaID, mediaType); err != nil {
		return err
	}

	var count int64
	s.db.WithContext(ctx).Model(&MediaLike{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&count)
	unlocked := shared.TrackAndNotify(ctx, s.achieveSvc, s.notifSvc, userID, "like_count", int(count))
	s.createAchievementFeedActivities(ctx, userID, unlocked)

	if s.notifSvc != nil {
		if actor, err := s.getUserSummary(userID); err == nil {
			var title string
			switch mediaType {
			case movie_module.MediaMovie:
				if details, err := tmdbmodule.GetMovieByID(mediaID); err == nil && details != nil {
					title = details.Title
				}
			case movie_module.MediaSeries:
				if details, err := tmdbmodule.GetSeriesByID(mediaID); err == nil && details != nil {
					title = details.Name
				}
			}
			_ = s.notifSvc.PushFollowingLikedMedia(
				ctx,
				userID,
				actor.Username,
				uint(mediaID),
				title,
			)
		}
	}

	if s.feedSvc != nil {
		mt := string(mediaType)
		_ = s.feedSvc.CreateActivity(ctx, userID, feed_module.ActivityMediaLiked, feed_module.ActivityPayload{
			MediaID:   &mediaID,
			MediaType: &mt,
		})
	}

	return nil
}

func (s *Service) Unlike(ctx context.Context, userID uint, mediaID int, mediaType movie_module.MediaType) error {
	if err := s.repo.Delete(userID, mediaID, mediaType); err != nil {
		return err
	}

	if s.feedSvc != nil {
		mt := string(mediaType)
		_ = s.feedSvc.DeleteMediaActivity(ctx, userID, feed_module.ActivityMediaLiked, mediaID, mt)
	}

	return nil
}

func (s *Service) GetLikes(ctx context.Context, ownerID, requesterID uint) ([]LikeResponse, error) {
	if s.policy != nil {
		canView, err := s.policy.CanViewProfileSection(ctx, requesterID, ownerID, privacy_policy.SectionLibrary)
		if err != nil {
			return nil, err
		}
		if !canView {
			return nil, ErrForbidden
		}
	}

	likes, err := s.repo.FindByUser(ownerID)
	if err != nil {
		return nil, err
	}

	responses := make([]LikeResponse, len(likes))
	summaries := make([]*shared.MediaSummary, len(likes))
	for i, like := range likes {
		media := shared.MediaSummary{
			ID:        like.MediaID,
			MediaType: shared.MediaType(like.MediaType),
		}

		switch like.MediaType {
		case movie_module.MediaMovie:
			if details, err := tmdbmodule.GetMovieByID(like.MediaID); err == nil && details != nil {
				media.Title = details.Title
				media.PosterURL = details.PosterPath
				media.Genres = details.Genres
				media.VoteAverage = float32(details.VoteAverage)
			}
		case movie_module.MediaSeries:
			if details, err := tmdbmodule.GetSeriesByID(like.MediaID); err == nil && details != nil {
				media.Title = details.Name
				media.PosterURL = details.PosterPath
				media.Genres = details.Genres
				media.VoteAverage = float32(details.VoteAverage)
			}
		}

		responses[i] = LikeResponse{
			ID:        like.ID,
			UserID:    like.UserID,
			Media:     media,
			CreatedAt: like.CreatedAt,
		}
		summaries[i] = &responses[i].Media
	}
	if err := shared.AttachRatings(ctx, s.db, summaries); err != nil {
		return nil, err
	}

	return responses, nil
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

type UserSummary struct {
	Username string
}

func (s *Service) getUserSummary(userID uint) (*UserSummary, error) {
	var u struct {
		Username string
	}
	err := s.db.Table("users").Where("id = ?", userID).Select("username").First(&u).Error
	return &UserSummary{Username: u.Username}, err
}

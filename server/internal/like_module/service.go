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
	if !shared.IsValidMediaType(shared.MediaType(mediaType)) {
		return errors.New("invalid media type")
	}
	if mediaID <= 0 {
		return errors.New("invalid media id")
	}
	if err := shared.CheckMediaAccess(ctx, userID, mediaID, shared.MediaType(mediaType)); err != nil {
		return err
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
		media := shared.FetchMediaSummary(mediaID, shared.MediaType(mediaType))
		_ = s.notifSvc.PushFollowingLikedMedia(
			ctx,
			userID,
			notification_module.MediaRef{ID: uint(mediaID), Type: string(mediaType), Title: media.Title},
		)
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

// GuestLikesLimit caps how many likes a guest gets per request, which bounds the per-item
// TMDB lookups an unauthenticated caller can trigger.
const GuestLikesLimit = 100

// GetLikes returns the visible likes and whether the list was cut at GuestLikesLimit.
func (s *Service) GetLikes(ctx context.Context, ownerID, requesterID uint) ([]LikeResponse, bool, error) {
	if s.policy != nil {
		canView, err := s.policy.CanViewProfileSection(ctx, requesterID, ownerID, privacy_policy.SectionLibrary)
		if err != nil {
			return nil, false, err
		}
		if !canView {
			return nil, false, ErrForbidden
		}
	}

	limit := 0
	if requesterID == 0 {
		limit = GuestLikesLimit + 1 // one extra row tells us whether more exist
	}
	likes, err := s.repo.FindByUser(ownerID, limit)
	if err != nil {
		return nil, false, err
	}
	hasMore := limit > 0 && len(likes) > GuestLikesLimit
	if hasMore {
		likes = likes[:GuestLikesLimit]
	}
	vis, err := shared.NewMediaVisibility(ctx, requesterID)
	if err != nil {
		return nil, false, err
	}
	likes, err = shared.KeepVisible(ctx, vis, likes, func(l MediaLike) shared.VisibleRef {
		return shared.VisibleRef{Media: shared.MediaKey{ID: l.MediaID, Type: shared.MediaType(l.MediaType)}, OwnerID: l.UserID}
	})
	if err != nil {
		return nil, false, err
	}

	keys := make([]shared.MediaKey, len(likes))
	for i, like := range likes {
		keys[i] = shared.MediaKey{ID: like.MediaID, Type: shared.MediaType(like.MediaType)}
	}
	summaryByKey, err := shared.MediaSummaries(ctx, keys)
	if err != nil {
		return nil, false, err
	}
	responses := make([]LikeResponse, len(likes))
	summaries := make([]*shared.MediaSummary, len(likes))
	for i, like := range likes {
		responses[i] = LikeResponse{
			ID:        like.ID,
			UserID:    like.UserID,
			Media:     summaryByKey[keys[i]],
			CreatedAt: like.CreatedAt,
		}
		summaries[i] = &responses[i].Media
	}
	if err := shared.AttachRatings(ctx, s.db, summaries); err != nil {
		return nil, false, err
	}

	return responses, hasMore, nil
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

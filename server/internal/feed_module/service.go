package feed_module

import (
	"context"
	"errors"

	noti "github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/arinsuda/movie-hub/internal/shared"
	"github.com/arinsuda/movie-hub/internal/shared/storage"
	tmdbmodule "github.com/arinsuda/movie-hub/internal/tmdb_module"
	"gorm.io/gorm"
)

type Service interface {
	CreateActivity(ctx context.Context, actorID uint, activityType ActivityType, payload ActivityPayload) error
	GetFeed(ctx context.Context, userID uint, pq PaginationQuery) (*FeedListResponse, error)
	GetUserActivities(ctx context.Context, targetUserID, requesterID uint, pq PaginationQuery) (*FeedListResponse, error)
	UpdateActivityVisibility(ctx context.Context, activityID, requesterID uint, visibility privacy_policy.ActivityVisibility) error
	DeleteActivity(ctx context.Context, activityID, requesterID uint) error
	GetSettings(ctx context.Context, userID uint) (*ActivitySettingsResponse, error)
	UpdateSettings(ctx context.Context, userID uint, req UpdateActivitySettingsRequest) (*ActivitySettingsResponse, error)
	CountNewFeedItems(ctx context.Context, userID uint, afterActivityID uint, category string) (int64, error)
	IsActivityEnabled(ctx context.Context, userID uint, activityType ActivityType) (bool, error)

	DeleteReviewActivity(ctx context.Context, actorID uint, reviewID uint) error
	DeleteCommentActivity(ctx context.Context, actorID uint, commentID uint) error
	DeleteWatchLogActivity(ctx context.Context, actorID uint, watchLogID uint) error
	DeleteMediaActivity(ctx context.Context, actorID uint, activityType ActivityType, mediaID int, mediaType string) error
	DeleteFollowActivity(ctx context.Context, followerID uint, followeeID uint) error
}

type service struct {
	db    *gorm.DB
	repo  *repository
	hub   *noti.Hub
	minio *storage.MinIOClient
}

func NewService(db *gorm.DB, hub *noti.Hub, minio *storage.MinIOClient) Service {
	return &service{
		db:    db,
		repo:  newRepository(db),
		hub:   hub,
		minio: minio,
	}
}

func (s *service) CreateActivity(ctx context.Context, actorID uint, activityType ActivityType, payload ActivityPayload) error {
	enabled, err := s.repo.IsEnabled(ctx, actorID, activityType)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	event := &ActivityEvent{
		ActorID:       actorID,
		Type:          activityType,
		MediaID:       payload.MediaID,
		MediaType:     payload.MediaType,
		ReviewID:      payload.ReviewID,
		CommentID:     payload.CommentID,
		AchievementID: payload.AchievementID,
		LibraryItemID: payload.LibraryItemID,
		WatchLogID:    payload.WatchLogID,
		TargetUserID:  payload.TargetUserID,
		Message:       payload.Message,
		Visibility:    privacy_policy.VisibilityDefault,
		IsVisible:     true,
	}

	if err := s.repo.CreateOrRestore(ctx, event); err != nil {
		return err
	}

	s.broadcastRefresh(ctx, actorID)
	return nil
}

func (s *service) GetFeed(ctx context.Context, userID uint, pq PaginationQuery) (*FeedListResponse, error) {
	pq.Normalize()
	rows, total, err := s.repo.FindFeed(ctx, userID, pq)
	if err != nil {
		return nil, err
	}
	return s.withRatings(ctx, toFeedListResponse(rows, pq, total, s.minio))
}

func (s *service) GetUserActivities(ctx context.Context, targetUserID, requesterID uint, pq PaginationQuery) (*FeedListResponse, error) {
	pq.Normalize()
	rows, total, err := s.repo.FindUserActivities(ctx, targetUserID, requesterID, pq)
	if err != nil {
		return nil, err
	}
	return s.withRatings(ctx, toFeedListResponse(rows, pq, total, s.minio))
}

func (s *service) UpdateActivityVisibility(ctx context.Context, activityID, requesterID uint, visibility privacy_policy.ActivityVisibility) error {
	if visibility != privacy_policy.VisibilityDefault &&
		visibility != privacy_policy.VisibilityPublic &&
		visibility != privacy_policy.VisibilityFollowers &&
		visibility != privacy_policy.VisibilityPrivate {
		return errors.New("invalid visibility value")
	}

	err := s.repo.UpdateOwnedActivityVisibility(ctx, activityID, requesterID, visibility)
	if err == nil {
		s.broadcastActivityUpdate(ctx, requesterID, activityID, visibility)
	}
	return err
}

func (s *service) DeleteActivity(ctx context.Context, activityID, requesterID uint) error {
	err := s.repo.DeleteOwnedActivity(ctx, activityID, requesterID)
	if err == nil {
		s.broadcastActivityDelete(ctx, requesterID, activityID)
	}
	return err
}

func (s *service) DeleteReviewActivity(ctx context.Context, actorID uint, reviewID uint) error {
	err := s.repo.DeleteReviewActivity(ctx, actorID, reviewID)
	if err == nil {
		s.broadcastRefresh(ctx, actorID)
	}
	return err
}

func (s *service) DeleteCommentActivity(ctx context.Context, actorID uint, commentID uint) error {
	err := s.repo.DeleteCommentActivity(ctx, actorID, commentID)
	if err == nil {
		s.broadcastRefresh(ctx, actorID)
	}
	return err
}

func (s *service) DeleteWatchLogActivity(ctx context.Context, actorID uint, watchLogID uint) error {
	err := s.repo.DeleteWatchLogActivity(ctx, actorID, watchLogID)
	if err == nil {
		s.broadcastRefresh(ctx, actorID)
	}
	return err
}

func (s *service) DeleteMediaActivity(ctx context.Context, actorID uint, activityType ActivityType, mediaID int, mediaType string) error {
	err := s.repo.DeleteMediaActivity(ctx, actorID, activityType, mediaID, mediaType)
	if err == nil {
		s.broadcastRefresh(ctx, actorID)
	}
	return err
}

func (s *service) DeleteFollowActivity(ctx context.Context, followerID uint, followeeID uint) error {
	err := s.repo.DeleteFollowActivity(ctx, followerID, followeeID)
	if err == nil {
		s.broadcastRefresh(ctx, followerID)
	}
	return err
}

func (s *service) GetSettings(ctx context.Context, userID uint) (*ActivitySettingsResponse, error) {
	rows, err := s.repo.FindSettings(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toSettingsResponse(rows), nil
}

func (s *service) UpdateSettings(ctx context.Context, userID uint, req UpdateActivitySettingsRequest) (*ActivitySettingsResponse, error) {
	updates := map[ActivityType]*bool{
		ActivityReviewCreated:       req.ReviewCreated,
		ActivityReviewCommented:     req.ReviewCommented,
		ActivityReviewLiked:         req.ReviewLiked,
		ActivityMediaLiked:          req.MediaLiked,
		ActivityWatchlistAdded:      req.WatchlistAdded,
		ActivityWatchedAdded:        req.WatchedAdded,
		ActivityAchievementUnlocked: req.AchievementUnlocked,
		ActivityUserFollowed:        req.UserFollowed,
		ActivityWatchLogCreated:     req.WatchLogCreated,
		ActivityRewatched:           req.Rewatched,
	}

	for activityType, value := range updates {
		if value == nil {
			continue
		}
		if err := s.repo.UpsertSetting(ctx, userID, activityType, *value); err != nil {
			return nil, err
		}
	}

	return s.GetSettings(ctx, userID)
}

func (s *service) CountNewFeedItems(ctx context.Context, userID uint, afterActivityID uint, category string) (int64, error) {
	return s.repo.CountNewFeedItems(ctx, userID, afterActivityID, category)
}

func (s *service) broadcastRefresh(ctx context.Context, actorID uint) {
	if s.hub == nil {
		return
	}
	s.hub.EmitFeedRefresh(actorID)
	var followerIDs []uint
	if err := s.repo.db.WithContext(ctx).Table("user_follows").Where("followee_id = ? AND status = 'accepted'", actorID).Pluck("follower_id", &followerIDs).Error; err == nil {
		for _, followerID := range followerIDs {
			s.hub.EmitFeedRefresh(followerID)
		}
	}
}

func (s *service) broadcastActivityUpdate(ctx context.Context, actorID uint, activityID uint, visibility privacy_policy.ActivityVisibility) {
	if s.hub == nil {
		return
	}
	s.hub.EmitFeedUpdated(actorID, activityID, string(visibility))
	var followerIDs []uint
	if err := s.repo.db.WithContext(ctx).Table("user_follows").Where("followee_id = ? AND status = 'accepted'", actorID).Pluck("follower_id", &followerIDs).Error; err == nil {
		for _, followerID := range followerIDs {
			s.hub.EmitFeedUpdated(followerID, activityID, string(visibility))
		}
	}
}

func (s *service) broadcastActivityDelete(ctx context.Context, actorID uint, activityID uint) {
	if s.hub == nil {
		return
	}
	s.hub.EmitFeedDeleted(actorID, activityID)
	var followerIDs []uint
	if err := s.repo.db.WithContext(ctx).Table("user_follows").Where("followee_id = ? AND status = 'accepted'", actorID).Pluck("follower_id", &followerIDs).Error; err == nil {
		for _, followerID := range followerIDs {
			s.hub.EmitFeedDeleted(followerID, activityID)
		}
	}
}

// ── Helpers ──────────────────────────────────────────────────────

func toSettingsResponse(rows []ActivityPrivacySetting) *ActivitySettingsResponse {
	enabled := make(map[ActivityType]bool, len(AllActivityTypes))
	for _, t := range AllActivityTypes {
		enabled[t] = privacy_policy.ActivityEnabledByDefault(t)
	}
	for _, row := range rows {
		enabled[row.ActivityType] = row.Enabled
	}

	return &ActivitySettingsResponse{
		ReviewCreated:       enabled[ActivityReviewCreated],
		ReviewCommented:     enabled[ActivityReviewCommented],
		ReviewLiked:         enabled[ActivityReviewLiked],
		MediaLiked:          enabled[ActivityMediaLiked],
		WatchlistAdded:      enabled[ActivityWatchlistAdded],
		WatchedAdded:        enabled[ActivityWatchedAdded],
		AchievementUnlocked: enabled[ActivityAchievementUnlocked],
		UserFollowed:        enabled[ActivityUserFollowed],
		WatchLogCreated:     enabled[ActivityWatchLogCreated],
		Rewatched:           enabled[ActivityRewatched],
	}
}

func toFeedListResponse(rows []feedRow, pq PaginationQuery, total int64, minio *storage.MinIOClient) *FeedListResponse {
	items := make([]FeedItemResponse, len(rows))
	for i, row := range rows {
		items[i] = toFeedItemResponse(row, minio)
	}
	return &FeedListResponse{
		Items:      items,
		Pagination: newPaginationMeta(pq.Page, pq.Limit, total),
	}
}

func toFeedItemResponse(row feedRow, minio *storage.MinIOClient) FeedItemResponse {
	item := FeedItemResponse{
		ID:   row.ID,
		Type: row.Type,
		Actor: ActorSummary{
			ID:          row.ActorID,
			Username:    row.ActorUsername,
			DisplayName: row.ActorDisplayName,
			AvatarURL:   presignAvatar(row.ActorAvatarURL, minio),
		},
		ReviewID:      row.ReviewID,
		CommentID:     row.CommentID,
		AchievementID: row.AchievementID,
		LibraryItemID: row.LibraryItemID,
		Message:       row.Message,
		Visibility:    row.Visibility,
		CreatedAt:     row.CreatedAt,
	}

	if row.TargetUserID != nil {
		item.TargetUser = &ActorSummary{
			ID:          *row.TargetUserID,
			Username:    row.TargetUserUsername,
			DisplayName: row.TargetUserDisplayName,
			AvatarURL:   presignAvatar(row.TargetUserAvatarURL, minio),
		}
	} else if row.Type == ActivityUserFollowed {
		// Target user was physically or soft deleted
		deletedUsername := "deleted_user"
		deletedDisplayName := "ผู้ใช้ที่ถูกลบ"
		item.TargetUser = &ActorSummary{
			ID:          0,
			Username:    deletedUsername,
			DisplayName: &deletedDisplayName,
			AvatarURL:   nil,
		}
	}

	if row.MediaID != nil && row.MediaType != nil {
		media := shared.MediaSummary{
			ID:        *row.MediaID,
			MediaType: shared.MediaType(*row.MediaType),
		}
		switch shared.MediaType(*row.MediaType) {
		case shared.MediaTypeMovie:
			if details, err := tmdbmodule.GetMovieByID(*row.MediaID); err == nil && details != nil {
				media.Title = details.Title
				media.PosterURL = tmdbmodule.ImageURL(details.PosterPath)
				media.Genres = details.Genres
				media.VoteAverage = float32(details.VoteAverage)
			}
		case shared.MediaTypeTV:
			if details, err := tmdbmodule.GetSeriesByID(*row.MediaID); err == nil && details != nil {
				media.Title = details.Name
				media.PosterURL = tmdbmodule.ImageURL(details.PosterPath)
				media.Genres = details.Genres
				media.VoteAverage = float32(details.VoteAverage)
			}
		}
		item.Media = &media
	}

	return item
}

// presignAvatar resolves a MinIO object key to a presigned URL.
// Returns the raw pointer unchanged if minio is nil or presigning fails.
func presignAvatar(avatarURL *string, minio *storage.MinIOClient) *string {
	if avatarURL == nil || *avatarURL == "" || minio == nil {
		return avatarURL
	}
	if presigned, err := minio.PresignURL(context.Background(), *avatarURL); err == nil {
		return &presigned
	}
	return avatarURL
}

func (s *service) IsActivityEnabled(ctx context.Context, userID uint, activityType ActivityType) (bool, error) {
	return s.repo.IsEnabled(ctx, userID, activityType)
}

// withRatings attaches popcorn ratings to every feed item that carries media (one query per media type).
func (s *service) withRatings(ctx context.Context, list *FeedListResponse) (*FeedListResponse, error) {
	var media []*shared.MediaSummary
	for i := range list.Items {
		if list.Items[i].Media != nil {
			media = append(media, list.Items[i].Media)
		}
	}
	if err := shared.AttachRatings(ctx, s.db, media); err != nil {
		return nil, err
	}
	return list, nil
}

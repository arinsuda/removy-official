package notification_module

import (
	"context"
	"errors"

	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	user_module "github.com/arinsuda/movie-hub/internal/user_module"
	"gorm.io/gorm"
)

type UserProvider interface {
	FindByID(id uint) (*user_module.User, int, int, int, error)
	FindFollowerIDs(userID uint) ([]uint, error)
}

type Service struct {
	repo         *repository
	userProvider UserProvider
	hub          *Hub
}

func NewService(db *gorm.DB, userProvider UserProvider, hub *Hub) *Service {
	return &Service{
		repo:         newRepository(db),
		userProvider: userProvider,
		hub:          hub,
	}
}

func (s *Service) PushWelcome(ctx context.Context, userID uint) error {
	return s.createAndEmit(ctx, []Notification{{UserID: userID, Type: NotifWelcome}})
}

func (s *Service) ListNotifications(ctx context.Context, userID uint, q ListNotificationsQuery) (*NotificationListResponse, error) {
	rows, total, err := s.repo.FindByUser(ctx, userID, q)
	if err != nil {
		return nil, err
	}
	unread, err := s.repo.CountUnread(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, pageSize := normalizePagination(q.Page, q.PageSize)

	responses := make([]NotificationResponse, 0, len(rows))
	for _, n := range rows {
		responses = append(responses, s.toResponse(n))
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return &NotificationListResponse{
		Notifications: responses,
		UnreadCount:   unread,
		Pagination: NotificationPaginationMeta{
			Page: page, Limit: pageSize, Total: total, TotalPages: totalPages,
		},
	}, nil
}

func (s *Service) GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	return s.repo.CountUnread(ctx, userID)
}

func (s *Service) GetUnreadCountByCategory(ctx context.Context, userID uint) ([]UnreadByCategoryResponse, error) {
	return s.repo.CountUnreadByCategory(ctx, userID)
}

func (s *Service) MarkRead(ctx context.Context, userID uint, req MarkReadRequest) error {
	if err := s.repo.MarkRead(ctx, userID, req.IDs); err != nil {
		return err
	}
	s.hub.EmitRead(userID, req.IDs)
	if count, err := s.repo.CountUnread(ctx, userID); err == nil {
		s.hub.EmitUnreadCount(userID, count)
	}
	return nil
}

func (s *Service) DeleteNotifications(ctx context.Context, userID uint, ids []uint) error {
	if err := s.repo.DeleteByUser(ctx, userID, ids); err != nil {
		return err
	}
	s.hub.EmitDeleted(userID, ids)
	return nil
}

func (s *Service) DeleteAllNotifications(ctx context.Context, userID uint) error {
	if err := s.repo.DeleteAllByUser(ctx, userID); err != nil {
		return err
	}
	s.hub.EmitDeleted(userID, []uint{})
	return nil
}

func (s *Service) createAndEmit(ctx context.Context, ns []Notification) error {
	if len(ns) == 0 {
		return nil
	}
	for i := range ns {
		if ns[i].Category == "" {
			ns[i].Category = categoryByType[ns[i].Type]
		}
	}
	if err := s.repo.CreateBatch(ctx, ns); err != nil {
		return err
	}
	for _, n := range ns {
		s.hub.EmitNew(n.UserID, s.toResponse(n))
	}
	return nil
}

func (s *Service) pushUserEvent(ctx context.Context, t NotificationType, targetUserID, actorID uint) error {
	n := Notification{
		UserID:    targetUserID,
		ActorID:   &actorID,
		Type:      t,
		TargetID:  &actorID,
		TargetRef: ptr("user"),
	}
	return s.createAndEmit(ctx, []Notification{n})
}

func (s *Service) PushFollowedYou(ctx context.Context, targetUserID, actorID uint) error {
	return s.pushUserEvent(ctx, NotifFollowedYou, targetUserID, actorID)
}

func (s *Service) PushFollowRequested(ctx context.Context, targetUserID, actorID uint) error {
	return s.pushUserEvent(ctx, NotifFollowRequested, targetUserID, actorID)
}

func (s *Service) PushFollowAccepted(ctx context.Context, targetUserID, actorID uint) error {
	return s.pushUserEvent(ctx, NotifFollowAccepted, targetUserID, actorID)
}

func (s *Service) PushMovieNowPlaying(ctx context.Context, userIDs []uint, media MediaRef) error {
	ns := make([]Notification, 0, len(userIDs))
	for _, uid := range userIDs {
		ns = append(ns, Notification{
			UserID:    uid,
			ActorID:   nil,
			Type:      NotifMovieNowPlaying,
			TargetID:  &media.ID,
			TargetRef: ptr(media.Type),
			Params:    media.params(),
		})
	}
	return s.createAndEmit(ctx, ns)
}

func (s *Service) PushFollowingActivity(
	ctx context.Context,
	actorID uint,
	notifType NotificationType,
	targetID *uint,
	targetRef *string,
	params *NotificationParams,
) error {
	// Map notification type to the corresponding feed activity type to check privacy settings.
	var activityType privacy_policy.ActivityType
	switch notifType {
	case NotifFollowingReviewed:
		activityType = privacy_policy.ActivityReviewCreated
	case NotifFollowingCommented:
		activityType = privacy_policy.ActivityReviewCommented
	case NotifFollowingLikedReview:
		activityType = privacy_policy.ActivityReviewLiked
	case NotifFollowingAddedWatchlist:
		activityType = privacy_policy.ActivityWatchlistAdded
	case NotifFollowingAddedWatched:
		activityType = privacy_policy.ActivityWatchedAdded
	case NotifFollowingMarkedHelpful:
		// review_liked setting controls marked helpful as well
		activityType = privacy_policy.ActivityReviewLiked
	case NotifFollowingLikedMedia:
		activityType = privacy_policy.ActivityMediaLiked
	}

	if activityType != "" {
		var setting struct {
			Enabled bool
		}
		err := s.repo.db.WithContext(ctx).Table("activity_privacy_settings").
			Select("enabled").
			Where("user_id = ? AND activity_type = ?", actorID, string(activityType)).
			First(&setting).Error

		enabled := false
		if err == nil {
			enabled = setting.Enabled
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			// Fallback to default activity sharing setting
			defaultActivitySharing := map[privacy_policy.ActivityType]bool{
				privacy_policy.ActivityReviewCreated:       true,
				privacy_policy.ActivityReviewCommented:     true,
				privacy_policy.ActivityReviewLiked:         false,
				privacy_policy.ActivityMediaLiked:          false,
				privacy_policy.ActivityWatchlistAdded:      false,
				privacy_policy.ActivityWatchedAdded:        false,
				privacy_policy.ActivityAchievementUnlocked: true,
				privacy_policy.ActivityUserFollowed:        false,
			}
			enabled = defaultActivitySharing[activityType]
		} else {
			return err
		}

		if !enabled {
			// Sharing is disabled for this activity type, do not send notifications to followers.
			return nil
		}
	}

	followerIDs, err := s.userProvider.FindFollowerIDs(actorID)
	if err != nil {
		return err
	}

	ns := make([]Notification, 0, len(followerIDs))
	for _, fid := range followerIDs {
		ns = append(ns, Notification{
			UserID:    fid,
			ActorID:   &actorID,
			Type:      notifType,
			TargetID:  targetID,
			TargetRef: targetRef,
			Params:    params,
		})
	}
	return s.createAndEmit(ctx, ns)
}

// pushFollowingMedia fans a media activity out to followers; the target is the media.
func (s *Service) pushFollowingMedia(ctx context.Context, t NotificationType, actorID uint, media MediaRef) error {
	return s.PushFollowingActivity(ctx, actorID, t, &media.ID, ptr(media.Type), media.params())
}

// pushFollowingReview fans a review activity out to followers; the target is the review.
func (s *Service) pushFollowingReview(ctx context.Context, t NotificationType, actorID, reviewID uint, media MediaRef) error {
	return s.PushFollowingActivity(ctx, actorID, t, &reviewID, ptr("review"), media.params())
}

func (s *Service) PushFollowingReviewed(ctx context.Context, actorID, reviewID uint, media MediaRef) error {
	return s.pushFollowingReview(ctx, NotifFollowingReviewed, actorID, reviewID, media)
}

func (s *Service) PushFollowingLikedReview(ctx context.Context, actorID, reviewID uint, media MediaRef) error {
	return s.pushFollowingReview(ctx, NotifFollowingLikedReview, actorID, reviewID, media)
}

func (s *Service) PushFollowingMarkedHelpful(ctx context.Context, actorID, reviewID uint, media MediaRef) error {
	return s.pushFollowingReview(ctx, NotifFollowingMarkedHelpful, actorID, reviewID, media)
}

func (s *Service) PushFollowingCommented(ctx context.Context, actorID, reviewID uint, media MediaRef) error {
	return s.pushFollowingReview(ctx, NotifFollowingCommented, actorID, reviewID, media)
}

func (s *Service) PushFollowingAddedWatchlist(ctx context.Context, actorID uint, media MediaRef) error {
	return s.pushFollowingMedia(ctx, NotifFollowingAddedWatchlist, actorID, media)
}

func (s *Service) PushFollowingAddedWatched(ctx context.Context, actorID uint, media MediaRef) error {
	return s.pushFollowingMedia(ctx, NotifFollowingAddedWatched, actorID, media)
}

func (s *Service) PushFollowingLikedMedia(ctx context.Context, actorID uint, media MediaRef) error {
	return s.pushFollowingMedia(ctx, NotifFollowingLikedMedia, actorID, media)
}

func (s *Service) toResponse(n Notification) NotificationResponse {
	resp := NotificationResponse{
		ID:        n.ID,
		Type:      n.Type,
		Category:  n.Category,
		Message:   n.Message,
		Params:    n.Params,
		IsRead:    n.IsRead,
		ReadAt:    n.ReadAt,
		TargetID:  n.TargetID,
		TargetRef: n.TargetRef,
		CreatedAt: n.CreatedAt,
	}

	if n.ActorID != nil {
		if actor, _, _, _, err := s.userProvider.FindByID(*n.ActorID); err == nil {
			resp.Actor = &ActorSummary{
				ID:          actor.ID,
				Username:    actor.Username,
				DisplayName: actor.DisplayName,
				AvatarURL:   actor.AvatarURL,
			}
		}
	}

	return resp
}

func ptr[T any](v T) *T { return &v }

// pushReviewOwner notifies a review's owner about someone else's action on it.
func (s *Service) pushReviewOwner(ctx context.Context, t NotificationType, reviewOwnerID, actorID, reviewID uint, media MediaRef) error {
	if reviewOwnerID == actorID {
		return nil
	}
	n := Notification{
		UserID: reviewOwnerID, ActorID: &actorID, Type: t,
		TargetID: &reviewID, TargetRef: ptr("review"),
		Params: media.params(),
	}
	return s.createAndEmit(ctx, []Notification{n})
}

func (s *Service) PushReviewLiked(ctx context.Context, reviewOwnerID, actorID, reviewID uint, media MediaRef) error {
	return s.pushReviewOwner(ctx, NotifReviewLiked, reviewOwnerID, actorID, reviewID, media)
}

func (s *Service) PushReviewMarkedHelpful(ctx context.Context, reviewOwnerID, actorID, reviewID uint, media MediaRef) error {
	return s.pushReviewOwner(ctx, NotifReviewMarkedHelpful, reviewOwnerID, actorID, reviewID, media)
}

func (s *Service) PushReviewCommented(ctx context.Context, reviewOwnerID, actorID, reviewID uint, media MediaRef) error {
	return s.pushReviewOwner(ctx, NotifReviewCommented, reviewOwnerID, actorID, reviewID, media)
}

func (s *Service) PushAchievementUnlocked(ctx context.Context, userID uint, a AchievementRef, exp int) error {
	n := Notification{
		UserID: userID, Type: NotifAchievementUnlocked,
		TargetID: &a.ID, TargetRef: ptr("achievement"),
		Params: &NotificationParams{ActionType: a.ActionType, TargetCount: a.TargetCount, Exp: exp, Name: a.Name},
	}
	return s.createAndEmit(ctx, []Notification{n})
}

func (s *Service) PushEmailVerified(ctx context.Context, userID uint) error {
	return s.createAndEmit(ctx, []Notification{{UserID: userID, Type: NotifEmailVerified}})
}

func (s *Service) PushPasswordChanged(ctx context.Context, userID uint) error {
	return s.createAndEmit(ctx, []Notification{{UserID: userID, Type: NotifPasswordChanged}})
}

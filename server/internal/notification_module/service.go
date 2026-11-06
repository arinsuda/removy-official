package notification_module

import (
	"context"
	"errors"
	"fmt"

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

func (s *Service) PushWelcome(ctx context.Context, userID uint, username string) error {
	n := Notification{
		UserID:  userID,
		Type:    NotifWelcome,
		Message: fmt.Sprintf("ยินดีต้อนรับสู่ POPCORN, %s! เริ่มสร้าง Portfolio สำหรับการดูภาพยนต์ ของคุณได้เลย", username),
	}
	return s.createAndEmit(ctx, []Notification{n})
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

func (s *Service) PushFollowedYou(ctx context.Context, targetUserID, actorID uint, actorUsername string) error {
	n := Notification{
		UserID:    targetUserID,
		ActorID:   &actorID,
		Type:      NotifFollowedYou,
		TargetID:  &actorID,
		TargetRef: ptr("user"),
		Message:   fmt.Sprintf("%s started following you", actorUsername),
	}
	return s.createAndEmit(ctx, []Notification{n})
}

func (s *Service) PushFollowRequested(ctx context.Context, targetUserID, actorID uint, actorUsername string) error {
	n := Notification{
		UserID:    targetUserID,
		ActorID:   &actorID,
		Type:      NotifFollowRequested,
		TargetID:  &actorID,
		TargetRef: ptr("user"),
		Message:   fmt.Sprintf("%s requested to follow you", actorUsername),
	}
	return s.createAndEmit(ctx, []Notification{n})
}

func (s *Service) PushFollowAccepted(ctx context.Context, targetUserID, actorID uint, actorUsername string) error {
	n := Notification{
		UserID:    targetUserID,
		ActorID:   &actorID,
		Type:      NotifFollowAccepted,
		TargetID:  &actorID,
		TargetRef: ptr("user"),
		Message:   fmt.Sprintf("%s accepted your follow request", actorUsername),
	}
	return s.createAndEmit(ctx, []Notification{n})
}

func (s *Service) PushMovieNowPlaying(ctx context.Context, userIDs []uint, movieID uint, movieTitle string) error {
	ns := make([]Notification, 0, len(userIDs))
	for _, uid := range userIDs {
		ns = append(ns, Notification{
			UserID:    uid,
			ActorID:   nil,
			Type:      NotifMovieNowPlaying,
			TargetID:  &movieID,
			TargetRef: ptr("movie"),
			Message:   fmt.Sprintf(`"%s" is now playing — time to watch!`, movieTitle),
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
	message string,
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
			Message:   message,
		})
	}
	return s.createAndEmit(ctx, ns)
}

func (s *Service) PushFollowingReviewed(ctx context.Context, actorID uint, actorUsername string, reviewID uint, movieTitle string) error {
	ref := "review"
	msg := fmt.Sprintf("%s reviewed %q", actorUsername, movieTitle)
	return s.PushFollowingActivity(ctx, actorID, NotifFollowingReviewed, &reviewID, &ref, msg)
}

func (s *Service) PushFollowingLikedReview(ctx context.Context, actorID uint, actorUsername string, reviewID uint, movieTitle string) error {
	ref := "review"
	msg := fmt.Sprintf("%s liked a review of %q", actorUsername, movieTitle)
	return s.PushFollowingActivity(ctx, actorID, NotifFollowingLikedReview, &reviewID, &ref, msg)
}

func (s *Service) PushFollowingAddedWatchlist(ctx context.Context, actorID uint, actorUsername string, movieID uint, movieTitle string) error {
	ref := "movie"
	msg := fmt.Sprintf("%s added %q to their watchlist", actorUsername, movieTitle)
	return s.PushFollowingActivity(ctx, actorID, NotifFollowingAddedWatchlist, &movieID, &ref, msg)
}

func (s *Service) PushFollowingAddedWatched(ctx context.Context, actorID uint, actorUsername string, movieID uint, movieTitle string) error {
	ref := "movie"
	msg := fmt.Sprintf("%s marked %q as watched", actorUsername, movieTitle)
	return s.PushFollowingActivity(ctx, actorID, NotifFollowingAddedWatched, &movieID, &ref, msg)
}

func (s *Service) PushFollowingLikedMedia(ctx context.Context, actorID uint, actorUsername string, movieID uint, movieTitle string) error {
	ref := "movie"
	msg := fmt.Sprintf("%s liked %q", actorUsername, movieTitle)
	return s.PushFollowingActivity(ctx, actorID, NotifFollowingLikedMedia, &movieID, &ref, msg)
}

func (s *Service) toResponse(n Notification) NotificationResponse {
	resp := NotificationResponse{
		ID:        n.ID,
		Type:      n.Type,
		Category:  n.Category,
		Message:   n.Message,
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

func (s *Service) PushReviewLiked(ctx context.Context, reviewOwnerID, actorID, reviewID uint, actorUsername, movieTitle string) error {
	if reviewOwnerID == actorID {
		return nil
	}
	ref := "review"
	n := Notification{
		UserID: reviewOwnerID, ActorID: &actorID, Type: NotifReviewLiked,
		TargetID: &reviewID, TargetRef: &ref,
		Message: fmt.Sprintf("%s liked your review of %q", actorUsername, movieTitle),
	}
	return s.createAndEmit(ctx, []Notification{n})
}

func (s *Service) PushReviewMarkedHelpful(ctx context.Context, reviewOwnerID, actorID, reviewID uint, actorUsername, movieTitle string) error {
	if reviewOwnerID == actorID {
		return nil
	}
	ref := "review"
	n := Notification{
		UserID: reviewOwnerID, ActorID: &actorID, Type: NotifReviewMarkedHelpful,
		TargetID: &reviewID, TargetRef: &ref,
		Message: fmt.Sprintf("%s marked your review of %q as helpful", actorUsername, movieTitle),
	}
	return s.createAndEmit(ctx, []Notification{n})
}

func (s *Service) PushReviewCommented(ctx context.Context, reviewOwnerID, actorID, reviewID uint, actorUsername, movieTitle string) error {
	if reviewOwnerID == actorID {
		return nil
	}
	ref := "review"
	n := Notification{
		UserID: reviewOwnerID, ActorID: &actorID, Type: NotifReviewCommented,
		TargetID: &reviewID, TargetRef: &ref,
		Message: fmt.Sprintf("%s commented on your review of %q", actorUsername, movieTitle),
	}
	return s.createAndEmit(ctx, []Notification{n})
}

func (s *Service) PushFollowingMarkedHelpful(ctx context.Context, actorID uint, actorUsername string, reviewID uint, movieTitle string) error {
	ref := "review"
	msg := fmt.Sprintf("%s marked a review of %q as helpful", actorUsername, movieTitle)
	return s.PushFollowingActivity(ctx, actorID, NotifFollowingMarkedHelpful, &reviewID, &ref, msg)
}

func (s *Service) PushFollowingCommented(ctx context.Context, actorID uint, actorUsername string, reviewID uint, movieTitle string) error {
	ref := "review"
	msg := fmt.Sprintf("%s commented on a review of %q", actorUsername, movieTitle)
	return s.PushFollowingActivity(ctx, actorID, NotifFollowingCommented, &reviewID, &ref, msg)
}

func (s *Service) PushAchievementUnlocked(ctx context.Context, userID, achievementID uint, achievementName string, expGained int) error {
	ref := "achievement"
	n := Notification{
		UserID: userID, Type: NotifAchievementUnlocked,
		TargetID: &achievementID, TargetRef: &ref,
		Message: fmt.Sprintf("Achievement unlocked: %s (+%d EXP)", achievementName, expGained),
	}
	return s.createAndEmit(ctx, []Notification{n})
}

func (s *Service) PushEmailVerified(ctx context.Context, userID uint) error {
	n := Notification{UserID: userID, Type: NotifEmailVerified, Message: "Your email has been verified"}
	return s.createAndEmit(ctx, []Notification{n})
}

func (s *Service) PushPasswordChanged(ctx context.Context, userID uint) error {
	n := Notification{
		UserID: userID, Type: NotifPasswordChanged,
		Message: "Your password was changed. If this wasn't you, please contact support immediately.",
	}
	return s.createAndEmit(ctx, []Notification{n})
}

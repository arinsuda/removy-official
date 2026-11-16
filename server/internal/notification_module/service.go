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

// mangaType is the media type whose titles are age-gated (AniList ids).
const mangaType = "manga"

// TitleKey names one title a notification is about.
type TitleKey struct {
	ID   int
	Type string
}

// TitleCheck says, per title, whether a notification about it may reach userID (userID 0: safe for everyone).
// The router wires it to shared.ReceivableMedia: shared imports this package, so this package cannot import shared.
type TitleCheck func(ctx context.Context, userID uint, titles []TitleKey) (map[TitleKey]bool, error)

// MangaRestriction says whether adult and unknown manga must be left out of userID's notification counts and
// pages. The router wires it to shared.NotificationMangaRestricted; until it is set, manga is always restricted (fail closed).
type MangaRestriction func(ctx context.Context, userID uint) (bool, error)

type Service struct {
	repo         *repository
	userProvider UserProvider
	hub          *Hub
	titles       TitleCheck
	restricted   MangaRestriction
}

func NewService(db *gorm.DB, userProvider UserProvider, hub *Hub) *Service {
	return &Service{
		repo:         newRepository(db),
		userProvider: userProvider,
		hub:          hub,
	}
}

// SetTitleCheck wires the age gate for adult manga. Until it is set, every manga title is refused (fail closed).
func (s *Service) SetTitleCheck(check TitleCheck) {
	s.titles = check
}

// SetMangaRestriction wires the viewer rule that the SQL counts and pages apply to manga notifications.
func (s *Service) SetMangaRestriction(check MangaRestriction) {
	s.restricted = check
}

// restrictManga reports whether adult and unknown manga is filtered out of userID's counts and pages in SQL.
func (s *Service) restrictManga(ctx context.Context, userID uint) (bool, error) {
	if s.restricted == nil {
		return true, nil
	}
	return s.restricted(ctx, userID)
}

// receivable reports, per title, whether a notification about it may be sent or shown to userID.
func (s *Service) receivable(ctx context.Context, userID uint, titles []TitleKey) (map[TitleKey]bool, error) {
	if s.titles == nil {
		out := make(map[TitleKey]bool, len(titles))
		for _, t := range titles {
			out[t] = t.Type != mangaType
		}
		return out, nil
	}
	return s.titles(ctx, userID, titles)
}

// mayReceive is the one-title form of receivable. An error reads as refused.
func (s *Service) mayReceive(ctx context.Context, userID uint, t TitleKey) (bool, error) {
	ok, err := s.receivable(ctx, userID, []TitleKey{t})
	return ok[t], err
}

func (s *Service) PushWelcome(ctx context.Context, userID uint) error {
	return s.createAndEmit(ctx, []Notification{{UserID: userID, Type: NotifWelcome}})
}

func (s *Service) ListNotifications(ctx context.Context, userID uint, q ListNotificationsQuery) (*NotificationListResponse, error) {
	restrict, err := s.restrictManga(ctx, userID)
	if err != nil {
		return nil, err
	}
	rows, total, err := s.repo.FindByUser(ctx, userID, q, restrict)
	if err != nil {
		return nil, err
	}
	rows, err = s.visibleRows(ctx, userID, rows)
	if err != nil {
		return nil, err
	}
	unread, err := s.repo.CountUnread(ctx, userID, restrict)
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

// visibleRows drops notifications about titles the viewer may not see, in one batched check. No owner
// exemption applies: the notification is about someone else's action on the title, not the viewer's own item.
func (s *Service) visibleRows(ctx context.Context, userID uint, rows []Notification) ([]Notification, error) {
	titles := make([]TitleKey, 0, len(rows))
	for _, n := range rows {
		if n.Params != nil {
			titles = append(titles, paramsTitle(n.Params))
		}
	}
	ok, err := s.receivable(ctx, userID, titles)
	if err != nil {
		return nil, err
	}
	out := make([]Notification, 0, len(rows))
	for _, n := range rows {
		if n.Params == nil || ok[paramsTitle(n.Params)] {
			out = append(out, n)
		}
	}
	return out, nil
}

func paramsTitle(p *NotificationParams) TitleKey {
	return TitleKey{ID: int(p.MediaID), Type: p.MediaType}
}

func (s *Service) GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	restrict, err := s.restrictManga(ctx, userID)
	if err != nil {
		return 0, err
	}
	return s.repo.CountUnread(ctx, userID, restrict)
}

func (s *Service) GetUnreadCountByCategory(ctx context.Context, userID uint) ([]UnreadByCategoryResponse, error) {
	restrict, err := s.restrictManga(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.CountUnreadByCategory(ctx, userID, restrict)
}

func (s *Service) MarkRead(ctx context.Context, userID uint, req MarkReadRequest) error {
	if err := s.repo.MarkRead(ctx, userID, req.IDs); err != nil {
		return err
	}
	s.hub.EmitRead(userID, req.IDs)
	// The unread badge is a best-effort push: a failed count leaves the client's last value in place.
	if restrict, err := s.restrictManga(ctx, userID); err == nil {
		if count, err := s.repo.CountUnread(ctx, userID, restrict); err == nil {
			s.hub.EmitUnreadCount(userID, count)
		}
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
	// Followers share one notification per activity, so only a title safe for everyone fans out. An adult or
	// unknown manga is skipped for all followers (guests are never eligible), instead of checking each follower.
	if params != nil {
		ok, err := s.mayReceive(ctx, 0, paramsTitle(params))
		if err != nil || !ok {
			return err
		}
	}

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
	resp := baseResponse(n)

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

// baseResponse maps a row to its response without resolving the actor.
func baseResponse(n Notification) NotificationResponse {
	return NotificationResponse{
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
}

func ptr[T any](v T) *T { return &v }

// pushReviewOwner notifies a review's owner about someone else's action on it.
func (s *Service) pushReviewOwner(ctx context.Context, t NotificationType, reviewOwnerID, actorID, reviewID uint, media MediaRef) error {
	if reviewOwnerID == actorID {
		return nil
	}
	// The owner is the only recipient, so an adult title is sent only when the owner may see it.
	ok, err := s.mayReceive(ctx, reviewOwnerID, TitleKey{ID: int(media.ID), Type: media.Type})
	if err != nil || !ok {
		return err
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

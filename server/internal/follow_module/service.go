package follow_module

import (
	"context"

	achievementsmodule "github.com/arinsuda/movie-hub/internal/achievements_module"
	"github.com/arinsuda/movie-hub/internal/feed_module"
	notification_module "github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/internal/shared"
	"gorm.io/gorm"
)

type Service struct {
	repo       *repository
	db         *gorm.DB
	achieveSvc achievementsmodule.Service
	notifSvc   *notification_module.Service
	feedSvc    feed_module.Service
}

func NewService(
	db *gorm.DB,
	achieve achievementsmodule.Service,
	notif *notification_module.Service,
	feed feed_module.Service,
) *Service {
	return &Service{
		repo:       newRepository(db),
		db:         db,
		achieveSvc: achieve,
		notifSvc:   notif,
		feedSvc:    feed,
	}
}

func (s *Service) Follow(ctx context.Context, requesterID, targetID uint) (*FollowResponse, error) {
	follow, err := s.repo.Follow(requesterID, targetID)
	if err != nil {
		return nil, err
	}

	switch follow.Status {
	case StatusAccepted:
		s.onFollowEstablished(ctx, requesterID, targetID)
		s.notifyNewFollower(ctx, requesterID, targetID)
	case StatusPending:
		s.notifyFollowRequested(ctx, requesterID, targetID)
	}

	return &FollowResponse{
		FollowerID: follow.FollowerID,
		FolloweeID: follow.FolloweeID,
		Status:     string(follow.Status),
	}, nil
}

func (s *Service) Unfollow(ctx context.Context, requesterID, targetID uint) error {
	err := s.repo.Unfollow(requesterID, targetID)
	if err != nil {
		return err
	}
	if s.feedSvc != nil {
		_ = s.feedSvc.DeleteFollowActivity(ctx, requesterID, targetID)
	}
	return nil
}

func (s *Service) AcceptFollow(ctx context.Context, requesterID, followerID uint) error {
	if err := s.repo.AcceptFollow(followerID, requesterID); err != nil {
		return err
	}

	s.onFollowEstablished(ctx, followerID, requesterID)
	s.notifyFollowAccepted(ctx, requesterID, followerID)

	return nil
}

func (s *Service) RejectFollow(ctx context.Context, requesterID, followerID uint) error {
	err := s.repo.RejectFollow(followerID, requesterID)
	if err != nil {
		return err
	}
	if s.feedSvc != nil {
		_ = s.feedSvc.DeleteFollowActivity(ctx, followerID, requesterID)
	}
	return nil
}

func (s *Service) GetFollowers(ctx context.Context, requesterID, userID uint) ([]UserSummary, error) {
	canView, err := s.repo.canViewFollowList(requesterID, userID)
	if err != nil {
		return nil, err
	}
	if !canView {
		return nil, ErrForbidden
	}

	rows, err := s.repo.GetFollowers(userID)
	if err != nil {
		return nil, err
	}
	return toSummaryList(rows), nil
}

func (s *Service) GetFollowing(ctx context.Context, requesterID, userID uint) ([]UserSummary, error) {
	canView, err := s.repo.canViewFollowList(requesterID, userID)
	if err != nil {
		return nil, err
	}
	if !canView {
		return nil, ErrForbidden
	}

	rows, err := s.repo.GetFollowing(userID)
	if err != nil {
		return nil, err
	}
	return toSummaryList(rows), nil
}

func (s *Service) GetPendingRequests(ctx context.Context, requesterID, userID uint) ([]UserSummary, error) {
	if requesterID != userID {
		return nil, ErrForbidden
	}
	rows, err := s.repo.GetPendingRequests(userID)
	if err != nil {
		return nil, err
	}
	return toSummaryList(rows), nil
}

func (s *Service) GetRelationshipStatus(ctx context.Context, requesterID, targetID uint) (*RelationshipStatus, error) {
	outbound, err := s.repo.GetStatus(requesterID, targetID)
	if err != nil {
		return nil, err
	}
	inbound, err := s.repo.GetStatus(targetID, requesterID)
	if err != nil {
		return nil, err
	}

	status := &RelationshipStatus{}
	if outbound != nil {
		status.IsFollowing = outbound.Status == StatusAccepted
		status.FollowStatus = string(outbound.Status)
	}
	if inbound != nil {
		status.IsFollowedBy = inbound.Status == StatusAccepted
	}
	return status, nil
}

func (s *Service) onFollowEstablished(ctx context.Context, followerID, followeeID uint) {
	var followingCount int64
	s.db.Model(&UserFollow{}).
		Where("follower_id = ? AND status = ?", followerID, StatusAccepted).
		Count(&followingCount)
	shared.TrackAndNotify(ctx, s.achieveSvc, s.notifSvc, followerID, "following_count", int(followingCount))

	var followerCount int64
	s.db.Model(&UserFollow{}).
		Where("followee_id = ? AND status = ?", followeeID, StatusAccepted).
		Count(&followerCount)
	shared.TrackAndNotify(ctx, s.achieveSvc, s.notifSvc, followeeID, "follower_count", int(followerCount))

	if s.feedSvc != nil {
		_ = s.feedSvc.CreateActivity(ctx, followerID, feed_module.ActivityUserFollowed, feed_module.ActivityPayload{
			TargetUserID: &followeeID,
		})
	}
}

func (s *Service) notifyNewFollower(ctx context.Context, followerID, followeeID uint) {
	if s.notifSvc == nil {
		return
	}
	_ = s.notifSvc.PushFollowedYou(ctx, followeeID, followerID)
}

func (s *Service) notifyFollowRequested(ctx context.Context, followerID, followeeID uint) {
	if s.notifSvc == nil {
		return
	}
	_ = s.notifSvc.PushFollowRequested(ctx, followeeID, followerID)
}

func (s *Service) notifyFollowAccepted(ctx context.Context, accepterID, followerID uint) {
	if s.notifSvc == nil {
		return
	}
	_ = s.notifSvc.PushFollowAccepted(ctx, followerID, accepterID)
}

func (s *Service) GetFollowStats(ctx context.Context, requesterID, targetID uint) (*FollowStatsResponse, error) {
	followers, err := s.repo.CountFollowers(targetID)
	if err != nil {
		return nil, err
	}
	following, err := s.repo.CountFollowing(targetID)
	if err != nil {
		return nil, err
	}

	isFollowing := false
	if requesterID != targetID {
		status, err := s.repo.GetStatus(requesterID, targetID)
		if err != nil {
			return nil, err
		}
		isFollowing = status != nil && status.Status == StatusAccepted
	}

	return &FollowStatsResponse{
		UserID:      targetID,
		Followers:   followers,
		Following:   following,
		IsFollowing: isFollowing,
	}, nil
}

func toSummaryList(rows []listRow) []UserSummary {
	result := make([]UserSummary, len(rows))
	for i, r := range rows {
		result[i] = UserSummary{
			ID:          r.ID,
			Username:    r.Username,
			DisplayName: r.DisplayName,
			AvatarURL:   r.AvatarURL,
			Status:      r.Status,
		}
	}
	return result
}

package admin_module

import (
	"context"
	"log"
	"strings"
	"time"

	notification_module "github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/internal/shared/storage"
)

type Service struct {
	repo  Repository
	hub   *notification_module.Hub
	minio *storage.MinIOClient
}

func NewService(repo Repository, hub *notification_module.Hub, minio *storage.MinIOClient) *Service {
	return &Service{
		repo:  repo,
		hub:   hub,
		minio: minio,
	}
}

func (s *Service) GetOverview() (*OverviewStats, error) {
	onlineCount := 0
	if s.hub != nil {
		onlineCount = s.hub.UniqueOnlineCount()
	}
	return s.repo.GetOverviewStats(onlineCount)
}

func (s *Service) GetGrowth() ([]GrowthPoint, error) {
	return s.repo.Get12MonthGrowthTrend()
}

func (s *Service) ListUsers(filter UserFilter) (PaginatedResponse[AdminUserRow], error) {
	items, total, err := s.repo.ListUsers(filter)
	if err != nil {
		return PaginatedResponse[AdminUserRow]{}, err
	}
	for i := range items {
		items[i].AvatarURL = presignAvatar(items[i].AvatarURL, s.minio)
	}
	return NewPaginatedResponse(items, total, filter.Page, filter.Limit), nil
}

func (s *Service) ListReviews(filter ReviewFilter) (PaginatedResponse[AdminReviewRow], error) {
	items, total, err := s.repo.ListReviews(filter)
	if err != nil {
		return PaginatedResponse[AdminReviewRow]{}, err
	}
	return NewPaginatedResponse(items, total, filter.Page, filter.Limit), nil
}

func (s *Service) ListAuditLogs(filter AuditLogFilter) (PaginatedResponse[AdminAuditLogRow], error) {
	items, total, err := s.repo.ListAuditLogs(filter)
	if err != nil {
		return PaginatedResponse[AdminAuditLogRow]{}, err
	}
	return NewPaginatedResponse(items, total, filter.Page, filter.Limit), nil
}

func (s *Service) UpdateUserRole(adminID, targetUserID uint, req UpdateRoleRequest) error {
	if req.Role != "admin" && req.Role != "user" {
		return ErrRoleNotFound
	}
	return s.repo.UpdateUserRole(adminID, targetUserID, req.Role, req.Reason)
}

func (s *Service) UpdateUserStatus(adminID, targetUserID uint, req UpdateStatusRequest) error {
	return s.repo.UpdateUserStatus(adminID, targetUserID, req.IsActive, req.Reason)
}

func (s *Service) DeleteUser(adminID, targetUserID uint, req DeleteUserRequest) error {
	return s.repo.DeleteUser(adminID, targetUserID, req.Reason)
}

// erasureCleanupTimeout bounds each MinIO call of the post-commit cleanup.
const erasureCleanupTimeout = 15 * time.Second

// EraseUser permanently removes a soft-deleted user. Storage cleanup runs after the commit; a
// failure is logged (keys only, no PII) and left as storage_cleanup "pending" in the audit row,
// because the account is already gone.
func (s *Service) EraseUser(adminID, targetUserID uint, req EraseUserRequest) error {
	res, out, err := s.repo.EraseUser(adminID, targetUserID, req.ConfirmUsername, req.Reason)
	if err != nil {
		return err
	}
	out.Publish(targetUserID)
	if s.minio == nil {
		return nil
	}
	ok := true
	for _, prefix := range res.AvatarPrefixes {
		ok = s.cleanup(prefix, s.minio.DeletePrefix) && ok
	}
	ok = s.cleanup(res.AvatarKey, s.minio.DeleteObject) && ok
	if ok {
		if err := s.repo.SetErasureCleanup(res.AuditID, CleanupDone); err != nil {
			log.Printf("admin: erase user %d: record cleanup failed: %v", targetUserID, err)
		}
	}
	return nil
}

// cleanup runs one storage call under its own timeout.
func (s *Service) cleanup(key string, del func(context.Context, string) error) bool {
	if key == "" {
		return true
	}
	ctx, cancel := context.WithTimeout(context.Background(), erasureCleanupTimeout)
	defer cancel()
	if err := del(ctx, key); err != nil {
		log.Printf("admin: erase: storage cleanup of %q failed: %v", key, err)
		return false
	}
	return true
}

func (s *Service) DeleteReview(adminID, reviewID uint, req DeleteReviewRequest) error {
	return s.repo.DeleteReview(adminID, reviewID, req.Reason)
}

func presignAvatar(avatarURL *string, minio *storage.MinIOClient) *string {
	if avatarURL == nil || *avatarURL == "" || minio == nil {
		return avatarURL
	}
	if strings.HasPrefix(*avatarURL, "http://") || strings.HasPrefix(*avatarURL, "https://") {
		return avatarURL
	}
	if presigned, err := minio.PresignURL(context.Background(), *avatarURL); err == nil {
		return &presigned
	}
	return avatarURL
}

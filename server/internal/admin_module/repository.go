package admin_module

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/arinsuda/movie-hub/internal/identity_policy"
	tmdbmodule "github.com/arinsuda/movie-hub/internal/tmdb_module"
	"github.com/arinsuda/movie-hub/internal/user_module"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrReviewNotFound         = errors.New("review not found")
	ErrRoleNotFound           = errors.New("invalid role")
	ErrSelfDeactivation       = errors.New("cannot deactivate your own account")
	ErrSelfDeletion           = errors.New("cannot delete your own account")
	ErrFinalAdminProtection   = errors.New("cannot demote or deactivate final active admin")
	ErrInactiveUserRoleChange = errors.New("cannot change role of inactive user")
	ErrUserAlreadyInStatus    = errors.New("user is already in requested status")
)

type Repository interface {
	GetOverviewStats(onlineUsers int) (*OverviewStats, error)
	Get12MonthGrowthTrend() ([]GrowthPoint, error)
	ListUsers(filter UserFilter) ([]AdminUserRow, int64, error)
	ListReviews(filter ReviewFilter) ([]AdminReviewRow, int64, error)
	ListAuditLogs(filter AuditLogFilter) ([]AdminAuditLogRow, int64, error)
	UpdateUserRole(adminID, targetUserID uint, newRoleName string, reason *string) error
	UpdateUserStatus(adminID, targetUserID uint, isActive bool, reason *string) error
	DeleteUser(adminID, targetUserID uint, reason *string) error
	DeleteReview(adminID, reviewID uint, reason *string) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

var _ Repository = (*gormRepository)(nil)

func (r *gormRepository) GetOverviewStats(onlineUsers int) (*OverviewStats, error) {
	now := time.Now().UTC()
	startOfCurrentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	startOfPrevMonth := startOfCurrentMonth.AddDate(0, -1, 0)
	endOfPrevMonth := startOfCurrentMonth.Add(-time.Nanosecond)
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	sevenDaysAgo := now.AddDate(0, 0, -7)
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	var totalReg, activeUsers, inactiveUsers int64
	if err := r.db.Table("users").Where("deleted_at IS NULL").Count(&totalReg).Error; err != nil {
		return nil, err
	}
	if err := r.db.Table("users").Where("deleted_at IS NULL AND is_active = true").Count(&activeUsers).Error; err != nil {
		return nil, err
	}
	if err := r.db.Table("users").Where("deleted_at IS NULL AND is_active = false").Count(&inactiveUsers).Error; err != nil {
		return nil, err
	}

	var currentMonthReg, prevMonthReg int64
	if err := r.db.Table("users").
		Where("created_at >= ? AND created_at <= ?", startOfCurrentMonth, now).
		Count(&currentMonthReg).Error; err != nil {
		return nil, err
	}
	if err := r.db.Table("users").
		Where("created_at >= ? AND created_at <= ?", startOfPrevMonth, endOfPrevMonth).
		Count(&prevMonthReg).Error; err != nil {
		return nil, err
	}

	absGrowth := currentMonthReg - prevMonthReg
	var growthPct *float64
	var status GrowthStatus

	if prevMonthReg == 0 {
		if currentMonthReg == 0 {
			zero := 0.0
			growthPct = &zero
			status = GrowthStatusNoChange
		} else {
			growthPct = nil
			status = GrowthStatusNoPreviousBaseline
		}
	} else {
		val := (float64(absGrowth) / float64(prevMonthReg)) * 100.0
		growthPct = &val
		if absGrowth > 0 {
			status = GrowthStatusGrowth
		} else if absGrowth < 0 {
			status = GrowthStatusDecline
		} else {
			status = GrowthStatusNoChange
		}
	}

	var totalActivities, activitiesToday, dau, wau, mau int64
	if err := r.db.Table("activity_events").Where("deleted_at IS NULL").Count(&totalActivities).Error; err != nil {
		totalActivities = 0
	}
	if err := r.db.Table("activity_events").Where("deleted_at IS NULL AND created_at >= ?", startOfToday).Count(&activitiesToday).Error; err != nil {
		activitiesToday = 0
	}
	if err := r.db.Table("activity_events").Where("deleted_at IS NULL AND created_at >= ?", startOfToday).Select("COUNT(DISTINCT actor_id)").Scan(&dau).Error; err != nil {
		dau = 0
	}
	if err := r.db.Table("activity_events").Where("deleted_at IS NULL AND created_at >= ?", sevenDaysAgo).Select("COUNT(DISTINCT actor_id)").Scan(&wau).Error; err != nil {
		wau = 0
	}
	if err := r.db.Table("activity_events").Where("deleted_at IS NULL AND created_at >= ?", thirtyDaysAgo).Select("COUNT(DISTINCT actor_id)").Scan(&mau).Error; err != nil {
		mau = 0
	}

	var totalReviews, totalLikes int64
	if err := r.db.Table("reviews").Where("deleted_at IS NULL").Count(&totalReviews).Error; err != nil {
		totalReviews = 0
	}
	if err := r.db.Table("media_likes").Where("deleted_at IS NULL").Count(&totalLikes).Error; err != nil {
		totalLikes = 0
	}

	return &OverviewStats{
		TotalRegisteredUsers:       totalReg,
		ActiveUsersCount:           activeUsers,
		InactiveUsersCount:         inactiveUsers,
		CurrentMonthRegistrations:  currentMonthReg,
		PreviousMonthRegistrations: prevMonthReg,
		AbsoluteGrowth:             absGrowth,
		GrowthPercentage:           growthPct,
		GrowthStatus:               status,
		UniqueOnlineUsers:          onlineUsers,
		TotalActivityEvents:        totalActivities,
		ActivityEventsToday:        activitiesToday,
		DauToday:                   dau,
		Wau7d:                      wau,
		Mau30d:                     mau,
		TotalReviews:               totalReviews,
		TotalMediaLikes:            totalLikes,
	}, nil
}

func (r *gormRepository) Get12MonthGrowthTrend() ([]GrowthPoint, error) {
	now := time.Now().UTC()
	startMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -11, 0)

	type monthCount struct {
		Month string `gorm:"column:month"`
		Count int64  `gorm:"column:count"`
	}
	var rows []monthCount

	err := r.db.Table("users").
		Select("to_char(created_at, 'YYYY-MM') AS month, COUNT(*) AS count").
		Where("created_at >= ?", startMonth).
		Group("month").
		Order("month ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	countMap := make(map[string]int64)
	for _, r := range rows {
		countMap[r.Month] = r.Count
	}

	result := make([]GrowthPoint, 12)
	curr := startMonth
	for i := 0; i < 12; i++ {
		mStr := curr.Format("2006-01")
		result[i] = GrowthPoint{
			Month:     mStr,
			UserCount: countMap[mStr],
		}
		curr = curr.AddDate(0, 1, 0)
	}

	return result, nil
}

func (r *gormRepository) ListUsers(filter UserFilter) ([]AdminUserRow, int64, error) {
	filter.Normalize()

	query := r.db.Table("users").
		Select(`
			users.id,
			users.username,
			users.email,
			users.display_name,
			users.avatar_url,
			roles.role_name AS role,
			users.is_active,
			users.created_at,
			COALESCE(s.review_count, 0) AS review_count
		`).
		Joins("JOIN roles ON roles.id = users.role_id").
		Joins("LEFT JOIN user_stats s ON s.user_id = users.id").
		Where("users.deleted_at IS NULL")

	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		query = query.Where("users.username ILIKE ? OR users.email ILIKE ?", s, s)
	}
	if filter.Role != "" && filter.Role != "all" {
		query = query.Where("roles.role_name = ?", filter.Role)
	}
	if filter.Status != "" && filter.Status != "all" {
		if filter.Status == "active" {
			query = query.Where("users.is_active = true")
		} else if filter.Status == "inactive" {
			query = query.Where("users.is_active = false")
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := fmt.Sprintf("users.%s %s", filter.SortBy, filter.SortOrder)
	if filter.SortBy == "review_count" {
		orderClause = fmt.Sprintf("review_count %s", filter.SortOrder)
	}

	type dbUserRow struct {
		ID          uint
		Username    string
		Email       string
		DisplayName *string
		AvatarURL   *string
		Role        string
		IsActive    bool
		CreatedAt   time.Time
		ReviewCount int
	}

	var rows []dbUserRow
	err := query.Order(orderClause).
		Offset(filter.Offset()).
		Limit(filter.Limit).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	result := make([]AdminUserRow, len(rows))
	for i, r := range rows {
		result[i] = AdminUserRow{
			ID:          r.ID,
			Username:    r.Username,
			Email:       r.Email,
			DisplayName: r.DisplayName,
			AvatarURL:   r.AvatarURL,
			Role:        r.Role,
			IsActive:    r.IsActive,
			CreatedAt:   r.CreatedAt.Format(time.RFC3339),
			ReviewCount: r.ReviewCount,
		}
	}

	return result, total, nil
}

func (r *gormRepository) ListReviews(filter ReviewFilter) ([]AdminReviewRow, int64, error) {
	filter.Normalize()

	query := r.db.Table("reviews").
		Select(`
			reviews.id,
			reviews.user_id,
			users.username,
			reviews.media_id,
			reviews.media_type,
			reviews.rating,
			reviews.body,
			reviews.is_public,
			reviews.like_count,
			reviews.created_at
		`).
		Joins("JOIN users ON users.id = reviews.user_id").
		Where("reviews.deleted_at IS NULL")

	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		query = query.Where("reviews.body ILIKE ? OR users.username ILIKE ?", s, s)
	}
	if filter.MediaType != "" && filter.MediaType != "all" {
		query = query.Where("reviews.media_type = ?", filter.MediaType)
	}
	if filter.Visibility != "" && filter.Visibility != "all" {
		if filter.Visibility == "public" {
			query = query.Where("reviews.is_public = true")
		} else if filter.Visibility == "private" {
			query = query.Where("reviews.is_public = false")
		}
	}
	if filter.MinRating != nil {
		query = query.Where("reviews.rating >= ?", *filter.MinRating)
	}
	if filter.MaxRating != nil {
		query = query.Where("reviews.rating <= ?", *filter.MaxRating)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := fmt.Sprintf("reviews.%s %s", filter.SortBy, filter.SortOrder)

	type dbReviewRow struct {
		ID        uint
		UserID    uint
		Username  string
		MediaID   int
		MediaType string
		Rating    float32
		Body      string
		IsPublic  bool
		LikeCount int
		CreatedAt time.Time
	}

	var rows []dbReviewRow
	err := query.Order(orderClause).
		Offset(filter.Offset()).
		Limit(filter.Limit).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	result := make([]AdminReviewRow, len(rows))
	for i, r := range rows {
		title, poster := fetchMediaTitleAndPoster(r.MediaID, r.MediaType)
		result[i] = AdminReviewRow{
			ID:         r.ID,
			UserID:     r.UserID,
			Username:   r.Username,
			MediaID:    r.MediaID,
			MediaType:  r.MediaType,
			MediaTitle: title,
			PosterURL:  poster,
			Rating:     r.Rating,
			Body:       r.Body,
			IsPublic:   r.IsPublic,
			LikeCount:  r.LikeCount,
			CreatedAt:  r.CreatedAt.Format(time.RFC3339),
		}
	}

	return result, total, nil
}

func fetchMediaTitleAndPoster(mediaID int, mediaType string) (title, posterURL string) {
	if mediaType == "movie" {
		m, err := tmdbmodule.GetMovieByID(mediaID)
		if err == nil && m != nil {
			t := m.EnglishTitle
			if t == "" {
				t = m.Title
			}
			return t, tmdbmodule.ImageURL(m.PosterPath)
		}
	} else if mediaType == "tv" {
		s, err := tmdbmodule.GetSeriesByID(mediaID)
		if err == nil && s != nil {
			t := s.EnglishTitle
			if t == "" {
				t = s.Name
			}
			return t, tmdbmodule.ImageURL(s.PosterPath)
		}
	}
	return "", ""
}

func (r *gormRepository) ListAuditLogs(filter AuditLogFilter) ([]AdminAuditLogRow, int64, error) {
	filter.Normalize()

	query := r.db.Table("admin_audit_logs").
		Select(`
			admin_audit_logs.id,
			admin_audit_logs.admin_id,
			users.username AS admin_username,
			admin_audit_logs.action,
			admin_audit_logs.target_type,
			admin_audit_logs.target_id,
			admin_audit_logs.reason,
			admin_audit_logs.meta_data,
			admin_audit_logs.created_at
		`).
		Joins("JOIN users ON users.id = admin_audit_logs.admin_id")

	if filter.Action != "" && filter.Action != "all" {
		query = query.Where("admin_audit_logs.action = ?", filter.Action)
	}
	if filter.TargetType != "" && filter.TargetType != "all" {
		query = query.Where("admin_audit_logs.target_type = ?", filter.TargetType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type dbAuditRow struct {
		ID            uint
		AdminID       uint
		AdminUsername string
		Action        AuditAction
		TargetType    string
		TargetID      uint
		Reason        *string
		MetaData      datatypes.JSON
		CreatedAt     time.Time
	}

	var rows []dbAuditRow
	err := query.Order("admin_audit_logs.created_at DESC, admin_audit_logs.id DESC").
		Offset(filter.Offset()).
		Limit(filter.Limit).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	result := make([]AdminAuditLogRow, len(rows))
	for i, r := range rows {
		result[i] = AdminAuditLogRow{
			ID:            r.ID,
			AdminID:       r.AdminID,
			AdminUsername: r.AdminUsername,
			Action:        r.Action,
			TargetType:    r.TargetType,
			TargetID:      r.TargetID,
			Reason:        r.Reason,
			MetaData:      r.MetaData,
			CreatedAt:     r.CreatedAt.Format(time.RFC3339),
		}
	}

	return result, total, nil
}

func (r *gormRepository) UpdateUserRole(adminID, targetUserID uint, newRoleName string, reason *string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Before any row lock: see identity_policy.LockAdminGuard.
		if err := identity_policy.LockAdminGuard(tx); err != nil {
			return err
		}
		type userRoleState struct {
			ID       uint
			IsActive bool
			RoleName string
		}
		var target userRoleState
		err := tx.Table("users").
			Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: "users"}}).
			Select("users.id, users.is_active, roles.role_name").
			Joins("JOIN roles ON roles.id = users.role_id").
			Where("users.id = ? AND users.deleted_at IS NULL", targetUserID).
			Scan(&target).Error
		if err != nil || target.ID == 0 {
			return ErrUserNotFound
		}

		if !target.IsActive {
			return ErrInactiveUserRoleChange
		}

		var newRoleID uint
		if err := tx.Table("roles").Select("id").Where("role_name = ?", newRoleName).Scan(&newRoleID).Error; err != nil || newRoleID == 0 {
			return ErrRoleNotFound
		}

		oldRoleName := target.RoleName
		if oldRoleName == newRoleName {
			return nil
		}

		if oldRoleName == string(user_module.RoleAdmin) && newRoleName != string(user_module.RoleAdmin) {
			activeAdminCount, err := identity_policy.CountActiveAdmins(tx, string(user_module.RoleAdmin))
			if err != nil {
				return err
			}
			if activeAdminCount <= 1 {
				return ErrFinalAdminProtection
			}
		}

		if err := tx.Table("users").Where("id = ?", targetUserID).Update("role_id", newRoleID).Error; err != nil {
			return err
		}

		metaMap := map[string]string{
			"old_role": oldRoleName,
			"new_role": newRoleName,
		}
		metaJS, _ := json.Marshal(metaMap)

		auditLog := AdminAuditLog{
			AdminID:    adminID,
			Action:     ActionUserRoleChanged,
			TargetType: "user",
			TargetID:   targetUserID,
			Reason:     reason,
			MetaData:   datatypes.JSON(metaJS),
			CreatedAt:  time.Now().UTC(),
		}
		return tx.Create(&auditLog).Error
	})
}

func (r *gormRepository) UpdateUserStatus(adminID, targetUserID uint, isActive bool, reason *string) error {
	if adminID == targetUserID && !isActive {
		return ErrSelfDeactivation
	}

	var out identity_policy.Outcome
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Before any row lock: see identity_policy.LockAdminGuard.
		if err := identity_policy.LockAdminGuard(tx); err != nil {
			return err
		}
		type userRoleState struct {
			ID       uint
			IsActive bool
			RoleName string
		}
		var target userRoleState
		err := tx.Table("users").
			Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: "users"}}).
			Select("users.id, users.is_active, roles.role_name").
			Joins("JOIN roles ON roles.id = users.role_id").
			Where("users.id = ? AND users.deleted_at IS NULL", targetUserID).
			Scan(&target).Error
		if err != nil || target.ID == 0 {
			return ErrUserNotFound
		}

		if target.IsActive == isActive {
			return ErrUserAlreadyInStatus
		}

		if !isActive && target.RoleName == string(user_module.RoleAdmin) {
			activeAdminCount, err := identity_policy.CountActiveAdmins(tx, string(user_module.RoleAdmin))
			if err != nil {
				return err
			}
			if activeAdminCount <= 1 {
				return ErrFinalAdminProtection
			}
		}

		// Any status change ends the target's sessions. Reactivation matters too: accounts
		// deactivated before sessions were revoked on deactivation may still hold tokens.
		if err := tx.Table("users").Where("id = ?", targetUserID).Updates(map[string]any{
			"is_active":     isActive,
			"token_version": gorm.Expr("token_version + 1"),
		}).Error; err != nil {
			return err
		}
		if err := identity_policy.RevokeSessionScope(tx, targetUserID); err != nil {
			return err
		}
		out.SessionsRevoked = true

		action := ActionUserDeactivated
		if isActive {
			action = ActionUserReactivated
		}

		metaMap := map[string]bool{
			"is_active": isActive,
		}
		metaJS, _ := json.Marshal(metaMap)

		auditLog := AdminAuditLog{
			AdminID:    adminID,
			Action:     action,
			TargetType: "user",
			TargetID:   targetUserID,
			Reason:     reason,
			MetaData:   datatypes.JSON(metaJS),
			CreatedAt:  time.Now().UTC(),
		}
		return tx.Create(&auditLog).Error
	})
	if err != nil {
		return err
	}
	out.Publish(targetUserID)
	return nil
}

func (r *gormRepository) DeleteUser(adminID, targetUserID uint, reason *string) error {
	if adminID == targetUserID {
		return ErrSelfDeletion
	}

	var out identity_policy.Outcome
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Before any row lock: see identity_policy.LockAdminGuard.
		if err := identity_policy.LockAdminGuard(tx); err != nil {
			return err
		}
		type userRoleState struct {
			ID       uint
			IsActive bool
			RoleName string
			Username string
			Email    string
		}
		var target userRoleState
		err := tx.Table("users").
			Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: "users"}}).
			Select("users.id, users.is_active, roles.role_name, users.username, users.email").
			Joins("JOIN roles ON roles.id = users.role_id").
			Where("users.id = ? AND users.deleted_at IS NULL", targetUserID).
			Scan(&target).Error
		if err != nil || target.ID == 0 {
			return ErrUserNotFound
		}

		if target.RoleName == string(user_module.RoleAdmin) {
			activeAdminCount, err := identity_policy.CountActiveAdmins(tx, string(user_module.RoleAdmin))
			if err != nil {
				return err
			}
			if activeAdminCount <= 1 {
				return ErrFinalAdminProtection
			}
		}

		now := time.Now().UTC()

		if err := tx.Table("users").Where("id = ?", targetUserID).Updates(map[string]any{
			"deleted_at":    now,
			"is_active":     false,
			"token_version": gorm.Expr("token_version + 1"),
		}).Error; err != nil {
			return err
		}
		if err := identity_policy.RevokeSessionScope(tx, targetUserID); err != nil {
			return err
		}
		out.SessionsRevoked = true

		_ = tx.Table("reviews").Where("user_id = ? AND deleted_at IS NULL", targetUserID).Update("deleted_at", now).Error
		_ = tx.Table("activity_events").Where("actor_id = ? AND deleted_at IS NULL", targetUserID).Update("deleted_at", now).Error
		_ = tx.Table("media_watch_logs").Where("user_id = ? AND deleted_at IS NULL", targetUserID).Update("deleted_at", now).Error
		_ = tx.Table("library_items").Where("user_id = ? AND deleted_at IS NULL", targetUserID).Update("deleted_at", now).Error

		metaMap := map[string]any{
			"target_username": target.Username,
			"target_email":    target.Email,
			"target_role":     target.RoleName,
		}
		metaJS, _ := json.Marshal(metaMap)

		auditLog := AdminAuditLog{
			AdminID:    adminID,
			Action:     ActionUserDeleted,
			TargetType: "user",
			TargetID:   targetUserID,
			Reason:     reason,
			MetaData:   datatypes.JSON(metaJS),
			CreatedAt:  now,
		}
		return tx.Create(&auditLog).Error
	})
	if err != nil {
		return err
	}
	out.Publish(targetUserID)
	return nil
}

func (r *gormRepository) DeleteReview(adminID, reviewID uint, reason *string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		type reviewRow struct {
			ID     uint
			UserID uint
			Body   string
			Rating float32
		}
		var rev reviewRow
		err := tx.Table("reviews").
			Select("id, user_id, body, rating").
			Where("id = ? AND deleted_at IS NULL", reviewID).
			Scan(&rev).Error
		if err != nil || rev.ID == 0 {
			return ErrReviewNotFound
		}

		if err := tx.Table("reviews").Where("id = ?", reviewID).Update("deleted_at", time.Now().UTC()).Error; err != nil {
			return err
		}

		metaMap := map[string]any{
			"author_user_id": rev.UserID,
			"rating":         rev.Rating,
		}
		metaJS, _ := json.Marshal(metaMap)

		auditLog := AdminAuditLog{
			AdminID:    adminID,
			Action:     ActionReviewDeleted,
			TargetType: "review",
			TargetID:   reviewID,
			Reason:     reason,
			MetaData:   datatypes.JSON(metaJS),
			CreatedAt:  time.Now().UTC(),
		}
		return tx.Create(&auditLog).Error
	})
}

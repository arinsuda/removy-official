package user_module

import (
	"errors"
	"strings"

	"github.com/arinsuda/movie-hub/internal/identity_policy"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound               = errors.New("user not found")
	ErrForbidden                  = errors.New("forbidden")
	ErrInvalidCredentials         = errors.New("invalid credentials")
	ErrOTPNotFound                = errors.New("otp not found")
	ErrOTPExpired                 = errors.New("otp expired")
	ErrOTPInvalid                 = errors.New("otp invalid")
	ErrOTPMaxAttempts             = errors.New("otp max attempts exceeded")
	ErrEmailAlreadyInUse          = errors.New("email already in use")
	ErrInvalidEmailFormat         = errors.New("invalid email format")
	ErrEmailSameAsCurrent         = errors.New("new email must be different from current email")
	ErrPendingEmailMissing        = errors.New("pending new email missing")
	ErrPasswordResetTokenNotFound = errors.New("password reset token not found")
	ErrPasswordResetTokenExpired  = errors.New("password reset token expired")
	ErrPasswordResetTokenInvalid  = errors.New("password reset token invalid")
	ErrPasswordNotSet             = errors.New("no password set, use password reset to create one")
	ErrSessionStale               = errors.New("session is no longer valid")
	ErrFinalAdmin                 = errors.New("cannot delete the final active admin")
	ErrWrongPassword              = errors.New("incorrect password")
	ErrConfirmMismatch            = errors.New("username confirmation does not match")
)

type repository struct {
	db *gorm.DB
}

func newRepository(db *gorm.DB) *repository {
	return &repository{db: db}
}

func NewPublicRepository(db *gorm.DB) *repository {
	return newRepository(db)
}

type userWithStats struct {
	User
	ReviewCount    int `gorm:"column:review_count"`
	FollowerCount  int `gorm:"column:follower_count"`
	FollowingCount int `gorm:"column:following_count"`
}

func (r *repository) FindByID(id uint) (*User, int, int, int, error) {
	var row userWithStats

	err := r.db.
		Table("users").
		Select(`
			users.*,
			COALESCE(s.review_count,   0) AS review_count,
			COALESCE(s.follower_count,  0) AS follower_count,
			COALESCE(s.following_count, 0) AS following_count
		`).
		Joins("LEFT JOIN user_stats s ON s.user_id = users.id").
		Joins("LEFT JOIN roles ON roles.id = users.role_id").
		Where("users.id = ? AND users.is_active = true", id).
		First(&row).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, 0, 0, 0, ErrUserNotFound
	}
	if err != nil {
		return nil, 0, 0, 0, err
	}

	if err := r.db.First(&row.User.Role, row.User.RoleID).Error; err != nil {
		return nil, 0, 0, 0, err
	}

	return &row.User, row.ReviewCount, row.FollowerCount, row.FollowingCount, nil
}

func (r *repository) FindByEmail(email string) (*User, error) {
	var user User
	err := r.db.
		Where("lower(email) = lower(?) AND is_active = true", strings.TrimSpace(email)).
		Order("(verified_email_at IS NOT NULL) DESC, id ASC").
		First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := r.db.First(&user.Role, user.RoleID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// findCredentials returns the active user's username and password hash (nil when none is set).
func (r *repository) findCredentials(id uint) (username string, password *string, err error) {
	var row struct {
		Username string
		Password *string
	}
	res := r.db.Raw(`SELECT username, password FROM users WHERE id = ? AND deleted_at IS NULL`, id).Scan(&row)
	if res.Error != nil {
		return "", nil, res.Error
	}
	if res.RowsAffected == 0 {
		return "", nil, ErrUserNotFound
	}
	return row.Username, row.Password, nil
}

func (r *repository) UpdateProfile(id uint, updates map[string]any) error {
	result := r.db.Model(&User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// DeleteUser deactivates the account and revokes its sessions in one transaction.
// The last active admin cannot remove itself.
func (r *repository) DeleteUser(id uint, sessionVersion int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := identity_policy.LockAdminGuard(tx); err != nil {
			return err
		}
		if err := identity_policy.EnsureSession(tx, id, sessionVersion, true); err != nil {
			if errors.Is(err, identity_policy.ErrSessionStale) {
				return ErrSessionStale
			}
			return err
		}

		var roleName string
		if err := tx.Table("users").
			Select("roles.role_name").
			Joins("JOIN roles ON roles.id = users.role_id").
			Where("users.id = ? AND users.deleted_at IS NULL", id).
			Scan(&roleName).Error; err != nil {
			return err
		}
		if roleName == string(RoleAdmin) {
			activeAdmins, err := identity_policy.CountActiveAdmins(tx, string(RoleAdmin))
			if err != nil {
				return err
			}
			if activeAdmins <= 1 {
				return ErrFinalAdmin
			}
		}

		result := tx.Exec(
			"UPDATE users SET is_active = false, token_version = token_version + 1, updated_at = now() WHERE id = ? AND deleted_at IS NULL",
			id,
		)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrUserNotFound
		}
		return identity_policy.RevokeSessionScope(tx, id)
	})
}

func (r *repository) UpsertEmailChangeRequest(req *EmailChangeRequest) error {
	return r.db.
		Where(EmailChangeRequest{UserID: req.UserID}).
		Assign(EmailChangeRequest{
			NewEmail:     req.NewEmail,
			OTPHash:      req.OTPHash,
			NewOTPHash:   req.NewOTPHash,
			ExpiresAt:    req.ExpiresAt,
			AttemptCount: 0,
		}).
		FirstOrCreate(req).Error
}

func (r *repository) FindEmailChangeRequest(userID uint) (*EmailChangeRequest, error) {
	var req EmailChangeRequest
	err := r.db.Where("user_id = ?", userID).First(&req).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrOTPNotFound
	}
	return &req, err
}

func (r *repository) IncrementOTPAttempt(id uint) error {
	return r.db.Model(&EmailChangeRequest{}).
		Where("id = ?", id).
		UpdateColumn("attempt_count", gorm.Expr("attempt_count + 1")).Error
}

func (r *repository) DeleteEmailChangeRequest(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&EmailChangeRequest{}).Error
}

func (r *repository) IsEmailTaken(email string, excludeUserID uint) (bool, error) {
	var count int64
	err := r.db.Model(&User{}).
		Where("lower(email) = lower(?) AND id != ? AND is_active = true", email, excludeUserID).
		Count(&count).Error
	return count > 0, err
}

func (r *repository) UpsertPasswordResetToken(record *PasswordResetToken) error {

	r.db.Where("user_id = ?", record.UserID).Delete(&PasswordResetToken{})
	return r.db.Create(record).Error
}

func (r *repository) FindPasswordResetTokenByUserID(userID uint) (*PasswordResetToken, error) {
	var record PasswordResetToken
	err := r.db.Where("user_id = ?", userID).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPasswordResetTokenNotFound
	}
	return &record, err
}

func (r *repository) DeletePasswordResetToken(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&PasswordResetToken{}).Error
}

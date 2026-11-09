package auth_module

import (
	"errors"
	"strings"
	"time"

	"github.com/arinsuda/movie-hub/internal/user_module"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func newRepository(db *gorm.DB) *repository {
	return &repository{db: db}
}

func (r *repository) CreateUser(user *user_module.User) error {
	err := r.db.Create(user).Error
	if err != nil {
		errStr := strings.ToLower(err.Error())
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "23505") {
			if strings.Contains(errStr, "username") {
				return ErrUsernameTaken
			}
			return ErrEmailTaken
		}
	}
	return err
}

func (r *repository) FindByUsername(username string) (*user_module.User, error) {
	return r.findUser("username = ?", username)
}

// FindByEmail matches case-insensitively. Mixed-case duplicates can exist until the
// unique index is case-insensitive, so the verified row wins, then the oldest.
func (r *repository) FindByEmail(email string) (*user_module.User, error) {
	var user user_module.User
	err := r.db.Preload("Role").
		Where("lower(email) = ?", NormalizeEmail(email)).
		Order("(verified_email_at IS NOT NULL) DESC, id ASC").
		First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

// NormalizeEmail is the single spelling used to store and compare addresses.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (r *repository) FindByID(id uint) (*user_module.User, error) {
	return r.findUser("id = ?", id)
}

func (r *repository) findUser(query string, args ...any) (*user_module.User, error) {
	var user user_module.User
	err := r.db.Preload("Role").Where(query, args...).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

func (r *repository) CreateEmailVerification(ev *user_module.EmailVerification) error {
	r.db.Where("user_id = ?", ev.UserID).Delete(&user_module.EmailVerification{})
	return r.db.Create(ev).Error
}

func (r *repository) FindEmailVerification(hashedToken string) (*user_module.EmailVerification, error) {
	var ev user_module.EmailVerification
	err := r.db.Where("token = ?", hashedToken).First(&ev).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidToken
	}
	return &ev, err
}

func (r *repository) DeleteEmailVerification(id uint) error {
	return r.db.Delete(&user_module.EmailVerification{}, id).Error
}

func (r *repository) CreateRefreshToken(rt *user_module.RefreshToken) error {
	return r.db.Create(rt).Error
}

func (r *repository) FindRefreshToken(hashedToken string) (*user_module.RefreshToken, error) {
	var rt user_module.RefreshToken
	err := r.db.Where("hashed_token = ?", hashedToken).First(&rt).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidToken
	}
	return &rt, err
}

func (r *repository) DeleteRefreshToken(hashedToken string) error {
	return r.db.Where("hashed_token = ?", hashedToken).Delete(&user_module.RefreshToken{}).Error
}

func (r *repository) DeleteAllRefreshTokens(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&user_module.RefreshToken{}).Error
}

func (r *repository) MarkFirstLoginIfNeeded(userID uint) (bool, error) {
	now := time.Now()
	result := r.db.Model(&user_module.User{}).
		Where("id = ? AND first_login_at IS NULL", userID).
		Update("first_login_at", &now)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidToken     = errors.New("invalid or expired token")
	ErrEmailTaken       = errors.New("email already in use")
	ErrUsernameTaken    = errors.New("username already in use")
	ErrWrongPassword    = errors.New("incorrect password")
	ErrEmailUnverified  = errors.New("email not verified")
	ErrPasswordMismatch = errors.New("passwords do not match")
)

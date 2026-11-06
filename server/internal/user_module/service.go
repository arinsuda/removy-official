package user_module

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"mime/multipart"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/arinsuda/movie-hub/internal/identity_policy"
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/arinsuda/movie-hub/internal/shared/storage"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	repo                *repository
	minio               *storage.MinIOClient
	statsSvc            StatsProvider
	mailer              Mailer
	emailVerifier       EmailVerificationSender
	passwordResetMailer PasswordResetMailer
	policy              privacy_policy.UserAccessPolicy
}

func NewService(
	db *gorm.DB,
	mc *storage.MinIOClient,
	statsSvc StatsProvider,
	mailer Mailer,
	emailVerifier EmailVerificationSender,
	passwordResetMailer PasswordResetMailer,
	policy privacy_policy.UserAccessPolicy,
) *Service {
	return &Service{
		repo:                newRepository(db),
		minio:               mc,
		statsSvc:            statsSvc,
		mailer:              mailer,
		emailVerifier:       emailVerifier,
		passwordResetMailer: passwordResetMailer,
		policy:              policy,
	}
}

func (s *Service) GetProfile(targetUserID, requesterID uint) (*UserProfileResponse, error) {
	user, reviewCount, followerCount, followingCount, err := s.repo.FindByID(targetUserID)
	if err != nil {
		return nil, err
	}

	level := s.fetchLevel(targetUserID)
	profile := toProfileResponse(user, reviewCount, followerCount, followingCount, level)

	if user.IsPrivate && requesterID != targetUserID {
		canView, err := s.policy.CanViewProfileSection(context.Background(), requesterID, targetUserID, privacy_policy.SectionProfile)
		if err != nil || !canView {
			profile.Bio = nil
			profile.FavoriteGenres = nil
			profile.Gender = ""
			profile.DateOfBirth = nil
		}
	}

	if profile.AvatarURL != nil && strings.HasPrefix(*profile.AvatarURL, "avatars/") {
		if url, err := s.minio.PresignURL(context.Background(), *profile.AvatarURL); err == nil {
			profile.AvatarURL = &url
		}
	}

	return profile, nil
}

func (s *Service) UpdateProfile(
	targetUserID, requesterID uint,
	req UpdateProfileRequest,
	avatarFile multipart.File,
	avatarHeader *multipart.FileHeader,
) (*UserProfileResponse, error) {
	if targetUserID != requesterID {
		return nil, ErrForbidden
	}

	updates := map[string]any{}
	if req.DisplayName != nil {
		updates["display_name"] = req.DisplayName
	}
	if req.Bio != nil {
		updates["bio"] = req.Bio
	}
	if req.DateOfBirth != nil {
		updates["date_of_birth"] = req.DateOfBirth
	}
	if req.Gender != "" {
		if !isValidGender(req.Gender) {
			return nil, errors.New("invalid gender")
		}
		updates["gender"] = req.Gender
	}
	if req.GenderOther != nil {
		updates["gender_other"] = req.GenderOther
	}
	if req.FavoriteGenres != nil {
		updates["favorite_genres"] = req.FavoriteGenres
	}
	if req.IsPrivate != nil {
		updates["is_private"] = req.IsPrivate
	}

	if avatarFile != nil && avatarHeader != nil {
		oldKey := s.currentAvatarKey(targetUserID)
		newKey, err := s.minio.UploadAvatar(context.Background(), targetUserID, avatarFile, avatarHeader)
		if err != nil {
			return nil, err
		}
		updates["avatar_url"] = newKey
		if oldKey != "" {
			_ = s.minio.DeleteObject(context.Background(), oldKey)
		}
	}
	log.Printf("DEBUG updates map: %+v", updates)
	if len(updates) > 0 {
		if err := s.repo.UpdateProfile(targetUserID, updates); err != nil {
			return nil, err
		}
	}

	return s.GetProfile(targetUserID, requesterID)
}

func (s *Service) DeleteUser(targetUserID, requesterID uint, requesterRole string) error {
	if targetUserID != requesterID && requesterRole != "admin" {
		return ErrForbidden
	}
	return s.repo.DeleteUser(targetUserID)
}

func (s *Service) UpdateFavoriteGenres(targetUserID, requesterID uint, genres []int) (*UserProfileResponse, error) {
	if targetUserID != requesterID {
		return nil, ErrForbidden
	}

	var stored *string
	if len(genres) > 0 {
		b, err := json.Marshal(genres)
		if err != nil {
			return nil, err
		}
		str := string(b)
		stored = &str
	}

	if err := s.repo.UpdateProfile(targetUserID, map[string]any{"favorite_genres": stored}); err != nil {
		return nil, err
	}
	return s.GetProfile(targetUserID, requesterID)
}

func (s *Service) RequestEmailChange(targetUserID, requesterID uint, newEmail string) error {
	if targetUserID != requesterID {
		return ErrForbidden
	}

	newEmail = strings.TrimSpace(strings.ToLower(newEmail))
	if !isValidEmail(newEmail) {
		return ErrInvalidEmailFormat
	}

	user, _, _, _, err := s.repo.FindByID(targetUserID)
	if err != nil {
		return err
	}

	if strings.EqualFold(user.Email, newEmail) {
		return ErrEmailSameAsCurrent
	}

	taken, err := s.repo.IsEmailTaken(newEmail, targetUserID)
	if err != nil {
		return err
	}
	if taken {
		return ErrEmailAlreadyInUse
	}

	otp, err := generateOTP()
	if err != nil {
		return err
	}

	otpHash, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	changeReq := &EmailChangeRequest{
		UserID:    targetUserID,
		NewEmail:  newEmail,
		OTPHash:   string(otpHash),
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	if err := s.repo.UpsertEmailChangeRequest(changeReq); err != nil {
		return err
	}

	if err := s.mailer.SendOTP(user.Email, otp); err != nil {
		_ = s.repo.DeleteEmailChangeRequest(targetUserID)
		return errors.New("failed to send OTP email")
	}

	return nil
}

const maxOTPAttempts = 5

func (s *Service) VerifyEmailChange(
	targetUserID, requesterID uint,
	otp string,
) (*UserProfileResponse, error) {
	if targetUserID != requesterID {
		return nil, ErrForbidden
	}

	req, err := s.repo.FindEmailChangeRequest(targetUserID)
	if err != nil {
		return nil, err
	}

	if req.IsExpired() {
		_ = s.repo.DeleteEmailChangeRequest(targetUserID)
		return nil, ErrOTPExpired
	}

	if req.AttemptCount >= maxOTPAttempts {
		_ = s.repo.DeleteEmailChangeRequest(targetUserID)
		return nil, ErrOTPMaxAttempts
	}

	if strings.TrimSpace(req.NewEmail) == "" {
		_ = s.repo.DeleteEmailChangeRequest(targetUserID)
		return nil, ErrPendingEmailMissing
	}

	if err := bcrypt.CompareHashAndPassword([]byte(req.OTPHash), []byte(otp)); err != nil {
		_ = s.repo.IncrementOTPAttempt(req.ID)
		return nil, ErrOTPInvalid
	}

	if err := s.repo.UpdateProfile(targetUserID, map[string]any{
		"email":             req.NewEmail,
		"verified_email_at": nil,
	}); err != nil {
		return nil, err
	}

	if err := s.repo.DeleteEmailChangeRequest(targetUserID); err != nil {
		return nil, err
	}

	if s.emailVerifier != nil {
		defer func() {
			_ = s.emailVerifier.SendVerification(targetUserID, req.NewEmail)
		}()
	}

	return s.GetProfile(targetUserID, requesterID)
}

func (s *Service) ChangePassword(targetUserID, requesterID uint, req ChangePasswordRequest) error {
	if targetUserID != requesterID {
		return ErrForbidden
	}

	if err := validateNewPassword(req.NewPassword, req.ConfirmPassword); err != nil {
		return err
	}

	user, _, _, _, err := s.repo.FindByID(targetUserID)
	if err != nil {
		return err
	}

	if user.Password != nil && *user.Password != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.OldPassword)); err != nil {
			return ErrInvalidCredentials
		}
		if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.NewPassword)); err == nil {
			return errors.New("new password must be different from current password")
		}
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(targetUserID, string(hashed))
}

func (s *Service) ForgotPassword(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	log.Printf("DEBUG ForgotPassword: email=%s", email)

	user, err := s.repo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil
		}
		return err
	}

	rawToken, err := generateSecureToken()
	if err != nil {
		return err
	}

	hashedToken, err := bcrypt.GenerateFromPassword([]byte(rawToken), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	record := &PasswordResetToken{
		UserID:      user.ID,
		HashedToken: string(hashedToken),
		ExpiresAt:   time.Now().Add(1 * time.Hour),
		Email:       &user.Email, // bound: only valid while this is still the account's address
	}
	if err := s.repo.UpsertPasswordResetToken(record); err != nil {
		return err
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s&uid=%d",
		getEnv("FRONTEND_BASE_URL", "http://localhost:3000"),
		rawToken,
		user.ID,
	)

	if err := s.passwordResetMailer.SendResetLink(user.Email, resetURL); err != nil {
		_ = s.repo.DeletePasswordResetToken(user.ID)
		return errors.New("failed to send reset email")
	}

	return nil
}

func (s *Service) ResetPassword(userID uint, rawToken string, req ResetPasswordRequest) error {
	if err := validateNewPassword(req.NewPassword, req.ConfirmPassword); err != nil {
		return err
	}

	record, err := s.repo.FindPasswordResetTokenByUserID(userID)
	if err != nil {
		return err
	}

	if record.IsExpired() {
		_ = s.repo.DeletePasswordResetToken(userID)
		return ErrPasswordResetTokenExpired
	}

	if err := bcrypt.CompareHashAndPassword([]byte(record.HashedToken), []byte(rawToken)); err != nil {
		return ErrPasswordResetTokenInvalid
	}

	// A token minted before the email column existed, or for another address, proves nothing.
	if record.Email == nil {
		return ErrPasswordResetTokenInvalid
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	hashedStr := string(hashed)

	return s.repo.db.Transaction(func(tx *gorm.DB) error {
		// Consume first: concurrent uses race on this row and only one deletes it.
		del := tx.Delete(&PasswordResetToken{}, record.ID)
		if del.Error != nil {
			return del.Error
		}
		if del.RowsAffected != 1 {
			return ErrPasswordResetTokenInvalid
		}
		_, err := identity_policy.ProveEmail(tx, identity_policy.Proof{
			UserID: userID, Email: *record.Email, Kind: identity_policy.KindPasswordReset, NewPasswordHash: &hashedStr,
		})
		switch {
		case errors.Is(err, identity_policy.ErrEmailMismatch):
			return ErrPasswordResetTokenInvalid
		case errors.Is(err, identity_policy.ErrUserNotFound), errors.Is(err, identity_policy.ErrUserInactive):
			return ErrUserNotFound
		}
		return err
	})
}

func validateNewPassword(newPassword, confirmPassword string) error {
	if newPassword == "" {
		return errors.New("new_password is required")
	}
	if len(newPassword) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if newPassword != confirmPassword {
		return errors.New("passwords do not match")
	}
	return nil
}

func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}

func generateOTP() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func isValidEmail(email string) bool {
	const pattern = `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func (s *Service) fetchLevel(userID uint) int {
	if s.statsSvc == nil {
		return 1
	}
	return s.statsSvc.GetLevel(userID)
}

func (s *Service) currentAvatarKey(userID uint) string {
	user, _, _, _, err := s.repo.FindByID(userID)
	if err != nil || user.AvatarURL == nil {
		return ""
	}
	if strings.HasPrefix(*user.AvatarURL, "avatars/") {
		return *user.AvatarURL
	}
	return ""
}

func toProfileResponse(u *User, reviewCount, followerCount, followingCount, level int) *UserProfileResponse {
	return &UserProfileResponse{
		ID:              u.ID,
		Username:        u.Username,
		Email:           u.Email,
		VerifiedEmailAt: u.VerifiedEmailAt,
		DisplayName:     u.DisplayName,
		Bio:             u.Bio,
		AvatarURL:       u.AvatarURL,
		Gender:          u.Gender,
		FavoriteGenres:  u.FavoriteGenres,
		DateOfBirth:     u.DateOfBirth,
		ReviewCount:     reviewCount,
		FollowerCount:   followerCount,
		FollowingCount:  followingCount,
		IsPrivate:       u.IsPrivate,
		Level:           level,
		Role:            string(u.Role.RoleName),
		CreatedAt:       u.CreatedAt,
	}
}

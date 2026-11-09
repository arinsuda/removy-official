package user_module

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
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
	reissuer            SessionReissuer
}

// SetSessionReissuer wires the session issuer used to keep the caller signed in after a password change.
func (s *Service) SetSessionReissuer(r SessionReissuer) { s.reissuer = r }

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

	if requesterID != targetUserID {
		// Email is account-identifying data; only the owner may read it.
		profile.Email = ""
		profile.VerifiedEmailAt = nil
		profile.OnboardedAt = nil
	}

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
	// An empty string clears the column (NULL).
	if req.DisplayName != nil {
		updates["display_name"] = nullIfEmpty(req.DisplayName)
	}
	if req.Bio != nil {
		updates["bio"] = nullIfEmpty(req.Bio)
	}
	if req.ClearDateOfBirth {
		updates["date_of_birth"] = nil
	} else if req.DateOfBirth != nil {
		updates["date_of_birth"] = req.DateOfBirth
	}
	if req.Gender != "" {
		if !isValidGender(req.Gender) {
			return nil, errors.New("invalid gender")
		}
		updates["gender"] = req.Gender
	}
	if req.GenderOther != nil {
		updates["gender_other"] = nullIfEmpty(req.GenderOther)
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
	if len(updates) > 0 {
		if err := s.repo.UpdateProfile(targetUserID, updates); err != nil {
			return nil, err
		}
	}

	return s.GetProfile(targetUserID, requesterID)
}

// DeleteUser deactivates the caller's own account and ends all of its sessions.
// Admins remove other users through the admin module, never through this path.
//
// Confirmation comes first and changes nothing when it fails: an account with a password needs
// that password; a Google-only account (no password) must type its username instead. The check
// runs before the transaction so bcrypt does not hold the admin guard lock; a password change in
// between bumps token_version and the transaction then rejects the stale session.
func (s *Service) DeleteUser(targetUserID, requesterID uint, sessionVersion int, req DeleteAccountRequest) error {
	if targetUserID != requesterID {
		return ErrForbidden
	}
	username, password, err := s.repo.findCredentials(targetUserID)
	if err != nil {
		return err
	}
	if password != nil && *password != "" {
		if req.Password == "" || bcrypt.CompareHashAndPassword([]byte(*password), []byte(req.Password)) != nil {
			return ErrWrongPassword
		}
	} else if !strings.EqualFold(strings.TrimSpace(req.ConfirmUsername), username) {
		return ErrConfirmMismatch
	}
	if err := s.repo.DeleteUser(targetUserID, sessionVersion); err != nil {
		return err
	}
	identity_policy.Outcome{SessionsRevoked: true}.Publish(targetUserID)
	return nil
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

	updates := map[string]any{
		"favorite_genres": stored,
		"onboarded_at":    gorm.Expr("COALESCE(onboarded_at, now())"),
	}
	if err := s.repo.UpdateProfile(targetUserID, updates); err != nil {
		return nil, err
	}
	return s.GetProfile(targetUserID, requesterID)
}

// RequestEmailChange mails one code to the current address and one to the new
// address. Both are needed to finish: the old one shows the session owner still
// controls the account, the new one proves the new address belongs to them.
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

	currentOTP, currentHash, err := newOTPWithHash()
	if err != nil {
		return err
	}
	newOTP, newHash, err := newOTPWithHash()
	if err != nil {
		return err
	}

	changeReq := &EmailChangeRequest{
		UserID:     targetUserID,
		NewEmail:   newEmail,
		OTPHash:    currentHash,
		NewOTPHash: &newHash,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	}

	if err := s.repo.UpsertEmailChangeRequest(changeReq); err != nil {
		return err
	}

	if err := s.mailer.SendOTP(user.Email, currentOTP); err != nil {
		_ = s.repo.DeleteEmailChangeRequest(targetUserID)
		return errors.New("failed to send OTP email")
	}
	if err := s.mailer.SendOTP(newEmail, newOTP); err != nil {
		_ = s.repo.DeleteEmailChangeRequest(targetUserID)
		return errors.New("failed to send OTP email")
	}

	return nil
}

func newOTPWithHash() (otp, hash string, err error) {
	otp, err = generateOTP()
	if err != nil {
		return "", "", err
	}
	h, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	return otp, string(h), nil
}

const maxOTPAttempts = 5

// VerifyEmailChange completes the change in one transaction: both codes must match,
// then the email and its proof are written together, so the account never passes
// through an unverified state and no stale token can outlive the old address.
func (s *Service) VerifyEmailChange(
	targetUserID, requesterID uint,
	sessionVersion int,
	otp, newEmailOTP string,
) (*UserProfileResponse, error) {
	if targetUserID != requesterID {
		return nil, ErrForbidden
	}

	var codeErr error
	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		// A request that passed RequireAuth just before a claim must not write after it.
		if err := identity_policy.EnsureSession(tx, targetUserID, sessionVersion, true); err != nil {
			if errors.Is(err, identity_policy.ErrSessionStale) {
				return ErrSessionStale
			}
			return err
		}

		var req EmailChangeRequest
		if err := tx.Where("user_id = ?", targetUserID).First(&req).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrOTPNotFound
			}
			return err
		}

		if req.IsExpired() {
			codeErr = ErrOTPExpired
			return tx.Where("user_id = ?", targetUserID).Delete(&EmailChangeRequest{}).Error
		}
		if req.AttemptCount >= maxOTPAttempts {
			codeErr = ErrOTPMaxAttempts
			return tx.Where("user_id = ?", targetUserID).Delete(&EmailChangeRequest{}).Error
		}
		if strings.TrimSpace(req.NewEmail) == "" {
			codeErr = ErrPendingEmailMissing
			return tx.Where("user_id = ?", targetUserID).Delete(&EmailChangeRequest{}).Error
		}
		// A request made before the second code existed cannot prove the new address.
		if req.NewOTPHash == nil {
			codeErr = ErrOTPNotFound
			return tx.Where("user_id = ?", targetUserID).Delete(&EmailChangeRequest{}).Error
		}

		if bcrypt.CompareHashAndPassword([]byte(req.OTPHash), []byte(otp)) != nil ||
			bcrypt.CompareHashAndPassword([]byte(*req.NewOTPHash), []byte(newEmailOTP)) != nil {
			codeErr = ErrOTPInvalid // committed below so the attempt counts
			return tx.Model(&EmailChangeRequest{}).Where("id = ?", req.ID).
				UpdateColumn("attempt_count", gorm.Expr("attempt_count + 1")).Error
		}

		var taken int64
		if err := tx.Model(&User{}).
			Where("lower(email) = lower(?) AND id != ? AND deleted_at IS NULL", req.NewEmail, targetUserID).
			Count(&taken).Error; err != nil {
			return err
		}
		if taken > 0 {
			return ErrEmailAlreadyInUse
		}

		// Both codes matched: the new address is proven, write it verified. Never NULL.
		if err := tx.Exec(`UPDATE users SET email = ?, verified_email_at = now(), updated_at = now() WHERE id = ?`,
			req.NewEmail, targetUserID).Error; err != nil {
			if isUniqueViolation(err) {
				return ErrEmailAlreadyInUse
			}
			return err
		}
		// Anything minted for the old address must not outlive it.
		for _, t := range []struct{ table, column string }{
			{"email_change_requests", "user_id"},
			{"link_pending", "local_user_id"},
			{"link_handoff_tokens", "user_id"},
			{"password_reset_tokens", "user_id"},
			{"email_verifications", "user_id"},
		} {
			if err := tx.Exec(`DELETE FROM `+t.table+` WHERE `+t.column+` = ?`, targetUserID).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if codeErr != nil {
		return nil, codeErr
	}

	return s.GetProfile(targetUserID, requesterID)
}

func isUniqueViolation(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "23505")
}

func (s *Service) ChangePassword(targetUserID, requesterID uint, sessionVersion int, req ChangePasswordRequest) error {
	if targetUserID != requesterID {
		return ErrForbidden
	}

	if err := validateNewPassword(req.NewPassword, req.ConfirmPassword); err != nil {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = s.repo.db.Transaction(func(tx *gorm.DB) error {
		// A request that passed RequireAuth just before a claim must not write after it.
		if err := identity_policy.EnsureSession(tx, targetUserID, sessionVersion, true); err != nil {
			if errors.Is(err, identity_policy.ErrSessionStale) {
				return ErrSessionStale
			}
			return err
		}

		var current struct{ Password *string }
		if err := tx.Raw(`SELECT password FROM users WHERE id = ? AND deleted_at IS NULL`, targetUserID).Scan(&current).Error; err != nil {
			return err
		}
		// First passwords are set only through a reset link (proof of the mailbox),
		// never through an authenticated session that may belong to a squatter.
		if current.Password == nil || *current.Password == "" {
			return ErrPasswordNotSet
		}
		if err := bcrypt.CompareHashAndPassword([]byte(*current.Password), []byte(req.OldPassword)); err != nil {
			return ErrInvalidCredentials
		}
		if err := bcrypt.CompareHashAndPassword([]byte(*current.Password), []byte(req.NewPassword)); err == nil {
			return errors.New("new password must be different from current password")
		}

		// A new password ends every session, this one included: the handler reissues it.
		if err := tx.Exec(`UPDATE users SET password = ?, token_version = token_version + 1, updated_at = now() WHERE id = ?`,
			string(hashed), targetUserID).Error; err != nil {
			return err
		}
		return identity_policy.RevokeSessionScope(tx, targetUserID)
	})
	if err != nil {
		return err
	}
	identity_policy.Outcome{SessionsRevoked: true}.Publish(targetUserID)
	return nil
}

func (s *Service) ForgotPassword(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))

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

	var out identity_policy.Outcome
	err = s.repo.db.Transaction(func(tx *gorm.DB) error {
		// Consume first: concurrent uses race on this row and only one deletes it.
		del := tx.Delete(&PasswordResetToken{}, record.ID)
		if del.Error != nil {
			return del.Error
		}
		if del.RowsAffected != 1 {
			return ErrPasswordResetTokenInvalid
		}
		var err error
		out, err = identity_policy.ProveEmail(tx, identity_policy.Proof{
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
	if err != nil {
		return err
	}
	out.Publish(userID)
	return nil
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

func nullIfEmpty(v *string) any {
	if v == nil || *v == "" {
		return nil
	}
	return *v
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
		OnboardedAt:     u.OnboardedAt,
	}
}

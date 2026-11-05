package auth_module

import (
	"bytes"
	"context"
	"crypto/subtle"
	"errors"
	"fmt"

	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/mailer"
	"github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/internal/shared/storage"
	"github.com/arinsuda/movie-hub/internal/user_module"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrGoogleOAuthDisabled             = errors.New("google authentication is disabled")
	ErrGoogleAccountLinkRequired       = errors.New("google_account_link_required")
	ErrGoogleIdentityAlreadyConnected  = errors.New("google_identity_already_connected")
	ErrCannotDisconnectFinalAuthMethod = errors.New("cannot disconnect your final authentication method")
	ErrInvalidOAuthState               = errors.New("invalid or mismatched oauth state")
	ErrUserInactive                    = errors.New("user account is inactive")
)

type Service struct {
	repo           *repository
	jwt            *jwtManager
	mailer         *mailer.Mailer
	minio          *storage.MinIOClient
	cfg            *config.Config
	userSvc        UserPasswordService
	notifSvc       *notification_module.Service
	db             *gorm.DB
	oauthStore     OAuthTransactionStore
	googleProvider GoogleOAuthProvider
	loginFallback  LoginFallback
}

// LoginFallback lets the account-linking module provision a local user whose
// credentials were proven against the remote platform. It returns (nil, nil)
// when the fallback does not apply or the credentials are not valid remotely.
type LoginFallback interface {
	LoginViaRemote(email, password string) (*user_module.User, error)
}

func (s *Service) SetLoginFallback(f LoginFallback) {
	s.loginFallback = f
}

type UserPasswordService interface {
	ForgotPassword(email string) error
	ResetPassword(userID uint, rawToken string, req user_module.ResetPasswordRequest) error
}

func (s *Service) SetUserService(userSvc UserPasswordService) {
	s.userSvc = userSvc
}

func (s *Service) SetGoogleOAuthProvider(provider GoogleOAuthProvider) {
	s.googleProvider = provider
}

func (s *Service) SetOAuthTransactionStore(store OAuthTransactionStore) {
	s.oauthStore = store
}

func NewService(db *gorm.DB, cfg *config.Config, m *mailer.Mailer, mc *storage.MinIOClient, notif *notification_module.Service) *Service {
	return &Service{
		repo:           newRepository(db),
		jwt:            newJWTManager(cfg.JWT),
		mailer:         m,
		minio:          mc,
		cfg:            cfg,
		notifSvc:       notif,
		db:             db,
		oauthStore:     NewMemoryOAuthTransactionStore(),
		googleProvider: NewRealGoogleOAuthProvider(),
	}
}

func (s *Service) Register(req RegisterRequest) (*user_module.User, error) {
	if req.Password != req.ConfirmPassword {
		return nil, ErrPasswordMismatch
	}

	if _, err := s.repo.FindByEmail(req.Email); err == nil {
		return nil, ErrEmailTaken
	}
	if _, err := s.repo.FindByUsername(req.Username); err == nil {
		return nil, ErrUsernameTaken
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("register: hash password: %w", err)
	}

	hashedStr := string(hashed)
	user := &user_module.User{
		Username: req.Username,
		Email:    req.Email,
		Password: &hashedStr,
		RoleID:   2,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("register: create user: %w", err)
	}

	go func() {
		if err := s.sendVerificationEmail(user); err != nil {
			log.Printf("ERROR: send verification email to %s: %v", user.Email, err)
		}
	}()

	created, err := s.repo.FindByID(user.ID)
	if err != nil {
		log.Printf("ERROR register: FindByID(%d) with Preload(Role) failed: %v", user.ID, err)
		return nil, fmt.Errorf("register: reload user: %w", err)
	}

	return created, nil
}

func (s *Service) Login(req LoginRequest, userAgent, ip string) (*TokenPair, *user_module.User, error) {
	user, err := s.repo.FindByEmail(req.Identifier)
	if err != nil {
		user, err = s.repo.FindByUsername(req.Identifier)
		if err != nil {
			// No local user: try the linked remote platform (provisions + links on success).
			// Any remote failure is treated as invalid credentials.
			if s.loginFallback == nil || !strings.Contains(req.Identifier, "@") {
				return nil, nil, ErrUserNotFound
			}
			fbUser, fbErr := s.loginFallback.LoginViaRemote(req.Identifier, req.Password)
			if fbErr != nil || fbUser == nil || !fbUser.IsActive {
				return nil, nil, ErrUserNotFound
			}
			user = fbUser
			return s.finishLogin(user, userAgent, ip)
		}
	}

	if !user.IsActive {
		return nil, nil, ErrUserNotFound
	}

	if user.Password == nil || *user.Password == "" {
		return nil, nil, ErrWrongPassword
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.Password)); err != nil {
		return nil, nil, ErrWrongPassword
	}

	if user.VerifiedEmailAt == nil {
		return nil, nil, ErrEmailUnverified
	}

	return s.finishLogin(user, userAgent, ip)
}

func (s *Service) finishLogin(user *user_module.User, userAgent, ip string) (*TokenPair, *user_module.User, error) {
	if user.AvatarURL != nil && strings.HasPrefix(*user.AvatarURL, "avatars/") {
		if url, err := s.minio.PresignURL(context.Background(), *user.AvatarURL); err == nil {
			user.AvatarURL = &url
		}
	}

	pair, err := s.IssueAndStoreTokens(user, userAgent, ip)
	if err != nil {
		return nil, nil, err
	}

	if isFirst, err := s.repo.MarkFirstLoginIfNeeded(user.ID); err == nil && isFirst && s.notifSvc != nil {
		_ = s.notifSvc.PushWelcome(context.Background(), user.ID, user.Username)
	}

	return pair, user, nil
}

func (s *Service) InitiateGoogleLogin(ctx context.Context, returnURL string) (authURL string, state string, err error) {
	if !s.cfg.Google.Enabled {
		return "", "", ErrGoogleOAuthDisabled
	}

	state, err = GenerateCryptoString(32)
	if err != nil {
		return "", "", err
	}

	nonce, err := GenerateCryptoString(32)
	if err != nil {
		return "", "", err
	}

	verifier, err := GenerateCryptoString(43)
	if err != nil {
		return "", "", err
	}

	challenge := ComputePKCEChallenge(verifier)
	validReturnURL := ValidateRelativeReturnURL(returnURL)

	tx := &OAuthTransaction{
		State:         state,
		StateHash:     HashState(state),
		Nonce:         nonce,
		CodeVerifier:  verifier,
		CodeChallenge: challenge,
		Intent:        IntentLogin,
		ReturnURL:     validReturnURL,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(10 * time.Minute),
	}

	if err := s.oauthStore.Create(ctx, tx); err != nil {
		return "", "", fmt.Errorf("failed to save oauth transaction: %w", err)
	}

	params := GoogleAuthorizationParams{
		ClientID:      s.cfg.Google.ClientID,
		RedirectURI:   s.cfg.Google.RedirectURL,
		State:         state,
		Nonce:         nonce,
		CodeChallenge: challenge,
	}

	authURL = s.googleProvider.AuthorizationURL(params)
	return authURL, state, nil
}

func (s *Service) InitiateGoogleLink(ctx context.Context, userID uint, returnURL string) (authURL string, state string, err error) {
	if !s.cfg.Google.Enabled {
		return "", "", ErrGoogleOAuthDisabled
	}

	var user user_module.User
	if err := s.db.Where("id = ? AND is_active = true", userID).First(&user).Error; err != nil {
		return "", "", ErrUserNotFound
	}

	state, err = GenerateCryptoString(32)
	if err != nil {
		return "", "", err
	}

	nonce, err := GenerateCryptoString(32)
	if err != nil {
		return "", "", err
	}

	verifier, err := GenerateCryptoString(43)
	if err != nil {
		return "", "", err
	}

	challenge := ComputePKCEChallenge(verifier)
	validReturnURL := ValidateRelativeReturnURL(returnURL)

	tx := &OAuthTransaction{
		State:         state,
		StateHash:     HashState(state),
		Nonce:         nonce,
		CodeVerifier:  verifier,
		CodeChallenge: challenge,
		Intent:        IntentLink,
		UserID:        userID,
		ReturnURL:     validReturnURL,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(10 * time.Minute),
	}

	if err := s.oauthStore.Create(ctx, tx); err != nil {
		return "", "", fmt.Errorf("failed to save oauth transaction: %w", err)
	}

	params := GoogleAuthorizationParams{
		ClientID:      s.cfg.Google.ClientID,
		RedirectURI:   s.cfg.Google.RedirectURL,
		State:         state,
		Nonce:         nonce,
		CodeChallenge: challenge,
	}

	authURL = s.googleProvider.AuthorizationURL(params)
	return authURL, state, nil
}

func (s *Service) HandleGoogleCallback(ctx context.Context, code, stateStr, cookieStateStr, userAgent, ip string) (*TokenPair, *user_module.User, string, error) {
	if !s.cfg.Google.Enabled {
		return nil, nil, "", ErrGoogleOAuthDisabled
	}

	if stateStr == "" || cookieStateStr == "" || subtle.ConstantTimeCompare([]byte(stateStr), []byte(cookieStateStr)) != 1 {
		return nil, nil, "", ErrInvalidOAuthState
	}

	tx, err := s.oauthStore.Consume(ctx, stateStr)
	if err != nil {
		return nil, nil, "", ErrOAuthTransactionNotFound
	}

	tokens, err := s.googleProvider.ExchangeCode(ctx, code, tx.CodeVerifier, s.cfg.Google.ClientID, s.cfg.Google.ClientSecret, s.cfg.Google.RedirectURL)
	if err != nil {
		return nil, nil, tx.ReturnURL, fmt.Errorf("token exchange failed: %w", err)
	}

	claims, err := s.googleProvider.VerifyIDToken(ctx, tokens.IDToken, s.cfg.Google.ClientID, tx.Nonce)
	if err != nil {
		return nil, nil, tx.ReturnURL, fmt.Errorf("id_token verification failed: %w", err)
	}

	var sessionUser *user_module.User
	var tokenPair *TokenPair

	err = s.db.Transaction(func(txDB *gorm.DB) error {
		if tx.Intent == IntentLogin {
			var identity user_module.UserIdentity
			err := txDB.Where("provider = ? AND provider_subject = ?", user_module.ProviderGoogle, claims.Subject).First(&identity).Error

			if err == nil {
				// 1. Existing Google Identity
				var u user_module.User
				if err := txDB.Where("id = ? AND is_active = true", identity.UserID).First(&u).Error; err != nil {
					return ErrUserInactive
				}
				sessionUser = &u
			} else {
				// 2. Check if REMOVY user with matching email exists
				normalizedEmail := strings.ToLower(strings.TrimSpace(claims.Email))
				var existingUser user_module.User
				err := txDB.Where("email = ? AND is_active = true", normalizedEmail).First(&existingUser).Error

				if err == nil {
					// User exists with matching email
					if IsAuthoritativeGoogleEmail(claims) {
						// Auto-link Google identity to authoritative email
						newIdentity := &user_module.UserIdentity{
							UserID:          existingUser.ID,
							Provider:        user_module.ProviderGoogle,
							ProviderSubject: claims.Subject,
							ProviderEmail:   claims.Email,
						}
						if err := txDB.Create(newIdentity).Error; err != nil {
							return fmt.Errorf("failed to link Google identity: %w", err)
						}

						if existingUser.VerifiedEmailAt == nil {
							now := time.Now().UTC()
							existingUser.VerifiedEmailAt = &now
							txDB.Model(&existingUser).Update("verified_email_at", now)
						}
						sessionUser = &existingUser
					} else {
						// Non-authoritative email: reject auto-link!
						return ErrGoogleAccountLinkRequired
					}
				} else {
					// 3. New Account: Create User & UserIdentity atomically
					now := time.Now().UTC()
					var verifiedAt *time.Time
					if claims.EmailVerified && IsAuthoritativeGoogleEmail(claims) {
						verifiedAt = &now
					}

					baseUsername := SanitizeUsernameBase(claims.Email)
					if claims.Name != "" {
						baseUsername = SanitizeUsernameBase(claims.Name)
					}

					finalUsername := baseUsername
					for i := 0; i < 5; i++ {
						var count int64
						txDB.Model(&user_module.User{}).Where("username = ?", finalUsername).Count(&count)
						if count == 0 {
							break
						}
						finalUsername = fmt.Sprintf("%s_%s", baseUsername[:min(len(baseUsername), 20)], GenerateRandomSuffix(4))
					}

					var dispName *string
					if claims.Name != "" {
						dispName = &claims.Name
					}

					newUser := &user_module.User{
						Username:        finalUsername,
						Email:           normalizedEmail,
						Password:        nil, // Passwordless account
						DisplayName:     dispName,
						VerifiedEmailAt: verifiedAt,
						RoleID:          2,
						IsActive:        true,
					}

					if err := txDB.Create(newUser).Error; err != nil {
						return fmt.Errorf("failed to create new user: %w", err)
					}

					newIdentity := &user_module.UserIdentity{
						UserID:          newUser.ID,
						Provider:        user_module.ProviderGoogle,
						ProviderSubject: claims.Subject,
						ProviderEmail:   claims.Email,
					}
					if err := txDB.Create(newIdentity).Error; err != nil {
						return fmt.Errorf("failed to create identity: %w", err)
					}

					sessionUser = newUser
				}
			}

			// Issue REMOVY token pair & store refresh token session inside transaction scope
			pair, err := s.issueAndStoreTokensTx(txDB, sessionUser, userAgent, ip)
			if err != nil {
				return fmt.Errorf("failed to issue session tokens: %w", err)
			}
			tokenPair = pair

		} else if tx.Intent == IntentLink {
			if tx.UserID == 0 {
				return errors.New("link transaction missing user id")
			}

			var identity user_module.UserIdentity
			err := txDB.Where("provider = ? AND provider_subject = ?", user_module.ProviderGoogle, claims.Subject).First(&identity).Error
			if err == nil {
				if identity.UserID != tx.UserID {
					return ErrGoogleIdentityAlreadyConnected
				}
				// Already connected to current user
				var u user_module.User
				txDB.First(&u, tx.UserID)
				sessionUser = &u
			} else {
				var count int64
				txDB.Model(&user_module.UserIdentity{}).Where("user_id = ? AND provider = ?", tx.UserID, user_module.ProviderGoogle).Count(&count)
				if count > 0 {
					return ErrGoogleIdentityAlreadyConnected
				}

				newIdentity := &user_module.UserIdentity{
					UserID:          tx.UserID,
					Provider:        user_module.ProviderGoogle,
					ProviderSubject: claims.Subject,
					ProviderEmail:   claims.Email,
				}
				if err := txDB.Create(newIdentity).Error; err != nil {
					return fmt.Errorf("failed to link google identity: %w", err)
				}
				var u user_module.User
				txDB.First(&u, tx.UserID)
				sessionUser = &u
			}
		}
		return nil
	})

	if err != nil {
		return nil, nil, tx.ReturnURL, err
	}

	// Attempt avatar import asynchronously (non-fatal)
	if sessionUser != nil && sessionUser.AvatarURL == nil && claims.Picture != "" {
		go s.importGoogleAvatar(context.Background(), claims.Picture, sessionUser.ID)
	}

	return tokenPair, sessionUser, tx.ReturnURL, nil
}

func (s *Service) DisconnectGoogleIdentity(ctx context.Context, userID uint) error {
	if !s.cfg.Google.Enabled {
		return ErrGoogleOAuthDisabled
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var u user_module.User
		if err := tx.Where("id = ? AND is_active = true", userID).First(&u).Error; err != nil {
			return ErrUserNotFound
		}

		var identityCount int64
		if err := tx.Model(&user_module.UserIdentity{}).Where("user_id = ?", userID).Count(&identityCount).Error; err != nil {
			return err
		}

		hasPassword := u.Password != nil && *u.Password != ""
		if !hasPassword && identityCount <= 1 {
			return ErrCannotDisconnectFinalAuthMethod
		}

		res := tx.Where("user_id = ? AND provider = ?", userID, user_module.ProviderGoogle).Delete(&user_module.UserIdentity{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("google identity not connected")
		}

		return nil
	})
}

func (s *Service) GetGoogleOAuthStatus(userID uint) (map[string]interface{}, error) {
	enabled := s.cfg.Google.Enabled
	if !enabled {
		return map[string]interface{}{
			"enabled":          false,
			"google_connected": false,
			"can_disconnect":   false,
		}, nil
	}

	var identity user_module.UserIdentity
	err := s.db.Where("user_id = ? AND provider = ?", userID, user_module.ProviderGoogle).First(&identity).Error
	connected := (err == nil)

	var u user_module.User
	_ = s.db.First(&u, userID)

	var identityCount int64
	s.db.Model(&user_module.UserIdentity{}).Where("user_id = ?", userID).Count(&identityCount)

	hasPassword := u.Password != nil && *u.Password != ""
	canDisconnect := connected && (hasPassword || identityCount > 1)

	return map[string]interface{}{
		"enabled":          true,
		"google_connected": connected,
		"can_disconnect":   canDisconnect,
		"google_email":     identity.ProviderEmail,
	}, nil
}

func (s *Service) importGoogleAvatar(ctx context.Context, pictureURL string, userID uint) {
	if s.minio == nil || pictureURL == "" {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pictureURL, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return
	}

	lr := io.LimitReader(resp.Body, 2*1024*1024)
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, lr); err != nil {
		return
	}

	ext := ".jpg"
	if strings.Contains(contentType, "png") {
		ext = ".png"
	} else if strings.Contains(contentType, "webp") {
		ext = ".webp"
	}

	objectKey := fmt.Sprintf("avatars/google_%d_%d%s", userID, time.Now().Unix(), ext)
	if _, err := s.minio.UploadBuffer(ctx, objectKey, bytes.NewReader(buf.Bytes()), int64(buf.Len()), contentType); err == nil {
		s.db.Model(&user_module.User{}).Where("id = ?", userID).Update("avatar_url", objectKey)
	}
}

func (s *Service) IssueAndStoreTokens(user *user_module.User, userAgent, ip string) (*TokenPair, error) {
	return s.issueAndStoreTokensTx(s.db, user, userAgent, ip)
}

func (s *Service) issueAndStoreTokensTx(tx *gorm.DB, user *user_module.User, userAgent, ip string) (*TokenPair, error) {
	role := string(user.Role.RoleName)
	if role == "" {
		var r user_module.Role
		if err := tx.First(&r, user.RoleID).Error; err == nil {
			role = string(r.RoleName)
		}
	}

	pair, err := s.jwt.Issue(user.ID, role)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	hashedRefresh := HashToken(pair.RefreshToken)
	rt := &user_module.RefreshToken{
		UserID:      user.ID,
		HashedToken: hashedRefresh,
		ExpiresAt:   time.Now().Add(s.cfg.JWT.RefreshTTL),
		UserAgent:   userAgent,
		IPAddress:   ip,
	}

	if err := tx.Create(rt).Error; err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	return pair, nil
}

func (s *Service) Refresh(rawRefresh, userAgent, ip string) (*TokenPair, *user_module.User, error) {
	claims, err := s.jwt.ParseRefresh(rawRefresh)
	if err != nil {
		return nil, nil, ErrInvalidToken
	}

	hashed := HashToken(rawRefresh)
	stored, err := s.repo.FindRefreshToken(hashed)
	if err != nil {
		return nil, nil, ErrInvalidToken
	}

	if stored.IsExpired() {
		_ = s.repo.DeleteRefreshToken(hashed)
		return nil, nil, ErrInvalidToken
	}

	user, err := s.repo.FindByID(claims.UserID)
	if err != nil {
		return nil, nil, ErrUserNotFound
	}

	if !user.IsActive {
		return nil, nil, ErrUserNotFound
	}

	_ = s.repo.DeleteRefreshToken(hashed)

	newPair, err := s.IssueAndStoreTokens(user, userAgent, ip)
	if err != nil {
		return nil, nil, err
	}

	return newPair, user, nil
}

func (s *Service) Logout(rawRefresh string) error {
	hashed := HashToken(rawRefresh)
	return s.repo.DeleteRefreshToken(hashed)
}

func (s *Service) LogoutAll(userID uint) error {
	return s.repo.DeleteAllRefreshTokens(userID)
}

func (s *Service) SendVerification(userID uint, email string) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		user = &user_module.User{
			Username: email,
			Email:    email,
		}
		user.ID = userID
	}
	return s.sendVerificationEmail(user)
}

func (s *Service) VerifyEmail(token string) error {
	record, err := s.repo.FindEmailVerification(token)
	if err != nil {
		return ErrInvalidToken
	}

	if record.IsExpired() {
		_ = s.repo.DeleteEmailVerification(record.ID)
		return ErrInvalidToken
	}

	if err := s.repo.MarkEmailVerified(record.UserID); err != nil {
		return err
	}

	_ = s.repo.DeleteEmailVerification(record.ID)
	return nil
}

func (s *Service) ResendVerification(email string) error {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil
	}

	if user.VerifiedEmailAt != nil {
		return ErrAlreadyVerified
	}

	return s.sendVerificationEmail(user)
}

func (s *Service) ForgotPassword(email string) error {
	if s.userSvc != nil {
		return s.userSvc.ForgotPassword(email)
	}
	return nil
}

func (s *Service) ResetPassword(userID uint, token string, req user_module.ResetPasswordRequest) error {
	if s.userSvc != nil {
		return s.userSvc.ResetPassword(userID, token, req)
	}
	return nil
}

func (s *Service) sendVerificationEmail(u *user_module.User) error {
	if !s.mailer.IsEnabled() {
		log.Printf("WARN mailer disabled, skipping verification email for user %d (%s)", u.ID, u.Email)
		return nil
	}

	rawToken, err := GenerateSecureToken()
	if err != nil {
		return err
	}

	ev := &user_module.EmailVerification{
		UserID:    u.ID,
		Token:     rawToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.repo.CreateEmailVerification(ev); err != nil {
		return err
	}

	verifyURL := fmt.Sprintf("%s/auth/verify-email?token=%s", s.cfg.AppBaseURL, rawToken)
	return s.mailer.SendVerificationEmail(u.Email, u.Username, verifyURL)
}

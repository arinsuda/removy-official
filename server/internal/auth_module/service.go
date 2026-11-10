package auth_module

import (
	"bytes"
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"github.com/arinsuda/movie-hub/internal/shared"

	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/identity_policy"
	"github.com/arinsuda/movie-hub/internal/mailer"
	"github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/internal/shared/fragment"
	"github.com/arinsuda/movie-hub/internal/shared/storage"
	"github.com/arinsuda/movie-hub/internal/user_module"
	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// VerifyEmailPath is the SPA route of the mailed verification link (secret goes in its fragment).
const VerifyEmailPath = "/auth/verify-email"

var (
	ErrGoogleOAuthDisabled             = errors.New("google authentication is disabled")
	ErrGoogleAccountLinkRequired       = errors.New("google_account_link_required")
	ErrGoogleIdentityAlreadyConnected  = errors.New("google_identity_already_connected")
	ErrCannotDisconnectFinalAuthMethod = errors.New("cannot disconnect your final authentication method")
	ErrInvalidOAuthState               = errors.New("invalid or mismatched oauth state")
	ErrUserInactive                    = errors.New("user account is inactive")
	// ErrGoogleEmailVerificationRequired: Google did not vouch for the address, so the account
	// exists but holds no session until its mailbox owner follows the verification link.
	ErrGoogleEmailVerificationRequired = errors.New("google_email_verification_required")
	// ErrVerifyWrongPassword: the typed password did not match; the verification token is still valid.
	ErrVerifyWrongPassword = errors.New("wrong_password")
	// ErrPasswordChoiceRequired: neither a password nor an explicit skip was sent.
	ErrPasswordChoiceRequired = errors.New("password_or_skip_required")
	// ErrVerifyAttemptsExceeded: too many wrong passwords on one token; the owner must resend the link.
	ErrVerifyAttemptsExceeded = errors.New("verify_attempts_exceeded")
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

	req.Email = NormalizeEmail(req.Email)
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
	req.Identifier = strings.TrimSpace(req.Identifier)
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
		_ = s.notifSvc.PushWelcome(context.Background(), user.ID)
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

// claimedVersion is the token_version of the caller's access token, not the current one:
// a reset between the token being issued and this call must not be inherited.
func (s *Service) InitiateGoogleLink(ctx context.Context, userID uint, claimedVersion int, returnURL string) (authURL string, state string, err error) {
	if !s.cfg.Google.Enabled {
		return "", "", ErrGoogleOAuthDisabled
	}

	st, found, err := identity_policy.ReadSession(s.db, userID)
	if err != nil {
		return "", "", err
	}
	if !found || !st.Active || !st.Verified {
		return "", "", ErrUserNotFound
	}
	if st.Version != claimedVersion {
		return "", "", identity_policy.ErrSessionStale
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
		UserVersion:   claimedVersion,
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
	var pendingVerification *user_module.User // account whose email nobody proved: verification mail, no session
	var proofs []identity_policy.Outcome
	var proofUsers []uint

	err = s.db.Transaction(func(txDB *gorm.DB) error {
		if tx.Intent == IntentLogin {
			var identity user_module.UserIdentity
			err := txDB.Where("provider = ? AND provider_subject = ?", user_module.ProviderGoogle, claims.Subject).First(&identity).Error

			err = s.findLiveGoogleIdentity(txDB, claims.Subject, &identity, err)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			if err == nil {
				// 1. Existing Google Identity
				var u user_module.User
				if err := txDB.Where("id = ? AND is_active = true", identity.UserID).First(&u).Error; err != nil {
					return ErrUserInactive
				}
				if u.VerifiedEmailAt == nil && isProvenByGoogle(claims, u.Email) {
					// Google vouches for the account's address: prove it through the policy
					// (wipes anything planted while it was unverified) and keep this identity.
					out, err := s.proveWithGoogle(txDB, &u, claims)
					if err != nil {
						return err
					}
					proofs, proofUsers = append(proofs, out), append(proofUsers, u.ID)
				}
				sessionUser = &u
			} else {
				// 2. Check if POPCORN user with matching email exists
				normalizedEmail := NormalizeEmail(claims.Email)
				var existingUser user_module.User
				err := txDB.Where("lower(email) = ? AND is_active = true", normalizedEmail).
					Order("(verified_email_at IS NOT NULL) DESC, id ASC").First(&existingUser).Error

				if err == nil {
					// User exists with matching email
					if isProvenByGoogle(claims, existingUser.Email) {
						// Authoritative Google email: proves the address (wiping anything planted
						// on an unverified account), then links this Google identity.
						out, err := s.proveWithGoogle(txDB, &existingUser, claims)
						if err != nil {
							return err
						}
						proofs, proofUsers = append(proofs, out), append(proofUsers, existingUser.ID)
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
						if shared.IsPgUniqueViolation(err) {
							return ErrGoogleIdentityAlreadyConnected // concurrent signup won the race
						}
						return fmt.Errorf("failed to create new user: %w", err)
					}

					newIdentity := &user_module.UserIdentity{
						UserID:          newUser.ID,
						Provider:        user_module.ProviderGoogle,
						ProviderSubject: claims.Subject,
						ProviderEmail:   claims.Email,
					}
					if err := txDB.Create(newIdentity).Error; err != nil {
						if shared.IsPgUniqueViolation(err) {
							return ErrGoogleIdentityAlreadyConnected
						}
						return fmt.Errorf("failed to create identity: %w", err)
					}

					sessionUser = newUser
				}
			}

			// No proof of the address, no session: the mailbox owner has to verify first.
			if sessionUser.VerifiedEmailAt == nil {
				pendingVerification = sessionUser
				return nil
			}

			// Issue POPCORN token pair & store refresh token session inside transaction scope
			pair, err := s.issueAndStoreTokensTx(txDB, sessionUser, userAgent, ip)
			if err != nil {
				return fmt.Errorf("failed to issue session tokens: %w", err)
			}
			tokenPair = pair

		} else if tx.Intent == IntentLink {
			if tx.UserID == 0 {
				return errors.New("link transaction missing user id")
			}
			// The callback is public and the transaction may be minutes old: only
			// proceed while the initiating session is still the account's session.
			if err := identity_policy.EnsureSession(txDB, tx.UserID, tx.UserVersion, true); err != nil {
				return err
			}

			var identity user_module.UserIdentity
			err := txDB.Where("provider = ? AND provider_subject = ?", user_module.ProviderGoogle, claims.Subject).First(&identity).Error
			err = s.findLiveGoogleIdentity(txDB, claims.Subject, &identity, err)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
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
					if shared.IsPgUniqueViolation(err) {
						return ErrGoogleIdentityAlreadyConnected
					}
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
	for i, out := range proofs {
		out.Publish(proofUsers[i])
	}

	if pendingVerification != nil {
		u := *pendingVerification
		go func() {
			if err := s.sendVerificationEmail(&u); err != nil {
				log.Printf("ERROR: send verification email to %s: %v", u.Email, err)
			}
		}()
		return nil, nil, tx.ReturnURL, ErrGoogleEmailVerificationRequired
	}

	// Attempt avatar import asynchronously (non-fatal)
	if sessionUser != nil && sessionUser.AvatarURL == nil && claims.Picture != "" {
		go s.importGoogleAvatar(context.Background(), claims.Picture, sessionUser.ID)
	}

	return tokenPair, sessionUser, tx.ReturnURL, nil
}

// maxStaleIdentityReads bounds the re-read loop when concurrent requests release the same stale row.
const maxStaleIdentityReads = 3

// findLiveGoogleIdentity continues a lookup of a Google identity (firstErr is its
// First() result). A row owned by a soft-deleted user is stale: admin delete normally
// releases it, so drop it and report gorm.ErrRecordNotFound (a first-time Google user).
// A merely deactivated (self-deleted) owner keeps the row. If another request already
// removed the stale row (0 rows affected), re-read instead of assuming.
func (s *Service) findLiveGoogleIdentity(txDB *gorm.DB, subject string, identity *user_module.UserIdentity, firstErr error) error {
	err := firstErr
	for i := 0; i < maxStaleIdentityReads; i++ {
		if err != nil {
			return err
		}
		var removed int64
		if err := txDB.Unscoped().Model(&user_module.User{}).Where("id = ? AND deleted_at IS NOT NULL", identity.UserID).Count(&removed).Error; err != nil {
			return err
		}
		if removed == 0 {
			return nil
		}
		res := txDB.Exec("DELETE FROM user_identities WHERE id = ?", identity.ID)
		if res.Error != nil {
			return fmt.Errorf("release stale google identity: %w", res.Error)
		}
		if res.RowsAffected > 0 {
			return gorm.ErrRecordNotFound
		}
		*identity = user_module.UserIdentity{}
		err = txDB.Where("provider = ? AND provider_subject = ?", user_module.ProviderGoogle, subject).First(identity).Error
	}
	return ErrGoogleIdentityAlreadyConnected
}

// isProvenByGoogle reports whether Google asserts ownership of email.
func isProvenByGoogle(claims *GoogleIdentityClaims, email string) bool {
	return claims.EmailVerified && IsAuthoritativeGoogleEmail(claims) && identity_policy.SameEmail(claims.Email, email)
}

// proveWithGoogle records Google's proof of u.Email and makes sure u has exactly
// this Google identity. It reloads u so the caller sees the post-proof state.
func (s *Service) proveWithGoogle(txDB *gorm.DB, u *user_module.User, claims *GoogleIdentityClaims) (identity_policy.Outcome, error) {
	out, err := identity_policy.ProveEmail(txDB, identity_policy.Proof{
		UserID: u.ID, Email: u.Email, Kind: identity_policy.KindGoogle,
	})
	if err != nil {
		return out, fmt.Errorf("prove email: %w", err)
	}

	// The user row is locked until commit, so this check cannot race another proof.
	var identity user_module.UserIdentity
	err = txDB.Where("user_id = ? AND provider = ?", u.ID, user_module.ProviderGoogle).First(&identity).Error
	switch {
	case err == nil:
		if identity.ProviderSubject != claims.Subject {
			return out, ErrGoogleIdentityAlreadyConnected
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		newIdentity := &user_module.UserIdentity{
			UserID:          u.ID,
			Provider:        user_module.ProviderGoogle,
			ProviderSubject: claims.Subject,
			ProviderEmail:   claims.Email,
		}
		if err := txDB.Create(newIdentity).Error; err != nil {
			if isUniqueViolation(err) {
				return out, ErrGoogleIdentityAlreadyConnected
			}
			return out, fmt.Errorf("failed to link Google identity: %w", err)
		}
	default:
		return out, err
	}
	return out, txDB.Preload("Role").First(u, u.ID).Error
}

func isUniqueViolation(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "23505")
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

// IssueAndStoreTokens starts a session for user in its own transaction.
func (s *Service) IssueAndStoreTokens(user *user_module.User, userAgent, ip string) (*TokenPair, error) {
	var pair *TokenPair
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var err error
		pair, err = s.issueAndStoreTokensTx(tx, user, userAgent, ip)
		return err
	})
	return pair, err
}

// issueAndStoreTokensTx is the only place a session is created. It share-locks
// the user row, so a concurrent ProveEmail (which bumps the epoch) finishes
// first, and refuses unless the account is active and its email is proven.
// The epoch in the tokens always comes from the locked row, never from the
// caller's possibly stale user struct. tx must be a transaction.
func (s *Service) issueAndStoreTokensTx(tx *gorm.DB, user *user_module.User, userAgent, ip string) (*TokenPair, error) {
	st, found, err := identity_policy.LockSession(tx, user.ID, false)
	if err != nil {
		return nil, fmt.Errorf("lock session: %w", err)
	}
	if !found || !st.Active {
		return nil, ErrUserNotFound
	}
	if !st.Verified {
		return nil, ErrEmailUnverified
	}

	role := string(user.Role.RoleName)
	if role == "" {
		var r user_module.Role
		if err := tx.First(&r, user.RoleID).Error; err == nil {
			role = string(r.RoleName)
		}
	}

	pair, err := s.jwt.Issue(user.ID, role, st.Version)
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
		if errors.Is(err, ErrInvalidToken) {
			return nil, nil, ErrInvalidToken
		}
		return nil, nil, fmt.Errorf("find refresh token: %w", err) // infrastructure failure, not a verdict on the token
	}

	if stored.IsExpired() {
		_ = s.repo.DeleteRefreshToken(hashed)
		return nil, nil, ErrInvalidToken
	}

	var newPair *TokenPair
	var user *user_module.User
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Lock the user first: a racing claim (ProveEmail) either finishes before
		// this point, and then the token row below is gone, or waits for us.
		if err := identity_policy.EnsureSession(tx, claims.UserID, claims.TokenVersion, true); err != nil {
			if errors.Is(err, identity_policy.ErrSessionStale) {
				return ErrInvalidToken
			}
			return err
		}

		del := tx.Where("hashed_token = ?", hashed).Delete(&user_module.RefreshToken{})
		if del.Error != nil {
			return del.Error
		}
		if del.RowsAffected != 1 { // already rotated or wiped: a refresh token works once
			return ErrInvalidToken
		}

		var u user_module.User
		if err := tx.Preload("Role").First(&u, claims.UserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrUserNotFound
			}
			return err
		}
		user = &u

		newPair, err = s.issueAndStoreTokensTx(tx, user, userAgent, ip)
		return err
	})
	if err != nil {
		return nil, nil, err
	}

	return newPair, user, nil
}

func (s *Service) Logout(rawRefresh string) error {
	hashed := HashToken(rawRefresh)
	return s.repo.DeleteRefreshToken(hashed)
}

// LogoutAll ends every session of the user: refresh tokens go and token_version moves on,
// so access tokens already issued stop working at once and live sockets are dropped.
func (s *Service) LogoutAll(userID uint) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`UPDATE users SET token_version = token_version + 1, updated_at = now() WHERE id = ? AND deleted_at IS NULL`, userID).Error; err != nil {
			return err
		}
		// Refresh tokens plus handoff, pending-link and mail tokens that could revive the session.
		return identity_policy.RevokeSessionScope(tx, userID)
	})
	if err != nil {
		return err
	}
	identity_policy.Outcome{SessionsRevoked: true}.Publish(userID)
	return nil
}

// ReissueSession starts a fresh session for userID and writes its cookies on c. It backs
// the password change, which revokes every session including the caller's own.
// expectedVersion is the epoch that change produced; it is checked under the row lock in
// the same transaction that issues the tokens, so a reset or claim that committed in
// between yields ErrSessionStale and no tokens.
func (s *Service) ReissueSession(c fiber.Ctx, userID uint, expectedVersion int) (string, string, error) {
	var u user_module.User
	if err := s.db.Preload("Role").First(&u, userID).Error; err != nil {
		return "", "", err
	}
	var pair *TokenPair
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := identity_policy.EnsureSession(tx, userID, expectedVersion, false); err != nil {
			return err
		}
		var err error
		pair, err = s.issueAndStoreTokensTx(tx, &u, c.Get("User-Agent"), c.IP())
		return err
	})
	if err != nil {
		return "", "", err
	}
	SetTokenCookies(c, s.cfg, pair)
	return pair.AccessToken, pair.RefreshToken, nil
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

// VerifyResult tells the client what the proof left behind.
type VerifyResult struct {
	// PasswordResetRequired: the address is proven but the current password was not,
	// so the account has no password until the owner sets one through a reset link.
	PasswordResetRequired bool
}

// MaxVerifyAttempts caps wrong-password tries per verification token.
const MaxVerifyAttempts = 5

// VerifyEmail consumes a verification token and proves the address it was
// mailed to. A matching password keeps the account intact. A non-empty wrong
// password changes nothing, keeps the token and counts one attempt (ErrVerifyWrongPassword).
// Only skipPassword ("I don't know the password") wipes planted credentials.
func (s *Service) VerifyEmail(token string, password *string, skipPassword bool) (VerifyResult, error) {
	if token == "" {
		return VerifyResult{}, ErrInvalidToken
	}
	if (password == nil || *password == "") && !skipPassword {
		return VerifyResult{}, ErrPasswordChoiceRequired
	}
	h := HashToken(token) // tokens are stored hashed; the raw value only lives in the mailed link
	// Look up by hash only: accepting the submitted value as-is would let anyone who can read
	// stored hashes verify an account without the mailed link.
	candidates := []string{h}
	// Reserve an attempt in its own committed statement: the proof below rolls back on a
	// wrong password, and the count must survive that. Parallel guesses cannot share a slot.
	reserve := s.db.Exec(`UPDATE email_verifications SET attempts = attempts + 1 WHERE token IN ? AND attempts < ?`, candidates, MaxVerifyAttempts)
	if reserve.Error != nil {
		return VerifyResult{}, reserve.Error
	}
	if reserve.RowsAffected == 0 {
		// Either no such token, or it is spent: reject it for good so the owner must resend.
		del := s.db.Exec(`DELETE FROM email_verifications WHERE token IN ? AND attempts >= ?`, candidates, MaxVerifyAttempts)
		if del.Error == nil && del.RowsAffected > 0 {
			return VerifyResult{}, ErrVerifyAttemptsExceeded
		}
		return VerifyResult{}, ErrInvalidToken
	}
	var out identity_policy.Outcome
	var userID uint
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var record user_module.EmailVerification
		if err := tx.Where("token IN ?", candidates).First(&record).Error; err != nil {
			return ErrInvalidToken
		}
		// Consume first: concurrent clicks race on this row, only one deletes it.
		del := tx.Delete(&user_module.EmailVerification{}, record.ID)
		if del.Error != nil {
			return del.Error
		}
		if del.RowsAffected != 1 || record.IsExpired() || record.Email == nil {
			return ErrInvalidToken
		}
		userID = record.UserID
		var err error
		out, err = identity_policy.ProveEmail(tx, identity_policy.Proof{
			UserID: record.UserID, Email: *record.Email, Kind: identity_policy.KindVerifyLink, PresentedPassword: password,
		})
		if err != nil {
			if errors.Is(err, identity_policy.ErrWrongPassword) {
				return ErrVerifyWrongPassword
			}
			if errors.Is(err, identity_policy.ErrEmailMismatch) || errors.Is(err, identity_policy.ErrUserNotFound) || errors.Is(err, identity_policy.ErrUserInactive) {
				return ErrInvalidToken
			}
			return err
		}
		return nil
	})
	if err != nil {
		return VerifyResult{}, err
	}
	out.Publish(userID)
	return VerifyResult{PasswordResetRequired: out.CredentialsWiped}, nil
}

func (s *Service) ResendVerification(email string) error {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil
	}

	// A verified account gets no mail and no distinct answer, so the endpoint
	// cannot be used to probe which addresses are registered or verified.
	if user.VerifiedEmailAt != nil {
		return nil
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
		Token:     HashToken(rawToken), // only the hash is stored; the mailed link carries the raw token
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Email:     &u.Email, // bound: the token only proves this address, not whatever the email becomes
	}

	if err := s.repo.CreateEmailVerification(ev); err != nil {
		return err
	}

	verifyURL := fragment.URL(s.cfg.AppBaseURL, VerifyEmailPath, url.Values{"token": {rawToken}})
	return s.mailer.SendVerificationEmail(u.Email, u.Username, verifyURL)
}

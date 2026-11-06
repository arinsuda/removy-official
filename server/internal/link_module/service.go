package link_module

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/auth_module"
	"github.com/arinsuda/movie-hub/internal/identity_policy"
	"github.com/arinsuda/movie-hub/internal/mailer"
	"github.com/arinsuda/movie-hub/internal/user_module"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrDisabled        = errors.New("account linking is disabled")
	ErrInvalidToken    = errors.New("invalid or expired token")
	ErrInvalidTarget   = errors.New("invalid target platform")
	ErrRemoteFailed    = errors.New("remote platform unavailable")
	ErrUserInactive    = errors.New("user account is inactive")
	ErrLinkConflict    = errors.New("account already linked to a different account")
	ErrPendingInvalid  = errors.New("link request invalid, expired or too many attempts")
	ErrBadCredentials  = errors.New("invalid credentials")
	ErrInvalidMethod   = errors.New("invalid confirmation method")
	ErrCodeNotAllowed  = errors.New("could not send code")
	ErrNotLinked       = errors.New("not linked")
	ErrEmailNotAllowed = errors.New("email taken")
	ErrSessionStale    = errors.New("session is no longer valid")
	// ErrRemoteUnverified: the other platform has not verified this email, so no local account is created for it.
	ErrRemoteUnverified = errors.New("remote email is not verified")
)

// ConfirmRequiredError is returned by Redeem when the user must prove ownership.
type ConfirmRequiredError struct {
	PendingID string
	Email     string
	Methods   []string
}

func (e *ConfirmRequiredError) Error() string { return CodeConfirmRC }

type Result struct {
	User        *user_module.User
	Provisioned bool
}

type Service struct {
	db     *gorm.DB
	cfg    config.LinkConfig
	auth   *auth_module.Service
	mailer *mailer.Mailer
	remote *remoteClient
}

func NewService(db *gorm.DB, cfg config.LinkConfig, auth *auth_module.Service, m *mailer.Mailer) *Service {
	if cfg.SelfPlatform != PlatformPlanzi {
		cfg.SelfPlatform = PlatformPopcorn
	}
	return &Service{db: db, cfg: cfg, auth: auth, mailer: m, remote: newRemoteClient(cfg)}
}

func (s *Service) Enabled() bool { return s.cfg.Enabled() }

func (s *Service) remotePlatform() string {
	if s.cfg.SelfPlatform == PlatformPopcorn {
		return PlatformPlanzi
	}
	return PlatformPopcorn
}

func hashHex(v string) string {
	h := sha256.Sum256([]byte(v))
	return hex.EncodeToString(h[:])
}

// ---- local user helpers ----

func (s *Service) findUserByEmail(db *gorm.DB, email string) (*user_module.User, error) {
	var u user_module.User
	err := db.Preload("Role").Where("lower(email) = lower(?)", strings.TrimSpace(email)).
		Order("(verified_email_at IS NOT NULL) DESC, id ASC").First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Service) loadUser(id uint) (*user_module.User, error) {
	var u user_module.User
	if err := s.db.Preload("Role").First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Service) toRemoteUser(u *user_module.User) *RemoteUser {
	ru := &RemoteUser{
		Platform:      s.cfg.SelfPlatform,
		UserID:        fmt.Sprintf("%d", u.ID),
		Email:         u.Email,
		EmailVerified: u.VerifiedEmailAt != nil,
	}
	if u.DisplayName != nil {
		ru.DisplayName = *u.DisplayName
	}
	if u.Password != nil {
		ru.PasswordHash = *u.Password
	}
	ru.IdentityResetAt = u.IdentityResetAt
	return ru
}

// ---- internal (server-to-server) endpoints ----

// Introspect atomically consumes a handoff token minted for the remote platform.
// It share-locks the user first, so a claim in flight (ProveEmail deletes the
// user's handoff tokens) finishes before the token is looked at, and only hands
// out accounts whose email is proven.
func (s *Service) Introspect(token string) (*RemoteUser, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}
	var ru *RemoteUser
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var uid uint
		res := tx.Raw(`SELECT user_id FROM link_handoff_tokens WHERE token_hash = ? AND target_platform = ?`,
			hashHex(token), s.remotePlatform()).Scan(&uid)
		if res.Error != nil || res.RowsAffected == 0 {
			return ErrInvalidToken
		}
		st, found, err := identity_policy.LockSession(tx, uid, false)
		if err != nil {
			return err
		}
		if !found || !st.Active {
			return ErrInvalidToken
		}
		var consumed uint
		res = tx.Raw(`UPDATE link_handoff_tokens SET used_at = ? WHERE token_hash = ? AND used_at IS NULL AND expires_at > ? AND target_platform = ? RETURNING user_id`,
			time.Now(), hashHex(token), time.Now(), s.remotePlatform()).Scan(&consumed)
		if res.Error != nil || res.RowsAffected == 0 {
			return ErrInvalidToken
		}
		if !st.Verified {
			return ErrInvalidToken
		}
		var u user_module.User
		if err := tx.Preload("Role").First(&u, uid).Error; err != nil {
			return ErrInvalidToken
		}
		ru = s.toRemoteUser(&u)
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrInvalidToken) {
			return nil, ErrInvalidToken
		}
		log.Printf("ERROR link: introspect: %v", err)
		return nil, ErrInvalidToken
	}
	return ru, nil
}

// VerifyPassword checks email+password of an active, verified local user. An
// unverified account's password was set by whoever registered the address first.
func (s *Service) VerifyPassword(email, password string) (*RemoteUser, bool) {
	u, err := s.findUserByEmail(s.db, email)
	if err != nil || !u.IsActive || u.VerifiedEmailAt == nil || u.Password == nil || *u.Password == "" {
		return nil, false
	}
	if bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(password)) != nil {
		return nil, false
	}
	return s.toRemoteUser(u), true
}

// ---- outgoing handoff ----

func (s *Service) CreateHandoff(userID uint, sessionVersion int, target string) (string, int, error) {
	if !s.Enabled() {
		return "", 0, ErrDisabled
	}
	if target != s.remotePlatform() {
		return "", 0, ErrInvalidTarget
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", 0, err
	}
	token := base64.RawURLEncoding.EncodeToString(buf)
	row := &HandoffToken{
		TokenHash:      hashHex(token),
		UserID:         userID,
		TargetPlatform: target,
		ExpiresAt:      time.Now().Add(HandoffTTL),
		CreatedAt:      time.Now(),
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Minting a handoff hands the account to another platform: only for a live session.
		if err := identity_policy.EnsureSession(tx, userID, sessionVersion, false); err != nil {
			if errors.Is(err, identity_policy.ErrSessionStale) {
				return ErrSessionStale
			}
			return err
		}
		return tx.Create(row).Error
	})
	if err != nil {
		return "", 0, err
	}
	// opportunistic cleanup of old rows
	s.db.Exec(`DELETE FROM link_handoff_tokens WHERE expires_at < ?`, time.Now().Add(-time.Hour))
	return s.cfg.RemoteWebURL + "/auth/handoff#t=" + token, int(HandoffTTL.Seconds()), nil
}

// ---- redeem ----

func (s *Service) findLinkByRemote(remoteUserID string) (*AccountLink, error) {
	var l AccountLink
	err := s.db.Where("remote_platform = ? AND remote_user_id = ?", s.remotePlatform(), remoteUserID).First(&l).Error
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (s *Service) Redeem(ctx context.Context, token string) (*Result, error) {
	if !s.Enabled() {
		return nil, ErrDisabled
	}
	if token == "" {
		return nil, ErrInvalidToken
	}
	ru, err := s.remote.Introspect(ctx, s.remotePlatform(), token)
	if err != nil {
		if errors.Is(err, ErrRemoteRejected) {
			return nil, ErrInvalidToken
		}
		log.Printf("WARN link: introspect failed: %v", err)
		return nil, ErrRemoteFailed
	}
	return s.resolveRemoteUser(ru)
}

func (s *Service) resolveRemoteUser(ru *RemoteUser) (*Result, error) {
	ru.Email = strings.ToLower(strings.TrimSpace(ru.Email))
	if ru.Email == "" || ru.UserID == "" {
		return nil, ErrInvalidToken
	}
	link, linkErr := s.findLinkByRemote(ru.UserID)
	alreadyLinked := linkErr == nil

	var local *user_module.User
	if alreadyLinked {
		u, err := s.loadUser(link.LocalUserID)
		if err != nil || !u.IsActive {
			return nil, ErrUserInactive
		}
		local = u
	} else {
		u, lerr := s.findUserByEmail(s.db, ru.Email)
		if lerr == nil {
			if !u.IsActive {
				return nil, ErrUserInactive
			}
			local = u
		}
	}
	localExists := local != nil
	localVerified := localExists && local.VerifiedEmailAt != nil

	switch Decide(alreadyLinked, localExists, localVerified, ru.EmailVerified) {
	case DecisionLoginLinked:
		return &Result{User: local}, nil
	case DecisionReject:
		return nil, ErrRemoteUnverified
	case DecisionProvision:
		u, err := s.provision(ru)
		if err != nil {
			return nil, err
		}
		return &Result{User: u, Provisioned: true}, nil
	case DecisionAutoLink:
		if err := s.createLink(s.db, local.ID, ru.UserID, ViaAuto); err != nil {
			return nil, err
		}
		return &Result{User: local}, nil
	default:
		// A link that already points at this local account is not a conflict; proving the
		// mailbox replaces it. Any other existing link to this account is.
		if !alreadyLinked && s.hasLinkTo(local.ID) {
			return nil, ErrLinkConflict
		}
		p := &LinkPending{
			ID:             uuid.NewString(),
			LocalUserID:    local.ID,
			RemotePlatform: s.remotePlatform(),
			RemoteUserID:   ru.UserID,
			RemoteEmail:    ru.Email,
			ExpiresAt:      time.Now().Add(PendingTTL),
			CreatedAt:      time.Now(),
		}
		if err := s.db.Create(p).Error; err != nil {
			return nil, err
		}
		return nil, &ConfirmRequiredError{
			PendingID: p.ID,
			Email:     ru.Email,
			Methods:   ConfirmMethods(local.Password != nil && *local.Password != "", localVerified),
		}
	}
}

func (s *Service) hasLinkTo(localUserID uint) bool {
	var n int64
	s.db.Model(&AccountLink{}).Where("local_user_id = ? AND remote_platform = ?", localUserID, s.remotePlatform()).Count(&n)
	return n > 0
}

func (s *Service) createLink(db *gorm.DB, localUserID uint, remoteUserID, via string) error {
	l := &AccountLink{
		LocalUserID:    localUserID,
		RemotePlatform: s.remotePlatform(),
		RemoteUserID:   remoteUserID,
		LinkedVia:      via,
		LinkedAt:       time.Now(),
	}
	if err := db.Create(l).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
			return ErrLinkConflict
		}
		return err
	}
	return nil
}

// provision creates a local user from the remote one and links it. Never grants admin.
func (s *Service) provision(ru *RemoteUser) (*user_module.User, error) {
	if !ru.EmailVerified {
		return nil, ErrRemoteUnverified // defence in depth: Decide already rejects this
	}
	hash := ru.PasswordHash
	if !strings.HasPrefix(hash, "$2") {
		raw := make([]byte, 32)
		if _, err := rand.Read(raw); err != nil {
			return nil, err
		}
		b, err := bcrypt.GenerateFromPassword([]byte(base64.RawURLEncoding.EncodeToString(raw)), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		hash = string(b)
	}
	var created *user_module.User
	err := s.db.Transaction(func(tx *gorm.DB) error {
		username := GenerateUsername(ru.Email, func(name string) bool {
			var n int64
			tx.Model(&user_module.User{}).Where("username = ?", name).Count(&n)
			return n > 0
		})
		u := &user_module.User{
			Username: username,
			Email:    strings.ToLower(strings.TrimSpace(ru.Email)),
			Password: &hash,
			RoleID:   2,
			IsActive: true,
		}
		if ru.DisplayName != "" {
			dn := ru.DisplayName
			if len(dn) > 100 {
				dn = dn[:100]
			}
			u.DisplayName = &dn
		}
		if ru.EmailVerified {
			now := time.Now()
			u.VerifiedEmailAt = &now
		}
		if err := tx.Create(u).Error; err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
				return ErrEmailNotAllowed
			}
			return err
		}
		if err := s.createLink(tx, u.ID, ru.UserID, ViaProvision); err != nil {
			return err
		}
		created = u
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.loadUser(created.ID)
}

// LoginViaRemote implements auth_module.LoginFallback.
func (s *Service) LoginViaRemote(email, password string) (*user_module.User, error) {
	if !s.Enabled() {
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	ru, err := s.remote.VerifyPassword(ctx, s.remotePlatform(), email, password)
	if err != nil {
		log.Printf("WARN link: remote verify-password failed: %v", err)
		return nil, nil
	}
	if ru == nil || !ru.EmailVerified {
		// An unverified remote account proves nothing: the same as wrong credentials.
		return nil, nil
	}
	// Never auto-link/sign into an existing local account via login fallback:
	// auth's FindByEmail is case-sensitive, so a case-variant email can reach here.
	if _, err := s.findUserByEmail(s.db, email); err == nil {
		return nil, nil
	}
	if _, err := s.findUserByEmail(s.db, ru.Email); err == nil {
		if _, lerr := s.findLinkByRemote(ru.UserID); lerr != nil {
			return nil, nil
		}
	}
	res, err := s.resolveRemoteUser(ru)
	if err != nil {
		var ce *ConfirmRequiredError
		if !errors.As(err, &ce) {
			log.Printf("WARN link: login fallback resolve failed: %v", err)
		}
		return nil, nil
	}
	return res.User, nil
}

// ---- confirm ----

func (s *Service) claimAttempt(pendingID string) (*LinkPending, error) {
	if _, err := uuid.Parse(pendingID); err != nil {
		return nil, ErrPendingInvalid
	}
	var p LinkPending
	res := s.db.Raw(`UPDATE link_pending SET attempts = attempts + 1 WHERE id = ? AND attempts < ? AND expires_at > ?
		RETURNING id, local_user_id, remote_platform, remote_user_id, remote_email, attempts, code_hash, code_expires_at, code_email, expires_at, created_at`,
		pendingID, MaxAttempts, time.Now()).Scan(&p)
	if res.Error != nil || res.RowsAffected == 0 {
		return nil, ErrPendingInvalid
	}
	return &p, nil
}

func (s *Service) Confirm(pendingID, method, password, code string) (*Result, error) {
	if !s.Enabled() {
		return nil, ErrDisabled
	}
	if method != MethodPasswd && method != MethodEmail {
		return nil, ErrInvalidMethod
	}
	p, err := s.claimAttempt(pendingID)
	if err != nil {
		return nil, err
	}
	u, err := s.loadUser(p.LocalUserID)
	if err != nil || !u.IsActive {
		return nil, ErrPendingInvalid
	}
	switch method {
	case MethodPasswd:
		if u.VerifiedEmailAt == nil || u.Password == nil || *u.Password == "" ||
			bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(password)) != nil {
			return nil, ErrBadCredentials
		}
	case MethodEmail:
		// The code only counts while it was mailed to the address the account has now.
		if p.CodeHash == nil || p.CodeExpiresAt == nil || time.Now().After(*p.CodeExpiresAt) || len(code) != 6 ||
			p.CodeEmail == nil || !identity_policy.SameEmail(*p.CodeEmail, u.Email) ||
			subtle.ConstantTimeCompare([]byte(*p.CodeHash), []byte(codeHash(p.ID, code))) != 1 {
			return nil, ErrBadCredentials
		}
	}
	via := ViaPassword
	if method == MethodEmail {
		via = ViaEmailCode
	}
	var proof identity_policy.Outcome
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if method == MethodEmail {
			// The code proves the mailbox, not the person: the policy wipes whatever an
			// unverified account carries before the link is created.
			var err error
			proof, err = identity_policy.ProveEmail(tx, identity_policy.Proof{
				UserID: u.ID, Email: *p.CodeEmail, Kind: identity_policy.KindLinkCode,
			})
			if err != nil {
				if errors.Is(err, identity_policy.ErrEmailMismatch) || errors.Is(err, identity_policy.ErrUserNotFound) || errors.Is(err, identity_policy.ErrUserInactive) {
					return ErrBadCredentials
				}
				return err
			}
		}
		if err := s.createLink(tx, u.ID, p.RemoteUserID, via); err != nil {
			return err
		}
		return tx.Exec(`DELETE FROM link_pending WHERE id = ?`, p.ID).Error
	})
	if err != nil {
		return nil, err
	}
	proof.Publish(u.ID)
	u, err = s.loadUser(u.ID)
	if err != nil {
		return nil, err
	}
	return &Result{User: u}, nil
}

func codeHash(pendingID, code string) string { return hashHex(pendingID + ":" + code) }

func (s *Service) SendCode(pendingID string) error {
	if !s.Enabled() {
		return ErrDisabled
	}
	if _, err := uuid.Parse(pendingID); err != nil {
		return ErrPendingInvalid
	}
	var p LinkPending
	if err := s.db.Where("id = ? AND attempts < ? AND expires_at > ?", pendingID, MaxAttempts, time.Now()).First(&p).Error; err != nil {
		return ErrPendingInvalid
	}
	u, err := s.loadUser(p.LocalUserID)
	if err != nil || !u.IsActive {
		return ErrPendingInvalid
	}
	n := randInt(1000000)
	code := fmt.Sprintf("%06d", n)
	exp := time.Now().Add(CodeTTL)
	if err := s.db.Model(&LinkPending{}).Where("id = ?", p.ID).
		Updates(map[string]any{"code_hash": codeHash(p.ID, code), "code_expires_at": exp, "code_email": u.Email}).Error; err != nil {
		return err
	}
	if s.mailer == nil || !s.mailer.IsEnabled() {
		log.Printf("WARN link: mailer disabled, code email not sent for user %d", u.ID)
		return nil
	}
	if err := s.mailer.SendLinkCodeEmail(u.Email, u.Username, code); err != nil {
		log.Printf("ERROR link: send code email: %v", err)
		return ErrCodeNotAllowed
	}
	return nil
}

// ---- status / unlink ----

func (s *Service) Status(userID uint) ([]AccountLink, error) {
	links := []AccountLink{}
	err := s.db.Where("local_user_id = ?", userID).Order("linked_at").Find(&links).Error
	return links, err
}

func (s *Service) Unlink(userID uint, platform string) error {
	res := s.db.Where("local_user_id = ? AND remote_platform = ?", userID, platform).Delete(&AccountLink{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotLinked
	}
	return nil
}

package auth_module

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/identity_policy"
	"github.com/arinsuda/movie-hub/internal/user_module"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	svc *Service
	cfg *config.Config
}

func NewHandler(svc *Service, cfg *config.Config) *Handler {
	return &Handler{svc: svc, cfg: cfg}
}

func (h *Handler) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	user, err := h.svc.Register(req)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(AuthResponse{
		User: toUserResponse(user),
	})
}

func (h *Handler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	pair, user, err := h.svc.Login(req, c.Get("User-Agent"), c.IP())
	if err != nil {
		return h.handleServiceError(c, err)
	}

	h.setTokenCookies(c, pair)

	return c.JSON(AuthResponse{
		User:         toUserResponse(user),
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}

func (h *Handler) Refresh(c fiber.Ctx) error {
	rawRefresh := c.Cookies("refresh_token")
	if rawRefresh == "" {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		_ = c.Bind().JSON(&body)
		if body.RefreshToken != "" {
			rawRefresh = body.RefreshToken
		}
	}
	if rawRefresh == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing refresh token"})
	}

	pair, user, err := h.svc.Refresh(rawRefresh, c.Get("User-Agent"), c.IP())
	if err != nil {
		h.clearTokenCookies(c)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
	}

	h.setTokenCookies(c, pair)

	return c.JSON(AuthResponse{
		User:         toUserResponse(user),
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}

func (h *Handler) Logout(c fiber.Ctx) error {
	rawRefresh := c.Cookies("refresh_token")
	if rawRefresh != "" {
		_ = h.svc.Logout(rawRefresh)
	}
	h.clearTokenCookies(c)
	return c.JSON(MessageResponse{Message: "logged out successfully"})
}

func (h *Handler) LogoutAll(c fiber.Ctx) error {
	claims := GetClaims(c)
	_ = h.svc.LogoutAll(claims.UserID)
	h.clearTokenCookies(c)
	return c.JSON(MessageResponse{Message: "logged out from all devices"})
}

// VerifyEmailLink answers GET /auth/verify-email without changing anything: mail
// scanners and link previews fetch URLs, and verification must be a deliberate POST.
func (h *Handler) VerifyEmailLink(c fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"code":  "VERIFY_EMAIL_REQUIRES_POST",
		"error": "open the verification link in the app to confirm your email",
	})
}

// VerifyEmail proves the mailbox. The optional password is the account's current
// password: a match keeps the account intact, a wrong one is rejected with the token
// kept. Planted credentials are wiped only for an explicit skip_password, and the client
// is then told to send the owner through password reset.
func (h *Handler) VerifyEmail(c fiber.Ctx) error {
	var req struct {
		Token        string `json:"token"`
		Password     string `json:"password"`
		SkipPassword bool   `json:"skip_password"`
	}
	if err := c.Bind().JSON(&req); err != nil || req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing token"})
	}

	var password *string
	if req.Password != "" {
		password = &req.Password
	}
	res, err := h.svc.VerifyEmail(req.Token, password, req.SkipPassword)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidToken):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid or expired token"})
		case errors.Is(err, ErrVerifyWrongPassword):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "WRONG_PASSWORD", "error": "wrong password"})
		case errors.Is(err, ErrPasswordChoiceRequired):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "PASSWORD_OR_SKIP_REQUIRED", "error": "enter your password or choose to skip"})
		case errors.Is(err, ErrVerifyAttemptsExceeded):
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"code": "VERIFY_ATTEMPTS_EXCEEDED", "error": "too many wrong passwords, request a new verification email"})
		}
		log.Printf("ERROR verify email: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(fiber.Map{
		"message":                 "email verified successfully",
		"password_reset_required": res.PasswordResetRequired,
	})
}

func (h *Handler) ResendVerification(c fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
	}
	if err := c.Bind().JSON(&body); err != nil || body.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email required"})
	}

	if err := h.svc.ResendVerification(body.Email); err != nil {
		if errors.Is(err, ErrAlreadyVerified) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "email already verified"})
		}
	}

	return c.JSON(MessageResponse{Message: "if the email exists, a verification link has been sent"})
}

func (h *Handler) ForgotPassword(c fiber.Ctx) error {
	var req user_module.ForgotPasswordRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email is required"})
	}

	_ = h.svc.ForgotPassword(req.Email)

	return c.JSON(MessageResponse{Message: "if an account with that email exists, a password reset link has been sent"})
}

func (h *Handler) ResetPassword(c fiber.Ctx) error {
	var req user_module.ResetPasswordRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "token is required"})
	}
	if req.UserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id is required"})
	}
	if req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "new_password is required"})
	}
	if req.ConfirmPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "confirm_password is required"})
	}

	if err := h.svc.ResetPassword(req.UserID, req.Token, req); err != nil {
		return h.handlePasswordResetError(c, err)
	}

	return c.JSON(MessageResponse{Message: "password has been reset successfully"})
}

func (h *Handler) GoogleLogin(c fiber.Ctx) error {
	if !h.cfg.Google.Enabled {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "google authentication is disabled"})
	}

	returnURL := c.Query("return_url")
	authURL, state, err := h.svc.InitiateGoogleLogin(c.Context(), returnURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to initiate google login"})
	}

	h.setOAuthStateCookie(c, state)

	return c.Redirect().To(authURL)
}

func (h *Handler) GoogleLink(c fiber.Ctx) error {
	if !h.cfg.Google.Enabled {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "google authentication is disabled"})
	}

	claims := GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req struct {
		ReturnURL string `json:"return_url"`
	}
	_ = c.Bind().JSON(&req)

	authURL, state, err := h.svc.InitiateGoogleLink(c.Context(), claims.UserID, req.ReturnURL)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	h.setOAuthStateCookie(c, state)

	return c.JSON(fiber.Map{
		"authorization_url": authURL,
	})
}

func (h *Handler) GoogleCallback(c fiber.Ctx) error {
	defer h.clearOAuthStateCookie(c)

	if !h.cfg.Google.Enabled {
		return c.Redirect().To(h.cfg.Google.FrontendErrorURL + "?error=google_oauth_disabled")
	}

	errParam := c.Query("error")
	if errParam != "" {
		code := "google_login_failed"
		if errParam == "access_denied" {
			code = "user_cancelled"
		}
		return c.Redirect().To(h.buildErrorRedirectURL(code, "/login"))
	}

	code := c.Query("code")
	stateParam := c.Query("state")
	cookieState := c.Cookies("popcorn_oauth_state")

	if code == "" || stateParam == "" || cookieState == "" {
		return c.Redirect().To(h.buildErrorRedirectURL("invalid_oauth_state", "/login"))
	}

	pair, _, returnURL, err := h.svc.HandleGoogleCallback(c.Context(), code, stateParam, cookieState, c.Get("User-Agent"), c.IP())
	if err != nil {
		errCode := "google_login_failed"
		switch {
		case errors.Is(err, ErrGoogleAccountLinkRequired):
			errCode = "google_account_link_required"
		case errors.Is(err, ErrGoogleIdentityAlreadyConnected):
			errCode = "google_identity_already_connected"
		case errors.Is(err, ErrUserInactive):
			errCode = "account_disabled"
		case errors.Is(err, ErrGoogleEmailVerificationRequired):
			errCode = "google_email_verification_required"
		case errors.Is(err, identity_policy.ErrSessionStale):
			errCode = "session_expired"
		case errors.Is(err, ErrInvalidOAuthState), errors.Is(err, ErrOAuthTransactionNotFound):
			errCode = "invalid_oauth_state"
		}
		return c.Redirect().To(h.buildErrorRedirectURL(errCode, returnURL))
	}

	if pair != nil {
		h.setTokenCookies(c, pair)
	}

	successURL := h.cfg.Google.FrontendSuccessURL
	if returnURL != "" && returnURL != "/" && strings.HasPrefix(returnURL, "/") && !strings.HasPrefix(returnURL, "//") && returnURL != "/q" {
		if u, err := url.Parse(successURL); err == nil {
			u.Path = returnURL
			successURL = u.String()
		}
	}

	if pair != nil {
		sep := "?"
		if strings.Contains(successURL, "?") {
			sep = "&"
		}
		successURL = fmt.Sprintf("%s%saccess_token=%s&refresh_token=%s", successURL, sep, url.QueryEscape(pair.AccessToken), url.QueryEscape(pair.RefreshToken))
	}

	return c.Redirect().To(successURL)
}

func (h *Handler) DisconnectGoogle(c fiber.Ctx) error {
	if !h.cfg.Google.Enabled {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "google authentication is disabled"})
	}

	claims := GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	if err := h.svc.DisconnectGoogleIdentity(c.Context(), claims.UserID); err != nil {
		if errors.Is(err, ErrCannotDisconnectFinalAuthMethod) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot disconnect your final authentication method"})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(MessageResponse{Message: "google account disconnected successfully"})
}

func (h *Handler) GoogleConfig(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"enabled": h.cfg.Google.Enabled,
	})
}

func (h *Handler) GoogleStatus(c fiber.Ctx) error {
	claims := GetClaims(c)
	if claims == nil || claims.UserID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	status, err := h.svc.GetGoogleOAuthStatus(claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get google oauth status"})
	}

	return c.JSON(status)
}

func (h *Handler) setOAuthStateCookie(c fiber.Ctx, state string) {
	c.Cookie(&fiber.Cookie{
		Name:     "popcorn_oauth_state",
		Value:    state,
		HTTPOnly: true,
		Secure:   h.cfg.Cookie.Secure,
		SameSite: "Lax",
		MaxAge:   600,
		Path:     "/api/auth/google",
	})
}

func (h *Handler) clearOAuthStateCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:    "popcorn_oauth_state",
		Value:   "",
		MaxAge:  -1,
		Expires: time.Unix(0, 0),
		Path:    "/api/auth/google",
	})
}

func (h *Handler) buildErrorRedirectURL(errCode string, returnPath string) string {
	errURL := h.cfg.Google.FrontendErrorURL
	u, err := url.Parse(errURL)
	if err != nil {
		return errURL + "?error=" + errCode
	}
	q := u.Query()
	q.Set("error", errCode)
	if returnPath != "" && returnPath != "/" {
		q.Set("return_url", returnPath)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (h *Handler) setTokenCookies(c fiber.Ctx, pair *TokenPair) {
	SetTokenCookies(c, h.cfg, pair)
}

// SetTokenCookies writes the access/refresh cookies for a freshly issued session.
func SetTokenCookies(c fiber.Ctx, cfg *config.Config, pair *TokenPair) {
	sameSite := cfg.Cookie.SameSite

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    pair.AccessToken,
		HTTPOnly: true,
		Secure:   cfg.Cookie.Secure,
		SameSite: sameSite,
		Domain:   cfg.Cookie.Domain,
		MaxAge:   int(cfg.JWT.AccessTTL.Seconds()),
		Path:     "/",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    pair.RefreshToken,
		HTTPOnly: true,
		Secure:   cfg.Cookie.Secure,
		SameSite: sameSite,
		Domain:   cfg.Cookie.Domain,
		MaxAge:   int(cfg.JWT.RefreshTTL.Seconds()),
		Path:     "/",
	})
}

func (h *Handler) clearTokenCookies(c fiber.Ctx) {
	for _, name := range []string{"access_token", "refresh_token"} {
		c.Cookie(&fiber.Cookie{
			Name:    name,
			Value:   "",
			MaxAge:  -1,
			Expires: time.Unix(0, 0),
			Path:    "/",
		})
	}
}

func (h *Handler) handlePasswordResetError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, user_module.ErrPasswordResetTokenNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "invalid or expired reset link"})
	case errors.Is(err, user_module.ErrPasswordResetTokenExpired):
		return c.Status(fiber.StatusGone).JSON(fiber.Map{"error": "reset link has expired, please request a new one"})
	case errors.Is(err, user_module.ErrPasswordResetTokenInvalid):
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "invalid reset token"})
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
}

func (h *Handler) handleServiceError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrPasswordMismatch):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, ErrEmailTaken), errors.Is(err, ErrUsernameTaken):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, ErrUserNotFound), errors.Is(err, ErrVerifyWrongPassword):
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	case errors.Is(err, ErrEmailUnverified):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "please verify your email first"})
	case errors.Is(err, ErrInvalidToken):
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
	default:
		log.Printf("ERROR [%s %s] unclassified service error: %v", c.Method(), c.Path(), err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}

func ToUserResponse(u *user_module.User) UserResponse { return toUserResponse(u) }

func toUserResponse(u *user_module.User) UserResponse {
	return UserResponse{
		ID:             u.ID,
		Username:       u.Username,
		Email:          u.Email,
		DisplayName:    u.DisplayName,
		Bio:            u.Bio,
		AvatarURL:      u.AvatarURL,
		DateOfBirth:    u.DateOfBirth,
		Gender:         string(u.Gender),
		IsVerified:     u.VerifiedEmailAt != nil,
		Role:           string(u.Role.RoleName),
		FavoriteGenres: u.FavoriteGenres,
		IsPrivate:      u.IsPrivate,
		Level:          1,
	}
}

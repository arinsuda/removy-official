package link_module

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/auth_module"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	svc *Service
	cfg *config.Config
}

func errJSON(c fiber.Ctx, status int, msg string) error {
	return c.Status(status).JSON(fiber.Map{"error": msg})
}

// requireSignature authenticates server-to-server calls via HMAC.
func (h *Handler) requireSignature(c fiber.Ctx) error {
	if !h.svc.Enabled() {
		return errJSON(c, fiber.StatusServiceUnavailable, "account linking is disabled")
	}
	err := VerifySignature(h.cfg.Link.Secrets, c.Get(HeaderTimestamp), c.Body(), c.Get(HeaderSignature), time.Now())
	if err != nil {
		return errJSON(c, fiber.StatusUnauthorized, "invalid signature")
	}
	return c.Next()
}

func (h *Handler) Introspect(c fiber.Ctx) error {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return errJSON(c, fiber.StatusBadRequest, "invalid request body")
	}
	ru, err := h.svc.Introspect(req.Token)
	if err != nil {
		return errJSON(c, fiber.StatusUnauthorized, "invalid token")
	}
	return c.JSON(ru)
}

func (h *Handler) VerifyPassword(c fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return errJSON(c, fiber.StatusBadRequest, "invalid request body")
	}
	ru, ok := h.svc.VerifyPassword(req.Email, req.Password)
	if !ok {
		return c.JSON(fiber.Map{"ok": false})
	}
	return c.JSON(fiber.Map{"ok": true, "user": ru})
}

func (h *Handler) CreateHandoff(c fiber.Ctx) error {
	claims := auth_module.GetClaims(c)
	if claims == nil {
		return errJSON(c, fiber.StatusUnauthorized, "unauthorized")
	}
	var req struct {
		Target string `json:"target"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return errJSON(c, fiber.StatusBadRequest, "invalid request body")
	}
	url, ttl, err := h.svc.CreateHandoff(claims.UserID, claims.TokenVersion, req.Target)
	if err != nil {
		return h.fail(c, err)
	}
	return c.JSON(fiber.Map{"redirect_url": url, "expires_in": ttl})
}

func (h *Handler) sessionResponse(c fiber.Ctx, res *Result) error {
	pair, err := h.svc.auth.IssueAndStoreTokens(res.User, c.Get("User-Agent"), c.IP())
	if errors.Is(err, auth_module.ErrEmailUnverified) {
		return errJSON(c, fiber.StatusForbidden, "email not verified")
	}
	if err != nil {
		log.Printf("ERROR link: issue tokens: %v", err)
		return errJSON(c, fiber.StatusInternalServerError, "internal server error")
	}
	auth_module.SetTokenCookies(c, h.cfg, pair)
	return c.JSON(fiber.Map{
		"user":          auth_module.ToUserResponse(res.User),
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"linked":        true,
		"provisioned":   res.Provisioned,
	})
}

func (h *Handler) Redeem(c fiber.Ctx) error {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.Bind().JSON(&req); err != nil || req.Token == "" {
		return errJSON(c, fiber.StatusBadRequest, "invalid request body")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	res, err := h.svc.Redeem(ctx, req.Token)
	if err != nil {
		return h.fail(c, err)
	}
	return h.sessionResponse(c, res)
}

func (h *Handler) Confirm(c fiber.Ctx) error {
	var req struct {
		PendingID string `json:"pending_id"`
		Method    string `json:"method"`
		Password  string `json:"password"`
		Code      string `json:"code"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return errJSON(c, fiber.StatusBadRequest, "invalid request body")
	}
	res, err := h.svc.Confirm(req.PendingID, req.Method, req.Password, req.Code)
	if err != nil {
		return h.fail(c, err)
	}
	return h.sessionResponse(c, res)
}

func (h *Handler) SendCode(c fiber.Ctx) error {
	var req struct {
		PendingID string `json:"pending_id"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return errJSON(c, fiber.StatusBadRequest, "invalid request body")
	}
	if err := h.svc.SendCode(req.PendingID); err != nil {
		return h.fail(c, err)
	}
	return c.JSON(fiber.Map{"message": "code sent"})
}

func (h *Handler) Status(c fiber.Ctx) error {
	claims := auth_module.GetClaims(c)
	if claims == nil {
		return errJSON(c, fiber.StatusUnauthorized, "unauthorized")
	}
	links, err := h.svc.Status(claims.UserID)
	if err != nil {
		return h.fail(c, err)
	}
	return c.JSON(fiber.Map{"links": links})
}

func (h *Handler) Unlink(c fiber.Ctx) error {
	claims := auth_module.GetClaims(c)
	if claims == nil {
		return errJSON(c, fiber.StatusUnauthorized, "unauthorized")
	}
	if err := h.svc.Unlink(claims.UserID, c.Params("platform")); err != nil {
		return h.fail(c, err)
	}
	return c.JSON(fiber.Map{"message": "unlinked"})
}

func (h *Handler) fail(c fiber.Ctx, err error) error {
	var ce *ConfirmRequiredError
	switch {
	case errors.As(err, &ce):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"code": CodeConfirmRC, "pending_id": ce.PendingID, "email": ce.Email, "methods": ce.Methods,
		})
	case errors.Is(err, ErrDisabled):
		return errJSON(c, fiber.StatusServiceUnavailable, err.Error())
	case errors.Is(err, ErrInvalidToken), errors.Is(err, ErrBadCredentials), errors.Is(err, ErrSessionStale):
		return errJSON(c, fiber.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrInvalidTarget), errors.Is(err, ErrInvalidMethod), errors.Is(err, ErrPendingInvalid):
		return errJSON(c, fiber.StatusBadRequest, err.Error())
	case errors.Is(err, ErrUserInactive):
		return errJSON(c, fiber.StatusForbidden, err.Error())
	case errors.Is(err, ErrRemoteUnverified):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"code": "REMOTE_EMAIL_UNVERIFIED", "error": err.Error()})
	case errors.Is(err, ErrLinkConflict), errors.Is(err, ErrEmailNotAllowed):
		return errJSON(c, fiber.StatusConflict, err.Error())
	case errors.Is(err, ErrNotLinked):
		return errJSON(c, fiber.StatusNotFound, err.Error())
	case errors.Is(err, ErrRemoteFailed), errors.Is(err, ErrCodeNotAllowed):
		return errJSON(c, fiber.StatusBadGateway, err.Error())
	default:
		log.Printf("ERROR link [%s %s]: %v", c.Method(), c.Path(), err)
		return errJSON(c, fiber.StatusInternalServerError, "internal server error")
	}
}

// Remote returns the other platform's public web URL so the handoff page can offer a way back.
// Public and token-free; hidden (404) when linking is disabled or no URL is configured.
func (h *Handler) Remote(c fiber.Ctx) error {
	if !h.svc.Enabled() || h.cfg.Link.RemoteWebURL == "" {
		return errJSON(c, fiber.StatusNotFound, "not found")
	}
	c.Set(fiber.HeaderCacheControl, "public, max-age=300")
	return c.JSON(fiber.Map{"web_url": h.cfg.Link.RemoteWebURL})
}

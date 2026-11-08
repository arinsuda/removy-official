package link_module

import (
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/auth_module"
	"github.com/arinsuda/movie-hub/internal/mailer"
	"github.com/arinsuda/movie-hub/internal/shared"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"gorm.io/gorm"
)

func rateLimit(hops, max int, window time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: window,
		KeyGenerator: func(c fiber.Ctx) string {
			return shared.ClientIP(c, hops)
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "too many requests"})
		},
	})
}

// RegisterRoutes mounts account-linking routes on the /api group and installs the login fallback.
func RegisterRoutes(api fiber.Router, db *gorm.DB, cfg *config.Config, m *mailer.Mailer, authSvc *auth_module.Service) *Service {
	svc := NewService(db, cfg.Link, authSvc, m)
	h := &Handler{svc: svc, cfg: cfg}
	mw := auth_module.NewMiddleware(cfg, db)
	authSvc.SetLoginFallback(svc)

	internal := api.Group("/internal/link", rateLimit(cfg.TrustedProxyHops, 120, time.Minute), h.requireSignature)
	internal.Post("/introspect", h.Introspect)
	internal.Post("/verify-password", h.VerifyPassword)

	public := rateLimit(cfg.TrustedProxyHops, 10, time.Minute)
	api.Post("/auth/handoff/redeem", public, h.Redeem)
	api.Post("/auth/handoff/confirm", public, h.Confirm)
	api.Post("/auth/handoff/send-code", rateLimit(cfg.TrustedProxyHops, 5, time.Minute), h.SendCode)

	api.Get("/link/remote", h.Remote)
	api.Post("/link/handoff", mw.RequireAuth, h.CreateHandoff)
	api.Get("/link/status", mw.RequireAuth, h.Status)
	api.Delete("/link/:platform", mw.RequireAuth, h.Unlink)
	return svc
}

package auth_module

import (
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/mailer"
	"github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/internal/shared"
	"github.com/arinsuda/movie-hub/internal/shared/ratelimit"
	"github.com/arinsuda/movie-hub/internal/shared/storage"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func RegisterRoutes(router fiber.Router, db *gorm.DB, cfg *config.Config, m *mailer.Mailer, s *storage.MinIOClient, notifSvc *notification_module.Service) *Service {
	svc := NewService(db, cfg, m, s, notifSvc)
	h := NewHandler(svc, cfg)
	mw := NewMiddleware(cfg, db)

	auth := router.Group("/auth")
	hops := cfg.TrustedProxyHops
	ipKey := func(c fiber.Ctx) string { return "ip:" + shared.ClientIP(c, hops) }
	emailKey := ratelimit.BodyFieldKey("email")

	auth.Post("/register", ratelimit.New(5, time.Hour, ipKey), h.Register)
	auth.Post("/login",
		ratelimit.New(10, time.Minute, ipKey),
		ratelimit.New(5, time.Minute, ratelimit.BodyFieldKey("identifier")),
		h.Login)
	auth.Post("/refresh", h.Refresh)
	auth.Post("/logout", h.Logout)
	auth.Get("/verify-email", h.VerifyEmailLink)
	auth.Post("/verify-email", ratelimit.New(20, time.Hour, ipKey), h.VerifyEmail)
	auth.Post("/resend-verification",
		ratelimit.New(3, time.Hour, ipKey),
		ratelimit.New(3, time.Hour, emailKey),
		h.ResendVerification)
	auth.Post("/logout-all", mw.RequireAuth, h.LogoutAll)
	auth.Post("/forgot-password",
		ratelimit.New(3, time.Hour, ipKey),
		ratelimit.New(3, time.Hour, emailKey),
		h.ForgotPassword)
	auth.Post("/reset-password", h.ResetPassword)

	// Google OAuth 2.0 routes
	auth.Get("/google/config", h.GoogleConfig)
	auth.Get("/google/login", h.GoogleLogin)
	auth.Get("/google/callback", h.GoogleCallback)
	auth.Post("/google/link", mw.RequireAuth, h.GoogleLink)
	auth.Delete("/google/link", mw.RequireAuth, h.DisconnectGoogle)
	auth.Get("/google/status", mw.RequireAuth, h.GoogleStatus)

	return svc
}

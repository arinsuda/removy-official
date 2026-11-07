package auth_module

import (
	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/mailer"
	"github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/internal/shared/storage"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func RegisterRoutes(router fiber.Router, db *gorm.DB, cfg *config.Config, m *mailer.Mailer, s *storage.MinIOClient, notifSvc *notification_module.Service) *Service {
	svc := NewService(db, cfg, m, s, notifSvc)
	h := NewHandler(svc, cfg)
	mw := NewMiddleware(cfg, db)

	auth := router.Group("/auth")
	auth.Post("/register", h.Register)
	auth.Post("/login", h.Login)
	auth.Post("/refresh", h.Refresh)
	auth.Post("/logout", h.Logout)
	auth.Get("/verify-email", h.VerifyEmailLink)
	auth.Post("/verify-email", h.VerifyEmail)
	auth.Post("/resend-verification", h.ResendVerification)
	auth.Post("/logout-all", mw.RequireAuth, h.LogoutAll)
	auth.Post("/forgot-password", h.ForgotPassword)
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

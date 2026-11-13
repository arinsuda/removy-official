package user_stats_module

import (
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func RegisterPublicRoutes(router fiber.Router, db *gorm.DB, policy privacy_policy.UserAccessPolicy, guards ...fiber.Handler) {
	svc := NewService(db)
	h := NewHandler(svc, policy)

	mw.GetRoute(router, "/users/:userId<int>/stats", h.GetUserStats, guards...)
}

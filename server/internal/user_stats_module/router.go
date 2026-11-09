package user_stats_module

import (
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func RegisterRoutes(router fiber.Router, db *gorm.DB, policy privacy_policy.UserAccessPolicy) {
	svc := NewService(db)
	h := NewHandler(svc, policy)

	router.Group("/users/:userId").Get("/stats", h.GetUserStats)
}

package like_module

import (
	achievementsmodule "github.com/arinsuda/movie-hub/internal/achievements_module"
	"github.com/arinsuda/movie-hub/internal/feed_module"
	"github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func New(
	db *gorm.DB,
	achieve achievementsmodule.Service,
	notif *notification_module.Service,
	feed feed_module.Service,
	policy privacy_policy.UserAccessPolicy,
) *Handler {
	svc := NewService(db, achieve, notif, feed, policy)
	return NewHandler(svc)
}

func RegisterPublicRoutes(router fiber.Router, h *Handler, guards ...fiber.Handler) {
	mw.GetRoute(router, "/users/:userId<int>/likes", h.GetLikes, guards...)
}

func RegisterRoutes(router fiber.Router, h *Handler) {
	router.Post("/stats/:mediaType/:mediaId/like", h.Like)
	router.Delete("/stats/:mediaType/:mediaId/like", h.Unlike)
}

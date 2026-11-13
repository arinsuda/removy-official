package follow_module

import (
	achievementsmodule "github.com/arinsuda/movie-hub/internal/achievements_module"
	"github.com/arinsuda/movie-hub/internal/feed_module"
	notification_module "github.com/arinsuda/movie-hub/internal/notification_module"
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
	mw.GetRoute(router, "/users/:userId<int>/followers", h.GetFollowers, guards...)
	mw.GetRoute(router, "/users/:userId<int>/following", h.GetFollowing, guards...)
}

func RegisterRoutes(router fiber.Router, h *Handler) {
	users := router.Group("/users")
	users.Post("/:userId/follow", h.Follow)
	users.Delete("/:userId/follow", h.Unfollow)
	users.Get("/:userId/follow-status", h.GetFollowStatus)
	users.Get("/:userId/follow-stats", h.GetFollowStats)
	users.Get("/:userId/follow-requests", h.GetPendingRequests)
	users.Post("/:userId/follow-requests/:followerId/accept", h.AcceptRequest)
	users.Delete("/:userId/follow-requests/:followerId", h.RejectRequest)
}

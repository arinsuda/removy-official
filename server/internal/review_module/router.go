package review_module

import (
	achievementsmodule "github.com/arinsuda/movie-hub/internal/achievements_module"
	"github.com/arinsuda/movie-hub/internal/feed_module"
	notification_module "github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/arinsuda/movie-hub/internal/shared/storage"
	stats "github.com/arinsuda/movie-hub/internal/user_stats_module"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// New builds the handler shared by RegisterPublicRoutes and RegisterRoutes.
func New(
	db *gorm.DB,
	mc *storage.MinIOClient,
	exp stats.ExpAdder,
	achieve achievementsmodule.Service,
	notif *notification_module.Service,
	feed feed_module.Service,
	policy privacy_policy.UserAccessPolicy,
) *Handler {
	return NewHandler(NewService(db, mc, exp, achieve, notif, feed, policy))
}

// RegisterPublicRoutes mounts the media review reads guests may use (viewer id 0: the service
// only returns public reviews of accounts the policy lets a guest see). Register it on the
// root /api router before any auth group.
func RegisterPublicRoutes(router fiber.Router, h *Handler, guards ...fiber.Handler) {
	mw.GetRoute(router, "/reviews/media/:mediaType/:mediaId", h.GetMediaReviews, guards...)
	mw.GetRoute(router, "/reviews/media/:mediaType/:mediaId/rating", h.GetMediaRating, guards...)
}

func RegisterRoutes(router fiber.Router, h *Handler) {
	reviews := router.Group("/reviews")
	reviews.Post("", h.CreateReview)
	reviews.Get("/user/:userId", h.GetUserReviews)
	// Distinct from the public /media/:mediaType/:mediaId[/rating] routes: the extra /me segment never matches them.
	reviews.Get("/media/:mediaType/:mediaId/me", h.GetMyMediaReview)
	reviews.Patch("/:reviewId", h.UpdateReview)
	reviews.Delete("/:reviewId", h.DeleteReview)

	reviews.Post("/:reviewId/like", h.LikeReview)
	reviews.Delete("/:reviewId/like", h.UnlikeReview)

	reviews.Post("/:reviewId/helpful", h.MarkHelpful)
	reviews.Delete("/:reviewId/helpful", h.UnmarkHelpful)

	reviews.Post("/:reviewId/comments", h.CreateComment)
	reviews.Get("/:reviewId/comments", h.GetComments)
	reviews.Patch("/comments/:commentId", h.UpdateComment)
	reviews.Delete("/:reviewId/comments/:commentId", h.DeleteComment)
}

package event_module

import (
	"strconv"

	"github.com/arinsuda/movie-hub/internal/shared/ratelimit"
	"github.com/arinsuda/movie-hub/internal/shared/storage"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func New(db *gorm.DB, mc *storage.MinIOClient) *Handler {
	svc := NewService(db, mc)
	return NewHandler(svc)
}

func RegisterPublicRoutes(router fiber.Router, h *Handler, guards ...fiber.Handler) {
	mw.GetRoute(router, "/events", h.ListPublic, guards...)
	mw.GetRoute(router, "/events/:id<int>", h.GetPublic, guards...)
}

func RegisterRoutes(router fiber.Router, h *Handler) {
	voteKey := func(c fiber.Ctx) string {
		claims := mw.GetClaims(c)
		if claims == nil {
			return ""
		}
		return "event-vote:" + strconv.FormatUint(uint64(claims.UserID), 10)
	}
	// One limiter for both vote routes: casting and withdrawing count together.
	voteLimit := ratelimit.New(VoteRateMax, VoteRateWindow, voteKey)

	events := router.Group("/events")
	events.Get("/popup", h.Popup)
	events.Put("/:id<int>/vote", voteLimit, h.Vote)
	events.Delete("/:id<int>/vote", voteLimit, h.CancelVote)
}

// RegisterAdminRoutes mounts the event admin routes on the guarded /admin group returned by
// admin_module.RegisterRoutes, so every route here inherits RequireCurrentAdmin. Paths are
// relative to that group (final paths: /api/admin/events/*).
func RegisterAdminRoutes(adminGroup fiber.Router, h *Handler) {
	events := adminGroup.Group("/events")
	events.Get("", h.ListAdmin)
	events.Post("", h.Create)
	events.Put("/:id<int>", h.Update)
	events.Delete("/:id<int>", h.Delete)
	events.Post("/:id<int>/image", h.SetImage)
	events.Delete("/:id<int>/image", h.RemoveImage)
	events.Post("/:id<int>/finalize", h.Finalize)
	events.Put("/:id<int>/nominees", h.SetNominees)
	events.Get("/suggest-nominees", h.SuggestNominees)
	events.Post("/generate-monthly", h.GenerateMonthly)
}

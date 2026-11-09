package media_stats_module

import (
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// New builds the handler shared by RegisterPublicRoutes and RegisterRoutes.
func New(db *gorm.DB) *Handler { return NewHandler(NewService(db)) }

// RegisterPublicRoutes mounts the read-only stats guests may see. Register it on the root
// /api router before any auth group; recording a view stays authenticated.
func RegisterPublicRoutes(router fiber.Router, h *Handler, guards ...fiber.Handler) {
	mw.GetRoute(router, "/stats/:media_type/:media_id", h.GetStats, guards...)
}

func RegisterRoutes(router fiber.Router, h *Handler) {
	stats := router.Group("/stats/:media_type/:media_id")
	stats.Post("/view", h.RecordView) // POST /stats/movie/123/view
}

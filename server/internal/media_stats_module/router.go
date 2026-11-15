package media_stats_module

import (
	"strconv"
	"time"

	"github.com/arinsuda/movie-hub/internal/shared/ratelimit"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

const (
	// ViewRateMax and ViewRateWindow bound how many view records one signed-in user may send.
	// A view is one per title and day, so the budget only stops scripted inflation of the total.
	ViewRateMax    = 60
	ViewRateWindow = time.Minute
)

// New builds the handler shared by RegisterPublicRoutes and RegisterRoutes.
func New(db *gorm.DB) *Handler { return NewHandler(NewService(db)) }

// RegisterPublicRoutes mounts the read-only stats guests may see. Register it on the root
// /api router before any auth group; recording a view stays authenticated.
func RegisterPublicRoutes(router fiber.Router, h *Handler, guards ...fiber.Handler) {
	mw.GetRoute(router, "/stats/:media_type/:media_id", h.GetStats, guards...)
}

func RegisterRoutes(router fiber.Router, h *Handler) {
	viewKey := func(c fiber.Ctx) string {
		claims := mw.GetClaims(c)
		if claims == nil {
			return ""
		}
		return "media-view:" + strconv.FormatUint(uint64(claims.UserID), 10)
	}
	viewLimit := ratelimit.New(ViewRateMax, ViewRateWindow, viewKey)

	stats := router.Group("/stats/:media_type/:media_id")
	stats.Post("/view", ValidateMediaParams, viewLimit, h.RecordView) // POST /stats/movie/123/view
}

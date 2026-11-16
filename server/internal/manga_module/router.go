package manga_module

import (
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
)

// RegisterPublicRoutes mounts the guest-readable manga endpoints on the root /api router, before the
// auth group. searchGuard is an extra stricter limiter applied to the search route only.
func RegisterPublicRoutes(router fiber.Router, h *Handler, searchGuard fiber.Handler, guards ...fiber.Handler) {
	searchGuards := append(append([]fiber.Handler{}, guards...), searchGuard)
	mw.GetRoute(router, "/manga", h.List, guards...)
	mw.GetRoute(router, "/manga/search", h.Search, searchGuards...)
	mw.GetRoute(router, "/manga/genres", h.Genres, guards...)
	mw.GetRoute(router, "/manga/:id<int>", h.GetByID, guards...)
	mw.GetRoute(router, "/manga/:id<int>/similar", h.GetSimilar, guards...)
}

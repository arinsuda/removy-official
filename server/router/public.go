package router

import (
	"time"

	"github.com/arinsuda/movie-hub/internal/media_stats_module"
	"github.com/arinsuda/movie-hub/internal/movie_module"
	"github.com/arinsuda/movie-hub/internal/review_module"
	"github.com/arinsuda/movie-hub/internal/shared"
	"github.com/arinsuda/movie-hub/internal/shared/ratelimit"
	"github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
)

// Guest read budget per client IP. Signed-in users are not counted: these routes were
// unlimited for them before guests existed, and an account is already a cost for abusers.
const (
	PublicReadRateMax    = 300
	PublicReadRateWindow = time.Minute
)

// PublicReadLimiter limits guest reads per IP (IPv6 collapsed to /64 by shared.RateLimitIP).
func PublicReadLimiter(trustedProxyHops int) fiber.Handler {
	return ratelimit.New(PublicReadRateMax, PublicReadRateWindow, func(c fiber.Ctx) string {
		if middleware.GetClaims(c) != nil {
			return ""
		}
		return "public:" + shared.RateLimitIP(c, trustedProxyHops)
	})
}

// RegisterPublicRoutes mounts the guest-readable endpoints (detail pages, media reviews and
// stats) with OptionalAuth and the guest limiter. It must run on the root /api router BEFORE
// the `protected := api.Group("/", RequireAuth)` line: in Fiber v3 that group is a USE on /api,
// so routes registered after it never run before authentication.
func RegisterPublicRoutes(
	api fiber.Router,
	optionalAuth, limiter fiber.Handler,
	movies *movie_module.Handler,
	reviews *review_module.Handler,
	stats *media_stats_module.Handler,
) {
	movie_module.RegisterPublicRoutes(api, movies, optionalAuth, limiter)
	review_module.RegisterPublicRoutes(api, reviews, optionalAuth, limiter)
	media_stats_module.RegisterPublicRoutes(api, stats, optionalAuth, limiter)
}

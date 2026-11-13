package router

import (
	"time"

	achievementsmodule "github.com/arinsuda/movie-hub/internal/achievements_module"
	"github.com/arinsuda/movie-hub/internal/bmol_module"
	"github.com/arinsuda/movie-hub/internal/event_module"
	"github.com/arinsuda/movie-hub/internal/follow_module"
	"github.com/arinsuda/movie-hub/internal/library_module"
	"github.com/arinsuda/movie-hub/internal/like_module"
	"github.com/arinsuda/movie-hub/internal/media_stats_module"
	"github.com/arinsuda/movie-hub/internal/movie_module"
	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/arinsuda/movie-hub/internal/review_module"
	"github.com/arinsuda/movie-hub/internal/shared"
	"github.com/arinsuda/movie-hub/internal/shared/ratelimit"
	"github.com/arinsuda/movie-hub/internal/user_module"
	"github.com/arinsuda/movie-hub/internal/user_stats_module"
	"github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// Guest read budget per client IP. Signed-in users are not counted: these routes were
// unlimited for them before guests existed, and an account is already a cost for abusers.
const (
	PublicReadRateMax    = 300
	PublicReadRateWindow = time.Minute
)

// Guests hitting TMDB-backed search get a tighter, separate budget per IP.
const (
	PublicSearchRateMax    = 30
	PublicSearchRateWindow = time.Minute
)

// PublicSearchLimiter limits guest /movies/search and /tv/search per IP; signed-in users are not counted.
func PublicSearchLimiter(trustedProxyHops int) fiber.Handler {
	return ratelimit.New(PublicSearchRateMax, PublicSearchRateWindow, func(c fiber.Ctx) string {
		if middleware.GetClaims(c) != nil {
			return ""
		}
		return "public-search:" + shared.RateLimitIP(c, trustedProxyHops)
	})
}

// PublicReadLimiter limits guest reads per IP (IPv6 collapsed to /64 by shared.RateLimitIP).
func PublicReadLimiter(trustedProxyHops int) fiber.Handler {
	return ratelimit.New(PublicReadRateMax, PublicReadRateWindow, func(c fiber.Ctx) string {
		if middleware.GetClaims(c) != nil {
			return ""
		}
		return "public:" + shared.RateLimitIP(c, trustedProxyHops)
	})
}

// PublicHandlers bundles the module handlers whose read-only routes are exposed to guests.
type PublicHandlers struct {
	Movies       *movie_module.Handler
	Reviews      *review_module.Handler
	MediaStats   *media_stats_module.Handler
	Users        *user_module.Handler
	Achievements *achievementsmodule.Module
	Library      *library_module.Handler
	Likes        *like_module.Handler
	BMOL         *bmol_module.Handler
	Follows      *follow_module.Handler
	Events       *event_module.Handler
	DB           *gorm.DB
	Policy       privacy_policy.UserAccessPolicy
}

// RegisterPublicRoutes mounts the guest-readable endpoints (detail pages, media reviews/stats,
// and public profile reads) with OptionalAuth and the guest limiter. It must run on the root /api
// router BEFORE the `protected := api.Group("/", RequireAuth)` line: in Fiber v3 that group is a
// USE on /api, so routes registered after it never run before authentication.
func RegisterPublicRoutes(
	api fiber.Router,
	optionalAuth, limiter, searchLimiter fiber.Handler,
	h PublicHandlers,
) {
	movie_module.RegisterPublicRoutes(api, h.Movies, searchLimiter, optionalAuth, limiter)
	review_module.RegisterPublicRoutes(api, h.Reviews, optionalAuth, limiter)
	media_stats_module.RegisterPublicRoutes(api, h.MediaStats, optionalAuth, limiter)
	user_module.RegisterPublicRoutes(api, h.Users, optionalAuth, limiter)
	user_stats_module.RegisterPublicRoutes(api, h.DB, h.Policy, optionalAuth, limiter)
	h.Achievements.RegisterPublicRoutes(api, optionalAuth, limiter)
	library_module.RegisterPublicRoutes(api, h.Library, optionalAuth, limiter)
	like_module.RegisterPublicRoutes(api, h.Likes, optionalAuth, limiter)
	bmol_module.RegisterPublicRoutes(api, h.BMOL, optionalAuth, limiter)
	follow_module.RegisterPublicRoutes(api, h.Follows, optionalAuth, limiter)
	event_module.RegisterPublicRoutes(api, h.Events, optionalAuth, limiter)
}

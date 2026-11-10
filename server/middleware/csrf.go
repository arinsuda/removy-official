package middleware

import (
	"strings"

	"github.com/arinsuda/movie-hub/internal/shared/apierr"
	"github.com/gofiber/fiber/v3"
)

// Header and value every unsafe /api request from the SPA carries (see client/src/api/index.ts).
const (
	HeaderRequestedWith = "X-Requested-With"
	RequestedWithXHR    = "XMLHttpRequest"
	msgXHRRequired      = "missing X-Requested-With header"
)

// CSRFExemptPrefixes lists /api path prefixes that skip the header check. Add a prefix only when
// the route is NOT authenticated by a browser cookie. Never exempt login, register, refresh or logout.
var CSRFExemptPrefixes = []string{
	// PLANZI server-to-server calls, authenticated by an HMAC signature (no cookies, no browser).
	"/api/internal/link/",
}

// RequireXHR rejects POST/PUT/PATCH/DELETE without X-Requested-With: XMLHttpRequest. The JSON
// binder ignores Content-Type, so a cross-site form (text/plain or multipart) would otherwise
// carry the SameSite=None cookie. A browser cannot add a custom header cross-site without a CORS
// preflight, which only the configured origins pass.
func RequireXHR(exempt []string) fiber.Handler {
	return func(c fiber.Ctx) error {
		switch c.Method() {
		case fiber.MethodPost, fiber.MethodPut, fiber.MethodPatch, fiber.MethodDelete:
		default:
			return c.Next()
		}
		if c.Get(HeaderRequestedWith) == RequestedWithXHR {
			return c.Next()
		}
		path := c.Path()
		for _, p := range exempt {
			if strings.HasPrefix(path, p) {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"code": apierr.ErrorCodeXHRRequired, "error": msgXHRRequired})
	}
}

package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

// SecurityHeaders sets baseline response headers for the API. Fiber serves only JSON (no
// HTML or file endpoints), so a deny-all CSP is safe on every response. Headers are set
// before the handler runs so error-handler responses carry them too. HSTS is only sent for
// https requests (directly or via X-Forwarded-Proto).
func SecurityHeaders() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "no-referrer")
		c.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		if c.Protocol() == "https" || strings.EqualFold(c.Get("X-Forwarded-Proto"), "https") {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		return c.Next()
	}
}

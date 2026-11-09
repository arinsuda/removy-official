package middleware

import "github.com/gofiber/fiber/v3"

// GetRoute registers a GET route that runs guards in order before handler.
//
// Public routes must be registered on the root /api router with per-route guards, never through
// api.Group("/", OptionalAuth, ...): in Fiber v3 a group with handlers registers a USE on its
// whole prefix, so it would run in front of every route registered after it.
func GetRoute(r fiber.Router, path string, handler fiber.Handler, guards ...fiber.Handler) {
	chain := make([]any, 0, len(guards)+1)
	for _, g := range guards {
		chain = append(chain, g)
	}
	chain = append(chain, handler)
	r.Get(path, chain[0], chain[1:]...)
}

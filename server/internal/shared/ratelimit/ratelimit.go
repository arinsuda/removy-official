// Package ratelimit holds in-memory limiters. It must not import internal/shared
// (shared -> notification_module -> user_module would form an import cycle).
package ratelimit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

// HashKey returns a SHA-256 hex digest of the normalized (trimmed, lowercased)
// value so identifiers and emails are never kept raw in limiter storage.
func HashKey(v string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(v))))
	return hex.EncodeToString(sum[:])
}

// New builds an in-memory fixed-window limiter. keyFn returns the bucket
// key; an empty key skips limiting for the request. Fiber sets Retry-After
// before LimitReached runs.
func New(max int, window time.Duration, keyFn func(fiber.Ctx) string) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: window,
		Next: func(c fiber.Ctx) bool {
			return keyFn(c) == ""
		},
		KeyGenerator: keyFn,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"code": "RATE_LIMITED", "error": "too many requests"})
		},
	})
}

// BodyFieldKey keys a limiter by the SHA-256 of a JSON body field. The body is
// read without consuming it. Requests without the field are not limited here.
func BodyFieldKey(field string) func(fiber.Ctx) string {
	return func(c fiber.Ctx) string {
		var body map[string]any
		if err := json.Unmarshal(c.Body(), &body); err != nil {
			return ""
		}
		s, _ := body[field].(string)
		if strings.TrimSpace(s) == "" {
			return ""
		}
		return "id:" + HashKey(s)
	}
}

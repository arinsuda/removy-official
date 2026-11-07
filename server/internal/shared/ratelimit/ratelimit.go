// Package ratelimit holds in-memory limiters. It must not import internal/shared
// (shared -> notification_module -> user_module would form an import cycle).
package ratelimit

import (
	"crypto/sha256"
	"encoding/hex"
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
	return build(max, window, keyFn, false)
}

// NewFailedOnly is New, but responses with status < 400 are not counted, so
// only failed attempts consume the budget.
func NewFailedOnly(max int, window time.Duration, keyFn func(fiber.Ctx) string) fiber.Handler {
	return build(max, window, keyFn, true)
}

func build(max int, window time.Duration, keyFn func(fiber.Ctx) string, failedOnly bool) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:                    max,
		Expiration:             window,
		SkipSuccessfulRequests: failedOnly,
		Next: func(c fiber.Ctx) bool {
			return keyFn(c) == ""
		},
		KeyGenerator: keyFn,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"code": "RATE_LIMITED", "error": "too many requests"})
		},
	})
}

// BodyKey keys a limiter by the SHA-256 of the identifier that extract pulls
// from the raw body. extract must unmarshal into the same DTO the handler
// binds, so field matching (case-insensitive, last value wins) is identical
// and the limiter cannot be bypassed with duplicate or re-cased keys. The body
// is not consumed. An empty extract result skips limiting.
func BodyKey(extract func(body []byte) string) func(fiber.Ctx) string {
	return func(c fiber.Ctx) string {
		s := extract(c.Body())
		if strings.TrimSpace(s) == "" {
			return ""
		}
		return "id:" + HashKey(s)
	}
}

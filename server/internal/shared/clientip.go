package shared

import (
	"net"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// ClientIP returns the caller's IP behind a reverse proxy. It takes the entry
// `hops` positions from the RIGHT of X-Forwarded-For (the proxy appends the
// connecting address, so the leftmost entries are client-controlled and never
// trusted). It falls back to c.IP() if the header is missing, too short, or the
// entry is not a valid IP.
func ClientIP(c fiber.Ctx, hops int) string {
	return clientIPFromHeader(c.Get(fiber.HeaderXForwardedFor), hops, c.IP())
}

func clientIPFromHeader(xff string, hops int, fallback string) string {
	if hops < 1 || strings.TrimSpace(xff) == "" {
		return fallback
	}
	parts := strings.Split(xff, ",")
	if len(parts) < hops {
		return fallback
	}
	candidate := strings.TrimSpace(parts[len(parts)-hops])
	ip := net.ParseIP(candidate)
	if ip == nil {
		return fallback
	}
	return ip.String()
}

// RateLimitIP is ClientIP shaped for limiter keys: IPv6 addresses collapse to
// their /64 so a client cannot rotate through its own prefix to dodge limits.
// IPv4 is unchanged.
func RateLimitIP(c fiber.Ctx, hops int) string {
	return limitKeyIP(ClientIP(c, hops))
}

func limitKeyIP(s string) string {
	ip := net.ParseIP(s)
	if ip == nil || ip.To4() != nil {
		return s
	}
	return ip.Mask(net.CIDRMask(64, 128)).String() + "/64"
}

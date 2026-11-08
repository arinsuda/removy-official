package router

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

const (
	shareProxyHost     = "image.tmdb.org"
	shareProxyPathBase = "/t/p/"
	shareProxyMaxBytes = 8 << 20

	shareProxyMaxInflight = 16 // concurrent upstream fetches
	shareProxyRatePerMin  = 30 // requests per IP per minute
	shareProxyMaxClients  = 10000
)

// Only these TMDB size segments are proxied; "original" and unknown sizes are refused.
var shareProxySizes = map[string]bool{"w92": true, "w185": true, "w342": true, "w500": true, "w780": true}

var shareProxyContentTypes = map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true, "image/avif": true}

// ipLimiter is a small per-key token bucket with a bounded map.
type ipLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    float64 // tokens per second
	burst   float64
	max     int
	now     func() time.Time
}

type bucket struct {
	tokens float64
	last   time.Time
}

func newIPLimiter(perMin int, maxClients int) *ipLimiter {
	return &ipLimiter{buckets: map[string]*bucket{}, rate: float64(perMin) / 60, burst: float64(perMin), max: maxClients, now: time.Now}
}

func (l *ipLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	b, ok := l.buckets[key]
	if !ok {
		if len(l.buckets) >= l.max {
			l.cleanup(now)
			if len(l.buckets) >= l.max {
				return false
			}
		}
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}
	b.tokens += now.Sub(b.last).Seconds() * l.rate
	if b.tokens > l.burst {
		b.tokens = l.burst
	}
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// cleanup drops buckets that have fully refilled (idle clients). Caller holds mu.
func (l *ipLimiter) cleanup(now time.Time) {
	for k, b := range l.buckets {
		if b.tokens+now.Sub(b.last).Seconds()*l.rate >= l.burst {
			delete(l.buckets, k)
		}
	}
}

// validShareImageURL accepts only https://image.tmdb.org/t/p/... with no
// credentials or custom port. Everything else is refused, so the proxy cannot
// be pointed at internal hosts (SSRF).
func validShareImageURL(raw string) (*url.URL, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil {
		return nil, false
	}
	if !strings.EqualFold(u.Hostname(), shareProxyHost) || (u.Port() != "" && u.Port() != "443") {
		return nil, false
	}
	if !strings.HasPrefix(u.Path, shareProxyPathBase) || strings.Contains(u.Path, "..") {
		return nil, false
	}
	return u, true
}

// shareImageSizeAllowed checks the size segment (/t/p/<size>/<file>).
func shareImageSizeAllowed(u *url.URL) bool {
	rest := strings.TrimPrefix(u.Path, shareProxyPathBase)
	size, file, found := strings.Cut(rest, "/")
	return found && file != "" && shareProxySizes[size]
}

// shareImageProxyHandler re-serves TMDB images with permissive CORS so the
// share cards can be rasterised without tainting the canvas.
func shareImageProxyHandler() fiber.Handler {
	return newShareImageProxyHandler(&http.Client{
		Timeout: 10 * time.Second,
		// A redirect could leave the allowlisted host; treat it as a failure.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}, shareProxyMaxInflight, shareProxyRatePerMin)
}

func newShareImageProxyHandler(client *http.Client, maxInflight, ratePerMin int) fiber.Handler {
	sem := make(chan struct{}, maxInflight)
	limiter := newIPLimiter(ratePerMin, shareProxyMaxClients)
	return func(c fiber.Ctx) error {
		if !limiter.allow(c.IP()) {
			c.Set("Retry-After", "60")
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "too many requests"})
		}
		rawURL := c.Query("url")
		if rawURL == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "url parameter is required"})
		}
		target, ok := validShareImageURL(rawURL)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "host not allowed"})
		}

		if !shareImageSizeAllowed(target) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "image size not allowed"})
		}

		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
		default:
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "proxy busy"})
		}

		// Rebuild the URL from the fixed host and path only; the query is dropped.
		resp, err := client.Get("https://" + shareProxyHost + target.EscapedPath())
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to fetch remote image"})
		}
		defer resp.Body.Close()
		contentType := strings.ToLower(strings.TrimSpace(strings.SplitN(resp.Header.Get("Content-Type"), ";", 2)[0]))
		if resp.StatusCode != http.StatusOK || !shareProxyContentTypes[contentType] {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to fetch remote image"})
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, shareProxyMaxBytes+1))
		if err != nil || len(body) > shareProxyMaxBytes {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to read image"})
		}

		c.Set("Content-Type", contentType)
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Cache-Control", "public, max-age=86400")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("Content-Security-Policy", "default-src 'none'")
		return c.Status(fiber.StatusOK).Send(body)
	}
}

package router

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/arinsuda/movie-hub/internal/shared"
)

const (
	shareProxyMaxBytes = 8 << 20
	shareProxyTimeout  = 10 * time.Second

	shareProxyMaxInflight = 16 // concurrent upstream fetches
	shareProxyRatePerMin  = 30 // requests per IP per minute
	shareProxyMaxClients  = 10000

	shareProxyRetryAfter = "60" // seconds, sent with 429
	shareProxyCacheHdr   = "public, max-age=86400"

	msgShareRateLimited    = "too many requests"
	msgShareURLRequired    = "url parameter is required"
	msgShareHostNotAllowed = "host not allowed"
	msgShareSizeNotAllowed = "image size not allowed"
	msgShareBusy           = "proxy busy"
	msgShareFetchFailed    = "failed to fetch remote image"
	msgShareReadFailed     = "failed to read image"
)

// shareProxyRule is one upstream image host the proxy serves. A URL must match a rule's exact host and path
// prefix; the size is the path segment right after the prefix (<pathBase><size>/<file>).
type shareProxyRule struct {
	host     string
	pathBase string
	sizes    map[string]bool
}

// shareProxyRules is the allowlist of upstream image hosts: TMDB posters (movie, TV) and AniList covers (manga).
// "original" and unknown sizes are refused.
var shareProxyRules = []shareProxyRule{
	{
		host:     "image.tmdb.org",
		pathBase: "/t/p/",
		sizes:    map[string]bool{"w92": true, "w185": true, "w342": true, "w500": true, "w780": true},
	},
	{
		host:     "s4.anilist.co",
		pathBase: "/file/anilistcdn/media/manga/cover/",
		sizes:    map[string]bool{"medium": true, "large": true, "extraLarge": true},
	},
}

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

// shareProxyRuleFor finds the rule whose host matches exactly (case-insensitive).
func shareProxyRuleFor(host string) (shareProxyRule, bool) {
	for _, rule := range shareProxyRules {
		if strings.EqualFold(host, rule.host) {
			return rule, true
		}
	}
	return shareProxyRule{}, false
}

// validShareImageURL accepts only https URLs on an allowlisted host, under that host's path prefix, with no
// credentials or custom port. Everything else is refused, so the proxy cannot be pointed at internal hosts (SSRF).
// The size is checked separately by shareImageSizeAllowed.
func validShareImageURL(raw string) (*url.URL, shareProxyRule, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil {
		return nil, shareProxyRule{}, false
	}
	if u.Port() != "" && u.Port() != "443" {
		return nil, shareProxyRule{}, false
	}
	rule, ok := shareProxyRuleFor(u.Hostname())
	if !ok || !strings.HasPrefix(u.Path, rule.pathBase) || strings.Contains(u.Path, "..") {
		return nil, shareProxyRule{}, false
	}
	return u, rule, true
}

// shareImageSizeAllowed checks the size segment (<pathBase><size>/<file>) against the rule's sizes.
func shareImageSizeAllowed(u *url.URL, rule shareProxyRule) bool {
	rest := strings.TrimPrefix(u.Path, rule.pathBase)
	size, file, found := strings.Cut(rest, "/")
	return found && file != "" && rule.sizes[size]
}

// shareImageProxyHandler re-serves upstream images with permissive CORS so the
// share cards can be rasterised without tainting the canvas.
func shareImageProxyHandler(proxyHops int) fiber.Handler {
	return newShareImageProxyHandler(&http.Client{
		Timeout: shareProxyTimeout,
		// A redirect could leave the allowlisted host; treat it as a failure.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}, shareProxyMaxInflight, shareProxyRatePerMin, proxyHops)
}

func newShareImageProxyHandler(client *http.Client, maxInflight, ratePerMin, proxyHops int) fiber.Handler {
	sem := make(chan struct{}, maxInflight)
	limiter := newIPLimiter(ratePerMin, shareProxyMaxClients)
	return func(c fiber.Ctx) error {
		if !limiter.allow(shared.RateLimitIP(c, proxyHops)) {
			c.Set("Retry-After", shareProxyRetryAfter)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": msgShareRateLimited})
		}
		rawURL := c.Query("url")
		if rawURL == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msgShareURLRequired})
		}
		target, rule, ok := validShareImageURL(rawURL)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": msgShareHostNotAllowed})
		}

		if !shareImageSizeAllowed(target, rule) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msgShareSizeNotAllowed})
		}

		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
		default:
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": msgShareBusy})
		}

		// Rebuild the URL from the fixed host and path only; the query is dropped.
		resp, err := client.Get("https://" + rule.host + target.EscapedPath())
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": msgShareFetchFailed})
		}
		defer resp.Body.Close()
		contentType := strings.ToLower(strings.TrimSpace(strings.SplitN(resp.Header.Get("Content-Type"), ";", 2)[0]))
		if resp.StatusCode != http.StatusOK || !shareProxyContentTypes[contentType] {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": msgShareFetchFailed})
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, shareProxyMaxBytes+1))
		if err != nil || len(body) > shareProxyMaxBytes {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": msgShareReadFailed})
		}

		c.Set("Content-Type", contentType)
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Cache-Control", shareProxyCacheHdr)
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("Content-Security-Policy", "default-src 'none'")
		return c.Status(fiber.StatusOK).Send(body)
	}
}

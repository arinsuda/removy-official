package router

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

const (
	shareProxyHost     = "image.tmdb.org"
	shareProxyPathBase = "/t/p/"
	shareProxyMaxBytes = 8 << 20
)

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

// shareImageProxyHandler re-serves TMDB images with permissive CORS so the
// share cards can be rasterised without tainting the canvas.
func shareImageProxyHandler() fiber.Handler {
	client := &http.Client{
		Timeout: 10 * time.Second,
		// A redirect could leave the allowlisted host; treat it as a failure.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	return func(c fiber.Ctx) error {
		rawURL := c.Query("url")
		if rawURL == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "url parameter is required"})
		}
		target, ok := validShareImageURL(rawURL)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "host not allowed"})
		}

		resp, err := client.Get(target.String())
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to fetch remote image"})
		}
		defer resp.Body.Close()
		contentType := resp.Header.Get("Content-Type")
		if resp.StatusCode != http.StatusOK || !strings.HasPrefix(contentType, "image/") {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to fetch remote image"})
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, shareProxyMaxBytes+1))
		if err != nil || len(body) > shareProxyMaxBytes {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to read image"})
		}

		c.Set("Content-Type", contentType)
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Cache-Control", "public, max-age=86400")
		return c.Status(fiber.StatusOK).Send(body)
	}
}

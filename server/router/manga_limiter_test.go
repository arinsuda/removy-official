package router

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mangaLimiterApp mounts one route behind MangaLimiter. The X-Test-User header stands in for
// OptionalAuth: when it holds a user id, that user's claims are set before the limiter runs.
func mangaLimiterApp() *fiber.App {
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		if id, err := strconv.Atoi(c.Get("X-Test-User")); err == nil && id > 0 {
			c.Locals("claims", &middleware.Claims{UserID: uint(id)})
		}
		return c.Next()
	})
	app.Get("/api/manga", MangaLimiter(0), func(c fiber.Ctx) error { return c.SendStatus(http.StatusOK) })
	return app
}

// hit sends one request as userID (0 = guest) and returns the status code.
func hit(t *testing.T, app *fiber.App, userID int) int {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/manga", nil)
	if userID > 0 {
		req.Header.Set("X-Test-User", strconv.Itoa(userID))
	}
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	return resp.StatusCode
}

func TestMangaLimiterCountsSignedInUsers(t *testing.T) {
	app := mangaLimiterApp()

	for i := range config.MangaRateMax {
		require.Equal(t, http.StatusOK, hit(t, app, 7), "request %d within the budget", i+1)
	}
	assert.Equal(t, http.StatusTooManyRequests, hit(t, app, 7), "signed-in user over the budget must be limited")
	assert.Equal(t, http.StatusOK, hit(t, app, 8), "another signed-in user has its own budget")
}

func TestMangaLimiterKeysGuestsByIP(t *testing.T) {
	app := mangaLimiterApp()

	for range config.MangaRateMax {
		require.Equal(t, http.StatusOK, hit(t, app, 0))
	}
	assert.Equal(t, http.StatusTooManyRequests, hit(t, app, 0), "guest over the per-IP budget must be limited")
	assert.Equal(t, http.StatusOK, hit(t, app, 9), "a signed-in user is not charged to the guest bucket")
}

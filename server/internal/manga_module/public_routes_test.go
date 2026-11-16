package manga_module

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPublicApp(t *testing.T) *fiber.App {
	t.Helper()
	up := newFakeAniList(t, standardUpstream)
	app := fiber.New()
	api := app.Group("/api")
	mw := middleware.NewAuthMiddleware(&config.Config{JWT: config.JWTConfig{AccessSecret: "test-secret"}}, nil)
	RegisterPublicRoutes(api, New(up.client(), matrixGate()), passThrough, mw.OptionalAuth)
	return app
}

func TestMangaPublicRoutesServeGuests(t *testing.T) {
	app := setupPublicApp(t)
	paths := []string{
		"/api/manga",
		"/api/manga?country=JP&format=manga&sort=popular&status=all&page=1",
		"/api/manga/search?q=dune",
		"/api/manga/genres",
		"/api/manga/1",
		"/api/manga/1/similar",
	}
	for _, p := range paths {
		t.Run("guest 200 "+p, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(http.MethodGet, p, nil))
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode, p)
		})
	}
}

func TestMangaPublicRoutesRejectInvalidBearer(t *testing.T) {
	app := setupPublicApp(t)
	for _, p := range []string{"/api/manga", "/api/manga/search?q=dune", "/api/manga/genres", "/api/manga/1", "/api/manga/1/similar"} {
		t.Run("invalid bearer 401 "+p, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, p, nil)
			req.Header.Set("Authorization", "Bearer not-a-jwt")
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, p)
		})
	}
}

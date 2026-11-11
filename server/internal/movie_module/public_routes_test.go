package movie_module

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/shared"
	tmdb "github.com/arinsuda/movie-hub/internal/tmdb_module"
	"github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubRatingStatsReader struct{}

func (stubRatingStatsReader) GetMediaRating(context.Context, shared.MediaIdentity) (shared.RatingStats, error) {
	return shared.RatingStats{}, nil
}

func (stubRatingStatsReader) GetBatchMediaRatings(context.Context, []shared.MediaIdentity) (map[shared.MediaKey]shared.RatingStats, error) {
	return map[shared.MediaKey]shared.RatingStats{}, nil
}

func startFakeTMDBServer(t *testing.T) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasPrefix(r.URL.Path, "/genre/"):
			_, _ = w.Write([]byte(`{"genres":[{"id":28,"name":"Action"}]}`))
		case strings.HasSuffix(r.URL.Path, "/watch/providers"):
			_, _ = w.Write([]byte(`{"id":603,"results":{}}`))
		case strings.HasSuffix(r.URL.Path, "/credits"):
			_, _ = w.Write([]byte(`{"id":603,"cast":[],"crew":[]}`))
		case strings.HasSuffix(r.URL.Path, "/videos"):
			_, _ = w.Write([]byte(`{"id":603,"results":[]}`))
		case strings.HasSuffix(r.URL.Path, "/movie_credits"), strings.HasSuffix(r.URL.Path, "/tv_credits"):
			_, _ = w.Write([]byte(`{"cast":[],"crew":[]}`))
		case strings.HasPrefix(r.URL.Path, "/person/"):
			_, _ = w.Write([]byte(`{"id":100,"name":"Keanu Reeves","biography":"Actor"}`))
		default:
			_, _ = w.Write([]byte(`{"id":603,"title":"Fake Movie","name":"Fake Show","overview":"o","poster_path":"/p.jpg","genres":[],"results":[],"page":1,"total_pages":1,"total_results":0}`))
		}
	}))
	t.Cleanup(srv.Close)
	tmdb.Init(&config.Config{TMDB: config.TMDBConfig{BaseURL: srv.URL, APIKey: "test"}})
}

func setupPublicTestApp(t *testing.T) *fiber.App {
	t.Helper()
	startFakeTMDBServer(t)
	app := fiber.New()
	api := app.Group("/api")
	mw := middleware.NewAuthMiddleware(&config.Config{JWT: config.JWTConfig{AccessSecret: "test-secret"}}, nil)
	h := New(stubRatingStatsReader{}, nil)
	RegisterPublicRoutes(api, h, func(c fiber.Ctx) error { return c.Next() }, mw.OptionalAuth)
	protected := api.Group("/", mw.RequireAuth)
	RegisterRoutes(protected, h)
	return app
}

func TestMovieModulePublicAndProtectedRoutes(t *testing.T) {
	app := setupPublicTestApp(t)

	publicPaths := []string{
		"/api/movies/popular",
		"/api/movies/now-playing",
		"/api/movies/top-rated",
		"/api/movies/upcoming",
		"/api/movies/search?q=dune",
		"/api/movies/genres",
		"/api/movies/603",
		"/api/movies/603/similar",
		"/api/movies/603/watch-providers",
		"/api/tv/popular",
		"/api/tv/now-airing",
		"/api/tv/search?q=dune",
		"/api/tv/genres",
		"/api/tv/1399",
		"/api/tv/1399/similar",
		"/api/tv/1399/watch-providers",
		"/api/actors/100",
		"/api/actors/100/movies",
		"/api/actors/100/tv",
	}

	for _, p := range publicPaths {
		t.Run("guest 200 "+p, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(http.MethodGet, p, nil))
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode, p)
		})
	}

	newlyPublicPaths := []string{
		"/api/movies/popular",
		"/api/movies/now-playing",
		"/api/movies/top-rated",
		"/api/movies/upcoming",
		"/api/movies/search?q=dune",
		"/api/movies/genres",
		"/api/tv/popular",
		"/api/tv/now-airing",
		"/api/tv/search?q=dune",
		"/api/tv/genres",
		"/api/actors/100",
		"/api/actors/100/movies",
		"/api/actors/100/tv",
	}

	for _, p := range newlyPublicPaths {
		t.Run("invalid bearer 401 "+p, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, p, nil)
			req.Header.Set("Authorization", "Bearer not-a-jwt")
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, p)
		})
	}

	protectedPaths := []string{
		"/api/movies/recommended",
		"/api/movies/upcoming/2026",
		"/api/tv/top-rated",
		"/api/actors/search?q=tom",
	}

	for _, p := range protectedPaths {
		t.Run("guest 401 protected "+p, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(http.MethodGet, p, nil))
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, p)
		})
	}
}

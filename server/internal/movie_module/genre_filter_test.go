package movie_module

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestGenreParamValidation(t *testing.T) {
	app := fiber.New()
	h := NewHandler(NewMovieService(nil))
	app.Get("/movies/popular", h.GetPopular)
	app.Get("/movies/now-playing", h.GetNowPlaying)
	app.Get("/movies/top-rated", h.GetTopRated)
	app.Get("/tv/popular", h.GetPopularSeries)
	app.Get("/tv/now-airing", h.GetNowAiringSeries)
	app.Get("/tv/top-rated", h.GetTopRatedSeries)

	for _, path := range []string{"/movies/popular", "/movies/now-playing", "/movies/top-rated", "/tv/popular", "/tv/now-airing", "/tv/top-rated"} {
		for _, bad := range []string{"abc", "28,", "28|35", "28;drop", "1,2,x", ",28"} {
			resp, err := app.Test(httptest.NewRequest("GET", path+"?with_genres="+bad, nil))
			assert.NoError(t, err)
			assert.Equal(t, 400, resp.StatusCode, path+" "+bad)
		}
	}
}

func TestUpcomingRejectsBadGenres(t *testing.T) {
	app := fiber.New()
	app.Get("/movies/upcoming", NewHandler(NewMovieService(nil)).GetUpcoming)
	for _, bad := range []string{"abc", "28,", "28|35", "1,2,x", ",28"} {
		resp, err := app.Test(httptest.NewRequest("GET", "/movies/upcoming?with_genres="+bad, nil))
		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode, bad)
	}
}

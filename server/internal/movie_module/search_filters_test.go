package movie_module

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestSearchRejectsBadYearAndGenre(t *testing.T) {
	app := fiber.New()
	h := NewHandler(NewMovieService(nil))
	app.Get("/movies/search", h.Search)
	app.Get("/tv/search", h.SearchSeries)

	for _, path := range []string{"/movies/search", "/tv/search"} {
		for _, bad := range []string{"year=abc", "year=1800", "year=2201", "year=-1", "genre=abc", "genre=0", "genre=1000000", "genre=28,35"} {
			resp, err := app.Test(httptest.NewRequest("GET", path+"?q=dune&"+bad, nil))
			assert.NoError(t, err)
			assert.Equal(t, 400, resp.StatusCode, path+" "+bad)
		}
	}
}

func TestSearchFiltersParse(t *testing.T) {
	app := fiber.New()
	app.Get("/x", func(c fiber.Ctx) error {
		f, ok := searchFilters(c)
		assert.True(t, ok)
		assert.Equal(t, 2021, f.Year)
		assert.Equal(t, 28, f.GenreID)
		return c.SendStatus(200)
	})
	resp, err := app.Test(httptest.NewRequest("GET", "/x?year=2021&genre=28", nil))
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

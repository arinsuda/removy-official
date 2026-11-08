package movie_module

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	tmdb "github.com/arinsuda/movie-hub/internal/tmdb_module"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func providersApp(calls *int) *fiber.App {
	svc := NewMovieService(nil)
	svc.fetchProviders = func(kind string, id int, region string) (*tmdb.WatchProviders, error) {
		*calls++
		return &tmdb.WatchProviders{Region: region, Flatrate: []tmdb.WatchProvider{}, Free: []tmdb.WatchProvider{},
			Ads: []tmdb.WatchProvider{}, Rent: []tmdb.WatchProvider{}, Buy: []tmdb.WatchProvider{}}, nil
	}
	h := NewHandler(svc)
	app := fiber.New()
	app.Get("/movies/:id/watch-providers", h.GetMovieWatchProviders)
	app.Get("/tv/:id/watch-providers", h.GetSeriesWatchProviders)
	return app
}

func TestWatchProvidersValidation(t *testing.T) {
	calls := 0
	app := providersApp(&calls)
	for _, u := range []string{
		"/movies/abc/watch-providers", "/tv/0/watch-providers",
		"/movies/1/watch-providers?region=th", "/tv/1/watch-providers?region=THA", "/tv/1/watch-providers?region=1A",
	} {
		resp, err := app.Test(httptest.NewRequest("GET", u, nil))
		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode, u)
	}
	assert.Equal(t, 0, calls)
}

func TestWatchProvidersDefaultRegionEmptyAndCache(t *testing.T) {
	calls := 0
	app := providersApp(&calls)
	for i := 0; i < 2; i++ {
		resp, err := app.Test(httptest.NewRequest("GET", "/movies/5/watch-providers", nil))
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		b, _ := io.ReadAll(resp.Body)
		var out struct {
			Providers tmdb.WatchProviders `json:"providers"`
		}
		assert.NoError(t, json.Unmarshal(b, &out))
		assert.Equal(t, "TH", out.Providers.Region)
		assert.Contains(t, string(b), `"flatrate":[]`)
	}
	assert.Equal(t, 1, calls)
}

func TestProvidersCacheExpiryAndBound(t *testing.T) {
	c := newProvidersCache()
	now := time.Now()
	c.now = func() time.Time { return now }
	c.set("a", &tmdb.WatchProviders{})
	_, ok := c.get("a")
	assert.True(t, ok)
	now = now.Add(providersTTL + time.Second)
	_, ok = c.get("a")
	assert.False(t, ok)
	for i := 0; i < providersMaxItems+10; i++ {
		c.set(string(rune('a'+i%26))+time.Duration(i).String(), &tmdb.WatchProviders{})
	}
	assert.LessOrEqual(t, len(c.items), providersMaxItems)
}

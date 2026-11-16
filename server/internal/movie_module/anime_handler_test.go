package movie_module

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func animeCurrentYear() int {
	return time.Now().In(config.AppLocation).Year()
}

func TestAnimeEndpointsGuest200(t *testing.T) {
	app := setupPublicTestApp(t)
	paths := []string{
		"/api/anime/tv",
		"/api/anime/tv?sort=airing&country=all&keyword=isekai&genre=10759",
		"/api/anime/tv?sort=newest&season=spring&year=2025&keyword=mecha&country=KR",
		"/api/anime/tv?decade=1990&sort=top_rated&page=2",
		"/api/anime/movies",
		"/api/anime/movies?sort=newest&year=2024&keyword=school",
		"/api/anime/movies?decade=2010&sort=top_rated",
	}
	for _, p := range paths {
		t.Run(p, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(http.MethodGet, p, nil))
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestAnimeEndpointsRejectBadInput(t *testing.T) {
	app := setupPublicTestApp(t)
	next := animeCurrentYear() + 2
	cases := []struct {
		name string
		path string
		code shared.ErrorCode
	}{
		{"unknown sort", "/api/anime/tv?sort=hot", shared.ErrorCodeInvalidSearch},
		{"airing on movies", "/api/anime/movies?sort=airing", shared.ErrorCodeInvalidSearch},
		{"unknown country", "/api/anime/tv?country=US", shared.ErrorCodeInvalidSearch},
		{"lowercase country", "/api/anime/tv?country=jp", shared.ErrorCodeInvalidSearch},
		{"unknown keyword", "/api/anime/tv?keyword=harem", shared.ErrorCodeInvalidSearch},
		{"genre 16 is implicit", "/api/anime/tv?genre=16", shared.ErrorCodeInvalidGenre},
		{"genre not a number", "/api/anime/tv?genre=abc", shared.ErrorCodeInvalidGenre},
		{"genre out of range", "/api/anime/movies?genre=9999999", shared.ErrorCodeInvalidGenre},
		{"decade before 1960", "/api/anime/tv?decade=1950", shared.ErrorCodeInvalidSearch},
		{"decade not a multiple of 10", "/api/anime/tv?decade=1975", shared.ErrorCodeInvalidSearch},
		{"decade in the future", fmt.Sprintf("/api/anime/tv?decade=%d", (animeCurrentYear()/10+1)*10), shared.ErrorCodeInvalidSearch},
		{"year too old", "/api/anime/tv?year=1959", shared.ErrorCodeInvalidSearch},
		{"year too far ahead", fmt.Sprintf("/api/anime/movies?year=%d", next), shared.ErrorCodeInvalidSearch},
		{"decade with year", "/api/anime/tv?decade=1990&year=1995", shared.ErrorCodeInvalidSearch},
		{"decade with season", "/api/anime/tv?decade=1990&season=fall&year=1995", shared.ErrorCodeInvalidSearch},
		{"season without year", "/api/anime/tv?season=summer", shared.ErrorCodeInvalidSearch},
		{"unknown season", "/api/anime/tv?season=monsoon&year=2020", shared.ErrorCodeInvalidSearch},
		{"season on movies", "/api/anime/movies?season=winter&year=2020", shared.ErrorCodeInvalidSearch},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(http.MethodGet, tc.path, nil))
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, http.StatusBadRequest, resp.StatusCode, tc.path)

			var body struct {
				Code string `json:"code"`
			}
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
			assert.Equal(t, string(tc.code), body.Code, tc.path)
		})
	}
}

func TestParseAnimeFilterAcceptsValidCombinations(t *testing.T) {
	app := setupPublicTestApp(t)
	cases := []struct {
		name   string
		path   string
		series bool
	}{
		{"airing tv", "/api/anime/tv?sort=airing", true},
		{"season with year on tv", "/api/anime/tv?season=fall&year=2025", true},
		{"decade alone", "/api/anime/movies?decade=1980", false},
		{"blank params are defaults", "/api/anime/tv?sort=&country=&keyword=&season=", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(http.MethodGet, tc.path, nil))
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode, tc.path)
		})
	}
}

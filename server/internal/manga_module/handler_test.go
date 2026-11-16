package manga_module

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// get performs a GET and returns the status and the decoded JSON body.
func get(t *testing.T, app *fiber.App, path string) (int, map[string]any) {
	t.Helper()
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var body map[string]any
	require.NoError(t, json.Unmarshal(raw, &body), string(raw))
	return resp.StatusCode, body
}

func errorCode(body map[string]any) string {
	code, _ := body["code"].(string)
	return code
}

func TestBadParametersAre400AndNeverReachAniList(t *testing.T) {
	up := newFakeAniList(t, standardUpstream)
	h := New(up.client(), matrixGate())
	app := viewerApp(h, viewerNoDOB)

	cases := []string{
		"/manga/search",          // q missing
		"/manga/search?q=%20%20", // q blank after trim
		"/manga/search?q=" + strings.Repeat("a", maxSearchRunes+1),
		"/manga/search?q=dune&page=0",
		"/manga/search?q=dune&page=abc",
		"/manga/search?q=dune&page=501",
		"/manga/search?q=dune&include_adult=yes",
		"/manga?country=us",
		"/manga?country=jp", // case-sensitive: JP only
		"/manga?format=comic",
		"/manga?sort=bogus",
		"/manga?status=paused",
		"/manga?genre=Ecchi",  // not in the allowlist
		"/manga?genre=sci-fi", // exact AniList spelling only
		"/manga?page=0",
		"/manga?page=-1",
		"/manga?include_adult=1",
	}
	for _, path := range cases {
		t.Run(path, func(t *testing.T) {
			status, body := get(t, app, path)
			assert.Equal(t, http.StatusBadRequest, status)
			assert.Equal(t, "INVALID_SEARCH", errorCode(body))
		})
	}
	assert.Zero(t, up.count(), "invalid parameters must not spend upstream budget")
}

func TestInvalidIDIs400(t *testing.T) {
	up := newFakeAniList(t, standardUpstream)
	app := viewerApp(New(up.client(), matrixGate()), viewerGuest)

	for _, path := range []string{"/manga/0", "/manga/0/similar"} {
		status, body := get(t, app, path)
		assert.Equal(t, http.StatusBadRequest, status, path)
		assert.Equal(t, "INVALID_ID", errorCode(body), path)
	}
	assert.Zero(t, up.count())
}

func TestMissingTitleIs404(t *testing.T) {
	up := newFakeAniList(t, func(upstreamCall) reply { return ok(`{"data":{"Media":null}}`) })
	app := viewerApp(New(up.client(), matrixGate()), viewerGuest)

	status, body := get(t, app, "/manga/77")
	assert.Equal(t, http.StatusNotFound, status)
	assert.Equal(t, "NOT_FOUND", errorCode(body))
}

func TestAniListNotFoundStatusIs404(t *testing.T) {
	up := newFakeAniList(t, func(upstreamCall) reply { return reply{status: http.StatusNotFound, body: `{}`} })
	app := viewerApp(New(up.client(), matrixGate()), viewerGuest)

	status, body := get(t, app, "/manga/77")
	assert.Equal(t, http.StatusNotFound, status)
	assert.Equal(t, "NOT_FOUND", errorCode(body))
}

func TestRateLimitIs503AndCoolsDownWithoutHTTP(t *testing.T) {
	up := newFakeAniList(t, func(upstreamCall) reply {
		return reply{status: http.StatusTooManyRequests, body: `{}`, retryAfter: "120"}
	})
	app := viewerApp(New(up.client(), matrixGate()), viewerGuest)

	for range 2 {
		status, body := get(t, app, "/manga/search?q=dune")
		assert.Equal(t, http.StatusServiceUnavailable, status)
		assert.Equal(t, "UPSTREAM_UNAVAILABLE", errorCode(body))
	}
	assert.Equal(t, 1, up.count(), "second request must be refused by the cooldown")
}

func TestSecondIdenticalListIsServedFromCache(t *testing.T) {
	up := newFakeAniList(t, standardUpstream)
	app := viewerApp(New(up.client(), matrixGate()), viewerGuest)

	for range 2 {
		status, _ := get(t, app, "/manga/search?q=dune&page=2")
		require.Equal(t, http.StatusOK, status)
	}
	assert.Equal(t, 1, up.count())
	assert.EqualValues(t, 2, up.last().Variables["page"])
}

func TestGenresNeedNoUpstream(t *testing.T) {
	up := newFakeAniList(t, standardUpstream)
	app := viewerApp(New(up.client(), matrixGate()), viewerGuest)

	status, body := get(t, app, "/manga/genres")
	assert.Equal(t, http.StatusOK, status)
	genres, ok := body["genres"].([]any)
	require.True(t, ok)
	assert.Contains(t, genres, "Sci-Fi")
	assert.NotContains(t, genres, "Ecchi")
	assert.Zero(t, up.count())
}

func TestDetailCarriesCleanedFields(t *testing.T) {
	up := newFakeAniList(t, standardUpstream)
	app := viewerApp(New(up.client(), matrixGate()), viewerGuest)

	status, body := get(t, app, "/manga/1")
	require.Equal(t, http.StatusOK, status)
	manga := body["manga"].(map[string]any)

	assert.Equal(t, "Hi\nthere & you", manga["description"])
	assert.Equal(t, "manhwa", manga["kind"])
	assert.Equal(t, "มังงะ 1", manga["display_title"])
	assert.Equal(t, "https://img.example/1.jpg", manga["cover_url"])
	assert.Equal(t, "", manga["banner_url"], "non-https banner must be dropped")
	assert.Equal(t, false, manga["is_adult"])

	tags := manga["tags"].([]any)
	require.Len(t, tags, 1, "spoiler and low-rank tags are dropped")
	assert.Equal(t, "Isekai", tags[0].(map[string]any)["name"])

	links := manga["read_links"].([]any)
	require.Len(t, links, 1, "only https STREAMING/INFO links survive")
	assert.Equal(t, "https://ex.example/read", links[0].(map[string]any)["url"])

	relations := manga["relations"].([]any)
	require.Len(t, relations, 1, "only ADAPTATION to ANIME survives")
	assert.Equal(t, "E", relations[0].(map[string]any)["title"])
}

func TestAdultMatrixForListSearchDetailAndSimilar(t *testing.T) {
	cases := []struct {
		name     string
		viewer   uint
		eligible bool
	}{
		{"guest", viewerGuest, false},
		{"no date of birth", viewerNoDOB, false},
		{"under 18", viewerUnder18, false},
		{"18 and over", viewerAdult, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			up := newFakeAniList(t, standardUpstream)
			app := viewerApp(New(up.client(), matrixGate()), tc.viewer)

			// Browse and search: adult items appear only when eligible and requested.
			for _, path := range []string{"/manga?include_adult=true", "/manga/search?q=dune&include_adult=true"} {
				status, body := get(t, app, path)
				require.Equal(t, http.StatusOK, status, path)
				assertAdultFilter(t, up.last().Variables, tc.eligible, path)
				for _, item := range body["results"].([]any) {
					isAdult := item.(map[string]any)["is_adult"].(bool)
					assert.False(t, isAdult && !tc.eligible, "adult item leaked to %s on %s", tc.name, path)
				}
			}

			// Single adult title: 403 unless eligible.
			status, body := get(t, app, "/manga/2")
			if tc.eligible {
				assert.Equal(t, http.StatusOK, status)
			} else {
				assert.Equal(t, http.StatusForbidden, status)
				assert.Equal(t, "AGE_RESTRICTED", errorCode(body))
			}

			// Recommendations: adult entries need include_adult and an eligible viewer; anime is never listed.
			for path, want := range map[string][]float64{
				"/manga/1/similar?include_adult=true": {3, 4},
				"/manga/1/similar":                    {3},
			} {
				status, body = get(t, app, path)
				require.Equal(t, http.StatusOK, status, path)
				ids := []float64{}
				for _, item := range body["results"].([]any) {
					ids = append(ids, item.(map[string]any)["id"].(float64))
				}
				if tc.eligible {
					assert.Equal(t, want, ids, path)
				} else {
					assert.Equal(t, []float64{3}, ids, path)
				}
			}
		})
	}
}

// assertAdultFilter checks the isAdult variable: an eligible viewer sends none (no filter), everyone
// else sends isAdult:false.
func assertAdultFilter(t *testing.T, vars map[string]any, eligible bool, path string) {
	t.Helper()
	if eligible {
		assert.NotContains(t, vars, "isAdult", path)
		return
	}
	assert.Equal(t, false, vars["isAdult"], path)
}

func TestAdultCacheEntriesAreSeparated(t *testing.T) {
	up := newFakeAniList(t, standardUpstream)
	h := New(up.client(), matrixGate())
	guest := viewerApp(h, viewerGuest)
	adult := viewerApp(h, viewerAdult)

	get(t, guest, "/manga?include_adult=true")
	get(t, adult, "/manga?include_adult=true")
	assert.Equal(t, 2, up.count(), "eligible request must not reuse the guest entry")
	assertAdultFilter(t, up.last().Variables, true, "eligible browse")

	status, body := get(t, guest, "/manga?include_adult=true")
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, 2, up.count(), "guest repeat is a cache hit on the non-adult entry")
	for _, item := range body["results"].([]any) {
		assert.False(t, item.(map[string]any)["is_adult"].(bool), "guest response must not contain adult items")
	}
}

func TestIncludeAdultIsIgnoredForIneligibleViewers(t *testing.T) {
	up := newFakeAniList(t, standardUpstream)
	app := viewerApp(New(up.client(), matrixGate()), viewerUnder18)

	status, _ := get(t, app, "/manga?include_adult=true")
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, false, up.last().Variables["isAdult"])
}

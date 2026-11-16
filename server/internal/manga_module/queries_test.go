package manga_module

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEveryQueryIsScopedToManga(t *testing.T) {
	for name, q := range map[string]string{"list": listQuery, "detail": detailQuery, "similar": similarQuery} {
		assert.Contains(t, q, "type: MANGA", name)
		assert.NotContains(t, q, "isAdult: false", name)
	}
	assert.Contains(t, listQuery, "isAdult: $isAdult", "browse and search read adult eligibility from a variable")
	// A Media(id) lookup with an isAdult filter returns 404 for titles that do not match it,
	// so detail and similar fetch by id alone and gate on the returned isAdult field.
	assert.NotContains(t, detailQuery, "isAdult:", "detail must not filter by isAdult")
	assert.NotContains(t, similarQuery, "isAdult:", "similar must not filter by isAdult")
}

func TestBrowseVariablesPerFilter(t *testing.T) {
	up := newFakeAniList(t, standardUpstream)
	app := viewerApp(New(up.client(), matrixGate()), viewerGuest)

	status, _ := get(t, app, "/manga")
	require.Equal(t, http.StatusOK, status)
	def := up.last().Variables
	assert.NotContains(t, def, "country")
	assert.NotContains(t, def, "genre")
	assert.NotContains(t, def, "status")
	assert.Equal(t, []any{"POPULARITY_DESC"}, def["sort"])
	assert.Equal(t, []any{"MANGA", "ONE_SHOT", "NOVEL"}, def["format_in"])
	assert.Equal(t, false, def["isAdult"])

	status, _ = get(t, app, "/manga?country=KR&genre=Action&status=finished&sort=newest&format=novel&page=2")
	require.Equal(t, http.StatusOK, status)
	vars := up.last().Variables
	assert.Equal(t, "KR", vars["country"])
	assert.Equal(t, "Action", vars["genre"])
	assert.Equal(t, "FINISHED", vars["status"])
	assert.Equal(t, []any{"START_DATE_DESC"}, vars["sort"])
	assert.Equal(t, []any{"NOVEL"}, vars["format_in"])
	assert.EqualValues(t, 2, vars["page"])
	assert.EqualValues(t, mangaPageSize, vars["perPage"])
}

func TestCountryAndSortMappings(t *testing.T) {
	up := newFakeAniList(t, standardUpstream)
	app := viewerApp(New(up.client(), matrixGate()), viewerGuest)

	cases := map[string][2]any{
		"/manga?country=CN":       {"country", "CN"},
		"/manga?country=JP":       {"country", "JP"},
		"/manga?sort=top_rated":   {"sort", []any{"SCORE_DESC"}},
		"/manga?sort=trending":    {"sort", []any{"TRENDING_DESC"}},
		"/manga?status=releasing": {"status", "RELEASING"},
		"/manga?format=one_shot":  {"format_in", []any{"ONE_SHOT"}},
		"/manga?genre=Sci-Fi":     {"genre", "Sci-Fi"},
	}
	for path, want := range cases {
		status, _ := get(t, app, path)
		require.Equal(t, http.StatusOK, status, path)
		assert.Equal(t, want[1], up.last().Variables[want[0].(string)], path)
	}
}

func TestSearchVariablesCarryQueryAndAdultFlag(t *testing.T) {
	up := newFakeAniList(t, standardUpstream)
	app := viewerApp(New(up.client(), matrixGate()), viewerAdult)

	status, _ := get(t, app, "/manga/search?q=%20Dune%20&page=3&include_adult=true")
	require.Equal(t, http.StatusOK, status)
	vars := up.last().Variables
	assert.Equal(t, "Dune", vars["search"])
	assert.EqualValues(t, 3, vars["page"])
	assert.Equal(t, []any{"SEARCH_MATCH"}, vars["sort"])
	assert.NotContains(t, vars, "isAdult", "eligible viewer including adult titles sends no isAdult filter")
	assert.NotContains(t, vars, "country")
}

func TestSimilarAndDetailFetchByIDThenGate(t *testing.T) {
	up := newFakeAniList(t, standardUpstream)
	app := viewerApp(New(up.client(), matrixGate()), viewerGuest)

	status, _ := get(t, app, "/manga/2/similar")
	assert.Equal(t, http.StatusForbidden, status)
	assert.NotContains(t, strings.Join(queriesSent(up), "\n"), "isAdult: true")
}

func queriesSent(f *fakeAniList) []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, 0, len(f.calls))
	for _, c := range f.calls {
		out = append(out, c.Query)
	}
	return out
}

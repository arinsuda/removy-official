package tmdbmodule

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// stubSearch swaps the package client for one that answers every request with body
// and records the query of the last request.
func stubSearch(t *testing.T, body string) *url.Values {
	t.Helper()
	var seen url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	prev := c
	t.Cleanup(func() { c = prev; srv.Close() })
	c = &client{http: srv.Client(), baseURL: srv.URL, apiKey: "k", language: "en-US"}
	return &seen
}

func TestSearchMoviesYearIsSentGenreFiltersPage(t *testing.T) {
	seen := stubSearch(t, `{"page":2,"total_pages":7,"total_results":130,"results":[
		{"id":7001,"title":"A","genre_ids":[28,12]},{"id":7002,"title":"B","genre_ids":[35]},{"id":7003,"title":"C","genre_ids":[28]}]}`)

	got, err := SearchMovies("dune", 2, "en-US", SearchFilters{Year: 2021, GenreID: 28})
	if err != nil {
		t.Fatal(err)
	}
	if (*seen).Get("primary_release_year") != "2021" || (*seen).Get("query") != "dune" || (*seen).Get("page") != "2" {
		t.Fatalf("params: %v", *seen)
	}
	if (*seen).Has("with_genres") {
		t.Fatalf("genre must not be sent to /search: %v", *seen)
	}
	if len(got.Results) != 2 || got.Results[0].ID != 7001 || got.Results[1].ID != 7003 {
		t.Fatalf("filtered: %+v", got.Results)
	}
	if got.Page != 2 || got.TotalPages != 7 {
		t.Fatalf("paging must follow TMDB: %+v", got)
	}
}

func TestSearchSeriesUsesFirstAirYearAndNoFilterKeepsAll(t *testing.T) {
	seen := stubSearch(t, `{"page":1,"total_pages":1,"total_results":2,"results":[
		{"id":8001,"name":"A","genre_ids":[18]},{"id":8002,"name":"B","genre_ids":[]}]}`)

	got, err := SearchSeries("show", 1, "", SearchFilters{Year: 2019})
	if err != nil {
		t.Fatal(err)
	}
	if (*seen).Get("first_air_date_year") != "2019" || (*seen).Has("primary_release_year") {
		t.Fatalf("params: %v", *seen)
	}
	if len(got.Results) != 2 {
		t.Fatalf("no genre filter must keep all: %+v", got.Results)
	}

	got, err = SearchSeries("show", 1, "", SearchFilters{})
	if err != nil || (*seen).Has("first_air_date_year") || len(got.Results) != 2 {
		t.Fatalf("empty filters: %v %v %+v", err, *seen, got)
	}
	got, _ = SearchSeries("show", 1, "", SearchFilters{GenreID: 99})
	if got.Results == nil || len(got.Results) != 0 {
		t.Fatalf("genre with no match must be an empty list, not null: %+v", got.Results)
	}
}

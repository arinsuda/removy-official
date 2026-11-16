package tmdbmodule

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

var animeToday = time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC)

func assertParams(t *testing.T, name string, got map[string][]string, want map[string]string) {
	t.Helper()
	for k, v := range want {
		vals, ok := got[k]
		if !ok || len(vals) == 0 {
			t.Fatalf("%s: missing %s (want %q)", name, k, v)
		}
		if vals[0] != v {
			t.Fatalf("%s: %s=%q, want %q", name, k, vals[0], v)
		}
	}
}

func TestAnimeParamsSortsPerType(t *testing.T) {
	cases := []struct {
		name   string
		series bool
		f      AnimeFilter
		want   map[string]string
		absent []string
	}{
		{
			name: "popular tv", series: true, f: AnimeFilter{Sort: AnimeSortPopular},
			want:   map[string]string{"sort_by": "popularity.desc", "vote_count.gte": "20", "with_genres": "16", "include_adult": "false", "with_origin_country": "JP"},
			absent: []string{"first_air_date.lte", "air_date.gte"},
		},
		{
			name: "top rated tv", series: true, f: AnimeFilter{Sort: AnimeSortTopRated},
			want: map[string]string{"sort_by": "vote_average.desc", "vote_count.gte": "200"},
		},
		{
			name: "top rated movie", series: false, f: AnimeFilter{Sort: AnimeSortTopRated},
			want: map[string]string{"sort_by": "vote_average.desc", "vote_count.gte": "300"},
		},
		{
			name: "airing tv window", series: true, f: AnimeFilter{Sort: AnimeSortAiring},
			want:   map[string]string{"sort_by": "popularity.desc", "air_date.gte": "2026-11-16", "air_date.lte": "2026-11-23"},
			absent: []string{"vote_count.gte"},
		},
		{
			name: "newest tv capped at today", series: true, f: AnimeFilter{Sort: AnimeSortNewest},
			want: map[string]string{"sort_by": "first_air_date.desc", "first_air_date.lte": "2026-11-16", "vote_count.gte": "5"},
		},
		{
			name: "newest movie capped at today", series: false, f: AnimeFilter{Sort: AnimeSortNewest},
			want: map[string]string{"sort_by": "primary_release_date.desc", "primary_release_date.lte": "2026-11-16", "vote_count.gte": "5"},
		},
		{
			name: "airing on movie falls back to popularity", series: false, f: AnimeFilter{Sort: AnimeSortAiring},
			want:   map[string]string{"sort_by": "popularity.desc", "vote_count.gte": "20"},
			absent: []string{"air_date.gte"},
		},
		{
			name: "empty sort defaults to popular", series: false, f: AnimeFilter{},
			want: map[string]string{"sort_by": "popularity.desc"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := AnimeParams(tc.series, tc.f, 2, animeToday)
			if p.Get("page") != "2" {
				t.Fatalf("page: %v", p)
			}
			assertParams(t, tc.name, p, tc.want)
			for _, k := range tc.absent {
				if p.Has(k) {
					t.Fatalf("%s: %s should be unset: %v", tc.name, k, p)
				}
			}
		})
	}
}

func TestAnimeParamsGenreIsAndNotOr(t *testing.T) {
	p := AnimeParams(true, AnimeFilter{GenreID: 10759}, 1, animeToday)
	if got := p.Get("with_genres"); got != "16,10759" {
		t.Fatalf("with_genres=%q, want 16,10759 (AND, never |)", got)
	}
}

func TestAnimeParamsKeywordAndCountry(t *testing.T) {
	p := AnimeParams(false, AnimeFilter{Keyword: "isekai", Country: AnimeCountryAll}, 1, animeToday)
	if p.Get("with_keywords") != "237451" {
		t.Fatalf("with_keywords=%q", p.Get("with_keywords"))
	}
	if p.Get("with_origin_country") != "JP|CN|KR|TW" {
		t.Fatalf("with_origin_country=%q", p.Get("with_origin_country"))
	}
	p = AnimeParams(false, AnimeFilter{Country: "KR"}, 1, animeToday)
	if p.Get("with_origin_country") != "KR" || p.Has("with_keywords") {
		t.Fatalf("country KR: %v", p)
	}
	if AnimeParams(true, AnimeFilter{}, 1, animeToday).Get("with_origin_country") != "JP" {
		t.Fatal("empty country must default to JP")
	}
}

func TestAnimeParamsDateWindows(t *testing.T) {
	cases := []struct {
		name   string
		series bool
		f      AnimeFilter
		want   map[string]string
	}{
		{"decade tv", true, AnimeFilter{Decade: 1990}, map[string]string{"first_air_date.gte": "1990-01-01", "first_air_date.lte": "1999-12-31", "vote_count.gte": "5"}},
		{"decade movie popular", false, AnimeFilter{Decade: 2020}, map[string]string{"primary_release_date.gte": "2020-01-01", "primary_release_date.lte": "2029-12-31", "vote_count.gte": "5"}},
		{"year movie", false, AnimeFilter{Year: 2024}, map[string]string{"primary_release_date.gte": "2024-01-01", "primary_release_date.lte": "2024-12-31"}},
		{"winter tv", true, AnimeFilter{Year: 2025, Season: "winter"}, map[string]string{"first_air_date.gte": "2025-01-01", "first_air_date.lte": "2025-03-31"}},
		{"spring tv", true, AnimeFilter{Year: 2025, Season: "spring"}, map[string]string{"first_air_date.gte": "2025-04-01", "first_air_date.lte": "2025-06-30"}},
		{"summer tv", true, AnimeFilter{Year: 2025, Season: "summer"}, map[string]string{"first_air_date.gte": "2025-07-01", "first_air_date.lte": "2025-09-30"}},
		{"fall tv", true, AnimeFilter{Year: 2025, Season: "fall"}, map[string]string{"first_air_date.gte": "2025-10-01", "first_air_date.lte": "2025-12-31"}},
		{"newest with past year keeps the year end", true, AnimeFilter{Sort: AnimeSortNewest, Year: 2024}, map[string]string{"first_air_date.lte": "2024-12-31", "vote_count.gte": "5"}},
		{"popular scoped floor is lower", true, AnimeFilter{Year: 2024}, map[string]string{"vote_count.gte": "5"}},
		{"top rated scoped floor", true, AnimeFilter{Sort: AnimeSortTopRated, Decade: 1980}, map[string]string{"vote_count.gte": "50"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := AnimeParams(tc.series, tc.f, 1, animeToday)
			assertParams(t, tc.name, p, tc.want)
		})
	}
}

func TestAnimeParamsAiringFloorlessForTV(t *testing.T) {
	p := AnimeParams(true, AnimeFilter{Sort: AnimeSortAiring}, 1, animeToday)
	if p.Has("vote_count.gte") {
		t.Fatalf("airing TV must not set a vote floor: %v", p)
	}
}

func TestAnimeParamsSeasonEndsOnRealLastDay(t *testing.T) {
	// Fall ends on Dec 31; December has 31 days, so day 0 of January is the last day.
	p := AnimeParams(true, AnimeFilter{Year: 2028, Season: "fall"}, 1, animeToday)
	if p.Get("first_air_date.lte") != "2028-12-31" {
		t.Fatalf("fall lte=%q", p.Get("first_air_date.lte"))
	}
	// Spring in a leap year ends on June 30, not an earlier day.
	p = AnimeParams(true, AnimeFilter{Year: 2028, Season: "spring"}, 1, animeToday)
	if p.Get("first_air_date.lte") != "2028-06-30" {
		t.Fatalf("spring lte=%q", p.Get("first_air_date.lte"))
	}
}

func TestAnimeKeywordAllowlist(t *testing.T) {
	if id, ok := AnimeKeywordID("mecha"); !ok || id != 10046 {
		t.Fatalf("mecha: %d %v", id, ok)
	}
	if _, ok := AnimeKeywordID("harem"); ok {
		t.Fatal("harem must not be allowed")
	}
}

func TestDiscoverAnimeCachesRawPage(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"page":1,"total_pages":1,"total_results":1,"results":[{"id":910001,"name":"Cached"}]}`))
	}))
	defer srv.Close()

	prev := c
	defer func() { c = prev }()
	c = &client{http: srv.Client(), baseURL: srv.URL, apiKey: "k", language: "en-US"}

	// Unique filter: no other test fetches this combination, so the cache starts cold.
	f := AnimeFilter{Sort: AnimeSortTopRated, Keyword: "sports", Country: "CN", Year: 1961, Season: "spring"}
	for i := 0; i < 2; i++ {
		got, err := DiscoverAnimeSeries(f, 1)
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Results) != 1 || got.Results[0].ID != 910001 {
			t.Fatalf("results: %+v", got.Results)
		}
	}
	if n := hits.Load(); n != 1 {
		t.Fatalf("TMDB hits=%d, want 1 (second call must be served from cache)", n)
	}

	// A different page is a different canonical request.
	if _, err := DiscoverAnimeSeries(f, 2); err != nil {
		t.Fatal(err)
	}
	if n := hits.Load(); n != 2 {
		t.Fatalf("TMDB hits=%d, want 2 after paging", n)
	}
}

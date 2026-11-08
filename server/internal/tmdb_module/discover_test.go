package tmdbmodule

import (
	"testing"
	"time"
)

func TestDiscoverParamsKeepTabOrdering(t *testing.T) {
	today := time.Date(2026, 11, 8, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		series bool
		tab    string
		want   map[string]string
	}{
		{false, TabPopular, map[string]string{"sort_by": "popularity.desc", "with_genres": "28|35"}},
		{false, TabTopRated, map[string]string{"sort_by": "vote_average.desc", "vote_count.gte": "300"}},
		{false, TabNow, map[string]string{"sort_by": "popularity.desc", "release_date.gte": "2026-09-24", "release_date.lte": "2026-11-08", "region": "TH"}},
		{true, TabTopRated, map[string]string{"sort_by": "vote_average.desc", "vote_count.gte": "200"}},
		{true, TabNow, map[string]string{"sort_by": "popularity.desc", "air_date.gte": "2026-11-08", "air_date.lte": "2026-11-15"}},
		{false, TabUpcoming, map[string]string{"sort_by": "primary_release_date.asc", "release_date.gte": "2026-11-08", "release_date.lte": "2027-03-08", "region": "TH", "with_release_type": "2|3"}},
		{true, "unknown", map[string]string{"sort_by": "popularity.desc"}},
	}
	for _, tc := range cases {
		p := DiscoverParams(tc.series, tc.tab, "28,35", 2, today)
		if p.Get("page") != "2" {
			t.Fatalf("page: %v", p)
		}
		for k, v := range tc.want {
			if k == "with_genres" && p.Get(k) != v {
				t.Fatalf("%s/%v genres %q", tc.tab, tc.series, p.Get(k))
			} else if k != "with_genres" && p.Get(k) != v {
				t.Fatalf("%s/%v %s=%q want %q", tc.tab, tc.series, k, p.Get(k), v)
			}
		}
	}
}

func TestDiscoverParamsUpcomingGenreOptional(t *testing.T) {
	today := time.Date(2026, 11, 8, 0, 0, 0, 0, time.UTC)
	if p := DiscoverParams(false, TabUpcoming, "", 1, today); p.Has("with_genres") {
		t.Fatalf("with_genres should be unset: %v", p)
	}
	if p := DiscoverParams(false, TabUpcoming, "28,35", 1, today); p.Get("with_genres") != "28|35" {
		t.Fatalf("with_genres: %v", p)
	}
}

func TestBangkokTodayRollsOverAtLocalMidnight(t *testing.T) {
	// 18:00 UTC on Nov 8 is already Nov 9 in Bangkok (UTC+7).
	got := bangkokToday(time.Date(2026, 11, 8, 18, 0, 0, 0, time.UTC))
	if got.Format("2006-01-02") != "2026-11-09" {
		t.Fatalf("got %s", got)
	}
}

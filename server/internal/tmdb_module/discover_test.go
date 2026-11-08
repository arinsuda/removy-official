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

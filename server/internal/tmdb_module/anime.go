package tmdbmodule

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Anime browse: TMDB discover filtered to animation, with a keyword and an origin country.
const (
	AnimeSortPopular  = "popular"
	AnimeSortTopRated = "top_rated"
	AnimeSortAiring   = "airing"
	AnimeSortNewest   = "newest"
	AnimeDefaultSort  = AnimeSortPopular

	AnimeGenreAnimation  = 16
	AnimeCountryAll      = "all"
	AnimeDefaultCountry  = "JP"
	animeAllCountryCodes = "JP|CN|KR|TW"

	animeCacheTTL = 30 * time.Minute
	animeCacheMax = 500

	// Minimum vote counts. The scoped values apply when a decade, year or season narrows
	// the window, because a single year has far fewer votes than the whole catalogue.
	animeVotesPopular        = 20
	animeVotesPopularScoped  = 5
	animeVotesTopRatedTV     = 200
	animeVotesTopRatedMovie  = 300
	animeVotesTopRatedScoped = 50
	animeVotesNewest         = 5
	animeAiringWindowDays    = 7
	animeDecadeYears         = 10
	animeFirstDayOfMonth     = 1
)

// animeKeywordIDs are TMDB keyword ids (verified). The client sends the name.
var animeKeywordIDs = map[string]int{
	"isekai":        237451,
	"mecha":         10046,
	"shoujo":        206437,
	"shounen":       207826,
	"seinen":        195668,
	"josei":         229074,
	"slice_of_life": 9914,
	"sports":        6075,
	"idol":          287664,
	"school":        190116,
}

// animeCountryCodes maps the accepted country parameter to TMDB with_origin_country.
var animeCountryCodes = map[string]string{
	"JP":            "JP",
	"CN":            "CN",
	"KR":            "KR",
	AnimeCountryAll: animeAllCountryCodes,
}

// animeSeasons are calendar quarters by first air or release date.
var animeSeasons = map[string][2]time.Month{
	"winter": {time.January, time.March},
	"spring": {time.April, time.June},
	"summer": {time.July, time.September},
	"fall":   {time.October, time.December},
}

// animeCache holds raw TMDB anime pages, keyed by the canonical request.
var animeCache = newDetailCache(animeCacheTTL, animeCacheMax)

// AnimeFilter narrows anime discover. Zero values mean "no filter".
type AnimeFilter struct {
	Sort    string // AnimeSort*; "" = popular
	GenreID int    // extra genre ANDed with animation; 0 = none
	Keyword string // key of animeKeywordIDs; "" = none
	Country string // JP, CN, KR or AnimeCountryAll; "" = JP
	Decade  int    // first year of a decade (1960, 1970, ...); 0 = none
	Year    int    // release year; 0 = none. Required with Season
	Season  string // winter, spring, summer or fall (TV only); "" = none
}

// IsAnimeSort reports whether s is a sort the anime discover accepts.
func IsAnimeSort(s string) bool {
	switch s {
	case AnimeSortPopular, AnimeSortTopRated, AnimeSortAiring, AnimeSortNewest:
		return true
	}
	return false
}

// IsAnimeCountry reports whether s is an accepted country parameter.
func IsAnimeCountry(s string) bool {
	_, ok := animeCountryCodes[s]
	return ok
}

// IsAnimeSeason reports whether s is a season name.
func IsAnimeSeason(s string) bool {
	_, ok := animeSeasons[s]
	return ok
}

// AnimeKeywordID returns the TMDB keyword id for an allowed keyword name.
func AnimeKeywordID(name string) (int, bool) {
	id, ok := animeKeywordIDs[name]
	return id, ok
}

// animeDateWindow returns the inclusive release window as YYYY-MM-DD ("" when unset).
func animeDateWindow(f AnimeFilter) (from, to string) {
	day := func(t time.Time) string { return t.Format("2006-01-02") }
	switch {
	case f.Season != "" && f.Year > 0:
		months := animeSeasons[f.Season]
		from = day(time.Date(f.Year, months[0], animeFirstDayOfMonth, 0, 0, 0, 0, time.UTC))
		// Day 0 of the month after the season's last month is that month's last day.
		to = day(time.Date(f.Year, months[1]+1, 0, 0, 0, 0, 0, time.UTC))
	case f.Decade > 0:
		from = day(time.Date(f.Decade, time.January, 1, 0, 0, 0, 0, time.UTC))
		to = day(time.Date(f.Decade+animeDecadeYears, time.January, 0, 0, 0, 0, 0, time.UTC))
	case f.Year > 0:
		from = day(time.Date(f.Year, time.January, 1, 0, 0, 0, 0, time.UTC))
		to = day(time.Date(f.Year+1, time.January, 0, 0, 0, 0, 0, time.UTC))
	}
	return from, to
}

// animeVoteFloor is the minimum vote count for a sort. Newest always uses its small floor.
func animeVoteFloor(series bool, sort string, scoped bool) int {
	switch sort {
	case AnimeSortTopRated:
		if scoped {
			return animeVotesTopRatedScoped
		}
		if series {
			return animeVotesTopRatedTV
		}
		return animeVotesTopRatedMovie
	case AnimeSortNewest:
		return animeVotesNewest
	default:
		if scoped {
			return animeVotesPopularScoped
		}
		return animeVotesPopular
	}
}

// AnimeParams builds /discover/{movie,tv} params for anime. The genre list is joined with ","
// (AND); DiscoverParams joins with "|" (OR), which would widen this filter.
func AnimeParams(series bool, f AnimeFilter, page int, today time.Time) url.Values {
	params := pageParams(page)
	withGenres := strconv.Itoa(AnimeGenreAnimation)
	if f.GenreID > 0 {
		withGenres += "," + strconv.Itoa(f.GenreID)
	}
	params.Set("with_genres", withGenres)
	params.Set("include_adult", "false")

	country, ok := animeCountryCodes[f.Country]
	if !ok {
		country = animeCountryCodes[AnimeDefaultCountry]
	}
	params.Set("with_origin_country", country)

	if id, ok := AnimeKeywordID(f.Keyword); ok {
		params.Set("with_keywords", strconv.Itoa(id))
	}

	dateKey := "primary_release_date"
	if series {
		dateKey = "first_air_date"
	}
	from, upper := animeDateWindow(f)
	scoped := from != ""
	if from != "" {
		params.Set(dateKey+".gte", from)
	}

	sort := f.Sort
	if sort == "" {
		sort = AnimeDefaultSort
	}
	day := today.Format("2006-01-02")

	switch sort {
	case AnimeSortTopRated:
		params.Set("sort_by", "vote_average.desc")
	case AnimeSortAiring:
		// Same window as the TV "now" tab. Airing is TV only; movies fall back to popularity.
		params.Set("sort_by", "popularity.desc")
		if series {
			params.Set("air_date.gte", day)
			params.Set("air_date.lte", today.AddDate(0, 0, animeAiringWindowDays).Format("2006-01-02"))
		}
	case AnimeSortNewest:
		params.Set("sort_by", dateKey+".desc")
		// Never list titles that have not started yet.
		if upper == "" || day < upper {
			upper = day
		}
	default:
		params.Set("sort_by", "popularity.desc")
	}

	if sort != AnimeSortAiring || !series {
		params.Set("vote_count.gte", strconv.Itoa(animeVoteFloor(series, sort, scoped)))
	}
	if upper != "" {
		params.Set(dateKey+".lte", upper)
	}
	return params
}

// DiscoverAnimeSeries lists anime TV, with popular, top rated, airing or newest order.
func DiscoverAnimeSeries(f AnimeFilter, page int) (*PaginatedResult[TVSeries], error) {
	return discoverAnime[TVSeries]("/discover/tv", true, f, page)
}

// DiscoverAnimeMovies lists anime movies with popular, top rated or newest order.
func DiscoverAnimeMovies(f AnimeFilter, page int) (*PaginatedResult[Movie], error) {
	return discoverAnime[Movie]("/discover/movie", false, f, page)
}

// discoverAnime fetches one anime page, serving repeats of the same canonical request from cache.
func discoverAnime[T any](path string, series bool, f AnimeFilter, page int) (*PaginatedResult[T], error) {
	params := AnimeParams(series, f, page, bangkokToday(time.Now()))
	key := detailKey{kind: path + "?" + params.Encode(), language: currentLanguage()}
	if v, ok := animeCache.get(key); ok {
		cached := v.(PaginatedResult[T])
		return &cached, nil
	}

	var result PaginatedResult[T]
	if err := get(path, params, &result); err != nil {
		return nil, fmt.Errorf("discoverAnime %s: %w", path, err)
	}
	animeCache.set(key, result)
	return &result, nil
}

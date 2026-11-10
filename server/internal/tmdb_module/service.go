package tmdbmodule

import (
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

const defaultRegion = "TH"

func normalizePage(page int) int {
	if page < 1 {
		return 1
	}

	if page > 500 {
		return 500
	}

	return page
}

func pageParams(page int) url.Values {
	params := url.Values{}
	params.Set("page", strconv.Itoa(normalizePage(page)))
	return params
}

func regionalPageParams(page int) url.Values {
	params := pageParams(page)
	params.Set("region", defaultRegion)
	return params
}

type RequestOptions struct {
	Language string
	Region   string
}

func (o RequestOptions) apply(params url.Values) url.Values {
	if params == nil {
		params = url.Values{}
	}

	if o.Language != "" {
		params.Set("language", o.Language)
	}

	region := o.Region
	if region == "" {
		region = defaultRegion
	}
	params.Set("region", region)

	return params
}

func listParams(page int, options RequestOptions) url.Values {
	params := url.Values{}
	params.Set("page", strconv.Itoa(normalizePage(page)))

	return options.apply(params)
}

func GetPopular(page int) (*PaginatedResult[Movie], error) {
	var result PaginatedResult[Movie]
	if err := get("/movie/popular", pageParams(page), &result); err != nil {
		return nil, fmt.Errorf("GetPopular: %w", err)
	}
	return &result, nil
}

func GetNowPlaying(
	page int,
	options RequestOptions,
) (*PaginatedResult[Movie], error) {
	var result PaginatedResult[Movie]

	params := listParams(page, options)

	if err := get(
		"/movie/now_playing",
		params,
		&result,
	); err != nil {
		return nil, fmt.Errorf("GetNowPlaying: %w", err)
	}

	return &result, nil
}

func GetTopRated(page int) (*PaginatedResult[Movie], error) {
	var result PaginatedResult[Movie]
	if err := get("/movie/top_rated", pageParams(page), &result); err != nil {
		return nil, fmt.Errorf("GetTopRated: %w", err)
	}
	return &result, nil
}

func GetUpcoming(page int) (*PaginatedResult[Movie], error) {
	var result PaginatedResult[Movie]

	if err := get(
		"/movie/upcoming",
		regionalPageParams(page),
		&result,
	); err != nil {
		return nil, fmt.Errorf("GetUpcoming: %w", err)
	}

	return &result, nil
}

func DiscoverUpcomingByYear(year, page int) (*PaginatedResult[Movie], error) {
	params := pageParams(page)

	params.Set("primary_release_year",
		fmt.Sprintf("%d", year))

	params.Set("sort_by", "popularity.desc")

	params.Set("include_adult", "false")
	params.Set("include_video", "false")

	var result PaginatedResult[Movie]

	err := get("/discover/movie", params, &result)

	return &result, err
}

func GetMovieByID(tmdbID int) (*MovieDetail, error) {
	key := detailKey{detailKindMovie, tmdbID, currentLanguage()}
	if v, ok := details.get(key); ok {
		cached := v.(MovieDetail)
		return &cached, nil
	}
	var movie MovieDetail
	path := fmt.Sprintf("/movie/%d", tmdbID)

	err := getWithFallback(path, nil, &movie, func() bool {
		return movie.Overview == ""
	})
	if err != nil {
		return nil, fmt.Errorf("GetMovieByID(%d): %w", tmdbID, err)
	}

	var enMovie struct {
		Title string `json:"title"`
	}
	enParams := url.Values{}
	enParams.Set("language", "en-US")
	if err := get(path, enParams, &enMovie); err == nil && enMovie.Title != "" {
		movie.EnglishTitle = enMovie.Title
	}

	details.set(key, movie)
	return &movie, nil
}

func GetCredits(tmdbID int) (*Credits, error) {
	var credits Credits
	if err := get(fmt.Sprintf("/movie/%d/credits", tmdbID), nil, &credits); err != nil {
		return nil, fmt.Errorf("GetCredits(%d): %w", tmdbID, err)
	}
	return &credits, nil
}

func GetVideos(tmdbID int) ([]Video, error) {
	var result VideoResult
	path := fmt.Sprintf("/movie/%d/videos", tmdbID)

	err := getWithFallback(path, nil, &result, func() bool {
		return len(result.Results) == 0
	})
	if err != nil {
		return nil, fmt.Errorf("GetVideos(%d): %w", tmdbID, err)
	}
	return result.Results, nil
}

func GetSimilar(tmdbID, page int) (*PaginatedResult[Movie], error) {
	var result PaginatedResult[Movie]
	path := fmt.Sprintf("/movie/%d/similar", tmdbID)

	if err := get(path, pageParams(page), &result); err != nil {
		return nil, fmt.Errorf("GetSimilar(%d): %w", tmdbID, err)
	}
	return &result, nil
}

// SearchFilters narrows a title search. Zero values mean "no filter".
//
// Year is applied by TMDB itself, so paging and totals stay exact. TMDB's search endpoints
// have no genre parameter, so GenreID is applied to the returned page: a page can come back
// with fewer than 20 titles (even none) while page < total_pages, and total_results is the
// unfiltered count (an upper bound). Callers keep paging by total_pages. A scan-ahead that
// refills short pages was rejected: it would make "page" mean different things per filter
// and cost several TMDB calls per request. /discover is not used because it cannot take a
// free-text query.
type SearchFilters struct {
	Year    int
	GenreID int
}

func (f SearchFilters) apply(params url.Values, yearKey string) {
	if f.Year > 0 {
		params.Set(yearKey, strconv.Itoa(f.Year))
	}
}

func filterByGenre[T any](results []T, genreID int, genreIDs func(T) []int) []T {
	if genreID <= 0 {
		return results
	}
	kept := make([]T, 0, len(results))
	for _, r := range results {
		if slices.Contains(genreIDs(r), genreID) {
			kept = append(kept, r)
		}
	}
	return kept
}

func SearchMovies(query string, page int, language string, filters SearchFilters) (*PaginatedResult[Movie], error) {
	params := pageParams(page)
	params.Set("query", query)
	if language != "" {
		params.Set("language", language)
	}
	filters.apply(params, "primary_release_year")

	var result PaginatedResult[Movie]
	if err := get("/search/movie", params, &result); err != nil {
		return nil, fmt.Errorf("SearchMovies(%q): %w", query, err)
	}
	result.Results = filterByGenre(result.Results, filters.GenreID, func(m Movie) []int { return m.GenreIDs })
	return &result, nil
}

func GetGenres() ([]Genre, error) {
	var result struct {
		Genres []Genre `json:"genres"`
	}
	if err := get("/genre/movie/list", nil, &result); err != nil {
		return nil, fmt.Errorf("GetGenres: %w", err)
	}
	return result.Genres, nil
}

func GetPopularSeries(page int) (*PaginatedResult[TVSeries], error) {
	var result PaginatedResult[TVSeries]
	if err := get("/tv/popular", pageParams(page), &result); err != nil {
		return nil, fmt.Errorf("GetPopularSeries: %w", err)
	}
	return &result, nil
}

func GetSeriesByID(tmdbID int) (*TVSeriesDetail, error) {
	key := detailKey{detailKindTV, tmdbID, currentLanguage()}
	if v, ok := details.get(key); ok {
		cached := v.(TVSeriesDetail)
		return &cached, nil
	}
	var series TVSeriesDetail
	path := fmt.Sprintf("/tv/%d", tmdbID)

	err := getWithFallback(path, nil, &series, func() bool {
		return series.Overview == ""
	})
	if err != nil {
		return nil, fmt.Errorf("GetSeriesByID(%d): %w", tmdbID, err)
	}

	var enSeries struct {
		Name string `json:"name"`
	}
	enParams := url.Values{}
	enParams.Set("language", "en-US")
	if err := get(path, enParams, &enSeries); err == nil && enSeries.Name != "" {
		series.EnglishTitle = enSeries.Name
	}

	details.set(key, series)
	return &series, nil
}

// GetSeriesIMDbID returns the IMDb id from /tv/{id}/external_ids ("" when TMDB has none).
func GetSeriesIMDbID(tmdbID int) (string, error) {
	var ids struct {
		IMDbID *string `json:"imdb_id"`
	}
	if err := get(fmt.Sprintf("/tv/%d/external_ids", tmdbID), nil, &ids); err != nil {
		return "", fmt.Errorf("GetSeriesIMDbID(%d): %w", tmdbID, err)
	}
	if ids.IMDbID == nil {
		return "", nil
	}
	return *ids.IMDbID, nil
}

// GetMovieIMDbID returns the IMDb id from /movie/{id}/external_ids ("" when TMDB has none).
func GetMovieIMDbID(tmdbID int) (string, error) {
	var ids struct {
		IMDbID *string `json:"imdb_id"`
	}
	if err := get(fmt.Sprintf("/movie/%d/external_ids", tmdbID), nil, &ids); err != nil {
		return "", fmt.Errorf("GetMovieIMDbID(%d): %w", tmdbID, err)
	}
	if ids.IMDbID == nil {
		return "", nil
	}
	return *ids.IMDbID, nil
}

func GetSeriesCredits(tmdbID int) (*Credits, error) {
	var credits Credits
	if err := get(fmt.Sprintf("/tv/%d/credits", tmdbID), nil, &credits); err != nil {
		return nil, fmt.Errorf("GetSeriesCredits(%d): %w", tmdbID, err)
	}
	return &credits, nil
}

func GetSeriesVideos(tmdbID int) ([]Video, error) {
	var result VideoResult
	path := fmt.Sprintf("/tv/%d/videos", tmdbID)

	err := getWithFallback(path, nil, &result, func() bool {
		return len(result.Results) == 0
	})
	if err != nil {
		return nil, fmt.Errorf("GetSeriesVideos(%d): %w", tmdbID, err)
	}
	return result.Results, nil
}

func GetSimilarSeries(tmdbID, page int) (*PaginatedResult[TVSeries], error) {
	var result PaginatedResult[TVSeries]
	path := fmt.Sprintf("/tv/%d/similar", tmdbID)

	if err := get(path, pageParams(page), &result); err != nil {
		return nil, fmt.Errorf("GetSimilarSeries(%d): %w", tmdbID, err)
	}
	return &result, nil
}

func GetNowAiringSeries(page int) (*PaginatedResult[TVSeries], error) {
	var result PaginatedResult[TVSeries]
	if err := get("/tv/on_the_air", pageParams(page), &result); err != nil {
		return nil, fmt.Errorf("GetNowAiringSeries: %w", err)
	}
	return &result, nil
}

func GetTopRatedSeries(page int) (*PaginatedResult[TVSeries], error) {
	var result PaginatedResult[TVSeries]
	if err := get("/tv/top_rated", pageParams(page), &result); err != nil {
		return nil, fmt.Errorf("GetTopRatedSeries: %w", err)
	}
	return &result, nil
}

func SearchSeries(query string, page int, language string, filters SearchFilters) (*PaginatedResult[TVSeries], error) {
	params := pageParams(page)
	params.Set("query", query)
	if language != "" {
		params.Set("language", language)
	}
	filters.apply(params, "first_air_date_year")

	var result PaginatedResult[TVSeries]
	if err := get("/search/tv", params, &result); err != nil {
		return nil, fmt.Errorf("SearchSeries(%q): %w", query, err)
	}
	result.Results = filterByGenre(result.Results, filters.GenreID, func(s TVSeries) []int { return s.GenreIDs })
	return &result, nil
}

func GetSeriesGenres() ([]Genre, error) {
	var result struct {
		Genres []Genre `json:"genres"`
	}
	if err := get("/genre/tv/list", nil, &result); err != nil {
		return nil, fmt.Errorf("GetSeriesGenres: %w", err)
	}
	return result.Genres, nil
}

// Browse tabs shared by the movie and TV discover helpers.
const (
	TabPopular  = "popular"
	TabNow      = "now"
	TabTopRated = "top_rated"
	TabUpcoming = "upcoming"
)

// upcomingWindowDays is how far ahead the upcoming discover looks.
const upcomingWindowDays = 120

// DiscoverParams builds /discover/{movie,tv} params that keep the ordering of the
// browse tab while filtering by genre (comma list becomes TMDB OR with "|").
func DiscoverParams(series bool, tab, withGenres string, page int, today time.Time) url.Values {
	params := pageParams(page)
	if withGenres != "" {
		params.Set("with_genres", strings.ReplaceAll(withGenres, ",", "|"))
	}
	params.Set("include_adult", "false")
	day := func(d time.Time) string { return d.Format("2006-01-02") }

	switch tab {
	case TabTopRated:
		params.Set("sort_by", "vote_average.desc")
		if series {
			params.Set("vote_count.gte", "200")
		} else {
			params.Set("vote_count.gte", "300")
		}
	case TabNow:
		params.Set("sort_by", "popularity.desc")
		if series {
			params.Set("air_date.gte", day(today))
			params.Set("air_date.lte", day(today.AddDate(0, 0, 7)))
		} else {
			params.Set("region", defaultRegion)
			params.Set("with_release_type", "2|3")
			params.Set("release_date.gte", day(today.AddDate(0, 0, -45)))
			params.Set("release_date.lte", day(today))
		}
	case TabUpcoming:
		params.Set("sort_by", "primary_release_date.asc")
		params.Set("region", defaultRegion)
		params.Set("with_release_type", "2|3")
		params.Set("release_date.gte", day(today))
		params.Set("release_date.lte", day(today.AddDate(0, 0, upcomingWindowDays)))
	default:
		params.Set("sort_by", "popularity.desc")
	}
	return params
}

// bangkokToday returns midnight of the current day in Asia/Bangkok.
func bangkokToday(now time.Time) time.Time {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		loc = time.FixedZone("ICT", 7*60*60)
	}
	n := now.In(loc)
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, loc)
}

// DiscoverUpcoming lists Thai releases from today on, soonest first, optionally by genre.
func DiscoverUpcoming(withGenres string, page int, options RequestOptions) (*PaginatedResult[Movie], error) {
	params := options.apply(DiscoverParams(false, TabUpcoming, withGenres, page, bangkokToday(time.Now())))
	var result PaginatedResult[Movie]
	if err := get("/discover/movie", params, &result); err != nil {
		return nil, fmt.Errorf("DiscoverUpcoming: %w", err)
	}
	return &result, nil
}

func DiscoverMovies(withGenres, tab string, page int) (*PaginatedResult[Movie], error) {
	var result PaginatedResult[Movie]
	if err := get("/discover/movie", DiscoverParams(false, tab, withGenres, page, time.Now()), &result); err != nil {
		return nil, fmt.Errorf("DiscoverMovies: %w", err)
	}
	return &result, nil
}

func DiscoverSeries(withGenres, tab string, page int) (*PaginatedResult[TVSeries], error) {
	var result PaginatedResult[TVSeries]
	if err := get("/discover/tv", DiscoverParams(true, tab, withGenres, page, time.Now()), &result); err != nil {
		return nil, fmt.Errorf("DiscoverSeries: %w", err)
	}
	return &result, nil
}

func GetPersonDetails(personID int, options RequestOptions) (*PersonDetail, error) {
	var result PersonDetail
	path := fmt.Sprintf("/person/%d", personID)
	params := options.apply(nil)

	if err := get(path, params, &result); err != nil {
		return nil, fmt.Errorf("GetPersonDetails(%d): %w", personID, err)
	}
	return &result, nil
}

func SearchPerson(query string, page int) (*PaginatedResult[Person], error) {
	params := pageParams(page)
	params.Set("query", query)

	var result PaginatedResult[Person]
	if err := get("/search/person", params, &result); err != nil {
		return nil, fmt.Errorf("SearchPerson(%q): %w", query, err)
	}
	return &result, nil
}

func GetPersonMovieCredits(personID int) (*PersonMovieCredits, error) {
	var result PersonMovieCredits
	path := fmt.Sprintf("/person/%d/movie_credits", personID)

	if err := get(path, nil, &result); err != nil {
		return nil, fmt.Errorf("GetPersonMovieCredits(%d): %w", personID, err)
	}
	return &result, nil
}

func GetPersonTVCredits(personID int) (*PersonTVCredits, error) {
	var result PersonTVCredits
	path := fmt.Sprintf("/person/%d/tv_credits", personID)

	if err := get(path, nil, &result); err != nil {
		return nil, fmt.Errorf("GetPersonTVCredits(%d): %w", personID, err)
	}
	return &result, nil
}

type MovieReleaseDate struct {
	Certification string `json:"certification"`
	ReleaseDate   string `json:"release_date"`
	Type          int    `json:"type"`
	Note          string `json:"note"`
}

type CountryReleaseDates struct {
	CountryCode  string             `json:"iso_3166_1"`
	ReleaseDates []MovieReleaseDate `json:"release_dates"`
}

type MovieReleaseDatesResult struct {
	ID      int                   `json:"id"`
	Results []CountryReleaseDates `json:"results"`
}

func GetMovieReleaseDates(tmdbID int) (*MovieReleaseDatesResult, error) {
	var result MovieReleaseDatesResult
	path := fmt.Sprintf("/movie/%d/release_dates", tmdbID)

	if err := get(path, nil, &result); err != nil {
		return nil, fmt.Errorf("GetMovieReleaseDates(%d): %w", tmdbID, err)
	}
	return &result, nil
}

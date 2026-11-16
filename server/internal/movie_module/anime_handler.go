package movie_module

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/shared"
	tmdb "github.com/arinsuda/movie-hub/internal/tmdb_module"
	"github.com/gofiber/fiber/v3"
)

// Anime browse bounds. The earliest year is the oldest decade the catalogue covers well;
// a year may be one ahead of the current year for announced releases.
const (
	animeMinYear         = 1960
	animeYearsAhead      = 1
	animeDecadeStep      = 10
	msgInvalidAnimeQuery = "anime filter is invalid"
)

var (
	errAnimeGenre  = errors.New("anime: genre")
	errAnimeFilter = errors.New("anime: filter")
)

// parseAnimeFilter reads and validates the anime query. Genre errors map to INVALID_GENRE,
// every other error to INVALID_SEARCH. decade, year and season are mutually exclusive except
// that season always needs a year.
func parseAnimeFilter(c fiber.Ctx, series bool, now time.Time) (tmdb.AnimeFilter, error) {
	var f tmdb.AnimeFilter
	currentYear := now.In(config.AppLocation).Year()

	f.Sort = queryOr(c, "sort", tmdb.AnimeDefaultSort)
	if !tmdb.IsAnimeSort(f.Sort) || (f.Sort == tmdb.AnimeSortAiring && !series) {
		return f, errAnimeFilter
	}

	f.Country = queryOr(c, "country", tmdb.AnimeDefaultCountry)
	if !tmdb.IsAnimeCountry(f.Country) {
		return f, errAnimeFilter
	}

	f.Keyword = queryOr(c, "keyword", "")
	if _, ok := tmdb.AnimeKeywordID(f.Keyword); f.Keyword != "" && !ok {
		return f, errAnimeFilter
	}

	genre, ok := readBoundedInt(c.Query("genre"), 1, searchGenreMax)
	if !ok || genre == tmdb.AnimeGenreAnimation {
		return f, errAnimeGenre
	}
	f.GenreID = genre

	decade, ok := readBoundedInt(c.Query("decade"), animeMinYear, currentYear)
	if !ok || decade%animeDecadeStep != 0 {
		return f, errAnimeFilter
	}
	year, ok := readBoundedInt(c.Query("year"), animeMinYear, currentYear+animeYearsAhead)
	if !ok {
		return f, errAnimeFilter
	}
	season := queryOr(c, "season", "")
	if season != "" && (!series || !tmdb.IsAnimeSeason(season) || year == 0) {
		return f, errAnimeFilter
	}
	if decade > 0 && (year > 0 || season != "") {
		return f, errAnimeFilter
	}

	f.Decade, f.Year, f.Season = decade, year, season
	return f, nil
}

// queryOr returns the trimmed query value, or def when it is absent or blank.
func queryOr(c fiber.Ctx, key, def string) string {
	if v := strings.TrimSpace(c.Query(key)); v != "" {
		return v
	}
	return def
}

// readBoundedInt parses an optional integer; blank is 0 and valid, anything else must be in [lo, hi].
func readBoundedInt(raw string, lo, hi int) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, true
	}
	n, err := strconv.Atoi(raw)
	return n, err == nil && n >= lo && n <= hi
}

func invalidAnime(c fiber.Ctx) error {
	return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidSearch, msgInvalidAnimeQuery)
}

func animeFilterError(c fiber.Ctx, err error) error {
	if errors.Is(err, errAnimeGenre) {
		return invalidGenre(c)
	}
	return invalidAnime(c)
}

// GetAnimeSeries lists anime TV, ranked by the sort parameter.
func (h *Handler) GetAnimeSeries(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	f, err := parseAnimeFilter(c, true, time.Now())
	if err != nil {
		return animeFilterError(c, err)
	}

	result, err := h.svc.DiscoverAnimeSeries(c.Context(), f, page)
	if err != nil {
		return upstreamFailed(c)
	}
	return c.JSON(result)
}

// GetAnimeMovies lists anime movies. The airing sort and season are TV only.
func (h *Handler) GetAnimeMovies(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	f, err := parseAnimeFilter(c, false, time.Now())
	if err != nil {
		return animeFilterError(c, err)
	}

	result, err := h.svc.DiscoverAnimeMovies(c.Context(), f, page)
	if err != nil {
		return upstreamFailed(c)
	}
	return c.JSON(result)
}

package manga_module

import (
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
)

// Query parameter names and bounds.
const (
	paramCountry      = "country"
	paramFormat       = "format"
	paramSort         = "sort"
	paramGenre        = "genre"
	paramStatus       = "status"
	paramPage         = "page"
	paramQuery        = "q"
	paramIncludeAdult = "include_adult"

	defaultPage    = 1
	maxPage        = 500
	mangaPageSize  = 20
	maxSearchRunes = 100
	similarPerPage = 20
)

// Filter values as the API accepts them. "all" means no filter.
const (
	filterAll = "all"

	countryJP = "JP"
	countryKR = "KR"
	// countryCN covers CN only: AniList takes one countryOfOrigin per query, so TW titles are not included yet.
	countryCN = "CN"
)

// Kinds derived from format and country of origin.
const (
	kindManga  = "manga"
	kindManhwa = "manhwa"
	kindManhua = "manhua"
	kindNovel  = "novel"
)

// countryFilters maps the API value to AniList countryOfOrigin ("" = no filter).
var countryFilters = map[string]string{
	filterAll: "",
	countryJP: countryJP,
	countryKR: countryKR,
	countryCN: countryCN,
}

// formatFilters maps the API value to AniList format_in.
var formatFilters = map[string][]string{
	"manga":    {"MANGA"},
	"one_shot": {"ONE_SHOT"},
	"novel":    {"NOVEL"},
	filterAll:  {"MANGA", "ONE_SHOT", "NOVEL"},
}

// sortFilters maps the API value to AniList sort enums.
var sortFilters = map[string]string{
	"popular":   "POPULARITY_DESC",
	"top_rated": "SCORE_DESC",
	"trending":  "TRENDING_DESC",
	"newest":    "START_DATE_DESC",
}

// statusFilters maps the API value to AniList MediaStatus ("" = no filter).
var statusFilters = map[string]string{
	filterAll:   "",
	"releasing": "RELEASING",
	"finished":  "FINISHED",
}

// mangaGenres is the allowlist of AniList genres offered for browsing. Ecchi and Hentai are
// left out: they are adult-leaning genres that the browse list does not offer to everyone.
var mangaGenres = []string{
	"Action", "Adventure", "Comedy", "Drama", "Fantasy", "Horror", "Mahou Shoujo", "Mecha",
	"Music", "Mystery", "Psychological", "Romance", "Sci-Fi", "Slice of Life", "Sports",
	"Supernatural", "Thriller",
}

// ListFilter is a validated browse request; zero values mean "no filter".
// IncludeAdult is what the client asked for; Adult is the effective flag after the viewer gate.
type ListFilter struct {
	Country      string
	Formats      []string
	Sort         string
	Genre        string
	Status       string
	Page         int
	IncludeAdult bool
	Adult        bool
}

// parseListFilter validates every browse parameter; ok is false when any is out of range.
func parseListFilter(c fiber.Ctx) (ListFilter, bool) {
	var f ListFilter
	var ok bool
	if f.Country, ok = countryFilters[lookupParam(c, paramCountry, filterAll)]; !ok {
		return f, false
	}
	if f.Formats, ok = formatFilters[lookupParam(c, paramFormat, filterAll)]; !ok {
		return f, false
	}
	if f.Sort, ok = sortFilters[lookupParam(c, paramSort, "popular")]; !ok {
		return f, false
	}
	if f.Status, ok = statusFilters[lookupParam(c, paramStatus, filterAll)]; !ok {
		return f, false
	}
	if f.Genre, ok = genreParam(c); !ok {
		return f, false
	}
	if f.Page, ok = pageParam(c); !ok {
		return f, false
	}
	if f.IncludeAdult, ok = includeAdultParam(c); !ok {
		return f, false
	}
	return f, true
}

// includeAdultParam reads ?include_adult; only the literal "true" or "false" (default) is accepted.
func includeAdultParam(c fiber.Ctx) (bool, bool) {
	switch strings.TrimSpace(c.Query(paramIncludeAdult, "false")) {
	case "true":
		return true, true
	case "false":
		return false, true
	default:
		return false, false
	}
}

func lookupParam(c fiber.Ctx, key, fallback string) string {
	return strings.TrimSpace(c.Query(key, fallback))
}

// genreParam returns the AniList genre name; empty means no genre filter.
func genreParam(c fiber.Ctx) (string, bool) {
	g := strings.TrimSpace(c.Query(paramGenre))
	if g == "" {
		return "", true
	}
	return g, slices.Contains(mangaGenres, g)
}

// pageParam reads ?page; ok is false unless it is an integer in [1, maxPage].
func pageParam(c fiber.Ctx) (int, bool) {
	raw := strings.TrimSpace(c.Query(paramPage, strconv.Itoa(defaultPage)))
	n, err := strconv.Atoi(raw)
	return n, err == nil && n >= defaultPage && n <= maxPage
}

// searchQuery trims ?q and reports whether it is present and within the length cap.
func searchQuery(c fiber.Ctx) (string, bool) {
	q := strings.TrimSpace(c.Query(paramQuery))
	return q, q != "" && utf8.RuneCountInString(q) <= maxSearchRunes
}

// Genres returns a copy of the browse genre allowlist.
func Genres() []string {
	return slices.Clone(mangaGenres)
}

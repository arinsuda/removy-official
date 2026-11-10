package movie_module

import (
	"math/rand/v2"
	"regexp"
	"strconv"
	"strings"

	"github.com/arinsuda/movie-hub/internal/shared"
	tmdb "github.com/arinsuda/movie-hub/internal/tmdb_module"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	svc *MovieService
}

func NewHandler(svc *MovieService) *Handler {
	return &Handler{svc: svc}
}

func resolveTMDBLanguage(acceptLanguage string) string {
	normalized := strings.ToLower(strings.TrimSpace(acceptLanguage))

	switch {
	case strings.HasPrefix(normalized, "en"):
		return "en-US"
	case strings.HasPrefix(normalized, "th"):
		return "th-TH"
	default:
		return "th-TH"
	}
}

var genreListRe = regexp.MustCompile(`^\d{1,6}(,\d{1,6}){0,19}$`)

// genresParam reads with_genres (or legacy genre); ok is false when it is not a comma list of ids.
func genresParam(c fiber.Ctx) (string, bool) {
	g := strings.TrimSpace(c.Query("with_genres"))
	if g == "" {
		g = strings.TrimSpace(c.Query("genre"))
	}
	if g == "" {
		return "", true
	}
	return g, genreListRe.MatchString(g)
}

// Bounds for the search year/genre filters (genre ids match genreListRe: up to 6 digits).
const (
	searchYearMin  = 1870
	searchYearMax  = 2200
	searchGenreMax = 999999
)

// searchFilters reads the optional ?year and ?genre of the search endpoints; ok is false
// when one is present but not an integer inside its bounds.
func searchFilters(c fiber.Ctx) (tmdb.SearchFilters, bool) {
	var f tmdb.SearchFilters
	read := func(key string, lo, hi int) (int, bool) {
		raw := strings.TrimSpace(c.Query(key))
		if raw == "" {
			return 0, true
		}
		n, err := strconv.Atoi(raw)
		return n, err == nil && n >= lo && n <= hi
	}
	var ok bool
	if f.Year, ok = read("year", searchYearMin, searchYearMax); !ok {
		return f, false
	}
	if f.GenreID, ok = read("genre", 1, searchGenreMax); !ok {
		return f, false
	}
	return f, true
}

// Error messages are for logs and developers; clients branch on the code.
const (
	msgInvalidGenre  = "with_genres must be a comma-separated list of genre ids"
	msgInvalidSearch = "year or genre is out of range"
	msgInvalidQuery  = "q is required"
	msgInvalidID     = "id must be a positive integer"
	msgNotFound      = "not found"
	msgUpstream      = "failed to load data from TMDB"
)

func invalidGenre(c fiber.Ctx) error {
	return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidGenre, msgInvalidGenre)
}
func invalidSearch(c fiber.Ctx) error {
	return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidSearch, msgInvalidSearch)
}
func invalidQuery(c fiber.Ctx) error {
	return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidQuery, msgInvalidQuery)
}
func invalidID(c fiber.Ctx) error {
	return shared.WriteAPIError(c, fiber.StatusBadRequest, shared.ErrorCodeInvalidID, msgInvalidID)
}
func notFound(c fiber.Ctx) error {
	return shared.WriteAPIError(c, fiber.StatusNotFound, shared.ErrorCodeNotFound, msgNotFound)
}
func upstreamFailed(c fiber.Ctx) error {
	return shared.WriteAPIError(c, fiber.StatusInternalServerError, shared.ErrorCodeUpstream, msgUpstream)
}

func (h *Handler) GetPopular(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	withGenres, ok := genresParam(c)
	if !ok {
		return invalidGenre(c)
	}

	if withGenres != "" {
		result, err := h.svc.DiscoverMovies(c.Context(), withGenres, tmdb.TabPopular, page)
		if err != nil {
			return upstreamFailed(c)
		}
		return c.JSON(result)
	}

	result, err := h.svc.GetPopular(c.Context(), page)
	if err != nil {
		return upstreamFailed(c)
	}
	return c.JSON(result)
}

func (h *Handler) GetNowPlaying(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	withGenres, ok := genresParam(c)
	if !ok {
		return invalidGenre(c)
	}

	if withGenres != "" {
		result, err := h.svc.DiscoverMovies(c.Context(), withGenres, tmdb.TabNow, page)
		if err != nil {
			return upstreamFailed(c)
		}
		return c.JSON(result)
	}

	options := tmdb.RequestOptions{
		Language: resolveTMDBLanguage(c.Get("Accept-Language")),
		Region:   "TH",
	}

	result, err := h.svc.GetNowPlaying(c.Context(), page, options)
	if err != nil {
		return upstreamFailed(c)
	}
	return c.JSON(result)
}

func (h *Handler) GetTopRated(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	withGenres, ok := genresParam(c)
	if !ok {
		return invalidGenre(c)
	}

	if withGenres != "" {
		result, err := h.svc.DiscoverMovies(c.Context(), withGenres, tmdb.TabTopRated, page)
		if err != nil {
			return upstreamFailed(c)
		}
		return c.JSON(result)
	}

	result, err := h.svc.GetTopRated(c.Context(), page)
	if err != nil {
		return upstreamFailed(c)
	}
	return c.JSON(result)
}

// GetUpcoming lists Thai releases from today on, soonest first. Date filtering
// and ordering happen in TMDB discover; results is always an array.
func (h *Handler) GetUpcoming(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	withGenres, ok := genresParam(c)
	if !ok {
		return invalidGenre(c)
	}

	options := tmdb.RequestOptions{
		Language: resolveTMDBLanguage(c.Get("Accept-Language")),
		Region:   "TH",
	}
	result, err := h.svc.GetUpcoming(c.Context(), withGenres, page, options)
	if err != nil {
		return upstreamFailed(c)
	}

	movies := result.Results
	if movies == nil {
		movies = []MovieDTO{}
	}
	return c.JSON(fiber.Map{
		"page":          result.Page,
		"results":       movies,
		"total_pages":   result.TotalPages,
		"total_results": result.TotalResults,
	})
}

// func (h *Handler) GetUpcoming(c fiber.Ctx) error {
// 	wantUpcoming := 10 // ยังไม่ฉาย
// 	wantReleased := 10 // ฉายแล้ว (ภายใน 30 วัน)

// 	today := time.Now().Truncate(24 * time.Hour)
// 	monthAgo := today.AddDate(0, 0, -30)

// 	var upcoming []tmdb.Movie
// 	var released []tmdb.Movie

// 	for page := 1; page <= 10; page++ { // cap ไว้ที่ 10 pages กันวน
// 		result, err := tmdb.GetUpcoming(page)
// 		if err != nil {
// 			break
// 		}

// 		for _, movie := range result.Results {
// 			if len(movie.ReleaseDate) < 10 {
// 				continue
// 			}
// 			releaseDate, err := time.Parse("2006-01-02", movie.ReleaseDate)
// 			if err != nil {
// 				continue
// 			}

// 			if releaseDate.After(today) || releaseDate.Equal(today) {
// 				// ยังไม่ฉาย
// 				upcoming = append(upcoming, movie)
// 			} else if releaseDate.After(monthAgo) {
// 				// ฉายแล้วแต่ไม่เกิน 30 วัน
// 				released = append(released, movie)
// 			}
// 		}

// 		// ได้ครบแล้ว หยุด
// 		if len(upcoming) >= wantUpcoming && len(released) >= wantReleased {
// 			break
// 		}

// 		// ไม่มี page ต่อแล้ว
// 		if page >= result.TotalPages {
// 			break
// 		}
// 	}

// 	// Trim ให้พอดี
// 	if len(upcoming) > wantUpcoming {
// 		upcoming = upcoming[:wantUpcoming]
// 	}
// 	if len(released) > wantReleased {
// 		released = released[:wantReleased]
// 	}

// 	// รวมกันแล้วส่ง — upcoming ก่อน, released ตามหลัง
// 	combined := append(upcoming, released...)

// 	return c.JSON(fiber.Map{
// 		"results":       combined,
// 		"total_results": len(combined),
// 		"page":          1,
// 		"total_pages":   1, // Frontend ไม่ต้อง paginate แล้ว
// 	})
// }

func (h *Handler) GetUpcomingByYear(c fiber.Ctx) error {
	year, _ := strconv.Atoi(c.Params("year"))
	page, _ := strconv.Atoi(c.Query("page", "1"))

	result, err := tmdb.DiscoverUpcomingByYear(year, page)

	if err != nil {
		return upstreamFailed(c)
	}

	return c.JSON(result)
}

func (h *Handler) Search(c fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return invalidQuery(c)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	filters, ok := searchFilters(c)
	if !ok {
		return invalidSearch(c)
	}

	result, err := h.svc.SearchMovies(c.Context(), query, page, resolveTMDBLanguage(c.Get("Accept-Language")), filters)
	if err != nil {
		return upstreamFailed(c)
	}
	return c.JSON(result)
}

func (h *Handler) GetGenres(c fiber.Ctx) error {
	genres, err := tmdb.GetGenres()
	if err != nil {
		return upstreamFailed(c)
	}
	return c.JSON(fiber.Map{"genres": genres})
}

func (h *Handler) GetByID(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return invalidID(c)
	}

	movie, err := h.svc.GetMovieByID(c.Context(), id)
	if err != nil {
		return notFound(c)
	}

	credits, _ := tmdb.GetCredits(id)
	videos, _ := tmdb.GetVideos(id)

	return c.JSON(fiber.Map{
		"movie":   movie,
		"credits": credits,
		"videos":  videos,
	})
}

func (h *Handler) GetSimilar(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return invalidID(c)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))

	result, err := h.svc.GetSimilar(c.Context(), id, page)
	if err != nil {
		return upstreamFailed(c)
	}
	return c.JSON(result)
}

func (h *Handler) GetPopularSeries(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	withGenres, ok := genresParam(c)
	if !ok {
		return invalidGenre(c)
	}
	if withGenres != "" {
		result, err := h.svc.DiscoverSeries(c.Context(), withGenres, tmdb.TabPopular, page)
		if err != nil {
			return upstreamFailed(c)
		}
		return c.JSON(result)
	}

	result, err := h.svc.GetPopularSeries(c.Context(), page)
	if err != nil {
		return upstreamFailed(c)
	}
	return c.JSON(result)
}

func (h *Handler) GetNowAiringSeries(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	withGenres, ok := genresParam(c)
	if !ok {
		return invalidGenre(c)
	}
	if withGenres != "" {
		result, err := h.svc.DiscoverSeries(c.Context(), withGenres, tmdb.TabNow, page)
		if err != nil {
			return upstreamFailed(c)
		}
		return c.JSON(result)
	}

	result, err := h.svc.GetNowAiringSeries(c.Context(), page)
	if err != nil {
		return upstreamFailed(c)
	}
	return c.JSON(result)
}

func (h *Handler) GetTopRatedSeries(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	withGenres, ok := genresParam(c)
	if !ok {
		return invalidGenre(c)
	}
	if withGenres != "" {
		result, err := h.svc.DiscoverSeries(c.Context(), withGenres, tmdb.TabTopRated, page)
		if err != nil {
			return upstreamFailed(c)
		}
		return c.JSON(result)
	}

	result, err := h.svc.GetTopRatedSeries(c.Context(), page)
	if err != nil {
		return upstreamFailed(c)
	}
	return c.JSON(result)
}

func (h *Handler) SearchSeries(c fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return invalidQuery(c)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	filters, ok := searchFilters(c)
	if !ok {
		return invalidSearch(c)
	}
	result, err := h.svc.SearchSeries(c.Context(), query, page, resolveTMDBLanguage(c.Get("Accept-Language")), filters)
	if err != nil {
		return upstreamFailed(c)
	}
	return c.JSON(result)
}

func (h *Handler) GetSeriesGenres(c fiber.Ctx) error {
	genres, err := tmdb.GetSeriesGenres()
	if err != nil {
		return upstreamFailed(c)
	}
	return c.JSON(fiber.Map{"genres": genres})
}

func (h *Handler) GetSeriesByID(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return invalidID(c)
	}

	series, err := h.svc.GetSeriesByID(c.Context(), id)
	if err != nil {
		return notFound(c)
	}

	credits, _ := tmdb.GetSeriesCredits(id)
	videos, _ := tmdb.GetSeriesVideos(id)

	return c.JSON(fiber.Map{
		"series":  series,
		"credits": credits,
		"videos":  videos,
	})
}

func (h *Handler) GetSimilarSeries(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return invalidID(c)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	result, err := h.svc.GetSimilarSeries(c.Context(), id, page)
	if err != nil {
		return upstreamFailed(c)
	}
	return c.JSON(result)
}

func (h *Handler) GetRecommended(c fiber.Ctx) error {
	genreIDs := c.Query("with_genres")
	if genreIDs == "" {
		return h.GetPopular(c)
	}

	var allResults []tmdb.Movie

	for page := 1; page <= 3; page++ {
		result, err := tmdb.DiscoverMovies(genreIDs, tmdb.TabPopular, page)
		if err != nil {
			break
		}
		allResults = append(allResults, result.Results...)
		if page >= result.TotalPages {
			break
		}
	}

	rand.Shuffle(len(allResults), func(i, j int) {
		allResults[i], allResults[j] = allResults[j], allResults[i]
	})

	if len(allResults) > 10 {
		allResults = allResults[:10]
	}

	enriched, err := h.svc.EnrichMovies(c.Context(), &tmdb.PaginatedResult[tmdb.Movie]{Results: allResults})
	if err != nil {
		return upstreamFailed(c)
	}

	return c.JSON(fiber.Map{
		"results":       enriched.Results,
		"total_results": len(allResults),
	})
}

func (h *Handler) SearchActor(c fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return invalidQuery(c)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))

	result, err := tmdb.SearchPerson(query, page)
	if err != nil {
		return upstreamFailed(c)
	}
	return c.JSON(result)
}

func (h *Handler) GetActorDetails(c fiber.Ctx) error {
	personID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return invalidID(c)
	}

	options := tmdb.RequestOptions{
		Language: resolveTMDBLanguage(c.Get("Accept-Language")),
	}

	result, err := tmdb.GetPersonDetails(personID, options)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return notFound(c)
		}
		return upstreamFailed(c)
	}
	return c.JSON(result)
}

func (h *Handler) GetMoviesByActor(c fiber.Ctx) error {
	personID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return invalidID(c)
	}

	result, err := h.svc.GetPersonMovieCredits(c.Context(), personID)
	if err != nil {
		return upstreamFailed(c)
	}
	return c.JSON(result)
}

func (h *Handler) GetSeriesByActor(c fiber.Ctx) error {
	personID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return invalidID(c)
	}

	result, err := h.svc.GetPersonTVCredits(c.Context(), personID)
	if err != nil {
		return upstreamFailed(c)
	}
	return c.JSON(result)
}

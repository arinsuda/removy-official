package movie_module

import (
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

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

func (h *Handler) GetPopular(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	withGenres := c.Query("with_genres")
	if withGenres == "" {
		withGenres = c.Query("genre")
	}

	if withGenres != "" {
		result, err := tmdb.DiscoverMovies(withGenres, page)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลหนังไม่สำเร็จ"})
		}
		return c.JSON(result)
	}

	result, err := h.svc.GetPopular(c.Context(), page)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลหนังไม่สำเร็จ"})
	}
	return c.JSON(result)
}

func (h *Handler) GetNowPlaying(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	withGenres := c.Query("with_genres")
	if withGenres == "" {
		withGenres = c.Query("genre")
	}

	if withGenres != "" {
		result, err := tmdb.DiscoverMovies(withGenres, page)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลหนังไม่สำเร็จ"})
		}
		return c.JSON(result)
	}

	options := tmdb.RequestOptions{
		Language: resolveTMDBLanguage(c.Get("Accept-Language")),
		Region:   "TH",
	}

	result, err := h.svc.GetNowPlaying(c.Context(), page, options)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลหนังไม่สำเร็จ"})
	}
	return c.JSON(result)
}

func (h *Handler) GetTopRated(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	withGenres := c.Query("with_genres")
	if withGenres == "" {
		withGenres = c.Query("genre")
	}

	if withGenres != "" {
		result, err := tmdb.DiscoverMovies(withGenres, page)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลหนังไม่สำเร็จ"})
		}
		return c.JSON(result)
	}

	result, err := h.svc.GetTopRated(c.Context(), page)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลหนังไม่สำเร็จ"})
	}
	return c.JSON(result)
}

func (h *Handler) GetUpcoming(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))

	result, err := h.svc.GetUpcoming(c.Context(), page)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลหนังไม่สำเร็จ"})
	}

	location, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลหนังไม่สำเร็จ"})
	}

	now := time.Now().In(location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)

	var upcomingOnly []MovieDTO

	for _, movie := range result.Results {
		if len(movie.ReleaseDate) < 10 {
			continue
		}
		releaseDate, err := time.ParseInLocation("2006-01-02", movie.ReleaseDate, location)
		if err != nil {
			continue
		}

		if releaseDate.After(today) || releaseDate.Equal(today) {
			upcomingOnly = append(upcomingOnly, movie)
		}
	}

	return c.JSON(fiber.Map{
		"page":          result.Page,
		"results":       upcomingOnly,
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
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to load upcoming movies",
		})
	}

	return c.JSON(result)
}

func (h *Handler) Search(c fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(400).JSON(fiber.Map{"error": "กรุณาระบุคำค้นหา"})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))

	result, err := h.svc.SearchMovies(c.Context(), query, page)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ค้นหาไม่สำเร็จ"})
	}
	return c.JSON(result)
}

func (h *Handler) GetGenres(c fiber.Ctx) error {
	genres, err := tmdb.GetGenres()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูล genre ไม่สำเร็จ"})
	}
	return c.JSON(fiber.Map{"genres": genres})
}

func (h *Handler) GetByID(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}

	movie, err := h.svc.GetMovieByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบหนัง"})
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
		return c.Status(400).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))

	result, err := tmdb.GetSimilar(id, page)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลไม่สำเร็จ"})
	}
	return c.JSON(result)
}

func (h *Handler) GetPopularSeries(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	result, err := tmdb.GetPopularSeries(page)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลซีรีส์ไม่สำเร็จ"})
	}
	return c.JSON(result)
}

func (h *Handler) GetNowAiringSeries(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	result, err := tmdb.GetNowAiringSeries(page)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลซีรีส์ไม่สำเร็จ"})
	}
	return c.JSON(result)
}

func (h *Handler) GetTopRatedSeries(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	result, err := tmdb.GetTopRatedSeries(page)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลซีรีส์ไม่สำเร็จ"})
	}
	return c.JSON(result)
}

func (h *Handler) SearchSeries(c fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(400).JSON(fiber.Map{"error": "กรุณาระบุคำค้นหา"})
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	result, err := tmdb.SearchSeries(query, page)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ค้นหาไม่สำเร็จ"})
	}
	return c.JSON(result)
}

func (h *Handler) GetSeriesGenres(c fiber.Ctx) error {
	genres, err := tmdb.GetSeriesGenres()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูล genre ไม่สำเร็จ"})
	}
	return c.JSON(fiber.Map{"genres": genres})
}

func (h *Handler) GetSeriesByID(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}

	series, err := tmdb.GetSeriesByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบซีรีส์"})
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
		return c.Status(400).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	result, err := tmdb.GetSimilarSeries(id, page)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลไม่สำเร็จ"})
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
		result, err := tmdb.DiscoverMovies(genreIDs, page)
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

	return c.JSON(fiber.Map{
		"results":       allResults,
		"total_results": len(allResults),
	})
}

func (h *Handler) SearchActor(c fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(400).JSON(fiber.Map{"error": "กรุณาระบุคำค้นหา"})
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))

	result, err := tmdb.SearchPerson(query, page)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ค้นหานักแสดงไม่สำเร็จ"})
	}
	return c.JSON(result)
}

func (h *Handler) GetActorDetails(c fiber.Ctx) error {
	personID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}

	options := tmdb.RequestOptions{
		Language: resolveTMDBLanguage(c.Get("Accept-Language")),
	}

	result, err := tmdb.GetPersonDetails(personID, options)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return c.Status(404).JSON(fiber.Map{"error": "ไม่พบข้อมูลนักแสดง"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลนักแสดงไม่สำเร็จ"})
	}
	return c.JSON(result)
}

func (h *Handler) GetMoviesByActor(c fiber.Ctx) error {
	personID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}

	result, err := tmdb.GetPersonMovieCredits(personID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลผลงานไม่สำเร็จ"})
	}
	return c.JSON(result)
}

func (h *Handler) GetSeriesByActor(c fiber.Ctx) error {
	personID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}

	result, err := tmdb.GetPersonTVCredits(personID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ดึงข้อมูลผลงานไม่สำเร็จ"})
	}
	return c.JSON(result)
}

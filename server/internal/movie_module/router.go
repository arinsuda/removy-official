package movie_module

import (
	external "github.com/arinsuda/movie-hub/internal/external_rating_module"
	"github.com/arinsuda/movie-hub/internal/shared"
	mw "github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
)

// New builds the handler shared by RegisterPublicRoutes and RegisterRoutes.
func New(ratingRepo shared.RatingStatsReader, ext *external.Service) *Handler {
	svc := NewMovieService(ratingRepo)
	svc.SetExternalRatings(ext)
	return NewHandler(svc)
}

// RegisterPublicRoutes mounts the movie, TV and actor endpoints guests may read.
// Register it on the root /api router before any auth group; dynamic ids are
// constrained to integers so /movies/recommended, /movies/upcoming/:year,
// /tv/top-rated and /actors/search fall through to RegisterRoutes.
//
// searchGuard is an extra stricter limiter applied (after guards) to the search routes only.
func RegisterPublicRoutes(router fiber.Router, h *Handler, searchGuard fiber.Handler, guards ...fiber.Handler) {
	searchGuards := append(append([]fiber.Handler{}, guards...), searchGuard)
	mw.GetRoute(router, "/movies/popular", h.GetPopular, guards...)
	mw.GetRoute(router, "/movies/now-playing", h.GetNowPlaying, guards...)
	mw.GetRoute(router, "/movies/top-rated", h.GetTopRated, guards...)
	mw.GetRoute(router, "/movies/upcoming", h.GetUpcoming, guards...)
	mw.GetRoute(router, "/movies/search", h.Search, searchGuards...)
	mw.GetRoute(router, "/movies/genres", h.GetGenres, guards...)
	mw.GetRoute(router, "/movies/:id<int>", h.GetByID, guards...)
	mw.GetRoute(router, "/movies/:id<int>/similar", h.GetSimilar, guards...)
	mw.GetRoute(router, "/movies/:id<int>/watch-providers", h.GetMovieWatchProviders, guards...)

	mw.GetRoute(router, "/tv/popular", h.GetPopularSeries, guards...)
	mw.GetRoute(router, "/tv/now-airing", h.GetNowAiringSeries, guards...)
	mw.GetRoute(router, "/tv/search", h.SearchSeries, searchGuards...)
	mw.GetRoute(router, "/tv/genres", h.GetSeriesGenres, guards...)
	mw.GetRoute(router, "/tv/:id<int>", h.GetSeriesByID, guards...)
	mw.GetRoute(router, "/tv/:id<int>/similar", h.GetSimilarSeries, guards...)
	mw.GetRoute(router, "/tv/:id<int>/watch-providers", h.GetSeriesWatchProviders, guards...)

	mw.GetRoute(router, "/actors/:id<int>", h.GetActorDetails, guards...)
	mw.GetRoute(router, "/actors/:id<int>/movies", h.GetMoviesByActor, guards...)
	mw.GetRoute(router, "/actors/:id<int>/tv", h.GetSeriesByActor, guards...)
}

// RegisterRoutes mounts the browse, personalized and actor-search endpoints on
// the authenticated router.
func RegisterRoutes(router fiber.Router, h *Handler) {
	movies := router.Group("/movies")
	movies.Get("/upcoming/:year", h.GetUpcomingByYear)
	movies.Get("/recommended", h.GetRecommended)

	series := router.Group("/tv")
	series.Get("/top-rated", h.GetTopRatedSeries)

	actors := router.Group("/actors")
	actors.Get("/search", h.SearchActor)
}

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

// RegisterPublicRoutes mounts the detail endpoints guests may read. Register it on the root
// /api router before any auth group; ids are constrained to integers so /movies/popular and
// the other fixed paths fall through to RegisterRoutes.
func RegisterPublicRoutes(router fiber.Router, h *Handler, guards ...fiber.Handler) {
	mw.GetRoute(router, "/movies/:id<int>", h.GetByID, guards...)
	mw.GetRoute(router, "/movies/:id<int>/similar", h.GetSimilar, guards...)
	mw.GetRoute(router, "/movies/:id<int>/watch-providers", h.GetMovieWatchProviders, guards...)

	mw.GetRoute(router, "/tv/:id<int>", h.GetSeriesByID, guards...)
	mw.GetRoute(router, "/tv/:id<int>/similar", h.GetSimilarSeries, guards...)
	mw.GetRoute(router, "/tv/:id<int>/watch-providers", h.GetSeriesWatchProviders, guards...)
}

// RegisterRoutes mounts the browse, search and actor endpoints on the authenticated router.
func RegisterRoutes(router fiber.Router, h *Handler) {
	movies := router.Group("/movies")
	movies.Get("/popular", h.GetPopular)
	movies.Get("/now-playing", h.GetNowPlaying)
	movies.Get("/top-rated", h.GetTopRated)
	movies.Get("/upcoming", h.GetUpcoming)
	movies.Get("/upcoming/:year", h.GetUpcomingByYear)
	movies.Get("/search", h.Search)
	movies.Get("/genres", h.GetGenres)
	movies.Get("/recommended", h.GetRecommended)

	series := router.Group("/tv")
	series.Get("/popular", h.GetPopularSeries)
	series.Get("/now-airing", h.GetNowAiringSeries)
	series.Get("/top-rated", h.GetTopRatedSeries)
	series.Get("/search", h.SearchSeries)
	series.Get("/genres", h.GetSeriesGenres)

	actors := router.Group("/actors")
	actors.Get("/search", h.SearchActor)
	actors.Get("/:id", h.GetActorDetails)
	actors.Get("/:id/movies", h.GetMoviesByActor)
	actors.Get("/:id/tv", h.GetSeriesByActor)
}

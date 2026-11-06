package movie_module

import (
	"github.com/arinsuda/movie-hub/internal/shared"
	tmdb "github.com/arinsuda/movie-hub/internal/tmdb_module"
)

type MediaType string
type ListType string

const (
	MediaMovie  MediaType = "movie"
	MediaSeries MediaType = "tv"
)

const (
	ListWatchlist ListType = "watchlist"
	ListWatched   ListType = "watched"
)

type RatingSourceResponse struct {
	Average   *float64 `json:"average"`
	Count     int      `json:"count"`
	Available bool     `json:"available"`
	Scale     float64  `json:"scale"`
}

type MediaRatingsResponse struct {
	POPCORN RatingSourceResponse `json:"popcorn"`
	TMDB    RatingSourceResponse `json:"tmdb"`
}

type MovieDTO struct {
	tmdb.Movie
	Ratings MediaRatingsResponse `json:"ratings"`
}

type MovieDetailDTO struct {
	tmdb.MovieDetail
	Ratings MediaRatingsResponse `json:"ratings"`
}

type TVSeriesDTO struct {
	tmdb.TVSeries
	Ratings MediaRatingsResponse `json:"ratings"`
}

func NewMediaRatingsResponse(popcornStats shared.RatingStats, tmdbAvg float32, tmdbVotes int) MediaRatingsResponse {
	var popcornAvg *float64
	if popcornStats.Average != nil {
		popcornAvg = popcornStats.Average
	}
	popcornAvailable := popcornAvg != nil && popcornStats.Count > 0

	var tmdbAvgPtr *float64
	if tmdbVotes > 0 {
		v := float64(tmdbAvg)
		tmdbAvgPtr = &v
	}
	tmdbAvailable := tmdbAvgPtr != nil && tmdbVotes > 0

	return MediaRatingsResponse{
		POPCORN: RatingSourceResponse{
			Average:   popcornAvg,
			Count:     popcornStats.Count,
			Available: popcornAvailable,
			Scale:     5.0,
		},
		TMDB: RatingSourceResponse{
			Average:   tmdbAvgPtr,
			Count:     tmdbVotes,
			Available: tmdbAvailable,
			Scale:     10.0,
		},
	}
}

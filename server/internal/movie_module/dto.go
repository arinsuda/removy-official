package movie_module

import (
	external "github.com/arinsuda/movie-hub/internal/external_rating_module"
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

// MediaRatingsResponse always carries popcorn and tmdb. IMDb and Rotten Tomatoes are
// only set on movie and TV detail responses.
type MediaRatingsResponse struct {
	POPCORN        RatingSourceResponse  `json:"popcorn"`
	TMDB           RatingSourceResponse  `json:"tmdb"`
	IMDb           *RatingSourceResponse `json:"imdb,omitempty"`
	RottenTomatoes *RatingSourceResponse `json:"rotten_tomatoes,omitempty"`
	// ExternalPending is set while the IMDb/RT lookup is still running; clients should ask again.
	ExternalPending bool `json:"external_pending,omitempty"`
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

type TVSeriesDetailDTO struct {
	tmdb.TVSeriesDetail
	Ratings MediaRatingsResponse `json:"ratings"`
}

type PersonMovieCreditItemDTO struct {
	tmdb.PersonMovieCreditItem
	Ratings MediaRatingsResponse `json:"ratings"`
}

type PersonMovieCreditsDTO struct {
	Cast []PersonMovieCreditItemDTO `json:"cast"`
	Crew []MovieDTO                 `json:"crew"`
}

type PersonTVCreditItemDTO struct {
	tmdb.PersonTVCreditItem
	Ratings MediaRatingsResponse `json:"ratings"`
}

type PersonTVCreditsDTO struct {
	Cast []PersonTVCreditItemDTO `json:"cast"`
	Crew []TVSeriesDTO           `json:"crew"`
}

func NewMediaRatingsResponse(popcornStats shared.RatingStats, tmdbAvg float32, tmdbVotes int) MediaRatingsResponse {
	// No reviews reads as 0 (not null) with available=false, so clients can show "0.0".
	popcornAvg := new(float64)
	if popcornStats.Average != nil {
		popcornAvg = popcornStats.Average
	}
	popcornAvailable := popcornStats.Average != nil && popcornStats.Count > 0

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

// WithExternal adds IMDb (scale 10) and Rotten Tomatoes (percent) sources.
func (r MediaRatingsResponse) WithExternal(ext external.Ratings) MediaRatingsResponse {
	imdb := RatingSourceResponse{Average: ext.IMDbRating, Count: ext.IMDbVotes, Available: ext.IMDbRating != nil, Scale: 10}
	rt := RatingSourceResponse{Available: ext.RTScore != nil, Scale: 100}
	if ext.RTScore != nil {
		v := float64(*ext.RTScore)
		rt.Average = &v
	}
	r.IMDb, r.RottenTomatoes = &imdb, &rt
	r.ExternalPending = ext.Pending
	return r
}

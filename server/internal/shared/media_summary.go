package shared

import (
	tmdbmodule "github.com/arinsuda/movie-hub/internal/tmdb_module"
)

type MediaSummary struct {
	ID          int                `json:"id"`
	Title       string             `json:"title"`
	PosterURL   string             `json:"poster_url"`
	MediaType   MediaType          `json:"media_type"`
	Genres      []tmdbmodule.Genre `json:"genres"`
	VoteAverage float32            `json:"vote_average"`
	Ratings     *SummaryRatings    `json:"ratings,omitempty"`
}

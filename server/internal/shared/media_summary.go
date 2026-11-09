package shared

import (
	tmdbmodule "github.com/arinsuda/movie-hub/internal/tmdb_module"
)

type MediaSummary struct {
	ID          int                `json:"id"`
	Title       string             `json:"title"`
	ReleaseDate string             `json:"release_date,omitempty"`
	PosterURL   string             `json:"poster_url"`
	MediaType   MediaType          `json:"media_type"`
	Genres      []tmdbmodule.Genre `json:"genres"`
	VoteAverage float32            `json:"vote_average"`
	Ratings     *SummaryRatings    `json:"ratings,omitempty"`
}

// FetchMediaSummary builds a summary from the (cached) TMDB detail; fields stay empty when TMDB fails.
func FetchMediaSummary(id int, mediaType MediaType) MediaSummary {
	media := MediaSummary{ID: id, MediaType: mediaType}
	switch mediaType {
	case MediaTypeMovie:
		if d, err := tmdbmodule.GetMovieByID(id); err == nil && d != nil {
			media.Title = d.Title
			media.ReleaseDate = d.ReleaseDate
			media.PosterURL = d.PosterPath
			media.Genres = d.Genres
			media.VoteAverage = float32(d.VoteAverage)
		}
	case MediaTypeTV:
		if d, err := tmdbmodule.GetSeriesByID(id); err == nil && d != nil {
			media.Title = d.Name
			media.ReleaseDate = d.FirstAirDate
			media.PosterURL = d.PosterPath
			media.Genres = d.Genres
			media.VoteAverage = float32(d.VoteAverage)
		}
	}
	return media
}

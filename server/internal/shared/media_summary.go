package shared

import (
	"context"

	tmdbmodule "github.com/arinsuda/movie-hub/internal/tmdb_module"
)

type MediaSummary struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	// EnglishTitle is server-side only (admin views fall back to it); it is not part of the API shape.
	EnglishTitle string             `json:"-"`
	ReleaseDate  string             `json:"release_date,omitempty"`
	PosterURL    string             `json:"poster_url"`
	MediaType    MediaType          `json:"media_type"`
	Genres       []tmdbmodule.Genre `json:"genres"`
	VoteAverage  float32            `json:"vote_average"`
	Kind         string             `json:"kind,omitempty"`
	Ratings      *SummaryRatings    `json:"ratings,omitempty"`
}

// FetchMediaSummary builds one summary. Lists and feeds should call MediaSummaries once instead, so manga
// is read in a single batch. Fields stay empty when the provider fails.
func FetchMediaSummary(id int, mediaType MediaType) MediaSummary {
	summaries, err := MediaSummaries(context.Background(), []MediaKey{{ID: id, Type: mediaType}})
	if err != nil {
		return MediaSummary{ID: id, MediaType: mediaType}
	}
	return summaries[MediaKey{ID: id, Type: mediaType}]
}

// fetchSummary builds a movie or TV summary through the resolver; fields stay empty when the provider fails.
func fetchSummary(id int, mediaType MediaType) MediaSummary {
	media := MediaSummary{ID: id, MediaType: mediaType}
	info, err := ResolveMedia(context.Background(), id, mediaType)
	if err != nil {
		return media
	}
	media.Title = info.Title
	media.EnglishTitle = info.EnglishTitle
	media.ReleaseDate = info.ReleaseDate
	media.PosterURL = info.PosterURL
	media.Genres = info.Genres
	media.VoteAverage = info.VoteAverage
	media.Kind = info.Kind
	return media
}

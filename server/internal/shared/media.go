package shared

import (
	"context"
	"math"
)

type MediaType string

const (
	MediaTypeMovie MediaType = "movie"
	MediaTypeTV    MediaType = "tv"
	// MediaTypeManga titles are AniList ids, not TMDB ids: the same number names a different title per type.
	MediaTypeManga MediaType = "manga"
)

// IsValidMediaType reports whether t is a media type that reviews, library, likes, BMOL, lists and stats accept.
// Watch logs and events stay movie and TV only and check their own list.
func IsValidMediaType(t MediaType) bool {
	switch t {
	case MediaTypeMovie, MediaTypeTV, MediaTypeManga:
		return true
	}
	return false
}

type MediaIdentity struct {
	ID   int
	Type MediaType
}

type MediaKey struct {
	ID   int
	Type MediaType
}

type RatingStats struct {
	Average *float64
	Count   int
}

type RatingStatsReader interface {
	GetMediaRating(ctx context.Context, id MediaIdentity) (RatingStats, error)
	GetBatchMediaRatings(ctx context.Context, ids []MediaIdentity) (map[MediaKey]RatingStats, error)
}

const (
	MinRating  = 0.5
	MaxRating  = 5.0
	RatingStep = 0.5
)

func IsValidRating(rating float32) bool {
	if rating < MinRating || rating > MaxRating {
		return false
	}
	valFloat := float64(rating) * 10.0
	valRounded := math.Round(valFloat)
	if math.Abs(valFloat-valRounded) > 1e-4 {
		return false
	}
	return int(valRounded)%5 == 0
}

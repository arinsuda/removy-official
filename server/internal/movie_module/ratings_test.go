package movie_module

import (
	"encoding/json"
	"testing"

	external "github.com/arinsuda/movie-hub/internal/external_rating_module"
	"github.com/arinsuda/movie-hub/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoReviewsGivesPopcornZeroUnavailable(t *testing.T) {
	r := NewMediaRatingsResponse(shared.RatingStats{Average: nil, Count: 0}, 0, 0)
	require.NotNil(t, r.POPCORN.Average)
	assert.Equal(t, 0.0, *r.POPCORN.Average)
	assert.False(t, r.POPCORN.Available)
	assert.Equal(t, 5.0, r.POPCORN.Scale)
	assert.False(t, r.TMDB.Available)
	assert.Nil(t, r.IMDb, "lists omit imdb")
}

func TestPopcornWithReviews(t *testing.T) {
	avg := 4.5
	r := NewMediaRatingsResponse(shared.RatingStats{Average: &avg, Count: 2}, 7.5, 100)
	assert.True(t, r.POPCORN.Available)
	assert.Equal(t, 4.5, *r.POPCORN.Average)
	assert.True(t, r.TMDB.Available)
}

func TestWithExternalShape(t *testing.T) {
	imdb, rt := 8.1, 94
	r := NewMediaRatingsResponse(shared.RatingStats{}, 0, 0).WithExternal(external.Ratings{IMDbRating: &imdb, IMDbVotes: 10, RTScore: &rt})
	b, _ := json.Marshal(r)
	var m map[string]map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, 10.0, m["imdb"]["scale"])
	assert.Equal(t, 94.0, m["rotten_tomatoes"]["average"])
	assert.Equal(t, 100.0, m["rotten_tomatoes"]["scale"])

	off := NewMediaRatingsResponse(shared.RatingStats{}, 0, 0).WithExternal(external.Ratings{})
	assert.False(t, off.IMDb.Available)
	assert.False(t, off.RottenTomatoes.Available)
}

//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/internal/external_rating_module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExternalRatingGormStoreRoundTrip(t *testing.T) {
	db := newTestDB(t)
	store := external_rating_module.NewGormStore(db)
	ctx := context.Background()

	got, err := store.Find(ctx, "movie", 42)
	require.NoError(t, err)
	assert.Nil(t, got)

	rating, votes, rt := 8.1, 1234567, 94
	require.NoError(t, store.Save(ctx, &external_rating_module.ExternalRating{
		MediaType: "movie", TMDBID: 42, IMDbID: "tt0000042",
		IMDbRating: &rating, IMDbVotes: &votes, RTScore: &rt,
		FetchedAt: time.Now().UTC(),
	}))
	got, err = store.Find(ctx, "movie", 42)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "tt0000042", got.IMDbID)
	assert.InDelta(t, 8.1, *got.IMDbRating, 1e-9)
	assert.Equal(t, 1234567, *got.IMDbVotes)
	assert.Equal(t, 94, *got.RTScore)

	// Upsert on the same (media_type, tmdb_id) updates in place.
	require.NoError(t, store.Save(ctx, &external_rating_module.ExternalRating{
		MediaType: "movie", TMDBID: 42, IMDbID: "tt0000042", NotFound: true,
		FetchedAt: time.Now().UTC(),
	}))
	got, err = store.Find(ctx, "movie", 42)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.True(t, got.NotFound)
	assert.Nil(t, got.IMDbRating)

	var n int64
	require.NoError(t, db.Table("external_ratings").Count(&n).Error)
	assert.EqualValues(t, 1, n)
}

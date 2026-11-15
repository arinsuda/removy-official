package movie_module

import (
	"context"
	"errors"
	"testing"

	"github.com/arinsuda/movie-hub/internal/shared"
	tmdb "github.com/arinsuda/movie-hub/internal/tmdb_module"
	"github.com/stretchr/testify/assert"
)

// stubAwardReader answers AwardsFor from a fixed map, or fails with err.
type stubAwardReader struct {
	awards map[shared.MediaKey][]shared.MediaAward
	err    error
}

func (s stubAwardReader) AwardsFor(context.Context, []shared.MediaIdentity) (map[shared.MediaKey][]shared.MediaAward, error) {
	return s.awards, s.err
}

func TestAwardsForNeverReturnsNil(t *testing.T) {
	ctx := context.Background()
	const id = 42
	won := []shared.MediaAward{{Kind: "monthly", Period: "2026-05"}}

	cases := []struct {
		name   string
		reader shared.AwardReader
		want   []shared.MediaAward
	}{
		{"no reader wired", nil, []shared.MediaAward{}},
		{"reader fails: logged, shown as none", stubAwardReader{err: errors.New("db down")}, []shared.MediaAward{}},
		{"title without awards", stubAwardReader{awards: map[shared.MediaKey][]shared.MediaAward{}}, []shared.MediaAward{}},
		{"title with awards", stubAwardReader{awards: map[shared.MediaKey][]shared.MediaAward{
			{ID: id, Type: shared.MediaTypeMovie}: won,
		}}, won},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewMovieService(nil)
			svc.SetAwardReader(tc.reader)
			got := svc.awardsFor(ctx, id, shared.MediaTypeMovie)
			assert.NotNil(t, got)
			assert.Equal(t, tc.want, got)
		})
	}
}

// countingAwardReader records how often AwardsFor ran and answers from a fixed map.
type countingAwardReader struct {
	calls  *int
	awards map[shared.MediaKey][]shared.MediaAward
	err    error
}

func (c countingAwardReader) AwardsFor(context.Context, []shared.MediaIdentity) (map[shared.MediaKey][]shared.MediaAward, error) {
	*c.calls++
	return c.awards, c.err
}

func TestBatchAwardsReadsOncePerPageAndToleratesErrors(t *testing.T) {
	ctx := context.Background()
	won := []shared.MediaAward{{Kind: "yearly", Period: "2026"}}
	byKey := map[shared.MediaKey][]shared.MediaAward{{ID: 2, Type: shared.MediaTypeTV}: won}

	calls := 0
	svc := NewMovieService(stubRatingStatsReader{})
	svc.SetAwardReader(countingAwardReader{calls: &calls, awards: byKey})
	res, err := svc.enrichSeries(ctx, &tmdb.PaginatedResult[tmdb.TVSeries]{
		Results: []tmdb.TVSeries{{ID: 1}, {ID: 2}, {ID: 3}},
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
	assert.Empty(t, res.Results[0].Awards)
	assert.Equal(t, won, res.Results[1].Awards)

	calls = 0
	svc.SetAwardReader(countingAwardReader{calls: &calls, err: errors.New("db down")})
	res, err = svc.enrichSeries(ctx, &tmdb.PaginatedResult[tmdb.TVSeries]{Results: []tmdb.TVSeries{{ID: 2}}})
	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
	assert.Empty(t, res.Results[0].Awards)

	svc.SetAwardReader(nil)
	res, err = svc.enrichSeries(ctx, &tmdb.PaginatedResult[tmdb.TVSeries]{Results: []tmdb.TVSeries{{ID: 2}}})
	assert.NoError(t, err)
	assert.Empty(t, res.Results[0].Awards)
}

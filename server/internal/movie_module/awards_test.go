package movie_module

import (
	"context"
	"errors"
	"testing"

	"github.com/arinsuda/movie-hub/internal/shared"
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

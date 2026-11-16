package review_module

import (
	"context"
	"errors"
	"testing"

	"github.com/arinsuda/movie-hub/internal/shared"
)

// reviewFakeLookup answers manga lookups for the review routing tests.
type reviewFakeLookup struct{ calls int }

func (f *reviewFakeLookup) LookupMedia(_ context.Context, id int) (shared.MediaInfo, error) {
	f.calls++
	return shared.MediaInfo{Title: "Manga Title", PosterURL: "https://cdn.example/m.jpg"}, nil
}

func TestValidateReviewRequestAcceptsManga(t *testing.T) {
	base := CreateReviewRequest{MediaID: 9, Rating: 4.0}
	for _, mt := range []string{"movie", "tv", "manga"} {
		base.MediaType = mt
		if err := validateReviewRequest(base); err != nil {
			t.Errorf("%s rejected: %v", mt, err)
		}
	}
	for _, mt := range []string{"", "book", "Manga", "anime"} {
		base.MediaType = mt
		if err := validateReviewRequest(base); !errors.Is(err, ErrInvalidMediaType) {
			t.Errorf("%q accepted: %v", mt, err)
		}
	}
}

func TestFetchMediaSummaryRoutesByType(t *testing.T) {
	lookup := &reviewFakeLookup{}
	shared.SetMangaSupport(lookup, nil, nil)
	t.Cleanup(func() { shared.SetMangaSupport(nil, nil, nil) })

	title, poster := fetchMediaSummary(9, "manga")
	if title != "Manga Title" || poster != "https://cdn.example/m.jpg" || lookup.calls != 1 {
		t.Fatalf("manga summary: title=%q poster=%q calls=%d", title, poster, lookup.calls)
	}

	// Regression: an unknown type used to fall through to the TV lookup. It must yield nothing.
	if title, poster := fetchMediaSummary(9, "anime"); title != "" || poster != "" || lookup.calls != 1 {
		t.Fatalf("unknown type must be empty and never reach a provider: title=%q poster=%q calls=%d", title, poster, lookup.calls)
	}
}

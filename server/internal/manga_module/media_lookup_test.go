package manga_module

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/arinsuda/movie-hub/internal/shared"
)

func TestLookupMediaMapsTitleAndAdultFlag(t *testing.T) {
	up := newFakeAniList(t, standardUpstream)
	svc := NewMangaService(up.client())

	sfw, err := svc.LookupMedia(context.Background(), 1)
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if sfw.IsAdult || sfw.Kind != "manhwa" || sfw.ReleaseDate != "2020" {
		t.Fatalf("sfw lookup: %+v", sfw)
	}
	if sfw.Title == "" || sfw.PosterURL != "https://img.example/1.jpg" || sfw.EnglishTitle != "" {
		t.Fatalf("sfw title/poster: %+v", sfw)
	}
	if sfw.StartYear != 2020 || len(sfw.GenreNames) != 1 || sfw.GenreNames[0] != "Action" {
		t.Fatalf("sfw year/genres: %+v", sfw)
	}

	// The lookup asks AniList with eligibility granted, so an adult title still resolves; the
	// viewer decision is shared.CheckMediaAccess, not this function.
	adult, err := svc.LookupMedia(context.Background(), adultTitleID)
	if err != nil {
		t.Fatalf("adult lookup must resolve: %v", err)
	}
	if !adult.IsAdult {
		t.Fatalf("adult flag lost: %+v", adult)
	}
}

func TestLookupMediaErrorMapping(t *testing.T) {
	cases := map[string]struct {
		reply reply
		want  error
	}{
		"no such title": {ok(`{"data":{"Media":null}}`), shared.ErrMediaNotFound},
		"upstream 503":  {reply{status: http.StatusServiceUnavailable, body: `{}`}, shared.ErrMediaUnavailable},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			up := newFakeAniList(t, func(upstreamCall) reply { return tc.reply })
			_, err := NewMangaService(up.client()).LookupMedia(context.Background(), 7)
			if !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
		})
	}
}

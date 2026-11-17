package reading_progress_module

import (
	"context"
	"errors"
	"testing"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/shared"
)

const (
	testMangaSFW   = 30
	testMangaAdult = 31
	testUserID     = 7
)

func mangaTarget(id int) Target {
	return Target{MediaID: id, MediaType: shared.MediaTypeManga}
}

func TestValidateProgress(t *testing.T) {
	cases := []struct {
		name string
		req  UpsertRequest
		want error
	}{
		{"chapter only", UpsertRequest{Chapter: 1}, nil},
		{"volume only", UpsertRequest{Volume: 1}, nil},
		{"chapter at cap", UpsertRequest{Chapter: config.ReadingProgressMaxChapter}, nil},
		{"volume at cap", UpsertRequest{Volume: config.ReadingProgressMaxVolume}, nil},
		{"chapter over cap", UpsertRequest{Chapter: config.ReadingProgressMaxChapter + 1, Volume: 1}, ErrOutOfRange},
		{"volume over cap", UpsertRequest{Chapter: 1, Volume: config.ReadingProgressMaxVolume + 1}, ErrOutOfRange},
		{"negative chapter", UpsertRequest{Chapter: -1, Volume: 1}, ErrOutOfRange},
		{"negative volume", UpsertRequest{Chapter: 1, Volume: -1}, ErrOutOfRange},
		{"both zero", UpsertRequest{}, ErrInvalidInput},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := validateProgress(c.req); !errors.Is(err, c.want) {
				t.Fatalf("got %v want %v", err, c.want)
			}
		})
	}
}

func TestValidateTarget(t *testing.T) {
	if err := validateTarget(mangaTarget(testMangaSFW)); err != nil {
		t.Fatalf("manga rejected: %v", err)
	}
	for _, mt := range []shared.MediaType{shared.MediaTypeMovie, shared.MediaTypeTV, ""} {
		if err := validateTarget(Target{MediaID: testMangaSFW, MediaType: mt}); !errors.Is(err, ErrInvalidMediaType) {
			t.Errorf("media type %q: got %v want ErrInvalidMediaType", mt, err)
		}
	}
	for _, id := range []int{0, -4} {
		if err := validateTarget(mangaTarget(id)); !errors.Is(err, ErrInvalidID) {
			t.Errorf("id %d: got %v want ErrInvalidID", id, err)
		}
	}
}

// stubMedia answers the shared media layer: testMangaAdult is adult, everything else is safe.
type stubMedia struct{}

func (stubMedia) LookupMedia(_ context.Context, id int) (shared.MediaInfo, error) {
	return shared.MediaInfo{Title: "title", IsAdult: id == testMangaAdult}, nil
}

type stubGate struct{ eligible bool }

func (g stubGate) Eligible(_ context.Context, _ uint) (bool, error) {
	return g.eligible, nil
}

// The service is built without a database: every case here is refused before the write.
func TestUpsertRefusesBeforeWrite(t *testing.T) {
	shared.SetMangaSupport(stubMedia{}, stubGate{eligible: false}, nil)
	t.Cleanup(func() { shared.SetMangaSupport(nil, nil, nil) })
	svc := NewService(nil)
	ctx := context.Background()
	in := UpsertRequest{Chapter: 3}

	if _, err := svc.Upsert(ctx, testUserID, Target{MediaID: testMangaSFW, MediaType: shared.MediaTypeMovie}, in); !errors.Is(err, ErrInvalidMediaType) {
		t.Errorf("movie: got %v want ErrInvalidMediaType", err)
	}
	if _, err := svc.Upsert(ctx, testUserID, mangaTarget(testMangaSFW), UpsertRequest{Chapter: config.ReadingProgressMaxChapter + 1}); !errors.Is(err, ErrOutOfRange) {
		t.Errorf("over cap: got %v want ErrOutOfRange", err)
	}
	if _, err := svc.Upsert(ctx, testUserID, mangaTarget(testMangaSFW), UpsertRequest{}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("both zero: got %v want ErrInvalidInput", err)
	}
	if _, err := svc.Upsert(ctx, testUserID, mangaTarget(testMangaAdult), in); !errors.Is(err, shared.ErrAgeRestricted) {
		t.Errorf("adult ineligible: got %v want ErrAgeRestricted", err)
	}
}

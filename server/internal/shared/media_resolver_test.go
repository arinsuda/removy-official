package shared

import (
	"context"
	"errors"
	"sync"
	"testing"
)

// fakeLookup answers AniList-style lookups from a fixed table and counts calls. Calls may arrive
// concurrently (the bounded fallback), so the counter is guarded.
type fakeLookup struct {
	titles map[int]MediaInfo
	calls  int
	err    error
	mu     sync.Mutex
}

func (f *fakeLookup) LookupMedia(_ context.Context, id int) (MediaInfo, error) {
	f.mu.Lock()
	f.calls++
	f.mu.Unlock()
	if f.err != nil {
		return MediaInfo{}, f.err
	}
	info, ok := f.titles[id]
	if !ok {
		return MediaInfo{}, ErrMediaNotFound
	}
	return info, nil
}

// fakeGate grants eligibility to the listed users only.
type fakeGate struct {
	eligible map[uint]bool
	err      error
}

func (g fakeGate) Eligible(_ context.Context, userID uint) (bool, error) {
	return g.eligible[userID], g.err
}

// wire installs a lookup and gate for one test and restores the empty wiring afterwards.
// No database is wired, so manga_meta is always a miss and every lookup goes to the fake.
func wire(t *testing.T, lookup MangaLookup, gate AdultGate) {
	t.Helper()
	SetMangaSupport(lookup, gate, nil)
	t.Cleanup(func() { SetMangaSupport(nil, nil, nil) })
}

const (
	sfwManga   = 30
	adultManga = 31
	movieID    = 603
)

func testLookup() *fakeLookup {
	return &fakeLookup{titles: map[int]MediaInfo{
		sfwManga:   {Title: "Sfw", PosterURL: "https://cdn.example/sfw.jpg", Kind: "manga"},
		adultManga: {Title: "Adult", PosterURL: "https://cdn.example/adult.jpg", Kind: "manhwa", IsAdult: true},
	}}
}

// hidden reports whether KeepVisible drops the single title id of type typ for v.
func hidden(t *testing.T, v MediaVisibility, id int, typ MediaType) bool {
	t.Helper()
	kept, err := KeepVisible(context.Background(), v, []int{id}, func(i int) VisibleRef {
		return VisibleRef{Media: MediaKey{ID: i, Type: typ}}
	})
	if err != nil {
		t.Fatalf("KeepVisible: %v", err)
	}
	return len(kept) == 0
}

func TestIsValidMediaType(t *testing.T) {
	for _, ok := range []MediaType{MediaTypeMovie, MediaTypeTV, MediaTypeManga} {
		if !IsValidMediaType(ok) {
			t.Errorf("%q should be valid", ok)
		}
	}
	for _, bad := range []MediaType{"", "book", "Manga", "movies", "series", "anime"} {
		if IsValidMediaType(bad) {
			t.Errorf("%q should be rejected", bad)
		}
	}
}

func TestResolveRoutesMangaToLookup(t *testing.T) {
	lookup := testLookup()
	wire(t, lookup, nil)

	info, err := ResolveMedia(context.Background(), sfwManga, MediaTypeManga)
	if err != nil {
		t.Fatalf("manga lookup: %v", err)
	}
	if info.Title != "Sfw" || lookup.calls != 1 {
		t.Fatalf("manga routed wrongly: title=%q calls=%d", info.Title, lookup.calls)
	}
}

func TestResolveUnknownTypeIsErrorNeverTV(t *testing.T) {
	lookup := testLookup()
	wire(t, lookup, nil)

	if _, err := ResolveMedia(context.Background(), sfwManga, "anime"); !errors.Is(err, ErrUnknownMediaType) {
		t.Fatalf("unknown type: want ErrUnknownMediaType, got %v", err)
	}
	if lookup.calls != 0 {
		t.Fatalf("unknown type must not reach any provider, calls=%d", lookup.calls)
	}
}

func TestResolveMangaWithoutLookupIsUnavailable(t *testing.T) {
	wire(t, nil, nil)
	if _, err := ResolveMedia(context.Background(), sfwManga, MediaTypeManga); !errors.Is(err, ErrMediaUnavailable) {
		t.Fatalf("want ErrMediaUnavailable, got %v", err)
	}
}

func TestFetchMediaSummaryManga(t *testing.T) {
	wire(t, testLookup(), nil)

	got := FetchMediaSummary(sfwManga, MediaTypeManga)
	if got.ID != sfwManga || got.MediaType != MediaTypeManga || got.Title != "Sfw" || got.Kind != "manga" {
		t.Fatalf("manga summary: %+v", got)
	}
	if got.PosterURL != "https://cdn.example/sfw.jpg" {
		t.Fatalf("manga poster must stay absolute, got %q", got.PosterURL)
	}
	if got.VoteAverage != 0 || got.Genres != nil {
		t.Fatalf("manga must not carry TMDB fields: %+v", got)
	}
}

func TestMediaPosterURL(t *testing.T) {
	cases := map[string]string{
		"":                          "",
		"https://cdn.example/a.jpg": "https://cdn.example/a.jpg",
		"http://cdn.example/a.jpg":  "http://cdn.example/a.jpg",
		"/abc123.jpg":               "https://image.tmdb.org/t/p/w500/abc123.jpg",
	}
	for in, want := range cases {
		if got := MediaPosterURL(in); got != want {
			t.Errorf("MediaPosterURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCheckMediaAccess(t *testing.T) {
	const adult, child, guest uint = 1, 2, 0
	gate := fakeGate{eligible: map[uint]bool{adult: true}}
	lookup := testLookup()
	wire(t, lookup, gate)
	ctx := context.Background()

	cases := []struct {
		name string
		user uint
		id   int
		typ  MediaType
		want error
	}{
		{"eligible adult title", adult, adultManga, MediaTypeManga, nil},
		{"adult title, minor", child, adultManga, MediaTypeManga, ErrAgeRestricted},
		{"adult title, guest", guest, adultManga, MediaTypeManga, ErrAgeRestricted},
		{"sfw title, guest", guest, sfwManga, MediaTypeManga, nil},
		{"unknown manga id", adult, 999, MediaTypeManga, ErrMediaNotFound},
		{"movie is never gated", guest, movieID, MediaTypeMovie, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lookup.calls = 0
			err := CheckMediaAccess(ctx, tc.user, tc.id, tc.typ)
			if !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
			if tc.typ != MediaTypeManga && lookup.calls != 0 {
				t.Fatalf("non-manga must not call AniList, calls=%d", lookup.calls)
			}
		})
	}
}

func TestCheckMediaAccessFailsClosed(t *testing.T) {
	ctx := context.Background()

	wire(t, &fakeLookup{err: ErrMediaUnavailable}, fakeGate{eligible: map[uint]bool{1: true}})
	if err := CheckMediaAccess(ctx, 1, adultManga, MediaTypeManga); !errors.Is(err, ErrMediaUnavailable) {
		t.Fatalf("unresolved title must be refused, got %v", err)
	}

	gateErr := errors.New("db down")
	wire(t, testLookup(), fakeGate{err: gateErr})
	if err := CheckMediaAccess(ctx, 1, adultManga, MediaTypeManga); !errors.Is(err, gateErr) {
		t.Fatalf("gate error must surface, got %v", err)
	}
}

func TestMediaVisibility(t *testing.T) {
	ctx := context.Background()
	wire(t, testLookup(), fakeGate{eligible: map[uint]bool{7: true}})

	eligible, err := NewMediaVisibility(ctx, 7)
	if err != nil {
		t.Fatal(err)
	}
	guest, err := NewMediaVisibility(ctx, 0)
	if err != nil {
		t.Fatal(err)
	}

	if hidden(t, eligible, adultManga, MediaTypeManga) {
		t.Error("eligible viewer must see adult manga")
	}
	if !hidden(t, guest, adultManga, MediaTypeManga) {
		t.Error("guest must not see adult manga")
	}
	if hidden(t, guest, sfwManga, MediaTypeManga) {
		t.Error("guest must see sfw manga")
	}
	if hidden(t, guest, movieID, MediaTypeMovie) {
		t.Error("movies are never hidden by age")
	}
	if !hidden(t, guest, 999, MediaTypeManga) {
		t.Error("unresolvable manga must be hidden for an ineligible viewer")
	}

	var zero MediaVisibility
	if !hidden(t, zero, adultManga, MediaTypeManga) {
		t.Error("zero value must fail closed")
	}
}

func TestKeepVisibleDropsAdultManga(t *testing.T) {
	ctx := context.Background()
	wire(t, testLookup(), fakeGate{})
	guest, err := NewMediaVisibility(ctx, 0)
	if err != nil {
		t.Fatal(err)
	}
	type row struct {
		ID  int
		Typ MediaType
	}
	in := []row{{sfwManga, MediaTypeManga}, {adultManga, MediaTypeManga}, {movieID, MediaTypeMovie}, {sfwManga, MediaTypeMovie}}
	out, err := KeepVisible(ctx, guest, in, func(r row) VisibleRef {
		return VisibleRef{Media: MediaKey{ID: r.ID, Type: r.Typ}}
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []row{{sfwManga, MediaTypeManga}, {movieID, MediaTypeMovie}, {sfwManga, MediaTypeMovie}}
	if len(out) != len(want) {
		t.Fatalf("got %v, want %v", out, want)
	}
	for i := range want {
		if out[i] != want[i] {
			t.Fatalf("got %v, want %v", out, want)
		}
	}
	if len(in) != 4 {
		t.Fatal("input must not be modified")
	}
}

func TestKeepVisibleNeverHidesOwnItems(t *testing.T) {
	ctx := context.Background()
	wire(t, testLookup(), fakeGate{})
	const owner uint = 5
	self, err := NewMediaVisibility(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	other, err := NewMediaVisibility(ctx, owner+1)
	if err != nil {
		t.Fatal(err)
	}

	own := []VisibleRef{{Media: MediaKey{ID: adultManga, Type: MediaTypeManga}, OwnerID: owner}}
	if out, _ := KeepVisible(ctx, self, own, func(r VisibleRef) VisibleRef { return r }); len(out) != 1 {
		t.Error("owner must keep their own adult title")
	}
	if out, _ := KeepVisible(ctx, other, own, func(r VisibleRef) VisibleRef { return r }); len(out) != 0 {
		t.Error("another viewer must not see the owner's adult title")
	}
	// OwnerID 0 (not one user's item, e.g. a notification) never gets the owner exemption.
	anonymous := []VisibleRef{{Media: MediaKey{ID: adultManga, Type: MediaTypeManga}}}
	if out, _ := KeepVisible(ctx, self, anonymous, func(r VisibleRef) VisibleRef { return r }); len(out) != 0 {
		t.Error("an item without an owner must be filtered for an ineligible viewer")
	}
}

func TestRestrictedFollowsOwnerAndEligibility(t *testing.T) {
	ctx := context.Background()
	wire(t, testLookup(), fakeGate{eligible: map[uint]bool{9: true}})
	guest, _ := NewMediaVisibility(ctx, 0)
	member, _ := NewMediaVisibility(ctx, 3)
	adult, _ := NewMediaVisibility(ctx, 9)

	if !guest.Restricted(3) || !member.Restricted(4) {
		t.Error("ineligible viewers must be restricted outside their own data")
	}
	if member.Restricted(3) {
		t.Error("the owner must never be restricted")
	}
	if adult.Restricted(4) || !adult.ShowsAdult() {
		t.Error("eligible viewers must not be restricted")
	}
}

func TestReceivableMediaAdultNeedsEligibility(t *testing.T) {
	ctx := context.Background()
	wire(t, testLookup(), fakeGate{eligible: map[uint]bool{9: true}})
	sfw := MediaKey{ID: sfwManga, Type: MediaTypeManga}
	adult := MediaKey{ID: adultManga, Type: MediaTypeManga}
	unknown := MediaKey{ID: 999, Type: MediaTypeManga}
	movie := MediaKey{ID: movieID, Type: MediaTypeMovie}

	guest, err := ReceivableMedia(ctx, 0, []MediaKey{sfw, movie, adult})
	if err != nil {
		t.Fatal(err)
	}
	if !guest[sfw] || !guest[movie] {
		t.Error("sfw manga and movies reach everyone")
	}
	if guest[adult] {
		t.Error("adult manga must not fan out to guests")
	}

	minor, err := ReceivableMedia(ctx, 3, []MediaKey{adult})
	if err != nil {
		t.Fatal(err)
	}
	if minor[adult] {
		t.Error("ineligible recipient must not get adult manga")
	}

	adultUser, err := ReceivableMedia(ctx, 9, []MediaKey{adult, unknown})
	if err != nil {
		t.Fatal(err)
	}
	if !adultUser[adult] {
		t.Error("eligible recipient gets adult manga")
	}
	if !adultUser[unknown] {
		t.Error("an eligible recipient may get a title that cannot be resolved (treated as adult)")
	}

	guestUnknown, err := ReceivableMedia(ctx, 0, []MediaKey{unknown})
	if err != nil {
		t.Fatal(err)
	}
	if guestUnknown[unknown] {
		t.Error("unresolvable manga must not reach an ineligible user")
	}
}

func TestMangaSafeSQLShape(t *testing.T) {
	got := MangaSafeSQL("li.media_type", "li.media_id")
	want := "(COALESCE(li.media_type, '') <> 'manga' OR li.media_id IN (SELECT media_id FROM manga_meta WHERE is_adult = false))"
	if got != want {
		t.Fatalf("MangaSafeSQL:\n got %s\nwant %s", got, want)
	}
}

func TestMangaRowsFallbackIsBounded(t *testing.T) {
	ctx := context.Background()
	titles := map[int]MediaInfo{}
	var ids []int
	for i := 1; i <= MangaMetaFallbackLimit+3; i++ {
		titles[i] = MediaInfo{Title: "t"}
		ids = append(ids, i)
	}
	lookup := &fakeLookup{titles: titles}
	wire(t, lookup, nil)

	rows, err := MangaRows(ctx, ids)
	if err != nil {
		t.Fatal(err)
	}
	if lookup.calls != MangaMetaFallbackLimit {
		t.Fatalf("fallback lookups = %d, want %d", lookup.calls, MangaMetaFallbackLimit)
	}
	if len(rows) != MangaMetaFallbackLimit {
		t.Fatalf("rows = %d, want %d; titles past the cap must stay unknown", len(rows), MangaMetaFallbackLimit)
	}
}

//go:build integration

package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/arinsuda/movie-hub/internal/shared"
)

// AniList ids used here. Manga ids share the number space with TMDB movie ids, so collideID is also a
// movie id in other fixtures.
const (
	mangaSFW     = 30
	mangaAdult   = 31
	mangaMissing = 999
	collideID    = 603
	// guestCapRows fills past library_module.GuestLibraryLimit (100) so the guest page is cut.
	guestCapRows = 101
)

// fakeAniList answers the shared media layer: every id is a safe title except mangaAdult (adult) and
// mangaMissing (no such title).
type fakeAniList struct{}

func (fakeAniList) LookupMedia(_ context.Context, id int) (shared.MediaInfo, error) {
	switch id {
	case mangaMissing:
		return shared.MediaInfo{}, shared.ErrMediaNotFound
	case mangaAdult:
		return shared.MediaInfo{Title: "Adult", IsAdult: true}, nil
	}
	return shared.MediaInfo{Title: fmt.Sprintf("Manga %d", id), PosterURL: "https://cdn.example/p.jpg"}, nil
}

// adultGate lists the users allowed to see adult titles.
type adultGate map[uint]bool

func (g adultGate) Eligible(_ context.Context, id uint) (bool, error) { return g[id], nil }

// mangaEnv builds the test app with the fake AniList and an age gate. newEnv wires modules directly,
// not through router.Register, so the shared media layer is installed here.
func mangaEnv(t *testing.T) (*env, adultGate) {
	t.Helper()
	e := newEnv(t)
	gate := adultGate{}
	shared.SetMangaSupport(fakeAniList{}, gate, e.db)
	t.Cleanup(func() { shared.SetMangaSupport(nil, nil, nil) })
	return e, gate
}

// mediaIDs returns media.id of every item under key in a list response.
func mediaIDs(t *testing.T, out map[string]any, key string) []int {
	t.Helper()
	raw, ok := out[key].([]any)
	if !ok {
		t.Fatalf("response has no %q list: %v", key, out)
	}
	ids := make([]int, 0, len(raw))
	for _, it := range raw {
		media, _ := it.(map[string]any)["media"].(map[string]any)
		ids = append(ids, int(media["id"].(float64)))
	}
	return ids
}

func containsInt(xs []int, want int) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

func TestMangaIDsDoNotCollideWithMovies(t *testing.T) {
	e, _ := mangaEnv(t)
	uid := seedUser(t, e.db, "collide@x.test", true, "collide-password1")
	tok := e.loginPair("collide@x.test", "collide-password1").AccessToken

	for _, typ := range []string{"movie", "manga"} {
		if code, out := e.do("POST", "/api/reviews", map[string]any{"media_id": collideID, "media_type": typ, "rating": 4.0, "body": typ}, tok); code != 201 {
			t.Fatalf("review %s: %d %v", typ, code, out)
		}
		if code, out := e.do("POST", "/api/library", map[string]any{"media_id": collideID, "media_type": typ, "list_type": "watchlist"}, tok); code != 201 {
			t.Fatalf("library %s: %d %v", typ, code, out)
		}
		if code, out := e.do("POST", fmt.Sprintf("/api/stats/%s/%d/like", typ, collideID), nil, tok); code != 204 {
			t.Fatalf("like %s: %d %v", typ, code, out)
		}
	}
	for _, table := range []string{"reviews", "library_items", "media_likes"} {
		if n := e.count(fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE user_id = ? AND media_id = ?`, table), uid, collideID); n != 2 {
			t.Errorf("%s: want one row per media type for the same id, got %d", table, n)
		}
	}
}

// mangaWrite is one write a viewer may attempt against an adult title.
type mangaWrite struct {
	method, path string
	body         map[string]any
}

func adultWrites(listID int) map[string]mangaWrite {
	return map[string]mangaWrite{
		"review":    {"POST", "/api/reviews", map[string]any{"media_id": mangaAdult, "media_type": "manga", "rating": 3.0, "body": "x"}},
		"library":   {"POST", "/api/library", map[string]any{"media_id": mangaAdult, "media_type": "manga", "list_type": "watchlist"}},
		"like":      {"POST", fmt.Sprintf("/api/stats/manga/%d/like", mangaAdult), nil},
		"view":      {"POST", fmt.Sprintf("/api/stats/manga/%d/view", mangaAdult), nil},
		"bmol":      {"POST", "/api/bmol", map[string]any{"media_id": mangaAdult, "media_type": "manga", "rank": 1}},
		"list item": {"POST", fmt.Sprintf("/api/lists/%d/items", listID), map[string]any{"media_id": mangaAdult, "media_type": "manga"}},
	}
}

func TestAdultMangaWritesNeedEligibility(t *testing.T) {
	e, gate := mangaEnv(t)
	seedUser(t, e.db, "minor@x.test", true, "minor-password1")
	adult := seedUser(t, e.db, "adult@x.test", true, "adult-password1")
	gate[adult] = true
	minorTok := e.loginPair("minor@x.test", "minor-password1").AccessToken
	adultTok := e.loginPair("adult@x.test", "adult-password1").AccessToken

	_, minorList := e.do("POST", "/api/lists", map[string]any{"name": "Shelf"}, minorTok)
	minorListID := int(minorList["list"].(map[string]any)["id"].(float64))
	_, adultList := e.do("POST", "/api/lists", map[string]any{"name": "Shelf"}, adultTok)
	adultListID := int(adultList["list"].(map[string]any)["id"].(float64))

	for name, w := range adultWrites(minorListID) {
		code, out := e.do(w.method, w.path, w.body, minorTok)
		if code != 403 || out["code"] != string(shared.ErrorCodeAgeRestricted) {
			t.Errorf("minor %s: want 403 AGE_RESTRICTED, got %d %v", name, code, out)
		}
	}
	if n := e.count(`SELECT COUNT(*) FROM reviews WHERE user_id = (SELECT id FROM users WHERE email = 'minor@x.test')`); n != 0 {
		t.Errorf("refused review was stored: %d rows", n)
	}

	for name, w := range adultWrites(adultListID) {
		if code, out := e.do(w.method, w.path, w.body, adultTok); code >= 300 {
			t.Errorf("eligible %s: got %d %v", name, code, out)
		}
	}

	// A safe title stays open to the minor, and an id AniList does not know is refused for everyone.
	if code, out := e.do("POST", "/api/library", map[string]any{"media_id": mangaSFW, "media_type": "manga", "list_type": "watchlist"}, minorTok); code != 201 {
		t.Errorf("minor sfw library: %d %v", code, out)
	}
	code, out := e.do("POST", "/api/library", map[string]any{"media_id": mangaMissing, "media_type": "manga", "list_type": "watchlist"}, adultTok)
	if code != 404 || out["code"] != string(shared.ErrorCodeNotFound) {
		t.Errorf("unknown manga: want 404 NOT_FOUND, got %d %v", code, out)
	}
}

func TestAdultMangaHiddenFromOtherViewers(t *testing.T) {
	e, gate := mangaEnv(t)
	owner := seedUser(t, e.db, "owner@x.test", true, "owner-password1")
	gate[owner] = true
	seedUser(t, e.db, "minor@x.test", true, "minor-password1")
	ownerTok := e.loginPair("owner@x.test", "owner-password1").AccessToken
	minorTok := e.loginPair("minor@x.test", "minor-password1").AccessToken

	_, listOut := e.do("POST", "/api/lists", map[string]any{"name": "Shelf"}, ownerTok)
	listID := int(listOut["list"].(map[string]any)["id"].(float64))
	for rank, id := range []int{mangaSFW, mangaAdult} {
		if code, out := e.do("POST", "/api/library", map[string]any{"media_id": id, "media_type": "manga", "list_type": "watchlist"}, ownerTok); code != 201 {
			t.Fatalf("library %d: %d %v", id, code, out)
		}
		if code, out := e.do("POST", fmt.Sprintf("/api/stats/manga/%d/like", id), nil, ownerTok); code != 204 {
			t.Fatalf("like %d: %d %v", id, code, out)
		}
		if code, out := e.do("POST", "/api/bmol", map[string]any{"media_id": id, "media_type": "manga", "rank": rank + 1}, ownerTok); code != 201 {
			t.Fatalf("bmol %d: %d %v", id, code, out)
		}
		if code, out := e.do("POST", "/api/reviews", map[string]any{"media_id": id, "media_type": "manga", "rating": 4.0, "body": "r"}, ownerTok); code != 201 {
			t.Fatalf("review %d: %d %v", id, code, out)
		}
		if code, out := e.do("POST", fmt.Sprintf("/api/lists/%d/items", listID), map[string]any{"media_id": id, "media_type": "manga"}, ownerTok); code != 201 {
			t.Fatalf("list item %d: %d %v", id, code, out)
		}
	}

	// guest is false for routes that need a session (lists are not guest-readable).
	reads := []struct {
		path, key string
		guest     bool
	}{
		{fmt.Sprintf("/api/library/user/%d", owner), "items", true},
		{fmt.Sprintf("/api/users/%d/likes", owner), "likes", true},
		{fmt.Sprintf("/api/bmol/user/%d", owner), "items", true},
		{fmt.Sprintf("/api/reviews/user/%d", owner), "reviews", true},
		{fmt.Sprintf("/api/lists/%d", listID), "items", false},
	}
	for _, viewer := range []struct{ name, tok string }{{"guest", ""}, {"minor", minorTok}} {
		for _, r := range reads {
			if viewer.tok == "" && !r.guest {
				continue
			}
			code, out := e.do("GET", r.path, nil, viewer.tok)
			if code != 200 {
				t.Fatalf("%s %s: %d %v", viewer.name, r.path, code, out)
			}
			if got := mediaIDs(t, out, r.key); len(got) != 1 || got[0] != mangaSFW {
				t.Errorf("%s %s: want only the safe title, got %v", viewer.name, r.path, got)
			}
		}
	}
	for _, r := range reads {
		code, out := e.do("GET", r.path, nil, ownerTok)
		if code != 200 {
			t.Fatalf("owner %s: %d %v", r.path, code, out)
		}
		if got := mediaIDs(t, out, r.key); !containsInt(got, mangaAdult) {
			t.Errorf("owner %s must still see the adult title, got %v", r.path, got)
		}
	}
}

func TestGuestCapHidesAdultTitles(t *testing.T) {
	e, gate := mangaEnv(t)
	owner := seedUser(t, e.db, "capped@x.test", true, "capped-password1")
	gate[owner] = true
	ownerTok := e.loginPair("capped@x.test", "capped-password1").AccessToken

	// Bulk rows share one timestamp, so the adult row added through the API is the newest row.
	// Every stored manga has a manga_meta row (the write path stores it), so the bulk titles get one too.
	if err := e.db.Exec(`INSERT INTO manga_meta (media_id, title, cover_url, kind, is_adult, updated_at)
		SELECT g, 'Bulk', '', 'manga', false, now() FROM generate_series(1000, ?) g`, 1000+guestCapRows-1).Error; err != nil {
		t.Fatal(err)
	}
	if err := e.db.Exec(`INSERT INTO library_items (created_at, updated_at, user_id, media_id, media_type, list_type)
		SELECT now(), now(), ?, g, 'manga', 'watchlist' FROM generate_series(1000, ?) g`, owner, 1000+guestCapRows-1).Error; err != nil {
		t.Fatal(err)
	}
	if code, out := e.do("POST", "/api/library", map[string]any{"media_id": mangaAdult, "media_type": "manga", "list_type": "watchlist"}, ownerTok); code != 201 {
		t.Fatalf("adult library add: %d %v", code, out)
	}

	code, out := e.do("GET", fmt.Sprintf("/api/library/user/%d", owner), nil, "")
	if code != 200 {
		t.Fatalf("guest library: %d %v", code, out)
	}
	// The guest page holds the newest 100 rows: the adult row plus 99 safe ones.
	got := mediaIDs(t, out, "items")
	if len(got) != 99 {
		t.Errorf("guest page: want 99 items after hiding the adult row, got %d", len(got))
	}
	if containsInt(got, mangaAdult) {
		t.Error("adult title leaked to a guest")
	}
	if more, _ := out["has_more"].(bool); !more {
		t.Error("guest cap must still report has_more")
	}
}

func TestPrivateProfileHidesMangaSections(t *testing.T) {
	e, gate := mangaEnv(t)
	owner := seedUser(t, e.db, "private@x.test", true, "private-password1")
	gate[owner] = true
	seedUser(t, e.db, "stranger@x.test", true, "stranger-password1")
	ownerTok := e.loginPair("private@x.test", "private-password1").AccessToken
	strangerTok := e.loginPair("stranger@x.test", "stranger-password1").AccessToken

	if code, out := e.do("POST", "/api/library", map[string]any{"media_id": mangaSFW, "media_type": "manga", "list_type": "watchlist"}, ownerTok); code != 201 {
		t.Fatalf("library: %d %v", code, out)
	}
	if code, out := e.do("POST", "/api/bmol", map[string]any{"media_id": mangaSFW, "media_type": "manga", "rank": 1}, ownerTok); code != 201 {
		t.Fatalf("bmol: %d %v", code, out)
	}
	if err := e.db.Exec(`UPDATE users SET is_private = true WHERE id = ?`, owner).Error; err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{fmt.Sprintf("/api/library/user/%d", owner), fmt.Sprintf("/api/bmol/user/%d", owner)} {
		if code, out := e.do("GET", path, nil, strangerTok); code != 403 {
			t.Errorf("stranger %s on a private profile: want 403, got %d %v", path, code, out)
		}
		if code, out := e.do("GET", path, nil, ownerTok); code != 200 {
			t.Errorf("owner %s: %d %v", path, code, out)
		}
	}
}

func TestUserListItemsAcceptManga(t *testing.T) {
	e, _ := mangaEnv(t)
	seedUser(t, e.db, "lister@x.test", true, "lister-password1")
	tok := e.loginPair("lister@x.test", "lister-password1").AccessToken

	_, listOut := e.do("POST", "/api/lists", map[string]any{"name": "Manga"}, tok)
	listID := int(listOut["list"].(map[string]any)["id"].(float64))
	if code, out := e.do("POST", fmt.Sprintf("/api/lists/%d/items", listID), map[string]any{"media_id": mangaSFW, "media_type": "manga"}, tok); code != 201 {
		t.Fatalf("manga list item: %d %v", code, out)
	}
	if code, out := e.do("POST", fmt.Sprintf("/api/lists/%d/items", listID), map[string]any{"media_id": mangaSFW, "media_type": "anime"}, tok); code != 400 {
		t.Fatalf("anime list item: want 400, got %d %v", code, out)
	}

	// The database constraint itself must allow the three media types. Migration 20261118 widens it.
	var def string
	if err := e.db.Raw(`SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conname = 'chk_user_list_items_media_type'`).Scan(&def).Error; err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"'movie'", "'tv'", "'manga'"} {
		if !strings.Contains(def, want) {
			t.Errorf("constraint %q lacks %s", def, want)
		}
	}
}

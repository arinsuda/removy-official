//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/config"
	tmdbmodule "github.com/arinsuda/movie-hub/internal/tmdb_module"
	"github.com/arinsuda/movie-hub/router"
	"github.com/gofiber/fiber/v3"
)

// useFakeTMDB points the TMDB client at a local server that answers every path with minimal valid JSON.
func useFakeTMDB(t *testing.T) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/watch/providers") {
			_, _ = w.Write([]byte(`{"id":1,"results":{}}`))
			return
		}
		_, _ = w.Write([]byte(`{"id":603,"title":"Fake Movie","name":"Fake Show","overview":"o","poster_path":"/p.jpg","genres":[],"results":[],"cast":[],"crew":[],"page":1,"total_pages":1,"total_results":0}`))
	}))
	t.Cleanup(srv.Close)
	tmdbmodule.Init(&config.Config{TMDB: config.TMDBConfig{BaseURL: srv.URL, APIKey: "test"}})
}

var guestPublicPaths = []string{
	"/api/auth/session",
	"/api/movies/popular", "/api/movies/now-playing", "/api/movies/top-rated",
	"/api/movies/upcoming", "/api/movies/search?q=dune", "/api/movies/genres",
	"/api/movies/603", "/api/movies/603/similar", "/api/movies/603/watch-providers",
	"/api/tv/popular", "/api/tv/now-airing", "/api/tv/search?q=dune", "/api/tv/genres",
	"/api/tv/1399", "/api/tv/1399/similar", "/api/tv/1399/watch-providers",
	"/api/actors/100", "/api/actors/100/movies", "/api/actors/100/tv",
	"/api/reviews/media/movies/603", "/api/reviews/media/movies/603/rating",
	"/api/stats/movie/603",
}

func TestGuestCanReadPublicRoutes(t *testing.T) {
	useFakeTMDB(t)
	e := newEnv(t)
	for _, p := range guestPublicPaths {
		if code, out := e.do("GET", p, nil, ""); code != 200 {
			t.Errorf("guest GET %s: %d %v", p, code, out)
		}
	}
	if code, out := e.do("GET", "/api/auth/session", nil, ""); code != 200 || out["has_session"] != false {
		t.Fatalf("guest GET /api/auth/session: %d %v, want has_session=false", code, out)
	}
	// A signed-in user keeps using the same routes.
	seedUser(t, e.db, "me@x.test", true, "my-password1")
	pair := e.loginPair("me@x.test", "my-password1")
	tok := pair.AccessToken
	for _, p := range guestPublicPaths {
		if code, out := e.do("GET", p, nil, tok); code != 200 {
			t.Errorf("user GET %s: %d %v", p, code, out)
		}
	}
}

func TestGuestIsRejectedFromEverythingElse(t *testing.T) {
	useFakeTMDB(t)
	e := newEnv(t)
	uid := seedUser(t, e.db, "me@x.test", true, "my-password1")
	tok := e.loginPair("me@x.test", "my-password1").AccessToken

	for _, c := range []struct{ method, path string }{
		{"GET", fmt.Sprintf("/api/library/user/%d", uid)},
		{"GET", "/api/lists"},
		{"POST", "/api/lists"},
		{"POST", "/api/stats/movie/603/view"},
		{"GET", "/api/movies/recommended"},
		{"GET", "/api/movies/upcoming/2026"},
		{"GET", "/api/tv/top-rated"},
		{"GET", "/api/actors/search?q=tom"},
		{"GET", "/api/feed"},
		{"GET", "/api/notifications"},
		{"GET", "/api/achievements"},
		{"GET", fmt.Sprintf("/api/reviews/user/%d", uid)},
		{"POST", "/api/reviews"},
		{"GET", fmt.Sprintf("/api/users/%d", uid)},
		{"GET", "/api/ping"},
	} {
		if code, _ := e.do(c.method, c.path, nil, ""); code != 401 {
			t.Errorf("guest %s %s: %d, want 401", c.method, c.path, code)
		}
	}
	// Protected routes under the same prefixes work when signed in and are not shadowed by public :id<int> routes.
	if code, out := e.do("GET", "/api/actors/search?q=tom", nil, tok); code != 200 {
		t.Errorf("user GET /actors/search: %d %v", code, out)
	}
	if code, out := e.do("GET", "/api/movies/recommended", nil, tok); code != 200 {
		t.Errorf("user GET /movies/recommended: %d %v", code, out)
	}
	if code, out := e.do("POST", "/api/stats/movie/603/view", nil, tok); code != 204 {
		t.Errorf("user POST view: %d %v", code, out)
	}
}

// /api/achievements is registered on the root router AFTER the auth group, so the group's USE
// already ran: it is auth-gated in practice (same order as router.Register).
func TestAchievementsListIsAuthGatedToday(t *testing.T) {
	e := newEnv(t)
	if code, _ := e.do("GET", "/api/achievements", nil, ""); code != 401 {
		t.Fatalf("guest GET /achievements: %d, want 401 (documented as auth-gated)", code)
	}
}

func TestGuestOnlySeesPublicReviewsOfOpenAccounts(t *testing.T) {
	e := newEnv(t)
	insert := func(email, visibility string, isPublic, privateAccount bool) uint {
		uid := seedUser(t, e.db, email, true, "pw-password1")
		if privateAccount {
			e.db.Exec(`UPDATE users SET is_private = true WHERE id = ?`, uid)
		}
		if err := e.db.Exec(`INSERT INTO reviews (created_at, updated_at, user_id, media_id, media_type, rating, body, is_public, visibility)
			VALUES (now(), now(), ?, 603, 'movie', 4.0, ?, ?, ?)`, uid, email, isPublic, visibility).Error; err != nil {
			t.Fatal(err)
		}
		return uid
	}
	insert("public@x.test", "public", true, false)
	insert("private@x.test", "private", false, false)
	insert("legacy@x.test", "private", true, false) // flags disagree: a guest must not see it
	lockedOwner := insert("locked@x.test", "public", true, true)

	bodies := func(token string) map[string]bool {
		code, out := e.do("GET", "/api/reviews/media/movies/603", nil, token)
		if code != 200 {
			t.Fatalf("GET reviews: %d %v", code, out)
		}
		got := map[string]bool{}
		for _, r := range out["reviews"].([]any) {
			got[r.(map[string]any)["body"].(string)] = true
		}
		return got
	}

	guest := bodies("")
	if len(guest) != 1 || !guest["public@x.test"] {
		t.Fatalf("guest sees %v, want only the public review of an open account", guest)
	}

	// An accepted follower of the locked account sees its review; a guest and a stranger do not.
	follower := seedUser(t, e.db, "f@x.test", true, "pw-password1")
	stranger := seedUser(t, e.db, "s@x.test", true, "pw-password1")
	_ = stranger
	e.db.Exec(`INSERT INTO user_follows (follower_id, followee_id, status, created_at, updated_at) VALUES (?, ?, 'accepted', now(), now())`, follower, lockedOwner)
	followerTok := e.loginPair("f@x.test", "pw-password1").AccessToken
	if got := bodies(followerTok); !got["locked@x.test"] || !got["public@x.test"] {
		t.Fatalf("follower sees %v", got)
	}
	if got := bodies(e.loginPair("s@x.test", "pw-password1").AccessToken); got["locked@x.test"] || got["private@x.test"] {
		t.Fatalf("stranger sees %v", got)
	}
}

// A token that is present but stale or malformed is 401 even on a public route, so the client
// refreshes instead of silently browsing as a guest.
func TestStaleOrInvalidTokenOnPublicRouteIs401(t *testing.T) {
	useFakeTMDB(t)
	e := newEnv(t)
	uid := seedUser(t, e.db, "me@x.test", true, "my-password1")
	old := e.loginPair("me@x.test", "my-password1").AccessToken
	if code, _ := e.do("GET", "/api/movies/603", nil, old); code != 200 {
		t.Fatalf("valid token: %d", code)
	}
	e.db.Exec(`UPDATE users SET token_version = token_version + 1 WHERE id = ?`, uid)
	for _, p := range []string{"/api/movies/603", "/api/movies/popular", "/api/movies/upcoming", "/api/tv/search?q=dune", "/api/actors/100"} {
		if code, _ := e.do("GET", p, nil, old); code != 401 {
			t.Fatalf("stale token on %s: %d, want 401", p, code)
		}
		if code, _ := e.do("GET", p, nil, "not-a-jwt"); code != 401 {
			t.Fatalf("garbage token on %s: %d, want 401", p, code)
		}
	}
}

func TestGuestReadsAreRateLimitedPerIP(t *testing.T) {
	e := newEnv(t)
	seedUser(t, e.db, "me@x.test", true, "my-password1")
	tok := e.loginPair("me@x.test", "my-password1").AccessToken

	var last int
	for i := 0; i < router.PublicReadRateMax+1; i++ {
		last, _ = e.do("GET", "/api/stats/movie/603", nil, "")
	}
	if last != 429 {
		t.Fatalf("request %d: %d, want 429", router.PublicReadRateMax+1, last)
	}
	if code, out := e.do("GET", "/api/stats/movie/603", nil, ""); code != 429 || out["code"] != "RATE_LIMITED" {
		t.Fatalf("still limited: %d %v", code, out)
	}
	if code, _ := e.do("GET", "/api/stats/movie/603", nil, tok); code != 200 {
		t.Fatalf("signed-in user shares the guest budget: %d", code)
	}
}

func TestMediaReviewsArePaginatedAndKeepPrivacy(t *testing.T) {
	e := newEnv(t)
	const total = 25
	for i := 0; i < total; i++ {
		uid := seedUser(t, e.db, fmt.Sprintf("r%d@x.test", i), true, "pw-password1")
		e.db.Exec(`INSERT INTO reviews (created_at, updated_at, user_id, media_id, media_type, rating, body, is_public, visibility)
			VALUES (now() - make_interval(secs => ?), now(), ?, 77, 'movie', 3.0, 'b', true, 'public')`, i, uid)
	}
	locked := seedUser(t, e.db, "locked@x.test", true, "pw-password1")
	e.db.Exec(`UPDATE users SET is_private = true WHERE id = ?`, locked)
	e.db.Exec(`INSERT INTO reviews (created_at, updated_at, user_id, media_id, media_type, rating, body, is_public, visibility)
		VALUES (now(), now(), ?, 77, 'movie', 3.0, 'b', true, 'public')`, locked)

	page := func(q string) (int, []any, map[string]any) {
		code, out := e.do("GET", "/api/reviews/media/movies/77"+q, nil, "")
		revs, _ := out["reviews"].([]any)
		pg, _ := out["pagination"].(map[string]any)
		return code, revs, pg
	}
	code, revs, pg := page("")
	if code != 200 || len(revs) != 20 || pg["total"] != float64(total) || pg["has_more"] != true || pg["page"] != float64(1) || pg["page_size"] != float64(20) {
		t.Fatalf("page 1: %d len=%d %v", code, len(revs), pg)
	}
	_, revs, pg = page("?page=2")
	if len(revs) != 5 || pg["has_more"] != false {
		t.Fatalf("page 2: len=%d %v", len(revs), pg)
	}
	_, revs, pg = page("?page_size=1000")
	if len(revs) != 50 && len(revs) != total || pg["page_size"] != float64(50) {
		t.Fatalf("page_size cap: len=%d %v", len(revs), pg)
	}
}

func TestMyMediaReview(t *testing.T) {
	e := newEnv(t)
	insert := func(email, visibility string) string {
		uid := seedUser(t, e.db, email, true, "pw-password1")
		if err := e.db.Exec(`INSERT INTO reviews (created_at, updated_at, user_id, media_id, media_type, rating, body, is_public, visibility)
			VALUES (now(), now(), ?, 88, 'movie', 4.0, ?, ?, ?)`, uid, "mine-"+email, visibility == "public", visibility).Error; err != nil {
			t.Fatal(err)
		}
		return e.loginPair(email, "pw-password1").AccessToken
	}
	privTok := insert("priv@x.test", "private")
	insert("pub@x.test", "public")
	seedUser(t, e.db, "none@x.test", true, "pw-password1")
	noneTok := e.loginPair("none@x.test", "pw-password1").AccessToken
	const path = "/api/reviews/media/movies/88/me"

	if code, _ := e.do("GET", path, nil, ""); code != 401 {
		t.Fatalf("guest: %d, want 401", code)
	}
	code, out := e.do("GET", path, nil, privTok)
	if r, _ := out["review"].(map[string]any); code != 200 || r == nil || r["body"] != "mine-priv@x.test" {
		t.Fatalf("own private review: %d %v", code, out)
	}
	code, out = e.do("GET", path, nil, noneTok)
	if v, present := out["review"]; code != 200 || !present || v != nil {
		t.Fatalf("no review: %d %v, want review null", code, out)
	}
	if code, out := e.do("GET", "/api/reviews/media/book/88/me", nil, noneTok); code != 400 || out["code"] != "INVALID_MEDIA_TYPE" {
		t.Fatalf("bad type: %d %v", code, out)
	}
	if code, out := e.do("GET", "/api/reviews/media/movies/abc/me", nil, noneTok); code != 400 || out["code"] != "INVALID_INPUT" {
		t.Fatalf("bad id: %d %v", code, out)
	}
	// The public list is untouched and still guest-readable.
	if code, _ := e.do("GET", "/api/reviews/media/movies/88", nil, ""); code != 200 {
		t.Fatalf("public list: %d", code)
	}
}

func TestMediaRatingDistribution(t *testing.T) {
	e := newEnv(t)

	// Empty media: guest can call it, returns 10 zero-filled steps with count 0.
	code, out := e.do("GET", "/api/reviews/media/movies/901/rating", nil, "")
	if code != 200 {
		t.Fatalf("guest GET empty rating: %d %v", code, out)
	}
	emptyRating, _ := out["rating"].(map[string]any)
	emptyDist, _ := emptyRating["distribution"].([]any)
	if len(emptyDist) != 10 || emptyRating["review_count"] != float64(0) || emptyRating["has_rating"] != false {
		t.Fatalf("empty rating: %v", emptyRating)
	}
	for i, raw := range emptyDist {
		b := raw.(map[string]any)
		if b["rating"] != float64(i+1)*0.5 || b["count"] != float64(0) {
			t.Fatalf("empty bucket[%d]: %v", i, b)
		}
	}

	insert := func(email string, rating float64, visibility string, privateAccount, deleted bool) {
		uid := seedUser(t, e.db, email, true, "pw-password1")
		if privateAccount {
			e.db.Exec(`UPDATE users SET is_private = true WHERE id = ?`, uid)
		}
		delSQL := "NULL"
		if deleted {
			delSQL = "now()"
		}
		if err := e.db.Exec(fmt.Sprintf(`INSERT INTO reviews (created_at, updated_at, deleted_at, user_id, media_id, media_type, rating, body, is_public, visibility)
			VALUES (now(), now(), %s, ?, 902, 'movie', ?, 'b', ?, ?)`, delSQL), uid, rating, visibility == "public", visibility).Error; err != nil {
			t.Fatal(err)
		}
	}
	insert("u1@x.test", 5.0, "public", false, false)
	insert("u2@x.test", 5.0, "public", true, false) // public review on private profile is counted in aggregate
	insert("u3@x.test", 3.5, "public", false, false)
	insert("u4@x.test", 1.0, "private", false, false)   // excluded: private review
	insert("u5@x.test", 2.0, "followers", false, false) // excluded: followers-only review
	insert("u6@x.test", 4.0, "public", false, true)     // excluded: soft-deleted review

	code, out = e.do("GET", "/api/reviews/media/movies/902/rating", nil, "")
	if code != 200 {
		t.Fatalf("guest GET rating: %d %v", code, out)
	}
	ratingObj, _ := out["rating"].(map[string]any)
	dist, _ := ratingObj["distribution"].([]any)
	if len(dist) != 10 {
		t.Fatalf("len(distribution) = %d, want 10", len(dist))
	}
	count := int(ratingObj["review_count"].(float64))
	if count != 3 {
		t.Fatalf("review_count = %d, want 3", count)
	}
	sum := 0
	byRating := map[float64]int{}
	for _, raw := range dist {
		b := raw.(map[string]any)
		c := int(b["count"].(float64))
		sum += c
		byRating[b["rating"].(float64)] = c
	}
	if sum != count {
		t.Fatalf("sum(distribution) = %d != review_count = %d", sum, count)
	}
	if byRating[5.0] != 2 || byRating[3.5] != 1 || byRating[1.0] != 0 || byRating[2.0] != 0 || byRating[4.0] != 0 {
		t.Fatalf("unexpected bucket counts: %v", byRating)
	}
}

func TestGuestViewerZeroHasNoPersonalFlags(t *testing.T) {
	useFakeTMDB(t)
	e := newEnv(t)
	author := seedUser(t, e.db, "author@x.test", true, "pw-password1")
	e.db.Exec(`INSERT INTO reviews (id, created_at, updated_at, user_id, media_id, media_type, rating, body, is_public, visibility)
		VALUES (501, now(), now(), ?, 603, 'movie', 4.5, 'Great!', true, 'public')`, author)
	// Even if bogus user_id=0 rows existed in likes/helpful/media_likes/library, guests must still get false/nil.
	liker := seedUser(t, e.db, "liker@x.test", true, "pw-password1")
	e.db.Exec(`INSERT INTO review_likes (review_id, user_id, created_at) VALUES (501, ?, now())`, liker)
	e.db.Exec(`INSERT INTO review_helpful (review_id, user_id, created_at) VALUES (501, ?, now())`, liker)
	e.db.Exec(`INSERT INTO media_likes (user_id, media_id, media_type, liked_at) VALUES (?, 603, 'movie', now())`, liker)

	code, out := e.do("GET", "/api/reviews/media/movies/603", nil, "")
	if code != 200 {
		t.Fatalf("guest reviews: %d %v", code, out)
	}
	revs := out["reviews"].([]any)
	if len(revs) != 1 {
		t.Fatalf("expected 1 review, got %d", len(revs))
	}
	r := revs[0].(map[string]any)
	if r["is_liked"] != false || r["is_helpful_voted"] != false {
		t.Fatalf("guest review must have is_liked=false and is_helpful_voted=false, got liked=%v helpful=%v", r["is_liked"], r["is_helpful_voted"])
	}

	code, statsOut := e.do("GET", "/api/stats/movie/603", nil, "")
	if code != 200 {
		t.Fatalf("guest stats: %d %v", code, statsOut)
	}
	stats := statsOut["stats"].(map[string]any)
	if stats["liked_at"] != nil || stats["watchlisted_at"] != nil {
		t.Fatalf("guest stats must neutralize viewer fields, got liked_at=%v watchlisted_at=%v", stats["liked_at"], stats["watchlisted_at"])
	}
}

func TestAuthSessionEndpointWithCookie(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "sess@x.test", true, "my-password1")
	pair := e.loginPair("sess@x.test", "my-password1")

	checkCookie := func(val string) bool {
		req := httptest.NewRequest(http.MethodGet, "/api/auth/session", nil)
		if val != "" {
			req.AddCookie(&http.Cookie{Name: "refresh_token", Value: val})
		}
		resp, err := e.app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()
		raw, _ := io.ReadAll(resp.Body)
		var out map[string]bool
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatalf("decode: %v", err)
		}
		return out["has_session"]
	}

	if checkCookie("") {
		t.Fatal("missing cookie must return has_session=false")
	}
	if !checkCookie(pair.RefreshToken) {
		t.Fatal("live refresh cookie must return has_session=true")
	}
	e.db.Exec(`UPDATE users SET token_version = token_version + 1 WHERE id = ?`, uid)
	if checkCookie(pair.RefreshToken) {
		t.Fatal("stale refresh cookie must return has_session=false")
	}
}

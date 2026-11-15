//go:build integration

package integration

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/event_module"
	"github.com/arinsuda/movie-hub/internal/media_stats_module"
	"gorm.io/gorm"
)

// strRef returns a pointer to a copy of s (event media type and period are nullable columns).
func strRef(s string) *string { return &s }

// seedEvent inserts an event row as given and returns its id.
func seedEvent(t *testing.T, db *gorm.DB, ev event_module.Event) uint {
	t.Helper()
	if ev.TitleEN == "" {
		ev.TitleEN, ev.TitleTH = "Round", "รอบ"
	}
	if err := db.Create(&ev).Error; err != nil {
		t.Fatalf("seed event: %v", err)
	}
	return ev.ID
}

// seedFinishedRound inserts a finalized, published monthly or yearly round that ended in the past.
func seedFinishedRound(t *testing.T, db *gorm.DB, typ event_module.EventType, mediaType, period string) uint {
	t.Helper()
	starts := time.Date(2030, 1, 1, 0, 0, 0, 0, config.AppLocation)
	return seedEvent(t, db, event_module.Event{
		Type: typ, MediaType: strRef(mediaType), Period: strRef(period),
		StartsAt: starts, EndsAt: starts.AddDate(0, 1, 0),
		Status: event_module.StatusPublished, IsFinalized: true,
	})
}

// seedNominee inserts a nominee at position pos and returns its id.
func seedNominee(t *testing.T, db *gorm.DB, eventID uint, mediaType string, pos, mediaID int, poster, backdrop string) uint {
	t.Helper()
	n := event_module.EventNominee{
		EventID: eventID, MediaID: mediaID, MediaType: mediaType, Position: pos,
		MediaTitle: fmt.Sprintf("Title %d", mediaID), PosterPath: poster, BackdropPath: backdrop,
	}
	if err := db.Create(&n).Error; err != nil {
		t.Fatalf("seed nominee: %v", err)
	}
	return n.ID
}

// seedAward inserts a winner row for a finished round.
func seedAward(t *testing.T, db *gorm.DB, eventID, nomineeID uint, mediaID int, mediaType string, kind event_module.AwardKind, period string, votes int) {
	t.Helper()
	a := event_module.EventAward{
		EventID: eventID, NomineeID: nomineeID, MediaID: mediaID, MediaType: mediaType,
		AwardKind: kind, Period: period, Votes: votes,
	}
	if err := db.Create(&a).Error; err != nil {
		t.Fatalf("seed award: %v", err)
	}
}

// seedView inserts one per-day view row (media_type movie). day is YYYY-MM-DD in config.AppTimeZone.
func seedView(t *testing.T, db *gorm.DB, userID uint, mediaID int, day string) {
	t.Helper()
	if err := db.Exec(`INSERT INTO media_view_days (user_id, media_type, media_id, day) VALUES (?, 'movie', ?, ?::date)`,
		userID, mediaID, day).Error; err != nil {
		t.Fatalf("seed view: %v", err)
	}
}

func TestMediaViewCountsOncePerDay(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "viewer@x.test", true, "pw-viewer")
	tok := e.loginPair("viewer@x.test", "pw-viewer").AccessToken
	const mediaID = 611

	for i := range 2 {
		if code, res := e.do("POST", fmt.Sprintf("/api/stats/movie/%d/view", mediaID), nil, tok); code != 204 {
			t.Fatalf("view %d: got %d %v", i, code, res)
		}
	}

	var days []string
	if err := e.db.Raw(`SELECT to_char(day, 'YYYY-MM-DD') FROM media_view_days
		WHERE user_id = ? AND media_id = ? AND media_type = 'movie'`, uid, mediaID).Scan(&days).Error; err != nil {
		t.Fatal(err)
	}
	if want := config.LocalDay(time.Now()).Format("2006-01-02"); len(days) != 1 || days[0] != want {
		t.Fatalf("expected one view row on Bangkok day %s, got %v", want, days)
	}

	var views int
	if err := e.db.Raw(`SELECT view_count FROM media_stats WHERE media_id = ? AND media_type = 'movie'`, mediaID).Scan(&views).Error; err != nil {
		t.Fatal(err)
	}
	if views != 2 {
		t.Fatalf("view_count should still count every view, got %d", views)
	}
}

func TestCommunityScoreCountsOnlyActiveViewsInWindow(t *testing.T) {
	e := newEnv(t)
	useFakeVotingTMDB(t, 6)
	e.seedAdmin("admin@x.test", "pw-admin")
	adminTok := e.loginPair("admin@x.test", "pw-admin").AccessToken

	// Round 2030-06 opens 15 Jun 2030 (Bangkok); its activity window is 15 May to 15 Jun.
	active := seedUser(t, e.db, "active@x.test", true, "pw-active")
	inactive := seedUser(t, e.db, "inactive@x.test", true, "pw-inactive")
	deleted := seedUser(t, e.db, "deleted@x.test", true, "pw-deleted")
	e.db.Exec(`UPDATE users SET is_active = false WHERE id = ?`, inactive)
	e.db.Exec(`UPDATE users SET deleted_at = now() WHERE id = ?`, deleted)

	seedView(t, e.db, active, 901, "2030-05-20") // counts
	seedView(t, e.db, active, 901, "2030-06-01") // counts: second day
	seedView(t, e.db, active, 902, "2030-05-01") // before the window
	seedView(t, e.db, active, 902, "2030-06-15") // the round's opening day is outside the window
	seedView(t, e.db, inactive, 903, "2030-05-20")
	seedView(t, e.db, deleted, 904, "2030-05-20")

	code, res := e.do("GET", "/api/admin/events/suggest-nominees?media_type=movie&period=2030-06", nil, adminTok)
	if code != 200 {
		t.Fatalf("suggest nominees failed (%d): %v", code, res)
	}
	found := map[int]float64{}
	for _, raw := range res["nominees"].([]any) {
		n := raw.(map[string]any)
		if n["source"] == string(event_module.SourceCommunity) {
			found[int(n["media_id"].(float64))] = n["score"].(float64)
		}
	}
	if got, ok := found[901]; !ok || math.Abs(got-0.2) > 1e-9 {
		t.Fatalf("expected community score 0.2 for media 901 (two active days), got %v (present=%v)", got, ok)
	}
	for _, id := range []int{902, 903, 904} {
		if _, ok := found[id]; ok {
			t.Errorf("media %d must not score: its views are outside the window or from an inactive or deleted user", id)
		}
	}
}

func TestMediaViewPurgeRemovesExpiredDays(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "purge@x.test", true, "pw-purge")
	now := time.Now()
	expired := config.LocalDay(now.Add(-media_stats_module.MediaViewRetention - 48*time.Hour)).Format("2006-01-02")
	recent := config.LocalDay(now.Add(-24 * time.Hour)).Format("2006-01-02")
	seedView(t, e.db, uid, 911, expired)
	seedView(t, e.db, uid, 912, recent)

	deleted, err := media_stats_module.NewViewPurger(e.db).RunOnce(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 expired row deleted, got %d", deleted)
	}
	var kept []int
	if err := e.db.Raw(`SELECT media_id FROM media_view_days WHERE user_id = ?`, uid).Scan(&kept).Error; err != nil {
		t.Fatal(err)
	}
	if len(kept) != 1 || kept[0] != 912 {
		t.Fatalf("expected only the recent view (912) to remain, got %v", kept)
	}
}

// Hard-deleting a user (what account erasure does) cascades to their view rows.
func TestMediaViewRowsCascadeOnUserDelete(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "cascade@x.test", true, "pw-cascade")
	seedView(t, e.db, uid, 921, "2030-05-20")

	if err := e.db.Exec(`DELETE FROM users WHERE id = ?`, uid).Error; err != nil {
		t.Fatal(err)
	}
	if n := e.count(`SELECT COUNT(*) FROM media_view_days WHERE user_id = ?`, uid); n != 0 {
		t.Fatalf("expected view rows to cascade with the user, %d left", n)
	}
}

func TestAwardsListingFiltersAndMovieDetail(t *testing.T) {
	e := newEnv(t)
	useFakeVotingTMDB(t, 6)

	// 2030-03 movie round: a tie between 911 and 912, both winners.
	march := seedFinishedRound(t, e.db, event_module.TypeVoteMonthly, "movie", "2030-03")
	n911 := seedNominee(t, e.db, march, "movie", 1, 911, "/p911.jpg", "")
	n912 := seedNominee(t, e.db, march, "movie", 2, 912, "", "/b912.jpg")
	seedNominee(t, e.db, march, "movie", 3, 913, "/p913.jpg", "")
	seedAward(t, e.db, march, n911, 911, "movie", event_module.AwardKindMonthly, "2030-03", 3)
	seedAward(t, e.db, march, n912, 912, "movie", event_module.AwardKindMonthly, "2030-03", 3)

	// 2030-03 tv round.
	tvRound := seedFinishedRound(t, e.db, event_module.TypeVoteMonthly, "tv", "2030-03")
	n921 := seedNominee(t, e.db, tvRound, "tv", 1, 921, "/p921.jpg", "")
	seedAward(t, e.db, tvRound, n921, 921, "tv", event_module.AwardKindMonthly, "2030-03", 5)

	// 2030-02 movie round: the winner used by the movie detail check.
	feb := seedFinishedRound(t, e.db, event_module.TypeVoteMonthly, "movie", "2030-02")
	n611 := seedNominee(t, e.db, feb, "movie", 1, 611, "", "")
	seedAward(t, e.db, feb, n611, 611, "movie", event_module.AwardKindMonthly, "2030-02", 1)

	// A yearly winner in 2029.
	yearly := seedFinishedRound(t, e.db, event_module.TypeVoteYearly, "movie", "2029")
	n931 := seedNominee(t, e.db, yearly, "movie", 1, 931, "", "")
	seedAward(t, e.db, yearly, n931, 931, "movie", event_module.AwardKindYearly, "2029", 9)

	// A soft-deleted round and a draft round must not appear.
	deletedRound := seedFinishedRound(t, e.db, event_module.TypeVoteMonthly, "movie", "2030-04")
	nDel := seedNominee(t, e.db, deletedRound, "movie", 1, 941, "", "")
	seedAward(t, e.db, deletedRound, nDel, 941, "movie", event_module.AwardKindMonthly, "2030-04", 4)
	e.db.Exec(`UPDATE events SET deleted_at = now() WHERE id = ?`, deletedRound)
	draft := seedEvent(t, e.db, event_module.Event{
		Type: event_module.TypeVoteMonthly, MediaType: strRef("movie"), Period: strRef("2030-05"),
		StartsAt: time.Date(2030, 5, 1, 0, 0, 0, 0, time.UTC), EndsAt: time.Date(2030, 6, 1, 0, 0, 0, 0, time.UTC),
		Status: event_module.StatusDraft, IsFinalized: true,
	})
	nDraft := seedNominee(t, e.db, draft, "movie", 1, 951, "", "")
	seedAward(t, e.db, draft, nDraft, 951, "movie", event_module.AwardKindMonthly, "2030-05", 2)

	// Guest read of the 2030 list: 911, 912 (tie), 921 and 611, newest period first.
	code, res := e.do("GET", "/api/events/awards?year=2030", nil, "")
	if code != 200 {
		t.Fatalf("guest awards read failed (%d): %v", code, res)
	}
	awards := res["awards"].([]any)
	if len(awards) != 4 {
		t.Fatalf("expected 4 winners in 2030 (tie counts twice), got %d: %v", len(awards), res)
	}
	if last := awards[3].(map[string]any); last["media_id"].(float64) != 611 {
		t.Errorf("expected the 2030-02 winner last (newest first), got %v", last)
	}
	if years := res["years"].([]any); len(years) != 2 || years[0].(float64) != 2030 || years[1].(float64) != 2029 {
		t.Errorf("expected years [2030 2029], got %v", years)
	}
	first := awards[0].(map[string]any)
	if first["poster_path"] != "/p911.jpg" || first["backdrop_path"] != nil {
		t.Errorf("poster/backdrop snapshot wrong for 911: %v", first)
	}

	cases := []struct {
		path      string
		wantCount int
	}{
		{"/api/events/awards?year=2030&media_type=tv", 1},
		{"/api/events/awards?year=2030&media_type=movie", 3},
		{"/api/events/awards?year=2029", 1},
	}
	for _, tc := range cases {
		code, res := e.do("GET", tc.path, nil, "")
		if code != 200 || len(res["awards"].([]any)) != tc.wantCount {
			t.Errorf("%s: expected %d winners, got %d (%v)", tc.path, tc.wantCount, code, res)
		}
	}

	for _, path := range []string{
		"/api/events/awards?media_type=bogus",
		"/api/events/awards?year=1999",
		"/api/events/awards?year=abc",
	} {
		code, res := e.do("GET", path, nil, "")
		if code != 400 || res["code"] != "INVALID_INPUT" {
			t.Errorf("%s: expected 400 INVALID_INPUT, got %d %v", path, code, res)
		}
	}

	// Movie detail carries the title's awards; a title without any gets an empty list, never null.
	// The detail payload is wrapped as {"movie": {...}}.
	code, res = e.do("GET", "/api/movies/611", nil, "")
	if code != 200 {
		t.Fatalf("movie detail failed (%d): %v", code, res)
	}
	if got := res["movie"].(map[string]any)["awards"].([]any); len(got) != 1 || got[0].(map[string]any)["period"] != "2030-02" {
		t.Errorf("expected one 2030-02 award on movie 611, got %v", got)
	}
	code, res = e.do("GET", "/api/movies/612", nil, "")
	if code != 200 {
		t.Fatalf("movie detail for 612 failed (%d): %v", code, res)
	}
	if got, ok := res["movie"].(map[string]any)["awards"].([]any); !ok || len(got) != 0 {
		t.Errorf("expected an empty awards list for movie 612, got %v", res["movie"].(map[string]any)["awards"])
	}
}

func TestEventCoverBackdropFromFirstNomineeWithOne(t *testing.T) {
	e := newEnv(t)
	now := time.Now().UTC()
	seedUser(t, e.db, "member@x.test", true, "pw-member")
	tok := e.loginPair("member@x.test", "pw-member").AccessToken

	// Active round whose position 1 has a backdrop: the cover is that backdrop.
	withFirst := seedEvent(t, e.db, event_module.Event{
		Type: event_module.TypeVoteMonthly, MediaType: strRef("movie"), Period: strRef("2030-07"),
		StartsAt: now.Add(-time.Hour), EndsAt: now.Add(24 * time.Hour),
		Status: event_module.StatusPublished, ShowPopup: true,
	})
	seedNominee(t, e.db, withFirst, "movie", 1, 961, "", "/cover-1.jpg")
	seedNominee(t, e.db, withFirst, "movie", 2, 962, "", "/cover-2.jpg")

	// Position 1 has no backdrop: the cover skips it and takes the next one by position.
	skipEmpty := seedEvent(t, e.db, event_module.Event{
		Type: event_module.TypeVoteMonthly, MediaType: strRef("movie"), Period: strRef("2030-08"),
		StartsAt: now.Add(-time.Hour), EndsAt: now.Add(24 * time.Hour),
		Status: event_module.StatusPublished,
	})
	seedNominee(t, e.db, skipEmpty, "movie", 1, 963, "", "")
	seedNominee(t, e.db, skipEmpty, "movie", 2, 964, "", "/cover-2b.jpg")

	// An announcement never gets a cover.
	announcement := seedEvent(t, e.db, event_module.Event{
		Type: event_module.TypeAnnouncement, TitleEN: "News",
		StartsAt: now.Add(-time.Hour), EndsAt: now.Add(24 * time.Hour),
		Status: event_module.StatusPublished,
	})

	// coversByID maps each event in a list response to its cover_backdrop_path.
	coversByID := func(list []any) map[uint]any {
		out := map[uint]any{}
		for _, raw := range list {
			ev := raw.(map[string]any)
			out[uint(ev["id"].(float64))] = ev["cover_backdrop_path"]
		}
		return out
	}

	code, res := e.do("GET", "/api/events", nil, "")
	if code != 200 {
		t.Fatalf("list events failed (%d): %v", code, res)
	}
	got := coversByID(res["events"].([]any))
	if got[withFirst] != "/cover-1.jpg" {
		t.Errorf("list: expected cover /cover-1.jpg, got %v", got[withFirst])
	}
	if got[skipEmpty] != "/cover-2b.jpg" {
		t.Errorf("list: expected cover /cover-2b.jpg, got %v", got[skipEmpty])
	}
	if got[announcement] != nil {
		t.Errorf("list: announcement must have no cover, got %v", got[announcement])
	}

	code, res = e.do("GET", fmt.Sprintf("/api/events/%d", withFirst), nil, "")
	if code != 200 {
		t.Fatalf("detail failed (%d): %v", code, res)
	}
	if c := res["event"].(map[string]any)["cover_backdrop_path"]; c != "/cover-1.jpg" {
		t.Errorf("detail: expected cover /cover-1.jpg, got %v", c)
	}

	code, res = e.do("GET", "/api/events/popup", nil, tok)
	if code != 200 {
		t.Fatalf("popup failed (%d): %v", code, res)
	}
	if c := coversByID(res["events"].([]any))[withFirst]; c != "/cover-1.jpg" {
		t.Errorf("popup: expected cover /cover-1.jpg, got %v", c)
	}
}

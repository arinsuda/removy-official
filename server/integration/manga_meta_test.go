//go:build integration

package integration

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/arinsuda/movie-hub/internal/analytics_module"
	"github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/internal/shared"
	"gorm.io/gorm"
)

// countingLookup answers like fakeAniList and counts the AniList calls made through the shared media layer.
type countingLookup struct {
	mu    sync.Mutex
	calls int
}

func (c *countingLookup) LookupMedia(ctx context.Context, id int) (shared.MediaInfo, error) {
	c.mu.Lock()
	c.calls++
	c.mu.Unlock()
	return fakeAniList{}.LookupMedia(ctx, id)
}

func (c *countingLookup) reset() {
	c.mu.Lock()
	c.calls = 0
	c.mu.Unlock()
}

func (c *countingLookup) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

// countingEnv re-wires the media layer with a counting lookup on the test database.
func countingEnv(t *testing.T, e *env, gate adultGate) *countingLookup {
	t.Helper()
	lookup := &countingLookup{}
	shared.SetMangaSupport(lookup, gate, e.db)
	t.Cleanup(func() { shared.SetMangaSupport(nil, nil, nil) })
	return lookup
}

// countMangaMeta counts the SELECT statements that touch manga_meta, so a test can assert that a request
// reads the table in batches and never once per item.
func countMangaMeta(t *testing.T, db *gorm.DB) *int {
	t.Helper()
	n := new(int)
	if err := db.Callback().Query().After("gorm:query").Register("test:count_manga_meta", func(tx *gorm.DB) {
		if strings.Contains(tx.Statement.SQL.String(), "manga_meta") {
			*n++
		}
	}); err != nil {
		t.Fatalf("register query counter: %v", err)
	}
	return n
}

// summaryCount reads summary.count from a user reviews response.
func summaryCount(t *testing.T, out map[string]any) int {
	t.Helper()
	summary, ok := out["summary"].(map[string]any)
	if !ok {
		t.Fatalf("response has no summary: %v", out)
	}
	return int(summary["count"].(float64))
}

// TestMangaMetaServesReadsWithoutAniList checks that writes store the title and that a later list read
// is served from manga_meta: no AniList call, and a constant number of manga_meta reads per request.
func TestMangaMetaServesReadsWithoutAniList(t *testing.T) {
	e, gate := mangaEnv(t)
	lookup := countingEnv(t, e, gate)
	owner := seedUser(t, e.db, "meta-owner@x.test", true, "meta-password1")
	gate[owner] = true
	seedUser(t, e.db, "meta-minor@x.test", true, "meta-password1")
	ownerTok := e.loginPair("meta-owner@x.test", "meta-password1").AccessToken
	minorTok := e.loginPair("meta-minor@x.test", "meta-password1").AccessToken

	for _, id := range []int{mangaSFW, mangaAdult} {
		if code, out := e.do("POST", "/api/library", map[string]any{"media_id": id, "media_type": "manga", "list_type": "watchlist"}, ownerTok); code != 201 {
			t.Fatalf("library %d: %d %v", id, code, out)
		}
	}
	if n := e.count(`SELECT COUNT(*) FROM manga_meta WHERE (media_id = ? AND is_adult = false) OR (media_id = ? AND is_adult = true)`, mangaSFW, mangaAdult); n != 2 {
		t.Fatalf("manga_meta rows after the writes: %d matches, want 2", n)
	}

	lookup.reset()
	statements := countMangaMeta(t, e.db)
	code, out := e.do("GET", fmt.Sprintf("/api/library/user/%d", owner), nil, minorTok)
	if code != 200 {
		t.Fatalf("minor library read: %d %v", code, out)
	}
	ids := mediaIDs(t, out, "items")
	if containsInt(ids, mangaAdult) || !containsInt(ids, mangaSFW) {
		t.Fatalf("minor must see the safe title only, got %v", ids)
	}
	if n := lookup.count(); n != 0 {
		t.Errorf("reads with stored rows called AniList %d times, want 0", n)
	}
	// One read for the visibility check and one for the summaries, whatever the number of items.
	if *statements > 2 {
		t.Errorf("manga_meta statements for one list request: %d, want at most 2", *statements)
	}
}

// TestOwnerSeesOwnAdultTitles checks that a viewer who later loses eligibility still sees their own adult
// items, while other viewers do not, and that review counts follow the same rule.
func TestOwnerSeesOwnAdultTitles(t *testing.T) {
	e, gate := mangaEnv(t)
	owner := seedUser(t, e.db, "own-owner@x.test", true, "own-password1")
	gate[owner] = true
	seedUser(t, e.db, "own-minor@x.test", true, "own-password1")
	ownerTok := e.loginPair("own-owner@x.test", "own-password1").AccessToken
	minorTok := e.loginPair("own-minor@x.test", "own-password1").AccessToken

	for _, id := range []int{mangaSFW, mangaAdult} {
		if code, out := e.do("POST", "/api/reviews", map[string]any{"media_id": id, "media_type": "manga", "rating": 4.0, "body": "text"}, ownerTok); code != 201 {
			t.Fatalf("review %d: %d %v", id, code, out)
		}
		if code, out := e.do("POST", "/api/library", map[string]any{"media_id": id, "media_type": "manga", "list_type": "watchlist"}, ownerTok); code != 201 {
			t.Fatalf("library %d: %d %v", id, code, out)
		}
	}
	// The age gate is revoked after the writes; the owner keeps their own titles.
	gate[owner] = false

	_, own := e.do("GET", fmt.Sprintf("/api/library/user/%d", owner), nil, ownerTok)
	if !containsInt(mediaIDs(t, own, "items"), mangaAdult) {
		t.Errorf("owner must still see their own adult title")
	}
	_, minorView := e.do("GET", fmt.Sprintf("/api/library/user/%d", owner), nil, minorTok)
	if containsInt(mediaIDs(t, minorView, "items"), mangaAdult) {
		t.Errorf("another viewer must not see the adult title")
	}

	_, ownReviews := e.do("GET", fmt.Sprintf("/api/reviews/user/%d", owner), nil, ownerTok)
	_, minorReviews := e.do("GET", fmt.Sprintf("/api/reviews/user/%d", owner), nil, minorTok)
	if got := summaryCount(t, ownReviews); got != 2 {
		t.Errorf("owner review count = %d, want 2", got)
	}
	if got := summaryCount(t, minorReviews); got != 1 {
		t.Errorf("other viewer review count = %d, want 1 (adult review excluded)", got)
	}
}

// TestListItemCountExcludesAdultForOthers checks that the list card count leaves out adult titles for
// viewers who may not see them, and keeps them for the owner.
func TestListItemCountExcludesAdultForOthers(t *testing.T) {
	e, gate := mangaEnv(t)
	owner := seedUser(t, e.db, "count-owner@x.test", true, "count-password1")
	gate[owner] = true
	seedUser(t, e.db, "count-minor@x.test", true, "count-password1")
	ownerTok := e.loginPair("count-owner@x.test", "count-password1").AccessToken
	minorTok := e.loginPair("count-minor@x.test", "count-password1").AccessToken

	_, listOut := e.do("POST", "/api/lists", map[string]any{"name": "Shelf"}, ownerTok)
	listID := int(listOut["list"].(map[string]any)["id"].(float64))
	for _, id := range []int{mangaSFW, mangaAdult} {
		if code, out := e.do("POST", fmt.Sprintf("/api/lists/%d/items", listID), map[string]any{"media_id": id, "media_type": "manga"}, ownerTok); code >= 300 {
			t.Fatalf("list item %d: %d %v", id, code, out)
		}
	}

	ownCount := listItemCount(t, e, "/api/lists", ownerTok)
	minorCount := listItemCount(t, e, fmt.Sprintf("/api/users/%d/lists", owner), minorTok)
	if ownCount != 2 {
		t.Errorf("owner item_count = %d, want 2", ownCount)
	}
	if minorCount != 1 {
		t.Errorf("other viewer item_count = %d, want 1 (adult title excluded)", minorCount)
	}
}

// listItemCount returns the item_count of the first list in a list response.
func listItemCount(t *testing.T, e *env, path, token string) int {
	t.Helper()
	code, out := e.do("GET", path, nil, token)
	if code != 200 {
		t.Fatalf("GET %s: %d %v", path, code, out)
	}
	lists, ok := out["lists"].([]any)
	if !ok || len(lists) == 0 {
		t.Fatalf("GET %s: no lists in %v", path, out)
	}
	return int(lists[0].(map[string]any)["item_count"].(float64))
}

// TestReviewLikeHelpfulCommentNeedEligibility checks that an ineligible viewer cannot act on an adult review.
func TestReviewLikeHelpfulCommentNeedEligibility(t *testing.T) {
	e, gate := mangaEnv(t)
	author := seedUser(t, e.db, "gate-author@x.test", true, "gate-password1")
	gate[author] = true
	seedUser(t, e.db, "gate-minor@x.test", true, "gate-password1")
	authorTok := e.loginPair("gate-author@x.test", "gate-password1").AccessToken
	minorTok := e.loginPair("gate-minor@x.test", "gate-password1").AccessToken

	code, out := e.do("POST", "/api/reviews", map[string]any{"media_id": mangaAdult, "media_type": "manga", "rating": 3.0, "body": "x"}, authorTok)
	if code != 201 {
		t.Fatalf("author review: %d %v", code, out)
	}
	reviewID := int(out["review"].(map[string]any)["id"].(float64))

	actions := map[string]struct {
		path string
		body map[string]any
	}{
		"like":    {fmt.Sprintf("/api/reviews/%d/like", reviewID), nil},
		"helpful": {fmt.Sprintf("/api/reviews/%d/helpful", reviewID), nil},
		"comment": {fmt.Sprintf("/api/reviews/%d/comments", reviewID), map[string]any{"body": "hi"}},
	}
	for name, a := range actions {
		code, out := e.do("POST", a.path, a.body, minorTok)
		if code != 403 || out["code"] != string(shared.ErrorCodeAgeRestricted) {
			t.Errorf("minor %s: want 403 AGE_RESTRICTED, got %d %v", name, code, out)
		}
	}

	// An eligible viewer passes the same gate.
	if code, out := e.do("POST", fmt.Sprintf("/api/reviews/%d/comments", reviewID), map[string]any{"body": "ok"}, authorTok); code != 201 {
		t.Errorf("eligible comment: %d %v", code, out)
	}
}

// TestAdultMangaNotificationsNeedEligibility is the fan-out matrix: an adult title never reaches a
// minor follower or a minor review owner, a safe title reaches everyone, and stored rows are filtered on read.
func TestAdultMangaNotificationsNeedEligibility(t *testing.T) {
	e, gate := mangaEnv(t)
	author := seedUser(t, e.db, "fan-author@x.test", true, "fan-password1")
	eligible := seedUser(t, e.db, "fan-eligible@x.test", true, "fan-password1")
	gate[eligible] = true
	minor := seedUser(t, e.db, "fan-minor@x.test", true, "fan-password1")

	svc := newNotifService(t, e.db, fakeUserProvider{followers: []uint{eligible, minor}}, nil)
	svc.SetTitleCheck(shared.NotificationTitleCheck)
	svc.SetMangaRestriction(shared.NotificationMangaRestricted)
	ctx := context.Background()
	manga := func(id int, title string) notification_module.MediaRef {
		return notification_module.MediaRef{ID: uint(id), Type: "manga", Title: title}
	}
	received := func(uid uint, targetID int) int64 {
		return e.count(`SELECT COUNT(*) FROM notifications WHERE user_id = ? AND target_id = ? AND deleted_at IS NULL`, uid, targetID)
	}

	// Follower fan-out: the adult title reaches no follower, the safe title reaches both.
	if err := svc.PushFollowingReviewed(ctx, author, 901, manga(mangaAdult, "Adult")); err != nil {
		t.Fatal(err)
	}
	if err := svc.PushFollowingReviewed(ctx, author, 902, manga(mangaSFW, "Sfw")); err != nil {
		t.Fatal(err)
	}
	if n := received(eligible, 901) + received(minor, 901); n != 0 {
		t.Errorf("adult fan-out reached %d followers, want 0", n)
	}
	if received(eligible, 902) != 1 || received(minor, 902) != 1 {
		t.Errorf("safe fan-out must reach both followers")
	}

	// Review owner: only an eligible owner is told about an adult title.
	if err := svc.PushReviewLiked(ctx, eligible, author, 903, manga(mangaAdult, "Adult")); err != nil {
		t.Fatal(err)
	}
	if err := svc.PushReviewLiked(ctx, minor, author, 904, manga(mangaAdult, "Adult")); err != nil {
		t.Fatal(err)
	}
	if received(eligible, 903) != 1 {
		t.Errorf("eligible owner must get the like notification")
	}
	if received(minor, 904) != 0 {
		t.Errorf("minor owner must not get the adult like notification")
	}

	// A row stored before eligibility was lost is hidden on read, and only for the minor.
	cat := notification_module.NotificationCategory("media")
	adultRef := "manga"
	adultID := uint(mangaAdult)
	for _, uid := range []uint{eligible, minor} {
		row := notification_module.Notification{
			UserID: uid, Type: notification_module.NotifFollowingLikedMedia, Category: cat,
			TargetID: &adultID, TargetRef: &adultRef,
			Params: &notification_module.NotificationParams{MediaID: mangaAdult, MediaType: "manga", MediaTitle: "Adult"},
		}
		if err := e.db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	minorList, err := svc.ListNotifications(ctx, minor, notification_module.ListNotificationsQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(minorList.Notifications) != 1 {
		t.Errorf("minor list: %d notifications, want 1 (the safe follow only)", len(minorList.Notifications))
	}
	eligibleList, err := svc.ListNotifications(ctx, eligible, notification_module.ListNotificationsQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(eligibleList.Notifications) != 3 {
		t.Errorf("eligible list: %d notifications, want 3", len(eligibleList.Notifications))
	}
}

// TestTrendingAndAnalyticsHideAdultManga checks the analytics read path: trending leaves adult titles out for
// ineligible viewers, and the single-title analytics is refused like any other read.
func TestTrendingAndAnalyticsHideAdultManga(t *testing.T) {
	e, gate := mangaEnv(t)
	owner := seedUser(t, e.db, "trend-owner@x.test", true, "trend-password1")
	gate[owner] = true
	minor := seedUser(t, e.db, "trend-minor@x.test", true, "trend-password1")
	ownerTok := e.loginPair("trend-owner@x.test", "trend-password1").AccessToken
	for _, id := range []int{mangaSFW, mangaAdult} {
		if code, out := e.do("POST", "/api/library", map[string]any{"media_id": id, "media_type": "manga", "list_type": "watchlist"}, ownerTok); code != 201 {
			t.Fatalf("library %d: %d %v", id, code, out)
		}
	}

	svc := analytics_module.NewService(e.db)
	ctx := context.Background()
	minorRows, err := svc.GetTrending(ctx, minor, "manga")
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]int, 0, len(minorRows))
	for _, r := range minorRows {
		ids = append(ids, r.MediaID)
	}
	if containsInt(ids, mangaAdult) || !containsInt(ids, mangaSFW) {
		t.Errorf("minor trending = %v, want the safe title only", ids)
	}

	ownerRows, err := svc.GetTrending(ctx, owner, "manga")
	if err != nil {
		t.Fatal(err)
	}
	if len(ownerRows) != 2 {
		t.Errorf("eligible trending = %d rows, want 2", len(ownerRows))
	}

	if _, err := svc.GetMediaAnalytics(ctx, minor, mangaAdult, "manga"); !errors.Is(err, shared.ErrAgeRestricted) {
		t.Errorf("adult analytics for a minor: want ErrAgeRestricted, got %v", err)
	}
	if _, err := svc.GetMediaAnalytics(ctx, minor, mangaSFW, "manga"); err != nil {
		t.Errorf("safe analytics for a minor: %v", err)
	}
}

// TestMangaMetaUpsertWritesYearAndGenres checks that the first write and a later refresh both store the start
// year and genre names, and that the genre list is capped.
func TestMangaMetaUpsertWritesYearAndGenres(t *testing.T) {
	e, gate := mangaEnv(t)
	countingEnv(t, e, gate)
	const id = 4201
	ctx := context.Background()

	shared.RecordMangaMeta(ctx, id, shared.MediaInfo{Title: "First", StartYear: 2019, GenreNames: []string{"Action"}})
	var row shared.MangaMeta
	if err := e.db.Where("media_id = ?", id).First(&row).Error; err != nil {
		t.Fatalf("read first write: %v", err)
	}
	if row.StartYear != 2019 || len(row.Genres) != 1 || row.Genres[0] != "Action" {
		t.Fatalf("first write year/genres: %d %v", row.StartYear, row.Genres)
	}

	many := make([]string, shared.MangaMetaMaxGenres+2)
	for i := range many {
		many[i] = fmt.Sprintf("Genre%d", i)
	}
	shared.RecordMangaMeta(ctx, id, shared.MediaInfo{Title: "Second", StartYear: 2021, GenreNames: many})
	row = shared.MangaMeta{}
	if err := e.db.Where("media_id = ?", id).First(&row).Error; err != nil {
		t.Fatalf("read refresh: %v", err)
	}
	if row.Title != "Second" || row.StartYear != 2021 || len(row.Genres) != shared.MangaMetaMaxGenres {
		t.Fatalf("refresh must update year and capped genres: title=%q year=%d genres=%d", row.Title, row.StartYear, len(row.Genres))
	}
}

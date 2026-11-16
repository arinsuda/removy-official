//go:build integration

package integration

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	anilist "github.com/arinsuda/movie-hub/internal/anilist_module"
	"github.com/arinsuda/movie-hub/internal/manga_module"
	"github.com/arinsuda/movie-hub/internal/notification_module"
	"github.com/arinsuda/movie-hub/internal/shared"
)

// Manga ids used only by this file, so no other fixture caches them.
const (
	freshSafe      = 4001
	freshAdult     = 4002
	staleFlips     = 4003
	staleSafeDown  = 4004
	staleAdultDown = 4005
	deadID         = 4006
	capIDFirst     = 4100
	detailAdult    = 4301
	notifSafe      = 4201
	notifAdult     = 4202
)

// scriptedLookup answers AniList lookups from a table that a test can change. An id missing from titles is not
// found; down makes every lookup fail as unavailable. It counts every lookup it answers.
type scriptedLookup struct {
	mu     sync.Mutex
	titles map[int]shared.MediaInfo
	down   bool
	calls  int
}

func (s *scriptedLookup) LookupMedia(_ context.Context, id int) (shared.MediaInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	if s.down {
		return shared.MediaInfo{}, shared.ErrMediaUnavailable
	}
	info, ok := s.titles[id]
	if !ok {
		return shared.MediaInfo{}, shared.ErrMediaNotFound
	}
	return info, nil
}

func (s *scriptedLookup) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

func (s *scriptedLookup) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = 0
}

// scriptedEnv wires the media layer with a scripted lookup on the test database.
func scriptedEnv(t *testing.T, e *env, gate adultGate, titles map[int]shared.MediaInfo) *scriptedLookup {
	t.Helper()
	lookup := &scriptedLookup{titles: titles}
	shared.SetMangaSupport(lookup, gate, e.db)
	t.Cleanup(func() { shared.SetMangaSupport(nil, nil, nil) })
	return lookup
}

// seedMangaMeta stores a manga_meta row whose updated_at is age in the past.
func seedMangaMeta(t *testing.T, e *env, id int, title string, adult bool, age time.Duration) {
	t.Helper()
	err := e.db.Exec(`INSERT INTO manga_meta (media_id, title, cover_url, kind, is_adult, updated_at)
		VALUES (?, ?, '', 'manga', ?, ?)
		ON CONFLICT (media_id) DO UPDATE SET title = EXCLUDED.title, is_adult = EXCLUDED.is_adult, updated_at = EXCLUDED.updated_at`,
		id, title, adult, time.Now().Add(-age)).Error
	if err != nil {
		t.Fatal(err)
	}
}

// mangaRef keys one manga id for KeepVisible.
func mangaRef(id int) shared.VisibleRef {
	return shared.VisibleRef{Media: shared.MediaKey{ID: id, Type: shared.MediaTypeManga}}
}

// keptManga returns the ids viewer may see, through the same KeepVisible path the list reads use.
func keptManga(t *testing.T, viewer uint, ids ...int) []int {
	t.Helper()
	vis, err := shared.NewMediaVisibility(context.Background(), viewer)
	if err != nil {
		t.Fatal(err)
	}
	kept, err := shared.KeepVisible(context.Background(), vis, ids, func(id int) shared.VisibleRef { return mangaRef(id) })
	if err != nil {
		t.Fatal(err)
	}
	return kept
}

// TestStaleMangaFlipsToAdultAndIsHidden checks that a stored safe row that AniList now reports as adult is
// refreshed on read, so a minor stops seeing the title.
func TestStaleMangaFlipsToAdultAndIsHidden(t *testing.T) {
	e, gate := mangaEnv(t)
	minor := seedUser(t, e.db, "stale-minor@x.test", true, "stale-password1")
	lookup := scriptedEnv(t, e, gate, map[int]shared.MediaInfo{staleFlips: {Title: "Now adult", IsAdult: true}})
	seedMangaMeta(t, e, staleFlips, "Was safe", false, 48*time.Hour)

	if kept := keptManga(t, minor, staleFlips); len(kept) != 0 {
		t.Errorf("stale row that became adult is still visible to a minor: %v", kept)
	}
	if n := lookup.count(); n != 1 {
		t.Errorf("AniList lookups for one stale row = %d, want 1", n)
	}
	if n := e.count(`SELECT COUNT(*) FROM manga_meta WHERE media_id = ? AND is_adult = true`, staleFlips); n != 1 {
		t.Errorf("stored row was not refreshed to adult")
	}
	if err := shared.CheckMediaAccess(context.Background(), minor, staleFlips, shared.MediaTypeManga); !errors.Is(err, shared.ErrAgeRestricted) {
		t.Errorf("write gate for a minor after refresh: want ErrAgeRestricted, got %v", err)
	}
}

// TestFreshMangaRowsNeedNoAniListCall checks that fresh stored rows answer every read without AniList.
func TestFreshMangaRowsNeedNoAniListCall(t *testing.T) {
	e, gate := mangaEnv(t)
	minor := seedUser(t, e.db, "fresh-minor@x.test", true, "fresh-password1")
	lookup := scriptedEnv(t, e, gate, map[int]shared.MediaInfo{})
	seedMangaMeta(t, e, freshSafe, "Safe", false, 0)
	seedMangaMeta(t, e, freshAdult, "Adult", true, time.Hour)

	if kept := keptManga(t, minor, freshSafe, freshAdult); len(kept) != 1 || kept[0] != freshSafe {
		t.Errorf("minor kept %v, want only %d", kept, freshSafe)
	}
	if err := shared.CheckMediaAccess(context.Background(), minor, freshSafe, shared.MediaTypeManga); err != nil {
		t.Errorf("safe fresh title write gate: %v", err)
	}
	if n := lookup.count(); n != 0 {
		t.Errorf("fresh rows called AniList %d times, want 0", n)
	}
}

// TestStaleRowKeepsStoredFlagsWhenAniListIsDown checks that a stale row is still used when the refresh fails:
// a safe title stays visible and an adult title stays hidden.
func TestStaleRowKeepsStoredFlagsWhenAniListIsDown(t *testing.T) {
	e, gate := mangaEnv(t)
	minor := seedUser(t, e.db, "down-minor@x.test", true, "down-password1")
	lookup := scriptedEnv(t, e, gate, map[int]shared.MediaInfo{})
	lookup.down = true
	seedMangaMeta(t, e, staleSafeDown, "Safe", false, 48*time.Hour)
	seedMangaMeta(t, e, staleAdultDown, "Adult", true, 48*time.Hour)

	if kept := keptManga(t, minor, staleSafeDown, staleAdultDown); len(kept) != 1 || kept[0] != staleSafeDown {
		t.Errorf("minor kept %v while AniList is down, want only %d", kept, staleSafeDown)
	}
	if n := lookup.count(); n != 2 {
		t.Errorf("refresh attempts = %d, want 2", n)
	}
	if n := e.count(`SELECT COUNT(*) FROM manga_meta WHERE media_id = ? AND is_adult = false AND updated_at < now() - interval '24 hours'`, staleSafeDown); n != 1 {
		t.Errorf("a failed refresh must leave the stored row unchanged")
	}
}

// TestMangaRefreshIsBoundedPerRequest checks that stale rows share the fallback cap: one call never makes more
// than MangaMetaFallbackLimit AniList lookups.
func TestMangaRefreshIsBoundedPerRequest(t *testing.T) {
	e, gate := mangaEnv(t)
	titles := map[int]shared.MediaInfo{}
	var ids []int
	for i := 0; i < shared.MangaMetaFallbackLimit+2; i++ {
		id := capIDFirst + i
		titles[id] = shared.MediaInfo{Title: fmt.Sprintf("Cap %d", id)}
		seedMangaMeta(t, e, id, "Old", false, 48*time.Hour)
		ids = append(ids, id)
	}
	lookup := scriptedEnv(t, e, gate, titles)
	if _, err := shared.MangaRows(context.Background(), ids); err != nil {
		t.Fatal(err)
	}
	if n := lookup.count(); n != shared.MangaMetaFallbackLimit {
		t.Errorf("AniList lookups for %d stale rows = %d, want %d", len(ids), n, shared.MangaMetaFallbackLimit)
	}
}

// TestNotFoundMarkerStopsRepeatLookups checks that an id AniList does not know is stored as a marker: it stays
// hidden and not found, costs one lookup, and is rechecked only after MangaMetaNotFoundRecheckAfter.
func TestNotFoundMarkerStopsRepeatLookups(t *testing.T) {
	e, gate := mangaEnv(t)
	minor := seedUser(t, e.db, "dead-minor@x.test", true, "dead-password1")
	owner := seedUser(t, e.db, "dead-owner@x.test", true, "dead-password1")
	gate[owner] = true
	lookup := scriptedEnv(t, e, gate, map[int]shared.MediaInfo{})

	if kept := keptManga(t, minor, deadID); len(kept) != 0 {
		t.Errorf("unknown title visible to a minor: %v", kept)
	}
	if n := e.count(`SELECT COUNT(*) FROM manga_meta WHERE media_id = ? AND is_adult = true AND title = ''`, deadID); n != 1 {
		t.Fatalf("not-found marker rows = %d, want 1", n)
	}
	if kept := keptManga(t, minor, deadID); len(kept) != 0 {
		t.Errorf("unknown title visible on second read: %v", kept)
	}
	if n := lookup.count(); n != 1 {
		t.Errorf("lookups for a dead id over two reads = %d, want 1", n)
	}
	// The write gate keeps the 404 for an unknown title, even for an eligible viewer.
	if err := shared.CheckMediaAccess(context.Background(), owner, deadID, shared.MediaTypeManga); !errors.Is(err, shared.ErrMediaNotFound) {
		t.Errorf("write gate on a marker: want ErrMediaNotFound, got %v", err)
	}
	if n := lookup.count(); n != 1 {
		t.Errorf("a fresh marker must not be looked up again, got %d lookups", n)
	}

	if err := e.db.Exec(`UPDATE manga_meta SET updated_at = now() - interval '7 hours' WHERE media_id = ?`, deadID).Error; err != nil {
		t.Fatal(err)
	}
	keptManga(t, minor, deadID)
	if n := lookup.count(); n != 2 {
		t.Errorf("a marker older than the recheck window must be looked up again, got %d lookups", n)
	}
}

// TestMangaDetailRecordsFlagsBeforeAgeGate checks that a manga detail fetch refreshes the stored row, even when
// the viewer is refused the title, so the next list read already hides it.
func TestMangaDetailRecordsFlagsBeforeAgeGate(t *testing.T) {
	e, _ := mangaEnv(t)
	seedMangaMeta(t, e, detailAdult, "Was safe", false, 48*time.Hour)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, fmt.Sprintf(`{"data":{"Media":{"id":%d,"type":"MANGA","isAdult":true,"title":{"romaji":"Adult","english":"Adult"}}}}`, detailAdult))
	}))
	t.Cleanup(srv.Close)
	client := anilist.NewClient(anilist.Options{Endpoint: srv.URL, HTTP: srv.Client()})
	svc := manga_module.NewMangaService(client)

	if _, err := svc.GetByID(context.Background(), detailAdult, false); !errors.Is(err, manga_module.ErrAgeRestricted) {
		t.Fatalf("minor detail: want ErrAgeRestricted, got %v", err)
	}
	if n := e.count(`SELECT COUNT(*) FROM manga_meta WHERE media_id = ? AND is_adult = true`, detailAdult); n != 1 {
		t.Errorf("refused detail fetch did not record the adult flag")
	}
}

// TestNotificationCountsHideAdultManga checks that unread counts, badges and page totals leave out adult manga
// for a viewer who may not see it, in SQL, and keep it for an eligible viewer.
func TestNotificationCountsHideAdultManga(t *testing.T) {
	e, gate := mangaEnv(t)
	eligible := seedUser(t, e.db, "count-notif-eligible@x.test", true, "count-notif-pw1")
	gate[eligible] = true
	minor := seedUser(t, e.db, "count-notif-minor@x.test", true, "count-notif-pw1")
	seedMangaMeta(t, e, notifSafe, "Safe", false, 0)
	seedMangaMeta(t, e, notifAdult, "Adult", true, 0)

	svc := newNotifService(t, e.db, fakeUserProvider{}, nil)
	svc.SetTitleCheck(shared.NotificationTitleCheck)
	svc.SetMangaRestriction(shared.NotificationMangaRestricted)

	cat := notification_module.NotificationCategory("media")
	for _, uid := range []uint{eligible, minor} {
		for _, p := range []*notification_module.NotificationParams{
			{MediaID: notifSafe, MediaType: "manga", MediaTitle: "Safe"},
			{MediaID: notifAdult, MediaType: "manga", MediaTitle: "Adult"},
			{MediaID: 5, MediaType: "movie", MediaTitle: "Movie"},
		} {
			row := notification_module.Notification{UserID: uid, Type: notification_module.NotifFollowingLikedMedia, Category: cat, Params: p}
			if err := e.db.Create(&row).Error; err != nil {
				t.Fatal(err)
			}
		}
	}

	ctx := context.Background()
	if n, err := svc.GetUnreadCount(ctx, minor); err != nil || n != 2 {
		t.Errorf("minor unread = %d (%v), want 2 (safe manga and movie)", n, err)
	}
	if n, err := svc.GetUnreadCount(ctx, eligible); err != nil || n != 3 {
		t.Errorf("eligible unread = %d (%v), want 3", n, err)
	}
	minorList, err := svc.ListNotifications(ctx, minor, notification_module.ListNotificationsQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if minorList.Pagination.Total != 2 || len(minorList.Notifications) != 2 || minorList.UnreadCount != 2 {
		t.Errorf("minor page: total %d, rows %d, unread %d; want 2, 2, 2",
			minorList.Pagination.Total, len(minorList.Notifications), minorList.UnreadCount)
	}
	byCat, err := svc.GetUnreadCountByCategory(ctx, minor)
	if err != nil {
		t.Fatal(err)
	}
	if len(byCat) != 1 || byCat[0].Count != 2 {
		t.Errorf("minor unread by category = %+v, want one category with 2", byCat)
	}
}

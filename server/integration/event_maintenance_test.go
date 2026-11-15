//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/database/migrate"
	"github.com/arinsuda/movie-hub/internal/event_module"
	"github.com/arinsuda/movie-hub/internal/media_stats_module"
	"gorm.io/gorm"
)

func eventPeriod(t *testing.T, db *gorm.DB, id uint) string {
	t.Helper()
	var period string
	if err := db.Raw(`SELECT period FROM events WHERE id = ?`, id).Scan(&period).Error; err != nil {
		t.Fatal(err)
	}
	return period
}

func eventFinalized(t *testing.T, db *gorm.DB, id uint) bool {
	t.Helper()
	var done bool
	if err := db.Raw(`SELECT is_finalized FROM events WHERE id = ?`, id).Scan(&done).Error; err != nil {
		t.Fatal(err)
	}
	return done
}

// Rows written under the older period rule carry the month before their window opens. The backfill
// realigns them to the Bangkok month of starts_at and copies the new period onto monthly awards.
func TestEventPeriodBackfillRealignsRoundsAndAwards(t *testing.T) {
	e := newEnv(t)
	bkk := config.AppLocation
	novStart := time.Date(2026, 11, 15, 0, 0, 0, 0, bkk)
	octStart := time.Date(2026, 10, 15, 0, 0, 0, 0, bkk)

	// a and b swap periods: moving a first would collide with b's current period in a unique index.
	a := seedEvent(t, e.db, event_module.Event{
		Type: event_module.TypeVoteMonthly, MediaType: strRef("movie"), Period: strRef("2026-10"),
		StartsAt: novStart, EndsAt: novStart.AddDate(0, 1, 0), Status: event_module.StatusPublished, IsFinalized: true,
	})
	b := seedEvent(t, e.db, event_module.Event{
		Type: event_module.TypeVoteMonthly, MediaType: strRef("movie"), Period: strRef("2026-11"),
		StartsAt: octStart, EndsAt: octStart.AddDate(0, 1, 0), Status: event_module.StatusPublished, IsFinalized: true,
	})
	// Opens 1 Nov 06:30 Bangkok, which is 31 Oct 23:30 UTC: the Bangkok month is November.
	c := seedEvent(t, e.db, event_module.Event{
		Type: event_module.TypeVoteMonthly, MediaType: strRef("tv"), Period: strRef("2026-10"),
		StartsAt: time.Date(2026, 11, 1, 6, 30, 0, 0, bkk), EndsAt: time.Date(2026, 12, 1, 6, 30, 0, 0, bkk),
		Status: event_module.StatusPublished,
	})
	nominee := seedNominee(t, e.db, a, "movie", 0, 101, "", "")
	seedAward(t, e.db, a, nominee, 101, "movie", event_module.AwardKindMonthly, "2026-10", 5)

	if err := migrate.BackfillEventPeriodsToStartMonth(e.db, config.AppTimeZone); err != nil {
		t.Fatalf("backfill: %v", err)
	}

	for id, want := range map[uint]string{a: "2026-11", b: "2026-10", c: "2026-11"} {
		if got := eventPeriod(t, e.db, id); got != want {
			t.Errorf("event %d period: want %s, got %s", id, want, got)
		}
	}
	var awardPeriod string
	if err := e.db.Raw(`SELECT period FROM event_awards WHERE event_id = ?`, a).Scan(&awardPeriod).Error; err != nil {
		t.Fatal(err)
	}
	if awardPeriod != "2026-11" {
		t.Errorf("award period: want 2026-11, got %s", awardPeriod)
	}

	// A second run changes nothing.
	if err := migrate.BackfillEventPeriodsToStartMonth(e.db, config.AppTimeZone); err != nil {
		t.Fatalf("second backfill: %v", err)
	}
	if got := eventPeriod(t, e.db, a); got != "2026-11" {
		t.Errorf("re-run changed period to %s", got)
	}
}

// Two live rounds of one media type that open in the same Bangkok month cannot share a period.
// The backfill gives it to the earliest and leaves the other NULL instead of failing boot.
func TestBackfillEventPeriodsSkipsCollisions(t *testing.T) {
	e := newEnv(t)
	bkk := config.AppLocation
	first := time.Date(2026, 8, 15, 0, 0, 0, 0, bkk)
	second := time.Date(2026, 8, 20, 0, 0, 0, 0, bkk)
	a := seedEvent(t, e.db, event_module.Event{
		Type: event_module.TypeVoteMonthly, MediaType: strRef("movie"), Period: strRef("2026-07"),
		StartsAt: first, EndsAt: first.AddDate(0, 1, 0), Status: event_module.StatusPublished,
	})
	b := seedEvent(t, e.db, event_module.Event{
		Type: event_module.TypeVoteMonthly, MediaType: strRef("movie"), Period: strRef("2026-09"),
		StartsAt: second, EndsAt: second.AddDate(0, 1, 0), Status: event_module.StatusPublished,
	})

	if err := migrate.BackfillEventPeriodsToStartMonth(e.db, config.AppTimeZone); err != nil {
		t.Fatalf("backfill must not fail on a collision: %v", err)
	}
	if got := eventPeriod(t, e.db, a); got != "2026-08" {
		t.Errorf("earliest round: want 2026-08, got %q", got)
	}
	var bPeriod *string
	if err := e.db.Raw(`SELECT period FROM events WHERE id = ?`, b).Scan(&bPeriod).Error; err != nil {
		t.Fatal(err)
	}
	if bPeriod != nil {
		t.Errorf("colliding round: want NULL, got %s", *bPeriod)
	}
}

// The worker settles ended, published, unfinalized rounds with no page visit. Open, draft and
// soft-deleted rounds are left alone, and a second pass settles nothing.
func TestEventFinalizerSettlesEndedRoundsWithoutVisit(t *testing.T) {
	e := newEnv(t)
	now := time.Now().UTC()
	ended := seedEvent(t, e.db, event_module.Event{
		Type: event_module.TypeVoteMonthly, MediaType: strRef("movie"), Period: strRef("2026-09"),
		StartsAt: now.Add(-30 * 24 * time.Hour), EndsAt: now.Add(-time.Hour), Status: event_module.StatusPublished,
	})
	open := seedEvent(t, e.db, event_module.Event{
		Type: event_module.TypeVoteMonthly, MediaType: strRef("movie"), Period: strRef("2026-10"),
		StartsAt: now.Add(-time.Hour), EndsAt: now.Add(24 * time.Hour), Status: event_module.StatusPublished,
	})
	draft := seedEvent(t, e.db, event_module.Event{
		Type: event_module.TypeVoteMonthly, MediaType: strRef("movie"), Period: strRef("2026-08"),
		StartsAt: now.Add(-30 * 24 * time.Hour), EndsAt: now.Add(-time.Hour), Status: event_module.StatusDraft,
	})
	deleted := seedEvent(t, e.db, event_module.Event{
		Type: event_module.TypeVoteMonthly, MediaType: strRef("movie"), Period: strRef("2026-07"),
		StartsAt: now.Add(-30 * 24 * time.Hour), EndsAt: now.Add(-time.Hour), Status: event_module.StatusPublished,
	})
	if err := e.db.Delete(&event_module.Event{}, deleted).Error; err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	settled, err := event_module.NewFinalizer(e.db).RunOnce(ctx)
	if err != nil {
		t.Fatalf("finalize pass: %v", err)
	}
	if settled != 1 {
		t.Fatalf("want 1 round settled, got %d", settled)
	}
	if !eventFinalized(t, e.db, ended) {
		t.Error("ended round was not finalized")
	}
	if eventFinalized(t, e.db, open) {
		t.Error("open round must not be finalized")
	}
	if eventFinalized(t, e.db, draft) {
		t.Error("draft round must not be finalized")
	}
	if eventFinalized(t, e.db, deleted) {
		t.Error("soft-deleted round must not be finalized")
	}

	if settled, err := event_module.NewFinalizer(e.db).RunOnce(ctx); err != nil || settled != 0 {
		t.Fatalf("second pass: want 0 settled and no error, got %d, %v", settled, err)
	}
}

// One user may record at most ViewRateMax views per window; the next one is 429.
func TestMediaViewRateLimit(t *testing.T) {
	e := newEnv(t)
	_, tok := newVoter(t, e, "view-rate@x.test")

	for i := 1; i <= media_stats_module.ViewRateMax; i++ {
		if code, res := e.do("POST", "/api/stats/movie/778/view", nil, tok); code != 204 {
			t.Fatalf("view %d of %d should pass, got %d: %v", i, media_stats_module.ViewRateMax, code, res)
		}
	}
	code, res := e.do("POST", "/api/stats/movie/778/view", nil, tok)
	if code != 429 || res["code"] != "RATE_LIMITED" {
		t.Fatalf("view over the budget: want 429 RATE_LIMITED, got %d: %v", code, res)
	}
}

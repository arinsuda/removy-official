//go:build integration

package integration

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/internal/event_module"
)

// castVote inserts a vote row directly (the HTTP path refuses votes outside the window).
func castVote(t *testing.T, e *env, eventID, userID, nomineeID uint) {
	t.Helper()
	if err := e.db.Exec(`INSERT INTO event_votes (event_id, user_id, nominee_id, created_at, updated_at) VALUES (?, ?, ?, now(), now())`,
		eventID, userID, nomineeID).Error; err != nil {
		t.Fatal(err)
	}
}

func endedRound(t *testing.T, e *env) (uint, uint) {
	t.Helper()
	now := time.Now().UTC()
	id := seedEvent(t, e.db, event_module.Event{
		Type: event_module.TypeVoteMonthly, MediaType: strRef("movie"), Period: strRef("2026-05"),
		StartsAt: now.Add(-48 * time.Hour), EndsAt: now.Add(-time.Hour), Status: event_module.StatusPublished,
	})
	nominee := seedNominee(t, e.db, id, "movie", 1, 11, "", "")
	seedNominee(t, e.db, id, "movie", 2, 22, "", "")
	return id, nominee
}

func TestFinalizeNotifiesEachActiveVoterOnce(t *testing.T) {
	e := newEnv(t)
	active := seedUser(t, e.db, "va@x.test", true, "pw")
	inactive := seedUser(t, e.db, "vb@x.test", true, "pw")
	deleted := seedUser(t, e.db, "vc@x.test", true, "pw")
	id, nominee := endedRound(t, e)
	for _, u := range []uint{active, inactive, deleted} {
		castVote(t, e, id, u, nominee)
	}
	e.db.Exec(`UPDATE users SET is_active = false WHERE id = ?`, inactive)
	e.db.Exec(`UPDATE users SET deleted_at = now() WHERE id = ?`, deleted)

	// Concurrent finalizers plus a second pass: lockEvent serializes them, so one row per voter.
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = event_module.NewFinalizer(e.db, nil).RunOnce(t.Context())
		}()
	}
	wg.Wait()
	if _, err := event_module.NewFinalizer(e.db, nil).RunOnce(t.Context()); err != nil {
		t.Fatal(err)
	}

	count := func(uid uint) int64 {
		return e.count(`SELECT COUNT(*) FROM notifications WHERE type = 'event_results' AND target_ref = 'event' AND target_id = ? AND user_id = ?`, id, uid)
	}
	if n := count(active); n != 1 {
		t.Errorf("active voter: want exactly 1 notification, got %d", n)
	}
	if n := count(inactive) + count(deleted); n != 0 {
		t.Errorf("inactive and deleted voters must get none, got %d", n)
	}
	var row struct {
		Category string
		Message  string
		ActorID  *uint
		Params   string
	}
	if err := e.db.Raw(`SELECT category, message, actor_id, params::text AS params FROM notifications WHERE user_id = ?`, active).Scan(&row).Error; err != nil {
		t.Fatal(err)
	}
	if row.Category != "system" || row.Message != "" || row.ActorID != nil {
		t.Errorf("unexpected row: %+v", row)
	}
	if want := e.count(`SELECT COUNT(*) FROM notifications WHERE user_id = ? AND params = '{"event_type":"vote_monthly","media_type":"movie","period":"2026-05"}'::jsonb`, active); want != 1 {
		t.Errorf("params must hold only event_type, media_type and period, got %s", row.Params)
	}
}

func TestFinalizeWithoutVotesCreatesNoNotifications(t *testing.T) {
	e := newEnv(t)
	endedRound(t, e)
	if _, err := event_module.NewFinalizer(e.db, nil).RunOnce(t.Context()); err != nil {
		t.Fatal(err)
	}
	if n := e.count(`SELECT COUNT(*) FROM notifications WHERE type = 'event_results'`); n != 0 {
		t.Errorf("want 0 notifications, got %d", n)
	}
}

func TestVoteRefusesNewAccountAndExposesEligibility(t *testing.T) {
	e := newEnv(t)
	e.seedAdmin("admin@x.test", "pw-admin")
	adminTok := e.loginPair("admin@x.test", "pw-admin").AccessToken
	eventID, nominees := openVotingRound(t, e, adminTok)

	newUID := seedUser(t, e.db, "new@x.test", true, "pw-new")
	newTok := e.loginPair("new@x.test", "pw-new").AccessToken
	code, res := vote(e, eventID, nominees[0], newTok)
	if code != 403 || res["code"] != "ACCOUNT_TOO_NEW" || res["eligible_at"] == nil || res["min_account_age_days"] != float64(event_module.MinVoterAccountAgeDays) {
		t.Fatalf("want 403 ACCOUNT_TOO_NEW with eligible_at and min_account_age_days, got %d: %v", code, res)
	}
	if n := e.count(`SELECT COUNT(*) FROM event_votes WHERE user_id = ?`, newUID); n != 0 {
		t.Errorf("a refused vote must not be stored, got %d", n)
	}
	if code, _ := e.do("DELETE", fmt.Sprintf("/api/events/%d/vote", eventID), nil, newTok); code != 200 {
		t.Errorf("withdrawing is not gated, got %d", code)
	}

	// The round closes in 24h, before this account turns old enough: no eligibility hint.
	_, short := e.do("GET", fmt.Sprintf("/api/events/%d", eventID), nil, newTok)
	if _, ok := short["event"].(map[string]any)["vote_eligible_at"]; ok {
		t.Errorf("eligibility after ends_at must not be exposed: %v", short["event"])
	}
	now := time.Now().UTC()
	// The window is frozen once open, so move the end directly.
	if err := e.db.Exec("UPDATE events SET ends_at = ? WHERE id = ?", now.Add(30*24*time.Hour), eventID).Error; err != nil {
		t.Fatal(err)
	}

	_, signed := e.do("GET", fmt.Sprintf("/api/events/%d", eventID), nil, newTok)
	if signed["event"].(map[string]any)["vote_eligible_at"] == nil {
		t.Error("a too-new signed-in viewer must see vote_eligible_at")
	}
	_, guest := e.do("GET", fmt.Sprintf("/api/events/%d", eventID), nil, "")
	if _, ok := guest["event"].(map[string]any)["vote_eligible_at"]; ok {
		t.Error("guests must not get vote_eligible_at")
	}

	_, oldTok := newVoter(t, e, "old@x.test")
	if code, res := vote(e, eventID, nominees[0], oldTok); code != 200 {
		t.Fatalf("an old account can vote, got %d: %v", code, res)
	}
	_, old := e.do("GET", fmt.Sprintf("/api/events/%d", eventID), nil, oldTok)
	if _, ok := old["event"].(map[string]any)["vote_eligible_at"]; ok {
		t.Error("an eligible viewer must not get vote_eligible_at")
	}

	// A past round never carries the hint.
	if err := e.db.Exec("UPDATE events SET starts_at = ?, ends_at = ? WHERE id = ?", now.Add(-48*time.Hour), now.Add(-24*time.Hour), eventID).Error; err != nil {
		t.Fatal(err)
	}
	_, past := e.do("GET", fmt.Sprintf("/api/events/%d", eventID), nil, newTok)
	if _, ok := past["event"].(map[string]any)["vote_eligible_at"]; ok {
		t.Error("a past round must not expose vote_eligible_at")
	}
}

func TestDeleteVotingRoundWithVotesOrResultsIsRefused(t *testing.T) {
	e := newEnv(t)
	e.seedAdmin("admin@x.test", "pw-admin")
	adminTok := e.loginPair("admin@x.test", "pw-admin").AccessToken
	del := func(id uint) (int, map[string]any) {
		return e.do("DELETE", fmt.Sprintf("/api/admin/events/%d", id), nil, adminTok)
	}

	eventID, nominees := openVotingRound(t, e, adminTok)
	_, tok := newVoter(t, e, "dv@x.test")
	if code, res := vote(e, eventID, nominees[0], tok); code != 200 {
		t.Fatalf("vote failed (%d): %v", code, res)
	}
	if code, res := del(eventID); code != 409 || res["code"] != "EVENT_NOT_DELETABLE" {
		t.Errorf("round with votes: want 409 EVENT_NOT_DELETABLE, got %d: %v", code, res)
	}
	_, list := e.do("GET", "/api/admin/events", nil, adminTok)
	for _, raw := range list["events"].([]any) {
		ev := raw.(map[string]any)
		if uint(ev["id"].(float64)) == eventID && ev["has_votes"] != true {
			t.Errorf("has_votes must be true for a round with votes: %v", ev)
		}
	}

	finalized := seedFinishedRound(t, e.db, event_module.TypeVoteMonthly, "movie", "2030-01")
	if code, res := del(finalized); code != 409 || res["code"] != "EVENT_NOT_DELETABLE" {
		t.Errorf("finalized round: want 409 EVENT_NOT_DELETABLE, got %d: %v", code, res)
	}

	now := time.Now().UTC()
	draft := seedEvent(t, e.db, event_module.Event{
		Type: event_module.TypeVoteMonthly, MediaType: strRef("movie"), Period: strRef("2031-01"),
		StartsAt: now.Add(time.Hour), EndsAt: now.Add(48 * time.Hour), Status: event_module.StatusDraft,
	})
	if code, res := del(draft); code != 204 {
		t.Errorf("draft without votes: want 204, got %d: %v", code, res)
	}
}

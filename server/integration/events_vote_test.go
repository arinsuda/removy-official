//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/event_module"
	tmdbmodule "github.com/arinsuda/movie-hub/internal/tmdb_module"
)

// useFakeVotingTMDB points the TMDB client at a local server that serves a fixed pool of movies.
func useFakeVotingTMDB(t *testing.T, count int) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/movie/now_playing"), strings.HasSuffix(r.URL.Path, "/movie/popular"):
			items := make([]string, 0, count)
			for i := 1; i <= count; i++ {
				id := 500 + i
				items = append(items, fmt.Sprintf(
					`{"id":%d,"title":"Candidate Movie %d","overview":"Overview %d","poster_path":"/p%d.jpg","backdrop_path":"/b%d.jpg","popularity":%d}`,
					id, i, i, i, i, 100-i,
				))
			}
			_, _ = w.Write([]byte(fmt.Sprintf(`{"page":1,"total_pages":1,"total_results":%d,"results":[%s]}`, count, strings.Join(items, ","))))
		default:
			_, _ = w.Write([]byte(`{"id":601,"title":"Community Pick","overview":"Top community film","poster_path":"/c.jpg","backdrop_path":"/cb.jpg","genres":[]}`))
		}
	}))
	t.Cleanup(srv.Close)
	tmdbmodule.Init(&config.Config{TMDB: config.TMDBConfig{BaseURL: srv.URL, APIKey: "test"}})
}

// roundBody is the admin update body for a monthly movie round (copy left blank).
func roundBody(period string, starts, ends time.Time, status string) map[string]any {
	return map[string]any{
		"type":           "vote_monthly",
		"media_type":     "movie",
		"period":         period,
		"starts_at":      starts.Format(time.RFC3339),
		"ends_at":        ends.Format(time.RFC3339),
		"show_popup":     true,
		"popup_priority": 10,
		"status":         status,
	}
}

// setWindow moves a round's window directly in the DB (the API refuses to edit an open round).
func setWindow(t *testing.T, e *env, eventID uint, starts, ends time.Time) {
	t.Helper()
	if err := e.db.Model(&event_module.Event{}).Where("id = ?", eventID).Updates(map[string]any{
		"starts_at": starts,
		"ends_at":   ends,
	}).Error; err != nil {
		t.Fatal(err)
	}
}

// openVotingRound generates the next monthly movie round, publishes it, and opens its window now.
// It returns the event id and nominee ids in position order.
func openVotingRound(t *testing.T, e *env, adminTok string) (uint, []uint) {
	t.Helper()
	useFakeVotingTMDB(t, 6)
	period := event_module.NextMonthlyPeriod(time.Now())

	code, genRes := e.do("POST", "/api/admin/events/generate-monthly?media_type=movie&period="+period, nil, adminTok)
	if code != 201 {
		t.Fatalf("generate-monthly failed (%d): %v", code, genRes)
	}
	eventID := uint(genRes["event"].(map[string]any)["id"].(float64))

	now := time.Now().UTC()
	code, pubRes := e.do("PUT", fmt.Sprintf("/api/admin/events/%d", eventID),
		roundBody(period, now.Add(-time.Hour), now.Add(24*time.Hour), "published"), adminTok)
	if code != 200 {
		t.Fatalf("publish round failed (%d): %v", code, pubRes)
	}

	var nomineeIDs []uint
	for _, n := range genRes["event"].(map[string]any)["nominees"].([]any) {
		nomineeIDs = append(nomineeIDs, uint(n.(map[string]any)["id"].(float64)))
	}
	return eventID, nomineeIDs
}

// newVoter seeds a verified user and returns its access token.
func newVoter(t *testing.T, e *env, email string) (uint, string) {
	t.Helper()
	uid := seedUser(t, e.db, email, true, "pw-"+email)
	return uid, e.loginPair(email, "pw-"+email).AccessToken
}

func vote(e *env, eventID, nomineeID uint, token string) (int, map[string]any) {
	return e.do("PUT", fmt.Sprintf("/api/events/%d/vote", eventID), map[string]any{"nominee_id": nomineeID}, token)
}

func TestEventsVotingAndNomineeFlow(t *testing.T) {
	e := newEnv(t)
	now := time.Now().UTC()
	period := event_module.NextMonthlyPeriod(now)
	otherPeriod := event_module.NextMonthlyPeriod(now.AddDate(0, 1, 0))

	e.seedAdmin("admin@x.test", "pw-admin")
	adminTok := e.loginPair("admin@x.test", "pw-admin").AccessToken

	seedUser(t, e.db, "voter1@x.test", true, "pw-voter1")
	voter1Tok := e.loginPair("voter1@x.test", "pw-voter1").AccessToken

	seedUser(t, e.db, "voter2@x.test", true, "pw-voter2")
	voter2Tok := e.loginPair("voter2@x.test", "pw-voter2").AccessToken

	// 0. When fewer than 5 real candidates exist, generate-monthly refuses with 422 NOT_ENOUGH_NOMINEES
	// (never fabricates fake placeholder movies).
	useFakeVotingTMDB(t, 2)
	code, shortRes := e.do("POST", "/api/admin/events/generate-monthly?media_type=movie&period="+period, nil, adminTok)
	if code != 422 || shortRes["code"] != "NOT_ENOUGH_NOMINEES" {
		t.Fatalf("expected 422 NOT_ENOUGH_NOMINEES when <5 candidates, got %d: %v", code, shortRes)
	}

	// Now provide a full pool of 6 TMDB movies.
	useFakeVotingTMDB(t, 6)

	// 1. Suggest nominees via admin
	code, sugRes := e.do("GET", "/api/admin/events/suggest-nominees?media_type=movie&period="+period, nil, adminTok)
	if code != 200 {
		t.Fatalf("suggest nominees failed (%d): %v", code, sugRes)
	}
	sugNominees, ok := sugRes["nominees"].([]any)
	if !ok || len(sugNominees) != 5 {
		t.Fatalf("expected 5 suggested nominees, got: %v", sugRes)
	}

	// 2. Generate monthly round (and verify concurrent duplicate returns 409, never 500)
	var wgGen sync.WaitGroup
	genCodes := make([]int, 2)
	genBodies := make([]map[string]any, 2)
	for i := 0; i < 2; i++ {
		wgGen.Add(1)
		go func(idx int) {
			defer wgGen.Done()
			genCodes[idx], genBodies[idx] = e.do("POST", "/api/admin/events/generate-monthly?media_type=movie&period="+period, nil, adminTok)
		}(i)
	}
	wgGen.Wait()

	var genRes map[string]any
	createdCount, conflictCount := 0, 0
	for i, c := range genCodes {
		switch c {
		case 201:
			createdCount++
			genRes = genBodies[i]
		case 409:
			conflictCount++
			if genBodies[i]["code"] != "EVENT_ALREADY_EXISTS" {
				t.Errorf("expected EVENT_ALREADY_EXISTS on conflict, got: %v", genBodies[i])
			}
		default:
			t.Fatalf("unexpected status from concurrent generate-monthly: %d (%v)", c, genBodies[i])
		}
	}
	if createdCount != 1 || conflictCount != 1 {
		t.Fatalf("expected 1x201 and 1x409 from concurrent generate-monthly, got codes=%v", genCodes)
	}

	// A generated round is a draft with blank copy: the admin writes it before publishing.
	evMap := genRes["event"].(map[string]any)
	eventID := uint(evMap["id"].(float64))
	if evMap["status"] != "draft" {
		t.Errorf("expected generated round to be a draft, got %v", evMap["status"])
	}
	evTitle := evMap["title"].(map[string]any)
	evBody := evMap["body"].(map[string]any)
	if evTitle["en"] != "" || evTitle["th"] != "" || evBody["en"] != "" || evBody["th"] != "" {
		t.Errorf("expected blank title and body on generated round, got title=%v body=%v", evTitle, evBody)
	}
	nominees := evMap["nominees"].([]any)
	if len(nominees) != 5 {
		t.Fatalf("expected 5 nominees, got: %d", len(nominees))
	}

	// 2b. While the round is UPCOMING (before starts_at), SetNominees is allowed.
	upcomingStarts := now.Add(2 * time.Hour).Truncate(time.Minute)
	upcomingEnds := now.Add(48 * time.Hour).Truncate(time.Minute)
	setWindow(t, e, eventID, upcomingStarts, upcomingEnds)

	customNominees := []map[string]any{
		{"media_id": 701, "media_type": "movie", "media_title": "Nominee One", "position": 1},
		{"media_id": 702, "media_type": "movie", "media_title": "Nominee Two", "position": 2},
		{"media_id": 703, "media_type": "movie", "media_title": "Nominee Three", "position": 3},
	}
	code, setRes := e.do("PUT", fmt.Sprintf("/api/admin/events/%d/nominees", eventID), map[string]any{"nominees": customNominees}, adminTok)
	if code != 200 {
		t.Fatalf("expected 200 editing nominees while upcoming, got %d: %v", code, setRes)
	}
	updatedNoms := setRes["nominees"].([]any)
	if len(updatedNoms) != 3 {
		t.Fatalf("expected 3 updated nominees, got %d", len(updatedNoms))
	}
	nom1ID := uint(updatedNoms[0].(map[string]any)["id"].(float64))
	nom2ID := uint(updatedNoms[1].(map[string]any)["id"].(float64))

	// Publish the round while it is still upcoming.
	code, pubRes := e.do("PUT", fmt.Sprintf("/api/admin/events/%d", eventID), roundBody(period, upcomingStarts, upcomingEnds, "published"), adminTok)
	if code != 200 {
		t.Fatalf("publish while upcoming failed (%d): %v", code, pubRes)
	}

	// Update event starts_at and ends_at to make it currently ACTIVE
	activeStarts := now.Add(-time.Hour).Truncate(time.Minute)
	activeEnds := now.Add(24 * time.Hour).Truncate(time.Minute)
	setWindow(t, e, eventID, activeStarts, activeEnds)

	// 2c. Once ACTIVE: premature finalization is blocked with 409 VOTING_NOT_ENDED
	code, earlyFin := e.do("POST", fmt.Sprintf("/api/admin/events/%d/finalize", eventID), nil, adminTok)
	if code != 409 || earlyFin["code"] != "VOTING_NOT_ENDED" {
		t.Fatalf("expected 409 VOTING_NOT_ENDED when finalizing active round, got %d: %v", code, earlyFin)
	}

	// 2d. Once ACTIVE: SetNominees is blocked with 409 VOTING_LOCKED
	code, lockedNoms := e.do("PUT", fmt.Sprintf("/api/admin/events/%d/nominees", eventID), map[string]any{"nominees": customNominees[:2]}, adminTok)
	if code != 409 || lockedNoms["code"] != "VOTING_LOCKED" {
		t.Fatalf("expected 409 VOTING_LOCKED when editing nominees on active round, got %d: %v", code, lockedNoms)
	}

	// 3. Guest views active voting event: vote counts, percentages, total votes MUST BE HIDDEN
	code, guestRes := e.do("GET", fmt.Sprintf("/api/events/%d", eventID), nil, "")
	if code != 200 {
		t.Fatalf("guest get event failed (%d): %v", code, guestRes)
	}
	guestEv := guestRes["event"].(map[string]any)
	if guestEv["total_votes"] != nil {
		t.Errorf("expected total_votes to be hidden for guest while active, got: %v", guestEv["total_votes"])
	}
	guestNoms := guestEv["nominees"].([]any)
	for _, n := range guestNoms {
		nm := n.(map[string]any)
		if nm["vote_count"] != nil {
			t.Errorf("expected vote_count to be hidden while active, got: %v", nm["vote_count"])
		}
		if nm["is_winner"] == true {
			t.Errorf("expected is_winner to be false while active")
		}
	}

	// 4. Guest attempts to vote -> 401 Unauthorized
	code, _ = e.do("PUT", fmt.Sprintf("/api/events/%d/vote", eventID), map[string]any{"nominee_id": nom1ID}, "")
	if code != 401 {
		t.Fatalf("expected 401 for guest voting, got %d", code)
	}

	// 5. Concurrent votes by Voter 1 (atomic upsert must never 500)
	var wgVote sync.WaitGroup
	for i := 0; i < 4; i++ {
		wgVote.Add(1)
		target := nom1ID
		if i%2 == 1 {
			target = nom2ID
		}
		go func(nid uint) {
			defer wgVote.Done()
			c, res := vote(e, eventID, nid, voter1Tok)
			if c != 200 {
				t.Errorf("concurrent vote failed (%d): %v", c, res)
			}
		}(target)
	}
	wgVote.Wait()

	// Settle Voter 1 on Nominee 1
	code, voteRes := vote(e, eventID, nom1ID, voter1Tok)
	if code != 200 {
		t.Fatalf("voter1 vote failed (%d): %v", code, voteRes)
	}
	if uint(voteRes["nominee_id"].(float64)) != nom1ID {
		t.Errorf("expected nominee_id %d, got %v", nom1ID, voteRes["nominee_id"])
	}

	// Voter 1 checks event: my_vote_nominee_id is nom1ID, but vote counts still hidden!
	code, voter1Get := e.do("GET", fmt.Sprintf("/api/events/%d", eventID), nil, voter1Tok)
	if code != 200 {
		t.Fatalf("voter1 get event failed (%d): %v", code, voter1Get)
	}
	v1Ev := voter1Get["event"].(map[string]any)
	if v1Ev["my_vote_nominee_id"] == nil || uint(v1Ev["my_vote_nominee_id"].(float64)) != nom1ID {
		t.Errorf("expected my_vote_nominee_id %d, got %v", nom1ID, v1Ev["my_vote_nominee_id"])
	}
	if v1Ev["total_votes"] != nil {
		t.Errorf("expected total_votes hidden from voter1 during active voting, got: %v", v1Ev["total_votes"])
	}

	// 6. Voter 1 changes vote to Nominee 2 (upsert)
	code, vote2Res := vote(e, eventID, nom2ID, voter1Tok)
	if code != 200 {
		t.Fatalf("voter1 change vote failed (%d): %v", code, vote2Res)
	}

	var voteCount int64
	if err := e.db.Table("event_votes").Where("event_id = ?", eventID).Count(&voteCount).Error; err != nil {
		t.Fatal(err)
	}
	if voteCount != 1 {
		t.Errorf("expected 1 vote row after change, got %d", voteCount)
	}

	// 7. Voter 2 votes for Nominee 1, then cancels
	code, _ = vote(e, eventID, nom1ID, voter2Tok)
	if code != 200 {
		t.Fatalf("voter2 vote failed (%d)", code)
	}
	code, cancelRes := e.do("DELETE", fmt.Sprintf("/api/events/%d/vote", eventID), nil, voter2Tok)
	if code != 200 {
		t.Fatalf("cancel vote failed (%d): %v", code, cancelRes)
	}
	if err := e.db.Table("event_votes").Where("event_id = ?", eventID).Count(&voteCount).Error; err != nil {
		t.Fatal(err)
	}
	if voteCount != 1 {
		t.Errorf("expected 1 vote row after cancel, got %d", voteCount)
	}

	// Set up a tie: voter1 -> nom1ID, voter2 -> nom2ID
	_, _ = vote(e, eventID, nom1ID, voter1Tok)
	_, _ = vote(e, eventID, nom2ID, voter2Tok)

	// 8. Admin copy edit (title/body) on an active round succeeds and preserves nominees + votes
	code, copyRes := e.do("PUT", fmt.Sprintf("/api/admin/events/%d", eventID), map[string]any{
		"type":           "vote_monthly",
		"media_type":     "movie",
		"period":         period,
		"title_en":       "Updated Poll Title",
		"title_th":       "อัปเดตชื่อโหวต",
		"body_en":        "Updated description",
		"body_th":        "คำอธิบายใหม่",
		"starts_at":      activeStarts.Format(time.RFC3339),
		"ends_at":        activeEnds.Format(time.RFC3339),
		"show_popup":     true,
		"popup_priority": 10,
		"status":         "published",
	}, adminTok)
	if code != 200 {
		t.Fatalf("expected 200 on copy edit of active round, got %d: %v", code, copyRes)
	}
	if err := e.db.Table("event_votes").Where("event_id = ?", eventID).Count(&voteCount).Error; err != nil {
		t.Fatal(err)
	}
	if voteCount != 2 {
		t.Fatalf("expected 2 votes preserved after copy edit, got %d", voteCount)
	}

	// Attempting to pass nominees or alter period via Update on an active round fails with 409 VOTING_LOCKED
	code, badUpdate := e.do("PUT", fmt.Sprintf("/api/admin/events/%d", eventID), map[string]any{
		"type":       "vote_monthly",
		"media_type": "movie",
		"period":     otherPeriod,
		"title_en":   "Updated Poll Title",
		"title_th":   "อัปเดตชื่อโหวต",
		"starts_at":  activeStarts.Format(time.RFC3339),
		"ends_at":    activeEnds.Format(time.RFC3339),
		"status":     "published",
	}, adminTok)
	if code != 409 || badUpdate["code"] != "VOTING_LOCKED" {
		t.Fatalf("expected 409 VOTING_LOCKED when changing period of active round, got %d: %v", code, badUpdate)
	}

	// 9. Simulating Voting End & Lazy Finalization (Co-winners Tie!)
	setWindow(t, e, eventID, now.Add(-48*time.Hour), now.Add(-time.Hour))

	// Voting is now closed: attempt to vote should fail with 400 VOTING_CLOSED
	code, closedRes := vote(e, eventID, nom1ID, voter1Tok)
	if code != 409 || closedRes["code"] != "VOTING_CLOSED" {
		t.Fatalf("expected 409 VOTING_CLOSED, got %d: %v", code, closedRes)
	}

	// Fetch event after end: triggers lazy finalization!
	code, pastRes := e.do("GET", fmt.Sprintf("/api/events/%d", eventID), nil, "")
	if code != 200 {
		t.Fatalf("get past event failed (%d): %v", code, pastRes)
	}
	pastEv := pastRes["event"].(map[string]any)
	if pastEv["total_votes"] == nil || int(pastEv["total_votes"].(float64)) != 2 {
		t.Errorf("expected total_votes == 2, got: %v", pastEv["total_votes"])
	}
	if pastEv["is_finalized"] != true {
		t.Errorf("expected is_finalized == true")
	}

	// Both nom1 and nom2 have 1 vote each and are co-winners!
	pastNoms := pastEv["nominees"].([]any)
	winnerCount := 0
	for _, n := range pastNoms {
		nm := n.(map[string]any)
		id := uint(nm["id"].(float64))
		if id == nom1ID || id == nom2ID {
			if nm["vote_count"] == nil || int(nm["vote_count"].(float64)) != 1 {
				t.Errorf("expected nominee %d to have 1 vote, got %v", id, nm["vote_count"])
			}
			if nm["is_winner"] != true {
				t.Errorf("expected nominee %d to be co-winner", id)
			}
			winnerCount++
		}
	}
	if winnerCount != 2 {
		t.Errorf("expected 2 co-winners, got %d", winnerCount)
	}

	// Verify EventAward table has 2 award records for this event
	var awardCount int64
	if err := e.db.Model(&event_module.EventAward{}).Where("event_id = ?", eventID).Count(&awardCount).Error; err != nil {
		t.Fatal(err)
	}
	if awardCount != 2 {
		t.Errorf("expected 2 EventAward rows for co-winners, got %d", awardCount)
	}
}

func TestEventsVoteOnFinalizedRoundIsClosed(t *testing.T) {
	e := newEnv(t)
	e.seedAdmin("admin@x.test", "pw-admin")
	adminTok := e.loginPair("admin@x.test", "pw-admin").AccessToken
	_, voterTok := newVoter(t, e, "voter@x.test")

	eventID, nominees := openVotingRound(t, e, adminTok)
	if code, res := vote(e, eventID, nominees[0], voterTok); code != 200 {
		t.Fatalf("vote while open failed (%d): %v", code, res)
	}

	// Still inside the window, but already settled: votes and withdrawals are refused.
	if err := e.db.Model(&event_module.Event{}).Where("id = ?", eventID).Update("is_finalized", true).Error; err != nil {
		t.Fatal(err)
	}
	if code, res := vote(e, eventID, nominees[1], voterTok); code != 409 || res["code"] != "VOTING_CLOSED" {
		t.Errorf("expected 409 VOTING_CLOSED voting on finalized round, got %d: %v", code, res)
	}
	if code, res := e.do("DELETE", fmt.Sprintf("/api/events/%d/vote", eventID), nil, voterTok); code != 409 || res["code"] != "VOTING_CLOSED" {
		t.Errorf("expected 409 VOTING_CLOSED cancelling on finalized round, got %d: %v", code, res)
	}
}

func TestEventsFinalizeExcludesInactiveAndDeletedVoters(t *testing.T) {
	e := newEnv(t)
	e.seedAdmin("admin@x.test", "pw-admin")
	adminTok := e.loginPair("admin@x.test", "pw-admin").AccessToken
	eventID, nominees := openVotingRound(t, e, adminTok)

	voterA, tokA := newVoter(t, e, "a@x.test")
	voterB, tokB := newVoter(t, e, "b@x.test")
	_, tokC := newVoter(t, e, "c@x.test")
	voterD, tokD := newVoter(t, e, "d@x.test")
	for _, v := range []struct {
		tok    string
		nomIdx int
	}{{tokA, 0}, {tokB, 0}, {tokC, 1}, {tokD, 1}} {
		if code, res := vote(e, eventID, nominees[v.nomIdx], v.tok); code != 200 {
			t.Fatalf("vote failed (%d): %v", code, res)
		}
	}

	// B is deactivated and D soft-deleted before settlement: neither counts.
	if err := e.db.Table("users").Where("id = ?", voterB).Update("is_active", false).Error; err != nil {
		t.Fatal(err)
	}
	if err := e.db.Table("users").Where("id = ?", voterD).Update("deleted_at", time.Now().UTC()).Error; err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	setWindow(t, e, eventID, now.Add(-48*time.Hour), now.Add(-time.Hour))
	if code, res := e.do("POST", fmt.Sprintf("/api/admin/events/%d/finalize", eventID), nil, adminTok); code != 200 {
		t.Fatalf("finalize failed (%d): %v", code, res)
	}

	_, res := e.do("GET", fmt.Sprintf("/api/events/%d", eventID), nil, "")
	ev := res["event"].(map[string]any)
	if got := int(ev["total_votes"].(float64)); got != 2 {
		t.Errorf("expected total_votes 2 (active voters A and C only), got %d", got)
	}
	assertNomineeVotes(t, ev, nominees[0], 1)
	assertNomineeVotes(t, ev, nominees[1], 1)

	// A voter deleted after settlement leaves the published counts unchanged.
	if err := e.db.Table("users").Where("id = ?", voterA).Update("deleted_at", time.Now().UTC()).Error; err != nil {
		t.Fatal(err)
	}
	_, after := e.do("GET", fmt.Sprintf("/api/events/%d", eventID), nil, "")
	afterEv := after["event"].(map[string]any)
	if got := int(afterEv["total_votes"].(float64)); got != 2 {
		t.Errorf("expected total_votes unchanged (2) after voter deletion, got %d", got)
	}
	assertNomineeVotes(t, afterEv, nominees[0], 1)
}

// assertNomineeVotes checks the settled vote_count of one nominee in a public event payload.
func assertNomineeVotes(t *testing.T, ev map[string]any, nomineeID uint, want int) {
	t.Helper()
	for _, n := range ev["nominees"].([]any) {
		nm := n.(map[string]any)
		if uint(nm["id"].(float64)) == nomineeID {
			if got := int(nm["vote_count"].(float64)); got != want {
				t.Errorf("nominee %d: expected vote_count %d, got %d", nomineeID, want, got)
			}
			return
		}
	}
	t.Errorf("nominee %d missing from event payload", nomineeID)
}

func TestEventsParallelFinalizeWritesOneAwardSet(t *testing.T) {
	e := newEnv(t)
	e.seedAdmin("admin@x.test", "pw-admin")
	adminTok := e.loginPair("admin@x.test", "pw-admin").AccessToken
	eventID, nominees := openVotingRound(t, e, adminTok)

	// Two voters tie on two nominees: two co-winners.
	_, tokA := newVoter(t, e, "a@x.test")
	_, tokB := newVoter(t, e, "b@x.test")
	if code, res := vote(e, eventID, nominees[0], tokA); code != 200 {
		t.Fatalf("vote failed (%d): %v", code, res)
	}
	if code, res := vote(e, eventID, nominees[1], tokB); code != 200 {
		t.Fatalf("vote failed (%d): %v", code, res)
	}

	now := time.Now().UTC()
	setWindow(t, e, eventID, now.Add(-48*time.Hour), now.Add(-time.Hour))

	const parallelFinalizers = 4
	codes := make([]int, parallelFinalizers)
	bodies := make([]map[string]any, parallelFinalizers)
	var wg sync.WaitGroup
	for i := 0; i < parallelFinalizers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			codes[idx], bodies[idx] = e.do("POST", fmt.Sprintf("/api/admin/events/%d/finalize", eventID), nil, adminTok)
		}(i)
	}
	wg.Wait()
	for i, c := range codes {
		if c != 200 {
			t.Errorf("parallel finalize %d: expected 200 (idempotent), got %d: %v", i, c, bodies[i])
		}
	}

	var awardCount int64
	if err := e.db.Model(&event_module.EventAward{}).Where("event_id = ?", eventID).Count(&awardCount).Error; err != nil {
		t.Fatal(err)
	}
	if awardCount != 2 {
		t.Errorf("expected exactly 2 EventAward rows after parallel finalize, got %d", awardCount)
	}
}

const (
	// lockWaitTimeout bounds how long an interleaving test waits for the other request to block on a row lock.
	lockWaitTimeout = 10 * time.Second
	lockPollEvery   = 10 * time.Millisecond
)

// waitForLockWaiter polls pg_stat_activity until a session of the test database waits on a lock,
// so the interleaving tests know the racing request has reached the row lock.
func waitForLockWaiter(t *testing.T, e *env) {
	t.Helper()
	deadline := time.Now().Add(lockWaitTimeout)
	for {
		var waiting int64
		if err := e.db.Raw(`SELECT count(*) FROM pg_stat_activity
			WHERE datname = current_database() AND wait_event_type = 'Lock'`).Scan(&waiting).Error; err != nil {
			t.Fatal(err)
		}
		if waiting > 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("no session started waiting on a lock")
		}
		time.Sleep(lockPollEvery)
	}
}

// Finalize takes the row lock first; a vote whose share lock is already held must be counted.
// The in-flight vote is inserted inside the test's own share-locked transaction, standing in for
// a vote that was admitted while the round was open.
func TestEventsFinalizeWaitsForInFlightVote(t *testing.T) {
	e := newEnv(t)
	e.seedAdmin("admin@x.test", "pw-admin")
	adminTok := e.loginPair("admin@x.test", "pw-admin").AccessToken
	voterID, _ := newVoter(t, e, "voter@x.test")
	eventID, nominees := openVotingRound(t, e, adminTok)

	now := time.Now().UTC()
	setWindow(t, e, eventID, now.Add(-48*time.Hour), now.Add(-time.Hour))

	tx := e.db.Begin()
	t.Cleanup(func() { tx.Rollback() })
	if err := tx.Exec("SELECT id FROM events WHERE id = ? FOR SHARE", eventID).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Create(&event_module.EventVote{EventID: eventID, UserID: voterID, NomineeID: nominees[0], CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	var code int
	var body map[string]any
	wg.Add(1)
	go func() {
		defer wg.Done()
		code, body = e.do("POST", fmt.Sprintf("/api/admin/events/%d/finalize", eventID), nil, adminTok)
	}()
	waitForLockWaiter(t, e)
	if err := tx.Commit().Error; err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	if code != 200 {
		t.Fatalf("finalize after in-flight vote failed (%d): %v", code, body)
	}

	_, res := e.do("GET", fmt.Sprintf("/api/events/%d", eventID), nil, "")
	ev := res["event"].(map[string]any)
	if got := int(ev["total_votes"].(float64)); got != 1 {
		t.Errorf("expected the in-flight vote counted (total_votes 1), got %d", got)
	}
	assertNomineeVotes(t, ev, nominees[0], 1)
}

// The reverse interleave: a vote that waits on a finalization in progress is refused once it commits.
// The test holds the row lock with an UPDATE that sets is_finalized, the effect finalizeTx commits.
func TestEventsVoteWaitsForFinalizeThenIsClosed(t *testing.T) {
	e := newEnv(t)
	e.seedAdmin("admin@x.test", "pw-admin")
	adminTok := e.loginPair("admin@x.test", "pw-admin").AccessToken
	voterID, voterTok := newVoter(t, e, "voter@x.test")
	eventID, nominees := openVotingRound(t, e, adminTok)

	tx := e.db.Begin()
	t.Cleanup(func() { tx.Rollback() })
	if err := tx.Exec("SELECT id FROM events WHERE id = ? FOR UPDATE", eventID).Error; err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	var code int
	var body map[string]any
	wg.Add(1)
	go func() {
		defer wg.Done()
		code, body = vote(e, eventID, nominees[0], voterTok)
	}()
	waitForLockWaiter(t, e)
	if err := tx.Model(&event_module.Event{}).Where("id = ?", eventID).Update("is_finalized", true).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	if code != 409 || body["code"] != "VOTING_CLOSED" {
		t.Errorf("expected 409 VOTING_CLOSED for a vote racing finalization, got %d: %v", code, body)
	}

	var votes int64
	if err := e.db.Model(&event_module.EventVote{}).Where("event_id = ? AND user_id = ?", eventID, voterID).Count(&votes).Error; err != nil {
		t.Fatal(err)
	}
	if votes != 0 {
		t.Errorf("expected no vote recorded after finalization, got %d", votes)
	}
}

// Once a vote exists, nominees cannot be replaced, even if the round's start was moved into the future.
func TestEventsNomineesLockedOnceVotesExist(t *testing.T) {
	e := newEnv(t)
	e.seedAdmin("admin@x.test", "pw-admin")
	adminTok := e.loginPair("admin@x.test", "pw-admin").AccessToken
	_, voterTok := newVoter(t, e, "voter@x.test")
	eventID, nominees := openVotingRound(t, e, adminTok)
	if code, res := vote(e, eventID, nominees[0], voterTok); code != 200 {
		t.Fatalf("vote failed (%d): %v", code, res)
	}

	// Moving the start into the future defeats the window check, so only the vote guard can refuse.
	period := event_module.NextMonthlyPeriod(time.Now())
	now := time.Now().UTC()
	starts, ends := now.Add(2*time.Hour), now.Add(48*time.Hour)
	setWindow(t, e, eventID, starts, ends)
	custom := []map[string]any{
		{"media_id": 701, "media_type": "movie", "media_title": "Nominee One", "position": 1},
	}

	code, res := e.do("PUT", fmt.Sprintf("/api/admin/events/%d/nominees", eventID), map[string]any{"nominees": custom}, adminTok)
	if code != 409 || res["code"] != "VOTING_LOCKED" {
		t.Errorf("expected 409 VOTING_LOCKED replacing nominees via /nominees, got %d: %v", code, res)
	}

	body := roundBody(period, starts, ends, "published")
	body["nominees"] = custom
	code, res = e.do("PUT", fmt.Sprintf("/api/admin/events/%d", eventID), body, adminTok)
	if code != 409 || res["code"] != "VOTING_LOCKED" {
		t.Errorf("expected 409 VOTING_LOCKED replacing nominees via update, got %d: %v", code, res)
	}

	var count int64
	if err := e.db.Model(&event_module.EventNominee{}).Where("event_id = ?", eventID).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != int64(len(nominees)) {
		t.Errorf("expected the %d original nominees untouched, got %d", len(nominees), count)
	}
}

// Only a published round can be finalized; a draft is refused even after its window has ended.
func TestEventsFinalizeRefusesDraftRound(t *testing.T) {
	e := newEnv(t)
	e.seedAdmin("admin@x.test", "pw-admin")
	adminTok := e.loginPair("admin@x.test", "pw-admin").AccessToken
	useFakeVotingTMDB(t, 6)
	period := event_module.NextMonthlyPeriod(time.Now())

	code, genRes := e.do("POST", "/api/admin/events/generate-monthly?media_type=movie&period="+period, nil, adminTok)
	if code != 201 {
		t.Fatalf("generate-monthly failed (%d): %v", code, genRes)
	}
	eventID := uint(genRes["event"].(map[string]any)["id"].(float64))

	now := time.Now().UTC()
	setWindow(t, e, eventID, now.Add(-48*time.Hour), now.Add(-time.Hour))
	code, res := e.do("POST", fmt.Sprintf("/api/admin/events/%d/finalize", eventID), nil, adminTok)
	if code != 409 || res["code"] != "EVENT_NOT_PUBLISHED" {
		t.Errorf("expected 409 EVENT_NOT_PUBLISHED finalizing a draft, got %d: %v", code, res)
	}

	var ev event_module.Event
	if err := e.db.First(&ev, eventID).Error; err != nil {
		t.Fatal(err)
	}
	if ev.IsFinalized {
		t.Error("draft round must stay unfinalized after a refused finalize")
	}
}

// A yearly round is not generated copy-less: a blank title is refused.
func TestEventsYearlyRoundRequiresTitle(t *testing.T) {
	e := newEnv(t)
	e.seedAdmin("admin@x.test", "pw-admin")
	adminTok := e.loginPair("admin@x.test", "pw-admin").AccessToken

	now := time.Now().UTC()
	body := map[string]any{
		"type":           "vote_yearly",
		"media_type":     "movie",
		"period":         fmt.Sprintf("%d", now.Year()),
		"title_en":       "",
		"title_th":       "",
		"starts_at":      now.Add(time.Hour).Format(time.RFC3339),
		"ends_at":        now.Add(48 * time.Hour).Format(time.RFC3339),
		"show_popup":     false,
		"popup_priority": 0,
		"status":         "draft",
	}
	code, res := e.do("POST", "/api/admin/events", body, adminTok)
	if code != 400 || res["code"] != "INVALID_INPUT" {
		t.Fatalf("expected 400 INVALID_INPUT for a blank yearly title, got %d: %v", code, res)
	}
	if msg, _ := res["error"].(string); !strings.Contains(msg, "title_en") {
		t.Errorf("expected the error to name title_en, got %q", msg)
	}
}

// movieNominee is one admin-supplied candidate for a movie round.
func movieNominee(mediaID int, title string) map[string]any {
	return map[string]any{"media_id": mediaID, "media_type": "movie", "media_title": title}
}

// TestEventsPublishRequiresTwoNominees: a voting round with fewer than two titles can be saved as a
// draft but not published, on create or on update.
func TestEventsPublishRequiresTwoNominees(t *testing.T) {
	e := newEnv(t)
	e.seedAdmin("admin@x.test", "pw-admin")
	adminTok := e.loginPair("admin@x.test", "pw-admin").AccessToken

	now := time.Now().UTC()
	period := event_module.NextMonthlyPeriod(now.AddDate(0, 3, 0))
	// Future window: once a round opens, its nominees are locked and the final publish would be refused.
	starts, ends := now.Add(48*time.Hour), now.Add(72*time.Hour)

	draft := roundBody(period, starts, ends, "draft")
	draft["nominees"] = []any{movieNominee(901, "Lone Pick")}
	code, res := e.do("POST", "/api/admin/events", draft, adminTok)
	if code != 201 {
		t.Fatalf("draft with one nominee should save, got %d: %v", code, res)
	}
	eventID := uint(res["event"].(map[string]any)["id"].(float64))

	// A second period: the draft already holds this one (type, media type, period is unique).
	otherPeriod := event_module.NextMonthlyPeriod(now.AddDate(0, 5, 0))
	published := roundBody(otherPeriod, starts, ends, "published")
	published["nominees"] = []any{movieNominee(901, "Lone Pick")}
	code, res = e.do("POST", "/api/admin/events", published, adminTok)
	if code != 422 || res["code"] != "NOT_ENOUGH_NOMINEES" {
		t.Fatalf("create published with one nominee: want 422 NOT_ENOUGH_NOMINEES, got %d: %v", code, res)
	}

	code, res = e.do("PUT", fmt.Sprintf("/api/admin/events/%d", eventID), roundBody(period, starts, ends, "published"), adminTok)
	if code != 422 || res["code"] != "NOT_ENOUGH_NOMINEES" {
		t.Fatalf("publish with one stored nominee: want 422 NOT_ENOUGH_NOMINEES, got %d: %v", code, res)
	}

	// The refused publish rolled back: the round is still a draft.
	var status string
	if err := e.db.Raw("SELECT status FROM events WHERE id = ?", eventID).Scan(&status).Error; err != nil {
		t.Fatal(err)
	}
	if status != "draft" {
		t.Fatalf("status after refused publish = %q, want draft", status)
	}

	twoNominees := roundBody(period, starts, ends, "published")
	twoNominees["nominees"] = []any{movieNominee(901, "Lone Pick"), movieNominee(902, "Second Pick")}
	code, res = e.do("PUT", fmt.Sprintf("/api/admin/events/%d", eventID), twoNominees, adminTok)
	if code != 200 {
		t.Fatalf("publish with two nominees should succeed, got %d: %v", code, res)
	}
}

// TestEventsSetNomineesKeepsPublishedRoundPublishable: a published upcoming round cannot be cut
// below two titles through the nominees endpoint, and the refused edit leaves the nominees intact.
func TestEventsSetNomineesKeepsPublishedRoundPublishable(t *testing.T) {
	e := newEnv(t)
	e.seedAdmin("admin@x.test", "pw-admin")
	adminTok := e.loginPair("admin@x.test", "pw-admin").AccessToken

	now := time.Now().UTC()
	period := event_module.NextMonthlyPeriod(now.AddDate(0, 3, 0))
	starts, ends := now.Add(48*time.Hour), now.Add(72*time.Hour)

	published := roundBody(period, starts, ends, "published")
	published["nominees"] = []any{movieNominee(911, "Pick A"), movieNominee(912, "Pick B")}
	code, res := e.do("POST", "/api/admin/events", published, adminTok)
	if code != 201 {
		t.Fatalf("publish with two nominees should succeed, got %d: %v", code, res)
	}
	eventID := uint(res["event"].(map[string]any)["id"].(float64))

	code, res = e.do("PUT", fmt.Sprintf("/api/admin/events/%d/nominees", eventID),
		map[string]any{"nominees": []any{movieNominee(911, "Pick A")}}, adminTok)
	if code != 422 || res["code"] != "NOT_ENOUGH_NOMINEES" {
		t.Fatalf("set one nominee on published round: want 422 NOT_ENOUGH_NOMINEES, got %d: %v", code, res)
	}

	var count int64
	if err := e.db.Raw("SELECT COUNT(*) FROM event_nominees WHERE event_id = ?", eventID).Scan(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("nominee count after refused edit = %d, want 2 (rolled back)", count)
	}
}

// TestEventsMediaTypeChangeRevalidatesNominees: switching a round's media type while its stored
// nominees are the other type is refused with 400, and nothing changes.
func TestEventsMediaTypeChangeRevalidatesNominees(t *testing.T) {
	e := newEnv(t)
	e.seedAdmin("admin@x.test", "pw-admin")
	adminTok := e.loginPair("admin@x.test", "pw-admin").AccessToken

	// A window that has not started: an open round is locked and refuses any type or media change first.
	now := time.Now().UTC()
	period := event_module.NextMonthlyPeriod(now.AddDate(0, 4, 0))
	starts, ends := now.Add(48*time.Hour), now.Add(72*time.Hour)
	body := roundBody(period, starts, ends, "draft")
	body["nominees"] = []any{movieNominee(911, "Movie A"), movieNominee(912, "Movie B")}
	code, res := e.do("POST", "/api/admin/events", body, adminTok)
	if code != 201 {
		t.Fatalf("create movie round failed (%d): %v", code, res)
	}
	eventID := uint(res["event"].(map[string]any)["id"].(float64))

	switched := roundBody(period, starts, ends, "draft")
	switched["media_type"] = "tv"
	code, res = e.do("PUT", fmt.Sprintf("/api/admin/events/%d", eventID), switched, adminTok)
	if code != 400 || res["code"] != "INVALID_INPUT" {
		t.Fatalf("media_type change over movie nominees: want 400 INVALID_INPUT, got %d: %v", code, res)
	}

	var mediaType string
	if err := e.db.Raw("SELECT media_type FROM events WHERE id = ?", eventID).Scan(&mediaType).Error; err != nil {
		t.Fatal(err)
	}
	if mediaType != "movie" {
		t.Fatalf("media_type after refused update = %q, want movie", mediaType)
	}
}

// TestEventsVoteRateLimit: one user may cast at most VoteRateMax votes per window; the next is 429.
func TestEventsVoteRateLimit(t *testing.T) {
	e := newEnv(t)
	e.seedAdmin("admin@x.test", "pw-admin")
	adminTok := e.loginPair("admin@x.test", "pw-admin").AccessToken
	eventID, nominees := openVotingRound(t, e, adminTok)
	_, voterTok := newVoter(t, e, "rate-voter@x.test")

	for i := 1; i <= event_module.VoteRateMax; i++ {
		if code, res := vote(e, eventID, nominees[0], voterTok); code != 200 {
			t.Fatalf("vote %d of %d should pass, got %d: %v", i, event_module.VoteRateMax, code, res)
		}
	}
	code, res := vote(e, eventID, nominees[0], voterTok)
	if code != 429 || res["code"] != "RATE_LIMITED" {
		t.Fatalf("vote over the budget: want 429 RATE_LIMITED, got %d: %v", code, res)
	}

	// The budget is per user: another voter is unaffected.
	_, otherTok := newVoter(t, e, "rate-other@x.test")
	if code, res := vote(e, eventID, nominees[0], otherTok); code != 200 {
		t.Fatalf("second user should not share the budget, got %d: %v", code, res)
	}
}

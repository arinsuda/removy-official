//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"
)

const (
	feedScopeAll       = "all"
	feedScopeFollowing = "following"
	feedScopeMe        = "me"
)

func (e *env) seedActivity(actor uint, visibility string, reviewID *uint) uint {
	e.t.Helper()
	var id uint
	err := e.db.Raw(`INSERT INTO activity_events (actor_id, type, review_id, message, visibility, is_visible, created_at, updated_at)
		VALUES (?, 'review_created', ?, 'msg', ?, true, now(), now()) RETURNING id`, actor, reviewID, visibility).Scan(&id).Error
	if err != nil {
		e.t.Fatal(err)
	}
	return id
}

func (e *env) seedReview(uid uint, mediaID int) uint {
	e.t.Helper()
	var id uint
	err := e.db.Raw(`INSERT INTO reviews (created_at, updated_at, user_id, media_id, media_type, rating, body, visibility)
		VALUES (now(), now(), ?, ?, 'movie', 4.0, 'body', 'public') RETURNING id`, uid, mediaID).Scan(&id).Error
	if err != nil {
		e.t.Fatal(err)
	}
	return id
}

func feedIDs(body map[string]any) map[uint]map[string]any {
	out := map[uint]map[string]any{}
	items, _ := body["items"].([]any)
	for _, it := range items {
		m := it.(map[string]any)
		out[uint(m["id"].(float64))] = m
	}
	return out
}

// Feed scopes: privacy never widens, own activity only under all/me, like state is batched.
func TestFeedScopeAndViewerLiked(t *testing.T) {
	e := newEnv(t)
	viewer := seedUser(t, e.db, "viewer@x.test", true, "password-one1")
	followed := seedUser(t, e.db, "followed@x.test", true, "password-two2")
	stranger := seedUser(t, e.db, "stranger@x.test", true, "password-three3")
	tok := e.loginPair("viewer@x.test", "password-one1").AccessToken
	if err := e.db.Exec(`INSERT INTO user_follows (follower_id, followee_id, status, created_at, updated_at) VALUES (?, ?, 'accepted', now(), now())`, viewer, followed).Error; err != nil {
		t.Fatal(err)
	}

	rev := e.seedReview(followed, 1001)
	unliked := e.seedReview(followed, 1002)
	if err := e.db.Exec(`INSERT INTO review_likes (review_id, user_id, created_at) VALUES (?, ?, now())`, rev, viewer).Error; err != nil {
		t.Fatal(err)
	}
	followedPub := e.seedActivity(followed, "public", &rev)
	followedUnliked := e.seedActivity(followed, "public", &unliked)
	followedPrivate := e.seedActivity(followed, "private", nil)
	strangerPub := e.seedActivity(stranger, "public", nil)
	strangerFollowers := e.seedActivity(stranger, "followers", nil)
	mine := e.seedActivity(viewer, "private", nil)
	followedFollowersOnly := e.seedActivity(followed, "followers", nil)

	pending := seedUser(t, e.db, "pending@x.test", true, "password-four4")
	if err := e.db.Exec(`INSERT INTO user_follows (follower_id, followee_id, status, created_at, updated_at) VALUES (?, ?, 'pending', now(), now())`, viewer, pending).Error; err != nil {
		t.Fatal(err)
	}
	pendingPub := e.seedActivity(pending, "public", nil)

	deactivated := seedUser(t, e.db, "gone@x.test", true, "password-five5")
	if err := e.db.Exec(`INSERT INTO user_follows (follower_id, followee_id, status, created_at, updated_at) VALUES (?, ?, 'accepted', now(), now())`, viewer, deactivated).Error; err != nil {
		t.Fatal(err)
	}
	deactivatedPub := e.seedActivity(deactivated, "public", nil)
	if err := e.db.Exec(`UPDATE users SET is_active = false WHERE id = ?`, deactivated).Error; err != nil {
		t.Fatal(err)
	}

	get := func(scope string) (int, map[uint]map[string]any) {
		path := "/api/feed"
		if scope != "" {
			path += "?scope=" + scope
		}
		code, body := e.do("GET", path, nil, tok)
		return code, feedIDs(body)
	}
	for _, scope := range []string{"", feedScopeAll, feedScopeFollowing, feedScopeMe} {
		code, got := get(scope)
		if code != http.StatusOK {
			t.Fatalf("scope %q: %d", scope, code)
		}
		for _, id := range []uint{strangerPub, strangerFollowers, followedPrivate, pendingPub, deactivatedPub} {
			if _, ok := got[id]; ok {
				t.Fatalf("scope %q leaked activity %d", scope, id)
			}
		}
		_, hasFollowed := got[followedPub]
		_, hasMine := got[mine]
		wantFollowed := scope != feedScopeMe
		if _, ok := got[followedFollowersOnly]; ok != wantFollowed {
			t.Fatalf("scope %q: followers-only row of a followed user visible=%v", scope, ok)
		}
		wantMine := scope == "" || scope == feedScopeAll || scope == feedScopeMe
		if hasFollowed != wantFollowed || hasMine != wantMine {
			t.Fatalf("scope %q: followed=%v mine=%v", scope, hasFollowed, hasMine)
		}
		if wantFollowed {
			if got[followedPub]["viewer_liked"] != true || got[followedUnliked]["viewer_liked"] != false {
				t.Fatalf("scope %q: viewer_liked wrong: %v %v", scope, got[followedPub]["viewer_liked"], got[followedUnliked]["viewer_liked"])
			}
		}
	}

	_, prof := e.do("GET", fmt.Sprintf("/api/users/%d/activities", followed), nil, tok)
	profItems := feedIDs(prof)
	if profItems[followedPub]["viewer_liked"] != true || profItems[followedUnliked]["viewer_liked"] != false {
		t.Fatalf("profile activities viewer_liked wrong: %v", profItems)
	}

	if code, _ := get("bogus"); code != http.StatusBadRequest {
		t.Fatalf("invalid scope must be 400, got %d", code)
	}
	code, _ := e.do("GET", fmt.Sprintf("/api/feed/new-count?after_id=0&scope=%s", "bogus"), nil, tok)
	if code != http.StatusBadRequest {
		t.Fatalf("new-count invalid scope must be 400, got %d", code)
	}
}

// New-count compares (created_at, id) of the newest item the client holds, and ignores own rows.
func TestFeedNewCountCursorAndOwnRows(t *testing.T) {
	e := newEnv(t)
	viewer := seedUser(t, e.db, "viewer@x.test", true, "password-one1")
	followed := seedUser(t, e.db, "followed@x.test", true, "password-two2")
	tok := e.loginPair("viewer@x.test", "password-one1").AccessToken
	if err := e.db.Exec(`INSERT INTO user_follows (follower_id, followee_id, status, created_at, updated_at) VALUES (?, ?, 'accepted', now(), now())`, viewer, followed).Error; err != nil {
		t.Fatal(err)
	}
	older := e.seedActivity(followed, "public", nil)
	newest := e.seedActivity(followed, "public", nil)

	count := func(scope string) float64 {
		_, body := e.do("GET", fmt.Sprintf("/api/feed/new-count?after_id=%d&scope=%s", newest, scope), nil, tok)
		return body["count"].(float64)
	}
	if got := count(feedScopeAll); got != 0 {
		t.Fatalf("nothing newer yet, got %v", got)
	}

	e.seedActivity(viewer, "public", nil)
	if got := count(feedScopeAll); got != 0 {
		t.Fatalf("own activity must not count, got %v", got)
	}
	if got := count(feedScopeMe); got != 0 {
		t.Fatalf("me scope never counts own rows, got %v", got)
	}

	// CreateOrRestore re-stamps created_at but keeps the (lower) id: still new.
	if err := e.db.Exec(`UPDATE activity_events SET created_at = now() + interval '1 minute' WHERE id = ?`, older).Error; err != nil {
		t.Fatal(err)
	}
	for _, scope := range []string{feedScopeAll, feedScopeFollowing} {
		if got := count(scope); got != 1 {
			t.Fatalf("restored older id must count under %s, got %v", scope, got)
		}
	}
}

//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

// patchProfile sends a multipart profile update.
func (e *env) patchProfile(uid uint, fields map[string]string, token string) (int, map[string]any) {
	e.t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	_ = w.Close()
	req := httptest.NewRequest("PATCH", fmt.Sprintf("/api/users/%d", uid), &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	resp, err := e.app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
	if err != nil {
		e.t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

// S1: only the owner receives the email of a profile.
func TestProfileEmailHiddenFromOthers(t *testing.T) {
	e := newEnv(t)
	owner := seedUser(t, e.db, "owner@x.test", true, "password-one1")
	seedUser(t, e.db, "visitor@x.test", true, "password-two2")
	own := e.loginPair("owner@x.test", "password-one1")
	vis := e.loginPair("visitor@x.test", "password-two2")

	_, body := e.do("GET", fmt.Sprintf("/api/users/%d", owner), nil, own.AccessToken)
	u, _ := body["user"].(map[string]any)
	if u["email"] != "owner@x.test" {
		t.Fatalf("owner must see own email: %v", u)
	}
	_, body = e.do("GET", fmt.Sprintf("/api/users/%d", owner), nil, vis.AccessToken)
	u, _ = body["user"].(map[string]any)
	if s, _ := u["email"].(string); s != "" {
		t.Fatalf("email leaked to visitor: %v", u["email"])
	}
	if u["verified_email_at"] != nil {
		t.Fatalf("verified_email_at leaked to visitor: %v", u["verified_email_at"])
	}
}

// S2: BMOL of a private account is visible only to the owner and accepted followers.
func TestBMOLGatedByPrivacy(t *testing.T) {
	e := newEnv(t)
	owner := seedUser(t, e.db, "owner@x.test", true, "password-one1")
	stranger := seedUser(t, e.db, "stranger@x.test", true, "password-two2")
	follower := seedUser(t, e.db, "follower@x.test", true, "password-thr3")
	if err := e.db.Exec(`UPDATE users SET is_private = true WHERE id = ?`, owner).Error; err != nil {
		t.Fatal(err)
	}
	e.db.Exec(`INSERT INTO bmol_items (created_at, updated_at, user_id, media_id, media_type, rank) VALUES (now(), now(), ?, 603, 'movie', 1)`, owner)
	e.db.Exec(`INSERT INTO user_follows (follower_id, followee_id, status, created_at, updated_at) VALUES (?, ?, 'pending', now(), now())`, follower, owner)

	path := fmt.Sprintf("/api/bmol/user/%d", owner)
	own := e.loginPair("owner@x.test", "password-one1")
	str := e.loginPair("stranger@x.test", "password-two2")
	fol := e.loginPair("follower@x.test", "password-thr3")
	_ = stranger

	if code, _ := e.do("GET", path, nil, str.AccessToken); code != 403 {
		t.Fatalf("stranger: want 403, got %d", code)
	}
	if code, _ := e.do("GET", path, nil, fol.AccessToken); code != 403 {
		t.Fatalf("pending follower: want 403, got %d", code)
	}
	e.db.Exec(`UPDATE user_follows SET status = 'accepted' WHERE follower_id = ?`, follower)
	if code, _ := e.do("GET", path, nil, fol.AccessToken); code != 200 {
		t.Fatalf("accepted follower: want 200, got %d", code)
	}
	if code, _ := e.do("GET", path, nil, own.AccessToken); code != 200 {
		t.Fatalf("owner: want 200, got %d", code)
	}
}

// Private profile sections answer 403 to strangers and pending followers, 200 to accepted
// followers and the owner, and 404 once the owner is deactivated.
func TestAchievementsAndLibraryGatedByPrivacy(t *testing.T) {
	e := newEnv(t)
	owner := seedUser(t, e.db, "owner@x.test", true, "password-one1")
	seedUser(t, e.db, "stranger@x.test", true, "password-two2")
	follower := seedUser(t, e.db, "follower@x.test", true, "password-thr3")
	if err := e.db.Exec(`UPDATE users SET is_private = true WHERE id = ?`, owner).Error; err != nil {
		t.Fatal(err)
	}
	e.db.Exec(`INSERT INTO user_follows (follower_id, followee_id, status, created_at, updated_at) VALUES (?, ?, 'pending', now(), now())`, follower, owner)

	str := e.loginPair("stranger@x.test", "password-two2")
	fol := e.loginPair("follower@x.test", "password-thr3")
	paths := []string{
		fmt.Sprintf("/api/users/%d/achievements", owner),
		fmt.Sprintf("/api/library/user/%d", owner),
	}
	for _, path := range paths {
		if code, body := e.do("GET", path, nil, str.AccessToken); code != 403 || body["error"] != "forbidden" {
			t.Fatalf("%s stranger: %d %v", path, code, body)
		}
		if code, _ := e.do("GET", path, nil, fol.AccessToken); code != 403 {
			t.Fatalf("%s pending follower: want 403, got %d", path, code)
		}
	}
	e.db.Exec(`UPDATE user_follows SET status = 'accepted' WHERE follower_id = ?`, follower)
	for _, path := range paths {
		if code, body := e.do("GET", path, nil, fol.AccessToken); code != 200 {
			t.Fatalf("%s accepted follower: %d %v", path, code, body)
		}
	}
	e.db.Exec(`UPDATE users SET is_active = false WHERE id = ?`, owner)
	for _, path := range paths {
		if code, _ := e.do("GET", path, nil, fol.AccessToken); code != 404 {
			t.Fatalf("%s deactivated owner: want 404, got %d", path, code)
		}
	}
}

// S3: empty keys clear fields; over-long or out-of-range values are rejected.
func TestUpdateProfileClearsAndValidates(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "me@x.test", true, "password-one1")
	tok := e.loginPair("me@x.test", "password-one1").AccessToken

	if code, b := e.patchProfile(uid, map[string]string{"display_name": "Me", "bio": "hello", "date_of_birth": "1990-05-01", "gender_other": "x"}, tok); code != 200 {
		t.Fatalf("set: %d %v", code, b)
	}
	if code, b := e.patchProfile(uid, map[string]string{"display_name": "", "bio": "", "date_of_birth": "", "gender_other": ""}, tok); code != 200 {
		t.Fatalf("clear: %d %v", code, b)
	}
	var row struct {
		DisplayName, Bio, GenderOther *string
		DateOfBirth                   *time.Time
	}
	e.db.Raw(`SELECT display_name, bio, gender_other, date_of_birth FROM users WHERE id = ?`, uid).Scan(&row)
	if row.DisplayName != nil || row.Bio != nil || row.GenderOther != nil || row.DateOfBirth != nil {
		t.Fatalf("fields not cleared: %+v", row)
	}

	long := func(n int) string { return string(bytes.Repeat([]byte("a"), n)) }
	for name, f := range map[string]map[string]string{
		"name too long":         {"display_name": long(61)},
		"bio too long":          {"bio": long(201)},
		"dob too early":         {"date_of_birth": "1899-12-31"},
		"gender_other too long": {"gender_other": long(101)},
		"dob future":            {"date_of_birth": time.Now().AddDate(1, 0, 0).Format("2006-01-02")},
	} {
		if code, _ := e.patchProfile(uid, f, tok); code != 400 {
			t.Errorf("%s: want 400, got %d", name, code)
		}
	}
}

// Cancelling a pending follow never changes follower_count (only accepted rows count).
func TestCancelPendingFollowKeepsFollowerCount(t *testing.T) {
	e := newEnv(t)
	owner := seedUser(t, e.db, "owner@x.test", true, "password-one1")
	f := seedUser(t, e.db, "f@x.test", true, "password-two2")
	count := func() int {
		var n int
		e.db.Raw(`SELECT COALESCE(follower_count,0) FROM user_stats WHERE user_id = ?`, owner).Scan(&n)
		return n
	}
	e.db.Exec(`INSERT INTO user_follows (follower_id, followee_id, status, created_at, updated_at) VALUES (?, ?, 'pending', now(), now())`, f, owner)
	before := count()
	e.db.Exec(`DELETE FROM user_follows WHERE follower_id = ? AND followee_id = ?`, f, owner)
	if after := count(); after != before || after != 0 {
		t.Fatalf("follower_count before=%d after=%d, want 0/0", before, after)
	}
}

// Library counts (stats) follow the library rule; a missing user is 404, not 500.
func TestStatsGatedAndMissingUser(t *testing.T) {
	e := newEnv(t)
	owner := seedUser(t, e.db, "owner@x.test", true, "password-one1")
	follower := seedUser(t, e.db, "follower@x.test", true, "password-two2")
	seedUser(t, e.db, "stranger@x.test", true, "password-thr3")
	e.db.Exec(`UPDATE users SET is_private = true WHERE id = ?`, owner)
	e.db.Exec(`INSERT INTO user_follows (follower_id, followee_id, status, created_at, updated_at) VALUES (?, ?, 'accepted', now(), now())`, follower, owner)
	own := e.loginPair("owner@x.test", "password-one1")
	fol := e.loginPair("follower@x.test", "password-two2")
	str := e.loginPair("stranger@x.test", "password-thr3")

	path := fmt.Sprintf("/api/users/%d/stats", owner)
	if code, _ := e.do("GET", path, nil, str.AccessToken); code != 403 {
		t.Fatalf("stranger: want 403, got %d", code)
	}
	for name, tok := range map[string]string{"follower": fol.AccessToken, "owner": own.AccessToken} {
		if code, _ := e.do("GET", path, nil, tok); code != 200 {
			t.Fatalf("%s: want 200, got %d", name, code)
		}
	}
	if code, _ := e.do("GET", "/api/bmol/user/999999", nil, own.AccessToken); code != 404 {
		t.Fatalf("bmol missing user: want 404, got %d", code)
	}
	if code, _ := e.do("GET", "/api/users/999999/stats", nil, own.AccessToken); code != 404 {
		t.Fatalf("stats missing user: want 404, got %d", code)
	}
}

// B6 regression test: preserves all 5 visibility cases on GET /api/reviews/user/:userId
// (owner, approved follower, other user, guest, private profile).
func TestUserReviewsVisibilityRegression(t *testing.T) {
	useFakeTMDB(t)
	e := newEnv(t)
	owner := seedUser(t, e.db, "rev-owner@x.test", true, "password-one1")
	follower := seedUser(t, e.db, "rev-follower@x.test", true, "password-two2")
	pending := seedUser(t, e.db, "rev-pending@x.test", true, "password-thr3")
	stranger := seedUser(t, e.db, "rev-stranger@x.test", true, "password-fou4")
	_ = stranger

	insertRev := func(mediaID int, vis, body string) {
		if err := e.db.Exec(`INSERT INTO reviews (created_at, updated_at, user_id, media_id, media_type, rating, body, is_public, visibility)
			VALUES (now(), now(), ?, ?, 'movie', 4.0, ?, ?, ?)`, owner, mediaID, body, vis == "public", vis).Error; err != nil {
			t.Fatal(err)
		}
	}
	insertRev(101, "public", "pub-rev")
	insertRev(102, "followers", "fol-rev")
	insertRev(103, "private", "priv-rev")

	e.db.Exec(`INSERT INTO user_follows (follower_id, followee_id, status, created_at, updated_at) VALUES (?, ?, 'accepted', now(), now())`, follower, owner)
	e.db.Exec(`INSERT INTO user_follows (follower_id, followee_id, status, created_at, updated_at) VALUES (?, ?, 'pending', now(), now())`, pending, owner)

	ownTok := e.loginPair("rev-owner@x.test", "password-one1").AccessToken
	folTok := e.loginPair("rev-follower@x.test", "password-two2").AccessToken
	pendTok := e.loginPair("rev-pending@x.test", "password-thr3").AccessToken
	strTok := e.loginPair("rev-stranger@x.test", "password-fou4").AccessToken
	path := fmt.Sprintf("/api/reviews/user/%d", owner)

	bodiesOf := func(out map[string]any) map[string]bool {
		res := map[string]bool{}
		revs, _ := out["reviews"].([]any)
		for _, r := range revs {
			m := r.(map[string]any)
			res[m["body"].(string)] = true
		}
		return res
	}

	// 1. Guest -> 401
	if code, _ := e.do("GET", path, nil, ""); code != 401 {
		t.Fatalf("guest GET %s: got %d, want 401", path, code)
	}

	// 2. Owner -> 200, sees public + followers + private, and can filter by visibility=private
	code, out := e.do("GET", path, nil, ownTok)
	if code != 200 {
		t.Fatalf("owner GET %s: %d %v", path, code, out)
	}
	ownBodies := bodiesOf(out)
	if !ownBodies["pub-rev"] || !ownBodies["fol-rev"] || !ownBodies["priv-rev"] {
		t.Fatalf("owner bodies = %v, want all 3", ownBodies)
	}
	code, out = e.do("GET", path+"?visibility=private", nil, ownTok)
	if code != 200 {
		t.Fatalf("owner GET private filter: %d %v", code, out)
	}
	ownPriv := bodiesOf(out)
	if len(ownPriv) != 1 || !ownPriv["priv-rev"] {
		t.Fatalf("owner private filter bodies = %v, want only priv-rev", ownPriv)
	}

	// 3. Other user (stranger) on public profile -> 200, sees only public even if visibility=private requested
	code, out = e.do("GET", path+"?visibility=private", nil, strTok)
	if code != 200 {
		t.Fatalf("stranger on public profile: %d %v", code, out)
	}
	strBodies := bodiesOf(out)
	if len(strBodies) != 1 || !strBodies["pub-rev"] {
		t.Fatalf("stranger bodies = %v, want only pub-rev", strBodies)
	}

	// 4 & 5. Switch owner to private profile: stranger & pending follower -> 403, approved follower & owner -> 200
	if err := e.db.Exec(`UPDATE users SET is_private = true WHERE id = ?`, owner).Error; err != nil {
		t.Fatal(err)
	}
	if code, _ := e.do("GET", path, nil, strTok); code != 403 {
		t.Fatalf("stranger on private profile: got %d, want 403", code)
	}
	if code, _ := e.do("GET", path, nil, pendTok); code != 403 {
		t.Fatalf("pending follower on private profile: got %d, want 403", code)
	}
	code, out = e.do("GET", path, nil, folTok)
	if code != 200 {
		t.Fatalf("approved follower on private profile: %d %v", code, out)
	}
	folBodies := bodiesOf(out)
	if len(folBodies) != 1 || !folBodies["pub-rev"] {
		t.Fatalf("approved follower bodies = %v, want only pub-rev", folBodies)
	}
	if code, _ := e.do("GET", path, nil, ownTok); code != 200 {
		t.Fatalf("owner on private profile: got %d, want 200", code)
	}
}

func TestUserReviewsPaginationAndFilters(t *testing.T) {
	useFakeTMDB(t)
	e := newEnv(t)
	owner := seedUser(t, e.db, "page-owner@x.test", true, "password-one1")
	ownTok := e.loginPair("page-owner@x.test", "password-one1").AccessToken

	for i := 0; i < 25; i++ {
		rating := 3.0
		body := fmt.Sprintf("review-%02d", i)
		if i == 0 {
			rating = 5.0
			body = "special 100% match_here"
		} else if i == 1 {
			rating = 1.5
			body = ""
		}
		if err := e.db.Exec(`INSERT INTO reviews (created_at, updated_at, user_id, media_id, media_type, rating, body, is_public, visibility)
			VALUES (now() - make_interval(secs => ?), now(), ?, ?, 'movie', ?, ?, true, 'public')`,
			i, owner, 200+i, rating, body).Error; err != nil {
			t.Fatal(err)
		}
	}

	path := fmt.Sprintf("/api/reviews/user/%d", owner)

	// Default page 1 (20 items, total 25, has_more true)
	code, out := e.do("GET", path, nil, ownTok)
	if code != 200 {
		t.Fatalf("page 1: %d %v", code, out)
	}
	revs, _ := out["reviews"].([]any)
	pg, _ := out["pagination"].(map[string]any)
	if len(revs) != 20 || pg["total"] != float64(25) || pg["has_more"] != true || pg["page"] != float64(1) || pg["page_size"] != float64(20) {
		t.Fatalf("page 1 unexpected: len=%d pg=%v", len(revs), pg)
	}

	// Page 2 (5 items, has_more false)
	code, out = e.do("GET", path+"?page=2", nil, ownTok)
	revs, _ = out["reviews"].([]any)
	pg, _ = out["pagination"].(map[string]any)
	if code != 200 || len(revs) != 5 || pg["has_more"] != false {
		t.Fatalf("page 2 unexpected: %d len=%d pg=%v", code, len(revs), pg)
	}

	// Filter high (rating >= 4.0) -> 1 item
	code, out = e.do("GET", path+"?filter=high", nil, ownTok)
	revs, _ = out["reviews"].([]any)
	if code != 200 || len(revs) != 1 {
		t.Fatalf("filter=high: %d len=%d", code, len(revs))
	}

	// Filter low (rating <= 2.5) -> 1 item
	code, out = e.do("GET", path+"?filter=low", nil, ownTok)
	revs, _ = out["reviews"].([]any)
	if code != 200 || len(revs) != 1 {
		t.Fatalf("filter=low: %d len=%d", code, len(revs))
	}

	// Filter withText -> 24 items (1 has empty body)
	code, out = e.do("GET", path+"?filter=withText", nil, ownTok)
	pg, _ = out["pagination"].(map[string]any)
	if code != 200 || pg["total"] != float64(24) {
		t.Fatalf("filter=withText: %d pg=%v", code, pg)
	}

	// Search with LIKE wildcards (% and _) escaped
	code, out = e.do("GET", path+"?search=100%25%20match_here", nil, ownTok)
	revs, _ = out["reviews"].([]any)
	if code != 200 || len(revs) != 1 {
		t.Fatalf("escaped search: %d len=%d", code, len(revs))
	}

	// Sort lowest -> first item has rating 1.5
	code, out = e.do("GET", path+"?sort=lowest", nil, ownTok)
	revs, _ = out["reviews"].([]any)
	if code != 200 || revs[0].(map[string]any)["rating"] != float64(1.5) {
		t.Fatalf("sort=lowest: %d first=%v", code, revs[0])
	}

	// Invalid sort / filter -> 400
	if code, _ := e.do("GET", path+"?sort=invalid", nil, ownTok); code != 400 {
		t.Fatalf("invalid sort: got %d, want 400", code)
	}
	if code, _ := e.do("GET", path+"?filter=invalid", nil, ownTok); code != 400 {
		t.Fatalf("invalid filter: got %d, want 400", code)
	}
}

// Summary counts cover the privacy/visibility scope but ignore rating filter, search and paging.
func TestUserReviewsSummaryScope(t *testing.T) {
	useFakeTMDB(t)
	e := newEnv(t)
	owner := seedUser(t, e.db, "sum-owner@x.test", true, "password-one1")
	other := seedUser(t, e.db, "sum-other@x.test", true, "password-two2")
	ownTok := e.loginPair("sum-owner@x.test", "password-one1").AccessToken
	othTok := e.loginPair("sum-other@x.test", "password-two2").AccessToken

	rows := []struct {
		media  int
		rating float64
		body   string
		vis    string
	}{
		{301, 5.0, "great", "public"},
		{302, 2.0, "", "public"},
		{303, 3.0, "meh", "public"},
		{304, 1.0, "secret", "private"},
	}
	for _, r := range rows {
		if err := e.db.Exec(`INSERT INTO reviews (created_at, updated_at, user_id, media_id, media_type, rating, body, is_public, visibility)
			VALUES (now(), now(), ?, ?, 'movie', ?, ?, ?, ?)`, owner, r.media, r.rating, r.body, r.vis == "public", r.vis).Error; err != nil {
			t.Fatal(err)
		}
	}
	path := fmt.Sprintf("/api/reviews/user/%d", owner)
	summaryOf := func(tok, query string) map[string]any {
		code, out := e.do("GET", path+query, nil, tok)
		if code != 200 {
			t.Fatalf("GET %s%s: %d %v", path, query, code, out)
		}
		s, _ := out["summary"].(map[string]any)
		return s
	}
	check := func(label string, s map[string]any, count, high, low, withText float64) {
		fc, _ := s["filter_counts"].(map[string]any)
		if s["count"] != count || fc["all"] != count || fc["high"] != high || fc["low"] != low || fc["withText"] != withText {
			t.Fatalf("%s summary = %v", label, s)
		}
	}

	check("owner", summaryOf(ownTok, ""), 4, 1, 2, 3)
	check("owner ignores filter/search/page", summaryOf(ownTok, "?filter=high&search=great&page_size=1"), 4, 1, 2, 3)
	check("non-owner sees public only", summaryOf(othTok, "?visibility=private"), 3, 1, 1, 2)
	_ = other
}

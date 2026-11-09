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

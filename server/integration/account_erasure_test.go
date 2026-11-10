//go:build integration

package integration

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/arinsuda/movie-hub/internal/admin_module"
	"github.com/arinsuda/movie-hub/internal/auth_module"
)

func (e *env) username(uid uint) string {
	e.t.Helper()
	var name string
	if err := e.db.Raw(`SELECT username FROM users WHERE id = ?`, uid).Scan(&name).Error; err != nil {
		e.t.Fatal(err)
	}
	return name
}

func (e *env) exec(sql string, args ...any) {
	e.t.Helper()
	if err := e.db.Exec(sql, args...).Error; err != nil {
		e.t.Fatalf("%s: %v", sql, err)
	}
}

func (e *env) erase(token string, target uint, confirm string) (int, map[string]any) {
	return e.do("POST", fmt.Sprintf("/api/admin/users/%d/erasure", target), map[string]any{"confirm_username": confirm}, token)
}

func (e *env) softDelete(token string, target uint) {
	e.t.Helper()
	if code, body := e.do("DELETE", fmt.Sprintf("/api/admin/users/%d", target), nil, token); code != 204 {
		e.t.Fatalf("soft delete: %d %v", code, body)
	}
}

// userRefCount sums rows in every base table whose user-referencing column holds uid.
func (e *env) userRefCount(uid uint) map[string]int64 {
	e.t.Helper()
	var cols []struct{ TableName, ColumnName string }
	err := e.db.Raw(`SELECT c.table_name, c.column_name FROM information_schema.columns c
		JOIN information_schema.tables t ON t.table_schema = c.table_schema AND t.table_name = c.table_name AND t.table_type = 'BASE TABLE'
		WHERE c.table_schema = 'public'
		AND c.column_name IN ('user_id', 'actor_id', 'follower_id', 'followee_id', 'local_user_id', 'admin_id', 'target_user_id')`).Scan(&cols).Error
	if err != nil || len(cols) == 0 {
		e.t.Fatalf("column sweep: n=%d err=%v", len(cols), err)
	}
	left := map[string]int64{}
	for _, c := range cols {
		if n := e.count(fmt.Sprintf(`SELECT count(*) FROM %s WHERE %s = ?`, c.TableName, c.ColumnName), uid); n != 0 {
			left[c.TableName+"."+c.ColumnName] = n
		}
	}
	return left
}

func TestAccountErasureRemovesEverything(t *testing.T) {
	e := newEnv(t)
	adminID := e.seedAdmin("admin@x.test", "admin-password1")
	admin := e.loginPair("admin@x.test", "admin-password1")
	other := seedUser(t, e.db, "other@x.test", true, "")
	third := seedUser(t, e.db, "third@x.test", true, "")
	const email, sub = "victim@gmail.com", "victim-sub"
	victim := seedUser(t, e.db, email, true, "")
	name := e.username(victim)
	e.exec(`UPDATE users SET avatar_url = ? WHERE id = ?`, fmt.Sprintf("avatars/%d/a.png", victim), victim)

	seedUserRefs(t, e.db, victim, other, "victim") // library, media like, list, notification, audit (as admin_id)
	e.seedGoogleIdentity(victim, sub, email)
	e.seedLinkRows(victim, "planzi-victim")
	e.seedRefreshToken(victim, "victim-rt")
	for _, pair := range [][2]uint{{victim, other}, {other, victim}} {
		e.exec(`INSERT INTO user_follows (follower_id, followee_id, created_at) VALUES (?, ?, now())`, pair[0], pair[1])
	}
	e.exec(`INSERT INTO notifications (created_at, updated_at, user_id, actor_id, type, category, message) VALUES (now(), now(), ?, ?, 'follow', 'social', 'm')`, victim, other)

	// The victim's reviews carry likes, helpful votes and comments from others.
	vr := e.seedReview(victim, 5001)
	e.seedActivity(victim, "public", &vr)
	e.exec(`INSERT INTO activity_events (actor_id, review_id, type, message, visibility, is_visible, created_at, updated_at)
		VALUES (?, ?, 'review_liked', 'm', 'public', true, now(), now())`, other, vr)
	e.exec(`INSERT INTO review_likes (review_id, user_id, created_at) VALUES (?, ?, now())`, vr, other)
	e.exec(`INSERT INTO review_helpfuls (review_id, user_id, created_at) VALUES (?, ?, now())`, vr, other)
	e.exec(`INSERT INTO review_comments (created_at, updated_at, review_id, user_id, body) VALUES (now(), now(), ?, ?, 'c')`, vr, other)

	// The victim's reactions on someone else's review, with counters that include them.
	or := e.seedReview(other, 5002)
	e.exec(`UPDATE reviews SET like_count = 2, helpful_count = 2, comment_count = 2 WHERE id = ?`, or)
	for _, u := range []uint{victim, third} {
		e.exec(`INSERT INTO review_likes (review_id, user_id, created_at) VALUES (?, ?, now())`, or, u)
		e.exec(`INSERT INTO review_helpfuls (review_id, user_id, created_at) VALUES (?, ?, now())`, or, u)
		e.exec(`INSERT INTO review_comments (created_at, updated_at, review_id, user_id, body) VALUES (now(), now(), ?, ?, 'c')`, or, u)
	}
	// An already soft-deleted comment was decremented when it was deleted: it must not count twice.
	e.exec(`INSERT INTO review_comments (created_at, updated_at, deleted_at, review_id, user_id, body) VALUES (now(), now(), now(), ?, ?, 'old')`, or, victim)

	// About the victim: activity naming them as target.
	e.exec(`INSERT INTO activity_events (actor_id, target_user_id, type, message, visibility, is_visible, created_at, updated_at)
		VALUES (?, ?, 'follow', 'm', 'public', true, now(), now())`, other, victim)

	if code, _ := e.do("DELETE", fmt.Sprintf("/api/admin/users/%d", victim), map[string]any{"reason": "spam by " + name}, admin.AccessToken); code != 204 {
		t.Fatalf("soft delete: %d", code)
	}
	if n := e.count(`SELECT count(*) FROM admin_audit_logs WHERE target_type = 'user' AND target_id = ? AND meta_data::text LIKE ?`, victim, "%"+name+"%"); n != 1 {
		t.Fatalf("precondition: delete audit log should hold the username, got %d", n)
	}

	code, body := e.do("POST", fmt.Sprintf("/api/admin/users/%d/erasure", victim),
		map[string]any{"confirm_username": name, "reason": "erased " + name}, admin.AccessToken)
	if code != 204 {
		t.Fatalf("erase: %d %v", code, body)
	}

	if n := e.count(`SELECT count(*) FROM users WHERE id = ?`, victim); n != 0 {
		t.Fatal("user row still there")
	}
	if left := e.userRefCount(victim); len(left) != 0 {
		t.Fatalf("rows still reference the erased user: %v", left)
	}
	for _, tbl := range []string{"review_likes", "review_helpfuls", "review_comments", "activity_events"} {
		if n := e.count(fmt.Sprintf(`SELECT count(*) FROM %s WHERE review_id = ?`, tbl), vr); n != 0 {
			t.Errorf("%s keeps %d rows of the erased user's review", tbl, n)
		}
	}
	if n := e.count(`SELECT count(*) FROM reviews WHERE user_id = ? OR id = ?`, victim, vr); n != 0 {
		t.Errorf("reviews left: %d", n)
	}

	var counters struct{ LikeCount, HelpfulCount, CommentCount int }
	e.db.Raw(`SELECT like_count, helpful_count, comment_count FROM reviews WHERE id = ?`, or).Scan(&counters)
	if counters.LikeCount != 1 || counters.HelpfulCount != 1 || counters.CommentCount != 1 {
		t.Errorf("counters on other's review = %+v, want 1/1/1", counters)
	}
	if n := e.count(`SELECT count(*) FROM review_likes WHERE review_id = ?`, or); n != 1 {
		t.Errorf("third's like lost: %d", n)
	}

	// Audit trail survives, without the person's identity.
	if n := e.count(`SELECT count(*) FROM admin_audit_logs WHERE target_type = 'user' AND target_id = ? AND action = 'USER_DELETED'`, victim); n != 1 {
		t.Errorf("delete audit log missing: %d", n)
	}
	var erased struct {
		AdminID *uint
		Meta    string
	}
	e.db.Raw(`SELECT admin_id, meta_data::text AS meta FROM admin_audit_logs WHERE target_type = 'user' AND target_id = ? AND action = 'USER_ERASED'`, victim).Scan(&erased)
	if erased.AdminID == nil || *erased.AdminID != adminID {
		t.Errorf("erase audit log admin = %v", erased.AdminID)
	}
	if n := e.count(`SELECT count(*) FROM admin_audit_logs WHERE target_type = 'user' AND target_id = ?
		AND (jsonb_exists(meta_data, 'target_username') OR jsonb_exists(meta_data, 'target_email') OR meta_data::text ILIKE ? OR meta_data::text ILIKE ?)`, victim, "%"+name+"%", "%victim@%"); n != 0 {
		t.Errorf("%d audit rows still carry the username or email", n)
	}
	// Free-text reasons on earlier rows about the target are redacted; the erase row keeps its own
	// reason (written knowing the person is being erased; the UI warns against personal data).
	if n := e.count(`SELECT count(*) FROM admin_audit_logs WHERE target_type = 'user' AND target_id = ? AND action <> 'USER_ERASED' AND reason IS NOT NULL`, victim); n != 0 {
		t.Errorf("%d earlier audit rows keep a reason", n)
	}
	if !strings.Contains(erased.Meta, `"storage_cleanup": "pending"`) {
		t.Errorf("erase audit meta lacks storage_cleanup=pending: %s", erased.Meta)
	}
	if err := admin_module.NewRepository(e.db).SetErasureCleanup(erasedAuditID(e, victim), admin_module.CleanupDone); err != nil {
		t.Fatal(err)
	}
	if n := e.count(`SELECT count(*) FROM admin_audit_logs WHERE target_id = ? AND action = 'USER_ERASED' AND meta_data->>'storage_cleanup' = 'done'`, victim); n != 1 {
		t.Errorf("cleanup status not updated: %d", n)
	}
	if n := e.count(`SELECT count(*) FROM admin_audit_logs WHERE admin_id = ? AND target_id = ?`, adminID, victim); n != 2 {
		t.Errorf("audit rows by admin for target: %d, want 2", n)
	}

	// The same email, username and Google subject are free again.
	nu, err := e.auth.Register(auth_module.RegisterRequest{Username: name, Email: email, Password: "second-pass1234", ConfirmPassword: "second-pass1234"})
	if err != nil || nu.ID == victim {
		t.Fatalf("re-register: user=%v err=%v", nu, err)
	}
	e.google.set(auth_module.GoogleIdentityClaims{Subject: sub, Email: "fresh@gmail.com", EmailVerified: true})
	if _, gu, err := e.googleLogin(); err != nil || gu.ID == victim {
		t.Fatalf("google signup: user=%v err=%v", gu, err)
	}
}

func TestAccountErasureGuards(t *testing.T) {
	e := newEnv(t)
	adminID := e.seedAdmin("admin@x.test", "admin-password1")
	admin := e.loginPair("admin@x.test", "admin-password1")
	live := seedUser(t, e.db, "live@x.test", true, "")
	removed := seedUser(t, e.db, "removed@x.test", true, "")
	e.softDelete(admin.AccessToken, removed)

	cases := []struct {
		name    string
		target  uint
		confirm string
		status  int
		code    string
	}{
		{"not soft-deleted", live, e.username(live), 409, "USER_NOT_REMOVED"},
		{"wrong confirmation", removed, e.username(removed) + "x", 400, "CONFIRMATION_MISMATCH"},
		{"case differs", removed, strings.ToUpper(e.username(removed)), 400, "CONFIRMATION_MISMATCH"},
		{"self", adminID, e.username(adminID), 403, "SELF_ERASURE"},
		{"missing", 987654, "nobody", 404, "USER_NOT_FOUND"},
	}
	for _, tc := range cases {
		code, body := e.erase(admin.AccessToken, tc.target, tc.confirm)
		if code != tc.status || body["code"] != tc.code {
			t.Errorf("%s: %d %v, want %d %s", tc.name, code, body, tc.status, tc.code)
		}
	}
	if n := e.count(`SELECT count(*) FROM users WHERE id IN (?, ?, ?)`, live, removed, adminID); n != 3 {
		t.Fatalf("a rejected erasure removed a user: %d left", n)
	}
	if n := e.count(`SELECT count(*) FROM admin_audit_logs WHERE action = 'USER_ERASED'`); n != 0 {
		t.Fatalf("rejected erasures wrote %d audit rows", n)
	}
}

// A plain user cannot reach the route.
func TestAccountErasureRequiresAdmin(t *testing.T) {
	e := newEnv(t)
	seedUser(t, e.db, "plain@x.test", true, "plain-password1")
	victim := seedUser(t, e.db, "v@x.test", true, "")
	plain := e.loginPair("plain@x.test", "plain-password1")
	if code, _ := e.erase(plain.AccessToken, victim, e.username(victim)); code != 403 {
		t.Fatalf("non-admin erasure: %d, want 403", code)
	}
	if code, _ := e.erase("", victim, e.username(victim)); code != 401 {
		t.Fatalf("anonymous erasure: %d, want 401", code)
	}
}

// Defensive guard: a row in an odd state (soft-deleted yet still an active admin) cannot be erased
// when no other active admin remains. Unreachable through the API, which clears is_active on delete.
func TestAccountErasureFinalAdminGuard(t *testing.T) {
	e := newEnv(t)
	actor := e.seedAdmin("actor@x.test", "admin-password1")
	odd := e.seedAdmin("odd@x.test", "admin-password1")
	e.exec(`UPDATE users SET deleted_at = now() WHERE id = ?`, odd)
	e.exec(`UPDATE users SET is_active = false WHERE id = ?`, actor) // no active admin left
	_, _, err := admin_module.NewRepository(e.db).EraseUser(actor, odd, e.username(odd), nil)
	if err != admin_module.ErrFinalAdminProtection {
		t.Fatalf("got %v, want ErrFinalAdminProtection", err)
	}
	if e.count(`SELECT count(*) FROM users WHERE id = ?`, odd) != 1 {
		t.Fatal("user erased")
	}
}

// Two erasures of one target: one wins, the other sees it gone. Erasure, delete and demotion
// all queue on the admin guard, so a mix of them neither deadlocks nor leaves zero active admins.
func TestAccountErasureConcurrency(t *testing.T) {
	e := newEnv(t)
	a := e.seedAdmin("a@x.test", "admin-password1")
	b := e.seedAdmin("b@x.test", "admin-password1")
	gone := e.seedAdmin("gone@x.test", "admin-password1")
	repo := admin_module.NewRepository(e.db)
	if err := repo.DeleteUser(a, gone, nil); err != nil {
		t.Fatal(err)
	}
	goneName := e.username(gone)

	errs := make([]error, 5)
	var wg sync.WaitGroup
	run := func(i int, f func() error) {
		wg.Add(1)
		go func() { defer wg.Done(); errs[i] = f() }()
	}
	run(0, func() error { _, _, err := repo.EraseUser(a, gone, goneName, nil); return err })
	run(1, func() error { _, _, err := repo.EraseUser(b, gone, goneName, nil); return err })
	run(2, func() error { return repo.DeleteUser(a, b, nil) })
	run(3, func() error { return repo.DeleteUser(b, a, nil) })
	run(4, func() error { return repo.UpdateUserStatus(a, b, false, nil) })
	wg.Wait()

	won := 0
	for i := 0; i < 2; i++ {
		switch errs[i] {
		case nil:
			won++
		case admin_module.ErrUserNotFound:
		default:
			t.Errorf("erase %d: %v", i, errs[i])
		}
	}
	if won != 1 {
		t.Errorf("%d erasures succeeded, want exactly 1", won)
	}
	if n := e.count(`SELECT count(*) FROM users u JOIN roles r ON r.id = u.role_id
		WHERE r.role_name = 'admin' AND u.is_active AND u.deleted_at IS NULL`); n < 1 {
		t.Fatalf("no active admin left: %d", n)
	}
}

func TestAdminListDeletedUsers(t *testing.T) {
	e := newEnv(t)
	e.seedAdmin("admin@x.test", "admin-password1")
	admin := e.loginPair("admin@x.test", "admin-password1")
	live := seedUser(t, e.db, "live@x.test", true, "")
	removed := seedUser(t, e.db, "removed@x.test", true, "")
	e.softDelete(admin.AccessToken, removed)

	ids := func(status string) map[float64]map[string]any {
		code, body := e.do("GET", "/api/admin/users?limit=100&status="+status, nil, admin.AccessToken)
		if code != 200 {
			t.Fatalf("list %s: %d %v", status, code, body)
		}
		out := map[float64]map[string]any{}
		for _, it := range body["items"].([]any) {
			row := it.(map[string]any)
			out[row["id"].(float64)] = row
		}
		return out
	}
	deleted := ids("deleted")
	if len(deleted) != 1 || deleted[float64(removed)] == nil || deleted[float64(removed)]["deleted_at"] == nil {
		t.Fatalf("status=deleted = %v", deleted)
	}
	for _, status := range []string{"all", "active", "inactive"} {
		got := ids(status)
		if got[float64(removed)] != nil {
			t.Errorf("status=%s lists a deleted user", status)
		}
		if status != "inactive" && got[float64(live)] == nil {
			t.Errorf("status=%s misses the live user", status)
		}
		if row := got[float64(live)]; row != nil && row["deleted_at"] != nil {
			t.Errorf("live user has deleted_at %v", row["deleted_at"])
		}
	}
	// Role, status and delete actions on a deleted user stay 404.
	for _, call := range [][3]string{
		{"PATCH", fmt.Sprintf("/api/admin/users/%d/status", removed), `{"is_active":true}`},
		{"PATCH", fmt.Sprintf("/api/admin/users/%d/role", removed), `{"role":"user"}`},
		{"DELETE", fmt.Sprintf("/api/admin/users/%d", removed), ""},
	} {
		if code, _ := e.do(call[0], call[1], call[2], admin.AccessToken); code != 404 {
			t.Errorf("%s %s on a deleted user: %d, want 404", call[0], call[1], code)
		}
	}
}

func erasedAuditID(e *env, target uint) uint {
	e.t.Helper()
	var id uint
	if err := e.db.Raw(`SELECT id FROM admin_audit_logs WHERE target_id = ? AND action = 'USER_ERASED'`, target).Scan(&id).Error; err != nil || id == 0 {
		e.t.Fatalf("erase audit row: id=%d err=%v", id, err)
	}
	return id
}

// The erase result lists the upload and Google-import avatar prefixes, with separators so
// user 1 never matches user 12.
func TestAccountErasureAvatarPrefixes(t *testing.T) {
	e := newEnv(t)
	actor := e.seedAdmin("actor@x.test", "admin-password1")
	victim := seedUser(t, e.db, "v@x.test", true, "")
	e.exec(`UPDATE users SET deleted_at = now(), is_active = false WHERE id = ?`, victim)
	res, _, err := admin_module.NewRepository(e.db).EraseUser(actor, victim, e.username(victim), nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{fmt.Sprintf("avatars/%d/", victim), fmt.Sprintf("avatars/google_%d_", victim)}
	if len(res.AvatarPrefixes) != 2 || res.AvatarPrefixes[0] != want[0] || res.AvatarPrefixes[1] != want[1] {
		t.Fatalf("prefixes = %v, want %v", res.AvatarPrefixes, want)
	}
}

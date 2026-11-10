//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/account_erasure"
	"github.com/arinsuda/movie-hub/internal/auth_module"
	"github.com/gofiber/fiber/v3"
)

const erasurePath = "/api/users/me/erasure-request"

// requestErasure posts like the SPA does: JSON plus the X-Requested-With header.
func (e *env) requestErasure(body any, token string) (int, map[string]any) {
	e.t.Helper()
	return e.doWith("POST", erasurePath, body, token, map[string]string{"X-Requested-With": "XMLHttpRequest"}, "application/json")
}

// doWith is do with an explicit Content-Type and extra headers (empty contentType omits it).
func (e *env) doWith(method, path string, body any, token string, headers map[string]string, contentType string) (int, map[string]any) {
	e.t.Helper()
	var rd io.Reader
	if s, ok := body.(string); ok {
		rd = strings.NewReader(s)
	} else {
		raw, _ := json.Marshal(body)
		rd = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, rd)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := e.app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
	if err != nil {
		e.t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	return resp.StatusCode, out
}

func (e *env) dueAt(uid uint) *time.Time {
	e.t.Helper()
	var row struct{ ErasureDueAt *time.Time }
	if err := e.db.Raw(`SELECT erasure_due_at FROM users WHERE id = ?`, uid).Scan(&row).Error; err != nil {
		e.t.Fatal(err)
	}
	return row.ErasureDueAt
}

func (e *env) purge() int {
	e.t.Helper()
	n, err := account_erasure.NewPurger(e.db, nil).RunOnce(context.Background())
	if err != nil {
		e.t.Fatal(err)
	}
	return n
}

func TestErasureRequestDeactivatesAndSchedules(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "me@x.test", true, "my-password1")
	p := e.loginPair("me@x.test", "my-password1")

	if code, body := e.requestErasure(map[string]string{"password": "wrong-password"}, p.AccessToken); code != 400 || body["code"] != "WRONG_PASSWORD" {
		t.Fatalf("wrong password: %d %v", code, body)
	}
	if e.dueAt(uid) != nil {
		t.Fatal("a refused request must not schedule anything")
	}

	before := time.Now()
	code, body := e.requestErasure(map[string]string{"password": "my-password1"}, p.AccessToken)
	if code != 202 || body["erasure_due_at"] == nil {
		t.Fatalf("request: %d %v", code, body)
	}
	due := e.dueAt(uid)
	if due == nil || due.Before(before.Add(config.ErasureGracePeriod-time.Minute)) || due.After(time.Now().Add(config.ErasureGracePeriod+time.Minute)) {
		t.Fatalf("erasure_due_at = %v", due)
	}
	var active bool
	e.db.Raw(`SELECT is_active FROM users WHERE id = ?`, uid).Scan(&active)
	if active {
		t.Fatal("account still active")
	}
	if code, _ := e.do("GET", "/api/ping", nil, p.AccessToken); code != 401 {
		t.Fatalf("access token survived: %d", code)
	}
	if _, _, err := e.auth.Refresh(p.RefreshToken, "ua", "ip"); err == nil {
		t.Fatal("refresh token survived")
	}
	reDue := regexp.MustCompile(`วันที่ <b[^>]*>([0-9-]+)</b>`)
	if got := e.mail.last(t, "me@x.test", reDue); got != due.UTC().Format("2006-01-02") {
		t.Fatalf("mail date = %s, want %s", got, due.UTC().Format("2006-01-02"))
	}
	if n := e.count(`SELECT count(*) FROM admin_audit_logs WHERE action = 'USER_ERASURE_REQUESTED' AND admin_id IS NULL AND target_id = ? AND meta_data ? 'erasure_due_at'`, uid); n != 1 {
		t.Fatalf("request audit rows = %d", n)
	}
	// Reserved during the grace period.
	if n := e.count(`SELECT count(*) FROM users WHERE email = 'me@x.test'`); n != 1 {
		t.Fatalf("email row gone during grace: %d", n)
	}
}

func TestErasureRequestPasswordlessNeedsUsername(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "g@x.test", true, "")
	p := e.accessToken(uid, 0)
	if code, body := e.requestErasure(map[string]string{"confirm_username": "nope"}, p); code != 400 || body["code"] == nil {
		t.Fatalf("mismatch: %d %v", code, body)
	}
	if code, _ := e.requestErasure(map[string]string{"confirm_username": e.username(uid)}, p); code != 202 {
		t.Fatalf("username confirm: %d", code)
	}
}

func TestErasureRequestLastAdminRefused(t *testing.T) {
	e := newEnv(t)
	uid := e.seedAdmin("admin@x.test", "admin-password1")
	p := e.loginPair("admin@x.test", "admin-password1")
	if code, body := e.requestErasure(map[string]string{"password": "admin-password1"}, p.AccessToken); code != 409 || body["code"] != "FINAL_ADMIN" {
		t.Fatalf("last admin: %d %v", code, body)
	}
	if e.dueAt(uid) != nil {
		t.Fatal("refused request scheduled erasure")
	}
}

func TestPurgeWorkerErasesDueAndOldDeleted(t *testing.T) {
	e := newEnv(t)
	adminID := e.seedAdmin("admin@x.test", "admin-password1")
	admin := e.loginPair("admin@x.test", "admin-password1")

	due := seedUser(t, e.db, "due@x.test", true, "")
	dueName := e.username(due)
	notDue := seedUser(t, e.db, "notdue@x.test", true, "")
	oldDeleted := seedUser(t, e.db, "old@x.test", true, "")
	freshDeleted := seedUser(t, e.db, "fresh@x.test", true, "")
	e.exec(`UPDATE users SET is_active = false, erasure_due_at = now() - interval '1 minute', avatar_url = ? WHERE id = ?`, fmt.Sprintf("avatars/%d/a.png", due), due)
	e.exec(`UPDATE users SET is_active = false, erasure_due_at = now() + interval '1 day' WHERE id = ?`, notDue)
	e.softDelete(admin.AccessToken, oldDeleted)
	e.softDelete(admin.AccessToken, freshDeleted)
	e.exec(`UPDATE users SET deleted_at = now() - interval '31 days' WHERE id = ?`, oldDeleted)
	e.seedReview(due, 7001)

	if n := e.purge(); n != 2 {
		t.Fatalf("purge erased=%d, want 2", n)
	}
	for _, id := range []uint{due, oldDeleted} {
		if c := e.count(`SELECT count(*) FROM users WHERE id = ?`, id); c != 0 {
			t.Fatalf("user %d not erased", id)
		}
	}
	if c := e.count(`SELECT count(*) FROM reviews WHERE user_id = ?`, due); c != 0 {
		t.Fatal("reviews must be hard-deleted")
	}
	for _, id := range []uint{notDue, freshDeleted, adminID} {
		if c := e.count(`SELECT count(*) FROM users WHERE id = ?`, id); c != 1 {
			t.Fatalf("user %d wrongly erased", id)
		}
	}
	if c := e.count(`SELECT count(*) FROM admin_audit_logs WHERE action = ? AND admin_id IS NULL AND target_id IN (?, ?)
		AND meta_data->>'storage_cleanup' = 'pending' AND meta_data::text NOT LIKE ?`,
		account_erasure.ActionUserErasedScheduled, due, oldDeleted, "%"+dueName+"%"); c != 2 {
		t.Fatalf("audit rows = %d, want 2", c)
	}
	// Re-signup with the freed email works.
	nu, err := e.auth.Register(auth_module.RegisterRequest{Username: dueName, Email: "due@x.test", Password: "second-pass1234", ConfirmPassword: "second-pass1234"})
	if err != nil || nu == nil {
		t.Fatalf("re-signup after purge: user=%v err=%v", nu, err)
	}
	if n := e.purge(); n != 0 {
		t.Fatalf("second run erased %d", n)
	}
}

// The route is a destructive POST behind a SameSite=None cookie: a cross-site form (text/plain,
// no custom header) must be refused and change nothing.
func TestErasureRequestRejectsCrossSiteStyleRequests(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "me@x.test", true, "my-password1")
	p := e.loginPair("me@x.test", "my-password1")
	body := `{"password":"my-password1"}`

	if code, res := e.doWith("POST", erasurePath, body, p.AccessToken, nil, "text/plain"); code != 415 || res["code"] != "UNSUPPORTED_MEDIA_TYPE" {
		t.Fatalf("text/plain: %d %v", code, res)
	}
	if code, res := e.doWith("POST", erasurePath, body, p.AccessToken, nil, "application/json"); code != 403 || res["code"] != "XHR_HEADER_REQUIRED" {
		t.Fatalf("no custom header: %d %v", code, res)
	}
	if code, _ := e.doWith("POST", erasurePath, body, p.AccessToken, map[string]string{"X-Requested-With": "XMLHttpRequest"}, ""); code != 415 {
		t.Fatalf("no content type: %d", code)
	}
	if e.dueAt(uid) != nil {
		t.Fatal("a rejected request scheduled erasure")
	}
}

// Several workers may run at once: every account is erased exactly once, none errors.
func TestPurgeWorkersRunConcurrently(t *testing.T) {
	e := newEnv(t)
	for i := 0; i < 6; i++ {
		uid := seedUser(t, e.db, fmt.Sprintf("due%d@x.test", i), true, "")
		e.exec(`UPDATE users SET is_active = false, erasure_due_at = now() - interval '1 minute' WHERE id = ?`, uid)
	}
	var wg sync.WaitGroup
	var total atomic.Int64
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n, err := account_erasure.NewPurger(e.db, nil).RunOnce(context.Background())
			if err != nil {
				t.Error(err)
			}
			total.Add(int64(n))
		}()
	}
	wg.Wait()
	if total.Load() != 6 {
		t.Fatalf("erased %d in total, want 6", total.Load())
	}
	if c := e.count(`SELECT count(*) FROM admin_audit_logs WHERE action = ?`, account_erasure.ActionUserErasedScheduled); c != 6 {
		t.Fatalf("audit rows = %d, want 6", c)
	}
}

func TestAdminReactivateClearsErasureDue(t *testing.T) {
	e := newEnv(t)
	e.seedAdmin("admin@x.test", "admin-password1")
	admin := e.loginPair("admin@x.test", "admin-password1")
	uid := seedUser(t, e.db, "me@x.test", true, "my-password1")
	e.exec(`UPDATE users SET is_active = false, erasure_due_at = now() + interval '1 day' WHERE id = ?`, uid)

	if code, body := e.do("PATCH", fmt.Sprintf("/api/admin/users/%d/status", uid), map[string]any{"is_active": true}, admin.AccessToken); code != 200 && code != 204 {
		t.Fatalf("reactivate: %d %v", code, body)
	}
	if e.dueAt(uid) != nil {
		t.Fatal("erasure_due_at not cleared")
	}
}

type fakeStore struct {
	fail    atomic.Bool
	deleted atomic.Int64
}

func (f *fakeStore) DeleteObject(context.Context, string) error { return f.call() }
func (f *fakeStore) DeletePrefix(context.Context, string) error { return f.call() }
func (f *fakeStore) call() error {
	if f.fail.Load() {
		return fmt.Errorf("storage down")
	}
	f.deleted.Add(1)
	return nil
}

// A failed avatar cleanup stays pending in the audit row and a later tick finishes it.
func TestPurgeRetriesPendingStorageCleanup(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "due@x.test", true, "")
	e.exec(`UPDATE users SET is_active = false, erasure_due_at = now() - interval '1 minute' WHERE id = ?`, uid)
	store := &fakeStore{}
	store.fail.Store(true)
	p := account_erasure.NewPurger(e.db, store)
	if n, err := p.RunOnce(context.Background()); err != nil || n != 1 {
		t.Fatalf("run: n=%d err=%v", n, err)
	}
	pending := `SELECT count(*) FROM admin_audit_logs WHERE action = ? AND meta_data->>'storage_cleanup' = 'pending'`
	if c := e.count(pending, account_erasure.ActionUserErasedScheduled); c != 1 {
		t.Fatalf("pending rows = %d, want 1", c)
	}
	store.fail.Store(false)
	if n := p.RetryCleanups(context.Background()); n != 1 {
		t.Fatalf("retried %d, want 1", n)
	}
	if c := e.count(pending, account_erasure.ActionUserErasedScheduled); c != 0 {
		t.Fatal("cleanup still pending after retry")
	}
}

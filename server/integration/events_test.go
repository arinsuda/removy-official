//go:build integration

package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/internal/admin_module"
	"github.com/arinsuda/movie-hub/internal/event_module"
)

func TestEventsPublicAndAdminFlow(t *testing.T) {
	e := newEnv(t)
	now := time.Now().UTC()

	e.seedAdmin("admin@x.test", "pw-admin")
	adminTok := e.loginPair("admin@x.test", "pw-admin").AccessToken

	seedUser(t, e.db, "user@x.test", true, "pw-user")
	userTok := e.loginPair("user@x.test", "pw-user").AccessToken

	// 1. Regular user forbidden on admin endpoints
	code, out := e.do("POST", "/api/admin/events", map[string]any{
		"title_en":  "Unauthorized Event",
		"title_th":  "กิจกรรมไม่ได้รับอนุญาต",
		"starts_at": now.Format(time.RFC3339),
		"ends_at":   now.Add(24 * time.Hour).Format(time.RFC3339),
	}, userTok)
	if code != 403 {
		t.Fatalf("expected 403 for non-admin, got %d: %v", code, out)
	}

	// 2. Admin creates a published active event with popup
	createBody := map[string]any{
		"title_en":       "Launch Celebration",
		"title_th":       "เฉลิมฉลองการเปิดตัว",
		"body_en":        "Welcome to POPCORN events!",
		"body_th":        "ยินดีต้อนรับสู่กิจกรรม POPCORN!",
		"link_url":       "/movies",
		"starts_at":      now.Add(-time.Hour).Format(time.RFC3339),
		"ends_at":        now.Add(48 * time.Hour).Format(time.RFC3339),
		"show_popup":     true,
		"popup_priority": 10,
		"status":         "published",
	}
	code, createRes := e.do("POST", "/api/admin/events", createBody, adminTok)
	if code != 201 {
		t.Fatalf("admin create event failed (%d): %v", code, createRes)
	}
	evMap, ok := createRes["event"].(map[string]any)
	if !ok {
		t.Fatalf("missing event in create response: %v", createRes)
	}
	eventID := uint(evMap["id"].(float64))

	// Also create a draft event
	draftBody := map[string]any{
		"title_en":  "Draft Event",
		"title_th":  "กิจกรรมร่าง",
		"starts_at": now.Add(time.Hour).Format(time.RFC3339),
		"ends_at":   now.Add(24 * time.Hour).Format(time.RFC3339),
		"status":    "draft",
	}
	code, draftRes := e.do("POST", "/api/admin/events", draftBody, adminTok)
	if code != 201 {
		t.Fatalf("admin create draft event failed (%d): %v", code, draftRes)
	}
	draftID := uint(draftRes["event"].(map[string]any)["id"].(float64))

	// 3. Guest list public events (state=active)
	code, listRes := e.do("GET", "/api/events?state=active", nil, "")
	if code != 200 {
		t.Fatalf("guest list events failed (%d): %v", code, listRes)
	}
	items := listRes["events"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 active published event, got %d", len(items))
	}
	item := items[0].(map[string]any)
	if uint(item["id"].(float64)) != eventID {
		t.Errorf("expected event id %d, got %v", eventID, item["id"])
	}

	// 4. Guest get public event by ID
	code, getRes := e.do("GET", fmt.Sprintf("/api/events/%d", eventID), nil, "")
	if code != 200 {
		t.Fatalf("guest get event failed (%d): %v", code, getRes)
	}

	// Draft event must return 404 for public
	code, _ = e.do("GET", fmt.Sprintf("/api/events/%d", draftID), nil, "")
	if code != 404 {
		t.Fatalf("expected 404 for draft event, got %d", code)
	}

	// 5. Popup endpoint: guest gets 401, signed-in user gets 200 with popup event
	code, _ = e.do("GET", "/api/events/popup", nil, "")
	if code != 401 {
		t.Fatalf("expected 401 for guest popup, got %d", code)
	}

	code, popupRes := e.do("GET", "/api/events/popup", nil, userTok)
	if code != 200 {
		t.Fatalf("user popup failed (%d): %v", code, popupRes)
	}
	popups := popupRes["events"].([]any)
	if len(popups) != 1 {
		t.Fatalf("expected 1 popup event, got %d", len(popups))
	}

	// 6. Admin updates event
	updateBody := createBody
	updateBody["title_en"] = "Launch Celebration Updated"
	code, updateRes := e.do("PUT", fmt.Sprintf("/api/admin/events/%d", eventID), updateBody, adminTok)
	if code != 200 {
		t.Fatalf("admin update event failed (%d): %v", code, updateRes)
	}
	updatedEv := updateRes["event"].(map[string]any)
	updatedTitle := updatedEv["title"].(map[string]any)["en"].(string)
	if updatedTitle != "Launch Celebration Updated" {
		t.Errorf("expected updated title, got %s", updatedTitle)
	}

	// 7. Admin deletes event
	code, _ = e.do("DELETE", fmt.Sprintf("/api/admin/events/%d", eventID), nil, adminTok)
	if code != 204 {
		t.Fatalf("admin delete event failed (%d)", code)
	}

	// Deleted event is now 404
	code, _ = e.do("GET", fmt.Sprintf("/api/events/%d", eventID), nil, "")
	if code != 404 {
		t.Fatalf("expected 404 for deleted event, got %d", code)
	}

	// 8. Verify audit logs recorded actions
	var auditCount int64
	e.db.Model(&admin_module.AdminAuditLog{}).
		Where("target_type = ? AND action IN (?)", "event", []string{
			string(admin_module.ActionEventCreated),
			string(admin_module.ActionEventUpdated),
			string(admin_module.ActionEventDeleted),
		}).Count(&auditCount)
	if auditCount < 3 {
		t.Errorf("expected at least 3 event audit logs, got %d", auditCount)
	}
}

func TestEventInputValidationFailures(t *testing.T) {
	e := newEnv(t)
	now := time.Now().UTC()

	e.seedAdmin("admin2@x.test", "pw-admin")
	adminTok := e.loginPair("admin2@x.test", "pw-admin").AccessToken

	cases := []struct {
		name string
		body map[string]any
	}{
		{
			name: "missing title_en",
			body: map[string]any{
				"title_en":  "",
				"title_th":  "ไทย",
				"starts_at": now.Format(time.RFC3339),
				"ends_at":   now.Add(time.Hour).Format(time.RFC3339),
			},
		},
		{
			name: "missing title_th",
			body: map[string]any{
				"title_en":  "English",
				"title_th":  "",
				"starts_at": now.Format(time.RFC3339),
				"ends_at":   now.Add(time.Hour).Format(time.RFC3339),
			},
		},
		{
			name: "ends_at before starts_at",
			body: map[string]any{
				"title_en":  "English",
				"title_th":  "ไทย",
				"starts_at": now.Add(2 * time.Hour).Format(time.RFC3339),
				"ends_at":   now.Add(time.Hour).Format(time.RFC3339),
			},
		},
		{
			name: "invalid link_url",
			body: map[string]any{
				"title_en":  "English",
				"title_th":  "ไทย",
				"link_url":  "//evil.com",
				"starts_at": now.Format(time.RFC3339),
				"ends_at":   now.Add(time.Hour).Format(time.RFC3339),
			},
		},
		{
			name: "invalid status",
			body: map[string]any{
				"title_en":  "English",
				"title_th":  "ไทย",
				"status":    "bogus",
				"starts_at": now.Format(time.RFC3339),
				"ends_at":   now.Add(time.Hour).Format(time.RFC3339),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, out := e.do("POST", "/api/admin/events", tc.body, adminTok)
			if code != 400 {
				t.Fatalf("expected 400, got %d: %v", code, out)
			}
			if codeVal, _ := out["code"].(string); codeVal != "INVALID_INPUT" {
				t.Errorf("expected INVALID_INPUT code, got %s", codeVal)
			}
		})
	}
}

// eventIDs returns the ids of the events under key in a list response.
func eventIDs(t *testing.T, out map[string]any, key string) []uint {
	t.Helper()
	items, ok := out[key].([]any)
	if !ok {
		t.Fatalf("missing %s in response: %v", key, out)
	}
	ids := make([]uint, 0, len(items))
	for _, it := range items {
		ids = append(ids, uint(it.(map[string]any)["id"].(float64)))
	}
	return ids
}

func assertIDs(t *testing.T, got, want []uint) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got ids %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got ids %v, want %v", got, want)
		}
	}
}

func createEvent(t *testing.T, e *env, adminTok string, body map[string]any) uint {
	t.Helper()
	code, out := e.do("POST", "/api/admin/events", body, adminTok)
	if code != 201 {
		t.Fatalf("create event failed (%d): %v", code, out)
	}
	return uint(out["event"].(map[string]any)["id"].(float64))
}

// eventWindow builds an admin create body whose window is [now+from, now+to).
func eventWindow(now time.Time, status string, from, to time.Duration) map[string]any {
	return map[string]any{
		"title_en":  status + " event",
		"title_th":  "กิจกรรม",
		"starts_at": now.Add(from).Format(time.RFC3339),
		"ends_at":   now.Add(to).Format(time.RFC3339),
		"status":    status,
	}
}

func TestEventsListHidesDraftAndCancelled(t *testing.T) {
	e := newEnv(t)
	now := time.Now().UTC()

	e.seedAdmin("admin3@x.test", "pw-admin")
	adminTok := e.loginPair("admin3@x.test", "pw-admin").AccessToken

	upcoming := createEvent(t, e, adminTok, eventWindow(now, "published", time.Hour, 2*time.Hour))
	createEvent(t, e, adminTok, eventWindow(now, "draft", 3*time.Hour, 4*time.Hour))
	createEvent(t, e, adminTok, eventWindow(now, "cancelled", 5*time.Hour, 6*time.Hour))

	active := createEvent(t, e, adminTok, eventWindow(now, "published", -time.Hour, 24*time.Hour))
	createEvent(t, e, adminTok, eventWindow(now, "draft", -time.Hour, 24*time.Hour))
	createEvent(t, e, adminTok, eventWindow(now, "cancelled", -time.Hour, time.Hour))

	past := createEvent(t, e, adminTok, eventWindow(now, "published", -3*time.Hour, -time.Hour))
	createEvent(t, e, adminTok, eventWindow(now, "draft", -3*time.Hour, -2*time.Hour))
	createEvent(t, e, adminTok, eventWindow(now, "cancelled", -3*time.Hour, -2*time.Hour))

	// Window edge: an event that ended one second ago is past, never active.
	justEnded := createEvent(t, e, adminTok, eventWindow(now, "published", -2*time.Hour, -time.Second))

	cases := []struct {
		state string
		want  []uint
	}{
		{"upcoming", []uint{upcoming}},
		{"active", []uint{active}},
		// Ordered by ends_at DESC, so the event that ended most recently comes first.
		{"past", []uint{justEnded, past}},
	}
	for _, tc := range cases {
		code, out := e.do("GET", "/api/events?state="+tc.state, nil, "")
		if code != 200 {
			t.Fatalf("list state=%s failed (%d): %v", tc.state, code, out)
		}
		t.Run(tc.state, func(t *testing.T) {
			assertIDs(t, eventIDs(t, out, "events"), tc.want)
		})
	}
}

func TestEventsPopupFiltersHiddenAndOutOfWindow(t *testing.T) {
	e := newEnv(t)
	now := time.Now().UTC()

	e.seedAdmin("admin4@x.test", "pw-admin")
	adminTok := e.loginPair("admin4@x.test", "pw-admin").AccessToken
	seedUser(t, e.db, "popup-user@x.test", true, "pw-user")
	userTok := e.loginPair("popup-user@x.test", "pw-user").AccessToken

	withPopup := func(status string, from, to time.Duration, showPopup bool) map[string]any {
		body := eventWindow(now, status, from, to)
		body["show_popup"] = showPopup
		return body
	}
	popupOn := createEvent(t, e, adminTok, withPopup("published", -time.Hour, 24*time.Hour, true))
	createEvent(t, e, adminTok, withPopup("published", -time.Hour, 24*time.Hour, false))
	createEvent(t, e, adminTok, withPopup("published", time.Hour, 2*time.Hour, true))
	createEvent(t, e, adminTok, withPopup("published", -3*time.Hour, -time.Hour, true))
	createEvent(t, e, adminTok, withPopup("draft", -time.Hour, 24*time.Hour, true))
	createEvent(t, e, adminTok, withPopup("cancelled", -time.Hour, 24*time.Hour, true))

	code, out := e.do("GET", "/api/events/popup", nil, userTok)
	if code != 200 {
		t.Fatalf("user popup failed (%d): %v", code, out)
	}
	assertIDs(t, eventIDs(t, out, "events"), []uint{popupOn})
}

// Suppress unused import warnings if any
var _ = event_module.TypeAnnouncement

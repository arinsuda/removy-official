//go:build integration

package integration

import (
	"fmt"
	"strings"
	"testing"
)

func TestBMOLLifecycle(t *testing.T) {
	e := newEnv(t)
	owner := seedUser(t, e.db, "owner@x.test", true, "password-one1")
	seedUser(t, e.db, "other@x.test", true, "password-two2")
	own := e.loginPair("owner@x.test", "password-one1").AccessToken
	oth := e.loginPair("other@x.test", "password-two2").AccessToken
	add := func(id, rank int, tok string) (int, map[string]any) {
		return e.do("POST", "/api/bmol", map[string]any{"media_id": id, "media_type": "movie", "rank": rank}, tok)
	}

	code, b := add(603, 1, own)
	if code != 201 {
		t.Fatalf("add: %d %v", code, b)
	}
	item := b["item"].(map[string]any)
	itemPath := fmt.Sprintf("/api/bmol/%v", item["id"])
	if code, _ := add(603, 2, own); code != 409 {
		t.Fatalf("duplicate: want 409, got %d", code)
	}
	if code, _ := e.do("DELETE", itemPath, nil, own); code != 204 {
		t.Fatalf("delete: %d", code)
	}
	// Re-add after soft delete restores the row instead of a 500.
	code, b = add(603, 3, own)
	if code != 201 {
		t.Fatalf("re-add: %d %v", code, b)
	}
	if n := e.count(`SELECT count(*) FROM bmol_items WHERE user_id = ? AND deleted_at IS NULL`, owner); n != 1 {
		t.Fatalf("want 1 live row, got %d", n)
	}
	itemPath = fmt.Sprintf("/api/bmol/%v", b["item"].(map[string]any)["id"])

	for name, body := range map[string]map[string]any{
		"rank 0":    {"media_id": 1, "media_type": "movie", "rank": 0},
		"media 0":   {"media_id": 0, "media_type": "movie", "rank": 1},
		"bad type":  {"media_id": 1, "media_type": "book", "rank": 1},
		"rank 1000": {"media_id": 1, "media_type": "movie", "rank": 1000},
	} {
		if code, _ := e.do("POST", "/api/bmol", body, own); code != 400 {
			t.Fatalf("%s: want 400, got %d", name, code)
		}
	}
	if code, _ := e.do("PUT", itemPath, map[string]any{}, own); code != 400 {
		t.Fatalf("empty update: want 400, got %d", code)
	}
	if code, _ := e.do("PUT", itemPath, map[string]any{"note": strings.Repeat("a", 281)}, own); code != 400 {
		t.Fatalf("long note: want 400, got %d", code)
	}
	if code, _ := e.do("PUT", itemPath, map[string]any{"rank": 2}, oth); code != 403 {
		t.Fatalf("non-owner PUT: want 403, got %d", code)
	}
	if code, _ := e.do("DELETE", itemPath, nil, oth); code != 403 {
		t.Fatalf("non-owner DELETE: want 403, got %d", code)
	}
	if code, b := e.do("PUT", itemPath, map[string]any{"note": "  masterpiece "}, own); code != 200 || b["item"].(map[string]any)["note"] != "masterpiece" {
		t.Fatalf("note update: %d %v", code, b)
	}
}

// The note is exposed only through the gated GET.
func TestBMOLNoteGated(t *testing.T) {
	e := newEnv(t)
	owner := seedUser(t, e.db, "owner@x.test", true, "password-one1")
	seedUser(t, e.db, "stranger@x.test", true, "password-two2")
	e.db.Exec(`UPDATE users SET is_private = true WHERE id = ?`, owner)
	e.db.Exec(`INSERT INTO bmol_items (created_at, updated_at, user_id, media_id, media_type, rank, note) VALUES (now(), now(), ?, 603, 'movie', 1, 'secret note')`, owner)
	path := fmt.Sprintf("/api/bmol/user/%d", owner)

	own := e.loginPair("owner@x.test", "password-one1").AccessToken
	str := e.loginPair("stranger@x.test", "password-two2").AccessToken
	code, b := e.do("GET", path, nil, str)
	if code != 403 || strings.Contains(fmt.Sprint(b), "secret note") {
		t.Fatalf("stranger: %d %v", code, b)
	}
	code, b = e.do("GET", path, nil, own)
	items, _ := b["items"].([]any)
	if code != 200 || len(items) != 1 || items[0].(map[string]any)["note"] != "secret note" {
		t.Fatalf("owner: %d %v", code, b)
	}
}

// Two concurrent re-adds of a soft-deleted title: exactly one wins, the other gets 409.
func TestBMOLRestoreRace(t *testing.T) {
	e := newEnv(t)
	owner := seedUser(t, e.db, "owner@x.test", true, "password-one1")
	own := e.loginPair("owner@x.test", "password-one1").AccessToken
	e.db.Exec(`INSERT INTO bmol_items (created_at, updated_at, deleted_at, user_id, media_id, media_type, rank) VALUES ('2020-01-01', now(), now(), ?, 603, 'movie', 1)`, owner)

	codes := make(chan int, 2)
	for i := 0; i < 2; i++ {
		go func() {
			code, _ := e.do("POST", "/api/bmol", map[string]any{"media_id": 603, "media_type": "movie", "rank": 2}, own)
			codes <- code
		}()
	}
	a, b := <-codes, <-codes
	if !((a == 201 && b == 409) || (a == 409 && b == 201)) {
		t.Fatalf("want one 201 and one 409, got %d and %d", a, b)
	}
	if n := e.count(`SELECT count(*) FROM bmol_items WHERE user_id = ? AND deleted_at IS NULL AND created_at > '2021-01-01'`, owner); n != 1 {
		t.Fatalf("want 1 live restored row, got %d", n)
	}
}

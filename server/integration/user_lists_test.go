//go:build integration

package integration

import (
	"fmt"
	"sync"
	"testing"

	"github.com/arinsuda/movie-hub/internal/list_module"
)

// listFixture holds one owner and the three kinds of viewers.
type listFixture struct {
	e                                  *env
	owner, follower, stranger          uint
	ownerTok, followerTok, strangerTok string
}

func newListFixture(t *testing.T) *listFixture {
	t.Helper()
	e := newEnv(t)
	f := &listFixture{e: e}
	f.owner = seedUser(t, e.db, "owner@x.test", true, "owner-password1")
	f.follower = seedUser(t, e.db, "follower@x.test", true, "follower-password1")
	f.stranger = seedUser(t, e.db, "stranger@x.test", true, "stranger-password1")
	if err := e.db.Exec(`INSERT INTO user_follows (follower_id, followee_id, status, created_at, updated_at) VALUES (?, ?, 'accepted', now(), now())`, f.follower, f.owner).Error; err != nil {
		t.Fatal(err)
	}
	f.ownerTok = e.loginPair("owner@x.test", "owner-password1").AccessToken
	f.followerTok = e.loginPair("follower@x.test", "follower-password1").AccessToken
	f.strangerTok = e.loginPair("stranger@x.test", "stranger-password1").AccessToken
	return f
}

func (f *listFixture) createList(visibility string) uint {
	f.e.t.Helper()
	code, out := f.e.do("POST", "/api/lists", map[string]any{"name": "L-" + visibility, "visibility": visibility}, f.ownerTok)
	if code != 201 {
		f.e.t.Fatalf("create %s: %d %v", visibility, code, out)
	}
	return uint(out["list"].(map[string]any)["id"].(float64))
}

func (f *listFixture) addItem(listID uint, mediaID int) (int, map[string]any) {
	return f.e.do("POST", fmt.Sprintf("/api/lists/%d/items", listID), map[string]any{"media_id": mediaID, "media_type": "movie"}, f.ownerTok)
}

func TestListDefaultsAndValidation(t *testing.T) {
	f := newListFixture(t)
	code, out := f.e.do("POST", "/api/lists", map[string]any{"name": "Plain"}, f.ownerTok)
	if l, _ := out["list"].(map[string]any); code != 201 || l["visibility"] != "public" {
		t.Fatalf("default visibility: %d %v", code, out)
	}
	// Duplicate names are allowed.
	if code, _ := f.e.do("POST", "/api/lists", map[string]any{"name": "Plain"}, f.ownerTok); code != 201 {
		t.Fatalf("duplicate name: %d", code)
	}
	for _, body := range []map[string]any{{"name": "  "}, {"name": "x", "visibility": "friends"}} {
		if code, _ := f.e.do("POST", "/api/lists", body, f.ownerTok); code != 400 {
			t.Fatalf("invalid body %v: %d", body, code)
		}
	}
	if n := f.e.count(`SELECT count(*) FROM pg_constraint WHERE conname IN ('chk_user_lists_visibility','chk_user_list_items_media_type')`); n != 2 {
		t.Fatalf("check constraints missing: %d", n)
	}
	if code, _ := f.e.do("GET", "/api/lists", nil, ""); code != 401 {
		t.Fatalf("guest: %d", code)
	}
}

// owner / follower / stranger x three visibilities x public and private accounts.
func TestListPrivacyMatrix(t *testing.T) {
	for _, privateAccount := range []bool{false, true} {
		f := newListFixture(t)
		ids := map[string]uint{}
		for _, v := range []string{"public", "followers", "private"} {
			ids[v] = f.createList(v)
		}
		if privateAccount {
			f.e.db.Exec(`UPDATE users SET is_private = true WHERE id = ?`, f.owner)
		}
		// want[viewer][visibility]
		type want map[string]bool
		viewers := []struct {
			name string
			tok  string
			want want
		}{
			{"owner", f.ownerTok, want{"public": true, "followers": true, "private": true}},
			{"follower", f.followerTok, want{"public": true, "followers": true, "private": false}},
			{"stranger", f.strangerTok, want{"public": !privateAccount, "followers": false, "private": false}},
		}
		for _, v := range viewers {
			visibleCount := 0
			for vis, id := range ids {
				code, out := f.e.do("GET", fmt.Sprintf("/api/lists/%d", id), nil, v.tok)
				wantCode := 404
				if v.want[vis] {
					wantCode = 200
					visibleCount++
				}
				if code != wantCode {
					t.Errorf("private=%v %s -> %s list: %d %v, want %d", privateAccount, v.name, vis, code, out, wantCode)
				}
			}
			code, out := f.e.do("GET", fmt.Sprintf("/api/users/%d/lists", f.owner), nil, v.tok)
			if v.name == "stranger" && privateAccount {
				if code != 403 {
					t.Errorf("locked profile list index: %d, want 403", code)
				}
				continue
			}
			lists, _ := out["lists"].([]any)
			if code != 200 || len(lists) != visibleCount {
				t.Errorf("private=%v %s index: %d, %d lists, want %d", privateAccount, v.name, code, len(lists), visibleCount)
			}
		}
	}
}

func TestListOfInactiveOwnerIs404(t *testing.T) {
	f := newListFixture(t)
	id := f.createList("public")
	f.e.db.Exec(`UPDATE users SET is_active = false WHERE id = ?`, f.owner)
	if code, _ := f.e.do("GET", fmt.Sprintf("/api/lists/%d", id), nil, f.strangerTok); code != 404 {
		t.Fatalf("list of inactive owner: %d, want 404", code)
	}
	if code, _ := f.e.do("GET", fmt.Sprintf("/api/users/%d/lists", f.owner), nil, f.strangerTok); code != 404 {
		t.Fatalf("list index of inactive owner: %d, want 404", code)
	}
}

func TestListMutationsAreOwnerOnly(t *testing.T) {
	f := newListFixture(t)
	id := f.createList("public")
	base := fmt.Sprintf("/api/lists/%d", id)
	if code, _ := f.addItem(id, 603); code != 201 {
		t.Fatal("seed item")
	}
	for _, c := range []struct {
		method, path string
		body         any
	}{
		{"PATCH", base, map[string]any{"name": "hijack"}},
		{"DELETE", base, nil},
		{"POST", base + "/items", map[string]any{"media_id": 1, "media_type": "movie"}},
		{"PUT", base + "/items/order", map[string]any{"item_ids": []uint{}}},
	} {
		if code, _ := f.e.do(c.method, c.path, c.body, f.strangerTok); code != 404 {
			t.Errorf("stranger %s %s: %d, want 404", c.method, c.path, code)
		}
	}
	if n := f.e.count(`SELECT count(*) FROM user_lists WHERE id = ? AND name = 'L-public' AND deleted_at IS NULL`, id); n != 1 {
		t.Fatal("a stranger changed the list")
	}
}

func TestListItemsDuplicateCoverAndDelete(t *testing.T) {
	f := newListFixture(t)
	id := f.createList("public")
	base := fmt.Sprintf("/api/lists/%d", id)

	code, out := f.addItem(id, 603)
	if code != 201 {
		t.Fatalf("add: %d %v", code, out)
	}
	itemID := uint(out["item"].(map[string]any)["id"].(float64))
	if code, out := f.addItem(id, 603); code != 409 || out["code"] != "ITEM_EXISTS" {
		t.Fatalf("duplicate: %d %v", code, out)
	}
	if code, out := f.e.do("POST", base+"/items", map[string]any{"media_id": 5, "media_type": "book"}, f.ownerTok); code != 400 {
		t.Fatalf("bad media type: %d %v", code, out)
	}

	if code, out := f.e.do("PATCH", base, map[string]any{"cover_item_id": itemID}, f.ownerTok); code != 200 {
		t.Fatalf("set cover: %d %v", code, out)
	}
	if code, _ := f.e.do("PATCH", base, map[string]any{"cover_item_id": 999999}, f.ownerTok); code != 404 {
		t.Fatalf("foreign cover item: %d", code)
	}
	if code, _ := f.e.do("DELETE", fmt.Sprintf("%s/items/%d", base, itemID), nil, f.ownerTok); code != 204 {
		t.Fatalf("remove item: %d", code)
	}
	if n := f.e.count(`SELECT count(*) FROM user_lists WHERE id = ? AND cover_item_id IS NULL`, id); n != 1 {
		t.Fatal("cover must reset when its item is removed")
	}
	if code, out := f.e.do("DELETE", fmt.Sprintf("%s/items/%d", base, itemID), nil, f.ownerTok); code != 404 || out["code"] != "ITEM_NOT_FOUND" {
		t.Fatalf("remove twice: %d %v", code, out)
	}
	// contains_media and item_id reflect membership for the item picker.
	_, addedOut := f.addItem(id, 27205)
	addedItemID := addedOut["item"].(map[string]any)["id"]
	_, out = f.e.do("GET", "/api/lists?media_type=movie&media_id=27205", nil, f.ownerTok)
	lists, _ := out["lists"].([]any)
	if len(lists) != 1 || lists[0].(map[string]any)["contains_media"] != true || lists[0].(map[string]any)["item_id"] != addedItemID {
		t.Fatalf("contains_media/item_id: %v", out)
	}

	if code, _ := f.e.do("DELETE", base, nil, f.ownerTok); code != 204 {
		t.Fatalf("delete list: %d", code)
	}
	if code, _ := f.e.do("GET", base, nil, f.ownerTok); code != 404 {
		t.Fatalf("deleted list readable: %d", code)
	}
	if n := f.e.count(`SELECT count(*) FROM user_lists WHERE id = ? AND deleted_at IS NOT NULL`, id); n != 1 {
		t.Fatal("list delete must be a soft delete")
	}
}

func TestListCaps(t *testing.T) {
	f := newListFixture(t)

	id := f.createList("public")
	f.e.db.Exec(`INSERT INTO user_list_items (list_id, media_type, media_id, position, created_at)
		SELECT ?, 'movie', g, g, now() FROM generate_series(1, ?) g`, id, list_module.MaxItemsPerList)
	if code, out := f.addItem(id, 999001); code != 422 || out["code"] != "LIST_FULL" {
		t.Fatalf("full list: %d %v", code, out)
	}

	f.e.db.Exec(`INSERT INTO user_lists (created_at, updated_at, user_id, name, visibility)
		SELECT now(), now(), ?, 'bulk', 'public' FROM generate_series(1, ?)`, f.owner, list_module.MaxListsPerUser-2)
	// One slot is left: of several concurrent creates exactly one wins.
	var wg sync.WaitGroup
	codes := make([]int, 6)
	for i := range codes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			codes[i], _ = f.e.do("POST", "/api/lists", map[string]any{"name": "race"}, f.ownerTok)
		}()
	}
	wg.Wait()
	created, limited := 0, 0
	for _, c := range codes {
		switch c {
		case 201:
			created++
		case 422:
			limited++
		}
	}
	if created != 1 || limited != len(codes)-1 {
		t.Fatalf("concurrent creates at the cap: %v", codes)
	}
	if n := f.e.count(`SELECT count(*) FROM user_lists WHERE user_id = ? AND deleted_at IS NULL`, f.owner); n != list_module.MaxListsPerUser {
		t.Fatalf("list count %d, want %d", n, list_module.MaxListsPerUser)
	}
}

func TestListConcurrentAddsGetUniquePositions(t *testing.T) {
	f := newListFixture(t)
	id := f.createList("public")
	const adds = 15
	var wg sync.WaitGroup
	codes := make([]int, adds)
	for i := 0; i < adds; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			codes[i], _ = f.addItem(id, 1000+i)
		}()
	}
	wg.Wait()
	for i, c := range codes {
		if c != 201 {
			t.Fatalf("add #%d: %d", i, c)
		}
	}
	if n := f.e.count(`SELECT count(DISTINCT position) FROM user_list_items WHERE list_id = ?`, id); n != adds {
		t.Fatalf("distinct positions %d, want %d", n, adds)
	}
}

func TestListReorder(t *testing.T) {
	f := newListFixture(t)
	id := f.createList("public")
	var ids []uint
	for i := 0; i < 3; i++ {
		_, out := f.addItem(id, 2000+i)
		ids = append(ids, uint(out["item"].(map[string]any)["id"].(float64)))
	}
	path := fmt.Sprintf("/api/lists/%d/items/order", id)

	for name, body := range map[string][]uint{
		"missing item": {ids[0], ids[1]},
		"duplicate":    {ids[0], ids[0], ids[1]},
		"foreign id":   {ids[0], ids[1], 999999},
		"extra":        {ids[0], ids[1], ids[2], 999999},
	} {
		if code, out := f.e.do("PUT", path, map[string]any{"item_ids": body}, f.ownerTok); code != 400 || out["code"] != "ORDER_MISMATCH" {
			t.Errorf("%s: %d %v", name, code, out)
		}
	}
	if code, out := f.e.do("PUT", path, map[string]any{"item_ids": []uint{ids[2], ids[0], ids[1]}}, f.ownerTok); code != 204 {
		t.Fatalf("reorder: %d %v", code, out)
	}
	_, out := f.e.do("GET", fmt.Sprintf("/api/lists/%d", id), nil, f.ownerTok)
	items, _ := out["items"].([]any)
	if len(items) != 3 || uint(items[0].(map[string]any)["id"].(float64)) != ids[2] || uint(items[2].(map[string]any)["id"].(float64)) != ids[1] {
		t.Fatalf("order after reorder: %v", items)
	}
}

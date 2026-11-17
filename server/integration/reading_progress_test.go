//go:build integration

package integration

import (
	"fmt"
	"sync"
	"testing"
)

// readingPath is the owner-only progress route for one manga.
func readingPath(id int) string {
	return fmt.Sprintf("/api/reading-progress/manga/%d", id)
}

// readingReader seeds a verified reader aged 30 (passes the adult gate) and returns their access token.
func readingReader(t *testing.T, e *env, email string) (uint, string) {
	t.Helper()
	const password = "reader-password1"
	uid := seedUser(t, e.db, email, true, password)
	setBirthDate(t, e, uid, yearsAgo(30))
	return uid, e.loginPair(email, password).AccessToken
}

func progressRows(e *env, uid uint, mediaID int) int64 {
	return e.count(`SELECT count(*) FROM reading_progress WHERE user_id = ? AND media_id = ? AND media_type = 'manga'`, uid, mediaID)
}

func TestReadingProgressUpsertReplacesOneRow(t *testing.T) {
	e := readGateEnv(t)
	uid, tok := readingReader(t, e, "reader@x.test")
	path := readingPath(mangaSFW)

	if code, out := e.do("GET", path, nil, tok); code != 200 || out["progress"] != nil {
		t.Fatalf("no progress yet: want 200 null, got %d %v", code, out)
	}
	code, out := e.do("PUT", path, map[string]any{"chapter": 3, "volume": 1}, tok)
	if code != 200 {
		t.Fatalf("create: %d %v", code, out)
	}
	code, out = e.do("PUT", path, map[string]any{"chapter": 12, "volume": 0}, tok)
	if code != 200 {
		t.Fatalf("update: %d %v", code, out)
	}
	p, _ := out["progress"].(map[string]any)
	if p["chapter"] != float64(12) || p["volume"] != float64(0) || p["media_type"] != "manga" {
		t.Errorf("full replace: got %v", p)
	}
	if n := progressRows(e, uid, mangaSFW); n != 1 {
		t.Errorf("rows after create and update: %d, want 1", n)
	}
}

func TestReadingProgressConcurrentWritesOneRow(t *testing.T) {
	e := readGateEnv(t)
	uid, tok := readingReader(t, e, "concurrent@x.test")
	path := readingPath(mangaSFW)

	const writers = 10
	codes := make([]int, writers)
	var wg sync.WaitGroup
	for i := range writers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			codes[i], _ = e.do("PUT", path, map[string]any{"chapter": i + 1, "volume": 1}, tok)
		}()
	}
	wg.Wait()

	for i, code := range codes {
		if code != 200 {
			t.Errorf("writer %d: status %d, want 200", i, code)
		}
	}
	if n := progressRows(e, uid, mangaSFW); n != 1 {
		t.Errorf("rows after %d concurrent PUTs: %d, want 1", writers, n)
	}
}

func TestReadingProgressDeleteIsIdempotent(t *testing.T) {
	e := readGateEnv(t)
	uid, tok := readingReader(t, e, "delete@x.test")
	path := readingPath(mangaSFW)

	if code, out := e.do("PUT", path, map[string]any{"chapter": 2}, tok); code != 200 {
		t.Fatalf("seed: %d %v", code, out)
	}
	for i := range 2 {
		if code, out := e.do("DELETE", path, nil, tok); code != 204 {
			t.Fatalf("delete %d: %d %v", i+1, code, out)
		}
	}
	if code, out := e.do("GET", path, nil, tok); code != 200 || out["progress"] != nil {
		t.Errorf("after delete: want 200 null, got %d %v", code, out)
	}
	if n := progressRows(e, uid, mangaSFW); n != 0 {
		t.Errorf("rows after delete: %d, want 0", n)
	}
}

func TestReadingProgressRequiresXHRHeader(t *testing.T) {
	e := readGateEnv(t)
	_, tok := readingReader(t, e, "xhr@x.test")

	// doWith sends no X-Requested-With header, unlike do.
	code, out := e.doWith("PUT", readingPath(mangaSFW), map[string]any{"chapter": 1}, tok, nil, "application/json")
	if code != 403 || out["code"] != "XHR_HEADER_REQUIRED" {
		t.Errorf("PUT without XHR header: want 403 XHR_HEADER_REQUIRED, got %d %v", code, out)
	}
}

func TestReadingProgressRefusesNonMangaAndBadIDs(t *testing.T) {
	e := readGateEnv(t)
	_, tok := readingReader(t, e, "types@x.test")

	cases := []struct {
		method, path, code string
	}{
		{"PUT", "/api/reading-progress/movie/603", "INVALID_MEDIA_TYPE"},
		{"GET", "/api/reading-progress/tv/5", "INVALID_MEDIA_TYPE"},
		{"DELETE", "/api/reading-progress/movie/603", "INVALID_MEDIA_TYPE"},
		{"GET", "/api/reading-progress/manga/0", "INVALID_ID"},
	}
	for _, c := range cases {
		body := map[string]any{"chapter": 1}
		code, out := e.do(c.method, c.path, body, tok)
		if code != 400 || out["code"] != c.code {
			t.Errorf("%s %s: want 400 %s, got %d %v", c.method, c.path, c.code, code, out)
		}
	}
}

func TestReadingProgressRangeAndBodyErrors(t *testing.T) {
	e := readGateEnv(t)
	uid, tok := readingReader(t, e, "range@x.test")
	path := readingPath(mangaSFW)

	cases := []struct {
		name string
		body any
		code string
	}{
		{"chapter over cap", map[string]any{"chapter": 10000}, "PROGRESS_OUT_OF_RANGE"},
		{"volume negative", map[string]any{"chapter": 1, "volume": -1}, "PROGRESS_OUT_OF_RANGE"},
		{"both zero", map[string]any{"chapter": 0, "volume": 0}, "INVALID_INPUT"},
		{"not JSON", "not json", "INVALID_BODY"},
	}
	for _, c := range cases {
		code, out := e.do("PUT", path, c.body, tok)
		if code != 400 || out["code"] != c.code {
			t.Errorf("%s: want 400 %s, got %d %v", c.name, c.code, code, out)
		}
	}
	if n := progressRows(e, uid, mangaSFW); n != 0 {
		t.Errorf("refused writes left %d rows", n)
	}
}

func TestReadingProgressAdultWriteNeedsEligibility(t *testing.T) {
	e := readGateEnv(t)
	seedUser(t, e.db, "nodob-progress@x.test", true, "nodob-password1")
	tok := e.loginPair("nodob-progress@x.test", "nodob-password1").AccessToken
	path := readingPath(mangaAdult)

	if code, out := e.do("PUT", path, map[string]any{"chapter": 1}, tok); code != 403 || out["code"] != "AGE_RESTRICTED" {
		t.Errorf("adult write without DOB: want 403 AGE_RESTRICTED, got %d %v", code, out)
	}
	if code, out := e.do("GET", path, nil, tok); code != 200 || out["progress"] != nil {
		t.Errorf("adult read is not age-gated: want 200 null, got %d %v", code, out)
	}
}

// The DB CHECKs back up the service validation for any writer that skips the API.
func TestReadingProgressChecksRejectBadRows(t *testing.T) {
	e := newEnv(t)
	uid := seedUser(t, e.db, "checks@x.test", true, "")

	insert := func(mediaID, chapter, volume int, mediaType string) error {
		return e.db.Exec(`INSERT INTO reading_progress (user_id, media_id, media_type, chapter, volume, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, now(), now())`, uid, mediaID, mediaType, chapter, volume).Error
	}
	if err := insert(mangaSFW, 9999, 999, "manga"); err != nil {
		t.Fatalf("caps are inclusive: %v", err)
	}
	rejected := []struct {
		name            string
		mediaType       string
		chapter, volume int
	}{
		{"movie media type", "movie", 1, 0},
		{"chapter over cap", "manga", 10000, 0},
		{"volume over cap", "manga", 1, 1000},
		{"negative chapter", "manga", -1, 1},
		{"both zero", "manga", 0, 0},
	}
	for i, c := range rejected {
		// Each case gets its own title so only the CHECK can refuse it, not the unique index.
		if err := insert(mangaSFW+1+i, c.chapter, c.volume, c.mediaType); err == nil {
			t.Errorf("%s: row accepted", c.name)
		}
	}
}

//go:build integration

package integration

import (
	"fmt"
	"testing"
	"time"

	manga_module "github.com/arinsuda/movie-hub/internal/manga_module"
	"github.com/arinsuda/movie-hub/internal/shared"
)

// readGateEnv wires the real date-of-birth gate, so the under-18 and 18+ cases run the production rule.
func readGateEnv(t *testing.T) *env {
	t.Helper()
	e := newEnv(t)
	shared.SetMangaSupport(fakeAniList{}, manga_module.NewDOBGate(e.db), e.db)
	t.Cleanup(func() { shared.SetMangaSupport(nil, nil, nil) })
	return e
}

// setBirthDate stores users.date_of_birth; nil clears it.
func setBirthDate(t *testing.T, e *env, uid uint, dob *time.Time) {
	t.Helper()
	if err := e.db.Exec(`UPDATE users SET date_of_birth = ? WHERE id = ?`, dob, uid).Error; err != nil {
		t.Fatal(err)
	}
}

func yearsAgo(n int) *time.Time {
	d := time.Now().AddDate(-n, 0, 0)
	return &d
}

func TestAdultMangaReadsNeedEligibility(t *testing.T) {
	e := readGateEnv(t)
	writer := seedUser(t, e.db, "writer@x.test", true, "writer-password1")
	setBirthDate(t, e, writer, yearsAgo(30))
	writerTok := e.loginPair("writer@x.test", "writer-password1").AccessToken

	seedUser(t, e.db, "nodob@x.test", true, "nodob-password1")
	minor := seedUser(t, e.db, "minor@x.test", true, "minor-password1")
	setBirthDate(t, e, minor, yearsAgo(17))
	adult := seedUser(t, e.db, "adult@x.test", true, "adult-password1")
	setBirthDate(t, e, adult, yearsAgo(18))

	// The writer is 30, so the adult title can carry a review and a comment.
	code, out := e.do("POST", "/api/reviews", map[string]any{"media_id": mangaAdult, "media_type": "manga", "rating": 4.0, "body": "adult review"}, writerTok)
	if code != 201 {
		t.Fatalf("adult review: %d %v", code, out)
	}
	reviewID := int(out["review"].(map[string]any)["id"].(float64))
	if code, out := e.do("POST", fmt.Sprintf("/api/reviews/%d/comments", reviewID), map[string]any{"body": "adult comment"}, writerTok); code != 201 {
		t.Fatalf("adult comment: %d %v", code, out)
	}

	// Comment listing sits behind the session, so only signed-in viewers reach it.
	publicReads := []string{
		fmt.Sprintf("/api/reviews/media/manga/%d", mangaAdult),
		fmt.Sprintf("/api/reviews/media/manga/%d/rating", mangaAdult),
		fmt.Sprintf("/api/stats/manga/%d", mangaAdult),
	}
	commentsPath := fmt.Sprintf("/api/reviews/%d/comments", reviewID)
	refused := []struct {
		name, tok string
	}{
		{"guest", ""},
		{"no DOB", e.loginPair("nodob@x.test", "nodob-password1").AccessToken},
		{"under 18", e.loginPair("minor@x.test", "minor-password1").AccessToken},
	}
	for _, v := range refused {
		for _, path := range publicReads {
			code, out := e.do("GET", path, nil, v.tok)
			if code != 403 || out["code"] != string(shared.ErrorCodeAgeRestricted) {
				t.Errorf("%s %s: want 403 AGE_RESTRICTED, got %d %v", v.name, path, code, out)
			}
		}
		if v.tok == "" {
			continue
		}
		if code, out := e.do("GET", commentsPath, nil, v.tok); code != 403 || out["code"] != string(shared.ErrorCodeAgeRestricted) {
			t.Errorf("%s comments: want 403 AGE_RESTRICTED, got %d %v", v.name, code, out)
		}
	}

	adultTok := e.loginPair("adult@x.test", "adult-password1").AccessToken
	for _, path := range append(publicReads, commentsPath) {
		if code, out := e.do("GET", path, nil, adultTok); code != 200 {
			t.Errorf("18+ %s: want 200, got %d %v", path, code, out)
		}
	}

	// Safe manga stays open to everyone, guests included.
	for _, path := range []string{
		fmt.Sprintf("/api/reviews/media/manga/%d", mangaSFW),
		fmt.Sprintf("/api/reviews/media/manga/%d/rating", mangaSFW),
		fmt.Sprintf("/api/stats/manga/%d", mangaSFW),
	} {
		if code, out := e.do("GET", path, nil, ""); code != 200 {
			t.Errorf("guest safe %s: want 200, got %d %v", path, code, out)
		}
	}
}

func TestMangaReadsFailClosedForUnknownTitle(t *testing.T) {
	e := readGateEnv(t)
	// An id AniList does not know is refused for everyone, as on the write side.
	code, out := e.do("GET", fmt.Sprintf("/api/stats/manga/%d", mangaMissing), nil, "")
	if code != 404 || out["code"] != string(shared.ErrorCodeNotFound) {
		t.Errorf("unknown manga stats: want 404 NOT_FOUND, got %d %v", code, out)
	}
}

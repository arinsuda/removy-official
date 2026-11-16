package manga_module

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	anilist "github.com/arinsuda/movie-hub/internal/anilist_module"
	"github.com/arinsuda/movie-hub/middleware"
	"github.com/gofiber/fiber/v3"
)

// adultTitleID is the fixture title that AniList flags as adult.
const adultTitleID = 2

// fixedNow anchors every age calculation in tests.
var fixedNow = time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)

// upstreamCall is one GraphQL request as the fake AniList received it.
type upstreamCall struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

// reply is one canned upstream response.
type reply struct {
	status     int
	body       string
	retryAfter string
}

// fakeAniList is a local stand-in for graphql.anilist.co that records every call.
type fakeAniList struct {
	srv   *httptest.Server
	mu    sync.Mutex
	calls []upstreamCall
}

func newFakeAniList(t *testing.T, respond func(upstreamCall) reply) *fakeAniList {
	t.Helper()
	f := &fakeAniList{}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var call upstreamCall
		_ = json.NewDecoder(r.Body).Decode(&call)
		f.mu.Lock()
		f.calls = append(f.calls, call)
		f.mu.Unlock()

		rep := respond(call)
		if rep.retryAfter != "" {
			w.Header().Set("Retry-After", rep.retryAfter)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(rep.status)
		_, _ = w.Write([]byte(rep.body))
	}))
	t.Cleanup(f.srv.Close)
	return f
}

// client returns an AniList client wired to this fake, with its own cache and budget.
func (f *fakeAniList) client() *anilist.Client {
	return anilist.NewClient(anilist.Options{Endpoint: f.srv.URL, HTTP: f.srv.Client()})
}

func (f *fakeAniList) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

func (f *fakeAniList) last() upstreamCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[len(f.calls)-1]
}

func ok(body string) reply { return reply{status: http.StatusOK, body: body} }

// standardUpstream answers like AniList for the fixtures in this file.
func standardUpstream(call upstreamCall) reply {
	switch {
	case strings.Contains(call.Query, "recommendations"):
		id := int(call.Variables["id"].(float64))
		return ok(fmt.Sprintf(`{"data":{"Media":{"id":%d,"type":"MANGA","isAdult":%t,"recommendations":{"nodes":[`, id, id == adultTitleID) +
			`{"mediaRecommendation":` + mediaFixture(3, false) + `},` +
			`{"mediaRecommendation":` + mediaFixture(4, true) + `},` +
			`{"mediaRecommendation":` + strings.Replace(mediaFixture(5, false), `"type": "MANGA"`, `"type": "ANIME"`, 1) + `}` +
			`]}}}}`)
	case call.Variables["id"] != nil:
		id := int(call.Variables["id"].(float64))
		return ok(fmt.Sprintf(`{"data":{"Media":%s}}`, mediaFixture(id, id == adultTitleID)))
	default:
		return ok(`{"data":{"Page":{"pageInfo":{"total":2,"currentPage":1,"lastPage":1},"media":[` +
			mediaFixture(1, false) + `,` + mediaFixture(adultTitleID, true) + `]}}}`)
	}
}

// mediaFixture is one AniList Media object with every field the queries select.
func mediaFixture(id int, adult bool) string {
	return fmt.Sprintf(`{
		"id": %d, "type": "MANGA", "isAdult": %t,
		"title": {"romaji": "Romaji %d", "english": "", "native": "Native %d"},
		"synonyms": ["Alt", "มังงะ %d"],
		"coverImage": {"large": "https://img.example/%d.jpg"},
		"bannerImage": "http://insecure.example/banner.jpg",
		"format": "MANGA", "countryOfOrigin": "KR", "status": "RELEASING",
		"startDate": {"year": 2020}, "chapters": 10, "volumes": null,
		"genres": ["Action"], "averageScore": 72,
		"description": "<p>Hi<br>there &amp; <b>you</b></p>",
		"tags": [
			{"name": "Isekai", "rank": 80, "isMediaSpoiler": false, "isGeneralSpoiler": false},
			{"name": "Twist", "rank": 90, "isMediaSpoiler": true, "isGeneralSpoiler": false},
			{"name": "Minor", "rank": 10, "isMediaSpoiler": false, "isGeneralSpoiler": false}
		],
		"externalLinks": [
			{"site": "Official", "url": "https://ex.example/read", "type": "STREAMING", "language": "English"},
			{"site": "Insecure", "url": "http://ex.example/read", "type": "INFO", "language": null},
			{"site": "Social", "url": "https://x.example", "type": "SOCIAL", "language": null}
		],
		"relations": {"edges": [
			{"relationType": "ADAPTATION", "node": {"id": 99, "type": "ANIME", "title": {"romaji": "R", "english": "E", "native": "N"}}},
			{"relationType": "SEQUEL", "node": {"id": 100, "type": "MANGA", "title": {"romaji": "S", "english": "", "native": ""}}}
		]},
		"recommendations": null
	}`, id, adult, id, id, id, id)
}

// viewer ids used by the gate matrix; guest is zero.
const (
	viewerGuest   uint = 0
	viewerNoDOB   uint = 1
	viewerUnder18 uint = 2
	viewerAdult   uint = 3
)

// fixedGate applies the real age rule to a fixed clock and a DOB table.
type fixedGate struct {
	dob map[uint]*time.Time
}

func (g fixedGate) Eligible(_ context.Context, id uint) (bool, error) {
	if id == 0 {
		return false, nil
	}
	return eligibleAt(g.dob[id], fixedNow), nil
}

// matrixGate: no DOB, under 18, and exactly 18 with one day to spare.
func matrixGate() fixedGate {
	noDOB := (*time.Time)(nil)
	under := fixedNow.AddDate(-17, 0, 0)
	adult := fixedNow.AddDate(-18, 0, -1)
	return fixedGate{dob: map[uint]*time.Time{
		viewerNoDOB:   noDOB,
		viewerUnder18: &under,
		viewerAdult:   &adult,
	}}
}

// passThrough stands in for OptionalAuth/limiters in handler-level tests.
func passThrough(c fiber.Ctx) error { return c.Next() }

// viewerApp mounts the public manga routes on a bare app; the viewer's claims are set when viewer != 0.
func viewerApp(h *Handler, viewer uint) *fiber.App {
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		if viewer != 0 {
			c.Locals("claims", &middleware.Claims{UserID: viewer})
		}
		return c.Next()
	})
	RegisterPublicRoutes(app, h, passThrough)
	return app
}

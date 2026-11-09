package tmdbmodule

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDetailCacheHitAndExpiry(t *testing.T) {
	now := time.Unix(1000, 0)
	c := newDetailCache(time.Minute, 10)
	c.now = func() time.Time { return now }
	k := detailKey{detailKindMovie, 1, "th-TH"}

	if _, ok := c.get(k); ok {
		t.Fatal("unexpected hit on empty cache")
	}
	c.set(k, "v")
	if v, ok := c.get(k); !ok || v != "v" {
		t.Fatalf("want hit, got %v %v", v, ok)
	}
	if _, ok := c.get(detailKey{detailKindTV, 1, "th-TH"}); ok {
		t.Fatal("tv key must not collide with movie key")
	}
	now = now.Add(time.Minute)
	if _, ok := c.get(k); ok {
		t.Fatal("entry should have expired")
	}
}

func TestDetailCacheBounded(t *testing.T) {
	c := newDetailCache(time.Hour, 3)
	for i := 0; i < 20; i++ {
		c.set(detailKey{detailKindMovie, i, "th-TH"}, i)
	}
	if len(c.entries) > 3 {
		t.Fatalf("cache grew to %d entries", len(c.entries))
	}
}

func TestDetailCacheKeysByLanguage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		title := map[string]string{"th-TH": "ชื่อไทย", "en-US": "English title"}[r.URL.Query().Get("language")]
		_, _ = w.Write([]byte(`{"id":4242,"title":"` + title + `","overview":"o"}`))
	}))
	defer srv.Close()

	prev := c
	defer func() { c = prev }()
	for _, lang := range []string{"th-TH", "en-US"} {
		c = &client{http: srv.Client(), baseURL: srv.URL, apiKey: "k", language: lang}
		got, err := GetMovieByID(4242)
		if err != nil {
			t.Fatal(err)
		}
		want := map[string]string{"th-TH": "ชื่อไทย", "en-US": "English title"}[lang]
		if got.Title != want {
			t.Fatalf("%s: title %q, want %q (language leaked from cache)", lang, got.Title, want)
		}
	}
	for _, lang := range []string{"th-TH", "en-US"} {
		if _, ok := details.get(detailKey{detailKindMovie, 4242, lang}); !ok {
			t.Fatalf("no cache entry for %s", lang)
		}
	}
}

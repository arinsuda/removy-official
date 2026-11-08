package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func TestValidShareImageURL(t *testing.T) {
	ok := []string{
		"https://image.tmdb.org/t/p/w500/abc.jpg",
		"https://IMAGE.TMDB.ORG/t/p/original/abc.jpg",
		"https://image.tmdb.org:443/t/p/w500/abc.jpg",
	}
	bad := []string{
		"",
		"http://image.tmdb.org/t/p/w500/abc.jpg",
		"https://image.tmdb.org.evil.com/t/p/w500/abc.jpg",
		"https://evil.com/t/p/w500/abc.jpg",
		"https://image.tmdb.org@127.0.0.1/t/p/w500/abc.jpg",
		"https://user:pw@image.tmdb.org/t/p/w500/abc.jpg",
		"https://image.tmdb.org:8443/t/p/w500/abc.jpg",
		"https://image.tmdb.org/other/abc.jpg",
		"https://image.tmdb.org/t/p/../../admin",
		"http://169.254.169.254/latest/meta-data",
		"http://localhost:8081/api/health",
		"file:///etc/passwd",
	}
	for _, u := range ok {
		if _, valid := validShareImageURL(u); !valid {
			t.Errorf("expected %q to be allowed", u)
		}
	}
	for _, u := range bad {
		if _, valid := validShareImageURL(u); valid {
			t.Errorf("expected %q to be rejected", u)
		}
	}
}

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func testProxyApp(rt http.RoundTripper, inflight, rate int) *fiber.App {
	app := fiber.New()
	app.Get("/p", newShareImageProxyHandler(&http.Client{Transport: rt}, inflight, rate, 0))
	return app
}

func okImage(ct string) rtFunc {
	return func(*http.Request) (*http.Response, error) {
		h := http.Header{}
		h.Set("Content-Type", ct)
		return &http.Response{StatusCode: 200, Header: h, Body: io.NopCloser(strings.NewReader("img"))}, nil
	}
}

func proxyGet(t *testing.T, app *fiber.App, raw string) *http.Response {
	t.Helper()
	resp, err := app.Test(httptest.NewRequest("GET", "/p?url="+url.QueryEscape(raw), nil))
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestShareProxySizeAllowlist(t *testing.T) {
	app := testProxyApp(okImage("image/jpeg"), 4, 1000)
	for _, size := range []string{"w92", "w185", "w342", "w500", "w780"} {
		if r := proxyGet(t, app, "https://image.tmdb.org/t/p/"+size+"/a.jpg"); r.StatusCode != 200 {
			t.Errorf("%s: got %d", size, r.StatusCode)
		}
	}
	for _, size := range []string{"original", "w1280", "h632", "foo"} {
		if r := proxyGet(t, app, "https://image.tmdb.org/t/p/"+size+"/a.jpg"); r.StatusCode != 400 {
			t.Errorf("%s: got %d, want 400", size, r.StatusCode)
		}
	}
}

func TestShareProxyContentTypeAndHeaders(t *testing.T) {
	for _, ct := range []string{"image/jpeg", "image/png", "image/webp", "image/avif", "image/jpeg; charset=x"} {
		r := proxyGet(t, testProxyApp(okImage(ct), 4, 1000), "https://image.tmdb.org/t/p/w500/a.jpg")
		if r.StatusCode != 200 {
			t.Fatalf("%s: got %d", ct, r.StatusCode)
		}
		if r.Header.Get("X-Content-Type-Options") != "nosniff" || r.Header.Get("Content-Security-Policy") != "default-src 'none'" {
			t.Errorf("%s: missing security headers", ct)
		}
	}
	for _, ct := range []string{"image/svg+xml", "text/html", "image/gif", ""} {
		if r := proxyGet(t, testProxyApp(okImage(ct), 4, 1000), "https://image.tmdb.org/t/p/w500/a.jpg"); r.StatusCode != 502 {
			t.Errorf("%q: got %d, want 502", ct, r.StatusCode)
		}
	}
}

func TestShareProxyDropsQuery(t *testing.T) {
	var got string
	rt := rtFunc(func(r *http.Request) (*http.Response, error) {
		got = r.URL.String()
		return okImage("image/png")(r)
	})
	r := proxyGet(t, testProxyApp(rt, 4, 1000), "https://image.tmdb.org/t/p/w500/a.jpg?x=1&api_key=z#frag")
	if r.StatusCode != 200 || got != "https://image.tmdb.org/t/p/w500/a.jpg" {
		t.Errorf("status %d, upstream %q", r.StatusCode, got)
	}
}

func TestShareProxyRateLimit(t *testing.T) {
	app := testProxyApp(okImage("image/png"), 4, 3)
	for i := 0; i < 3; i++ {
		if r := proxyGet(t, app, "https://image.tmdb.org/t/p/w500/a.jpg"); r.StatusCode != 200 {
			t.Fatalf("req %d: %d", i, r.StatusCode)
		}
	}
	if r := proxyGet(t, app, "https://image.tmdb.org/t/p/w500/a.jpg"); r.StatusCode != 429 {
		t.Errorf("got %d, want 429", r.StatusCode)
	}
}

func TestIPLimiterRefillAndBound(t *testing.T) {
	now := time.Unix(0, 0)
	l := newIPLimiter(60, 2)
	l.now = func() time.Time { return now }
	for i := 0; i < 60; i++ {
		if !l.allow("a") {
			t.Fatalf("burst %d denied", i)
		}
	}
	if l.allow("a") {
		t.Error("expected deny after burst")
	}
	now = now.Add(2 * time.Second)
	if !l.allow("a") {
		t.Error("expected refill")
	}
	l.allow("b")
	if l.allow("c") {
		t.Error("map bound should refuse new key while full of active clients")
	}
	now = now.Add(2 * time.Minute)
	if !l.allow("c") {
		t.Error("cleanup should free idle clients")
	}
}

func TestShareProxySemaphore(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	rt := rtFunc(func(r *http.Request) (*http.Response, error) {
		started <- struct{}{}
		<-release
		return okImage("image/png")(r)
	})
	app := testProxyApp(rt, 1, 1000)
	done := make(chan int)
	go func() { done <- proxyGet(t, app, "https://image.tmdb.org/t/p/w500/a.jpg").StatusCode }()
	<-started
	if r := proxyGet(t, app, "https://image.tmdb.org/t/p/w500/b.jpg"); r.StatusCode != 503 {
		t.Errorf("got %d, want 503", r.StatusCode)
	}
	close(release)
	if c := <-done; c != 200 {
		t.Errorf("first request: %d", c)
	}
}

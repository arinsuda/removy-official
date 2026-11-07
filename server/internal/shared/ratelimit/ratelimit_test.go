package ratelimit

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func ipKey(c fiber.Ctx) string { return "ip:" + c.Get("X-Test-IP") }

func newApp(keyFn func(fiber.Ctx) string, max int) *fiber.App {
	app := fiber.New()
	app.Post("/", New(max, time.Hour, keyFn), func(c fiber.Ctx) error { return c.SendString("ok") })
	return app
}

func post(t *testing.T, app *fiber.App, ip, body string) *fiberResp {
	t.Helper()
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-IP", ip)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	return &fiberResp{status: resp.StatusCode, retryAfter: resp.Header.Get("Retry-After"), body: string(b)}
}

type fiberResp struct {
	status     int
	retryAfter string
	body       string
}

func TestLimitExceededReturns429(t *testing.T) {
	app := newApp(ipKey, 3)
	for i := 0; i < 3; i++ {
		if r := post(t, app, "1.1.1.1", "{}"); r.status != 200 {
			t.Fatalf("request %d: %d", i+1, r.status)
		}
	}
	r := post(t, app, "1.1.1.1", "{}")
	if r.status != 429 {
		t.Fatalf("want 429, got %d", r.status)
	}
	if r.retryAfter == "" {
		t.Fatal("missing Retry-After")
	}
	if !strings.Contains(r.body, `"code":"RATE_LIMITED"`) || !strings.Contains(r.body, `"error":"too many requests"`) {
		t.Fatalf("body: %s", r.body)
	}
	if other := post(t, app, "2.2.2.2", "{}"); other.status != 200 {
		t.Fatalf("other IP blocked: %d", other.status)
	}
}

func TestBodyFieldKeyHashesAndNormalizes(t *testing.T) {
	var keys []string
	app := fiber.New()
	app.Post("/", func(c fiber.Ctx) error {
		keys = append(keys, BodyFieldKey("email")(c))
		return nil
	})
	for _, body := range []string{`{"email":"User@Example.com"}`, `{"email":"  user@example.com "}`, `{}`} {
		req := httptest.NewRequest("POST", "/", strings.NewReader(body))
		if _, err := app.Test(req); err != nil {
			t.Fatal(err)
		}
	}
	if keys[0] != keys[1] {
		t.Fatalf("not normalized: %q vs %q", keys[0], keys[1])
	}
	if strings.Contains(strings.ToLower(keys[0]), "example") || strings.Contains(keys[0], "@") {
		t.Fatalf("raw email in key: %q", keys[0])
	}
	if keys[0] != "id:"+HashKey("user@example.com") || len(HashKey("x")) != 64 {
		t.Fatalf("unexpected key %q", keys[0])
	}
	if keys[2] != "" {
		t.Fatalf("missing field should give empty key, got %q", keys[2])
	}
}

func TestEmptyKeySkipsLimit(t *testing.T) {
	app := newApp(BodyFieldKey("email"), 1)
	for i := 0; i < 3; i++ {
		if r := post(t, app, "1.1.1.1", "{}"); r.status != 200 {
			t.Fatalf("unkeyed request limited: %d", r.status)
		}
	}
	for i, want := range []int{200, 429} {
		if r := post(t, app, "1.1.1.1", `{"email":"a@b.co"}`); r.status != want {
			t.Fatalf("keyed request %d: got %d want %d", i, r.status, want)
		}
	}
}

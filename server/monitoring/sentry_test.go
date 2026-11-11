package monitoring

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v3"
)

func TestScrubEvent(t *testing.T) {
	e := &sentry.Event{
		Message: "failed for bob@example.com",
		Request: &sentry.Request{
			Method:      "GET",
			URL:         "https://x.test/api/a?token=abc#frag",
			Cookies:     "access_token=secret",
			Data:        "body",
			QueryString: "token=abc",
			Headers:     map[string]string{"Authorization": "Bearer x"},
		},
		User:        sentry.User{ID: "7", Email: "bob@example.com", IPAddress: "1.2.3.4"},
		Breadcrumbs: []*sentry.Breadcrumb{{Message: "mail a@b.co", Data: map[string]interface{}{"k": "v"}}},
	}
	got := ScrubEvent(e)
	if got.Message != "failed for [email]" || got.Breadcrumbs[0].Message != "mail [email]" {
		t.Fatalf("email not redacted: %q %q", got.Message, got.Breadcrumbs[0].Message)
	}
	r := got.Request
	if r.URL != "https://x.test/api/a" || r.Cookies != "" || r.Data != "" || r.QueryString != "" || r.Headers != nil {
		t.Fatalf("request not scrubbed: %+v", r)
	}
	if got.User.ID != "7" || got.User.Email != "" || got.User.IPAddress != "" {
		t.Fatalf("user not scrubbed: %+v", got.User)
	}
	if got.Breadcrumbs[0].Data != nil {
		t.Fatal("breadcrumb data kept")
	}
}

func TestShouldReport(t *testing.T) {
	for status, want := range map[int]bool{200: false, 401: false, 404: false, 429: false, 499: false, 500: true, 503: true} {
		if ShouldReport(status) != want {
			t.Errorf("ShouldReport(%d) != %v", status, want)
		}
	}
}

func TestInitDisabledWithoutDSN(t *testing.T) {
	if Init(Options{}) {
		t.Fatal("Init enabled without DSN")
	}
}

func TestMiddlewareRecoversPanic(t *testing.T) {
	app := fiber.New()
	app.Use(Middleware())
	app.Get("/boom", func(fiber.Ctx) error { panic("x") })
	app.Get("/ok", func(c fiber.Ctx) error { return c.SendStatus(http.StatusOK) })
	for path, want := range map[string]int{"/boom": 500, "/ok": 200} {
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
		if err != nil || resp.StatusCode != want {
			t.Fatalf("%s: status=%v err=%v", path, resp, err)
		}
	}
}

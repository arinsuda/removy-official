package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestSecurityHeaders(t *testing.T) {
	app := fiber.New()
	app.Use(SecurityHeaders())
	app.Get("/", func(c fiber.Ctx) error { return c.JSON(fiber.Map{"ok": true}) })

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "no-referrer",
	}
	for k, v := range want {
		if got := resp.Header.Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
	if resp.Header.Get("Strict-Transport-Security") == "" {
		t.Error("missing HSTS on https")
	}
	if resp.Header.Get("Content-Security-Policy") == "" {
		t.Error("missing CSP")
	}
	plain, _ := app.Test(httptest.NewRequest("GET", "/", nil))
	if plain.Header.Get("Strict-Transport-Security") != "" {
		t.Error("HSTS must not be sent over http")
	}
}

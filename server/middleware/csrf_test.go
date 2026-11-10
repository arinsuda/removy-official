package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func csrfApp() *fiber.App {
	app := fiber.New()
	api := app.Group("/api")
	api.Use(RequireXHR(CSRFExemptPrefixes))
	ok := func(c fiber.Ctx) error { return c.SendString("ok") }
	api.Get("/x", ok)
	api.Post("/x", ok)
	api.Put("/x", ok)
	api.Patch("/x", ok)
	api.Delete("/x", ok)
	api.Post("/internal/link/introspect", ok)
	api.Post("/auth/login", ok)
	return app
}

func status(t *testing.T, app *fiber.App, method, path string, hdr map[string]string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader("{}"))
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func TestRequireXHR(t *testing.T) {
	app := csrfApp()
	xhr := map[string]string{HeaderRequestedWith: RequestedWithXHR}
	cases := []struct {
		name, method, path string
		hdr                map[string]string
		want               int
	}{
		{"post without header", "POST", "/api/x", nil, 403},
		{"form text/plain without header", "POST", "/api/x", map[string]string{"Content-Type": "text/plain"}, 403},
		{"put without header", "PUT", "/api/x", nil, 403},
		{"patch without header", "PATCH", "/api/x", nil, 403},
		{"delete without header", "DELETE", "/api/x", nil, 403},
		{"login without header", "POST", "/api/auth/login", nil, 403},
		{"wrong header value", "POST", "/api/x", map[string]string{HeaderRequestedWith: "fetch"}, 403},
		{"post with header", "POST", "/api/x", xhr, 200},
		{"get without header", "GET", "/api/x", nil, 200},
		{"hmac link route without header", "POST", "/api/internal/link/introspect", nil, 200},
	}
	for _, tc := range cases {
		if got := status(t, app, tc.method, tc.path, tc.hdr); got != tc.want {
			t.Errorf("%s: got %d want %d", tc.name, got, tc.want)
		}
	}
}

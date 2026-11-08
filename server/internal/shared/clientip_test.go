package shared

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestClientIPFromHeader(t *testing.T) {
	const fb = "10.9.9.9"
	cases := []struct {
		name, xff string
		hops      int
		want      string
	}{
		{"spoofed leftmost ignored", "6.6.6.6, 203.0.113.7", 1, "203.0.113.7"},
		{"hops 1", "1.1.1.1, 2.2.2.2, 3.3.3.3", 1, "3.3.3.3"},
		{"hops 2", "1.1.1.1, 2.2.2.2, 3.3.3.3", 2, "2.2.2.2"},
		{"zero hops falls back", "1.1.1.1", 0, fb},
		{"missing header", "", 1, fb},
		{"too short", "1.1.1.1", 2, fb},
		{"invalid value", "1.1.1.1, not-an-ip", 1, fb},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := clientIPFromHeader(tc.xff, tc.hops, fb); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestClientIPUsesHeader(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error { return c.SendString(ClientIP(c, 1)) })
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "6.6.6.6, 203.0.113.7")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	if string(b) != "203.0.113.7" {
		t.Fatalf("got %q", b)
	}
}

func TestLimitKeyIP(t *testing.T) {
	cases := map[string]string{
		"203.0.113.7":                "203.0.113.7",
		"2001:db8:1:2:aaaa:bbbb:c:d": "2001:db8:1:2::/64",
		"2001:db8:1:2:1111:2222:3:4": "2001:db8:1:2::/64",
		"::ffff:203.0.113.7":         "::ffff:203.0.113.7",
		"not-an-ip":                  "not-an-ip",
	}
	for in, want := range cases {
		if got := limitKeyIP(in); got != want {
			t.Fatalf("%s: got %q want %q", in, got, want)
		}
	}
}

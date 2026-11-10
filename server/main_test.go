package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func TestMergeForwardedFor(t *testing.T) {
	h := http.Header{}
	h.Add("X-Forwarded-For", "6.6.6.6")
	h.Add("X-Forwarded-For", "1.1.1.1, 2.2.2.2")
	mergeForwardedFor(h)
	if got := h.Values("X-Forwarded-For"); len(got) != 1 || got[0] != "6.6.6.6, 1.1.1.1, 2.2.2.2" {
		t.Fatalf("got %q", got)
	}
	single := http.Header{"X-Forwarded-For": {"3.3.3.3"}}
	mergeForwardedFor(single)
	if single.Get("X-Forwarded-For") != "3.3.3.3" {
		t.Fatal("single value changed")
	}
}

func TestCORSPreflightAllowsRequestedWith(t *testing.T) {
	const origin = "https://app.example.test"
	app := fiber.New()
	app.Use(cors.New(corsConfig([]string{origin})))
	app.Post("/api/x", func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	req := httptest.NewRequest(http.MethodOptions, "/api/x", nil)
	req.Header.Set("Origin", origin)
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "x-requested-with,content-type")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	got := strings.ToLower(resp.Header.Get("Access-Control-Allow-Headers"))
	if !strings.Contains(got, "x-requested-with") {
		t.Fatalf("Allow-Headers = %q", got)
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != origin {
		t.Fatalf("origin not allowed: %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
}

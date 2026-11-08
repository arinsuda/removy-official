package link_module

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/arinsuda/movie-hub/config"
	"github.com/gofiber/fiber/v3"
)

func remoteStatus(t *testing.T, link config.LinkConfig) (int, string) {
	t.Helper()
	cfg := &config.Config{Link: link}
	h := &Handler{svc: &Service{cfg: link}, cfg: cfg}
	app := fiber.New()
	app.Get("/link/remote", h.Remote)
	resp, err := app.Test(httptest.NewRequest("GET", "/link/remote", nil))
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

func TestRemote(t *testing.T) {
	code, body := remoteStatus(t, config.LinkConfig{Secrets: []string{"s"}, RemoteWebURL: "https://planzi.example"})
	if code != 200 || body != `{"web_url":"https://planzi.example"}` {
		t.Fatalf("enabled: %d %s", code, body)
	}
	if code, _ := remoteStatus(t, config.LinkConfig{RemoteWebURL: "https://planzi.example"}); code != 404 {
		t.Fatalf("disabled should be 404, got %d", code)
	}
	if code, _ := remoteStatus(t, config.LinkConfig{Secrets: []string{"s"}}); code != 404 {
		t.Fatalf("no url should be 404, got %d", code)
	}
}

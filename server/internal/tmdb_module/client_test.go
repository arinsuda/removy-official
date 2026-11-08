package tmdbmodule

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetErrorNeverLeaksAPIKey(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	base := srv.URL
	srv.Close() // connection refused: *url.Error carrying the full URL

	prev := c
	defer func() { c = prev }()
	c = &client{http: &http.Client{Timeout: time.Second}, baseURL: base, apiKey: "SECRET-KEY-123", language: "en-US"}

	err := get("/movie/1", nil, &struct{}{})
	if err == nil {
		t.Fatal("expected an error")
	}
	if strings.Contains(err.Error(), "SECRET-KEY-123") || strings.Contains(err.Error(), "api_key") {
		t.Fatalf("error leaks api key: %v", err)
	}
}

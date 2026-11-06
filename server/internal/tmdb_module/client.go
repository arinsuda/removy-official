package tmdbmodule

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/arinsuda/movie-hub/config"
)

const (
	defaultLanguage  = "th-TH"
	fallbackLanguage = "en-US"
	imageBaseURL     = "https://image.tmdb.org/t/p/w500"
)

type client struct {
	http     *http.Client
	baseURL  string
	apiKey   string
	language string
}

var c *client

func Init(cfg *config.Config) {
	c = &client{
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL:  cfg.TMDB.BaseURL,
		apiKey:   cfg.TMDB.APIKey,
		language: defaultLanguage,
	}
}

func ImageURL(path string) string {
	if path == "" {
		return ""
	}
	return imageBaseURL + path
}

func get(path string, params url.Values, target any) error {
	if c == nil {
		return fmt.Errorf("tmdb: client not initialized, call Init() first")
	}
	if params == nil {
		params = url.Values{}
	}
	params.Set("api_key", c.apiKey)
	if params.Get("language") == "" {
		params.Set("language", c.language)
	}

	fullURL := fmt.Sprintf("%s%s?%s", c.baseURL, path, params.Encode())

	resp, err := c.http.Get(fullURL)
	if err != nil {
		// *url.Error embeds the full request URL, including api_key; keep only the cause.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return fmt.Errorf("tmdb: GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tmdb: GET %s: status %d: %s", path, resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("tmdb: GET %s: decode failed: %w", path, err)
	}

	return nil
}

func getWithFallback(path string, params url.Values, target any, isEmpty func() bool) error {
	if err := get(path, params, target); err != nil {
		return err
	}

	if isEmpty() {
		if params == nil {
			params = url.Values{}
		}
		params.Set("language", fallbackLanguage)
		return get(path, params, target)
	}

	return nil
}

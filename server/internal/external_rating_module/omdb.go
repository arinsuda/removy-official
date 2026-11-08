package external_rating_module

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var errOMDbNotFound = errors.New("omdb: title not found")

// omdbResult is the parsed subset of an OMDb response. Nil fields mean "N/A".
type omdbResult struct {
	IMDbRating *float64
	IMDbVotes  *int
	RTScore    *int
}

type omdbResponse struct {
	Response   string `json:"Response"`
	IMDbRating string `json:"imdbRating"`
	IMDbVotes  string `json:"imdbVotes"`
	Ratings    []struct {
		Source string `json:"Source"`
		Value  string `json:"Value"`
	} `json:"Ratings"`
}

// parseOMDb decodes an OMDb body. It returns errOMDbNotFound when OMDb says Response is False.
func parseOMDb(body []byte) (omdbResult, error) {
	var raw omdbResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return omdbResult{}, fmt.Errorf("omdb: decode failed: %w", err)
	}
	if strings.EqualFold(raw.Response, "False") {
		return omdbResult{}, errOMDbNotFound
	}

	var out omdbResult
	if v, ok := parseNumber(raw.IMDbRating); ok {
		out.IMDbRating = &v
	}
	if v, ok := parseNumber(strings.ReplaceAll(raw.IMDbVotes, ",", "")); ok {
		n := int(v)
		out.IMDbVotes = &n
	}
	for _, r := range raw.Ratings {
		switch r.Source {
		case "Internet Movie Database":
			// Fallback when the top-level field is missing: "8.1/10".
			if out.IMDbRating == nil {
				if v, ok := parseNumber(strings.TrimSuffix(r.Value, "/10")); ok {
					out.IMDbRating = &v
				}
			}
		case "Rotten Tomatoes":
			if v, ok := parseNumber(strings.TrimSuffix(r.Value, "%")); ok {
				n := int(v)
				out.RTScore = &n
			}
		}
	}
	return out, nil
}

func parseNumber(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "N/A") {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// omdbClient talks to OMDb. The API key is never included in returned errors.
type omdbClient struct {
	http    *http.Client
	baseURL string
	apiKey  string
}

func (c *omdbClient) fetch(ctx context.Context, imdbID string) (omdbResult, error) {
	q := url.Values{}
	q.Set("apikey", c.apiKey)
	q.Set("i", imdbID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/?"+q.Encode(), nil)
	if err != nil {
		return omdbResult{}, errors.New("omdb: build request failed")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err // *url.Error embeds the full URL including the key
		}
		return omdbResult{}, fmt.Errorf("omdb: request failed: %w", err)
	}
	defer resp.Body.Close()

	// OMDb answers unknown ids with 200 + Response False; other statuses are transient or key problems.
	if resp.StatusCode != http.StatusOK {
		return omdbResult{}, fmt.Errorf("omdb: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return omdbResult{}, fmt.Errorf("omdb: read failed: %w", err)
	}
	return parseOMDb(body)
}

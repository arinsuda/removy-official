// Package anilist_module is the only caller of the AniList GraphQL API.
package anilist_module

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"golang.org/x/sync/singleflight"
)

var (
	// ErrNotFound means AniList has no such title (HTTP 404 or an empty Media).
	ErrNotFound = errors.New("anilist: not found")
	// ErrUnavailable covers network failures, 429 and local budget exhaustion, 5xx and bad payloads.
	ErrUnavailable = errors.New("anilist: unavailable")
)

// Options configures a Client. Zero values fall back to the config package.
type Options struct {
	Endpoint string
	HTTP     *http.Client
	Now      func() time.Time
}

// Client posts GraphQL queries to AniList with a local budget, a 429 cooldown,
// a bounded TTL cache and singleflight de-duplication of identical in-flight queries.
type Client struct {
	endpoint string
	http     *http.Client
	now      func() time.Time
	limiter  *tokenBucket
	cache    *ttlCache
	// misses remembers keys that AniList answered with ErrNotFound, for config.AniListNotFoundTTL.
	misses *ttlCache
	group  singleflight.Group

	mu            sync.Mutex
	cooldownUntil time.Time
}

// NewClient builds a Client; tests pass Endpoint and HTTP to stay off the network.
func NewClient(opts Options) *Client {
	if opts.Endpoint == "" {
		opts.Endpoint = config.AniListGraphQLURL
	}
	if opts.HTTP == nil {
		opts.HTTP = &http.Client{Timeout: config.AniListRequestTimeout}
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	return &Client{
		endpoint: opts.Endpoint,
		http:     opts.HTTP,
		now:      opts.Now,
		limiter:  newTokenBucket(config.AniListRequestsPerMinute, config.AniListRequestsPerMinute, opts.Now),
		cache:    newTTLCache(config.AniListCacheMaxEntries, opts.Now),
		misses:   newTTLCache(config.AniListCacheMaxEntries, opts.Now),
	}
}

// Query runs a GraphQL query and returns its raw "data" object. Successful responses are cached for
// ttl under scope, the query text and the variables; ErrNotFound is cached for config.AniListNotFoundTTL.
// Identical concurrent calls share one request. scope separates entries whose content depends on the
// viewer (for example adult titles). It is never sent upstream.
func (c *Client) Query(ctx context.Context, scope, query string, vars map[string]any, ttl time.Duration) (json.RawMessage, error) {
	key, err := cacheKey(scope, query, vars)
	if err != nil {
		return nil, ErrUnavailable
	}
	if raw, ok, err := c.fromCache(key); err != nil || ok {
		return raw, err
	}
	v, err, _ := c.group.Do(key, func() (any, error) {
		// A call that finished just before we joined has already cached its result; use it.
		if raw, ok, err := c.fromCache(key); err != nil || ok {
			return raw, err
		}
		// The shared fetch must not die with the first caller's request; the HTTP timeout bounds it.
		raw, err := c.fetch(context.WithoutCancel(ctx), query, vars)
		if errors.Is(err, ErrNotFound) {
			c.misses.set(key, nil, config.AniListNotFoundTTL)
		}
		if err != nil {
			return nil, err
		}
		c.cache.set(key, raw, ttl)
		return raw, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(json.RawMessage), nil
}

// fromCache returns a cached payload (ok) or ErrNotFound for a recently missed key; neither means fetch.
func (c *Client) fromCache(key string) (raw json.RawMessage, ok bool, err error) {
	if _, missed := c.misses.get(key); missed {
		return nil, true, ErrNotFound
	}
	raw, ok = c.cache.get(key)
	return raw, ok, nil
}

func cacheKey(scope, query string, vars map[string]any) (string, error) {
	b, err := json.Marshal(vars) // map keys are sorted, so equal variables give equal keys
	if err != nil {
		return "", err
	}
	return scope + "\x00" + query + "\x00" + string(b), nil
}

func (c *Client) fetch(ctx context.Context, query string, vars map[string]any) (json.RawMessage, error) {
	if c.coolingDown() || !c.limiter.take() {
		return nil, ErrUnavailable
	}

	body, err := json.Marshal(map[string]any{"query": query, "variables": vars})
	if err != nil {
		return nil, ErrUnavailable
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, ErrUnavailable
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		// Transport errors embed the URL; keep only the sentinel so logs stay clean.
		return nil, ErrUnavailable
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusTooManyRequests:
		c.cooldown(retryAfter(resp.Header, c.now()))
		return nil, ErrUnavailable
	case resp.StatusCode == http.StatusNotFound:
		return nil, ErrNotFound
	case resp.StatusCode != http.StatusOK:
		return nil, ErrUnavailable
	}

	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, config.AniListMaxResponseBytes)).Decode(&envelope); err != nil {
		return nil, ErrUnavailable
	}
	if len(envelope.Errors) > 0 && (len(envelope.Data) == 0 || string(envelope.Data) == "null") {
		return nil, ErrUnavailable
	}
	return envelope.Data, nil
}

func (c *Client) coolingDown() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now().Before(c.cooldownUntil)
}

// cooldown only ever extends the pause, so a late short Retry-After cannot shorten an earlier one.
func (c *Client) cooldown(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if until := c.now().Add(d); until.After(c.cooldownUntil) {
		c.cooldownUntil = until
	}
}

// retryAfter reads Retry-After as seconds or an HTTP date, falling back to the default and clamping to the max.
func retryAfter(h http.Header, now time.Time) time.Duration {
	raw := strings.TrimSpace(h.Get("Retry-After"))
	var d time.Duration
	if secs, err := strconv.Atoi(raw); err == nil {
		d = time.Duration(secs) * time.Second
	} else if at, err := http.ParseTime(raw); err == nil {
		d = at.Sub(now)
	}
	if d <= 0 {
		d = config.AniListRetryAfterDefault
	}
	return min(d, config.AniListRetryAfterMax)
}

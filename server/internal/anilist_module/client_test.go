package anilist_module

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeClock is a settable clock for cache TTLs and cooldowns.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (f *fakeClock) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

func (f *fakeClock) Advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = f.now.Add(d)
}

// reply is one canned upstream response.
type reply struct {
	status     int
	body       string
	retryAfter string
}

// newStub starts a local server that answers every request with respond and counts hits.
func newStub(t *testing.T, respond func(n int64) reply) (*httptest.Server, *atomic.Int64) {
	t.Helper()
	var hits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		rep := respond(n)
		if rep.retryAfter != "" {
			w.Header().Set("Retry-After", rep.retryAfter)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(rep.status)
		_, _ = w.Write([]byte(rep.body))
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

func newTestClient(srv *httptest.Server, clock *fakeClock) *Client {
	return NewClient(Options{Endpoint: srv.URL, HTTP: srv.Client(), Now: clock.Now})
}

func okData(_ int64) reply { return reply{status: http.StatusOK, body: `{"data":{"ok":true}}`} }

func TestQueryCachesUntilTTLExpires(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1000, 0)}
	srv, hits := newStub(t, okData)
	c := newTestClient(srv, clock)

	for range 2 {
		raw, err := c.Query(context.Background(), "", "q", map[string]any{"id": 1}, time.Minute)
		require.NoError(t, err)
		assert.JSONEq(t, `{"ok":true}`, string(raw))
	}
	assert.EqualValues(t, 1, hits.Load(), "second call must come from cache")

	clock.Advance(time.Minute)
	_, err := c.Query(context.Background(), "", "q", map[string]any{"id": 1}, time.Minute)
	require.NoError(t, err)
	assert.EqualValues(t, 2, hits.Load(), "expired entry must refetch")
}

func TestQueryScopeSeparatesCacheEntries(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1000, 0)}
	srv, hits := newStub(t, okData)
	c := newTestClient(srv, clock)

	_, _ = c.Query(context.Background(), "sfw", "q", nil, time.Minute)
	_, _ = c.Query(context.Background(), "adult", "q", nil, time.Minute)
	_, _ = c.Query(context.Background(), "sfw", "q", nil, time.Minute)
	assert.EqualValues(t, 2, hits.Load(), "same query under another scope must not reuse the cached entry")
}

func TestNotFoundIsCachedBrieflyThenRefetched(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1000, 0)}
	srv, hits := newStub(t, func(int64) reply { return reply{status: http.StatusNotFound, body: `{}`} })
	c := newTestClient(srv, clock)

	for range 3 {
		_, err := c.Query(context.Background(), "", "q", map[string]any{"id": 9}, time.Hour)
		require.ErrorIs(t, err, ErrNotFound)
	}
	assert.EqualValues(t, 1, hits.Load(), "repeated misses must be answered from the negative cache")

	clock.Advance(config.AniListNotFoundTTL - time.Second)
	_, _ = c.Query(context.Background(), "", "q", map[string]any{"id": 9}, time.Hour)
	assert.EqualValues(t, 1, hits.Load(), "miss stays cached until AniListNotFoundTTL")

	clock.Advance(time.Second)
	_, _ = c.Query(context.Background(), "", "q", map[string]any{"id": 9}, time.Hour)
	assert.EqualValues(t, 2, hits.Load(), "after AniListNotFoundTTL the title is fetched again")
}

func TestUnavailableIsNotNegativelyCached(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1000, 0)}
	srv, hits := newStub(t, func(int64) reply { return reply{status: http.StatusInternalServerError, body: `{}`} })
	c := newTestClient(srv, clock)

	for range 2 {
		_, err := c.Query(context.Background(), "", "q", nil, time.Hour)
		require.ErrorIs(t, err, ErrUnavailable)
	}
	assert.EqualValues(t, 2, hits.Load(), "only ErrNotFound is remembered, upstream failures retry")
}

func TestQueryKeysByVariables(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1000, 0)}
	srv, hits := newStub(t, okData)
	c := newTestClient(srv, clock)

	_, _ = c.Query(context.Background(), "", "q", map[string]any{"id": 1}, time.Minute)
	_, _ = c.Query(context.Background(), "", "q", map[string]any{"id": 2}, time.Minute)
	assert.EqualValues(t, 2, hits.Load(), "different variables are different cache keys")
}

func TestConcurrentIdenticalQueriesShareOneRequest(t *testing.T) {
	release := make(chan struct{})
	srv, hits := newStub(t, func(int64) reply {
		<-release
		return reply{status: http.StatusOK, body: `{"data":{"ok":true}}`}
	})
	c := newTestClient(srv, &fakeClock{now: time.Unix(1000, 0)})

	const callers = 5
	var wg sync.WaitGroup
	errs := make(chan error, callers)
	for range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := c.Query(context.Background(), "", "q", map[string]any{"id": 9}, time.Minute)
			errs <- err
		}()
	}
	close(release)
	wg.Wait()
	close(errs)
	for err := range errs {
		assert.NoError(t, err)
	}
	assert.EqualValues(t, 1, hits.Load())
}

func TestRateLimitedResponseStartsCooldownWithoutHTTP(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1000, 0)}
	srv, hits := newStub(t, func(n int64) reply {
		if n == 1 {
			return reply{status: http.StatusTooManyRequests, body: `{}`, retryAfter: "30"}
		}
		return okData(n)
	})
	c := newTestClient(srv, clock)

	_, err := c.Query(context.Background(), "", "q", nil, time.Minute)
	assert.ErrorIs(t, err, ErrUnavailable)

	_, err = c.Query(context.Background(), "", "other", nil, time.Minute)
	assert.ErrorIs(t, err, ErrUnavailable)
	assert.EqualValues(t, 1, hits.Load(), "during cooldown no request leaves the process")

	clock.Advance(31 * time.Second)
	_, err = c.Query(context.Background(), "", "other", nil, time.Minute)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, hits.Load())
}

func TestRetryAfterParsing(t *testing.T) {
	now := time.Unix(1000, 0)
	cases := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{"missing uses default", "", config.AniListRetryAfterDefault},
		{"garbage uses default", "soon", config.AniListRetryAfterDefault},
		{"zero uses default", "0", config.AniListRetryAfterDefault},
		{"seconds", "45", 45 * time.Second},
		{"huge is clamped", "86400", config.AniListRetryAfterMax},
		{"http date", now.Add(90 * time.Second).UTC().Format(http.TimeFormat), 90 * time.Second},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := http.Header{}
			if tc.value != "" {
				h.Set("Retry-After", tc.value)
			}
			assert.Equal(t, tc.want, retryAfter(h, now))
		})
	}
}

func TestLocalBudgetStopsAfterOneMinuteOfBurst(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1000, 0)}
	srv, hits := newStub(t, okData)
	c := newTestClient(srv, clock)

	for i := range config.AniListRequestsPerMinute {
		_, err := c.Query(context.Background(), "", "q", map[string]any{"id": i}, time.Minute)
		require.NoError(t, err)
	}
	_, err := c.Query(context.Background(), "", "q", map[string]any{"id": -1}, time.Minute)
	assert.ErrorIs(t, err, ErrUnavailable)
	assert.EqualValues(t, config.AniListRequestsPerMinute, hits.Load())
}

func TestTokenBucketRefillsOverTime(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1000, 0)}
	b := newTokenBucket(2, 60, clock.Now) // two tokens, one per second
	assert.True(t, b.take())
	assert.True(t, b.take())
	assert.False(t, b.take())
	clock.Advance(time.Second)
	assert.True(t, b.take())
	assert.False(t, b.take())
}

func TestStatusAndPayloadMapping(t *testing.T) {
	cases := []struct {
		name    string
		rep     reply
		wantErr error
	}{
		{"404 is not found", reply{status: http.StatusNotFound, body: `{}`}, ErrNotFound},
		{"500 is unavailable", reply{status: http.StatusInternalServerError, body: `{}`}, ErrUnavailable},
		{"errors without data", reply{status: http.StatusOK, body: `{"errors":[{"message":"bad"}]}`}, ErrUnavailable},
		{"malformed json", reply{status: http.StatusOK, body: `{"data":`}, ErrUnavailable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, _ := newStub(t, func(int64) reply { return tc.rep })
			c := newTestClient(srv, &fakeClock{now: time.Unix(1000, 0)})
			_, err := c.Query(context.Background(), "", "q", nil, time.Minute)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestTransportErrorIsUnavailableWithoutURL(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	endpoint := srv.URL
	srv.Close() // connection refused

	c := NewClient(Options{Endpoint: endpoint, HTTP: &http.Client{Timeout: time.Second}})
	_, err := c.Query(context.Background(), "", "q", nil, time.Minute)
	require.ErrorIs(t, err, ErrUnavailable)
	assert.NotContains(t, fmt.Sprint(err), endpoint)
}

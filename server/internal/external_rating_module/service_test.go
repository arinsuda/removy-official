package external_rating_module

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memStore struct {
	mu   sync.Mutex
	rows map[string]*ExternalRating
}

func newMemStore() *memStore { return &memStore{rows: map[string]*ExternalRating{}} }

func key(t string, id int) string { return t + ":" + strconv.Itoa(id) }

func (m *memStore) Find(_ context.Context, t string, id int) (*ExternalRating, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r, ok := m.rows[key(t, id)]; ok {
		c := *r
		return &c, nil
	}
	return nil, nil
}

func (m *memStore) Save(_ context.Context, r *ExternalRating) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c := *r
	m.rows[key(r.MediaType, r.TMDBID)] = &c
	return nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

const okBody = `{"Response":"True","imdbRating":"8.1","imdbVotes":"1,234,567","Ratings":[{"Source":"Internet Movie Database","Value":"8.1/10"},{"Source":"Rotten Tomatoes","Value":"94%"}]}`

// newTestService stubs the OMDb HTTP client; no real network is used.
func newTestService(store Store, calls *int32, handler func() (string, error)) *Service {
	s := NewService(store, "test-key", func(context.Context, string, int) (string, error) { return "tt0000001", nil })
	s.client.http = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		atomic.AddInt32(calls, 1)
		body, err := handler()
		if err != nil {
			return nil, err
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}}, nil
	})}
	return s
}

func TestParseOMDb(t *testing.T) {
	res, err := parseOMDb([]byte(okBody))
	require.NoError(t, err)
	assert.InDelta(t, 8.1, *res.IMDbRating, 1e-9)
	assert.Equal(t, 1234567, *res.IMDbVotes)
	assert.Equal(t, 94, *res.RTScore)

	res, err = parseOMDb([]byte(`{"Response":"True","imdbRating":"N/A","imdbVotes":"N/A","Ratings":[{"Source":"Rotten Tomatoes","Value":"N/A"}]}`))
	require.NoError(t, err)
	assert.Nil(t, res.IMDbRating)
	assert.Nil(t, res.IMDbVotes)
	assert.Nil(t, res.RTScore)

	_, err = parseOMDb([]byte(`{"Response":"False","Error":"Incorrect IMDb ID."}`))
	assert.ErrorIs(t, err, errOMDbNotFound)

	_, err = parseOMDb([]byte(`not json`))
	assert.Error(t, err)
}

func TestDisabledWithoutKey(t *testing.T) {
	s := NewService(newMemStore(), "", nil)
	assert.False(t, s.Enabled())
	assert.Equal(t, Ratings{}, s.Get(context.Background(), "movie", 1, "tt1"))
}

func TestMissFetchesThenFreshCacheIsUsed(t *testing.T) {
	var calls int32
	s := newTestService(newMemStore(), &calls, func() (string, error) { return okBody, nil })

	r := s.Get(context.Background(), "movie", 7, "tt7")
	require.NotNil(t, r.IMDbRating)
	assert.Equal(t, 94, *r.RTScore)
	assert.Equal(t, 1234567, r.IMDbVotes)

	s.Get(context.Background(), "movie", 7, "tt7")
	assert.EqualValues(t, 1, atomic.LoadInt32(&calls), "second call must hit the cache")
}

func TestStaleCacheRefreshesAndFallsBackOnError(t *testing.T) {
	store := newMemStore()
	var calls int32
	fail := false
	s := newTestService(store, &calls, func() (string, error) {
		if fail {
			return "", errors.New("boom")
		}
		return okBody, nil
	})
	base := time.Now()
	s.now = func() time.Time { return base }
	s.Get(context.Background(), "movie", 7, "tt7")

	// 6 days: still fresh. 8 days: stale, refresh fails, stale value is returned.
	s.now = func() time.Time { return base.Add(6 * 24 * time.Hour) }
	s.Get(context.Background(), "movie", 7, "tt7")
	assert.EqualValues(t, 1, atomic.LoadInt32(&calls))

	fail = true
	s.now = func() time.Time { return base.Add(8 * 24 * time.Hour) }
	r := s.Get(context.Background(), "movie", 7, "tt7")
	assert.EqualValues(t, 2, atomic.LoadInt32(&calls))
	require.NotNil(t, r.IMDbRating, "stale cache served on upstream error")
}

func TestNotFoundUsesShortTTL(t *testing.T) {
	var calls int32
	s := newTestService(newMemStore(), &calls, func() (string, error) { return `{"Response":"False"}`, nil })
	base := time.Now()
	s.now = func() time.Time { return base }

	assert.Equal(t, Ratings{}, s.Get(context.Background(), "movie", 9, "tt9"))
	s.now = func() time.Time { return base.Add(23 * time.Hour) }
	s.Get(context.Background(), "movie", 9, "tt9")
	assert.EqualValues(t, 1, atomic.LoadInt32(&calls), "not_found still fresh at 23h")

	s.now = func() time.Time { return base.Add(25 * time.Hour) }
	s.Get(context.Background(), "movie", 9, "tt9")
	assert.EqualValues(t, 2, atomic.LoadInt32(&calls), "not_found expires after 1 day")
}

func TestBudgetExhaustionReturnsStaleOrUnavailable(t *testing.T) {
	store := newMemStore()
	var calls int32
	s := newTestService(store, &calls, func() (string, error) { return okBody, nil })
	s.budget.limit = 1
	base := time.Now()
	s.now = func() time.Time { return base }

	require.NotNil(t, s.Get(context.Background(), "movie", 1, "tt1").IMDbRating)
	// Budget used: a new title is unavailable, no HTTP call.
	assert.Equal(t, Ratings{}, s.Get(context.Background(), "movie", 2, "tt2"))
	assert.EqualValues(t, 1, atomic.LoadInt32(&calls))

	// A stale title still returns its stale value.
	s.now = func() time.Time { return base.Add(10 * 24 * time.Hour) }
	s.budget.day = s.now().UTC().Format("2006-01-02")
	s.budget.used = 1
	require.NotNil(t, s.Get(context.Background(), "movie", 1, "tt1").IMDbRating)
	assert.EqualValues(t, 1, atomic.LoadInt32(&calls))

	// A new UTC day resets the budget.
	s.now = func() time.Time { return base.Add(48 * time.Hour) }
	s.Get(context.Background(), "movie", 2, "tt2")
	assert.EqualValues(t, 2, atomic.LoadInt32(&calls))
}

func TestTimeoutFallsBackQuickly(t *testing.T) {
	store := newMemStore()
	var calls int32
	release := make(chan struct{})
	s := newTestService(store, &calls, func() (string, error) { <-release; return okBody, nil })
	defer close(release)
	s.timeout = 50 * time.Millisecond

	start := time.Now()
	r := s.Get(context.Background(), "movie", 3, "tt3")
	assert.Less(t, time.Since(start), time.Second)
	assert.Equal(t, Ratings{}, r)
}

func TestTVResolvesIMDbIDWhenUnknown(t *testing.T) {
	var calls int32
	var resolved int32
	s := newTestService(newMemStore(), &calls, func() (string, error) { return okBody, nil })
	s.resolver = func(context.Context, string, int) (string, error) {
		atomic.AddInt32(&resolved, 1)
		return "tt4", nil
	}
	r := s.Get(context.Background(), "tv", 4, "")
	require.NotNil(t, r.IMDbRating)
	assert.EqualValues(t, 1, resolved)

	// TMDB has no IMDb id -> not_found without calling OMDb.
	s.resolver = func(context.Context, string, int) (string, error) { return "", nil }
	assert.Equal(t, Ratings{}, s.Get(context.Background(), "tv", 5, ""))
	assert.EqualValues(t, 1, atomic.LoadInt32(&calls))
}

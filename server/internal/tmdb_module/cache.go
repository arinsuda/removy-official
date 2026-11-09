package tmdbmodule

import (
	"sync"
	"time"
)

const (
	detailCacheTTL = 10 * time.Minute
	detailCacheMax = 2000
)

// detailKey includes the resolved TMDB language so one caller's language never leaks to another.
type detailKey struct {
	kind     string
	id       int
	language string
}

type detailEntry struct {
	value   any
	expires time.Time
}

// detailCache is a small bounded in-process TTL cache for TMDB detail lookups.
type detailCache struct {
	mu      sync.Mutex
	ttl     time.Duration
	max     int
	now     func() time.Time
	entries map[detailKey]detailEntry
}

func newDetailCache(ttl time.Duration, max int) *detailCache {
	return &detailCache{ttl: ttl, max: max, now: time.Now, entries: make(map[detailKey]detailEntry)}
}

func (d *detailCache) get(k detailKey) (any, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	e, ok := d.entries[k]
	if !ok {
		return nil, false
	}
	if !d.now().Before(e.expires) {
		delete(d.entries, k)
		return nil, false
	}
	return e.value, true
}

func (d *detailCache) set(k detailKey, v any) {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := d.now()
	if _, exists := d.entries[k]; !exists && len(d.entries) >= d.max {
		for key, e := range d.entries {
			if !now.Before(e.expires) {
				delete(d.entries, key)
			}
		}
		// Still full: evict an arbitrary entry to stay bounded.
		for key := range d.entries {
			if len(d.entries) < d.max {
				break
			}
			delete(d.entries, key)
		}
	}
	d.entries[k] = detailEntry{value: v, expires: now.Add(d.ttl)}
}

const (
	detailKindMovie = "movie"
	detailKindTV    = "tv"
)

var details = newDetailCache(detailCacheTTL, detailCacheMax)

// currentLanguage is the TMDB language detail requests resolve to (the client default).
func currentLanguage() string {
	if c == nil {
		return defaultLanguage
	}
	return c.language
}

package anilist_module

import (
	"encoding/json"
	"sync"
	"time"
)

// ttlCache is a small bounded in-process TTL cache of raw GraphQL "data" objects.
type ttlCache struct {
	mu      sync.Mutex
	max     int
	now     func() time.Time
	entries map[string]cacheEntry
}

type cacheEntry struct {
	value   json.RawMessage
	expires time.Time
}

func newTTLCache(max int, now func() time.Time) *ttlCache {
	return &ttlCache{max: max, now: now, entries: make(map[string]cacheEntry)}
}

func (c *ttlCache) get(key string) (json.RawMessage, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if !c.now().Before(e.expires) {
		delete(c.entries, key)
		return nil, false
	}
	return e.value, true
}

func (c *ttlCache) set(key string, value json.RawMessage, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	if _, exists := c.entries[key]; !exists && len(c.entries) >= c.max {
		for k, e := range c.entries {
			if !now.Before(e.expires) {
				delete(c.entries, k)
			}
		}
		// Still full: evict arbitrary entries to stay bounded.
		for k := range c.entries {
			if len(c.entries) < c.max {
				break
			}
			delete(c.entries, k)
		}
	}
	c.entries[key] = cacheEntry{value: value, expires: now.Add(ttl)}
}

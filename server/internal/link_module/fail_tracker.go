package link_module

import (
	"sync"
	"time"
)

const (
	verifyFailMax    = 30
	verifyFailWindow = 15 * time.Minute
	verifyFailKeyCap = 10000
)

// failTracker counts failed attempts per key in a fixed window (in-memory, per process).
// The zero value is ready to use.
type failTracker struct {
	mu      sync.Mutex
	entries map[string]failEntry
}

type failEntry struct {
	count   int
	resetAt time.Time
}

// allow reports whether key is still under the failure budget.
func (t *failTracker) allow(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	e, ok := t.entries[key]
	if !ok || time.Now().After(e.resetAt) {
		return true
	}
	return e.count < verifyFailMax
}

// fail records one failed attempt for key.
func (t *failTracker) fail(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	if t.entries == nil {
		t.entries = make(map[string]failEntry)
	}
	if len(t.entries) >= verifyFailKeyCap {
		for k, e := range t.entries {
			if now.After(e.resetAt) {
				delete(t.entries, k)
			}
		}
	}
	e := t.entries[key]
	if now.After(e.resetAt) {
		e = failEntry{resetAt: now.Add(verifyFailWindow)}
	}
	e.count++
	t.entries[key] = e
}

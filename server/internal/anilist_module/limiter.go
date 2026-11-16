package anilist_module

import (
	"sync"
	"time"
)

// tokenBucket is a local request budget: capacity tokens, refilled continuously.
type tokenBucket struct {
	mu           sync.Mutex
	capacity     float64
	tokens       float64
	refillPerSec float64
	last         time.Time
	now          func() time.Time
}

func newTokenBucket(capacity int, perMinute int, now func() time.Time) *tokenBucket {
	return &tokenBucket{
		capacity:     float64(capacity),
		tokens:       float64(capacity),
		refillPerSec: float64(perMinute) / 60,
		last:         now(),
		now:          now,
	}
}

// take spends one token and reports whether one was available. It never waits.
func (b *tokenBucket) take() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	t := b.now()
	b.tokens = min(b.capacity, b.tokens+t.Sub(b.last).Seconds()*b.refillPerSec)
	b.last = t
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

package media_stats_module

import (
	"context"
	"log"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"gorm.io/gorm"
)

const (
	// MediaViewRetention is how long per-day view rows are kept (3 months). Voting reads one
	// lookback month before a round opens, so this leaves room.
	MediaViewRetention = 90 * 24 * time.Hour
	// viewPurgeInterval is the purge worker tick.
	viewPurgeInterval = time.Hour
	// viewPurgeBatch caps rows deleted per statement so one pass never holds a long lock.
	viewPurgeBatch = 10000
)

// ViewPurger deletes media_view_days rows older than MediaViewRetention.
type ViewPurger struct {
	db  *gorm.DB
	now func() time.Time
}

func NewViewPurger(db *gorm.DB) *ViewPurger {
	return &ViewPurger{db: db, now: time.Now}
}

// Run purges every viewPurgeInterval until ctx is cancelled.
func (p *ViewPurger) Run(ctx context.Context) {
	ticker := time.NewTicker(viewPurgeInterval)
	defer ticker.Stop()
	for {
		p.tick(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// tick runs one pass and never lets a panic kill the worker goroutine.
func (p *ViewPurger) tick(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("view purge: recovered from panic: %v", r)
		}
	}()
	deleted, err := p.RunOnce(ctx)
	if err != nil {
		log.Printf("view purge: %v", err)
		return
	}
	if deleted > 0 {
		log.Printf("view purge: deleted=%d", deleted)
	}
}

// RunOnce deletes expired rows in batches until none are left or ctx ends. A row is expired when
// its Bangkok day is before the Bangkok day MediaViewRetention ago.
func (p *ViewPurger) RunOnce(ctx context.Context) (int64, error) {
	cutoff := config.LocalDay(p.now().Add(-MediaViewRetention))
	var total int64
	for {
		if err := ctx.Err(); err != nil {
			return total, err
		}
		res := p.db.WithContext(ctx).Exec(`DELETE FROM media_view_days WHERE ctid IN (
			SELECT ctid FROM media_view_days WHERE day < ? LIMIT ?)`, cutoff, viewPurgeBatch)
		if res.Error != nil {
			return total, res.Error
		}
		total += res.RowsAffected
		if res.RowsAffected < viewPurgeBatch {
			return total, nil
		}
	}
}

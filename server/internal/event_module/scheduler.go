package event_module

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/arinsuda/movie-hub/internal/admin_module"
	"github.com/arinsuda/movie-hub/internal/shared"
	"gorm.io/gorm"
)

// RoundScheduler creates the next monthly and yearly voting rounds as drafts shortly before they open
// (autoGenerateLead), and publishes the drafts it created (auto_generated) just before they open
// (autoPublishLead). Admins opt out of a round by cancelling or deleting its draft; their own drafts
// are never published. Every step is idempotent: the unique (type, media_type, period) index settles
// concurrent creators and FOR UPDATE settles concurrent publishers, so several instances are safe.
type RoundScheduler struct {
	svc     *Service
	suggest SuggestFunc
}

// NewRoundScheduler needs no image store: it never touches banners.
func NewRoundScheduler(db *gorm.DB) *RoundScheduler {
	svc := NewService(db, nil)
	return &RoundScheduler{svc: svc, suggest: svc.SuggestNominees}
}

// NewRoundSchedulerWith injects the clock and the nominee source, so tests need neither real time nor TMDB.
func NewRoundSchedulerWith(db *gorm.DB, now func() time.Time, suggest SuggestFunc) *RoundScheduler {
	svc := NewService(db, nil)
	svc.now = now
	return &RoundScheduler{svc: svc, suggest: suggest}
}

// Run looks for work every autoRoundInterval until ctx is cancelled.
func (r *RoundScheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(autoRoundInterval)
	defer ticker.Stop()
	for {
		r.tick(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// tick runs one pass and never lets a panic kill the worker goroutine.
func (r *RoundScheduler) tick(ctx context.Context) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("round scheduler: recovered from panic: %v", rec)
		}
	}()
	generated, published, err := r.RunOnce(ctx)
	if err != nil {
		log.Printf("round scheduler: %v", err)
	}
	if generated > 0 || published > 0 {
		log.Printf("round scheduler: generated=%d published=%d", generated, published)
	}
}

// RunOnce creates the due drafts, then publishes the due auto drafts. Failures are collected and do
// not stop the pass.
func (r *RoundScheduler) RunOnce(ctx context.Context) (generated, published int, err error) {
	var errs []error
	now := r.svc.now().UTC()
	for _, mt := range []shared.MediaType{shared.MediaTypeMovie, shared.MediaTypeTV} {
		mediaType := string(mt)
		for _, kind := range []EventType{TypeVoteMonthly, TypeVoteYearly} {
			if err := ctx.Err(); err != nil {
				return generated, published, errors.Join(append(errs, err)...)
			}
			created, err := r.generateDue(ctx, now, kind, mediaType)
			if err != nil {
				errs = append(errs, fmt.Errorf("generate %s %s: %w", kind, mediaType, err))
			}
			if created {
				generated++
			}
		}
	}

	var ids []uint
	if err := r.svc.db.WithContext(ctx).Model(&Event{}).
		Where("auto_generated = ? AND status = ? AND type IN ? AND starts_at <= ? AND ends_at > ?",
			true, StatusDraft, votingTypes(), now.Add(autoPublishLead), now).
		Order("starts_at ASC, id ASC").Limit(autoRoundBatch).
		Pluck("id", &ids).Error; err != nil {
		return generated, published, errors.Join(append(errs, err)...)
	}
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return generated, published, errors.Join(append(errs, err)...)
		}
		ok, err := r.svc.autoPublishTx(ctx, id)
		if err != nil {
			errs = append(errs, fmt.Errorf("publish event %d: %w", id, err))
		}
		if ok {
			published++
		}
	}
	return generated, published, errors.Join(errs...)
}

// generateDue creates the next round of `kind` once it is within autoGenerateLead of opening.
// An existing round, pending monthly rounds or too few TMDB candidates mean "nothing to do now".
func (r *RoundScheduler) generateDue(ctx context.Context, now time.Time, kind EventType, mediaType string) (bool, error) {
	var (
		period   string
		startsAt time.Time
		err      error
	)
	if kind == TypeVoteMonthly {
		period = NextMonthlyPeriod(now)
		startsAt, _, err = MonthlyWindow(period)
	} else {
		period = NextYearlyPeriod(now)
		startsAt, _, err = YearlyWindow(period)
	}
	if err != nil {
		return false, err
	}
	if now.Before(startsAt.Add(-autoGenerateLead)) {
		return false, nil
	}

	if kind == TypeVoteMonthly {
		_, err = r.svc.generateMonthly(ctx, 0, mediaType, period, true, r.suggest)
	} else {
		_, err = r.svc.generateYearly(ctx, 0, mediaType, period, true)
	}
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, ErrNotEnoughNominees):
		if !now.Before(startsAt.Add(-autoRoundInterval)) && now.Before(startsAt.Add(autoRoundInterval)) {
			log.Printf("round scheduler: %s %s round %s not generated: fewer than %d suggested nominees",
				kind, mediaType, period, nomineesPerRound)
		}
		return false, nil
	case errors.Is(err, ErrEventExists), errors.Is(err, ErrMonthlyRoundsPending):
		return false, nil
	}
	return false, err
}

type publishOutcome int

const (
	publishSkipped publishOutcome = iota // not due, not an auto draft, or changed since the query
	publishDone
	publishTooFewNominees
)

// autoPublishTx publishes one auto-generated draft that is due. The row is locked first and the clock
// is read after the lock, so an admin edit, cancel or delete that landed first wins.
func (s *Service) autoPublishTx(ctx context.Context, id uint) (bool, error) {
	outcome := publishSkipped
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ev Event
		if err := lockEvent(tx, id, &ev); err != nil {
			return err
		}
		now := s.now().UTC()
		if !ev.AutoGenerated || ev.Status != StatusDraft || !ev.Type.IsVoting() ||
			now.Before(ev.StartsAt.Add(-autoPublishLead)) || !now.Before(ev.EndsAt) {
			return nil
		}
		var nominees int64
		if err := tx.Model(&EventNominee{}).Where("event_id = ?", id).Count(&nominees).Error; err != nil {
			return err
		}
		if nominees < minNomineesToPublish {
			if now.Before(ev.StartsAt.Add(autoRoundInterval)) {
				outcome = publishTooFewNominees
			}
			return nil
		}
		if err := tx.Model(&ev).Updates(map[string]any{"status": StatusPublished, "updated_at": now}).Error; err != nil {
			return err
		}
		outcome = publishDone
		return writeAudit(tx, 0, admin_module.ActionEventAutoPublished, id, map[string]any{"nominees": nominees})
	})
	switch {
	case errors.Is(err, ErrNotFound):
		return false, nil // deleted since the query
	case err != nil:
		return false, err
	}
	if outcome == publishTooFewNominees {
		log.Printf("round scheduler: event %d stays draft: fewer than %d nominees", id, minNomineesToPublish)
	}
	return outcome == publishDone, nil
}

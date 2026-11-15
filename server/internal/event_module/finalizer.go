package event_module

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/arinsuda/movie-hub/monitoring"
	"gorm.io/gorm"
)

const (
	// finalizeInterval is how often the worker looks for ended rounds nobody has opened.
	finalizeInterval = 5 * time.Minute
	// finalizeBatch caps rounds settled per tick.
	finalizeBatch = 50
)

// Finalizer settles ended voting rounds in the background, so a round closes even when no one
// opens its page. Each round is settled by finalizeTx in its own transaction, which locks the row
// with FOR UPDATE, so several instances can run the worker safely.
type Finalizer struct {
	svc    *Service
	report *monitoring.JobReporter
}

// NewFinalizer settles rounds only, so it needs no image store.
func NewFinalizer(db *gorm.DB, emitter NotificationEmitter) *Finalizer {
	svc := NewService(db, nil)
	svc.SetEmitter(emitter)
	return &Finalizer{svc: svc, report: monitoring.NewJobReporter(monitoring.JobRoundFinalizer)}
}

// Run settles rounds every finalizeInterval until ctx is cancelled.
func (f *Finalizer) Run(ctx context.Context) {
	ticker := time.NewTicker(finalizeInterval)
	defer ticker.Stop()
	for {
		f.tick(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// tick runs one pass and never lets a panic kill the worker goroutine.
func (f *Finalizer) tick(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("round finalize: recovered from panic: %v", r)
			f.report.Error(monitoring.OpPanic, fmt.Errorf("panic: %v", r))
		}
	}()
	settled, err := f.RunOnce(ctx)
	if err != nil {
		log.Printf("round finalize: %v", err)
		f.report.Error(monitoring.OpRun, err)
	}
	if settled > 0 {
		log.Printf("round finalize: settled=%d", settled)
	}
}

// RunOnce settles up to finalizeBatch published, unfinalized voting rounds whose window has ended.
// A round that changed since the query (not ended, unpublished) is skipped; the next tick sees its
// current state. Other failures are collected and do not stop the batch.
func (f *Finalizer) RunOnce(ctx context.Context) (int, error) {
	var ids []uint
	err := f.svc.db.WithContext(ctx).Model(&Event{}).
		Scopes(publishedScope).
		Where("type IN ? AND is_finalized = ? AND ends_at <= ?", votingTypes(), false, f.svc.now().UTC()).
		Order("ends_at ASC, id ASC").Limit(finalizeBatch).
		Pluck("id", &ids).Error
	if err != nil {
		return 0, err
	}

	settled := 0
	var errs []error
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return settled, errors.Join(append(errs, err)...)
		}
		err := f.svc.finalizeTx(ctx, 0, id)
		switch {
		case err == nil:
			settled++
		case errors.Is(err, ErrVotingNotEnded), errors.Is(err, ErrEventNotPublished),
			errors.Is(err, ErrInvalidEventType), errors.Is(err, ErrNotFound):
			// Changed since the query; the next tick decides again.
		default:
			errs = append(errs, fmt.Errorf("event %d: %w", id, err))
		}
	}
	return settled, errors.Join(errs...)
}

// votingTypes lists the event types that settle into awards.
func votingTypes() []string {
	return []string{string(TypeVoteMonthly), string(TypeVoteYearly)}
}

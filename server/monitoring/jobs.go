package monitoring

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
)

const (
	// JobReportInterval is the most often one job+op may send a Sentry event; repeats inside it are
	// counted and the next event carries the count.
	JobReportInterval = time.Hour

	// Background job names (Sentry tag "job").
	JobRoundScheduler = "round_scheduler"
	JobRoundFinalizer = "round_finalizer"
	JobAccountPurger  = "account_purger"
	JobViewPurger     = "view_purger"

	// Operation names (Sentry tag "op"), shared so the dedupe key and fingerprint stay stable.
	OpPanic            = "panic"
	OpRun              = "run"
	OpTooFewNominees   = "too_few_nominees"
	OpLateRoundDrafted = "late_round_drafted"
	OpUserErase        = "user_erase"
	OpRecordCleanup    = "record_cleanup"
	OpStorageCleanup   = "storage_cleanup"

	tagJob           = "job"
	tagOp            = "op"
	tagSuppressed    = "suppressed"
	suppressedPrefix = " (suppressed repeats: "
	suppressedSuffix = ")"
)

// jobCapture sends one report; tests replace it.
type jobCapture func(job, op string, level sentry.Level, err error, suppressed int)

// JobReporter reports background-job failures to Sentry, at most one event per op per JobReportInterval.
// Reported errors must carry no user ids, storage keys or emails; keep that detail in logs.
// It is a no-op while Sentry is off.
type JobReporter struct {
	job     string
	capture jobCapture
	enabled func() bool
	now     func() time.Time

	mu   sync.Mutex
	seen map[string]*jobState
}

type jobState struct {
	last       time.Time
	suppressed int
}

// NewJobReporter returns a reporter for one background job.
func NewJobReporter(job string) *JobReporter {
	return &JobReporter{
		job:     job,
		capture: captureJob,
		enabled: func() bool { return sentry.CurrentHub().Client() != nil },
		now:     time.Now,
		seen:    map[string]*jobState{},
	}
}

// Error reports a failed operation. nil and context.Canceled (shutdown) are ignored.
func (r *JobReporter) Error(op string, err error) {
	if err == nil || errors.Is(err, context.Canceled) {
		return
	}
	r.report(op, sentry.LevelError, err)
}

// Warn reports a notable condition that is not a failure.
func (r *JobReporter) Warn(op string) {
	if r == nil {
		return
	}
	r.report(op, sentry.LevelWarning, errors.New(r.job+": "+op))
}

func (r *JobReporter) report(op string, level sentry.Level, err error) {
	if r == nil || !r.enabled() {
		return
	}
	now := r.now()
	r.mu.Lock()
	st, ok := r.seen[op]
	if !ok {
		st = &jobState{}
		r.seen[op] = st
	}
	if ok && now.Sub(st.last) < JobReportInterval {
		st.suppressed++
		r.mu.Unlock()
		return
	}
	suppressed := st.suppressed
	st.last, st.suppressed = now, 0
	r.mu.Unlock()
	r.capture(r.job, op, level, err, suppressed)
}

func captureJob(job, op string, level sentry.Level, err error, suppressed int) {
	// Workers run in separate goroutines; a cloned hub keeps one job's scope off another's event.
	hub := sentry.CurrentHub().Clone()
	hub.WithScope(func(s *sentry.Scope) {
		s.SetTag(tagJob, job)
		s.SetTag(tagOp, op)
		s.SetLevel(level)
		s.SetFingerprint([]string{job, op})
		if suppressed > 0 {
			s.SetTag(tagSuppressed, strconv.Itoa(suppressed))
			err = errors.New(err.Error() + suppressedPrefix + strconv.Itoa(suppressed) + suppressedSuffix)
		}
		hub.CaptureException(err)
	})
}

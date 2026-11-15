package monitoring

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
)

type sent struct {
	job, op    string
	suppressed int
}

func testReporter(enabled bool, at *time.Time, out *[]sent) *JobReporter {
	r := NewJobReporter(JobRoundScheduler)
	r.enabled = func() bool { return enabled }
	r.now = func() time.Time { return *at }
	r.capture = func(job, op string, _ sentry.Level, _ error, suppressed int) {
		*out = append(*out, sent{job, op, suppressed})
	}
	return r
}

func TestJobReporterDedupeAndSuppressedCount(t *testing.T) {
	at := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)
	var out []sent
	r := testReporter(true, &at, &out)
	boom := errors.New("boom")

	r.Error(OpRun, boom)
	r.Error(OpRun, boom)
	at = at.Add(JobReportInterval - time.Second)
	r.Error(OpRun, boom)
	r.Warn(OpLateRoundDrafted) // different op: own budget
	if len(out) != 2 || out[0].suppressed != 0 {
		t.Fatalf("want 2 events, first clean; got %+v", out)
	}
	at = at.Add(time.Second)
	r.Error(OpRun, boom)
	if len(out) != 3 || out[2].op != OpRun || out[2].suppressed != 2 {
		t.Fatalf("want third event with 2 suppressed; got %+v", out)
	}
	at = at.Add(JobReportInterval)
	r.Error(OpRun, boom)
	if out[3].suppressed != 0 {
		t.Fatalf("suppressed count must reset; got %+v", out[3])
	}
}

func TestJobReporterIgnoresCanceledNilAndOff(t *testing.T) {
	at := time.Now()
	var out []sent
	r := testReporter(true, &at, &out)
	r.Error(OpRun, nil)
	r.Error(OpRun, context.Canceled)
	r.Error(OpRun, fmt.Errorf("wrapped: %w", context.Canceled))
	if len(out) != 0 {
		t.Fatalf("nothing should be reported; got %+v", out)
	}
	off := testReporter(false, &at, &out)
	off.Error(OpRun, errors.New("boom"))
	off.Warn(OpRun)
	if len(out) != 0 {
		t.Fatalf("reporter must be silent when Sentry is off; got %+v", out)
	}
	var nilReporter *JobReporter
	nilReporter.Error(OpRun, errors.New("boom")) // must not panic
}

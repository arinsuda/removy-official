package event_module

import (
	"errors"
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/config"
)

func bangkok(y int, m time.Month, d, h, min int) time.Time {
	return time.Date(y, m, d, h, min, 0, 0, config.AppLocation)
}

func TestCurrentRoundMonthly(t *testing.T) {
	cases := []struct {
		name string
		at   time.Time
		want string
	}{
		{"opening instant", bangkok(2026, time.November, 15, 0, 0), "2026-11"},
		{"just before the cutoff", bangkok(2026, time.November, 14, 23, 59), "2026-10"},
		{"year boundary", bangkok(2027, time.January, 3, 12, 0), "2026-12"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			period, starts, ends, ok := currentRound(TypeVoteMonthly, tc.at)
			if !ok || period != tc.want {
				t.Fatalf("got %q ok=%v, want %q", period, ok, tc.want)
			}
			if tc.at.Before(starts) || !tc.at.Before(ends) {
				t.Errorf("window %v-%v does not contain %v", starts, ends, tc.at)
			}
		})
	}
}

func TestCurrentRoundYearly(t *testing.T) {
	cases := []struct {
		name   string
		at     time.Time
		period string
		ok     bool
	}{
		{"before the window", bangkok(2027, time.January, 15, 23, 59), "", false},
		{"opening instant", bangkok(2027, time.January, 16, 0, 0), "2026", true},
		{"last minute", bangkok(2027, time.February, 14, 23, 59), "2026", true},
		{"closing instant", bangkok(2027, time.February, 15, 0, 0), "", false},
		{"mid year", bangkok(2026, time.July, 1, 0, 0), "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			period, _, _, ok := currentRound(TypeVoteYearly, tc.at)
			if ok != tc.ok || period != tc.period {
				t.Fatalf("got %q ok=%v, want %q ok=%v", period, ok, tc.period, tc.ok)
			}
		})
	}
}

func TestLateStartWindow(t *testing.T) {
	ends := bangkok(2026, time.December, 15, 0, 0)
	now := bangkok(2026, time.December, 11, 22, 30).Add(45 * time.Second)

	starts, err := lateStartWindow(now, ends)
	if err != nil {
		t.Fatal(err)
	}
	if want := bangkok(2026, time.December, 11, 23, 30); !starts.Equal(want) {
		t.Errorf("starts = %v, want %v (minute truncated + lead)", starts, want)
	}

	// Exactly lateStartMinWindow left is allowed; one minute less is not.
	atLimit := ends.Add(-lateStartMinWindow - lateStartLead)
	if _, err := lateStartWindow(atLimit, ends); err != nil {
		t.Errorf("exact minimum window: unexpected %v", err)
	}
	if _, err := lateStartWindow(atLimit.Add(time.Minute), ends); !errors.Is(err, ErrVotingWindowTooShort) {
		t.Errorf("one minute short: want ErrVotingWindowTooShort, got %v", err)
	}
}

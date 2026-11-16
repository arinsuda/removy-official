package manga_module

import (
	"testing"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/stretchr/testify/assert"
)

func TestEligibleAtBoundary(t *testing.T) {
	now := time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		dob  *time.Time
		want bool
	}{
		{"no birth date", nil, false},
		{"born yesterday", ptrTime(now.AddDate(0, 0, -1)), false},
		{"17 years and 364 days", ptrTime(now.AddDate(-18, 0, 1)), false},
		{"turns 18 today", ptrTime(now.AddDate(-18, 0, 0)), true},
		{"18 years and a day", ptrTime(now.AddDate(-18, 0, -1)), true},
		{"well past 18", ptrTime(now.AddDate(-40, 0, 0)), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, eligibleAt(tc.dob, now))
		})
	}
}

// bangkok builds a midnight-based instant in config.AppLocation, the way the DOB column is read.
func bangkok(y int, m time.Month, d, hour, minute int) time.Time {
	return time.Date(y, m, d, hour, minute, 0, 0, config.AppLocation)
}

func TestEligibilityUnlocksOnTheBangkokBirthday(t *testing.T) {
	dob := ptrTime(bangkok(2008, time.November, 16, 0, 0))
	cases := []struct {
		name string
		now  time.Time
		want bool
	}{
		{"day before, last minute", bangkok(2026, time.November, 15, 23, 59), false},
		{"birthday, first minute", bangkok(2026, time.November, 16, 0, 0), true},
		{"birthday, late evening", bangkok(2026, time.November, 16, 23, 59), true},
		// 2026-11-15 18:00 UTC is already 16 November in Bangkok, so the unlock must not wait for UTC.
		{"UTC still on the day before", time.Date(2026, time.November, 15, 18, 0, 0, 0, time.UTC), true},
		{"UTC 17:00 is already 17 November in Bangkok", time.Date(2026, time.November, 16, 17, 0, 0, 0, time.UTC), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, eligibleAt(dob, tc.now))
		})
	}
}

func TestEligibilityForLeapDayBirth(t *testing.T) {
	dob := ptrTime(bangkok(2008, time.February, 29, 0, 0))
	assert.False(t, eligibleAt(dob, bangkok(2026, time.February, 28, 23, 59)), "the day before 1 March in a non-leap year")
	assert.True(t, eligibleAt(dob, bangkok(2026, time.March, 1, 0, 0)), "non-leap years use 1 March")
	assert.True(t, eligibleAt(dob, bangkok(2028, time.February, 29, 0, 0)), "leap years use 29 February")
}

func TestGuestGateNeverTouchesTheDatabase(t *testing.T) {
	ok, err := dobGate{}.Eligible(t.Context(), 0)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestScopeFollowsEligibility(t *testing.T) {
	assert.Equal(t, scopeAdult, scopeOf(true))
	assert.Equal(t, scopeSFW, scopeOf(false))
}

func ptrTime(t time.Time) *time.Time { return &t }

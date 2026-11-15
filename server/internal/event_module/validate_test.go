package event_module

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func validTestInput() EventInput {
	now := time.Now().UTC()
	return EventInput{
		TitleEN:       "Summer Movie Festival",
		TitleTH:       "เทศกาลภาพยนตร์ฤดูร้อน",
		BodyEN:        "Join our annual festival!",
		BodyTH:        "ร่วมสนุกกับเทศกาลประจำปี!",
		LinkURL:       "/movies",
		StartsAt:      now.Add(-time.Hour),
		EndsAt:        now.Add(24 * time.Hour),
		ShowPopup:     true,
		PopupPriority: 10,
		Status:        StatusPublished,
	}
}

func TestEventInputNormalize_Valid(t *testing.T) {
	in := validTestInput()
	if err := in.normalize(); err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
	if in.TitleEN != "Summer Movie Festival" {
		t.Errorf("unexpected TitleEN: %s", in.TitleEN)
	}
}

func TestEventInputNormalize_TitleValidation(t *testing.T) {
	// Empty English title
	in := validTestInput()
	in.TitleEN = "   "
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "title_en") {
		t.Errorf("expected title_en error, got: %v", err)
	}

	// Empty Thai title
	in = validTestInput()
	in.TitleTH = ""
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "title_th") {
		t.Errorf("expected title_th error, got: %v", err)
	}

	// Too long English title (>120 runes)
	in = validTestInput()
	in.TitleEN = strings.Repeat("A", 121)
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "title_en") {
		t.Errorf("expected title_en length error, got: %v", err)
	}

	// 120 Thai runes is fine
	in = validTestInput()
	in.TitleTH = strings.Repeat("ก", 120)
	if err := in.normalize(); err != nil {
		t.Errorf("expected 120 runes valid, got: %v", err)
	}

	// 121 Thai runes fails
	in.TitleTH = strings.Repeat("ก", 121)
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "title_th") {
		t.Errorf("expected title_th length error, got: %v", err)
	}
}

func TestEventInputNormalize_BodyValidation(t *testing.T) {
	in := validTestInput()
	in.BodyEN = strings.Repeat("B", 2001)
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "body_en") {
		t.Errorf("expected body_en error, got: %v", err)
	}

	in = validTestInput()
	in.BodyTH = strings.Repeat("ข", 2001)
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "body_th") {
		t.Errorf("expected body_th error, got: %v", err)
	}
}

func TestEventInputNormalize_LinkURL(t *testing.T) {
	cases := []struct {
		url   string
		valid bool
	}{
		{"", true},
		{"/events", true},
		{"/movies/123", true},
		{"https://example.com/details", true},
		{"http://example.com", true},
		{"//evil.com", false},
		{"javascript:alert(1)", false},
		{"ftp://files.example.com", false},
		{"/with space", false},
		{"https://user:pass@example.com", false},
	}

	for _, c := range cases {
		in := validTestInput()
		in.LinkURL = c.url
		err := in.normalize()
		if c.valid && err != nil {
			t.Errorf("expected url %q to be valid, got: %v", c.url, err)
		}
		if !c.valid && err == nil {
			t.Errorf("expected url %q to be invalid, but passed", c.url)
		}
	}
}

func TestEventInputNormalize_Dates(t *testing.T) {
	now := time.Now().UTC()
	// ends_at before starts_at
	in := validTestInput()
	in.StartsAt = now.Add(time.Hour)
	in.EndsAt = now
	if err := in.normalize(); err == nil {
		t.Error("expected ends_at before starts_at to fail")
	}

	// starts_at == ends_at
	in = validTestInput()
	in.StartsAt = now
	in.EndsAt = now
	if err := in.normalize(); err == nil {
		t.Error("expected equal dates to fail")
	}

	// Zero dates
	in = validTestInput()
	in.StartsAt = time.Time{}
	if err := in.normalize(); err == nil {
		t.Error("expected zero starts_at to fail")
	}
}

func TestEventInputNormalize_PriorityAndStatus(t *testing.T) {
	in := validTestInput()
	in.PopupPriority = 101
	if err := in.normalize(); err == nil {
		t.Error("expected priority > 100 to fail")
	}

	in = validTestInput()
	in.PopupPriority = -101
	if err := in.normalize(); err == nil {
		t.Error("expected priority < -100 to fail")
	}

	in = validTestInput()
	in.Status = "unknown"
	if err := in.normalize(); err == nil {
		t.Error("expected invalid status to fail")
	}

	in = validTestInput()
	in.Status = ""
	if err := in.normalize(); err != nil || in.Status != StatusDraft {
		t.Errorf("expected empty status to default to draft, got: %v, status: %s", err, in.Status)
	}
}

func TestEventStateAt(t *testing.T) {
	now := time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)
	ev := Event{
		StartsAt: time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC),
		EndsAt:   time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC),
	}

	// During
	if ev.StateAt(now) != StateActive {
		t.Errorf("expected active, got %s", ev.StateAt(now))
	}

	// Before
	if ev.StateAt(now.Add(-10*24*time.Hour)) != StateUpcoming {
		t.Errorf("expected upcoming, got %s", ev.StateAt(now.Add(-10*24*time.Hour)))
	}

	// After
	if ev.StateAt(now.Add(10*24*time.Hour)) != StatePast {
		t.Errorf("expected past, got %s", ev.StateAt(now.Add(10*24*time.Hour)))
	}
}

func TestEventInputNormalize_VotingRounds(t *testing.T) {
	in := validTestInput()
	in.Type = TypeVoteMonthly
	mediaType := "movie"
	period := "2026-11"
	in.MediaType = &mediaType
	in.Period = &period
	in.Nominees = []NomineeInput{
		{MediaID: 101, MediaType: "movie", MediaTitle: "Dune", Position: 1},
		{MediaID: 102, MediaType: "movie", MediaTitle: "Oppenheimer", Position: 2},
	}

	if err := in.normalize(); err != nil {
		t.Fatalf("expected valid monthly vote input, got: %v", err)
	}

	// Invalid media type
	badMediaType := "anime"
	in.MediaType = &badMediaType
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "media_type") {
		t.Errorf("expected media_type error, got: %v", err)
	}
	in.MediaType = &mediaType

	// Invalid monthly period
	badPeriod := "2026-13"
	in.Period = &badPeriod
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "period") {
		t.Errorf("expected period error for month 13, got: %v", err)
	}

	badPeriod2 := "invalid-date"
	in.Period = &badPeriod2
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "period") {
		t.Errorf("expected period error, got: %v", err)
	}
	in.Period = &period

	// Nominee validation
	in.Nominees[0].MediaID = 0
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "nominees.media_id") {
		t.Errorf("expected nominee media_id error, got: %v", err)
	}
	in.Nominees[0].MediaID = 101

	in.Nominees[0].MediaTitle = ""
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "nominees.media_title") {
		t.Errorf("expected nominee media_title error, got: %v", err)
	}
	in.Nominees[0].MediaTitle = "Dune"

	// Duplicate nominee MediaID in same round fails
	in.Nominees[1].MediaID = 101
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "nominees.media_id") {
		t.Errorf("expected duplicate nominee media_id error, got: %v", err)
	}
	in.Nominees[1].MediaID = 102

	// Nominee MediaType mismatch with event MediaType fails
	in.Nominees[1].MediaType = "tv"
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "nominees.media_type") {
		t.Errorf("expected nominee media_type mismatch error, got: %v", err)
	}
}

func TestNormalizeNominees_ImageAndOverviewFields(t *testing.T) {
	tooLongPath := "/" + strings.Repeat("a", maxNomineePathLength)
	tooLongOverview := strings.Repeat("ก", maxNomineeOverviewRunes+1)
	cases := []struct {
		name      string
		poster    string
		backdrop  string
		overview  string
		wantField string
	}{
		{name: "empty paths and overview are allowed", wantField: ""},
		{name: "relative TMDB path", poster: "/abc.jpg", backdrop: "/def.jpg", overview: "Plot", wantField: ""},
		{name: "poster with scheme", poster: "https://evil.test/x.jpg", wantField: "nominees.poster_path"},
		{name: "poster protocol-relative", poster: "//evil.test/x.jpg", wantField: "nominees.poster_path"},
		{name: "poster without leading slash", poster: "abc.jpg", wantField: "nominees.poster_path"},
		{name: "poster with whitespace", poster: "/a b.jpg", wantField: "nominees.poster_path"},
		{name: "poster with backslash", poster: "/a\\b.jpg", wantField: "nominees.poster_path"},
		{name: "poster over length cap", poster: tooLongPath, wantField: "nominees.poster_path"},
		{name: "backdrop with scheme", backdrop: "javascript:alert(1)", wantField: "nominees.backdrop_path"},
		{name: "backdrop protocol-relative", backdrop: "//evil.test/b.jpg", wantField: "nominees.backdrop_path"},
		{name: "backdrop over length cap", backdrop: tooLongPath, wantField: "nominees.backdrop_path"},
		{name: "overview over rune cap", overview: tooLongOverview, wantField: "nominees.overview"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			noms := []NomineeInput{{
				MediaID: 1, MediaType: "movie", MediaTitle: "Dune",
				PosterPath: tc.poster, BackdropPath: tc.backdrop, Overview: tc.overview,
			}}
			err := normalizeNominees(noms, "movie")
			if tc.wantField == "" {
				if err != nil {
					t.Fatalf("expected valid nominee, got: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantField) {
				t.Fatalf("expected %s error, got: %v", tc.wantField, err)
			}
			if !errors.Is(err, ErrInvalidInput) {
				t.Errorf("expected ErrInvalidInput wrapping, got: %v", err)
			}
		})
	}
}

func TestCheckPublishable(t *testing.T) {
	cases := []struct {
		name    string
		status  EventStatus
		evType  EventType
		count   int
		wantErr bool
	}{
		{name: "published voting round with two nominees", status: StatusPublished, evType: TypeVoteMonthly, count: 2},
		{name: "published voting round with one nominee", status: StatusPublished, evType: TypeVoteMonthly, count: 1, wantErr: true},
		{name: "published voting round with none", status: StatusPublished, evType: TypeVoteYearly, count: 0, wantErr: true},
		{name: "draft round may have one nominee", status: StatusDraft, evType: TypeVoteMonthly, count: 1},
		{name: "cancelled round may have none", status: StatusCancelled, evType: TypeVoteMonthly, count: 0},
		{name: "published announcement needs no nominees", status: StatusPublished, evType: TypeAnnouncement, count: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkPublishable(&Event{Status: tc.status, Type: tc.evType}, tc.count)
			if tc.wantErr && !errors.Is(err, ErrNotEnoughNominees) {
				t.Fatalf("expected ErrNotEnoughNominees, got: %v", err)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
		})
	}
}

func TestMonthlyWindow(t *testing.T) {
	starts, ends, err := MonthlyWindow("2026-11")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantStarts := time.Date(2026, 10, votingCutoffDay, 0, 0, 0, 0, votingLocation).UTC()
	wantEnds := time.Date(2026, 11, votingCutoffDay, 0, 0, 0, 0, votingLocation).UTC()
	if !starts.Equal(wantStarts) {
		t.Errorf("expected starts %v, got: %v", wantStarts, starts)
	}
	if !ends.Equal(wantEnds) {
		t.Errorf("expected ends %v, got: %v", wantEnds, ends)
	}

	// January boundary (month roll back to previous year December)
	startsJan, endsJan, err := MonthlyWindow("2026-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantStartsJan := time.Date(2025, 12, votingCutoffDay, 0, 0, 0, 0, votingLocation).UTC()
	wantEndsJan := time.Date(2026, 1, votingCutoffDay, 0, 0, 0, 0, votingLocation).UTC()
	if !startsJan.Equal(wantStartsJan) {
		t.Errorf("expected starts %v, got: %v", wantStartsJan, startsJan)
	}
	if !endsJan.Equal(wantEndsJan) {
		t.Errorf("expected ends %v, got: %v", wantEndsJan, endsJan)
	}
}

func TestCurrentMonthlyPeriod(t *testing.T) {
	// Before the 15th in Bangkok -> current calendar month
	before := time.Date(2026, 5, 14, 23, 59, 0, 0, votingLocation)
	if got := CurrentMonthlyPeriod(before); got != "2026-05" {
		t.Errorf("expected 2026-05 before cutoff, got %s", got)
	}
	// On/after the 15th in Bangkok -> next calendar month
	after := time.Date(2026, 5, 15, 0, 0, 0, 0, votingLocation)
	if got := CurrentMonthlyPeriod(after); got != "2026-06" {
		t.Errorf("expected 2026-06 on/after cutoff, got %s", got)
	}
	// December after the 15th rolls over to January of next year
	decAfter := time.Date(2026, 12, 20, 12, 0, 0, 0, votingLocation)
	if got := CurrentMonthlyPeriod(decAfter); got != "2027-01" {
		t.Errorf("expected 2027-01 in late December, got %s", got)
	}
}

func TestCheckVotingLock(t *testing.T) {
	movie := "movie"
	tv := "tv"
	period := "2026-05"
	otherPeriod := "2026-06"
	starts := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	ends := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)

	ev := Event{
		Type:      TypeVoteMonthly,
		MediaType: &movie,
		Period:    &period,
		StartsAt:  starts,
		EndsAt:    ends,
	}

	baseIn := EventInput{
		Type:      TypeVoteMonthly,
		MediaType: &movie,
		Period:    &period,
		TitleEN:   "Updated Title",
		TitleTH:   "อัปเดตชื่อ",
		StartsAt:  starts,
		EndsAt:    ends,
	}

	// Before starts_at: changing nominees or period is allowed
	inBefore := baseIn
	inBefore.Period = &otherPeriod
	inBefore.Nominees = []NomineeInput{{MediaID: 1, MediaType: "movie", MediaTitle: "A"}}
	if err := checkVotingLock(&ev, inBefore, starts.Add(-time.Minute)); err != nil {
		t.Errorf("expected edits before starts_at to be allowed, got: %v", err)
	}

	// Once active (now >= starts_at): title/body edits without touching structural fields are allowed
	activeNow := starts.Add(time.Minute)
	if err := checkVotingLock(&ev, baseIn, activeNow); err != nil {
		t.Errorf("expected copy edit after starts_at to be allowed, got: %v", err)
	}

	// Once active: replacing nominees is rejected with ErrVotingLocked
	inNoms := baseIn
	inNoms.Nominees = []NomineeInput{{MediaID: 1, MediaType: "movie", MediaTitle: "A"}}
	if err := checkVotingLock(&ev, inNoms, activeNow); err != ErrVotingLocked {
		t.Errorf("expected ErrVotingLocked when editing nominees after open, got: %v", err)
	}

	// Once active: changing media_type, period, type, or ends_at is rejected
	inMedia := baseIn
	inMedia.MediaType = &tv
	if err := checkVotingLock(&ev, inMedia, activeNow); err != ErrVotingLocked {
		t.Errorf("expected ErrVotingLocked when changing media_type, got: %v", err)
	}

	inEnds := baseIn
	inEnds.EndsAt = ends.Add(time.Hour)
	if err := checkVotingLock(&ev, inEnds, activeNow); err != ErrVotingLocked {
		t.Errorf("expected ErrVotingLocked when shifting ends_at, got: %v", err)
	}
}

func TestEventInputNormalize_VotingRoundAllowsEmptyCopy(t *testing.T) {
	mediaType := "movie"
	period := "2026-11"
	in := validTestInput()
	in.Type = TypeVoteMonthly
	in.MediaType = &mediaType
	in.Period = &period
	in.TitleEN, in.TitleTH, in.BodyEN, in.BodyTH = "", "", "", ""
	if err := in.normalize(); err != nil {
		t.Fatalf("expected blank voting round copy to be valid, got: %v", err)
	}

	// Length limits still apply to voting rounds.
	in.TitleEN = strings.Repeat("A", maxTitleRunes+1)
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "title_en") {
		t.Errorf("expected title_en length error for voting round, got: %v", err)
	}
}

func TestEventInputNormalize_YearlyRoundRequiresTitles(t *testing.T) {
	mediaType := "movie"
	period := "2026"
	in := validTestInput()
	in.Type = TypeVoteYearly
	in.MediaType = &mediaType
	in.Period = &period
	in.TitleEN = ""
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "title_en") {
		t.Errorf("expected title_en required for yearly round, got: %v", err)
	}
}

func TestEventInputNormalize_AnnouncementRequiresCopy(t *testing.T) {
	in := validTestInput()
	in.TitleEN = ""
	if err := in.normalize(); err == nil || !strings.Contains(err.Error(), "title_en") {
		t.Errorf("expected title_en required for announcement, got: %v", err)
	}
}

func TestNextMonthlyPeriod(t *testing.T) {
	cases := []struct {
		name string
		at   time.Time
		want string
	}{
		{"bangkok day 14 late evening opens next month", time.Date(2026, 5, 14, 23, 59, 0, 0, votingLocation), "2026-06"},
		{"bangkok day 15 midnight skips a month", time.Date(2026, 5, 15, 0, 0, 0, 0, votingLocation), "2026-07"},
		{"utc 16:59 on day 14 is bangkok 23:59 on day 14", time.Date(2026, 5, 14, 16, 59, 59, 0, time.UTC), "2026-06"},
		{"utc 17:00 on day 14 is bangkok 00:00 on day 15", time.Date(2026, 5, 14, 17, 0, 0, 0, time.UTC), "2026-07"},
		{"december before cutoff rolls to january", time.Date(2026, 12, 14, 12, 0, 0, 0, votingLocation), "2027-01"},
		{"december after cutoff rolls to february", time.Date(2026, 12, 20, 12, 0, 0, 0, votingLocation), "2027-02"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := NextMonthlyPeriod(tc.at); got != tc.want {
				t.Errorf("NextMonthlyPeriod(%v) = %s, want %s", tc.at, got, tc.want)
			}
		})
	}
}

func TestWinningTallies(t *testing.T) {
	if got := winningTallies(nil); len(got) != 0 {
		t.Errorf("expected no winners without votes, got: %v", got)
	}
	if got := winningTallies([]voteTally{{NomineeID: 1, Votes: 0}}); len(got) != 0 {
		t.Errorf("expected no winners when top tally is zero, got: %v", got)
	}
	tied := []voteTally{{NomineeID: 1, Votes: 3}, {NomineeID: 2, Votes: 3}, {NomineeID: 3, Votes: 1}}
	got := winningTallies(tied)
	if len(got) != 2 || got[0].NomineeID != 1 || got[1].NomineeID != 2 {
		t.Errorf("expected co-winners 1 and 2, got: %v", got)
	}
}

package event_module

import (
	"context"
	"sort"
	"time"

	"github.com/arinsuda/movie-hub/config"
)

// ---------------------------------------------------------------------------
// Yearly windows
// ---------------------------------------------------------------------------

// YearlyWindow returns the UTC voting window of yearly round `period` (YYYY, the year awarded): from
// 00:00 on yearlyStartDay of yearlyStartMonth of the following year to 00:00 on yearlyEndDay of
// yearlyEndMonth, in config.AppTimeZone.
func YearlyWindow(period string) (time.Time, time.Time, error) {
	year, err := time.ParseInLocation(yearlyPeriodLayout, period, config.AppLocation)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	next := year.Year() + 1
	startsAt := time.Date(next, yearlyStartMonth, yearlyStartDay, 0, 0, 0, 0, config.AppLocation)
	endsAt := time.Date(next, yearlyEndMonth, yearlyEndDay, 0, 0, 0, 0, config.AppLocation)
	return startsAt.UTC(), endsAt.UTC(), nil
}

// NextYearlyPeriod is the yearly round that opens next after t: the previous Bangkok year until that
// round starts, then the current year.
func NextYearlyPeriod(t time.Time) string {
	year := t.In(config.AppLocation).Year()
	prev := time.Date(year-1, time.January, 1, 0, 0, 0, 0, config.AppLocation).Format(yearlyPeriodLayout)
	if startsAt, _, err := YearlyWindow(prev); err == nil && t.Before(startsAt) {
		return prev
	}
	return time.Date(year, time.January, 1, 0, 0, 0, 0, config.AppLocation).Format(yearlyPeriodLayout)
}

// currentRound is the round of `kind` whose official window contains now. Monthly windows tile time,
// so a monthly round always exists; ok is false for a yearly round outside its Jan 16 - Feb 15 window.
func currentRound(kind EventType, now time.Time) (period string, startsAt, endsAt time.Time, ok bool) {
	if kind == TypeVoteMonthly {
		period = CurrentMonthlyPeriod(now)
		s, e, err := MonthlyWindow(period)
		return period, s, e, err == nil
	}
	year := now.In(config.AppLocation).Year() - 1 // yearly round Y opens in Y+1
	period = time.Date(year, time.January, 1, 0, 0, 0, 0, config.AppLocation).Format(yearlyPeriodLayout)
	s, e, err := YearlyWindow(period)
	if err != nil || now.Before(s) || !now.Before(e) {
		return "", time.Time{}, time.Time{}, false
	}
	return period, s, e, true
}

// lateStartWindow computes the opening time of a late-started round: lateStartLead from now (minute
// precision) with the official end kept. ErrVotingWindowTooShort if less than lateStartMinWindow is left.
func lateStartWindow(now, endsAt time.Time) (time.Time, error) {
	startsAt := now.Truncate(time.Minute).Add(lateStartLead)
	if endsAt.Sub(startsAt) < lateStartMinWindow {
		return time.Time{}, ErrVotingWindowTooShort
	}
	return startsAt, nil
}

// ---------------------------------------------------------------------------
// Round generation
// ---------------------------------------------------------------------------

// SuggestFunc proposes monthly nominees; tests inject a stub instead of TMDB.
type SuggestFunc func(ctx context.Context, mediaType, period string) ([]SuggestedNominee, error)

// genOpts tunes one generation. auto marks a scheduler run; autoPublish lets the worker publish the
// draft; lateStart opens the CURRENT round lateStartLead from now instead of at its official start.
type genOpts struct {
	auto        bool
	autoPublish bool
	lateStart   bool
}

// roundSpec identifies one round to create.
type roundSpec struct {
	eventType EventType
	mediaType string
	period    string
	startsAt  time.Time
	endsAt    time.Time
}

// generateRound creates a draft voting round. `opts.auto` marks a scheduler run: it has no admin, flags
// the row auto_generated and checks for an existing round Unscoped, so a deleted draft is never recreated
// behind the admin's back. `precheck` runs after the existence check (yearly: monthly rounds settled);
// `nominees` builds the list. The unique (type, media_type, period) index settles concurrent callers.
func (s *Service) generateRound(ctx context.Context, adminID uint, spec roundSpec, opts genOpts,
	precheck func(context.Context) error, nominees func(context.Context) ([]NomineeInput, error)) (AdminEventResponse, error) {
	now := s.now().UTC()
	if opts.lateStart {
		current, _, endsAt, ok := currentRound(spec.eventType, now)
		if !ok || current != spec.period {
			return AdminEventResponse{}, invalid("period")
		}
		startsAt, err := lateStartWindow(now, endsAt)
		if err != nil {
			return AdminEventResponse{}, err
		}
		spec.startsAt, spec.endsAt = startsAt, endsAt
	}
	if !now.Before(spec.startsAt) {
		return AdminEventResponse{}, invalid("period")
	}

	db := s.db.WithContext(ctx)
	if opts.auto {
		db = db.Unscoped()
	}
	var count int64
	if err := db.Model(&Event{}).
		Where("type = ? AND media_type = ? AND period = ?", spec.eventType, spec.mediaType, spec.period).
		Count(&count).Error; err != nil {
		return AdminEventResponse{}, err
	}
	if count > 0 {
		return AdminEventResponse{}, ErrEventExists
	}
	if precheck != nil {
		if err := precheck(ctx); err != nil {
			return AdminEventResponse{}, err
		}
	}
	noms, err := nominees(ctx)
	if err != nil {
		return AdminEventResponse{}, err
	}

	mediaType, period := spec.mediaType, spec.period
	return s.Create(ctx, adminID, EventInput{
		Type:          spec.eventType,
		MediaType:     &mediaType,
		Period:        &period,
		StartsAt:      spec.startsAt,
		EndsAt:        spec.endsAt,
		ShowPopup:     true,
		PopupPriority: generatedPopupPriority,
		Status:        StatusDraft,
		Nominees:      noms,
		autoGenerated: opts.auto,
		autoPublish:   opts.autoPublish,
		lateStart:     opts.lateStart,
	})
}

// GenerateMonthlyVote creates the monthly round for (mediaType, period) as a draft with
// nomineesPerRound suggested titles. The round has no copy: the client renders it from i18n.
// invalid("period") if the round already opened; ErrEventExists if it already exists;
// ErrNotEnoughNominees if suggestions could not fill it. With lateStart the period must be the
// current round, which then opens lateStartLead from now (ErrVotingWindowTooShort if too little is left).
func (s *Service) GenerateMonthlyVote(ctx context.Context, adminID uint, mediaType, period string, lateStart bool) (AdminEventResponse, error) {
	return s.generateMonthly(ctx, adminID, mediaType, period, genOpts{lateStart: lateStart}, s.SuggestNominees)
}

func (s *Service) generateMonthly(ctx context.Context, adminID uint, mediaType, period string, opts genOpts, suggest SuggestFunc) (AdminEventResponse, error) {
	if !validMediaType(mediaType) {
		return AdminEventResponse{}, invalid("media_type")
	}
	startsAt, endsAt, err := MonthlyWindow(period)
	if err != nil {
		return AdminEventResponse{}, invalid("period")
	}
	spec := roundSpec{eventType: TypeVoteMonthly, mediaType: mediaType, period: period, startsAt: startsAt, endsAt: endsAt}
	return s.generateRound(ctx, adminID, spec, opts, nil, func(ctx context.Context) ([]NomineeInput, error) {
		suggested, err := suggest(ctx, mediaType, period)
		if err != nil {
			return nil, err
		}
		if len(suggested) < nomineesPerRound {
			return nil, ErrNotEnoughNominees
		}
		nominees := make([]NomineeInput, 0, len(suggested))
		for i, sug := range suggested {
			nominees = append(nominees, NomineeInput{
				MediaID:      sug.MediaID,
				MediaType:    sug.MediaType,
				Position:     i + 1,
				MediaTitle:   sug.MediaTitle,
				PosterPath:   sug.PosterPath,
				BackdropPath: sug.BackdropPath,
				Overview:     sug.Overview,
			})
		}
		return nominees, nil
	})
}

// GenerateYearlyVote creates the yearly round of `period` (YYYY, the year awarded) as a draft whose
// nominees are that year's monthly winners. It may have fewer than minNomineesToPublish nominees
// (the admin can edit the draft). invalid("period") if the round already opened; ErrEventExists if it
// exists; ErrMonthlyRoundsPending while a published monthly round of the year is unsettled.
func (s *Service) GenerateYearlyVote(ctx context.Context, adminID uint, mediaType, period string, lateStart bool) (AdminEventResponse, error) {
	return s.generateYearly(ctx, adminID, mediaType, period, genOpts{lateStart: lateStart})
}

func (s *Service) generateYearly(ctx context.Context, adminID uint, mediaType, period string, opts genOpts) (AdminEventResponse, error) {
	if !validMediaType(mediaType) {
		return AdminEventResponse{}, invalid("media_type")
	}
	if !yearlyPeriodRe.MatchString(period) {
		return AdminEventResponse{}, invalid("period")
	}
	startsAt, endsAt, err := YearlyWindow(period)
	if err != nil {
		return AdminEventResponse{}, invalid("period")
	}
	spec := roundSpec{eventType: TypeVoteYearly, mediaType: mediaType, period: period, startsAt: startsAt, endsAt: endsAt}
	return s.generateRound(ctx, adminID, spec, opts,
		func(ctx context.Context) error { return s.checkMonthlySettled(ctx, mediaType, period) },
		func(ctx context.Context) ([]NomineeInput, error) { return s.yearlyNominees(ctx, mediaType, period) })
}

// checkMonthlySettled refuses a yearly round while a published, non-deleted monthly round of the same
// year and media type is still unfinalized (its winner would be missing from the nominees).
func (s *Service) checkMonthlySettled(ctx context.Context, mediaType, year string) error {
	var pending int64
	err := s.db.WithContext(ctx).Model(&Event{}).
		Scopes(publishedScope).
		Where("type = ? AND media_type = ? AND is_finalized = ? AND period LIKE ?",
			TypeVoteMonthly, mediaType, false, year+"-%").
		Count(&pending).Error
	if err != nil {
		return err
	}
	if pending > 0 {
		return ErrMonthlyRoundsPending
	}
	return nil
}

// monthlyWin is one monthly award joined to its nominee snapshot.
type monthlyWin struct {
	MediaID      int
	Period       string
	Votes        int
	MediaTitle   string
	PosterPath   string
	BackdropPath string
	Overview     string
}

// yearlyNominees lists the monthly winners of `year` for a media type as nominees, best first.
// Only published, non-deleted rounds count; nominee data is copied from the winning round (no TMDB).
func (s *Service) yearlyNominees(ctx context.Context, mediaType, year string) ([]NomineeInput, error) {
	var wins []monthlyWin
	err := s.db.WithContext(ctx).Table("event_awards a").
		Select("a.media_id, a.period, a.votes, n.media_title, n.poster_path, n.backdrop_path, n.overview").
		Joins("JOIN events e ON e.id = a.event_id").
		Joins("JOIN event_nominees n ON n.id = a.nominee_id").
		Where("e.deleted_at IS NULL AND e.status = ?", StatusPublished).
		Where("a.award_kind = ? AND a.media_type = ? AND a.period LIKE ?", AwardKindMonthly, mediaType, year+"-%").
		Scan(&wins).Error
	if err != nil {
		return nil, err
	}
	return rankYearlyNominees(wins, mediaType), nil
}

// rankYearlyNominees groups wins by title and ranks them by wins DESC, total votes DESC, earliest
// winning period ASC (then media id, for a stable order), capped at maxNomineesPerRound. The snapshot
// comes from the title's latest win.
func rankYearlyNominees(wins []monthlyWin, mediaType string) []NomineeInput {
	type tally struct {
		wins, votes int
		first       string
		latest      monthlyWin
	}
	byMedia := make(map[int]*tally, len(wins))
	for _, w := range wins {
		t, ok := byMedia[w.MediaID]
		if !ok {
			byMedia[w.MediaID] = &tally{wins: 1, votes: w.Votes, first: w.Period, latest: w}
			continue
		}
		t.wins++
		t.votes += w.Votes
		if w.Period < t.first {
			t.first = w.Period
		}
		if w.Period > t.latest.Period {
			t.latest = w
		}
	}
	ranked := make([]*tally, 0, len(byMedia))
	for _, t := range byMedia {
		ranked = append(ranked, t)
	}
	sort.Slice(ranked, func(i, j int) bool {
		a, b := ranked[i], ranked[j]
		switch {
		case a.wins != b.wins:
			return a.wins > b.wins
		case a.votes != b.votes:
			return a.votes > b.votes
		case a.first != b.first:
			return a.first < b.first
		}
		return a.latest.MediaID < b.latest.MediaID
	})
	if len(ranked) > maxNomineesPerRound {
		ranked = ranked[:maxNomineesPerRound]
	}
	out := make([]NomineeInput, 0, len(ranked))
	for i, t := range ranked {
		out = append(out, NomineeInput{
			MediaID:      t.latest.MediaID,
			MediaType:    mediaType,
			Position:     i + 1,
			MediaTitle:   t.latest.MediaTitle,
			PosterPath:   t.latest.PosterPath,
			BackdropPath: t.latest.BackdropPath,
			Overview:     t.latest.Overview,
		})
	}
	return out
}

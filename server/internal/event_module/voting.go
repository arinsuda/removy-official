package event_module

import (
	"context"
	"errors"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/admin_module"
	"github.com/arinsuda/movie-hub/internal/shared"
	tmdbmodule "github.com/arinsuda/movie-hub/internal/tmdb_module"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// eventsPeriodUniqueIndex is created by migration events_type_media_period_unique_v1.
	eventsPeriodUniqueIndex = "idx_events_type_media_period"
	percentScale            = 100
)

// ---------------------------------------------------------------------------
// Nominee persistence helpers
// ---------------------------------------------------------------------------

func (s *Service) loadNominees(ctx context.Context, eventID uint) ([]EventNominee, error) {
	return loadNomineesTx(s.db.WithContext(ctx), eventID)
}

func loadNomineesTx(db *gorm.DB, eventID uint) ([]EventNominee, error) {
	var rows []EventNominee
	err := db.Where("event_id = ?", eventID).Order("position ASC, id ASC").Find(&rows).Error
	return rows, err
}

// loadNomineesFor batch-loads nominees for several events (one query, no N+1).
func (s *Service) loadNomineesFor(ctx context.Context, eventIDs []uint) (map[uint][]EventNominee, error) {
	out := make(map[uint][]EventNominee, len(eventIDs))
	if len(eventIDs) == 0 {
		return out, nil
	}
	var rows []EventNominee
	if err := s.db.WithContext(ctx).Where("event_id IN ?", eventIDs).
		Order("event_id ASC, position ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, n := range rows {
		out[n.EventID] = append(out[n.EventID], n)
	}
	return out, nil
}

// replaceNominees swaps an event's nominee list. Callers must have checked that the
// round has not opened: deleting a nominee cascades to its votes and awards.
func replaceNominees(tx *gorm.DB, eventID uint, in []NomineeInput, now time.Time) ([]EventNominee, error) {
	if err := tx.Where("event_id = ?", eventID).Delete(&EventNominee{}).Error; err != nil {
		return nil, err
	}
	rows := make([]EventNominee, 0, len(in))
	for i, n := range in {
		pos := n.Position
		if pos <= 0 {
			pos = i + 1
		}
		rows = append(rows, EventNominee{
			EventID:      eventID,
			MediaID:      n.MediaID,
			MediaType:    n.MediaType,
			Position:     pos,
			MediaTitle:   n.MediaTitle,
			PosterPath:   n.PosterPath,
			BackdropPath: n.BackdropPath,
			Overview:     n.Overview,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}
	if len(rows) == 0 {
		return rows, nil
	}
	if err := tx.Omit(clause.Associations).Create(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// checkNoVotes refuses a nominee change once anyone has voted: replacing or deleting a
// nominee cascades to its votes and awards.
func checkNoVotes(tx *gorm.DB, eventID uint) error {
	var voted bool
	if err := tx.Raw("SELECT EXISTS(SELECT 1 FROM event_votes WHERE event_id = ?)", eventID).Row().Scan(&voted); err != nil {
		return err
	}
	if voted {
		return ErrVotingLocked
	}
	return nil
}

// revalidateNominees keeps the nominees valid after an update changed the event's type or
// media type without sending a new list. Switching to an announcement clears them; a voting
// round keeps them only if they still match its media type.
func revalidateNominees(tx *gorm.DB, eventID uint, in EventInput, now time.Time) ([]EventNominee, error) {
	if err := checkNoVotes(tx, eventID); err != nil {
		return nil, err
	}
	if !in.Type.IsVoting() {
		return replaceNominees(tx, eventID, nil, now)
	}
	existing, err := loadNomineesTx(tx, eventID)
	if err != nil {
		return nil, err
	}
	for _, n := range existing {
		if n.MediaType != *in.MediaType {
			return nil, invalid("nominees.media_type")
		}
	}
	return existing, nil
}

// checkPublishable refuses a published voting round with fewer than minNomineesToPublish titles.
// Drafts may be saved with any number of nominees.
func checkPublishable(ev *Event, nomineeCount int) error {
	if ev.Status == StatusPublished && ev.Type.IsVoting() && nomineeCount < minNomineesToPublish {
		return ErrNotEnoughNominees
	}
	return nil
}

func toNomineeResponse(n *EventNominee) NomineeResponse {
	return NomineeResponse{
		ID:           n.ID,
		EventID:      n.EventID,
		MediaID:      n.MediaID,
		MediaType:    n.MediaType,
		Position:     n.Position,
		MediaTitle:   n.MediaTitle,
		PosterPath:   optional(n.PosterPath),
		BackdropPath: optional(n.BackdropPath),
		Overview:     optional(n.Overview),
	}
}

type voteTally struct {
	NomineeID uint
	Votes     int
}

// tallyVotes counts votes per nominee, ignoring voters who are soft-deleted or deactivated.
// Tallies are sorted by votes DESC, so the leaders come first.
func tallyVotes(db *gorm.DB, eventID uint) ([]voteTally, error) {
	var tallies []voteTally
	err := db.Model(&EventVote{}).
		Select("event_votes.nominee_id, COUNT(*) AS votes").
		Joins("JOIN users u ON u.id = event_votes.user_id").
		Where("event_votes.event_id = ? AND u.deleted_at IS NULL AND u.is_active", eventID).
		Group("event_votes.nominee_id").
		Order("votes DESC, event_votes.nominee_id ASC").
		Scan(&tallies).Error
	return tallies, err
}

// winningTallies returns the top-scoring tallies (co-winners). Empty when nobody voted.
func winningTallies(tallies []voteTally) []voteTally {
	if len(tallies) == 0 || tallies[0].Votes == 0 {
		return nil
	}
	top := tallies[0].Votes
	end := 0
	for end < len(tallies) && tallies[end].Votes == top {
		end++
	}
	return tallies[:end]
}

// finalVotes is the settled count of a nominee; nil (round not settled) reads as 0.
func finalVotes(n EventNominee) int {
	if n.FinalVotes == nil {
		return 0
	}
	return *n.FinalVotes
}

func votePercentage(count, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(count) / float64(total) * percentScale
}

// mapEventUniqueErr turns a duplicate (type, media_type, period) insert into ErrEventExists.
func mapEventUniqueErr(err error) error {
	if shared.IsPgUniqueViolationOn(err, eventsPeriodUniqueIndex) {
		return ErrEventExists
	}
	return err
}

// checkVotingLock freezes a voting round once it opens: its type, media type, period,
// window and nominees may no longer change (votes and awards depend on them).
// Title, body, image, popup settings and status stay editable.
func checkVotingLock(ev *Event, in EventInput, now time.Time) error {
	if !ev.Type.IsVoting() && !in.Type.IsVoting() {
		return nil
	}
	if now.Before(ev.StartsAt) {
		return nil
	}
	if in.Nominees != nil ||
		in.Type != ev.Type ||
		!equalStringPtr(in.MediaType, ev.MediaType) ||
		!equalStringPtr(in.Period, ev.Period) ||
		!sameFormMinute(in.StartsAt, ev.StartsAt) ||
		!sameFormMinute(in.EndsAt, ev.EndsAt) {
		return ErrVotingLocked
	}
	return nil
}

func equalStringPtr(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// sameFormMinute compares at the admin form's precision (datetime-local has minutes).
func sameFormMinute(a, b time.Time) bool {
	return a.Truncate(time.Minute).Equal(b.Truncate(time.Minute))
}

// ---------------------------------------------------------------------------
// Voting
// ---------------------------------------------------------------------------

// openRoundForVoting loads a published voting round inside tx and checks it is open.
// FOR SHARE blocks finalization (FOR UPDATE, see lockEvent) and nominee edits until tx ends.
// The clock is read after the lock, so a round that finalized while the caller waited is refused.
// It returns the instant the check ran.
func openRoundForVoting(tx *gorm.DB, eventID uint, nowFn func() time.Time) (time.Time, error) {
	var ev Event
	err := tx.Clauses(clause.Locking{Strength: clause.LockingStrengthShare}).
		Scopes(publishedScope).First(&ev, eventID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return time.Time{}, ErrNotFound
	}
	if err != nil {
		return time.Time{}, err
	}
	if !ev.Type.IsVoting() {
		return time.Time{}, ErrInvalidEventType
	}
	now := nowFn().UTC()
	if ev.IsFinalized {
		return time.Time{}, ErrVotingClosed
	}
	if now.Before(ev.StartsAt) {
		return time.Time{}, ErrVotingNotStarted
	}
	if !now.Before(ev.EndsAt) {
		return time.Time{}, ErrVotingClosed
	}
	return now, nil
}

// Vote records or changes the user's single vote for the round (atomic upsert).
func (s *Service) Vote(ctx context.Context, userID, eventID, nomineeID uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now, err := openRoundForVoting(tx, eventID, s.now)
		if err != nil {
			return err
		}
		var exists int64
		if err := tx.Model(&EventNominee{}).Where("id = ? AND event_id = ?", nomineeID, eventID).Count(&exists).Error; err != nil {
			return err
		}
		if exists == 0 {
			return ErrNomineeNotFound
		}
		vote := EventVote{EventID: eventID, UserID: userID, NomineeID: nomineeID, CreatedAt: now, UpdatedAt: now}
		return tx.Omit(clause.Associations).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "event_id"}, {Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"nominee_id", "updated_at"}),
		}).Create(&vote).Error
	})
}

// CancelVote removes the user's vote while the round is open. Idempotent.
func (s *Service) CancelVote(ctx context.Context, userID, eventID uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := openRoundForVoting(tx, eventID, s.now); err != nil {
			return err
		}
		return tx.Where("event_id = ? AND user_id = ?", eventID, userID).Delete(&EventVote{}).Error
	})
}

// ---------------------------------------------------------------------------
// Finalization
// ---------------------------------------------------------------------------

// Finalize settles an ended round on an admin's request. It refuses before ends_at.
func (s *Service) Finalize(ctx context.Context, adminID, eventID uint) error {
	return s.finalizeTx(ctx, adminID, eventID)
}

// finalizeTx settles an ended round: it stores every nominee's final tally and records the
// top-scoring nominees as (co-)winners. Idempotent: lockEvent serializes callers, and a
// finalized round is a no-op.
func (s *Service) finalizeTx(ctx context.Context, adminID, eventID uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ev Event
		if err := lockEvent(tx, eventID, &ev); err != nil {
			return err
		}
		if !ev.Type.IsVoting() {
			return ErrInvalidEventType
		}
		if ev.IsFinalized {
			return nil
		}
		if ev.Status != StatusPublished {
			return ErrEventNotPublished
		}
		now := s.now().UTC()
		if now.Before(ev.EndsAt) {
			return ErrVotingNotEnded
		}

		nominees, err := loadNomineesTx(tx, eventID)
		if err != nil {
			return err
		}
		tallies, err := tallyVotes(tx, eventID)
		if err != nil {
			return err
		}
		counts := make(map[uint]int, len(tallies))
		for _, t := range tallies {
			counts[t.NomineeID] = t.Votes
		}
		for _, n := range nominees {
			if err := tx.Model(&EventNominee{}).Where("id = ?", n.ID).Update("final_votes", counts[n.ID]).Error; err != nil {
				return err
			}
		}

		awards := buildAwards(ev, nominees, winningTallies(tallies), now)
		if len(awards) > 0 {
			if err := tx.Omit(clause.Associations).Create(&awards).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&ev).Update("is_finalized", true).Error; err != nil {
			return err
		}
		if adminID > 0 {
			return writeAudit(tx, adminID, admin_module.ActionEventFinalized, eventID, map[string]any{"winners": len(awards)})
		}
		return nil
	})
}

// buildAwards turns the winning tallies into award rows for the round's kind and period.
func buildAwards(ev Event, nominees []EventNominee, winners []voteTally, now time.Time) []EventAward {
	kind := AwardKindMonthly
	if ev.Type == TypeVoteYearly {
		kind = AwardKindYearly
	}
	period := ""
	if ev.Period != nil {
		period = *ev.Period
	}
	byID := make(map[uint]EventNominee, len(nominees))
	for _, n := range nominees {
		byID[n.ID] = n
	}
	awards := make([]EventAward, 0, len(winners))
	for _, w := range winners {
		nom, ok := byID[w.NomineeID]
		if !ok {
			continue
		}
		awards = append(awards, EventAward{
			EventID:   ev.ID,
			NomineeID: nom.ID,
			MediaID:   nom.MediaID,
			MediaType: nom.MediaType,
			AwardKind: kind,
			Period:    period,
			Votes:     w.Votes,
			CreatedAt: now,
		})
	}
	return awards
}

// ---------------------------------------------------------------------------
// Nominee editing
// ---------------------------------------------------------------------------

// SetNominees replaces a round's nominees. Only allowed before the round opens.
func (s *Service) SetNominees(ctx context.Context, adminID, eventID uint, nominees []NomineeInput) ([]EventNominee, error) {
	var rows []EventNominee
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ev Event
		if err := lockEvent(tx, eventID, &ev); err != nil {
			return err
		}
		// Read the clock after the row lock, so a round that opened while we waited is refused.
		now := s.now().UTC()
		if !ev.Type.IsVoting() || ev.MediaType == nil {
			return ErrInvalidEventType
		}
		if !now.Before(ev.StartsAt) {
			return ErrVotingLocked
		}
		if err := normalizeNominees(nominees, *ev.MediaType); err != nil {
			return err
		}
		if err := checkNoVotes(tx, eventID); err != nil {
			return err
		}
		var err error
		if rows, err = replaceNominees(tx, eventID, nominees, now); err != nil {
			return err
		}
		// A published round must keep enough titles even when only its nominees change.
		if err := checkPublishable(&ev, len(rows)); err != nil {
			return err
		}
		return writeAudit(tx, adminID, admin_module.ActionNomineesUpdated, eventID, map[string]any{"count": len(rows)})
	})
	return rows, err
}

// ---------------------------------------------------------------------------
// Monthly rounds
// ---------------------------------------------------------------------------

// MonthlyWindow returns the UTC voting window of monthly round `period` (YYYY-MM): from 00:00 on
// votingCutoffDay of `period` to 00:00 on votingCutoffDay of the following month, in config.AppTimeZone.
// Round P is therefore open from the 15th of P until the 15th of P+1.
func MonthlyWindow(period string) (time.Time, time.Time, error) {
	month, err := time.ParseInLocation(monthlyPeriodLayout, period, config.AppLocation)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	startsAt := time.Date(month.Year(), month.Month(), votingCutoffDay, 0, 0, 0, 0, config.AppLocation)
	endsAt := startsAt.AddDate(0, 1, 0)
	return startsAt.UTC(), endsAt.UTC(), nil
}

// currentRoundMonth is the period (first day, in config.AppTimeZone) of the monthly round whose window
// contains t. Before votingCutoffDay the window that contains t started in the previous month.
func currentRoundMonth(t time.Time) time.Time {
	local := t.In(config.AppLocation)
	month := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, config.AppLocation)
	if local.Day() < votingCutoffDay {
		month = month.AddDate(0, -1, 0)
	}
	return month
}

// CurrentMonthlyPeriod is the monthly round whose window contains t.
func CurrentMonthlyPeriod(t time.Time) string {
	return currentRoundMonth(t).Format(monthlyPeriodLayout)
}

// NextMonthlyPeriod is the round after CurrentMonthlyPeriod(t): the next one to open.
func NextMonthlyPeriod(t time.Time) string {
	return currentRoundMonth(t).AddDate(0, 1, 0).Format(monthlyPeriodLayout)
}

// communityScoreSQL ranks titles by activity from active, non-deleted accounts. Views are counted per
// day (media_view_days has one row per user, title and Bangkok day), so the day bounds are dates.
// Weights are bound as DOUBLE PRECISION: a bare ? next to COUNT(*) is typed bigint and would
// truncate fractional weights such as 0.1 to 0.
const communityScoreSQL = `
	SELECT media_id, SUM(score) AS score FROM (
		SELECT r.media_id, COUNT(*) * CAST(? AS DOUBLE PRECISION) AS score
		FROM reviews r JOIN users u ON u.id = r.user_id
		WHERE r.media_type = ? AND r.deleted_at IS NULL AND r.created_at >= ? AND r.created_at < ?
		  AND u.deleted_at IS NULL AND u.is_active
		GROUP BY r.media_id
		UNION ALL
		SELECT l.media_id, COUNT(*) * CAST(? AS DOUBLE PRECISION) AS score
		FROM library_items l JOIN users u ON u.id = l.user_id
		WHERE l.media_type = ? AND l.deleted_at IS NULL AND l.created_at >= ? AND l.created_at < ?
		  AND u.deleted_at IS NULL AND u.is_active
		GROUP BY l.media_id
		UNION ALL
		SELECT m.media_id, COUNT(*) * CAST(? AS DOUBLE PRECISION) AS score
		FROM media_likes m JOIN users u ON u.id = m.user_id
		WHERE m.media_type = ? AND m.deleted_at IS NULL AND m.created_at >= ? AND m.created_at < ?
		  AND u.deleted_at IS NULL AND u.is_active
		GROUP BY m.media_id
		UNION ALL
		SELECT v.media_id, COUNT(*) * CAST(? AS DOUBLE PRECISION) AS score
		FROM media_view_days v JOIN users u ON u.id = v.user_id
		WHERE v.media_type = ? AND v.day >= ? AND v.day < ?
		  AND u.deleted_at IS NULL AND u.is_active
		GROUP BY v.media_id
	) sub
	GROUP BY media_id
	ORDER BY score DESC, media_id ASC
	LIMIT ?`

type communityScore struct {
	MediaID int
	Score   float64
}

// tmdbListItem is the subset of a TMDB list entry a suggestion needs.
type tmdbListItem struct {
	ID                                   int
	Title, PosterPath, BackdropPath, Ovw string
	Popularity                           float64
}

// SuggestNominees proposes up to nomineesPerRound real titles for a monthly round:
// communityNomineeSlots from POPCORN activity in the month before voting opens, then
// TMDB new releases, then TMDB popular titles. It never invents placeholders; fewer
// results simply mean fewer candidates were found.
func (s *Service) SuggestNominees(ctx context.Context, mediaType, period string) ([]SuggestedNominee, error) {
	if !validMediaType(mediaType) {
		return nil, invalid("media_type")
	}
	startsAt, _, err := MonthlyWindow(period)
	if err != nil {
		return nil, invalid("period")
	}
	activityStart := startsAt.AddDate(0, -activityLookbackMonths, 0)
	viewFrom, viewTo := config.LocalDay(activityStart), config.LocalDay(startsAt)

	var community []communityScore
	if err := s.db.WithContext(ctx).Raw(communityScoreSQL,
		reviewScoreWeight, mediaType, activityStart, startsAt,
		libraryScoreWeight, mediaType, activityStart, startsAt,
		likeScoreWeight, mediaType, activityStart, startsAt,
		viewScoreWeight, mediaType, viewFrom, viewTo,
		communityCandidatePool,
	).Scan(&community).Error; err != nil {
		return nil, err
	}

	result := make([]SuggestedNominee, 0, nomineesPerRound)
	seen := make(map[int]bool, nomineesPerRound)
	add := func(n SuggestedNominee) {
		if len(result) < nomineesPerRound && !seen[n.MediaID] {
			seen[n.MediaID] = true
			result = append(result, n)
		}
	}

	for _, c := range community {
		if len(result) >= communityNomineeSlots {
			break
		}
		if seen[c.MediaID] {
			continue
		}
		if nom, ok := fetchMediaInfo(c.MediaID, mediaType, c.Score); ok {
			add(nom)
		}
	}

	for _, src := range []NomineeSource{SourceRelease, SourcePopular} {
		if len(result) >= nomineesPerRound {
			break
		}
		for _, item := range tmdbList(mediaType, src) {
			add(SuggestedNominee{
				MediaID:      item.ID,
				MediaType:    mediaType,
				MediaTitle:   item.Title,
				PosterPath:   item.PosterPath,
				BackdropPath: item.BackdropPath,
				Overview:     item.Ovw,
				Score:        item.Popularity,
				Source:       src,
			})
		}
	}
	return result, nil
}

// tmdbList returns the first page of the TMDB list behind src, or nil when TMDB fails.
func tmdbList(mediaType string, src NomineeSource) []tmdbListItem {
	var out []tmdbListItem
	if mediaType == string(shared.MediaTypeMovie) {
		var res *tmdbmodule.PaginatedResult[tmdbmodule.Movie]
		var err error
		if src == SourceRelease {
			res, err = tmdbmodule.GetNowPlaying(tmdbFirstPage, tmdbmodule.RequestOptions{})
		} else {
			res, err = tmdbmodule.GetPopular(tmdbFirstPage)
		}
		if err != nil || res == nil {
			return nil
		}
		for _, m := range res.Results {
			out = append(out, tmdbListItem{ID: m.ID, Title: m.Title, PosterPath: m.PosterPath, BackdropPath: m.BackdropPath, Ovw: m.Overview, Popularity: float64(m.Popularity)})
		}
		return out
	}
	var res *tmdbmodule.PaginatedResult[tmdbmodule.TVSeries]
	var err error
	if src == SourceRelease {
		res, err = tmdbmodule.GetNowAiringSeries(tmdbFirstPage)
	} else {
		res, err = tmdbmodule.GetPopularSeries(tmdbFirstPage)
	}
	if err != nil || res == nil {
		return nil
	}
	for _, tv := range res.Results {
		out = append(out, tmdbListItem{ID: tv.ID, Title: tv.Name, PosterPath: tv.PosterPath, BackdropPath: tv.BackdropPath, Ovw: tv.Overview, Popularity: float64(tv.Popularity)})
	}
	return out
}

// fetchMediaInfo resolves a community title through TMDB; ok=false if it cannot.
func fetchMediaInfo(mediaID int, mediaType string, score float64) (SuggestedNominee, bool) {
	nom := SuggestedNominee{MediaID: mediaID, MediaType: mediaType, Score: score, Source: SourceCommunity}
	if mediaType == string(shared.MediaTypeMovie) {
		detail, err := tmdbmodule.GetMovieByID(mediaID)
		if err != nil || detail == nil || detail.Title == "" {
			return nom, false
		}
		nom.MediaTitle, nom.PosterPath, nom.BackdropPath, nom.Overview = detail.Title, detail.PosterPath, detail.BackdropPath, detail.Overview
		return nom, true
	}
	detail, err := tmdbmodule.GetSeriesByID(mediaID)
	if err != nil || detail == nil || detail.Name == "" {
		return nom, false
	}
	nom.MediaTitle, nom.PosterPath, nom.BackdropPath, nom.Overview = detail.Name, detail.PosterPath, detail.BackdropPath, detail.Overview
	return nom, true
}

// GenerateMonthlyVote creates the monthly round for (mediaType, period) as a draft with
// nomineesPerRound suggested titles. The round has no copy: the admin writes it before publishing.
// invalid("period") if the round already opened; ErrEventExists if it already exists;
// ErrNotEnoughNominees if suggestions could not fill it.
func (s *Service) GenerateMonthlyVote(ctx context.Context, adminID uint, mediaType, period string) (AdminEventResponse, error) {
	if !validMediaType(mediaType) {
		return AdminEventResponse{}, invalid("media_type")
	}
	startsAt, endsAt, err := MonthlyWindow(period)
	if err != nil {
		return AdminEventResponse{}, invalid("period")
	}
	if !s.now().Before(startsAt) {
		return AdminEventResponse{}, invalid("period")
	}

	// Cheap pre-check; the unique index still settles concurrent requests (mapEventUniqueErr).
	var count int64
	if err := s.db.WithContext(ctx).Model(&Event{}).
		Where("type = ? AND media_type = ? AND period = ?", TypeVoteMonthly, mediaType, period).
		Count(&count).Error; err != nil {
		return AdminEventResponse{}, err
	}
	if count > 0 {
		return AdminEventResponse{}, ErrEventExists
	}

	suggested, err := s.SuggestNominees(ctx, mediaType, period)
	if err != nil {
		return AdminEventResponse{}, err
	}
	if len(suggested) < nomineesPerRound {
		return AdminEventResponse{}, ErrNotEnoughNominees
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

	return s.Create(ctx, adminID, EventInput{
		Type:          TypeVoteMonthly,
		MediaType:     &mediaType,
		Period:        &period,
		StartsAt:      startsAt,
		EndsAt:        endsAt,
		ShowPopup:     true,
		PopupPriority: generatedPopupPriority,
		Status:        StatusDraft,
		Nominees:      nominees,
	})
}

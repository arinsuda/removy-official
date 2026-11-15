package migrate

import (
	"log"

	"gorm.io/gorm"
)

// BackfillEventPeriodsToStartMonth realigns monthly voting rounds with the current rule: a round's
// period is the calendar month (in tz) its window opens in, so round P opens on the 15th of P.
//
// Order matters. idx_events_type_media_period is a non-deferrable unique index that Postgres checks
// row by row, so a single UPDATE that swaps two periods can collide with a row not yet moved. Rows
// that change are therefore cleared to NULL first (the partial index ignores NULL) and filled in
// a second statement. Awards then copy the event's period. Every step derives from starts_at, so a
// re-run changes nothing.
func BackfillEventPeriodsToStartMonth(tx *gorm.DB, tz string) error {
	if err := tx.Exec(`
		UPDATE events SET period = NULL
		WHERE type = 'vote_monthly' AND period IS NOT NULL
		  AND period <> to_char(starts_at AT TIME ZONE ?, 'YYYY-MM')`, tz).Error; err != nil {
		return err
	}
	// Two live rounds of one media type that open in the same month would both claim one period
	// and fail the unique index, blocking boot. Only the earliest (starts_at, id) takes the period;
	// the rest keep NULL so an admin can fix them, and the count is logged instead of aborting.
	fill := tx.Exec(`
		WITH target AS (
			SELECT id, deleted_at, to_char(starts_at AT TIME ZONE ?, 'YYYY-MM') AS p,
			       ROW_NUMBER() OVER (
			           PARTITION BY media_type, to_char(starts_at AT TIME ZONE ?, 'YYYY-MM'), (deleted_at IS NULL)
			           ORDER BY starts_at, id) AS rn
			FROM events
			WHERE type = 'vote_monthly' AND period IS NULL
		)
		UPDATE events e SET period = t.p
		FROM target t
		WHERE e.id = t.id
		  AND (t.deleted_at IS NOT NULL OR (t.rn = 1 AND NOT EXISTS (
		      SELECT 1 FROM events o
		      WHERE o.type = 'vote_monthly' AND o.media_type IS NOT DISTINCT FROM e.media_type
		        AND o.period = t.p AND o.deleted_at IS NULL)))`, tz, tz)
	if fill.Error != nil {
		return fill.Error
	}
	var skipped int64
	if err := tx.Raw(`SELECT COUNT(*) FROM events
		WHERE type = 'vote_monthly' AND period IS NULL AND deleted_at IS NULL`).Scan(&skipped).Error; err != nil {
		return err
	}
	if skipped > 0 {
		log.Printf("WARN event period backfill: %d monthly round(s) left without a period (same media type and month as another round); fix them in admin", skipped)
	}
	return tx.Exec(`
		UPDATE event_awards a SET period = e.period
		FROM events e
		WHERE e.id = a.event_id AND e.type = 'vote_monthly' AND e.period IS NOT NULL
		  AND a.award_kind = 'monthly' AND a.period <> e.period`).Error
}

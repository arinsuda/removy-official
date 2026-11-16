package manga_module

import (
	"context"
	"errors"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"gorm.io/gorm"
)

// MangaAdultMinAge is the age from which a signed-in viewer may see adult titles.
const MangaAdultMinAge = 18

// Scopes separate cache entries whose content depends on the viewer's eligibility.
const (
	scopeAdult = "adult"
	scopeSFW   = "sfw"
	// detailScope is used for single-title and recommendation lookups: AniList is asked for adult
	// titles, so the cached payload is the same for everyone and the viewer gate runs after the cache.
	detailScope = scopeAdult
)

// AdultGate decides whether a viewer may see adult titles. Guests (userID 0) are never eligible.
type AdultGate interface {
	Eligible(ctx context.Context, userID uint) (bool, error)
}

// dobGate reads users.date_of_birth for the signed-in viewer.
type dobGate struct {
	db  *gorm.DB
	now func() time.Time
}

// NewDOBGate returns the production gate: eligible when the account has a birth date and is at least MangaAdultMinAge.
func NewDOBGate(db *gorm.DB) AdultGate {
	return dobGate{db: db, now: func() time.Time { return time.Now().In(config.AppLocation) }}
}

func (g dobGate) Eligible(ctx context.Context, userID uint) (bool, error) {
	if userID == 0 {
		return false, nil
	}
	var row struct {
		DateOfBirth *time.Time
	}
	err := g.db.WithContext(ctx).Table("users").
		Select("date_of_birth").
		Where("id = ? AND deleted_at IS NULL", userID).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return eligibleAt(row.DateOfBirth, g.now()), nil
}

// eligibleAt reports whether someone born on dob has reached MangaAdultMinAge by now. Both instants are
// reduced to calendar days in config.AppLocation, so the unlock happens on the same day for everyone.
// A missing birth date is never eligible. A 29 February birthday falls on 1 March in non-leap years.
func eligibleAt(dob *time.Time, now time.Time) bool {
	if dob == nil {
		return false
	}
	born := dob.In(config.AppLocation)
	eighteenth := time.Date(born.Year()+MangaAdultMinAge, born.Month(), born.Day(), 0, 0, 0, 0, config.AppLocation)
	today := now.In(config.AppLocation)
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, config.AppLocation)
	return !todayStart.Before(eighteenth)
}

// scopeOf names the cache scope for a list request.
func scopeOf(adult bool) string {
	if adult {
		return scopeAdult
	}
	return scopeSFW
}

package event_module

import (
	"time"
	// Embeds the IANA database so votingLocation resolves on hosts without zoneinfo.
	_ "time/tzdata"
)

// Voting round rules. Change them here; nothing else hard-codes these values.
const (
	// nomineesPerRound is how many titles an auto-generated round offers.
	nomineesPerRound = 5
	// communityNomineeSlots is how many of those come from POPCORN activity;
	// the rest are new releases (then popular titles) from TMDB.
	communityNomineeSlots = 3
	// communityCandidatePool is how many top community titles are fetched before
	// TMDB lookups weed out ones that no longer resolve.
	communityCandidatePool = 10
	// maxNomineesPerRound bounds an admin-edited nominee list.
	maxNomineesPerRound = 20

	// Community score = reviews*w + library adds*w + likes*w over the lookback window.
	reviewScoreWeight  = 5.0
	libraryScoreWeight = 3.0
	likeScoreWeight    = 2.0

	// votingCutoffDay is the day of month a monthly round closes (and the next opens),
	// at 00:00 in votingTimeZone.
	votingCutoffDay = 15
	// activityLookbackMonths: nominees reflect activity in the month before voting opens.
	activityLookbackMonths = 1
	// votingTimeZone is where round boundaries fall on local midnight (Thailand).
	votingTimeZone = "Asia/Bangkok"

	// generatedPopupPriority places auto-generated rounds above routine announcements.
	generatedPopupPriority = 10

	monthlyPeriodLayout = "2006-01"
	yearlyPeriodLayout  = "2006"

	// tmdbFirstPage is the only TMDB list page suggestion lookups read.
	tmdbFirstPage = 1
)

// AwardKind labels what a winner won.
type AwardKind string

const (
	AwardKindMonthly AwardKind = "monthly"
	AwardKindYearly  AwardKind = "yearly"
)

// NomineeSource explains where a suggested nominee came from (shown to admins).
type NomineeSource string

const (
	SourceCommunity NomineeSource = "community"
	SourceRelease   NomineeSource = "release"
	SourcePopular   NomineeSource = "popular"
)

// votingLocation is loaded once; tzdata is embedded so this cannot fail at runtime.
var votingLocation = mustLoadLocation(votingTimeZone)

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic("event_module: cannot load time zone " + name + ": " + err.Error())
	}
	return loc
}

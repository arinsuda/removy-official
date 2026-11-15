package event_module

import (
	"time"
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
	// minNomineesToPublish is the fewest titles a voting round needs before it can be published.
	minNomineesToPublish = 2

	// VoteRateMax and VoteRateWindow bound how often one signed-in user may cast or withdraw a vote.
	// Votes are idempotent upserts, so the budget only stops scripted churn.
	VoteRateMax    = 30
	VoteRateWindow = time.Minute

	// Community score = reviews*w + library adds*w + likes*w + view days*w over the lookback window.
	// A view counts once per user, title and day, so one habitual viewer cannot dominate a title.
	reviewScoreWeight  = 5.0
	libraryScoreWeight = 3.0
	likeScoreWeight    = 2.0
	viewScoreWeight    = 0.1

	// votingCutoffDay is the day of month a monthly round opens (and the previous one closes),
	// at 00:00 in config.AppTimeZone.
	votingCutoffDay = 15
	// activityLookbackMonths: nominees reflect activity in the month before voting opens.
	activityLookbackMonths = 1
	// generatedPopupPriority places auto-generated rounds above routine announcements.
	generatedPopupPriority = 10

	// Yearly round Y (the year awarded) opens 16 January of Y+1 and closes 15 February of Y+1 at
	// 00:00 in config.AppTimeZone, after December's monthly round (closes 15 January) has settled.
	yearlyStartMonth = time.January
	yearlyStartDay   = 16
	yearlyEndMonth   = time.February
	yearlyEndDay     = 15

	// autoGenerateLead is how long before a round opens the scheduler creates its draft.
	autoGenerateLead = 72 * time.Hour
	// autoRoundInterval is how often the scheduler looks for rounds to create and publish.
	autoRoundInterval = 5 * time.Minute
	// autoPublishLead is how long before a round opens the scheduler publishes its auto draft.
	autoPublishLead = autoRoundInterval
	// autoRoundLockPrefix namespaces the advisory lock that serializes auto round creation.
	autoRoundLockPrefix = "event_auto_round"
	// autoRoundBatch caps drafts published per tick.
	autoRoundBatch = 50

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

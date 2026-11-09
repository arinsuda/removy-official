package privacy_policy

import (
	"sort"
	"strings"
)

// activityDefaultEnabled is the single source of truth for whether an activity type is shown
// when the actor never saved a preference. Types missing here default to disabled.
var activityDefaultEnabled = map[ActivityType]bool{
	ActivityReviewCreated:       true,
	ActivityReviewCommented:     true,
	ActivityReviewLiked:         false,
	ActivityMediaLiked:          false,
	ActivityWatchlistAdded:      false,
	ActivityWatchedAdded:        false,
	ActivityAchievementUnlocked: true,
	ActivityUserFollowed:        false,
	ActivityWatchLogCreated:     false,
	ActivityRewatched:           false,
}

// ActivityEnabledByDefault reports the default for t when no preference row exists.
func ActivityEnabledByDefault(t ActivityType) bool { return activityDefaultEnabled[t] }

// ActivityEnabledSQL returns a SQL boolean expression for feed queries: the actor's saved
// preference (alias aps) or, without one, the default for the event type in typeColumn.
// Only compile-time constants are interpolated.
func ActivityEnabledSQL(typeColumn string) string {
	var on []string
	for t, enabled := range activityDefaultEnabled {
		if enabled {
			on = append(on, "'"+string(t)+"'")
		}
	}
	sort.Strings(on)
	return "(aps.enabled IS NULL AND " + typeColumn + " IN (" + strings.Join(on, ", ") + ")) OR aps.enabled = true"
}

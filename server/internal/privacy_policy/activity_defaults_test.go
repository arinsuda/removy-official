package privacy_policy

import "testing"

func TestActivityDefaults(t *testing.T) {
	if !ActivityEnabledByDefault(ActivityReviewCreated) || ActivityEnabledByDefault(ActivityRewatched) {
		t.Fatal("unexpected defaults")
	}
	want := "(aps.enabled IS NULL AND ae.type IN ('achievement_unlocked', 'review_commented', 'review_created')) OR aps.enabled = true"
	if got := ActivityEnabledSQL("ae.type"); got != want {
		t.Fatalf("got %q", got)
	}
}

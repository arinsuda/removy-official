package privacy_policy

type ProfileSection string

const (
	SectionProfile      ProfileSection = "profile"
	SectionReviews      ProfileSection = "reviews"
	SectionLibrary      ProfileSection = "library"
	SectionAchievements ProfileSection = "achievements"
	SectionFeed         ProfileSection = "feed"
)

type ActivityVisibility string

const (
	VisibilityDefault   ActivityVisibility = "default"
	VisibilityPublic    ActivityVisibility = "public"
	VisibilityFollowers ActivityVisibility = "followers"
	VisibilityPrivate   ActivityVisibility = "private"
)

type ActivityType string

const (
	ActivityReviewCreated       ActivityType = "review_created"
	ActivityReviewCommented     ActivityType = "review_commented"
	ActivityReviewLiked         ActivityType = "review_liked"
	ActivityMediaLiked          ActivityType = "media_liked"
	ActivityWatchlistAdded      ActivityType = "watchlist_added"
	ActivityWatchedAdded        ActivityType = "watched_added"
	ActivityWatchLogCreated     ActivityType = "watch_log_created"
	ActivityRewatched           ActivityType = "rewatched"
	ActivityAchievementUnlocked ActivityType = "achievement_unlocked"
	ActivityUserFollowed        ActivityType = "user_followed"
)

const (
	EventFeedRefresh         = "feed:refresh_required"
	EventFeedActivityUpdated = "feed:activity_updated"
	EventFeedActivityRemoved = "feed:activity_removed"
)

// ListVisibility is who may open a user-made list (on top of the owner's profile privacy).
type ListVisibility string

const (
	ListPublic    ListVisibility = "public"
	ListFollowers ListVisibility = "followers"
	ListPrivate   ListVisibility = "private"
)

// Valid reports whether v is one of the known list visibilities.
func (v ListVisibility) Valid() bool {
	return v == ListPublic || v == ListFollowers || v == ListPrivate
}

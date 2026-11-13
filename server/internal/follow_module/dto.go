package follow_module

type FollowResponse struct {
	FollowerID uint   `json:"follower_id"`
	FolloweeID uint   `json:"followee_id"`
	Status     string `json:"status"`
}

type FollowStatsResponse struct {
	UserID      uint  `json:"user_id"`
	Followers   int64 `json:"followers"`
	Following   int64 `json:"following"`
	IsFollowing bool  `json:"is_following"`
}

type UserSummary struct {
	ID          uint    `json:"id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
	Status      string  `json:"status,omitempty"`
}

type RelationshipStatus struct {
	IsFollowing  bool   `json:"is_following"`
	FollowStatus string `json:"follow_status,omitempty"`
	IsFollowedBy bool   `json:"is_followed_by"`
}

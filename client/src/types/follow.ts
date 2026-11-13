import type { UserSummary } from "./user"
import type { MediaSummary } from "./movie"

export interface FollowStatsResponse {
  user_id: number
  followers: number
  following: number
  is_following: boolean
}

export interface FollowRelationshipStatus {
  is_following: boolean
  follow_status?: "pending" | "accepted"
  is_followed_by: boolean
}

export interface FollowActionResponse {
  follower_id: number
  followee_id: number
  status: "pending" | "accepted"
}

export interface FollowUserSummary {
  id: number
  username: string
  display_name: string | null
  avatar_url: string | null
  status?: "pending" | "accepted"
}

export interface FeedReviewResponse {
  id: number
  user: UserSummary
  media: MediaSummary
  rating: number
  body: string
  like_count: number
  comment_count: number
  is_liked: boolean
  created_at: string
}

export interface FeedItemResponse {
  review: FeedReviewResponse
  created_at: string
}

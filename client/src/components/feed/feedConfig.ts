import type { ActivityType, FeedScope } from "@/types/feed"

/** Feed categories understood by GET /feed; "" means all. */
export const FEED_CATEGORIES = ["", "reviews", "lists", "social"] as const

/** Feed scopes understood by GET /feed. */
export const FEED_SCOPES: FeedScope[] = ["all", "following", "me"]
export const DEFAULT_FEED_SCOPE: FeedScope = "all"

/** The API reports a follow target that no longer exists with id 0. */
export const DELETED_USER_ID = 0

/** Reviews longer than this get a clamp with a read-more toggle. */
export const REVIEW_EXCERPT_CHARS = 180

export type Visibility = "default" | "public" | "followers" | "private"
export const VISIBILITY_OPTIONS: Visibility[] = ["default", "public", "followers", "private"]

/** Activity types that carry a review the viewer can like. */
export const LIKEABLE_TYPES: ActivityType[] = ["review_created"]

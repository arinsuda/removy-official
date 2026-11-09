import type { ActivityType } from "@/types/feed"

/** Feed categories understood by GET /feed; "" means all. */
export const FEED_CATEGORIES = ["", "reviews", "lists", "social"] as const

/** Reviews longer than this get a clamp with a read-more toggle. */
export const REVIEW_EXCERPT_CHARS = 180

export type Visibility = "default" | "public" | "followers" | "private"
export const VISIBILITY_OPTIONS: Visibility[] = ["default", "public", "followers", "private"]

/** Activity types that carry a review the viewer can like. */
export const LIKEABLE_TYPES: ActivityType[] = ["review_created"]

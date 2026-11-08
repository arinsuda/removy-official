import type { MediaType, Visibility } from "./common"
import type { Genre } from "./movie"

export interface ShareMediaContext {
  id: number
  mediaType: MediaType
  title: string
  posterPath: string | null
  posterUrl: string | null
  releaseYear: string | null
  genres: Genre[]
  voteAverage: number
  popcornRating: number | null
  /** Review count behind popcornRating; used for the tier. Optional for older callers. */
  popcornCount?: number | null
}

export interface ShareReviewContext {
  id: number
  authorDisplayName: string
  authorUsername: string
  authorAvatarUrl: string | null
  rating: number
  body: string
  visibility: Visibility
  createdAt: string
}

/** The signed-in user's own activity on the title. Build it only for the owner. */
export interface ShareStatsContext {
  rating: number | null
  watchCount: number
  lastWatchedOn: string | null
  reviewBody: string | null
  reviewVisibility: Visibility | null
}

export type ShareTemplateKey = "poster" | "take" | "sticker" | "square" | "review"
export type ShareTheme = "dark" | "light" | "brand"

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
  /** Manga genres arrive as names only (AniList has no genre ids); shown when `genres` is empty. */
  genreNames?: string[]
  voteAverage: number
  popcornRating: number | null
  /** Review count behind popcornRating; used for the tier. Optional for older callers. */
  popcornCount?: number | null
  /** Adult manga: the poster is replaced by the fallback card unless the viewer may see adult titles. */
  isAdult?: boolean
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

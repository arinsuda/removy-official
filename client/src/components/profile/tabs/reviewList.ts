import type { ReviewResponse } from "@/types"

/** Rating chip thresholds; must equal HighRatingMin / LowRatingMax in server review_module/dto.go. */
export const HIGH_RATING_MIN = 4
export const LOW_RATING_MAX = 2.5
/** Delay before a typed search is sent to the server. */
export const REVIEW_SEARCH_DEBOUNCE_MS = 350

export const REVIEW_FILTERS: Record<string, (r: ReviewResponse) => boolean> = {
  high: (r) => r.rating >= HIGH_RATING_MIN,
  low: (r) => r.rating <= LOW_RATING_MAX,
  withText: (r) => r.body.trim() !== "",
}

export const REVIEW_SORT_KEYS = ["newest", "oldest", "highest", "lowest"] as const

export interface ReviewSummary {
  count: number
  /** Mean rating rounded to 1 decimal; 0 when there are no reviews. */
  average: number
  /** Genre that appears in the most reviewed titles, or null. */
  topGenre: string | null
}

export function summarizeReviews(reviews: ReviewResponse[]): ReviewSummary {
  if (!reviews.length) return { count: 0, average: 0, topGenre: null }
  const average = Math.round((reviews.reduce((sum, r) => sum + r.rating, 0) / reviews.length) * 10) / 10
  const genres = new Map<string, number>()
  for (const r of reviews) for (const g of r.media.genres ?? []) genres.set(g.name, (genres.get(g.name) ?? 0) + 1)
  // Ties resolve to the genre seen first (Map keeps insertion order).
  let topGenre: string | null = null
  let best = 0
  for (const [name, n] of genres) {
    if (n > best) {
      best = n
      topGenre = name
    }
  }
  return { count: reviews.length, average, topGenre }
}

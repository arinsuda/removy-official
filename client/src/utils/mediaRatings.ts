import type { MediaRatings } from "@/types/movie"

interface LiveStats {
  average_rating: number
  review_count: number
  has_rating: boolean
}

/**
 * Detail pages load ratings once with the title, but the media-stats call is
 * refreshed after reviews change. Prefer the live POPCORN numbers when present.
 */
export function withLiveStats(ratings: MediaRatings | null | undefined, stats: LiveStats): MediaRatings | null {
  if (!ratings) return null
  // No reviews left: show POPCORN 0.0 rather than a stale server value.
  if (!stats.has_rating) return { ...ratings, popcorn: { average: 0, count: 0, available: false, scale: 5 } }
  return {
    ...ratings,
    popcorn: { average: stats.average_rating, count: stats.review_count, available: true, scale: 5 },
  }
}

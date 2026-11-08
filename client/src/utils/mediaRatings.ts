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
  if (!stats.has_rating) return ratings
  return {
    ...ratings,
    popcorn: { average: stats.average_rating, count: stats.review_count, available: true, scale: 5 },
  }
}

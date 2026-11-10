import type { RatingBucket, ReviewResponse } from "@/types/review"

export interface DistributionBucket {
  /** Rating value, 0.5 to 5.0. */
  value: number
  count: number
  /** Share of the tallest bucket (0-100), so the biggest bar is always full. */
  pct: number
}

export interface RatingDistributionResult {
  total: number
  buckets: DistributionBucket[]
}

const RATING_STEPS = 10

function toDistributionResult(counts: number[], total: number): RatingDistributionResult {
  const max = Math.max(...counts, 0)
  return {
    total,
    buckets: counts.map((count, i) => ({
      value: (i + 1) / 2,
      count,
      pct: max ? Math.round((count / max) * 100) : 0,
    })),
  }
}

/**
 * Histogram of ratings 0.5-5.0 from the server's pre-aggregated 10-step buckets.
 */
export function ratingDistributionFromBuckets(buckets: RatingBucket[]): RatingDistributionResult {
  const counts = new Array<number>(RATING_STEPS).fill(0)
  let total = 0
  for (const b of buckets) {
    if (b.count <= 0) continue
    const idx = Math.round(b.rating * 2) - 1
    if (idx < 0 || idx >= RATING_STEPS) continue
    counts[idx] = (counts[idx] ?? 0) + b.count
    total += b.count
  }
  return toDistributionResult(counts, total)
}

/**
 * Histogram of ratings 0.5-5.0 from the visible PUBLIC reviews only. Followers
 * and private reviews are left out so the chart matches the POPCORN average.
 */
export function ratingDistribution(
  reviews: Pick<ReviewResponse, "rating" | "visibility">[],
  options: { /** Count every review regardless of visibility (profile views). */ includeAll?: boolean } = {},
): RatingDistributionResult {
  const counts = new Array<number>(RATING_STEPS).fill(0)
  let total = 0
  for (const r of reviews) {
    if (!options.includeAll && r.visibility !== "public") continue
    const idx = Math.round(r.rating * 2) - 1
    if (idx < 0 || idx >= RATING_STEPS) continue
    counts[idx] = (counts[idx] ?? 0) + 1
    total++
  }
  return toDistributionResult(counts, total)
}

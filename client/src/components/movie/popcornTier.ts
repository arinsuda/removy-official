export type PopcornTier = "none" | "stale" | "fresh" | "hot"

export const POPCORN_TIER_THRESHOLDS = { fresh: 2.5, hot: 4 } as const

/** Round to the 1 decimal shown in the UI so the tier always matches the visible number. */
export function roundScore(avg: number | null | undefined): number {
  return avg && Number.isFinite(avg) ? Math.round(avg * 10) / 10 : 0
}

/**
 * Visual tier of a POPCORN score (0-5).
 * none: no reviews; stale: below 2.5; fresh: 2.5-3.9; hot: 4.0 and above.
 */
export function popcornTier(avg: number | null | undefined, count: number | null | undefined): PopcornTier {
  const score = roundScore(avg)
  if (!count || count <= 0 || score <= 0) return "none"
  if (score < POPCORN_TIER_THRESHOLDS.fresh) return "stale"
  if (score < POPCORN_TIER_THRESHOLDS.hot) return "fresh"
  return "hot"
}

import type { ReviewResponse } from "@/types/review"
import { ratingDistribution, type RatingDistributionResult } from "@/components/detail/ratingDistribution"

export type TastePersona = "generous" | "balanced" | "tough" | "new"

export interface TasteSummary {
  /** Mean rating, or null when there are no reviews. */
  average: number | null
  persona: TastePersona
  topGenres: { id: number; name: string; count: number }[]
  distribution: RatingDistributionResult
}

/** Below this many reviews a taste label would be noise. */
export const MIN_REVIEWS_FOR_PERSONA = 3
const GENEROUS_AT = 4
const TOUGH_AT = 3

type TasteReview = Pick<ReviewResponse, "rating" | "visibility"> & {
  media?: { genres?: { id: number; name: string }[] }
}

export function tasteSummary(reviews: TasteReview[]): TasteSummary {
  const count = reviews.length
  const average = count ? reviews.reduce((sum, r) => sum + r.rating, 0) / count : null

  let persona: TastePersona = "new"
  if (average !== null && count >= MIN_REVIEWS_FOR_PERSONA) {
    persona = average >= GENEROUS_AT ? "generous" : average < TOUGH_AT ? "tough" : "balanced"
  }

  const genres = new Map<number, { id: number; name: string; count: number }>()
  for (const r of reviews) {
    for (const g of r.media?.genres ?? []) {
      const entry = genres.get(g.id) ?? { id: g.id, name: g.name, count: 0 }
      entry.count++
      genres.set(g.id, entry)
    }
  }
  const topGenres = [...genres.values()].sort((a, b) => b.count - a.count || a.name.localeCompare(b.name)).slice(0, 3)

  return { average, persona, topGenres, distribution: ratingDistribution(reviews, { includeAll: true }) }
}

import { popcornTier, type PopcornTier } from "@/components/movie/popcornTier"
import type { HomeMedia } from "./homeMedia"

export const HERO_MIN_HOT = 3
export const HERO_MAX = 5

export function mediaTier(item: HomeMedia): PopcornTier {
  const p = item.ratings?.popcorn
  return popcornTier(p?.average, p?.count)
}

/** First occurrence wins; the same title can arrive from several TMDB lists. */
export function dedupeMedia(items: HomeMedia[]): HomeMedia[] {
  const seen = new Set<string>()
  return items.filter((m) => !seen.has(m.key) && !!seen.add(m.key))
}

/** Interleave lists so movies and TV alternate instead of one type dominating. */
export function interleave(...lists: HomeMedia[][]): HomeMedia[] {
  const out: HomeMedia[] = []
  const longest = Math.max(0, ...lists.map((l) => l.length))
  for (let i = 0; i < longest; i++) {
    for (const l of lists) {
      const item = l[i]
      if (item) out.push(item)
    }
  }
  return dedupeMedia(out)
}

/** Items of one tier. Hot and fresh sort by score then review count; others keep list order. */
export function filterByTier(items: HomeMedia[], tier: PopcornTier, limit = 20): HomeMedia[] {
  const matches = items.filter((m) => mediaTier(m) === tier)
  if (tier === "hot" || tier === "fresh") {
    const score = (m: HomeMedia) => m.ratings?.popcorn?.average ?? 0
    const count = (m: HomeMedia) => m.ratings?.popcorn?.count ?? 0
    matches.sort((a, b) => score(b) - score(a) || count(b) - count(a))
  }
  return matches.slice(0, limit)
}

export function countByTier(items: HomeMedia[]): Record<PopcornTier, number> {
  const counts: Record<PopcornTier, number> = { none: 0, stale: 0, fresh: 0, hot: 0 }
  for (const m of items) counts[mediaTier(m)]++
  return counts
}

export interface HeroPick {
  items: HomeMedia[]
  /** "hot" when the slides are ป๊อปร้อน titles, "popular" for the TMDB fallback. */
  source: "hot" | "popular"
}

/**
 * Hero slides: ป๊อปร้อน titles (with a backdrop) when at least HERO_MIN_HOT
 * qualify, otherwise the first HERO_MAX of the TMDB popular fallback.
 */
export function pickHeroItems(pool: HomeMedia[], fallback: HomeMedia[]): HeroPick {
  const hot = filterByTier(dedupeMedia(pool), "hot", Infinity).filter((m) => m.backdrop_path)
  if (hot.length >= HERO_MIN_HOT) return { items: hot.slice(0, HERO_MAX), source: "hot" }
  return { items: fallback.slice(0, HERO_MAX), source: "popular" }
}

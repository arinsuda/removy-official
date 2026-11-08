import type { HomeMedia } from "@/components/home/homeMedia"
import { mediaTier } from "@/components/home/popcornHome"
import type { BrowseTierFilter } from "./browseQuery"

export type BrowseSort = "default" | "popcorn" | "tmdb" | "date"
export const BROWSE_SORTS: BrowseSort[] = ["default", "popcorn", "tmdb", "date"]

const popcornScore = (m: HomeMedia) => m.ratings?.popcorn?.average ?? 0
const popcornCount = (m: HomeMedia) => m.ratings?.popcorn?.count ?? 0

const comparators: Record<Exclude<BrowseSort, "default">, (a: HomeMedia, b: HomeMedia) => number> = {
  // Unrated titles (no reviews) always sink below rated ones; ties break on review count.
  popcorn: (a, b) =>
    Number(popcornCount(b) > 0) - Number(popcornCount(a) > 0) ||
    popcornScore(b) - popcornScore(a) ||
    popcornCount(b) - popcornCount(a),
  tmdb: (a, b) => (b.vote_average || 0) - (a.vote_average || 0),
  // Missing dates last.
  date: (a, b) => Number(!!b.date) - Number(!!a.date) || b.date.localeCompare(a.date),
}

/** Sorted copy. "default" keeps the TMDB list order; Array.sort is stable so ties keep it too. */
export function sortMedia(items: HomeMedia[], sort: BrowseSort): HomeMedia[] {
  if (sort === "default") return items
  return [...items].sort(comparators[sort])
}

export function filterTier(items: HomeMedia[], tier: BrowseTierFilter): HomeMedia[] {
  return tier === "all" ? items : items.filter((m) => mediaTier(m) === tier)
}

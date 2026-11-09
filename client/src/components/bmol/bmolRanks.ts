import type { BMOLItemResponse } from "@/types"

/** Ranks beyond this fold away behind "Show all" on the shelf. */
export const SHELF_VISIBLE_RANK = 10

export type BmolTier = "gold" | "silver" | "bronze" | "shelf"

export interface RankGroup {
  rank: number
  items: BMOLItemResponse[]
}

/** Several titles may share a rank (a tie bucket). Groups are sorted by rank, then by insertion. */
export function groupByRank(items: BMOLItemResponse[]): RankGroup[] {
  const map = new Map<number, BMOLItemResponse[]>()
  for (const item of items) {
    const list = map.get(item.rank)
    if (list) list.push(item)
    else map.set(item.rank, [item])
  }
  // Legacy rows with rank < 1 sort last so they stay reachable at the end of the list.
  const order = (rank: number) => (rank < 1 ? Number.POSITIVE_INFINITY : rank)
  return [...map.entries()].sort((a, b) => order(a[0]) - order(b[0])).map(([rank, list]) => ({ rank, items: list }))
}

export function tierOf(rank: number): BmolTier {
  if (rank === 1) return "gold"
  if (rank === 2) return "silver"
  if (rank === 3) return "bronze"
  return "shelf"
}

/** Smallest positive rank that nobody uses yet. */
export function nextFreeRank(items: BMOLItemResponse[]): number {
  const used = new Set(items.map((i) => i.rank))
  let rank = 1
  while (used.has(rank)) rank++
  return rank
}

/**
 * Rank an item lands on when moved one step. Up joins the previous group;
 * down joins the next group, or opens a new rank after the last group when
 * the item is not alone there. Returns null when there is nowhere to go.
 */
export function moveTarget(groups: RankGroup[], rank: number, direction: "up" | "down"): number | null {
  const index = groups.findIndex((g) => g.rank === rank)
  if (index === -1) return null
  if (direction === "up") return index > 0 ? groups[index - 1]!.rank : null
  const next = groups[index + 1]
  // A legacy rank < 1 is never a destination, and "down" never moves upward.
  if (next && next.rank >= 1) return next.rank
  return rank >= 1 && rank < 999 && groups[index]!.items.length > 1 ? rank + 1 : null
}

/** Items in display order (rank, then insertion), flattened from the groups. */
export function flattenGroups(groups: RankGroup[]): BMOLItemResponse[] {
  return groups.flatMap((g) => g.items)
}

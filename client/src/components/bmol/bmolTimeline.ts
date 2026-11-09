import type { BMOLItemResponse } from "@/types"

/** Bucket for titles without a usable release date; always rendered last. */
export const UNKNOWN_YEAR = "unknown" as const

export interface YearGroup {
  year: number | typeof UNKNOWN_YEAR
  items: BMOLItemResponse[]
}

const YEAR = /^(\d{4})-/

export function releaseYear(item: BMOLItemResponse): number | null {
  const m = YEAR.exec(item.media.release_date ?? "")
  return m ? Number(m[1]) : null
}

/**
 * Groups titles by release year. Years run newest first; unknown dates come last.
 * Inside a year, titles keep their BMOL rank order (then release date).
 */
export function groupByReleaseYear(items: BMOLItemResponse[], order: "desc" | "asc" = "desc"): YearGroup[] {
  const byYear = new Map<number | typeof UNKNOWN_YEAR, BMOLItemResponse[]>()
  for (const item of items) {
    const key = releaseYear(item) ?? UNKNOWN_YEAR
    const list = byYear.get(key)
    if (list) list.push(item)
    else byYear.set(key, [item])
  }
  const sign = order === "desc" ? -1 : 1
  const rankOrder = (r: number) => (r < 1 ? Number.POSITIVE_INFINITY : r)
  const within = (a: BMOLItemResponse, b: BMOLItemResponse) =>
    rankOrder(a.rank) - rankOrder(b.rank) || (a.media.release_date ?? "").localeCompare(b.media.release_date ?? "")
  const groups: YearGroup[] = []
  for (const [year, list] of byYear) groups.push({ year, items: [...list].sort(within) })
  return groups.sort((a, b) => {
    if (a.year === UNKNOWN_YEAR) return 1
    if (b.year === UNKNOWN_YEAR) return -1
    return sign * (a.year - b.year)
  })
}

import type { HomeMedia } from "@/components/home/homeMedia"

export type UpcomingGroupKind = "out" | "week" | "next" | "month"

export interface UpcomingGroup {
  /** Stable DOM/rail id suffix, e.g. "out", "week", "m-2026-12". */
  id: string
  kind: UpcomingGroupKind
  /** Localized month name with year; only set for kind "month". */
  monthLabel?: string
  items: HomeMedia[]
}

const DAY_MS = 86_400_000

/** Parse "YYYY-MM-DD" as a local calendar day (new Date(str) would shift it by timezone). */
export function parseLocalDate(value: string): Date | null {
  const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(value)
  if (!m) return null
  const d = new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]))
  return Number.isNaN(d.getTime()) ? null : d
}

export const startOfDay = (d: Date) => new Date(d.getFullYear(), d.getMonth(), d.getDate())

/** Whole calendar days from today to date (negative when past); DST-safe. */
export function daysUntil(date: Date, today: Date): number {
  const utc = (d: Date) => Date.UTC(d.getFullYear(), d.getMonth(), d.getDate())
  return Math.round((utc(date) - utc(today)) / DAY_MS)
}

/** The Monday that opens next week (weeks run Monday to Sunday). */
function nextMonday(today: Date): Date {
  const sinceMonday = (today.getDay() + 6) % 7
  return new Date(today.getFullYear(), today.getMonth(), today.getDate() - sinceMonday + 7)
}

/**
 * Buckets titles by release date: already out (before today), this week, next week
 * (weeks start Monday), then one group per month. Groups keep date order and empty
 * groups are omitted. Titles without a valid date are dropped.
 */
export function groupByRelease(items: HomeMedia[], today: Date, locale: string): UpcomingGroup[] {
  const day0 = startOfDay(today)
  const mondayNext = nextMonday(day0)
  const mondayAfter = new Date(mondayNext.getFullYear(), mondayNext.getMonth(), mondayNext.getDate() + 7)
  const monthName = new Intl.DateTimeFormat(locale, { month: "long", year: "numeric" })

  const dated = items
    .map((item) => ({ item, date: parseLocalDate(item.date) }))
    .filter((e): e is { item: HomeMedia; date: Date } => e.date !== null)
    .sort((a, b) => a.date.getTime() - b.date.getTime())

  const groups: UpcomingGroup[] = []
  const byId = new Map<string, UpcomingGroup>()
  for (const { item, date } of dated) {
    let group: Omit<UpcomingGroup, "items">
    if (date < day0) group = { id: "out", kind: "out" }
    else if (date < mondayNext) group = { id: "week", kind: "week" }
    else if (date < mondayAfter) group = { id: "next", kind: "next" }
    else {
      const mm = String(date.getMonth() + 1).padStart(2, "0")
      group = { id: `m-${date.getFullYear()}-${mm}`, kind: "month", monthLabel: monthName.format(date) }
    }
    let target = byId.get(group.id)
    if (!target) {
      target = { ...group, items: [] }
      byId.set(group.id, target)
      groups.push(target)
    }
    target.items.push(item)
  }
  return groups
}

/** Release window (days after the first upcoming release) the hero picks its title from. */
const HERO_WINDOW_DAYS = 14

/**
 * The most popular title among those releasing soonest. Out-now and backdrop-less
 * titles are skipped so the hero always counts down to a real, imageable release.
 */
export function pickHero(items: HomeMedia[], today: Date): HomeMedia | null {
  const day0 = startOfDay(today)
  const candidates = items
    .map((item) => ({ item, date: parseLocalDate(item.date) }))
    .filter((e): e is { item: HomeMedia; date: Date } => e.date !== null && e.date >= day0 && !!e.item.backdrop_path)
  if (!candidates.length) return null
  const first = Math.min(...candidates.map((e) => e.date.getTime()))
  const limit = first + HERO_WINDOW_DAYS * DAY_MS
  const popularity = (m: HomeMedia) => ("popularity" in m.source ? (m.source.popularity ?? 0) : 0)
  return candidates
    .filter((e) => e.date.getTime() <= limit)
    .reduce((best, e) => (popularity(e.item) > popularity(best.item) ? e : best)).item
}

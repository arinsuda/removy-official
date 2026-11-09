import type { ProfileEntry } from "./profileMedia"

export interface WatchedEntry extends ProfileEntry {
  /** Times watched; only known for the owner (watch logs are private). */
  count?: number
  /** The profile owner's review rating for this title, if public to the viewer. */
  rating?: number
}

export interface MonthGroup {
  /** "YYYY-MM" in local time. */
  key: string
  label: string
  items: WatchedEntry[]
}

export function monthKey(iso: string): string {
  const d = new Date(iso)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}`
}

/** One entry per title; when a title appears twice the most recent watch wins. */
export function dedupeWatched<T extends ProfileEntry>(entries: T[]): T[] {
  const latest = new Map<string, T>()
  for (const e of entries) {
    const prev = latest.get(e.media.key)
    if (!prev || new Date(e.at).getTime() > new Date(prev.at).getTime()) latest.set(e.media.key, e)
  }
  return [...latest.values()]
}

export function yearsOf(entries: ProfileEntry[]): number[] {
  return [...new Set(entries.map((e) => new Date(e.at).getFullYear()))].sort((a, b) => b - a)
}

export function inYear(entries: WatchedEntry[], year: number): WatchedEntry[] {
  return entries.filter((e) => new Date(e.at).getFullYear() === year)
}

/** Groups consecutive entries that share a month, keeping the incoming (sorted) order. */
export function groupByMonth(entries: WatchedEntry[], locale: string): MonthGroup[] {
  const fmt = new Intl.DateTimeFormat(locale, { month: "long", year: "numeric" })
  const groups: MonthGroup[] = []
  for (const e of entries) {
    const key = monthKey(e.at)
    const last = groups[groups.length - 1]
    if (last?.key === key) last.items.push(e)
    else groups.push({ key, label: fmt.format(new Date(e.at)), items: [e] })
  }
  return groups
}

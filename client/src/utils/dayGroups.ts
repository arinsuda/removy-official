export type DayBucket = "today" | "yesterday" | "earlier"

export interface DayGroup<T> {
  bucket: DayBucket
  items: T[]
}

const startOfDay = (d: Date) => new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime()

export function dayBucket(iso: string, now = new Date()): DayBucket {
  const dayDiff = Math.round((startOfDay(now) - startOfDay(new Date(iso))) / 86_400_000)
  if (dayDiff <= 0) return "today"
  return dayDiff === 1 ? "yesterday" : "earlier"
}

/** Groups items (already newest first) into Today / Yesterday / Earlier, dropping empty groups. */
export function groupByDay<T>(items: T[], dateOf: (item: T) => string, now = new Date()): DayGroup<T>[] {
  const groups: DayGroup<T>[] = []
  for (const item of items) {
    const bucket = dayBucket(dateOf(item), now)
    const last = groups[groups.length - 1]
    if (last && last.bucket === bucket) last.items.push(item)
    else groups.push({ bucket, items: [item] })
  }
  return groups
}

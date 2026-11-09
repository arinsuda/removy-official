export type ReleaseStatus = "released" | "upcoming" | "unknown"

const DAY = /^(\d{4})-(\d{2})-(\d{2})/

/** Local calendar day as YYYY-MM-DD, so "today" follows the viewer's timezone, not UTC. */
export function localDay(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, "0")
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

/** A title released today counts as released. Missing or malformed dates are unknown. */
export function releaseStatus(date: string | null | undefined, today: Date = new Date()): ReleaseStatus {
  const m = date ? DAY.exec(date) : null
  if (!m) return "unknown"
  return `${m[1]}-${m[2]}-${m[3]}` > localDay(today) ? "upcoming" : "released"
}

/** Formats a YYYY-MM-DD date in the given locale without a timezone shift; "" when unknown. */
export function formatReleaseDate(date: string | null | undefined, locale: string): string {
  const m = date ? DAY.exec(date) : null
  if (!m) return ""
  const local = new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]))
  return new Intl.DateTimeFormat(locale, { dateStyle: "medium" }).format(local)
}

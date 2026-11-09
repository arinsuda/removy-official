const UNITS: [Intl.RelativeTimeFormatUnit, number][] = [
  ["year", 60 * 60 * 24 * 365],
  ["month", 60 * 60 * 24 * 30],
  ["week", 60 * 60 * 24 * 7],
  ["day", 60 * 60 * 24],
  ["hour", 60 * 60],
  ["minute", 60],
]

/** Under this many seconds the result is "now" (Intl numeric: "auto"). */
const NOW_THRESHOLD_SECONDS = 30

/** Relative time ("3 hours ago") in the given locale; no hard-coded wording. */
export function formatRelativeTime(iso: string, locale: string, now = Date.now()): string {
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ""
  const rtf = new Intl.RelativeTimeFormat(locale, { numeric: "auto" })
  const diffSeconds = Math.round((date.getTime() - now) / 1000)

  if (Math.abs(diffSeconds) < NOW_THRESHOLD_SECONDS) return rtf.format(0, "second")

  for (const [unit, secondsInUnit] of UNITS) {
    if (Math.abs(diffSeconds) >= secondsInUnit) {
      return rtf.format(Math.round(diffSeconds / secondsInUnit), unit)
    }
  }
  return rtf.format(diffSeconds, "second")
}

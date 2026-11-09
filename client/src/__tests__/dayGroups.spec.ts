import { describe, expect, it } from "vitest"
import { dayBucket, groupByDay } from "@/utils/dayGroups"
import { formatRelativeTime } from "@/utils/relativeTime"

const now = new Date(2026, 10, 9, 12, 0, 0)
const at = (daysAgo: number, hour = 9) => new Date(2026, 10, 9 - daysAgo, hour).toISOString()

describe("dayGroups", () => {
  it("buckets by calendar day, not by 24h windows", () => {
    expect(dayBucket(at(0, 0), now)).toBe("today")
    expect(dayBucket(at(1, 23), now)).toBe("yesterday")
    expect(dayBucket(at(2), now)).toBe("earlier")
    expect(dayBucket(at(30), now)).toBe("earlier")
  })

  it("groups consecutive items and drops empty groups", () => {
    const items = [{ t: at(0) }, { t: at(0, 3) }, { t: at(2) }, { t: at(9) }]
    const groups = groupByDay(items, (i) => i.t, now)
    expect(groups.map((g) => [g.bucket, g.items.length])).toEqual([
      ["today", 2],
      ["earlier", 2],
    ])
  })
})

describe("formatRelativeTime", () => {
  it("uses Intl in the requested locale with no hard-coded words", () => {
    const t = now.getTime()
    expect(formatRelativeTime(new Date(t - 10_000).toISOString(), "en", t)).toBe("now")
    expect(formatRelativeTime(new Date(t - 3 * 3_600_000).toISOString(), "en", t)).toBe("3 hours ago")
    expect(formatRelativeTime(new Date(t - 3 * 3_600_000).toISOString(), "th", t)).toContain("3")
    expect(formatRelativeTime("nope", "en", t)).toBe("")
  })
})

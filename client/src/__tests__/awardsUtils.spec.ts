import { describe, expect, it } from "vitest"
import type { EventAward } from "@/types/event"
import { groupAwards } from "@/utils/awards"
import { EVENT_COVER_SIZE, eventCoverUrl, formatAwardPeriod } from "@/utils/events"

const award = (over: Partial<EventAward>): EventAward => ({
  event_id: 1,
  kind: "monthly",
  period: "2026-11",
  media_id: 10,
  media_type: "movie",
  media_title: "Title",
  poster_path: null,
  backdrop_path: null,
  votes: 3,
  ...over,
})

describe("groupAwards", () => {
  it("puts yearly sections first, then monthly by newest period", () => {
    const groups = groupAwards([
      award({ kind: "monthly", period: "2026-09", media_id: 1 }),
      award({ kind: "yearly", period: "2025", media_id: 2 }),
      award({ kind: "monthly", period: "2026-11", media_id: 3 }),
      award({ kind: "yearly", period: "2026", media_id: 4 }),
    ])
    expect(groups.map((g) => `${g.kind}:${g.period}`)).toEqual([
      "yearly:2026",
      "yearly:2025",
      "monthly:2026-11",
      "monthly:2026-09",
    ])
  })

  it("keeps co-winners of one round together", () => {
    const groups = groupAwards([
      award({ period: "2026-11", media_id: 1 }),
      award({ period: "2026-11", media_id: 2 }),
    ])
    expect(groups).toHaveLength(1)
    expect(groups[0]!.awards.map((a) => a.media_id)).toEqual([1, 2])
  })

  it("returns no sections for no winners", () => {
    expect(groupAwards([])).toEqual([])
  })
})

describe("formatAwardPeriod", () => {
  it("formats monthly periods by the month they start", () => {
    expect(formatAwardPeriod("2026-11", "en", "long")).toBe("November 2026")
    expect(formatAwardPeriod("2026-11", "en", "short")).toBe("Nov 26")
  })

  it("formats yearly periods as the year only", () => {
    expect(formatAwardPeriod("2026", "en", "long")).toBe("2026")
    expect(formatAwardPeriod("2026", "en", "short")).toBe("2026")
  })

  it("uses the Buddhist era for Thai", () => {
    expect(formatAwardPeriod("2026", "th", "long")).toBe("พ.ศ. 2569")
  })
})

describe("eventCoverUrl", () => {
  it("prefers the uploaded image", () => {
    expect(eventCoverUrl({ image_url: "https://cdn.example/a.jpg", cover_backdrop_path: "/b.jpg" })).toBe(
      "https://cdn.example/a.jpg",
    )
  })

  it("falls back to the TMDB backdrop at the cover size", () => {
    expect(eventCoverUrl({ image_url: null, cover_backdrop_path: "/b.jpg" })).toBe(
      `https://image.tmdb.org/t/p/${EVENT_COVER_SIZE}/b.jpg`,
    )
  })

  it("returns null when there is neither image", () => {
    expect(eventCoverUrl({ image_url: null, cover_backdrop_path: null })).toBeNull()
  })
})
